package worker

import (
	"context"
	"fmt"
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

	// Worker management - using multi-connection workers
	workersMu sync.RWMutex
	workers   []*MultiConnectionWorker

	// Connection to worker mapping
	connectionMap sync.Map // map[types.ConnectionKey]int - maps connection keys to worker indices

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
		workers:           make([]*MultiConnectionWorker, 0),
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
	p.workers = make([]*MultiConnectionWorker, 0)
	p.workersMu.Unlock()

	// Clear the connection map
	p.connectionMap = sync.Map{}

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

	// Set the destination IP in the packet
	packet.DestIP = destIP

	// Get or create a worker for this connection
	workerIndex, err := p.getOrCreateWorkerForConnection(connKey)
	if err != nil {
		atomic.AddInt64(&p.metrics.PacketsDropped, 1)
		atomic.AddInt64(&p.metrics.Errors, 1)
		return err
	}

	// Get the worker
	p.workersMu.RLock()
	if workerIndex >= len(p.workers) {
		p.workersMu.RUnlock()
		return fmt.Errorf("invalid worker index: %d", workerIndex)
	}
	worker := p.workers[workerIndex]
	p.workersMu.RUnlock()

	// Dispatch the packet to the worker
	if !worker.EnqueuePacket(packet, connKey) {
		atomic.AddInt64(&p.metrics.PacketsDropped, 1)
		atomic.AddInt64(&p.metrics.Errors, 1)
		return types.ErrWorkerChannelFull
	}

	atomic.AddInt64(&p.metrics.PacketsProcessed, 1)
	return nil
}

// getOrCreateWorkerForConnection gets or creates a worker for a connection
func (p *MultiWorkerPoolV2) getOrCreateWorkerForConnection(connKey types.ConnectionKey) (int, error) {
	// Check if we already have a worker for this connection
	if workerIndexVal, exists := p.connectionMap.Load(connKey); exists {
		return workerIndexVal.(int), nil
	}

	// Find the least loaded worker or create a new one
	return p.assignConnectionToWorker(connKey)
}

// assignConnectionToWorker assigns a connection to the least loaded worker or creates a new one
func (p *MultiWorkerPoolV2) assignConnectionToWorker(connKey types.ConnectionKey) (int, error) {
	p.workersMu.Lock()
	defer p.workersMu.Unlock()

	// If we have no workers, create one
	if len(p.workers) == 0 {
		return p.createNewWorker(connKey)
	}

	// If we have reached the maximum number of workers, find the least loaded one
	if p.config.MaxWorkersPerPeer > 0 && len(p.workers) >= p.config.MaxWorkersPerPeer {
		return p.findLeastLoadedWorker(connKey)
	}

	// Check if any existing worker has capacity
	for i, worker := range p.workers {
		// Check if the worker has capacity (less than 75% buffer utilization)
		if worker.GetBufferUtilization() < 75 {
			// Assign the connection to this worker
			p.connectionMap.Store(connKey, i)
			return i, nil
		}
	}

	// All workers are loaded, create a new one
	return p.createNewWorker(connKey)
}

// findLeastLoadedWorker finds the worker with the least load
func (p *MultiWorkerPoolV2) findLeastLoadedWorker(connKey types.ConnectionKey) (int, error) {
	// Find the worker with the lowest buffer utilization
	leastLoadedIndex := 0
	leastLoad := 100 // Start with maximum load

	for i, worker := range p.workers {
		load := worker.GetBufferUtilization()
		if load < leastLoad {
			leastLoad = load
			leastLoadedIndex = i
		}
	}

	// Assign the connection to the least loaded worker
	p.connectionMap.Store(connKey, leastLoadedIndex)
	return leastLoadedIndex, nil
}

