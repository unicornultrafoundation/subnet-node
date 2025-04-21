package vpn_test

import (
	"errors"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/testutil"
)

// TestStreamServicePerformance tests the performance of the stream service
func TestStreamServicePerformance(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Create a mock stream service with minimal latency and failure rate
	mockStreamConfig := &testutil.MockServiceConfig{
		Latency:     1 * time.Millisecond,
		Jitter:      0,
		FailureRate: 0.01, // 1% failure rate
		PacketLoss:  0.01, // 1% packet loss
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

	// Create a stream service with optimized parameters for performance testing
	streamService := stream.NewStreamService(
		mockStreamService,
		50,                   // maxStreamsPerPeer
		10,                   // minStreamsPerPeer
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

	// Define test cases with different performance configurations
	testCases := []struct {
		name   string
		config *testutil.PerformanceConfig
	}{
		{
			name:   "Low Load",
			config: &testutil.PerformanceConfig{
				PacketSize:       1024,
				PacketsPerSecond: 100,
				Duration:         2 * time.Second,
				Concurrency:      1,
				WarmupDuration:   500 * time.Millisecond,
				CooldownDuration: 500 * time.Millisecond,
			},
		},
		{
			name:   "Medium Load",
			config: &testutil.PerformanceConfig{
				PacketSize:       1024,
				PacketsPerSecond: 500,
				Duration:         2 * time.Second,
				Concurrency:      2,
				WarmupDuration:   500 * time.Millisecond,
				CooldownDuration: 500 * time.Millisecond,
			},
		},
		{
			name:   "High Load",
			config: &testutil.PerformanceConfig{
				PacketSize:       1024,
				PacketsPerSecond: 1000,
				Duration:         2 * time.Second,
				Concurrency:      4,
				WarmupDuration:   500 * time.Millisecond,
				CooldownDuration: 500 * time.Millisecond,
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run the performance test
			result := testutil.RunStreamPerformanceTest(t, streamService, peerID, tc.config)

			// Verify that the performance meets the requirements
			// These are just example requirements, adjust based on your system
			switch tc.name {
			case "Low Load":
				testutil.VerifyPerformanceRequirements(t, result, 90, 10*time.Millisecond)
			case "Medium Load":
				testutil.VerifyPerformanceRequirements(t, result, 450, 20*time.Millisecond)
			case "High Load":
				testutil.VerifyPerformanceRequirements(t, result, 900, 30*time.Millisecond)
			}
		})
	}
}

// TestPacketDispatcherPerformance tests the performance of the packet dispatcher
func TestPacketDispatcherPerformance(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Create a test fixture with good network conditions
	fixture := testutil.NewTestFixture(t, &testutil.NetworkCondition{
		Name:       "performance",
		Latency:    1 * time.Millisecond,
		Jitter:     0,
		PacketLoss: 0.01,
		Bandwidth:  1000000, // 1 Gbps
	}, 3)
	defer fixture.Cleanup()

	// Define test cases with different performance configurations
	testCases := []struct {
		name   string
		config *testutil.PerformanceConfig
	}{
		{
			name:   "Low Load",
			config: &testutil.PerformanceConfig{
				PacketSize:       1024,
				PacketsPerSecond: 100,
				Duration:         2 * time.Second,
				Concurrency:      1,
				WarmupDuration:   500 * time.Millisecond,
				CooldownDuration: 500 * time.Millisecond,
			},
		},
		{
			name:   "Medium Load",
			config: &testutil.PerformanceConfig{
				PacketSize:       1024,
				PacketsPerSecond: 500,
				Duration:         2 * time.Second,
				Concurrency:      2,
				WarmupDuration:   500 * time.Millisecond,
				CooldownDuration: 500 * time.Millisecond,
			},
		},
		{
			name:   "High Load",
			config: &testutil.PerformanceConfig{
				PacketSize:       1024,
				PacketsPerSecond: 1000,
				Duration:         2 * time.Second,
				Concurrency:      4,
				WarmupDuration:   500 * time.Millisecond,
				CooldownDuration: 500 * time.Millisecond,
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run the performance test
			result := testutil.RunDispatcherPerformanceTest(t, fixture.Dispatcher, "192.168.1.1", tc.config)

			// Verify that the performance meets the requirements
			// These are just example requirements, adjust based on your system
			switch tc.name {
			case "Low Load":
				testutil.VerifyPerformanceRequirements(t, result, 90, 20*time.Millisecond)
			case "Medium Load":
				testutil.VerifyPerformanceRequirements(t, result, 450, 30*time.Millisecond)
			case "High Load":
				testutil.VerifyPerformanceRequirements(t, result, 900, 50*time.Millisecond)
			}
		})
	}
}

// TestResiliencePerformance tests the performance of the resilience patterns
func TestResiliencePerformance(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Define test cases with different resilience configurations
	testCases := []struct {
		name   string
		config *resilience.ResilienceConfig
	}{
		{
			name:   "Fast Recovery",
			config: &resilience.ResilienceConfig{
				CircuitBreakerFailureThreshold: 3,
				CircuitBreakerResetTimeout:     100 * time.Millisecond,
				CircuitBreakerSuccessThreshold: 2,
				RetryMaxAttempts:               2,
				RetryInitialInterval:           10 * time.Millisecond,
				RetryMaxInterval:               50 * time.Millisecond,
			},
		},
		{
			name:   "Balanced",
			config: &resilience.ResilienceConfig{
				CircuitBreakerFailureThreshold: 5,
				CircuitBreakerResetTimeout:     500 * time.Millisecond,
				CircuitBreakerSuccessThreshold: 3,
				RetryMaxAttempts:               3,
				RetryInitialInterval:           50 * time.Millisecond,
				RetryMaxInterval:               200 * time.Millisecond,
			},
		},
		{
			name:   "High Resilience",
			config: &resilience.ResilienceConfig{
				CircuitBreakerFailureThreshold: 10,
				CircuitBreakerResetTimeout:     1 * time.Second,
				CircuitBreakerSuccessThreshold: 5,
				RetryMaxAttempts:               5,
				RetryInitialInterval:           100 * time.Millisecond,
				RetryMaxInterval:               500 * time.Millisecond,
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a resilience service with the test configuration
			resilienceService := testutil.SetupResilienceService(t, tc.config)

			// Create a circuit breaker
			circuitBreaker := resilienceService.GetCircuitBreakerManager().GetBreaker("test-circuit")
			require.NotNil(t, circuitBreaker)

			// Measure the time it takes to execute operations with the circuit breaker
			start := time.Now()

			// Execute a series of operations
			const numOperations = 1000
			successCount := 0
			failureCount := 0

			for i := 0; i < numOperations; i++ {
				// Simulate some operations succeeding and some failing
				shouldFail := i%5 == 0 // 20% failure rate

				err := circuitBreaker.Execute(func() error {
					if shouldFail {
						return errors.New("test error")
					}
					return nil
				})

				if err == nil {
					successCount++
				} else {
					failureCount++
				}
			}

			// Calculate the duration
			duration := time.Since(start)

			// Calculate operations per second
			opsPerSecond := float64(numOperations) / duration.Seconds()

			// Log results
			t.Logf("Resilience performance test results:")
			t.Logf("  Configuration: %s", tc.name)
			t.Logf("  Total operations: %d", numOperations)
			t.Logf("  Successful operations: %d (%.2f%%)", successCount, float64(successCount)/float64(numOperations)*100)
			t.Logf("  Failed operations: %d (%.2f%%)", failureCount, float64(failureCount)/float64(numOperations)*100)
			t.Logf("  Duration: %v", duration)
			t.Logf("  Operations per second: %.2f", opsPerSecond)

			// Verify that the performance meets the requirements
			// These are just example requirements, adjust based on your system
			switch tc.name {
			case "Fast Recovery":
				require.GreaterOrEqual(t, opsPerSecond, 5000.0, "Operations per second should be at least 5000")
			case "Balanced":
				require.GreaterOrEqual(t, opsPerSecond, 2000.0, "Operations per second should be at least 2000")
			case "High Resilience":
				require.GreaterOrEqual(t, opsPerSecond, 1000.0, "Operations per second should be at least 1000")
			}
		})
	}
}
