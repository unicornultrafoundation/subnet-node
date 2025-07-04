package types

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// OrderStatus represents the status of an order
type OrderStatus uint8

const (
	OrderStatusOpen OrderStatus = iota
	OrderStatusAccepted
	OrderStatusClosed
	OrderStatusExpired
	OrderStatusCancelled
)

// BidStatus represents the status of a bid
type BidStatus uint8

const (
	BidStatusActive BidStatus = iota
	BidStatusAccepted
	BidStatusCancelled
	BidStatusExpired
)

// Order represents a resource order from the smart contract
type Order struct {
	ID                        *big.Int
	MachineType               *big.Int
	Owner                     common.Address
	Status                    OrderStatus
	CreatedAt                 *big.Int
	Duration                  *big.Int
	MinBidPrice               *big.Int
	MaxBidPrice               *big.Int
	AcceptedBidPricePerSecond *big.Int
	ParentOrderId             *big.Int
	PaymentToken              common.Address
	CpuCores                  *big.Int
	GpuCores                  *big.Int
	GpuMemory                 *big.Int
	MemoryMB                  *big.Int
	DiskGB                    *big.Int
	UploadMbps                *big.Int
	DownloadMbps              *big.Int
	Region                    *big.Int
	Specs                     string
	AcceptedProviderId        *big.Int
	AcceptedMachineId         *big.Int
	StartAt                   *big.Int
	ExpiredAt                 *big.Int
	LastPaidAt                *big.Int
}

// Bid represents a bid submitted by a provider
type Bid struct {
	Id             *big.Int
	Provider       common.Address
	PricePerSecond *big.Int
	Status         BidStatus
	CreatedAt      *big.Int
	ProviderId     *big.Int
	MachineId      *big.Int
}

// Machine represents a provider's machine resource
type Machine struct {
	ID                   *big.Int
	Active               bool
	MachineType          *big.Int
	Region               *big.Int
	CpuCores             *big.Int
	GpuCores             *big.Int
	GpuMemory            *big.Int
	MemoryMB             *big.Int
	DiskGB               *big.Int
	UploadSpeed          *big.Int
	DownloadSpeed        *big.Int
	CreatedAt            *big.Int
	UpdatedAt            *big.Int
	StakeAmount          *big.Int
	RemovedAt            *big.Int
	UnlockTime           *big.Int
	WithdrawalProcessed  bool
	Metadata             string
	CpuPricePerSecond    *big.Int
	GpuPricePerSecond    *big.Int
	MemoryPricePerSecond *big.Int
	DiskPricePerSecond   *big.Int
}

// Provider represents a provider's information
type Provider struct {
	Operator           common.Address
	Registered         bool
	Reputation         *big.Int
	MachineCount       *big.Int
	CreatedAt          *big.Int
	UpdatedAt          *big.Int
	TotalStaked        *big.Int
	PendingWithdrawals *big.Int
	SlashedAmount      *big.Int
	TokenId            *big.Int
	Metadata           string
	IsSlashed          bool
	IsActive           bool
	Verified           bool
}

// BidStrategy represents the bidding strategy configuration
type BidStrategy struct {
	MinProfitMargin   float64 // Minimum profit margin percentage
	MaxProfitMargin   float64 // Maximum profit margin percentage
	CompetitiveFactor float64 // Factor to make bids more competitive
	MarketAdjustment  float64 // Market price adjustment factor
	ResourceWeight    ResourceWeight
}

// ResourceWeight represents the weight of different resources in pricing
type ResourceWeight struct {
	CPU     float64
	GPU     float64
	Memory  float64
	Disk    float64
	Network float64
}

// BidEngineConfig represents the configuration for the bid engine
type BidEngineConfig struct {
	// Contract addresses
	BidMarketAddress common.Address
	ProviderAddress  common.Address

	// Provider information
	ProviderID     *big.Int
	ProviderWallet common.Address

	// Bidding configuration
	BidStrategy       BidStrategy
	MaxConcurrentBids int
	BidTimeout        time.Duration

	// Monitoring configuration
	OrderSyncInterval time.Duration
	BidCheckInterval  time.Duration

	// Logging configuration
	LogLevel string
	LogFile  string
}

// BidResult represents the result of a bid submission
type BidResult struct {
	OrderID   *big.Int
	BidIndex  *big.Int
	Success   bool
	Error     error
	TxHash    common.Hash
	Timestamp time.Time
}

// OrderEventType represents the type of order event
type OrderEventType string

const (
	OrderEventNew      OrderEventType = "new"
	OrderEventClosed   OrderEventType = "closed"
	OrderEventExpired  OrderEventType = "expired"
	OrderEventUpdated  OrderEventType = "updated"
	OrderEventAccepted OrderEventType = "accepted"
)

// OrderEvent represents an order lifecycle event
type OrderEvent struct {
	Type      OrderEventType
	OrderID   *big.Int
	Order     *Order
	Timestamp time.Time
	Data      map[string]interface{}
}

// ResourceUsage represents current resource usage
type ResourceUsage struct {
	CPUUsed     *big.Int
	GPUUsed     *big.Int
	MemoryUsed  *big.Int
	DiskUsed    *big.Int
	NetworkUsed *big.Int
}

// MarketData represents market information for pricing
type MarketData struct {
	AveragePricePerSecond *big.Int
	MinPricePerSecond     *big.Int
	MaxPricePerSecond     *big.Int
	TotalOrders           *big.Int
	ActiveOrders          *big.Int
	LastUpdated           time.Time
}

// ResourceAllocation represents stored resource allocation data
type ResourceAllocation struct {
	OrderID   *big.Int       `json:"order_id"`
	Usage     *ResourceUsage `json:"usage"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// Error definitions
var (
	ErrMachineNotFound           = fmt.Errorf("machine not found")
	ErrResourcesAlreadyAllocated = fmt.Errorf("resources already allocated")
	ErrInsufficientResources     = fmt.Errorf("insufficient resources")
	ErrResourcesNotAllocated     = fmt.Errorf("resources not allocated")
	ErrOrderNotTracked           = fmt.Errorf("order not tracked")
	ErrOrderExpired              = fmt.Errorf("order expired")
	ErrBidNotTracked             = fmt.Errorf("bid not tracked")
)

// IsOrderReadyToClose kiểm tra order đã hết hạn + 1 ngày chưa
func IsOrderReadyToClose(order *Order, now int64) bool {
	if order == nil || order.ExpiredAt.Int64() == 0 {
		return false
	}
	const oneDay = int64(86402)
	return now > order.ExpiredAt.Int64()+oneDay
}

func (order *Order) IsOrderWithinBiddingTime() bool {
	now := time.Now().Unix()
	const biddingTimeLimit = int64(300) // 5 minutes in seconds
	timeSinceCreation := now - order.CreatedAt.Int64()
	return timeSinceCreation <= biddingTimeLimit
}
