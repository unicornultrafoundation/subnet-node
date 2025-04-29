package utils

// BufferPoolInterface defines the interface for buffer pools
type BufferPoolInterface interface {
	// Get gets a buffer from the pool
	Get() []byte

	// Put puts a buffer back into the pool
	Put(buf []byte)

	// Size returns the size of the buffers in the pool
	Size() int
}

// Ensure both buffer pool implementations satisfy the interface
var _ BufferPoolInterface = (*BufferPool)(nil)
var _ BufferPoolInterface = (*EnhancedBufferPool)(nil)

// Global enhanced buffer pool instance
var globalEnhancedBufferPool *EnhancedBufferPool

// CreateBufferPool creates a buffer pool with the specified size and options
// If useEnhanced is true, it creates an EnhancedBufferPool, otherwise a standard BufferPool
func CreateBufferPool(size int, useEnhanced bool, prealloc int) BufferPoolInterface {
	if useEnhanced {
		enhancedPool := NewEnhancedBufferPool(size, prealloc)
		globalEnhancedBufferPool = enhancedPool
		return enhancedPool
	}
	return NewBufferPool(size)
}

// GetEnhancedBufferPool returns the global enhanced buffer pool instance if it exists
func GetEnhancedBufferPool() (*EnhancedBufferPool, bool) {
	if globalEnhancedBufferPool != nil {
		return globalEnhancedBufferPool, true
	}
	return nil, false
}
