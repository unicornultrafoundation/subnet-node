package proxy

import (
	"context"
	"fmt"
	"net"

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

	// ACL configuration
	allowAll    bool
	allowedNets []*net.IPNet
}

// Initializes the Peer Service.
func New(peerHost p2phost.Host, peerId peer.ID, cfg *config.C) (*Service, error) {
	parsedConfig, err := ParseProxyConfig(cfg.GetMap("proxy", map[string]any{}))
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxy config: %v", err)
	}

	service := &Service{
		PeerId:   peerId,
		PeerHost: peerHost,
		Cfg:      cfg,
		IsEnable: cfg.GetBool("proxy.enable", false),
		ProxyCfg: parsedConfig,
		stopChan: make(chan struct{}),
	}

	// Initialize ACL settings
	if err := service.initializeACL(); err != nil {
		return nil, fmt.Errorf("failed to initialize ACL: %v", err)
	}

	return service, nil
}

// initializeACL sets up ACL configuration from config
func (s *Service) initializeACL() error {
	// Check if allow_all is enabled
	s.allowAll = s.Cfg.GetBool("proxy.acl.allow_all", false)

	// Parse allowed networks from config
	allowedNetworks := s.Cfg.GetStringSlice("proxy.acl.allowed_networks", []string{})
	for _, network := range allowedNetworks {
		if _, ipnet, err := net.ParseCIDR(network); err == nil {
			s.allowedNets = append(s.allowedNets, ipnet)
		}
	}

	log.Infof("ACL initialized - allowAll: %v, allowedNets: %d",
		s.allowAll, len(s.allowedNets))

	return nil
}

// allowConnection implements ACL checking with allowAll and allowedNets only
func (s *Service) allowConnection(sourceIP, appID string, appAllowIPs []string) bool {
	// If allow_all is enabled, permit everything
	if s.allowAll {
		log.Debugf("Connection from %s allowed by allow_all policy (AppId: %s)", sourceIP, appID)
		return true
	}

	ip := net.ParseIP(sourceIP)
	if ip == nil {
		log.Warnf("Invalid IP format: %s", sourceIP)
		return false
	}

	// Check traditional app-level AllowIPs first (for backward compatibility)
	if len(appAllowIPs) > 0 && !isIPAllowed(sourceIP, appAllowIPs) {
		log.Debugf("Connection from %s denied by app AllowIPs (AppId: %s)", sourceIP, appID)
		return false
	}

	// Check if IP is in configured allowed networks
	for _, allowedNet := range s.allowedNets {
		if allowedNet.Contains(ip) {
			log.Debugf("Connection from %s allowed by configured network %s (AppId: %s)", sourceIP, allowedNet.String(), appID)
			return true
		}
	}

	// Default: deny if not in allowed networks and allow_all is false
	log.Debugf("Connection from %s denied by ACL policy (AppId: %s)", sourceIP, appID)
	return false
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
