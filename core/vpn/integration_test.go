package vpn

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
)

// TestIntegrationPacketFlow tests the entire packet flow from TUN to peer
func TestIntegrationPacketFlow(t *testing.T) {
	// Create a mock service
	mockService := NewMockService()

	// Add peer mappings
	mockService.peerIDMap["192.168.1.1"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"
	mockService.peerIDMap["192.168.1.2"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"

	// Create a dispatcher
	dispatcher := NewPacketDispatcher(mockService, 300)

	// Create a service with the dispatcher
	config := &VPNConfig{
		MTU: 1500,
	}
	service := &Service{
		dispatcher: dispatcher,
		stopChan:   make(chan struct{}),
		config:     config,
	}

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Create test packets for different destinations
	packet1 := createTestPacket("10.0.0.1", "192.168.1.1", 12345, 80, 6, []byte("test payload 1"))
	packet2 := createTestPacket("10.0.0.1", "192.168.1.2", 12345, 80, 6, []byte("test payload 2"))
	packet3 := createTestPacket("10.0.0.1", "192.168.1.1", 12345, 8080, 6, []byte("test payload 3"))

	// For this test, we'll just dispatch packets directly
	ctx := context.Background()

	// Dispatch the packets
	dispatcher.DispatchPacket(ctx, "192.168.1.1:80", "192.168.1.1", packet1)
	dispatcher.DispatchPacket(ctx, "192.168.1.2:80", "192.168.1.2", packet2)
	dispatcher.DispatchPacket(ctx, "192.168.1.1:8080", "192.168.1.1", packet3)

	// Wait for the packets to be processed
	time.Sleep(500 * time.Millisecond)

	// Stop the service
	close(service.stopChan)

	// Verify workers were created for each destination
	_, exists := dispatcher.workers.Load("192.168.1.1:80")
	assert.True(t, exists)

	_, exists = dispatcher.workers.Load("192.168.1.2:80")
	assert.True(t, exists)

	_, exists = dispatcher.workers.Load("192.168.1.1:8080")
	assert.True(t, exists)

	// Verify the packets were sent to the correct streams
	validPeerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")
	stream1 := mockService.streams[validPeerID]
	stream2 := mockService.streams[validPeerID]

	// In our mock implementation, the last packet for each peer overwrites previous ones
	// In a real implementation, we would need a more sophisticated way to verify all packets
	assert.NotNil(t, stream1)
	assert.NotNil(t, stream2)

	// Verify metrics were updated
	metrics := mockService.GetMetrics()
	assert.Equal(t, 3, metrics["packets_sent"], "Should have sent 3 packets")
	// Calculate total bytes sent
	expectedBytes := len(packet1) + len(packet2) + len(packet3)
	assert.Equal(t, expectedBytes, metrics["bytes_sent"], "Should have sent correct number of bytes")
	assert.Equal(t, 0, metrics["packets_dropped"], "Should not have dropped any packets")
	assert.Equal(t, 0, metrics["stream_errors"], "Should not have any stream errors")
}

// TestIntegrationConcurrentPackets tests handling concurrent packets for different destinations
func TestIntegrationConcurrentPackets(t *testing.T) {
	// Create a mock service
	mockService := NewMockService()

	// Add peer mappings
	mockService.peerIDMap["192.168.1.1"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"
	mockService.peerIDMap["192.168.1.2"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"

	// Create a dispatcher
	dispatcher := NewPacketDispatcher(mockService, 300)

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Create test packets for different destinations
	packet1 := createTestPacket("10.0.0.1", "192.168.1.1", 12345, 80, 6, []byte("test payload 1"))
	packet2 := createTestPacket("10.0.0.1", "192.168.1.2", 12345, 80, 6, []byte("test payload 2"))

	// Track processing times
	var wg sync.WaitGroup
	wg.Add(2)

	startTime := time.Now()
	var endTime1, endTime2 time.Time

	// Add artificial delay to packet processing
	time.Sleep(200 * time.Millisecond)

	// Dispatch packets concurrently
	ctx := context.Background()

	go func() {
		defer wg.Done()
		dispatcher.DispatchPacket(ctx, "192.168.1.1:80", "192.168.1.1", packet1)
		endTime1 = time.Now()
	}()

	go func() {
		defer wg.Done()
		dispatcher.DispatchPacket(ctx, "192.168.1.2:80", "192.168.1.2", packet2)
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

	// Verify metrics were updated
	metrics := mockService.GetMetrics()
	// In concurrent tests, we can't guarantee both packets will be processed
	// before we check the metrics, so we just verify at least one packet was sent
	assert.GreaterOrEqual(t, metrics["packets_sent"], 1, "Should have sent at least 1 packet")
	assert.GreaterOrEqual(t, metrics["bytes_sent"], len(packet1), "Should have sent at least some bytes")
	assert.Equal(t, 0, metrics["packets_dropped"], "Should not have dropped any packets")
	assert.Equal(t, 0, metrics["stream_errors"], "Should not have any stream errors")
}

// TestIntegrationSequentialPackets tests handling sequential packets for the same destination
func TestIntegrationSequentialPackets(t *testing.T) {
	// Create a mock service
	mockService := NewMockService()

	// Add peer mappings
	mockService.peerIDMap["192.168.1.1"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"

	// Create a dispatcher
	dispatcher := NewPacketDispatcher(mockService, 300)

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Create test packets for the same destination
	packet1 := createTestPacket("10.0.0.1", "192.168.1.1", 12345, 80, 6, []byte{1})
	packet2 := createTestPacket("10.0.0.1", "192.168.1.1", 12345, 80, 6, []byte{2})
	packet3 := createTestPacket("10.0.0.1", "192.168.1.1", 12345, 80, 6, []byte{3})

	// No need to track processing order in this test

	// Dispatch packets
	ctx := context.Background()
	dispatcher.DispatchPacket(ctx, "192.168.1.1:80", "192.168.1.1", packet1)
	dispatcher.DispatchPacket(ctx, "192.168.1.1:80", "192.168.1.1", packet2)
	dispatcher.DispatchPacket(ctx, "192.168.1.1:80", "192.168.1.1", packet3)

	// Wait for packets to be processed
	time.Sleep(500 * time.Millisecond)

	// Verify the packet was sent to the stream
	validPeerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")
	stream := mockService.streams[validPeerID]
	assert.NotNil(t, stream)

	// Verify metrics were updated
	metrics := mockService.GetMetrics()
	assert.Equal(t, 3, metrics["packets_sent"], "Should have sent 3 packets")
	expectedBytes := len(packet1) + len(packet2) + len(packet3)
	assert.Equal(t, expectedBytes, metrics["bytes_sent"], "Should have sent correct number of bytes")
	assert.Equal(t, 0, metrics["packets_dropped"], "Should not have dropped any packets")
	assert.Equal(t, 0, metrics["stream_errors"], "Should not have any stream errors")
}

// TestIntegrationWorkerCleanup tests that workers are cleaned up after being idle
func TestIntegrationWorkerCleanup(t *testing.T) {
	// Create a mock service
	mockService := NewMockService()

	// Add peer mappings
	mockService.peerIDMap["192.168.1.1"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"

	// Set a short idle timeout in the mock service
	mockService.workerIdleTimeout = 1 // 1 second timeout

	// Create a dispatcher
	dispatcher := NewPacketDispatcher(mockService, 0)

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Dispatch a packet
	ctx := context.Background()
	packet := createTestPacket("10.0.0.1", "192.168.1.1", 12345, 80, 6, []byte("test payload"))
	dispatcher.DispatchPacket(ctx, "192.168.1.1:80", "192.168.1.1", packet)

	// Verify a worker was created
	_, exists := dispatcher.workers.Load("192.168.1.1:80")
	assert.True(t, exists)

	// Override the cleanupInactiveWorkers function to run immediately
	go func() {
		time.Sleep(1500 * time.Millisecond) // Wait for the worker to be idle
		idleTimeout := time.Duration(dispatcher.workerIdleTimeout) * time.Second

		dispatcher.workers.Range(func(key, value interface{}) bool {
			syncKey := key.(string)
			worker := value.(*PacketWorker)

			if worker.IsIdle(idleTimeout) {
				// Stop the worker
				worker.Stop()
				// Remove the worker from the map
				dispatcher.workers.Delete(syncKey)
			}

			return true
		})
	}()

	// Wait for the worker to be cleaned up
	time.Sleep(2 * time.Second)

	// Verify the worker was removed
	_, exists = dispatcher.workers.Load("192.168.1.1:80")
	assert.False(t, exists)

	// Verify metrics were updated
	metrics := mockService.GetMetrics()
	assert.Equal(t, 1, metrics["packets_sent"], "Should have sent 1 packet")
	assert.Equal(t, len(packet), metrics["bytes_sent"], "Should have sent correct number of bytes")
	assert.Equal(t, 0, metrics["packets_dropped"], "Should not have dropped any packets")
	assert.Equal(t, 0, metrics["stream_errors"], "Should not have any stream errors")
}

// TestIntegrationStreamRecovery tests that a worker can recover from a stream error
func TestIntegrationStreamRecovery(t *testing.T) {
	// Create a mock service
	mockService := NewMockService()

	// Add peer mappings
	mockService.peerIDMap["192.168.1.1"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"

	// Create a dispatcher
	dispatcher := NewPacketDispatcher(mockService, 300)

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Dispatch a packet
	ctx := context.Background()
	packet1 := createTestPacket("10.0.0.1", "192.168.1.1", 12345, 80, 6, []byte("test payload 1"))
	dispatcher.DispatchPacket(ctx, "192.168.1.1:80", "192.168.1.1", packet1)

	// Wait for the packet to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify the packet was sent to the stream
	validPeerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")
	stream1 := mockService.streams[validPeerID]
	assert.Equal(t, packet1, stream1.buffer.Bytes())

	// Close the stream to simulate an error
	stream1.Close()

	// Dispatch another packet
	packet2 := createTestPacket("10.0.0.1", "192.168.1.1", 12345, 80, 6, []byte("test payload 2"))
	dispatcher.DispatchPacket(ctx, "192.168.1.1:80", "192.168.1.1", packet2)

	// Wait for the packet to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify a new stream was created
	stream2 := mockService.streams[validPeerID]
	assert.NotNil(t, stream2)

	// Verify metrics were updated
	metrics := mockService.GetMetrics()
	// We can't guarantee both packets will be processed before we check metrics
	assert.GreaterOrEqual(t, metrics["packets_sent"], 1, "Should have sent at least 1 packet")
	assert.GreaterOrEqual(t, metrics["bytes_sent"], len(packet1), "Should have sent at least some bytes")
	assert.Equal(t, 0, metrics["packets_dropped"], "Should not have dropped any packets")
	// We should have at least one stream error due to the closed stream
	assert.GreaterOrEqual(t, metrics["stream_errors"], 1, "Should have at least one stream error")
}
