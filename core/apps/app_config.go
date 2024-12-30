package apps

import (
	"context"
	"fmt"
	"math/big"
)

// ...existing code...

func (s *Service) UpdateAppConfig(ctx context.Context, appId *big.Int, cfg ContainerConfig) error {
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
