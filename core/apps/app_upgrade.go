package apps

import (
	"context"
	"fmt"
	"strings"
	"time"

	ctypes "github.com/docker/docker/api/types/container"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

func (s *Service) startUpgradeAppVersion(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Infof("Starting upgrade app version for all running containers...")
			s.upgradeAppVersion(ctx)
		case <-s.stopChan:
			log.Infof("Stopping upgrade app version for all containers")
			return
		case <-ctx.Done():
			log.Infof("Context canceled, stopping upgrade app version")
			return
		}
	}
}

func (s *Service) upgradeAppVersion(ctx context.Context) error {
	containers, err := s.dockerClient.ContainerList(ctx, ctypes.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to fetch running containers: %v", err)
	}

	for _, container := range containers {
		// Get container ID (assuming appID is same as container ID)
		containerId := strings.TrimPrefix(container.Names[0], "/")

		if !strings.HasPrefix(containerId, "subnet-") {
			continue
		}

		appId, err := atypes.GetAppIdFromContainerId(containerId)

		if err != nil {
			log.Debugf("Failed to get appId from containerId %s: %v", containerId, err)
			continue
		}

		// Get running Container image version
		runningImageVersion := container.Image

		// Get latest App image version from smart contract
		app, err := s.GetApp(ctx, appId)
		if err != nil {
			return fmt.Errorf("failed to fetch app details: %w", err)
		}

		latestImageVersion := app.Metadata.ContainerConfig.Image

		if runningImageVersion != latestImageVersion {
			return s.RestartContainer(ctx, appId) // this will upgrade the container with the new image version
		}
	}

	return nil
}
