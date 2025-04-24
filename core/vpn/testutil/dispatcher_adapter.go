package testutil

import (
	"context"
	"time"

	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
)

// DispatcherAdapter adapts the dispatch.Dispatcher to the packet.Dispatcher interface
type DispatcherAdapter struct {
	dispatcher *dispatch.Dispatcher
}

// NewDispatcher creates a new dispatcher adapter
func NewDispatcher(
	discoveryService api.PeerDiscoveryService,
	streamService api.StreamService,
	poolService api.StreamPoolService,
	resilienceService *resilience.ResilienceService,
) *DispatcherAdapter {
	// Create a dispatcher config
	config := &dispatch.Config{
		MaxStreamsPerPeer:     10,
		StreamIdleTimeout:     5 * time.Minute,
		StreamCleanupInterval: 1 * time.Minute,
		PacketBufferSize:      100,
	}

	// Create a new dispatcher
	dispatcher := dispatch.NewDispatcher(
		discoveryService,
		streamService,
		config,
		resilienceService,
	)

	return &DispatcherAdapter{
		dispatcher: dispatcher,
	}
}

// DispatchPacket dispatches a packet to the appropriate worker
func (a *DispatcherAdapter) DispatchPacket(ctx context.Context, syncKey, destIP string, packet []byte) error {
	// Convert the syncKey to a ConnectionKey
	connKey := types.ConnectionKey(syncKey)

	// Dispatch the packet
	return a.dispatcher.DispatchPacket(ctx, connKey, destIP, packet)
}

// DispatchPacketWithCallback dispatches a packet and provides a callback channel for the result
func (a *DispatcherAdapter) DispatchPacketWithCallback(ctx context.Context, syncKey, destIP string, packet []byte, doneCh chan error) error {
	// Convert the syncKey to a ConnectionKey
	connKey := types.ConnectionKey(syncKey)

	// Dispatch the packet with callback
	return a.dispatcher.DispatchPacketWithCallback(ctx, connKey, destIP, packet, doneCh)
}

// Start starts the dispatcher
func (a *DispatcherAdapter) Start() {
	a.dispatcher.Start()
}

// Stop stops the dispatcher
func (a *DispatcherAdapter) Stop() {
	a.dispatcher.Stop()
}

// GetMetrics returns the dispatcher's metrics
func (a *DispatcherAdapter) GetMetrics() map[string]int64 {
	return a.dispatcher.GetMetrics()
}
