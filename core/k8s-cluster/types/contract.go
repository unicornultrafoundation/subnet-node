package types

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// BidTrackerInterface defines the interface for bid tracking operations
type BidTrackerInterface interface {
	// Bid management
	AddBid(ctx context.Context, deploymentID string, provider common.Address, amount *big.Int, duration time.Duration) error
	GetBids(ctx context.Context, deploymentID string) ([]*BidSubmittedEvent, error)
	GetBid(ctx context.Context, deploymentID string, provider common.Address) (*BidSubmittedEvent, error)
	RemoveBid(ctx context.Context, deploymentID string, provider common.Address) error
	ClearBids(ctx context.Context, deploymentID string) error

	// Bid validation
	ValidateBid(ctx context.Context, deploymentID string, provider common.Address, amount *big.Int, duration time.Duration) error
	IsBidValid(ctx context.Context, deploymentID string, provider common.Address) (bool, error)

	// Bid status
	GetBidStatus(ctx context.Context, deploymentID string, provider common.Address) (BidStatus, error)
	UpdateBidStatus(ctx context.Context, deploymentID string, provider common.Address, status BidStatus) error
}

// BidStatus represents the status of a bid
type BidStatus string

const (
	BidStatusActive   BidStatus = "active"
	BidStatusSelected BidStatus = "selected"
	BidStatusRejected BidStatus = "rejected"
	BidStatusExpired  BidStatus = "expired"
)

// ContractInterface defines the interface for contract operations
type ContractInterface interface {
	// Payment methods
	ParsePaymentReceived(log ethtypes.Log) (*PaymentReceivedEvent, error)

	// Event subscription methods
	SubscribeToDeploymentRequested(ctx context.Context, handler func(*DeploymentRequestedEvent) error) error
	SubscribeToBidSubmitted(ctx context.Context, handler func(*BidSubmittedEvent) error) error
	SubscribeToProviderSelected(ctx context.Context, handler func(*ProviderSelectedEvent) error) error
	SubscribeToDeploymentCompleted(ctx context.Context, handler func(*DeploymentCompletedEvent) error) error
	SubscribeToDeploymentTerminated(ctx context.Context, handler func(*DeploymentTerminatedEvent) error) error

	// Contract call methods
	RequestDeployment(ctx context.Context, opts *bind.TransactOpts, deploymentID string, sdlHash string, minBid *big.Int, maxBid *big.Int, duration *big.Int) (*ethtypes.Transaction, error)
	SubmitBid(ctx context.Context, opts *bind.TransactOpts, deploymentID string, amount *big.Int, duration *big.Int) (*ethtypes.Transaction, error)
	SelectProvider(ctx context.Context, opts *bind.TransactOpts, deploymentID string, provider common.Address, bidIndex *big.Int) (*ethtypes.Transaction, error)
	CompleteDeployment(ctx context.Context, opts *bind.TransactOpts, deploymentID string) (*ethtypes.Transaction, error)
	TerminateDeployment(ctx context.Context, opts *bind.TransactOpts, deploymentID string) (*ethtypes.Transaction, error)
	LockPayment(ctx context.Context, opts *bind.TransactOpts, deploymentID string, amount *big.Int, provider common.Address) (*ethtypes.Transaction, error)
	ReleasePayment(ctx context.Context, opts *bind.TransactOpts, deploymentID string) (*ethtypes.Transaction, error)
	RefundPayment(ctx context.Context, opts *bind.TransactOpts, deploymentID string) (*ethtypes.Transaction, error)
	InitiateDispute(ctx context.Context, opts *bind.TransactOpts, deploymentID string, reason string) (*ethtypes.Transaction, error)
	ReleaseEscrow(ctx context.Context, opts *bind.TransactOpts, deploymentID string) (*ethtypes.Transaction, error)
	RefundEscrow(ctx context.Context, opts *bind.TransactOpts, deploymentID string) (*ethtypes.Transaction, error)

	// Query methods
	GetDeployment(ctx context.Context, deploymentID string) (*Deployment, error)
	GetBids(ctx context.Context, deploymentID string) ([]*BidSubmittedEvent, error)
	GetDeploymentRequestedEvents(ctx context.Context, fromBlock, toBlock uint64) ([]*DeploymentRequestedEvent, error)
	GetBidSubmittedEvents(ctx context.Context, fromBlock, toBlock uint64) ([]*BidSubmittedEvent, error)
	GetProviderSelectedEvents(ctx context.Context, fromBlock, toBlock uint64) ([]*ProviderSelectedEvent, error)
	GetDeploymentCompletedEvents(ctx context.Context, fromBlock, toBlock uint64) ([]*DeploymentCompletedEvent, error)
	GetDeploymentTerminatedEvents(ctx context.Context, fromBlock, toBlock uint64) ([]*DeploymentTerminatedEvent, error)
}

// ContractConfig holds the contract configuration
type ContractConfig struct {
	Client       *ethclient.Client
	ContractAddr common.Address
	BidTracker   BidTrackerInterface
	PaymentMgr   PaymentManagerInterface
	Logger       LoggerInterface
}

// LoggerInterface defines the interface for logging operations
type LoggerInterface interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	With(fields ...interface{}) LoggerInterface
}

// ContractError represents a contract operation error
type ContractError struct {
	Code    string
	Message string
	Err     error
}

func (e *ContractError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Common error codes
const (
	ErrCodeInvalidInput      = "INVALID_INPUT"
	ErrCodeTransactionFailed = "TRANSACTION_FAILED"
	ErrCodeEventNotFound     = "EVENT_NOT_FOUND"
	ErrCodeBidInvalid        = "BID_INVALID"
	ErrCodePaymentFailed     = "PAYMENT_FAILED"
	ErrCodeDeploymentFailed  = "DEPLOYMENT_FAILED"
)

// PaymentReceivedEvent represents a payment received event
type PaymentReceivedEvent struct {
	DeploymentID string
	Requester    common.Address
	Provider     common.Address
	Amount       *big.Int
}

func (e *PaymentReceivedEvent) GetDeploymentID() string      { return e.DeploymentID }
func (e *PaymentReceivedEvent) GetRequester() common.Address { return e.Requester }
func (e *PaymentReceivedEvent) GetProvider() common.Address  { return e.Provider }
func (e *PaymentReceivedEvent) GetAmount() *big.Int          { return e.Amount }
func (e *PaymentReceivedEvent) GetDuration() uint64          { return 0 }
