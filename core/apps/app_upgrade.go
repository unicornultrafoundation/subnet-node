package apps

import (
	"context"
	"fmt"
	"time"

	"github.com/containerd/containerd/namespaces"
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
	// Fetch all running containers
	ctx = namespaces.WithNamespace(ctx, NAMESPACE)
	containers, err := s.containerdClient.Containers(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch running containers: %v", err)
	}

	for _, container := range containers {
		// Get container ID (assuming appID is same as container ID)
		containerId := container.ID()
		appId, err := getAppIdFromContainerId(containerId)

		if err != nil {
			log.Errorf("Failed to get appId from containerId %s: %v", containerId, err)
			continue
		}

		// Get running Container image version
		containerInfo, err := container.Info(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch container info: %w", err)
		}
		runningImageVersion := containerInfo.Image

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
