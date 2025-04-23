package dispatch

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
)

// MockPeerDiscovery is a mock implementation of the PeerDiscoveryService interface
type MockPeerDiscovery struct {
	mock.Mock
}

func (m *MockPeerDiscovery) GetPeerID(ctx context.Context, destIP string) (string, error) {
	args := m.Called(ctx, destIP)
	return args.String(0), args.Error(1)
}

func (m *MockPeerDiscovery) GetPeerIDByRegistry(ctx context.Context, destIP string) (string, error) {
	return m.GetPeerID(ctx, destIP)
}

func (m *MockPeerDiscovery) SyncPeerIDToDHT(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPeerDiscovery) GetVirtualIP(ctx context.Context, peerID string) (string, error) {
	args := m.Called(ctx, peerID)
	return args.String(0), args.Error(1)
}

func (m *MockPeerDiscovery) StoreMappingInDHT(ctx context.Context, peerID string) error {
	args := m.Called(ctx, peerID)
	return args.Error(0)
}

func (m *MockPeerDiscovery) VerifyVirtualIPHasRegistered(ctx context.Context, virtualIP string) error {
	args := m.Called(ctx, virtualIP)
	return args.Error(0)
}

// MockStreamService is a mock implementation of the StreamService interface
type MockStreamService struct {
	mock.Mock
}

func (m *MockStreamService) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (api.VPNStream, error) {
	args := m.Called(ctx, peerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(api.VPNStream), args.Error(1)
}

// MockStream is a mock implementation of the VPNStream interface
type MockStream struct {
	mock.Mock
	closed bool
}

func (m *MockStream) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockStream) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockStream) Close() error {
	args := m.Called()
	m.closed = true
	return args.Error(0)
}

func (m *MockStream) Reset() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStream) SetDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockStream) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockStream) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func TestDispatcher_DispatchPacket(t *testing.T) {
	// Create mocks
	mockPeerDiscovery := new(MockPeerDiscovery)
	mockStreamService := new(MockStreamService)
	mockStream := new(MockStream)

	// Set up mock expectations
	testPeerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	mockPeerDiscovery.On("GetPeerID", mock.Anything, "192.168.1.1").Return(testPeerID.String(), nil)
	mockStreamService.On("CreateNewVPNStream", mock.Anything, testPeerID).Return(mockStream, nil)
	mockStream.On("Write", mock.Anything).Return(10, nil)
	mockStream.On("Close").Return(nil)

	// Create dispatcher config
	config := &DispatcherConfig{
		MaxStreamsPerPeer:     10,
		StreamIdleTimeout:     5 * time.Minute,
		StreamCleanupInterval: 1 * time.Minute,
		WorkerIdleTimeout:     300, // seconds
		WorkerCleanupInterval: 1 * time.Minute,
		WorkerBufferSize:      100,
		PacketBufferSize:      100,
	}

	// Create dispatcher
	dispatcher := NewDispatcher(mockPeerDiscovery, mockStreamService, config, nil)

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Create test packet
	packet := []byte{1, 2, 3, 4, 5}
	connKey := types.ConnectionKey("12345:192.168.1.1:80")

	// Dispatch packet
	err := dispatcher.DispatchPacket(context.Background(), connKey, "192.168.1.1", packet)

	// Assert no error
	assert.NoError(t, err)

	// Wait for packet to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify metrics
	metrics := dispatcher.GetMetrics()
	assert.Equal(t, int64(1), metrics["packets_dispatched"])
}

func TestDispatcher_DispatchPacket_Error(t *testing.T) {
	// Create mocks
	mockPeerDiscovery := new(MockPeerDiscovery)
	mockStreamService := new(MockStreamService)

	// Set up mock expectations for error case
	mockPeerDiscovery.On("GetPeerID", mock.Anything, "192.168.1.2").Return("", errors.New("peer not found"))

	// Create dispatcher config
	config := &DispatcherConfig{
		MaxStreamsPerPeer:     10,
		StreamIdleTimeout:     5 * time.Minute,
		StreamCleanupInterval: 1 * time.Minute,
		WorkerIdleTimeout:     300, // seconds
		WorkerCleanupInterval: 1 * time.Minute,
		WorkerBufferSize:      100,
		PacketBufferSize:      100,
	}

	// Create dispatcher
	dispatcher := NewDispatcher(mockPeerDiscovery, mockStreamService, config, nil)

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Create test packet
	packet := []byte{1, 2, 3, 4, 5}
	connKey := types.ConnectionKey("12345:192.168.1.2:80")

	// Dispatch packet
	err := dispatcher.DispatchPacket(context.Background(), connKey, "192.168.1.2", packet)

	// Assert error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no peer mapping found")

	// Verify metrics
	metrics := dispatcher.GetMetrics()
	assert.Equal(t, int64(0), metrics["packets_dispatched"])
	assert.Equal(t, int64(1), metrics["packets_dropped"])
}

func TestDispatcher_DispatchPacketWithCallback(t *testing.T) {
	// Create mocks
	mockPeerDiscovery := new(MockPeerDiscovery)
	mockStreamService := new(MockStreamService)
	mockStream := new(MockStream)

	// Set up mock expectations
	testPeerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	mockPeerDiscovery.On("GetPeerID", mock.Anything, "192.168.1.1").Return(testPeerID.String(), nil)
	mockStreamService.On("CreateNewVPNStream", mock.Anything, testPeerID).Return(mockStream, nil)
	mockStream.On("Write", mock.Anything).Return(10, nil)
	mockStream.On("Close").Return(nil)

	// Create dispatcher config
	config := &DispatcherConfig{
		MaxStreamsPerPeer:     10,
		StreamIdleTimeout:     5 * time.Minute,
		StreamCleanupInterval: 1 * time.Minute,
		WorkerIdleTimeout:     300, // seconds
		WorkerCleanupInterval: 1 * time.Minute,
		WorkerBufferSize:      100,
		PacketBufferSize:      100,
	}

	// Create dispatcher
	dispatcher := NewDispatcher(mockPeerDiscovery, mockStreamService, config, nil)

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Create test packet
	packet := []byte{1, 2, 3, 4, 5}
	connKey := types.ConnectionKey("12345:192.168.1.1:80")

	// Create callback channel
	doneCh := make(chan error, 1)

	// Dispatch packet with callback
	err := dispatcher.DispatchPacketWithCallback(context.Background(), connKey, "192.168.1.1", packet, doneCh)

	// Assert no error from dispatch
	assert.NoError(t, err)

	// Wait for callback
	callbackErr := <-doneCh

	// Assert no error from callback
	assert.NoError(t, callbackErr)

	// Verify metrics
	metrics := dispatcher.GetMetrics()
	assert.Equal(t, int64(1), metrics["packets_dispatched"])
}
