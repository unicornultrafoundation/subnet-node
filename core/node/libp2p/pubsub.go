package libp2p

import (
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/unicornultrafoundation/subnet-node/core/node/helpers"
	"go.uber.org/fx"
)

func FloodSub(pubsubOptions ...pubsub.Option) interface{} {
	return func(mctx helpers.MetricsCtx, lc fx.Lifecycle, host host.Host, disc discovery.Discovery) (service *pubsub.PubSub, err error) {
		return pubsub.NewFloodSub(helpers.LifecycleCtx(mctx, lc), host, append(pubsubOptions, pubsub.WithDiscovery(disc))...)
	}
}

func GossipSub(pubsubOptions ...pubsub.Option) interface{} {
	return func(mctx helpers.MetricsCtx, lc fx.Lifecycle, host host.Host, disc discovery.Discovery) (service *pubsub.PubSub, err error) {
		return pubsub.NewGossipSub(helpers.LifecycleCtx(mctx, lc), host, append(
			pubsubOptions,
			pubsub.WithDiscovery(disc),
			pubsub.WithFloodPublish(true))...,
		)
	}
}
