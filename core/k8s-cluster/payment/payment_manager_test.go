package payment

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/events"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
)

func TestPaymentManager(t *testing.T) {
	// Create temporary directory for store
	storeDir, err := os.MkdirTemp("", "payment-manager-test")
	require.NoError(t, err)
	defer os.RemoveAll(storeDir)

	// Create test deployment
	dep := &types.ManagedDeployment{
		ID:        "test-deployment",
		Requester: common.HexToAddress("0x123"),
		Status:    types.DeploymentStatusRunning,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create event bus
	eventBus := events.NewEventBus[types.MarketplaceEvent]()

	// Create payment manager
	manager, err := NewPaymentManager(&Config{
		EthNodeURL:     "http://localhost:8545",
		ContractAddr:   common.HexToAddress("0x789"),
		StoreDir:       storeDir,
		CheckInterval:  time.Second,
		ProviderAddr:   common.HexToAddress("0x456"),
		PrivateKeyPath: "/tmp/private.key",
	}, eventBus.(*events.DefaultEventBus[types.MarketplaceEvent]))
	require.NoError(t, err)
	require.NotNil(t, manager)

	// Start manager
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = manager.Start(ctx)
	require.NoError(t, err)

	// Create test payment
	payment := &types.Payment{
		DeploymentID: dep.ID,
		Requester:    dep.Requester,
		Provider:     common.HexToAddress("0x456"),
		Amount:       big.NewInt(500),
		Status:       "pending",
		TxHash:       common.HexToHash("0xabc"),
		BlockNumber:  1,
		Timestamp:    time.Now(),
		Version:      1,
	}

	// Test storing payment
	err = manager.store.StorePayment(payment)
	require.NoError(t, err)

	// Test getting payment
	retrieved, err := manager.GetPayment(dep.ID)
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
	history, err := manager.GetPaymentHistory(dep.ID)
	require.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, payment.DeploymentID, history[0].DeploymentID)
	assert.Equal(t, payment.Amount, history[0].Amount)
	assert.Equal(t, payment.Status, history[0].Status)

	// Stop manager
	manager.Stop()
}

func TestPaymentManagerEscrow(t *testing.T) {
	// Create temporary directory for store
	storeDir, err := os.MkdirTemp("", "payment-manager-escrow-test")
	require.NoError(t, err)
	defer os.RemoveAll(storeDir)

	// Create test deployment
	dep := &types.ManagedDeployment{
		ID:           "test-deployment",
		Requester:    common.HexToAddress("0x123"),
		Status:       types.DeploymentStatusRunning,
		Version:      1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		HealthStatus: types.HealthStatusUnknown,
	}

	// Create event bus
	eventBus := events.NewEventBus[types.MarketplaceEvent]()

	// Create payment manager
	manager, err := NewPaymentManager(&Config{
		EthNodeURL:     "http://localhost:8545",
		ContractAddr:   common.HexToAddress("0x789"),
		StoreDir:       storeDir,
		CheckInterval:  time.Second,
		ProviderAddr:   common.HexToAddress("0x456"),
		PrivateKeyPath: "/tmp/private.key",
	}, eventBus.(*events.DefaultEventBus[types.MarketplaceEvent]))
	require.NoError(t, err)
	require.NotNil(t, manager)

	// Start manager
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = manager.Start(ctx)
	require.NoError(t, err)

	// Create test escrow
	escrow := &types.Escrow{
		DeploymentID: dep.ID,
		Requester:    dep.Requester,
		Provider:     common.HexToAddress("0x456"),
		Amount:       big.NewInt(500),
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Version:      1,
	}

	// Test storing escrow
	err = manager.store.StoreEscrow(escrow)
	require.NoError(t, err)

	// Test getting escrow
	retrieved, err := manager.GetEscrow(dep.ID)
	require.NoError(t, err)
	assert.Equal(t, escrow.DeploymentID, retrieved.DeploymentID)
	assert.Equal(t, escrow.Requester, retrieved.Requester)
	assert.Equal(t, escrow.Provider, retrieved.Provider)
	assert.Equal(t, escrow.Amount, retrieved.Amount)
	assert.Equal(t, escrow.Status, retrieved.Status)
	assert.Equal(t, escrow.CreatedAt.Unix(), retrieved.CreatedAt.Unix())
	assert.Equal(t, escrow.UpdatedAt.Unix(), retrieved.UpdatedAt.Unix())
	assert.Equal(t, escrow.Version, retrieved.Version)

	// Test escrow history
	history, err := manager.GetEscrowHistory(dep.ID)
	require.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, escrow.DeploymentID, history[0].DeploymentID)
	assert.Equal(t, escrow.Amount, history[0].Amount)
	assert.Equal(t, escrow.Status, history[0].Status)

	// Test active escrows
	active, err := manager.store.GetActiveEscrows()
	require.NoError(t, err)
	assert.Len(t, active, 1)
	assert.Equal(t, escrow.DeploymentID, active[0].DeploymentID)
	assert.Equal(t, escrow.Status, active[0].Status)

	// Stop manager
	manager.Stop()
}

func TestPaymentManagerInvalid(t *testing.T) {
	// Create temporary directory for store
	storeDir, err := os.MkdirTemp("", "payment-manager-invalid-test")
	require.NoError(t, err)
	defer os.RemoveAll(storeDir)

	// Create event bus
	eventBus := events.NewEventBus[types.MarketplaceEvent]()

	// Create payment manager
	manager, err := NewPaymentManager(&Config{
		EthNodeURL:     "http://localhost:8545",
		ContractAddr:   common.HexToAddress("0x789"),
		StoreDir:       storeDir,
		CheckInterval:  time.Second,
		ProviderAddr:   common.HexToAddress("0x456"),
		PrivateKeyPath: "/tmp/private.key",
	}, eventBus.(*events.DefaultEventBus[types.MarketplaceEvent]))
	require.NoError(t, err)
	require.NotNil(t, manager)

	// Test getting non-existent payment
	_, err = manager.GetPayment("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no payment found")

	// Test getting non-existent escrow
	_, err = manager.GetEscrow("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no escrow found")

	// Test getting payment history for non-existent deployment
	_, err = manager.GetPaymentHistory("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no payment found")

	// Test getting escrow history for non-existent deployment
	_, err = manager.GetEscrowHistory("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no escrow found")
}
