package manager

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	ds "github.com/ipfs/go-datastore"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/bidengine/storage"
	"github.com/unicornultrafoundation/subnet-node/bidengine/types"
)

// ExtendedBid represents a bid with additional fields for internal use
type ExtendedBid struct {
	*types.Bid
	OrderID         string
	TransactionHash string
	SubmittedAt     time.Time
	CancelledAt     time.Time
}

// Manager manages bid submission and tracking
type Manager struct {
	config          *types.BidEngineConfig
	bidMarket       types.BidMarketContract
	logger          *logrus.Entry
	metrics         types.Metrics
	storage         types.Storage
	resourceManager types.ResourceManager
	pricingService  types.PricingEngine
	orderMonitor    types.OrderMonitor

	mu        sync.RWMutex
	isRunning bool
	ctx       context.Context
	cancel    context.CancelFunc

	// Tracking
	pendingBids map[string]*ExtendedBid
	bidResults  map[string]*types.BidResult
}

// NewBidManager creates a new BidManagerService instance
func NewManager(
	config *types.BidEngineConfig,
	bidMarket types.BidMarketContract,
	logger *logrus.Entry,
	metrics types.Metrics,
	datastore ds.Datastore,
	resourceManager types.ResourceManager,
	pricingService types.PricingEngine,
	orderMonitor types.OrderMonitor,
) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		config:          config,
		bidMarket:       bidMarket,
		logger:          logger,
		metrics:         metrics,
		storage:         storage.NewStorage(datastore, logger),
		resourceManager: resourceManager,
		pricingService:  pricingService,
		orderMonitor:    orderMonitor,
		ctx:             ctx,
		cancel:          cancel,
		pendingBids:     make(map[string]*ExtendedBid),
		bidResults:      make(map[string]*types.BidResult),
	}
}

// Start starts the bid manager
func (bm *Manager) Start(ctx context.Context) error {
	bm.mu.Lock()
	if bm.isRunning {
		bm.mu.Unlock()
		return nil
	}

	bm.logger.Info("Starting BidManager")
	bm.isRunning = true
	bm.mu.Unlock()

	// Load persisted bids from storage (without lock)
	if err := bm.loadPersistedBids(ctx); err != nil {
		bm.logger.Warn("Failed to load persisted bids", "error", err)
	}

	bm.orderMonitor.RegisterEventHandler(types.OrderEventNew, bm.handleOrderCreate)
	bm.orderMonitor.RegisterEventHandler(types.OrderEventClosed, bm.handleOrderClosed)
	bm.orderMonitor.RegisterEventHandler(types.OrderEventExpired, bm.handleOrderExpired)
	bm.orderMonitor.RegisterEventHandler(types.OrderEventAccepted, bm.handleOrderAccepted)

	bm.logger.Info("BidManager started successfully")
	return nil
}

// Stop stops the bid manager
func (bm *Manager) Stop(ctx context.Context) error {
	bm.mu.Lock()
	if !bm.isRunning {
		bm.mu.Unlock()
		return nil
	}

	bm.logger.Info("Stopping BidManager")
	bm.isRunning = false
	bm.mu.Unlock()

	// Cancel context
	bm.cancel()

	bm.logger.Info("BidManager stopped successfully")
	return nil
}

// handleOrderCreate handles new order events and attempts to place bids on them
// This function is registered as an event handler for OrderEventNew events.
// When a new order is created, it automatically tries to submit a bid if:
// - The order is open for bidding
// - A suitable machine is available
// - The calculated price is within the order's price limits
// - No existing bid has been placed for this order
func (bm *Manager) handleOrderCreate(event *types.OrderEvent) {
	// Extract order from event
	order := event.Order
	if order == nil {
		bm.logger.Warn("Received order event without order data", "orderID", event.OrderID)
		return
	}

	bm.logger.Info("Received new order event",
		"orderID", order.ID.String(),
		"status", order.Status,
		"machineType", order.MachineType.String())

	// Try to bid on the new order
	if err := bm.TryBidOnOrder(bm.ctx, order); err != nil {
		bm.logger.Warn("Failed to bid on new order",
			"orderID", order.ID.String(),
			"error", err)
	}
}

