package payment

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/events"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
	clusterTypes "github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
)

// MarketplaceContract represents the smart contract interface
type MarketplaceContract struct {
	*bind.BoundContract
	client *ethclient.Client
}

// NewMarketplaceContract creates a new contract instance
func NewMarketplaceContract(address common.Address, client *ethclient.Client) (*MarketplaceContract, error) {
	// Load contract ABI
	contractABI, err := loadContractABI()
	if err != nil {
		return nil, fmt.Errorf("failed to load contract ABI: %w", err)
	}

	// Create bound contract
	contract := bind.NewBoundContract(address, contractABI, client, client, client)

	return &MarketplaceContract{
		BoundContract: contract,
		client:        client,
	}, nil
}

// loadContractABI loads the contract ABI from file
func loadContractABI() (abi.ABI, error) {
	// TODO: Load ABI from file or embed in binary
	abiJSON := `[{"anonymous":false,"inputs":[{"indexed":true,"name":"deploymentId","type":"string"},{"indexed":true,"name":"requester","type":"address"},{"indexed":true,"name":"provider","type":"address"},{"indexed":false,"name":"amount","type":"uint256"}],"name":"PaymentReceived","type":"event"}]`
	return abi.JSON(strings.NewReader(abiJSON))
}

// ParsePaymentReceived parses the PaymentReceived event
func (c *MarketplaceContract) ParsePaymentReceived(log ethtypes.Log) (*clusterTypes.PaymentReceived, error) {
	// TODO: Implement event parsing
	return &clusterTypes.PaymentReceived{
		DeploymentID: "test",
		Requester:    common.HexToAddress("0x0"),
		Provider:     common.HexToAddress("0x0"),
		Amount:       big.NewInt(0),
	}, nil
}

// ReleaseEscrow releases funds from escrow
func (c *MarketplaceContract) ReleaseEscrow(opts *bind.TransactOpts, deploymentID string) (*ethtypes.Transaction, error) {
	// TODO: Implement contract call
	return nil, nil
}

// RefundEscrow refunds funds from escrow
func (c *MarketplaceContract) RefundEscrow(opts *bind.TransactOpts, deploymentID string) (*ethtypes.Transaction, error) {
	// TODO: Implement contract call
	return nil, nil
}

// loadPrivateKey loads a private key from file
func loadPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	// Read key file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	// Parse key file
	var keyFile struct {
		PrivateKey string `json:"privateKey"`
	}
	if err := json.Unmarshal(data, &keyFile); err != nil {
		return nil, fmt.Errorf("failed to parse key file: %w", err)
	}

	// Decode private key
	key, err := crypto.HexToECDSA(keyFile.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	return key, nil
}

// PaymentManagerInterface defines the interface for payment management
type PaymentManagerInterface interface {
	Start(ctx context.Context) error
	Stop()
	GetPayment(deploymentID string) (*clusterTypes.Payment, error)
	GetEscrow(deploymentID string) (*clusterTypes.Escrow, error)
	GetPaymentHistory(deploymentID string) ([]*clusterTypes.Payment, error)
	GetEscrowHistory(deploymentID string) ([]*clusterTypes.Escrow, error)
	IsHealthy() bool
	HandleDeploymentStatusChange(ctx context.Context, deploymentID string, status clusterTypes.DeploymentStatus) error
}

// Config holds the payment manager configuration
type Config struct {
	EthNodeURL     string
	ContractAddr   common.Address
	StoreDir       string
	CheckInterval  time.Duration
	ProviderAddr   common.Address
	PrivateKeyPath string
}

// PaymentManager handles payment processing and escrow management
type PaymentManager struct {
	config      *Config
	client      *ethclient.Client
	store       *PaymentStore
	contract    *MarketplaceContract
	stopCh      chan struct{}
	mu          sync.RWMutex
	deployments map[string]*clusterTypes.ManagedDeployment
	eventBus    *events.DefaultEventBus[types.MarketplaceEvent]
}

