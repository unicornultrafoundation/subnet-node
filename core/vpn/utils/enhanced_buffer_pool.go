package utils

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// EnhancedBufferPool is an improved buffer pool with better memory management
type EnhancedBufferPool struct {
	// Size of each buffer
	size int
	// Number of buffers to preallocate
	prealloc int
	// Pool of buffers
	pool sync.Pool
	// Stats
	stats struct {
		gets    int64
		puts    int64
		misses  int64
		creates int64
	}
}

// NewEnhancedBufferPool creates a new enhanced buffer pool
func NewEnhancedBufferPool(size int, prealloc int) *EnhancedBufferPool {
	if prealloc <= 0 {
		// Default to number of CPUs * 4 for preallocation
		prealloc = runtime.NumCPU() * 4
	}

	p := &EnhancedBufferPool{
		size:     size,
		prealloc: prealloc,
		pool: sync.Pool{
			New: func() any {
				buf := make([]byte, size)
				return buf
			},
		},
	}

	// Preallocate buffers
	p.preallocate()

	return p
}

// preallocate creates and puts a number of buffers into the pool
func (p *EnhancedBufferPool) preallocate() {
	for i := 0; i < p.prealloc; i++ {
		buf := make([]byte, p.size)
		p.pool.Put(buf)
	}
	atomic.AddInt64(&p.stats.creates, int64(p.prealloc))
}

// Get gets a buffer from the pool
func (p *EnhancedBufferPool) Get() []byte {
	atomic.AddInt64(&p.stats.gets, 1)
	
	// Get a buffer from the pool
	buf := p.pool.Get().([]byte)
	
	// Check if we need to resize the buffer
	if len(buf) != p.size {
		// Buffer size has changed, create a new one
		atomic.AddInt64(&p.stats.misses, 1)
		atomic.AddInt64(&p.stats.creates, 1)
		return make([]byte, p.size)
	}
	
	// Clear the buffer to prevent data leakage
	for i := range buf {
		buf[i] = 0
	}
	
	return buf
}

// Put puts a buffer back into the pool
func (p *EnhancedBufferPool) Put(buf []byte) {
	atomic.AddInt64(&p.stats.puts, 1)
	
	// Only put buffers of the correct size back into the pool
	if cap(buf) >= p.size {
		// Reslice to the correct size
		p.pool.Put(buf[:p.size])
	}
}

// Size returns the size of the buffers in the pool
func (p *EnhancedBufferPool) Size() int {
	return p.size
}

// Stats returns statistics about the buffer pool
func (p *EnhancedBufferPool) Stats() map[string]int64 {
	return map[string]int64{
		"gets":    atomic.LoadInt64(&p.stats.gets),
		"puts":    atomic.LoadInt64(&p.stats.puts),
		"misses":  atomic.LoadInt64(&p.stats.misses),
		"creates": atomic.LoadInt64(&p.stats.creates),
	}
}