// SubmitBid submits a bid to the blockchain
func (bm *Manager) SubmitBid(ctx context.Context, orderID *big.Int, pricePerSecond *big.Int, machineID *big.Int) (*types.BidResult, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Validate inputs
	if orderID == nil {
		return nil, fmt.Errorf("order ID cannot be nil")
	}

	if machineID == nil {
		return nil, fmt.Errorf("machine ID cannot be nil")
	}

	if pricePerSecond == nil || pricePerSecond.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("price must be greater than zero")
	}

	// Use provider ID from configuration
	providerID := bm.config.ProviderID
	if providerID == nil {
		return nil, fmt.Errorf("provider ID not configured")
	}

	// Submit bid to blockchain
	tx, err := bm.bidMarket.SubmitBid(ctx, orderID, pricePerSecond, providerID, machineID)
	if err != nil {
		bm.logger.Error("Failed to submit bid to blockchain",
			"orderID", orderID.String(),
			"pricePerSecond", pricePerSecond.String(),
			"providerID", providerID.String(),
			"machineID", machineID.String(),
			"error", err)
		return nil, fmt.Errorf("failed to submit bid: %w", err)
	}

	// Wait for transaction confirmation and get bid index
	bidIndex, err := bm.bidMarket.GetBidIndexFromTransaction(ctx, tx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid index from transaction: %w", err)
	}

	// Allocate resources immediately after successful bid submission
	if err := bm.allocateResourcesForBid(ctx, orderID, machineID); err != nil {
		bm.logger.Warn("Failed to allocate resources for bid",
			"orderID", orderID.String(),
			"error", err)
		// Don't fail the bid submission if resource allocation fails
	}

	// Create extended bid
	extendedBid := &ExtendedBid{
		Bid: &types.Bid{
			Id:             bidIndex,
			Provider:       bm.config.ProviderWallet,
			PricePerSecond: pricePerSecond,
			Status:         types.BidStatusActive,
			CreatedAt:      big.NewInt(time.Now().Unix()),
			ProviderId:     providerID,
			MachineId:      machineID,
		},
		OrderID:         orderID.String(),
		TransactionHash: tx.Hash().String(),
		SubmittedAt:     time.Now(),
	}

	// Track bid using orderID only (one bid per provider per order)
	bidKey := orderID.String()
	bm.pendingBids[bidKey] = extendedBid

	// Save bid to storage with actual bidIndex
	if err := bm.storage.SaveBid(ctx, extendedBid.Bid, orderID.String(), int(bidIndex.Int64())); err != nil {
		bm.logger.Warn("Failed to save bid to storage", "error", err)
	}

	// Create bid result
	bidResult := &types.BidResult{
		OrderID:   orderID,
		BidIndex:  bidIndex,
		Success:   true,
		TxHash:    tx.Hash(),
		Timestamp: time.Now(),
	}

	// Store bid result
	bm.bidResults[orderID.String()] = bidResult

	// Track the bid for monitoring
	if err := bm.trackBidInternal(ctx, orderID, bidIndex); err != nil {
		bm.logger.Warn("Failed to track bid", "error", err)
	}

	bm.logger.Info("Bid submitted successfully",
		"orderID", orderID.String(),
		"bidIndex", bidIndex.String(),
		"pricePerSecond", pricePerSecond.String(),
		"machineID", machineID.String(),
		"txHash", tx.Hash().String())

	bm.metrics.IncrementBidsSubmitted()
	return bidResult, nil
}

// CancelBid cancels a pending bid
func (bm *Manager) CancelBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Cancel bid on blockchain
	tx, err := bm.bidMarket.CancelBid(ctx, orderID, bidIndex)
	if err != nil {
		bm.logger.Error("Failed to cancel bid on blockchain",
			"orderID", orderID.String(),
			"bidIndex", bidIndex.String(),
			"error", err)
		return fmt.Errorf("failed to cancel bid on blockchain: %w", err)
	}

	// Update bid status using orderID only
	bidKey := orderID.String()
	if bid, exists := bm.pendingBids[bidKey]; exists {
		bid.Bid.Status = types.BidStatusCancelled
		bid.CancelledAt = time.Now()

		// Update bid in storage with correct bidIndex
		if err := bm.storage.UpdateBid(ctx, bid.Bid, bid.OrderID, int(bidIndex.Int64())); err != nil {
			bm.logger.Warn("Failed to update bid in storage", "error", err)
		}

		// Cleanup the cancelled bid
		bm.mu.Unlock()
		bm.cleanupBid(orderID, "bid_cancelled_by_provider")
		// Re-lock to continue with cleanup
		bm.mu.Lock()
	}

	bm.logger.Info("Bid cancelled successfully",
		"orderID", orderID.String(),
		"bidIndex", bidIndex.String(),
		"txHash", tx.Hash().String())

	return nil
}

