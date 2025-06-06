package types

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

// PaymentManagerInterface defines the interface for payment management
type PaymentManagerInterface interface {
	Start(ctx context.Context) error
	Stop()
	GetPayment(deploymentID string) (*Payment, error)
	GetEscrow(deploymentID string) (*Escrow, error)
	GetPaymentHistory(deploymentID string) ([]*Payment, error)
	GetEscrowHistory(deploymentID string) ([]*Escrow, error)
	IsHealthy() bool
	HandleDeploymentStatusChange(ctx context.Context, deploymentID string, status DeploymentStatus) error
}

// Payment represents a payment transaction
type Payment struct {
	DeploymentID string
	Requester    common.Address
	Provider     common.Address
	Amount       *big.Int
	Status       string
	TxHash       common.Hash
	BlockNumber  uint64
	Timestamp    time.Time
	Version      uint64
}

// Escrow represents an escrow account
type Escrow struct {
	DeploymentID string
	Requester    common.Address
	Provider     common.Address
	Amount       *big.Int
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Version      uint64
}

// PaymentStatusInfo represents the payment status of a deployment
type PaymentStatusInfo struct {
	CurrentBalance *big.Int
	LastPayment    time.Time
	PaymentHistory []*Payment
	EscrowBalance  *big.Int
	EscrowHistory  []*Escrow
}

// PaymentReceived represents the PaymentReceived event
type PaymentReceived struct {
	DeploymentID string
	Requester    common.Address
	Provider     common.Address
	Amount       *big.Int
}

// PaymentState represents the state of a payment
type PaymentState string

const (
	PaymentStatePending   PaymentState = "pending"
	PaymentStateCompleted PaymentState = "completed"
	PaymentStateFailed    PaymentState = "failed"
	PaymentStateRefunded  PaymentState = "refunded"
)

// PaymentCreatedEvent represents a payment creation event
type PaymentCreatedEvent struct {
	BaseEvent
	DeploymentID string
	Requester    common.Address
	Provider     common.Address
	Amount       *big.Int
}

// PaymentReleasedEvent represents a payment release event
type PaymentReleasedEvent struct {
	BaseEvent
	DeploymentID string
	Requester    common.Address
	Provider     common.Address
	Amount       *big.Int
}

// PaymentFailedEvent represents a payment failure event
type PaymentFailedEvent struct {
	BaseEvent
	DeploymentID string
	Requester    common.Address
	Provider     common.Address
	Amount       *big.Int
	Reason       string
	FailedAt     time.Time
}

// PaymentRefundedEvent represents a payment refund event
type PaymentRefundedEvent struct {
	BaseEvent
	DeploymentID string
	Requester    common.Address
	Provider     common.Address
	Amount       *big.Int
	Reason       string
	RefundedAt   time.Time
}

// PaymentDisputedEvent represents a payment dispute event
type PaymentDisputedEvent struct {
	DeploymentID string
	Requester    common.Address
	Provider     common.Address
	Amount       *big.Int
	Reason       string
	DisputedAt   time.Time
	Timestamp    time.Time
}

// PaymentDisputeResolvedEvent represents a payment dispute resolution event
type PaymentDisputeResolvedEvent struct {
	DeploymentID string
	Requester    common.Address
	Provider     common.Address
	Amount       *big.Int
	Resolution   string
	ResolvedAt   time.Time
	Timestamp    time.Time
}

// PaymentScheduledEvent represents a scheduled payment event
type PaymentScheduledEvent struct {
	BaseEvent
	DeploymentID string
	Requester    common.Address
	Provider     common.Address
	Amount       *big.Int
	ScheduledAt  time.Time
}

// PaymentCancelledEvent represents a cancelled payment event
type PaymentCancelledEvent struct {
	DeploymentID string
	Requester    common.Address
	Provider     common.Address
	Amount       *big.Int
	Reason       string
	CancelledAt  time.Time
	Timestamp    time.Time
}

func (e *PaymentFailedEvent) GetType() MarketplaceEventType {
	return "PaymentFailed"
}

func (e *PaymentCreatedEvent) GetType() MarketplaceEventType {
	return "PaymentCreated"
}

func (e *PaymentReleasedEvent) GetType() MarketplaceEventType {
	return "PaymentReleased"
}

func (e *PaymentRefundedEvent) GetType() MarketplaceEventType {
	return "PaymentRefunded"
}

func (e *PaymentDisputedEvent) GetType() MarketplaceEventType {
	return "PaymentDisputed"
}

func (e *PaymentDisputeResolvedEvent) GetType() MarketplaceEventType {
	return "PaymentDisputeResolved"
}

func (e *PaymentScheduledEvent) GetType() MarketplaceEventType {
	return "PaymentScheduled"
}

func (e *PaymentCancelledEvent) GetType() MarketplaceEventType {
	return "PaymentCancelled"
}
