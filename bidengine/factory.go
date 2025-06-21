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
	"github.com/unicornultrafoundation/subnet-node/config"
)

// NewBidEngine creates a new BidEngine instance with all dependencies
func NewBidEngine(
	ctx context.Context,
	config *BidEngineConfig,
	bidMarket BidMarketContract,
	provider ProviderContract,
	datastore ds.Datastore,
) (*BidEngine, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if bidMarket == nil {
		return nil, fmt.Errorf("bidMarket contract cannot be nil")
	}

	if provider == nil {
		return nil, fmt.Errorf("provider contract cannot be nil")
	}

	if datastore == nil {
		return nil, fmt.Errorf("datastore cannot be nil")
	}

	// Create logger
	logger, err := NewLogger(config.LogLevel, config.LogFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Create metrics
	metrics := NewMetrics()

	// Create pricing engine
	pricingEngine := NewPricingEngine(config, logger)

	// Create resource manager
	resourceManager := NewResourceManager(config, provider, logger, metrics, datastore)

	// Create order monitor
	orderMonitor := NewOrderMonitor(config, bidMarket, logger, metrics, datastore)

	// Create bid manager
	bidManager := NewBidManager(config, bidMarket, logger, metrics, datastore)

	// Create storage
	storage := NewStorage(datastore, logger)

	// Create bid engine
	bidEngine := &BidEngine{
		config:          config,
		bidMarket:       bidMarket,
		provider:        provider,
		pricingEngine:   pricingEngine,
		resourceManager: resourceManager,
		orderMonitor:    orderMonitor,
		bidManager:      bidManager,
		metrics:         metrics,
		storage:         storage,
		logger:          logger,
	}

	// Load persisted data
	if err := bidEngine.loadPersistedData(ctx); err != nil {
		logger.Warn("Failed to load persisted data", "error", err)
	}

	return bidEngine, nil
}

// NewBidEngineFromConfig creates a new BidEngine instance from configuration
func NewBidEngineFromConfig(
	config *BidEngineConfig,
	client *ethclient.Client,
	transactor *bind.TransactOpts,
	datastore ds.Datastore,
) (*BidEngine, error) {
	// Create bid market contract
	bidMarket, err := NewBidMarketContract(client, config.BidMarketAddress, transactor)
	if err != nil {
		return nil, err
	}

	// Create provider contract
	provider, err := NewProviderContract(client, config.ProviderAddress, transactor)
	if err != nil {
		return nil, err
	}

	// Create and return the bid engine
	return NewBidEngine(
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
	bidMarketAddress := common.HexToAddress(cfg.GetString("contracts.bid_market", ""))
	if bidMarketAddress == (common.Address{}) {
		return nil, fmt.Errorf("bid market address not found in config")
	}

	bidMarket, err := NewBidMarketContract(client, bidMarketAddress, transactor)
	if err != nil {
		return nil, fmt.Errorf("failed to create bid market contract: %w", err)
	}

	// Create provider contract
	providerAddress := common.HexToAddress(cfg.GetString("contracts.provider", ""))
	if providerAddress == (common.Address{}) {
		return nil, fmt.Errorf("provider address not found in config")
	}

	provider, err := NewProviderContract(client, providerAddress, transactor)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider contract: %w", err)
	}

	// Create and return the bid engine
	return NewBidEngine(
		context.Background(),
		bidEngineConfig,
		bidMarket,
		provider,
		datastore,
	)
}

// ParseBidEngineConfigFromC parses BidEngineConfig from config.C
func ParseBidEngineConfigFromC(cfg *config.C) (*BidEngineConfig, error) {
	config := &BidEngineConfig{}

	// Parse contract addresses
	config.BidMarketAddress = common.HexToAddress(cfg.GetString("contracts.bid_market", ""))
	config.ProviderAddress = common.HexToAddress(cfg.GetString("contracts.provider", ""))

	// Parse provider_id if present
	if cfg.IsSet("bidengine.provider_id") {
		if v := cfg.Get("bidengine.provider_id"); v != nil {
			switch val := v.(type) {
			case int:
				config.ProviderID = big.NewInt(int64(val))
			case int64:
				config.ProviderID = big.NewInt(val)
			case float64:
				config.ProviderID = big.NewInt(int64(val))
			case string:
				if i, ok := new(big.Int).SetString(val, 10); ok {
					config.ProviderID = i
				}
			}
		}
	}
	// Parse provider_wallet if present
	if cfg.IsSet("bidengine.provider_wallet") {
		if v := cfg.GetString("bidengine.provider_wallet", ""); v != "" {
			config.ProviderWallet = common.HexToAddress(v)
		}
	}

	// If not set, use default values (should be set by application)
	if config.ProviderID == nil {
		config.ProviderID = big.NewInt(0)
	}
	if config.ProviderWallet == (common.Address{}) {
		config.ProviderWallet = common.Address{}
	}

	// Parse profit margins from percentage values
	minBidPercent := cfg.GetInt("bidengine.min_bid_percent", 70)
	maxBidPercent := cfg.GetInt("bidengine.max_bid_percent", 95)
	config.BidStrategy.MinProfitMargin = float64(minBidPercent) / 100.0
	config.BidStrategy.MaxProfitMargin = float64(maxBidPercent) / 100.0

	// Parse strategy parameters (use new structure if available, fallback to defaults)
	if cfg.IsSet("bidengine.strategy.min_profit_margin") {
		if val := cfg.Get("bidengine.strategy.min_profit_margin"); val != nil {
			if f, ok := val.(float64); ok {
				config.BidStrategy.MinProfitMargin = f
			}
		}
	}
	if cfg.IsSet("bidengine.strategy.max_profit_margin") {
		if val := cfg.Get("bidengine.strategy.max_profit_margin"); val != nil {
			if f, ok := val.(float64); ok {
				config.BidStrategy.MaxProfitMargin = f
			}
		}
	}
	if cfg.IsSet("bidengine.strategy.competitive_factor") {
		if val := cfg.Get("bidengine.strategy.competitive_factor"); val != nil {
			if f, ok := val.(float64); ok {
				config.BidStrategy.CompetitiveFactor = f
			}
		}
	} else {
		config.BidStrategy.CompetitiveFactor = 0.10 // Default value
	}
	if cfg.IsSet("bidengine.strategy.market_adjustment") {
		if val := cfg.Get("bidengine.strategy.market_adjustment"); val != nil {
			if f, ok := val.(float64); ok {
				config.BidStrategy.MarketAdjustment = f
			}
		}
	} else {
		config.BidStrategy.MarketAdjustment = 0.05 // Default value
	}

	// Parse resource weights (use new structure if available, fallback to defaults)
	config.BidStrategy.ResourceWeight = ResourceWeight{
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
				config.BidStrategy.ResourceWeight.CPU = f
			}
		}
		if val := cfg.Get("bidengine.strategy.resource_weights.gpu"); val != nil {
			if f, ok := val.(float64); ok {
				config.BidStrategy.ResourceWeight.GPU = f
			}
		}
		if val := cfg.Get("bidengine.strategy.resource_weights.memory"); val != nil {
			if f, ok := val.(float64); ok {
				config.BidStrategy.ResourceWeight.Memory = f
			}
		}
		if val := cfg.Get("bidengine.strategy.resource_weights.disk"); val != nil {
			if f, ok := val.(float64); ok {
				config.BidStrategy.ResourceWeight.Disk = f
			}
		}
		if val := cfg.Get("bidengine.strategy.resource_weights.network"); val != nil {
			if f, ok := val.(float64); ok {
				config.BidStrategy.ResourceWeight.Network = f
			}
		}
	}

	// Parse operational parameters
	config.MaxConcurrentBids = cfg.GetInt("bidengine.max_concurrent_bids", 10)
	config.BidTimeout = cfg.GetDuration("bidengine.bid_timeout", 30*time.Second)
	config.OrderSyncInterval = cfg.GetDuration("bidengine.order_sync_interval", 30*time.Second)
	config.BidCheckInterval = cfg.GetDuration("bidengine.bid_check_interval", 60*time.Second)

	// Parse logging configuration
	config.LogLevel = cfg.GetString("logging.level", "INFO")
	config.LogFile = cfg.GetString("logging.file_path", "")

	// Validate the configuration
	if err := ValidateBidEngineConfig(config); err != nil {
		return nil, fmt.Errorf("invalid bid engine configuration: %w", err)
	}

	return config, nil
}

