package router

import (
	"hash/fnv"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

// ShardedCache implements a sharded TTL cache for connection routes
type ShardedCache struct {
	shards    []*CacheShard
	shardMask uint32
}

// CacheShard represents a single shard of the cache
type CacheShard struct {
	cache *cache.Cache
	mu    sync.RWMutex
}

// NewShardedCache creates a new sharded cache
func NewShardedCache(shardCount int, defaultExpiration, cleanupInterval time.Duration) *ShardedCache {
	// Ensure shard count is a power of 2
	shardCount = nextPowerOfTwo(shardCount)
	shards := make([]*CacheShard, shardCount)

	for i := 0; i < shardCount; i++ {
		shards[i] = &CacheShard{
			cache: cache.New(defaultExpiration, cleanupInterval),
		}
	}

	return &ShardedCache{
		shards:    shards,
		shardMask: uint32(shardCount - 1),
	}
}

// getShard returns the appropriate shard for a key
func (c *ShardedCache) getShard(key string) *CacheShard {
	h := fnv.New32a()
	h.Write([]byte(key))
	shardIndex := h.Sum32() & c.shardMask
	return c.shards[shardIndex]
}

// Get retrieves an item from the cache
func (c *ShardedCache) Get(key string) (interface{}, bool) {
	shard := c.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	return shard.cache.Get(key)
}

// Set adds an item to the cache
func (c *ShardedCache) Set(key string, value interface{}, d time.Duration) {
	shard := c.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	shard.cache.Set(key, value, d)
}

// Delete removes an item from the cache
func (c *ShardedCache) Delete(key string) {
	shard := c.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	shard.cache.Delete(key)
}

// ItemCount returns the number of items in the cache
func (c *ShardedCache) ItemCount() int {
	count := 0
	for _, shard := range c.shards {
		shard.mu.RLock()
		count += shard.cache.ItemCount()
		shard.mu.RUnlock()
	}
	return count
}

// OnEvicted sets a callback function to be called when an item is evicted
func (c *ShardedCache) OnEvicted(f func(string, interface{})) {
	for _, shard := range c.shards {
		shard.mu.Lock()
		shard.cache.OnEvicted(f)
		shard.mu.Unlock()
	}
}

// nextPowerOfTwo returns the next power of 2 >= n
func nextPowerOfTwo(n int) int {
	if n <= 1 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	return n + 1
}
