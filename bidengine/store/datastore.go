package store

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/bidengine/types"
	"github.com/unicornultrafoundation/subnet-node/repo"
)

const (
	// BidKeyPrefix is the namespace prefix for bid storage
	BidKeyPrefix = "/bids/"
)

// Store implements BidDatastore using repo.Datastore
type Store struct {
	ds    repo.Datastore
	log   *logrus.Logger
	mutex sync.RWMutex
}

// NewStore creates a new datastore-based bid storage
func NewStore(ds repo.Datastore, log *logrus.Logger) (*Store, error) {
	if ds == nil {
		return nil, fmt.Errorf("datastore cannot be nil")
	}

	return &Store{
		ds:  ds,
		log: log,
	}, nil
}

// bidKey generates a store key for a bid
func bidKey(bidID string) datastore.Key {
	return datastore.NewKey(BidKeyPrefix + bidID)
}

// SaveBid stores or updates a bid in the datastore
func (s *Store) SaveBid(ctx context.Context, bid *types.Bid) error {
	if bid == nil || bid.ID == nil {
		return fmt.Errorf("cannot save nil bid or bid with nil ID")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Marshal bid to JSON
	data, err := json.Marshal(bid)
	if err != nil {
		return fmt.Errorf("failed to marshal bid: %v", err)
	}

	// Save to datastore
	key := bidKey(bid.ID.String())
	if err := s.ds.Put(ctx, key, data); err != nil {
		return fmt.Errorf("failed to write bid to datastore: %v", err)
	}

	s.log.WithField("bidId", bid.ID.String()).Debug("Saved bid to datastore")
	return nil
}

// GetBidByID retrieves a bid by its ID
func (s *Store) GetBidByID(ctx context.Context, bidID string) (*types.Bid, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	key := bidKey(bidID)
	data, err := s.ds.Get(ctx, key)
	if err != nil {
		if err == datastore.ErrNotFound {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to read bid from datastore: %v", err)
	}

	var bid types.Bid
	if err := json.Unmarshal(data, &bid); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bid: %v", err)
	}

	return &bid, nil
}

// ListActiveBids retrieves all active bids
func (s *Store) ListActiveBids(ctx context.Context) ([]*types.Bid, error) {
	return s.listBidsWithFilter(ctx, func(bid *types.Bid) bool {
		return bid.Status != types.BidStatusExpired && bid.Status != types.BidStatusRejected
	})
}

// ListBidsByStatus retrieves bids by their status
func (s *Store) ListBidsByStatus(ctx context.Context, status types.BidStatus) ([]*types.Bid, error) {
	return s.listBidsWithFilter(ctx, func(bid *types.Bid) bool {
		return bid.Status == status
	})
}

// ListAllBids retrieves all bids from the datastore
func (s *Store) ListAllBids(ctx context.Context) ([]*types.Bid, error) {
	return s.listBidsWithFilter(ctx, func(bid *types.Bid) bool {
		return true // Return all bids
	})
}

// listBidsWithFilter is a helper function to list bids with a filter function
func (s *Store) listBidsWithFilter(ctx context.Context, filter func(*types.Bid) bool) ([]*types.Bid, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var bids []*types.Bid

	// Query all keys with the bid prefix
	results, err := s.ds.Query(ctx, query.Query{
		Prefix: BidKeyPrefix,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query datastore for bids: %v", err)
	}
	defer results.Close()

	// Process each entry
	for {
		result, ok := results.NextSync()
		if !ok {
			break
		}
		if result.Error != nil {
			s.log.WithError(result.Error).WithField("key", result.Key).Warn("Failed to read bid from datastore")
			continue
		}

		// Unmarshal the bid
		var bid types.Bid
		if err := json.Unmarshal(result.Value, &bid); err != nil {
			s.log.WithError(err).WithField("key", result.Key).Warn("Failed to unmarshal bid")
			continue
		}

		// Apply filter
		if filter(&bid) {
			bids = append(bids, &bid)
		}
	}

	return bids, nil
}

// DeleteBid removes a bid from the datastore
func (s *Store) DeleteBid(ctx context.Context, bidID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	key := bidKey(bidID)
	if err := s.ds.Delete(ctx, key); err != nil {
		if err == datastore.ErrNotFound {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to delete bid from datastore: %v", err)
	}

	s.log.WithField("bidId", bidID).Debug("Deleted bid from datastore")
	return nil
}
