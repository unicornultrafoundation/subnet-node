package router

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/pool"
)

var adapterLog = logrus.WithField("service", "vpn-router-adapter")

// RouterAdapter adapts the StreamRouter to implement the DispatcherService interface
type RouterAdapter struct {
	router *StreamRouter
}

// NewRouterAdapter creates a new router adapter
func NewRouterAdapter(router *StreamRouter) *RouterAdapter {
	return &RouterAdapter{
		router: router,
	}
}

// DispatchPacket implements the DispatcherService interface
func (a *RouterAdapter) DispatchPacket(ctx context.Context, syncKey, destIP string, packet []byte) {
	err := a.router.DispatchPacket(ctx, packet)
	if err != nil {
		adapterLog.Debugf("Failed to dispatch packet: %v", err)
	}
}

// Start implements the DispatcherService interface
func (a *RouterAdapter) Start() {
	// The router is already started when created, nothing to do here
	adapterLog.Info("Router adapter started")
}

// Stop implements the DispatcherService interface
func (a *RouterAdapter) Stop() {
	a.router.Shutdown()
	adapterLog.Info("Router adapter stopped")
}

// CreateStreamRouter creates a new StreamRouter with the given configuration and dependencies
func CreateStreamRouter(
	config *StreamRouterConfig,
	streamPool pool.PoolServiceExtension,
	peerDiscovery api.PeerDiscoveryService,
) *StreamRouter {
	return NewStreamRouter(config, streamPool, peerDiscovery)
}

// CreateRouterAdapter creates a new RouterAdapter with the given configuration and dependencies
func CreateRouterAdapter(
	config *StreamRouterConfig,
	streamPool pool.PoolServiceExtension,
	peerDiscovery api.PeerDiscoveryService,
) *RouterAdapter {
	router := CreateStreamRouter(config, streamPool, peerDiscovery)
	return NewRouterAdapter(router)
}
