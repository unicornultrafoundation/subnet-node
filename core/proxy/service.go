package proxy

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"

	p2phost "github.com/libp2p/go-libp2p/core/host"
)

var log = logrus.WithField("service", "proxy")

type Service struct {
	PeerId   peer.ID
	Cfg      *config.C
	PeerHost p2phost.Host // the network host (server+client)
	IsEnable bool
	ProxyCfg ProxyConfig

	stopChan chan struct{} // Channel to stop background tasks
}

// Initializes the Peer Service.
func New(peerHost p2phost.Host, peerId peer.ID, cfg *config.C) (*Service, error) {
	parsedConfig, err := ParseProxyConfig(cfg.GetMap("proxy", map[interface{}]interface{}{}))

	if err != nil {
		return nil, fmt.Errorf("failed to parse proxy config: %v", err)
	}

	return &Service{
		PeerId:   peerId,
		PeerHost: peerHost,
		Cfg:      cfg,
		IsEnable: cfg.GetBool("proxy.enable", false),
		ProxyCfg: parsedConfig,
		stopChan: make(chan struct{}),
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	if !s.IsEnable || len(s.ProxyCfg.Peers) == 0 {
		return nil
	}

	// Start forwarding traffic for each mapping
	for _, peer := range s.ProxyCfg.Peers {
		for _, app := range peer.Apps {
			for _, ports := range app.ParsedPorts {
				go s.forwardTraffic(peer.ParsedId, app.ID, ports)
			}
		}
	}

	log.Info("Proxy Service started successfully.")
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	if !s.IsEnable || len(s.ProxyCfg.Peers) == 0 {
		return nil
	}

	log.Info("Stopping Proxy Service...")

	// Close stopChan to stop all background tasks
	close(s.stopChan)

	log.Info("Proxy Service stopped successfully.")
	return nil
}
