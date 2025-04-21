package types

import (
	"context"
	"io"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// VPNStream is an interface for the network.Stream to make it easier to mock
type VPNStream interface {
	io.ReadWriteCloser
	Reset() error
	SetDeadline(time.Time) error
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

// Ensure network.Stream implements VPNStream
var _ VPNStream = (network.Stream)(nil)

// Service handles stream creation and management
type Service interface {
	// CreateNewVPNStream creates a new stream to a peer
	CreateNewVPNStream(ctx context.Context, peerID peer.ID) (VPNStream, error)
}

// PoolService handles stream pooling
type PoolService interface {
	// GetStream gets a stream from the pool or creates a new one
	GetStream(ctx context.Context, peerID peer.ID) (VPNStream, error)
	// ReleaseStream returns a stream to the pool
	ReleaseStream(peerID peer.ID, stream VPNStream, healthy bool)
}

// HealthService handles stream health checking and warming
//
// This interface combines the functionality of health checking and stream warming,
// providing a unified interface for managing stream health. It includes methods
// for starting and stopping the health checker and stream warmer, as well as
// retrieving health metrics.
type HealthService interface {
	// StartHealthChecker starts the health checker
	StartHealthChecker()
	// StopHealthChecker stops the health checker
	StopHealthChecker()
	// StartStreamWarmer starts the stream warmer
	StartStreamWarmer()
	// StopStreamWarmer stops the stream warmer
	StopStreamWarmer()
	// GetHealthMetrics returns the health metrics
	GetHealthMetrics() map[string]map[string]int64
}

// MetricsService handles metrics for the stream service
//
// This interface provides methods for retrieving metrics from different components
// of the stream service. It includes methods for getting stream pool metrics and
// health metrics, providing a unified interface for monitoring the service.
type MetricsService interface {
	// GetStreamPoolMetrics returns the stream pool metrics
	GetStreamPoolMetrics() map[string]int64
	// GetHealthMetrics returns the health metrics
	GetHealthMetrics() map[string]map[string]int64
}

// LifecycleService handles service lifecycle
//
// This interface provides methods for managing the lifecycle of a service,
// including starting and stopping the service. It's implemented by various
// components in the stream package to provide a consistent way to manage
// their lifecycle.
type LifecycleService interface {
	// Start starts the service
	Start()
	// Stop stops the service
	Stop()
}
