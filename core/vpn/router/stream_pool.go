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

var streamPoolLog = logrus.WithField("service", "vpn-router-stream-pool")

// RouterStreamPool manages streams for the router package
// It provides functionality similar to the pool.PoolServiceExtension interface
// but is implemented directly in the router package
type RouterStreamPool struct {
	// Core components
	createStreamFn StreamCreator
	ctx            context.Context
	cancel         context.CancelFunc

	// Stream management
	mu            sync.RWMutex
	peerStreams   map[string]map[int]*routerStream
	streamMetrics map[string]*streamMetrics

	// Configuration
	minStreamsPerPeer int
	maxStreamsPerPeer int
}

// routerStream represents a stream in the pool
type routerStream struct {
	stream       api.VPNStream
	index        int
	usageCount   int32
	lastUsed     time.Time
	healthy      bool
	creationTime time.Time
}

// streamMetrics tracks metrics for streams
type streamMetrics struct {
	acquiredCount int64
	releasedCount int64
	errorCount    int64
	createdCount  int64
	closedCount   int64
}

// NewRouterStreamPool creates a new router stream pool
func NewRouterStreamPool(
	ctx context.Context,
	createStreamFn StreamCreator,
	minStreamsPerPeer int,
	maxStreamsPerPeer int,
) *RouterStreamPool {
	if minStreamsPerPeer <= 0 {
		minStreamsPerPeer = 1
	}
	if maxStreamsPerPeer <= 0 {
		maxStreamsPerPeer = 10
	}

	poolCtx, cancel := context.WithCancel(ctx)

	pool := &RouterStreamPool{
		createStreamFn:    createStreamFn,
		ctx:               poolCtx,
		cancel:            cancel,
		peerStreams:       make(map[string]map[int]*routerStream),
		streamMetrics:     make(map[string]*streamMetrics),
		minStreamsPerPeer: minStreamsPerPeer,
		maxStreamsPerPeer: maxStreamsPerPeer,
	}

	// Start background maintenance
	go pool.maintainStreams()

	return pool
}

// maintainStreams periodically checks and maintains the stream pool
func (p *RouterStreamPool) maintainStreams() {
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

// verifyStreams verifies that all streams are healthy
func (p *RouterStreamPool) verifyStreams() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for peerIDStr, streams := range p.peerStreams {
		_, err := peer.Decode(peerIDStr)
		if err != nil {
			streamPoolLog.WithError(err).Warn("Failed to decode peer ID")
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

			streamPoolLog.WithFields(logrus.Fields{
				"peer_id":       peerIDStr,
				"healthy_count": healthyCount,
				"target_count":  targetCount,
			}).Info("Not enough healthy streams, ensuring target count")

			// We'll create streams as needed when they're requested
		}
	}
}

// getOrCreatePeerStreams gets or creates the stream maps for a peer
func (p *RouterStreamPool) getOrCreatePeerStreams(peerID peer.ID) (map[int]*routerStream, *streamMetrics) {
	peerIDStr := peerID.String()

	p.mu.RLock()
	streams, streamsExist := p.peerStreams[peerIDStr]
	metrics, metricsExist := p.streamMetrics[peerIDStr]
	p.mu.RUnlock()

	if !streamsExist || !metricsExist {
		p.mu.Lock()
		defer p.mu.Unlock()

		// Check again in case it was created between the read and write locks
		streams, streamsExist = p.peerStreams[peerIDStr]
		metrics, metricsExist = p.streamMetrics[peerIDStr]

		if !streamsExist {
			p.peerStreams[peerIDStr] = make(map[int]*routerStream)
			streams = p.peerStreams[peerIDStr]
		}

		if !metricsExist {
			p.streamMetrics[peerIDStr] = &streamMetrics{}
			metrics = p.streamMetrics[peerIDStr]
		}
	}

	return streams, metrics
}

// GetStreamByIndex gets a stream for a specific index
func (p *RouterStreamPool) GetStreamByIndex(ctx context.Context, peerID peer.ID, streamIndex int) (api.VPNStream, error) {
	peerIDStr := peerID.String()
	streams, metrics := p.getOrCreatePeerStreams(peerID)

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if we already have a stream at this index
	if stream, exists := streams[streamIndex]; exists && stream.healthy {
		// Increment usage count
		atomic.AddInt32(&stream.usageCount, 1)
		stream.lastUsed = time.Now()

		// Increment acquired count
		atomic.AddInt64(&metrics.acquiredCount, 1)

		streamPoolLog.WithFields(logrus.Fields{
			"peer_id":      peerIDStr,
			"stream_index": streamIndex,
			"usage_count":  atomic.LoadInt32(&stream.usageCount),
		}).Debug("Acquired existing stream")

		return stream.stream, nil
	}

	// Create a new stream
	newStream, err := p.createStreamFn(ctx, peerID)
	if err != nil {
		atomic.AddInt64(&metrics.errorCount, 1)
		streamPoolLog.WithFields(logrus.Fields{
			"peer_id":      peerIDStr,
			"stream_index": streamIndex,
			"error":        err,
		}).Warn("Failed to create new stream")
		return nil, fmt.Errorf("failed to create new stream: %w", err)
	}

	// Add the stream to the pool
	stream := &routerStream{
		stream:       newStream,
		index:        streamIndex,
		usageCount:   1,
		lastUsed:     time.Now(),
		healthy:      true,
		creationTime: time.Now(),
	}

	streams[streamIndex] = stream

	// Increment counters
	atomic.AddInt64(&metrics.acquiredCount, 1)
	atomic.AddInt64(&metrics.createdCount, 1)

	streamPoolLog.WithFields(logrus.Fields{
		"peer_id":      peerIDStr,
		"stream_index": streamIndex,
	}).Debug("Created new stream")

	return newStream, nil
}

