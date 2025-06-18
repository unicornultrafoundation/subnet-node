package node

import (
	"context"

	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/core/deployer"
	"go.uber.org/fx"
)

// DeployerService provides a lifecycle-managed Deployer service
func DeployerService(lc fx.Lifecycle, cfg *config.C, acc *account.AccountService) (*deployer.Service, error) {
	if !cfg.GetBool("deployer.enable", false) {
		return nil, nil
	}

	service, err := deployer.NewService(cfg, acc)
	if err != nil {
		return nil, err
	}

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
