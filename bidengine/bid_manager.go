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
) *BidManagerService {
	ctx, cancel := context.WithCancel(context.Background())

	return &BidManagerService{
		config:      config,
		bidMarket:   bidMarket,
		logger:      logger,
		metrics:     metrics,
		storage:     NewStorage(datastore, logger),
		ctx:         ctx,
		cancel:      cancel,
		pendingBids: make(map[string]*ExtendedBid),
		bidResults:  make(map[string]*BidResult),
	}
}

// Start starts the bid manager
func (bm *BidManagerService) Start(ctx context.Context) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if bm.isRunning {
		return nil
	}

	bm.logger.Info("Starting BidManager")
	bm.isRunning = true

	// Load persisted bids from storage
	if err := bm.loadPersistedBids(ctx); err != nil {
		bm.logger.Warn("Failed to load persisted bids", "error", err)
	}

	// Start monitoring goroutine
	go bm.bidMonitoringLoop(ctx)

	bm.logger.Info("BidManager started successfully")
	return nil
}

// Stop stops the bid manager
func (bm *BidManagerService) Stop(ctx context.Context) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if !bm.isRunning {
		return nil
	}

	bm.logger.Info("Stopping BidManager")
	bm.isRunning = false

	// Save current state to storage
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

	// Use a default provider ID for now (this should come from config)
	providerID := big.NewInt(1)

	// Submit bid to blockchain
	tx, err := bm.bidMarket.SubmitBid(nil, orderID, pricePerSecond, providerID, machineID)
	if err != nil {
		bm.logger.Error("Failed to submit bid to blockchain",
			"orderID", orderID.String(),
			"machineID", machineID.String(),
			"error", err)
		return nil, fmt.Errorf("failed to submit bid to blockchain: %w", err)
	}

	// Get bid index from transaction (placeholder)
	bidIndex := big.NewInt(0)

	// Create extended bid
	extendedBid := &ExtendedBid{
		Bid: &Bid{
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

	// Track bid
	bidKey := fmt.Sprintf("%s-%s", orderID.String(), machineID.String())
	bm.pendingBids[bidKey] = extendedBid

	// Save bid to storage
	if err := bm.storage.SaveBid(ctx, extendedBid.Bid, orderID.String(), 0); err != nil {
		bm.logger.Warn("Failed to save bid to storage", "error", err)
	}

	bm.logger.Info("Bid submitted successfully",
		"orderID", orderID.String(),
		"machineID", machineID.String(),
		"price", pricePerSecond,
		"txHash", tx.Hash().String(),
		"bidIndex", bidIndex)

	bm.metrics.IncrementBidsSubmitted()

	// Create bid result
	result := &BidResult{
		OrderID:   orderID,
		BidIndex:  bidIndex,
		Success:   true,
		TxHash:    tx.Hash(),
		Timestamp: time.Now(),
	}

	return result, nil
}

// CancelBid cancels a pending bid
func (bm *BidManagerService) CancelBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Cancel bid on blockchain
	tx, err := bm.bidMarket.CancelBid(nil, orderID, bidIndex)
	if err != nil {
		bm.logger.Error("Failed to cancel bid on blockchain",
			"orderID", orderID.String(),
			"bidIndex", bidIndex.String(),
			"error", err)
		return fmt.Errorf("failed to cancel bid on blockchain: %w", err)
	}

	// Update bid status
	bidKey := fmt.Sprintf("%s-%s", orderID.String(), bidIndex.String())
	if bid, exists := bm.pendingBids[bidKey]; exists {
		bid.Status = BidStatusCancelled
		bid.CancelledAt = time.Now()

		// Update bid in storage
		if err := bm.storage.UpdateBid(ctx, bid.Bid, bid.OrderID, 0); err != nil {
			bm.logger.Warn("Failed to update bid in storage", "error", err)
		}

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
	bidKey := fmt.Sprintf("%s-%s", orderID.String(), bidIndex.String())

	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Check if bid is already tracked
	if _, exists := bm.pendingBids[bidKey]; exists {
		return fmt.Errorf("bid is already tracked")
	}

	// Create placeholder extended bid for tracking
	extendedBid := &ExtendedBid{
		Bid: &Bid{
			Status:    BidStatusActive,
			CreatedAt: big.NewInt(time.Now().Unix()),
		},
		OrderID: orderID.String(),
	}

	bm.pendingBids[bidKey] = extendedBid
	bm.logger.Info("Bid tracked", "orderID", orderID.String(), "bidIndex", bidIndex.String())
	return nil
}

// UntrackBid untracks a bid
func (bm *BidManagerService) UntrackBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) error {
	bidKey := fmt.Sprintf("%s-%s", orderID.String(), bidIndex.String())

	bm.mu.Lock()
	defer bm.mu.Unlock()

	if _, exists := bm.pendingBids[bidKey]; !exists {
		return fmt.Errorf("bid is not tracked")
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

	for bidKey := range bm.pendingBids {
		// Parse orderID and bidIndex from key
		// This is a simplified approach - in real implementation you'd need proper parsing
		orderID, _ := new(big.Int).SetString(bidKey[:len(bidKey)/2], 10)
		bidIndex, _ := new(big.Int).SetString(bidKey[len(bidKey)/2:], 10)

		if result[orderID] == nil {
			result[orderID] = []*big.Int{}
		}
		result[orderID] = append(result[orderID], bidIndex)
	}

	return result, nil
}

// MonitorBidStatus monitors the status of a specific bid
func (bm *BidManagerService) MonitorBidStatus(ctx context.Context, orderID *big.Int, bidIndex *big.Int) error {
	bidKey := fmt.Sprintf("%s-%s", orderID.String(), bidIndex.String())

	bm.mu.RLock()
	bid, exists := bm.pendingBids[bidKey]
	bm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("bid is not tracked")
	}

	return bm.checkBidStatus(ctx, bid)
}

