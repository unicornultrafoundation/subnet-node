package dispatch

import "time"

// Config contains configuration for the packet dispatcher
type Config struct {
	// Stream pool configuration
	MinStreamsPerPeer     int
	MaxStreamsPerPeer     int
	StreamIdleTimeout     time.Duration
	StreamCleanupInterval time.Duration

	// Worker pool configuration
	WorkerIdleTimeout     time.Duration
	WorkerCleanupInterval time.Duration
	WorkerBufferSize      int
	MaxWorkersPerPeer     int

	// Packet buffer size for stream channels
	PacketBufferSize int
}
