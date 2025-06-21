package bidengine

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	ds "github.com/ipfs/go-datastore"
)

// ExtendedBid represents a bid with additional fields for internal use
type ExtendedBid struct {
	*Bid
	OrderID         string
	TransactionHash string
	SubmittedAt     time.Time
	CancelledAt     time.Time
}

// BidManagerService manages bid submission and tracking
type BidManagerService struct {
	config          *BidEngineConfig
	bidMarket       BidMarketContract
	logger          Logger
	metrics         Metrics
	storage         *Storage
	resourceManager ResourceManager
	pricingService  PricingEngine

	mu        sync.RWMutex
	isRunning bool
	ctx       context.Context
	cancel    context.CancelFunc

	// Tracking
	pendingBids map[string]*ExtendedBid
	bidResults  map[string]*BidResult
}

// NewBidManager creates a new BidManagerService instance
func NewBidManager(
	config *BidEngineConfig,
	bidMarket BidMarketContract,
	logger Logger,
	metrics Metrics,
	datastore ds.Datastore,
	resourceManager ResourceManager,
	pricingService PricingEngine,
) *BidManagerService {
	ctx, cancel := context.WithCancel(context.Background())

	return &BidManagerService{
		config:          config,
		bidMarket:       bidMarket,
		logger:          logger,
		metrics:         metrics,
		storage:         NewStorage(datastore, logger),
		resourceManager: resourceManager,
		pricingService:  pricingService,
		ctx:             ctx,
		cancel:          cancel,
		pendingBids:     make(map[string]*ExtendedBid),
		bidResults:      make(map[string]*BidResult),
	}
}

// Start starts the bid manager
func (bm *BidManagerService) Start(ctx context.Context) error {
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

	// Start monitoring goroutine
	go bm.bidMonitoringLoop(ctx)

	// Start expiry check goroutine
	go bm.expiryCheckLoop(ctx)

	bm.logger.Info("BidManager started successfully")
	return nil
}

// Stop stops the bid manager
func (bm *BidManagerService) Stop(ctx context.Context) error {
	bm.mu.Lock()
	if !bm.isRunning {
		bm.mu.Unlock()
		return nil
	}

	bm.logger.Info("Stopping BidManager")
	bm.isRunning = false
	bm.mu.Unlock()

	// Save current state to storage (without lock)
	if err := bm.savePersistedBids(ctx); err != nil {
		bm.logger.Warn("Failed to save persisted bids", "error", err)
	}

	// Cancel context
	bm.cancel()

	bm.logger.Info("BidManager stopped successfully")
	return nil
}

// SubmitBid submits a bid to the blockchain
func (bm *BidManagerService) SubmitBid(ctx context.Context, orderID *big.Int, pricePerSecond *big.Int, machineID *big.Int) (*BidResult, error) {
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
		Bid: &Bid{
			Id:             bidIndex,
			Provider:       bm.config.ProviderWallet,
			PricePerSecond: pricePerSecond,
			Status:         BidStatusActive,
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
	bidResult := &BidResult{
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
func (bm *BidManagerService) CancelBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) error {
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
		bid.Status = BidStatusCancelled
		bid.CancelledAt = time.Now()

		// Update bid in storage with correct bidIndex
		if err := bm.storage.UpdateBid(ctx, bid.Bid, bid.OrderID, int(bidIndex.Int64())); err != nil {
			bm.logger.Warn("Failed to update bid in storage", "error", err)
		}

		// Deallocate resources for cancelled bid
		// Unlock before calling deallocateResourcesForBid to avoid deadlock
		bm.mu.Unlock()
		if err := bm.deallocateResourcesForBid(ctx, orderID); err != nil {
			bm.logger.Warn("Failed to deallocate resources for cancelled bid",
				"orderID", orderID.String(),
				"error", err)
		}
		// Re-lock to continue with cleanup
		bm.mu.Lock()

		delete(bm.pendingBids, bidKey)
	}

	bm.logger.Info("Bid cancelled successfully",
		"orderID", orderID.String(),
		"bidIndex", bidIndex.String(),
		"txHash", tx.Hash().String())

	return nil
}

// TrackBid tracks a bid
func (bm *BidManagerService) TrackBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	return bm.trackBidInternal(ctx, orderID, bidIndex)
}

