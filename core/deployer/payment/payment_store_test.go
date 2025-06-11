package payment

import (
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
)

func TestPaymentStore(t *testing.T) {
	// Create temporary directory for store
	storeDir, err := os.MkdirTemp("", "payment-store-test")
	require.NoError(t, err)
	defer os.RemoveAll(storeDir)

	// Create store
	store, err := NewPaymentStore(storeDir)
	require.NoError(t, err)
	require.NotNil(t, store)

	// Create test payment
	payment := &types.Payment{
		DeploymentID: "test-deployment",
		Requester:    common.HexToAddress("0x123"),
		Provider:     common.HexToAddress("0x456"),
		Amount:       big.NewInt(500),
		Status:       "pending",
		TxHash:       common.HexToHash("0xabc"),
		BlockNumber:  1,
		Timestamp:    time.Now(),
		Version:      1,
	}

	// Test storing payment
	err = store.StorePayment(payment)
	require.NoError(t, err)

	// Test getting payment
	retrieved, err := store.GetPayment(payment.DeploymentID)
	require.NoError(t, err)
	assert.Equal(t, payment.DeploymentID, retrieved.DeploymentID)
	assert.Equal(t, payment.Requester, retrieved.Requester)
	assert.Equal(t, payment.Provider, retrieved.Provider)
	assert.Equal(t, payment.Amount, retrieved.Amount)
	assert.Equal(t, payment.Status, retrieved.Status)
	assert.Equal(t, payment.TxHash, retrieved.TxHash)
	assert.Equal(t, payment.BlockNumber, retrieved.BlockNumber)
	assert.Equal(t, payment.Timestamp.Unix(), retrieved.Timestamp.Unix())
	assert.Equal(t, payment.Version, retrieved.Version)

	// Test payment history
	history, err := store.GetPaymentHistory(payment.DeploymentID)
	require.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, payment.DeploymentID, history[0].DeploymentID)
	assert.Equal(t, payment.Amount, history[0].Amount)
	assert.Equal(t, payment.Status, history[0].Status)

	// Create test escrow
	escrow := &types.Escrow{
		DeploymentID: "test-deployment",
		Requester:    common.HexToAddress("0x123"),
		Provider:     common.HexToAddress("0x456"),
		Amount:       big.NewInt(500),
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Version:      1,
	}

	// Test storing escrow
	err = store.StoreEscrow(escrow)
	require.NoError(t, err)

	// Test getting escrow
	retrievedEscrow, err := store.GetEscrow(escrow.DeploymentID)
	require.NoError(t, err)
	assert.Equal(t, escrow.DeploymentID, retrievedEscrow.DeploymentID)
	assert.Equal(t, escrow.Requester, retrievedEscrow.Requester)
	assert.Equal(t, escrow.Provider, retrievedEscrow.Provider)
	assert.Equal(t, escrow.Amount, retrievedEscrow.Amount)
	assert.Equal(t, escrow.Status, retrievedEscrow.Status)
	assert.Equal(t, escrow.CreatedAt.Unix(), retrievedEscrow.CreatedAt.Unix())
	assert.Equal(t, escrow.UpdatedAt.Unix(), retrievedEscrow.UpdatedAt.Unix())
	assert.Equal(t, escrow.Version, retrievedEscrow.Version)

	// Test escrow history
	escrowHistory, err := store.GetEscrowHistory(escrow.DeploymentID)
	require.NoError(t, err)
	assert.Len(t, escrowHistory, 1)
	assert.Equal(t, escrow.DeploymentID, escrowHistory[0].DeploymentID)
	assert.Equal(t, escrow.Amount, escrowHistory[0].Amount)
	assert.Equal(t, escrow.Status, escrowHistory[0].Status)

	// Test active escrows
	active, err := store.GetActiveEscrows()
	require.NoError(t, err)
	assert.Len(t, active, 1)
	assert.Equal(t, escrow.DeploymentID, active[0].DeploymentID)
	assert.Equal(t, escrow.Status, active[0].Status)
}

func TestPaymentStoreConcurrent(t *testing.T) {
	// Create temporary directory for store
	storeDir, err := os.MkdirTemp("", "payment-store-concurrent-test")
	require.NoError(t, err)
	defer os.RemoveAll(storeDir)

	// Create store
	store, err := NewPaymentStore(storeDir)
	require.NoError(t, err)
	require.NotNil(t, store)

	// Create multiple test payments
	payments := make([]*types.Payment, 10)
	for i := 0; i < 10; i++ {
		payments[i] = &types.Payment{
			DeploymentID: fmt.Sprintf("test-deployment-%d", i),
			Requester:    common.HexToAddress("0x123"),
			Provider:     common.HexToAddress("0x456"),
			Amount:       big.NewInt(500),
			Status:       "pending",
			TxHash:       common.HexToHash(fmt.Sprintf("0xabc%d", i)),
			BlockNumber:  uint64(i),
			Timestamp:    time.Now(),
			Version:      1,
		}
	}

	// Test concurrent store operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(i int) {
			err := store.StorePayment(payments[i])
			require.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all payments were stored
	for i := 0; i < 10; i++ {
		retrieved, err := store.GetPayment(payments[i].DeploymentID)
		require.NoError(t, err)
		assert.Equal(t, payments[i].DeploymentID, retrieved.DeploymentID)
		assert.Equal(t, payments[i].Amount, retrieved.Amount)
		assert.Equal(t, payments[i].Status, retrieved.Status)
	}
}

func TestPaymentStoreInvalid(t *testing.T) {
	// Create temporary directory for store
	storeDir, err := os.MkdirTemp("", "payment-store-invalid-test")
	require.NoError(t, err)
	defer os.RemoveAll(storeDir)

	// Create store
	store, err := NewPaymentStore(storeDir)
	require.NoError(t, err)
	require.NotNil(t, store)

	// Test getting non-existent payment
	_, err = store.GetPayment("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no payment found")

	// Test getting non-existent escrow
	_, err = store.GetEscrow("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no escrow found")

	// Test getting payment history for non-existent deployment
	_, err = store.GetPaymentHistory("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no payment found")

	// Test getting escrow history for non-existent deployment
	_, err = store.GetEscrowHistory("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no escrow found")
}
