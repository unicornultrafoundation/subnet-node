package vpn

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PacketDispatcher manages workers for different destination IP:Port combinations
// and routes packets to the appropriate worker
type PacketDispatcher struct {
	// Service references for accessing service methods
	peerDiscovery PeerDiscoveryService
	stream        StreamService
	config        ConfigService
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
	// Channel to signal dispatcher shutdown
	stopChan chan struct{}
	// Whether the dispatcher is running
	running bool
}

// NewPacketDispatcher creates a new packet dispatcher
func NewPacketDispatcher(service VPNService, workerIdleTimeout int) *PacketDispatcher {
	ctx, cancel := context.WithCancel(context.Background())

	return &PacketDispatcher{
		peerDiscovery:     service,
		stream:            service,
		config:            service,
		ctx:               ctx,
		cancel:            cancel,
		workerIdleTimeout: service.GetWorkerIdleTimeout(),
		stopChan:          make(chan struct{}),
		running:           false,
	}
}

// Start starts the dispatcher and its cleanup routine
func (d *PacketDispatcher) Start() {
	if d.running {
		return
	}

	d.running = true

	// Start the cleanup routine
	go d.cleanupInactiveWorkers()

	log.Infof("Packet dispatcher started")
}

// Stop stops the dispatcher and all its workers
func (d *PacketDispatcher) Stop() {
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
func (d *PacketDispatcher) DispatchPacket(ctx context.Context, syncKey, destIP string, packet []byte) {
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
		// No DoneCh since we don't need to wait for completion
	}

	// Try to add the packet to the worker's queue
	if !worker.EnqueuePacket(packetObj) {
		log.Debugf("Worker channel full for %s, dropping packet", syncKey)
	}
}

// getOrCreateWorker gets an existing worker or creates a new one for the given sync key
func (d *PacketDispatcher) getOrCreateWorker(ctx context.Context, syncKey, destIP string) (*PacketWorker, error) {
	d.workersMu.Lock()
	defer d.workersMu.Unlock()

	// Check if worker already exists
	if workerVal, exists := d.workers.Load(syncKey); exists {
		worker := workerVal.(*PacketWorker)
		// Update last activity time
		worker.UpdateLastActivity()
		return worker, nil
	}

	// Create a new worker
	// Get peer ID for the destination IP
	peerID, err := d.peerDiscovery.GetPeerID(ctx, destIP)
	if err != nil {
		return nil, fmt.Errorf("no peer mapping found for IP %s: %v", destIP, err)
	}

	parsedPeerID, err := ParsePeerID(peerID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse peerid %s: %v", peerID, err)
	}

	// Create a new stream to the peer
	stream, err := d.stream.CreateNewVPNStream(ctx, parsedPeerID)
	if err != nil {
		return nil, fmt.Errorf("failed to create P2P stream: %v", err)
	}

	// Create a new worker context with cancel function
	workerCtx, cancel := context.WithCancel(d.ctx)

	// Create a new worker
	worker := NewPacketWorker(
		syncKey,
		destIP,
		parsedPeerID,
		stream,
		workerCtx,
		cancel,
		d.config.GetWorkerBufferSize(),
	)

	// Store the worker
	d.workers.Store(syncKey, worker)

	// Start the worker with a service adapter
	serviceAdapter := &ServiceAdapter{
		PeerDiscovery: d.peerDiscovery,
		Stream:        d.stream,
		Retry:         d.peerDiscovery.(RetryService),   // Assuming peerDiscovery also implements RetryService
		Metrics:       d.peerDiscovery.(MetricsService), // Assuming peerDiscovery also implements MetricsService
		Config:        d.config,
	}
	worker.Start(serviceAdapter)

	log.Debugf("Created new worker for %s", syncKey)
	return worker, nil
}

// cleanupInactiveWorkers periodically checks for and removes inactive workers
func (d *PacketDispatcher) cleanupInactiveWorkers() {
	ticker := time.NewTicker(1 * time.Minute)
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
				worker := value.(*PacketWorker)

				shouldCleanup := false

				// Check if worker has been idle for too long
				if worker.IsIdle(idleTimeout) {
					log.Debugf("Cleaning up inactive worker for %s", syncKey)
					shouldCleanup = true
				}

				// Check if worker has too many errors
				if worker.ErrorCount > 100 {
					log.Warnf("Cleaning up worker %s due to excessive errors", syncKey)
					shouldCleanup = true
				}

				// Check if worker has processed too few packets (might be stuck)
				if worker.PacketCount > 0 && worker.ErrorCount > worker.PacketCount/2 {
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
func (d *PacketDispatcher) stopAllWorkers() {
	d.workersMu.Lock()
	defer d.workersMu.Unlock()

	d.workers.Range(func(key, value interface{}) bool {
		worker := value.(*PacketWorker)

		// Stop the worker
		worker.Stop()

		// Remove the worker from the map
		d.workers.Delete(key)

		return true
	})
}
