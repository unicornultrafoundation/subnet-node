package k8s

import (
	"context"
	"errors"
	"fmt"

	"github.com/unicornultrafoundation/subnet-node/core/k8s/kube/builder"
	etypes "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/expiry"
	apclient "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/provider/client"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/apitypes"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/sdl"
	dtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/deployment/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"

	kubeclienterrors "github.com/unicornultrafoundation/subnet-node/core/k8s/kube/errors"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
)

func (s *service) RequestDeployment(ctx context.Context, deploymentID dtypes.DeploymentID, sdlManifest sdl.SDL) error {
	// Check if the deployment already has a active lease
	status, err := s.expiryService.CheckDeploymentExpiry(ctx, fmt.Sprintf("%d", deploymentID.DSeq))
	if err != nil {
		return err
	}
	if status.Status != etypes.DeploymentExpiryStatusActive {
		message := fmt.Sprintf("Deployment does not have a active lease with ID %d. Lease status: %s", deploymentID.DSeq, status.Status)
		if status.Status == etypes.DeploymentExpiryStatusExpired {
			message = fmt.Sprintf("Lease %s is expired. You need to add more funds to this lease to keep it active or it will be deleted in %d seconds", deploymentID.DSeq, status.TimeLeft)
		}
		return errors.New(message)
	}
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

		expiryStatus, err := s.expiryService.CheckDeploymentExpiry(ctx, fmt.Sprintf("%d", leaseID.DSeq))
		if err != nil {
			return nil, fmt.Errorf("failed to check deployment expiry: %w", err)
		}

		deploymentStatuses = append(deploymentStatuses, apitypes.DeploymentStatus{
			LeaseID:        leaseID,
			ManifestGroups: manifestGroups,
			Status:         expiryStatus.Status,
			TimeLeft:       expiryStatus.TimeLeft,
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
			if expose.Global {
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

	expiryStatus, err := s.expiryService.CheckDeploymentExpiry(ctx, fmt.Sprintf("%d", leaseID.DSeq))
	if err != nil {
		return result, err
	}
	result.Status = expiryStatus.Status
	result.TimeLeft = expiryStatus.TimeLeft

	return result, nil
}

func (s *service) DeleteDeployment(lid mtypes.LeaseID) error {
	if err := s.bus.Publish(mtypes.EventLeaseClosed{
		ID: lid,
	}); err != nil {
		return fmt.Errorf("send lease closed request failed: %w", err)
	}

	return nil
}
