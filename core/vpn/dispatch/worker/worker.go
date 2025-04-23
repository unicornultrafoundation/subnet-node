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

var workerLog = logrus.WithField("service", "vpn-worker")

// Worker handles packets for a specific connection key and implements the SingleConnectionWorker interface
type Worker struct {
	// ConnectionKey is the unique key for this worker (sourcePort:destinationIP:destinationPort)
	ConnectionKey types.ConnectionKey
	// DestIP is the destination IP address this worker handles
	DestIP string
	// PeerID is the libp2p peer ID associated with this destination
	PeerID peer.ID
	// StreamManager provides access to stream management
	StreamManager StreamManagerInterface
	// PacketChan is the channel for receiving packets to be processed
	PacketChan chan *types.QueuedPacket
	// LastActivity is the timestamp of the last activity
	LastActivity time.Time
	// Ctx is the context for this worker
	Ctx context.Context
	// Cancel is the cancel function for the worker context
	Cancel context.CancelFunc
	// Mu is the mutex for protecting worker state
	Mu sync.RWMutex
	// Running indicates whether the worker is running
	Running bool
	// ResilienceService provides resilience patterns
	ResilienceService *resilience.ResilienceService
	// Metrics for this worker
	Metrics types.WorkerMetrics
}

// NewWorker creates a new worker for a specific connection key
func NewWorker(
	connKey types.ConnectionKey,
	destIP string,
	peerID peer.ID,
	streamManager StreamManagerInterface,
	ctx context.Context,
	cancel context.CancelFunc,
	bufferSize int,
	resilienceService *resilience.ResilienceService,
) *Worker {
	// Use default resilience service if none provided
	if resilienceService == nil {
		resilienceService = resilience.NewResilienceService(nil)
	}

	return &Worker{
		ConnectionKey:     connKey,
		DestIP:            destIP,
		PeerID:            peerID,
		StreamManager:     streamManager,
		PacketChan:        make(chan *types.QueuedPacket, bufferSize),
		LastActivity:      time.Now(),
		Ctx:               ctx,
		Cancel:            cancel,
		Running:           true,
		ResilienceService: resilienceService,
	}
}

// Start begins the worker's packet processing loop
func (w *Worker) Start() {
	go w.run()
}

// Stop terminates the worker's processing loop
func (w *Worker) Stop() {
	w.Cancel()
}

// EnqueuePacket adds a packet to the worker's queue
func (w *Worker) EnqueuePacket(packet *types.QueuedPacket) bool {
	select {
	case w.PacketChan <- packet:
		return true
	default:
		// Channel is full
		return false
	}
}

// run is the main processing loop for the worker
func (w *Worker) run() {
	logger := workerLog.WithFields(logrus.Fields{
		"worker_id": string(w.ConnectionKey),
		"dest_ip":   w.DestIP,
		"peer_id":   w.PeerID.String(),
	})

	logger.Debug("Worker started")

	defer func() {
		w.Mu.Lock()
		w.Running = false
		w.Mu.Unlock()

		logger.Debug("Worker stopped")
	}()

	for {
		select {
		case <-w.Ctx.Done():
			logger.Debug("Worker context cancelled, stopping")
			return
		case packet, ok := <-w.PacketChan:
			if !ok {
				// Channel was closed
				logger.Debug("Packet channel closed, stopping worker")
				return
			}

			w.Mu.Lock()
			// Update last activity time
			w.LastActivity = time.Now()
			w.Mu.Unlock()

			// Process the packet
			packetSize := len(packet.Data)
			packetLogger := logger.WithFields(logrus.Fields{
				"packet_size": packetSize,
				"dest_ip":     packet.DestIP,
			})

			// Debug logging only when needed
			if workerLog.Logger.GetLevel() >= logrus.DebugLevel {
				packetLogger.Debug("Processing packet")
			}

			// Process the packet with resilience patterns
			err := w.processPacket(packet)
			if err != nil {
				packetLogger.WithError(err).Warn("Failed to process packet")

				// Signal the error on the done channel if provided
				if packet.DoneCh != nil {
					packet.DoneCh <- err
					close(packet.DoneCh)
				}

				// Update error metrics
				atomic.AddInt64(&w.Metrics.ErrorCount, 1)
			} else {
				// Signal success on the done channel if provided
				if packet.DoneCh != nil {
					packet.DoneCh <- nil
					close(packet.DoneCh)
				}

				// Update packet metrics
				atomic.AddInt64(&w.Metrics.PacketCount, 1)
				atomic.AddInt64(&w.Metrics.BytesSent, int64(packetSize))
			}
		}
	}
}

// processPacket processes a packet with resilience patterns
func (w *Worker) processPacket(packet *types.QueuedPacket) error {
	// Create a breaker ID for this operation
	breakerId := w.ResilienceService.FormatPeerBreakerId(w.PeerID, "send_packet")

	// Execute with resilience patterns
	err, _ := w.ResilienceService.ExecuteWithResilience(
		packet.Ctx,
		breakerId,
		func() error {
			// Send the packet through the stream manager
			return w.StreamManager.SendPacket(packet.Ctx, w.ConnectionKey, w.PeerID, packet)
		},
	)

	return err
}

// IsRunning returns whether the worker is running
func (w *Worker) IsRunning() bool {
	w.Mu.RLock()
	defer w.Mu.RUnlock()
	return w.Running
}

// GetLastActivity returns the timestamp of the last activity
func (w *Worker) GetLastActivity() time.Time {
	w.Mu.RLock()
	defer w.Mu.RUnlock()
	return w.LastActivity
}

// GetMetrics returns the worker's metrics
func (w *Worker) GetMetrics() types.WorkerMetrics {
	return types.WorkerMetrics{
		PacketCount: atomic.LoadInt64(&w.Metrics.PacketCount),
		ErrorCount:  atomic.LoadInt64(&w.Metrics.ErrorCount),
		BytesSent:   atomic.LoadInt64(&w.Metrics.BytesSent),
	}
}

// Close implements io.Closer
func (w *Worker) Close() error {
	w.Stop()
	return nil
}
