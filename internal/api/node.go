package api

import (
	"context"

	"github.com/unicornultrafoundation/subnet-node/core/node/resource"
)

type NodeAPI struct {
	resource *resource.Service
}

// NewNodeAPI creates a new instance of NodeAPI.
func NewNodeAPI(resource *resource.Service) *NodeAPI {
	return &NodeAPI{resource: resource}
}

func (api *NodeAPI) GetResource(ctx context.Context) (*resource.ResourceInfo, error) {
	return api.resource.GetResource(), nil
}
