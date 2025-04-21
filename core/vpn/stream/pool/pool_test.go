package pool_test

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/pool"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/testutil"
)

func TestStreamPool(t *testing.T) {
	// Create a mock stream service
	mockService := new(testutil.MockStreamService)

	// Create a mock stream
	mockStream := new(testutil.MockStream)

	// Create a peer ID
	peerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")

	// Set up the mock service to return the mock stream
	mockService.On("CreateNewVPNStream", mock.Anything, peerID).Return(mockStream, nil)

	// Set up the mock stream to handle Close
	mockStream.On("Close").Return(nil)

	// Create a stream pool
	streamPool := pool.NewStreamPool(
		mockService,
		10,
		3,
		5*time.Minute,
	)

	// Test getting a stream
	stream, err := streamPool.GetStream(context.Background(), peerID)
	assert.NoError(t, err)
	assert.Equal(t, mockStream, stream)

	// Test releasing a stream
	streamPool.ReleaseStream(peerID, stream, true)

	// Test getting stream count
	count := streamPool.GetStreamCount(peerID)
	assert.Equal(t, 1, count)

	// Test getting active stream count
	activeCount := streamPool.GetActiveStreamCount(peerID)
	assert.Equal(t, 0, activeCount)

	// Test getting metrics
	metrics := streamPool.GetMetrics()
	assert.NotNil(t, metrics)

	// Test closing the pool
	streamPool.Close()
}

func TestStreamPoolManager(t *testing.T) {
	// Create a mock stream service
	mockService := new(testutil.MockStreamService)

	// Create mock streams
	mockStream1 := new(testutil.MockStream)
	mockStream2 := new(testutil.MockStream)

	// Create peer IDs
	peerID1, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	peerID2, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5M")

	// Set up the mock service to return different streams for different peers
	mockService.On("CreateNewVPNStream", mock.Anything, peerID1).Return(mockStream1, nil)
	mockService.On("CreateNewVPNStream", mock.Anything, peerID2).Return(mockStream2, nil)

	// Set up the mock streams to handle Close
	mockStream1.On("Close").Return(nil)
	mockStream2.On("Close").Return(nil)

	// Create a stream pool manager
	manager := pool.NewStreamPoolManager(
		mockService,
		10,
		3,
		5*time.Minute,
		1*time.Minute,
	)

	// Test getting streams for different peers
	stream1, err := manager.GetStream(context.Background(), peerID1)
	assert.NoError(t, err)
	assert.Equal(t, mockStream1, stream1)

	stream2, err := manager.GetStream(context.Background(), peerID2)
	assert.NoError(t, err)
	assert.Equal(t, mockStream2, stream2)

	// Test releasing streams
	manager.ReleaseStream(peerID1, stream1, true)
	manager.ReleaseStream(peerID2, stream2, true)

	// Test getting stream count
	count1 := manager.GetStreamCount(peerID1)
	assert.Equal(t, 1, count1)

	count2 := manager.GetStreamCount(peerID2)
	assert.Equal(t, 1, count2)

	// Test getting active stream count
	activeCount1 := manager.GetActiveStreamCount(peerID1)
	assert.Equal(t, 0, activeCount1)

	activeCount2 := manager.GetActiveStreamCount(peerID2)
	assert.Equal(t, 0, activeCount2)

	// Test getting metrics
	metrics := manager.GetStreamPoolMetrics()
	assert.NotNil(t, metrics)

	// Test getting min streams per peer
	minStreams := manager.GetMinStreamsPerPeer()
	assert.Equal(t, 3, minStreams)

	// Start the manager
	manager.Start()

	// Stop the manager
	manager.Stop()
}
