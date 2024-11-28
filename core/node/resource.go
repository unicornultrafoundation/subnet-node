package node

import (
	"context"
	"time"

	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/node/resource"
	"go.uber.org/fx"
)

func ResourceService(lc fx.Lifecycle, id peer.ID, pubsub *pubsub.PubSub, DHT *ddht.DHT, cfg *config.C) *resource.Service {
	srv := &resource.Service{
		Identity:   id,
		PubSub:     pubsub,
		DHT:        DHT,
		UpdateFreq: 1 * time.Minute,
		IsProvider: cfg.GetBool("provider.enable", false),
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return srv.Stop()
		},
		OnStart: func(ctx context.Context) error {
			return srv.Start()
		},
	})

	return srv
}
