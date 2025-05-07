package node

import (
	"context"

	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/core/vpn"
	"go.uber.org/fx"
)

func VPNService(lc fx.Lifecycle, cfg *config.C, peerHost p2phost.Host, dht *ddht.DHT, accountService *account.AccountService) (*vpn.Service, error) {
	srv := vpn.New(cfg, peerHost, dht, accountService)

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return srv.Stop()
		},
		OnStart: func(ctx context.Context) error {
			return srv.Start(ctx)
		},
	})

	return srv, nil
}
