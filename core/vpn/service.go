// Package vpn provides a secure, peer-to-peer Virtual Private Network implementation
// for the Subnet Node. It creates an encrypted overlay network that allows nodes to
// communicate securely regardless of their physical location or network configuration.
//
// The VPN service leverages libp2p for peer-to-peer communication and establishes
// TUN interfaces on participating nodes to route traffic through the secure overlay network.
//
// For detailed documentation, see the docs/vpn.md file.
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
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/sirupsen/logrus"
	"github.com/songgao/water"
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
	"github.com/unicornultrafoundation/subnet-node/core/vpn/utils"
)

// Logger for the VPN service
var log = logrus.WithField("service", "vpn")

// Service is the main VPN service that coordinates all VPN functionality.
// It manages the lifecycle of all VPN components and provides the main interface
// for starting, stopping, and monitoring the VPN service.
type Service struct {
	mu             sync.RWMutex
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
	resilienceService *resilience.ResilienceService
	metricsService    *metrics.MetricsServiceImpl
	bufferPool        *utils.BufferPool
	streamService     *stream.StreamService

	// Context management
	serviceCtx    context.Context
	serviceCancel context.CancelFunc

	// Resource management
	resourceManager *utils.ResourceManager

	// Stop channel for graceful shutdown
	stopChan chan struct{}
}

// New creates a new VPN service with the provided configuration and dependencies.
func New(cfg *config.C, peerHost host.Host, dht *ddht.DHT, accountService *account.AccountService) *Service {
	// Create the configuration service
	configService := vpnconfig.NewConfigService(cfg)

	// Create the buffer pool
	bufferPool := utils.NewBufferPool(configService.GetMTU())

	// Create the metrics service
	metricsService := metrics.NewMetricsService()

	// Create the resource manager
	resourceManager := utils.NewResourceManager()

	// Create the service
	service := &Service{
		cfg:             cfg,
		configService:   configService,
		isProvider:      cfg.GetBool("provider.enable", false),
		accountService:  accountService,
		peerHost:        peerHost,
		dht:             dht,
		metricsService:  metricsService,
		bufferPool:      bufferPool,
		resourceManager: resourceManager,
		stopChan:        make(chan struct{}),
	}

	// Create the peer discovery service directly from libp2p components
	service.peerDiscovery = discovery.NewPeerDiscoveryFromLibp2p(
		peerHost,
		dht,
		configService.GetVirtualIP(),
		accountService,
	)

	// Create the TUN service
	tunConfig := &vpnnetwork.TUNConfig{
		MTU:       configService.GetMTU(),
		VirtualIP: configService.GetVirtualIP(),
		Subnet:    configService.GetSubnet(),
		Routes:    configService.GetRoutes(),
	}
	service.tunService = vpnnetwork.NewTUNService(tunConfig)

	// Create the resilience service with configuration from the config service
	resilienceConfig := &resilience.ResilienceConfig{
		CircuitBreakerFailureThreshold: configService.GetCircuitBreakerFailureThreshold(),
		CircuitBreakerResetTimeout:     configService.GetCircuitBreakerResetTimeout(),
		CircuitBreakerSuccessThreshold: configService.GetCircuitBreakerSuccessThreshold(),
		RetryMaxAttempts:               configService.GetRetryMaxAttempts(),
		RetryInitialInterval:           configService.GetRetryInitialInterval(),
		RetryMaxInterval:               configService.GetRetryMaxInterval(),
	}
	service.resilienceService = resilience.NewResilienceService(resilienceConfig)

	// Create a stream service adapter that implements types.Service
	streamServiceAdapter := &StreamServiceAdapter{service: service}

	// Create a stream service config with values from the config service
	streamConfig := &stream.StreamServiceConfig{
		MinStreamsPerPeer:      configService.GetMinStreamsPerPeer(),
		StreamIdleTimeout:      configService.GetStreamIdleTimeout(),
		CleanupInterval:        configService.GetCleanupInterval(),
		HealthCheckInterval:    configService.GetHealthCheckInterval(),
		HealthCheckTimeout:     configService.GetHealthCheckTimeout(),
		MaxConsecutiveFailures: configService.GetMaxConsecutiveFailures(),
		WarmInterval:           configService.GetWarmInterval(),
	}

	// Create the stream service using the adapter and explicit config
	service.streamService = stream.CreateStreamService(streamServiceAdapter, streamConfig)

	// Register the stream service with the resource manager
	service.resourceManager.Register(service.streamService)

	// Create the packet dispatcher with stream pooling
	service.dispatcher = packet.NewDispatcher(
		service.peerDiscovery,
		streamServiceAdapter,
		service.streamService,
		configService.GetWorkerIdleTimeout(),
		configService.GetWorkerCleanupInterval(),
		configService.GetWorkerBufferSize(),
		service.resilienceService,
	)

	// Register the dispatcher with the resource manager
	service.resourceManager.Register(service.dispatcher)

	// Create a metrics adapter for the network services
	metricsAdapter := vpnnetwork.NewMetricsAdapter(service.metricsService)

	// Create the client service
	service.clientService = vpnnetwork.NewClientService(
		service.tunService,
		service.dispatcher,
		metricsAdapter,
		service.bufferPool,
	)

	// Register the client service with the resource manager
	service.resourceManager.Register(service.clientService)

	// Create the server service
	serverConfig := &vpnnetwork.ServerConfig{
		MTU:            configService.GetMTU(),
		UnallowedPorts: configService.GetUnallowedPorts(),
	}
	service.serverService = vpnnetwork.NewServerService(serverConfig, metricsAdapter)

	return service
}

