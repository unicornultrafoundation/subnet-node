package apps

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
)

// Starts a container for the specified app using containerd.
func (s *Service) RunApp(ctx context.Context, appId *big.Int, envVars map[string]string) (*App, error) {
	// Set the namespace for the container
	ctx = namespaces.WithNamespace(ctx, NAMESPACE)

	if _, err := s.RegisterNode(ctx, appId); err != nil {
		return nil, err
	}

	// Retrieve app details from the Ethereum contract
	app, err := s.GetApp(ctx, appId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch app details: %w", err)
	}

	if app.Status != NotFound {
		return app, nil
	}

	app.Metadata.ContainerConfig.Env = envVars

	imageName := app.Metadata.ContainerConfig.Image
	if !strings.Contains(imageName, "/") {
		imageName = "docker.io/library/" + imageName
	}

	// Pull the image for the app
	image, err := s.containerdClient.Pull(ctx, imageName, containerd.WithPullUnpack)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	// Create a new container for the app
	specOpts := []oci.SpecOpts{
		oci.WithImageConfig(image),
		//oci.WithHostHostsFile,  // Use host's /etc/hosts
		//oci.WithHostResolvconf, // Use host's /etc/resolv.conf
	}

	// Add environment variables to the container spec
	for key, value := range envVars {
		specOpts = append(specOpts, oci.WithEnv([]string{fmt.Sprintf("%s=%s", key, value)}))
	}

	// Add volume to the container spec
	// specOpts = append(specOpts, oci.WithMounts([]specs.Mount{
	// 	{
	// 		Source:      "/host/path",
	// 		Destination: "/container/path",
	// 		Type:        "bind",
	// 		Options:     []string{"rbind", "rw"},
	// 	},
	// }))

	container, err := s.containerdClient.NewContainer(
		ctx,
		app.ContainerId(),
		containerd.WithNewSnapshot(app.ContainerId()+"-snapshot", image),
		containerd.WithNewSpec(specOpts...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Start the container
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	_, err = task.Wait(ctx)
	if err != nil {
		return nil, err
	}

	err = task.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start task: %w", err)
	}

	// Ensure the container has an IP address assigned using CNI
	ip, err := s.GetContainerIP(ctx, appId)
	if err != nil || ip == "" {
		return nil, fmt.Errorf("failed to get container IP: %w", err)
	}

	log.Printf("App %s started successfully: Container ID: %s, IP: %s", app.Name, container.ID(), ip)

	app.Status, err = s.GetContainerStatus(ctx, appId)
	if err != nil {
		return nil, err
	}

	// Save container configuration to datastore using proto
	err = s.SaveContainerConfigProto(ctx, appId, app.Metadata.ContainerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to save container config: %w", err)
	}

	return app, nil
}
