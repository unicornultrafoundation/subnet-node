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

	//
	// Update NodeResourceUsage
	//
	requestCPU, requestMemory, requestStorage, err := s.parseResourceRequest(app.Metadata.ContainerConfig.Resources.Requests)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource request: %w", err)
	}

	resourceUsage, err := s.getNodeResourceUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to get node resource usage: %w", err)
	}

	// Subtract resource usage from NodeResourceUsage
	newResourceUsage := atypes.ResourceUsage{
		UsedCpu:     new(big.Int).Sub(resourceUsage.UsedCpu, requestCPU),
		UsedMemory:  new(big.Int).Sub(resourceUsage.UsedMemory, requestMemory),
		UsedStorage: new(big.Int).Sub(resourceUsage.UsedStorage, requestStorage),
	}

	// Ensure values are not less than zero
	if newResourceUsage.UsedCpu.Cmp(big.NewInt(0)) < 0 {
		newResourceUsage.UsedCpu.SetInt64(0)
	}
	if newResourceUsage.UsedMemory.Cmp(big.NewInt(0)) < 0 {
		newResourceUsage.UsedMemory.SetInt64(0)
	}
	if newResourceUsage.UsedStorage.Cmp(big.NewInt(0)) < 0 {
		newResourceUsage.UsedStorage.SetInt64(0)
	}

	s.setNodeResourceUsage(newResourceUsage)

	return app, nil
}