// Start initializes and starts the VPN service.
// It waits for peer connections before setting up the TUN interface and other components.
// If the VPN service is disabled, this method returns immediately.
func (s *Service) Start(ctx context.Context) error {
	// Check if the service is enabled
	s.mu.RLock()
	if !s.configService.GetEnable() {
		s.mu.RUnlock()
		log.Infoln("VPN service is disabled")
		return nil
	}
	s.mu.RUnlock()

	// Lock for write to modify service state
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Infoln("Starting VPN service...")

	// Create a context with timeout for peer connection
	connectCtx, cancel := context.WithTimeout(ctx, s.configService.GetPeerConnectionTimeout())
	defer cancel()

	// Wait until there are some peers connected or timeout
	if err := s.waitUntilPeerConnected(connectCtx, s.peerHost); err != nil {
		return fmt.Errorf("failed to connect to peers: %w", err)
	}

	// Create a service context that can be cancelled when the service stops
	s.serviceCtx, s.serviceCancel = context.WithCancel(context.Background())

	// Create child contexts that inherit from the service context
	// When the service context is cancelled, these will be cancelled automatically
	dhtCtx := s.serviceCtx
	tunCtx := s.serviceCtx
	clientCtx := s.serviceCtx

	go func() {
		err := s.start(s.serviceCtx, dhtCtx, tunCtx, clientCtx)
		if err != nil {
			log.Errorf("Something went wrong when running VPN: %v", err)
		}
	}()

	return nil
}

// waitUntilPeerConnected blocks until at least one peer is connected to the host.
// It uses a simple retry loop with unlimited retry attempts and records metrics.
func (s *Service) waitUntilPeerConnected(ctx context.Context, host host.Host) error {
	log.Debug("Waiting for peer connection (unlimited attempts)")

	// Use a simple retry loop with unlimited attempts
	var attempts int
	breakerId := "peer_connection"

	// Manual retry loop with unlimited attempts
	for attempts = 1; ; attempts++ { // No upper limit
		// Check if we have any peers connected
		if len(host.Network().Peers()) > 0 {
			log.Infof("Connected to %d peers", len(host.Network().Peers()))
			// Record success for metrics purposes
			s.resilienceService.GetCircuitBreakerManager().GetBreaker(breakerId).RecordSuccess()
			return nil // Success, stop retrying
		}

		// Check if context is canceled
		if ctx.Err() != nil {
			return ctx.Err() // Exit the function on context cancellation
		}

		// No peers connected yet, wait before retrying
		log.Debugf("No peers connected yet, retry attempt %d (waiting for peers...)", attempts)
		select {
		case <-ctx.Done():
			return ctx.Err() // Exit the function on context cancellation
		case <-time.After(time.Duration(attempts) * 100 * time.Millisecond): // Simple backoff
			// Continue to next attempt
		}
	}
}

// start is the internal implementation of the Start method.
// It initializes and starts all VPN components in sequence.
func (s *Service) start(_, dhtCtx, tunCtx, clientCtx context.Context) error {
	// Start the stream service
	s.streamService.Start()

	// Start the packet dispatcher
	s.dispatcher.Start()

	// Create a context with timeout for DHT sync
	dhtSyncCtx, cancel := context.WithTimeout(dhtCtx, s.configService.GetDHTSyncTimeout())
	defer cancel()

	// Wait to connect to peer to sync DHT with retry
	err := s.syncPeerIDToDHTWithRetry(dhtSyncCtx)
	if err != nil {
		return fmt.Errorf("failed to sync peer ID to DHT: %w", err)
	}

	// Set up the TUN interface with retry
	tunSetupCtx, cancel := context.WithTimeout(tunCtx, s.configService.GetTUNSetupTimeout())
	defer cancel()

	// Setup TUN interface with retry
	iface, err := s.setupTUNWithRetry(tunSetupCtx)
	if err != nil {
		return fmt.Errorf("failed to setup TUN interface: %w", err)
	}

	// Register the TUN interface with the resource manager
	s.resourceManager.Register(iface)

	// Set up the stream handler for incoming P2P streams
	s.peerHost.SetStreamHandler(protocol.ID(s.configService.GetProtocol()), func(netStream network.Stream) {
		// Handle the incoming stream as a VPN stream
		s.serverService.HandleStream(netStream, iface)
	})

	// Start the client service with its own context and pass the existing TUN interface
	return s.clientService.Start(clientCtx, iface)
}

