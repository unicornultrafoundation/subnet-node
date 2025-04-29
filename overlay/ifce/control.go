package ifce

import (
	"context"
	"net/netip"

	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/overlay"
	"go.uber.org/fx"
)

type Control struct {
	f      *Interface
	l      *logrus.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

func (c *Control) Context() context.Context {
	return c.ctx
}

func (c *Control) Start() error {
	// Activate the interface
	c.f.activate()
	// Start reading packets.
	c.f.run()
	return nil
}

func (c *Control) Stop() error {
	c.cancel()

	if err := c.f.Close(); err != nil {
		c.l.WithError(err).Error("Close interface failed")
	}
	c.l.Info("Goodbye")
	return nil
}

func OverlayService(lc fx.Lifecycle, peerHost p2phost.Host, cfg *config.C) (*Control, error) {

	// Create a new logger
	logger := logrus.New()

	prefix, err := netip.ParsePrefix(cfg.GetString("vpn.virtual_ip", "") + "/16")
	if err != nil {
		return nil, err
	}

	routines := 1

	inside, err := overlay.NewDeviceFromConfig(cfg, logrus.New(), []netip.Prefix{prefix}, routines)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())

	// Create a new interface
	ifce := NewInterface(ctx, &InterfaceConfig{
		Inside:   inside,
		Routines: routines,
		P2phost:  peerHost,
		Logger:   logger,
	})

	// Create a new control instance
	control := &Control{
		f:      ifce,
		l:      logger,
		ctx:    ctx,
		cancel: cancel,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return control.Start()
		},
		OnStop: func(ctx context.Context) error {
			return control.Stop()
		},
	})

	return control, nil
}
