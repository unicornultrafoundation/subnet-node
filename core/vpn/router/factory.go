package router

import (
	"time"

	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/pool"
)

// Factory creates StreamRouter components
type Factory struct {
	cfg interface {
		GetBool(key string, defaultValue bool) bool
		GetInt(key string, defaultValue int) int
		GetString(key string, defaultValue string) string
		GetDuration(key string, defaultValue time.Duration) time.Duration
	}
}

// NewFactory creates a new factory
func NewFactory(cfg interface {
	GetBool(key string, defaultValue bool) bool
	GetInt(key string, defaultValue int) int
	GetString(key string, defaultValue string) string
	GetDuration(key string, defaultValue time.Duration) time.Duration
}) *Factory {
	return &Factory{
		cfg: cfg,
	}
}

// CreateRouterAdapter creates a new RouterAdapter with the given dependencies
func (f *Factory) CreateRouterAdapter(
	streamPool pool.PoolServiceExtension,
	peerDiscovery api.PeerDiscoveryService,
) *RouterAdapter {
	// Create the config service
	configService := NewConfigService(f.cfg)

	// Get the router config
	routerConfig := configService.GetStreamRouterConfig()

	// Create the router
	router := NewStreamRouter(routerConfig, streamPool, peerDiscovery)

	// Create the adapter
	return NewRouterAdapter(router)
}
