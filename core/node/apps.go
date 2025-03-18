package node

import (
	"context"

	"github.com/ipfs/go-datastore"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/core/apps"
	"github.com/unicornultrafoundation/subnet-node/core/docker"
	"github.com/unicornultrafoundation/subnet-node/p2p"
	"go.uber.org/fx"
)

func AppService(lc fx.Lifecycle, cfg *config.C, P2P *p2p.P2P, dataStore datastore.Datastore, acc *account.AccountService, peerId peer.ID, peerHost p2phost.Host, docker *docker.Service) *apps.Service {
	srv := apps.New(peerHost, peerId, cfg, P2P, dataStore, acc, docker)
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
