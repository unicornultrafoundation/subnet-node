package network

import (
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/config"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/utils"
)

// Note: We're using the standard BufferPool for now
// In the future, we could implement an adapter to use the EnhancedBufferPool

// CreateOptimizedClientService creates a new client service with optimized components
func CreateOptimizedClientService(cfg *config.VPNConfig, streamService api.StreamService, peerDiscovery api.PeerDiscoveryService) (*ClientService, error) {
	// Create a standard buffer pool
	bufferSize := 2048 // Default buffer size
	bufferPool := utils.NewBufferPool(bufferSize)

	// Create a TUN service
	tunConfig := &TUNConfig{
		MTU:       cfg.MTU,
		VirtualIP: cfg.VirtualIP,
		Subnet:    strconv.Itoa(cfg.Subnet),
		Routes:    cfg.Routes,
	}
	tunService := NewTUNService(tunConfig)

	// Create a dispatcher config
	dispatcherConfig := &dispatch.Config{
		MaxStreamsPerPeer:     cfg.MaxStreamsPerPeer,
		StreamIdleTimeout:     cfg.StreamIdleTimeout,
		StreamCleanupInterval: cfg.CleanupInterval,
		PacketBufferSize:      cfg.PacketBufferSize,
		UsageCountWeight:      cfg.UsageCountWeight,
		BufferUtilWeight:      cfg.BufferUtilWeight,
		BufferUtilThreshold:   cfg.BufferUtilThreshold,
		UsageCountThreshold:   cfg.UsageCountThreshold,
	}

	// Create resilience service
	resilienceConfig := &resilience.ResilienceConfig{
		CircuitBreakerFailureThreshold: cfg.CircuitBreakerFailureThreshold,
		CircuitBreakerResetTimeout:     cfg.CircuitBreakerResetTimeout,
		CircuitBreakerSuccessThreshold: cfg.CircuitBreakerSuccessThreshold,
		RetryMaxAttempts:               cfg.RetryMaxAttempts,
		RetryInitialInterval:           cfg.RetryInitialInterval,
		RetryMaxInterval:               cfg.RetryMaxInterval,
	}
	resilienceService := resilience.NewResilienceService(resilienceConfig)

	// Create the dispatcher
	dispatcher := dispatch.NewDispatcher(peerDiscovery, streamService, dispatcherConfig, resilienceService)

	logrus.Info("Created optimized client service with enhanced dispatcher")

	// Create the client service
	return NewClientService(tunService, dispatcher, bufferPool), nil
}

// CreateOptimizedServerService creates a new server service with optimized components
func CreateOptimizedServerService(cfg *config.VPNConfig) (*ServerService, error) {
	// Create server config
	serverConfig := &ServerConfig{
		MTU:            cfg.MTU,
		UnallowedPorts: cfg.UnallowedPorts,
	}

	// Create the server service
	return NewServerService(serverConfig), nil
}
