package utils

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLockFreeQueue(t *testing.T) {
	// Create a new queue
	q := NewLockFreeQueue()

	// Test empty queue
	assert.True(t, q.IsEmpty())
	assert.Equal(t, 0, q.Size())
	assert.Nil(t, q.Dequeue())

	// Test enqueue and dequeue
	q.Enqueue(1)
	assert.False(t, q.IsEmpty())
	assert.Equal(t, 1, q.Size())
	assert.Equal(t, 1, q.Dequeue())
	assert.True(t, q.IsEmpty())
	assert.Equal(t, 0, q.Size())

	// Test multiple enqueues and dequeues
	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)
	assert.Equal(t, 3, q.Size())
	assert.Equal(t, 1, q.Dequeue())
	assert.Equal(t, 2, q.Size())
	assert.Equal(t, 2, q.Dequeue())
	assert.Equal(t, 1, q.Size())
	assert.Equal(t, 3, q.Dequeue())
	assert.Equal(t, 0, q.Size())
	assert.Nil(t, q.Dequeue())
}

func TestLockFreeQueueConcurrency(t *testing.T) {
	// Create a new queue
	q := NewLockFreeQueue()

	// Test concurrent enqueues
	concurrency := 100
	itemsPerGoroutine := 1000
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < itemsPerGoroutine; j++ {
				q.Enqueue(id*itemsPerGoroutine + j)
			}
		}(i)
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
				item := q.Dequeue()
				if item != nil {
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

func TestLockFreeQueueMixedOperations(t *testing.T) {
	// Create a new queue
	q := NewLockFreeQueue()

	// Test concurrent enqueues and dequeues
	concurrency := 100
	iterations := 1000
	var wg sync.WaitGroup
	wg.Add(concurrency * 2) // Half for enqueues, half for dequeues

	// Start enqueuers
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				q.Enqueue(id*iterations + j)
			}
		}(i)
	}

	// Give enqueuers a head start
	for i := 0; i < 1000; i++ {
		q.Enqueue(i)
	}

	// Start dequeuers
	dequeued := make(chan interface{}, concurrency*iterations)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				for {
					item := q.Dequeue()
					if item != nil {
						dequeued <- item
						break
					}
					// If dequeue returns nil, try again
				}
			}
		}()
	}

	wg.Wait()
	close(dequeued)

	// Count dequeued items
	count := 0
	for range dequeued {
		count++
	}

	// Verify expected number of items were dequeued
	assert.Equal(t, concurrency*iterations+1000, count)
}

func BenchmarkLockFreeQueueSerial(b *testing.B) {
	q := NewLockFreeQueue()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		q.Enqueue(i)
	}

	for i := 0; i < b.N; i++ {
		q.Dequeue()
	}
}

func BenchmarkLockFreeQueueParallel(b *testing.B) {
	q := NewLockFreeQueue()
	b.ResetTimer()

	// Half the goroutines enqueue, half dequeue
	b.RunParallel(func(pb *testing.PB) {
		// Each goroutine alternates between enqueue and dequeue
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				q.Enqueue(i)
			} else {
				q.Dequeue()
			}
			i++
		}
	})
}
