package pool

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/metrics"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

var log = logrus.WithField("service", "vpn-stream-pool")

// StreamPool manages a pool of streams for each peer
type StreamPool struct {
	// Map of peer ID to a list of streams
	pools map[string][]*pooledStream
	// Mutex to protect access to pools
	mu sync.RWMutex
	// Stream service for creating new streams
	streamService types.Service
	// Maximum streams per peer
	maxStreamsPerPeer int
	// Minimum streams per peer
	minStreamsPerPeer int
	// Stream idle timeout
	streamIdleTimeout time.Duration
	// Metrics for the stream pool
	metrics *metrics.StreamPoolMetrics
	// Context for the pool
	ctx context.Context
	// Cancel function for the pool context
	cancel context.CancelFunc
	// Resilience service for stream operations
	resilienceService *resilience.ResilienceService
}

// pooledStream represents a stream in the pool
type pooledStream struct {
	// The actual stream
	stream types.VPNStream
	// Last time the stream was used
	lastUsed time.Time
	// Whether the stream is in use
	inUse bool
}

// NewStreamPool creates a new stream pool
func NewStreamPool(
	streamService types.Service,
	maxStreamsPerPeer int,
	minStreamsPerPeer int,
	streamIdleTimeout time.Duration,
) *StreamPool {
	ctx, cancel := context.WithCancel(context.Background())

	// Create a resilience service with default configuration
	// Use values that are reasonable for stream operations
	resilienceConfig := &resilience.ResilienceConfig{
		CircuitBreakerFailureThreshold: 5,
		CircuitBreakerResetTimeout:     30 * time.Second,
		CircuitBreakerSuccessThreshold: 2,
		RetryMaxAttempts:               3,
		RetryInitialInterval:           500 * time.Millisecond,
		RetryMaxInterval:               5 * time.Second,
	}
	resilienceService := resilience.NewResilienceService(resilienceConfig)

	return &StreamPool{
		pools:             make(map[string][]*pooledStream),
		streamService:     streamService,
		maxStreamsPerPeer: maxStreamsPerPeer,
		minStreamsPerPeer: minStreamsPerPeer,
		streamIdleTimeout: streamIdleTimeout,
		metrics:           metrics.NewStreamPoolMetrics(),
		ctx:               ctx,
		cancel:            cancel,
		resilienceService: resilienceService,
	}
}

// GetStream gets a stream from the pool or creates a new one
func (p *StreamPool) GetStream(ctx context.Context, peerID peer.ID) (types.VPNStream, error) {
	peerIDStr := peerID.String()

	// Try to get an existing stream from the pool
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if we have a pool for this peer
	pool, exists := p.pools[peerIDStr]
	if !exists {
		// Create a new pool for this peer
		p.pools[peerIDStr] = make([]*pooledStream, 0, p.maxStreamsPerPeer)
		pool = p.pools[peerIDStr]
	}

	// Try to find an available stream
	for i, ps := range pool {
		if !ps.inUse {
			// Mark the stream as in use
			pool[i].inUse = true
			pool[i].lastUsed = time.Now()

			// Update metrics
			p.metrics.IncrementStreamsAcquired()

			return ps.stream, nil
		}
	}

	// No available stream, create a new one if we haven't reached the maximum
	if len(pool) < p.maxStreamsPerPeer {
		// Use the unified resilience function for stream creation
		var newStream types.VPNStream

		logger := log.WithFields(logrus.Fields{
			"peer_id":   peerIDStr,
			"operation": "create_stream",
			"pool_size": len(pool),
			"max_size":  p.maxStreamsPerPeer,
		})

		logger.Debug("Creating new stream for peer")

		err, _ := p.resilienceService.ExecuteWithResilience(
			ctx,
			p.resilienceService.FormatPeerBreakerId(peerID, "create_stream"),
			func() error {
				// Create a new stream
				var createErr error
				newStream, createErr = p.streamService.CreateNewVPNStream(ctx, peerID)
				if createErr != nil {
					logger.WithError(createErr).Warn("Failed to create new stream, will retry")
					return createErr // Return the error to trigger retry
				}
				logger.Debug("Successfully created new stream")
				return nil // Success
			},
		)

		if err != nil {
			// Update metrics
			p.metrics.IncrementAcquisitionFailures()
			logger.WithError(err).Error("Failed to create new stream after retries")
			return nil, fmt.Errorf("failed to create new stream for peer %s: %v", peerIDStr, err)
		}

		// Add the stream to the pool
		ps := &pooledStream{
			stream:   newStream,
			lastUsed: time.Now(),
			inUse:    true,
		}
		p.pools[peerIDStr] = append(pool, ps)

		// Update metrics
		p.metrics.IncrementStreamsCreated()
		p.metrics.IncrementStreamsAcquired()

		return newStream, nil
	}

	// We've reached the maximum number of streams for this peer
	// Update metrics
	p.metrics.IncrementAcquisitionFailures()

	return nil, fmt.Errorf("maximum number of streams reached for peer %s", peerIDStr)
}

