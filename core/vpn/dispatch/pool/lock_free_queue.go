package pool

import (
	"sync/atomic"
	"unsafe"

	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
)

// node represents a node in the lock-free queue
type node struct {
	value *types.QueuedPacket
	next  unsafe.Pointer
}

// LockFreeQueue is a lock-free queue implementation for overflow packets
type LockFreeQueue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
	size int64
}

// NewLockFreeQueue creates a new lock-free queue
func NewLockFreeQueue() *LockFreeQueue {
	n := unsafe.Pointer(&node{})
	return &LockFreeQueue{
		head: n,
		tail: n,
		size: 0,
	}
}

// Enqueue adds a packet to the queue
func (q *LockFreeQueue) Enqueue(packet *types.QueuedPacket) {
	n := &node{value: packet}
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
	var packets []*types.QueuedPacket
	for {
		packet := q.Dequeue()
		if packet == nil {
			break
		}
		packets = append(packets, packet)
	}
	return packets
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
