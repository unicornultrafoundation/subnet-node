package bidengine

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	ds "github.com/ipfs/go-datastore"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/bidengine/contracts"
	managerpkg "github.com/unicornultrafoundation/subnet-node/bidengine/manager"
	"github.com/unicornultrafoundation/subnet-node/bidengine/metrics"
	"github.com/unicornultrafoundation/subnet-node/bidengine/order"
	pricingpkg "github.com/unicornultrafoundation/subnet-node/bidengine/princing"
	resourcepkg "github.com/unicornultrafoundation/subnet-node/bidengine/resource"
	storagepkg "github.com/unicornultrafoundation/subnet-node/bidengine/storage"
	"github.com/unicornultrafoundation/subnet-node/bidengine/types"
	"github.com/unicornultrafoundation/subnet-node/config"
)

// CreateBidEngine creates a new BidEngine instance with all dependencies
func CreateBidEngine(
	ctx context.Context,
	config *types.BidEngineConfig,
	bidMarket types.BidMarketContract,
	provider types.ProviderContract,
	datastore ds.Datastore,
) (*BidEngine, error) {

	if bidMarket == nil {
		return nil, fmt.Errorf("bidMarket contract cannot be nil")
	}

	if provider == nil {
		return nil, fmt.Errorf("provider contract cannot be nil")
	}

	if datastore == nil {
		return nil, fmt.Errorf("datastore cannot be nil")
	}

	// Create a simple logger (placeholder - should be implemented)
	logger := logrus.WithFields(logrus.Fields{
		"service": "bidengine",
	})

	// Create storage
	storage := storagepkg.NewStorage(datastore, logger)

	// Create metrics
	metricsService := metrics.NewMetrics()

	// Create pricing engine
	princingEngine := pricingpkg.NewEngine(config, logger)

	// Create resource manager
	resourceManager := resourcepkg.NewManager(config, provider, logger, metricsService, storage)

	// Create order monitor with bid manager as auto bidder
	orderMonitor := order.NewMonitor(config, bidMarket, logger, metricsService, storage)

	// Create bid manager first
	bidManager := managerpkg.NewManager(config, bidMarket, logger, metricsService, datastore, resourceManager, princingEngine, orderMonitor)

	// Create bid engine using the new constructor
	bidEngine := NewBidEngine(
		config,
		bidMarket,
		provider,
		princingEngine,
		resourceManager,
		orderMonitor,
		bidManager,
		storage,
		logger,
		metricsService,
	)

	return bidEngine, nil
}

// NewBidEngineFromConfig creates a new BidEngine instance from configuration
func NewBidEngineFromConfig(
	config *types.BidEngineConfig,
	client *ethclient.Client,
	transactor *bind.TransactOpts,
	datastore ds.Datastore,
) (*BidEngine, error) {
	// Create bid market contract
	bidMarket, err := contracts.NewBidMarketContract(client, config.BidMarketAddress, transactor)
	if err != nil {
		return nil, err
	}

	// Create provider contract
	provider, err := contracts.NewProviderContract(client, config.ProviderAddress, transactor)
	if err != nil {
		return nil, err
	}

	// Create and return the bid engine
	return CreateBidEngine(
		context.Background(),
		config,
		bidMarket,
		provider,
		datastore,
	)
}

// NewBidEngineFromConfigC creates a new BidEngine instance from config.C
func NewBidEngineFromConfigC(
	cfg *config.C,
	client *ethclient.Client,
	transactor *bind.TransactOpts,
	datastore ds.Datastore,
) (*BidEngine, error) {
	// Parse configuration from config.C
	bidEngineConfig, err := ParseBidEngineConfigFromC(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bid engine config: %w", err)
	}

	// Create bid market contract
	bidMarketAddress := common.HexToAddress(cfg.GetString("contracts.bid_market", config.DefaultBidMarketAddr))
	if bidMarketAddress == (common.Address{}) {
		return nil, fmt.Errorf("bid market address not found in config")
	}

	bidMarket, err := contracts.NewBidMarketContract(client, bidMarketAddress, transactor)
	if err != nil {
		return nil, fmt.Errorf("failed to create bid market contract: %w", err)
	}

	// Create provider contract
	providerAddress := common.HexToAddress(cfg.GetString("contracts.provider", config.DefaultSubnetProviderAddr))
	if providerAddress == (common.Address{}) {
		return nil, fmt.Errorf("provider address not found in config")
	}

	provider, err := contracts.NewProviderContract(client, providerAddress, transactor)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider contract: %w", err)
	}

	// Create and return the bid engine
	return CreateBidEngine(
		context.Background(),
		bidEngineConfig,
		bidMarket,
		provider,
		datastore,
	)
}