// Stop gracefully shuts down the VPN service and all its components.
// It logs metrics before stopping and uses the resource manager to close all resources.
func (s *Service) Stop() error {
	// Check if the service is enabled
	s.mu.RLock()
	if !s.configService.GetEnable() {
		s.mu.RUnlock()
		return nil
	}
	s.mu.RUnlock()

	// Lock for write to modify service state
	s.mu.Lock()
	defer s.mu.Unlock()

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

	// Cancel the service context to signal all components to stop
	// This will automatically cancel all child contexts
	if s.serviceCancel != nil {
		s.serviceCancel()
	}

	// Close stopChan to signal all background tasks to stop
	// This needs to happen before closing resources so components waiting on
	// this channel can begin their shutdown process
	close(s.stopChan)

	// Note: We rely on the resource manager to properly close all resources

	// Close all resources using the resource manager
	// This will close the client service, dispatcher, and stream service
	// since they're registered with the resource manager and implement io.Closer
	if err := s.resourceManager.Close(); err != nil {
		log.Warnf("Error closing resources: %v - continuing shutdown", err)
	}

	log.Infoln("VPN service stopped successfully!")
	return nil
}

// GetMetrics returns the current performance metrics for the VPN service.
func (s *Service) GetMetrics() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get all metrics from the metrics service
	return s.metricsService.GetAllMetrics()
}

// setupTUNWithRetry attempts to set up the TUN interface with circuit breaker and retry protection.
func (s *Service) setupTUNWithRetry(ctx context.Context) (*water.Interface, error) {
	log.Debug("Setting up TUN interface using resilience service with circuit breaker and retry protection")

	var iface *water.Interface

	// Use ExecuteWithResilience for better fault tolerance and metrics
	breakerId := "tun_setup"
	err, attempts := s.resilienceService.ExecuteWithResilience(ctx, breakerId, func() error {
		// Attempt to set up the TUN interface
		var err error
		iface, err = s.tunService.SetupTUN()
		if err != nil {
			log.Warnf("Failed to setup TUN interface with MTU %d, will retry: %v", s.configService.GetMTU(), err)
			return err // Return the error to trigger retry
		}

		log.Info("Successfully set up TUN interface")
		return nil // Success, stop retrying
	})

	if err != nil {
		log.Warnf("Failed to set up TUN interface after %d attempts: %v", attempts, err)
	} else if attempts > 1 {
		log.Infof("Successfully set up TUN interface after %d attempts", attempts)
	}

	return iface, err
}

// syncPeerIDToDHTWithRetry attempts to sync the peer ID to the DHT with circuit breaker and retry protection.
func (s *Service) syncPeerIDToDHTWithRetry(ctx context.Context) error {
	log.Debug("Syncing peer ID to DHT using resilience service with circuit breaker and retry protection")

	// Use ExecuteWithResilience for better fault tolerance and metrics
	breakerId := "dht_sync"
	err, attempts := s.resilienceService.ExecuteWithResilience(ctx, breakerId, func() error {
		// Attempt to sync peer ID to DHT
		err := s.peerDiscovery.SyncPeerIDToDHT(ctx)
		if err != nil {
			log.Warnf("Failed to sync peer ID to DHT, will retry: %v", err)
			return err // Return the error to trigger retry
		}

		log.Info("Successfully synced peer ID to DHT")
		return nil // Success, stop retrying
	})

	if err != nil {
		log.Warnf("Failed to sync peer ID to DHT after %d attempts: %v", attempts, err)
	} else if attempts > 1 {
		log.Infof("Successfully synced peer ID to DHT after %d attempts", attempts)
	}

	return err
}

// IsEnabled returns whether the VPN service is enabled
func (s *Service) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.configService.GetEnable()
}

// StreamServiceAdapter implements the types.Service interface to break the circular
// dependency between Service and StreamService. It allows the StreamService to create
// new VPN streams without directly depending on the main Service implementation.
type StreamServiceAdapter struct {
	service *Service
}

// CreateNewVPNStream implements the types.Service interface by creating a new
// libp2p stream to the specified peer using the VPN protocol with circuit breaker and retry protection.
func (a *StreamServiceAdapter) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (types.VPNStream, error) {
	// Use ExecuteWithResilience for better fault tolerance and metrics
	var stream types.VPNStream

	// Create a breaker ID for this peer operation
	breakerId := a.service.resilienceService.FormatPeerBreakerId(peerID, "create_stream")

	err, attempts := a.service.resilienceService.ExecuteWithResilience(ctx, breakerId, func() error {
		// Attempt to create a new stream to the peer
		var err error
		stream, err = a.service.peerHost.NewStream(ctx, peerID, protocol.ID(a.service.configService.GetProtocol()))
		if err != nil {
			log.Debugf("Failed to create P2P stream to peer %s, will retry: %v", peerID.String(), err)
			return err // Return the error to trigger retry
		}

		return nil // Success, stop retrying
	})

	if err != nil {
		log.Warnf("Failed to create P2P stream to peer %s after %d attempts: %v", peerID.String(), attempts, err)
		return nil, fmt.Errorf("failed to create P2P stream to peer %s after %d attempts: %w", peerID.String(), attempts, err)
	} else if attempts > 1 {
		log.Debugf("Successfully created P2P stream to peer %s after %d attempts", peerID.String(), attempts)
	}

	return stream, nil
}
