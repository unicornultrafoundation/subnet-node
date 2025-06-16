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
	providerId    *big.Int // Our provider ID
	store         *Store   // Added datastore interface

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
	if b.store != nil {
		bids, err := b.store.ListActiveBids(ctx)
		if err != nil {
			b.log.WithError(err).Warn("Failed to load active bids from datastore")
		} else {
			b.log.WithField("count", len(bids)).Info("Loaded active bids from datastore")
			b.bidsMutex.Lock()
			for _, bid := range bids {
				// Only load non-finalized bids into memory
				if bid.Status != BidStatusExpired && bid.Status != BidStatusRejected {
					b.activeBids[bid.ID.String()] = bid
				}
			}
			b.bidsMutex.Unlock()
		}
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
