package packet

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/discovery"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

var log = logrus.WithField("service", "vpn-packet")

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
	peerDiscovery discovery.PeerDiscoveryService
	streamService types.Service
	// Enhanced stream services
	poolService      types.PoolService
	multiplexService types.MultiplexService
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
	// Whether to use multiplexing
	useMultiplexing bool
}

// NewDispatcher creates a new packet dispatcher
func NewDispatcher(
	peerDiscovery discovery.PeerDiscoveryService,
	streamService types.Service,
	workerIdleTimeout int,
	workerCleanupInterval time.Duration,
	workerBufferSize int,
) *Dispatcher {
	ctx, cancel := context.WithCancel(context.Background())

	return &Dispatcher{
		peerDiscovery:         peerDiscovery,
		streamService:         streamService,
		ctx:                   ctx,
		cancel:                cancel,
		workerIdleTimeout:     workerIdleTimeout,
		workerCleanupInterval: workerCleanupInterval,
		workerBufferSize:      workerBufferSize,
		stopChan:              make(chan struct{}),
		running:               false,
		useMultiplexing:       false,
	}
}

// NewEnhancedDispatcher creates a new packet dispatcher with enhanced stream services
func NewEnhancedDispatcher(
	peerDiscovery discovery.PeerDiscoveryService,
	streamService types.Service,
	poolService types.PoolService,
	multiplexService types.MultiplexService,
	workerIdleTimeout int,
	workerCleanupInterval time.Duration,
	workerBufferSize int,
	useMultiplexing bool,
) *Dispatcher {
	ctx, cancel := context.WithCancel(context.Background())

	return &Dispatcher{
		peerDiscovery:         peerDiscovery,
		streamService:         streamService,
		poolService:           poolService,
		multiplexService:      multiplexService,
		ctx:                   ctx,
		cancel:                cancel,
		workerIdleTimeout:     workerIdleTimeout,
		workerCleanupInterval: workerCleanupInterval,
		workerBufferSize:      workerBufferSize,
		stopChan:              make(chan struct{}),
		running:               false,
		useMultiplexing:       useMultiplexing,
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

	log.Infof("Packet dispatcher started")
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

	log.Infof("Packet dispatcher stopped")
}

// DispatchPacket dispatches a packet to the appropriate worker
func (d *Dispatcher) DispatchPacket(ctx context.Context, syncKey, destIP string, packet []byte) {
	// Get or create a worker for this destination
	worker, err := d.getOrCreateWorker(ctx, syncKey, destIP)
	if err != nil {
		log.Debugf("Failed to get worker for %s: %v", syncKey, err)
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
		log.Debugf("Worker channel full for %s, dropping packet", syncKey)
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
		return nil, fmt.Errorf("no peer mapping found for IP %s: %v", destIP, err)
	}

	parsedPeerID, err := discovery.ParsePeerID(peerID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse peerid %s: %v", peerID, err)
	}

	// Create a new worker context with cancel function
	workerCtx, cancel := context.WithCancel(d.ctx)

	// Create a new worker based on available services
	var worker *Worker

	if d.poolService != nil && d.multiplexService != nil && d.useMultiplexing {
		// Use multiplexing if available and enabled
		worker = NewMultiplexedWorker(
			syncKey,
			destIP,
			parsedPeerID,
			d.multiplexService,
			workerCtx,
			cancel,
			d.workerBufferSize,
		)
		log.Debugf("Created new multiplexed worker for %s", syncKey)
	} else if d.poolService != nil {
		// Use stream pooling if available
		worker = NewPooledWorker(
			syncKey,
			destIP,
			parsedPeerID,
			d.poolService,
			workerCtx,
			cancel,
			d.workerBufferSize,
		)
		log.Debugf("Created new pooled worker for %s", syncKey)
	} else {
		// Fall back to direct stream creation
		vpnStream, err := d.streamService.CreateNewVPNStream(ctx, parsedPeerID)
		if err != nil {
			return nil, fmt.Errorf("failed to create P2P stream: %v", err)
		}

		worker = NewWorker(
			syncKey,
			destIP,
			parsedPeerID,
			vpnStream,
			workerCtx,
			cancel,
			d.workerBufferSize,
		)
		log.Debugf("Created new direct worker for %s", syncKey)
	}

	// Store the worker in the map
	d.workers.Store(syncKey, worker)

	// Start the worker
	worker.Start()

	return worker, nil
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
			d.workers.Range(func(key, value interface{}) bool {
				syncKey := key.(string)
				worker := value.(*Worker)

				shouldCleanup := false

				// Check if worker has been idle for too long
				if worker.IsIdle(idleTimeout) {
					log.Debugf("Cleaning up inactive worker for %s", syncKey)
					shouldCleanup = true
				}

				// Check if worker has too many errors
				if worker.GetErrorCount() > 100 {
					log.Warnf("Cleaning up worker %s due to excessive errors", syncKey)
					shouldCleanup = true
				}

				// Check if worker has processed too few packets (might be stuck)
				packetCount := worker.GetPacketCount()
				errorCount := worker.GetErrorCount()
				if packetCount > 0 && errorCount > packetCount/2 {
					log.Warnf("Cleaning up worker %s due to high error rate", syncKey)
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
	d.workers.Range(func(key, value interface{}) bool {
		worker := value.(*Worker)

		// Stop the worker
		worker.Stop()

		// Remove the worker from the map
		d.workers.Delete(key)

		return true
	})
}
