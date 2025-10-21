package node

import (
	"context"

	"github.com/unicornultrafoundation/subnet-node/core/vpn"
	vpnconfig "github.com/unicornultrafoundation/subnet-node/core/vpn/config"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/discovery"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatcher"
	ipmanager "github.com/unicornultrafoundation/subnet-node/core/vpn/ip_manager"
	"go.uber.org/fx"
)

func VPNService(lc fx.Lifecycle, ipManager ipmanager.IPManager, configService vpnconfig.ConfigService, discoveryService discovery.DiscoveryService, dispatcherService dispatcher.DispatcherService) *vpn.Service {
	srv := vpn.NewService(ipManager, configService, discoveryService, dispatcherService)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return srv.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return srv.Stop(ctx)
		},
	})

	return srv
}
