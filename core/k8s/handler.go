package k8s

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/unicornultrafoundation/subnet-node/core/k8s/kube/builder"
	apclient "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/provider/client"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/apitypes"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/sdl"
	dtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/deployment/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"

	kubeclienterrors "github.com/unicornultrafoundation/subnet-node/core/k8s/kube/errors"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
)

func (s *service) RequestDeployment(ctx context.Context, deploymentID dtypes.DeploymentID, sdlManifest sdl.SDL) error {
	return s.manifestService.Submit(ctx, deploymentID, sdlManifest)
}

func (s *service) GetAllLeaseStatus(ctx context.Context) ([]apitypes.DeploymentStatus, error) {
	ctx = context.WithValue(ctx, builder.SettingsKey, s.config.ClusterSettings[builder.SettingsKey])

	deploymentStatuses := make([]apitypes.DeploymentStatus, 0)

	for _, manager := range s.managers {
		leaseID := manager.deployment.LeaseID()
		lease, err := s.client.LeaseStatus(ctx, leaseID)
		if err != nil {
			return nil, err
		}
		manifestGroups := make([]apitypes.ManifestGroup, 0)

		for group, services := range lease {
			manifestGroups = append(manifestGroups, apitypes.ManifestGroup{
				Group:    group,
				Services: services,
			})
		}

		deploymentStatuses = append(deploymentStatuses, apitypes.DeploymentStatus{
			LeaseID:        leaseID,
			ManifestGroups: manifestGroups,
		})
	}
	return deploymentStatuses, nil
}

func (s *service) GetLeaseStatus(ctx context.Context, leaseID mtypes.LeaseID) (apclient.LeaseStatus, error) {
	ctx = context.WithValue(ctx, builder.SettingsKey, s.config.ClusterSettings[builder.SettingsKey])

	result := apclient.LeaseStatus{}

	found, manifestGroup, err := s.client.GetManifestGroup(ctx, leaseID)
	if err != nil {
		return result, err
	}

	if !found {
		return result, nil
	}

	hasForwardedPorts := false
portManifestGroupSearchLoop:
	for _, service := range manifestGroup.Services {
		for _, expose := range service.Expose {
			if expose.Global && expose.ExternalPort != 80 {
				hasForwardedPorts = true
				break portManifestGroupSearchLoop
			}
		}
	}
	if hasForwardedPorts {
		result.ForwardedPorts, err = s.client.ForwardedPortStatus(ctx, leaseID)
		if err != nil {
			return result, err
		}
	}

	result.Services, err = s.client.LeaseStatus(ctx, leaseID)
	if err != nil {
		if errors.Is(err, kubeclienterrors.ErrNoDeploymentForLease) {
			return result, err
		}
		if errors.Is(err, kubeclienterrors.ErrLeaseNotFound) {
			return result, err
		}
		if kubeErrors.IsNotFound(err) {
			return result, err
		}
		return result, err
	}

	return result, nil
}

func (s *service) DeleteDeployment(lid mtypes.LeaseID) error {
	managerKey := mtypes.LeaseIDToKey(lid)
	if manager := s.managers[managerKey]; manager != nil {
		if err := manager.teardown(); err != nil {
			return fmt.Errorf("tearing down lease deployment: %w", err)
		}
		return nil
	}

	// unreserve resources if no manager present yet.
	if strings.EqualFold(lid.Provider, s.session.Provider().Address().Hex()) {
		s.log.Info("unreserving unmanaged order", "lease", lid)
		err := s.inventory.unreserve(lid.OrderID())
		if err != nil && !errors.Is(errReservationNotFound, err) {
			return fmt.Errorf("unreserve failed: %w", err)
		}
	}

	return nil
}
