package pool

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
)

var managerV2Log = logrus.WithField("service", "stream-manager-v2")

// StreamManagerV2 is an improved implementation of StreamManager with reduced mutex usage
type StreamManagerV2 struct {
	// Core components
	streamPool StreamPoolInterfaceV2
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

// connectionShard represents a shard of the connection map
type connectionShard struct {
	sync.RWMutex
	// Map of connection key -> connection state
	connections map[types.ConnectionKey]*ConnectionStateV2
}

// NewStreamManagerV2 creates a new stream manager with reduced mutex usage
func NewStreamManagerV2(streamPool StreamPoolInterfaceV2) *StreamManagerV2 {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize the connection shards
	connectionShards := [16]*connectionShard{}
	for i := 0; i < 16; i++ {
		connectionShards[i] = &connectionShard{
			connections: make(map[types.ConnectionKey]*ConnectionStateV2),
		}
	}

	return &StreamManagerV2{
		streamPool:       streamPool,
		ctx:              ctx,
		cancel:           cancel,
		connectionShards: connectionShards,
	}
}

// Start starts the stream manager
func (m *StreamManagerV2) Start() {
	managerV2Log.Info("Starting stream manager")
}

// Stop stops the stream manager
func (m *StreamManagerV2) Stop() {
	managerV2Log.Info("Stopping stream manager")
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
func (m *StreamManagerV2) GetOrCreateStreamForConnection(
	ctx context.Context,
	connKey types.ConnectionKey,
	peerID peer.ID,
) (*StreamChannelV2, error) {
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
		newConn := &ConnectionStateV2{
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
func (m *StreamManagerV2) ReleaseConnection(connKey types.ConnectionKey, healthy bool) {
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
func (m *StreamManagerV2) SendPacket(
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
			managerV2Log.WithFields(logrus.Fields{
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
func (m *StreamManagerV2) GetMetrics() map[string]int64 {
	return map[string]int64{
		"streams_created":   atomic.LoadInt64(&m.metrics.StreamsCreated),
		"streams_released":  atomic.LoadInt64(&m.metrics.StreamsReleased),
		"packets_processed": atomic.LoadInt64(&m.metrics.PacketsProcessed),
		"errors":            atomic.LoadInt64(&m.metrics.Errors),
	}
}

// GetConnectionCount returns the number of active connections
func (m *StreamManagerV2) GetConnectionCount() int {
	count := 0
	for _, shard := range m.connectionShards {
		shard.RLock()
		count += len(shard.connections)
		shard.RUnlock()
	}
	return count
}

// getConnectionShardIndex returns the shard index for a connection key
func (m *StreamManagerV2) getConnectionShardIndex(connKey types.ConnectionKey) int {
	// Simple hash function to distribute connections across shards
	var hash uint32
	for i := 0; i < len(connKey); i++ {
		hash = hash*31 + uint32(connKey[i])
	}
	return int(hash % 16)
}

// Close implements io.Closer
func (m *StreamManagerV2) Close() error {
	m.Stop()
	return nil
}
