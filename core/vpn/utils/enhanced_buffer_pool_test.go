package utils

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnhancedBufferPool(t *testing.T) {
	// Create a new buffer pool
	bufferSize := 1500
	prealloc := 10
	pool := NewEnhancedBufferPool(bufferSize, prealloc)

	// Test getting a buffer
	buf1 := pool.Get()
	assert.Equal(t, bufferSize, len(buf1))

	// Test getting multiple buffers
	var buffers [][]byte
	poolSize := 20 // More than prealloc to test creation
	for i := 0; i < poolSize; i++ {
		buf := pool.Get()
		assert.Equal(t, bufferSize, len(buf))
		buffers = append(buffers, buf)
	}

	// Test returning buffers to the pool
	for _, buf := range buffers {
		pool.Put(buf)
	}

	// Test getting buffers after returning them
	for i := 0; i < poolSize; i++ {
		buf := pool.Get()
		assert.Equal(t, bufferSize, len(buf))
	}

	// Test stats
	stats := pool.Stats()
	assert.GreaterOrEqual(t, stats["gets"], int64(1+poolSize+poolSize))
	assert.GreaterOrEqual(t, stats["puts"], int64(poolSize))
	assert.GreaterOrEqual(t, stats["creates"], int64(prealloc))
}

func TestEnhancedBufferPoolConcurrency(t *testing.T) {
	// Create a new buffer pool
	bufferSize := 1500
	prealloc := 10
	pool := NewEnhancedBufferPool(bufferSize, prealloc)

	// Test concurrent gets and puts
	concurrency := 100
	iterations := 1000
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				buf := pool.Get()
				assert.Equal(t, bufferSize, len(buf))
				
				// Write to the buffer to ensure it's usable
				for k := 0; k < 10; k++ {
					buf[k] = byte(k)
				}
				
				pool.Put(buf)
			}
		}()
	}

	wg.Wait()

	// Test stats after concurrent usage
	stats := pool.Stats()
	assert.Equal(t, int64(concurrency*iterations), stats["gets"])
	assert.Equal(t, int64(concurrency*iterations), stats["puts"])
}

func TestEnhancedBufferPoolWithWrongSizeBuffer(t *testing.T) {
	// Create a new buffer pool
	bufferSize := 1500
	pool := NewEnhancedBufferPool(bufferSize, 10)

	// Put a buffer with wrong size
	wrongSizeBuffer := make([]byte, 2000)
	pool.Put(wrongSizeBuffer)

	// Get a buffer and check its size
	buf := pool.Get()
	assert.Equal(t, bufferSize, len(buf))

	// Put a buffer with smaller size
	smallerBuffer := make([]byte, 1000)
	pool.Put(smallerBuffer)

	// Get a buffer and check its size
	buf = pool.Get()
	assert.Equal(t, bufferSize, len(buf))
}

func BenchmarkEnhancedBufferPool(b *testing.B) {
	bufferSize := 1500
	pool := NewEnhancedBufferPool(bufferSize, 1000)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := pool.Get()
			pool.Put(buf)
		}
	})
}

func BenchmarkStandardBufferPool(b *testing.B) {
	bufferSize := 1500
	pool := NewBufferPool(bufferSize)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := pool.Get()
			pool.Put(buf)
		}
	})
}

func BenchmarkNoPool(b *testing.B) {
	bufferSize := 1500

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := make([]byte, bufferSize)
			_ = buf
		}
	})
}
