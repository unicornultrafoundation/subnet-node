package packet

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/discovery"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
)

var dispatcherLog = logrus.WithField("service", "vpn-packet")

// DispatcherService defines the interface for packet dispatching
type DispatcherService interface {
	// DispatchPacket dispatches a packet to the appropriate worker
	DispatchPacket(ctx context.Context, syncKey, destIP string, packet []byte)
	// Start starts the dispatcher
	Start()
	// Stop stops the dispatcher
	Stop()
}

// Dispatcher manages workers for different destination IP:Port combinations
type Dispatcher struct {
	// Service references for accessing service methods
	peerDiscovery api.PeerDiscoveryService
	streamService api.StreamService
	// Enhanced stream services
	poolService api.StreamPoolService
	// Context for the dispatcher
	ctx context.Context
	// Cancel function for the dispatcher context
	cancel context.CancelFunc
	// Map of workers for each destination IP:Port
	workers sync.Map
	// Mutex to protect access to workers
	workersMu sync.Mutex
	// Worker idle timeout in seconds
	workerIdleTimeout int
	// Worker cleanup interval
	workerCleanupInterval time.Duration
	// Worker buffer size
	workerBufferSize int
	// Channel to signal dispatcher shutdown
	stopChan chan struct{}
	// Whether the dispatcher is running
	running bool
	// Resilience service
	resilienceService *resilience.ResilienceService
}

// NewDispatcher creates a new packet dispatcher
func NewDispatcher(
	peerDiscovery api.PeerDiscoveryService,
	streamService api.StreamService,
	poolService api.StreamPoolService,
	workerIdleTimeout int,
	workerCleanupInterval time.Duration,
	workerBufferSize int,
	resilienceService *resilience.ResilienceService,
) *Dispatcher {
	ctx, cancel := context.WithCancel(context.Background())

	// Use default resilience service if none provided
	if resilienceService == nil {
		resilienceService = resilience.NewResilienceService(nil)
	}

	return &Dispatcher{
		peerDiscovery:         peerDiscovery,
		streamService:         streamService,
		poolService:           poolService,
		ctx:                   ctx,
		cancel:                cancel,
		workerIdleTimeout:     workerIdleTimeout,
		workerCleanupInterval: workerCleanupInterval,
		workerBufferSize:      workerBufferSize,
		stopChan:              make(chan struct{}),
		running:               false,
		resilienceService:     resilienceService,
	}
}

// Start starts the dispatcher and its cleanup routine
func (d *Dispatcher) Start() {
	if d.running {
		return
	}

	d.running = true

	// Start the cleanup routine
	go d.cleanupInactiveWorkers()

	dispatcherLog.Infof("Packet dispatcher started")
}

// Stop stops the dispatcher and all its workers
func (d *Dispatcher) Stop() {
	if !d.running {
		return
	}

	// Signal the cleanup routine to stop
	close(d.stopChan)

	// Cancel the dispatcher context
	d.cancel()

	// Stop all workers
	d.stopAllWorkers()

	d.running = false

	dispatcherLog.Infof("Packet dispatcher stopped")
}

// Close implements the io.Closer interface
func (d *Dispatcher) Close() error {
	d.Stop()
	return nil
}

// DispatchPacket dispatches a packet to the appropriate worker
func (d *Dispatcher) DispatchPacket(ctx context.Context, syncKey, destIP string, packet []byte) {
	// Get or create a worker for this destination
	worker, err := d.getOrCreateWorker(ctx, syncKey, destIP)
	if err != nil {
		dispatcherLog.Debugf("Failed to get worker for %s: %v", syncKey, err)
		return
	}

	// Create a packet object
	packetObj := &QueuedPacket{
		Ctx:    ctx,
		DestIP: destIP,
		Data:   packet,
	}

	// Try to add the packet to the worker's queue
	if !worker.EnqueuePacket(packetObj) {
		dispatcherLog.Debugf("Worker channel full for %s, dropping packet", syncKey)
	}
}

// DispatchPacketWithCallback dispatches a packet to the appropriate worker and provides a callback channel for the result
func (d *Dispatcher) DispatchPacketWithCallback(ctx context.Context, syncKey, destIP string, packet []byte, doneCh chan error) {
	// Get or create a worker for this destination
	worker, err := d.getOrCreateWorker(ctx, syncKey, destIP)
	if err != nil {
		dispatcherLog.Debugf("Failed to get worker for %s: %v", syncKey, err)
		// Signal the error on the done channel
		if doneCh != nil {
			doneCh <- fmt.Errorf("failed to get worker: %w", err)
			close(doneCh)
		}
		return
	}

	// Create a packet object with the done channel
	packetObj := &QueuedPacket{
		Ctx:    ctx,
		DestIP: destIP,
		Data:   packet,
		DoneCh: doneCh,
	}

	// Try to add the packet to the worker's queue
	if !worker.EnqueuePacket(packetObj) {
		dispatcherLog.Debugf("Worker channel full for %s, dropping packet", syncKey)
		// Signal the error on the done channel
		if doneCh != nil {
			doneCh <- fmt.Errorf("worker queue full")
			close(doneCh)
		}
	}
}

