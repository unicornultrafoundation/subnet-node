package peers

import (
	"context"
	"fmt"
	"net/http"

	gql "github.com/Khan/genqlient/graphql"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/graphql"
)

var log = logrus.WithField("service", "peers")

type Service struct {
	PeerHost p2phost.Host
	cfg      *config.C
	gqlUrl   string
	client   gql.Client
	enable   bool
}

// Initializes the Service
func New(cfg *config.C, peerHost p2phost.Host) *Service {
	return &Service{
		cfg:      cfg,
		PeerHost: peerHost,
		gqlUrl:   cfg.GetString("subgraph.graphql_url", ""),
		enable:   cfg.GetBool("subgraph.enable", false),
	}
}

func (s *Service) Start() error {
	if !s.enable || len(s.gqlUrl) == 0 {
		return nil
	}

	s.client = gql.NewClient(s.gqlUrl, http.DefaultClient)
	return nil
}

func (s *Service) Stop() error {
	return nil
}

func (s *Service) GetPeerMultiaddresses(peerID string) ([]string, error) {
	pid, err := peer.Decode(peerID)
	if err != nil {
		return nil, err
	}

	addrs := s.PeerHost.Peerstore().Addrs(pid)
	var multiAddrs []string
	for _, addr := range addrs {
		multiAddrs = append(multiAddrs, addr.String())
	}

	if len(multiAddrs) == 0 {
		return nil, fmt.Errorf("no multiaddress found for %s", peerID)
	}

	return multiAddrs, nil
}

func (s *Service) GetAppPeers(ctx context.Context, appId string) ([]PeerMultiAddress, error) {
	var peers []PeerMultiAddress

	peersData, err := graphql.GetAppPeers(ctx, s.client, appId)
	if err != nil {
		return peers, fmt.Errorf("failed to get app peers: %v", peersData)
	}

	if len(peersData.AppPeers) == 0 {
		return peers, nil
	}

	peers = make([]PeerMultiAddress, len(peersData.AppPeers))

	for index, peer := range peersData.AppPeers {
		peers[index].PeerID = peer.Peer.PeerId

		multiaddrs, err := s.GetPeerMultiaddresses(peer.Peer.PeerId)
		if err != nil {
			log.Debugf("failed to get multiaddrs for appId %v: %v", appId, err)
			continue
		}

		peers[index].MultiAddresses = multiaddrs
	}

	return peers, nil
}
