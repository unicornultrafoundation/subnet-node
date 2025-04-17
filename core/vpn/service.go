package vpn

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/core/apps"
)

const VPNProtocol = "/vpn/1.0.0"

var log = logrus.WithField("service", "vpn")

type Service struct {
	cfg            *config.C
	enable         bool
	IsProvider     bool
	accountService *account.AccountService
	Apps           *apps.Service
	PeerHost       p2phost.Host
	DHT            *ddht.DHT
	virtualIP      string
	subnet         int
	routes         []string
	mtu            int
	exposedPorts   map[string]bool // All ports exposed by running apps
	unallowedPorts map[string]bool
	streamCache    map[string]network.Stream
	peerIDCache    *cache.Cache

	stopChan chan struct{} // Channel to stop background tasks
}

// Initializes the Service
func New(cfg *config.C, peerHost p2phost.Host, dht *ddht.DHT, apps *apps.Service, accountService *account.AccountService) *Service {
	unallowedPortList := cfg.GetStringSlice("vpn.unallowed_ports", []string{})
	unallowedPorts := make(map[string]bool, len(unallowedPortList))
	for _, port := range unallowedPortList {
		unallowedPorts[port] = true
	}

	return &Service{
		cfg:            cfg,
		Apps:           apps,
		PeerHost:       peerHost,
		accountService: accountService,
		enable:         cfg.GetBool("vpn.enable", false),
		mtu:            cfg.GetInt("vpn.mtu", 1400),
		virtualIP:      cfg.GetString("vpn.virtual_ip", ""),
		subnet:         cfg.GetInt("vpn.subnet", 8),
		routes:         cfg.GetStringSlice("vpn.routes", []string{"10.0.0.0/8"}),
		unallowedPorts: unallowedPorts,
		IsProvider:     cfg.GetBool("provider.enable", false),
		DHT:            dht,
		streamCache:    make(map[string]network.Stream),
		peerIDCache:    cache.New(1*time.Minute, 30*time.Second),
		exposedPorts:   make(map[string]bool),
		stopChan:       make(chan struct{}),
	}
}

func (s *Service) Start(ctx context.Context) error {
	if !s.enable {
		return nil
	}

	if s.virtualIP == "" {
		log.Error("virtual IP is not set")
		return nil
	}

	if len(s.routes) == 0 {
		log.Error("routes are not set")
		return nil
	}

	// Wait until there are some peers connected
	WaitUntilPeerConnected(ctx, s.PeerHost)
	time.Sleep(1 * time.Second)

	go func() {
		if s.IsProvider {
			// Update for the first time
			s.UpdateAllAppExposedPorts(ctx)

			go s.startUpdatingExposedPorts(ctx)
		}

		err := s.start(ctx)
		if err != nil {
			log.Errorf("Something went wrong when running VPN: %v", err)
		}
	}()

	return nil
}

func (s *Service) start(ctx context.Context) error {
	// Wait to connect to peer to sync DHT
	err := s.SyncPeerIDToDHT(ctx)
	if err != nil {
		return fmt.Errorf("failed to sync peer id to DHT: %v", err)
	}

	iface, err := s.SetupTUN()
	if err != nil {
		return fmt.Errorf("failed to setup TUN interface: %v", err)
	}
	defer iface.Close()

	s.HandleP2PTraffic(iface)

	// Listen packets from TUN interface
	return s.listenFromTUN(ctx, iface)
}

func (s *Service) Stop() error {
	if !s.enable || s.virtualIP == "" || len(s.routes) == 0 {
		return nil
	}

	if len(s.streamCache) > 0 {
		// Close all cached streams
		var wg sync.WaitGroup
		for _, stream := range s.streamCache {
			wg.Add(1)
			go func(str io.Closer) {
				defer wg.Done()
				str.Close()
			}(stream)
		}
		wg.Wait()
	}

	// Close stopChan to stop all background tasks
	close(s.stopChan)

	log.Infoln("VPN service stopped successfully!")
	return nil
}
