package health

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/metrics"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/pool"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

// StreamWarmer ensures that there are enough streams in the pool
type StreamWarmer struct {
	// Stream pool manager
	poolManager *pool.StreamPoolManager
	// Stream service for creating new streams
	streamService types.Service
	// Warm interval
	warmInterval time.Duration
	// Minimum streams per peer
	minStreamsPerPeer int
	// Context for the warmer
	ctx context.Context
	// Cancel function for the warmer context
	cancel context.CancelFunc
	// Mutex to protect access to the warmer
	mu sync.Mutex
	// Whether the warmer is running
	running bool
	// Channel to signal warmer shutdown
	stopChan chan struct{}
	// Metrics for the stream warmer
	metrics *metrics.HealthMetrics
	// Known peers
	knownPeers map[string]peer.ID
}

// NewStreamWarmer creates a new stream warmer
func NewStreamWarmer(
	poolManager *pool.StreamPoolManager,
	streamService types.Service,
	warmInterval time.Duration,
	minStreamsPerPeer int,
) *StreamWarmer {
	ctx, cancel := context.WithCancel(context.Background())

	return &StreamWarmer{
		poolManager:       poolManager,
		streamService:     streamService,
		warmInterval:      warmInterval,
		minStreamsPerPeer: minStreamsPerPeer,
		ctx:               ctx,
		cancel:            cancel,
		running:           false,
		stopChan:          make(chan struct{}),
		metrics:           metrics.NewHealthMetrics(),
		knownPeers:        make(map[string]peer.ID),
	}
}

// Start starts the stream warmer
func (w *StreamWarmer) Start() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return
	}

	w.running = true

	// Start the warm goroutine
	go w.runWarm()

	log.Infof("Stream warmer started with interval %s", w.warmInterval)
}

// Stop stops the stream warmer
func (w *StreamWarmer) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	w.running = false

	// Cancel the context
	w.cancel()

	// Signal the warm goroutine to stop
	close(w.stopChan)

	log.Info("Stream warmer stopped")
}

// runWarm runs the warm loop
func (w *StreamWarmer) runWarm() {
	ticker := time.NewTicker(w.warmInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Warm streams
			w.warmStreams()
		case <-w.ctx.Done():
			return
		case <-w.stopChan:
			return
		}
	}
}

// warmStreams ensures that there are enough streams in the pool
func (w *StreamWarmer) warmStreams() {
	// Update metrics
	w.metrics.IncrementWarmsPerformed()

	// Get a list of known peers
	w.mu.Lock()
	peers := make([]peer.ID, 0, len(w.knownPeers))
	for _, peerID := range w.knownPeers {
		peers = append(peers, peerID)
	}
	w.mu.Unlock()

	// For each peer, ensure that there are enough streams
	for _, peerID := range peers {
		// Check if we have enough streams
		streamCount := w.poolManager.GetStreamCount(peerID)

		// If we don't have enough streams, create more
		if streamCount < w.minStreamsPerPeer {
			w.ensureMinStreams(peerID, w.minStreamsPerPeer-streamCount)
		}
	}
}

// ensureMinStreams ensures that we have at least minStreams streams for a peer
func (w *StreamWarmer) ensureMinStreams(peerID peer.ID, count int) {
	// Create a new context for this operation
	ctx, cancel := context.WithTimeout(w.ctx, 5*time.Second)
	defer cancel()

	// Create new streams
	for i := 0; i < count; i++ {
		// Create a new stream
		stream, err := w.streamService.CreateNewVPNStream(ctx, peerID)
		if err != nil {
			log.Errorf("Failed to create new stream for peer %s: %v", peerID.String(), err)
			w.metrics.IncrementWarmFailures()
			return
		}

		// Release the stream to the pool
		w.poolManager.ReleaseStream(peerID, stream, true)

		// Update metrics
		w.metrics.IncrementStreamsWarmed()
	}
}

// AddPeer adds a peer to the list of known peers
func (w *StreamWarmer) AddPeer(peerID peer.ID) {
	w.mu.Lock()
	defer w.mu.Unlock()

	peerIDStr := peerID.String()
	w.knownPeers[peerIDStr] = peerID
}

// RemovePeer removes a peer from the list of known peers
func (w *StreamWarmer) RemovePeer(peerID peer.ID) {
	w.mu.Lock()
	defer w.mu.Unlock()

	peerIDStr := peerID.String()
	delete(w.knownPeers, peerIDStr)
}

// GetMetrics returns the current metrics
func (w *StreamWarmer) GetMetrics() map[string]int64 {
	return map[string]int64{
		"warms_performed": w.metrics.WarmsPerformed,
		"streams_warmed":  w.metrics.StreamsWarmed,
		"warm_failures":   w.metrics.WarmFailures,
	}
}

// StartStreamWarmer starts the stream warmer (for StreamWarmerService interface)
func (w *StreamWarmer) StartStreamWarmer() {
	w.Start()
}

// StopStreamWarmer stops the stream warmer (for StreamWarmerService interface)
func (w *StreamWarmer) StopStreamWarmer() {
	w.Stop()
}

// Ensure StreamWarmer implements the required interfaces
var (
	_ types.StreamWarmerService    = (*StreamWarmer)(nil)
	_ types.StreamLifecycleService = (*StreamWarmer)(nil)
)
