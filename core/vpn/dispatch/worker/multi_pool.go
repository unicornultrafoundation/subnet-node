package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
)

var multiPoolLog = logrus.WithField("service", "multi-worker-pool")

// workerOp represents an operation on the worker pool
type workerOp struct {
	opType     string // "get_worker", "dispatch", "get_count", "get_metrics", "cleanup"
	connKey    types.ConnectionKey
	destIP     string
	ctx        context.Context
	packet     *types.QueuedPacket
	resultChan chan workerOpResult
}

// workerOpResult represents the result of a worker operation
type workerOpResult struct {
	worker  MultiConnectionWorkerInterface
	index   int64
	count   int
	metrics map[string]int64
	err     error
}

// MultiWorkerPool manages a pool of multi-connection workers for a specific peer ID
type MultiWorkerPool struct {
	// PeerID is the peer ID this pool is responsible for
	PeerID peer.ID
	// StreamManager provides access to stream management
	StreamManager StreamManagerInterface
	// Workers is a slice of multi-connection workers (managed by workerManager)
	workers []MultiConnectionWorkerInterface
	// ConnectionMap maps connection keys to worker indices
	ConnectionMap sync.Map
	// Context for the worker pool
	Ctx context.Context
	// Cancel function for the worker pool context
	Cancel context.CancelFunc
	// Worker idle timeout in seconds
	WorkerIdleTimeout int
	// Worker cleanup interval
	WorkerCleanupInterval time.Duration
	// Worker buffer size
	WorkerBufferSize int
	// Maximum workers per peer
	MaxWorkersPerPeer int
	// Channel to signal worker pool shutdown
	StopChan chan struct{}
	// Whether the worker pool is running (0 = not running, 1 = running)
	running int32
	// Resilience service
	ResilienceService *resilience.ResilienceService
	// Channel for worker operations
	opChan chan workerOp
	// Metrics
	Metrics struct {
		WorkersCreated int64
		WorkersRemoved int64
		PacketsHandled int64
		PacketsDropped int64
		Errors         int64
	}
}

// MultiWorkerPoolConfig contains configuration for the multi-worker pool
type MultiWorkerPoolConfig struct {
	WorkerIdleTimeout     int
	WorkerCleanupInterval time.Duration
	WorkerBufferSize      int
	MaxWorkersPerPeer     int
}

// NewMultiWorkerPool creates a new multi-worker pool for a specific peer ID
func NewMultiWorkerPool(
	peerID peer.ID,
	streamManager StreamManagerInterface,
	config *MultiWorkerPoolConfig,
	resilienceService *resilience.ResilienceService,
) *MultiWorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	// Use default resilience service if none provided
	if resilienceService == nil {
		resilienceService = resilience.NewResilienceService(nil)
	}

	pool := &MultiWorkerPool{
		PeerID:                peerID,
		StreamManager:         streamManager,
		workers:               make([]MultiConnectionWorkerInterface, 0),
		Ctx:                   ctx,
		Cancel:                cancel,
		WorkerIdleTimeout:     config.WorkerIdleTimeout,
		WorkerCleanupInterval: config.WorkerCleanupInterval,
		WorkerBufferSize:      config.WorkerBufferSize,
		MaxWorkersPerPeer:     config.MaxWorkersPerPeer,
		StopChan:              make(chan struct{}),
		ResilienceService:     resilienceService,
		opChan:                make(chan workerOp, 100), // Buffer for operations
	}

	// Initialize atomic values
	atomic.StoreInt32(&pool.running, 0) // Not running initially

	return pool
}

// Start starts the worker pool and its management routines
func (p *MultiWorkerPool) Start() {
	if atomic.LoadInt32(&p.running) == 1 {
		return
	}

	atomic.StoreInt32(&p.running, 1)

	// Start the worker manager routine
	go p.workerManager()

	// Start the cleanup routine
	go p.cleanupInactiveWorkers()

	multiPoolLog.WithField("peer_id", p.PeerID.String()).Info("Multi-worker pool started")
}

