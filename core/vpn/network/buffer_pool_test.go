package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferPool(t *testing.T) {
	// Create a new buffer pool
	bufferSize := 1500
	poolSize := 10
	pool := NewBufferPool(bufferSize, poolSize)

	// Test getting a buffer
	buf1 := pool.Get()
	assert.Equal(t, bufferSize, len(buf1))

	// Test getting multiple buffers
	var buffers [][]byte
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

	// Test getting more buffers than the pool size
	// This should create new buffers without error
	for i := 0; i < poolSize*2; i++ {
		buf := pool.Get()
		assert.Equal(t, bufferSize, len(buf))
	}

	// Test putting more buffers than the pool size
	// This should not cause any errors
	for i := 0; i < poolSize*2; i++ {
		pool.Put(make([]byte, bufferSize))
	}

	// Test that the pool is still working
	buf2 := pool.Get()
	assert.Equal(t, bufferSize, len(buf2))
}

func TestBufferPoolConcurrency(t *testing.T) {
	// Create a new buffer pool
	bufferSize := 1500
	poolSize := 10
	pool := NewBufferPool(bufferSize, poolSize)

	// Test concurrent gets and puts
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			buf := pool.Get()
			assert.Equal(t, bufferSize, len(buf))
			pool.Put(buf)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 100; i++ {
		<-done
	}
}
