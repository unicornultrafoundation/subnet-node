package vpn

import (
	"context"
	"fmt"
	"sync"
	"time"

	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	vpnconfig "github.com/unicornultrafoundation/subnet-node/core/vpn/config"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/discovery"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/metrics"
	vpnnetwork "github.com/unicornultrafoundation/subnet-node/core/vpn/network"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/packet"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

const VPNProtocol = "/vpn/1.0.0"

var log = logrus.WithField("service", "vpn")

// Service is the main VPN service
type Service struct {
	mu             sync.Mutex
	cfg            *config.C
	configService  vpnconfig.ConfigService
	isProvider     bool
	accountService *account.AccountService
	peerHost       host.Host
	dht            *ddht.DHT

	// Core components
	peerDiscovery     *discovery.PeerDiscovery
	tunService        *vpnnetwork.TUNService
	clientService     *vpnnetwork.ClientService
	serverService     *vpnnetwork.ServerService
	dispatcher        *packet.Dispatcher
	circuitBreakerMgr *resilience.CircuitBreakerManager
	retryManager      *resilience.RetryManager
	metricsService    *metrics.MetricsServiceImpl
	bufferPool        *vpnnetwork.BufferPool
	streamService     *stream.StreamService

	// Stop channel for graceful shutdown
	stopChan chan struct{}
}

// New creates a new VPN service
func New(cfg *config.C, peerHost host.Host, dht *ddht.DHT, accountService *account.AccountService) *Service {
	// Create the configuration service
	configService := vpnconfig.NewConfigService(cfg)

	// Create the buffer pool
	bufferPool := vpnnetwork.NewBufferPool(configService.GetMTU(), 100)

	// Create the metrics service
	metricsService := metrics.NewMetricsService()

	// Create the service
	service := &Service{
		cfg:            cfg,
		configService:  configService,
		isProvider:     cfg.GetBool("provider.enable", false),
		accountService: accountService,
		peerHost:       peerHost,
		dht:            dht,
		metricsService: metricsService,
		bufferPool:     bufferPool,
		stopChan:       make(chan struct{}),
	}

	// Create the peer discovery service
	service.peerDiscovery = discovery.NewPeerDiscovery(peerHost, dht, configService.GetVirtualIP())

	// Create the TUN service
	tunConfig := &vpnnetwork.TUNConfig{
		MTU:       configService.GetMTU(),
		VirtualIP: configService.GetVirtualIP(),
		Subnet:    configService.GetSubnet(),
		Routes:    configService.GetRoutes(),
	}
	service.tunService = vpnnetwork.NewTUNService(tunConfig)

	// Create the circuit breaker manager
	service.circuitBreakerMgr = resilience.NewCircuitBreakerManager(
		configService.GetCircuitBreakerFailureThreshold(),
		configService.GetCircuitBreakerResetTimeout(),
		configService.GetCircuitBreakerSuccessThreshold(),
	)

	// Create the retry manager
	service.retryManager = resilience.NewRetryManager(
		configService.GetRetryMaxAttempts(),
		configService.GetRetryInitialInterval(),
		configService.GetRetryMaxInterval(),
	)

	// Create a stream service adapter that implements types.Service
	streamServiceAdapter := &StreamServiceAdapter{service: service}

	// Create the stream service using the adapter
	service.streamService = stream.CreateStreamServiceWithConfigService(streamServiceAdapter, configService)

	// Create the packet dispatcher with enhanced stream services
	service.dispatcher = packet.NewEnhancedDispatcher(
		service.peerDiscovery,
		streamServiceAdapter,
		service.streamService,
		service.streamService,
		configService.GetWorkerIdleTimeout(),
		configService.GetWorkerCleanupInterval(),
		configService.GetWorkerBufferSize(),
		configService.GetMultiplexingEnabled(),
	)

	// Create a metrics adapter for the network services
	metricsAdapter := vpnnetwork.NewMetricsAdapter(service.metricsService)

	// Create the client service
	service.clientService = vpnnetwork.NewClientService(
		service.tunService,
		service.dispatcher,
		metricsAdapter,
		service.bufferPool,
	)

	// Create the server service
	serverConfig := &vpnnetwork.ServerConfig{
		MTU:            configService.GetMTU(),
		UnallowedPorts: configService.GetUnallowedPorts(),
	}
	service.serverService = vpnnetwork.NewServerService(serverConfig, metricsAdapter)

	return service
}

// Start starts the VPN service
func (s *Service) Start(ctx context.Context) error {
	if !s.configService.GetEnable() {
		return nil
	}

	// Wait until there are some peers connected
	waitUntilPeerConnected(ctx, s.peerHost)
	time.Sleep(1 * time.Second)

	// Start the packet dispatcher
	s.dispatcher.Start()

	// Start the stream service
	s.streamService.Start()

	go func() {
		err := s.start(ctx)
		if err != nil {
			log.Errorf("Something went wrong when running VPN: %v", err)
		}
	}()

	return nil
}

// waitUntilPeerConnected waits until at least one peer is connected
func waitUntilPeerConnected(ctx context.Context, host host.Host) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if len(host.Network().Peers()) > 0 {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (s *Service) start(ctx context.Context) error {
	// Wait to connect to peer to sync DHT
	err := s.peerDiscovery.SyncPeerIDToDHT(ctx)
	if err != nil {
		return fmt.Errorf("failed to sync peer id to DHT: %v", err)
	}

	// Set up the TUN interface
	iface, err := s.tunService.SetupTUN()
	if err != nil {
		return fmt.Errorf("failed to setup TUN interface: %v", err)
	}
	defer iface.Close()

	// Set up the stream handler for incoming P2P traffic
	s.peerHost.SetStreamHandler(VPNProtocol, func(netStream network.Stream) {
		// Use the stream directly as a VPNStream
		s.serverService.HandleStream(netStream, iface)
	})

	// Start the stream service
	s.streamService.Start()

	// Start the packet dispatcher
	s.dispatcher.Start()

	// Start the client service
	return s.clientService.Start(ctx)
}

// Stop stops the VPN service
func (s *Service) Stop() error {
	if !s.configService.GetEnable() {
		return nil
	}

	// Log metrics before stopping
	metrics := s.metricsService.GetAllMetrics()
	log.WithFields(logrus.Fields{
		"packets_received": metrics["packets_received"],
		"packets_sent":     metrics["packets_sent"],
		"packets_dropped":  metrics["packets_dropped"],
		"bytes_received":   metrics["bytes_received"],
		"bytes_sent":       metrics["bytes_sent"],
		"stream_errors":    metrics["stream_errors"],
	}).Info("VPN metrics")

	// Log circuit breaker metrics before stopping
	log.WithFields(logrus.Fields{
		"circuit_open_count":  metrics["circuit_open_count"],
		"circuit_close_count": metrics["circuit_close_count"],
		"circuit_reset_count": metrics["circuit_reset_count"],
		"request_block_count": metrics["request_block_count"],
		"request_allow_count": metrics["request_allow_count"],
		"active_breakers":     metrics["active_breakers"],
	}).Info("Circuit breaker metrics")

	// Stop the client service
	if err := s.clientService.Stop(); err != nil {
		log.Errorf("Error stopping client service: %v", err)
	}

	// Stop the packet dispatcher
	s.dispatcher.Stop()

	// Stop the stream service
	s.streamService.Stop()

	// Close stopChan to stop all background tasks
	close(s.stopChan)

	log.Infoln("VPN service stopped successfully!")
	return nil
}

// GetMetrics returns the current VPN metrics
func (s *Service) GetMetrics() map[string]int64 {
	// Get all metrics from the metrics service
	return s.metricsService.GetAllMetrics()
}

// StreamServiceAdapter is an adapter that implements the types.Service interface
// It's used to break the circular dependency between Service and StreamService
type StreamServiceAdapter struct {
	service *Service
}

// CreateNewVPNStream implements the types.Service interface
func (a *StreamServiceAdapter) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (types.VPNStream, error) {
	// Create a new stream to the peer
	stream, err := a.service.peerHost.NewStream(ctx, peerID, VPNProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to create P2P stream: %v", err)
	}

	return stream, nil
}
