package apps

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	ctypes "github.com/docker/docker/api/types/container"
	itypes "github.com/docker/docker/api/types/image"
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
			// 1. Pull the latest image first
			out, err := s.dockerClient.ImagePull(ctx, latestImageVersion, itypes.PullOptions{})
			if err != nil {
				return fmt.Errorf("failed to pull image %v: %w", latestImageVersion, err)
			}
			defer out.Close()
			io.Copy(io.Discard, out)

			// 2. Then restart the container (which will use the new image)
			err = s.RestartContainer(ctx, appId)
			if err != nil {
				return fmt.Errorf("failed to upgrade container %v: %w", appId, err)
			}
		}
	}

	return nil
}
