package node

import (
	"context"

	"github.com/ipfs/go-datastore"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/bidengine"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"go.uber.org/fx"
)

func BidengineService(lc fx.Lifecycle, cfg *config.C, acc *account.AccountService, dataStore datastore.Datastore) (*bidengine.BidEngine, error) {
	if !cfg.GetBool("bidengine.enable", false) {
		return nil, nil
	}

	transactOpts, err := acc.NewKeyedTransactor()
	if err != nil {
		return nil, err
	}

	srv, err := bidengine.NewBidEngineFromConfigC(
		cfg,
		acc.GetClient(),
		transactOpts,
		dataStore,
	)

	if err != nil {
		return nil, err
	}

	// Register reload callback for BidEngine configuration
	cfg.RegisterReloadCallback(func(c *config.C) {
		// Check if BidEngine configuration has changed
		if c.HasChanged("bidengine") {
			log := logrus.WithField("service", "bidengine")
			log.Info("BidEngine configuration changed, restarting service...")

			ctx := context.Background()

			// Stop the service
			if err := srv.Stop(ctx); err != nil {
				log.WithError(err).Error("Failed to stop BidEngine service during config reload")
				return
			}

			// Restart the service
			if err := srv.Start(ctx); err != nil {
				log.WithError(err).Error("Failed to restart BidEngine service after config reload")
				return
			}

			log.Info("BidEngine service restarted successfully")
		}
	})

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return srv.Stop(ctx)
		},
		OnStart: func(ctx context.Context) error {
			return srv.Start(ctx)
		},
	})

	return srv, nil
}