// workerManager handles worker operations through the operation channel
func (p *MultiWorkerPool) workerManager() {
	logger := multiPoolLog.WithField("peer_id", p.PeerID.String())
	logger.Debug("Worker manager started")

	for {
		select {
		case <-p.StopChan:
			logger.Debug("Worker manager stopping")
			return
		case op := <-p.opChan:
			switch op.opType {
			case "get_worker":
				worker, idx, err := p.getOrCreateWorkerInternal(op.ctx, op.connKey, op.destIP)
				op.resultChan <- workerOpResult{worker: worker, index: idx, err: err}
			case "dispatch":
				worker, _, err := p.getOrCreateWorkerInternal(op.ctx, op.connKey, op.destIP)
				if err != nil {
					atomic.AddInt64(&p.Metrics.Errors, 1)
					op.resultChan <- workerOpResult{err: err}
					continue
				}

				// Try to add the packet to the worker's queue
				if !worker.EnqueuePacket(op.packet, op.connKey) {
					atomic.AddInt64(&p.Metrics.PacketsDropped, 1)
					op.resultChan <- workerOpResult{err: types.ErrWorkerQueueFull}
					continue
				}

				// Update metrics
				atomic.AddInt64(&p.Metrics.PacketsHandled, 1)
				op.resultChan <- workerOpResult{err: nil}
			case "get_count":
				op.resultChan <- workerOpResult{count: len(p.workers)}
			case "get_metrics":
				metrics := p.getMetricsInternal()
				op.resultChan <- workerOpResult{metrics: metrics}
			case "cleanup":
				p.cleanupWorkersInternal(op.ctx)
				op.resultChan <- workerOpResult{}
			case "stop_all":
				// Stop all workers
				for _, worker := range p.workers {
					worker.Stop()
				}
				p.workers = make([]MultiConnectionWorkerInterface, 0)
				op.resultChan <- workerOpResult{}
			}
		}
	}
}

// Stop stops the worker pool and all its workers
func (p *MultiWorkerPool) Stop() {
	if atomic.LoadInt32(&p.running) == 0 {
		return
	}

	// Signal the cleanup routine to stop
	close(p.StopChan)
	atomic.StoreInt32(&p.running, 0)

	// Cancel the context to stop all workers
	p.Cancel()

	// Stop all workers through the worker manager
	resultChan := make(chan workerOpResult, 1)
	p.opChan <- workerOp{
		opType:     "stop_all",
		resultChan: resultChan,
	}
	<-resultChan // Wait for the operation to complete

	// Clear the connection map
	p.ConnectionMap = sync.Map{}

	multiPoolLog.WithField("peer_id", p.PeerID.String()).Info("Multi-worker pool stopped")
}

// getOrCreateWorker gets an existing worker or creates a new one for a connection key
func (p *MultiWorkerPool) getOrCreateWorker(
	ctx context.Context,
	connKey types.ConnectionKey,
	destIP string,
) (MultiConnectionWorkerInterface, error) {
	// Use the worker manager to get or create a worker
	resultChan := make(chan workerOpResult, 1)
	p.opChan <- workerOp{
		opType:     "get_worker",
		connKey:    connKey,
		destIP:     destIP,
		ctx:        ctx,
		resultChan: resultChan,
	}

	// Wait for the result
	result := <-resultChan
	if result.err != nil {
		return nil, result.err
	}

	// Assign the connection key to this worker
	p.ConnectionMap.Store(connKey, result.index)

	return result.worker, nil
}

