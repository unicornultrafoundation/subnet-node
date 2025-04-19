package multiplex

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

// MultiplexerManager manages stream multiplexers for different peers
type MultiplexerManager struct {
	// Map of peer ID to stream multiplexer
	multiplexers map[string]*StreamMultiplexer
	// Mutex to protect access to multiplexers
	mu sync.RWMutex
	// Stream service for creating new streams
	streamService types.Service
	// Stream pool service for managing streams
	streamPoolService types.PoolService
	// Maximum streams per multiplexer
	maxStreamsPerMultiplexer int
	// Minimum streams per multiplexer
	minStreamsPerMultiplexer int
	// Auto-scaling interval
	autoScalingInterval time.Duration
	// Context for the manager
	ctx context.Context
	// Cancel function for the manager context
	cancel context.CancelFunc
	// Channel to signal manager shutdown
	stopChan chan struct{}
	// Metrics for the manager
	metrics struct {
		// Multiplexers created
		MultiplexersCreated int64
		// Multiplexers closed
		MultiplexersClosed int64
		// Scale up operations
		ScaleUpOperations int64
		// Scale down operations
		ScaleDownOperations int64
	}
}

// NewMultiplexerManager creates a new multiplexer manager
func NewMultiplexerManager(
	streamService types.Service,
	streamPoolService types.PoolService,
	maxStreamsPerMultiplexer int,
	minStreamsPerMultiplexer int,
	autoScalingInterval time.Duration,
) *MultiplexerManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &MultiplexerManager{
		multiplexers:             make(map[string]*StreamMultiplexer),
		streamService:            streamService,
		streamPoolService:        streamPoolService,
		maxStreamsPerMultiplexer: maxStreamsPerMultiplexer,
		minStreamsPerMultiplexer: minStreamsPerMultiplexer,
		autoScalingInterval:      autoScalingInterval,
		ctx:                      ctx,
		cancel:                   cancel,
		stopChan:                 make(chan struct{}),
	}
}

// Start starts the multiplexer manager
func (m *MultiplexerManager) Start() {
	go m.runAutoScaling()
	log.Infof("Multiplexer manager started with auto-scaling interval %s", m.autoScalingInterval)
}

// Stop stops the multiplexer manager
func (m *MultiplexerManager) Stop() {
	m.cancel()
	close(m.stopChan)

	// Close all multiplexers
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, multiplexer := range m.multiplexers {
		multiplexer.Close()
		atomic.AddInt64(&m.metrics.MultiplexersClosed, 1)
	}

	m.multiplexers = make(map[string]*StreamMultiplexer)
	log.Info("Multiplexer manager stopped")
}

// SendPacket sends a packet using the appropriate multiplexer
func (m *MultiplexerManager) SendPacket(ctx context.Context, peerID peer.ID, packet []byte) error {
	multiplexer := m.getOrCreateMultiplexer(ctx, peerID)

	err := multiplexer.SendPacket(ctx, packet)
	if err != nil {
		return err
	}

	return nil
}

// getOrCreateMultiplexer gets or creates a multiplexer for a peer
func (m *MultiplexerManager) getOrCreateMultiplexer(ctx context.Context, peerID peer.ID) *StreamMultiplexer {
	peerIDStr := peerID.String()

	// Fast path: check if multiplexer exists
	m.mu.RLock()
	multiplexer, exists := m.multiplexers[peerIDStr]
	m.mu.RUnlock()

	if exists {
		// Check if the multiplexer has any streams
		if multiplexer.GetStreamCount() == 0 {
			// Try to recover the multiplexer
			err := multiplexer.recoverStreams(ctx)
			if err != nil {
				log.Errorf("Failed to recover multiplexer for peer %s: %v", peerIDStr, err)
			}
		}
		return multiplexer
	}

	// Slow path: create new multiplexer
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check again in case another goroutine created the multiplexer
	multiplexer, exists = m.multiplexers[peerIDStr]
	if exists {
		// Check if the multiplexer has any streams
		if multiplexer.GetStreamCount() == 0 {
			// Try to recover the multiplexer
			err := multiplexer.recoverStreams(ctx)
			if err != nil {
				log.Errorf("Failed to recover multiplexer for peer %s: %v", peerIDStr, err)
			}
		}
		return multiplexer
	}

	// Create new multiplexer
	multiplexer = NewStreamMultiplexer(
		peerID,
		m.streamService,
		m.streamPoolService,
		m.maxStreamsPerMultiplexer,
		m.minStreamsPerMultiplexer,
	)

	// Initialize the multiplexer
	err := multiplexer.Initialize(ctx)
	if err != nil {
		log.Errorf("Failed to initialize multiplexer for peer %s: %v", peerIDStr, err)
		// If initialization failed but we have at least one stream, we can continue
		if multiplexer.GetStreamCount() == 0 {
			// Try to recover the multiplexer
			err = multiplexer.recoverStreams(ctx)
			if err != nil {
				log.Errorf("Failed to recover multiplexer for peer %s: %v", peerIDStr, err)
				// Return an empty multiplexer, it will try to recover on first use
			}
		}
	}

	m.multiplexers[peerIDStr] = multiplexer
	atomic.AddInt64(&m.metrics.MultiplexersCreated, 1)

	return multiplexer
}

