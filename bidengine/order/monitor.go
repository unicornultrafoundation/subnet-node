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

	// Last order ID tracking for polling
	lastOrderID *big.Int
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
		lastOrderID:       nil,
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
					om.logger.WithFields(logrus.Fields{
						"eventType": event.Type,
						"orderID":   event.OrderID.String(),
						"panic":     r,
					}).Error("Event handler panicked")
				}
			}()
			h(event)
		}(handler)
	}

	// Also send to event channel for external consumers
	select {
	case om.orderEvents <- event:
	default:
		om.logger.WithFields(logrus.Fields{
			"eventType": event.Type,
			"orderID":   event.OrderID.String(),
		}).Warn("Order event channel is full, dropping event")
	}
}

// Start starts the order monitor
func (om *Monitor) Start(ctx context.Context) error {

	if om.isRunning {
		return nil
	}

	om.logger.Info("Starting OrderMonitor")
	om.isRunning = true

	// Load persisted orders from storage
	if err := om.loadPersistedOrders(ctx); err != nil {
		om.logger.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Failed to load persisted orders")
	}

	// Load last order ID from storage
	if err := om.loadLastOrderID(ctx); err != nil {
		om.logger.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Failed to load last order ID")
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

// watchOrderCreatedEvents polls for new order creation events
func (om *Monitor) watchOrderCreatedEvents(ctx context.Context) {
	om.logger.Info("Starting to poll for new orders")

	// Poll every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			om.logger.Info("Order polling stopped")
			return
		case <-ticker.C:
			if err := om.pollForNewOrders(ctx); err != nil {
				om.logger.WithFields(logrus.Fields{
					"error": err,
				}).Warn("Failed to poll for new orders")
			}
		}
	}
}

// isOrderWithinBiddingTime checks if an order is still within the bidding time limit (5 minutes from creation)
func (om *Monitor) isOrderWithinBiddingTime(order *types.Order) bool {
	return order.IsOrderWithinBiddingTime()
}

// pollForNewOrders polls for new orders since the last known order ID
func (om *Monitor) pollForNewOrders(ctx context.Context) error {
	om.mu.RLock()
	lastOrderID := om.lastOrderID
	om.mu.RUnlock()

	// Get current order count from blockchain
	orderCount, err := om.bidMarket.GetOrderCount(ctx)
	if err != nil {
		return fmt.Errorf("failed to get order count: %w", err)
	}

	// If no last order ID, start from the latest order
	if lastOrderID == nil {
		if orderCount.Cmp(big.NewInt(0)) > 0 {
			// Start from the latest order (orderCount - 1)
			lastOrderID = new(big.Int).Sub(orderCount, big.NewInt(1))
			om.logger.WithFields(logrus.Fields{
				"lastOrderID": lastOrderID.String(),
			}).Info("No last order ID found, starting from latest order")
		} else {
			om.logger.Debug("No orders exist yet")
			return nil
		}
	}

	// Check for new orders
	currentOrderID := new(big.Int).Add(lastOrderID, big.NewInt(1))

	om.logger.WithFields(logrus.Fields{
		"currentOrderID": currentOrderID.String(),
		"orderCount":     orderCount.String(),
	}).Info("Checking for new orders")

	// If lastOrderID equals orderCount, no new orders
	if lastOrderID.Cmp(orderCount) == 0 {
		om.logger.Debug("No new orders found")
		return nil
	}

	// Process orders from currentOrderID to orderCount-1
	for currentOrderID.Cmp(orderCount) <= 0 {
		om.logger.WithFields(logrus.Fields{
			"orderID": currentOrderID.String(),
		}).Debug("Processing order")

		// Get order details
		order, err := om.bidMarket.GetOrder(ctx, currentOrderID)
		if err != nil {
			om.logger.WithFields(logrus.Fields{
				"orderID": currentOrderID.String(),
				"error":   err,
			}).Error("Failed to get order details")
			currentOrderID.Add(currentOrderID, big.NewInt(1))
			continue
		}

		// Only track open orders that are still within bidding time (5 minutes from creation)
		if order.Status == types.OrderStatusOpen {
			// Check if order is still within bidding time
			if om.isOrderWithinBiddingTime(order) {
				// Check if we're already tracking this order
				om.mu.RLock()
				_, alreadyTracked := om.trackedOrders[currentOrderID.String()]
				om.mu.RUnlock()

				if !alreadyTracked {
					om.logger.WithFields(logrus.Fields{
						"orderID": currentOrderID.String(),
					}).Info("New order found via polling and still within bidding time")

					// Track the new order
					if err := om.TrackOrder(ctx, currentOrderID); err != nil {
						om.logger.WithFields(logrus.Fields{
							"orderID": currentOrderID.String(),
							"error":   err,
						}).Warn("Failed to track new order from polling")
					}
				}
			} else {
				om.logger.WithFields(logrus.Fields{
					"orderID": currentOrderID.String(),
				}).Debug("Order found but bidding time has expired, skipping")
			}
		}

		currentOrderID.Add(currentOrderID, big.NewInt(1))
	}

	// Update last order ID to the last processed order (orderCount - 1)
	if orderCount.Cmp(big.NewInt(0)) > 0 {

		om.mu.Lock()
		om.lastOrderID = orderCount
		om.mu.Unlock()

		// Save to storage
		if err := om.storage.SaveLastOrderID(ctx, orderCount); err != nil {
			om.logger.WithFields(logrus.Fields{
				"lastOrderID": orderCount.String(),
				"error":       err,
			}).Warn("Failed to save last order ID")
		}

		om.logger.WithFields(logrus.Fields{
			"lastOrderID": orderCount.String(),
		}).Debug("Updated last order ID")
	}

	return nil
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
	om.logger.WithFields(logrus.Fields{
		"type":    event.Type,
		"orderID": event.OrderID.String(),
	}).Info("Received contract event")

	switch event.Type {
	case "OrderCreated":
		om.handleOrderCreatedEvent(ctx, event)
	default:
		om.logger.WithFields(logrus.Fields{
			"type": event.Type,
		}).Debug("Ignoring contract event type")
	}
}

