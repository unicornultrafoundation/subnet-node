package types

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

// BidStatus represents the status of a bid
type BidStatus int

const (
	BidStatusPending BidStatus = iota
	BidStatusAccepted
	BidStatusRejected
	BidStatusExpired
)

type Provider struct {
	ID                 *uint256.Int   `json:"id"`
	Name               string         `json:"name"`
	Description        string         `json:"description"`
	Owner              common.Address `json:"owner"`
	Operator           common.Address
	Registered         bool
	Reputation         *uint256.Int
	MachineCount       *uint256.Int
	CreatedAt          *uint256.Int
	UpdatedAt          *uint256.Int
	TotalStaked        *uint256.Int
	PendingWithdrawals *uint256.Int
	SlashedAmount      *uint256.Int
	TokenId            *uint256.Int
	Metadata           string
	IsSlashed          bool
	IsActive           bool
}

// BidRequirements defines the resource requirements for a bid
type BidRequirements struct {
	MinCPUCores      *uint256.Int `json:"minCpuCores"`
	MinMemoryMB      *uint256.Int `json:"minMemoryMb"`
	MinDiskGB        *uint256.Int `json:"minDiskGb"`
	MinGPUCores      *uint256.Int `json:"minGpuCores"`
	MinUploadSpeed   *uint256.Int `json:"minUploadSpeed"`
	MinDownloadSpeed *uint256.Int `json:"minDownloadSpeed"`
	Region           *uint256.Int `json:"region"`
	MachineType      *uint256.Int `json:"machineType"`
}

// Bid represents a bid placed on the BidMarket contract
type Bid struct {
	ID           *uint256.Int     `json:"id"`           // Using order ID as bid ID for tracking
	OrderId      *uint256.Int     `json:"orderId"`      // Order ID this bid is for
	ProviderId   *uint256.Int     `json:"providerId"`   // Provider ID making the bid
	MachineId    *uint256.Int     `json:"machineId"`    // Machine ID offered
	PricePerSec  *uint256.Int     `json:"amount"`       // Bid amount
	Status       BidStatus        `json:"status"`       // Current status of the bid
	CreatedAt    time.Time        `json:"createdAt"`    // When the bid was created
	UpdatedAt    time.Time        `json:"updatedAt"`    // When the bid was last updated
	ExpirationAt time.Time        `json:"expirationAt"` // When the bid expires
	Requirements *BidRequirements `json:"requirements"` // Resource requirements
	TxHash       string           `json:"txHash"`       // Transaction hash of the bid submission
	Renter       common.Address   `json:"renter"`       // Optional reference data from renter
}

// OrderInfo contains information about an order from the BidMarket contract
type OrderInfo struct {
	OrderID             *uint256.Int
	RequesterID         common.Address
	Requirements        *BidRequirements
	MaxPrice            *uint256.Int
	MinPrice            *uint256.Int
	CreatedAt           time.Time
	ExpirationAt        time.Time
	AcceptedMachineId   *uint256.Int
	AcceptedProviderId  *uint256.Int
	AcceptedPricePerSec *uint256.Int
	Status              uint8
}

