package worker

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
)

var poolLog = logrus.WithField("service", "worker-pool")

// WorkerPool manages workers for different connection keys
type WorkerPool struct {
	// PeerID is the peer ID this pool is responsible for
	PeerID peer.ID
	// StreamManager provides access to stream management
	StreamManager StreamManagerInterface
	// Workers is a map of connection key to worker
	Workers sync.Map
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

// WorkerPoolConfig contains configuration for the worker pool
type WorkerPoolConfig struct {
	WorkerIdleTimeout     int
	WorkerCleanupInterval time.Duration
	WorkerBufferSize      int
}

// NewWorkerPool creates a new worker pool for a specific peer ID
func NewWorkerPool(
	peerID peer.ID,
	streamManager StreamManagerInterface,
	config *WorkerPoolConfig,
	resilienceService *resilience.ResilienceService,
) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	// Use default resilience service if none provided
	if resilienceService == nil {
		resilienceService = resilience.NewResilienceService(nil)
	}

	return &WorkerPool{
		PeerID:                peerID,
		StreamManager:         streamManager,
		Ctx:                   ctx,
		Cancel:                cancel,
		WorkerIdleTimeout:     config.WorkerIdleTimeout,
		WorkerCleanupInterval: config.WorkerCleanupInterval,
		WorkerBufferSize:      config.WorkerBufferSize,
		StopChan:              make(chan struct{}),
		Running:               false,
		ResilienceService:     resilienceService,
	}
}

// Start starts the worker pool and its cleanup routine
func (p *WorkerPool) Start() {
	if p.Running {
		return
	}

	p.Running = true

	// Start the cleanup routine
	go p.cleanupInactiveWorkers()

	poolLog.WithField("peer_id", p.PeerID.String()).Info("Worker pool started")
}

// Stop stops the worker pool and all its workers
func (p *WorkerPool) Stop() {
	if !p.Running {
		return
	}

	// Signal the cleanup routine to stop
	close(p.StopChan)
	p.Running = false

	// Cancel the context to stop all workers
	p.Cancel()

	// Stop all workers
	p.Workers.Range(func(key, value interface{}) bool {
		worker := value.(*Worker)
		worker.Stop()
		return true
	})

	poolLog.WithField("peer_id", p.PeerID.String()).Info("Worker pool stopped")
}

// GetOrCreateWorker gets or creates a worker for a connection key
func (p *WorkerPool) GetOrCreateWorker(
	ctx context.Context,
	connKey types.ConnectionKey,
	destIP string,
) (*Worker, error) {
	// Check if we already have a worker for this connection key
	if worker, ok := p.Workers.Load(connKey); ok {
		w := worker.(*Worker)
		if w.IsRunning() {
			return w, nil
		}
		// Worker exists but is not running, remove it
		p.Workers.Delete(connKey)
	}

	// Create a new worker
	workerCtx, workerCancel := context.WithCancel(p.Ctx)
	worker := NewWorker(
		connKey,
		destIP,
		p.PeerID,
		p.StreamManager,
		workerCtx,
		workerCancel,
		p.WorkerBufferSize,
		p.ResilienceService,
	)

	// Store the worker
	p.Workers.Store(connKey, worker)

	// Start the worker
	worker.Start()

	// Update metrics
	atomic.AddInt64(&p.Metrics.WorkersCreated, 1)

	return worker, nil
}

// DispatchPacket dispatches a packet to the appropriate worker
func (p *WorkerPool) DispatchPacket(
	ctx context.Context,
	connKey types.ConnectionKey,
	destIP string,
	packet *types.QueuedPacket,
) error {
	// Get or create a worker for this connection key
	worker, err := p.GetOrCreateWorker(ctx, connKey, destIP)
	if err != nil {
		atomic.AddInt64(&p.Metrics.Errors, 1)
		return err
	}

	// Try to add the packet to the worker's queue
	if !worker.EnqueuePacket(packet) {
		atomic.AddInt64(&p.Metrics.PacketsDropped, 1)
		return types.ErrWorkerQueueFull
	}

	// Update metrics
	atomic.AddInt64(&p.Metrics.PacketsHandled, 1)
	return nil
}

// cleanupInactiveWorkers periodically checks for and removes inactive workers
func (p *WorkerPool) cleanupInactiveWorkers() {
	ticker := time.NewTicker(p.WorkerCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.StopChan:
			return
		case <-ticker.C:
			idleTimeout := time.Duration(p.WorkerIdleTimeout) * time.Second
			now := time.Now()

			// Check each worker
			p.Workers.Range(func(key, value interface{}) bool {
				worker := value.(*Worker)
				lastActivity := worker.GetLastActivity()

				// If the worker has been idle for too long, stop it
				if now.Sub(lastActivity) > idleTimeout {
					worker.Stop()
					p.Workers.Delete(key)
					atomic.AddInt64(&p.Metrics.WorkersRemoved, 1)
				}

				return true
			})
		}
	}
}

// GetWorkerCount returns the number of active workers
func (p *WorkerPool) GetWorkerCount() int {
	count := 0
	p.Workers.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// GetMetrics returns the worker pool's metrics
func (p *WorkerPool) GetMetrics() map[string]int64 {
	return map[string]int64{
		"workers_created": atomic.LoadInt64(&p.Metrics.WorkersCreated),
		"workers_removed": atomic.LoadInt64(&p.Metrics.WorkersRemoved),
		"packets_handled": atomic.LoadInt64(&p.Metrics.PacketsHandled),
		"packets_dropped": atomic.LoadInt64(&p.Metrics.PacketsDropped),
		"errors":          atomic.LoadInt64(&p.Metrics.Errors),
		"active_workers":  int64(p.GetWorkerCount()),
	}
}

// Close implements io.Closer
func (p *WorkerPool) Close() error {
	p.Stop()
	return nil
}
