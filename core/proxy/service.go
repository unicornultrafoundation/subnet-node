package proxy

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/p2p"

	dockerCli "github.com/docker/docker/client"
	p2phost "github.com/libp2p/go-libp2p/core/host"
)

var log = logrus.WithField("service", "proxy")

type Service struct {
	peerId       peer.ID
	cfg          *config.C
	dockerClient *dockerCli.Client
	P2P          *p2p.P2P
	PeerHost     p2phost.Host `optional:"true"` // the network host (server+client)
	IsEnable     bool
	Ports        []PortMapping
	RemotePeerId peer.ID
	AppId        string

	stopChan chan struct{} // Channel to stop background tasks
}

// Initializes the Peer Service.
func New(peerHost p2phost.Host, peerId peer.ID, cfg *config.C, P2P *p2p.P2P) *Service {
	return &Service{
		peerId:   peerId,
		PeerHost: peerHost,
		P2P:      P2P,
		cfg:      cfg,
		IsEnable: cfg.GetBool("proxy.enable", false),
		AppId:    cfg.GetString("proxy.app_id", ""),
		stopChan: make(chan struct{}),
	}
}

func (s *Service) Start(ctx context.Context) error {
	if !s.IsEnable {
		return nil
	}

	encodedPeerId := s.cfg.GetString("proxy.peer_id", "")
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
	portMappings := s.cfg.GetStringSlice("proxy.ports", []string{})
	if len(portMappings) == 0 {
		return fmt.Errorf("proxy.ports was not set in config")
	}
	s.Ports, err = ParsePortMappings(portMappings)

	if err != nil {
		return fmt.Errorf("failed to parse port mapping from Proxy Service config: %v", err)
	}

	// Connect to docker daemon
	s.dockerClient, err = dockerCli.NewClientWithOpts(dockerCli.FromEnv, dockerCli.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("error connecting to docker: %v", err)
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

	// Close the docker client
	if s.dockerClient != nil {
		err := s.dockerClient.Close()
		if err != nil {
			return fmt.Errorf("failed to close docker client: %w", err)
		}
	}

	log.Info("Proxy Service stopped successfully.")
	return nil
}
