package pool

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
)

// MockStreamPoolV2 is a mock implementation of the StreamPoolInterfaceV2
type MockStreamPoolV2 struct {
	mock.Mock
}

func (m *MockStreamPoolV2) GetStreamChannel(ctx context.Context, peerID peer.ID) (*StreamChannelV2, error) {
	args := m.Called(ctx, peerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*StreamChannelV2), args.Error(1)
}

func (m *MockStreamPoolV2) ReleaseStreamChannel(peerID peer.ID, streamChannel *StreamChannelV2, healthy bool) {
	m.Called(peerID, streamChannel, healthy)
}

func (m *MockStreamPoolV2) Start() {
	m.Called()
}

func (m *MockStreamPoolV2) Stop() {
	m.Called()
}

func (m *MockStreamPoolV2) GetStreamMetrics() map[string]map[string]int64 {
	args := m.Called()
	return args.Get(0).(map[string]map[string]int64)
}

func (m *MockStreamPoolV2) GetStreamCount(peerID peer.ID) int {
	args := m.Called(peerID)
	return args.Int(0)
}

func (m *MockStreamPoolV2) GetTotalStreamCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockStreamPoolV2) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStreamPoolV2) GetPacketBufferSize() int {
	args := m.Called()
	return args.Int(0)
}

func TestStreamManagerV2_GetOrCreateStreamForConnection(t *testing.T) {
	// Create mocks
	mockStreamPool := new(MockStreamPoolV2)
	mockStream := new(MockStream)

	// Create a stream channel
	streamCtx, streamCancel := context.WithCancel(context.Background())
	streamChannel := &StreamChannelV2{
		Stream:       mockStream,
		PacketChan:   make(chan *types.QueuedPacket, 10),
		lastActivity: time.Now().UnixNano(),
		healthy:      1, // 1 = healthy
		ctx:          streamCtx,
		cancel:       streamCancel,
	}

	// Set up mock expectations
	testPeerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	mockStreamPool.On("GetStreamChannel", mock.Anything, testPeerID).Return(streamChannel, nil)
	mockStreamPool.On("ReleaseStreamChannel", mock.Anything, mock.Anything, mock.Anything).Return()
	mockStreamPool.On("Start").Return()
	mockStreamPool.On("Stop").Return()
	mockStreamPool.On("GetStreamMetrics").Return(map[string]map[string]int64{})
	mockStreamPool.On("GetStreamCount", mock.Anything).Return(1)
	mockStreamPool.On("GetTotalStreamCount").Return(1)
	mockStreamPool.On("Close").Return(nil)
	mockStreamPool.On("GetPacketBufferSize").Return(100)
	mockStream.On("Write", mock.Anything).Return(10, nil)
	mockStream.On("Close").Return(nil)

	// Create stream manager
	streamManager := NewStreamManagerV2(mockStreamPool)

	// Start the stream manager
	streamManager.Start()
	defer streamManager.Stop()

	// Get a stream for a connection
	connKey := types.ConnectionKey("12345:192.168.1.1:80")
	resultChannel, err := streamManager.GetOrCreateStreamForConnection(context.Background(), connKey, testPeerID)

	// Assert no error
	assert.NoError(t, err)
	assert.Equal(t, streamChannel, resultChannel)

	// Get the same stream again
	resultChannel2, err := streamManager.GetOrCreateStreamForConnection(context.Background(), connKey, testPeerID)

	// Assert no error and same stream
	assert.NoError(t, err)
	assert.Equal(t, streamChannel, resultChannel2)

	// Verify metrics
	metrics := streamManager.GetMetrics()
	assert.Equal(t, int64(1), metrics["streams_created"])

	// Verify connection count
	assert.Equal(t, 1, streamManager.GetConnectionCount())
}

