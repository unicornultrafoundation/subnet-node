package types

import (
	"context"
)

// BatchDispatcher is an interface for dispatchers that support batch processing
type BatchDispatcher interface {
	// DispatchPacketBatch dispatches a batch of packets
	DispatchPacketBatch(
		packets []*QueuedPacket,
		connKeys []ConnectionKey,
		destIPs []string,
		contexts []context.Context,
	) error
}
