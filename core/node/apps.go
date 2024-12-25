package node

import (
	"context"

	"github.com/ipfs/go-datastore"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/apps"
	"github.com/unicornultrafoundation/subnet-node/p2p"
	"go.uber.org/fx"
)

func AppService(lc fx.Lifecycle, cfg *config.C, P2P *p2p.P2P, dataStore datastore.Datastore) *apps.Service {
	srv := apps.New(cfg, P2P, dataStore)
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return srv.Stop(ctx)
		},
		OnStart: func(ctx context.Context) error {
			return srv.Start(ctx)
		},
	})

	return srv
}
