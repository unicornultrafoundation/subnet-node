package node

import (
	"context"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/unicornultrafoundation/subnet-node/core/vpn"
	vpnconfig "github.com/unicornultrafoundation/subnet-node/core/vpn/config"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/discovery"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatcher"
	ipmanager "github.com/unicornultrafoundation/subnet-node/core/vpn/ip_manager"
	"github.com/unicornultrafoundation/subnet-node/firewall"
	"go.uber.org/fx"
)

func VPNService(lc fx.Lifecycle, ipManager ipmanager.IPManager, configService vpnconfig.ConfigService, discoveryService discovery.DiscoveryService, dispatcherService dispatcher.DispatcherService, peerHost host.Host, firewall firewall.FirewallInterface) *vpn.Service {
	srv := vpn.NewService(ipManager, configService, discoveryService, dispatcherService, peerHost, firewall)

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