// TrackBid tracks a bid
func (bm *Manager) TrackBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	return bm.trackBidInternal(ctx, orderID, bidIndex)
}

// trackBidInternal is an internal method for tracking a bid
func (bm *Manager) trackBidInternal(ctx context.Context, orderID *big.Int, bidIndex *big.Int) error {
	// Track bid using orderID only
	bidKey := orderID.String()
	if _, exists := bm.pendingBids[bidKey]; exists {
		return fmt.Errorf("bid already tracked for order %s", orderID.String())
	}

	// Create extended bid for tracking
	extendedBid := &ExtendedBid{
		Bid: &types.Bid{
			Id: bidIndex,
		},
		OrderID:     orderID.String(),
		SubmittedAt: time.Now(),
	}

	bm.pendingBids[bidKey] = extendedBid
	bm.logger.Info("Bid tracked", "orderID", orderID.String(), "bidIndex", bidIndex.String())

	return nil
}

// UntrackBid untracks a bid
func (bm *Manager) UntrackBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	return bm.untrackBidInternal(ctx, orderID, bidIndex)
}

// untrackBidInternal is an internal method for untracking a bid
func (bm *Manager) untrackBidInternal(ctx context.Context, orderID *big.Int, bidIndex *big.Int) error {
	// Untrack bid using orderID only
	bidKey := orderID.String()
	if _, exists := bm.pendingBids[bidKey]; !exists {
		return fmt.Errorf("bid not tracked for order %s", orderID.String())
	}

	delete(bm.pendingBids, bidKey)
	bm.logger.Info("Bid untracked", "orderID", orderID.String(), "bidIndex", bidIndex.String())

	return nil
}

// GetTrackedBids returns all tracked bids
func (bm *Manager) GetTrackedBids(ctx context.Context) (map[*big.Int][]*big.Int, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	result := make(map[*big.Int][]*big.Int)

	for bidKey, extendedBid := range bm.pendingBids {
		// Parse orderID from bidKey (which is now just orderID)
		orderID, ok := new(big.Int).SetString(bidKey, 10)
		if !ok {
			bm.logger.Warn("Invalid orderID in bidKey", "bidKey", bidKey)
			continue
		}

		// Add bidIndex to the list for this orderID
		if result[orderID] == nil {
			result[orderID] = make([]*big.Int, 0)
		}
		result[orderID] = append(result[orderID], extendedBid.Bid.Id)
	}

	return result, nil
}

// GetBid retrieves a bid by order ID and bid index
func (bm *Manager) GetBid(ctx context.Context, orderID string, bidIndex int) (*types.Bid, error) {
	// Try to get from storage first
	bids, err := bm.storage.GetBids(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bids from storage: %w", err)
	}

	if bidIndex < len(bids) {
		return bids[bidIndex], nil
	}

	return nil, fmt.Errorf("bid not found: orderID=%s, bidIndex=%d", orderID, bidIndex)
}

// GetBids retrieves all bids for an order
func (bm *Manager) GetBids(ctx context.Context, orderID string) ([]*types.Bid, error) {
	return bm.storage.GetBids(ctx, orderID)
}

// GetPendingBids returns all pending bids
func (bm *Manager) GetPendingBids(ctx context.Context) []*types.Bid {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	var pendingBids []*types.Bid
	for _, extendedBid := range bm.pendingBids {
		if extendedBid.Bid.Status == types.BidStatusActive {
			pendingBids = append(pendingBids, extendedBid.Bid)
		}
	}

	return pendingBids
}

