package apps

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	ctypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	vtypes "github.com/docker/docker/api/types/volume"
	"github.com/docker/go-connections/nat"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

func (s *Service) UpdateAppConfig(ctx context.Context, appId *big.Int, cfg atypes.ContainerConfig) error {
	// app, err := s.GetApp(ctx, appId)
	// if err != nil {
	// 	return fmt.Errorf("failed to fetch app details: %w", err)
	// }

	// app.Metadata.ContainerConfig = cfg

	// // Save container configuration to datastore using proto
	// err = s.SaveContainerConfigProto(ctx, appId, app.Metadata.ContainerConfig)
	// if err != nil {
	// 	return fmt.Errorf("failed to save container config: %w", err)
	// }

	return nil
}

// createDockerVolume creates a Docker volume with the given name
func (s *Service) createDockerVolume(volumeName string) error {
	_, err := s.dockerClient.VolumeCreate(context.Background(), vtypes.CreateOptions{
		Name: volumeName,
	})
	if err != nil {
		return fmt.Errorf("failed to create docker volume: %w", err)
	}
	return nil
}

// setupVolumeMounts creates volume mounts for container
func (s *Service) setupVolumeMounts(appId *big.Int, volumes []atypes.Volume) ([]mount.Mount, error) {
	if len(volumes) == 0 {
		return nil, nil
	}

	mounts := make([]mount.Mount, 0, len(volumes))
	for _, volume := range volumes {
		volumeName := fmt.Sprintf("%s_%s", appId.String(), volume.Name)

		// Create the volume using Docker client
		err := s.createDockerVolume(volumeName)
		if err != nil {
			return nil, err
		}

		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeVolume,
			Source: volumeName,
			Target: volume.MountPath,
		})
	}

	return mounts, nil
}

// parseContainerPortMapping parses a port mapping string in format "hostPort:containerPort[/protocol]"
func parseContainerPortMapping(portMapping string) (hostPort string, containerPort nat.Port, err error) {
	// Parse host:container format
	portMappingParts := strings.Split(portMapping, ":")
	if len(portMappingParts) != 2 {
		return "", nat.Port(""), fmt.Errorf("invalid port mapping format: %s", portMapping)
	}

	hostPort = portMappingParts[0]
	containerPortWithProto := portMappingParts[1]

	// Default protocol is tcp if not specified
	proto := "tcp"
	containerPortStr := containerPortWithProto

	// Check if protocol is specified (containerPort/protocol)
	if protoParts := strings.Split(containerPortWithProto, "/"); len(protoParts) > 1 {
		containerPortStr = protoParts[0]
		proto = protoParts[1]
	}

	// Create nat.Port object
	containerPort, err = nat.NewPort(proto, containerPortStr)
	if err != nil {
		return "", nat.Port(""), fmt.Errorf("invalid port specification: %s: %w", portMapping, err)
	}

	return hostPort, containerPort, nil
}

// setupPortBindings creates port bindings for container
func setupPortBindings(ports []string) (nat.PortMap, error) {
	if len(ports) == 0 {
		return nil, nil
	}

	portMap := nat.PortMap{}

	for _, portMapping := range ports {
		hostPort, containerPort, err := parseContainerPortMapping(portMapping)
		if err != nil {
			return nil, err
		}

		// Add port binding to map
		portMap[containerPort] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0", // Bind to all interfaces
				HostPort: hostPort,
			},
		}
	}

	return portMap, nil
}

// containerHostConfig creates a host configuration for a container
func (s *Service) containerHostConfig(appId *big.Int, containerConfig atypes.ContainerConfig) (*ctypes.HostConfig, error) {
	// Create basic host config
	hostConfig := &ctypes.HostConfig{
		RestartPolicy: ctypes.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	// Setup volume mounts
	mounts, err := s.setupVolumeMounts(appId, containerConfig.Volumes)
	if err != nil {
		return nil, err
	}
	if mounts != nil {
		hostConfig.Mounts = mounts
	}

	// Setup port bindings
	portMap, err := setupPortBindings(containerConfig.Ports)
	if err != nil {
		return nil, err
	}
	if portMap != nil {
		hostConfig.PortBindings = portMap
	}

	return hostConfig, nil
}

// processEnvValue processes an environment variable value and replaces placeholders
func (s *Service) processEnvValue(value string) string {
	// Replace known placeholders with actual values
	if value == "${PEER_ID}" {
		return s.peerId.String()
	}

	// Return the original value if no placeholder is matched
	return value
}

// formatEnvPair formats a key-value pair into the "KEY=value" format
func formatEnvPair(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

// containerEnvConfig creates environment variable configuration for a container
func (s *Service) containerEnvConfig(containerConfig atypes.ContainerConfig) ([]string, error) {
	if len(containerConfig.Env) == 0 {
		return []string{}, nil
	}

	// Preallocate the slice with the expected capacity
	envs := make([]string, 0, len(containerConfig.Env))

	// Process each environment variable
	for key, value := range containerConfig.Env {
		// Process the value to replace placeholders
		processedValue := s.processEnvValue(value)

		// Add formatted environment variable to list
		envs = append(envs, formatEnvPair(key, processedValue))
	}

	return envs, nil
}
