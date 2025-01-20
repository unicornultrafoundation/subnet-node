package api

import (
	"context"

	"github.com/unicornultrafoundation/subnet-node/core"
	"github.com/unicornultrafoundation/subnet-node/core/apps"
	"github.com/unicornultrafoundation/subnet-node/core/node/resource"
)

type NodeAPI struct {
	resource *resource.Service
	node     *core.SubnetNode
	app      *apps.Service
}

// NewNodeAPI creates a new instance of NodeAPI.
func NewNodeAPI(resource *resource.Service, appS *apps.Service, node *core.SubnetNode) *NodeAPI {
	return &NodeAPI{resource: resource, node: node, app: appS}
}

func (api *NodeAPI) GetResource(ctx context.Context) (*resource.ResourceInfo, error) {
	return api.resource.GetResource(), nil
}

func (api *NodeAPI) GetPeerId(ctx context.Context) (string, error) {
	return api.resource.PeerId().String(), nil
}

func (api *NodeAPI) Restart(ctx context.Context) error {
	return api.node.Close()
}
