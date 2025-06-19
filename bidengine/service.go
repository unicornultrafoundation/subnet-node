package bidengine

import (
	"context"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/holiman/uint256"
	"github.com/sirupsen/logrus"
	bconfig "github.com/unicornultrafoundation/subnet-node/bidengine/config"
	"github.com/unicornultrafoundation/subnet-node/bidengine/contracts"
	"github.com/unicornultrafoundation/subnet-node/bidengine/market"
	bs "github.com/unicornultrafoundation/subnet-node/bidengine/provider"
	"github.com/unicornultrafoundation/subnet-node/bidengine/store"
	"github.com/unicornultrafoundation/subnet-node/bidengine/types"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/repo"
)

// Service is responsible for managing bids on the BidMarket contract
type Service struct {
	cfg           *config.C
	client        *ethclient.Client
	auth          *bind.TransactOpts
	log           *logrus.Logger
	provider      types.ProviderService
	bidMarket     types.MarketService
	providerAddr  common.Address
	bidMarketAddr common.Address
	activeBids    map[string]*types.Bid
	bidsMutex     sync.RWMutex
	bidConfig     types.BidConfig
	providerId    *uint256.Int       // Our provider ID
	store         types.BidDatastore // Added datastore interface

	shutdown chan struct{}
	wg       sync.WaitGroup
}

// NewService creates a new instance of the Service
func NewService(cfg *config.C, log *logrus.Logger, acc account.Service, ds repo.Datastore) (*Service, error) {
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

	store, err := store.NewStore(ds, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create bid store: %v", err)
	}

	return &Service{
		cfg:           cfg,
		client:        client,
		log:           log,
		provider:      bs.NewProviderService(provider),
		bidMarket:     market.NewMarketService(bidMarket),
		providerAddr:  providerAddr,
		bidMarketAddr: bidMarketAddr,
		activeBids:    make(map[string]*types.Bid),
		shutdown:      make(chan struct{}),
		auth:          auth,
		store:         store,
	}, nil
}

// Start initializes and starts the bid engine processes
func (b *Service) Start(ctx context.Context) error {
	b.log.Info("Starting BidEngine...")

	// Load bid configuration
	b.bidConfig = bconfig.LoadBidConfig(b.cfg)

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
func (b *Service) Stop(ctx context.Context) error {
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
