package payment

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/events"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
)

// MockPaymentManager implements PaymentManagerInterface for local/demo runs
type MockPaymentManager struct {
	store     map[string]*types.Escrow
	eventBus  *events.DefaultEventBus[interface{}]
	mu        sync.RWMutex
	stopCh    chan struct{}
	isHealthy bool
}

// NewMockPaymentManager creates a new mock payment manager
func NewMockPaymentManager(eventBus *events.DefaultEventBus[interface{}]) *MockPaymentManager {
	return &MockPaymentManager{
		store:     make(map[string]*types.Escrow),
		eventBus:  eventBus,
		stopCh:    make(chan struct{}),
		isHealthy: true,
	}
}

// Start starts the mock payment manager
func (m *MockPaymentManager) Start(ctx context.Context) error {
	// In mock mode, we don't need to do anything
	return nil
}

// Stop stops the mock payment manager
func (m *MockPaymentManager) Stop() {
	close(m.stopCh)
}

// IsHealthy returns true if the payment manager is healthy
func (m *MockPaymentManager) IsHealthy() bool {
	return m.isHealthy
}

// GetPayment returns a payment for a deployment
func (m *MockPaymentManager) GetPayment(deploymentID string) (*types.Payment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	escrow, exists := m.store[deploymentID]
	if !exists {
		return nil, fmt.Errorf("payment not found for deployment %s", deploymentID)
	}
	return &types.Payment{
		DeploymentID: deploymentID,
		Status:       escrow.Status,
		Amount:       escrow.Amount,
	}, nil
}

// GetPaymentHistory returns payment history for a deployment
func (m *MockPaymentManager) GetPaymentHistory(deploymentID string) ([]*types.Payment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	escrow, exists := m.store[deploymentID]
	if !exists {
		return nil, fmt.Errorf("payment not found for deployment %s", deploymentID)
	}
	return []*types.Payment{{
		DeploymentID: deploymentID,
		Status:       escrow.Status,
		Amount:       escrow.Amount,
	}}, nil
}

// GetEscrowHistory returns escrow history for a deployment
func (m *MockPaymentManager) GetEscrowHistory(deploymentID string) ([]*types.Escrow, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	escrow, exists := m.store[deploymentID]
	if !exists {
		return nil, fmt.Errorf("escrow not found for deployment %s", deploymentID)
	}
	return []*types.Escrow{escrow}, nil
}

// HandleDeploymentStatusChange handles deployment status changes
func (m *MockPaymentManager) HandleDeploymentStatusChange(ctx context.Context, deploymentID string, status types.DeploymentStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create or update escrow record
	escrow, exists := m.store[deploymentID]
	if !exists {
		escrow = &types.Escrow{
			DeploymentID: deploymentID,
			Status:       string(types.PaymentStatusPending),
			CreatedAt:    time.Now(),
		}
		m.store[deploymentID] = escrow
	}

	// Update escrow status based on deployment status
	switch status {
	case types.DeploymentStatusRunning:
		escrow.Status = string(types.PaymentStatusPending)
	case types.DeploymentStatusCompleted:
		escrow.Status = string(types.PaymentStatusCompleted)
	case types.DeploymentStatusTerminated:
		escrow.Status = string(types.PaymentStatusRefunded)
	default:
		escrow.Status = string(types.PaymentStatusPending)
	}

	escrow.UpdatedAt = time.Now()

	// Publish appropriate event
	switch escrow.Status {
	case string(types.PaymentStatusCompleted):
		m.eventBus.Publish(ctx, "PaymentCompleted", &types.PaymentReleasedEvent{
			BaseEvent: types.BaseEvent{
				Timestamp: time.Now(),
			},
			DeploymentID: deploymentID,
			Amount:       escrow.Amount,
		})
	case string(types.PaymentStatusRefunded):
		m.eventBus.Publish(ctx, "PaymentRefunded", &types.PaymentRefundedEvent{
			BaseEvent: types.BaseEvent{
				Timestamp: time.Now(),
			},
			DeploymentID: deploymentID,
			Amount:       escrow.Amount,
			Reason:       "Deployment terminated",
			RefundedAt:   time.Now(),
		})
	}

	return nil
}

// GetEscrow returns escrow information for a deployment
func (m *MockPaymentManager) GetEscrow(deploymentID string) (*types.Escrow, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	escrow, exists := m.store[deploymentID]
	if !exists {
		return nil, fmt.Errorf("escrow not found for deployment %s", deploymentID)
	}

	return escrow, nil
}
