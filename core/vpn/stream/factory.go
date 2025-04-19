package stream

import (
	"github.com/unicornultrafoundation/subnet-node/core/vpn/config"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

// CreateStreamService creates a new stream service with all components
func CreateStreamService(
	streamService types.Service,
	vpnConfig *config.VPNConfig,
) *StreamService {
	// Create the stream service
	service := NewStreamService(
		streamService,
		vpnConfig.MaxStreamsPerPeer,
		vpnConfig.MinStreamsPerPeer,
		vpnConfig.StreamIdleTimeout,
		vpnConfig.CleanupInterval,
		vpnConfig.HealthCheckInterval,
		vpnConfig.HealthCheckTimeout,
		vpnConfig.MaxConsecutiveFailures,
		vpnConfig.WarmInterval,
		vpnConfig.MaxStreamsPerMultiplexer,
		vpnConfig.MinStreamsPerMultiplexer,
		vpnConfig.AutoScalingInterval,
		vpnConfig.MultiplexingEnabled,
	)

	return service
}

// CreateStreamServiceWithConfigService creates a new stream service with all components using a config service
func CreateStreamServiceWithConfigService(
	streamService types.Service,
	configService config.ConfigService,
) *StreamService {
	// Create the stream service
	service := NewStreamService(
		streamService,
		configService.GetMaxStreamsPerPeer(),
		configService.GetMinStreamsPerPeer(),
		configService.GetStreamIdleTimeout(),
		configService.GetCleanupInterval(),
		configService.GetHealthCheckInterval(),
		configService.GetHealthCheckTimeout(),
		configService.GetMaxConsecutiveFailures(),
		configService.GetWarmInterval(),
		configService.GetMaxStreamsPerMultiplexer(),
		configService.GetMinStreamsPerMultiplexer(),
		configService.GetAutoScalingInterval(),
		configService.GetMultiplexingEnabled(),
	)

	return service
}
