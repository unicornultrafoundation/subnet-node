package utils

import (
	"hash/fnv"
	"sync"
	"sync/atomic"
)

// ShardedMap is a map that uses multiple shards to reduce lock contention
type ShardedMap struct {
	shards    []*mapShard
	shardMask uint32
	size      int64
}

// mapShard is a single shard of the sharded map
type mapShard struct {
	items map[string]interface{}
	mu    sync.RWMutex
}

// NewShardedMap creates a new sharded map with the specified number of shards
// The number of shards should be a power of 2 for efficient modulo operations
func NewShardedMap(shardCount int) *ShardedMap {
	// Ensure shard count is a power of 2
	if shardCount <= 0 {
		shardCount = 16
	}
	
	// Round up to the next power of 2
	shardCount--
	shardCount |= shardCount >> 1
	shardCount |= shardCount >> 2
	shardCount |= shardCount >> 4
	shardCount |= shardCount >> 8
	shardCount |= shardCount >> 16
	shardCount++
	
	// Create shards
	shards := make([]*mapShard, shardCount)
	for i := 0; i < shardCount; i++ {
		shards[i] = &mapShard{
			items: make(map[string]interface{}),
		}
	}
	
	return &ShardedMap{
		shards:    shards,
		shardMask: uint32(shardCount - 1),
		size:      0,
	}
}

// getShardIndex returns the shard index for a key
func (m *ShardedMap) getShardIndex(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32() & m.shardMask
}

// getShard returns the shard for a key
func (m *ShardedMap) getShard(key string) *mapShard {
	return m.shards[m.getShardIndex(key)]
}

// Get gets a value from the map
func (m *ShardedMap) Get(key string) (interface{}, bool) {
	shard := m.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	
	val, ok := shard.items[key]
	return val, ok
}

// Set sets a value in the map
func (m *ShardedMap) Set(key string, value interface{}) {
	shard := m.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	
	_, exists := shard.items[key]
	shard.items[key] = value
	
	if !exists {
		atomic.AddInt64(&m.size, 1)
	}
}

// Delete deletes a key from the map
func (m *ShardedMap) Delete(key string) {
	shard := m.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	
	_, exists := shard.items[key]
	delete(shard.items, key)
	
	if exists {
		atomic.AddInt64(&m.size, -1)
	}
}

// Size returns the number of items in the map
func (m *ShardedMap) Size() int {
	return int(atomic.LoadInt64(&m.size))
}

// ForEach iterates over all items in the map
// The callback function should not modify the map
func (m *ShardedMap) ForEach(callback func(key string, value interface{}) bool) {
	for _, shard := range m.shards {
		shard.mu.RLock()
		for k, v := range shard.items {
			shard.mu.RUnlock()
			if !callback(k, v) {
				return
			}
			shard.mu.RLock()
		}
		shard.mu.RUnlock()
	}
}

// Keys returns all keys in the map
func (m *ShardedMap) Keys() []string {
	keys := make([]string, 0, m.Size())
	
	for _, shard := range m.shards {
		shard.mu.RLock()
		for k := range shard.items {
			keys = append(keys, k)
		}
		shard.mu.RUnlock()
	}
	
	return keys
}

// Clear removes all items from the map
func (m *ShardedMap) Clear() {
	for _, shard := range m.shards {
		shard.mu.Lock()
		shard.items = make(map[string]interface{})
		shard.mu.Unlock()
	}
	
	atomic.StoreInt64(&m.size, 0)
}

// GetOrCompute gets a value from the map, or computes and stores it if not present
func (m *ShardedMap) GetOrCompute(key string, compute func() interface{}) interface{} {
	// First try to get the value without locking for writing
	if val, ok := m.Get(key); ok {
		return val
	}
	
	// Value not found, compute it
	shard := m.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	
	// Check again in case another goroutine set the value while we were waiting
	if val, ok := shard.items[key]; ok {
		return val
	}
	
	// Compute the value
	val := compute()
	shard.items[key] = val
	atomic.AddInt64(&m.size, 1)
	
	return val
}
