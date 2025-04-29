package utils

import (
	"sync"
)

// BufferPool is a pool of byte buffers
type BufferPool struct {
	size int
	pool sync.Pool
}

// NewBufferPool creates a new buffer pool with buffers of the specified size
func NewBufferPool(size int) *BufferPool {
	return &BufferPool{
		size: size,
		pool: sync.Pool{
			New: func() any {
				return make([]byte, size)
			},
		},
	}
}

// Get gets a buffer from the pool
func (p *BufferPool) Get() []byte {
	return p.pool.Get().([]byte)
}

// Put puts a buffer back into the pool
func (p *BufferPool) Put(buf []byte) {
	if cap(buf) >= p.size {
		p.pool.Put(buf[:p.size])
	}
}

// Size returns the size of the buffers in the pool
func (p *BufferPool) Size() int {
	return p.size
}
