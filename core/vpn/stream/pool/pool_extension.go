package pool

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

// PoolServiceExtension extends the PoolService with additional methods
type PoolServiceExtension interface {
	// GetStreamByIndex gets a stream by index from the pool
	GetStreamByIndex(ctx context.Context, peerID peer.ID, index int) (types.VPNStream, error)
	// ReleaseStreamByIndex releases a stream by index
	ReleaseStreamByIndex(peerID peer.ID, index int, close bool)
	// SetTargetStreamsForPeer sets the target number of streams for a peer
	SetTargetStreamsForPeer(peerID peer.ID, targetStreams int)
}
