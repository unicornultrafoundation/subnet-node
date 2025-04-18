package vpn

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

// PeerDiscoveryService handles peer discovery and mapping
type PeerDiscoveryService interface {
	// GetPeerID gets the peer ID for a destination IP
	GetPeerID(ctx context.Context, destIP string) (string, error)
}

// StreamService handles stream creation and management
type StreamService interface {
	// CreateNewVPNStream creates a new stream to a peer
	CreateNewVPNStream(ctx context.Context, peerID peer.ID) (VPNStream, error)
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
}

// VPNService is a composite interface that combines all the service interfaces
// This is provided for backward compatibility and for components that need all services
type VPNService interface {
	PeerDiscoveryService
	StreamService
	RetryService
	MetricsService
	ConfigService
}

// Ensure Service implements all interfaces
var (
	_ VPNService           = (*Service)(nil)
	_ PeerDiscoveryService = (*Service)(nil)
	_ StreamService        = (*Service)(nil)
	_ RetryService         = (*Service)(nil)
	_ MetricsService       = (*Service)(nil)
	_ ConfigService        = (*Service)(nil)
)
