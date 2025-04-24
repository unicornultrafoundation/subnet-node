package pool

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
)

var managerLog = logrus.WithField("service", "stream-manager")

// StreamManager is an improved implementation of StreamManager with reduced mutex usage
type StreamManager struct {
	// Core components
	streamPool StreamPoolInterface
	ctx        context.Context
	cancel     context.CancelFunc

	// Connection management
	connectionShards [16]*connectionShard

	// Metrics
	metrics struct {
		StreamsCreated   int64
		StreamsReleased  int64
		PacketsProcessed int64
		Errors           int64
	}
}

// ConnectionState represents the state of a connection
type ConnectionState struct {
	Key           types.ConnectionKey
	PeerID        peer.ID
	StreamChannel *StreamChannel
	LastActivity  int64
}

// connectionShard represents a shard of the connection map
type connectionShard struct {
	sync.RWMutex
	// Map of connection key -> connection state
	connections map[types.ConnectionKey]*ConnectionState
}

// NewStreamManager creates a new stream manager with reduced mutex usage
func NewStreamManager(streamPool StreamPoolInterface) *StreamManager {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize the connection shards
	connectionShards := [16]*connectionShard{}
	for i := 0; i < 16; i++ {
		connectionShards[i] = &connectionShard{
			connections: make(map[types.ConnectionKey]*ConnectionState),
		}
	}

	return &StreamManager{
		streamPool:       streamPool,
		ctx:              ctx,
		cancel:           cancel,
		connectionShards: connectionShards,
	}
}

// Start starts the stream manager
func (m *StreamManager) Start() {
	managerLog.Info("Starting stream manager")

	// Start the connection stats logger
	go m.connectionStatsLogger()
}

// Stop stops the stream manager
func (m *StreamManager) Stop() {
	managerLog.Info("Stopping stream manager")
	m.cancel()

	// Release all connections
	for _, shard := range m.connectionShards {
		shard.Lock()
		for connKey, conn := range shard.connections {
			if conn.StreamChannel != nil {
				m.streamPool.ReleaseStreamChannel(conn.PeerID, conn.StreamChannel, true)
			}
			delete(shard.connections, connKey)
		}
		shard.Unlock()
	}
}

// GetOrCreateStreamForConnection gets or creates a stream for a connection
func (m *StreamManager) GetOrCreateStreamForConnection(
	ctx context.Context,
	connKey types.ConnectionKey,
	peerID peer.ID,
) (*StreamChannel, error) {
	shardIdx := m.getConnectionShardIndex(connKey)
	shard := m.connectionShards[shardIdx]

	// Check if we already have a stream for this connection
	shard.RLock()
	conn, exists := shard.connections[connKey]
	if exists && conn.StreamChannel != nil {
		// Check if the stream is still healthy
		if conn.StreamChannel.IsHealthy() {
			stream := conn.StreamChannel
			shard.RUnlock()
			return stream, nil
		}

		// Stream is unhealthy, release it
		if conn.StreamChannel != nil {
			// We need to release the read lock before calling ReleaseStreamChannel
			// to avoid potential deadlocks
			shard.RUnlock()
			m.streamPool.ReleaseStreamChannel(conn.PeerID, conn.StreamChannel, false)

			// Re-acquire the lock for writing
			shard.Lock()
			// Check if the connection still exists
			conn, exists = shard.connections[connKey]
			if exists {
				conn.StreamChannel = nil
				shard.connections[connKey] = conn
			}
			shard.Unlock()
		} else {
			shard.RUnlock()
		}
	} else {
		shard.RUnlock()
	}

	// Get a stream from the pool
	streamChannel, err := m.streamPool.GetStreamChannel(ctx, peerID)
	if err != nil {
		return nil, err
	}

	// Create or update the connection state
	shard.Lock()
	if !exists {
		// Create a new connection state
		newConn := &ConnectionState{
			Key:           connKey,
			PeerID:        peerID,
			StreamChannel: streamChannel,
			LastActivity:  0, // Will be updated on first packet
		}
		shard.connections[connKey] = newConn
	} else {
		// Update the existing connection state
		conn.StreamChannel = streamChannel
		shard.connections[connKey] = conn
	}
	shard.Unlock()

	// Update metrics
	atomic.AddInt64(&m.metrics.StreamsCreated, 1)

	return streamChannel, nil
}

// ReleaseConnection releases a connection and its associated stream
func (m *StreamManager) ReleaseConnection(connKey types.ConnectionKey, healthy bool) {
	shardIdx := m.getConnectionShardIndex(connKey)
	shard := m.connectionShards[shardIdx]

	shard.Lock()
	conn, exists := shard.connections[connKey]
	if !exists || conn.StreamChannel == nil {
		shard.Unlock()
		return
	}

	// Get the stream channel and peer ID before releasing the lock
	streamChannel := conn.StreamChannel
	peerID := conn.PeerID

	// Clear the stream channel reference
	conn.StreamChannel = nil

	// Remove the connection if unhealthy
	if !healthy {
		delete(shard.connections, connKey)
	} else {
		// Update the connection state
		shard.connections[connKey] = conn
	}
	shard.Unlock()

	// Release the stream
	m.streamPool.ReleaseStreamChannel(peerID, streamChannel, healthy)

	// Update metrics
	atomic.AddInt64(&m.metrics.StreamsReleased, 1)
}

