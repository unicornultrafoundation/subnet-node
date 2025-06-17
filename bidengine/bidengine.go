package bidengine

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/bidengine/contracts"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/repo"
)

// BidEngine is responsible for managing bids on the BidMarket contract
type BidEngine struct {
	cfg           *config.C
	client        *ethclient.Client
	auth          *bind.TransactOpts
	log           *logrus.Logger
	provider      *contracts.Provider
	bidMarket     *contracts.BidMarket
	providerAddr  common.Address
	bidMarketAddr common.Address
	activeBids    map[string]*Bid
	bidsMutex     sync.RWMutex
	bidConfig     BidConfig
	providerId    *big.Int     // Our provider ID
	store         BidDatastore // Added datastore interface

	shutdown chan struct{}
	wg       sync.WaitGroup
}

// NewBidEngine creates a new instance of the BidEngine
func NewBidEngine(cfg *config.C, log *logrus.Logger, acc account.Service, ds repo.Datastore) (*BidEngine, error) {
	client := acc.GetClient()

	// Get contract addresses from config
	providerAddrStr := cfg.GetString("contracts.provider", "")
	if providerAddrStr == "" {
		return nil, fmt.Errorf("contracts.provider address is not set in config")
	}
	providerAddr := common.HexToAddress(providerAddrStr)

	bidMarketAddrStr := cfg.GetString("contracts.bid_market", "")
	if bidMarketAddrStr == "" {
		return nil, fmt.Errorf("contracts.bid_market address is not set in config")
	}
	bidMarketAddr := common.HexToAddress(bidMarketAddrStr)

	// Initialize contracts
	provider, err := contracts.NewProvider(providerAddr, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider contract: %v", err)
	}

	bidMarket, err := contracts.NewBidMarket(bidMarketAddr, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create bid market contract: %v", err)
	}

	auth, err := acc.NewKeyedTransactor()
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %v", err)
	}

	store, err := NewStore(ds, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create bid store: %v", err)
	}

	return &BidEngine{
		cfg:           cfg,
		client:        client,
		log:           log,
		provider:      provider,
		bidMarket:     bidMarket,
		providerAddr:  providerAddr,
		bidMarketAddr: bidMarketAddr,
		activeBids:    make(map[string]*Bid),
		shutdown:      make(chan struct{}),
		auth:          auth,
		store:         store,
	}, nil
}

// Start initializes and starts the bid engine processes
func (b *BidEngine) Start(ctx context.Context) error {
	b.log.Info("Starting BidEngine...")

	// Load bid configuration
	b.loadBidConfig()

	// Load active bids from datastore
	bids, err := b.store.ListActiveBids(ctx)
	if err != nil {
		b.log.WithError(err).Warn("Failed to load active bids from datastore")
	} else {
		b.log.WithField("count", len(bids)).Info("Loaded active bids from datastore")
		b.bidsMutex.Lock()
		for _, bid := range bids {
			b.activeBids[bid.ID.String()] = bid
		}
		b.bidsMutex.Unlock()
	}

	// Start monitoring for active bids
	b.wg.Add(1)
	go b.monitorActiveBids(ctx)

	// Start watching for new orders directly from the blockchain
	if err := b.watchNewOrders(ctx); err != nil {
		return fmt.Errorf("failed to start watching orders: %v", err)
	}

	return nil
}

// Stop gracefully stops the bid engine
func (b *BidEngine) Stop(ctx context.Context) error {
	b.log.Info("Stopping BidEngine...")
	close(b.shutdown)

	// Wait for all goroutines to finish with a timeout
	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		b.log.Info("BidEngine stopped gracefully")
	case <-ctx.Done():
		b.log.Warn("BidEngine stop timed out, some goroutines may still be running")
	}

	return nil
}

// GetActiveBids returns all active bids
func (b *BidEngine) GetActiveBids() []*Bid {
	b.bidsMutex.RLock()
	defer b.bidsMutex.RUnlock()

	bids := make([]*Bid, 0, len(b.activeBids))
	for _, bid := range b.activeBids {
		bids = append(bids, bid)
	}

	return bids
}

