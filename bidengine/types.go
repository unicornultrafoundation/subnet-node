package bidengine

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// BidStatus represents the current status of a bid
type BidStatus int

const (
	BidStatusPending BidStatus = iota
	BidStatusActive
	BidStatusAccepted
	BidStatusRejected
	BidStatusExpired
	BidStatusFailed
)

// String returns the string representation of the bid status
func (s BidStatus) String() string {
	switch s {
	case BidStatusPending:
		return "Pending"
	case BidStatusActive:
		return "Active"
	case BidStatusAccepted:
		return "Accepted"
	case BidStatusRejected:
		return "Rejected"
	case BidStatusExpired:
		return "Expired"
	case BidStatusFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// Bid represents a bid placed by the bidengine
type Bid struct {
	ID           *big.Int
	OrderId      *big.Int
	ProviderId   *big.Int
	MachineId    *big.Int
	Amount       *big.Int
	Status       BidStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ExpirationAt time.Time
	AcceptedAt   *time.Time
	Requirements *BidRequirements
	TxHash       string
}

// BidRequirements represents the requirements for a bid
type BidRequirements struct {
	MachineType      *big.Int
	MinCPUCores      *big.Int
	MinMemoryMB      *big.Int
	MinDiskGB        *big.Int
	MinGPUCores      *big.Int
	MinUploadSpeed   *big.Int
	MinDownloadSpeed *big.Int
	Region           *big.Int // Optional region requirement
}

// OrderInfo represents an order available for bidding
type OrderInfo struct {
	OrderID      *big.Int
	RequesterID  common.Address
	Requirements *BidRequirements
	MaxPrice     *big.Int
	MinPrice     *big.Int
	CreatedAt    time.Time
	ExpirationAt time.Time
	Status       uint8
	BidCount     int
}

// BidConfig represents the configuration for placing bids
type BidConfig struct {
	MinBidPercent   int     // Min percentage of max price (e.g., 70 means 70% of max price)
	MaxBidPercent   int     // Max percentage of max price (e.g., 95 means 95% of max price)
	PriceFactor     float64 // Factor to adjust bid price based on requirements
	MinRequirements *BidRequirements
	MaxProfit       *big.Int // Maximum profit margin we want

	// Resource cost parameters in wei
	CpuCorePrice       *big.Int // Cost per CPU core
	MemoryGBPrice      *big.Int // Cost per GB of RAM
	DiskGBPrice        *big.Int // Cost per GB of disk
	GpuCorePrice       *big.Int // Cost per GPU core
	UploadSpeedPrice   *big.Int // Cost per upload speed unit
	DownloadSpeedPrice *big.Int // Cost per download speed unit

	// Machine type cost multipliers
	// Different machine types have different cost models
	MachineTypeMultipliers map[int64]*big.Int // Maps machine type ID to cost multiplier
}
