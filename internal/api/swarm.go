package api

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/unicornultrafoundation/subnet-node/core/coreiface"
)

type SwarmAPI struct {
	api coreiface.SwarmAPI
}

// NewAdminAPI creates a new instance of SwarmAPI.
func NewSwarmAPI(api coreiface.SwarmAPI) *SwarmAPI {
	return &SwarmAPI{api: api}
}

type PeersResult struct {
	Peers []PeerResult `json:"peers"`
}

type PeerResult struct {
	Addr      string            `json:"addr"`
	Peer      string            `json:"peer"`
	Latency   time.Duration     `json:"latency"`
	Direction network.Direction `json:"direction"`
	Streams   []StreamResult    `json:"streams"`
}

type StreamResult struct {
	Protocol string `json:"protocol"`
}

func (api *SwarmAPI) Peers(ctx context.Context) (*PeersResult, error) {
	peers, err := api.api.Peers(ctx)
	if err != nil {
		return nil, err
	}
	peersResult := &PeersResult{
		Peers: []PeerResult{},
	}
	for _, peer := range peers {
		latency, err := peer.Latency()
		if err != nil {
			return nil, err
		}
		peerResult := PeerResult{
			Addr:      peer.Address().String(),
			Peer:      peer.ID().String(),
			Latency:   latency,
			Direction: peer.Direction(),
			Streams:   []StreamResult{},
		}

		streams, err := peer.Streams()
		if err != nil {
			return nil, err
		}
		for _, stream := range streams {
			peerResult.Streams = append(peerResult.Streams, StreamResult{
				string(stream),
			})
		}

		peersResult.Peers = append(peersResult.Peers, peerResult)
	}

	return peersResult, nil
}
