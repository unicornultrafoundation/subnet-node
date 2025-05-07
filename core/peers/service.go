package peers

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	gql "github.com/Khan/genqlient/graphql"
	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/common/cidutil"
	"github.com/unicornultrafoundation/subnet-node/config"
)

var log = logrus.WithField("service", "peers")

type Service struct {
	PeerHost p2phost.Host
	cfg      *config.C
	gqlUrl   string
	client   gql.Client
	enable   bool
	dht      *ddht.DHT `optional:"true"`
}

// Initializes the Service
func New(cfg *config.C, peerHost p2phost.Host, dht *ddht.DHT) *Service {
	return &Service{
		cfg:      cfg,
		PeerHost: peerHost,
		gqlUrl:   cfg.GetString("subgraph.graphql_url", ""),
		enable:   cfg.GetBool("subgraph.enable", false),
		dht:      dht,
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
	appId = strings.TrimPrefix(appId, "0x")

	appIdInt, ok := new(big.Int).SetString(appId, 16)
	if !ok {
		return nil, fmt.Errorf("invalid app id: %s", appId)
	}

	// Get DHT key
	cid, err := cidutil.GenerateCID(appIdInt.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to generate DHT key: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	providers := s.dht.FindProvidersAsync(ctx, cid, 100)

	// Collect providers
	var result []PeerMultiAddress
	for p := range providers {
		multiAddrs, err := s.GetPeerMultiaddresses(p.ID.String())
		if err != nil {
			log.Debugf("failed to get multiaddrs for appId %v: %v", appId, err)
			continue
		}

		result = append(result, PeerMultiAddress{
			PeerID:         p.ID.String(),
			MultiAddresses: multiAddrs,
		})
	}

	return result, nil

}
