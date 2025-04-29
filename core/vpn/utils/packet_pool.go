package utils

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// PacketData represents a packet with its data and metadata
type PacketData struct {
	// The actual packet data
	Data []byte
	// The size of the packet
	Size int
	// Whether the packet is in use
	inUse int32
}

// PacketPool is a pool of packet data structures
type PacketPool struct {
	// Size of each packet buffer
	bufferSize int
	// Number of packets to preallocate
	prealloc int
	// Pool of packet data
	pool []*PacketData
	// Mutex to protect the pool
	mu sync.Mutex
	// Stats
	stats struct {
		gets    int64
		puts    int64
		misses  int64
		creates int64
	}
}

// NewPacketPool creates a new packet pool
func NewPacketPool(bufferSize int, prealloc int) *PacketPool {
	if prealloc <= 0 {
		// Default to number of CPUs * 8 for preallocation
		prealloc = runtime.NumCPU() * 8
	}

	p := &PacketPool{
		bufferSize: bufferSize,
		prealloc:   prealloc,
		pool:       make([]*PacketData, 0, prealloc),
	}

	// Preallocate packets
	p.preallocate()

	return p
}

// preallocate creates and puts a number of packets into the pool
func (p *PacketPool) preallocate() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := 0; i < p.prealloc; i++ {
		packet := &PacketData{
			Data:  make([]byte, p.bufferSize),
			Size:  0,
			inUse: 0,
		}
		p.pool = append(p.pool, packet)
	}
	atomic.AddInt64(&p.stats.creates, int64(p.prealloc))
}

// Get gets a packet from the pool
func (p *PacketPool) Get() *PacketData {
	atomic.AddInt64(&p.stats.gets, 1)

	// Try to get a packet from the pool
	p.mu.Lock()
	var packet *PacketData
	for i := 0; i < len(p.pool); i++ {
		if atomic.CompareAndSwapInt32(&p.pool[i].inUse, 0, 1) {
			packet = p.pool[i]
			break
		}
	}
	p.mu.Unlock()

	// If no packet was available, create a new one
	if packet == nil {
		atomic.AddInt64(&p.stats.misses, 1)
		atomic.AddInt64(&p.stats.creates, 1)
		packet = &PacketData{
			Data:  make([]byte, p.bufferSize),
			Size:  0,
			inUse: 1,
		}
	}

	// Reset the packet
	packet.Size = 0
	// Clear the data to prevent data leakage
	for i := range packet.Data {
		packet.Data[i] = 0
	}

	return packet
}

// Put puts a packet back into the pool
func (p *PacketPool) Put(packet *PacketData) {
	if packet == nil {
		return
	}

	atomic.AddInt64(&p.stats.puts, 1)

	// Mark the packet as not in use
	atomic.StoreInt32(&packet.inUse, 0)

	// If the packet is not from our pool, add it
	p.mu.Lock()
	found := false
	for i := 0; i < len(p.pool); i++ {
		if p.pool[i] == packet {
			found = true
			break
		}
	}
	if !found {
		p.pool = append(p.pool, packet)
	}
	p.mu.Unlock()
}

// Stats returns statistics about the packet pool
func (p *PacketPool) Stats() map[string]int64 {
	return map[string]int64{
		"gets":    atomic.LoadInt64(&p.stats.gets),
		"puts":    atomic.LoadInt64(&p.stats.puts),
		"misses":  atomic.LoadInt64(&p.stats.misses),
		"creates": atomic.LoadInt64(&p.stats.creates),
		"size":    int64(len(p.pool)),
	}
}
