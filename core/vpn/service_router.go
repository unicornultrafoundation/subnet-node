package vpn

import (
	"github.com/unicornultrafoundation/subnet-node/core/vpn/packet"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/router"
)

// createDispatcher creates the appropriate packet dispatcher based on configuration
func (s *Service) createDispatcher() packet.DispatcherService {
	// Check if the StreamRouter is enabled
	routerIntegration := router.NewIntegration(s.cfg)
	if routerIntegration.IsEnabled() {
		log.Info("Using StreamRouter for packet dispatching")
		// Create an adapter that implements PoolServiceExtension
		poolAdapter := &StreamServicePoolAdapter{service: s}
		return routerIntegration.CreateDispatcherService(
			poolAdapter,
			s.peerDiscovery,
		)
	}

	// Use the default dispatcher
	log.Info("Using default dispatcher for packet dispatching")
	return packet.NewDispatcher(
		s.peerDiscovery,
		&StreamServiceAdapter{service: s},
		s.streamService,
		s.configService.GetWorkerIdleTimeout(),
		s.configService.GetWorkerCleanupInterval(),
		s.configService.GetWorkerBufferSize(),
		s.resilienceService,
	)
}
