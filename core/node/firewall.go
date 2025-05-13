package node

import (
	"context"
	"net/netip"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/firewall"
	"go.uber.org/fx"
)

func FirewallService(lc fx.Lifecycle, cfg *config.C) (firewall.FirewallInterface, error) {
	l := logrus.WithField("module", "firewall").Logger
	srv, err := firewall.NewFirewallFromConfig(l, cfg, []netip.Prefix{})
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return srv.Destroy()
		},
		OnStart: func(_ context.Context) error {
			return srv.Start()
		},
	})

	return srv, nil
}
