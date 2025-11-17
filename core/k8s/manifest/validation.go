package manifest

import (
	"context"
	"fmt"

	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/manifest"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/sdl"
	dtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/deployment/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/manifest/v1"
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

	groupNames := make([]string, 0)

	for _, group := range groups {
		groupNames = append(groupNames, group.GetName())
	}

	// Check that hostnames are not in use
	if err = s.checkHostnamesForManifest(manifest, groupNames, deploymentID); err != nil {
		return err
	}

	return nil
}

func (m *Service) checkHostnamesForManifest(requestManifest mtypes.Manifest, groupNames []string, deploymentID dtypes.DeploymentID) error {
	// Check if the hostnames are available. Do not block forever
	ownerAddr, err := deploymentID.GetOwnerAddress()
	if err != nil {
		return err
	}

	allHostnames := make([]string, 0)

	for _, mgroup := range requestManifest.GetGroups() {
		for _, groupName := range groupNames {
			// Only check leases with a matching deployment ID & group name
			if groupName != mgroup.GetName() {
				continue
			}

			allHostnames = append(allHostnames, manifest.AllHostnamesOfManifestGroup(mgroup)...)
			if !m.config.HTTPServicesRequireAtLeastOneHost {
				continue
			}
			// For each service that exposes via an Ingress, then require a hostname
			for _, service := range mgroup.Services {
				for _, expose := range service.Expose {
					if expose.IsIngress() && len(expose.Hosts) == 0 {
						return fmt.Errorf("%w: service %q exposed on %d:%s must have a hostname",
							errManifestRejected, service.Name, expose.GetExternalPort(), expose.Proto)
					}
				}
			}
		}
	}

	return m.hostnameService.CanReserveHostnames(allHostnames, ownerAddr)
}
