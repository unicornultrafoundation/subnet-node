package coreiface

import "github.com/unicornultrafoundation/subnet-node/core/coreiface/options"

type CoreAPI interface {
	PubSub() PubSubAPI

	Routing() RoutingAPI

	// Swarm returns an implementation of Swarm API
	Swarm() SwarmAPI

	// WithOptions creates new instance of CoreAPI based on this instance with
	// a set of options applied
	WithOptions(...options.ApiOption) (CoreAPI, error)
}
