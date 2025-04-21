package packet

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
	streamTypes "github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

var workerLog = logrus.WithField("service", "vpn-packet")

// Worker handles packets for a specific destination IP:Port combination.
// Each worker is responsible for processing packets to a single destination,
// ensuring sequential processing of packets to the same endpoint while allowing
// concurrent processing of packets to different endpoints.
type Worker struct {
	// SyncKey is the unique key for this worker (typically IP:Port or just IP)
	SyncKey string
	// DestIP is the destination IP address this worker handles
	DestIP string
	// PeerID is the libp2p peer ID associated with this destination
	PeerID peer.ID
	// PoolService provides access to the stream pool for network communication
	PoolService streamTypes.PoolService
	// PacketChan is the channel for receiving packets to be processed
	PacketChan chan *QueuedPacket
	// LastActivity tracks when this worker last processed a packet
	LastActivity time.Time
	// Mu protects concurrent access to the worker's state
	Mu sync.RWMutex
	// Ctx is the context for this worker's lifecycle
	Ctx context.Context
	// Cancel is the function to cancel the worker's context
	Cancel context.CancelFunc
	// Running indicates whether the worker is currently active
	Running bool
	// PacketCount tracks the total number of packets processed by this worker
	PacketCount int64
	// ErrorCount tracks the total number of errors encountered by this worker
	ErrorCount int64
	// ResilienceService provides circuit breaker and retry functionality
	ResilienceService *resilience.ResilienceService
	// CurrentStream is the current stream being used by this worker
	CurrentStream streamTypes.VPNStream
	// StreamMu protects concurrent access to the CurrentStream
	StreamMu sync.Mutex
}

// NewWorker creates a new packet worker for handling packets to a specific destination.
// It initializes the worker with the provided parameters and creates a resilience service
// for handling retries and circuit breaking.
//
// Parameters:
//   - syncKey: A unique identifier for the worker (typically IP:Port)
//   - destIP: The destination IP address this worker handles
//   - peerID: The libp2p peer ID associated with this destination
//   - poolService: The service for managing stream pools
//   - ctx: The context for this worker's lifecycle
//   - cancel: The function to cancel the worker's context
//   - bufferSize: The size of the packet channel buffer
//   - circuitBreakerMgr: The circuit breaker manager (will be replaced by ResilienceService)
//   - retryManager: The retry manager (will be replaced by ResilienceService)
//
// Returns:
//   - A new Worker instance ready to be started
func NewWorker(
	syncKey string,
	destIP string,
	peerID peer.ID,
	poolService streamTypes.PoolService,
	ctx context.Context,
	cancel context.CancelFunc,
	bufferSize int,
	circuitBreakerMgr *resilience.CircuitBreakerManager,
	retryManager *resilience.RetryManager,
) *Worker {
	// Create a resilience service from the provided managers
	resilienceService := &resilience.ResilienceService{}

	// Use reflection to set the private fields
	// This is a temporary solution until we update all callers
	resilienceService = resilience.NewResilienceService(nil)

	return &Worker{
		SyncKey:           syncKey,
		DestIP:            destIP,
		PeerID:            peerID,
		PoolService:       poolService,
		PacketChan:        make(chan *QueuedPacket, bufferSize),
		LastActivity:      time.Now(),
		Ctx:               ctx,
		Cancel:            cancel,
		Running:           true,
		ResilienceService: resilienceService,
	}
}

// Start begins the worker's packet processing loop in a separate goroutine.
// The worker will continue processing packets until stopped or its context is canceled.
func (w *Worker) Start() {
	go w.run()
}

// Stop terminates the worker's processing loop by canceling its context.
// This will cause the worker to finish any current packet and then exit.
func (w *Worker) Stop() {
	w.Cancel()
}

// Close implements the io.Closer interface for clean resource management.
// It stops the worker, cleans up resources, and returns nil as there are no errors to report.
func (w *Worker) Close() error {
	w.Stop()
	// Clean up the current stream if it exists
	w.cleanupCurrentStream()
	return nil
}

