package pool

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
)

func TestLockFreeQueue(t *testing.T) {
	// Create a new queue
	q := NewLockFreeQueue()

	// Test empty queue
	assert.True(t, q.IsEmpty())
	assert.Equal(t, 0, q.Size())
	assert.Nil(t, q.Dequeue())

	// Create a test packet
	packet := &types.QueuedPacket{
		Data:   []byte{1, 2, 3},
		DestIP: "192.168.1.1",
	}

	// Test enqueue and dequeue
	q.Enqueue(packet)
	assert.False(t, q.IsEmpty())
	assert.Equal(t, 1, q.Size())
	
	p := q.Dequeue()
	assert.Equal(t, packet, p)
	assert.True(t, q.IsEmpty())
	assert.Equal(t, 0, q.Size())

	// Test multiple enqueues and dequeues
	q.Enqueue(packet)
	q.Enqueue(packet)
	q.Enqueue(packet)
	assert.Equal(t, 3, q.Size())
	
	p = q.Dequeue()
	assert.Equal(t, packet, p)
	assert.Equal(t, 2, q.Size())
	
	p = q.Dequeue()
	assert.Equal(t, packet, p)
	assert.Equal(t, 1, q.Size())
	
	p = q.Dequeue()
	assert.Equal(t, packet, p)
	assert.Equal(t, 0, q.Size())
	
	p = q.Dequeue()
	assert.Nil(t, p)
}

func TestLockFreeQueueConcurrency(t *testing.T) {
	// Create a new queue
	q := NewLockFreeQueue()

	// Test concurrent enqueues
	concurrency := 100
	itemsPerGoroutine := 1000
	var wg sync.WaitGroup
	wg.Add(concurrency)

	// Create a test packet
	packet := &types.QueuedPacket{
		Data:   []byte{1, 2, 3},
		DestIP: "192.168.1.1",
	}

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < itemsPerGoroutine; j++ {
				q.Enqueue(packet)
			}
		}()
	}

	wg.Wait()

	// Verify size
	assert.Equal(t, concurrency*itemsPerGoroutine, q.Size())

	// Test concurrent dequeues
	wg.Add(concurrency)
	counts := make([]int, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < itemsPerGoroutine; j++ {
				p := q.Dequeue()
				if p != nil {
					counts[id]++
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all items were dequeued
	totalDequeued := 0
	for _, count := range counts {
		totalDequeued += count
	}
	assert.Equal(t, concurrency*itemsPerGoroutine, totalDequeued)
	assert.Equal(t, 0, q.Size())
}

func TestDrainToSlice(t *testing.T) {
	// Create a new queue
	q := NewLockFreeQueue()

	// Create a test packet
	packet := &types.QueuedPacket{
		Data:   []byte{1, 2, 3},
		DestIP: "192.168.1.1",
	}

	// Enqueue some packets
	for i := 0; i < 10; i++ {
		q.Enqueue(packet)
	}

	// Drain to slice
	packets := q.DrainToSlice()
	
	// Verify
	assert.Equal(t, 10, len(packets))
	assert.Equal(t, 0, q.Size())
	
	// Verify all packets are the same
	for _, p := range packets {
		assert.Equal(t, packet, p)
	}
}

func BenchmarkLockFreeQueueEnqueueDequeue(b *testing.B) {
	q := NewLockFreeQueue()
	packet := &types.QueuedPacket{
		Data:   []byte{1, 2, 3},
		DestIP: "192.168.1.1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(packet)
		q.Dequeue()
	}
}

func BenchmarkLockFreeQueueParallel(b *testing.B) {
	q := NewLockFreeQueue()
	packet := &types.QueuedPacket{
		Data:   []byte{1, 2, 3},
		DestIP: "192.168.1.1",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			q.Enqueue(packet)
			q.Dequeue()
		}
	})
}
