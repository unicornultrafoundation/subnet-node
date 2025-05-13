package node

import (
	"context"

	"github.com/ipfs/go-datastore"
	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/core/apps"
	"github.com/unicornultrafoundation/subnet-node/core/apps/verifier"
	"github.com/unicornultrafoundation/subnet-node/core/docker"
	"github.com/unicornultrafoundation/subnet-node/p2p"
	"go.uber.org/fx"
)

func AppService(lc fx.Lifecycle, cfg *config.C, P2P *p2p.P2P, dataStore datastore.Datastore, acc *account.AccountService, peerId peer.ID, peerHost p2phost.Host, docker *docker.Service, dht *ddht.DHT) *apps.Service {
	if !cfg.GetBool("provider.enable", false) {
		return nil
	}

	srv := apps.New(peerHost, peerId, cfg, P2P, dataStore, acc, docker, dht)
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

func VerifierService(lc fx.Lifecycle, cfg *config.C, P2P *p2p.P2P, dataStore datastore.Datastore, acc *account.AccountService, peerId peer.ID, peerHost p2phost.Host, docker *docker.Service, dht *ddht.DHT) *verifier.Verifier {
	if !cfg.GetBool("verifier.enable", false) {
		return nil
	}
	srv := verifier.NewVerifier(dataStore, peerHost, P2P, acc, dht)
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return nil
		},
		OnStart: func(ctx context.Context) error {
			return srv.Register()
		},
	})

	return srv
}