// getOrCreateWorker gets an existing worker or creates a new one for the given sync key
func (d *Dispatcher) getOrCreateWorker(ctx context.Context, syncKey, destIP string) (*Worker, error) {
	// Check if worker already exists (fast path, no lock)
	if workerVal, exists := d.workers.Load(syncKey); exists {
		worker := workerVal.(*Worker)
		// Update last activity time
		worker.UpdateLastActivity()
		return worker, nil
	}

	d.workersMu.Lock()
	defer d.workersMu.Unlock()

	// Check again in case another goroutine created the worker while we were waiting
	if workerVal, exists := d.workers.Load(syncKey); exists {
		worker := workerVal.(*Worker)
		worker.UpdateLastActivity()
		return worker, nil
	}

	// Create a new worker
	// Get peer ID for the destination IP
	peerID, err := d.peerDiscovery.GetPeerID(ctx, destIP)
	if err != nil {
		return nil, fmt.Errorf("no peer mapping found for IP %s: %w", destIP, err)
	}

	parsedPeerID, err := discovery.ParsePeerID(peerID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse peer ID %s for IP %s: %w", peerID, destIP, err)
	}

	// Create a new worker context with cancel function
	workerCtx, cancel := context.WithCancel(d.ctx)

	// Create a new worker based on available services
	var newWorker *Worker

	// Ensure cancel is called on error paths
	defer func() {
		if newWorker == nil {
			// If we're returning without a worker, cancel the context
			cancel()
		}
	}()

	// Create a new worker
	newWorker = NewWorker(
		syncKey,
		destIP,
		parsedPeerID,
		d.poolService,
		workerCtx,
		cancel,
		d.workerBufferSize,
		d.resilienceService.GetCircuitBreakerManager(),
		d.resilienceService.GetRetryManager(),
	)
	dispatcherLog.Debugf("Created new worker for %s", syncKey)

	// Store the worker in the map
	d.workers.Store(syncKey, newWorker)

	// Start the worker
	newWorker.Start()

	return newWorker, nil
}

// cleanupInactiveWorkers periodically checks for and removes inactive workers
func (d *Dispatcher) cleanupInactiveWorkers() {
	ticker := time.NewTicker(d.workerCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-d.stopChan:
			return
		case <-ticker.C:
			idleTimeout := time.Duration(d.workerIdleTimeout) * time.Second

			// Check all workers for inactivity
			d.workers.Range(func(key, value any) bool {
				syncKey := key.(string)
				worker := value.(*Worker)

				shouldCleanup := false

				// Check if worker has been idle for too long
				if worker.IsIdle(idleTimeout) {
					dispatcherLog.Debugf("Cleaning up inactive worker for %s", syncKey)
					shouldCleanup = true
				}

				// Check if worker has too many errors
				if worker.GetErrorCount() > 100 {
					dispatcherLog.Warnf("Cleaning up worker %s due to excessive errors", syncKey)
					shouldCleanup = true
				}

				// Check if worker has processed too few packets (might be stuck)
				packetCount := worker.GetPacketCount()
				errorCount := worker.GetErrorCount()
				if packetCount > 0 && errorCount > packetCount/2 {
					dispatcherLog.Warnf("Cleaning up worker %s due to high error rate", syncKey)
					shouldCleanup = true
				}

				if shouldCleanup {
					// Stop the worker
					worker.Stop()

					// Remove the worker from the map
					d.workers.Delete(syncKey)
				}

				return true
			})
		}
	}
}

// stopAllWorkers stops all active workers
func (d *Dispatcher) stopAllWorkers() {
	d.workers.Range(func(key, value any) bool {
		worker := value.(*Worker)

		// Stop the worker
		worker.Stop()

		// Remove the worker from the map
		d.workers.Delete(key)

		return true
	})
}

// GetWorkerMetrics returns metrics for all active workers
func (d *Dispatcher) GetWorkerMetrics() map[string]WorkerMetrics {
	metrics := make(map[string]WorkerMetrics)

	d.workers.Range(func(key, value any) bool {
		syncKey := key.(string)
		worker := value.(*Worker)

		metrics[syncKey] = WorkerMetrics{
			PacketCount: worker.GetPacketCount(),
			ErrorCount:  worker.GetErrorCount(),
		}

		return true
	})

	return metrics
}