func TestStreamManagerV2_SendPacket(t *testing.T) {
	// Create mocks
	mockStreamPool := new(MockStreamPoolV2)
	mockStream := new(MockStream)

	// Create a stream channel
	streamCtx, streamCancel := context.WithCancel(context.Background())
	streamChannel := &StreamChannelV2{
		Stream:       mockStream,
		PacketChan:   make(chan *types.QueuedPacket, 10),
		lastActivity: time.Now().UnixNano(),
		healthy:      1, // 1 = healthy
		ctx:          streamCtx,
		cancel:       streamCancel,
	}

	// Start a goroutine to read from the packet channel
	go func() {
		for packet := range streamChannel.PacketChan {
			// Simulate processing
			if packet.DoneCh != nil {
				packet.DoneCh <- nil
				close(packet.DoneCh)
			}
		}
	}()

	// Set up mock expectations
	testPeerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	mockStreamPool.On("GetStreamChannel", mock.Anything, testPeerID).Return(streamChannel, nil)
	mockStreamPool.On("ReleaseStreamChannel", mock.Anything, mock.Anything, mock.Anything).Return()
	mockStreamPool.On("Start").Return()
	mockStreamPool.On("Stop").Return()
	mockStreamPool.On("GetStreamMetrics").Return(map[string]map[string]int64{})
	mockStreamPool.On("GetStreamCount", mock.Anything).Return(1)
	mockStreamPool.On("GetTotalStreamCount").Return(1)
	mockStreamPool.On("Close").Return(nil)
	mockStreamPool.On("GetPacketBufferSize").Return(100)
	mockStream.On("Write", mock.Anything).Return(10, nil)
	mockStream.On("Close").Return(nil)

	// Create stream manager
	streamManager := NewStreamManagerV2(mockStreamPool)

	// Start the stream manager
	streamManager.Start()
	defer streamManager.Stop()

	// Create a packet
	packet := &types.QueuedPacket{
		Ctx:    context.Background(),
		DestIP: "192.168.1.1",
		Data:   []byte{1, 2, 3, 4, 5},
		DoneCh: make(chan error, 1),
	}

	// Send the packet
	connKey := types.ConnectionKey("12345:192.168.1.1:80")
	err := streamManager.SendPacket(context.Background(), connKey, testPeerID, packet)

	// Assert no error
	assert.NoError(t, err)

	// Wait for packet to be processed
	callbackErr := <-packet.DoneCh

	// Assert no error from callback
	assert.NoError(t, callbackErr)

	// Verify metrics
	metrics := streamManager.GetMetrics()
	assert.Equal(t, int64(1), metrics["packets_processed"])
}

func TestStreamManagerV2_ReleaseConnection(t *testing.T) {
	// Create mocks
	mockStreamPool := new(MockStreamPoolV2)
	mockStream := new(MockStream)

	// Create a stream channel
	streamCtx, streamCancel := context.WithCancel(context.Background())
	streamChannel := &StreamChannelV2{
		Stream:       mockStream,
		PacketChan:   make(chan *types.QueuedPacket, 10),
		lastActivity: time.Now().UnixNano(),
		healthy:      1, // 1 = healthy
		ctx:          streamCtx,
		cancel:       streamCancel,
	}

	// Set up mock expectations
	testPeerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	mockStreamPool.On("GetStreamChannel", mock.Anything, testPeerID).Return(streamChannel, nil)
	mockStreamPool.On("ReleaseStreamChannel", testPeerID, streamChannel, false).Return()
	mockStreamPool.On("Start").Return()
	mockStreamPool.On("Stop").Return()
	mockStreamPool.On("GetStreamMetrics").Return(map[string]map[string]int64{})
	mockStreamPool.On("GetStreamCount", mock.Anything).Return(1)
	mockStreamPool.On("GetTotalStreamCount").Return(1)
	mockStreamPool.On("Close").Return(nil)
	mockStreamPool.On("GetPacketBufferSize").Return(100)
	mockStream.On("Write", mock.Anything).Return(10, nil)
	mockStream.On("Close").Return(nil)

	// Create stream manager
	streamManager := NewStreamManagerV2(mockStreamPool)

	// Start the stream manager
	streamManager.Start()
	defer streamManager.Stop()

	// Get a stream for a connection
	connKey := types.ConnectionKey("12345:192.168.1.1:80")
	_, err := streamManager.GetOrCreateStreamForConnection(context.Background(), connKey, testPeerID)

	// Assert no error
	assert.NoError(t, err)

	// Release the connection as unhealthy
	streamManager.ReleaseConnection(connKey, false)

	// Verify the connection was removed
	assert.Equal(t, 0, streamManager.GetConnectionCount())

	// Verify metrics
	metrics := streamManager.GetMetrics()
	assert.Equal(t, int64(1), metrics["streams_released"])
}