// CheckBidExpiry checks if a bid has expired
func (bm *BidManagerService) CheckBidExpiry(ctx context.Context, orderID *big.Int, bidIndex *big.Int) error {
	// This would check if the bid has expired based on the order's bidding deadline
	// For now, we'll just log the check
	bm.logger.Debug("Checking bid expiry", "orderID", orderID.String(), "bidIndex", bidIndex.String())
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
	pendingBids := make([]*ExtendedBid, 0, len(bm.pendingBids))
	for _, bid := range bm.pendingBids {
		if bid.Status == BidStatusActive {
			pendingBids = append(pendingBids, bid)
		}
	}
	bm.mu.RUnlock()

	for _, bid := range pendingBids {
		if err := bm.checkBidStatus(ctx, bid); err != nil {
			bm.logger.Warn("Failed to check bid status",
				"orderID", bid.OrderID,
				"machineID", bid.MachineId.String(),
				"error", err)
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

	// Get bids from blockchain
	blockchainBids, err := bm.bidMarket.GetBids(nil, orderID)
	if err != nil {
		return fmt.Errorf("failed to get bids from blockchain: %w", err)
	}

	// Find our bid and check its status
	for i, blockchainBid := range blockchainBids {
		if blockchainBid.Provider.String() == bid.Provider.String() {
			// Update bid status based on blockchain status
			oldStatus := bid.Status
			bid.Status = blockchainBid.Status

			if bid.Status != oldStatus {
				bm.logger.Info("Bid status updated",
					"orderID", bid.OrderID,
					"machineID", bid.MachineId.String(),
					"oldStatus", oldStatus,
					"newStatus", bid.Status)

				// Update bid in storage
				if err := bm.storage.UpdateBid(ctx, bid.Bid, bid.OrderID, i); err != nil {
					bm.logger.Warn("Failed to update bid in storage", "error", err)
				}

				// Record metrics
				if bid.Status == BidStatusAccepted {
					bm.metrics.IncrementBidsAccepted()
				} else if bid.Status == BidStatusCancelled {
					bm.metrics.IncrementBidsRejected()
				}
			}

			break
		}
	}

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

	for _, order := range orders {
		bids, err := bm.storage.GetBids(ctx, order.ID.String())
		if err != nil {
			bm.logger.Warn("Failed to load bids for order", "orderID", order.ID, "error", err)
			continue
		}

		for i, bid := range bids {
			if bid.Status == BidStatusActive {
				bidKey := fmt.Sprintf("%s-%s", order.ID.String(), bid.MachineId.String())
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
