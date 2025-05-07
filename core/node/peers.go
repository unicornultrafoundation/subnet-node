package node

import (
	"context"

	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/peers"
	"go.uber.org/fx"
)

func PeerService(lc fx.Lifecycle, cfg *config.C, peerHost p2phost.Host, dht *ddht.DHT) (*peers.Service, error) {
	srv := peers.New(cfg, peerHost, dht)

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return srv.Stop()
		},
		OnStart: func(_ context.Context) error {
			return srv.Start()
		},
	})

	return srv, nil
}
