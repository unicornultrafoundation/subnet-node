package types

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// MarketplaceEventType represents the type of a marketplace event
type MarketplaceEventType string

const (
	// Event types
	MarketplaceEventTypeDeploymentRequested  MarketplaceEventType = "DeploymentRequested"
	MarketplaceEventTypeBidSubmitted         MarketplaceEventType = "BidSubmitted"
	MarketplaceEventTypeProviderSelected     MarketplaceEventType = "ProviderSelected"
	MarketplaceEventTypeDeploymentCompleted  MarketplaceEventType = "DeploymentCompleted"
	MarketplaceEventTypeDeploymentTerminated MarketplaceEventType = "DeploymentTerminated"
	MarketplaceEventTypePaymentCreated       MarketplaceEventType = "PaymentCreated"
	MarketplaceEventTypePaymentReleased      MarketplaceEventType = "PaymentReleased"
	MarketplaceEventTypePaymentRefunded      MarketplaceEventType = "PaymentRefunded"
	MarketplaceEventTypeDisputeInitiated     MarketplaceEventType = "DisputeInitiated"
)

// MarketplaceEvent is the interface that all marketplace events must implement
type MarketplaceEvent interface {
	GetType() MarketplaceEventType
	GetTimestamp() time.Time
	GetBlockNum() uint64
	GetTxHash() string
}

// BaseEvent contains common fields for all events
type BaseEvent struct {
	Timestamp time.Time
	BlockNum  uint64
	TxHash    string
}

// GetTimestamp returns the event timestamp
func (e BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// GetBlockNum returns the event block number
func (e BaseEvent) GetBlockNum() uint64 {
	return e.BlockNum
}

// GetTxHash returns the event transaction hash
func (e BaseEvent) GetTxHash() string {
	return e.TxHash
}

// Deployment represents a deployment in the system
type Deployment struct {
	ID               string
	Requester        common.Address
	Provider         common.Address
	Status           DeploymentStatus
	SDLHash          string
	SelectedProvider common.Address
	SelectedBid      *BidSubmittedEvent
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// DeploymentRequestedEvent represents a deployment request event
type DeploymentRequestedEvent struct {
	BaseEvent
	DeploymentID string
	Requester    common.Address
	SDLHash      string
	MinBid       *big.Int
	MaxBid       *big.Int
	Duration     time.Duration
}

// GetType returns the event type
func (e *DeploymentRequestedEvent) GetType() MarketplaceEventType {
	return MarketplaceEventTypeDeploymentRequested
}
func (e *DeploymentRequestedEvent) GetTimestamp() time.Time { return e.Timestamp }
func (e *DeploymentRequestedEvent) GetBlockNum() uint64     { return e.BlockNum }
func (e *DeploymentRequestedEvent) GetTxHash() string       { return e.TxHash }

// BidSubmittedEvent represents a bid submitted event
type BidSubmittedEvent struct {
	BaseEvent
	DeploymentID string
	Provider     common.Address
	Amount       *big.Int
	Duration     time.Duration
	Status       BidStatus
	SDLHash      string
	Requester    common.Address
}

// GetType returns the event type
func (e *BidSubmittedEvent) GetType() MarketplaceEventType {
	return MarketplaceEventTypeBidSubmitted
}
func (e *BidSubmittedEvent) GetTimestamp() time.Time { return e.Timestamp }
func (e *BidSubmittedEvent) GetBlockNum() uint64     { return e.BlockNum }
func (e *BidSubmittedEvent) GetTxHash() string       { return e.TxHash }

// ProviderSelectedEvent represents a provider selected event
type ProviderSelectedEvent struct {
	BaseEvent
	DeploymentID string
	Provider     common.Address
	Amount       *big.Int
	Duration     time.Duration
}

// GetType returns the event type
func (e *ProviderSelectedEvent) GetType() MarketplaceEventType {
	return MarketplaceEventTypeProviderSelected
}
func (e *ProviderSelectedEvent) GetTimestamp() time.Time { return e.Timestamp }
func (e *ProviderSelectedEvent) GetBlockNum() uint64     { return e.BlockNum }
func (e *ProviderSelectedEvent) GetTxHash() string       { return e.TxHash }

// DeploymentCompletedEvent represents a deployment completed event
type DeploymentCompletedEvent struct {
	BaseEvent
	DeploymentID string
	Provider     common.Address
	Status       DeploymentStatus
}

// GetType returns the event type
func (e *DeploymentCompletedEvent) GetType() MarketplaceEventType {
	return MarketplaceEventTypeDeploymentCompleted
}
func (e *DeploymentCompletedEvent) GetTimestamp() time.Time { return e.Timestamp }
func (e *DeploymentCompletedEvent) GetBlockNum() uint64     { return e.BlockNum }
func (e *DeploymentCompletedEvent) GetTxHash() string       { return e.TxHash }

// DeploymentTerminatedEvent represents a deployment terminated event
type DeploymentTerminatedEvent struct {
	BaseEvent
	DeploymentID string
	Provider     common.Address
	Requester    common.Address
	Reason       string
}

// GetType returns the event type
func (e *DeploymentTerminatedEvent) GetType() MarketplaceEventType {
	return MarketplaceEventTypeDeploymentTerminated
}
func (e *DeploymentTerminatedEvent) GetTimestamp() time.Time { return e.Timestamp }
func (e *DeploymentTerminatedEvent) GetBlockNum() uint64     { return e.BlockNum }
func (e *DeploymentTerminatedEvent) GetTxHash() string       { return e.TxHash }

// DisputeInitiatedEvent represents a dispute initiated event
type DisputeInitiatedEvent struct {
	BaseEvent
	DeploymentID string
	Provider     common.Address
	Requester    common.Address
	Reason       string
}

// GetType returns the event type
func (e *DisputeInitiatedEvent) GetType() MarketplaceEventType {
	return MarketplaceEventTypeDisputeInitiated
}
func (e *DisputeInitiatedEvent) GetTimestamp() time.Time { return e.Timestamp }
func (e *DisputeInitiatedEvent) GetBlockNum() uint64     { return e.BlockNum }
func (e *DisputeInitiatedEvent) GetTxHash() string       { return e.TxHash }

// EventHandler represents a function that handles marketplace events
type EventHandler func(ctx context.Context, event MarketplaceEvent) error

// EventBus represents an event bus for marketplace events
type EventBus interface {
	// Subscribe subscribes to marketplace events
	Subscribe() <-chan MarketplaceEvent
	// Publish publishes a marketplace event
	Publish(ctx context.Context, event MarketplaceEvent) error
	// Close closes the event bus
	Close() error
}

const (
	// Event types
	EventTypeDeploymentRequested  MarketplaceEventType = "DeploymentRequested"
	EventTypeBidSubmitted         MarketplaceEventType = "BidSubmitted"
	EventTypeProviderSelected     MarketplaceEventType = "ProviderSelected"
	EventTypeDeploymentCompleted  MarketplaceEventType = "DeploymentCompleted"
	EventTypeDeploymentTerminated MarketplaceEventType = "DeploymentTerminated"
	EventTypePaymentCreated       MarketplaceEventType = "PaymentCreated"
	EventTypePaymentReleased      MarketplaceEventType = "PaymentReleased"
	EventTypePaymentRefunded      MarketplaceEventType = "PaymentRefunded"
	EventTypeDisputeInitiated     MarketplaceEventType = "DisputeInitiated"
)
