package apps

import (
	"context"
	"math/big"
)

func (s *Service) GetApp(ctx context.Context, appId *big.Int) (*App, error) {
	subnetApp, err := s.accountService.AppStore().Apps(nil, appId)
	if err != nil {
		return nil, err
	}

	appStatus, err := s.GetContainerStatus(ctx, appId)
	if err != nil {
		return nil, err
	}

	app := convertToApp(subnetApp, appId, appStatus)

	// Retrieve metadata from datastore if available
	metadata, err := s.GetContainerConfigProto(ctx, appId)
	if err == nil {
		if app.Metadata == nil {
			app.Metadata = new(AppMetadata)
		}
		app.Metadata.ContainerConfig.Env = metadata.ContainerConfig.Env
	}

	if appStatus == Running {
		ip, err := s.GetContainerIP(ctx, appId)
		if err != nil {
			return nil, err
		}
		app.IP = ip
	}

	return app, nil
}