// cleanupBid properly cleans up a bid by removing it from tracking and deallocating resources
// This function should be called when a bid is no longer needed (rejected, cancelled, expired, etc.)
func (bm *Manager) cleanupBid(orderID *big.Int, reason string) {
	bidKey := orderID.String()

	bm.mu.Lock()
	bid, exists := bm.pendingBids[bidKey]
	bm.mu.Unlock()

	if !exists {
		bm.logger.Debug("No pending bid found for cleanup", "orderID", orderID.String(), "reason", reason)
		return
	}

	bm.logger.Info("Cleaning up bid",
		"orderID", orderID.String(),
		"bidIndex", bid.Bid.Id.String(),
		"reason", reason)

	// Remove from pending bids
	bm.mu.Lock()
	delete(bm.pendingBids, bidKey)
	bm.mu.Unlock()

	// Deallocate resources
	if err := bm.deallocateResourcesForBid(bm.ctx, orderID); err != nil {
		bm.logger.Error("Failed to deallocate resources during bid cleanup",
			"orderID", orderID.String(),
			"reason", reason,
			"error", err)
	}

	// Update bid status in storage
	bid.Bid.Status = types.BidStatusCancelled
	if err := bm.storage.UpdateBid(bm.ctx, bid.Bid, bid.OrderID, int(bid.Bid.Id.Int64())); err != nil {
		bm.logger.Warn("Failed to update bid status in storage during cleanup",
			"orderID", orderID.String(),
			"error", err)
	}

	// Record metrics for rejected bid
	bm.metrics.IncrementBidsRejected()

	bm.logger.Info("Bid cleanup completed",
		"orderID", orderID.String(),
		"bidIndex", bid.Bid.Id.String(),
		"reason", reason)
}

// handleOrderClosed handles order closed events
// This function is registered as an event handler for OrderEventClosed events.
// When an order is closed, it removes the order from pending bids tracking.
func (bm *Manager) handleOrderClosed(event *types.OrderEvent) {
	// Extract order from event
	order := event.Order
	if order == nil {
		bm.logger.Warn("Received order event without order data", "orderID", event.OrderID)
		return
	}

	bm.logger.Info("Received order closed event", "orderID", order.ID.String())

	// Cleanup bid if exists
	bm.cleanupBid(order.ID, "order_closed")
}

// handleOrderExpired handles order expired events
// This function is registered as an event handler for OrderEventExpired events.
// When an order expires, it removes the order from pending bids tracking.
func (bm *Manager) handleOrderExpired(event *types.OrderEvent) {
	// Extract order from event
	order := event.Order
	if order == nil {
		bm.logger.Warn("Received order event without order data", "orderID", event.OrderID)
		return
	}

	bm.logger.Info("Received order expired event", "orderID", order.ID.String())

	// Cleanup bid if exists
	bm.cleanupBid(order.ID, "order_expired")
}

// handleOrderAccepted handles order accepted events
// This function is registered as an event handler for OrderEventAccepted events.
// When an order is accepted, it checks if our bid was accepted or another provider's bid.
// If our bid was accepted, it updates the bid status and records metrics.
// If another provider's bid was accepted, it untracks the order since it's no longer relevant.
func (bm *Manager) handleOrderAccepted(event *types.OrderEvent) {
	// Extract order from event
	order := event.Order
	if order == nil {
		bm.logger.Warn("Received order event without order data", "orderID", event.OrderID)
		return
	}

	bm.logger.Info("Received order accepted event", "orderID", order.ID.String())

	// Check if we have a pending bid for this order
	bm.mu.Lock()
	bidKey := order.ID.String()
	if bid, exists := bm.pendingBids[bidKey]; exists {
		// Check if our bid was accepted by comparing provider IDs
		if order.AcceptedProviderId != nil && bid.Bid.ProviderId != nil &&
			order.AcceptedProviderId.Cmp(bid.Bid.ProviderId) == 0 {
			// Our bid was accepted
			oldBidStatus := bid.Bid.Status
			bid.Bid.Status = types.BidStatusAccepted

			// Update bid in storage
			if err := bm.storage.UpdateBid(bm.ctx, bid.Bid, bid.OrderID, int(bid.Bid.Id.Int64())); err != nil {
				bm.logger.Warn("Failed to update bid in storage", "error", err)
			}

			// Record metrics for accepted bid
			if oldBidStatus != types.BidStatusAccepted {
				bm.metrics.IncrementBidsAccepted()
			}

			bm.logger.Info("Our bid was accepted",
				"orderID", order.ID.String(),
				"bidIndex", bid.Bid.Id.String(),
				"providerID", bid.Bid.ProviderId.String(),
				"machineID", bid.Bid.MachineId.String())
		} else {
			// Another provider's bid was accepted, not ours
			bm.logger.Info("Another provider's bid was accepted, untracking order",
				"orderID", order.ID.String(),
				"acceptedProviderID", order.AcceptedProviderId,
				"ourProviderID", bid.Bid.ProviderId)

			// Cleanup our rejected bid
			bm.mu.Unlock()
			bm.cleanupBid(order.ID, "bid_rejected_by_another_provider")

			// Untrack the order since it's no longer relevant to us
			if err := bm.orderMonitor.UntrackOrder(bm.ctx, order.ID); err != nil {
				bm.logger.Warn("Failed to untrack accepted order", "orderID", order.ID.String(), "error", err)
			}
			return
		}
	} else {
		// No pending bid for this order, it was accepted by another provider
		bm.logger.Info("Order accepted by another provider (no pending bid), untracking order", "orderID", order.ID.String())

		// Untrack the order since it's no longer relevant to us
		bm.mu.Unlock()
		if err := bm.orderMonitor.UntrackOrder(bm.ctx, order.ID); err != nil {
			bm.logger.Warn("Failed to untrack accepted order", "orderID", order.ID.String(), "error", err)
		}
		return
	}
	bm.mu.Unlock()
}