// handleOrderCreatedEvent handles OrderCreated events from contract
func (om *Monitor) handleOrderCreatedEvent(ctx context.Context, event *types.OrderEvent) {
	om.logger.WithFields(logrus.Fields{
		"orderID": event.OrderID.String(),
	}).Info("Handling OrderCreated event")

	// Get the full order details from contract
	order, err := om.bidMarket.GetOrder(ctx, event.OrderID)
	if err != nil {
		om.logger.WithFields(logrus.Fields{
			"orderID": event.OrderID.String(),
			"error":   err,
		}).Warn("Failed to get order details for created event")
		return
	}

	// Only track open orders that are still within bidding time (5 minutes from creation)
	if order.Status == types.OrderStatusOpen {
		// Check if order is still within bidding time
		if om.isOrderWithinBiddingTime(order) {
			// Check if we're already tracking this order
			om.mu.RLock()
			_, alreadyTracked := om.trackedOrders[event.OrderID.String()]
			om.mu.RUnlock()

			if !alreadyTracked {
				om.logger.WithFields(logrus.Fields{
					"orderID": event.OrderID.String(),
				}).Info("New order created via event and still within bidding time")

				// Track the new order
				if err := om.TrackOrder(ctx, event.OrderID); err != nil {
					om.logger.WithFields(logrus.Fields{
						"orderID": event.OrderID.String(),
						"error":   err,
					}).Warn("Failed to track new order from event")
				}
			}
		} else {
			om.logger.WithFields(logrus.Fields{
				"orderID": event.OrderID.String(),
			}).Debug("Order created via event but bidding time has expired, skipping")
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
		om.logger.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Failed to save order to storage")
	}

	om.logger.WithFields(logrus.Fields{
		"orderID": orderID.String(),
	}).Info("Started tracking order")
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
		om.logger.WithFields(logrus.Fields{
			"orderID": orderID.String(),
			"error":   err,
		}).Warn("Failed to delete order from storage")
	}

	delete(om.trackedOrders, orderIDStr)
	om.logger.WithFields(logrus.Fields{
		"orderID": orderID.String(),
	}).Info("Stopped tracking order")
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

	// Check if order is still within bidding time
	if newOrder.Status == types.OrderStatusOpen && !om.isOrderWithinBiddingTime(newOrder) {
		newOrder.Status = types.OrderStatusClosed
	} else if newOrder.Status == types.OrderStatusAccepted {
		// Check if order is ready to close
		if types.IsOrderReadyToClose(newOrder, time.Now().Unix()) {
			newOrder.Status = types.OrderStatusExpired
		}
	}

	// Update tracked order
	om.mu.Lock()
	om.trackedOrders[orderIDStr] = newOrder
	om.mu.Unlock()

	// Update order in storage
	if err := om.storage.UpdateOrder(ctx, newOrder); err != nil {
		om.logger.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Failed to update order in storage")
	}
	// Check if order status has changed
	if oldOrder.Status != newOrder.Status {
		om.handleOrderStatusChange(orderID, oldOrder, newOrder)
	}
	return nil
}