// ParseBidEngineConfigFromC parses BidEngineConfig from config.C
func ParseBidEngineConfigFromC(cfg *config.C) (*types.BidEngineConfig, error) {
	bcfg := &types.BidEngineConfig{}

	// Parse contract addresses
	bcfg.BidMarketAddress = common.HexToAddress(cfg.GetString("contracts.bid_market", config.DefaultBidMarketAddr))
	bcfg.ProviderAddress = common.HexToAddress(cfg.GetString("contracts.provider", config.DefaultSubnetProviderAddr))

	// Parse provider_id if present
	if cfg.IsSet("provider.id") {
		if v := cfg.Get("provider.id"); v != nil {
			switch val := v.(type) {
			case int:
				bcfg.ProviderID = big.NewInt(int64(val))
			case int64:
				bcfg.ProviderID = big.NewInt(val)
			case float64:
				bcfg.ProviderID = big.NewInt(int64(val))
			case string:
				if i, ok := new(big.Int).SetString(val, 10); ok {
					bcfg.ProviderID = i
				}
			}
		}
	}

	// If not set, use default values (should be set by application)
	if bcfg.ProviderID == nil {
		bcfg.ProviderID = big.NewInt(0)
	}
	if bcfg.ProviderWallet == (common.Address{}) {
		bcfg.ProviderWallet = common.Address{}
	}

	// Parse profit margins from percentage values
	minBidPercent := cfg.GetInt("bidengine.min_bid_percent", 70)
	maxBidPercent := cfg.GetInt("bidengine.max_bid_percent", 95)
	bcfg.BidStrategy.MinProfitMargin = float64(minBidPercent) / 100.0
	bcfg.BidStrategy.MaxProfitMargin = float64(maxBidPercent) / 100.0

	// Parse strategy parameters (use new structure if available, fallback to defaults)
	if cfg.IsSet("bidengine.strategy.min_profit_margin") {
		if val := cfg.Get("bidengine.strategy.min_profit_margin"); val != nil {
			if f, ok := val.(float64); ok {
				bcfg.BidStrategy.MinProfitMargin = f
			}
		}
	}
	if cfg.IsSet("bidengine.strategy.max_profit_margin") {
		if val := cfg.Get("bidengine.strategy.max_profit_margin"); val != nil {
			if f, ok := val.(float64); ok {
				bcfg.BidStrategy.MaxProfitMargin = f
			}
		}
	}
	if cfg.IsSet("bidengine.strategy.competitive_factor") {
		if val := cfg.Get("bidengine.strategy.competitive_factor"); val != nil {
			if f, ok := val.(float64); ok {
				bcfg.BidStrategy.CompetitiveFactor = f
			}
		}
	} else {
		bcfg.BidStrategy.CompetitiveFactor = 0.10 // Default value
	}
	if cfg.IsSet("bidengine.strategy.market_adjustment") {
		if val := cfg.Get("bidengine.strategy.market_adjustment"); val != nil {
			if f, ok := val.(float64); ok {
				bcfg.BidStrategy.MarketAdjustment = f
			}
		}
	} else {
		bcfg.BidStrategy.MarketAdjustment = 0.05 // Default value
	}

	// Parse resource weights (use new structure if available, fallback to defaults)
	bcfg.BidStrategy.ResourceWeight = types.ResourceWeight{
		CPU:     0.25, // Default values
		GPU:     0.35,
		Memory:  0.20,
		Disk:    0.15,
		Network: 0.05,
	}

	// Try to get resource weights from config
	if cfg.IsSet("bidengine.strategy.resource_weights") {
		if val := cfg.Get("bidengine.strategy.resource_weights.cpu"); val != nil {
			if f, ok := val.(float64); ok {
				bcfg.BidStrategy.ResourceWeight.CPU = f
			}
		}
		if val := cfg.Get("bidengine.strategy.resource_weights.gpu"); val != nil {
			if f, ok := val.(float64); ok {
				bcfg.BidStrategy.ResourceWeight.GPU = f
			}
		}
		if val := cfg.Get("bidengine.strategy.resource_weights.memory"); val != nil {
			if f, ok := val.(float64); ok {
				bcfg.BidStrategy.ResourceWeight.Memory = f
			}
		}
		if val := cfg.Get("bidengine.strategy.resource_weights.disk"); val != nil {
			if f, ok := val.(float64); ok {
				bcfg.BidStrategy.ResourceWeight.Disk = f
			}
		}
		if val := cfg.Get("bidengine.strategy.resource_weights.network"); val != nil {
			if f, ok := val.(float64); ok {
				bcfg.BidStrategy.ResourceWeight.Network = f
			}
		}
	}

	// Parse operational parameters
	bcfg.MaxConcurrentBids = cfg.GetInt("bidengine.max_concurrent_bids", 10)
	bcfg.BidTimeout = cfg.GetDuration("bidengine.bid_timeout", 30*time.Second)
	bcfg.OrderSyncInterval = cfg.GetDuration("bidengine.order_sync_interval", 30*time.Second)
	bcfg.BidCheckInterval = cfg.GetDuration("bidengine.bid_check_interval", 60*time.Second)

	// Parse logging configuration
	bcfg.LogLevel = cfg.GetString("logging.level", "INFO")
	bcfg.LogFile = cfg.GetString("logging.file_path", "")

	// Validate the configuration
	if err := ValidateBidEngineConfig(bcfg); err != nil {
		return nil, fmt.Errorf("invalid bid engine configuration: %w", err)
	}

	return bcfg, nil
}

