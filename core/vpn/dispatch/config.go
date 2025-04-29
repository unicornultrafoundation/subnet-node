package dispatch

import "time"

// Config contains configuration for the packet dispatcher
type Config struct {
	// Stream pool configuration
	MaxStreamsPerPeer     int
	StreamIdleTimeout     time.Duration
	StreamCleanupInterval time.Duration

	// Packet buffer size for stream channels
	PacketBufferSize int

	// Load balancing configuration
	UsageCountWeight    float64
	BufferUtilWeight    float64
	BufferUtilThreshold int
	UsageCountThreshold int
}
