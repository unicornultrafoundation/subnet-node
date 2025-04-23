package packet

import (
	"context"
)

// DispatcherService defines the interface for packet dispatching
type DispatcherService interface {
	// DispatchPacket dispatches a packet to the appropriate worker
	DispatchPacket(ctx context.Context, connKey string, destIP string, packet []byte) error

	// Start starts the dispatcher
	Start()

	// Stop stops the dispatcher
	Stop()
}
