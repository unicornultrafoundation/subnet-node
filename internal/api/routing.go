package api

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/coreiface"
	"github.com/unicornultrafoundation/subnet-node/core/node/resource"
)

type RoutingAPI struct {
	api coreiface.RoutingAPI
}

// RoutingAPI creates a new instance of RoutingAPI.
func NewRoutingAPI(api coreiface.RoutingAPI) *RoutingAPI {
	return &RoutingAPI{api: api}
}

func (api *RoutingAPI) GetResource(ctx context.Context, peerID peer.ID) (resource.ResourceInfo, error) {
	return api.api.GetResource(ctx, peerID)
}

func (api *RoutingAPI) GetResources(ctx context.Context, peerIDs []peer.ID) ([]resource.ResourceInfo, error) {
	resources := []resource.ResourceInfo{}

	for _, peerID := range peerIDs {
		res, err := api.api.GetResource(ctx, peerID)
		if err != nil {
			return nil, err
		}
		resources = append(resources, res)
	}
	return resources, nil
}
