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

// MultiWorkerPool manages a pool of multi-connection workers for a specific peer ID
type MultiWorkerPool struct {
	// PeerID is the peer ID this pool is responsible for
	PeerID peer.ID
	// StreamManager provides access to stream management
	StreamManager StreamManagerInterface
	// Workers is a slice of multi-connection workers
	Workers []MultiConnectionWorkerInterface
	// WorkersMu protects the workers slice
	WorkersMu sync.RWMutex
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
	// Whether the worker pool is running
	Running bool
	// Resilience service
	ResilienceService *resilience.ResilienceService
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

	return &MultiWorkerPool{
		PeerID:                peerID,
		StreamManager:         streamManager,
		Workers:               make([]MultiConnectionWorkerInterface, 0),
		Ctx:                   ctx,
		Cancel:                cancel,
		WorkerIdleTimeout:     config.WorkerIdleTimeout,
		WorkerCleanupInterval: config.WorkerCleanupInterval,
		WorkerBufferSize:      config.WorkerBufferSize,
		MaxWorkersPerPeer:     config.MaxWorkersPerPeer,
		StopChan:              make(chan struct{}),
		Running:               false,
		ResilienceService:     resilienceService,
	}
}

// Start starts the worker pool and its cleanup routine
func (p *MultiWorkerPool) Start() {
	if p.Running {
		return
	}

	p.Running = true

	// Start the cleanup routine
	go p.cleanupInactiveWorkers()

	multiPoolLog.WithField("peer_id", p.PeerID.String()).Info("Multi-worker pool started")
}

// Stop stops the worker pool and all its workers
func (p *MultiWorkerPool) Stop() {
	if !p.Running {
		return
	}

	// Signal the cleanup routine to stop
	close(p.StopChan)
	p.Running = false

	// Cancel the context to stop all workers
	p.Cancel()

	// Stop all workers
	p.WorkersMu.Lock()
	for _, worker := range p.Workers {
		worker.Stop()
	}
	p.Workers = make([]MultiConnectionWorkerInterface, 0)
	p.WorkersMu.Unlock()

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
	// Check if we already have a worker assigned to this connection key
	if workerIdx, ok := p.ConnectionMap.Load(connKey); ok {
		p.WorkersMu.RLock()
		if int(workerIdx.(int64)) < len(p.Workers) {
			worker := p.Workers[workerIdx.(int64)]
			p.WorkersMu.RUnlock()
			if worker.IsRunning() {
				return worker, nil
			}
		} else {
			p.WorkersMu.RUnlock()
		}
		// Worker doesn't exist or isn't running, remove the mapping
		p.ConnectionMap.Delete(connKey)
	}

	// Find the least loaded worker or create a new one
	worker, workerIdx, err := p.findOrCreateLeastLoadedWorker(ctx, destIP)
	if err != nil {
		return nil, err
	}

	// Assign the connection key to this worker
	p.ConnectionMap.Store(connKey, workerIdx)

	return worker, nil
}