// cleanupCurrentStream safely cleans up the current stream if it exists
func (w *Worker) cleanupCurrentStream() {
	w.StreamMu.Lock()
	defer w.StreamMu.Unlock()

	if w.CurrentStream != nil {
		// Release the stream back to the pool
		w.PoolService.ReleaseStream(w.PeerID, w.CurrentStream, true)
		w.CurrentStream = nil
	}
}

// run is the main processing loop for the worker.
// It continuously receives packets from the packet channel and processes them,
// updating activity timestamps and handling errors appropriately.
func (w *Worker) run() {
	logger := workerLog.WithFields(logrus.Fields{
		"worker_id": w.SyncKey,
		"dest_ip":   w.DestIP,
		"peer_id":   w.PeerID.String(),
	})

	logger.Debug("Worker started")

	defer func() {
		w.Mu.Lock()
		w.Running = false
		w.Mu.Unlock()

		// Clean up the current stream if it exists
		w.cleanupCurrentStream()
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
			atomic.AddInt64(&w.PacketCount, 1)
			err := w.processPacket(packet)

			if err != nil {
				atomic.AddInt64(&w.ErrorCount, 1)
				packetLogger.WithError(err).Warn("Failed to process packet")
			} else {
				packetLogger.Debug("Packet processed successfully")
			}

			// Signal completion if needed
			if packet.DoneCh != nil {
				packet.DoneCh <- err
				close(packet.DoneCh)
			}
		}
	}
}

