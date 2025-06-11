package marketplace

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// PaymentState represents the state of a payment
type PaymentState string

const (
	PaymentStatePending   PaymentState = "pending"
	PaymentStateLocked    PaymentState = "locked"
	PaymentStateReleased  PaymentState = "released"
	PaymentStateRefunded  PaymentState = "refunded"
	PaymentStateDisputed  PaymentState = "disputed"
	PaymentStateCompleted PaymentState = "completed"
)

// Payment represents a payment transaction
type Payment struct {
	DeploymentID string
	Amount       *big.Int
	State        PaymentState
	Provider     common.Address
	Client       common.Address
	LockTime     time.Time
	ReleaseTime  time.Time
	RefundTime   time.Time
	TxHash       common.Hash
}

// PaymentManager handles payment processing and escrow management
type PaymentManager struct {
	contract *MarketplaceContract
	mu       sync.RWMutex
	payments map[string]*Payment // deploymentID -> payment
}

// NewPaymentManager creates a new payment manager
func NewPaymentManager(contract *MarketplaceContract) *PaymentManager {
	return &PaymentManager{
		contract: contract,
		payments: make(map[string]*Payment),
	}
}

// LockPayment locks funds in escrow for a deployment
func (pm *PaymentManager) LockPayment(ctx context.Context, deploymentID string, amount *big.Int, provider common.Address) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if payment already exists
	if _, exists := pm.payments[deploymentID]; exists {
		return errors.New("payment already exists for deployment")
	}

	// Create payment record
	payment := &Payment{
		DeploymentID: deploymentID,
		Amount:       amount,
		State:        PaymentStatePending,
		Provider:     provider,
		LockTime:     time.Now(),
	}

	// Lock funds in contract
	tx, err := pm.contract.LockPayment(ctx, nil, deploymentID, amount, provider)
	if err != nil {
		return err
	}

	// Update payment state
	payment.State = PaymentStateLocked
	payment.TxHash = tx.Hash()
	pm.payments[deploymentID] = payment

	return nil
}

// ReleasePayment releases funds from escrow to the provider
func (pm *PaymentManager) ReleasePayment(ctx context.Context, deploymentID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	payment, exists := pm.payments[deploymentID]
	if !exists {
		return errors.New("payment not found")
	}

	if payment.State != PaymentStateLocked {
		return errors.New("payment not in locked state")
	}

	// Release funds from contract
	tx, err := pm.contract.ReleasePayment(ctx, nil, deploymentID)
	if err != nil {
		return err
	}

	// Update payment state
	payment.State = PaymentStateReleased
	payment.ReleaseTime = time.Now()
	payment.TxHash = tx.Hash()

	return nil
}

// RefundPayment refunds funds from escrow to the client
func (pm *PaymentManager) RefundPayment(ctx context.Context, deploymentID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	payment, exists := pm.payments[deploymentID]
	if !exists {
		return errors.New("payment not found")
	}

	if payment.State != PaymentStateLocked {
		return errors.New("payment not in locked state")
	}

	// Refund funds from contract
	tx, err := pm.contract.RefundPayment(ctx, nil, deploymentID)
	if err != nil {
		return err
	}

	// Update payment state
	payment.State = PaymentStateRefunded
	payment.RefundTime = time.Now()
	payment.TxHash = tx.Hash()

	return nil
}

// InitiateDispute initiates a dispute for a payment
func (pm *PaymentManager) InitiateDispute(ctx context.Context, deploymentID string, reason string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	payment, exists := pm.payments[deploymentID]
	if !exists {
		return errors.New("payment not found")
	}

	if payment.State != PaymentStateLocked {
		return errors.New("payment not in locked state")
	}

	// Initiate dispute in contract
	tx, err := pm.contract.InitiateDispute(ctx, nil, deploymentID, reason)
	if err != nil {
		return err
	}

	// Update payment state
	payment.State = PaymentStateDisputed
	payment.TxHash = tx.Hash()

	return nil
}

// GetPayment returns the payment for a deployment
func (pm *PaymentManager) GetPayment(deploymentID string) (*Payment, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	payment, exists := pm.payments[deploymentID]
	if !exists {
		return nil, errors.New("payment not found")
	}

	return payment, nil
}

// GetPaymentsByProvider returns all payments for a provider
func (pm *PaymentManager) GetPaymentsByProvider(provider common.Address) []*Payment {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var payments []*Payment
	for _, payment := range pm.payments {
		if payment.Provider == provider {
			payments = append(payments, payment)
		}
	}

	return payments
}

// GetPaymentsByClient returns all payments for a client
func (pm *PaymentManager) GetPaymentsByClient(client common.Address) []*Payment {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var payments []*Payment
	for _, payment := range pm.payments {
		if payment.Client == client {
			payments = append(payments, payment)
		}
	}

	return payments
}

// GetPaymentsByState returns all payments in a specific state
func (pm *PaymentManager) GetPaymentsByState(state PaymentState) []*Payment {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var payments []*Payment
	for _, payment := range pm.payments {
		if payment.State == state {
			payments = append(payments, payment)
		}
	}

	return payments
}