// SendPacket sends a packet through the appropriate stream
func (m *StreamManager) SendPacket(
	ctx context.Context,
	connKey types.ConnectionKey,
	peerID peer.ID,
	packet *types.QueuedPacket,
) error {
	// Get or create a stream for this connection
	streamChannel, err := m.GetOrCreateStreamForConnection(ctx, connKey, peerID)
	if err != nil {
		atomic.AddInt64(&m.metrics.Errors, 1)
		return err
	}

	// Add a recovery mechanism to prevent panics from closed channels
	defer func() {
		if r := recover(); r != nil {
			managerLog.WithFields(logrus.Fields{
				"conn_key": connKey,
				"peer_id":  peerID,
				"panic":    r,
			}).Error("Recovered from panic in SendPacket")

			// Mark the stream as unhealthy
			if streamChannel != nil {
				streamChannel.SetHealthy(false)
				m.ReleaseConnection(connKey, false)
			}

			// Update error metrics
			atomic.AddInt64(&m.metrics.Errors, 1)
		}
	}()

	// Check if the stream is still healthy before sending
	if !streamChannel.IsHealthy() {
		atomic.AddInt64(&m.metrics.Errors, 1)
		return types.ErrNoHealthyStreams
	}

	// Send the packet to the stream's channel with additional safety checks
	select {
	case streamChannel.PacketChan <- packet:
		// Packet sent successfully
		atomic.AddInt64(&m.metrics.PacketsProcessed, 1)
		return nil
	case <-ctx.Done():
		// Context cancelled
		atomic.AddInt64(&m.metrics.Errors, 1)
		return types.ErrContextCancelled
	case <-streamChannel.ctx.Done():
		// Stream context cancelled, stream is shutting down
		atomic.AddInt64(&m.metrics.Errors, 1)
		// Mark the stream as unhealthy
		streamChannel.SetHealthy(false)
		m.ReleaseConnection(connKey, false)
		return types.ErrStreamChannelClosed
	default:
		// Channel full, packet dropped
		atomic.AddInt64(&m.metrics.Errors, 1)
		return types.ErrStreamChannelFull
	}
}

// GetMetrics returns the manager's metrics
func (m *StreamManager) GetMetrics() map[string]int64 {
	return map[string]int64{
		"streams_created":   atomic.LoadInt64(&m.metrics.StreamsCreated),
		"streams_released":  atomic.LoadInt64(&m.metrics.StreamsReleased),
		"packets_processed": atomic.LoadInt64(&m.metrics.PacketsProcessed),
		"errors":            atomic.LoadInt64(&m.metrics.Errors),
	}
}

// GetConnectionCount returns the number of active connections
func (m *StreamManager) GetConnectionCount() int {
	count := 0
	for _, shard := range m.connectionShards {
		shard.RLock()
		count += len(shard.connections)
		shard.RUnlock()
	}
	return count
}

// getConnectionShardIndex returns the shard index for a connection key
func (m *StreamManager) getConnectionShardIndex(connKey types.ConnectionKey) int {
	// Simple hash function to distribute connections across shards
	var hash uint32
	for i := 0; i < len(connKey); i++ {
		hash = hash*31 + uint32(connKey[i])
	}
	return int(hash % 16)
}

// Close implements io.Closer
func (m *StreamManager) Close() error {
	m.Stop()
	return nil
}

// connectionStatsLogger periodically logs statistics about connection-to-stream mappings
func (m *StreamManager) connectionStatsLogger() {
	// Create a ticker for periodic logging (every 30 seconds)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			// Log connection stats
			m.logConnectionStats()
		}
	}
}

// logConnectionStats logs detailed statistics about connection-to-stream mappings
func (m *StreamManager) logConnectionStats() {
	// Log overall stats
	connCount := m.GetConnectionCount()
	managerLog.WithFields(logrus.Fields{
		"connection_count":  connCount,
		"streams_created":   atomic.LoadInt64(&m.metrics.StreamsCreated),
		"streams_released":  atomic.LoadInt64(&m.metrics.StreamsReleased),
		"packets_processed": atomic.LoadInt64(&m.metrics.PacketsProcessed),
		"errors":            atomic.LoadInt64(&m.metrics.Errors),
	}).Info("Stream manager statistics")

	// Log detailed connection-to-stream mappings
	for shardIdx, shard := range m.connectionShards {
		shard.RLock()

		// Skip empty shards
		if len(shard.connections) == 0 {
			shard.RUnlock()
			continue
		}

		// Create a list of connection details
		connDetails := make([]map[string]interface{}, 0, len(shard.connections))
		for connKey, conn := range shard.connections {
			// Create connection detail
			detail := map[string]interface{}{
				"conn_key": string(connKey),
				"peer_id":  conn.PeerID.String(),
			}

			// Add stream information if available
			if conn.StreamChannel != nil {
				detail["stream_id"] = fmt.Sprintf("%p", conn.StreamChannel)
				detail["stream_healthy"] = conn.StreamChannel.IsHealthy()
				detail["buffer_util"] = conn.StreamChannel.GetBufferUtilization()
				detail["packet_count"] = atomic.LoadInt64(&conn.StreamChannel.Metrics.PacketCount)
				detail["error_count"] = atomic.LoadInt64(&conn.StreamChannel.Metrics.ErrorCount)
			} else {
				detail["stream_id"] = "none"
			}

			connDetails = append(connDetails, detail)
		}
		shard.RUnlock()

		// Log shard connection stats
		managerLog.WithFields(logrus.Fields{
			"shard_index":      shardIdx,
			"connection_count": len(connDetails),
			"connections":      connDetails,
		}).Info("Connection-to-stream mappings")
	}
}
