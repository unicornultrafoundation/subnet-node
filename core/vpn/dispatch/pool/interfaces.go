package pool

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

// StreamPoolInterface defines the interface for stream pooling
type StreamPoolInterface interface {
	// GetStreamChannel gets a stream channel for a peer
	GetStreamChannel(ctx context.Context, peerID peer.ID) (*StreamChannel, error)
	
	// ReleaseStreamChannel marks a stream channel as unhealthy if it's not healthy
	ReleaseStreamChannel(peerID peer.ID, streamChannel *StreamChannel, healthy bool)
	
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
}
