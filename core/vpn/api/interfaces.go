package api

import (
	"context"
	"math/big"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
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
	// GetVirtualIP gets the virtual IP for a peer ID
	GetVirtualIP(ctx context.Context, peerID string) (string, error)
	// StoreMappingInDHT stores a mapping in the DHT
	StoreMappingInDHT(ctx context.Context, peerID string) error
	// VerifyVirtualIPHasRegistered verifies if a virtual IP is registered to the current peer ID
	VerifyVirtualIPHasRegistered(ctx context.Context, virtualIP string) error
	// GetPeerIDByRegistry gets a peer ID from the registry
	GetPeerIDByRegistry(ctx context.Context, destIP string) (string, error)
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
	// IncrementPacketsReceived increments the packets received counter
	IncrementPacketsReceived(bytes int)
	// IncrementCircuitOpenDrops increments the circuit open drops counter
	IncrementCircuitOpenDrops()
	// GetAllMetrics returns all metrics
	GetAllMetrics() map[string]int64
}

// DHTService is an interface for DHT operations
type DHTService interface {
	// GetValue gets a value from the DHT
	GetValue(ctx context.Context, key string) ([]byte, error)
	// PutValue puts a value in the DHT
	PutValue(ctx context.Context, key string, value []byte) error
}

// PeerstoreService is an interface for the peerstore service
type PeerstoreService interface {
	// PrivKey returns the private key for a peer ID
	PrivKey(p peer.ID) crypto.PrivKey
}

// HostService is an interface for the host service
type HostService interface {
	// ID returns the peer ID of the host
	ID() peer.ID
	// Peerstore returns the peerstore of the host
	Peerstore() PeerstoreService
}

// AccountService is an interface for the account service
type AccountService interface {
	// IPRegistry returns the IP registry contract
	IPRegistry() IPRegistry
}

// IPRegistry is an interface for the IP registry contract
type IPRegistry interface {
	// GetPeer gets the peer ID for a token ID
	GetPeer(opts any, tokenID *big.Int) (string, error)
}

// ConfigService handles configuration retrieval
type ConfigService interface {
	// GetWorkerBufferSize returns the worker buffer size
	GetWorkerBufferSize() int
	// GetWorkerIdleTimeout returns the worker idle timeout in seconds
	GetWorkerIdleTimeout() int
	// GetWorkerCleanupInterval returns the interval for worker cleanup
	GetWorkerCleanupInterval() time.Duration

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
	// GetEnable returns whether the VPN service is enabled
	GetEnable() bool
	// GetMTU returns the MTU for the TUN interface
	GetMTU() int
	// GetVirtualIP returns the virtual IP for this node
	GetVirtualIP() string
	// GetSubnet returns the subnet for the TUN interface
	GetSubnet() string
	// GetRoutes returns the routes for the TUN interface
	GetRoutes() []string
	// GetUnallowedPorts returns the unallowed ports
	GetUnallowedPorts() map[string]bool
	// GetRetryMaxAttempts returns the maximum number of retry attempts
	GetRetryMaxAttempts() int
	// GetRetryInitialInterval returns the initial retry interval
	GetRetryInitialInterval() time.Duration
	// GetRetryMaxInterval returns the maximum retry interval
	GetRetryMaxInterval() time.Duration
}
