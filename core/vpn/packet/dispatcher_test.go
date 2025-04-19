package packet

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

// MockPeerDiscovery is a mock implementation of the PeerDiscoveryService interface
type MockPeerDiscovery struct {
	mock.Mock
	peerIDMap map[string]string
}

func NewMockPeerDiscovery() *MockPeerDiscovery {
	return &MockPeerDiscovery{
		peerIDMap: make(map[string]string),
	}
}

func (m *MockPeerDiscovery) GetPeerID(ctx context.Context, destIP string) (string, error) {
	peerID, exists := m.peerIDMap[destIP]
	if !exists {
		return "", errors.New("peer not found")
	}
	return peerID, nil
}

func (m *MockPeerDiscovery) SyncPeerIDToDHT(ctx context.Context) error {
	return nil
}

// MockStreamService is a mock implementation of the StreamService interface
type MockStreamService struct {
	mock.Mock
}

func (m *MockStreamService) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (types.VPNStream, error) {
	return &MockStream{}, nil
}

// MockStream is a mock implementation of the VPNStream interface
type MockStream struct {
	mock.Mock
	closed bool
	mu     sync.Mutex
}

func (m *MockStream) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (m *MockStream) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *MockStream) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *MockStream) Reset() error {
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

func TestDispatcherCreation(t *testing.T) {
	mockPeerDiscovery := NewMockPeerDiscovery()
	mockStreamService := &MockStreamService{}

	dispatcher := NewDispatcher(
		mockPeerDiscovery,
		mockStreamService,
		300,
		5*time.Second,
		100,
	)

	assert.NotNil(t, dispatcher)
	assert.Equal(t, 300, dispatcher.workerIdleTimeout)
	assert.Equal(t, 5*time.Second, dispatcher.workerCleanupInterval)
	assert.Equal(t, 100, dispatcher.workerBufferSize)
	assert.False(t, dispatcher.running)
}

func TestDispatcherStartStop(t *testing.T) {
	mockPeerDiscovery := NewMockPeerDiscovery()
	mockStreamService := &MockStreamService{}

	dispatcher := NewDispatcher(
		mockPeerDiscovery,
		mockStreamService,
		300,
		5*time.Second,
		100,
	)

	// Test starting the dispatcher
	dispatcher.Start()
	assert.True(t, dispatcher.running)

	// Test stopping the dispatcher
	dispatcher.Stop()
	assert.False(t, dispatcher.running)
}

func TestWorkerCreation(t *testing.T) {
	mockPeerDiscovery := NewMockPeerDiscovery()
	mockStreamService := &MockStreamService{}

	dispatcher := NewDispatcher(
		mockPeerDiscovery,
		mockStreamService,
		300,
		5*time.Second,
		100,
	)

	// Add a peer mapping
	mockPeerDiscovery.peerIDMap["192.168.1.1"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"

	// Create a worker
	ctx := context.Background()
	worker, err := dispatcher.getOrCreateWorker(ctx, "192.168.1.1:80", "192.168.1.1")

	assert.NoError(t, err)
	assert.NotNil(t, worker)
	assert.Equal(t, "192.168.1.1:80", worker.SyncKey)
	assert.Equal(t, "192.168.1.1", worker.DestIP)
}

func TestWorkerReuse(t *testing.T) {
	mockPeerDiscovery := NewMockPeerDiscovery()
	mockStreamService := &MockStreamService{}

	dispatcher := NewDispatcher(
		mockPeerDiscovery,
		mockStreamService,
		300,
		5*time.Second,
		100,
	)

	// Add a peer mapping
	mockPeerDiscovery.peerIDMap["192.168.1.1"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"

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

func TestPacketDispatch(t *testing.T) {
	mockPeerDiscovery := NewMockPeerDiscovery()
	mockStreamService := &MockStreamService{}

	dispatcher := NewDispatcher(
		mockPeerDiscovery,
		mockStreamService,
		300,
		5*time.Second,
		100,
	)

	// Add a peer mapping
	mockPeerDiscovery.peerIDMap["192.168.1.1"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Create a test packet
	packet := []byte("test packet")

	// Dispatch the packet
	ctx := context.Background()
	dispatcher.DispatchPacket(ctx, "192.168.1.1:80", "192.168.1.1", packet)

	// Verify a worker was created
	_, exists := dispatcher.workers.Load("192.168.1.1:80")
	assert.True(t, exists)
}

func TestCleanupInactiveWorkers(t *testing.T) {
	mockPeerDiscovery := NewMockPeerDiscovery()
	mockStreamService := &MockStreamService{}

	// Use a short cleanup interval for testing
	dispatcher := NewDispatcher(
		mockPeerDiscovery,
		mockStreamService,
		1,                    // 1 second idle timeout
		100*time.Millisecond, // 100ms cleanup interval
		100,
	)

	// Add a peer mapping
	mockPeerDiscovery.peerIDMap["192.168.1.1"] = "12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q"

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Create a worker
	ctx := context.Background()
	_, err := dispatcher.getOrCreateWorker(ctx, "192.168.1.1:80", "192.168.1.1")
	assert.NoError(t, err)

	// Verify the worker exists
	_, exists := dispatcher.workers.Load("192.168.1.1:80")
	assert.True(t, exists)

	// Wait for the worker to be cleaned up
	time.Sleep(1500 * time.Millisecond)

	// Verify the worker was removed
	_, exists = dispatcher.workers.Load("192.168.1.1:80")
	assert.False(t, exists)
}
