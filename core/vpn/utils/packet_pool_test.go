package utils

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPacketPool(t *testing.T) {
	// Create a new packet pool
	bufferSize := 1500
	prealloc := 10
	pool := NewPacketPool(bufferSize, prealloc)

	// Test getting a packet
	packet1 := pool.Get()
	assert.Equal(t, bufferSize, len(packet1.Data))
	assert.Equal(t, 0, packet1.Size)

	// Test getting multiple packets
	var packets []*PacketData
	poolSize := 20 // More than prealloc to test creation
	for i := 0; i < poolSize; i++ {
		packet := pool.Get()
		assert.Equal(t, bufferSize, len(packet.Data))
		assert.Equal(t, 0, packet.Size)
		packets = append(packets, packet)
	}

	// Test returning packets to the pool
	for _, packet := range packets {
		pool.Put(packet)
	}

	// Test getting packets after returning them
	for i := 0; i < poolSize; i++ {
		packet := pool.Get()
		assert.Equal(t, bufferSize, len(packet.Data))
		assert.Equal(t, 0, packet.Size)
	}

	// Test stats
	stats := pool.Stats()
	assert.GreaterOrEqual(t, stats["gets"], int64(1+poolSize+poolSize))
	assert.GreaterOrEqual(t, stats["puts"], int64(poolSize))
	assert.GreaterOrEqual(t, stats["creates"], int64(prealloc))
	assert.GreaterOrEqual(t, stats["size"], int64(prealloc))
}

func TestPacketPoolConcurrency(t *testing.T) {
	// Create a new packet pool
	bufferSize := 1500
	prealloc := 10
	pool := NewPacketPool(bufferSize, prealloc)

	// Test concurrent gets and puts
	concurrency := 100
	iterations := 100
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				packet := pool.Get()
				assert.Equal(t, bufferSize, len(packet.Data))
				
				// Write to the packet to ensure it's usable
				packet.Size = 10
				for k := 0; k < 10; k++ {
					packet.Data[k] = byte(k)
				}
				
				pool.Put(packet)
			}
		}()
	}

	wg.Wait()

	// Test stats after concurrent usage
	stats := pool.Stats()
	assert.Equal(t, int64(concurrency*iterations), stats["gets"])
	assert.Equal(t, int64(concurrency*iterations), stats["puts"])
}

func BenchmarkPacketPool(b *testing.B) {
	bufferSize := 1500
	pool := NewPacketPool(bufferSize, 1000)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			packet := pool.Get()
			packet.Size = 100
			pool.Put(packet)
		}
	})
}

func BenchmarkPacketAllocation(b *testing.B) {
	bufferSize := 1500

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			packet := &PacketData{
				Data:  make([]byte, bufferSize),
				Size:  100,
				inUse: 1,
			}
			_ = packet
		}
	})
}
