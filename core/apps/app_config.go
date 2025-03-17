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

func (s *Service) containerHostConfig(containerConfig atypes.ContainerConfig) (*ctypes.HostConfig, error) {
	//
	// HANDLE MOUNT VOLUME
	//
	volumes := containerConfig.Volumes

	hostConfig := &ctypes.HostConfig{}
	mounts := []mount.Mount{}

	for _, volume := range volumes {

		// Create the volume using Docker CLI
		_, err := s.dockerClient.VolumeCreate(context.Background(), vtypes.CreateOptions{
			Name: volume.Name,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create docker volume: %w", err)
		}

		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeVolume,
			Source: volume.Name,
			Target: volume.MountPath,
		})
	}
	hostConfig.Mounts = mounts

	return hostConfig, nil
}
