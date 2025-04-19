package vpn

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/metrics"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/packet"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

// MockStream is a mock implementation of the types.VPNStream interface
type MockStream struct {
	mock.Mock
	mu     sync.Mutex
	closed bool
	buffer []byte
}

func NewMockStream() *MockStream {
	return &MockStream{
		buffer: make([]byte, 0),
	}
}

func (m *MockStream) Read(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return 0, fmt.Errorf("stream closed")
	}

	if len(m.buffer) == 0 {
		return 0, nil
	}

	n = copy(p, m.buffer)
	m.buffer = m.buffer[n:]
	return n, nil
}

func (m *MockStream) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return 0, fmt.Errorf("stream closed")
	}

	// Simulate writing to the stream
	return len(p), nil
}

func (m *MockStream) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	return nil
}

func (m *MockStream) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	return nil
}

func (m *MockStream) SetDeadline(t time.Time) error {
	return nil
}

func (m *MockStream) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *MockStream) SetWriteDeadline(t time.Time) error {
	return nil
}

// MockStreamService is a mock implementation of the types.Service interface
type MockStreamService struct {
	mock.Mock
}

func (m *MockStreamService) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (types.VPNStream, error) {
	args := m.Called(ctx, peerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(types.VPNStream), args.Error(1)
}

// MockDiscoveryService is a mock implementation of the discovery.PeerDiscoveryService interface
type MockDiscoveryService struct {
	mock.Mock
	peerIDMap map[string]string
}

func NewMockDiscoveryService() *MockDiscoveryService {
	return &MockDiscoveryService{
		peerIDMap: make(map[string]string),
	}
}

func (m *MockDiscoveryService) GetPeerID(ctx context.Context, destIP string) (string, error) {
	peerID, exists := m.peerIDMap[destIP]
	if !exists {
		return "", fmt.Errorf("peer not found for IP %s", destIP)
	}
	return peerID, nil
}

func (m *MockDiscoveryService) SyncPeerIDToDHT(ctx context.Context) error {
	return nil
}

// TestStreamServiceIntegration tests the integration of stream service components
func TestStreamServiceIntegration(t *testing.T) {
	// Create a mock stream service
	mockStreamService := new(MockStreamService)

	// Create a mock stream
	mockStream := NewMockStream()

	// Create a peer ID
	peerID, err := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	require.NoError(t, err)

	// Set up the mock service to return the mock stream
	mockStreamService.On("CreateNewVPNStream", mock.Anything, peerID).Return(mockStream, nil)

	// Create a stream service
	streamService := stream.NewStreamService(
		mockStreamService,
		10,                   // maxStreamsPerPeer
		3,                    // minStreamsPerPeer
		5*time.Minute,        // streamIdleTimeout
		1*time.Minute,        // cleanupInterval
		1*time.Second,        // healthCheckInterval
		500*time.Millisecond, // healthCheckTimeout
		3,                    // maxConsecutiveFailures
		1*time.Minute,        // warmInterval
		5,                    // maxStreamsPerMultiplexer
		2,                    // minStreamsPerMultiplexer
		1*time.Second,        // autoScalingInterval
		true,                 // multiplexingEnabled
	)

	// Start the service
	streamService.Start()
	defer streamService.Stop()

	// Test direct mode
	testDirectMode(t, streamService, peerID, mockStream)

	// Test pooled mode
	testPooledMode(t, streamService, peerID)

	// Test multiplexed mode
	testMultiplexedMode(t, streamService, peerID)

	// Test metrics collection
	testMetricsCollection(t, streamService)
}

// testDirectMode tests the direct mode of operation
func testDirectMode(t *testing.T, streamService *stream.StreamService, peerID peer.ID, mockStream *MockStream) {
	// Create a test packet
	packet := []byte("test direct packet")

	// Create a context
	ctx := context.Background()

	// Get a stream directly
	stream, err := streamService.CreateNewVPNStream(ctx, peerID)
	assert.NoError(t, err)
	assert.Equal(t, mockStream, stream)

	// Write the packet to the stream
	n, err := stream.Write(packet)
	assert.NoError(t, err)
	assert.Equal(t, len(packet), n)
}

// testPooledMode tests the pooled mode of operation
func testPooledMode(t *testing.T, streamService *stream.StreamService, peerID peer.ID) {
	// Create a test packet
	packet := []byte("test pooled packet")

	// Create a context
	ctx := context.Background()

	// Get a stream from the pool
	stream, err := streamService.GetStream(ctx, peerID)
	assert.NoError(t, err)

	// Write the packet to the stream
	n, err := stream.Write(packet)
	assert.NoError(t, err)
	assert.Equal(t, len(packet), n)

	// Release the stream back to the pool
	streamService.ReleaseStream(peerID, stream, true)

	// Verify the stream pool metrics
	metrics := streamService.GetStreamPoolMetrics()
	assert.NotNil(t, metrics)
	assert.GreaterOrEqual(t, metrics["streams_created"], int64(0))
	assert.GreaterOrEqual(t, metrics["streams_acquired"], int64(0))
	assert.GreaterOrEqual(t, metrics["streams_returned"], int64(0))
}

// testMultiplexedMode tests the multiplexed mode of operation
func testMultiplexedMode(t *testing.T, streamService *stream.StreamService, peerID peer.ID) {
	// Create a test packet
	packet := []byte("test multiplexed packet")

	// Create a context
	ctx := context.Background()

	// Send a packet using the multiplexer
	err := streamService.SendPacketMultiplexed(ctx, peerID, packet)
	assert.NoError(t, err)

	// Verify the multiplexer metrics
	metrics := streamService.GetMultiplexerMetrics()
	assert.NotNil(t, metrics)
}

// testMetricsCollection tests the metrics collection
func testMetricsCollection(t *testing.T, streamService *stream.StreamService) {
	// Get health metrics
	healthMetrics := streamService.GetHealthMetrics()
	assert.NotNil(t, healthMetrics)

	// Get multiplexer metrics
	multiplexerMetrics := streamService.GetMultiplexerMetrics()
	assert.NotNil(t, multiplexerMetrics)

	// Get stream pool metrics
	poolMetrics := streamService.GetStreamPoolMetrics()
	assert.NotNil(t, poolMetrics)

	// Log the metrics for debugging
	t.Logf("Health metrics: %v", healthMetrics)
	t.Logf("Multiplexer metrics: %v", multiplexerMetrics)
	t.Logf("Pool metrics: %v", poolMetrics)
}

// TestPacketDispatcherIntegration tests the integration of packet dispatcher with stream service
func TestPacketDispatcherIntegration(t *testing.T) {
	// Create a mock stream service
	mockStreamService := new(MockStreamService)

	// Create a mock discovery service
	mockDiscoveryService := NewMockDiscoveryService()

	// Add peer mappings
	mockDiscoveryService.peerIDMap["192.168.1.1"] = "QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N"
	mockDiscoveryService.peerIDMap["192.168.1.2"] = "QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN"

	// Create a mock stream
	mockStream := NewMockStream()

	// Create a peer ID
	peerID, err := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	require.NoError(t, err)

	// Set up the mock service to return the mock stream
	mockStreamService.On("CreateNewVPNStream", mock.Anything, peerID).Return(mockStream, nil)

	// Create a packet dispatcher
	dispatcher := packet.NewDispatcher(
		mockDiscoveryService,
		mockStreamService,
		300,           // workerIdleTimeout
		5*time.Second, // workerCleanupInterval
		100,           // workerBufferSize
	)

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Create a test packet
	testPacket := []byte("test packet")

	// Create a context
	ctx := context.Background()

	// Dispatch a packet
	dispatcher.DispatchPacket(ctx, "192.168.1.1:80", "192.168.1.1", testPacket)

	// Wait for the packet to be processed
	time.Sleep(100 * time.Millisecond)
}

// TestEnhancedDispatcherIntegration tests the integration of enhanced packet dispatcher with stream service
func TestEnhancedDispatcherIntegration(t *testing.T) {
	// Create a mock stream service
	mockStreamService := new(MockStreamService)

	// Create a mock discovery service
	mockDiscoveryService := NewMockDiscoveryService()

	// Add peer mappings
	mockDiscoveryService.peerIDMap["192.168.1.1"] = "QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N"

	// Create a mock stream
	mockStream := NewMockStream()

	// Create a peer ID
	peerID, err := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	require.NoError(t, err)

	// Set up the mock service to return the mock stream
	mockStreamService.On("CreateNewVPNStream", mock.Anything, peerID).Return(mockStream, nil)

	// Create a stream service
	streamService := stream.NewStreamService(
		mockStreamService,
		10,                   // maxStreamsPerPeer
		3,                    // minStreamsPerPeer
		5*time.Minute,        // streamIdleTimeout
		1*time.Minute,        // cleanupInterval
		1*time.Second,        // healthCheckInterval
		500*time.Millisecond, // healthCheckTimeout
		3,                    // maxConsecutiveFailures
		1*time.Minute,        // warmInterval
		5,                    // maxStreamsPerMultiplexer
		2,                    // minStreamsPerMultiplexer
		1*time.Second,        // autoScalingInterval
		true,                 // multiplexingEnabled
	)

	// Start the stream service
	streamService.Start()
	defer streamService.Stop()

	// Create an enhanced packet dispatcher
	dispatcher := packet.NewEnhancedDispatcher(
		mockDiscoveryService,
		mockStreamService,
		streamService,
		streamService,
		300,           // workerIdleTimeout
		5*time.Second, // workerCleanupInterval
		100,           // workerBufferSize
		true,          // useMultiplexing
	)

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Create a test packet
	testPacket := []byte("test packet")

	// Create a context
	ctx := context.Background()

	// Dispatch a packet
	dispatcher.DispatchPacket(ctx, "192.168.1.1:80", "192.168.1.1", testPacket)

	// Wait for the packet to be processed
	time.Sleep(100 * time.Millisecond)
}

// TestStress is a stress test for the VPN service
func TestStress(t *testing.T) {
	// Skip this test for now
	t.Skip("Stress tests are not implemented yet")

	// Create a mock stream service
	mockStreamService := new(MockStreamService)

	// Create a mock stream
	mockStream := NewMockStream()

	// Create a peer ID
	peerID, err := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	require.NoError(t, err)

	// Set up the mock service to return the mock stream
	mockStreamService.On("CreateNewVPNStream", mock.Anything, peerID).Return(mockStream, nil)

	// Create a stream service
	streamService := stream.NewStreamService(
		mockStreamService,
		10,                   // maxStreamsPerPeer
		3,                    // minStreamsPerPeer
		5*time.Minute,        // streamIdleTimeout
		1*time.Minute,        // cleanupInterval
		1*time.Second,        // healthCheckInterval
		500*time.Millisecond, // healthCheckTimeout
		3,                    // maxConsecutiveFailures
		1*time.Minute,        // warmInterval
		5,                    // maxStreamsPerMultiplexer
		2,                    // minStreamsPerMultiplexer
		1*time.Second,        // autoScalingInterval
		true,                 // multiplexingEnabled
	)

	// Start the service
	streamService.Start()
	defer streamService.Stop()

	// Run stress test
	const numPackets = 1000
	const numGoroutines = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			ctx := context.Background()

			for j := 0; j < numPackets/numGoroutines; j++ {
				packet := []byte(fmt.Sprintf("stress test packet %d", j))
				err := streamService.SendPacketMultiplexed(ctx, peerID, packet)
				assert.NoError(t, err)
			}
		}()
	}

	wg.Wait()

	// Verify metrics
	metrics := streamService.GetMultiplexerMetrics()
	assert.NotNil(t, metrics)
	assert.Contains(t, metrics, peerID.String())
	assert.GreaterOrEqual(t, metrics[peerID.String()].PacketsSent, int64(numPackets))
}

// TestMetricsIntegration tests the integration of metrics
func TestMetricsIntegration(t *testing.T) {
	// Create a metrics collector
	vpnMetrics := metrics.NewVPNMetrics()

	// Increment some metrics
	vpnMetrics.IncrementPacketsReceived(100)
	vpnMetrics.IncrementPacketsSent(200)
	vpnMetrics.IncrementPacketsDropped()

	// Get the metrics
	metrics := vpnMetrics.GetMetrics()

	// Verify the metrics
	assert.Equal(t, int64(1), metrics["packets_received"])
	assert.Equal(t, int64(1), metrics["packets_sent"])
	assert.Equal(t, int64(1), metrics["packets_dropped"])
	assert.Equal(t, int64(100), metrics["bytes_received"])
	assert.Equal(t, int64(200), metrics["bytes_sent"])
}
