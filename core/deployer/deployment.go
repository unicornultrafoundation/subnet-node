package deployer

import (
	"context"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
)

func (s *Service) RequestDeployment(ctx context.Context, deploymentRequest *types.DeploymentRequest) (*types.DeploymentResponse, error) {
	return s.deploymentService.RequestDeployment(ctx, deploymentRequest)
}

func (s *Service) GetDeployment(ctx context.Context, orderID string) (*types.DeploymentResponse, error) {
	return s.deploymentService.GetDeployment(ctx, orderID)
}

func (s *Service) CleanupDeployment(ctx context.Context, orderID string) error {
	return s.deploymentService.CleanupDeployment(ctx, orderID)
}

func (s *Service) GetDeployments(ctx context.Context, requester string) ([]*types.DeploymentResponse, error) {
	return s.deploymentService.GetDeployments(ctx, requester)
}
