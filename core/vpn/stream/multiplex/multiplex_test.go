package multiplex_test

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/multiplex"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

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

// MockPoolService is a mock implementation of the types.PoolService interface
type MockPoolService struct {
	mock.Mock
}

func (m *MockPoolService) GetStream(ctx context.Context, peerID peer.ID) (types.VPNStream, error) {
	args := m.Called(ctx, peerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(types.VPNStream), args.Error(1)
}

func (m *MockPoolService) ReleaseStream(peerID peer.ID, s types.VPNStream, healthy bool) {
	m.Called(peerID, s, healthy)
}

// MockStream is a mock implementation of the types.VPNStream interface
type MockStream struct {
	mock.Mock
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

func TestStreamMultiplexer(t *testing.T) {
	// Create a mock stream service
	mockStreamService := new(MockStreamService)

	// Create a mock pool service
	mockPoolService := new(MockPoolService)

	// Create a peer ID
	peerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")

	// Create mock streams
	mockStream1 := new(MockStream)
	mockStream2 := new(MockStream)
	mockStream3 := new(MockStream)

	// Set up the mock service to return the mock streams
	mockStreamService.On("CreateNewVPNStream", mock.Anything, peerID).Return(mockStream1, nil).Once()
	mockStreamService.On("CreateNewVPNStream", mock.Anything, peerID).Return(mockStream2, nil).Once()
	mockStreamService.On("CreateNewVPNStream", mock.Anything, peerID).Return(mockStream3, nil).Once()

	// Set up the mock streams
	mockStream1.On("Write", mock.Anything).Return(10, nil)
	mockStream2.On("Write", mock.Anything).Return(10, nil)
	mockStream3.On("Write", mock.Anything).Return(10, nil)
	mockStream1.On("Close").Return(nil)
	mockStream2.On("Close").Return(nil)
	mockStream3.On("Close").Return(nil)

	// Create a stream multiplexer
	multiplexer := multiplex.NewStreamMultiplexer(
		peerID,
		mockStreamService,
		mockPoolService,
		5,
		2,
	)

	// Initialize the multiplexer
	err := multiplexer.Initialize(context.Background())
	assert.NoError(t, err)

	// Test sending packets
	packet := []byte("test packet")
	for i := 0; i < 5; i++ {
		err = multiplexer.SendPacket(context.Background(), packet)
		assert.NoError(t, err)
	}

	// Check metrics
	metrics := multiplexer.GetMetrics()
	assert.Equal(t, int64(5), metrics.PacketsSent)
	assert.Equal(t, int64(55), metrics.BytesSent) // 5 packets * 11 bytes each (including overhead)
	assert.Equal(t, int64(2), metrics.StreamsCreated)

	// Test scaling up
	err = multiplexer.ScaleUp(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 3, multiplexer.GetStreamCount())

	// Test scaling down
	err = multiplexer.ScaleDown()
	assert.NoError(t, err)
	assert.Equal(t, 2, multiplexer.GetStreamCount())

	// Test closing the multiplexer
	multiplexer.Close()
}

func TestMultiplexerManager(t *testing.T) {
	// Create a mock stream service
	mockStreamService := new(MockStreamService)

	// Create a mock pool service
	mockPoolService := new(MockPoolService)

	// Create a peer ID
	peerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")

	// Create mock streams
	mockStream1 := new(MockStream)
	mockStream2 := new(MockStream)

	// Set up the mock service to return the mock streams
	mockStreamService.On("CreateNewVPNStream", mock.Anything, peerID).Return(mockStream1, nil).Once()
	mockStreamService.On("CreateNewVPNStream", mock.Anything, peerID).Return(mockStream2, nil).Once()

	// Set up the mock streams
	mockStream1.On("Write", mock.Anything).Return(10, nil)
	mockStream2.On("Write", mock.Anything).Return(10, nil)
	mockStream1.On("Close").Return(nil)
	mockStream2.On("Close").Return(nil)

	// Create a multiplexer manager
	manager := multiplex.NewMultiplexerManager(
		mockStreamService,
		mockPoolService,
		5,
		2,
		1*time.Second,
	)

	// Start the manager
	manager.Start()

	// Test sending packets
	packet := []byte("test packet")
	err := manager.SendPacket(context.Background(), peerID, packet)
	assert.NoError(t, err)

	// Check metrics
	metrics := manager.GetMultiplexerMetrics()
	assert.NotNil(t, metrics)

	managerMetrics := manager.GetManagerMetrics()
	assert.NotNil(t, managerMetrics)

	// Stop the manager
	manager.Stop()
}
