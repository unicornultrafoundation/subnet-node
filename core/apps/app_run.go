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

	if app.Status != atypes.NotFound {
		return app, nil
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
	envs := []string{}
	for key, value := range app.Metadata.ContainerConfig.Env {
		envs = append(envs, fmt.Sprintf("%s=%s", key, value))
	}

	for key, value := range app.Metadata.ContainerConfig.Env {
		envs = append(envs, fmt.Sprintf("%s=%s", key, value))
	}

	// Create a new container for the app
	container, err := s.dockerClient.ContainerCreate(ctx, &ctypes.Config{
		Image: imageName,
		Cmd:   app.Metadata.ContainerConfig.Command,
		Env:   envs,
	}, nil, nil, nil, app.ContainerId())
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

	return app, nil
}
