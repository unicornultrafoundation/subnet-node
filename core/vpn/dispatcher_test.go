package vpn

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStream is a mock implementation of the VPNStream interface
type MockStream struct {
	mock.Mock
	buffer     bytes.Buffer
	closed     bool
	mu         sync.Mutex
	writeCount int64 // Count of successful writes
}

// Ensure MockStream implements VPNStream
var _ VPNStream = (*MockStream)(nil)

func (m *MockStream) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockStream) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, fmt.Errorf("stream closed")
	}
	n, err = m.buffer.Write(p)
	if err == nil {
		m.writeCount++
	}
	return n, err
}

func (m *MockStream) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *MockStream) Reset() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStream) SetDeadline(time.Time) error {
	return nil
}

func (m *MockStream) SetReadDeadline(time.Time) error {
	return nil
}

func (m *MockStream) SetWriteDeadline(time.Time) error {
	return nil
}

// MockService is a mock implementation of the VPNServiceInterface
type MockService struct {
	mock.Mock
	peerIDMap         map[string]string
	streams           map[peer.ID]*MockStream
	mu                sync.Mutex // Mutex for thread safety
	workerIdleTimeout int        // Worker idle timeout in seconds
	workerBufferSize  int        // Worker buffer size

	// Metrics tracking
	metrics struct {
		packetsSent    int
		packetsDropped int
		streamErrors   int
		bytesSent      int
		metricsMu      sync.Mutex
	}
}

// Ensure MockService implements all interfaces
var (
	_ VPNService           = (*MockService)(nil)
	_ PeerDiscoveryService = (*MockService)(nil)
	_ StreamService        = (*MockService)(nil)
	_ RetryService         = (*MockService)(nil)
	_ MetricsService       = (*MockService)(nil)
	_ ConfigService        = (*MockService)(nil)
)

func NewMockService() *MockService {
	return &MockService{
		peerIDMap:         make(map[string]string),
		streams:           make(map[peer.ID]*MockStream),
		workerIdleTimeout: 300, // Default 5 minutes
		workerBufferSize:  100, // Default buffer size
	}
}

func (m *MockService) GetPeerID(ctx context.Context, destIP string) (string, error) {
	peerID, ok := m.peerIDMap[destIP]
	if !ok {
		return "", fmt.Errorf("no peer mapping found for IP %s", destIP)
	}
	return peerID, nil
}

func (m *MockService) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (VPNStream, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stream, ok := m.streams[peerID]
	if !ok {
		stream = &MockStream{}
		m.streams[peerID] = stream
	}
	return stream, nil
}

func (m *MockService) RetryOperation(ctx context.Context, operation func() error) error {
	return operation()
}

// IncrementStreamErrors implements the MetricsService
func (m *MockService) IncrementStreamErrors() {
	m.metrics.metricsMu.Lock()
	defer m.metrics.metricsMu.Unlock()
	m.metrics.streamErrors++
}

// IncrementPacketsSent implements the MetricsService
func (m *MockService) IncrementPacketsSent(bytes int) {
	m.metrics.metricsMu.Lock()
	defer m.metrics.metricsMu.Unlock()
	m.metrics.packetsSent++
	m.metrics.bytesSent += bytes
}

// IncrementPacketsDropped implements the MetricsService
func (m *MockService) IncrementPacketsDropped() {
	m.metrics.metricsMu.Lock()
	defer m.metrics.metricsMu.Unlock()
	m.metrics.packetsDropped++
}

// GetWorkerBufferSize implements the ConfigService
func (m *MockService) GetWorkerBufferSize() int {
	return m.workerBufferSize
}

// GetWorkerIdleTimeout implements the ConfigService
func (m *MockService) GetWorkerIdleTimeout() int {
	return m.workerIdleTimeout
}

// GetMetrics returns the current metrics
func (m *MockService) GetMetrics() map[string]int {
	m.metrics.metricsMu.Lock()
	defer m.metrics.metricsMu.Unlock()

	return map[string]int{
		"packets_sent":    m.metrics.packetsSent,
		"packets_dropped": m.metrics.packetsDropped,
		"stream_errors":   m.metrics.streamErrors,
		"bytes_sent":      m.metrics.bytesSent,
	}
}

// Helper function to create a test packet
func createTestPacket(srcIP, dstIP string, srcPort, dstPort int, protocol uint8, payload []byte) []byte {
	// This is a simplified packet creation for testing
	// In a real scenario, you would create a proper IP packet

	// Calculate the header size and total packet size
	headerSize := 24 // 20 bytes for IP header + 4 bytes for ports
	packetSize := headerSize + len(payload)

	packet := make([]byte, packetSize)

	// Set source IP (bytes 12-15)
	copy(packet[12:16], net.ParseIP(srcIP).To4())

	// Set destination IP (bytes 16-19)
	copy(packet[16:20], net.ParseIP(dstIP).To4())

	// Set protocol (byte 9)
	packet[9] = protocol

	// Set source port (bytes 20-21)
	binary.BigEndian.PutUint16(packet[20:22], uint16(srcPort))

	// Set destination port (bytes 22-23)
	binary.BigEndian.PutUint16(packet[22:24], uint16(dstPort))

	// Add payload after the header
	if len(payload) > 0 {
		copy(packet[headerSize:], payload)
	}

	return packet
}

