package coreiface

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/coreiface/options"
	"github.com/unicornultrafoundation/subnet-node/core/node/resource"
)

type CoreAPI interface {
	Identity() peer.ID

	PubSub() PubSubAPI

	Routing() RoutingAPI

	Resource() *resource.Service

	// Swarm returns an implementation of Swarm API
	Swarm() SwarmAPI

	// WithOptions creates new instance of CoreAPI based on this instance with
	// a set of options applied
	WithOptions(...options.ApiOption) (CoreAPI, error)
}
