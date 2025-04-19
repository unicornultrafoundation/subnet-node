package stream_test

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

// MockServiceImpl is a mock implementation of the types.Service interface
type MockServiceImpl struct {
	mock.Mock
}

func (m *MockServiceImpl) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (types.VPNStream, error) {
	args := m.Called(ctx, peerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(types.VPNStream), args.Error(1)
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

func TestStreamService(t *testing.T) {
	// Create a mock stream service
	mockStreamService := new(MockServiceImpl)

	// Create a peer ID
	peerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")

	// Create a mock stream
	mockStream := new(MockStream)

	// Set up the mock service to return the mock stream
	mockStreamService.On("CreateNewVPNStream", mock.Anything, peerID).Return(mockStream, nil)

	// Set up the mock stream
	mockStream.On("Write", mock.Anything).Return(10, nil)
	mockStream.On("Close").Return(nil)

	// Create a stream service
	service := stream.NewStreamService(
		mockStreamService,
		10,
		3,
		5*time.Minute,
		1*time.Minute,
		1*time.Second,
		500*time.Millisecond,
		3,
		1*time.Minute,
		5,
		2,
		1*time.Second,
		true,
	)

	// Start the service
	service.Start()

	// Test CreateNewVPNStream
	s, err := service.CreateNewVPNStream(context.Background(), peerID)
	assert.NoError(t, err)
	assert.Equal(t, mockStream, s)

	// Test GetStream
	s, err = service.GetStream(context.Background(), peerID)
	assert.NoError(t, err)
	assert.Equal(t, mockStream, s)

	// Test ReleaseStream
	service.ReleaseStream(peerID, s, true)

	// Test SendPacketMultiplexed
	packet := []byte("test packet")
	err = service.SendPacketMultiplexed(context.Background(), peerID, packet)
	assert.NoError(t, err)

	// Test GetMultiplexerMetrics
	metrics := service.GetMultiplexerMetrics()
	assert.NotNil(t, metrics)

	// Test GetStreamPoolMetrics
	poolMetrics := service.GetStreamPoolMetrics()
	assert.NotNil(t, poolMetrics)

	// Test GetHealthMetrics
	healthMetrics := service.GetHealthMetrics()
	assert.NotNil(t, healthMetrics)

	// Stop the service
	service.Stop()
}

func TestStreamServiceWithoutMultiplexing(t *testing.T) {
	// Create a mock stream service
	mockStreamService := new(MockServiceImpl)

	// Create a peer ID
	peerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")

	// Create a mock stream
	mockStream := new(MockStream)

	// Set up the mock service to return the mock stream
	mockStreamService.On("CreateNewVPNStream", mock.Anything, peerID).Return(mockStream, nil)

	// Set up the mock stream
	mockStream.On("Write", mock.Anything).Return(10, nil)
	mockStream.On("Close").Return(nil)

	// Create a stream service without multiplexing
	service := stream.NewStreamService(
		mockStreamService,
		10,
		3,
		5*time.Minute,
		1*time.Minute,
		1*time.Second,
		500*time.Millisecond,
		3,
		1*time.Minute,
		5,
		2,
		1*time.Second,
		false,
	)

	// Start the service
	service.Start()

	// Test SendPacketMultiplexed (should fall back to regular stream pool)
	packet := []byte("test packet")
	err := service.SendPacketMultiplexed(context.Background(), peerID, packet)
	assert.NoError(t, err)

	// Test GetMultiplexerMetrics (should return empty map)
	metrics := service.GetMultiplexerMetrics()
	assert.Empty(t, metrics)

	// Stop the service
	service.Stop()
}
