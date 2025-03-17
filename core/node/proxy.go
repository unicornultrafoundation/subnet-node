package node

import (
	"context"

	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/proxy"
	"github.com/unicornultrafoundation/subnet-node/p2p"
	"go.uber.org/fx"
)

func ProxyService(lc fx.Lifecycle, cfg *config.C, P2P *p2p.P2P, peerId peer.ID, peerHost p2phost.Host) *proxy.Service {
	srv := proxy.New(peerHost, peerId, cfg, P2P)
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