// handleOrderStatusChange handles order status changes and emits appropriate events
func (om *Monitor) handleOrderStatusChange(orderID *big.Int, oldOrder, newOrder *types.Order) {
	om.logger.WithFields(logrus.Fields{
		"orderID":   orderID.String(),
		"oldStatus": oldOrder.Status,
		"newStatus": newOrder.Status,
	}).Info("Order status changed")

	switch newOrder.Status {
	case types.OrderStatusClosed:
		om.emitEvent(&types.OrderEvent{
			Type:      types.OrderEventClosed,
			OrderID:   orderID,
			Order:     newOrder,
			Timestamp: time.Now(),
		})
		// Stop tracking closed orders
		if err := om.UntrackOrder(om.ctx, orderID); err != nil {
			om.logger.WithFields(logrus.Fields{
				"orderID": orderID.String(),
				"error":   err,
			}).Error("Failed to untrack closed order")
		}

	case types.OrderStatusExpired:
		om.emitEvent(&types.OrderEvent{
			Type:      types.OrderEventExpired,
			OrderID:   orderID,
			Order:     newOrder,
			Timestamp: time.Now(),
		})
		// Stop tracking expired orders
		if err := om.UntrackOrder(om.ctx, orderID); err != nil {
			om.logger.WithFields(logrus.Fields{
				"orderID": orderID.String(),
				"error":   err,
			}).Error("Failed to untrack expired order")
		}

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
		if err := om.UntrackOrder(om.ctx, orderID); err != nil {
			om.logger.WithFields(logrus.Fields{
				"orderID": orderID.String(),
				"error":   err,
			}).Error("Failed to untrack cancelled order")
		}
	case types.OrderStatusAccepted:
		om.emitEvent(&types.OrderEvent{
			Type:      types.OrderEventAccepted,
			OrderID:   orderID,
			Order:     newOrder,
			Timestamp: time.Now(),
		})
	}
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
			om.logger.WithFields(logrus.Fields{
				"orderID": orderID.String(),
				"error":   err,
			}).Warn("Failed to sync order status")
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
			om.logger.WithFields(logrus.Fields{
				"type":    event.Type,
				"orderID": event.OrderID.String(),
			}).Debug("Processing order event")
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

	om.logger.WithFields(logrus.Fields{
		"trackedOrders": len(om.trackedOrders),
	}).Info("Loaded persisted orders")
	return nil
}

// loadLastOrderID loads the last processed order ID from storage
func (om *Monitor) loadLastOrderID(ctx context.Context) error {
	om.logger.Info("Loading last order ID from storage")

	lastOrderID, err := om.storage.GetLastOrderID(ctx)
	if err != nil {
		return fmt.Errorf("failed to load last order ID: %w", err)
	}

	om.mu.Lock()
	om.lastOrderID = lastOrderID
	om.mu.Unlock()

	if lastOrderID != nil {
		om.logger.WithFields(logrus.Fields{
			"lastOrderID": lastOrderID.String(),
		}).Info("Loaded last order ID from storage")
	} else {
		om.logger.Info("No last order ID found in storage, will start from latest")
	}

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
