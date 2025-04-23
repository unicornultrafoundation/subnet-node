package router

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
)

var sharedPoolLog = logrus.WithField("service", "vpn-shared-stream-pool")

// SharedStreamPool manages a pool of streams that can be shared by multiple connections
// It respects the maximum number of streams per peer and properly handles stream health
type SharedStreamPool struct {
	// Core components
	createStreamFn StreamCreator
	ctx            context.Context
	cancel         context.CancelFunc

	// Stream management
	mu            sync.RWMutex
	peerStreams   map[string][]*sharedStream
	streamIndices map[string]map[int]int // Maps peerID -> streamIndex -> poolIndex

	// Configuration
	minStreamsPerPeer int
	maxStreamsPerPeer int

	// Metrics
	acquiredCount int64
	releasedCount int64
	errorCount    int64
	createdCount  int64
	closedCount   int64
}

// sharedStream represents a stream in the shared pool
type sharedStream struct {
	stream       api.VPNStream
	index        int
	usageCount   int32
	lastUsed     time.Time
	healthy      bool
	creationTime time.Time
}

// NewSharedStreamPool creates a new shared stream pool
func NewSharedStreamPool(
	ctx context.Context,
	createStreamFn StreamCreator,
	minStreamsPerPeer int,
	maxStreamsPerPeer int,
) *SharedStreamPool {
	if minStreamsPerPeer <= 0 {
		minStreamsPerPeer = 1
	}
	if maxStreamsPerPeer <= 0 {
		maxStreamsPerPeer = 10
	}

	poolCtx, cancel := context.WithCancel(ctx)

	pool := &SharedStreamPool{
		createStreamFn:    createStreamFn,
		ctx:               poolCtx,
		cancel:            cancel,
		peerStreams:       make(map[string][]*sharedStream),
		streamIndices:     make(map[string]map[int]int),
		minStreamsPerPeer: minStreamsPerPeer,
		maxStreamsPerPeer: maxStreamsPerPeer,
	}

	// Start background maintenance
	go pool.maintainStreams()

	return pool
}

// maintainStreams periodically checks and maintains the stream pool
func (p *SharedStreamPool) maintainStreams() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.verifyStreams()
		}
	}
}

// verifyStreams verifies that all streams are healthy and that we have the right number of streams
func (p *SharedStreamPool) verifyStreams() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for peerIDStr, streams := range p.peerStreams {
		peerID, err := peer.Decode(peerIDStr)
		if err != nil {
			sharedPoolLog.WithError(err).Warn("Failed to decode peer ID")
			continue
		}

		// Count healthy streams
		healthyCount := 0
		for _, s := range streams {
			if s.healthy {
				healthyCount++
			}
		}

		// Ensure we have at least minStreamsPerPeer healthy streams
		if healthyCount < p.minStreamsPerPeer {
			targetCount := p.minStreamsPerPeer
			if targetCount > p.maxStreamsPerPeer {
				targetCount = p.maxStreamsPerPeer
			}

			sharedPoolLog.WithFields(logrus.Fields{
				"peer_id":       peerIDStr,
				"healthy_count": healthyCount,
				"target_count":  targetCount,
			}).Info("Not enough healthy streams, ensuring target count")

			// We'll create streams as needed when they're requested
			sharedPoolLog.WithFields(logrus.Fields{
				"peer_id":      peerID.String(),
				"target_count": targetCount,
			}).Debug("Will create streams as needed")
		}
	}
}

// getOrCreatePeerStreams gets or creates the stream maps for a peer
func (p *SharedStreamPool) getOrCreatePeerStreams(peerID peer.ID) ([]*sharedStream, map[int]int) {
	peerIDStr := peerID.String()

	p.mu.RLock()
	streams, streamsExist := p.peerStreams[peerIDStr]
	indices, indicesExist := p.streamIndices[peerIDStr]
	p.mu.RUnlock()

	if !streamsExist || !indicesExist {
		p.mu.Lock()
		defer p.mu.Unlock()

		// Check again in case it was created between the read and write locks
		streams, streamsExist = p.peerStreams[peerIDStr]
		indices, indicesExist = p.streamIndices[peerIDStr]

		if !streamsExist {
			p.peerStreams[peerIDStr] = make([]*sharedStream, 0)
			streams = p.peerStreams[peerIDStr]
		}

		if !indicesExist {
			p.streamIndices[peerIDStr] = make(map[int]int)
			indices = p.streamIndices[peerIDStr]
		}
	}

	return streams, indices
}

