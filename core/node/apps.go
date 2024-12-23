package node

import (
	"context"

	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/apps"
	"github.com/unicornultrafoundation/subnet-node/p2p"
	"go.uber.org/fx"
)

func AppService(lc fx.Lifecycle, cfg *config.C, P2P *p2p.P2P) *apps.Service {
	srv := apps.New(cfg, P2P)
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
