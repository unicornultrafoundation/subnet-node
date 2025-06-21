package bidengine

import (
	"context"
	"fmt"
	"sync"
)

// BidEngine is the main service that coordinates bidding and resource management
type BidEngine struct {
	config *BidEngineConfig

	// Contract interfaces
	bidMarket BidMarketContract
	provider  ProviderContract

	// Core components
	pricingEngine   PricingEngine
	resourceManager ResourceManager
	orderMonitor    OrderMonitor
	bidManager      BidManager
	storage         *Storage

	// Utilities
	logger  Logger
	metrics Metrics

	// Internal state
	mu              sync.RWMutex
	isRunning       bool
	stopChan        chan struct{}
	orderEventsChan chan *OrderEvent
	bidEventsChan   chan *BidResult

	// Tracking
	trackedOrders map[string]*Order
	trackedBids   map[string]*BidResult

	ctx    context.Context
	cancel context.CancelFunc
}

// Start starts the bid engine
func (be *BidEngine) Start(ctx context.Context) error {
	be.mu.Lock()
	defer be.mu.Unlock()

	if be.isRunning {
		return nil
	}

	be.logger.Info("Starting BidEngine")
	be.isRunning = true

	// Load persisted data from storage
	if err := be.loadPersistedData(ctx); err != nil {
		be.logger.Warn("Failed to load persisted data", "error", err)
	}

	be.logger.Info("BidEngine started successfully")
	return nil
}

// Stop stops the bid engine
func (be *BidEngine) Stop(ctx context.Context) error {
	be.mu.Lock()
	defer be.mu.Unlock()

	if !be.isRunning {
		return nil
	}

	be.logger.Info("Stopping BidEngine")
	be.isRunning = false
	close(be.stopChan)

	// Save current state to storage
	if err := be.savePersistedData(ctx); err != nil {
		be.logger.Warn("Failed to save persisted data", "error", err)
	}

	// Cancel context
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

// loadPersistedData loads data from storage on startup
func (be *BidEngine) loadPersistedData(ctx context.Context) error {
	be.logger.Info("Loading persisted data from storage")

	// Load orders
	orders, err := be.storage.ListOrders(ctx)
	if err != nil {
		return fmt.Errorf("failed to load orders: %w", err)
	}

	be.mu.Lock()
	for _, order := range orders {
		orderKey := order.ID.String()
		be.trackedOrders[orderKey] = order
		be.logger.Debug("Loaded order from storage", "orderID", orderKey)
	}
	be.mu.Unlock()

	// Load machines
	machines, err := be.storage.ListMachines(ctx)
	if err != nil {
		return fmt.Errorf("failed to load machines: %w", err)
	}

	// Register machines with resource manager
	for _, machine := range machines {
		if err := be.resourceManager.RegisterMachine(ctx, machine); err != nil {
			be.logger.Warn("Failed to register machine from storage", "machineID", machine.ID, "error", err)
		}
	}

	be.logger.Info("Loaded persisted data", "orders", len(orders), "machines", len(machines))
	return nil
}

// savePersistedData saves current state to storage
func (be *BidEngine) savePersistedData(ctx context.Context) error {
	be.logger.Info("Saving current state to storage")

	be.mu.RLock()
	orders := make([]*Order, 0, len(be.trackedOrders))
	for _, order := range be.trackedOrders {
		orders = append(orders, order)
	}
	be.mu.RUnlock()

	// Save orders
	for _, order := range orders {
		if err := be.storage.UpdateOrder(ctx, order); err != nil {
			be.logger.Warn("Failed to save order", "orderID", order.ID, "error", err)
		}
	}

	// Save machines
	machines := be.resourceManager.GetAllMachines(ctx)
	for _, machine := range machines {
		if err := be.storage.SaveMachine(ctx, machine); err != nil {
			be.logger.Warn("Failed to save machine", "machineID", machine.ID, "error", err)
		}
	}

	return nil
}

// GetStats returns bid engine statistics
func (be *BidEngine) GetStats() map[string]interface{} {
	be.mu.RLock()
	defer be.mu.RUnlock()

	return map[string]interface{}{
		"isRunning":     be.isRunning,
		"trackedOrders": len(be.trackedOrders),
		"trackedBids":   len(be.trackedBids),
	}
}

// GetBidManager returns the bid manager
func (be *BidEngine) GetBidManager() BidManager {
	return be.bidManager
}

// GetOrderMonitor returns the order monitor
func (be *BidEngine) GetOrderMonitor() OrderMonitor {
	return be.orderMonitor
}

// GetResourceManager returns the resource manager
func (be *BidEngine) GetResourceManager() ResourceManager {
	return be.resourceManager
}

// GetPricingEngine returns the pricing engine
func (be *BidEngine) GetPricingEngine() PricingEngine {
	return be.pricingEngine
}

// GetLogger returns the logger
func (be *BidEngine) GetLogger() Logger {
	return be.logger
}

// GetMetrics returns the metrics
func (be *BidEngine) GetMetrics() Metrics {
	return be.metrics
}
