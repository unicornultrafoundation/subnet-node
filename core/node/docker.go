package node

import (
	"context"

	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/docker"
	"go.uber.org/fx"
)

// DockerService provides a lifecycle-managed Docker service
func DockerService(lc fx.Lifecycle, cfg *config.C) (*docker.Service, error) {
	service, err := docker.NewDockerService(cfg)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println("DockerService started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("DockerService stopped")
			return nil
		},
	})

	return service, nil
}
