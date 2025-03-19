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
	PeerId       peer.ID
	Cfg          *config.C
	PeerHost     p2phost.Host `optional:"true"` // the network host (server+client)
	IsEnable     bool
	Ports        []PortMapping
	RemotePeerId peer.ID
	AppId        string

	stopChan chan struct{} // Channel to stop background tasks
}

// Initializes the Peer Service.
func New(peerHost p2phost.Host, peerId peer.ID, cfg *config.C) *Service {
	return &Service{
		PeerId:   peerId,
		PeerHost: peerHost,
		Cfg:      cfg,
		IsEnable: cfg.GetBool("proxy.enable", false),
		AppId:    cfg.GetString("proxy.app_id", ""),
		stopChan: make(chan struct{}),
	}
}

func (s *Service) Start(ctx context.Context) error {
	if !s.IsEnable {
		return nil
	}

	encodedPeerId := s.Cfg.GetString("proxy.peer_id", "")
	if encodedPeerId == "" || len(encodedPeerId) == 0 {
		return fmt.Errorf("proxy.peer_id was not set in config")
	}
	peerId, err := peer.Decode(encodedPeerId)
	if err != nil {
		return fmt.Errorf("failed to decode peerID %s: %v", encodedPeerId, err)
	}
	s.RemotePeerId = peerId

	if s.AppId == "" || len(s.AppId) == 0 {
		return fmt.Errorf("proxy.app_id was not set in config")
	}

	// Valdiate port mapping
	portMappings := s.Cfg.GetStrings("proxy.ports", []string{})
	if len(portMappings) == 0 {
		return fmt.Errorf("proxy.ports was not set in config")
	}
	s.Ports, err = ParsePortMappings(portMappings)

	if err != nil {
		return fmt.Errorf("failed to parse port mapping from Proxy Service config: %v", err)
	}

	// Start forwarding traffic for each mapping
	for _, mapping := range s.Ports {
		go s.forwardTraffic(mapping)
	}

	log.Info("Proxy Service started successfully.")
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	if !s.IsEnable {
		return nil
	}

	log.Info("Stopping Proxy Service...")

	// Close stopChan to stop all background tasks
	close(s.stopChan)

	log.Info("Proxy Service stopped successfully.")
	return nil
}