// TestDispatcherCreation tests the creation of a new dispatcher
func TestDispatcherCreation(t *testing.T) {
	mockService := NewMockService()
	dispatcher := NewPacketDispatcher(mockService, 300)

	assert.NotNil(t, dispatcher)
	assert.Equal(t, 300, dispatcher.workerIdleTimeout)
	assert.False(t, dispatcher.running)
}

// TestDispatcherStartStop tests starting and stopping the dispatcher
func TestDispatcherStartStop(t *testing.T) {
	mockService := NewMockService()
	dispatcher := NewPacketDispatcher(mockService, 300)

	// Start the dispatcher
	dispatcher.Start()
	assert.True(t, dispatcher.running)

	// Stop the dispatcher
	dispatcher.Stop()
	assert.False(t, dispatcher.running)
}

// TestWorkerCreation tests the creation of a worker
func TestWorkerCreation(t *testing.T) {
	mockService := NewMockService()
	dispatcher := NewPacketDispatcher(mockService, 300)

	// Add a peer mapping
	mockService.peerIDMap["192.168.1.1"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"

	// Create a worker
	ctx := context.Background()
	worker, err := dispatcher.getOrCreateWorker(ctx, "192.168.1.1:80", "192.168.1.1")

	assert.NoError(t, err)
	assert.NotNil(t, worker)
	assert.Equal(t, "192.168.1.1:80", worker.SyncKey)
	assert.Equal(t, "192.168.1.1", worker.DestIP)
	validPeerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")
	assert.Equal(t, validPeerID, worker.PeerID)
	assert.NotNil(t, worker.Stream)
	assert.NotNil(t, worker.PacketChan)
	assert.True(t, worker.Running)
}

// TestWorkerReuse tests that the same worker is reused for the same sync key
func TestWorkerReuse(t *testing.T) {
	mockService := NewMockService()
	dispatcher := NewPacketDispatcher(mockService, 300)

	// Add a peer mapping
	mockService.peerIDMap["192.168.1.1"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"

	// Create a worker
	ctx := context.Background()
	worker1, err := dispatcher.getOrCreateWorker(ctx, "192.168.1.1:80", "192.168.1.1")
	assert.NoError(t, err)

	// Get the same worker again
	worker2, err := dispatcher.getOrCreateWorker(ctx, "192.168.1.1:80", "192.168.1.1")
	assert.NoError(t, err)

	// Verify it's the same worker
	assert.Equal(t, worker1, worker2)
}

// TestPacketDispatch tests dispatching a packet to a worker
func TestPacketDispatch(t *testing.T) {
	mockService := NewMockService()
	dispatcher := NewPacketDispatcher(mockService, 300)

	// Add a peer mapping
	mockService.peerIDMap["192.168.1.1"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Create a test packet
	packet := createTestPacket("10.0.0.1", "192.168.1.1", 12345, 80, 6, []byte("test payload"))

	// Dispatch the packet
	ctx := context.Background()
	dispatcher.DispatchPacket(ctx, "192.168.1.1:80", "192.168.1.1", packet)

	// Verify a worker was created
	_, exists := dispatcher.workers.Load("192.168.1.1:80")
	assert.True(t, exists)

	// Wait for the packet to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify the packet was sent to the stream
	validPeerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")
	stream := mockService.streams[validPeerID]
	assert.Equal(t, packet, stream.buffer.Bytes())

	// Verify metrics were updated
	metrics := mockService.GetMetrics()
	assert.Equal(t, 1, metrics["packets_sent"], "Should have sent 1 packet")
	assert.Equal(t, len(packet), metrics["bytes_sent"], "Should have sent correct number of bytes")
	assert.Equal(t, 0, metrics["packets_dropped"], "Should not have dropped any packets")
	assert.Equal(t, 0, metrics["stream_errors"], "Should not have any stream errors")
}

// MockTUNInterface is a mock implementation of the water.Interface
type MockTUNInterface struct {
	packets [][]byte
	index   int
}

func (m *MockTUNInterface) Read(p []byte) (n int, err error) {
	if m.index >= len(m.packets) {
		// Block until the test is done
		select {}
	}

	packet := m.packets[m.index]
	m.index++

	copy(p, packet)
	return len(packet), nil
}

func (m *MockTUNInterface) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *MockTUNInterface) Close() error {
	return nil
}

func (m *MockTUNInterface) Name() string {
	return "mock-tun"
}
