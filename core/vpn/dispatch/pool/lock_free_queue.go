package pool

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
)

// node represents a node in the lock-free queue
type node struct {
	value *types.QueuedPacket
	next  unsafe.Pointer
}

// nodePool is a pool of nodes to reduce GC pressure
type nodePool struct {
	pool sync.Pool
	// Stats for monitoring
	stats struct {
		gets    int64
		puts    int64
		creates int64
	}
}

// newNodePool creates a new node pool
func newNodePool() *nodePool {
	p := &nodePool{}
	p.pool = sync.Pool{
		New: func() any {
			atomic.AddInt64(&p.stats.creates, 1)
			return &node{}
		},
	}
	return p
}

// get gets a node from the pool
func (p *nodePool) get() *node {
	atomic.AddInt64(&p.stats.gets, 1)
	n := p.pool.Get().(*node)

	// Ensure the node is clean
	if n.value != nil || n.next != nil {
		// Node wasn't properly cleaned, create a new one
		n = &node{}
		atomic.AddInt64(&p.stats.creates, 1)
	}

	return n
}

// put puts a node back into the pool
func (p *nodePool) put(n *node) {
	if n == nil {
		return
	}

	atomic.AddInt64(&p.stats.puts, 1)

	// Clear the node to prevent memory leaks
	n.value = nil
	n.next = nil
	p.pool.Put(n)
}

// getStats returns statistics about the node pool
func (p *nodePool) getStats() map[string]int64 {
	return map[string]int64{
		"gets":    atomic.LoadInt64(&p.stats.gets),
		"puts":    atomic.LoadInt64(&p.stats.puts),
		"creates": atomic.LoadInt64(&p.stats.creates),
	}
}

// Global node pool
var globalNodePool = newNodePool()

// LockFreeQueue is a lock-free queue implementation for overflow packets
type LockFreeQueue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
	size int64

	// Stats
	stats struct {
		enqueues int64
		dequeues int64
		nodeGets int64
		nodePuts int64
	}
}

// NewLockFreeQueue creates a new lock-free queue
func NewLockFreeQueue() *LockFreeQueue {
	// Get a sentinel node from the pool
	n := globalNodePool.get()
	n.next = nil
	n.value = nil

	return &LockFreeQueue{
		head: unsafe.Pointer(n),
		tail: unsafe.Pointer(n),
		size: 0,
	}
}

// Enqueue adds a packet to the queue
func (q *LockFreeQueue) Enqueue(packet *types.QueuedPacket) {
	// Track enqueue operation
	atomic.AddInt64(&q.stats.enqueues, 1)

	// Get a node from the pool
	n := globalNodePool.get()
	n.value = packet
	n.next = nil

	// Track node allocation
	atomic.AddInt64(&q.stats.nodeGets, 1)

	for {
		tail := load(&q.tail)
		next := load(&tail.next)
		if tail == load(&q.tail) { // Are tail and next consistent?
			if next == nil {
				// Try to link node at the end of the linked list
				if cas(&tail.next, next, n) {
					// Enqueue is done. Try to swing tail to the inserted node
					cas(&q.tail, tail, n)
					atomic.AddInt64(&q.size, 1)
					return
				}
			} else {
				// Tail was not pointing to the last node
				// Try to swing tail to the next node
				cas(&q.tail, tail, next)
			}
		}
	}
}

// Dequeue removes and returns the first packet in the queue
// Returns nil if the queue is empty
func (q *LockFreeQueue) Dequeue() *types.QueuedPacket {
	// Track dequeue operation
	atomic.AddInt64(&q.stats.dequeues, 1)

	for {
		head := load(&q.head)
		tail := load(&q.tail)
		next := load(&head.next)
		if head == load(&q.head) { // Are head, tail, and next consistent?
			if head == tail { // Is queue empty or tail falling behind?
				if next == nil { // Is queue empty?
					return nil
				}
				// Tail is falling behind. Try to advance it
				cas(&q.tail, tail, next)
			} else {
				// Read value before CAS, otherwise another dequeue might free the next node
				value := next.value
				if cas(&q.head, head, next) {
					atomic.AddInt64(&q.size, -1)

					// Return the old head node to the pool
					globalNodePool.put(head)

					// Track node return
					atomic.AddInt64(&q.stats.nodePuts, 1)

					return value
				}
			}
		}
	}
}

// Size returns the number of packets in the queue
func (q *LockFreeQueue) Size() int {
	return int(atomic.LoadInt64(&q.size))
}

// IsEmpty returns true if the queue is empty
func (q *LockFreeQueue) IsEmpty() bool {
	return q.Size() == 0
}

// DrainToSlice returns all packets from the queue and clears it
func (q *LockFreeQueue) DrainToSlice() []*types.QueuedPacket {
	// Pre-allocate the slice based on the current size to avoid reallocations
	size := q.Size()
	if size <= 0 {
		return nil
	}

	packets := make([]*types.QueuedPacket, 0, size)

	// Use a safety counter to prevent infinite loops in case of race conditions
	safetyCounter := size * 2

	for i := 0; i < safetyCounter; i++ {
		packet := q.Dequeue()
		if packet == nil {
			break
		}
		packets = append(packets, packet)

		// If we've drained more than we expected, break to avoid potential issues
		if len(packets) >= size*2 {
			break
		}
	}

	return packets
}

// Clear removes all packets from the queue
func (q *LockFreeQueue) Clear() {
	// Create a new empty queue
	n := globalNodePool.get()
	n.next = nil
	n.value = nil

	// Store old head and tail for cleanup
	oldHead := load(&q.head)

	// Update queue pointers
	q.head = unsafe.Pointer(n)
	q.tail = unsafe.Pointer(n)
	atomic.StoreInt64(&q.size, 0)

	// Return the old head node to the pool if it's not nil
	if oldHead != nil {
		globalNodePool.put(oldHead)
		atomic.AddInt64(&q.stats.nodePuts, 1)
	}
}

// GetStats returns statistics about the queue
func (q *LockFreeQueue) GetStats() map[string]int64 {
	stats := map[string]int64{
		"size":     atomic.LoadInt64(&q.size),
		"enqueues": atomic.LoadInt64(&q.stats.enqueues),
		"dequeues": atomic.LoadInt64(&q.stats.dequeues),
		"nodeGets": atomic.LoadInt64(&q.stats.nodeGets),
		"nodePuts": atomic.LoadInt64(&q.stats.nodePuts),
	}

	// Add node pool stats
	nodePoolStats := globalNodePool.getStats()
	for k, v := range nodePoolStats {
		stats["nodePool_"+k] = v
	}

	return stats
}

// Helper functions for atomic operations on unsafe.Pointer

// load atomically loads an unsafe.Pointer
func load(p *unsafe.Pointer) (n *node) {
	return (*node)(atomic.LoadPointer(p))
}

// cas performs a compare-and-swap operation on an unsafe.Pointer
func cas(p *unsafe.Pointer, old, new *node) bool {
	return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}
