package router

import (
	"context"
	"time"

	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/packet"
)

// Factory creates StreamRouter instances
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

// CreateRouterAdapter creates a DispatcherService implementation using the StreamRouter
func (f *Factory) CreateRouterAdapter(
	createStreamFn StreamCreator,
	peerDiscovery api.PeerDiscoveryService,
) packet.DispatcherService {
	// Get configuration from config service
	configService := NewConfigService(f.cfg)
	config := configService.GetStreamRouterConfig()

	// Create the router with the stream creator function
	router := NewStreamRouter(config, createStreamFn, peerDiscovery)

	// Create and return the adapter
	return &RouterAdapter{
		router: router,
	}
}

// RouterAdapter adapts the StreamRouter to the DispatcherService interface
type RouterAdapter struct {
	router *StreamRouter
}

// DispatchPacket implements the DispatcherService interface
func (a *RouterAdapter) DispatchPacket(ctx context.Context, syncKey, destIP string, packetData []byte) {
	// We ignore the error here because the DispatcherService interface doesn't return an error
	_ = a.router.DispatchPacket(ctx, packetData)
}

// Start implements the DispatcherService interface
func (a *RouterAdapter) Start() {
	// The router starts automatically when created
}

// Stop implements the DispatcherService interface
func (a *RouterAdapter) Stop() {
	a.router.Shutdown()
}
