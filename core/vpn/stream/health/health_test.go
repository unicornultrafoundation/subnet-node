package health_test

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/health"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/pool"
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

// MockStream is a mock implementation of the types.VPNStream interface
type MockStream struct {
	mock.Mock
	healthy bool
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

// MockPoolManager is a mock implementation of the pool.StreamPoolManager
type MockPoolManager struct {
	mock.Mock
	minStreamsPerPeer int
}

func (m *MockPoolManager) GetStream(ctx context.Context, peerID peer.ID) (types.VPNStream, error) {
	args := m.Called(ctx, peerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(types.VPNStream), args.Error(1)
}

func (m *MockPoolManager) ReleaseStream(peerID peer.ID, s types.VPNStream, healthy bool) {
	m.Called(peerID, s, healthy)
}

func (m *MockPoolManager) GetMetrics() map[string]int64 {
	args := m.Called()
	return args.Get(0).(map[string]int64)
}

func (m *MockPoolManager) GetStreamCount(peerID peer.ID) int {
	args := m.Called(peerID)
	return args.Int(0)
}

func (m *MockPoolManager) GetActiveStreamCount(peerID peer.ID) int {
	args := m.Called(peerID)
	return args.Int(0)
}

func (m *MockPoolManager) GetMinStreamsPerPeer() int {
	return m.minStreamsPerPeer
}

func (m *MockPoolManager) Start() {
	m.Called()
}

func (m *MockPoolManager) Stop() {
	m.Called()
}

func TestHealthChecker(t *testing.T) {
	// Create a mock stream service
	mockService := new(MockStreamService)

	// Create a real pool manager
	poolManager := pool.NewStreamPoolManager(
		mockService,
		10,
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
	mockService := new(MockStreamService)

	// Create a peer ID
	peerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")

	// Create a mock stream
	mockStream := new(MockStream)

	// Set up the mock service to return the mock stream
	mockService.On("CreateNewVPNStream", mock.Anything, peerID).Return(mockStream, nil)

	// Create a real pool manager
	poolManager := pool.NewStreamPoolManager(
		mockService,
		10,
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

func TestHealthIntegration(t *testing.T) {
	// Create a mock stream service
	mockService := new(MockStreamService)

	// Create a peer ID
	peerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")

	// Create mock streams
	healthyStream := &MockStream{healthy: true}
	unhealthyStream := &MockStream{healthy: false}

	// Set up the mock service to return streams
	mockService.On("CreateNewVPNStream", mock.Anything, peerID).Return(healthyStream, nil)

	// Set up the healthy stream to pass health checks
	healthyStream.On("SetDeadline", mock.Anything).Return(nil)
	healthyStream.On("Write", mock.Anything).Return(1, nil)
	healthyStream.On("Close").Return(nil)

	// Set up the unhealthy stream to fail health checks
	unhealthyStream.On("SetDeadline", mock.Anything).Return(nil)
	unhealthyStream.On("Write", mock.Anything).Return(0, assert.AnError)
	unhealthyStream.On("Close").Return(nil)

	// Create a stream pool manager
	poolManager := pool.NewStreamPoolManager(
		mockService,
		10,
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

	// Create a stream warmer
	warmer := health.NewStreamWarmer(
		poolManager,
		mockService,
		100*time.Millisecond,
		3,
	)

	// Start the health checker and warmer
	healthChecker.Start()
	warmer.Start()

	// Wait for some checks and warming to be performed
	time.Sleep(250 * time.Millisecond)

	// Stop the health checker and warmer
	healthChecker.Stop()
	warmer.Stop()

	// Check metrics
	healthMetrics := healthChecker.GetMetrics()
	assert.Greater(t, healthMetrics["checks_performed"], int64(0))

	warmerMetrics := warmer.GetMetrics()
	assert.Greater(t, warmerMetrics["warms_performed"], int64(0))
}