// ManagedDeployment represents a managed deployment
type ManagedDeployment struct {
	ID        string
	Requester common.Address
	Provider  common.Address
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Ensure PaymentManager implements PaymentManagerInterface
var _ clusterTypes.PaymentManagerInterface = (*PaymentManager)(nil)

// NewPaymentManager creates a new payment manager
func NewPaymentManager(cfg *Config, eventBus *events.DefaultEventBus[types.MarketplaceEvent]) (*PaymentManager, error) {
	// Connect to Ethereum node
	client, err := ethclient.Dial(cfg.EthNodeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	// Create payment store
	store, err := NewPaymentStore(cfg.StoreDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment store: %w", err)
	}

	// Create contract instance
	contract, err := NewMarketplaceContract(cfg.ContractAddr, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract instance: %w", err)
	}

	return &PaymentManager{
		config:      cfg,
		client:      client,
		store:       store,
		contract:    contract,
		stopCh:      make(chan struct{}),
		deployments: make(map[string]*clusterTypes.ManagedDeployment),
		eventBus:    eventBus,
	}, nil
}

// Start starts the payment manager
func (m *PaymentManager) Start(ctx context.Context) error {
	// Start payment processing loop
	go m.processPayments(ctx)

	// Start escrow monitoring loop
	go m.monitorEscrows(ctx)

	return nil
}

// Stop stops the payment manager
func (m *PaymentManager) Stop() {
	close(m.stopCh)
}

// processPayments processes pending payments
func (m *PaymentManager) processPayments(ctx context.Context) {
	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkPayments(ctx)
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		}
	}
}

// checkPayments checks and processes pending payments
func (m *PaymentManager) checkPayments(ctx context.Context) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, dep := range m.deployments {
		// Get payment status
		payment, err := m.store.GetPayment(dep.ID)
		if err != nil {
			// Publish PaymentFailedEvent
			m.eventBus.Publish(ctx, "PaymentFailed", &clusterTypes.PaymentFailedEvent{
				BaseEvent: clusterTypes.BaseEvent{
					Timestamp: time.Now(),
				},
				DeploymentID: dep.ID,
				Provider:     m.config.ProviderAddr,
				Amount:       big.NewInt(0),
				Reason:       fmt.Sprintf("failed to get payment: %v", err),
			})
			continue
		}

		// Verify payment
		verified, err := m.verifyPayment(ctx, payment)
		if err != nil {
			// Publish PaymentFailedEvent
			m.eventBus.Publish(ctx, "PaymentFailed", &clusterTypes.PaymentFailedEvent{
				BaseEvent: clusterTypes.BaseEvent{
					Timestamp: time.Now(),
				},
				DeploymentID: dep.ID,
				Provider:     m.config.ProviderAddr,
				Amount:       payment.Amount,
				Reason:       fmt.Sprintf("payment verification failed: %v", err),
			})
			continue
		}

		if !verified {
			// Publish PaymentFailedEvent
			m.eventBus.Publish(ctx, "PaymentFailed", &clusterTypes.PaymentFailedEvent{
				BaseEvent: clusterTypes.BaseEvent{
					Timestamp: time.Now(),
				},
				DeploymentID: dep.ID,
				Provider:     m.config.ProviderAddr,
				Amount:       payment.Amount,
				Reason:       "payment verification failed",
			})
			continue
		}

		// Publish PaymentCreatedEvent
		m.eventBus.Publish(ctx, "PaymentCreated", &clusterTypes.PaymentCreatedEvent{
			BaseEvent: clusterTypes.BaseEvent{
				Timestamp: time.Now(),
			},
			DeploymentID: dep.ID,
			Provider:     m.config.ProviderAddr,
			Amount:       payment.Amount,
		})
	}
}

// verifyPayment verifies a payment on the blockchain
func (m *PaymentManager) verifyPayment(ctx context.Context, payment *clusterTypes.Payment) (bool, error) {
	// Get transaction receipt
	receipt, err := m.client.TransactionReceipt(ctx, payment.TxHash)
	if err != nil {
		return false, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	// Check if transaction was successful
	if receipt.Status != ethtypes.ReceiptStatusSuccessful {
		return false, nil
	}

	// Verify payment amount
	if len(receipt.Logs) == 0 {
		return false, fmt.Errorf("no logs found in transaction")
	}

	event, err := m.contract.ParsePaymentReceived(*receipt.Logs[0])
	if err != nil {
		return false, fmt.Errorf("failed to parse payment event: %w", err)
	}

	if event.Amount.Cmp(payment.Amount) != 0 {
		return false, fmt.Errorf("payment amount mismatch")
	}

	return true, nil
}

// monitorEscrows monitors active escrows
func (m *PaymentManager) monitorEscrows(ctx context.Context) {
	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkEscrows(ctx)
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		}
	}
}