// deallocateResourcesForBid deallocates resources for a bid
func (bm *Manager) deallocateResourcesForBid(ctx context.Context, orderID *big.Int) error {
	// Stop the resource first
	if err := bm.resourceManager.StopResource(ctx, orderID); err != nil {
		bm.logger.Warn("Failed to stop resource",
			"orderID", orderID.String(),
			"error", err)
	}

	// Deallocate resources
	if err := bm.resourceManager.DeallocateResources(ctx, orderID); err != nil {
		return fmt.Errorf("failed to deallocate resources: %w", err)
	}

	bm.logger.Info("Resources deallocated for bid",
		"orderID", orderID.String())

	return nil
}

// loadPersistedBids loads bids from storage on startup
func (bm *Manager) loadPersistedBids(ctx context.Context) error {
	bm.logger.Info("Loading persisted bids from storage")

	// Get all orders to load their bids
	orders, err := bm.storage.ListOrders(ctx)
	if err != nil {
		return fmt.Errorf("failed to load orders: %w", err)
	}

	bm.mu.Lock()
	defer bm.mu.Unlock()

	for _, order := range orders {
		bids, err := bm.storage.GetBids(ctx, order.ID.String())
		if err != nil {
			bm.logger.Warn("Failed to load bids for order", "orderID", order.ID, "error", err)
			continue
		}

		for i, bid := range bids {
			if bid.Status == types.BidStatusActive {
				// Use orderID only as key (consistent with new approach)
				bidKey := order.ID.String()
				extendedBid := &ExtendedBid{
					Bid:     bid,
					OrderID: order.ID.String(),
				}
				bm.pendingBids[bidKey] = extendedBid
				bm.logger.Debug("Loaded pending bid from storage", "orderID", order.ID.String(), "bidIndex", i)
			}
		}
	}

	bm.logger.Info("Loaded persisted bids", "pendingBids", len(bm.pendingBids))
	return nil
}

// GetStats returns bid manager statistics
func (bm *Manager) GetStats() map[string]interface{} {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	return map[string]interface{}{
		"pendingBids": len(bm.pendingBids),
		"bidResults":  len(bm.bidResults),
		"isRunning":   bm.isRunning,
	}
}

// allocateResourcesForBid allocates resources for a bid
func (bm *Manager) allocateResourcesForBid(ctx context.Context, orderID *big.Int, machineID *big.Int) error {
	// Get order details to determine resource requirements
	order, err := bm.bidMarket.GetOrder(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order details: %w", err)
	}

	// Get machine details
	machine, err := bm.resourceManager.GetMachine(ctx, machineID)
	if err != nil {
		return fmt.Errorf("failed to get machine details: %w", err)
	}

	// Create resource usage based on order requirements
	resourceUsage := &types.ResourceUsage{
		CPUUsed:     order.CpuCores,
		GPUUsed:     order.GpuCores,
		MemoryUsed:  order.MemoryMB,
		DiskUsed:    order.DiskGB,
		NetworkUsed: order.UploadMbps, // Using upload speed as network usage
	}

	// Allocate resources
	if err := bm.resourceManager.AllocateResources(ctx, orderID, machine, resourceUsage); err != nil {
		return fmt.Errorf("failed to allocate resources: %w", err)
	}

	// Start the resource
	if err := bm.resourceManager.StartResource(ctx, orderID, machine); err != nil {
		return fmt.Errorf("failed to start resource: %w", err)
	}

	bm.logger.Info("Resources allocated and started for bid",
		"orderID", orderID.String(),
		"machineID", machineID.String(),
		"resourceUsage", resourceUsage)

	return nil
}

