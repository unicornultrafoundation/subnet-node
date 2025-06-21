package bidengine

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	ds "github.com/ipfs/go-datastore"
)

// OrderMonitorService monitors orders and their lifecycle
type OrderMonitorService struct {
	config    *BidEngineConfig
	bidMarket BidMarketContract
	logger    Logger
	metrics   Metrics
	storage   *Storage

	mu        sync.RWMutex
	isRunning bool
	ctx       context.Context
	cancel    context.CancelFunc

	// Tracking
	trackedOrders map[string]*Order
	orderEvents   chan *OrderEvent
}

// NewOrderMonitor creates a new OrderMonitorService instance
func NewOrderMonitor(
	config *BidEngineConfig,
	bidMarket BidMarketContract,
	logger Logger,
	metrics Metrics,
	datastore ds.Datastore,
) *OrderMonitorService {
	ctx, cancel := context.WithCancel(context.Background())

	return &OrderMonitorService{
		config:        config,
		bidMarket:     bidMarket,
		logger:        logger,
		metrics:       metrics,
		storage:       NewStorage(datastore, logger),
		ctx:           ctx,
		cancel:        cancel,
		trackedOrders: make(map[string]*Order),
		orderEvents:   make(chan *OrderEvent, 100),
	}
}

// Start starts the order monitor
func (om *OrderMonitorService) Start(ctx context.Context) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	if om.isRunning {
		return nil
	}

	om.logger.Info("Starting OrderMonitor")
	om.isRunning = true

	// Load persisted orders from storage
	if err := om.loadPersistedOrders(ctx); err != nil {
		om.logger.Warn("Failed to load persisted orders", "error", err)
	}

	// Start monitoring goroutines
	go om.orderSyncLoop(ctx)
	go om.orderExpiryCheckLoop(ctx)
	go om.eventProcessingLoop(ctx)

	om.logger.Info("OrderMonitor started successfully")
	return nil
}

// Stop stops the order monitor
func (om *OrderMonitorService) Stop(ctx context.Context) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	if !om.isRunning {
		return nil
	}

	om.logger.Info("Stopping OrderMonitor")
	om.isRunning = false

	// Save current state to storage
	if err := om.savePersistedOrders(ctx); err != nil {
		om.logger.Warn("Failed to save persisted orders", "error", err)
	}

	// Cancel context
	om.cancel()

	om.logger.Info("OrderMonitor stopped successfully")
	return nil
}

// TrackOrder starts tracking an order
func (om *OrderMonitorService) TrackOrder(ctx context.Context, orderID *big.Int) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	orderIDStr := orderID.String()

	// Check if order is already being tracked
	if _, exists := om.trackedOrders[orderIDStr]; exists {
		return fmt.Errorf("order %s is already being tracked", orderIDStr)
	}

	// Get order from blockchain
	order, err := om.bidMarket.GetOrder(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order from blockchain: %w", err)
	}

	// Track order
	om.trackedOrders[orderIDStr] = order

	// Save order to storage
	if err := om.storage.SaveOrder(ctx, order); err != nil {
		om.logger.Warn("Failed to save order to storage", "error", err)
	}

	om.logger.Info("Started tracking order", "orderID", orderID)
	om.metrics.IncrementOrdersTracked()
	return nil
}

// UntrackOrder stops tracking an order
func (om *OrderMonitorService) UntrackOrder(ctx context.Context, orderID *big.Int) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	orderIDStr := orderID.String()

	if _, exists := om.trackedOrders[orderIDStr]; !exists {
		return fmt.Errorf("order %s is not being tracked", orderIDStr)
	}

	delete(om.trackedOrders, orderIDStr)
	om.logger.Info("Stopped tracking order", "orderID", orderID)
	return nil
}

// GetTrackedOrders returns all currently tracked orders
func (om *OrderMonitorService) GetTrackedOrders(ctx context.Context) ([]*big.Int, error) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	var orderIDs []*big.Int
	for orderIDStr := range om.trackedOrders {
		orderID, _ := new(big.Int).SetString(orderIDStr, 10)
		orderIDs = append(orderIDs, orderID)
	}

	return orderIDs, nil
}

// MonitorOrderStatus monitors the status of a specific order
func (om *OrderMonitorService) MonitorOrderStatus(ctx context.Context, orderID *big.Int) error {
	om.mu.RLock()
	orderIDStr := orderID.String()
	_, exists := om.trackedOrders[orderIDStr]
	om.mu.RUnlock()

	if !exists {
		return fmt.Errorf("order %s is not being tracked", orderIDStr)
	}

	// Get current order status from blockchain
	order, err := om.bidMarket.GetOrder(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order status: %w", err)
	}

	// Update tracked order
	om.mu.Lock()
	om.trackedOrders[orderIDStr] = order
	om.mu.Unlock()

	// Update order in storage
	if err := om.storage.UpdateOrder(ctx, order); err != nil {
		om.logger.Warn("Failed to update order in storage", "error", err)
	}

	om.logger.Debug("Updated order status", "orderID", orderID, "status", order.Status)
	return nil
}

