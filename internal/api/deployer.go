package api

import (
	"context"

	"github.com/unicornultrafoundation/subnet-node/core/deployer"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
)

type DeployerAPI struct {
	deployerService *deployer.Service
}

func NewDeployerAPI(deployerService *deployer.Service) *DeployerAPI {
	return &DeployerAPI{deployerService: deployerService}
}

func (api *DeployerAPI) RequestDeployment(ctx context.Context, req *types.DeploymentRequest) (*types.Deployment, error) {
	return api.deployerService.RequestDeployment(ctx, req)
}

func (api *DeployerAPI) GetDeployment(ctx context.Context, orderID string) (*types.Deployment, error) {
	return api.deployerService.GetDeployment(ctx, orderID)
}

func (api *DeployerAPI) CleanupDeployment(ctx context.Context, orderID string) error {
	return api.deployerService.CleanupDeployment(ctx, orderID)
}
