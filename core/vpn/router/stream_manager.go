package router

import (
	"context"
	"hash/fnv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
)

var streamManagerLog = logrus.WithField("service", "vpn-stream-manager")

// StreamManager manages streams for the VPN router
// It handles stream acquisition, health tracking, and load balancing
type StreamManager struct {
	// Core components
	ctx context.Context

	// Stream management
	streamsMu          sync.RWMutex
	streamsPerPeer     map[peer.ID]int // Target number of streams per peer
	connectionToStream map[string]int  // Maps connection keys to stream indices

	// Stream pool for efficient stream management
	streamPool *RouterStreamPool

	// Configuration
	minStreamsPerPeer int
	maxStreamsPerPeer int

	// Metrics
	acquiredCount int64
	releasedCount int64
	errorCount    int64
}

// NewStreamManager creates a new stream manager
func NewStreamManager(
	ctx context.Context,
	createStreamFn StreamCreator,
	minStreamsPerPeer int,
	maxStreamsPerPeer int,
) *StreamManager {
	if minStreamsPerPeer <= 0 {
		minStreamsPerPeer = 1
	}
	if maxStreamsPerPeer <= 0 {
		maxStreamsPerPeer = 10
	}

	// Create the router stream pool
	streamPool := NewRouterStreamPool(ctx, createStreamFn, minStreamsPerPeer, maxStreamsPerPeer)

	return &StreamManager{
		ctx:                ctx,
		streamsPerPeer:     make(map[peer.ID]int),
		connectionToStream: make(map[string]int),
		streamPool:         streamPool,
		minStreamsPerPeer:  minStreamsPerPeer,
		maxStreamsPerPeer:  maxStreamsPerPeer,
	}
}

// GetStream gets a stream for a connection
// It ensures that packets with the same connection key always use the same stream
func (m *StreamManager) GetStream(peerID peer.ID, connectionKey string) (api.VPNStream, int, error) {
	// First check if we already have a stream index for this connection key using read lock
	m.streamsMu.RLock()
	streamIndex, exists := m.connectionToStream[connectionKey]
	m.streamsMu.RUnlock()

	// If this is a new connection, assign it to a stream
	if !exists {
		// Need to acquire write lock to update maps
		m.streamsMu.Lock()

		// Check again in case another goroutine assigned it while we were waiting for the lock
		streamIndex, exists = m.connectionToStream[connectionKey]
		if !exists {
			// Get target stream count for this peer
			targetCount := m.streamsPerPeer[peerID]
			if targetCount < m.minStreamsPerPeer {
				targetCount = m.minStreamsPerPeer
				m.streamsPerPeer[peerID] = targetCount
			}

			// Use consistent hash to determine stream index
			h := fnv.New32a()
			h.Write([]byte(connectionKey))
			streamIndex = int(h.Sum32() % uint32(targetCount))

			// Store the assignment for future packets with this connection key
			m.connectionToStream[connectionKey] = streamIndex

			streamManagerLog.WithFields(logrus.Fields{
				"peer_id":    peerID.String(),
				"index":      streamIndex,
				"connection": connectionKey,
			}).Debug("Assigned new connection to stream")
		}
		m.streamsMu.Unlock()
	}

	// Get the stream from the stream pool - do this outside the lock
	stream, err := m.streamPool.GetStreamByIndex(m.ctx, peerID, streamIndex)
	if err != nil {
		// If we can't get the stream, log the error
		streamManagerLog.WithFields(logrus.Fields{
			"peer_id": peerID.String(),
			"index":   streamIndex,
			"error":   err,
		}).Warn("Failed to get stream from shared pool")

		// Increment error count
		atomic.AddInt64(&m.errorCount, 1)

		return nil, -1, err
	}

	// Increment acquired count
	atomic.AddInt64(&m.acquiredCount, 1)

	streamManagerLog.WithFields(logrus.Fields{
		"peer_id":    peerID.String(),
		"index":      streamIndex,
		"connection": connectionKey,
	}).Debug("Acquired stream for connection")

	return stream, streamIndex, nil
}

