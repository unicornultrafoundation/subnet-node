package vpn_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/unicornultrafoundation/subnet-node/core/vpn/metrics"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/packet"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/testutil"
)

// TestStressMetrics tests the metrics collection under stress
func TestStressMetrics(t *testing.T) {
	// Create a metrics collector
	vpnMetrics := metrics.NewVPNMetrics()

	// Increment some metrics in parallel
	const numGoroutines = 5
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				vpnMetrics.IncrementPacketsReceived(100)
				vpnMetrics.IncrementPacketsSent(200)
				vpnMetrics.IncrementPacketsDropped()
			}
		}()
	}

	wg.Wait()

	// Get the metrics
	metrics := vpnMetrics.GetMetrics()

	// Verify the metrics
	assert.Equal(t, int64(numGoroutines*numOperations), metrics["packets_received"])
	assert.Equal(t, int64(numGoroutines*numOperations), metrics["packets_sent"])
	assert.Equal(t, int64(numGoroutines*numOperations), metrics["packets_dropped"])
	assert.Equal(t, int64(numGoroutines*numOperations*100), metrics["bytes_received"])
	assert.Equal(t, int64(numGoroutines*numOperations*200), metrics["bytes_sent"])
}

// TestStressStreamService tests the stream service under stress
func TestStressStreamService(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Create a mock stream service with some latency and failure rate
	mockStreamConfig := &testutil.MockServiceConfig{
		Latency:     1 * time.Millisecond,
		Jitter:      0,
		FailureRate: 0.01, // 1% failure rate
	}
	mockStreamService := testutil.NewMockStreamService(mockStreamConfig)

	// Create a peer ID
	peerID, err := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	require.NoError(t, err)

	// Set up the mock stream service to return a mock stream
	mockStream := testutil.SetupMockStream(t, &testutil.MockStreamConfig{
		Latency:     1 * time.Millisecond,
		Jitter:      0,
		PacketLoss:  0.01,
		FailureRate: 0.01,
	})
	mockStreamService.On("CreateNewVPNStream", mock.Anything, peerID).Return(mockStream, nil)

	// Create a stream service with realistic parameters
	streamService := stream.NewStreamService(
		mockStreamService,
		5,                    // minStreamsPerPeer
		5*time.Minute,        // streamIdleTimeout
		30*time.Second,       // cleanupInterval
		5*time.Second,        // healthCheckInterval
		500*time.Millisecond, // healthCheckTimeout
		3,                    // maxConsecutiveFailures
		30*time.Second,       // warmInterval
	)

	// Start the service
	streamService.Start()
	defer streamService.Stop()

	// Run a single test with fixed parameters
	stressTestStreamService(t, streamService, peerID, 1024, 50)

	// Get and log the service metrics
	healthMetrics := streamService.GetHealthMetrics()
	poolMetrics := streamService.GetStreamPoolMetrics()

	t.Logf("Health metrics: %v", healthMetrics)
	t.Logf("Pool metrics: %v", poolMetrics)
	t.Logf("Mock service stats: %v", mockStreamService.GetStats())
}

// stressTestStreamService runs a simplified stress test on the stream service
func stressTestStreamService(t *testing.T, streamService *stream.StreamService, peerID peer.ID, packetSize, packetsPerSecond int) {
	// Use a fixed test duration
	testDuration := 1 * time.Second
	totalPackets := int(float64(packetsPerSecond) * testDuration.Seconds())

	// Create a fixed-size packet
	packet := testutil.CreateTestPacket(packetSize)

	// Start time
	startTime := time.Now()

	// Track packet count
	var successCount int64

	// Send packets at the specified rate
	ctx := context.Background()

	// Use a single goroutine for simplicity
	for i := 0; i < totalPackets; i++ {
		// Get a stream from the pool
		stream, err := streamService.GetStream(ctx, peerID)
		if err == nil {
			// Write to the stream
			_, err = stream.Write(packet)
			// Release the stream back to the pool
			streamService.ReleaseStream(peerID, stream, err == nil)
		}

		// Count successful operations
		if err == nil {
			atomic.AddInt64(&successCount, 1)
		}

		// Add a small sleep to avoid overwhelming the system
		time.Sleep(time.Duration(20) * time.Millisecond)
	}

	// End time
	_ = time.Since(startTime) // Calculate duration but don't use it

	// Log basic results
	t.Logf("Total packets: %d, Successful: %d", totalPackets, atomic.LoadInt64(&successCount))
}