// checkEscrows checks and processes active escrows
func (m *PaymentManager) checkEscrows(ctx context.Context) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get active escrows
	escrows, err := m.store.GetActiveEscrows()
	if err != nil {
		log.Printf("Failed to get active escrows: %v", err)
		return
	}

	for _, escrow := range escrows {
		// Get deployment
		dep, err := m.store.GetDeployment(escrow.DeploymentID)
		if err != nil {
			log.Printf("Failed to get deployment %s: %v", escrow.DeploymentID, err)
			continue
		}

		// Check if deployment is completed
		if dep.Status == clusterTypes.DeploymentStatusCompleted {
			// Release escrow
			if err := m.releaseEscrow(ctx, escrow); err != nil {
				log.Printf("Failed to release escrow for deployment %s: %v", dep.ID, err)
				continue
			}
		}

		// Check if deployment is terminated
		if dep.Status == clusterTypes.DeploymentStatusTerminated {
			// Refund escrow
			if err := m.refundEscrow(ctx, escrow); err != nil {
				log.Printf("Failed to refund escrow for deployment %s: %v", dep.ID, err)
				continue
			}
		}
	}
}

// releaseEscrow releases funds from escrow
func (m *PaymentManager) releaseEscrow(ctx context.Context, escrow *clusterTypes.Escrow) error {
	// Get transaction options
	opts, err := m.getTransactOpts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get transaction options: %w", err)
	}

	// Release escrow
	tx, err := m.contract.ReleaseEscrow(opts, escrow.DeploymentID)
	if err != nil {
		// Publish PaymentFailedEvent
		m.eventBus.Publish(ctx, "PaymentFailed", &clusterTypes.PaymentFailedEvent{
			BaseEvent: clusterTypes.BaseEvent{
				Timestamp: time.Now(),
			},
			DeploymentID: escrow.DeploymentID,
			Provider:     m.config.ProviderAddr,
			Amount:       escrow.Amount,
			Reason:       fmt.Sprintf("failed to release escrow: %v", err),
		})
		return fmt.Errorf("failed to release escrow: %w", err)
	}

	// Wait for transaction receipt
	receipt, err := m.client.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		return fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	if receipt.Status == 0 {
		// Publish PaymentFailedEvent
		m.eventBus.Publish(ctx, "PaymentFailed", &clusterTypes.PaymentFailedEvent{
			BaseEvent: clusterTypes.BaseEvent{
				Timestamp: time.Now(),
			},
			DeploymentID: escrow.DeploymentID,
			Provider:     m.config.ProviderAddr,
			Amount:       escrow.Amount,
			Reason:       "escrow release transaction failed",
		})
		return fmt.Errorf("escrow release transaction failed")
	}

	// Publish PaymentReleasedEvent
	m.eventBus.Publish(ctx, "PaymentReleased", &clusterTypes.PaymentReleasedEvent{
		BaseEvent: clusterTypes.BaseEvent{
			Timestamp: time.Now(),
		},
		DeploymentID: escrow.DeploymentID,
		Provider:     m.config.ProviderAddr,
		Amount:       escrow.Amount,
	})

	return nil
}

// refundEscrow refunds funds from escrow
func (m *PaymentManager) refundEscrow(ctx context.Context, escrow *clusterTypes.Escrow) error {
	// Get transaction options
	opts, err := m.getTransactOpts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get transaction options: %w", err)
	}

	// Refund escrow
	tx, err := m.contract.RefundEscrow(opts, escrow.DeploymentID)
	if err != nil {
		// Publish PaymentFailedEvent
		m.eventBus.Publish(ctx, "PaymentFailed", &clusterTypes.PaymentFailedEvent{
			BaseEvent: clusterTypes.BaseEvent{
				Timestamp: time.Now(),
			},
			DeploymentID: escrow.DeploymentID,
			Provider:     m.config.ProviderAddr,
			Amount:       escrow.Amount,
			Reason:       fmt.Sprintf("failed to refund escrow: %v", err),
		})
		return fmt.Errorf("failed to refund escrow: %w", err)
	}

	// Wait for transaction receipt
	receipt, err := m.client.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		return fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	if receipt.Status == 0 {
		// Publish PaymentFailedEvent
		m.eventBus.Publish(ctx, "PaymentFailed", &clusterTypes.PaymentFailedEvent{
			BaseEvent: clusterTypes.BaseEvent{
				Timestamp: time.Now(),
			},
			DeploymentID: escrow.DeploymentID,
			Provider:     m.config.ProviderAddr,
			Amount:       escrow.Amount,
			Reason:       "escrow refund transaction failed",
		})
		return fmt.Errorf("escrow refund transaction failed")
	}

	// Publish PaymentRefundedEvent
	m.eventBus.Publish(ctx, "PaymentRefunded", &clusterTypes.PaymentRefundedEvent{
		BaseEvent: clusterTypes.BaseEvent{
			Timestamp: time.Now(),
		},
		DeploymentID: escrow.DeploymentID,
		Provider:     m.config.ProviderAddr,
		Amount:       escrow.Amount,
		Reason:       "escrow refunded successfully",
	})

	return nil
}