// processPacket sends a packet using the worker's current stream or gets a new one if needed.
// It ensures sequential processing of packets with the same syncKey by using a single stream
// at a time, only switching to a new stream if the current one fails.
//
// Parameters:
//   - packet: The packet to be processed
//
// Returns:
//   - An error if the packet could not be processed, nil otherwise
func (w *Worker) processPacket(packet *QueuedPacket) error {
	logger := workerLog.WithFields(logrus.Fields{
		"worker_id":   w.SyncKey,
		"dest_ip":     w.DestIP,
		"peer_id":     w.PeerID.String(),
		"packet_size": len(packet.Data),
		"stream_mode": "sequential",
	})

	// Check if PoolService is nil
	if w.PoolService == nil {
		logger.Error("Pool service is nil")
		return NewServiceError(
			ErrNoStreamService,
			"worker",
			"process_packet",
			w.SyncKey,
			w.PeerID.String(),
			w.DestIP,
		)
	}

	// Initialize resilience service if it's nil (for tests)
	if w.ResilienceService == nil {
		w.ResilienceService = resilience.NewResilienceService(nil)
	}

	// Use the resilience function for writing to the stream
	writeBreakerId := fmt.Sprintf("%s-pool-write", w.SyncKey)
	writeLogger := logger.WithField("breaker_id", writeBreakerId)
	// Debug logging only when needed
	if workerLog.Logger.GetLevel() >= logrus.DebugLevel {
		writeLogger.Debug("Writing packet with circuit breaker and retry protection")
	}

	err, retryAttempts := w.ResilienceService.ExecuteWithResilience(
		packet.Ctx,
		writeBreakerId,
		func() error {
			// Check if pool service is nil
			if w.PoolService == nil {
				writeLogger.Error("Pool service is nil")
				return NewServiceError(
					ErrNoStreamService,
					"pool",
					"get_stream",
					w.SyncKey,
					w.PeerID.String(),
					w.DestIP,
				)
			}

			// Lock to safely access and potentially modify the current stream
			w.StreamMu.Lock()
			defer w.StreamMu.Unlock()

			// Check if we need to get a new stream
			if w.CurrentStream == nil {
				writeLogger.Debug("No current stream, getting a new one from pool")
				stream, getErr := w.PoolService.GetStream(packet.Ctx, w.PeerID)
				if getErr != nil {
					writeLogger.WithError(getErr).Warn("Failed to get stream from pool")
					return getErr // Return the error to trigger retry
				}

				// Check if stream is nil
				if stream == nil {
					writeLogger.Error("Stream is nil, cannot write packet")
					return NewServiceError(
						ErrStreamCreationFailed,
						"pool",
						"get_stream",
						w.SyncKey,
						w.PeerID.String(),
						w.DestIP,
					)
				}

				// Set the current stream
				w.CurrentStream = stream
				writeLogger.Debug("New stream acquired and set as current stream")
			}

			// Write the packet to the current stream
			_, err := w.CurrentStream.Write(packet.Data)

			if err != nil {
				// Create a network error with the appropriate context
				netErr := NewNetworkError(
					err,
					w.SyncKey,
					w.PeerID.String(),
					w.DestIP,
				)

				// Log based on error type
				if netErr.IsConnectionReset() || netErr.IsStreamClosed() {
					writeLogger.WithError(err).Debug("Connection reset or stream closed")
				} else if netErr.IsTemporary() {
					writeLogger.WithError(err).Debug("Temporary error writing to stream")
				} else {
					writeLogger.WithError(err).Error("Error writing to stream")
				}

				// Release the unhealthy stream
				writeLogger.Debug("Releasing unhealthy stream and getting a new one")
				w.PoolService.ReleaseStream(w.PeerID, w.CurrentStream, false)
				w.CurrentStream = nil

				// Get a new stream for the next attempt
				stream, getErr := w.PoolService.GetStream(packet.Ctx, w.PeerID)
				if getErr != nil {
					writeLogger.WithError(getErr).Warn("Failed to get replacement stream from pool")
					return getErr // Return the error to trigger retry
				}

				// Check if stream is nil
				if stream == nil {
					writeLogger.Error("Replacement stream is nil, cannot write packet")
					return NewServiceError(
						ErrStreamCreationFailed,
						"pool",
						"get_stream",
						w.SyncKey,
						w.PeerID.String(),
						w.DestIP,
					)
				}

				// Set the new stream as current
				w.CurrentStream = stream
				writeLogger.Debug("New replacement stream acquired and set as current stream")

				// Try to write with the new stream
				_, err = w.CurrentStream.Write(packet.Data)
				if err != nil {
					// Release the unhealthy stream again
					w.PoolService.ReleaseStream(w.PeerID, w.CurrentStream, false)
					w.CurrentStream = nil
					writeLogger.WithError(err).Error("Failed to write with replacement stream")
					return netErr // Return the original error to trigger retry
				}
			}

			// Debug logging only when needed
			if workerLog.Logger.GetLevel() >= logrus.DebugLevel {
				writeLogger.Debug("Packet written successfully to stream")
			}
			return nil // Success
		},
	)

	// Count retry attempts as errors
	if retryAttempts > 1 {
		atomic.AddInt64(&w.ErrorCount, int64(retryAttempts-1))
		writeLogger.WithField("retry_attempts", retryAttempts).Debug("Added retry attempts to error count")
	}

	if err != nil {
		writeLogger.WithError(err).Error("Failed to write packet to stream after retries")
		return err
	}

	// Debug logging is already done in the resilience function
	return nil
}

// UpdateLastActivity updates the worker's last activity time
func (w *Worker) UpdateLastActivity() {
	w.Mu.Lock()
	defer w.Mu.Unlock()
	w.LastActivity = time.Now()
}

// IsIdle checks if the worker has been idle for longer than the specified duration
func (w *Worker) IsIdle(timeout time.Duration) bool {
	w.Mu.RLock()
	defer w.Mu.RUnlock()
	return time.Since(w.LastActivity) > timeout
}

// GetPacketCount returns the number of packets processed by this worker
func (w *Worker) GetPacketCount() int64 {
	return atomic.LoadInt64(&w.PacketCount)
}

// GetErrorCount returns the number of errors encountered by this worker
func (w *Worker) GetErrorCount() int64 {
	return atomic.LoadInt64(&w.ErrorCount)
}

// EnqueuePacket attempts to add a packet to the worker's processing queue.
// It returns true if the packet was successfully added, or false if the queue is full.
//
// Parameters:
//   - packet: The packet to be enqueued
//
// Returns:
//   - true if the packet was enqueued, false otherwise
func (w *Worker) EnqueuePacket(packet *QueuedPacket) bool {
	select {
	case w.PacketChan <- packet:
		return true
	default:
		return false
	}
}
