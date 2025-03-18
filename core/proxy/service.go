package proxy

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
	Ports        []string
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
		Ports:    cfg.GetStringSlice("proxy.ports", []string{}),
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
	if len(s.Ports) == 0 {
		return fmt.Errorf("proxy.ports was not set in config")
	}
	isPortMappingValid := s.validatePorts()

	if !isPortMappingValid {
		log.Fatal("Invalid port mapping from Proxy Service config")
	}

	// Connect to docker daemon
	s.dockerClient, err = dockerCli.NewClientWithOpts(dockerCli.FromEnv, dockerCli.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("error connecting to docker: %v", err)
	}

	// Start forwarding traffic for each mapping
	for _, mapping := range s.Ports {
		ports := strings.Split(mapping, ":")
		go s.forwardTraffic(ports[0], s.AppId, ports[1])
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

func (s *Service) validatePorts() bool {
	portMap := make(map[string]bool)
	hasDuplicate := false
	hasInvalidPort := false

	for _, mapping := range s.Ports {
		parts := strings.Split(mapping, ":")
		if len(parts) != 2 {
			fmt.Println("❌ Invalid port mapping format:", mapping)
			hasInvalidPort = true
			continue
		}

		localPort := parts[0] // Extract local port

		// Validate that the local port is a number and within range
		portNum, err := strconv.Atoi(localPort)
		if err != nil || portNum < 1 || portNum > 65535 {
			fmt.Println("❌ Invalid port number:", localPort)
			hasInvalidPort = true
			continue
		}

		// Check for duplicate local ports
		if portMap[localPort] {
			hasDuplicate = true
		}
		portMap[localPort] = true
	}

	return !(hasDuplicate || hasInvalidPort)
}
