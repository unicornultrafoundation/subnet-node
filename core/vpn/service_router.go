package vpn

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/packet"
)

// createDispatcher creates the appropriate packet dispatcher based on configuration
func (s *Service) createDispatcher() packet.DispatcherService {
	// Use the new dispatcher from the dispatch package
	logrus.Info("Using new dispatcher for packet dispatching")

	// Create dispatcher configuration
	config := &dispatch.Config{
		// Stream pool configuration
		MaxStreamsPerPeer:     s.configService.GetMaxStreamsPerPeer(),
		StreamIdleTimeout:     s.configService.GetStreamIdleTimeout(),
		StreamCleanupInterval: s.configService.GetCleanupInterval(),
		PacketBufferSize:      s.configService.GetPacketBufferSize(),
	}

	// Create and return the new dispatcher
	dispatcher := dispatch.NewDispatcher(
		s.peerDiscovery,
		s.streamService,
		config,
		s.resilienceService,
	)

	// Wrap the new dispatcher with an adapter to implement the packet.DispatcherService interface
	return &dispatcherAdapter{
		dispatcher: dispatcher,
	}
}

// dispatcherAdapter adapts the new dispatcher to the packet.DispatcherService interface
type dispatcherAdapter struct {
	dispatcher interface {
		DispatchPacket(ctx context.Context, connKey types.ConnectionKey, destIP string, packet []byte) error
		Start()
		Stop()
	}
}

// DispatchPacket implements the packet.DispatcherService interface
func (a *dispatcherAdapter) DispatchPacket(ctx context.Context, syncKey, destIP string, packet []byte) error {
	// Convert the syncKey to a ConnectionKey
	// The syncKey format is typically sourcePort:destIP:destPort
	connKey := types.ConnectionKey(syncKey)

	// Dispatch the packet using the new dispatcher
	err := a.dispatcher.DispatchPacket(ctx, connKey, destIP, packet)
	if err != nil {
		logrus.Debugf("Failed to dispatch packet: %v", err)
	}
	return err
}

// Start implements the packet.DispatcherService interface
func (a *dispatcherAdapter) Start() {
	a.dispatcher.Start()
}

// Stop implements the packet.DispatcherService interface
func (a *dispatcherAdapter) Stop() {
	a.dispatcher.Stop()
}