// findOrCreateLeastLoadedWorker finds the least loaded worker or creates a new one
func (p *MultiWorkerPool) findOrCreateLeastLoadedWorker(
	ctx context.Context,
	destIP string,
) (MultiConnectionWorkerInterface, int64, error) {
	p.WorkersMu.RLock()
	workerCount := len(p.Workers)

	// If we have workers, find the least loaded one
	if workerCount > 0 {
		leastConnections := int(^uint(0) >> 1) // Max int
		leastLoadedIdx := 0

		for i, worker := range p.Workers {
			if !worker.IsRunning() {
				continue
			}

			connections := worker.GetConnectionCount()
			if connections < leastConnections {
				leastConnections = connections
				leastLoadedIdx = i
			}
		}

		// If we found a running worker and it's not at capacity, use it
		if p.Workers[leastLoadedIdx].IsRunning() {
			worker := p.Workers[leastLoadedIdx]
			p.WorkersMu.RUnlock()
			return worker, int64(leastLoadedIdx), nil
		}
	}
	p.WorkersMu.RUnlock()

	// No suitable worker found, create a new one if we haven't reached the limit
	p.WorkersMu.Lock()
	defer p.WorkersMu.Unlock()

	// Check again in case another goroutine created a worker while we were waiting for the lock
	if len(p.Workers) > 0 {
		for i, worker := range p.Workers {
			if worker.IsRunning() {
				return worker, int64(i), nil
			}
		}
	}

	// Check if we've reached the maximum number of workers
	if len(p.Workers) >= p.MaxWorkersPerPeer {
		// Find any worker that's running, even if it's heavily loaded
		for i, worker := range p.Workers {
			if worker.IsRunning() {
				return worker, int64(i), nil
			}
		}

		// If we get here, all workers are stopped, so we'll replace the first one
		if len(p.Workers) > 0 {
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
			p.Workers[workerIdx] = worker

			// Update metrics
			atomic.AddInt64(&p.Metrics.WorkersCreated, 1)

			return worker, int64(workerIdx), nil
		}

		return nil, 0, fmt.Errorf("maximum number of workers reached for peer %s", p.PeerID.String())
	}

	// Create a new worker
	workerIdx := len(p.Workers)
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
	p.Workers = append(p.Workers, worker)

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
	// Get or create a worker for this connection key
	worker, err := p.getOrCreateWorker(ctx, connKey, destIP)
	if err != nil {
		atomic.AddInt64(&p.Metrics.Errors, 1)
		return err
	}

	// Try to add the packet to the worker's queue
	if !worker.EnqueuePacket(packet, connKey) {
		atomic.AddInt64(&p.Metrics.PacketsDropped, 1)
		return types.ErrWorkerQueueFull
	}

	// Update metrics
	atomic.AddInt64(&p.Metrics.PacketsHandled, 1)
	return nil
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
			idleTimeout := time.Duration(p.WorkerIdleTimeout) * time.Second
			now := time.Now()

			p.WorkersMu.Lock()
			// Check each worker
			for i := 0; i < len(p.Workers); i++ {
				if i >= len(p.Workers) {
					break
				}

				worker := p.Workers[i]
				lastActivity := worker.GetLastActivity()

				// If the worker has been idle for too long, stop it
				if now.Sub(lastActivity) > idleTimeout {
					worker.Stop()

					// Remove the worker from the slice
					p.Workers = append(p.Workers[:i], p.Workers[i+1:]...)
					i-- // Adjust index since we removed an element

					atomic.AddInt64(&p.Metrics.WorkersRemoved, 1)
				}
			}
			p.WorkersMu.Unlock()

			// Clean up connection map entries for non-existent workers
			p.ConnectionMap.Range(func(key, value interface{}) bool {
				workerIdx := value.(int64)

				p.WorkersMu.RLock()
				validIdx := int(workerIdx) < len(p.Workers)
				p.WorkersMu.RUnlock()

				if !validIdx {
					p.ConnectionMap.Delete(key)
				}
				return true
			})
		}
	}
}

// GetWorkerCount returns the number of active workers
func (p *MultiWorkerPool) GetWorkerCount() int {
	p.WorkersMu.RLock()
	defer p.WorkersMu.RUnlock()
	return len(p.Workers)
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
	metrics := map[string]int64{
		"workers_created":  atomic.LoadInt64(&p.Metrics.WorkersCreated),
		"workers_removed":  atomic.LoadInt64(&p.Metrics.WorkersRemoved),
		"packets_handled":  atomic.LoadInt64(&p.Metrics.PacketsHandled),
		"packets_dropped":  atomic.LoadInt64(&p.Metrics.PacketsDropped),
		"errors":           atomic.LoadInt64(&p.Metrics.Errors),
		"active_workers":   int64(p.GetWorkerCount()),
		"connection_count": int64(p.GetConnectionCount()),
	}

	// Add worker-specific metrics
	p.WorkersMu.RLock()
	for i, worker := range p.Workers {
		workerMetrics := worker.GetMetrics()
		metrics[fmt.Sprintf("worker_%d_packets", i)] = workerMetrics.PacketCount
		metrics[fmt.Sprintf("worker_%d_errors", i)] = workerMetrics.ErrorCount
		metrics[fmt.Sprintf("worker_%d_bytes", i)] = workerMetrics.BytesSent
		metrics[fmt.Sprintf("worker_%d_connections", i)] = int64(worker.GetConnectionCount())
	}
	p.WorkersMu.RUnlock()

	return metrics
}

// Close implements io.Closer
func (p *MultiWorkerPool) Close() error {
	p.Stop()
	return nil
}
