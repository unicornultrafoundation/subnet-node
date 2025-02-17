package apps

import (
	"context"
	"fmt"
	"math/big"

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
