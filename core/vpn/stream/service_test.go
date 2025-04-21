package stream_test

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/testutil"
)

func TestStreamService(t *testing.T) {
	// Create a mock stream service
	mockStreamService := new(testutil.MockStreamService)

	// Create a peer ID
	peerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")

	// Create a mock stream
	mockStream := new(testutil.MockStream)

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

	// Test GetStreamPoolMetrics
	poolMetrics := service.GetStreamPoolMetrics()
	assert.NotNil(t, poolMetrics)

	// Test GetHealthMetrics
	healthMetrics := service.GetHealthMetrics()
	assert.NotNil(t, healthMetrics)

	// Stop the service
	service.Stop()
}
