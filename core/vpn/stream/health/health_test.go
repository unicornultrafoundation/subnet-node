package health_test

import (
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/health"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/pool"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/testutil"
)

func TestHealthChecker(t *testing.T) {
	// Create a mock stream service
	mockService := new(testutil.MockStreamService)

	// Create a real pool manager
	poolManager := pool.NewStreamPoolManager(
		mockService,
		3,
		5*time.Minute,
		1*time.Minute,
	)

	// Create a health checker
	healthChecker := health.NewHealthChecker(
		poolManager,
		100*time.Millisecond,
		50*time.Millisecond,
		3,
	)

	// Start the health checker
	healthChecker.Start()

	// Wait for some health checks to be performed
	time.Sleep(250 * time.Millisecond)

	// Stop the health checker
	healthChecker.Stop()

	// Check metrics
	metrics := healthChecker.GetMetrics()
	assert.Greater(t, metrics["checks_performed"], int64(0))
}

func TestStreamWarmer(t *testing.T) {
	// Create a mock stream service
	mockService := new(testutil.MockStreamService)

	// Create a peer ID
	peerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")

	// Create a mock stream
	mockStream := new(testutil.MockStream)

	// Set up the mock service to return the mock stream
	mockService.On("CreateNewVPNStream", mock.Anything, peerID).Return(mockStream, nil)

	// Create a real pool manager
	poolManager := pool.NewStreamPoolManager(
		mockService,
		3,
		5*time.Minute,
		1*time.Minute,
	)

	// Create a stream warmer
	warmer := health.NewStreamWarmer(
		poolManager,
		mockService,
		100*time.Millisecond,
		3,
	)

	// Add a peer to warm
	warmer.AddPeer(peerID)

	// Start the warmer
	warmer.Start()

	// Wait for some warming to be performed
	time.Sleep(250 * time.Millisecond)

	// Stop the warmer
	warmer.Stop()

	// Check metrics
	metrics := warmer.GetMetrics()
	assert.Greater(t, metrics["warms_performed"], int64(0))
}
