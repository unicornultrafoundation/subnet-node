package worker

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
)

// WorkerInterface defines the common interface for all worker types
type WorkerInterface interface {
	// Start begins the worker's packet processing loop
	Start()
	// Stop terminates the worker's processing loop
	Stop()
	// IsRunning returns whether the worker is running
	IsRunning() bool
	// GetLastActivity returns the timestamp of the last activity
	GetLastActivity() time.Time
	// GetMetrics returns the worker's metrics
	GetMetrics() types.WorkerMetrics
	// Close implements io.Closer
	Close() error
}

// MultiConnectionWorkerInterface is the interface for workers that handle multiple connections
type MultiConnectionWorkerInterface interface {
	WorkerInterface
	// EnqueuePacket adds a packet to the worker's queue for a specific connection key
	EnqueuePacket(packet *types.QueuedPacket, connKey types.ConnectionKey) bool
	// GetConnectionCount returns the number of connections this worker is handling
	GetConnectionCount() int
	// GetConnectionMetrics returns metrics for all connections
	GetConnectionMetrics() map[string]types.WorkerMetrics
}

// StreamManagerInterface defines the interface for stream management
type StreamManagerInterface interface {
	// SendPacket sends a packet through the appropriate stream
	SendPacket(ctx context.Context, connKey types.ConnectionKey, peerID peer.ID, packet *types.QueuedPacket) error

	// Start starts the stream manager
	Start()

	// Stop stops the stream manager
	Stop()

	// GetMetrics returns the stream manager's metrics
	GetMetrics() map[string]int64

	// GetConnectionCount returns the number of active connections
	GetConnectionCount() int

	// Close implements io.Closer
	Close() error
}
