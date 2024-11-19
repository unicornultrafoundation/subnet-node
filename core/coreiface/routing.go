package coreiface

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/coreiface/options"
)

// RoutingAPI specifies the interface to the routing layer.
type RoutingAPI interface {
	// Get retrieves the best value for a given key
	Get(context.Context, string) ([]byte, error)

	// Put sets a value for a given key
	Put(ctx context.Context, key string, value []byte, opts ...options.RoutingPutOption) error

	// FindPeer queries the routing system for all the multiaddresses associated
	// with the given [peer.ID].
	FindPeer(context.Context, peer.ID) (peer.AddrInfo, error)
}