// ReleaseStream releases a stream for a connection
// It decrements the usage count but doesn't close the stream unless it's unhealthy
func (m *StreamManager) ReleaseStream(peerID peer.ID, streamIndex int, connectionKey string, healthy bool) {
	// Only lock when we need to modify the map
	if connectionKey != "" {
		m.streamsMu.Lock()
		delete(m.connectionToStream, connectionKey)
		m.streamsMu.Unlock()

		streamManagerLog.WithFields(logrus.Fields{
			"connection": connectionKey,
			"index":      streamIndex,
		}).Debug("Removed connection-to-stream mapping")
	}

	// Release the stream in the stream pool - no need for lock here
	m.streamPool.ReleaseStreamByIndex(peerID, streamIndex, !healthy)

	// Log the release
	if !healthy {
		streamManagerLog.WithFields(logrus.Fields{
			"peer_id":    peerID.String(),
			"index":      streamIndex,
			"connection": connectionKey,
		}).Warn("Released unhealthy stream")
	} else {
		streamManagerLog.WithFields(logrus.Fields{
			"peer_id":    peerID.String(),
			"index":      streamIndex,
			"connection": connectionKey,
		}).Debug("Released stream for connection")
	}

	// Increment released count - atomic operation, no lock needed
	atomic.AddInt64(&m.releasedCount, 1)
}

// SetTargetStreamsForPeer sets the target number of streams for a peer
func (m *StreamManager) SetTargetStreamsForPeer(peerID peer.ID, targetCount int) {
	// Ensure target count is within bounds
	if targetCount < m.minStreamsPerPeer {
		targetCount = m.minStreamsPerPeer
	}
	if targetCount > m.maxStreamsPerPeer {
		targetCount = m.maxStreamsPerPeer
	}

	m.streamsMu.Lock()
	defer m.streamsMu.Unlock()

	// Update target count
	oldCount := m.streamsPerPeer[peerID]
	m.streamsPerPeer[peerID] = targetCount

	// Notify the stream pool
	m.streamPool.SetTargetStreamsForPeer(peerID, targetCount)

	streamManagerLog.WithFields(logrus.Fields{
		"peer_id":      peerID.String(),
		"old_count":    oldCount,
		"target_count": targetCount,
	}).Info("Updated target stream count for peer")
}

// VerifyStreamCounts verifies that our stream counts match reality
func (m *StreamManager) VerifyStreamCounts() {
	m.streamsMu.Lock()
	defer m.streamsMu.Unlock()

	// Check each peer's stream count
	for peerID, targetCount := range m.streamsPerPeer {
		// Ensure target count is within bounds
		if targetCount < m.minStreamsPerPeer {
			targetCount = m.minStreamsPerPeer
			m.streamsPerPeer[peerID] = targetCount
		}
		if targetCount > m.maxStreamsPerPeer {
			targetCount = m.maxStreamsPerPeer
			m.streamsPerPeer[peerID] = targetCount
		}

		// Notify the stream pool
		m.streamPool.SetTargetStreamsForPeer(peerID, targetCount)

		streamManagerLog.WithFields(logrus.Fields{
			"peer_id":      peerID.String(),
			"target_count": targetCount,
		}).Debug("Verified stream count for peer")
	}
}

// GetStreamStats returns statistics about stream usage
func (m *StreamManager) GetStreamStats(peerID peer.ID) map[string]interface{} {
	// Use read lock for accessing maps
	m.streamsMu.RLock()

	// Get target stream count
	targetCount := m.streamsPerPeer[peerID]
	if targetCount < m.minStreamsPerPeer {
		targetCount = m.minStreamsPerPeer
	}

	// Count connections per stream - do this while holding the read lock
	connectionsPerStream := make(map[int]int)
	totalConnections := 0
	for connKey, streamIdx := range m.connectionToStream {
		if strings.Contains(connKey, peerID.String()) {
			connectionsPerStream[streamIdx]++
			totalConnections++
		}
	}
	m.streamsMu.RUnlock()

	// Get stats from the stream pool - do this outside the lock
	streamPoolStats := m.streamPool.GetStreamStats(peerID)

	// Combine stats - no lock needed for atomic operations
	stats := map[string]interface{}{
		"target_count":           targetCount,
		"connections_per_stream": connectionsPerStream,
		"total_connections":      totalConnections,
		"acquired":               atomic.LoadInt64(&m.acquiredCount),
		"released":               atomic.LoadInt64(&m.releasedCount),
		"errors":                 atomic.LoadInt64(&m.errorCount),
	}

	// Add stream pool stats
	for k, v := range streamPoolStats {
		stats[k] = v
	}

	return stats
}
