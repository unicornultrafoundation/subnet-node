package types

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type MarketplaceEventType string

const (
	MarketplaceEventTypeDeploymentRequested       MarketplaceEventType = "DeploymentRequested"
	MarketplaceEventTypeDeploymentRequestReceived MarketplaceEventType = "DeploymentRequestReceived"
	MarketplaceEventTypeBidSubmitted              MarketplaceEventType = "BidSubmitted"
	MarketplaceEventTypeProviderSelected          MarketplaceEventType = "ProviderSelected"
	MarketplaceEventTypeDeploymentApproved        MarketplaceEventType = "DeploymentApproved"
	MarketplaceEventTypeDeploymentCompleted       MarketplaceEventType = "DeploymentCompleted"
	MarketplaceEventTypeDeploymentTerminated      MarketplaceEventType = "DeploymentTerminated"
)

// MarketplaceEvent is the interface that all marketplace events must implement
type MarketplaceEvent interface {
	GetType() MarketplaceEventType
	GetTimestamp() time.Time
}

// BaseEvent contains common fields for all events
type BaseEvent struct {
	Timestamp time.Time
}

// GetTimestamp returns the event timestamp
func (e BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// Deployment represents a deployment in the system
type Deployment struct {
	ID        string
	Requester common.Address
	Provider  common.Address
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DeploymentRequestedEvent represents a deployment request event
type DeploymentRequestedEvent struct {
	BaseEvent
	DeploymentID string
	Requester    common.Address
	SDLHash      string
	MaxPrice     *big.Int
}

// GetType returns the event type
func (e *DeploymentRequestedEvent) GetType() MarketplaceEventType {
	return MarketplaceEventTypeDeploymentRequested
}

// DeploymentRequestReceivedEvent represents a deployment request received event
type DeploymentRequestReceivedEvent struct {
	BaseEvent
	DeploymentID string
	Requester    common.Address
	SDLHash      string
	MaxPrice     *big.Int
}

// GetType returns the event type
func (e *DeploymentRequestReceivedEvent) GetType() MarketplaceEventType {
	return MarketplaceEventTypeDeploymentRequestReceived
}

// BidSubmittedEvent represents a bid submitted event
type BidSubmittedEvent struct {
	BaseEvent
	DeploymentID string
	Provider     common.Address
	Amount       *big.Int
	Duration     time.Duration
}

// GetType returns the event type
func (e *BidSubmittedEvent) GetType() MarketplaceEventType {
	return MarketplaceEventTypeBidSubmitted
}

// ProviderSelectedEvent represents a provider selected event
type ProviderSelectedEvent struct {
	BaseEvent
	DeploymentID string
	Provider     common.Address
	Amount       *big.Int
}

// GetType returns the event type
func (e *ProviderSelectedEvent) GetType() MarketplaceEventType {
	return MarketplaceEventTypeProviderSelected
}

// DeploymentApprovedEvent represents a deployment approved event
type DeploymentApprovedEvent struct {
	BaseEvent
	DeploymentID string
	Provider     common.Address
	Requester    common.Address
}

// GetType returns the event type
func (e *DeploymentApprovedEvent) GetType() MarketplaceEventType {
	return MarketplaceEventTypeDeploymentApproved
}

// DeploymentCompletedEvent represents a deployment completed event
type DeploymentCompletedEvent struct {
	BaseEvent
	DeploymentID string
	Provider     common.Address
	Status       string
}

// GetType returns the event type
func (e *DeploymentCompletedEvent) GetType() MarketplaceEventType {
	return MarketplaceEventTypeDeploymentCompleted
}

// DeploymentTerminatedEvent represents a deployment terminated event
type DeploymentTerminatedEvent struct {
	BaseEvent
	DeploymentID string
	Provider     common.Address
	Requester    common.Address
}

// GetType returns the event type
func (e *DeploymentTerminatedEvent) GetType() MarketplaceEventType {
	return MarketplaceEventTypeDeploymentTerminated
}