// loadBidConfig loads the bid configuration from the config file
func (b *BidEngine) loadBidConfig() {
	b.bidConfig = BidConfig{
		MinBidPercent: b.cfg.GetInt("bidengine.min_bid_percent", 70),
		MaxBidPercent: b.cfg.GetInt("bidengine.max_bid_percent", 95),
		PriceFactor:   float64(b.cfg.GetInt("bidengine.price_factor", 1)),
		MinRequirements: &BidRequirements{
			MinCPUCores:      big.NewInt(int64(b.cfg.GetInt("bidengine.min_requirements.cpu_cores", 1))),
			MinMemoryMB:      big.NewInt(int64(b.cfg.GetInt("bidengine.min_requirements.memory_mb", 1024))),
			MinDiskGB:        big.NewInt(int64(b.cfg.GetInt("bidengine.min_requirements.disk_gb", 10))),
			MinGPUCores:      big.NewInt(int64(b.cfg.GetInt("bidengine.min_requirements.gpu_cores", 0))),
			MinUploadSpeed:   big.NewInt(int64(b.cfg.GetInt("bidengine.min_requirements.upload_speed", 0))),
			MinDownloadSpeed: big.NewInt(int64(b.cfg.GetInt("bidengine.min_requirements.download_speed", 0))),
		},
		MaxProfit: new(big.Int).SetInt64(int64(b.cfg.GetInt("bidengine.max_profit", 1000000000000000000))), // Default 1 ETH

		// Resource cost parameters with defaults in wei
		CpuCorePrice:       new(big.Int).SetInt64(int64(b.cfg.GetInt("bidengine.resource_cost.cpu_core", 1000000000000000))),     // 0.001 ETH default
		MemoryGBPrice:      new(big.Int).SetInt64(int64(b.cfg.GetInt("bidengine.resource_cost.memory_gb", 2000000000000000))),    // 0.002 ETH default
		DiskGBPrice:        new(big.Int).SetInt64(int64(b.cfg.GetInt("bidengine.resource_cost.disk_gb", 100000000000000))),       // 0.0001 ETH default
		GpuCorePrice:       new(big.Int).SetInt64(int64(b.cfg.GetInt("bidengine.resource_cost.gpu_core", 5000000000000000))),     // 0.005 ETH default
		UploadSpeedPrice:   new(big.Int).SetInt64(int64(b.cfg.GetInt("bidengine.resource_cost.upload_speed", 50000000000000))),   // 0.00005 ETH default
		DownloadSpeedPrice: new(big.Int).SetInt64(int64(b.cfg.GetInt("bidengine.resource_cost.download_speed", 50000000000000))), // 0.00005 ETH default

		// Initialize machine type cost multipliers with defaults
		MachineTypeMultipliers: make(map[int64]*big.Int),
	}

	// Load machine type multipliers from config
	// Default: shared=1.0, VM=1.5, bare metal=2.5
	b.bidConfig.MachineTypeMultipliers[1] = big.NewInt(int64(b.cfg.GetInt("bidengine.machine_type_multipliers.shared", 10)))
	b.bidConfig.MachineTypeMultipliers[2] = big.NewInt(int64(b.cfg.GetInt("bidengine.machine_type_multipliers.vm", 15)))
	b.bidConfig.MachineTypeMultipliers[3] = big.NewInt(int64(b.cfg.GetInt("bidengine.machine_type_multipliers.bare_metal", 25)))
}

// MachineResources represents the available resources of a machine
type MachineResources struct {
	CpuCores      *big.Int
	MemoryMB      *big.Int
	DiskGB        *big.Int
	GpuCores      *big.Int
	UploadSpeed   *big.Int
	DownloadSpeed *big.Int
	Active        bool
	Region        *big.Int
	MachineType   *big.Int
}

