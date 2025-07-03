package order

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/bidengine/types"
)

// Monitor monitors orders and their lifecycle
type Monitor struct {
	config    *types.BidEngineConfig
	bidMarket types.BidMarketContract
	logger    *logrus.Entry
	metrics   types.Metrics
	storage   types.Storage

	mu        sync.RWMutex
	isRunning bool
	ctx       context.Context
	cancel    context.CancelFunc

	// Tracking
	trackedOrders map[string]*types.Order
	orderEvents   chan *types.OrderEvent

	// Event handlers
	eventHandlers map[types.OrderEventType][]func(*types.OrderEvent)

	// Event watching
	contractEventSink chan *types.OrderEvent
}

// NewMonitor creates a new Monitor instance
func NewMonitor(
	config *types.BidEngineConfig,
	bidMarket types.BidMarketContract,
	logger *logrus.Entry,
	metrics types.Metrics,
	storage types.Storage,
) *Monitor {
	ctx, cancel := context.WithCancel(context.Background())

	om := &Monitor{
		config:            config,
		bidMarket:         bidMarket,
		logger:            logger,
		metrics:           metrics,
		storage:           storage,
		ctx:               ctx,
		cancel:            cancel,
		trackedOrders:     make(map[string]*types.Order),
		orderEvents:       make(chan *types.OrderEvent, 100),
		eventHandlers:     make(map[types.OrderEventType][]func(*types.OrderEvent)),
		contractEventSink: make(chan *types.OrderEvent, 50),
	}

	return om
}

// RegisterEventHandler registers an event handler for a specific event type
func (om *Monitor) RegisterEventHandler(eventType types.OrderEventType, handler func(*types.OrderEvent)) {
	om.mu.Lock()
	defer om.mu.Unlock()

	if om.eventHandlers[eventType] == nil {
		om.eventHandlers[eventType] = make([]func(*types.OrderEvent), 0)
	}
	om.eventHandlers[eventType] = append(om.eventHandlers[eventType], handler)
}

// emitEvent emits an order event to all registered handlers
func (om *Monitor) emitEvent(event *types.OrderEvent) {
	om.mu.RLock()
	handlers := om.eventHandlers[event.Type]
	om.mu.RUnlock()

	for _, handler := range handlers {
		go func(h func(*types.OrderEvent)) {
			defer func() {
				if r := recover(); r != nil {
					om.logger.Error("Event handler panicked", "eventType", event.Type, "orderID", event.OrderID, "panic", r)
				}
			}()
			h(event)
		}(handler)
	}

	// Also send to event channel for external consumers
	select {
	case om.orderEvents <- event:
	default:
		om.logger.Warn("Order event channel is full, dropping event", "eventType", event.Type, "orderID", event.OrderID)
	}
}

// Start starts the order monitor
func (om *Monitor) Start(ctx context.Context) error {
	if om.isRunning {
		return nil
	}

	om.logger.Info("Starting OrderMonitor")
	om.mu.Lock()
	om.isRunning = true
	om.mu.Unlock()

	// Load persisted orders from storage
	if err := om.loadPersistedOrders(ctx); err != nil {
		om.logger.Warn("Failed to load persisted orders", "error", err)
	}

	// Start monitoring goroutines
	go om.orderStatusSyncLoop(ctx) // Sync status and check expiry of tracked orders
	go om.eventProcessingLoop(ctx) // Process events
	go om.startEventWatching(ctx)

	om.logger.Info("OrderMonitor started successfully")
	return nil
}

// startEventWatching starts watching contract events
func (om *Monitor) startEventWatching(ctx context.Context) {
	om.logger.Info("Starting contract event watching")

	// Start watching OrderCreated events only
	go om.watchOrderCreatedEvents(ctx)

	// Process contract events
	go om.processContractEvents(ctx)
}

