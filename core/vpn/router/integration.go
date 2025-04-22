package router

import (
	"time"

	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/packet"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/pool"
)

// Integration provides methods to integrate the StreamRouter with the VPN service
type Integration struct {
	cfg interface {
		GetBool(key string, defaultValue bool) bool
		GetInt(key string, defaultValue int) int
		GetString(key string, defaultValue string) string
		GetDuration(key string, defaultValue time.Duration) time.Duration
	}
}

// NewIntegration creates a new integration
func NewIntegration(cfg interface {
	GetBool(key string, defaultValue bool) bool
	GetInt(key string, defaultValue int) int
	GetString(key string, defaultValue string) string
	GetDuration(key string, defaultValue time.Duration) time.Duration
}) *Integration {
	return &Integration{
		cfg: cfg,
	}
}

// CreateDispatcherService creates a DispatcherService implementation using the StreamRouter
func (i *Integration) CreateDispatcherService(
	streamPool pool.PoolServiceExtension,
	peerDiscovery api.PeerDiscoveryService,
) packet.DispatcherService {
	// Create the factory
	factory := NewFactory(i.cfg)

	// Create the router adapter
	return factory.CreateRouterAdapter(streamPool, peerDiscovery)
}

// IsEnabled returns whether the StreamRouter is enabled
func (i *Integration) IsEnabled() bool {
	return i.cfg.GetBool("vpn.router.enable", false)
}
