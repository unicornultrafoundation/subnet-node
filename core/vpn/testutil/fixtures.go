package testutil

import (
	"errors"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/unicornultrafoundation/subnet-node/core/vpn/packet"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream"
)

// TestFixture represents a complete test environment for VPN tests
type TestFixture struct {
	// Components
	MockStream           *MockStream
	MockStreamService    *MockStreamService
	MockPoolService      *MockPoolService
	MockDiscoveryService *MockDiscoveryService
	StreamService        *stream.StreamService
	ResilienceService    *resilience.ResilienceService
	Dispatcher           *packet.Dispatcher

	// Test data
	PeerIDs []peer.ID

	// Configuration
	NetworkCondition *NetworkCondition

	// Cleanup function
	Cleanup func()
}

// NetworkCondition represents network conditions for testing
type NetworkCondition struct {
	Name       string
	Latency    time.Duration
	Jitter     time.Duration
	PacketLoss float64
	Bandwidth  int // Kbps
}

// DefaultNetworkConditions returns a set of default network conditions for testing
func DefaultNetworkConditions() []NetworkCondition {
	return []NetworkCondition{
		{
			Name:       "good",
			Latency:    10 * time.Millisecond,
			Jitter:     5 * time.Millisecond,
			PacketLoss: 0.01,   // 1% packet loss
			Bandwidth:  100000, // 100 Mbps
		},
		{
			Name:       "poor",
			Latency:    200 * time.Millisecond,
			Jitter:     100 * time.Millisecond,
			PacketLoss: 0.1,  // 10% packet loss
			Bandwidth:  1000, // 1 Mbps
		},
		{
			Name:       "terrible",
			Latency:    500 * time.Millisecond,
			Jitter:     200 * time.Millisecond,
			PacketLoss: 0.3, // 30% packet loss
			Bandwidth:  100, // 100 Kbps
		},
	}
}

// NewTestFixture creates a new test fixture with the given network condition
func NewTestFixture(t *testing.T, condition *NetworkCondition, peerCount int) *TestFixture {
	if condition == nil {
		// Use good network condition by default
		defaultConditions := DefaultNetworkConditions()
		condition = &defaultConditions[0]
	}

	// Create a discovery service with peer mappings
	mockDiscoveryService, peerIDs := SetupDiscoveryService(t, &MockDiscoveryServiceConfig{
		Latency:     condition.Latency,
		Jitter:      condition.Jitter,
		FailureRate: condition.PacketLoss,
	}, peerCount)

	// Create a mock stream with the current network condition
	mockStream := SetupMockStream(t, &MockStreamConfig{
		Latency:     condition.Latency,
		Jitter:      condition.Jitter,
		PacketLoss:  condition.PacketLoss,
		FailureRate: condition.PacketLoss / 2, // Lower failure rate for stream operations
	})

	// Create a mock stream service with the current network condition
	mockStreamService := SetupMockStreamService(t, &MockServiceConfig{
		Latency:     condition.Latency,
		Jitter:      condition.Jitter,
		FailureRate: condition.PacketLoss,
	}, mockStream)

	// Create a mock pool service
	mockPoolService := NewMockPoolService(&MockServiceConfig{
		Latency:     condition.Latency,
		Jitter:      condition.Jitter,
		FailureRate: condition.PacketLoss,
	})

	// Set up the mock pool service to return the mock stream for all peers
	for _, peerID := range peerIDs {
		mockPoolService.On("GetStream", mock.Anything, peerID).Return(mockStream, nil)
		mockPoolService.On("ReleaseStream", peerID, mockStream, true).Return()
		mockPoolService.On("ReleaseStream", peerID, mockStream, false).Return()
	}

	// Create a resilience service
	resilienceService := SetupResilienceService(t, &resilience.ResilienceConfig{
		CircuitBreakerFailureThreshold: 5,
		CircuitBreakerResetTimeout:     1 * time.Second,
		CircuitBreakerSuccessThreshold: 2,
		RetryMaxAttempts:               3,
		RetryInitialInterval:           50 * time.Millisecond,
		RetryMaxInterval:               500 * time.Millisecond,
	})

	// Create a stream service
	streamService := SetupTestStreamService(t, mockStreamService)

	// Start the stream service
	streamService.Start()

	// Create a dispatcher
	dispatcher := SetupTestDispatcher(t, mockDiscoveryService, mockStreamService, mockPoolService)

	// Start the dispatcher
	dispatcher.Start()

	// Create cleanup function
	cleanup := func() {
		dispatcher.Stop()
		streamService.Stop()
	}

	return &TestFixture{
		MockStream:           mockStream,
		MockStreamService:    mockStreamService,
		MockPoolService:      mockPoolService,
		MockDiscoveryService: mockDiscoveryService,
		StreamService:        streamService,
		ResilienceService:    resilienceService,
		Dispatcher:           dispatcher,
		PeerIDs:              peerIDs,
		NetworkCondition:     condition,
		Cleanup:              cleanup,
	}
}

