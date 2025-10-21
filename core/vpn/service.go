package vpn

import (
	"context"
	"net/netip"
	"sync"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/config"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/discovery"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatcher"
	ipmanager "github.com/unicornultrafoundation/subnet-node/core/vpn/ip_manager"
	vpnnetwork "github.com/unicornultrafoundation/subnet-node/core/vpn/network"
	"github.com/unicornultrafoundation/subnet-node/firewall"
)

var log = logrus.WithField("service", "vpn")

type Service struct {
	// dependencies
	ipManager        ipmanager.IPManager
	observer         ipmanager.IPManagerObserver
	configService    config.ConfigService
	discoveryService discovery.DiscoveryService
	dispatcher       dispatcher.DispatcherService
	peerHost         host.Host
	firewall         firewall.FirewallInterface

	// runtime state
	ip string
	mu sync.RWMutex

	// packet stack
	tun      *vpnnetwork.TUNService
	inbound  *vpnnetwork.InboundPacketService
	outbound *vpnnetwork.OutboundPacketService

	// outbound lifecycle context
	clientCtx    context.Context
	clientCancel context.CancelFunc

	stopChan chan struct{}
}

func NewService(ipManager ipmanager.IPManager, configService config.ConfigService, discoveryService discovery.DiscoveryService, dispatcherService dispatcher.DispatcherService, peerHost host.Host, firewall firewall.FirewallInterface) *Service {
	s := &Service{
		ipManager:        ipManager,
		configService:    configService,
		discoveryService: discoveryService,
		dispatcher:       dispatcherService,
		stopChan:         make(chan struct{}),
		peerHost:         peerHost,
		firewall:         firewall,
	}

	return s
}

func (s *Service) Start(ctx context.Context) error {
	observer, err := s.ipManager.WatchIP()
	if err != nil {
		return err
	}
	s.observer = observer

	// start watching for IP updates
	go func() {
		for {
			select {
			case ip := <-s.observer.GetChannel():
				s.handleIPUpdate(ip)
			case <-s.stopChan:
				return
			}
		}
	}()

	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	s.mu.Lock()
	// close observer channel safely
	if s.observer != nil {
		s.observer.Close()
		s.observer = nil
	}
	s.mu.Unlock()

	// teardown runtime stack
	s.teardownStack()

	// drop inbound reference
	if s.inbound != nil {
		s.peerHost.RemoveStreamHandler(protocol.ID(s.configService.GetProtocol()))
		_ = s.inbound.Close()
		s.inbound = nil
	}

	if s.tun != nil {
		err := s.firewall.RemoveNetwork(s.tun.GetCIDR())
		if err != nil {
			return err
		}
	}

	close(s.stopChan)
	return nil
}

// handleIPUpdate serializes and applies IP updates to the runtime stack.
func (s *Service) handleIPUpdate(newIP string) {
	s.mu.Lock()
	currentIP := s.ip
	s.mu.Unlock()

	// Ignore duplicate IP updates
	if newIP == currentIP {
		return
	}

	// Empty IP indicates we should tear down the stack
	if newIP == "" {
		log.Info("tearing down VPN stack on empty IP")
		s.teardownStack()
		s.mu.Lock()
		s.ip = ""
		s.mu.Unlock()
		return
	}

	log.WithFields(logrus.Fields{
		"currentIP": currentIP,
		"newIP":     newIP,
	}).Info("rebuilding VPN stack on IP change")

	if err := s.rebuildStack(newIP); err != nil {
		log.WithError(err).Error("failed to rebuild VPN stack on IP change")
		return
	}

	s.mu.Lock()
	s.ip = newIP
	s.mu.Unlock()
}

// teardownStack stops outbound readers first, then closes the TUN device.
func (s *Service) teardownStack() {
	s.mu.Lock()
	// cancel outbound context to stop readers cleanly
	if s.clientCancel != nil {
		s.clientCancel()
		s.clientCancel = nil
		s.clientCtx = nil
	}
	// drop outbound reference
	if s.outbound != nil {
		_ = s.outbound.Close()
		s.outbound = nil
	}

	// inbound has no long-lived goroutines; keep instance to rebind
	tun := s.tun
	s.tun = nil
	s.mu.Unlock()

	if tun != nil {
		_ = tun.Close()
	}
}

// buildTUN constructs a TUN service for the provided IP and sets it up.
func (s *Service) buildTUN(ip string) (*vpnnetwork.TUNService, error) {
	cfg := &vpnnetwork.TUNConfig{
		MTU:       s.configService.GetMTU(),
		VirtualIP: ip,
		Subnet:    s.configService.GetSubnet(),
		Routes:    s.configService.GetRoutes(),
		Routines:  s.configService.GetRoutines(),
	}
	tun := vpnnetwork.NewTUNService(cfg)
	if err := tun.SetupTUN(); err != nil {
		return nil, err
	}
	return tun, nil
}

// rebuildStack recreates the TUN and rewires services for the new IP.
func (s *Service) rebuildStack(newIP string) error {
	oldCidr := netip.Prefix{}
	s.mu.RLock()
	if s.ip == newIP && s.tun != nil {
		s.mu.RUnlock()
		return nil
	}
	if s.tun != nil {
		oldCidr = s.tun.GetCIDR()
	}
	s.mu.RUnlock()

	// Stop old stack first
	s.teardownStack()

	// Create and setup new TUN
	tun, err := s.buildTUN(newIP)
	if err != nil {
		return err
	}

	if oldCidr.String() != tun.GetCIDR().String() {
		err := s.firewall.AddNetwork(tun.GetCIDR())
		if err != nil {
			return err
		}
	}

	s.mu.Lock()
	s.tun = tun
	// bind inbound (create once)
	if s.inbound == nil {
		s.inbound = vpnnetwork.NewInboundPacketService(s.tun, s.configService, s.firewall)
		s.peerHost.SetStreamHandler(protocol.ID(s.configService.GetProtocol()), func(netStream network.Stream) {
			// Handle the incoming stream as a VPN stream
			s.inbound.HandleStream(netStream)
		})
	} else {
		s.inbound.SetTUNService(s.tun)
	}
	// start fresh outbound with new readers
	s.clientCtx, s.clientCancel = context.WithCancel(context.Background())
	s.outbound = vpnnetwork.NewOutboundPacketService(s.tun, s.dispatcher, s.configService, s.firewall)
	clientCtx := s.clientCtx
	outbound := s.outbound
	s.mu.Unlock()

	// start outbound outside the lock
	if err := outbound.Start(clientCtx); err != nil {
		return err
	}
	return nil
}