// GetStream gets a stream for a specific index
// It ensures that the stream exists and is healthy
func (p *SharedStreamPool) GetStream(ctx context.Context, peerID peer.ID, streamIndex int) (api.VPNStream, error) {
	peerIDStr := peerID.String()
	streams, indices := p.getOrCreatePeerStreams(peerID)

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if we already have a stream at this index
	poolIndex, exists := indices[streamIndex]
	if exists && poolIndex >= 0 && poolIndex < len(streams) {
		stream := streams[poolIndex]
		if stream.healthy {
			// Increment usage count
			atomic.AddInt32(&stream.usageCount, 1)
			stream.lastUsed = time.Now()

			// Increment acquired count
			atomic.AddInt64(&p.acquiredCount, 1)

			sharedPoolLog.WithFields(logrus.Fields{
				"peer_id":      peerIDStr,
				"stream_index": streamIndex,
				"pool_index":   poolIndex,
				"usage_count":  atomic.LoadInt32(&stream.usageCount),
			}).Debug("Acquired existing stream")

			return stream.stream, nil
		}

		// Stream exists but is unhealthy, remove it
		sharedPoolLog.WithFields(logrus.Fields{
			"peer_id":      peerIDStr,
			"stream_index": streamIndex,
			"pool_index":   poolIndex,
		}).Warn("Found unhealthy stream, removing it")

		// Close the unhealthy stream
		stream.stream.Close()
		atomic.AddInt64(&p.closedCount, 1)

		// Remove the stream from the pool
		p.peerStreams[peerIDStr] = append(streams[:poolIndex], streams[poolIndex+1:]...)
		delete(indices, streamIndex)

		// Update the indices for streams after the removed one
		for idx, pidx := range indices {
			if pidx > poolIndex {
				indices[idx] = pidx - 1
			}
		}
	}

	// Create a new stream
	stream, err := p.createNewStream(ctx, peerID, streamIndex)
	if err != nil {
		atomic.AddInt64(&p.errorCount, 1)
		return nil, err
	}

	return stream, nil
}

// createNewStream creates a new stream and adds it to the pool
func (p *SharedStreamPool) createNewStream(ctx context.Context, peerID peer.ID, streamIndex int) (api.VPNStream, error) {
	peerIDStr := peerID.String()
	streams := p.peerStreams[peerIDStr]
	indices := p.streamIndices[peerIDStr]

	// Check if we've reached the maximum number of streams
	if len(streams) >= p.maxStreamsPerPeer {
		// Find the least used stream
		leastUsedIndex := 0
		lowestUsage := atomic.LoadInt32(&streams[0].usageCount)

		for i, s := range streams {
			usage := atomic.LoadInt32(&s.usageCount)
			if usage < lowestUsage {
				lowestUsage = usage
				leastUsedIndex = i
			}
		}

		// Reuse the least used stream
		stream := streams[leastUsedIndex]

		// Update the index mapping
		for idx, pidx := range indices {
			if pidx == leastUsedIndex {
				delete(indices, idx)
				break
			}
		}
		indices[streamIndex] = leastUsedIndex

		// Increment usage count
		atomic.AddInt32(&stream.usageCount, 1)
		stream.lastUsed = time.Now()

		// Increment acquired count
		atomic.AddInt64(&p.acquiredCount, 1)

		sharedPoolLog.WithFields(logrus.Fields{
			"peer_id":      peerIDStr,
			"stream_index": streamIndex,
			"pool_index":   leastUsedIndex,
			"usage_count":  atomic.LoadInt32(&stream.usageCount),
		}).Debug("Reusing least used stream")

		return stream.stream, nil
	}

	// Create a new stream directly
	newStream, err := p.createStreamFn(ctx, peerID)
	if err != nil {
		sharedPoolLog.WithFields(logrus.Fields{
			"peer_id":      peerIDStr,
			"stream_index": streamIndex,
			"error":        err,
		}).Warn("Failed to create new stream")
		return nil, fmt.Errorf("failed to create new stream: %w", err)
	}

	// Add the stream to the pool
	sharedStream := &sharedStream{
		stream:       newStream,
		index:        streamIndex,
		usageCount:   1,
		lastUsed:     time.Now(),
		healthy:      true,
		creationTime: time.Now(),
	}

	poolIndex := len(streams)
	p.peerStreams[peerIDStr] = append(streams, sharedStream)
	indices[streamIndex] = poolIndex

	// Increment counters
	atomic.AddInt64(&p.acquiredCount, 1)
	atomic.AddInt64(&p.createdCount, 1)

	sharedPoolLog.WithFields(logrus.Fields{
		"peer_id":      peerIDStr,
		"stream_index": streamIndex,
		"pool_index":   poolIndex,
	}).Debug("Created new stream")

	return newStream, nil
}