// BidConfig contains configuration parameters for the bidding process
type BidConfig struct {
	MinBidPercent   int              // Minimum percentage of max price to bid
	MaxBidPercent   int              // Maximum percentage of max price to bid
	PriceFactor     float64          // Factor to adjust bid price
	MinRequirements *BidRequirements // Minimum requirements to consider bidding
	MaxProfit       *uint256.Int     // Maximum profit to make on a bid
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

// Machine represents a machine available for serving orders
type Machine struct {
	ID            *uint256.Int `json:"id"`            // Unique machine ID
	ProviderId    *uint256.Int `json:"providerId"`    // Provider ID that owns this machine
	CpuCores      *uint256.Int `json:"cpuCores"`      // Number of CPU cores
	MemoryMB      *uint256.Int `json:"memoryMB"`      // Memory in MB
	DiskGB        *uint256.Int `json:"diskGB"`        // Disk space in GB
	GpuCores      *uint256.Int `json:"gpuCores"`      // Number of GPU cores
	UploadSpeed   *uint256.Int `json:"uploadSpeed"`   // Upload speed in Mbps
	DownloadSpeed *uint256.Int `json:"downloadSpeed"` // Download speed in Mbps
	Active        bool         `json:"active"`        // Whether the machine is active
	Region        *uint256.Int `json:"region"`        // Geographic region ID
	MachineType   *uint256.Int `json:"machineType"`   // Type of machine (VM, bare metal, etc.)

	// Resource pricing per second
	CpuPricePerSec    *uint256.Int `json:"cpuPricePerSec"`    // Price per CPU core per second
	MemoryPricePerSec *uint256.Int `json:"memoryPricePerSec"` // Price per MB of memory per second
	DiskPricePerSec   *uint256.Int `json:"diskPricePerSec"`   // Price per GB of disk per second
	GpuPricePerSec    *uint256.Int `json:"gpuPricePerSec"`    // Price per GPU core per second

	// Additional fields for tracking
	Utilization        float64           `json:"utilization"`        // Current utilization percentage (0-100)
	AllocatedResources *MachineResources `json:"allocatedResources"` // Resources currently allocated to bids
	ActiveBids         []*uint256.Int    `json:"activeBids"`         // List of active bid IDs using this machine
}

// MachineResources represents the resources of a machine
type MachineResources struct {
	CpuCores *uint256.Int `json:"cpuCores"`
	MemoryMB *uint256.Int `json:"memoryMB"`
	DiskGB   *uint256.Int `json:"diskGB"`
	GpuCores *uint256.Int `json:"gpuCores"`
}

// CalculateRemainingResources calculates the remaining available resources on a machine
func (m *Machine) CalculateRemainingResources() *MachineResources {
	if !m.Active {
		return &MachineResources{
			CpuCores: uint256.NewInt(0),
			MemoryMB: uint256.NewInt(0),
			DiskGB:   uint256.NewInt(0),
			GpuCores: uint256.NewInt(0),
		}
	}

	// If no resources are allocated yet, initialize with zeros
	if m.AllocatedResources == nil {
		m.AllocatedResources = &MachineResources{
			CpuCores: uint256.NewInt(0),
			MemoryMB: uint256.NewInt(0),
			DiskGB:   uint256.NewInt(0),
			GpuCores: uint256.NewInt(0),
		}
	}

	// Calculate remaining resources
	remaining := &MachineResources{
		CpuCores: new(uint256.Int).Sub(m.CpuCores, m.AllocatedResources.CpuCores),
		MemoryMB: new(uint256.Int).Sub(m.MemoryMB, m.AllocatedResources.MemoryMB),
		DiskGB:   new(uint256.Int).Sub(m.DiskGB, m.AllocatedResources.DiskGB),
		GpuCores: new(uint256.Int).Sub(m.GpuCores, m.AllocatedResources.GpuCores),
	}

	// Ensure no negative values
	if remaining.CpuCores.Lt(uint256.NewInt(0)) {
		remaining.CpuCores = uint256.NewInt(0)
	}
	if remaining.MemoryMB.Lt(uint256.NewInt(0)) {
		remaining.MemoryMB = uint256.NewInt(0)
	}
	if remaining.DiskGB.Lt(uint256.NewInt(0)) {
		remaining.DiskGB = uint256.NewInt(0)
	}
	if remaining.GpuCores.Lt(uint256.NewInt(0)) {
		remaining.GpuCores = uint256.NewInt(0)
	}

	return remaining
}

// CanSatisfyRequirements checks if the machine can satisfy the given resource requirements
func (m *Machine) CanSatisfyRequirements(requirements *BidRequirements) bool {
	if !m.Active {
		return false
	}

	// Get remaining resources
	remaining := m.CalculateRemainingResources()

	// Check if all requirements are satisfied
	if remaining.CpuCores.Lt(requirements.MinCPUCores) ||
		remaining.MemoryMB.Lt(requirements.MinMemoryMB) ||
		remaining.DiskGB.Lt(requirements.MinDiskGB) ||
		remaining.GpuCores.Lt(requirements.MinGPUCores) ||
		!m.Region.Eq(requirements.Region) ||
		!m.MachineType.Eq(requirements.MachineType) {
		return false
	}

	return true
}

// CalculateUtilization calculates the current resource utilization of the machine
func (m *Machine) CalculateUtilization() float64 {
	if !m.Active || m.AllocatedResources == nil {
		return 0.0
	}

	// Calculate utilization as average percentage of key resources
	cpuUtil := float64(m.AllocatedResources.CpuCores.Uint64()) / float64(m.CpuCores.Uint64())
	memUtil := float64(m.AllocatedResources.MemoryMB.Uint64()) / float64(m.MemoryMB.Uint64())
	diskUtil := float64(m.AllocatedResources.DiskGB.Uint64()) / float64(m.DiskGB.Uint64())

	// Average the utilization metrics (weighted more toward CPU and memory)
	utilization := (cpuUtil*0.4 + memUtil*0.4 + diskUtil*0.2) * 100.0
	if utilization > 100.0 {
		utilization = 100.0
	}

	m.Utilization = utilization
	return utilization
}

// CalculateHourlyCost calculates the cost per hour for running a workload with the specified requirements
func (m *Machine) CalculateHourlyCost(requirements *BidRequirements) *uint256.Int {
	// Calculate cost for each resource type per second
	cpuCost := new(uint256.Int).Mul(m.CpuPricePerSec, requirements.MinCPUCores)
	memoryCost := new(uint256.Int).Mul(m.MemoryPricePerSec, requirements.MinMemoryMB)
	diskCost := new(uint256.Int).Mul(m.DiskPricePerSec, requirements.MinDiskGB)
	gpuCost := new(uint256.Int).Mul(m.GpuPricePerSec, requirements.MinGPUCores)

	// Total cost per second
	costPerSecond := new(uint256.Int).Add(cpuCost, memoryCost)
	costPerSecond = new(uint256.Int).Add(costPerSecond, diskCost)
	costPerSecond = new(uint256.Int).Add(costPerSecond, gpuCost)

	// Convert to cost per hour (multiply by 3600 seconds)
	secondsPerHour := uint256.NewInt(3600)
	costPerHour := new(uint256.Int).Mul(costPerSecond, secondsPerHour)

	return costPerHour
}

// CalculateDailyCost calculates the cost per day for running a workload with the specified requirements
func (m *Machine) CalculateDailyCost(requirements *BidRequirements) *uint256.Int {
	// Get hourly cost
	hourlyCost := m.CalculateHourlyCost(requirements)

	// Multiply by 24 hours
	hoursPerDay := uint256.NewInt(24)
	dailyCost := new(uint256.Int).Mul(hourlyCost, hoursPerDay)

	return dailyCost
}

// CalculateMonthlyCost calculates the approximate cost per month (30 days) for running a workload
func (m *Machine) CalculateMonthlyCost(requirements *BidRequirements) *uint256.Int {
	// Get daily cost
	dailyCost := m.CalculateDailyCost(requirements)

	// Multiply by 30 days (approximate month)
	daysPerMonth := uint256.NewInt(30)
	monthlyCost := new(uint256.Int).Mul(dailyCost, daysPerMonth)

	return monthlyCost
}

// MachineManager defines the interface for machine management operations
type MachineManager interface {
	// GetMachine retrieves a machine by provider ID and machine ID
	GetMachine(ctx context.Context, providerId, machineId *uint256.Int) (*Machine, error)

	// ListMachines retrieves all machines for a provider
	ListMachines(ctx context.Context, providerId *uint256.Int) ([]*Machine, error)

	// UpdateMachine updates machine information
	UpdateMachine(ctx context.Context, machine *Machine) error

	// AllocateResources allocates resources on a machine for a bid
	AllocateResources(ctx context.Context, providerId, machineId *uint256.Int, bidId *uint256.Int, resources *BidRequirements) error

	// ReleaseResources releases resources allocated to a bid
	ReleaseResources(ctx context.Context, providerId, machineId *uint256.Int, bidId *uint256.Int) error
}