// watchOrderCreatedEvents watches for new order creation events
func (om *Monitor) watchOrderCreatedEvents(ctx context.Context) {
	om.logger.Info("Starting to watch OrderCreated events")

	err := om.bidMarket.WatchOrderCreated(ctx, om.contractEventSink)
	if err != nil {
		om.logger.Error("Failed to watch OrderCreated events", "error", err)
		// Fallback to polling only
		return
	}

	om.logger.Info("OrderCreated event watching started successfully")
}

// processContractEvents processes events from the contract
func (om *Monitor) processContractEvents(ctx context.Context) {
	om.logger.Info("Starting contract event processing")

	for {
		select {
		case <-ctx.Done():
			om.logger.Info("Contract event processing stopped")
			return
		case event := <-om.contractEventSink:
			om.handleContractEvent(ctx, event)
		}
	}
}

// handleContractEvent handles events from the smart contract
func (om *Monitor) handleContractEvent(ctx context.Context, event *types.OrderEvent) {
	om.logger.Info("Received contract event",
		"type", event.Type,
		"orderID", event.OrderID.String())

	switch event.Type {
	case "OrderCreated":
		om.handleOrderCreatedEvent(ctx, event)
	default:
		om.logger.Debug("Ignoring contract event type", "type", event.Type)
	}
}

// handleOrderCreatedEvent handles OrderCreated events from contract
func (om *Monitor) handleOrderCreatedEvent(ctx context.Context, event *types.OrderEvent) {
	om.logger.Info("Handling OrderCreated event", "orderID", event.OrderID.String())

	// Get the full order details from contract
	order, err := om.bidMarket.GetOrder(ctx, event.OrderID)
	if err != nil {
		om.logger.Warn("Failed to get order details for created event",
			"orderID", event.OrderID, "error", err)
		return
	}

	// Only track open orders
	if order.Status == types.OrderStatusOpen {
		// Check if we're already tracking this order
		om.mu.RLock()
		_, alreadyTracked := om.trackedOrders[event.OrderID.String()]
		om.mu.RUnlock()

		if !alreadyTracked {
			om.logger.Info("New order created via event", "orderID", event.OrderID.String())

			// Track the new order
			if err := om.TrackOrder(ctx, event.OrderID); err != nil {
				om.logger.Warn("Failed to track new order from event",
					"orderID", event.OrderID, "error", err)
			}
		}
	}
}

// Stop stops the order monitor
func (om *Monitor) Stop(ctx context.Context) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	if !om.isRunning {
		return nil
	}

	om.logger.Info("Stopping OrderMonitor")
	om.isRunning = false

	// Cancel context
	om.cancel()

	om.logger.Info("OrderMonitor stopped successfully")
	return nil
}

// TrackOrder starts tracking an order
func (om *Monitor) TrackOrder(ctx context.Context, orderID *big.Int) error {
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

	// Create event before unlocking
	event := &types.OrderEvent{
		Type:      types.OrderEventNew,
		OrderID:   orderID,
		Order:     order,
		Timestamp: time.Now(),
	}

	// Unlock before emitting event to avoid deadlock
	om.mu.Unlock()

	// Emit new order event (no lock needed)
	om.emitEvent(event)

	// Re-lock for defer
	om.mu.Lock()

	return nil
}

// UntrackOrder stops tracking an order
func (om *Monitor) UntrackOrder(ctx context.Context, orderID *big.Int) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	orderIDStr := orderID.String()

	if _, exists := om.trackedOrders[orderIDStr]; !exists {
		return fmt.Errorf("order %s is not being tracked", orderIDStr)
	}

	// Delete order from storage
	if err := om.storage.DeleteOrder(ctx, orderIDStr); err != nil {
		om.logger.Warn("Failed to delete order from storage", "orderID", orderID, "error", err)
	}

	delete(om.trackedOrders, orderIDStr)
	om.logger.Info("Stopped tracking order", "orderID", orderID)
	return nil
}

