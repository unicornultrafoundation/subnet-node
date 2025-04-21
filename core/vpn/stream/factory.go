package stream

import (
	"time"

	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

// StreamServiceConfig holds all configuration parameters for the stream service
type StreamServiceConfig struct {
	MinStreamsPerPeer      int
	StreamIdleTimeout      time.Duration
	CleanupInterval        time.Duration
	HealthCheckInterval    time.Duration
	HealthCheckTimeout     time.Duration
	MaxConsecutiveFailures int
	WarmInterval           time.Duration
}

// DefaultStreamServiceConfig returns a configuration with sensible defaults
func DefaultStreamServiceConfig() *StreamServiceConfig {
	return &StreamServiceConfig{
		MinStreamsPerPeer:      2,
		StreamIdleTimeout:      300 * time.Second, // 5 minutes
		CleanupInterval:        60 * time.Second,  // 1 minute
		HealthCheckInterval:    30 * time.Second,  // 30 seconds
		HealthCheckTimeout:     5 * time.Second,   // 5 seconds
		MaxConsecutiveFailures: 3,
		WarmInterval:           60 * time.Second, // 1 minute
	}
}

// CreateStreamService creates a new stream service with explicit configuration
func CreateStreamService(
	streamService types.Service,
	config *StreamServiceConfig,
) *StreamService {
	if config == nil {
		config = DefaultStreamServiceConfig()
	}

	return NewStreamService(
		streamService,
		config.MinStreamsPerPeer,
		config.StreamIdleTimeout,
		config.CleanupInterval,
		config.HealthCheckInterval,
		config.HealthCheckTimeout,
		config.MaxConsecutiveFailures,
		config.WarmInterval,
	)
}
