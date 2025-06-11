package bid

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
)

// BidTracker tracks bids for deployments
type BidTracker struct {
	mu     sync.RWMutex
	bids   map[string]map[common.Address]*types.BidSubmittedEvent
	logger *zap.Logger
}

// NewBidTracker creates a new bid tracker
func NewBidTracker(logger *zap.Logger) *BidTracker {
	return &BidTracker{
		bids:   make(map[string]map[common.Address]*types.BidSubmittedEvent),
		logger: logger,
	}
}

// AddBid adds a bid to the tracker
func (t *BidTracker) AddBid(ctx context.Context, deploymentID string, provider common.Address, amount *big.Int, duration time.Duration) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if deploymentID == "" {
		return &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "deploymentID cannot be empty",
		}
	}
	if provider == (common.Address{}) {
		return &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "provider cannot be empty",
		}
	}
	if amount == nil || amount.Sign() <= 0 {
		return &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "amount must be positive",
		}
	}
	if duration <= 0 {
		return &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "duration must be positive",
		}
	}

	// Initialize deployment bids map if it doesn't exist
	if _, exists := t.bids[deploymentID]; !exists {
		t.bids[deploymentID] = make(map[common.Address]*types.BidSubmittedEvent)
	}

	// Create bid event
	bid := &types.BidSubmittedEvent{
		BaseEvent: types.BaseEvent{
			Timestamp: time.Now(),
		},
		DeploymentID: deploymentID,
		Provider:     provider,
		Amount:       amount,
		Duration:     duration,
	}

	// Store bid
	t.bids[deploymentID][provider] = bid

	t.logger.Info("bid added",
		zap.String("deploymentID", deploymentID),
		zap.String("provider", provider.String()),
		zap.String("amount", amount.String()),
		zap.String("duration", duration.String()),
	)

	return nil
}

// GetBids gets all bids for a deployment
func (t *BidTracker) GetBids(ctx context.Context, deploymentID string) ([]*types.BidSubmittedEvent, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if deploymentID == "" {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "deploymentID cannot be empty",
		}
	}

	bids, exists := t.bids[deploymentID]
	if !exists {
		return nil, nil
	}

	result := make([]*types.BidSubmittedEvent, 0, len(bids))
	for _, bid := range bids {
		result = append(result, bid)
	}

	return result, nil
}

// GetBid gets a specific bid
func (t *BidTracker) GetBid(ctx context.Context, deploymentID string, provider common.Address) (*types.BidSubmittedEvent, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if deploymentID == "" {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "deploymentID cannot be empty",
		}
	}
	if provider == (common.Address{}) {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "provider cannot be empty",
		}
	}

	bids, exists := t.bids[deploymentID]
	if !exists {
		return nil, nil
	}

	return bids[provider], nil
}

// RemoveBid removes a bid
func (t *BidTracker) RemoveBid(ctx context.Context, deploymentID string, provider common.Address) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if deploymentID == "" {
		return &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "deploymentID cannot be empty",
		}
	}
	if provider == (common.Address{}) {
		return &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "provider cannot be empty",
		}
	}

	bids, exists := t.bids[deploymentID]
	if !exists {
		return nil
	}

	delete(bids, provider)
	if len(bids) == 0 {
		delete(t.bids, deploymentID)
	}

	t.logger.Info("bid removed",
		zap.String("deploymentID", deploymentID),
		zap.String("provider", provider.String()),
	)

	return nil
}

// ClearBids clears all bids for a deployment
func (t *BidTracker) ClearBids(ctx context.Context, deploymentID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if deploymentID == "" {
		return &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "deploymentID cannot be empty",
		}
	}

	delete(t.bids, deploymentID)

	t.logger.Info("bids cleared",
		zap.String("deploymentID", deploymentID),
	)

	return nil
}

// ValidateBid validates a bid
func (t *BidTracker) ValidateBid(ctx context.Context, deploymentID string, provider common.Address, amount *big.Int, duration time.Duration) error {
	if deploymentID == "" {
		return &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "deploymentID cannot be empty",
		}
	}
	if provider == (common.Address{}) {
		return &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "provider cannot be empty",
		}
	}
	if amount == nil || amount.Sign() <= 0 {
		return &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "amount must be positive",
		}
	}
	if duration <= 0 {
		return &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "duration must be positive",
		}
	}

	// Check if bid already exists
	existingBid, err := t.GetBid(ctx, deploymentID, provider)
	if err != nil {
		return err
	}
	if existingBid != nil {
		return &types.ContractError{
			Code:    types.ErrCodeBidInvalid,
			Message: "bid already exists",
		}
	}

	return nil
}

// IsBidValid checks if a bid is valid
func (t *BidTracker) IsBidValid(ctx context.Context, deploymentID string, provider common.Address) (bool, error) {
	bid, err := t.GetBid(ctx, deploymentID, provider)
	if err != nil {
		return false, err
	}
	return bid != nil, nil
}

// GetBidStatus gets the status of a bid
func (t *BidTracker) GetBidStatus(ctx context.Context, deploymentID string, provider common.Address) (types.BidStatus, error) {
	bid, err := t.GetBid(ctx, deploymentID, provider)
	if err != nil {
		return "", err
	}
	if bid == nil {
		return "", &types.ContractError{
			Code:    types.ErrCodeBidInvalid,
			Message: "bid not found",
		}
	}
	return types.BidStatusActive, nil // TODO: Implement proper status tracking
}

// UpdateBidStatus updates the status of a bid
func (t *BidTracker) UpdateBidStatus(ctx context.Context, deploymentID string, provider common.Address, status types.BidStatus) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if deploymentID == "" {
		return &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "deploymentID cannot be empty",
		}
	}
	if provider == (common.Address{}) {
		return &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "provider cannot be empty",
		}
	}
	if status == "" {
		return &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "status cannot be empty",
		}
	}

	bids, exists := t.bids[deploymentID]
	if !exists {
		return &types.ContractError{
			Code:    types.ErrCodeBidInvalid,
			Message: "deployment not found",
		}
	}

	bid, exists := bids[provider]
	if !exists {
		return &types.ContractError{
			Code:    types.ErrCodeBidInvalid,
			Message: "bid not found",
		}
	}

	bid.Status = status
	bid.BaseEvent.Timestamp = time.Now()

	t.logger.Info("bid status updated",
		zap.String("deploymentID", deploymentID),
		zap.String("provider", provider.String()),
		zap.String("status", string(status)),
	)

	return nil
}