// GetTrackedOrders returns all currently tracked orders
func (om *Monitor) GetTrackedOrders(ctx context.Context) ([]*big.Int, error) {
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
func (om *Monitor) MonitorOrderStatus(ctx context.Context, orderID *big.Int) error {
	om.mu.RLock()
	orderIDStr := orderID.String()
	oldOrder, exists := om.trackedOrders[orderIDStr]
	om.mu.RUnlock()

	if !exists {
		return fmt.Errorf("order %s is not being tracked", orderIDStr)
	}

	// Get current order status from blockchain
	newOrder, err := om.bidMarket.GetOrder(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order status: %w", err)
	}

	// Check if order status has changed
	if oldOrder.Status != newOrder.Status {
		om.handleOrderStatusChange(orderID, oldOrder, newOrder)
	}

	// Update tracked order
	om.mu.Lock()
	om.trackedOrders[orderIDStr] = newOrder
	om.mu.Unlock()

	// Update order in storage
	if err := om.storage.UpdateOrder(ctx, newOrder); err != nil {
		om.logger.Warn("Failed to update order in storage", "error", err)
	}

	// Emit updated event
	om.emitEvent(&types.OrderEvent{
		Type:      types.OrderEventUpdated,
		OrderID:   orderID,
		Order:     newOrder,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"oldStatus": oldOrder.Status,
			"newStatus": newOrder.Status,
		},
	})

	return nil
}

// handleOrderStatusChange handles order status changes and emits appropriate events
func (om *Monitor) handleOrderStatusChange(orderID *big.Int, oldOrder, newOrder *types.Order) {
	om.logger.Info("Order status changed",
		"orderID", orderID.String(),
		"oldStatus", oldOrder.Status,
		"newStatus", newOrder.Status)

	switch newOrder.Status {
	case types.OrderStatusClosed:
		om.emitEvent(&types.OrderEvent{
			Type:      types.OrderEventClosed,
			OrderID:   orderID,
			Order:     newOrder,
			Timestamp: time.Now(),
		})
		// Stop tracking closed orders
		go func() {
			if err := om.UntrackOrder(om.ctx, orderID); err != nil {
				om.logger.Error("Failed to untrack closed order", "orderID", orderID.String(), "error", err)
			}
		}()

	case types.OrderStatusExpired:
		om.emitEvent(&types.OrderEvent{
			Type:      types.OrderEventExpired,
			OrderID:   orderID,
			Order:     newOrder,
			Timestamp: time.Now(),
		})
		// Stop tracking expired orders
		go func() {
			if err := om.UntrackOrder(om.ctx, orderID); err != nil {
				om.logger.Error("Failed to untrack expired order", "orderID", orderID.String(), "error", err)
			}
		}()

	case types.OrderStatusCancelled:
		om.emitEvent(&types.OrderEvent{
			Type:      types.OrderEventClosed,
			OrderID:   orderID,
			Order:     newOrder,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"reason": "cancelled",
			},
		})
		// Stop tracking cancelled orders
		go func() {
			if err := om.UntrackOrder(om.ctx, orderID); err != nil {
				om.logger.Error("Failed to untrack cancelled order", "orderID", orderID.String(), "error", err)
			}
		}()
	case types.OrderStatusAccepted:
		om.emitEvent(&types.OrderEvent{
			Type:      types.OrderEventAccepted,
			OrderID:   orderID,
			Order:     newOrder,
			Timestamp: time.Now(),
		})
	}
}

