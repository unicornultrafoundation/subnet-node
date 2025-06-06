package bid

import (
	"sync"

	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
)

// BidTracker tracks bids for deployments
type BidTracker struct {
	mu       sync.RWMutex
	bids     map[string][]*types.BidSubmittedEvent
	requests map[string]*types.DeploymentRequestedEvent
}

// NewBidTracker creates a new bid tracker
func NewBidTracker() *BidTracker {
	return &BidTracker{
		bids:     make(map[string][]*types.BidSubmittedEvent),
		requests: make(map[string]*types.DeploymentRequestedEvent),
	}
}

func (t *BidTracker) GetBids(deploymentID string) []*types.BidSubmittedEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.bids[deploymentID]
}

func (t *BidTracker) GetRequests(deploymentID string) *types.DeploymentRequestedEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.requests[deploymentID]
}

func (t *BidTracker) StoreBid(deploymentID string, bid *types.BidSubmittedEvent) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.bids[deploymentID] = append(t.bids[deploymentID], bid)
}

func (t *BidTracker) StoreRequest(deploymentID string, request *types.DeploymentRequestedEvent) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.requests[deploymentID] = request
}
