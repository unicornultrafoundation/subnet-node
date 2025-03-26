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

// Starts a container for the specified app using docker.
func (s *Service) RunApp(ctx context.Context, appId *big.Int) (*atypes.App, error) {
	// Retrieve app details from the Ethereum contract
	app, err := s.GetApp(ctx, appId)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch app details: %w", err)
	}

	// Return if found app
	if app.Status != atypes.NotFound {
		return app, nil
	}

	// Fetch the machine's hardware and software capabilities
	deviceCap, err := s.getDeviceCapability()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch app details: %w", err)
	}

	// Check if the machine's hardware and software capabilities can run this app.
	resourceUsage, err := s.getNodeResourceUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to get node resource usage: %w", err)
	}

	requestCPU, requestMemory, requestStorage, err := s.validateNodeCompatibility(resourceUsage, app.Metadata.ContainerConfig.Resources.Requests, deviceCap)
	// Return if validation fails
	if err != nil {
		return nil, fmt.Errorf("failed to validate node compatibility: %w", err)
	}

	imageName := app.Metadata.ContainerConfig.Image

	// Pull the image for the app
	out, err := s.dockerClient.ImagePull(ctx, imageName, itypes.PullOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	defer out.Close()

	io.Copy(io.Discard, out)

	// Add environment variables to the container spec
	envs, err := s.containerEnvConfig(app.Metadata.ContainerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create container env config: %w", err)
	}

	hostConfig, err := s.containerHostConfig(appId, app.Metadata.ContainerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create container host config: %w", err)
	}

	// Create a new container for the app
	container, err := s.dockerClient.ContainerCreate(ctx, &ctypes.Config{
		Image: imageName,
		Cmd:   app.Metadata.ContainerConfig.Command,
		Env:   envs,
	}, hostConfig, nil, nil, app.ContainerId())
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	err = s.dockerClient.ContainerStart(ctx, container.ID, ctypes.StartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	fmt.Println("Container started with ID:", container.ID)

	// Ensure the container has an IP address assigned using CNI
	ip, err := s.GetContainerIP(ctx, appId)
	if err != nil || ip == "" {
		return nil, fmt.Errorf("failed to get container IP: %w", err)
	}

	log.Printf("App %s started successfully: Docker Container ID: %s, IP: %s", app.Name, container.ID, ip)

	app.Status, err = s.GetContainerStatus(ctx, appId)
	if err != nil {
		return nil, err
	}

	// Store new resource usage into cache
	if app.Status == atypes.Running {
		// calculate new resource usage equal to resourceUsage + app request
		newResourceUsage := atypes.ResourceUsage{
			UsedCpu:     new(big.Int).Add(resourceUsage.UsedCpu, requestCPU),
			UsedMemory:  new(big.Int).Add(resourceUsage.UsedMemory, requestMemory),
			UsedStorage: new(big.Int).Add(resourceUsage.UsedStorage, requestStorage),
		}

		s.setNodeResourceUsage(newResourceUsage)
	}

	s.AddNewRunningApp(ctx, appId)

	return app, nil
}
