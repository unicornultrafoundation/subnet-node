package apps

import (
	"context"
	"math/big"

	dtypes "github.com/docker/docker/api/types"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

func (s *Service) GetApp(ctx context.Context, appId *big.Int) (*atypes.App, error) {
	// Fetch subnet app from cache or smart contract
	app, err := s.getSubnetAppByID(ctx, appId)
	if err != nil {
		return nil, err
	}

	if app.Metadata == nil {
		app.Metadata = new(atypes.AppMetadata)
	}

	// Add GitHub app data to the app
	gitHubApp, err := s.getGitHubAppByID(appId.Int64())
	if err != nil {
		return nil, err
	}

	app.Description = gitHubApp.Description
	app.Logo = gitHubApp.Logo
	app.BannerUrls = gitHubApp.BannerUrls
	app.DefaultBannerIndex = gitHubApp.DefaultBannerIndex
	app.Website = gitHubApp.Website
	app.Metadata.ContainerConfig = gitHubApp.Container

	// // Retrieve metadata from datastore if available
	// metadata, err := s.GetContainerConfigProto(ctx, appId)
	// if err == nil {
	// 	app.Metadata.ContainerConfig.Env = metadata.ContainerConfig.Env
	// }

	// Fetch the current status of the container
	appStatus, err := s.GetContainerStatus(ctx, appId)
	if err != nil {
		return nil, err
	}
	app.Status = appStatus

	if app.Status == atypes.Running {
		ip, err := s.GetContainerIP(ctx, appId)
		if err != nil {
			return nil, err
		}
		app.IP = ip
	}

	return app, nil
}

func (s *Service) ContainerInspect(ctx context.Context, appId *big.Int) (dtypes.ContainerJSON, error) {
	containerId := atypes.GetContainerIdFromAppId(appId)
	return s.dockerClient.ContainerInspect(ctx, containerId)
}