// runAutoScaling runs the auto-scaling loop
func (m *MultiplexerManager) runAutoScaling() {
	ticker := time.NewTicker(m.autoScalingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.performAutoScaling()
		case <-m.ctx.Done():
			return
		case <-m.stopChan:
			return
		}
	}
}

// performAutoScaling performs auto-scaling of multiplexers
func (m *MultiplexerManager) performAutoScaling() {
	// Get a copy of the multiplexers
	m.mu.RLock()
	multiplexers := make([]*StreamMultiplexer, 0, len(m.multiplexers))
	for _, multiplexer := range m.multiplexers {
		multiplexers = append(multiplexers, multiplexer)
	}
	m.mu.RUnlock()

	for _, multiplexer := range multiplexers {
		metrics := multiplexer.GetMetrics()

		// Scale up if the multiplexer is busy and has capacity
		if metrics.PacketsSent > 1000 && metrics.AvgLatency > 50 {
			err := multiplexer.ScaleUp(m.ctx)
			if err == nil {
				atomic.AddInt64(&m.metrics.ScaleUpOperations, 1)
			}
		}

		// Scale down if the multiplexer is idle
		if metrics.PacketsSent < 100 && multiplexer.GetStreamCount() > m.minStreamsPerMultiplexer {
			multiplexer.ScaleDown()
			atomic.AddInt64(&m.metrics.ScaleDownOperations, 1)
		}

		// Reset metrics for the next interval
		atomic.StoreInt64(&metrics.PacketsSent, 0)
	}
}

// GetMultiplexerMetrics returns the metrics for all multiplexers
func (m *MultiplexerManager) GetMultiplexerMetrics() map[string]*types.MultiplexerMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := make(map[string]*types.MultiplexerMetrics, len(m.multiplexers))

	for peerIDStr, multiplexer := range m.multiplexers {
		metrics[peerIDStr] = multiplexer.GetMetrics()
	}

	return metrics
}

// GetManagerMetrics returns the metrics for the manager
func (m *MultiplexerManager) GetManagerMetrics() map[string]int64 {
	return map[string]int64{
		"multiplexers_created":  atomic.LoadInt64(&m.metrics.MultiplexersCreated),
		"multiplexers_closed":   atomic.LoadInt64(&m.metrics.MultiplexersClosed),
		"scale_up_operations":   atomic.LoadInt64(&m.metrics.ScaleUpOperations),
		"scale_down_operations": atomic.LoadInt64(&m.metrics.ScaleDownOperations),
	}
}

// StartMultiplexer starts the multiplexer manager (for MultiplexerService interface)
func (m *MultiplexerManager) StartMultiplexer() {
	m.Start()
}

// StopMultiplexer stops the multiplexer manager (for MultiplexerService interface)
func (m *MultiplexerManager) StopMultiplexer() {
	m.Stop()
}

// SendPacketMultiplexed sends a packet using the multiplexer (for MultiplexerService interface)
func (m *MultiplexerManager) SendPacketMultiplexed(ctx context.Context, peerID peer.ID, packet []byte) error {
	return m.SendPacket(ctx, peerID, packet)
}

// Ensure MultiplexerManager implements the required interfaces
var (
	_ types.MultiplexerService        = (*MultiplexerManager)(nil)
	_ types.MultiplexerMetricsService = (*MultiplexerManager)(nil)
	_ types.StreamLifecycleService    = (*MultiplexerManager)(nil)
)
