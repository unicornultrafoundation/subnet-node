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
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/packet/testutil"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

// MockPeerDiscovery is a mock implementation of the api.PeerDiscoveryService interface
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

func (m *MockPeerDiscovery) GetVirtualIP(ctx context.Context, peerID string) (string, error) {
	return "10.0.0.1", nil
}

func (m *MockPeerDiscovery) StoreMappingInDHT(ctx context.Context, peerID string) error {
	return nil
}

func (m *MockPeerDiscovery) VerifyVirtualIPHasRegistered(ctx context.Context, virtualIP string) error {
	return nil
}

func (m *MockPeerDiscovery) GetPeerIDByRegistry(ctx context.Context, destIP string) (string, error) {
	return m.GetPeerID(ctx, destIP)
}

// Ensure MockPeerDiscovery implements the api.PeerDiscoveryService interface
var _ api.PeerDiscoveryService = (*MockPeerDiscovery)(nil)

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

	mockPoolService := &testutil.MockPoolService{}

	// Create a resilience service for testing
	resilienceService := resilience.NewResilienceService(nil)

	dispatcher := NewDispatcher(
		mockPeerDiscovery,
		mockStreamService,
		mockPoolService,
		300,
		5*time.Second,
		100,
		resilienceService,
	)

	assert.NotNil(t, dispatcher)
	// We can't access private fields directly, so we'll just check that the dispatcher is created
}

func TestDispatcherStartStop(t *testing.T) {
	mockPeerDiscovery := NewMockPeerDiscovery()
	mockStreamService := &MockStreamService{}

	mockPoolService := &testutil.MockPoolService{}

	// Create a resilience service for testing
	resilienceService := resilience.NewResilienceService(nil)

	dispatcher := NewDispatcher(
		mockPeerDiscovery,
		mockStreamService,
		mockPoolService,
		300,
		5*time.Second,
		100,
		resilienceService,
	)

	// Test starting the dispatcher
	dispatcher.Start()

	// Test stopping the dispatcher
	dispatcher.Stop()
}

func TestPacketDispatch(t *testing.T) {
	mockPeerDiscovery := NewMockPeerDiscovery()
	mockStreamService := &MockStreamService{}

	mockPoolService := &testutil.MockPoolService{}

	// Create a resilience service for testing
	resilienceService := resilience.NewResilienceService(nil)

	dispatcher := NewDispatcher(
		mockPeerDiscovery,
		mockStreamService,
		mockPoolService,
		300,
		5*time.Second,
		100,
		resilienceService,
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

	// We can't directly access the workers map, but we can test that the dispatch doesn't panic
}
