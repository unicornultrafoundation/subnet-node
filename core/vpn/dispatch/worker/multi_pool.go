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
	logger.Info("Worker manager started")

	// Track operation statistics for monitoring
	var opStats struct {
		totalOps        int64
		dispatchOps     int64
		successfulOps   int64
		failedOps       int64
		lastStatsReport time.Time
	}
	opStats.lastStatsReport = time.Now()

	// Create a ticker for periodic stats reporting
	statsTicker := time.NewTicker(1 * time.Minute)
	defer statsTicker.Stop()

	for {
		select {
		case <-p.StopChan:
			logger.Debug("Worker manager stopping")
			return

		case <-statsTicker.C:
			// Report operation statistics periodically
			if atomic.LoadInt64(&opStats.totalOps) > 0 {
				logger.WithFields(logrus.Fields{
					"total_ops":      atomic.LoadInt64(&opStats.totalOps),
					"dispatch_ops":   atomic.LoadInt64(&opStats.dispatchOps),
					"successful_ops": atomic.LoadInt64(&opStats.successfulOps),
					"failed_ops":     atomic.LoadInt64(&opStats.failedOps),
					"duration":       time.Since(opStats.lastStatsReport).String(),
					"active_workers": len(p.workers),
				}).Info("Worker manager statistics")

				// Reset stats
				atomic.StoreInt64(&opStats.totalOps, 0)
				atomic.StoreInt64(&opStats.dispatchOps, 0)
				atomic.StoreInt64(&opStats.successfulOps, 0)
				atomic.StoreInt64(&opStats.failedOps, 0)
				opStats.lastStatsReport = time.Now()
			}

		case op := <-p.opChan:
			// Track operation statistics
			atomic.AddInt64(&opStats.totalOps, 1)

			// Create operation-specific logger
			opLogger := logger.WithFields(logrus.Fields{
				"op_type": op.opType,
			})

			// Add connection details if available
			if op.connKey != "" {
				opLogger = opLogger.WithField("conn_key", string(op.connKey))
			}
			if op.destIP != "" {
				opLogger = opLogger.WithField("dest_ip", op.destIP)
			}

			// Handle the operation
			switch op.opType {
			case "get_worker":
				worker, idx, err := p.getOrCreateWorkerInternal(op.ctx, op.connKey, op.destIP)
				op.resultChan <- workerOpResult{worker: worker, index: idx, err: err}

				if err != nil {
					atomic.AddInt64(&opStats.failedOps, 1)
					opLogger.WithError(err).Warn("Failed to get or create worker")
				} else {
					atomic.AddInt64(&opStats.successfulOps, 1)
					opLogger.WithField("worker_idx", idx).Debug("Successfully got or created worker")
				}

			case "dispatch":
				atomic.AddInt64(&opStats.dispatchOps, 1)

				// Get or create a worker for this connection
				worker, workerIdx, err := p.getOrCreateWorkerInternal(op.ctx, op.connKey, op.destIP)
				if err != nil {
					atomic.AddInt64(&p.Metrics.Errors, 1)
					atomic.AddInt64(&opStats.failedOps, 1)
					opLogger.WithError(err).Warn("Failed to get or create worker for dispatch")
					op.resultChan <- workerOpResult{err: err}
					continue
				}

				// Try to add the packet to the worker's queue
				if !worker.EnqueuePacket(op.packet, op.connKey) {
					atomic.AddInt64(&p.Metrics.PacketsDropped, 1)
					atomic.AddInt64(&opStats.failedOps, 1)
					opLogger.WithField("worker_idx", workerIdx).Warn("Worker queue full, packet dropped")
					op.resultChan <- workerOpResult{err: types.ErrWorkerQueueFull}
					continue
				}

				// Update metrics
				atomic.AddInt64(&p.Metrics.PacketsHandled, 1)
				atomic.AddInt64(&opStats.successfulOps, 1)
				opLogger.WithField("worker_idx", workerIdx).Debug("Successfully dispatched packet to worker")
				op.resultChan <- workerOpResult{err: nil}

			case "get_count":
				op.resultChan <- workerOpResult{count: len(p.workers)}
				atomic.AddInt64(&opStats.successfulOps, 1)

			case "get_metrics":
				metrics := p.getMetricsInternal()
				op.resultChan <- workerOpResult{metrics: metrics}
				atomic.AddInt64(&opStats.successfulOps, 1)

			case "cleanup":
				opLogger.Debug("Starting worker cleanup")
				p.cleanupWorkersInternal(op.ctx)
				op.resultChan <- workerOpResult{}
				atomic.AddInt64(&opStats.successfulOps, 1)
				opLogger.WithField("active_workers", len(p.workers)).Debug("Completed worker cleanup")

			case "stop_all":
				// Stop all workers
				opLogger.Debug("Stopping all workers")
				for i, worker := range p.workers {
					worker.Stop()
					opLogger.WithField("worker_idx", i).Debug("Stopped worker")
				}
				p.workers = make([]MultiConnectionWorkerInterface, 0)
				op.resultChan <- workerOpResult{}
				atomic.AddInt64(&opStats.successfulOps, 1)
				opLogger.Debug("All workers stopped")

			default:
				opLogger.Warn("Unknown operation type")
				op.resultChan <- workerOpResult{err: fmt.Errorf("unknown operation type: %s", op.opType)}
				atomic.AddInt64(&opStats.failedOps, 1)
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

// Note: We've removed the getOrCreateWorker method as it's no longer needed.
// All worker operations are now handled through the worker manager routine.

// getOrCreateWorkerInternal is the internal implementation of getOrCreateWorker
// It's called by the worker manager routine
func (p *MultiWorkerPool) getOrCreateWorkerInternal(
	ctx context.Context,
	connKey types.ConnectionKey,
	destIP string,
) (MultiConnectionWorkerInterface, int64, error) {
	// Check if we already have a worker assigned to this connection key
	if workerIdx, ok := p.ConnectionMap.Load(connKey); ok {
		// Since we're in the worker manager goroutine, we have exclusive access to the workers slice
		// This eliminates the race condition between checking and using the worker
		workerIdxInt := int(workerIdx.(int64))
		if workerIdxInt < len(p.workers) {
			worker := p.workers[workerIdxInt]
			if worker.IsRunning() {
				// Log connection affinity for debugging
				multiPoolLog.WithFields(logrus.Fields{
					"peer_id":    p.PeerID.String(),
					"conn_key":   string(connKey),
					"worker_idx": workerIdxInt,
				}).Debug("Maintaining connection affinity to existing worker")
				return worker, workerIdx.(int64), nil
			}
		}
		// Worker doesn't exist or isn't running, remove the mapping
		p.ConnectionMap.Delete(connKey)
		multiPoolLog.WithFields(logrus.Fields{
			"peer_id":  p.PeerID.String(),
			"conn_key": string(connKey),
		}).Debug("Removed mapping for non-existent or stopped worker")
	}

	// If we have workers, find the least loaded one
	if len(p.workers) > 0 {
		leastLoadScore := float64(^uint(0) >> 1) // Max float64
		leastLoadedIdx := 0
		foundRunningWorker := false
		runningWorkers := 0

		// Define thresholds for creating new workers
		const (
			// Buffer utilization threshold (percentage)
			bufferUtilThreshold = 70
			// Connection count threshold
			connCountThreshold = 10
		)

		// Track if any worker is approaching resource limits
		anyWorkerNearingCapacity := false

		for i, worker := range p.workers {
			if !worker.IsRunning() {
				continue
			}

			runningWorkers++
			foundRunningWorker = true

			// Get worker metrics
			connections := worker.GetConnectionCount()
			bufferUtil := worker.GetBufferUtilization()
			metrics := worker.GetMetrics()

			// Calculate a load score that considers multiple factors
			// This is a weighted score where higher values mean more load
			loadScore := (float64(connections) * 1.0) + (float64(bufferUtil) * 0.5)

			// Check if this worker is nearing capacity
			if bufferUtil > bufferUtilThreshold ||
				connections > connCountThreshold {
				anyWorkerNearingCapacity = true

				// Log detailed metrics when a worker is nearing capacity
				multiPoolLog.WithFields(logrus.Fields{
					"worker_id":          i,
					"peer_id":            p.PeerID.String(),
					"buffer_utilization": bufferUtil,
					"connections":        connections,
					"packets_handled":    metrics.PacketCount,
				}).Info("Worker nearing capacity")
			}

			// Find the worker with the lowest load score
			if loadScore < leastLoadScore {
				leastLoadScore = loadScore
				leastLoadedIdx = i
			}
		}

		// If we found a running worker, decide whether to use it or create a new one
		if foundRunningWorker {
			// Create a new worker if:
			// 1. We have capacity for more workers (haven't reached MaxWorkersPerPeer)
			// 2. At least one worker is nearing capacity
			// 3. MaxWorkersPerPeer is greater than 1 (we're allowed to have multiple workers)
			if runningWorkers < p.MaxWorkersPerPeer &&
				p.MaxWorkersPerPeer > 1 &&
				anyWorkerNearingCapacity &&
				len(p.workers) < p.MaxWorkersPerPeer {
				// Log the decision to create a new worker
				multiPoolLog.WithFields(logrus.Fields{
					"peer_id":         p.PeerID.String(),
					"running_workers": runningWorkers,
					"max_workers":     p.MaxWorkersPerPeer,
				}).Info("Creating new worker due to resource utilization")

				// Create a new worker instead of reusing an existing one
				goto createNewWorker
			}

			// Use the least loaded worker
			worker := p.workers[leastLoadedIdx]
			return worker, int64(leastLoadedIdx), nil
		}
	}

	// Label for creating a new worker
createNewWorker:

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

			// Log worker replacement
			multiPoolLog.WithFields(logrus.Fields{
				"peer_id":       p.PeerID.String(),
				"worker_id":     workerID,
				"worker_idx":    workerIdx,
				"total_workers": len(p.workers),
				"max_workers":   p.MaxWorkersPerPeer,
			}).Info("Replaced worker")

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

	// Log worker creation
	multiPoolLog.WithFields(logrus.Fields{
		"peer_id":       p.PeerID.String(),
		"worker_id":     workerID,
		"worker_idx":    workerIdx,
		"total_workers": len(p.workers),
		"max_workers":   p.MaxWorkersPerPeer,
	}).Info("Created new worker")

	return worker, int64(workerIdx), nil
}

// DispatchPacket dispatches a packet to the appropriate worker
func (p *MultiWorkerPool) DispatchPacket(
	ctx context.Context,
	connKey types.ConnectionKey,
	destIP string,
	packet *types.QueuedPacket,
) error {
	// Add packet tracking information
	logger := multiPoolLog.WithFields(logrus.Fields{
		"peer_id":  p.PeerID.String(),
		"conn_key": string(connKey),
		"dest_ip":  destIP,
	})

	// Use the worker manager to dispatch the packet
	resultChan := make(chan workerOpResult, 1)

	// Create a context with timeout to prevent blocking indefinitely
	dispatchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Send the operation to the worker manager
	select {
	case p.opChan <- workerOp{
		opType:     "dispatch",
		connKey:    connKey,
		destIP:     destIP,
		ctx:        ctx, // Use the original context for the actual operation
		packet:     packet,
		resultChan: resultChan,
	}:
		// Operation sent successfully
	case <-dispatchCtx.Done():
		// Timeout or context cancelled
		atomic.AddInt64(&p.Metrics.PacketsDropped, 1)
		atomic.AddInt64(&p.Metrics.Errors, 1)
		logger.Warn("Timed out while trying to dispatch packet to worker manager")
		return types.ErrWorkerManagerBusy
	}

	// Wait for the result with timeout
	select {
	case result := <-resultChan:
		if result.err != nil {
			logger.WithError(result.err).Debug("Failed to dispatch packet")
		} else {
			logger.Debug("Successfully dispatched packet")
		}
		return result.err
	case <-dispatchCtx.Done():
		// Timeout or context cancelled
		atomic.AddInt64(&p.Metrics.PacketsDropped, 1)
		atomic.AddInt64(&p.Metrics.Errors, 1)
		logger.Warn("Timed out while waiting for dispatch result")
		return types.ErrWorkerManagerTimeout
	}
}

// cleanupInactiveWorkers periodically checks for and removes inactive workers
func (p *MultiWorkerPool) cleanupInactiveWorkers() {
	cleanupTicker := time.NewTicker(p.WorkerCleanupInterval)
	defer cleanupTicker.Stop()

	// Create a separate ticker for status logging (every minute)
	statusTicker := time.NewTicker(1 * time.Minute)
	defer statusTicker.Stop()

	for {
		select {
		case <-p.StopChan:
			return
		case <-cleanupTicker.C:
			// Use the worker manager to clean up inactive workers
			resultChan := make(chan workerOpResult, 1)
			p.opChan <- workerOp{
				opType:     "cleanup",
				ctx:        p.Ctx,
				resultChan: resultChan,
			}
			<-resultChan // Wait for the operation to complete
		case <-statusTicker.C:
			// Log the current status of the worker pool
			metrics := p.GetMetrics()
			multiPoolLog.WithFields(logrus.Fields{
				"peer_id":         p.PeerID.String(),
				"active_workers":  metrics["active_workers"],
				"max_workers":     p.MaxWorkersPerPeer,
				"connections":     metrics["connection_count"],
				"packets_handled": metrics["packets_handled"],
			}).Info("Worker pool status")
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
			multiPoolLog.WithFields(logrus.Fields{
				"peer_id":       p.PeerID.String(),
				"worker_idx":    i,
				"idle_duration": now.Sub(lastActivity).String(),
				"idle_timeout":  idleTimeout.String(),
			}).Info("Marking idle worker for removal")
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
		// Create a map of old indices to new indices for connection remapping
		indexMap := make(map[int]int)
		newIdx := 0
		for oldIdx := 0; oldIdx < len(p.workers); oldIdx++ {
			// Check if this worker is being kept
			isRemoved := false
			for _, removedIdx := range workersToRemove {
				if oldIdx == removedIdx {
					isRemoved = true
					break
				}
			}

			if !isRemoved {
				// This worker is being kept, map its old index to its new index
				indexMap[oldIdx] = newIdx
				newIdx++
			}
		}

		// Update the workers slice
		p.workers = activeWorkers

		// Update connection map with new worker indices
		p.ConnectionMap.Range(func(key, value any) bool {
			oldWorkerIdx := int(value.(int64))

			// Check if this worker is still active
			if newWorkerIdx, exists := indexMap[oldWorkerIdx]; exists {
				// Update the connection map with the new index
				p.ConnectionMap.Store(key, int64(newWorkerIdx))
				multiPoolLog.WithFields(logrus.Fields{
					"conn_key":       key,
					"old_worker_idx": oldWorkerIdx,
					"new_worker_idx": newWorkerIdx,
				}).Debug("Remapped connection to new worker index")
			} else {
				// Worker was removed, delete the connection mapping
				p.ConnectionMap.Delete(key)
				multiPoolLog.WithFields(logrus.Fields{
					"conn_key":   key,
					"worker_idx": oldWorkerIdx,
				}).Debug("Removed connection mapping for removed worker")
			}
			return true
		})
	} else {
		// Even if no workers were removed, still clean up any invalid mappings
		p.ConnectionMap.Range(func(key, value any) bool {
			workerIdx := int(value.(int64))
			validIdx := workerIdx < len(p.workers)

			if !validIdx {
				p.ConnectionMap.Delete(key)
				multiPoolLog.WithFields(logrus.Fields{
					"conn_key":   key,
					"worker_idx": workerIdx,
				}).Debug("Removed connection mapping for invalid worker index")
			}
			return true
		})
	}
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
	p.ConnectionMap.Range(func(_, _ any) bool {
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
