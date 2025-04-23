package apps

import (
	"context"
	"fmt"
	"math/big"

	ctypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	vtypes "github.com/docker/docker/api/types/volume"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

// ...existing code...

func (s *Service) UpdateAppConfig(ctx context.Context, appId *big.Int, cfg atypes.ContainerConfig) error {
	app, err := s.GetApp(ctx, appId)
	if err != nil {
		return fmt.Errorf("failed to fetch app details: %w", err)
	}

	app.Metadata.ContainerConfig = cfg

	// Save container configuration to datastore using proto
	err = s.SaveContainerConfigProto(ctx, appId, app.Metadata.ContainerConfig)
	if err != nil {
		return fmt.Errorf("failed to save container config: %w", err)
	}

	return nil
}

func (s *Service) containerHostConfig(appId *big.Int, containerConfig atypes.ContainerConfig) (*ctypes.HostConfig, error) {
	//
	// HANDLE MOUNT VOLUME
	//
	volumes := containerConfig.Volumes

	hostConfig := &ctypes.HostConfig{
		RestartPolicy: ctypes.RestartPolicy{
			Name: "unless-stopped",
		},
	}
	mounts := []mount.Mount{}

	for _, volume := range volumes {

		volumeName := fmt.Sprintf("%s_%s", appId.String(), volume.Name)

		// Create the volume using Docker CLI
		_, err := s.dockerClient.VolumeCreate(context.Background(), vtypes.CreateOptions{
			Name: volumeName,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create docker volume: %w", err)
		}

		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeVolume,
			Source: volumeName,
			Target: volume.MountPath,
		})
	}
	hostConfig.Mounts = mounts

	return hostConfig, nil
}

func (s *Service) containerEnvConfig(containerConfig atypes.ContainerConfig) ([]string, error) {
	containerEnvs := containerConfig.Env
	// Add environment variables to the container spec
	envs := []string{}
	for key, value := range containerEnvs {
		if value == "${PEER_ID}" {
			value = s.peerId.String()
		}

		envs = append(envs, fmt.Sprintf("%s=%s", key, value))
	}

	return envs, nil
}
