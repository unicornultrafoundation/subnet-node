package manifest

import (
	"context"

	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/sdl"
	dtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/deployment/v1"
)

func (s *Service) ValidateRequest(ctx context.Context, deploymentID dtypes.DeploymentID, sdlManifest sdl.SDL) error {
	manifest, err := sdlManifest.Manifest()
	if err != nil {
		return err
	}

	err = manifest.Validate()
	if err != nil {
		return err
	}

	groups, err := sdlManifest.DeploymentGroups()
	if err != nil {
		return err
	}

	groupSpecs := make([]dtypes.GroupSpec, len(groups))
	for i, group := range groups {
		groupSpecs[i] = *group
	}

	if err = dtypes.ValidateDeploymentGroups(groupSpecs); err != nil {
		return err
	}

	if err = manifest.CheckAgainstGSpecs(groups); err != nil {
		return err
	}

	return nil
}
