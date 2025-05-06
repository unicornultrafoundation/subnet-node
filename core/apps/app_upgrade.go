package apps

import (
	"context"
	"fmt"
	"math/big"
	"time"
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
	runningAppList, err := s.GetRunningAppListProto(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch running app list for upgrading: %v", err)
	}

	for _, appIdBytes := range runningAppList.AppIds {
		appId := new(big.Int).SetBytes(appIdBytes)

		container, err := s.ContainerInspect(ctx, appId)
		if err != nil {
			log.Debugf("Failed to get container from appId %s: %v", appId, err)
			continue
		}

		// Get running Container image name and tag
		// Container.Image contains SHA256 digest, so we need to extract from Config.Image instead
		runningImageVersion := container.Config.Image

		// Get latest App image version from smart contract
		app, err := s.GetApp(ctx, appId)
		if err != nil {
			return fmt.Errorf("failed to fetch app details: %w", err)
		}

		latestImageVersion := app.Metadata.ContainerConfig.Image

		log.Debugf("Checking app %s - current image: %s, latest image: %s",
			app.Name, runningImageVersion, latestImageVersion)

		if runningImageVersion != latestImageVersion {
			log.Infof("Upgrading app %s from %s to %s",
				app.Name, runningImageVersion, latestImageVersion)
			return s.RestartContainer(ctx, appId) // this will upgrade the container with the new image version
		}
	}

	return nil
}