// DefaultBidEngineConfig creates a default configuration for the bid engine
func DefaultBidEngineConfig(
	bidMarketAddress common.Address,
	providerAddress common.Address,
	providerID *big.Int,
	providerWallet common.Address,
) *types.BidEngineConfig {
	return &types.BidEngineConfig{
		BidMarketAddress: bidMarketAddress,
		ProviderAddress:  providerAddress,
		ProviderID:       providerID,
		ProviderWallet:   providerWallet,
		BidStrategy: types.BidStrategy{
			MinProfitMargin:   0.05, // 5% minimum profit margin
			MaxProfitMargin:   0.20, // 20% maximum profit margin
			CompetitiveFactor: 0.10, // 10% competitive factor
			MarketAdjustment:  0.05, // 5% market adjustment
			ResourceWeight: types.ResourceWeight{
				CPU:     0.25,
				GPU:     0.35,
				Memory:  0.20,
				Disk:    0.15,
				Network: 0.05,
			},
		},
		MaxConcurrentBids: 10,
		BidTimeout:        30 * time.Second,
		OrderSyncInterval: 30 * time.Second,
		BidCheckInterval:  60 * time.Second,
		LogLevel:          "INFO",
		LogFile:           "",
	}
}

// ValidateBidEngineConfig validates the bid engine configuration
func ValidateBidEngineConfig(config *types.BidEngineConfig) error {
	if config.BidMarketAddress == (common.Address{}) {
		return fmt.Errorf("bid market address is required")
	}
	if config.ProviderAddress == (common.Address{}) {
		return fmt.Errorf("provider address is required")
	}
	if config.ProviderID == nil {
		return fmt.Errorf("provider ID is required")
	}

	if config.MaxConcurrentBids <= 0 {
		return fmt.Errorf("max concurrent bids must be positive")
	}
	if config.BidTimeout <= 0 {
		return fmt.Errorf("bid timeout must be positive")
	}
	if config.OrderSyncInterval <= 0 {
		return fmt.Errorf("order sync interval must be positive")
	}
	if config.BidCheckInterval <= 0 {
		return fmt.Errorf("bid check interval must be positive")
	}

	// Validate bid strategy
	if config.BidStrategy.MinProfitMargin < 0 {
		return fmt.Errorf("minimum profit margin cannot be negative")
	}
	if config.BidStrategy.MaxProfitMargin < config.BidStrategy.MinProfitMargin {
		return fmt.Errorf("maximum profit margin must be greater than minimum profit margin")
	}
	if config.BidStrategy.CompetitiveFactor < 0 || config.BidStrategy.CompetitiveFactor > 1 {
		return fmt.Errorf("competitive factor must be between 0 and 1")
	}
	if config.BidStrategy.MarketAdjustment < 0 || config.BidStrategy.MarketAdjustment > 1 {
		return fmt.Errorf("market adjustment must be between 0 and 1")
	}

	// Validate resource weights
	weights := config.BidStrategy.ResourceWeight
	totalWeight := weights.CPU + weights.GPU + weights.Memory + weights.Disk + weights.Network
	if totalWeight <= 0 {
		return fmt.Errorf("resource weights must sum to a positive value")
	}

	return nil
}

// SimpleLogger is a basic logger implementation
type SimpleLogger struct{}

func (l *SimpleLogger) Debug(msg string, fields ...interface{}) {}
func (l *SimpleLogger) Info(msg string, fields ...interface{})  {}
func (l *SimpleLogger) Warn(msg string, fields ...interface{})  {}
func (l *SimpleLogger) Error(msg string, fields ...interface{}) {}
func (l *SimpleLogger) Fatal(msg string, fields ...interface{}) {}
