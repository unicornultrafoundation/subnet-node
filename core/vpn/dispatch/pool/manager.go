package pool

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
)

var managerLog = logrus.WithField("service", "stream-manager")

// StreamManager manages stream assignments to connection keys
type StreamManager struct {
	// Core components
	streamPool StreamPoolInterface
	ctx        context.Context
	cancel     context.CancelFunc

	// Connection to stream mapping using sync.Map for concurrent access
	connections sync.Map // map[types.ConnectionKey]*ConnectionState

	// Metrics
	metrics struct {
		StreamsCreated   int64
		StreamsReleased  int64
		PacketsProcessed int64
		Errors           int64
	}
}

// ConnectionState tracks the state of a connection
type ConnectionState struct {
	// The connection key
	Key types.ConnectionKey
	// The peer ID
	PeerID peer.ID
	// The assigned stream channel
	StreamChannel *StreamChannel
	// Last activity time
	LastActivity int64 // Unix timestamp
}

// NewStreamManager creates a new stream manager
func NewStreamManager(streamPool StreamPoolInterface) *StreamManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &StreamManager{
		streamPool: streamPool,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start starts the stream manager
func (m *StreamManager) Start() {
	managerLog.Info("Starting stream manager")
}

// Stop stops the stream manager
func (m *StreamManager) Stop() {
	managerLog.Info("Stopping stream manager")
	m.cancel()

	// Release all connections
	m.connections.Range(func(key, value interface{}) bool {
		connKey := key.(types.ConnectionKey)
		conn := value.(*ConnectionState)

		if conn.StreamChannel != nil {
			m.streamPool.ReleaseStreamChannel(conn.PeerID, conn.StreamChannel, true)
		}
		m.connections.Delete(connKey)
		return true
	})
}

// GetOrCreateStreamForConnection gets or creates a stream for a connection
func (m *StreamManager) GetOrCreateStreamForConnection(
	ctx context.Context,
	connKey types.ConnectionKey,
	peerID peer.ID,
) (*StreamChannel, error) {
	// Check if we already have a stream for this connection
	value, exists := m.connections.Load(connKey)

	if exists {
		conn := value.(*ConnectionState)
		if conn.StreamChannel != nil {
			// Check if the stream is still healthy
			healthy := atomic.LoadInt32(&conn.StreamChannel.healthy) == 1

			if healthy {
				return conn.StreamChannel, nil
			}

			// Stream is unhealthy, release it
			if conn.StreamChannel != nil {
				m.streamPool.ReleaseStreamChannel(conn.PeerID, conn.StreamChannel, false)
				conn.StreamChannel = nil
			}
		}
	}

	// Get a stream from the pool
	streamChannel, err := m.streamPool.GetStreamChannel(ctx, peerID)
	if err != nil {
		return nil, err
	}

	// Create or update the connection state
	if !exists {
		// Create a new connection state
		newConn := &ConnectionState{
			Key:           connKey,
			PeerID:        peerID,
			StreamChannel: streamChannel,
			LastActivity:  0, // Will be updated on first packet
		}
		m.connections.Store(connKey, newConn)
	} else {
		// Update the existing connection state
		conn := value.(*ConnectionState)
		conn.StreamChannel = streamChannel
		m.connections.Store(connKey, conn)
	}

	// Update metrics
	atomic.AddInt64(&m.metrics.StreamsCreated, 1)

	return streamChannel, nil
}

// ReleaseConnection releases a connection and its associated stream
func (m *StreamManager) ReleaseConnection(connKey types.ConnectionKey, healthy bool) {
	value, exists := m.connections.Load(connKey)
	if !exists {
		return
	}

	conn := value.(*ConnectionState)
	if conn.StreamChannel == nil {
		return
	}

	// Release the stream
	m.streamPool.ReleaseStreamChannel(conn.PeerID, conn.StreamChannel, healthy)
	conn.StreamChannel = nil

	// Remove the connection if unhealthy
	if !healthy {
		m.connections.Delete(connKey)
	} else {
		// Update the connection state
		m.connections.Store(connKey, conn)
	}

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

	// Send the packet to the stream's channel
	select {
	case streamChannel.PacketChan <- packet:
		// Packet sent successfully
		atomic.AddInt64(&m.metrics.PacketsProcessed, 1)
		return nil
	case <-ctx.Done():
		// Context cancelled
		atomic.AddInt64(&m.metrics.Errors, 1)
		return types.ErrContextCancelled
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
	m.connections.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// Close implements io.Closer
func (m *StreamManager) Close() error {
	m.Stop()
	return nil
}
