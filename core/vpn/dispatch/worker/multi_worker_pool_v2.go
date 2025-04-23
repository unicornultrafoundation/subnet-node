package worker

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/pool"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
)

var multiWorkerPoolV2Log = logrus.WithField("service", "multi-worker-pool-v2")

// MultiWorkerPoolV2 manages multiple workers for a peer using V2 components
type MultiWorkerPoolV2 struct {
	// Core components
	peerID        peer.ID
	streamManager *pool.StreamManagerV2

	// Configuration
	config *MultiWorkerPoolConfig

	// Context for the worker pool
	ctx    context.Context
	cancel context.CancelFunc

	// Worker management
	workersMu sync.RWMutex
	workers   map[types.ConnectionKey]*PacketWorkerV2

	// Lifecycle management
	stopChan chan struct{}
	running  bool

	// Resilience service
	resilienceService *resilience.ResilienceService

	// Metrics
	metrics struct {
		WorkersCreated   int64
		WorkersRemoved   int64
		PacketsProcessed int64
		PacketsDropped   int64
		Errors           int64
	}
}

// NewMultiWorkerPoolV2 creates a new multi-worker pool using V2 components
func NewMultiWorkerPoolV2(
	peerID peer.ID,
	streamManager *pool.StreamManagerV2,
	config *MultiWorkerPoolConfig,
	resilienceService *resilience.ResilienceService,
) *MultiWorkerPoolV2 {
	ctx, cancel := context.WithCancel(context.Background())

	// Use default resilience service if none provided
	if resilienceService == nil {
		resilienceService = resilience.NewResilienceService(nil)
	}

	return &MultiWorkerPoolV2{
		peerID:            peerID,
		streamManager:     streamManager,
		config:            config,
		ctx:               ctx,
		cancel:            cancel,
		workers:           make(map[types.ConnectionKey]*PacketWorkerV2),
		stopChan:          make(chan struct{}),
		running:           false,
		resilienceService: resilienceService,
	}
}

// Start starts the worker pool and its cleanup routine
func (p *MultiWorkerPoolV2) Start() {
	if p.running {
		return
	}

	p.running = true

	// Start the cleanup routine
	go p.cleanupIdleWorkers()

	multiWorkerPoolV2Log.WithField("peer_id", p.peerID.String()).Info("Multi-worker pool V2 started")
}

// Stop stops the worker pool and all its workers
func (p *MultiWorkerPoolV2) Stop() {
	if !p.running {
		return
	}

	// Signal shutdown
	close(p.stopChan)
	p.running = false

	// Cancel the context to stop all operations
	p.cancel()

	// Stop all workers
	p.workersMu.Lock()
	for _, worker := range p.workers {
		worker.Stop()
	}
	p.workers = make(map[types.ConnectionKey]*PacketWorkerV2)
	p.workersMu.Unlock()

	multiWorkerPoolV2Log.WithField("peer_id", p.peerID.String()).Info("Multi-worker pool V2 stopped")
}

// DispatchPacket dispatches a packet to the appropriate worker
func (p *MultiWorkerPoolV2) DispatchPacket(
	ctx context.Context,
	connKey types.ConnectionKey,
	destIP string,
	packet *types.QueuedPacket,
) error {
	if !p.running {
		return types.ErrWorkerPoolStopped
	}

	// Get or create a worker for this connection
	worker, err := p.getOrCreateWorker(connKey)
	if err != nil {
		atomic.AddInt64(&p.metrics.PacketsDropped, 1)
		atomic.AddInt64(&p.metrics.Errors, 1)
		return err
	}

	// Dispatch the packet to the worker
	err = worker.ProcessPacket(ctx, destIP, packet)
	if err != nil {
		atomic.AddInt64(&p.metrics.PacketsDropped, 1)
		atomic.AddInt64(&p.metrics.Errors, 1)
		return err
	}

	atomic.AddInt64(&p.metrics.PacketsProcessed, 1)
	return nil
}