// createNewWorker creates a new worker
func (p *MultiWorkerPoolV2) createNewWorker(connKey types.ConnectionKey) (int, error) {
	// Check if we've reached the maximum number of workers
	if p.config.MaxWorkersPerPeer > 0 && len(p.workers) >= p.config.MaxWorkersPerPeer {
		return 0, types.ErrMaxWorkersReached
	}

	// Create a new worker context
	workerCtx, workerCancel := context.WithCancel(p.ctx)

	// Create a new worker
	workerID := fmt.Sprintf("worker-%s-%d", p.peerID.String()[:8], len(p.workers))

	// Create a StreamManagerAdapter to adapt StreamManagerV2 to StreamManagerInterface
	streamManagerAdapter := &StreamManagerV2Adapter{
		streamManager: p.streamManager,
	}

	worker := NewMultiConnectionWorker(
		workerID,
		"", // destIP is set per packet
		p.peerID,
		streamManagerAdapter,
		workerCtx,
		workerCancel,
		p.config.WorkerBufferSize,
		p.resilienceService,
	)

	// Add the worker to the pool
	workerIndex := len(p.workers)
	p.workers = append(p.workers, worker)

	// Assign the connection to this worker
	p.connectionMap.Store(connKey, workerIndex)

	// Start the worker
	worker.Start()

	// Update metrics
	atomic.AddInt64(&p.metrics.WorkersCreated, 1)

	multiWorkerPoolV2Log.WithFields(logrus.Fields{
		"peer_id":     p.peerID.String(),
		"worker_id":   workerID,
		"worker_mode": "multi-connection",
	}).Debug("Created new multi-connection worker")

	return workerIndex, nil
}

// StreamManagerV2Adapter adapts StreamManagerV2 to StreamManagerInterface
type StreamManagerV2Adapter struct {
	streamManager *pool.StreamManagerV2
}

// SendPacket implements StreamManagerInterface
func (a *StreamManagerV2Adapter) SendPacket(
	ctx context.Context,
	connKey types.ConnectionKey,
	peerID peer.ID,
	packet *types.QueuedPacket,
) error {
	return a.streamManager.SendPacket(ctx, connKey, peerID, packet)
}

// Start implements StreamManagerInterface
func (a *StreamManagerV2Adapter) Start() {
	a.streamManager.Start()
}

// Stop implements StreamManagerInterface
func (a *StreamManagerV2Adapter) Stop() {
	a.streamManager.Stop()
}

// GetMetrics implements StreamManagerInterface
func (a *StreamManagerV2Adapter) GetMetrics() map[string]int64 {
	return a.streamManager.GetMetrics()
}

// GetConnectionCount implements StreamManagerInterface
func (a *StreamManagerV2Adapter) GetConnectionCount() int {
	return a.streamManager.GetConnectionCount()
}

// Close implements StreamManagerInterface
func (a *StreamManagerV2Adapter) Close() error {
	return a.streamManager.Close()
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

	// We need at least one worker, so don't remove the last one
	if len(p.workers) <= 1 {
		return
	}

	// Check each worker
	for i := len(p.workers) - 1; i >= 1; i-- { // Start from the end, keep at least one worker
		worker := p.workers[i]

		// Check if the worker is idle and has no connections
		lastActivity := worker.GetLastActivity()
		if now.Sub(lastActivity) > time.Duration(p.config.WorkerIdleTimeout)*time.Second &&
			worker.GetConnectionCount() == 0 {
			// Stop the worker
			worker.Stop()

			// Remove the worker from the slice
			p.workers = append(p.workers[:i], p.workers[i+1:]...)

			// Update metrics
			atomic.AddInt64(&p.metrics.WorkersRemoved, 1)

			multiWorkerPoolV2Log.WithFields(logrus.Fields{
				"peer_id":   p.peerID.String(),
				"worker_id": worker.WorkerID,
			}).Debug("Removed idle multi-connection worker")
		}
	}

	// Rebuild the connection map if needed
	p.rebuildConnectionMap()
}

// rebuildConnectionMap rebuilds the connection map after workers have been removed
func (p *MultiWorkerPoolV2) rebuildConnectionMap() {
	// Create a new connection map
	newConnectionMap := sync.Map{}

	// Iterate through the old connection map
	p.connectionMap.Range(func(key, value interface{}) bool {
		connKey := key.(types.ConnectionKey)
		workerIndex := value.(int)

		// If the worker index is out of bounds, reassign to worker 0
		if workerIndex >= len(p.workers) {
			newConnectionMap.Store(connKey, 0)
		} else {
			newConnectionMap.Store(connKey, workerIndex)
		}

		return true
	})

	// Replace the old connection map
	p.connectionMap = newConnectionMap
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
