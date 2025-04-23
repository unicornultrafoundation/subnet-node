package vpn

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/packet"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/router"
)

// createDispatcher creates the appropriate packet dispatcher based on configuration
func (s *Service) createDispatcher() packet.DispatcherService {
	// Check if the StreamRouter is enabled
	routerIntegration := router.NewIntegration(s.cfg)
	if routerIntegration.IsEnabled() {
		log.Info("Using StreamRouter for packet dispatching")

		// Create a stream creator function that wraps the streamService's CreateNewVPNStream method
		createStreamFn := func(ctx context.Context, peerID peer.ID) (api.VPNStream, error) {
			return s.streamService.CreateNewVPNStream(ctx, peerID)
		}

		return routerIntegration.CreateDispatcherService(
			createStreamFn,
			s.peerDiscovery,
		)
	}

	// Use the default dispatcher
	log.Info("Using default dispatcher for packet dispatching")
	return packet.NewDispatcher(
		s.peerDiscovery,
		s.streamService,
		s.streamService,
		s.configService.GetWorkerIdleTimeout(),
		s.configService.GetWorkerCleanupInterval(),
		s.configService.GetWorkerBufferSize(),
		s.resilienceService,
	)
}
