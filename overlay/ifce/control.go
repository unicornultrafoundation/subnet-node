package ifce

import (
	"context"
	"net/netip"
	"time"

	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p/core/host"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/discovery"
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
	if err := c.waitUntilPeerConnected(c.ctx, c.f.p2phost); err != nil {
		return nil
	}

	if err := c.f.peerDiscovery.SyncPeerIDToDHT(c.ctx); err != nil {
		return err
	}

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

func OverlayService(lc fx.Lifecycle, dht *ddht.DHT, acc *account.AccountService, peerHost p2phost.Host, cfg *config.C) (*Control, error) {

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

	peerDiscovery := discovery.NewPeerDiscoveryFromLibp2p(
		peerHost,
		dht,
		cfg.GetString("vpn.virtual_ip", ""),
		acc,
	)

	// Create a new interface
	ifce := NewInterface(ctx, &InterfaceConfig{
		Inside:        inside,
		Routines:      routines,
		P2phost:       peerHost,
		Logger:        logger,
		PeerDiscovery: peerDiscovery,
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

func (c *Control) waitUntilPeerConnected(ctx context.Context, host host.Host) error {
	c.l.Debug("Waiting for peer connection (unlimited attempts)")

	// Use a simple retry loop with unlimited attempts
	var attempts int

	// Manual retry loop with unlimited attempts
	for attempts = 1; ; attempts++ { // No upper limit
		// Check if we have any peers connected
		if len(host.Network().Peers()) > 0 {
			c.l.Infof("Connected to %d peers", len(host.Network().Peers()))
			// Record success for metrics purposes
			return nil // Success, stop retrying
		}

		// Check if context is canceled
		if ctx.Err() != nil {
			return ctx.Err() // Exit the function on context cancellation
		}

		// No peers connected yet, wait before retrying
		c.l.Debugf("No peers connected yet, retry attempt %d (waiting for peers...)", attempts)
		select {
		case <-ctx.Done():
			return ctx.Err() // Exit the function on context cancellation
		case <-time.After(time.Duration(attempts) * 100 * time.Millisecond): // Simple backoff
			// Continue to next attempt
		}
	}
}
