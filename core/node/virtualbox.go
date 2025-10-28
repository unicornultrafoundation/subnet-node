package node

import (
	"context"

	"github.com/ipfs/go-datastore"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox"
	"go.uber.org/fx"
)

// VirtualBoxService provides a lifecycle-managed VirtualBox service
func VirtualBoxService(lc fx.Lifecycle, cfg *config.C, ds datastore.Datastore) (*virtualbox.VirtualboxService, error) {
	if !cfg.GetBool("virtualbox.enable", false) {
		return nil, nil
	}

	vboxConfig, err := virtualbox.GetVirtualBoxConfig(cfg)
	if err != nil {
		return nil, err
	}

	service, err := virtualbox.NewService(ds, vboxConfig)
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
