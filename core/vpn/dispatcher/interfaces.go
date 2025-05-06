package dispatcher

import "context"

// DispatcherService defines the interface for packet dispatching
type DispatcherService interface {
	// DispatchPacket dispatches a packet to the appropriate stream
	// queueID is used to select the appropriate stream for the packet
	DispatchPacket(ctx context.Context, packet []byte, queueID int) error

	// Start starts the dispatcher
	Start()

	// Close stops the dispatcher
	Close() error
}
