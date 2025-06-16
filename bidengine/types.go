package bidengine

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// BidStatus represents the status of a bid
type BidStatus int

const (
	BidStatusPending BidStatus = iota
	BidStatusAccepted
	BidStatusRejected
	BidStatusExpired
)

// BidRequirements defines the resource requirements for a bid
type BidRequirements struct {
	MinCPUCores      *big.Int `json:"minCpuCores"`
	MinMemoryMB      *big.Int `json:"minMemoryMb"`
	MinDiskGB        *big.Int `json:"minDiskGb"`
	MinGPUCores      *big.Int `json:"minGpuCores"`
	MinUploadSpeed   *big.Int `json:"minUploadSpeed"`
	MinDownloadSpeed *big.Int `json:"minDownloadSpeed"`
	Region           *big.Int `json:"region"`
	MachineType      *big.Int `json:"machineType"`
}

// Bid represents a bid placed on the BidMarket contract
type Bid struct {
	ID           *big.Int         `json:"id"`           // Using order ID as bid ID for tracking
	OrderId      *big.Int         `json:"orderId"`      // Order ID this bid is for
	ProviderId   *big.Int         `json:"providerId"`   // Provider ID making the bid
	MachineId    *big.Int         `json:"machineId"`    // Machine ID offered
	Amount       *big.Int         `json:"amount"`       // Bid amount
	Status       BidStatus        `json:"status"`       // Current status of the bid
	CreatedAt    time.Time        `json:"createdAt"`    // When the bid was created
	UpdatedAt    time.Time        `json:"updatedAt"`    // When the bid was last updated
	ExpirationAt time.Time        `json:"expirationAt"` // When the bid expires
	Requirements *BidRequirements `json:"requirements"` // Resource requirements
	TxHash       string           `json:"txHash"`       // Transaction hash of the bid submission
}

// OrderInfo contains information about an order from the BidMarket contract
type OrderInfo struct {
	OrderID      *big.Int
	RequesterID  common.Address
	Requirements *BidRequirements
	MaxPrice     *big.Int
	MinPrice     *big.Int
	CreatedAt    time.Time
	ExpirationAt time.Time
	Status       uint8
}

// BidConfig contains configuration parameters for the bidding process
type BidConfig struct {
	MinBidPercent          int                // Minimum percentage of max price to bid
	MaxBidPercent          int                // Maximum percentage of max price to bid
	PriceFactor            float64            // Factor to adjust bid price
	MinRequirements        *BidRequirements   // Minimum requirements to consider bidding
	MaxProfit              *big.Int           // Maximum profit to make on a bid
	CpuCorePrice           *big.Int           // Cost per CPU core
	MemoryGBPrice          *big.Int           // Cost per GB of memory
	DiskGBPrice            *big.Int           // Cost per GB of disk
	GpuCorePrice           *big.Int           // Cost per GPU core
	UploadSpeedPrice       *big.Int           // Cost per unit of upload speed
	DownloadSpeedPrice     *big.Int           // Cost per unit of download speed
	MachineTypeMultipliers map[int64]*big.Int // Cost multipliers by machine type
}

// BidDatastore defines the interface for bid storage operations
type BidDatastore interface {
	// SaveBid stores or updates a bid in the datastore
	SaveBid(ctx context.Context, bid *Bid) error

	// GetBidByID retrieves a bid by its ID
	GetBidByID(ctx context.Context, bidID string) (*Bid, error)

	// ListActiveBids retrieves all active bids
	ListActiveBids(ctx context.Context) ([]*Bid, error)

	// ListBidsByStatus retrieves bids by their status
	ListBidsByStatus(ctx context.Context, status BidStatus) ([]*Bid, error)

	// DeleteBid removes a bid from the datastore
	DeleteBid(ctx context.Context, bidID string) error
}
