package apps

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	ctypes "github.com/docker/docker/api/types/container"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

func (s *Service) RestartStoppedContainers(ctx context.Context) error {
	// Fetch all running containers
	containers, err := s.dockerClient.ContainerList(ctx, ctypes.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to fetch running containers: %w", err)
	}

	for _, container := range containers {
		// Get container ID (assuming appID is same as container ID)
		containerId := strings.TrimPrefix(container.Names[0], "/")

		if !strings.HasPrefix(containerId, "subnet-") {
			continue
		}

		appId, err := atypes.GetAppIdFromContainerId(containerId)

		if err != nil {
			log.Debugf("failed to get appId from containerId %s: %v", containerId, err)
			continue
		}

		status, err := s.GetContainerStatus(ctx, appId)

		if err != nil {
			log.Errorf("failed to get status from containerId %s: %v", containerId, err)
			continue
		}

		if status == atypes.Stopped {
			err := s.RestartContainer(ctx, appId)

			if err != nil {
				log.Errorf("failed to restart containerId %s: %v", containerId, err)
				continue
			}
		}
	}

	return nil
}

func (s *Service) RestartContainer(ctx context.Context, appId *big.Int) error {
	_, err := s.RemoveApp(ctx, appId)

	if err != nil {
		return fmt.Errorf("failed to remove appId %v: %v", appId, err)
	}

	_, err = s.RunApp(ctx, appId)
	if err != nil {
		return fmt.Errorf("failed to run appId %v: %v", appId, err)
	}

	return nil
}
