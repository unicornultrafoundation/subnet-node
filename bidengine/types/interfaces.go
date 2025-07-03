package types

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// BidMarketContract defines the interface for interacting with the bid market contract
type BidMarketContract interface {
	// Order management
	GetOrder(ctx context.Context, orderID *big.Int) (*Order, error)
	GetOrderCount(ctx context.Context) (*big.Int, error)
	OrderCount(ctx context.Context) (*big.Int, error)
	Orders(ctx context.Context, orderID *big.Int) (*Order, error)
	GetBids(ctx context.Context, orderID *big.Int) ([]Bid, error)
	IsBiddingOpen(ctx context.Context, orderID *big.Int) (bool, error)
	GetRemainingBidTime(ctx context.Context, orderID *big.Int) (*big.Int, error)

	// Bid submission
	SubmitBid(ctx context.Context, orderID *big.Int, pricePerSecond *big.Int, providerID *big.Int, machineID *big.Int) (*types.Transaction, error)
	GetBidIndexFromTransaction(ctx context.Context, tx *types.Transaction, orderID *big.Int) (*big.Int, error)
	CancelBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) (*types.Transaction, error)

	// Order lifecycle
	AcceptBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) (*types.Transaction, error)
	CancelOrder(ctx context.Context, orderID *big.Int) (*types.Transaction, error)
	CloseOrder(ctx context.Context, orderID *big.Int, reason string) (*types.Transaction, error)
	ExtendOrder(ctx context.Context, orderID *big.Int, amount *big.Int) (*types.Transaction, error)

	// Resource management
	GetUsedResource(ctx context.Context, providerID *big.Int, machineID *big.Int) (*ResourceUsage, error)
	ReleaseOrderResource(ctx context.Context, orderID *big.Int) (*types.Transaction, error)

	// Event monitoring
	WatchOrderCreated(ctx context.Context, sink chan<- *OrderEvent) error
	WatchOrderClosed(ctx context.Context, sink chan<- *OrderEvent) error
	WatchOrderExpired(ctx context.Context, sink chan<- *OrderEvent) error
	WatchBidSubmitted(ctx context.Context, sink chan<- *OrderEvent) error
	WatchBidAccepted(ctx context.Context, sink chan<- *OrderEvent) error
	WatchBidCancelled(ctx context.Context, sink chan<- *OrderEvent) error

	// OrderBids retrieves a single bid by orderID and bidIndex
	OrderBids(ctx context.Context, orderID *big.Int, bidIndex *big.Int) (*Bid, error)
}

// ProviderContract defines the interface for interacting with the provider contract
type ProviderContract interface {
	// Provider information
	GetProvider(ctx context.Context, providerID *big.Int) (*Provider, error)
	GetProviderOwner(ctx context.Context, providerID *big.Int) (common.Address, error)
	IsProviderOperatorOrOwner(ctx context.Context, providerID *big.Int, account common.Address) (bool, error)
	IsVerified(ctx context.Context, providerID *big.Int) (bool, error)

	// Machine management
	GetMachines(ctx context.Context, providerID *big.Int) ([]*Machine, error)
	GetActiveMachinesPaginated(ctx context.Context, providerID *big.Int, start *big.Int, limit *big.Int) ([]Machine, error)
	IsMachineActive(ctx context.Context, providerID *big.Int, machineID *big.Int) (bool, error)
	GetMachineResourcePrice(ctx context.Context, providerID *big.Int, machineID *big.Int) (*ResourceUsage, error)

	// Machine operations
	AddMachine(ctx context.Context, providerID *big.Int, machine Machine) (*types.Transaction, error)
	UpdateMachine(ctx context.Context, providerID *big.Int, machineID *big.Int, machine Machine) (*types.Transaction, error)
	RemoveMachine(ctx context.Context, providerID *big.Int, machineID *big.Int) (*types.Transaction, error)
	SetMachineResourcePrice(ctx context.Context, providerID *big.Int, machineID *big.Int, prices *ResourceUsage) (*types.Transaction, error)

	// Validation
	ValidateMachineRequirements(ctx context.Context, machineType *big.Int, providerID *big.Int, machineID *big.Int, requirements *ResourceUsage) (bool, error)
}

// PricingEngine defines the interface for calculating bid prices
type PricingEngine interface {
	// Calculate optimal bid price
	CalculateBidPrice(ctx context.Context, order *Order, machine *Machine, marketData *MarketData) (*big.Int, error)

	// Market analysis
	AnalyzeMarket(ctx context.Context, orders []*Order) (*MarketData, error)

	// Resource pricing
	CalculateResourcePrice(ctx context.Context, machine *Machine, usage *ResourceUsage) (*big.Int, error)

	// Strategy adjustment
	AdjustPriceForStrategy(ctx context.Context, basePrice *big.Int, strategy *BidStrategy) (*big.Int, error)
}

