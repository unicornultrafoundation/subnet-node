package types

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
)

// Global packet pool instance
var GlobalPacketPool = NewQueuedPacketPool()

// QueuedPacketPool is a pool of QueuedPacket objects
type QueuedPacketPool struct {
	// Pool of QueuedPacket objects
	pool sync.Pool
	// Stats
	stats struct {
		gets    int64
		puts    int64
		creates int64
	}
}

// NewQueuedPacketPool creates a new QueuedPacket pool
func NewQueuedPacketPool() *QueuedPacketPool {
	// Create the pool first
	p := &QueuedPacketPool{}

	// Then initialize the sync.Pool with a reference to p
	p.pool = sync.Pool{
		New: func() any {
			atomic.AddInt64(&p.stats.creates, 1)
			return &QueuedPacket{}
		},
	}

	// Preallocate some packets
	p.preallocate(runtime.NumCPU() * 16)

	return p
}

// preallocate creates and puts a number of packets into the pool
func (p *QueuedPacketPool) preallocate(count int) {
	for i := 0; i < count; i++ {
		packet := &QueuedPacket{}
		p.pool.Put(packet)
	}
	atomic.AddInt64(&p.stats.creates, int64(count))
}

// Get gets a packet from the pool
func (p *QueuedPacketPool) Get() *QueuedPacket {
	atomic.AddInt64(&p.stats.gets, 1)
	return p.pool.Get().(*QueuedPacket)
}

// GetWithData creates a packet with the provided data
func (p *QueuedPacketPool) GetWithData(ctx context.Context, destIP string, data []byte) *QueuedPacket {
	packet := p.Get()
	packet.Ctx = ctx
	packet.DestIP = destIP
	packet.Data = data

	// Set the ReturnToPool function to return the data to the appropriate pool
	packet.ReturnToPool = func(data []byte) {
		// This is a placeholder for actual buffer pool return logic
		// The actual implementation would depend on which buffer pool the data came from
		// For now, we just nil out the data to help with garbage collection
		// In a real implementation, this would call something like bufferPool.Put(data)
	}

	return packet
}

// Put puts a packet back into the pool
func (p *QueuedPacketPool) Put(packet *QueuedPacket) {
	if packet == nil {
		return
	}

	atomic.AddInt64(&p.stats.puts, 1)

	// Return the data to its original pool if a return function is provided
	if packet.ReturnToPool != nil && packet.Data != nil {
		packet.ReturnToPool(packet.Data)
	}

	// Clear the packet to prevent data leakage and memory leaks
	packet.Ctx = nil
	packet.DestIP = ""
	packet.Data = nil
	packet.ReturnToPool = nil

	// If there's a done channel, close it to prevent leaks
	if packet.DoneCh != nil {
		select {
		case packet.DoneCh <- nil:
			// Successfully sent nil (success) to the channel
		default:
			// Channel might be full or closed, which is fine
		}
		close(packet.DoneCh)
		packet.DoneCh = nil
	}

	p.pool.Put(packet)
}

// Stats returns statistics about the packet pool
func (p *QueuedPacketPool) Stats() map[string]int64 {
	return map[string]int64{
		"gets":    atomic.LoadInt64(&p.stats.gets),
		"puts":    atomic.LoadInt64(&p.stats.puts),
		"creates": atomic.LoadInt64(&p.stats.creates),
	}
}