// DefaultBidEngineConfig creates a default configuration for the bid engine
func DefaultBidEngineConfig(
	bidMarketAddress common.Address,
	providerAddress common.Address,
	providerID *big.Int,
	providerWallet common.Address,
) *BidEngineConfig {
	return &BidEngineConfig{
		BidMarketAddress: bidMarketAddress,
		ProviderAddress:  providerAddress,
		ProviderID:       providerID,
		ProviderWallet:   providerWallet,
		BidStrategy: BidStrategy{
			MinProfitMargin:   0.05, // 5% minimum profit margin
			MaxProfitMargin:   0.20, // 20% maximum profit margin
			CompetitiveFactor: 0.10, // 10% competitive factor
			MarketAdjustment:  0.05, // 5% market adjustment
			ResourceWeight: ResourceWeight{
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
func ValidateBidEngineConfig(config *BidEngineConfig) error {
	if config.BidMarketAddress == (common.Address{}) {
		return fmt.Errorf("bid market address is required")
	}
	if config.ProviderAddress == (common.Address{}) {
		return fmt.Errorf("provider address is required")
	}
	if config.ProviderID == nil {
		return fmt.Errorf("provider ID is required")
	}
	if config.ProviderWallet == (common.Address{}) {
		return fmt.Errorf("provider wallet is required")
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
