package node

import (
	"context"

	"github.com/ipfs/go-datastore"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/kvm"
	"github.com/unicornultrafoundation/subnet-node/core/node/resource"
	"go.uber.org/fx"
)

func KVMService(lc fx.Lifecycle, cfg *config.C, ds datastore.Datastore, resourceSvc *resource.Service) (*kvm.Service, error) {
	// Check if KVM service is enabled
	if !cfg.GetBool("kvm.enabled", false) {
		log.Info("KVM service is disabled")
		return nil, nil
	}

	// Create logger for KVM service
	logger := logrus.WithField("service", "kvm")

	// Create KVM service (either real libvirt or simulation based on config)
	var srv *kvm.Service
	if resourceSvc != nil {
		srv = kvm.NewService(cfg, logger, ds, *resourceSvc)
	} else {
		// If resource service is nil, we can't create KVM service
		log.Warn("Resource service is not available, KVM service will be disabled")
		return nil, nil
	}

	// Register lifecycle hooks
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting KVM service")
			return srv.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping KVM service")
			return srv.Stop(ctx)
		},
	})

	return srv, nil
}