// TestStressHighConcurrentStreams tests the ability to handle a high number of concurrent streams to the same peer
// after removing the maxStreamsPerPeer restriction
func TestStressHighConcurrentStreams(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Create a mock stream service with minimal latency and failure rate
	mockStreamConfig := &testutil.MockServiceConfig{
		Latency:     1 * time.Millisecond,
		Jitter:      0,
		FailureRate: 0.01, // 1% failure rate
	}
	mockStreamService := testutil.NewMockStreamService(mockStreamConfig)

	// Create a single peer ID to test high concurrency to the same peer
	peerID, err := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	require.NoError(t, err)

	// Set up the mock stream service to return a unique mock stream for each call
	// This simulates creating many unique streams to the same peer
	mockStream := testutil.SetupMockStream(t, &testutil.MockStreamConfig{
		Latency:     1 * time.Millisecond,
		Jitter:      0,
		PacketLoss:  0.01,
		FailureRate: 0.01,
	})
	mockStreamService.On("CreateNewVPNStream", mock.Anything, peerID).Return(mockStream, nil).Maybe()

	// Create a stream service with realistic parameters
	streamService := stream.NewStreamService(
		mockStreamService,
		5,                    // minStreamsPerPeer
		5*time.Minute,        // streamIdleTimeout
		30*time.Second,       // cleanupInterval
		5*time.Second,        // healthCheckInterval
		500*time.Millisecond, // healthCheckTimeout
		3,                    // maxConsecutiveFailures
		30*time.Second,       // warmInterval
	)

	// Start the service
	streamService.Start()
	defer streamService.Stop()

	// Test parameters
	testDuration := 2 * time.Second
	concurrentStreams := 100 // We'll create 100 concurrent streams to the same peer

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	// Create a wait group to synchronize goroutines
	var wg sync.WaitGroup
	wg.Add(concurrentStreams)

	// Track success and failure counts
	var successCount, failureCount int64

	// Start time
	startTime := time.Now()

	// Launch concurrent goroutines to create and use streams
	for i := 0; i < concurrentStreams; i++ {
		go func(id int) {
			defer wg.Done()

			// Create a fixed-size packet
			packet := testutil.CreateTestPacket(1024)

			// Try to get a stream from the pool
			stream, err := streamService.GetStream(ctx, peerID)
			if err != nil {
				atomic.AddInt64(&failureCount, 1)
				return
			}

			// Write to the stream
			_, err = stream.Write(packet)
			if err != nil {
				atomic.AddInt64(&failureCount, 1)
			} else {
				atomic.AddInt64(&successCount, 1)
			}

			// Release the stream back to the pool
			streamService.ReleaseStream(peerID, stream, err == nil)

			// Simulate some work with the stream
			time.Sleep(100 * time.Millisecond)
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// End time
	duration := time.Since(startTime)

	// Get the stream count for the peer
	streamCount := streamService.GetStreamCount(peerID)
	activeStreamCount := streamService.GetActiveStreamCount(peerID)

	// Get and log the service metrics
	healthMetrics := streamService.GetHealthMetrics()
	poolMetrics := streamService.GetStreamPoolMetrics()

	// Log results
	t.Logf("Test completed in %v", duration)
	t.Logf("Concurrent streams: %d", concurrentStreams)
	t.Logf("Success count: %d, Failure count: %d", successCount, failureCount)
	t.Logf("Stream count for peer: %d, Active stream count: %d", streamCount, activeStreamCount)
	t.Logf("Health metrics: %v", healthMetrics)
	t.Logf("Pool metrics: %v", poolMetrics)
	t.Logf("Mock service stats: %v", mockStreamService.GetStats())

	// Verify that we were able to create more than the previous limit (10) of streams
	assert.Greater(t, streamCount, 10, "Should be able to create more than the previous limit of streams")

	// Verify that most operations succeeded
	assert.Greater(t, successCount, int64(concurrentStreams*8/10), "Most operations should succeed")
}

// TestStressPacketDispatcher tests the packet dispatcher under stress
func TestStressPacketDispatcher(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Create mock services with some latency and failure rate
	mockStreamConfig := &testutil.MockServiceConfig{
		Latency:     1 * time.Millisecond,
		Jitter:      0,
		FailureRate: 0.01, // 1% failure rate
	}
	mockStreamService := testutil.NewMockStreamService(mockStreamConfig)

	mockDiscoveryConfig := &testutil.MockDiscoveryServiceConfig{
		Latency:     2 * time.Millisecond,
		Jitter:      0,
		FailureRate: 0.02, // 2% failure rate
	}
	mockDiscoveryService, peerIDs := testutil.SetupDiscoveryService(t, mockDiscoveryConfig, 3)

	// Set up the mock stream service to return a mock stream for each peer ID
	mockStream := testutil.SetupMockStream(t, &testutil.MockStreamConfig{
		Latency:     1 * time.Millisecond,
		Jitter:      0,
		PacketLoss:  0.01,
		FailureRate: 0.01,
	})

	// Set up expectations for each peer ID
	for _, peerID := range peerIDs {
		mockStreamService.On("CreateNewVPNStream", mock.Anything, peerID).Return(mockStream, nil)
	}

	// Create a stream service
	streamService := stream.NewStreamService(
		mockStreamService,
		5,                    // minStreamsPerPeer
		5*time.Minute,        // streamIdleTimeout
		30*time.Second,       // cleanupInterval
		5*time.Second,        // healthCheckInterval
		500*time.Millisecond, // healthCheckTimeout
		3,                    // maxConsecutiveFailures
		30*time.Second,       // warmInterval
	)

	// Start the stream service
	streamService.Start()
	defer streamService.Stop()

	// Create a resilience service for testing
	resilienceConfig := &resilience.ResilienceConfig{
		CircuitBreakerFailureThreshold: 5,
		CircuitBreakerResetTimeout:     1 * time.Second,
		CircuitBreakerSuccessThreshold: 2,
		RetryMaxAttempts:               3,
		RetryInitialInterval:           50 * time.Millisecond,
		RetryMaxInterval:               500 * time.Millisecond,
	}
	resilienceService := resilience.NewResilienceService(resilienceConfig)

	// Create a packet dispatcher
	dispatcher := packet.NewDispatcher(
		mockDiscoveryService,
		mockStreamService,
		streamService,
		300,           // workerIdleTimeout
		5*time.Second, // workerCleanupInterval
		1000,          // workerBufferSize
		resilienceService,
	)

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Run a single stress test with fixed parameters
	stressTestPacketDispatcher(t, dispatcher, 1024, 1)

	// Log statistics
	t.Logf("Mock stream service stats: %v", mockStreamService.GetStats())
	t.Logf("Mock discovery service stats: %v", mockDiscoveryService.GetStats())
}

// stressTestPacketDispatcher runs a simplified stress test on the packet dispatcher
func stressTestPacketDispatcher(t *testing.T, dispatcher *packet.Dispatcher, packetSize, destinationCount int) {
	// Create destinations
	destinations := make([]string, destinationCount)
	for i := 0; i < destinationCount; i++ {
		destinations[i] = fmt.Sprintf("192.168.1.%d", (i%3)+1)
	}

	// Create a fixed-size packet
	packet := testutil.CreateTestPacket(packetSize)

	// Test parameters - reduced for faster execution
	testDuration := 2 * time.Second
	packetsPerSecond := 100
	totalPackets := int(float64(packetsPerSecond) * testDuration.Seconds())

	// Start time
	startTime := time.Now()

	// Track packet count
	var sentCount int64

	// Use a single goroutine for simplicity
	ctx := context.Background()

	// Create a fixed sync key for testing
	destIP := destinations[0]
	syncKey := fmt.Sprintf("%s:%d", destIP, 1000)

	for i := 0; i < totalPackets; i++ {
		// Create a done channel to get the result
		doneCh := make(chan error, 1)

		// Dispatch the packet with a done channel
		dispatcher.DispatchPacketWithCallback(ctx, syncKey, destIP, packet, doneCh)

		// Increment sent count
		atomic.AddInt64(&sentCount, 1)

		// Wait for the result or timeout
		select {
		case <-doneCh:
			// The packet was processed (successfully or with error)
		case <-time.After(100 * time.Millisecond):
			// Timeout - packet was not processed in time
		}

		// Add a small sleep to avoid overwhelming the system
		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	// End time
	_ = time.Since(startTime) // Calculate duration but don't use it

	// Get worker metrics
	workerMetrics := dispatcher.GetWorkerMetrics()

	// Calculate basic metrics
	var totalPacketCount int64
	var totalErrorCount int64

	// Process the metrics
	for _, metrics := range workerMetrics {
		totalPacketCount += metrics.PacketCount
		totalErrorCount += metrics.ErrorCount
	}

	// Calculate success rate
	successRate := 0.0
	if totalPacketCount > 0 {
		successRate = 100.0 * (float64(totalPacketCount-totalErrorCount) / float64(totalPacketCount))
	}

	// Log basic results
	t.Logf("Packets sent: %d, Processed: %d, Success rate: %.2f%%",
		sentCount, totalPacketCount, successRate)

	// Verify that packets were processed
	assert.Greater(t, totalPacketCount, int64(0), "No packets were processed")
}
