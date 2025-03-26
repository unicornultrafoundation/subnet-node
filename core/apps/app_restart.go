package apps

import (
	"context"
	"fmt"
	"io"
	"math/big"

	ctypes "github.com/docker/docker/api/types/container"
	itypes "github.com/docker/docker/api/types/image"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

// RestartInactiveApps restarts all running apps in db which are not currently running in Docker
func (s *Service) RestartInactiveApps(ctx context.Context) error {
	runningAppList, err := s.GetRunningAppListProto(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch running app list for restarting: %v", err)
	}

	for _, appIdBytes := range runningAppList.AppIds {
		appId := new(big.Int).SetBytes(appIdBytes)
		status, err := s.GetContainerStatus(ctx, appId)
		if err != nil {
			log.Errorf("failed to get container status from appId %s: %v", appId, err)
			continue
		}

		if status != atypes.Running {
			log.Infof("AppId %v is somehow not running as expected. Restarting...", appId)
			err := s.RestartContainer(ctx, appId)

			if err != nil {
				log.Errorf("failed to restart appId %s: %v", appId, err)
				continue
			}
		}
	}

	return nil
}

// Stop, remove container without removing mount volume
// Start container again with latest Images
func (s *Service) RestartContainer(ctx context.Context, appId *big.Int) error {
	// Get app details
	app, err := s.GetApp(ctx, appId)
	if err != nil {
		return fmt.Errorf("failed to fetch app details: %w", err)
	}

	containerId := app.ContainerId()

	//  Stop the container
	if err := s.dockerClient.ContainerStop(ctx, containerId, ctypes.StopOptions{}); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	// Remove the container but keep volumes (RemoveOptions.RemoveVolumes = false)
	if err := s.dockerClient.ContainerRemove(ctx, containerId, ctypes.RemoveOptions{
		RemoveVolumes: false,
	}); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	// Pull the image for the app
	imageName := app.Metadata.ContainerConfig.Image
	out, err := s.dockerClient.ImagePull(ctx, imageName, itypes.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer out.Close()
	io.Copy(io.Discard, out)

	// Add environment variables to the container spec
	envs, err := s.containerEnvConfig(app.Metadata.ContainerConfig)
	if err != nil {
		return fmt.Errorf("failed to create container env config: %w", err)
	}

	// Create new container config with the same volume mounts
	hostConfig, err := s.containerHostConfig(appId, app.Metadata.ContainerConfig)
	if err != nil {
		return fmt.Errorf("failed to create host config: %w", err)
	}

	// Create new container with same configuration
	// hostConfig includes the existed volume mounts
	container, err := s.dockerClient.ContainerCreate(ctx,
		&ctypes.Config{
			Image: imageName,
			Cmd:   app.Metadata.ContainerConfig.Command,
			Env:   envs,
		},
		hostConfig, nil, nil, containerId)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Start the new container
	if err := s.dockerClient.ContainerStart(ctx, container.ID, ctypes.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	log.Printf("App %s restarted successfully: Docker Container ID: %s", app.Name, container.ID)
	return nil
}