// getOrCreateWorkerInternal is the internal implementation of getOrCreateWorker
// It's called by the worker manager routine
func (p *MultiWorkerPool) getOrCreateWorkerInternal(
	ctx context.Context,
	connKey types.ConnectionKey,
	destIP string,
) (MultiConnectionWorkerInterface, int64, error) {
	// Check if we already have a worker assigned to this connection key
	if workerIdx, ok := p.ConnectionMap.Load(connKey); ok {
		if int(workerIdx.(int64)) < len(p.workers) {
			worker := p.workers[workerIdx.(int64)]
			if worker.IsRunning() {
				return worker, workerIdx.(int64), nil
			}
		}
		// Worker doesn't exist or isn't running, remove the mapping
		p.ConnectionMap.Delete(connKey)
	}

	// If we have workers, find the least loaded one
	if len(p.workers) > 0 {
		leastConnections := int(^uint(0) >> 1) // Max int
		leastLoadedIdx := 0
		foundRunningWorker := false

		for i, worker := range p.workers {
			if !worker.IsRunning() {
				continue
			}

			foundRunningWorker = true
			connections := worker.GetConnectionCount()
			if connections < leastConnections {
				leastConnections = connections
				leastLoadedIdx = i
			}
		}

		// If we found a running worker, use it
		if foundRunningWorker {
			worker := p.workers[leastLoadedIdx]
			return worker, int64(leastLoadedIdx), nil
		}
	}

	// Check if we've reached the maximum number of workers
	if len(p.workers) >= p.MaxWorkersPerPeer {
		// If we get here, all workers are stopped, so we'll replace the first one
		if len(p.workers) > 0 {
			workerIdx := 0
			// Create a new worker to replace the stopped one
			workerCtx, workerCancel := context.WithCancel(ctx)
			workerID := fmt.Sprintf("%s-worker-%d", p.PeerID.String()[:8], workerIdx)
			worker := NewMultiConnectionWorker(
				workerID,
				destIP,
				p.PeerID,
				p.StreamManager,
				workerCtx,
				workerCancel,
				p.WorkerBufferSize,
				p.ResilienceService,
			)

			// Start the worker
			worker.Start()

			// Replace the worker
			p.workers[workerIdx] = worker

			// Update metrics
			atomic.AddInt64(&p.Metrics.WorkersCreated, 1)

			return worker, int64(workerIdx), nil
		}

		return nil, 0, fmt.Errorf("maximum number of workers reached for peer %s", p.PeerID.String())
	}

	// Create a new worker
	workerIdx := len(p.workers)
	workerCtx, workerCancel := context.WithCancel(ctx)
	workerID := fmt.Sprintf("%s-worker-%d", p.PeerID.String()[:8], workerIdx)
	worker := NewMultiConnectionWorker(
		workerID,
		destIP,
		p.PeerID,
		p.StreamManager,
		workerCtx,
		workerCancel,
		p.WorkerBufferSize,
		p.ResilienceService,
	)

	// Start the worker
	worker.Start()

	// Add the worker to the pool
	p.workers = append(p.workers, worker)

	// Update metrics
	atomic.AddInt64(&p.Metrics.WorkersCreated, 1)

	return worker, int64(workerIdx), nil
}

// DispatchPacket dispatches a packet to the appropriate worker
func (p *MultiWorkerPool) DispatchPacket(
	ctx context.Context,
	connKey types.ConnectionKey,
	destIP string,
	packet *types.QueuedPacket,
) error {
	// Use the worker manager to dispatch the packet
	resultChan := make(chan workerOpResult, 1)
	p.opChan <- workerOp{
		opType:     "dispatch",
		connKey:    connKey,
		destIP:     destIP,
		ctx:        ctx,
		packet:     packet,
		resultChan: resultChan,
	}

	// Wait for the result
	result := <-resultChan
	return result.err
}

// cleanupInactiveWorkers periodically checks for and removes inactive workers
func (p *MultiWorkerPool) cleanupInactiveWorkers() {
	ticker := time.NewTicker(p.WorkerCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.StopChan:
			return
		case <-ticker.C:
			// Use the worker manager to clean up inactive workers
			resultChan := make(chan workerOpResult, 1)
			p.opChan <- workerOp{
				opType:     "cleanup",
				ctx:        p.Ctx,
				resultChan: resultChan,
			}
			<-resultChan // Wait for the operation to complete
		}
	}
}

