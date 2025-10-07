package node

import (
	"context"

	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/core/vpn"
	"github.com/unicornultrafoundation/subnet-node/firewall"
	"go.uber.org/fx"
)

func VPNService(lc fx.Lifecycle, cfg *config.C, peerHost p2phost.Host, dht *ddht.DHT, accountService *account.AccountService, firewall firewall.FirewallInterface) (*vpn.Service, error) {
	srv := vpn.New(cfg, peerHost, dht, accountService, firewall)

	// Register reload callback for VPN configuration
	cfg.RegisterReloadCallback(func(c *config.C) {
		// Check if VPN configuration has changed
		if c.HasChanged("vpn") {
			log := logrus.WithField("service", "vpn")
			log.Info("VPN configuration changed, restarting service...")

			// Stop the service
			if err := srv.Stop(); err != nil {
				log.WithError(err).Error("Failed to stop VPN service during config reload")
				return
			}

			// Restart the service
			ctx := context.Background()
			if err := srv.Start(ctx); err != nil {
				log.WithError(err).Error("Failed to restart VPN service after config reload")
				return
			}

			log.Info("VPN service restarted successfully")
		}
	})

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
