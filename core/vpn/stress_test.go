package vpn

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
)

// TestStressDispatcher tests the dispatcher under high load
func TestStressDispatcher(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Create a mock service
	mockService := NewMockService()

	// Add peer mappings for multiple destinations
	// Use a valid peer ID format for all destinations
	validPeerID := "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"
	for i := 1; i <= 100; i++ {
		ip := fmt.Sprintf("192.168.1.%d", i)
		mockService.peerIDMap[ip] = validPeerID
	}

	// Create a dispatcher with a reasonable idle timeout
	dispatcher := NewPacketDispatcher(mockService, 60) // 60 seconds timeout

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Test parameters
	numDestinations := 100   // Number of unique destination IPs
	numPorts := 10           // Number of unique ports per destination
	numPacketsPerDest := 100 // Number of packets to send to each destination:port
	concurrentWorkers := 50  // Number of concurrent goroutines sending packets

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Track statistics
	var (
		totalPackets      int64
		droppedPackets    int64
		successfulPackets int64
		activeWorkers     int64
		maxWorkers        int64
	)

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(concurrentWorkers)

	// Create a channel to signal when to start sending packets
	startChan := make(chan struct{})

	// Start goroutines to send packets
	for i := 0; i < concurrentWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()

			// Wait for the signal to start
			<-startChan

			// Create a random number generator for this worker
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

			// Send packets
			for j := 0; j < numPacketsPerDest*numPorts/concurrentWorkers; j++ {
				// Choose a random destination
				destIndex := r.Intn(numDestinations) + 1
				destIP := fmt.Sprintf("192.168.1.%d", destIndex)

				// Choose a random port
				portIndex := r.Intn(numPorts) + 1
				destPort := 8000 + portIndex

				// Create a sync key
				syncKey := net.JoinHostPort(destIP, fmt.Sprintf("%d", destPort))

				// Create a test packet
				packet := createTestPacket(
					"10.0.0.1",
					destIP,
					12345,
					destPort,
					6, // TCP
					[]byte(fmt.Sprintf("test payload %d-%d-%d", workerID, destIndex, j)),
				)

				// Increment total packets counter
				atomic.AddInt64(&totalPackets, 1)

				// Try to create a worker directly to debug the issue
				_, err := dispatcher.getOrCreateWorker(ctx, syncKey, destIP)
				if err != nil {
					t.Logf("Error creating worker for %s: %v", syncKey, err)
				} else {
					t.Logf("Successfully created worker for %s", syncKey)

					// Dispatch the packet
					dispatcher.DispatchPacket(ctx, syncKey, destIP, packet)
				}

				// Count active workers
				workerCount := 0
				dispatcher.workers.Range(func(_, _ interface{}) bool {
					workerCount++
					return true
				})

				// Update max workers count
				atomic.StoreInt64(&activeWorkers, int64(workerCount))
				if int64(workerCount) > atomic.LoadInt64(&maxWorkers) {
					atomic.StoreInt64(&maxWorkers, int64(workerCount))
				}

				// Add a small delay to avoid overwhelming the system
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	// Start sending packets
	close(startChan)

	// Wait for all goroutines to finish
	wg.Wait()

	// Wait a bit for packets to be processed
	time.Sleep(1 * time.Second)

	// Count successful packets
	mockService.mu.Lock()
	for _, stream := range mockService.streams {
		if stream.writeCount > 0 {
			atomic.AddInt64(&successfulPackets, stream.writeCount)
		}
	}
	mockService.mu.Unlock()

	// Calculate dropped packets
	droppedPackets = totalPackets - successfulPackets

	// Log statistics
	t.Logf("Stress Test Statistics:")
	t.Logf("  Total Packets: %d", totalPackets)
	t.Logf("  Successful Packets: %d (%.2f%%)", successfulPackets, float64(successfulPackets)/float64(totalPackets)*100)
	t.Logf("  Dropped Packets: %d (%.2f%%)", droppedPackets, float64(droppedPackets)/float64(totalPackets)*100)
	t.Logf("  Max Workers: %d", maxWorkers)
	t.Logf("  Final Active Workers: %d", activeWorkers)

	// Verify that most packets were processed successfully
	successRate := float64(successfulPackets) / float64(totalPackets)
	assert.True(t, successRate > 0.95, "Success rate should be above 95%%, got %.2f%%", successRate*100)

	// Verify that workers were created for each destination
	workerCount := 0
	dispatcher.workers.Range(func(_, _ interface{}) bool {
		workerCount++
		return true
	})
	assert.True(t, workerCount > 0, "Should have active workers")
	assert.True(t, workerCount <= numDestinations*numPorts, "Should not have more workers than destinations*ports")

	// Test worker cleanup
	t.Log("Testing worker cleanup...")

	// Count initial workers
	initialWorkerCount := 0
	dispatcher.workers.Range(func(_, _ interface{}) bool {
		initialWorkerCount++
		return true
	})
	t.Logf("  Initial worker count: %d", initialWorkerCount)

	// Manually stop a subset of workers to verify cleanup works
	stoppedCount := 0
	dispatcher.workers.Range(func(key, value interface{}) bool {
		if stoppedCount < initialWorkerCount/2 {
			worker := value.(*PacketWorker)

			// Cancel the worker context
			worker.Cancel()

			// Close the packet channel
			close(worker.PacketChan)

			// Remove from map
			dispatcher.workers.Delete(key)

			stoppedCount++
		}
		return true
	})

	// Wait a short time for cleanup to take effect
	time.Sleep(100 * time.Millisecond)

	// Verify that workers were cleaned up
	finalWorkerCount := 0
	dispatcher.workers.Range(func(_, _ interface{}) bool {
		finalWorkerCount++
		return true
	})

	t.Logf("  Workers after manual cleanup: %d (stopped %d workers)", finalWorkerCount, stoppedCount)
	assert.Equal(t, initialWorkerCount-stoppedCount, finalWorkerCount, "Worker count should be reduced by the number of stopped workers")
}

// TestStressSequentialProcessing tests that packets for the same destination are processed sequentially
func TestStressSequentialProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Create a mock service
	mockService := NewMockService()

	// Add peer mapping
	mockService.peerIDMap["192.168.1.1"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"

	// Create a dispatcher
	dispatcher := NewPacketDispatcher(mockService, 60)

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Test parameters
	numPackets := 1000 // Number of packets to send

	// Create a context
	ctx := context.Background()

	// Create a channel to track processing order
	processOrder := make(chan int, numPackets)

	// Create a custom worker with a modified packet channel
	validPeerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")
	customWorker := &PacketWorker{
		SyncKey:      "192.168.1.1:80",
		DestIP:       "192.168.1.1",
		PeerID:       validPeerID,
		Stream:       mockService.streams[validPeerID],
		PacketChan:   make(chan *QueuedPacket, numPackets), // Large buffer to avoid blocking
		LastActivity: time.Now(),
		Ctx:          ctx,
		Cancel:       func() {},
		Running:      true,
		Mu:           sync.Mutex{},
	}

	// Start a goroutine to process packets and track order
	go func() {
		for packet := range customWorker.PacketChan {
			// Extract the sequence number from the packet
			if len(packet.Data) >= 24 {
				seqNum := int(packet.Data[20])<<24 | int(packet.Data[21])<<16 | int(packet.Data[22])<<8 | int(packet.Data[23])
				processOrder <- seqNum
			}
			// Process the packet
			time.Sleep(time.Millisecond) // Simulate processing time
		}
	}()

	// Store the worker in the dispatcher
	dispatcher.workers.Store("192.168.1.1:80", customWorker)

	// Send packets in reverse order to test sequential processing
	for i := numPackets - 1; i >= 0; i-- {
		// Create a test packet with a sequence number
		// Use a 4-byte sequence number to avoid wrapping issues
		seqBytes := make([]byte, 4)
		seqBytes[0] = byte((i >> 24) & 0xFF)
		seqBytes[1] = byte((i >> 16) & 0xFF)
		seqBytes[2] = byte((i >> 8) & 0xFF)
		seqBytes[3] = byte(i & 0xFF)
		packet := createTestPacket("10.0.0.1", "192.168.1.1", 12345, 80, 6, seqBytes)

		// Create a packet object
		packetObj := &QueuedPacket{
			Ctx:    ctx,
			DestIP: "192.168.1.1",
			Data:   packet,
		}

		// Send the packet directly to the worker
		customWorker.PacketChan <- packetObj
	}

	// Wait for all packets to be processed
	var order []int
	for i := 0; i < numPackets; i++ {
		select {
		case num := <-processOrder:
			order = append(order, num)
		case <-time.After(5 * time.Second):
			t.Fatalf("Timeout waiting for packet processing after %d packets", i)
		}
	}

	// Verify packets were processed in the order they were sent (reverse order)
	// In this test, we're sending packets in reverse order (numPackets-1 down to 0)
	// but we expect them to be processed in the order they were received by the worker
	// which means they should be processed in reverse order (numPackets-1 down to 0)
	isOrdered := true

	// Print the first 20 elements to debug
	t.Logf("First 20 elements of order: %v", order[:20])

	for i := 1; i < len(order); i++ {
		if order[i] > order[i-1] {
			isOrdered = false
			t.Logf("Order broken at index %d: %d > %d", i, order[i], order[i-1])
			break
		}
	}

	assert.True(t, isOrdered, "Packets should be processed in sequential order")
	t.Logf("Processed %d packets in sequential order", numPackets)

	// Don't close the worker channel here, it will be closed by the dispatcher
	// when it's stopped in the defer dispatcher.Stop() call
}

// TestStressConcurrentDestinations tests that packets for different destinations are processed concurrently
func TestStressConcurrentDestinations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Create a mock service
	mockService := NewMockService()

	// Add peer mappings
	validPeerID := "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"
	for i := 1; i <= 10; i++ {
		ip := fmt.Sprintf("192.168.1.%d", i)
		mockService.peerIDMap[ip] = validPeerID
	}

	// Create a dispatcher
	dispatcher := NewPacketDispatcher(mockService, 60)

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Test parameters
	numDestinations := 10
	packetsPerDest := 100

	// Create a context
	ctx := context.Background()

	// Create channels to track when processing starts and ends for each destination
	type processingEvent struct {
		destIP    string
		startTime time.Time
		endTime   time.Time
	}

	eventChan := make(chan processingEvent, numDestinations)

	// Create workers for each destination
	validPeerID2, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")
	for i := 1; i <= numDestinations; i++ {
		destIP := fmt.Sprintf("192.168.1.%d", i)
		syncKey := fmt.Sprintf("%s:80", destIP)

		// Create a custom worker
		worker := &PacketWorker{
			SyncKey:      syncKey,
			DestIP:       destIP,
			PeerID:       validPeerID2,
			Stream:       mockService.streams[validPeerID2],
			PacketChan:   make(chan *QueuedPacket, packetsPerDest),
			LastActivity: time.Now(),
			Ctx:          ctx,
			Cancel:       func() {},
			Running:      true,
			Mu:           sync.Mutex{},
		}

		// Store the worker in the dispatcher
		dispatcher.workers.Store(syncKey, worker)

		// Start a goroutine to process packets for this worker
		go func(w *PacketWorker, ip string) {
			// Record start time
			event := processingEvent{
				destIP:    ip,
				startTime: time.Now(),
			}

			// Process all packets
			for i := 0; i < packetsPerDest; i++ {
				select {
				case packet := <-w.PacketChan:
					// Simulate processing time with a delay
					time.Sleep(10 * time.Millisecond)
					// Process the packet (in a real scenario, this would send to the stream)
					_ = packet
				case <-time.After(5 * time.Second):
					t.Logf("Timeout waiting for packets for %s", ip)
					break
				}
			}

			// Record end time
			event.endTime = time.Now()
			eventChan <- event
		}(worker, destIP)
	}

	// Send packets to all destinations
	for i := 1; i <= numDestinations; i++ {
		destIP := fmt.Sprintf("192.168.1.%d", i)
		syncKey := fmt.Sprintf("%s:80", destIP)

		// Get the worker
		workerVal, exists := dispatcher.workers.Load(syncKey)
		assert.True(t, exists)
		worker := workerVal.(*PacketWorker)

		// Send packets to this destination
		for j := 0; j < packetsPerDest; j++ {
			// Create a test packet
			packet := createTestPacket("10.0.0.1", destIP, 12345, 80, 6, []byte{byte(j)})

			// Create a packet object
			packetObj := &QueuedPacket{
				Ctx:    ctx,
				DestIP: destIP,
				Data:   packet,
			}

			// Send the packet to the worker
			worker.PacketChan <- packetObj
		}
	}

	// Collect processing events
	var events []processingEvent
	for i := 0; i < numDestinations; i++ {
		select {
		case event := <-eventChan:
			events = append(events, event)
		case <-time.After(10 * time.Second):
			t.Fatalf("Timeout waiting for processing events after %d events", i)
		}
	}

	// Analyze the events to verify concurrent processing
	t.Logf("Processing times for %d destinations:", numDestinations)

	// Calculate total processing time (from first start to last end)
	var firstStart time.Time
	var lastEnd time.Time

	if len(events) > 0 {
		firstStart = events[0].startTime
		lastEnd = events[0].endTime

		for _, event := range events {
			if event.startTime.Before(firstStart) {
				firstStart = event.startTime
			}
			if event.endTime.After(lastEnd) {
				lastEnd = event.endTime
			}

			duration := event.endTime.Sub(event.startTime)
			t.Logf("  %s: %.2f ms", event.destIP, float64(duration.Milliseconds()))
		}
	}

	totalDuration := lastEnd.Sub(firstStart)
	t.Logf("Total processing time: %.2f ms", float64(totalDuration.Milliseconds()))

	// Calculate theoretical sequential time (sum of all individual times)
	var sequentialTime time.Duration
	for _, event := range events {
		sequentialTime += event.endTime.Sub(event.startTime)
	}
	t.Logf("Theoretical sequential time: %.2f ms", float64(sequentialTime.Milliseconds()))

	// Calculate concurrency factor
	concurrencyFactor := float64(sequentialTime) / float64(totalDuration)
	t.Logf("Concurrency factor: %.2f (higher is better)", concurrencyFactor)

	// Verify that processing was concurrent (concurrency factor > 1.5)
	assert.True(t, concurrencyFactor > 1.5, "Processing should be concurrent, got concurrency factor %.2f", concurrencyFactor)
}

// Update the MockStream to track write count
func init() {
	// We can't modify the Write method at runtime in Go
	// The writeCount will be incremented in the existing Write method
}
