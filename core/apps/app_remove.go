package apps

import (
	"context"
	"fmt"
	"math/big"

	ctypes "github.com/docker/docker/api/types/container"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

// Stops and removes the container for the specified app.
func (s *Service) RemoveApp(ctx context.Context, appId *big.Int) (*atypes.App, error) {
	// Retrieve app details from the Ethereum contract
	app, err := s.GetApp(ctx, appId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch app details: %w", err)
	}

	containerID := app.ContainerId()

	// Stop container
	if err := s.dockerClient.ContainerStop(ctx, containerID, ctypes.StopOptions{}); err != nil {
		return nil, fmt.Errorf("failed to stop docker container: %w", err)
	}

	// Remove container
	if err := s.dockerClient.ContainerRemove(ctx, containerID, ctypes.RemoveOptions{}); err != nil {
		return nil, fmt.Errorf("failed to remove docker container: %w", err)
	}

	log.Printf("App %s removed successfully: Container ID: %s", app.Symbol, app.ContainerId())
	app.Status = atypes.NotFound
	return app, nil
}