// CheckOrderExpiry checks if an order has expired
func (om *Monitor) CheckOrderExpiry(ctx context.Context, orderID *big.Int) error {
	om.mu.RLock()
	orderIDStr := orderID.String()
	order, exists := om.trackedOrders[orderIDStr]
	om.mu.RUnlock()

	if !exists {
		return fmt.Errorf("order %s is not being tracked", orderIDStr)
	}

	now := time.Now().Unix()

	// Check if order is still open but has exceeded bidding time (5 minutes)
	if order.Status == types.OrderStatusOpen {
		const biddingTimeLimit = int64(300) // 5 minutes in seconds
		timeSinceCreation := now - order.CreatedAt.Int64()

		if timeSinceCreation > biddingTimeLimit {
			om.logger.Info("Order has exceeded bidding time limit, treating as cancelled",
				"orderID", orderID,
				"createdAt", order.CreatedAt.Int64(),
				"currentTime", now,
				"timeSinceCreation", timeSinceCreation,
				"biddingTimeLimit", biddingTimeLimit)

			// Update order status to cancelled
			order.Status = types.OrderStatusCancelled

			// Update in storage
			if err := om.storage.UpdateOrder(ctx, order); err != nil {
				om.logger.Warn("Failed to update cancelled order in storage", "error", err)
			}

			// Emit cancelled event
			om.emitEvent(&types.OrderEvent{
				Type:      types.OrderEventClosed,
				OrderID:   orderID,
				Order:     order,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"reason": "bidding_time_expired",
				},
			})

			// Stop tracking cancelled order
			go func() {
				if err := om.UntrackOrder(om.ctx, orderID); err != nil {
					om.logger.Error("Failed to untrack cancelled order", "orderID", orderID.String(), "error", err)
				}
			}()

			return nil
		}
	}

	// Check if order has expired and grace period passed
	if types.IsOrderReadyToClose(order, now) {
		om.logger.Info("Order has expired and grace period passed",
			"orderID", orderID,
			"expiredAt", order.ExpiredAt,
			"currentTime", now)

		// Update order status
		order.Status = types.OrderStatusExpired

		// Update in storage
		if err := om.storage.UpdateOrder(ctx, order); err != nil {
			om.logger.Warn("Failed to update expired order in storage", "error", err)
		}

		// Emit expired event
		om.emitEvent(&types.OrderEvent{
			Type:      types.OrderEventExpired,
			OrderID:   orderID,
			Order:     order,
			Timestamp: time.Now(),
		})

		// Stop tracking expired order
		go func() {
			if err := om.UntrackOrder(om.ctx, orderID); err != nil {
				om.logger.Error("Failed to untrack expired order", "orderID", orderID.String(), "error", err)
			}
		}()
	} else {
		om.logger.Debug("Order has expired but grace period not passed yet",
			"orderID", orderID,
			"expiredAt", order.ExpiredAt,
			"currentTime", now)
	}

	return nil
}

// orderStatusSyncLoop syncs status of tracked orders and checks for expiry
func (om *Monitor) orderStatusSyncLoop(ctx context.Context) {
	// More frequent sync if event watching is disabled
	interval := 30 * time.Second

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			om.syncTrackedOrders(ctx)
			om.checkOrderExpiries(ctx)
		}
	}
}

// syncTrackedOrders syncs all tracked orders with blockchain
func (om *Monitor) syncTrackedOrders(ctx context.Context) {
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

// checkOrderExpiries checks all tracked orders for expiry
func (om *Monitor) checkOrderExpiries(ctx context.Context) {
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
func (om *Monitor) eventProcessingLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-om.orderEvents:
			om.logger.Debug("Processing order event",
				"type", event.Type,
				"orderID", event.OrderID.String())
		}
	}
}

// loadPersistedOrders loads orders from storage on startup
func (om *Monitor) loadPersistedOrders(ctx context.Context) error {
	om.logger.Info("Loading persisted orders from storage")

	orders, err := om.storage.ListOrders(ctx)
	if err != nil {
		return fmt.Errorf("failed to load orders: %w", err)
	}

	om.mu.Lock()
	for _, order := range orders {
		om.trackedOrders[order.ID.String()] = order
	}
	om.mu.Unlock()

	om.logger.Info("Loaded persisted orders", "trackedOrders", len(om.trackedOrders))
	return nil
}

// GetStats returns order monitor statistics
func (om *Monitor) GetStats() map[string]interface{} {
	om.mu.RLock()
	defer om.mu.RUnlock()

	return map[string]interface{}{
		"trackedOrders": len(om.trackedOrders),
		"isRunning":     om.isRunning,
	}
}