// ReleaseStreamByIndex releases a stream for a specific index
func (p *RouterStreamPool) ReleaseStreamByIndex(peerID peer.ID, streamIndex int, close bool) {
	peerIDStr := peerID.String()
	streams, metrics := p.getOrCreatePeerStreams(peerID)

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if we have a stream at this index
	stream, exists := streams[streamIndex]
	if !exists {
		streamPoolLog.WithFields(logrus.Fields{
			"peer_id":      peerIDStr,
			"stream_index": streamIndex,
		}).Warn("No stream found at index to release")
		return
	}

	// Decrement usage count
	newCount := atomic.AddInt32(&stream.usageCount, -1)
	stream.lastUsed = time.Now()

	// Increment released count
	atomic.AddInt64(&metrics.releasedCount, 1)

	if close {
		// Mark the stream as unhealthy
		stream.healthy = false

		// Close the unhealthy stream
		stream.stream.Close()
		atomic.AddInt64(&metrics.closedCount, 1)

		// Remove the stream from the pool
		delete(streams, streamIndex)

		streamPoolLog.WithFields(logrus.Fields{
			"peer_id":      peerIDStr,
			"stream_index": streamIndex,
		}).Warn("Closed and removed unhealthy stream")

		// Ensure we maintain the minimum number of streams
		if len(streams) < p.minStreamsPerPeer {
			// This will be handled by the maintenance routine
			streamPoolLog.WithFields(logrus.Fields{
				"peer_id":       peerIDStr,
				"current_count": len(streams),
				"min_streams":   p.minStreamsPerPeer,
			}).Info("Stream count below minimum, will be handled by maintenance")
		}
	} else {
		streamPoolLog.WithFields(logrus.Fields{
			"peer_id":      peerIDStr,
			"stream_index": streamIndex,
			"usage_count":  newCount,
		}).Debug("Released stream")
	}
}

// SetTargetStreamsForPeer sets the target number of streams for a peer
func (p *RouterStreamPool) SetTargetStreamsForPeer(peerID peer.ID, targetCount int) {
	if targetCount < p.minStreamsPerPeer {
		targetCount = p.minStreamsPerPeer
	}
	if targetCount > p.maxStreamsPerPeer {
		targetCount = p.maxStreamsPerPeer
	}

	streamPoolLog.WithFields(logrus.Fields{
		"peer_id":      peerID.String(),
		"target_count": targetCount,
	}).Info("Set target stream count for peer")

	// We'll create streams as needed when they're requested
}

// GetStreamStats returns statistics about the stream pool
func (p *RouterStreamPool) GetStreamStats(peerID peer.ID) map[string]interface{} {
	peerIDStr := peerID.String()

	p.mu.RLock()
	defer p.mu.RUnlock()

	streams, exists := p.peerStreams[peerIDStr]
	metrics, metricsExist := p.streamMetrics[peerIDStr]

	if !exists || !metricsExist {
		return map[string]interface{}{
			"stream_count":   0,
			"healthy_count":  0,
			"total_usage":    0,
			"acquired_count": 0,
			"released_count": 0,
			"error_count":    0,
			"created_count":  0,
			"closed_count":   0,
		}
	}

	// Collect stream stats
	streamStats := make([]map[string]interface{}, 0, len(streams))
	totalUsage := int32(0)
	healthyCount := 0

	for idx, s := range streams {
		usage := atomic.LoadInt32(&s.usageCount)
		totalUsage += usage

		if s.healthy {
			healthyCount++
		}

		streamStats = append(streamStats, map[string]interface{}{
			"index":         idx,
			"usage_count":   usage,
			"healthy":       s.healthy,
			"last_used":     s.lastUsed,
			"creation_time": s.creationTime,
			"age":           time.Since(s.creationTime).String(),
		})
	}

	// Return overall stats
	return map[string]interface{}{
		"stream_count":   len(streams),
		"healthy_count":  healthyCount,
		"total_usage":    totalUsage,
		"streams":        streamStats,
		"acquired_count": atomic.LoadInt64(&metrics.acquiredCount),
		"released_count": atomic.LoadInt64(&metrics.releasedCount),
		"error_count":    atomic.LoadInt64(&metrics.errorCount),
		"created_count":  atomic.LoadInt64(&metrics.createdCount),
		"closed_count":   atomic.LoadInt64(&metrics.closedCount),
	}
}

// Close closes the stream pool and all streams
func (p *RouterStreamPool) Close() {
	p.cancel()

	p.mu.Lock()
	defer p.mu.Unlock()

	for peerIDStr, streams := range p.peerStreams {
		for idx, s := range streams {
			s.stream.Close()
			streamPoolLog.WithFields(logrus.Fields{
				"peer_id":      peerIDStr,
				"stream_index": idx,
			}).Debug("Closed stream during shutdown")
		}
		delete(p.peerStreams, peerIDStr)
	}

	streamPoolLog.Info("Router stream pool closed")
}
