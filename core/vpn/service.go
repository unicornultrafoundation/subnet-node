package vpn

import (
	"time"

	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/docker"
)

const VPNProtocol = "/vpn/1.0.0"

var log = logrus.WithField("service", "vpn")

type Service struct {
	PeerHost         p2phost.Host
	cfg              *config.C
	enable           bool
	IsProvider       bool
	DHT              *ddht.DHT
	ConnectingPeerID peer.ID            // The PeerId that VPN Client is connecting to
	requestMap       map[string]peer.ID // Virtual IP -> Sender PeerID
	responseMap      map[string]string  // Actual application IP -> Virtual IP
	dockerClient     docker.DockerClient
	firstOctet       int

	stopChan chan struct{} // Channel to stop background tasks
}

// Initializes the Service
func New(cfg *config.C, peerHost p2phost.Host, dht *ddht.DHT, docker *docker.Service) *Service {
	return &Service{
		cfg:          cfg,
		PeerHost:     peerHost,
		enable:       cfg.GetBool("vpn.enable", false),
		IsProvider:   cfg.GetBool("provider.enable", false),
		DHT:          dht,
		dockerClient: *docker.GetClient(),
		requestMap:   make(map[string]peer.ID),
		responseMap:  make(map[string]string),
		stopChan:     make(chan struct{}),
	}
}

func (s *Service) Start() error {
	if !s.enable {
		return nil
	}

	s.firstOctet = 10

	go func() {
		// Wait to connect to peer to sync DHT
		time.Sleep(5 * time.Second)
		virtualIP, err := s.SyncPeerIDToDHT()
		if err != nil {
			log.Errorf("failed to sync peer id to DHT: %v", err)
			return
		}

		iface, err := s.CreateTUN(virtualIP, "8")
		if err != nil {
			log.Errorf("failed to create TUN: %v", err)
			return
		}
		defer iface.Close()

		s.HandleP2PTraffic(iface)

		connectPeerIdStr := s.cfg.GetString("vpn.default_peer_id", "")
		if len(connectPeerIdStr) != 0 {
			connectPeerId, err := peer.Decode(connectPeerIdStr)
			if err == nil {
				s.ConnectingPeerID = connectPeerId
			}
		}

		// Listen packets from TUN
		s.listenFromTUN(iface)
	}()

	log.Infof("VPN service started successfully!")
	return nil
}

func (s *Service) Stop() error {
	if !s.enable {
		return nil
	}

	// Close stopChan to stop all background tasks
	close(s.stopChan)

	log.Infof("VPN service stopped successfully!")
	return nil
}