// trackBidInternal is an internal method for tracking a bid
func (bm *BidManagerService) trackBidInternal(ctx context.Context, orderID *big.Int, bidIndex *big.Int) error {
	// Track bid using orderID only
	bidKey := orderID.String()
	if _, exists := bm.pendingBids[bidKey]; exists {
		return fmt.Errorf("bid already tracked for order %s", orderID.String())
	}

	// Create extended bid for tracking
	extendedBid := &ExtendedBid{
		Bid: &Bid{
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
func (bm *BidManagerService) UntrackBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	return bm.untrackBidInternal(ctx, orderID, bidIndex)
}

// untrackBidInternal is an internal method for untracking a bid
func (bm *BidManagerService) untrackBidInternal(ctx context.Context, orderID *big.Int, bidIndex *big.Int) error {
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
func (bm *BidManagerService) GetTrackedBids(ctx context.Context) (map[*big.Int][]*big.Int, error) {
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

// MonitorBidStatus monitors the status of a specific bid
func (bm *BidManagerService) MonitorBidStatus(ctx context.Context, orderID *big.Int, bidIndex *big.Int) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Check if bid is tracked using orderID only
	bidKey := orderID.String()
	if _, exists := bm.pendingBids[bidKey]; !exists {
		return fmt.Errorf("bid not tracked for order %s", orderID.String())
	}

	// Start monitoring (implementation would depend on your monitoring strategy)
	bm.logger.Info("Bid status monitoring started", "orderID", orderID.String(), "bidIndex", bidIndex.String())

	return nil
}

// CheckBidExpiry checks if a bid has expired
func (bm *BidManagerService) CheckBidExpiry(ctx context.Context, orderID *big.Int, bidIndex *big.Int) error {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	// Check bid expiry using orderID only
	bidKey := orderID.String()
	_, exists := bm.pendingBids[bidKey]
	if !exists {
		return fmt.Errorf("bid not tracked for order %s", orderID.String())
	}

	// Check if bid has expired (implementation would depend on your expiry logic)
	bm.logger.Info("Bid expiry check", "orderID", orderID.String(), "bidIndex", bidIndex.String())

	return nil
}

// GetBid retrieves a bid by order ID and bid index
func (bm *BidManagerService) GetBid(ctx context.Context, orderID string, bidIndex int) (*Bid, error) {
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
func (bm *BidManagerService) GetBids(ctx context.Context, orderID string) ([]*Bid, error) {
	return bm.storage.GetBids(ctx, orderID)
}

// GetPendingBids returns all pending bids
func (bm *BidManagerService) GetPendingBids(ctx context.Context) []*Bid {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	var pendingBids []*Bid
	for _, extendedBid := range bm.pendingBids {
		if extendedBid.Status == BidStatusActive {
			pendingBids = append(pendingBids, extendedBid.Bid)
		}
	}

	return pendingBids
}

// bidMonitoringLoop monitors bid statuses
func (bm *BidManagerService) bidMonitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(bm.config.BidCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			bm.checkBidStatuses(ctx)
		}
	}
}

// checkBidStatuses checks the status of pending bids
func (bm *BidManagerService) checkBidStatuses(ctx context.Context) {
	bm.mu.RLock()
	bids := make([]*ExtendedBid, 0, len(bm.pendingBids))
	for bidKey, bid := range bm.pendingBids {
		// Parse orderID from bidKey (which is now just orderID)
		orderID, ok := new(big.Int).SetString(bidKey, 10)
		if !ok {
			bm.logger.Warn("Invalid orderID in bidKey", "bidKey", bidKey)
			continue
		}
		// Create a copy to avoid race conditions
		bidCopy := &ExtendedBid{
			Bid:             bid.Bid,
			OrderID:         orderID.String(),
			TransactionHash: bid.TransactionHash,
			SubmittedAt:     bid.SubmittedAt,
			CancelledAt:     bid.CancelledAt,
		}
		bids = append(bids, bidCopy)
	}
	bm.mu.RUnlock()

	for _, bid := range bids {
		if err := bm.checkBidStatus(ctx, bid); err != nil {
			bm.logger.Warn("Failed to check bid status", "error", err)
		}
	}
}

// checkBidStatus checks the status of a specific bid
func (bm *BidManagerService) checkBidStatus(ctx context.Context, bid *ExtendedBid) error {
	// Convert string ID to big.Int
	orderID, ok := new(big.Int).SetString(bid.OrderID, 10)
	if !ok {
		return fmt.Errorf("invalid order ID: %s", bid.OrderID)
	}

	// Get the bid from blockchain by orderID and bidIndex
	blockchainBid, err := bm.bidMarket.OrderBids(ctx, orderID, bid.Bid.Id)
	if err != nil {
		return fmt.Errorf("failed to get bid from blockchain: %w", err)
	}

	// Update bid status based on blockchain status
	oldStatus := bid.Status
	newStatus := blockchainBid.Status

	if newStatus != oldStatus {
		bm.logger.Info("Bid status updated",
			"orderID", bid.OrderID,
			"machineID", bid.MachineId.String(),
			"oldStatus", oldStatus,
			"newStatus", newStatus)

		// Lock to update the actual bid in pendingBids
		bm.mu.Lock()
		bidKey := bid.OrderID
		if actualBid, exists := bm.pendingBids[bidKey]; exists {
			actualBid.Status = newStatus
			// Update bid in storage
			if err := bm.storage.UpdateBid(ctx, actualBid.Bid, actualBid.OrderID, int(actualBid.Bid.Id.Int64())); err != nil {
				bm.logger.Warn("Failed to update bid in storage", "error", err)
			}

			// Handle resource deallocation when bid is cancelled or expired
			if (newStatus == BidStatusCancelled || newStatus == BidStatusExpired) &&
				(oldStatus != BidStatusCancelled && oldStatus != BidStatusExpired) {
				// Unlock before calling deallocateResourcesForBid to avoid deadlock
				bm.mu.Unlock()
				if err := bm.deallocateResourcesForBid(ctx, orderID); err != nil {
					bm.logger.Error("Failed to deallocate resources for cancelled/expired bid",
						"orderID", bid.OrderID,
						"error", err)
				}
				// Re-lock to continue with metrics
				bm.mu.Lock()
			}

			// Record metrics
			switch newStatus {
			case BidStatusAccepted:
				bm.metrics.IncrementBidsAccepted()
			case BidStatusCancelled:
				bm.metrics.IncrementBidsRejected()
			case BidStatusExpired:
				bm.metrics.IncrementBidsRejected()
			}
		}
		bm.mu.Unlock()
	}

	return nil
}

// deallocateResourcesForBid deallocates resources for a bid
func (bm *BidManagerService) deallocateResourcesForBid(ctx context.Context, orderID *big.Int) error {
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
func (bm *BidManagerService) loadPersistedBids(ctx context.Context) error {
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
			if bid.Status == BidStatusActive {
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

// savePersistedBids saves current bid state to storage
func (bm *BidManagerService) savePersistedBids(ctx context.Context) error {
	bm.logger.Info("Saving persisted bids to storage")

	bm.mu.RLock()
	defer bm.mu.RUnlock()

	for bidKey, extendedBid := range bm.pendingBids {
		// Extract bid index from key (this is a simplified approach)
		// In a real implementation, you'd need to track bid indices properly
		bidIndex := 0 // Placeholder

		if err := bm.storage.UpdateBid(ctx, extendedBid.Bid, extendedBid.OrderID, bidIndex); err != nil {
			bm.logger.Warn("Failed to save bid to storage", "bidKey", bidKey, "error", err)
		}
	}

	return nil
}

// GetStats returns bid manager statistics
func (bm *BidManagerService) GetStats() map[string]interface{} {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	return map[string]interface{}{
		"pendingBids": len(bm.pendingBids),
		"bidResults":  len(bm.bidResults),
		"isRunning":   bm.isRunning,
	}
}

// CheckOrderExpiry checks if an order has expired and closes it if necessary
func (bm *BidManagerService) CheckOrderExpiry(ctx context.Context, orderID *big.Int) error {
	// Get order from blockchain to check expiry
	order, err := bm.bidMarket.GetOrder(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order from blockchain: %w", err)
	}

	now := time.Now().Unix()
	if IsOrderReadyToClose(order, now) {
		bm.logger.Info("Order has expired and 1 day grace period passed, closing it",
			"orderID", orderID.String(),
			"expiredAt", order.ExpiredAt.String(),
			"currentTime", now)

		// Deallocate resources first
		if err := bm.deallocateResourcesForBid(ctx, orderID); err != nil {
			bm.logger.Warn("Failed to deallocate resources for expired order",
				"orderID", orderID.String(),
				"error", err)
		}

		// Close the order on blockchain
		tx, err := bm.bidMarket.CloseOrder(ctx, orderID, "Order expired")
		if err != nil {
			return fmt.Errorf("failed to close expired order: %w", err)
		}

		bm.logger.Info("Order closed successfully",
			"orderID", orderID.String(),
			"txHash", tx.Hash().String())

		// Remove from pending bids if exists
		bm.mu.Lock()
		bidKey := orderID.String()
		if _, exists := bm.pendingBids[bidKey]; exists {
			delete(bm.pendingBids, bidKey)
			bm.logger.Info("Removed expired order from pending bids",
				"orderID", orderID.String())
		}
		bm.mu.Unlock()
	} else {
		bm.logger.Debug("Order has expired but grace period not passed yet",
			"orderID", orderID.String(),
			"expiredAt", order.ExpiredAt.String(),
			"currentTime", now)
	}

	return nil
}

// allocateResourcesForBid allocates resources for a bid
func (bm *BidManagerService) allocateResourcesForBid(ctx context.Context, orderID *big.Int, machineID *big.Int) error {
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
	resourceUsage := &ResourceUsage{
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

// expiryCheckLoop periodically checks for expired orders and closes them
func (bm *BidManagerService) expiryCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			bm.mu.RLock()
			orderIDs := make([]*big.Int, 0, len(bm.pendingBids))
			for bidKey := range bm.pendingBids {
				orderID, ok := new(big.Int).SetString(bidKey, 10)
				if ok {
					orderIDs = append(orderIDs, orderID)
				}
			}
			bm.mu.RUnlock()
			for _, orderID := range orderIDs {
				_ = bm.CheckOrderExpiry(ctx, orderID)
			}
		}
	}
}

// TryBidOnOrder attempts to bid on an order if conditions are met
func (bm *BidManagerService) TryBidOnOrder(ctx context.Context, order *Order) error {
	// Check if order is open for bidding
	if order.Status != OrderStatusOpen {
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
func (bm *BidManagerService) findSuitableMachine(ctx context.Context, order *Order) (*Machine, error) {
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
		required := &ResourceUsage{
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
