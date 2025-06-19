package api

import (
	"context"

	"github.com/unicornultrafoundation/subnet-node/core/deployer"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
)

type DeployerResponse[T any] struct {
	Data  T     `json:"data"`
	Error error `json:"error"`
}

type DeployerAPI struct {
	deployerService *deployer.Service
}

func NewDeployerAPI(deployerService *deployer.Service) *DeployerAPI {
	return &DeployerAPI{deployerService: deployerService}
}

func (api *DeployerAPI) RequestDeployment(ctx context.Context, req *types.DeploymentRequest) *DeployerResponse[*types.Deployment] {
	return WrapResponse(api.deployerService.RequestDeployment(ctx, req))
}

func (api *DeployerAPI) GetDeployment(ctx context.Context, orderID string) *DeployerResponse[*types.Deployment] {
	return WrapResponse(api.deployerService.GetDeployment(ctx, orderID))
}

func (api *DeployerAPI) CleanupDeployment(ctx context.Context, orderID string) *DeployerResponse[string] {
	return WrapResponse("Cleaned up deployment", api.deployerService.CleanupDeployment(ctx, orderID))
}

func (api *DeployerAPI) GetDeployments(ctx context.Context, requester string) *DeployerResponse[[]*types.Deployment] {
	return WrapResponse(api.deployerService.GetDeployments(ctx, requester))
}

func WrapResponse[T any](data T, err error) *DeployerResponse[T] {
	if err != nil {
		return &DeployerResponse[T]{Error: err}
	}
	return &DeployerResponse[T]{Data: data}
}
