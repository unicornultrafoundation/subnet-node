package stream

import (
	"time"

	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

// CreateStreamService creates a new stream service with all components
//
// This function accepts either a VPNConfig-like struct or a ConfigService-like interface
// to get the configuration values.
//
// The configuration source can be:
// 1. A struct with fields like MaxStreamsPerPeer, MinStreamsPerPeer, etc.
// 2. An interface with getter methods like GetMaxStreamsPerPeer(), GetMinStreamsPerPeer(), etc.
func CreateStreamService(
	streamService types.Service,
	config interface{},
) *StreamService {
	var (
		maxStreamsPerPeer      int
		minStreamsPerPeer      int
		streamIdleTimeout      time.Duration
		cleanupInterval        time.Duration
		healthCheckInterval    time.Duration
		healthCheckTimeout     time.Duration
		maxConsecutiveFailures int
		warmInterval           time.Duration
	)

	// Get configuration values based on the type of config provided
	switch c := config.(type) {
	case *struct {
		MaxStreamsPerPeer      int
		MinStreamsPerPeer      int
		StreamIdleTimeout      time.Duration
		CleanupInterval        time.Duration
		HealthCheckInterval    time.Duration
		HealthCheckTimeout     time.Duration
		MaxConsecutiveFailures int
		WarmInterval           time.Duration
	}:
		// Direct access to struct fields
		maxStreamsPerPeer = c.MaxStreamsPerPeer
		minStreamsPerPeer = c.MinStreamsPerPeer
		streamIdleTimeout = c.StreamIdleTimeout
		cleanupInterval = c.CleanupInterval
		healthCheckInterval = c.HealthCheckInterval
		healthCheckTimeout = c.HealthCheckTimeout
		maxConsecutiveFailures = c.MaxConsecutiveFailures
		warmInterval = c.WarmInterval
	case interface {
		GetMaxStreamsPerPeer() int
		GetMinStreamsPerPeer() int
		GetStreamIdleTimeout() time.Duration
		GetCleanupInterval() time.Duration
		GetHealthCheckInterval() time.Duration
		GetHealthCheckTimeout() time.Duration
		GetMaxConsecutiveFailures() int
		GetWarmInterval() time.Duration
	}:
		// Access through getter methods
		maxStreamsPerPeer = c.GetMaxStreamsPerPeer()
		minStreamsPerPeer = c.GetMinStreamsPerPeer()
		streamIdleTimeout = c.GetStreamIdleTimeout()
		cleanupInterval = c.GetCleanupInterval()
		healthCheckInterval = c.GetHealthCheckInterval()
		healthCheckTimeout = c.GetHealthCheckTimeout()
		maxConsecutiveFailures = c.GetMaxConsecutiveFailures()
		warmInterval = c.GetWarmInterval()
	default:
		// Use the package-level logger from service.go
		log.Error("Invalid configuration type provided to CreateStreamService")
		return nil
	}

	// Create the stream service
	service := NewStreamService(
		streamService,
		maxStreamsPerPeer,
		minStreamsPerPeer,
		streamIdleTimeout,
		cleanupInterval,
		healthCheckInterval,
		healthCheckTimeout,
		maxConsecutiveFailures,
		warmInterval,
	)

	return service
}