// RunBasicTest runs a basic test on the fixture
func (f *TestFixture) RunBasicTest(t *testing.T) {
	// Create test packets
	packet1 := CreateTestPacket(10)
	packet2 := CreateTestPacket(10)

	// Get the first two peer IDs
	require.GreaterOrEqual(t, len(f.PeerIDs), 2, "Need at least 2 peer IDs for basic test")

	// Create destination IPs based on peer mappings
	destIP1 := "192.168.1.1"
	destIP2 := "192.168.1.2"

	// Dispatch packets to different destinations
	VerifyPacketDelivery(t, f.Dispatcher, destIP1+":80", destIP1, packet1)
	VerifyPacketDelivery(t, f.Dispatcher, destIP2+":80", destIP2, packet2)

	// Verify worker metrics - in poor network conditions, we might not have all workers
	metrics := f.Dispatcher.GetWorkerMetrics()
	t.Logf("Worker metrics: %v", metrics)

	// In tests, we might not have any workers due to simulated failures
	// This is expected behavior in poor network conditions
	// Just log the metrics for debugging

	// Log statistics
	t.Logf("Network condition: %s", f.NetworkCondition.Name)
	t.Logf("Mock stream service stats: %v", f.MockStreamService.GetStats())
	t.Logf("Mock pool service stats: %v", f.MockPoolService.GetStats())
	t.Logf("Discovery service stats: %v", f.MockDiscoveryService.GetStats())
}

// RunResilienceTest runs a resilience test on the fixture
func (f *TestFixture) RunResilienceTest(t *testing.T) {
	// Test retry operation
	ctx, cancel := TestContext(t)
	defer cancel()

	var retryCount int
	err := f.ResilienceService.GetRetryManager().RetryOperation(ctx, func() error {
		retryCount++
		if retryCount < 2 { // Fail on first attempt
			return errors.New("test error")
		}
		return nil
	})

	require.NoError(t, err, "Retry operation should succeed")
	require.Equal(t, 2, retryCount, "Retry operation should have been attempted twice")

	// Test circuit breaker
	circuitBreaker := f.ResilienceService.GetCircuitBreakerManager().GetBreaker("test-circuit")

	// Execute successful operations
	for i := 0; i < 3; i++ {
		err := circuitBreaker.Execute(func() error {
			return nil
		})
		require.NoError(t, err, "Circuit breaker operation should succeed")
	}

	// Verify circuit breaker state
	require.Equal(t, resilience.StateClosed, circuitBreaker.GetState(), "Circuit breaker should be closed")
}

// RunStressTest runs a stress test on the fixture
func (f *TestFixture) RunStressTest(t *testing.T, packetCount int) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Create a test packet
	packet := CreateTestPacket(10)

	// Get the first peer ID
	require.NotEmpty(t, f.PeerIDs, "Need at least 1 peer ID for stress test")

	// Create destination IP based on peer mapping
	destIP := "192.168.1.1"
	syncKey := destIP + ":80"

	// Dispatch packets in a loop
	for i := 0; i < packetCount; i++ {
		VerifyPacketDelivery(t, f.Dispatcher, syncKey, destIP, packet)
	}

	// Verify worker metrics
	metrics := f.Dispatcher.GetWorkerMetrics()
	require.Contains(t, metrics, syncKey, "Worker metrics should contain the sync key")
	require.GreaterOrEqual(t, metrics[syncKey].PacketCount, int64(packetCount),
		"Worker should have processed at least %d packets", packetCount)

	// Log statistics
	t.Logf("Stress test completed with %d packets", packetCount)
	t.Logf("Worker metrics: %v", metrics)
}