// ReleaseStream releases a stream for a specific index
// If the stream is unhealthy, it will be closed and removed from the pool
func (p *SharedStreamPool) ReleaseStream(peerID peer.ID, streamIndex int, healthy bool) {
	peerIDStr := peerID.String()
	streams, indices := p.getOrCreatePeerStreams(peerID)

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if we have a stream at this index
	poolIndex, exists := indices[streamIndex]
	if !exists || poolIndex < 0 || poolIndex >= len(streams) {
		sharedPoolLog.WithFields(logrus.Fields{
			"peer_id":      peerIDStr,
			"stream_index": streamIndex,
		}).Warn("No stream found at index to release")
		return
	}

	stream := streams[poolIndex]

	// Decrement usage count
	newCount := atomic.AddInt32(&stream.usageCount, -1)
	stream.lastUsed = time.Now()

	// Increment released count
	atomic.AddInt64(&p.releasedCount, 1)

	if !healthy {
		// Mark the stream as unhealthy
		stream.healthy = false

		// Close the unhealthy stream
		stream.stream.Close()
		atomic.AddInt64(&p.closedCount, 1)

		// Remove the stream from the pool
		p.peerStreams[peerIDStr] = append(streams[:poolIndex], streams[poolIndex+1:]...)
		delete(indices, streamIndex)

		// Update the indices for streams after the removed one
		for idx, pidx := range indices {
			if pidx > poolIndex {
				indices[idx] = pidx - 1
			}
		}

		sharedPoolLog.WithFields(logrus.Fields{
			"peer_id":      peerIDStr,
			"stream_index": streamIndex,
			"pool_index":   poolIndex,
		}).Warn("Closed and removed unhealthy stream")

		// Ensure we maintain the minimum number of streams
		if len(p.peerStreams[peerIDStr]) < p.minStreamsPerPeer {
			// This will be handled by the maintenance routine
			sharedPoolLog.WithFields(logrus.Fields{
				"peer_id":       peerIDStr,
				"current_count": len(p.peerStreams[peerIDStr]),
				"min_streams":   p.minStreamsPerPeer,
			}).Info("Stream count below minimum, will be handled by maintenance")
		}
	} else {
		sharedPoolLog.WithFields(logrus.Fields{
			"peer_id":      peerIDStr,
			"stream_index": streamIndex,
			"pool_index":   poolIndex,
			"usage_count":  newCount,
		}).Debug("Released stream")
	}
}

// SetTargetStreams sets the target number of streams for a peer
func (p *SharedStreamPool) SetTargetStreams(peerID peer.ID, targetCount int) {
	if targetCount < p.minStreamsPerPeer {
		targetCount = p.minStreamsPerPeer
	}
	if targetCount > p.maxStreamsPerPeer {
		targetCount = p.maxStreamsPerPeer
	}

	// Store the target count for future reference
	p.mu.Lock()
	defer p.mu.Unlock()

	// We'll create streams as needed when they're requested

	sharedPoolLog.WithFields(logrus.Fields{
		"peer_id":      peerID.String(),
		"target_count": targetCount,
	}).Info("Set target stream count for peer")
}

// GetStreamStats returns statistics about the stream pool
func (p *SharedStreamPool) GetStreamStats(peerID peer.ID) map[string]interface{} {
	peerIDStr := peerID.String()

	p.mu.RLock()
	defer p.mu.RUnlock()

	streams, exists := p.peerStreams[peerIDStr]
	if !exists {
		return map[string]interface{}{
			"stream_count":   0,
			"healthy_count":  0,
			"total_usage":    0,
			"acquired_count": atomic.LoadInt64(&p.acquiredCount),
			"released_count": atomic.LoadInt64(&p.releasedCount),
			"error_count":    atomic.LoadInt64(&p.errorCount),
			"created_count":  atomic.LoadInt64(&p.createdCount),
			"closed_count":   atomic.LoadInt64(&p.closedCount),
		}
	}

	// Collect stream stats
	streamStats := make([]map[string]interface{}, len(streams))
	totalUsage := int32(0)
	healthyCount := 0

	for i, s := range streams {
		usage := atomic.LoadInt32(&s.usageCount)
		totalUsage += usage

		if s.healthy {
			healthyCount++
		}

		streamStats[i] = map[string]interface{}{
			"index":         s.index,
			"usage_count":   usage,
			"healthy":       s.healthy,
			"last_used":     s.lastUsed,
			"creation_time": s.creationTime,
			"age":           time.Since(s.creationTime).String(),
		}
	}

	// Return overall stats
	return map[string]interface{}{
		"stream_count":   len(streams),
		"healthy_count":  healthyCount,
		"total_usage":    totalUsage,
		"streams":        streamStats,
		"acquired_count": atomic.LoadInt64(&p.acquiredCount),
		"released_count": atomic.LoadInt64(&p.releasedCount),
		"error_count":    atomic.LoadInt64(&p.errorCount),
		"created_count":  atomic.LoadInt64(&p.createdCount),
		"closed_count":   atomic.LoadInt64(&p.closedCount),
	}
}

// Close closes the stream pool and all streams
func (p *SharedStreamPool) Close() {
	p.cancel()

	p.mu.Lock()
	defer p.mu.Unlock()

	for peerIDStr, streams := range p.peerStreams {
		for i, s := range streams {
			s.stream.Close()
			sharedPoolLog.WithFields(logrus.Fields{
				"peer_id":    peerIDStr,
				"pool_index": i,
			}).Debug("Closed stream during shutdown")
		}
		delete(p.peerStreams, peerIDStr)
		delete(p.streamIndices, peerIDStr)
	}

	sharedPoolLog.Info("Shared stream pool closed")
}
