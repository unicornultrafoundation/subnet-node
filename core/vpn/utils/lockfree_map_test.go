package utils

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShardedMap(t *testing.T) {
	// Create a new map
	m := NewShardedMap(16)

	// Test empty map
	assert.Equal(t, 0, m.Size())
	_, ok := m.Get("key")
	assert.False(t, ok)

	// Test set and get
	m.Set("key1", "value1")
	assert.Equal(t, 1, m.Size())
	val, ok := m.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", val)

	// Test overwrite
	m.Set("key1", "value2")
	assert.Equal(t, 1, m.Size())
	val, ok = m.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value2", val)

	// Test delete
	m.Delete("key1")
	assert.Equal(t, 0, m.Size())
	_, ok = m.Get("key1")
	assert.False(t, ok)

	// Test multiple keys
	for i := 0; i < 100; i++ {
		m.Set(fmt.Sprintf("key%d", i), i)
	}
	assert.Equal(t, 100, m.Size())
	for i := 0; i < 100; i++ {
		val, ok := m.Get(fmt.Sprintf("key%d", i))
		assert.True(t, ok)
		assert.Equal(t, i, val)
	}

	// Test ForEach
	count := 0
	m.ForEach(func(key string, value interface{}) bool {
		count++
		return true
	})
	assert.Equal(t, 100, count)

	// Test ForEach with early termination
	count = 0
	m.ForEach(func(key string, value interface{}) bool {
		count++
		return count < 50
	})
	assert.Equal(t, 50, count)

	// Test Keys
	keys := m.Keys()
	assert.Equal(t, 100, len(keys))

	// Test Clear
	m.Clear()
	assert.Equal(t, 0, m.Size())
	_, ok = m.Get("key0")
	assert.False(t, ok)
}

func TestShardedMapConcurrency(t *testing.T) {
	// Create a new map
	m := NewShardedMap(16)

	// Test concurrent set and get
	concurrency := 100
	itemsPerGoroutine := 100
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < itemsPerGoroutine; j++ {
				key := fmt.Sprintf("key%d-%d", id, j)
				m.Set(key, id*itemsPerGoroutine+j)
			}
		}(i)
	}

	wg.Wait()

	// Verify size
	assert.Equal(t, concurrency*itemsPerGoroutine, m.Size())

	// Verify all items are present
	for i := 0; i < concurrency; i++ {
		for j := 0; j < itemsPerGoroutine; j++ {
			key := fmt.Sprintf("key%d-%d", i, j)
			val, ok := m.Get(key)
			assert.True(t, ok)
			assert.Equal(t, i*itemsPerGoroutine+j, val)
		}
	}

	// Test concurrent delete
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < itemsPerGoroutine; j++ {
				key := fmt.Sprintf("key%d-%d", id, j)
				m.Delete(key)
			}
		}(i)
	}

	wg.Wait()

	// Verify all items are deleted
	assert.Equal(t, 0, m.Size())
}

func TestShardedMapGetOrCompute(t *testing.T) {
	// Create a new map
	m := NewShardedMap(16)

	// Test GetOrCompute
	computeCount := 0
	for i := 0; i < 10; i++ {
		val := m.GetOrCompute("key", func() interface{} {
			computeCount++
			return "computed"
		})
		assert.Equal(t, "computed", val)
	}
	assert.Equal(t, 1, computeCount)
	assert.Equal(t, 1, m.Size())

	// Test concurrent GetOrCompute
	m.Clear()
	var wg sync.WaitGroup
	concurrency := 100
	wg.Add(concurrency)
	computeCount = 0
	var mu sync.Mutex

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			val := m.GetOrCompute("key", func() interface{} {
				mu.Lock()
				computeCount++
				mu.Unlock()
				return "computed"
			})
			assert.Equal(t, "computed", val)
		}()
	}

	wg.Wait()
	assert.Equal(t, 1, computeCount)
	assert.Equal(t, 1, m.Size())
}

func BenchmarkShardedMapSerial(b *testing.B) {
	m := NewShardedMap(16)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		m.Set(key, i)
	}

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		m.Get(key)
	}

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		m.Delete(key)
	}
}

func BenchmarkShardedMapParallel(b *testing.B) {
	m := NewShardedMap(16)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d", i)
			switch i % 3 {
			case 0:
				m.Set(key, i)
			case 1:
				m.Get(key)
			case 2:
				m.Delete(key)
			}
			i++
		}
	})
}

func BenchmarkStandardMapParallel(b *testing.B) {
	var mu sync.RWMutex
	m := make(map[string]interface{})
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d", i)
			switch i % 3 {
			case 0:
				mu.Lock()
				m[key] = i
				mu.Unlock()
			case 1:
				mu.RLock()
				_ = m[key]
				mu.RUnlock()
			case 2:
				mu.Lock()
				delete(m, key)
				mu.Unlock()
			}
			i++
		}
	})
}