// CheckOrderExpiry checks if an order has expired
func (om *OrderMonitorService) CheckOrderExpiry(ctx context.Context, orderID *big.Int) error {
	om.mu.RLock()
	orderIDStr := orderID.String()
	order, exists := om.trackedOrders[orderIDStr]
	om.mu.RUnlock()

	if !exists {
		return fmt.Errorf("order %s is not being tracked", orderIDStr)
	}

	// Check if order has expired
	if order.ExpiredAt != nil {
		currentTime := big.NewInt(time.Now().Unix())
		if currentTime.Cmp(order.ExpiredAt) > 0 {
			om.logger.Info("Order has expired", "orderID", orderID, "expiredAt", order.ExpiredAt)

			// Update order status
			order.Status = OrderStatusExpired

			// Update in storage
			if err := om.storage.UpdateOrder(ctx, order); err != nil {
				om.logger.Warn("Failed to update expired order in storage", "error", err)
			}

			// Send expiry event
			om.orderEvents <- &OrderEvent{
				Type:      "expired",
				OrderID:   orderID,
				Timestamp: time.Now(),
				Data:      order,
			}
		}
	}

	return nil
}

// HandleOrderEvent handles order events
func (om *OrderMonitorService) HandleOrderEvent(ctx context.Context, event *OrderEvent) error {
	om.logger.Info("Handling order event",
		"type", event.Type,
		"orderID", event.OrderID)

	switch event.Type {
	case "created":
		// Start tracking new order
		if err := om.TrackOrder(ctx, event.OrderID); err != nil {
			om.logger.Warn("Failed to track new order", "orderID", event.OrderID, "error", err)
		}

	case "closed", "expired", "cancelled":
		// Stop tracking completed order
		if err := om.UntrackOrder(ctx, event.OrderID); err != nil {
			om.logger.Warn("Failed to untrack completed order", "orderID", event.OrderID, "error", err)
		}

		// Record completion metrics
		om.metrics.IncrementOrdersCompleted()
	}

	return nil
}

// orderSyncLoop periodically syncs orders with blockchain
func (om *OrderMonitorService) orderSyncLoop(ctx context.Context) {
	ticker := time.NewTicker(om.config.OrderSyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			om.syncOrders(ctx)
		}
	}
}

// syncOrders syncs all tracked orders with blockchain
func (om *OrderMonitorService) syncOrders(ctx context.Context) {
	om.mu.RLock()
	orderIDs := make([]*big.Int, 0, len(om.trackedOrders))
	for orderIDStr := range om.trackedOrders {
		orderID, _ := new(big.Int).SetString(orderIDStr, 10)
		orderIDs = append(orderIDs, orderID)
	}
	om.mu.RUnlock()

	for _, orderID := range orderIDs {
		if err := om.MonitorOrderStatus(ctx, orderID); err != nil {
			om.logger.Warn("Failed to sync order status", "orderID", orderID, "error", err)
		}
	}
}

// orderExpiryCheckLoop periodically checks for expired orders
func (om *OrderMonitorService) orderExpiryCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(om.config.OrderSyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			om.checkOrderExpiries(ctx)
		}
	}
}

// checkOrderExpiries checks all tracked orders for expiry
func (om *OrderMonitorService) checkOrderExpiries(ctx context.Context) {
	om.mu.RLock()
	orderIDs := make([]*big.Int, 0, len(om.trackedOrders))
	for orderIDStr := range om.trackedOrders {
		orderID, _ := new(big.Int).SetString(orderIDStr, 10)
		orderIDs = append(orderIDs, orderID)
	}
	om.mu.RUnlock()

	for _, orderID := range orderIDs {
		if err := om.CheckOrderExpiry(ctx, orderID); err != nil {
			om.logger.Warn("Failed to check order expiry", "orderID", orderID, "error", err)
		}
	}
}

// eventProcessingLoop processes order events
func (om *OrderMonitorService) eventProcessingLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-om.orderEvents:
			if err := om.HandleOrderEvent(ctx, event); err != nil {
				om.logger.Warn("Failed to handle order event", "error", err)
			}
		}
	}
}

// loadPersistedOrders loads orders from storage on startup
func (om *OrderMonitorService) loadPersistedOrders(ctx context.Context) error {
	om.logger.Info("Loading persisted orders from storage")

	orders, err := om.storage.ListOrders(ctx)
	if err != nil {
		return fmt.Errorf("failed to load orders: %w", err)
	}

	for _, order := range orders {
		orderIDStr := order.ID.String()
		om.trackedOrders[orderIDStr] = order
		om.logger.Debug("Loaded order from storage", "orderID", orderIDStr)
	}

	om.logger.Info("Loaded persisted orders", "count", len(orders))
	return nil
}

// savePersistedOrders saves current order state to storage
func (om *OrderMonitorService) savePersistedOrders(ctx context.Context) error {
	om.logger.Info("Saving persisted orders to storage")

	om.mu.RLock()
	defer om.mu.RUnlock()

	for orderIDStr, order := range om.trackedOrders {
		if err := om.storage.UpdateOrder(ctx, order); err != nil {
			om.logger.Warn("Failed to save order to storage", "orderID", orderIDStr, "error", err)
		}
	}

	return nil
}

// GetStats returns order monitor statistics
func (om *OrderMonitorService) GetStats() map[string]interface{} {
	om.mu.RLock()
	defer om.mu.RUnlock()

	return map[string]interface{}{
		"trackedOrders": len(om.trackedOrders),
		"isRunning":     om.isRunning,
	}
}
