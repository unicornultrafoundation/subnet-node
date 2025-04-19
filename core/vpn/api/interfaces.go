package api

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

// VPNService is the main interface for the VPN system
type VPNService interface {
	// Start starts the VPN service
	Start(ctx context.Context) error
	// Stop stops the VPN service
	Stop() error
	// GetMetrics returns the current VPN metrics
	GetMetrics() map[string]int64
}

// PeerDiscoveryService handles peer discovery and mapping
type PeerDiscoveryService interface {
	// GetPeerID gets the peer ID for a destination IP
	GetPeerID(ctx context.Context, destIP string) (string, error)
	// SyncPeerIDToDHT syncs the peer ID to the DHT
	SyncPeerIDToDHT(ctx context.Context) error
}

// StreamService handles stream creation and management
type StreamService interface {
	// CreateNewVPNStream creates a new stream to a peer
	CreateNewVPNStream(ctx context.Context, peerID peer.ID) (types.VPNStream, error)
}

// StreamPoolService handles stream pooling
type StreamPoolService interface {
	// GetStream gets a stream from the pool or creates a new one
	GetStream(ctx context.Context, peerID peer.ID) (types.VPNStream, error)
	// ReleaseStream returns a stream to the pool
	ReleaseStream(peerID peer.ID, stream types.VPNStream, healthy bool)
}

// CircuitBreakerService handles circuit breaker operations
type CircuitBreakerService interface {
	// Execute executes an operation with circuit breaker protection
	Execute(destination string, operation func() error) error
	// Reset resets a circuit breaker for a destination
	Reset(destination string)
	// GetBreakerStates returns the states of all circuit breakers
	GetBreakerStates() map[string]resilience.CircuitState
}

// StreamHealthService handles stream health monitoring
type StreamHealthService interface {
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

// StreamMultiplexService handles stream multiplexing
type StreamMultiplexService interface {
	// StartMultiplexer starts the multiplexer manager
	StartMultiplexer()
	// StopMultiplexer stops the multiplexer manager
	StopMultiplexer()
	// SendPacket sends a packet using the multiplexer
	SendPacketMultiplexed(ctx context.Context, peerID peer.ID, packet []byte) error
	// GetMultiplexerMetrics returns the multiplexer metrics
	GetMultiplexerMetrics() map[string]*types.MultiplexerMetrics
}

// RetryService handles retry operations with backoff
type RetryService interface {
	// RetryOperation retries an operation with exponential backoff
	RetryOperation(ctx context.Context, operation func() error) error
}

// MetricsService handles metrics collection
type MetricsService interface {
	// IncrementStreamErrors increments the stream errors counter
	IncrementStreamErrors()
	// IncrementPacketsSent increments the packets sent counter
	IncrementPacketsSent(bytes int)
	// IncrementPacketsDropped increments the packets dropped counter
	IncrementPacketsDropped()
}

// ConfigService handles configuration retrieval
type ConfigService interface {
	// GetWorkerBufferSize returns the worker buffer size
	GetWorkerBufferSize() int
	// GetWorkerIdleTimeout returns the worker idle timeout in seconds
	GetWorkerIdleTimeout() int
	// GetWorkerCleanupInterval returns the interval for worker cleanup
	GetWorkerCleanupInterval() time.Duration
	// GetMaxStreamsPerPeer returns the maximum number of streams per peer
	GetMaxStreamsPerPeer() int
	// GetMinStreamsPerPeer returns the minimum number of streams per peer
	GetMinStreamsPerPeer() int
	// GetStreamIdleTimeout returns the stream idle timeout
	GetStreamIdleTimeout() time.Duration
	// GetCircuitBreakerFailureThreshold returns the circuit breaker failure threshold
	GetCircuitBreakerFailureThreshold() int
	// GetCircuitBreakerResetTimeout returns the circuit breaker reset timeout
	GetCircuitBreakerResetTimeout() time.Duration
	// GetCircuitBreakerSuccessThreshold returns the circuit breaker success threshold
	GetCircuitBreakerSuccessThreshold() int
	// GetHealthCheckInterval returns the health check interval
	GetHealthCheckInterval() time.Duration
	// GetHealthCheckTimeout returns the health check timeout
	GetHealthCheckTimeout() time.Duration
	// GetMaxConsecutiveFailures returns the maximum consecutive failures
	GetMaxConsecutiveFailures() int
	// GetWarmInterval returns the warm interval
	GetWarmInterval() time.Duration
	// GetMaxStreamsPerMultiplexer returns the maximum number of streams per multiplexer
	GetMaxStreamsPerMultiplexer() int
	// GetMinStreamsPerMultiplexer returns the minimum number of streams per multiplexer
	GetMinStreamsPerMultiplexer() int
	// GetAutoScalingInterval returns the auto-scaling interval
	GetAutoScalingInterval() time.Duration
	// GetMultiplexingEnabled returns whether multiplexing is enabled
	GetMultiplexingEnabled() bool
}
