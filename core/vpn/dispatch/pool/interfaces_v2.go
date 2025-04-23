package pool

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

// StreamPoolInterfaceV2 defines the interface for stream pooling with reduced mutex usage
type StreamPoolInterfaceV2 interface {
	// GetStreamChannel gets a stream channel for a peer
	GetStreamChannel(ctx context.Context, peerID peer.ID) (*StreamChannelV2, error)

	// ReleaseStreamChannel marks a stream channel as unhealthy if it's not healthy
	ReleaseStreamChannel(peerID peer.ID, streamChannel *StreamChannelV2, healthy bool)

	// Start starts the stream pool
	Start()

	// Stop stops the stream pool
	Stop()

	// GetStreamMetrics returns metrics for all streams
	GetStreamMetrics() map[string]map[string]int64

	// GetStreamCount returns the number of streams for a peer
	GetStreamCount(peerID peer.ID) int

	// GetTotalStreamCount returns the total number of streams
	GetTotalStreamCount() int

	// Close implements io.Closer
	Close() error

	// GetPacketBufferSize returns the packet buffer size for streams
	GetPacketBufferSize() int
}
