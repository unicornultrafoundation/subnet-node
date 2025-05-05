package dispatcher

import "context"

// DispatcherService defines the interface for packet dispatching
type DispatcherService interface {
	// DispatchPacket dispatches a packet to the appropriate stream
	// queueID is used to select the appropriate stream for the packet
	DispatchPacket(ctx context.Context, packet []byte, queueID int) error

	// Start starts the dispatcher
	Start()

	// Stop stops the dispatcher
	Stop()

	// GetMetrics returns the current metrics for the dispatcher
	GetMetrics() map[string]uint64
}