// getOrCreateWorker gets or creates a worker for a connection
func (p *MultiWorkerPoolV2) getOrCreateWorker(connKey types.ConnectionKey) (*PacketWorkerV2, error) {
	// Check if we already have a worker for this connection
	p.workersMu.RLock()
	worker, exists := p.workers[connKey]
	p.workersMu.RUnlock()

	if exists {
		return worker, nil
	}

	// Check if we've reached the maximum number of workers
	p.workersMu.RLock()
	workerCount := len(p.workers)
	p.workersMu.RUnlock()

	if p.config.MaxWorkersPerPeer > 0 && workerCount >= p.config.MaxWorkersPerPeer {
		return nil, types.ErrMaxWorkersReached
	}

	// Create a new worker
	p.workersMu.Lock()
	defer p.workersMu.Unlock()

	// Check again in case another goroutine created it while we were waiting for the lock
	if worker, exists = p.workers[connKey]; exists {
		return worker, nil
	}

	// Create worker configuration
	workerConfig := &PacketWorkerConfig{
		BufferSize:  p.config.WorkerBufferSize,
		IdleTimeout: time.Duration(p.config.WorkerIdleTimeout) * time.Second,
	}

	// Create a new worker
	worker = NewPacketWorkerV2(p.ctx, p.peerID, connKey, p.streamManager, workerConfig, p.resilienceService)

	// Store the worker
	p.workers[connKey] = worker

	// Start the worker
	worker.Start()

	// Update metrics
	atomic.AddInt64(&p.metrics.WorkersCreated, 1)

	multiWorkerPoolV2Log.WithFields(logrus.Fields{
		"peer_id":     p.peerID.String(),
		"conn_key":    connKey.String(),
		"worker_mode": "connection",
	}).Debug("Created new packet worker V2")

	return worker, nil
}

// cleanupIdleWorkers periodically removes idle workers
func (p *MultiWorkerPoolV2) cleanupIdleWorkers() {
	ticker := time.NewTicker(p.config.WorkerCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.removeIdleWorkers()
		case <-p.stopChan:
			return
		}
	}
}

// removeIdleWorkers removes idle workers
func (p *MultiWorkerPoolV2) removeIdleWorkers() {
	now := time.Now()

	p.workersMu.Lock()
	defer p.workersMu.Unlock()

	for connKey, worker := range p.workers {
		if worker.IsIdle(now) {
			// Stop the worker
			worker.Stop()

			// Remove the worker from the map
			delete(p.workers, connKey)

			// Update metrics
			atomic.AddInt64(&p.metrics.WorkersRemoved, 1)

			multiWorkerPoolV2Log.WithFields(logrus.Fields{
				"peer_id":  p.peerID.String(),
				"conn_key": connKey.String(),
			}).Debug("Removed idle packet worker V2")
		}
	}
}

// GetMetrics returns the worker pool's metrics
func (p *MultiWorkerPoolV2) GetMetrics() map[string]int64 {
	return map[string]int64{
		"workers_created":   atomic.LoadInt64(&p.metrics.WorkersCreated),
		"workers_removed":   atomic.LoadInt64(&p.metrics.WorkersRemoved),
		"packets_processed": atomic.LoadInt64(&p.metrics.PacketsProcessed),
		"packets_dropped":   atomic.LoadInt64(&p.metrics.PacketsDropped),
		"errors":            atomic.LoadInt64(&p.metrics.Errors),
		"active_workers":    int64(p.GetWorkerCount()),
	}
}

// GetWorkerCount returns the number of active workers
func (p *MultiWorkerPoolV2) GetWorkerCount() int {
	p.workersMu.RLock()
	defer p.workersMu.RUnlock()
	return len(p.workers)
}

// Close implements io.Closer
func (p *MultiWorkerPoolV2) Close() error {
	p.Stop()
	return nil
}
