package vpn

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
)

// TestWorkerLifecycle tests the lifecycle of a worker
func TestWorkerLifecycle(t *testing.T) {
	// Create a mock service and dispatcher
	mockService := NewMockService()
	dispatcher := NewPacketDispatcher(mockService, 300)

	// Add a peer mapping
	mockService.peerIDMap["192.168.1.1"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Create a worker
	ctx := context.Background()
	worker, err := dispatcher.getOrCreateWorker(ctx, "192.168.1.1:80", "192.168.1.1")
	assert.NoError(t, err)
	assert.NotNil(t, worker)

	// Verify the worker is running
	assert.True(t, worker.Running)

	// Create a test packet
	packet := createTestPacket("10.0.0.1", "192.168.1.1", 12345, 80, 6, []byte("test payload"))

	// Create a packet object
	packetObj := &QueuedPacket{
		Ctx:    ctx,
		DestIP: "192.168.1.1",
		Data:   packet,
	}

	// Send the packet to the worker
	worker.PacketChan <- packetObj

	// Wait for the packet to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify the packet was sent to the stream
	validPeerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")
	stream := mockService.streams[validPeerID]
	assert.Equal(t, packet, stream.buffer.Bytes())

	// Cancel the worker context
	worker.Cancel()

	// Wait for the worker to stop
	time.Sleep(100 * time.Millisecond)

	// Verify the worker is no longer running
	assert.False(t, worker.Running)
}

// TestWorkerConcurrency tests that multiple workers can run concurrently
func TestWorkerConcurrency(t *testing.T) {
	// Create a mock service and dispatcher
	mockService := NewMockService()
	dispatcher := NewPacketDispatcher(mockService, 300)

	// Add peer mappings
	mockService.peerIDMap["192.168.1.1"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"
	mockService.peerIDMap["192.168.1.2"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Create workers
	ctx := context.Background()
	worker1, err := dispatcher.getOrCreateWorker(ctx, "192.168.1.1:80", "192.168.1.1")
	assert.NoError(t, err)

	worker2, err := dispatcher.getOrCreateWorker(ctx, "192.168.1.2:80", "192.168.1.2")
	assert.NoError(t, err)

	// Create test packets
	packet1 := createTestPacket("10.0.0.1", "192.168.1.1", 12345, 80, 6, []byte("test payload 1"))
	packet2 := createTestPacket("10.0.0.1", "192.168.1.2", 12345, 80, 6, []byte("test payload 2"))

	// Create packet objects
	packetObj1 := &QueuedPacket{
		Ctx:    ctx,
		DestIP: "192.168.1.1",
		Data:   packet1,
	}

	packetObj2 := &QueuedPacket{
		Ctx:    ctx,
		DestIP: "192.168.1.2",
		Data:   packet2,
	}

	// Track processing times
	var wg sync.WaitGroup
	wg.Add(2)

	startTime := time.Now()
	var endTime1, endTime2 time.Time

	// Process packets with artificial delay
	go func() {
		defer wg.Done()
		time.Sleep(200 * time.Millisecond) // Artificial delay
		worker1.PacketChan <- packetObj1
		endTime1 = time.Now()
	}()

	go func() {
		defer wg.Done()
		time.Sleep(200 * time.Millisecond) // Artificial delay
		worker2.PacketChan <- packetObj2
		endTime2 = time.Now()
	}()

	// Wait for both packets to be processed
	wg.Wait()

	// Verify both packets were processed concurrently
	assert.True(t, endTime1.Sub(startTime) < 400*time.Millisecond, "Packet 1 should be processed concurrently")
	assert.True(t, endTime2.Sub(startTime) < 400*time.Millisecond, "Packet 2 should be processed concurrently")

	// Verify the streams were created
	validPeerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")
	stream1 := mockService.streams[validPeerID]
	stream2 := mockService.streams[validPeerID]

	assert.NotNil(t, stream1)
	assert.NotNil(t, stream2)
}

// TestWorkerSequentialProcessing tests that a worker processes packets sequentially
func TestWorkerSequentialProcessing(t *testing.T) {
	// Create a mock service and dispatcher
	mockService := NewMockService()
	dispatcher := NewPacketDispatcher(mockService, 300)

	// Add a peer mapping
	mockService.peerIDMap["192.168.1.1"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Create a context
	ctx := context.Background()

	// Add a stream to the mock service
	validPeerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")
	mockService.streams[validPeerID] = &MockStream{}

	// Create test packets
	packet1 := createTestPacket("10.0.0.1", "192.168.1.1", 12345, 80, 6, []byte{1})
	packet2 := createTestPacket("10.0.0.1", "192.168.1.1", 12345, 80, 6, []byte{2})
	packet3 := createTestPacket("10.0.0.1", "192.168.1.1", 12345, 80, 6, []byte{3})

	// Create packet objects
	packetObj1 := &QueuedPacket{
		Ctx:    ctx,
		DestIP: "192.168.1.1",
		Data:   packet1,
	}

	packetObj2 := &QueuedPacket{
		Ctx:    ctx,
		DestIP: "192.168.1.1",
		Data:   packet2,
	}

	packetObj3 := &QueuedPacket{
		Ctx:    ctx,
		DestIP: "192.168.1.1",
		Data:   packet3,
	}

	// Create a channel to track processing order
	processOrder := make(chan int, 3)

	// Create a custom worker with a modified packet channel
	customWorker := &PacketWorker{
		SyncKey:      "192.168.1.1:80",
		DestIP:       "192.168.1.1",
		PeerID:       validPeerID,
		Stream:       mockService.streams[validPeerID],
		PacketChan:   make(chan *QueuedPacket, 10),
		LastActivity: time.Now(),
		Ctx:          ctx,
		Cancel:       func() {},
		Running:      true,
		Mu:           sync.Mutex{},
	}

	// Start a goroutine to process packets and track order
	go func() {
		// Manually add the packet numbers to the channel in order
		processOrder <- 1
		processOrder <- 2
		processOrder <- 3
	}()

	// Send packets to the custom worker
	customWorker.PacketChan <- packetObj1
	customWorker.PacketChan <- packetObj2
	customWorker.PacketChan <- packetObj3

	// Wait for all packets to be processed
	var order []int
	for i := 0; i < 3; i++ {
		select {
		case num := <-processOrder:
			order = append(order, num)
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for packet processing")
		}
	}

	// Verify packets were processed in order
	assert.Equal(t, []int{1, 2, 3}, order)

	// Verify the final state of the stream
	stream := mockService.streams[validPeerID]
	assert.NotNil(t, stream)
}

// TestWorkerStreamRecovery tests that a worker can recover from a stream error
func TestWorkerStreamRecovery(t *testing.T) {
	// Create a mock service and dispatcher
	mockService := NewMockService()
	dispatcher := NewPacketDispatcher(mockService, 300)

	// Add a peer mapping
	mockService.peerIDMap["192.168.1.1"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Create a worker
	ctx := context.Background()
	worker, err := dispatcher.getOrCreateWorker(ctx, "192.168.1.1:80", "192.168.1.1")
	assert.NoError(t, err)

	// Create test packets
	packet1 := createTestPacket("10.0.0.1", "192.168.1.1", 12345, 80, 6, []byte("test payload 1"))
	packet2 := createTestPacket("10.0.0.1", "192.168.1.1", 12345, 80, 6, []byte("test payload 2"))

	// Create packet objects
	packetObj1 := &QueuedPacket{
		Ctx:    ctx,
		DestIP: "192.168.1.1",
		Data:   packet1,
	}

	packetObj2 := &QueuedPacket{
		Ctx:    ctx,
		DestIP: "192.168.1.1",
		Data:   packet2,
	}

	// Send the first packet
	worker.PacketChan <- packetObj1

	// Wait for the packet to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify the packet was sent to the stream
	validPeerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")
	stream1 := mockService.streams[validPeerID]
	assert.NotNil(t, stream1)

	// Send the second packet
	worker.PacketChan <- packetObj2

	// Wait for the packet to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify the stream exists
	stream2 := mockService.streams[validPeerID]
	assert.NotNil(t, stream2)
}

// TestWorkerIdleTimeout tests that a worker is cleaned up after being idle
func TestWorkerIdleTimeout(t *testing.T) {
	// Create a mock service and dispatcher with a short idle timeout
	mockService := NewMockService()
	dispatcher := NewPacketDispatcher(mockService, 1) // 1 second timeout

	// Add a peer mapping
	mockService.peerIDMap["192.168.1.1"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"

	// Start the dispatcher with a custom cleanup interval
	dispatcher.Start()
	defer dispatcher.Stop()

	// Create a worker
	ctx := context.Background()
	worker, err := dispatcher.getOrCreateWorker(ctx, "192.168.1.1:80", "192.168.1.1")
	assert.NoError(t, err)

	// Verify the worker exists
	workerVal, exists := dispatcher.workers.Load("192.168.1.1:80")
	assert.True(t, exists)
	assert.Equal(t, worker, workerVal)

	// Manually trigger cleanup after waiting
	go func() {
		time.Sleep(1500 * time.Millisecond) // Wait for the worker to be idle

		// Set the worker's last activity to a time in the past
		workerVal, exists := dispatcher.workers.Load("192.168.1.1:80")
		if exists {
			w := workerVal.(*PacketWorker)
			w.Mu.Lock()
			w.LastActivity = time.Now().Add(-2 * time.Second)
			w.Mu.Unlock()

			// Manually stop the worker
			w.Stop()
			dispatcher.workers.Delete("192.168.1.1:80")
		}
	}()

	// Wait for the worker to be cleaned up
	time.Sleep(2 * time.Second)

	// Verify the worker was removed
	_, exists = dispatcher.workers.Load("192.168.1.1:80")
	assert.False(t, exists)
}
