package router

import (
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Tests
func TestStreamRouter_GetOrCreateRoute(t *testing.T) {
	// Create mocks
	mockPool := new(MockPoolServiceExt)
	mockPeerDiscovery := new(MockPeerDiscoveryService)

	// Create test config with short TTL
	config := &StreamRouterConfig{
		MinStreamsPerPeer:        1,
		MaxStreamsPerPeer:        5,
		ThroughputThreshold:      1000,
		ScaleUpThreshold:         0.8,
		ScaleDownThreshold:       0.3,
		ScalingInterval:          100 * time.Millisecond,
		MinWorkers:               2,
		MaxWorkers:               4,
		InitialWorkers:           2,
		WorkerQueueSize:          10,
		WorkerScaleInterval:      100 * time.Millisecond,
		WorkerScaleUpThreshold:   0.75,
		WorkerScaleDownThreshold: 0.25,
		ConnectionTTL:            100 * time.Millisecond, // Very short for testing
		CleanupInterval:          50 * time.Millisecond,  // Very short for testing
		CacheShardCount:          4,
	}

	// Setup mock expectations
	testPeerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")

	// Create router
	router := NewStreamRouter(config, mockPool, mockPeerDiscovery)
	defer router.Shutdown()

	// Test getOrCreateRoute
	connKey := "10.0.0.1:12345:192.168.1.1:80"
	route, err := router.getOrCreateRoute(connKey, testPeerID)

	// Verify route was created
	assert.NoError(t, err)
	assert.NotNil(t, route)
	assert.Equal(t, connKey, route.connKey)
	assert.Equal(t, testPeerID, route.peerID)
	assert.Equal(t, 0, route.streamIndex)

	// Verify cache has the route
	cachedRoute, found := router.connectionCache.Get(connKey)
	assert.True(t, found)
	assert.Equal(t, route, cachedRoute)

	// Test connection cache expiration
	time.Sleep(200 * time.Millisecond) // Wait for cache to expire

	// Check that connection was removed from cache
	_, found = router.connectionCache.Get(connKey)
	assert.False(t, found, "Connection should be removed from cache after expiration")

	// Verify expectations
	mockPool.AssertExpectations(t)
}

func TestStreamRouter_GetStreamForRoute(t *testing.T) {
	// Create mocks
	mockPool := new(MockPoolServiceExt)
	mockPeerDiscovery := new(MockPeerDiscoveryService)
	mockStream := new(MockVPNStream)

	// Create test config
	config := DefaultStreamRouterConfig()

	// Setup mock expectations
	testPeerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	mockStream.On("Read", mock.Anything).Return(0, nil)
	mockStream.On("Reset").Return(nil)
	mockStream.On("SetDeadline", mock.Anything).Return(nil)
	mockStream.On("SetReadDeadline", mock.Anything).Return(nil)
	mockStream.On("SetWriteDeadline", mock.Anything).Return(nil)
	mockPool.On("GetStreamByIndex", mock.Anything, testPeerID, 0).Return(mockStream, nil)

	// Create router
	router := NewStreamRouter(config, mockPool, mockPeerDiscovery)
	defer router.Shutdown()

	// Create a test route
	route := &ConnectionRoute{
		connKey:      "10.0.0.1:12345:192.168.1.1:80",
		peerID:       testPeerID,
		streamIndex:  0,
		lastActivity: time.Now(),
		packetCount:  0,
		workerID:     0,
	}

	// Get stream for route
	stream, err := router.getStreamForRoute(route)

	// Verify stream was returned
	assert.NoError(t, err)
	assert.Equal(t, mockStream, stream)

	// Verify expectations
	mockPool.AssertExpectations(t)
}