// cleanupWorkersInternal is the internal implementation of cleanupInactiveWorkers
// It's called by the worker manager routine
func (p *MultiWorkerPool) cleanupWorkersInternal(_ context.Context) {
	idleTimeout := time.Duration(p.WorkerIdleTimeout) * time.Second
	now := time.Now()

	// Find workers to remove
	var workersToRemove []int
	var activeWorkers []MultiConnectionWorkerInterface

	// Identify workers to remove
	for i, worker := range p.workers {
		lastActivity := worker.GetLastActivity()

		// If the worker has been idle for too long, mark it for removal
		if now.Sub(lastActivity) > idleTimeout {
			workersToRemove = append(workersToRemove, i)
		} else {
			activeWorkers = append(activeWorkers, worker)
		}
	}

	// Stop workers that need to be removed
	for _, idx := range workersToRemove {
		p.workers[idx].Stop()
		atomic.AddInt64(&p.Metrics.WorkersRemoved, 1)
	}

	// If we have workers to remove, update the workers slice
	if len(workersToRemove) > 0 {
		p.workers = activeWorkers
	}

	// Clean up connection map entries for non-existent workers
	p.ConnectionMap.Range(func(key, value interface{}) bool {
		workerIdx := value.(int64)
		validIdx := int(workerIdx) < len(p.workers)

		if !validIdx {
			p.ConnectionMap.Delete(key)
		}
		return true
	})
}

// getMetricsInternal is the internal implementation of GetMetrics
// It's called by the worker manager routine
func (p *MultiWorkerPool) getMetricsInternal() map[string]int64 {
	metrics := map[string]int64{
		"workers_created":  atomic.LoadInt64(&p.Metrics.WorkersCreated),
		"workers_removed":  atomic.LoadInt64(&p.Metrics.WorkersRemoved),
		"packets_handled":  atomic.LoadInt64(&p.Metrics.PacketsHandled),
		"packets_dropped":  atomic.LoadInt64(&p.Metrics.PacketsDropped),
		"errors":           atomic.LoadInt64(&p.Metrics.Errors),
		"active_workers":   int64(len(p.workers)),
		"connection_count": int64(p.GetConnectionCount()),
	}

	// Add worker-specific metrics
	for i, worker := range p.workers {
		workerMetrics := worker.GetMetrics()
		metrics[fmt.Sprintf("worker_%d_packets", i)] = workerMetrics.PacketCount
		metrics[fmt.Sprintf("worker_%d_errors", i)] = workerMetrics.ErrorCount
		metrics[fmt.Sprintf("worker_%d_bytes", i)] = workerMetrics.BytesSent
		metrics[fmt.Sprintf("worker_%d_connections", i)] = int64(worker.GetConnectionCount())
	}

	return metrics
}

// GetWorkerCount returns the number of active workers
func (p *MultiWorkerPool) GetWorkerCount() int {
	// Use the worker manager to get the worker count
	resultChan := make(chan workerOpResult, 1)
	p.opChan <- workerOp{
		opType:     "get_count",
		resultChan: resultChan,
	}

	// Wait for the result
	result := <-resultChan
	return result.count
}

// GetConnectionCount returns the total number of connections across all workers
func (p *MultiWorkerPool) GetConnectionCount() int {
	count := 0
	p.ConnectionMap.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// GetMetrics returns the worker pool's metrics
func (p *MultiWorkerPool) GetMetrics() map[string]int64 {
	// Use the worker manager to get the metrics
	resultChan := make(chan workerOpResult, 1)
	p.opChan <- workerOp{
		opType:     "get_metrics",
		resultChan: resultChan,
	}

	// Wait for the result
	result := <-resultChan
	return result.metrics
}

// Close implements io.Closer
func (p *MultiWorkerPool) Close() error {
	p.Stop()
	return nil
}