// TryBidOnOrder attempts to bid on an order if conditions are met
func (bm *Manager) TryBidOnOrder(ctx context.Context, order *types.Order) error {
	// Check if order is open for bidding
	if order.Status != types.OrderStatusOpen {
		bm.logger.Debug("Order not open for bidding", "orderID", order.ID, "status", order.Status)
		return nil
	}

	// Check if bidding is still open
	isOpen, err := bm.bidMarket.IsBiddingOpen(ctx, order.ID)
	if err != nil {
		return fmt.Errorf("failed to check if bidding is open: %w", err)
	}
	if !isOpen {
		bm.logger.Debug("Bidding is closed for order", "orderID", order.ID)
		return nil
	}

	// Check if we already have a bid for this order
	bm.mu.RLock()
	bidKey := order.ID.String()
	_, hasBid := bm.pendingBids[bidKey]
	bm.mu.RUnlock()

	if hasBid {
		bm.logger.Debug("Already have a bid for order", "orderID", order.ID)
		return nil
	}

	// Find suitable machine for this order
	machine, err := bm.findSuitableMachine(ctx, order)
	if err != nil {
		bm.logger.Debug("No suitable machine found for order", "orderID", order.ID, "error", err)
		return nil // Not an error, just no suitable machine
	}

	// Calculate bid price using PricingService
	pricePerSecond, err := bm.pricingService.CalculateBidPrice(ctx, order, machine, nil) // TODO: Get market data
	if err != nil {
		return fmt.Errorf("failed to calculate bid price: %w", err)
	}

	// Check if price is within order limits
	if pricePerSecond.Cmp(order.MaxBidPrice) > 0 {
		bm.logger.Debug("Calculated price exceeds order max price",
			"orderID", order.ID,
			"calculatedPrice", pricePerSecond,
			"maxPrice", order.MaxBidPrice)
		return nil
	}

	// Submit bid
	bm.logger.Info("Attempting to bid on order",
		"orderID", order.ID,
		"pricePerSecond", pricePerSecond,
		"machineID", machine.ID)

	result, err := bm.SubmitBid(ctx, order.ID, pricePerSecond, machine.ID)
	if err != nil {
		return fmt.Errorf("failed to submit bid: %w", err)
	}

	bm.logger.Info("Successfully submitted bid",
		"orderID", order.ID,
		"bidIndex", result.BidIndex,
		"txHash", result.TxHash)

	return nil
}

// findSuitableMachine finds a machine that can fulfill the order requirements
func (bm *Manager) findSuitableMachine(ctx context.Context, order *types.Order) (*types.Machine, error) {
	machines := bm.resourceManager.GetAllMachines(ctx)

	for _, machine := range machines {
		if !machine.Active {
			continue
		}

		// Check if machine type matches
		if machine.MachineType.Cmp(order.MachineType) != 0 {
			continue
		}

		// Check if region matches (if specified)
		if order.Region != nil && machine.Region.Cmp(order.Region) != 0 {
			continue
		}

		// Check if machine has sufficient resources
		required := &types.ResourceUsage{
			CPUUsed:    order.CpuCores,
			GPUUsed:    order.GpuCores,
			MemoryUsed: order.MemoryMB,
			DiskUsed:   order.DiskGB,
		}

		canAllocate, err := bm.resourceManager.CanAllocateResources(ctx, machine, required)
		if err != nil {
			bm.logger.Debug("Failed to check resource allocation", "machineID", machine.ID, "error", err)
			continue
		}

		if canAllocate {
			return machine, nil
		}
	}

	return nil, fmt.Errorf("no suitable machine found")
}
