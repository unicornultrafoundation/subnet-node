package vpn

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/pool"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

// StreamServiceAdapter adapts the StreamService to implement the PoolServiceExtension interface
type StreamServicePoolAdapter struct {
	service *Service
}

// GetStreamByIndex gets a stream by index from the pool
func (a *StreamServicePoolAdapter) GetStreamByIndex(ctx context.Context, peerID peer.ID, index int) (types.VPNStream, error) {
	// For now, we'll just create a new stream since the current StreamService doesn't support getting by index
	return a.service.streamService.CreateNewVPNStream(ctx, peerID)
}

// ReleaseStreamByIndex releases a stream by index
func (a *StreamServicePoolAdapter) ReleaseStreamByIndex(peerID peer.ID, index int, close bool) {
	// No-op for now since we don't have a way to track streams by index
}

// SetTargetStreamsForPeer sets the target number of streams for a peer
func (a *StreamServicePoolAdapter) SetTargetStreamsForPeer(peerID peer.ID, targetStreams int) {
	// No-op for now since we don't have a way to set target streams
}

// Ensure StreamServicePoolAdapter implements the PoolServiceExtension interface
var _ pool.PoolServiceExtension = (*StreamServicePoolAdapter)(nil)
