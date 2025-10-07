package node

import (
	"context"

	"github.com/ipfs/go-datastore"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/deployer"
	"go.uber.org/fx"
)

// DeployerService provides a lifecycle-managed Deployer service
func DeployerService(lc fx.Lifecycle, cfg *config.C, ds datastore.Datastore) (*deployer.Service, error) {
	if !cfg.GetBool("deployer.enable", false) {
		return nil, nil
	}

	service, err := deployer.NewService(cfg, ds)
	if err != nil {
		return nil, err
	}

	// Register reload callback for Deployer configuration
	cfg.RegisterReloadCallback(func(c *config.C) {
		// Check if deployer configuration has changed
		if c.HasChanged("deployer") {
			log := logrus.WithField("service", "deployer")
			log.Info("Deployer configuration changed, restarting service...")

			ctx := context.Background()

			// Stop the service
			if err := service.Stop(ctx); err != nil {
				log.WithError(err).Error("Failed to stop Deployer service during config reload")
				return
			}

			// Restart the service
			if err := service.Start(ctx); err != nil {
				log.WithError(err).Error("Failed to restart Deployer service after config reload")
				return
			}

			log.Info("Deployer service restarted successfully")
		}
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return service.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return service.Stop(ctx)
		},
	})

	return service, nil
}
