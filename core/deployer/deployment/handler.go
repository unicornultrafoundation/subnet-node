package deployment

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	"go.uber.org/zap"
)

func (s *DeploymentService) RequestDeployment(ctx context.Context, deploymentRequest *types.DeploymentRequest) (*types.Deployment, error) {
	if err := deploymentRequest.Manifest.Validate(); err != nil {
		s.Logger.Error("Invalid deployment request", zap.Error(err))
		return nil, err
	}

	manifest, err := deploymentRequest.Manifest.ToManifest()
	if err != nil {
		s.Logger.Error("Failed to convert manifest to kubernetes manifest", zap.Error(err))
		return nil, err
	}

	requester := common.HexToAddress(deploymentRequest.Requester)

	deployment := types.Deployment{
		ID:        deploymentRequest.OrderID,
		Manifest:  manifest,
		Requester: requester,
	}

	// Deploy the deployment request
	if err := s.KubeClient.Deploy(ctx, deployment); err != nil {
		s.Logger.Error("Failed to deploy deployment request", zap.Error(err))
		return nil, err
	}

	s.Logger.Info("Deployment request deployed", zap.Any("order_id", deploymentRequest.OrderID))

	return s.KubeClient.GetDeployment(ctx, deploymentRequest.OrderID)
}

func (s *DeploymentService) GetDeployment(ctx context.Context, orderID string) (*types.Deployment, error) {
	return s.KubeClient.GetDeployment(ctx, orderID)
}

func (s *DeploymentService) CleanupDeployment(ctx context.Context, orderID string) error {
	return s.KubeClient.CleanupResources(ctx, orderID)
}