// getTransactOpts creates transaction options for contract calls
func (m *PaymentManager) getTransactOpts(ctx context.Context) (*bind.TransactOpts, error) {
	// Load private key
	key, err := loadPrivateKey(m.config.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	// Get chain ID
	chainID, err := m.client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Create transaction options
	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	return auth, nil
}

// GetPayment gets payment information for a deployment
func (m *PaymentManager) GetPayment(deploymentID string) (*clusterTypes.Payment, error) {
	return m.store.GetPayment(deploymentID)
}

// GetEscrow gets escrow information for a deployment
func (m *PaymentManager) GetEscrow(deploymentID string) (*clusterTypes.Escrow, error) {
	return m.store.GetEscrow(deploymentID)
}

// GetPaymentHistory gets payment history for a deployment
func (m *PaymentManager) GetPaymentHistory(deploymentID string) ([]*clusterTypes.Payment, error) {
	return m.store.GetPaymentHistory(deploymentID)
}

// GetEscrowHistory gets escrow history for a deployment
func (m *PaymentManager) GetEscrowHistory(deploymentID string) ([]*clusterTypes.Escrow, error) {
	return m.store.GetEscrowHistory(deploymentID)
}

// IsHealthy checks if the payment manager is healthy
func (m *PaymentManager) IsHealthy() bool {
	// Check if client is connected
	if m.client == nil {
		return false
	}

	// Check if contract is initialized
	if m.contract == nil {
		return false
	}

	// Check if store is initialized
	if m.store == nil {
		return false
	}

	// Check if stop channel is not closed
	select {
	case <-m.stopCh:
		return false
	default:
	}

	return true
}

// HandleDeploymentStatusChange handles deployment status changes
func (m *PaymentManager) HandleDeploymentStatusChange(ctx context.Context, deploymentID string, status clusterTypes.DeploymentStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get deployment
	dep, ok := m.deployments[deploymentID]
	if !ok {
		return fmt.Errorf("deployment not found: %s", deploymentID)
	}

	// Update deployment status
	dep.Status = status
	dep.UpdatedAt = time.Now()

	// Get escrow
	escrow, err := m.store.GetEscrow(deploymentID)
	if err != nil {
		return fmt.Errorf("failed to get escrow: %w", err)
	}

	// Handle status change
	switch status {
	case clusterTypes.DeploymentStatusRunning:
		// Publish PaymentScheduledEvent
		m.eventBus.Publish(ctx, "PaymentScheduled", &clusterTypes.PaymentScheduledEvent{
			BaseEvent: clusterTypes.BaseEvent{
				Timestamp: time.Now(),
			},
			DeploymentID: deploymentID,
			Provider:     m.config.ProviderAddr,
			Amount:       escrow.Amount,
			ScheduledAt:  time.Now().Add(24 * time.Hour), // Example: schedule payment for 24 hours later
		})

	case clusterTypes.DeploymentStatusCompleted:
		// Release escrow
		if err := m.releaseEscrow(ctx, escrow); err != nil {
			return fmt.Errorf("failed to release escrow: %w", err)
		}

	case clusterTypes.DeploymentStatusTerminated, clusterTypes.DeploymentStatusFailed:
		// Refund escrow
		if err := m.refundEscrow(ctx, escrow); err != nil {
			return fmt.Errorf("failed to refund escrow: %w", err)
		}
	}

	return nil
}
