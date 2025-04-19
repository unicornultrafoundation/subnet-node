package stream

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/health"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/multiplex"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/pool"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

var log = logrus.WithField("service", "vpn-stream")

// StreamService implements the Service, PoolService, HealthCheckerService, StreamWarmerService, HealthMetricsService,
// MultiplexerService, MultiplexerMetricsService, StreamPoolMetricsService, and StreamLifecycleService interfaces
type StreamService struct {
	// Stream service for creating new streams
	streamService types.Service
	// Stream pool manager
	poolManager *pool.StreamPoolManager
	// Health checker
	healthChecker *health.HealthChecker
	// Stream warmer
	streamWarmer *health.StreamWarmer
	// Multiplexer manager
	multiplexManager *multiplex.MultiplexerManager
	// Whether multiplexing is enabled
	multiplexingEnabled bool
	// Mutex to protect access to the service
	mu sync.Mutex
}

// NewStreamService creates a new stream service
func NewStreamService(
	streamService types.Service,
	maxStreamsPerPeer int,
	minStreamsPerPeer int,
	streamIdleTimeout time.Duration,
	cleanupInterval time.Duration,
	healthCheckInterval time.Duration,
	healthCheckTimeout time.Duration,
	maxConsecutiveFailures int,
	warmInterval time.Duration,
	maxStreamsPerMultiplexer int,
	minStreamsPerMultiplexer int,
	autoScalingInterval time.Duration,
	multiplexingEnabled bool,
) *StreamService {
	// Create the stream pool manager
	poolManager := pool.NewStreamPoolManager(
		streamService,
		maxStreamsPerPeer,
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

	// Create the multiplexer manager
	multiplexManager := multiplex.NewMultiplexerManager(
		streamService,
		poolManager,
		maxStreamsPerMultiplexer,
		minStreamsPerMultiplexer,
		autoScalingInterval,
	)

	return &StreamService{
		streamService:       streamService,
		poolManager:         poolManager,
		healthChecker:       healthChecker,
		streamWarmer:        streamWarmer,
		multiplexManager:    multiplexManager,
		multiplexingEnabled: multiplexingEnabled,
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

	// Start the multiplexer manager if multiplexing is enabled
	if s.multiplexingEnabled {
		s.multiplexManager.Start()
	}

	log.Info("Stream service started")
}

// Stop stops the stream service
func (s *StreamService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop the multiplexer manager if multiplexing is enabled
	if s.multiplexingEnabled {
		s.multiplexManager.Stop()
	}

	// Stop the stream warmer
	s.streamWarmer.Stop()

	// Stop the health checker
	s.healthChecker.Stop()

	// Stop the stream pool manager
	s.poolManager.Stop()

	log.Info("Stream service stopped")
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

// StartMultiplexer starts the multiplexer manager
func (s *StreamService) StartMultiplexer() {
	if s.multiplexingEnabled {
		s.multiplexManager.Start()
	}
}

// StopMultiplexer stops the multiplexer manager
func (s *StreamService) StopMultiplexer() {
	if s.multiplexingEnabled {
		s.multiplexManager.Stop()
	}
}

// SendPacketMultiplexed sends a packet using the multiplexer
func (s *StreamService) SendPacketMultiplexed(ctx context.Context, peerID peer.ID, packet []byte) error {
	if !s.multiplexingEnabled {
		// If multiplexing is disabled, fall back to the regular stream pool
		stream, err := s.GetStream(ctx, peerID)
		if err != nil {
			return err
		}

		_, err = stream.Write(packet)
		if err != nil {
			// Release the stream as unhealthy
			s.ReleaseStream(peerID, stream, false)
			return err
		}

		// Release the stream as healthy
		s.ReleaseStream(peerID, stream, true)
		return nil
	}

	return s.multiplexManager.SendPacket(ctx, peerID, packet)
}

// GetMultiplexerMetrics returns the multiplexer metrics
func (s *StreamService) GetMultiplexerMetrics() map[string]*types.MultiplexerMetrics {
	if !s.multiplexingEnabled {
		return make(map[string]*types.MultiplexerMetrics)
	}

	return s.multiplexManager.GetMultiplexerMetrics()
}

// GetStreamPoolMetrics returns the stream pool metrics
func (s *StreamService) GetStreamPoolMetrics() map[string]int64 {
	return s.poolManager.GetMetrics()
}

// Ensure StreamService implements all the required interfaces
var (
	_ types.Service                   = (*StreamService)(nil)
	_ types.PoolService               = (*StreamService)(nil)
	_ types.HealthCheckerService      = (*StreamService)(nil)
	_ types.StreamWarmerService       = (*StreamService)(nil)
	_ types.HealthMetricsService      = (*StreamService)(nil)
	_ types.MultiplexerService        = (*StreamService)(nil)
	_ types.MultiplexerMetricsService = (*StreamService)(nil)
	_ types.StreamPoolMetricsService  = (*StreamService)(nil)
	_ types.StreamLifecycleService    = (*StreamService)(nil)
	_ types.HealthService             = (*StreamService)(nil)
	_ types.MultiplexService          = (*StreamService)(nil)
)
