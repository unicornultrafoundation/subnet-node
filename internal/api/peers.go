package api

import (
	"context"

	"github.com/unicornultrafoundation/subnet-node/core/peers"
)

type PeersAPI struct {
	peersService *peers.Service
}

// NewPeersAPI creates a new instance of PeersAPI.
func NewPeersAPI(peersService *peers.Service) *PeersAPI {
	return &PeersAPI{peersService: peersService}
}

// GetPeers retrieves the Peers's Multiaddress of the appId
func (api *PeersAPI) GetAppPeers(ctx context.Context, appId string) ([]peers.PeerMultiAddress, error) {

	return api.peersService.GetAppPeers(ctx, appId)
}