// ReleaseStream returns a stream to the pool
func (p *StreamPool) ReleaseStream(peerID peer.ID, s types.VPNStream, healthy bool) {
	peerIDStr := peerID.String()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if we have a pool for this peer
	pool, exists := p.pools[peerIDStr]
	if !exists {
		// This shouldn't happen, but just in case
		log.Warnf("Trying to release a stream for unknown peer %s", peerIDStr)
		if !healthy {
			// Close the stream if it's unhealthy
			s.Close()
		}
		return
	}

	// Find the stream in the pool
	for i, ps := range pool {
		if ps.stream == s {
			if healthy {
				// Mark the stream as not in use
				pool[i].inUse = false
				pool[i].lastUsed = time.Now()

				// Update metrics
				p.metrics.IncrementStreamsReturned()
			} else {
				// Close the unhealthy stream
				ps.stream.Close()

				// Remove the stream from the pool
				p.pools[peerIDStr] = slices.Delete(pool, i, i+1)

				// Update metrics
				p.metrics.IncrementUnhealthyStreams()
				p.metrics.IncrementStreamsClosed()

				// Create a new stream if we're below the minimum
				if len(p.pools[peerIDStr]) < p.minStreamsPerPeer {
					go p.ensureMinStreams(peerID)
				}
			}
			return
		}
	}

	// Stream not found in the pool
	log.Warnf("Trying to release a stream that is not in the pool for peer %s", peerIDStr)
	if !healthy {
		// Close the stream if it's unhealthy
		s.Close()
	}
}

// ensureMinStreams ensures that we have at least minStreamsPerPeer streams for a peer
func (p *StreamPool) ensureMinStreams(peerID peer.ID) {
	peerIDStr := peerID.String()

	// Create a new context for this operation
	ctx, cancel := context.WithTimeout(p.ctx, 5*time.Second)
	defer cancel()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if we have a pool for this peer
	pool, exists := p.pools[peerIDStr]
	if !exists {
		// Create a new pool for this peer
		p.pools[peerIDStr] = make([]*pooledStream, 0, p.maxStreamsPerPeer)
		pool = p.pools[peerIDStr]
	}

	// Create new streams until we reach the minimum
	for len(pool) < p.minStreamsPerPeer {
		// Use the unified resilience function for stream creation
		var newStream types.VPNStream

		logger := log.WithFields(logrus.Fields{
			"peer_id":           peerIDStr,
			"operation":         "ensure_min_streams",
			"current_pool_size": len(pool),
			"min_size":          p.minStreamsPerPeer,
		})

		logger.Debug("Creating new stream to ensure minimum pool size")

		err, _ := p.resilienceService.ExecuteWithResilience(
			ctx,
			p.resilienceService.FormatPeerBreakerId(peerID, "ensure_min_streams"),
			func() error {
				// Create a new stream
				var createErr error
				newStream, createErr = p.streamService.CreateNewVPNStream(ctx, peerID)
				if createErr != nil {
					logger.WithError(createErr).Warn("Failed to create new stream, will retry")
					return createErr // Return the error to trigger retry
				}
				logger.Debug("Successfully created new stream")
				return nil // Success
			},
		)

		if err != nil {
			logger.WithError(err).Error("Failed to create new stream after retries")
			return
		}

		// Add the stream to the pool
		ps := &pooledStream{
			stream:   newStream,
			lastUsed: time.Now(),
			inUse:    false,
		}
		p.pools[peerIDStr] = append(pool, ps)
		pool = p.pools[peerIDStr]

		// Update metrics
		p.metrics.IncrementStreamsCreated()
	}
}

