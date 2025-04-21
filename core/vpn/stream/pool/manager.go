package pool

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

// StreamPoolManager manages the stream pool and provides periodic cleanup
type StreamPoolManager struct {
	// The stream pool
	pool *StreamPool
	// Cleanup interval
	cleanupInterval time.Duration
	// Context for the manager
	ctx context.Context
	// Cancel function for the manager context
	cancel context.CancelFunc
	// Mutex to protect access to the manager
	mu sync.Mutex
	// Whether the manager is running
	running bool
	// Channel to signal manager shutdown
	stopChan chan struct{}
}

// NewStreamPoolManager creates a new stream pool manager
func NewStreamPoolManager(
	streamService types.Service,
	minStreamsPerPeer int,
	streamIdleTimeout time.Duration,
	cleanupInterval time.Duration,
) *StreamPoolManager {
	ctx, cancel := context.WithCancel(context.Background())

	pool := NewStreamPool(
		streamService,
		minStreamsPerPeer,
		streamIdleTimeout,
	)

	return &StreamPoolManager{
		pool:            pool,
		cleanupInterval: cleanupInterval,
		ctx:             ctx,
		cancel:          cancel,
		running:         false,
		stopChan:        make(chan struct{}),
	}
}

// Start starts the stream pool manager
func (m *StreamPoolManager) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return
	}

	m.running = true

	// Start the cleanup goroutine
	go m.runCleanup()

	log.Infof("Stream pool manager started with cleanup interval %s", m.cleanupInterval)
}

// Stop stops the stream pool manager
func (m *StreamPoolManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.running = false

	// Cancel the context
	m.cancel()

	// Signal the cleanup goroutine to stop
	close(m.stopChan)

	// Close the pool
	m.pool.Close()

	log.Info("Stream pool manager stopped")
}

// Close implements the io.Closer interface
func (m *StreamPoolManager) Close() error {
	m.Stop()
	return nil
}

// runCleanup runs the cleanup loop
func (m *StreamPoolManager) runCleanup() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Cleanup idle streams
			m.pool.CleanupIdleStreams()
		case <-m.ctx.Done():
			return
		case <-m.stopChan:
			return
		}
	}
}

// GetStream gets a stream from the pool or creates a new one
func (m *StreamPoolManager) GetStream(ctx context.Context, peerID peer.ID) (types.VPNStream, error) {
	return m.pool.GetStream(ctx, peerID)
}

// ReleaseStream returns a stream to the pool
func (m *StreamPoolManager) ReleaseStream(peerID peer.ID, stream types.VPNStream, healthy bool) {
	m.pool.ReleaseStream(peerID, stream, healthy)
}

// GetStreamPoolMetrics returns the stream pool metrics (for MetricsService interface)
func (m *StreamPoolManager) GetStreamPoolMetrics() map[string]int64 {
	return m.pool.GetMetrics()
}

// GetHealthMetrics returns empty health metrics (for MetricsService interface)
func (m *StreamPoolManager) GetHealthMetrics() map[string]map[string]int64 {
	// Pool manager doesn't have health metrics
	return map[string]map[string]int64{}
}

// Ensure StreamPoolManager implements the required interfaces
var (
	_ types.PoolService      = (*StreamPoolManager)(nil)
	_ types.MetricsService   = (*StreamPoolManager)(nil)
	_ types.LifecycleService = (*StreamPoolManager)(nil)
)

// GetStreamCount returns the number of streams for a peer
func (m *StreamPoolManager) GetStreamCount(peerID peer.ID) int {
	return m.pool.GetStreamCount(peerID)
}

// GetActiveStreamCount returns the number of active streams for a peer
func (m *StreamPoolManager) GetActiveStreamCount(peerID peer.ID) int {
	return m.pool.GetActiveStreamCount(peerID)
}

// GetMinStreamsPerPeer returns the minimum number of streams per peer
func (m *StreamPoolManager) GetMinStreamsPerPeer() int {
	return m.pool.minStreamsPerPeer
}

// GetAllPeers returns a list of all peers in the pool
func (m *StreamPoolManager) GetAllPeers() []peer.ID {
	return m.pool.GetAllPeers()
}
