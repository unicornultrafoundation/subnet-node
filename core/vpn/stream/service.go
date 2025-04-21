package stream

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/health"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/pool"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

var log = logrus.WithField("service", "vpn-stream")

// StreamService implements the Service, PoolService, HealthService, MetricsService,
// and LifecycleService interfaces. It provides a unified interface for managing streams,
// including stream pooling, health checking, and metrics collection.
type StreamService struct {
	// Stream service for creating new streams
	streamService types.Service
	// Stream pool manager
	poolManager *pool.StreamPoolManager
	// Health checker
	healthChecker *health.HealthChecker
	// Stream warmer
	streamWarmer *health.StreamWarmer
	// Mutex to protect access to the service
	mu sync.Mutex
}

// NewStreamService creates a new stream service
func NewStreamService(
	streamService types.Service,
	minStreamsPerPeer int,
	streamIdleTimeout time.Duration,
	cleanupInterval time.Duration,
	healthCheckInterval time.Duration,
	healthCheckTimeout time.Duration,
	maxConsecutiveFailures int,
	warmInterval time.Duration,
) *StreamService {
	// Create the stream pool manager
	poolManager := pool.NewStreamPoolManager(
		streamService,
		minStreamsPerPeer,
		streamIdleTimeout,
		cleanupInterval,
	)

	// Create the health checker
	healthChecker := health.NewHealthChecker(
		poolManager,
		healthCheckInterval,
		healthCheckTimeout,
		maxConsecutiveFailures,
	)

	// Create the stream warmer
	streamWarmer := health.NewStreamWarmer(
		poolManager,
		streamService,
		warmInterval,
		minStreamsPerPeer,
	)

	return &StreamService{
		streamService: streamService,
		poolManager:   poolManager,
		healthChecker: healthChecker,
		streamWarmer:  streamWarmer,
	}
}

// Start starts the stream service
func (s *StreamService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Start the stream pool manager
	s.poolManager.Start()

	// Start the health checker
	s.healthChecker.Start()

	// Start the stream warmer
	s.streamWarmer.Start()

	log.Info("Stream service started")
}

// Stop stops the stream service
func (s *StreamService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop the stream warmer
	s.streamWarmer.Stop()

	// Stop the health checker
	s.healthChecker.Stop()

	// Stop the stream pool manager
	s.poolManager.Stop()

	log.Info("Stream service stopped")
}

// Close implements the io.Closer interface
func (s *StreamService) Close() error {
	s.Stop()
	return nil
}

// CreateNewVPNStream creates a new stream to a peer
func (s *StreamService) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (types.VPNStream, error) {
	return s.streamService.CreateNewVPNStream(ctx, peerID)
}

// GetStream gets a stream from the pool or creates a new one
func (s *StreamService) GetStream(ctx context.Context, peerID peer.ID) (types.VPNStream, error) {
	// Add the peer to the stream warmer
	s.streamWarmer.AddPeer(peerID)

	return s.poolManager.GetStream(ctx, peerID)
}

// ReleaseStream returns a stream to the pool
func (s *StreamService) ReleaseStream(peerID peer.ID, stream types.VPNStream, healthy bool) {
	s.poolManager.ReleaseStream(peerID, stream, healthy)
}

// StartHealthChecker starts the health checker
func (s *StreamService) StartHealthChecker() {
	s.healthChecker.Start()
}

// StopHealthChecker stops the health checker
func (s *StreamService) StopHealthChecker() {
	s.healthChecker.Stop()
}

// StartStreamWarmer starts the stream warmer
func (s *StreamService) StartStreamWarmer() {
	s.streamWarmer.Start()
}

// StopStreamWarmer stops the stream warmer
func (s *StreamService) StopStreamWarmer() {
	s.streamWarmer.Stop()
}

// GetHealthMetrics returns the health metrics
func (s *StreamService) GetHealthMetrics() map[string]map[string]int64 {
	return map[string]map[string]int64{
		"health_checker": s.healthChecker.GetMetrics(),
		"stream_warmer":  s.streamWarmer.GetMetrics(),
	}
}

// GetStreamPoolMetrics returns the stream pool metrics
func (s *StreamService) GetStreamPoolMetrics() map[string]int64 {
	return s.poolManager.GetStreamPoolMetrics()
}

// GetStreamCount returns the number of streams for a peer
func (s *StreamService) GetStreamCount(peerID peer.ID) int {
	return s.poolManager.GetStreamCount(peerID)
}

// GetActiveStreamCount returns the number of active streams for a peer
func (s *StreamService) GetActiveStreamCount(peerID peer.ID) int {
	return s.poolManager.GetActiveStreamCount(peerID)
}

// Ensure StreamService implements all the required interfaces
var (
	_ types.Service          = (*StreamService)(nil)
	_ types.PoolService      = (*StreamService)(nil)
	_ types.HealthService    = (*StreamService)(nil)
	_ types.MetricsService   = (*StreamService)(nil)
	_ types.LifecycleService = (*StreamService)(nil)
)
