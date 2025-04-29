package dispatch

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/pool"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
)

// DispatcherService defines the interface for packet dispatching
type DispatcherService interface {
	// DispatchPacket dispatches a packet to the appropriate stream
	DispatchPacket(ctx context.Context, connKey types.ConnectionKey, destIP string, packet []byte) error

	// DispatchPacketWithCallback dispatches a packet and provides a callback channel for the result
	DispatchPacketWithCallback(ctx context.Context, connKey types.ConnectionKey, destIP string, packet []byte, doneCh chan error) error

	// DispatchPacketWithFuncCallback dispatches a packet and provides a function callback for the result
	DispatchPacketWithFuncCallback(ctx context.Context, connKey types.ConnectionKey, destIP string, packet []byte, callback func(error)) error

	// Start starts the dispatcher
	Start()

	// Stop stops the dispatcher
	Stop()

	// GetMetrics returns the dispatcher's metrics
	GetMetrics() map[string]int64
}

// StreamManagerService defines the interface for stream management
type StreamManagerService interface {
	// SendPacket sends a packet through the appropriate stream
	SendPacket(ctx context.Context, connKey types.ConnectionKey, peerID peer.ID, packet *types.QueuedPacket) error

	// Start starts the stream manager
	Start()

	// Stop stops the stream manager
	Stop()

	// GetMetrics returns the stream manager's metrics
	GetMetrics() map[string]int64
}

// StreamPoolService defines the interface for stream pooling
type StreamPoolService interface {
	// GetStreamChannel gets a stream channel for a peer
	GetStreamChannel(ctx context.Context, peerID peer.ID) (*pool.StreamChannel, error)

	// ReleaseStreamChannel marks a stream channel as unhealthy if it's not healthy
	ReleaseStreamChannel(peerID peer.ID, streamChannel *pool.StreamChannel, healthy bool)

	// Start starts the stream pool
	Start()

	// Stop stops the stream pool
	Stop()

	// GetStreamMetrics returns metrics for all streams
	GetStreamMetrics() map[string]map[string]int64
}
