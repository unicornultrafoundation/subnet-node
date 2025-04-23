package pool

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
)

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

func TestStreamPool_GetStreamChannel(t *testing.T) {
	// Create mocks
	mockStreamService := new(MockStreamService)
	mockStream := new(MockStream)

	// Set up mock expectations
	testPeerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	mockStreamService.On("CreateNewVPNStream", mock.Anything, testPeerID).Return(mockStream, nil)
	mockStream.On("Write", mock.Anything).Return(10, nil)
	mockStream.On("Close").Return(nil)

	// Create stream pool config
	config := &StreamPoolConfig{
		MaxStreamsPerPeer: 10,
		StreamIdleTimeout: 5 * time.Minute,
		CleanupInterval:   1 * time.Minute,
		PacketBufferSize:  100,
	}

	// Create stream pool
	streamPool := NewStreamPool(mockStreamService, config)

	// Start the stream pool
	streamPool.Start()
	defer streamPool.Stop()

	// Get a stream channel
	streamChannel, err := streamPool.GetStreamChannel(context.Background(), testPeerID)

	// Assert no error
	assert.NoError(t, err)
	assert.NotNil(t, streamChannel)
	assert.Equal(t, mockStream, streamChannel.Stream)

	// Get another stream channel for the same peer
	streamChannel2, err := streamPool.GetStreamChannel(context.Background(), testPeerID)

	// Assert no error and same stream channel (since it's the least used)
	assert.NoError(t, err)
	assert.Equal(t, streamChannel, streamChannel2)

	// Verify stream count
	assert.Equal(t, 1, streamPool.GetStreamCount(testPeerID))
	assert.Equal(t, 1, streamPool.GetTotalStreamCount())
}

func TestStreamPool_GetStreamChannel_Error(t *testing.T) {
	// Create mocks
	mockStreamService := new(MockStreamService)

	// Set up mock expectations for error case
	testPeerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	mockStreamService.On("CreateNewVPNStream", mock.Anything, testPeerID).Return(nil, errors.New("stream creation failed"))

	// Create stream pool config
	config := &StreamPoolConfig{
		MaxStreamsPerPeer: 10,
		StreamIdleTimeout: 5 * time.Minute,
		CleanupInterval:   1 * time.Minute,
		PacketBufferSize:  100,
	}

	// Create stream pool
	streamPool := NewStreamPool(mockStreamService, config)

	// Start the stream pool
	streamPool.Start()
	defer streamPool.Stop()

	// Try to get a stream channel
	streamChannel, err := streamPool.GetStreamChannel(context.Background(), testPeerID)

	// Assert error
	assert.Error(t, err)
	assert.Nil(t, streamChannel)
	assert.Contains(t, err.Error(), "stream creation failed")

	// Verify stream count
	assert.Equal(t, 0, streamPool.GetStreamCount(testPeerID))
	assert.Equal(t, 0, streamPool.GetTotalStreamCount())
}

func TestStreamPool_ReleaseStreamChannel(t *testing.T) {
	// Create mocks
	mockStreamService := new(MockStreamService)
	mockStream := new(MockStream)

	// Set up mock expectations
	testPeerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	mockStreamService.On("CreateNewVPNStream", mock.Anything, testPeerID).Return(mockStream, nil)
	mockStream.On("Write", mock.Anything).Return(10, nil)
	mockStream.On("Close").Return(nil)

	// Create stream pool config
	config := &StreamPoolConfig{
		MaxStreamsPerPeer: 10,
		StreamIdleTimeout: 5 * time.Minute,
		CleanupInterval:   1 * time.Minute,
		PacketBufferSize:  100,
	}

	// Create stream pool
	streamPool := NewStreamPool(mockStreamService, config)

	// Start the stream pool
	streamPool.Start()
	defer streamPool.Stop()

	// Get a stream channel
	streamChannel, err := streamPool.GetStreamChannel(context.Background(), testPeerID)

	// Assert no error
	assert.NoError(t, err)
	assert.NotNil(t, streamChannel)

	// Release the stream channel as unhealthy
	streamPool.ReleaseStreamChannel(testPeerID, streamChannel, false)

	// Verify stream channel is marked as unhealthy
	streamChannel.mu.RLock()
	healthy := streamChannel.Healthy
	streamChannel.mu.RUnlock()
	assert.False(t, healthy)

	// Get another stream channel for the same peer
	streamChannel2, err := streamPool.GetStreamChannel(context.Background(), testPeerID)

	// Assert no error and different stream channel
	assert.NoError(t, err)
	assert.NotEqual(t, streamChannel, streamChannel2)
}

func TestStreamPool_CleanupIdleStreams(t *testing.T) {
	// Create mocks
	mockStreamService := new(MockStreamService)
	mockStream := new(MockStream)

	// Set up mock expectations
	testPeerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	mockStreamService.On("CreateNewVPNStream", mock.Anything, testPeerID).Return(mockStream, nil)
	mockStream.On("Write", mock.Anything).Return(10, nil)
	mockStream.On("Close").Return(nil)

	// Create stream pool config with short timeout for testing
	config := &StreamPoolConfig{
		MaxStreamsPerPeer: 10,
		StreamIdleTimeout: 100 * time.Millisecond,
		CleanupInterval:   50 * time.Millisecond,
		PacketBufferSize:  100,
	}

	// Create stream pool
	streamPool := NewStreamPool(mockStreamService, config)

	// Start the stream pool
	streamPool.Start()
	defer streamPool.Stop()

	// Get a stream channel
	streamChannel, err := streamPool.GetStreamChannel(context.Background(), testPeerID)

	// Assert no error
	assert.NoError(t, err)
	assert.NotNil(t, streamChannel)

	// Verify stream count
	assert.Equal(t, 1, streamPool.GetStreamCount(testPeerID))

	// Wait for cleanup to run
	time.Sleep(200 * time.Millisecond)

	// Verify stream was removed
	assert.Equal(t, 0, streamPool.GetStreamCount(testPeerID))
}