// ResourceManager defines the interface for resource management operations
type ResourceManager interface {
	// Machine management
	RegisterMachine(ctx context.Context, machine *Machine) error
	UnregisterMachine(ctx context.Context, machineID *big.Int) error
	GetAllMachines(ctx context.Context) []*Machine
	GetMachine(ctx context.Context, machineID *big.Int) (*Machine, error)

	// Resource allocation
	AllocateResources(ctx context.Context, orderID *big.Int, machine *Machine, usage *ResourceUsage) error
	DeallocateResources(ctx context.Context, orderID *big.Int) error
	GetCurrentUsage(ctx context.Context, machine *Machine) (*ResourceUsage, error)
	GetAvailableResources(ctx context.Context, machine *Machine) (*ResourceUsage, error)
	CanAllocateResources(ctx context.Context, machine *Machine, required *ResourceUsage) (bool, error)

	// Resource lifecycle
	StartResource(ctx context.Context, orderID *big.Int, machine *Machine) error
	StopResource(ctx context.Context, orderID *big.Int) error
}

// OrderMonitor defines the interface for monitoring orders
type OrderMonitor interface {
	// Order tracking
	TrackOrder(ctx context.Context, orderID *big.Int) error
	UntrackOrder(ctx context.Context, orderID *big.Int) error
	GetTrackedOrders(ctx context.Context) ([]*big.Int, error)

	// Event handling
	RegisterEventHandler(eventType OrderEventType, handler func(*OrderEvent))
}

// BidManager defines the interface for managing bids
type BidManager interface {
	// Bid tracking
	TrackBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) error
	UntrackBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) error
	GetTrackedBids(ctx context.Context) (map[*big.Int][]*big.Int, error)

	// Bid lifecycle
	SubmitBid(ctx context.Context, orderID *big.Int, pricePerSecond *big.Int, machineID *big.Int) (*BidResult, error)
	CancelBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) error
}

// Storage defines the interface for persistent storage operations
type Storage interface {
	// Order operations
	SaveOrder(ctx context.Context, order *Order) error
	GetOrder(ctx context.Context, orderID string) (*Order, error)
	ListOrders(ctx context.Context) ([]*Order, error)
	UpdateOrder(ctx context.Context, order *Order) error
	DeleteOrder(ctx context.Context, orderID string) error

	// Bid operations
	SaveBid(ctx context.Context, bid *Bid, orderID string, bidIndex int) error
	GetBids(ctx context.Context, orderID string) ([]*Bid, error)
	UpdateBid(ctx context.Context, bid *Bid, orderID string, bidIndex int) error

	// Machine operations
	SaveMachine(ctx context.Context, machine *Machine) error
	GetMachine(ctx context.Context, machineID string) (*Machine, error)
	ListMachines(ctx context.Context) ([]*Machine, error)

	// Market data operations
	SaveMarketData(ctx context.Context, marketData *MarketData) error
	GetMarketData(ctx context.Context) (*MarketData, error)

	// Resource allocation operations
	SaveResourceAllocation(ctx context.Context, allocation *ResourceAllocation) error
	ListResourceAllocations(ctx context.Context) ([]*ResourceAllocation, error)

	// Last order ID tracking
	SaveLastOrderID(ctx context.Context, orderID *big.Int) error
	GetLastOrderID(ctx context.Context) (*big.Int, error)
}

// Logger defines the interface for logging
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
}

// Metrics defines the interface for collecting metrics
type Metrics interface {
	// Bid metrics
	IncrementBidsSubmitted()
	IncrementBidsAccepted()
	IncrementBidsRejected()
	RecordBidLatency(duration float64)

	// Order metrics
	IncrementOrdersTracked()
	IncrementOrdersCompleted()
	RecordOrderDuration(duration float64)

	// Resource metrics
	RecordResourceUtilization(usage *ResourceUsage)
	RecordResourceAllocation(usage *ResourceUsage)

	// Financial metrics
	RecordRevenue(amount *big.Int)
	RecordProfit(amount *big.Int)
	RecordCost(amount *big.Int)
}

// AutoBidder defines the interface for automatic bidding on orders
type AutoBidder interface {
	// TryBidOnOrder attempts to bid on an order if conditions are met
	TryBidOnOrder(ctx context.Context, order *Order) error
}