// GetMachineRemainingResources retrieves the remaining resources of a machine after accounting for active bids
func (b *BidEngine) GetMachineRemainingResources(ctx context.Context, providerId, machineId *big.Int) (*MachineResources, error) {
	// Get machine details from the blockchain
	machine, err := b.provider.ProviderMachines(&bind.CallOpts{Context: ctx}, providerId, machineId)
	if err != nil {
		return nil, fmt.Errorf("failed to get machine details: %v", err)
	}

	// Initialize with total resources
	resources := &MachineResources{
		CpuCores:      new(big.Int).Set(machine.CpuCores),
		MemoryMB:      new(big.Int).Set(machine.MemoryMB),
		DiskGB:        new(big.Int).Set(machine.DiskGB),
		GpuCores:      new(big.Int).Set(machine.GpuCores),
		UploadSpeed:   new(big.Int).Set(machine.UploadSpeed),
		DownloadSpeed: new(big.Int).Set(machine.DownloadSpeed),
		Active:        machine.Active,
		Region:        new(big.Int).Set(machine.Region),
		MachineType:   new(big.Int).Set(machine.MachineType),
	}

	// Return early if machine is not active
	if !machine.Active {
		return resources, nil
	}

	// Calculate resources used by active bids
	usedCPU := big.NewInt(0)
	usedMemory := big.NewInt(0)
	usedDisk := big.NewInt(0)
	usedGPU := big.NewInt(0)
	usedUpload := big.NewInt(0)
	usedDownload := big.NewInt(0)

	// Lock for reading the active bids map
	b.bidsMutex.RLock()
	for _, bid := range b.activeBids {
		// Only count bids that are pending or accepted for this specific machine
		if (bid.Status == BidStatusPending || bid.Status == BidStatusAccepted) &&
			bid.MachineId.Cmp(machineId) == 0 && bid.ProviderId.Cmp(providerId) == 0 {

			usedCPU = new(big.Int).Add(usedCPU, bid.Requirements.MinCPUCores)
			usedMemory = new(big.Int).Add(usedMemory, bid.Requirements.MinMemoryMB)
			usedDisk = new(big.Int).Add(usedDisk, bid.Requirements.MinDiskGB)
			usedGPU = new(big.Int).Add(usedGPU, bid.Requirements.MinGPUCores)
			usedUpload = new(big.Int).Add(usedUpload, bid.Requirements.MinUploadSpeed)
			usedDownload = new(big.Int).Add(usedDownload, bid.Requirements.MinDownloadSpeed)
		}
	}
	b.bidsMutex.RUnlock()

	// Calculate remaining resources by subtracting used resources
	resources.CpuCores = new(big.Int).Sub(resources.CpuCores, usedCPU)
	resources.MemoryMB = new(big.Int).Sub(resources.MemoryMB, usedMemory)
	resources.DiskGB = new(big.Int).Sub(resources.DiskGB, usedDisk)
	resources.GpuCores = new(big.Int).Sub(resources.GpuCores, usedGPU)
	resources.UploadSpeed = new(big.Int).Sub(resources.UploadSpeed, usedUpload)
	resources.DownloadSpeed = new(big.Int).Sub(resources.DownloadSpeed, usedDownload)

	// Ensure we don't return negative values (could happen if there's a resource tracking issue)
	if resources.CpuCores.Sign() < 0 {
		resources.CpuCores = big.NewInt(0)
	}
	if resources.MemoryMB.Sign() < 0 {
		resources.MemoryMB = big.NewInt(0)
	}
	if resources.DiskGB.Sign() < 0 {
		resources.DiskGB = big.NewInt(0)
	}
	if resources.GpuCores.Sign() < 0 {
		resources.GpuCores = big.NewInt(0)
	}
	if resources.UploadSpeed.Sign() < 0 {
		resources.UploadSpeed = big.NewInt(0)
	}
	if resources.DownloadSpeed.Sign() < 0 {
		resources.DownloadSpeed = big.NewInt(0)
	}

	return resources, nil
}

// GetMachineBids retrieves all bids associated with a specific machine
func (b *BidEngine) GetMachineBids(ctx context.Context, providerId, machineId *big.Int, filterStatus []BidStatus) ([]*Bid, error) {
	var result []*Bid

	// First check in-memory active bids
	b.bidsMutex.RLock()
	for _, bid := range b.activeBids {
		if bid.ProviderId.Cmp(providerId) == 0 && bid.MachineId.Cmp(machineId) == 0 {
			// If filterStatus is provided, check if bid status matches any of the requested statuses
			if len(filterStatus) > 0 {
				statusMatched := false
				for _, status := range filterStatus {
					if bid.Status == status {
						statusMatched = true
						break
					}
				}
				if !statusMatched {
					continue
				}
			}

			// Add a copy of the bid to the result to avoid race conditions
			bidCopy := *bid
			result = append(result, &bidCopy)
		}
	}
	b.bidsMutex.RUnlock()

	b.log.WithFields(logrus.Fields{
		"providerId":   providerId,
		"machineId":    machineId,
		"statusFilter": filterStatus,
		"bidCount":     len(result),
	}).Debug("Retrieved bids for machine")

	return result, nil
}

// GetActiveAndPendingBids retrieves all active and pending bids for a machine
func (b *BidEngine) GetActiveAndPendingBids(ctx context.Context, providerId, machineId *big.Int) ([]*Bid, error) {
	return b.GetMachineBids(ctx, providerId, machineId, []BidStatus{BidStatusAccepted, BidStatusPending})
}

// GetAcceptedBids retrieves all accepted bids for a machine
func (b *BidEngine) GetAcceptedBids(ctx context.Context, providerId, machineId *big.Int) ([]*Bid, error) {
	return b.GetMachineBids(ctx, providerId, machineId, []BidStatus{BidStatusAccepted})
}
