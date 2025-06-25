package bidengine

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/bidengine/types"
)

// BidEngine is the main service that coordinates bidding and resource management
type BidEngine struct {
	config *types.BidEngineConfig

	// Contract interfaces
	bidMarket types.BidMarketContract
	provider  types.ProviderContract

	// Core components
	pricingEngine   types.PricingEngine
	resourceManager types.ResourceManager
	orderMonitor    types.OrderMonitor
	bidManager      types.BidManager
	storage         types.Storage

	// Utilities
	logger  *logrus.Entry
	metrics types.Metrics

	// Internal state
	mu        sync.RWMutex
	isRunning bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewBidEngine creates a new BidEngine instance
func NewBidEngine(
	config *types.BidEngineConfig,
	bidMarket types.BidMarketContract,
	provider types.ProviderContract,
	pricingEngine types.PricingEngine,
	resourceManager types.ResourceManager,
	orderMonitor types.OrderMonitor,
	bidManager types.BidManager,
	storage types.Storage,
	logger *logrus.Entry,
	metrics types.Metrics,
) *BidEngine {
	ctx, cancel := context.WithCancel(context.Background())

	return &BidEngine{
		config:          config,
		bidMarket:       bidMarket,
		provider:        provider,
		pricingEngine:   pricingEngine,
		resourceManager: resourceManager,
		orderMonitor:    orderMonitor,
		bidManager:      bidManager,
		storage:         storage,
		logger:          logger,
		metrics:         metrics,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Start starts the bid engine and all its components
func (be *BidEngine) Start(ctx context.Context) error {
	be.mu.Lock()
	defer be.mu.Unlock()

	if be.isRunning {
		return nil
	}

	be.logger.Info("Starting BidEngine")

	// Start components if they have Start method
	if starter, ok := be.resourceManager.(interface{ Start(context.Context) error }); ok {
		if err := starter.Start(ctx); err != nil {
			return fmt.Errorf("failed to start resource manager: %w", err)
		}
	}

	if starter, ok := be.orderMonitor.(interface{ Start(context.Context) error }); ok {
		if err := starter.Start(ctx); err != nil {
			return fmt.Errorf("failed to start order monitor: %w", err)
		}
	}

	if starter, ok := be.bidManager.(interface{ Start(context.Context) error }); ok {
		if err := starter.Start(ctx); err != nil {
			return fmt.Errorf("failed to start bid manager: %w", err)
		}
	}

	be.isRunning = true
	be.logger.Info("BidEngine started successfully")
	return nil
}

// Stop stops the bid engine and all its components
func (be *BidEngine) Stop(ctx context.Context) error {
	be.mu.Lock()
	defer be.mu.Unlock()

	if !be.isRunning {
		return nil
	}

	be.logger.Info("Stopping BidEngine")

	// Stop components in reverse order if they have Stop method
	if stopper, ok := be.bidManager.(interface{ Stop(context.Context) error }); ok {
		if err := stopper.Stop(ctx); err != nil {
			be.logger.Warn("Failed to stop bid manager", "error", err)
		}
	}

	if stopper, ok := be.orderMonitor.(interface{ Stop(context.Context) error }); ok {
		if err := stopper.Stop(ctx); err != nil {
			be.logger.Warn("Failed to stop order monitor", "error", err)
		}
	}

	if stopper, ok := be.resourceManager.(interface{ Stop(context.Context) error }); ok {
		if err := stopper.Stop(ctx); err != nil {
			be.logger.Warn("Failed to stop resource manager", "error", err)
		}
	}

	be.isRunning = false
	be.cancel()

	be.logger.Info("BidEngine stopped successfully")
	return nil
}

// IsRunning returns whether the bid engine is running
func (be *BidEngine) IsRunning() bool {
	be.mu.RLock()
	defer be.mu.RUnlock()
	return be.isRunning
}

// ForceSync forces synchronization of all components
func (be *BidEngine) ForceSync(ctx context.Context) error {
	be.logger.Info("Forcing synchronization of all components")

	// Force sync components if they have ForceSync method
	if syncer, ok := be.resourceManager.(interface{ ForceSync(context.Context) error }); ok {
		if err := syncer.ForceSync(ctx); err != nil {
			be.logger.Warn("Failed to force sync resource manager", "error", err)
		}
	}

	if syncer, ok := be.orderMonitor.(interface{ ForceSync(context.Context) error }); ok {
		if err := syncer.ForceSync(ctx); err != nil {
			be.logger.Warn("Failed to force sync order monitor", "error", err)
		}
	}

	if syncer, ok := be.bidManager.(interface{ ForceSync(context.Context) error }); ok {
		if err := syncer.ForceSync(ctx); err != nil {
			be.logger.Warn("Failed to force sync bid manager", "error", err)
		}
	}

	be.logger.Info("Force synchronization completed")
	return nil
}

// GetStats returns bid engine statistics
func (be *BidEngine) GetStats() map[string]interface{} {
	be.mu.RLock()
	defer be.mu.RUnlock()

	stats := map[string]interface{}{
		"isRunning": be.isRunning,
	}

	// Get stats from components if they have GetStats method
	if statser, ok := be.resourceManager.(interface{ GetStats() map[string]interface{} }); ok {
		if resourceStats := statser.GetStats(); resourceStats != nil {
			stats["resourceManager"] = resourceStats
		}
	}

	if statser, ok := be.orderMonitor.(interface{ GetStats() map[string]interface{} }); ok {
		if orderStats := statser.GetStats(); orderStats != nil {
			stats["orderMonitor"] = orderStats
		}
	}

	if statser, ok := be.bidManager.(interface{ GetStats() map[string]interface{} }); ok {
		if bidStats := statser.GetStats(); bidStats != nil {
			stats["bidManager"] = bidStats
		}
	}

	return stats
}

// GetBidManager returns the bid manager
func (be *BidEngine) GetBidManager() types.BidManager {
	return be.bidManager
}

// GetOrderMonitor returns the order monitor
func (be *BidEngine) GetOrderMonitor() types.OrderMonitor {
	return be.orderMonitor
}

// GetResourceManager returns the resource manager
func (be *BidEngine) GetResourceManager() types.ResourceManager {
	return be.resourceManager
}

// GetPricingEngine returns the pricing engine
func (be *BidEngine) GetPricingEngine() types.PricingEngine {
	return be.pricingEngine
}

// GetMetrics returns the metrics
func (be *BidEngine) GetMetrics() types.Metrics {
	return be.metrics
}

// GetBidMarket returns the bid market contract
func (be *BidEngine) GetBidMarket() types.BidMarketContract {
	return be.bidMarket
}

// GetProvider returns the provider contract
func (be *BidEngine) GetProvider() types.ProviderContract {
	return be.provider
}

// GetStorage returns the storage
func (be *BidEngine) GetStorage() types.Storage {
	return be.storage
}