// CleanupIdleStreams removes idle streams that have exceeded the idle timeout
func (p *StreamPool) CleanupIdleStreams() {
	p.mu.Lock()
	defer p.mu.Unlock()

	logger := log.WithFields(logrus.Fields{
		"operation": "cleanup_idle_streams",
		"timeout":   p.streamIdleTimeout,
		"pool_size": len(p.pools),
	})

	logger.Debug("Starting cleanup of idle streams")
	now := time.Now()

	// Check each peer's pool
	for peerIDStr, pool := range p.pools {
		peerLogger := logger.WithFields(logrus.Fields{
			"peer_id":   peerIDStr,
			"pool_size": len(pool),
		})

		// Keep track of which streams to remove
		toRemove := make([]int, 0)

		// Check each stream in the pool
		for i, ps := range pool {
			// Skip streams that are in use
			if ps.inUse {
				continue
			}

			// Check if the stream has been idle for too long
			idleTime := now.Sub(ps.lastUsed)
			if idleTime > p.streamIdleTimeout {
				streamLogger := peerLogger.WithFields(logrus.Fields{
					"stream_idx": i,
					"idle_time":  idleTime,
				})
				streamLogger.Debug("Closing idle stream")

				// Close the stream
				if err := ps.stream.Close(); err != nil {
					streamLogger.WithError(err).Warn("Error closing idle stream")
				}

				// Mark the stream for removal
				toRemove = append(toRemove, i)

				// Update metrics
				p.metrics.IncrementStreamsClosed()
			}
		}

		// Remove the streams (in reverse order to avoid index issues)
		if len(toRemove) > 0 {
			peerLogger.WithField("streams_to_remove", len(toRemove)).Debug("Removing idle streams")
			for i := len(toRemove) - 1; i >= 0; i-- {
				idx := toRemove[i]
				pool = slices.Delete(pool, idx, idx+1)
			}
		}

		// Update the pool
		if len(pool) == 0 {
			// Remove the peer if there are no streams
			peerLogger.Debug("Removing peer from pool as it has no streams left")
			delete(p.pools, peerIDStr)
		} else {
			p.pools[peerIDStr] = pool
		}
	}

	logger.WithField("remaining_peers", len(p.pools)).Debug("Completed cleanup of idle streams")
}

// GetMetrics returns the current metrics
func (p *StreamPool) GetMetrics() map[string]int64 {
	return p.metrics.GetMetrics()
}

// GetStreamCount returns the number of streams for a peer
func (p *StreamPool) GetStreamCount(peerID peer.ID) int {
	peerIDStr := peerID.String()

	p.mu.RLock()
	defer p.mu.RUnlock()

	// Check if we have a pool for this peer
	pool, exists := p.pools[peerIDStr]
	if !exists {
		return 0
	}

	return len(pool)
}

// GetActiveStreamCount returns the number of active streams for a peer
func (p *StreamPool) GetActiveStreamCount(peerID peer.ID) int {
	peerIDStr := peerID.String()

	p.mu.RLock()
	defer p.mu.RUnlock()

	// Check if we have a pool for this peer
	pool, exists := p.pools[peerIDStr]
	if !exists {
		return 0
	}

	// Count active streams
	count := 0
	for _, ps := range pool {
		if ps.inUse {
			count++
		}
	}

	return count
}

// GetAllPeers returns a list of all peers in the pool
func (p *StreamPool) GetAllPeers() []peer.ID {
	p.mu.RLock()
	defer p.mu.RUnlock()

	logger := log.WithFields(logrus.Fields{
		"operation": "get_all_peers",
		"pool_size": len(p.pools),
	})

	logger.Debug("Retrieving all peers from pool")

	// Create a list of peer IDs
	peers := make([]peer.ID, 0, len(p.pools))

	// Add each peer ID to the list
	for peerIDStr := range p.pools {
		peerLogger := logger.WithField("peer_id_str", peerIDStr)
		peerID, err := peer.Decode(peerIDStr)
		if err != nil {
			peerLogger.WithError(err).Error("Failed to decode peer ID")
			continue
		}

		peers = append(peers, peerID)
	}

	logger.WithField("peer_count", len(peers)).Debug("Retrieved all peers from pool")
	return peers
}

// Close closes all streams in the pool
func (p *StreamPool) Close() error {
	// Cancel the context
	p.cancel()

	p.mu.Lock()
	defer p.mu.Unlock()

	logger := log.WithFields(logrus.Fields{
		"operation": "close_pool",
		"pool_size": len(p.pools),
	})

	logger.Info("Closing all streams in the pool")

	// Close all streams
	for peerIDStr, pool := range p.pools {
		peerLogger := logger.WithFields(logrus.Fields{
			"peer_id":   peerIDStr,
			"pool_size": len(pool),
		})

		peerLogger.Debug("Closing all streams for peer")
		for i, ps := range pool {
			streamLogger := peerLogger.WithField("stream_idx", i)
			if err := ps.stream.Close(); err != nil {
				streamLogger.WithError(err).Warn("Error closing stream")
			} else {
				streamLogger.Debug("Stream closed successfully")
			}

			// Update metrics
			p.metrics.IncrementStreamsClosed()
		}

		// Remove the peer
		peerLogger.Debug("Removed peer from pool")
		delete(p.pools, peerIDStr)
	}

	logger.Info("All streams closed and pool emptied")
	return nil
}
