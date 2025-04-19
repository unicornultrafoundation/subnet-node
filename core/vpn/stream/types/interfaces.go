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

// HealthCheckerService handles stream health checking
type HealthCheckerService interface {
	// StartHealthChecker starts the health checker
	StartHealthChecker()
	// StopHealthChecker stops the health checker
	StopHealthChecker()
}

// StreamWarmerService handles stream warming
type StreamWarmerService interface {
	// StartStreamWarmer starts the stream warmer
	StartStreamWarmer()
	// StopStreamWarmer stops the stream warmer
	StopStreamWarmer()
}

// HealthMetricsService handles health metrics
type HealthMetricsService interface {
	// GetHealthMetrics returns the health metrics
	GetHealthMetrics() map[string]map[string]int64
}

// MultiplexerService handles stream multiplexing operations
type MultiplexerService interface {
	// StartMultiplexer starts the multiplexer manager
	StartMultiplexer()
	// StopMultiplexer stops the multiplexer manager
	StopMultiplexer()
	// SendPacketMultiplexed sends a packet using the multiplexer
	SendPacketMultiplexed(ctx context.Context, peerID peer.ID, packet []byte) error
}

// MultiplexerMetricsService handles multiplexer metrics
type MultiplexerMetricsService interface {
	// GetMultiplexerMetrics returns the multiplexer metrics
	GetMultiplexerMetrics() map[string]*MultiplexerMetrics
}

// StreamPoolMetricsService handles stream pool metrics
type StreamPoolMetricsService interface {
	// GetStreamPoolMetrics returns the stream pool metrics
	GetStreamPoolMetrics() map[string]int64
}

// StreamLifecycleService handles stream service lifecycle
type StreamLifecycleService interface {
	// Start starts the stream service
	Start()
	// Stop stops the stream service
	Stop()
}

// HealthService is a composite interface for backward compatibility
type HealthService interface {
	HealthCheckerService
	StreamWarmerService
	HealthMetricsService
}

// MultiplexService is a composite interface for backward compatibility
type MultiplexService interface {
	MultiplexerService
	MultiplexerMetricsService
}
