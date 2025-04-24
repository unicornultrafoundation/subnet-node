package worker

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/pool"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
)

var multiWorkerLog = logrus.WithField("service", "vpn-multi-worker")

// ConnectionState tracks the state of a connection
type ConnectionState struct {
	// The connection key
	Key types.ConnectionKey
	// Last activity time
	LastActivity time.Time
	// Packet count for this connection
	PacketCount int64
	// Error count for this connection
	ErrorCount int64
	// Bytes sent for this connection
	BytesSent int64
}

// MultiConnectionWorker handles packets for multiple connection keys to the same destination
type MultiConnectionWorker struct {
	// WorkerID is a unique identifier for this worker
	WorkerID string
	// DestIP is the destination IP address this worker handles
	DestIP string
	// PeerID is the libp2p peer ID associated with this destination
	PeerID peer.ID
	// StreamManager provides access to stream management
	StreamManager StreamManagerInterface
	// PacketChan is the channel for receiving packets to be processed
	PacketChan chan *types.QueuedPacket
	// Ctx is the context for this worker
	Ctx context.Context
	// Cancel is the cancel function for the worker context
	Cancel context.CancelFunc
	// Running indicates whether the worker is running (0 = not running, 1 = running)
	running int32
	// ResilienceService provides resilience patterns
	ResilienceService *resilience.ResilienceService
	// Metrics for this worker
	Metrics types.WorkerMetrics
	// Connections maps connection keys to their state (using sync.Map for concurrent access)
	connections sync.Map
	// LastActivity is the timestamp of the last activity (Unix nano time)
	lastActivity int64
}

// NewMultiConnectionWorker creates a new worker that handles multiple connection keys
func NewMultiConnectionWorker(
	workerID string,
	destIP string,
	peerID peer.ID,
	streamManager StreamManagerInterface,
	ctx context.Context,
	cancel context.CancelFunc,
	bufferSize int,
	resilienceService *resilience.ResilienceService,
) *MultiConnectionWorker {
	// Use default resilience service if none provided
	if resilienceService == nil {
		resilienceService = resilience.NewResilienceService(nil)
	}

	worker := &MultiConnectionWorker{
		WorkerID:          workerID,
		DestIP:            destIP,
		PeerID:            peerID,
		StreamManager:     streamManager,
		PacketChan:        make(chan *types.QueuedPacket, bufferSize),
		Ctx:               ctx,
		Cancel:            cancel,
		ResilienceService: resilienceService,
	}

	// Initialize atomic values
	atomic.StoreInt32(&worker.running, 1)
	atomic.StoreInt64(&worker.lastActivity, time.Now().UnixNano())

	return worker
}

// Start begins the worker's packet processing loop
func (w *MultiConnectionWorker) Start() {
	go w.run()
}

// Stop terminates the worker's processing loop
func (w *MultiConnectionWorker) Stop() {
	atomic.StoreInt32(&w.running, 0)
	w.Cancel()
}

// EnqueuePacket adds a packet to the worker's queue
func (w *MultiConnectionWorker) EnqueuePacket(packet *types.QueuedPacket, connKey types.ConnectionKey) bool {
	// Update the connection state or create a new one
	now := time.Now()

	// Get existing connection state or create a new one
	value, _ := w.connections.LoadOrStore(connKey, &ConnectionState{
		Key:          connKey,
		LastActivity: now,
	})

	// Update last activity time
	connState := value.(*ConnectionState)
	connState.LastActivity = now

	// Try to enqueue the packet
	select {
	case w.PacketChan <- packet:
		return true
	default:
		// Channel is full
		return false
	}
}

// run is the main processing loop for the worker
func (w *MultiConnectionWorker) run() {
	logger := multiWorkerLog.WithFields(logrus.Fields{
		"worker_id": w.WorkerID,
		"dest_ip":   w.DestIP,
		"peer_id":   w.PeerID.String(),
	})

	logger.Debug("Multi-connection worker started")

	// Add panic recovery to prevent worker crashes
	defer func() {
		if r := recover(); r != nil {
			logger.WithField("panic", r).Error("Recovered from panic in worker run loop")
		}
		atomic.StoreInt32(&w.running, 0)
		logger.Debug("Multi-connection worker stopped")
	}()

	// Track worker statistics
	var stats struct {
		packetsProcessed int64
		packetsDropped   int64
		errors           int64
		lastStatsReport  time.Time
	}
	stats.lastStatsReport = time.Now()

	// Create a ticker for periodic stats reporting
	statsTicker := time.NewTicker(1 * time.Minute)
	defer statsTicker.Stop()

	for {
		select {
		case <-w.Ctx.Done():
			logger.Debug("Worker context cancelled, stopping")
			return

		case <-statsTicker.C:
			// Report worker statistics periodically
			if atomic.LoadInt64(&stats.packetsProcessed) > 0 || atomic.LoadInt64(&stats.errors) > 0 {
				// Get connection count
				connCount := 0
				w.connections.Range(func(_, _ any) bool {
					connCount++
					return true
				})

				logger.WithFields(logrus.Fields{
					"packets_processed": atomic.LoadInt64(&stats.packetsProcessed),
					"packets_dropped":   atomic.LoadInt64(&stats.packetsDropped),
					"errors":            atomic.LoadInt64(&stats.errors),
					"connections":       connCount,
					"duration":          time.Since(stats.lastStatsReport).String(),
					"buffer_util":       w.GetBufferUtilization(),
				}).Info("Worker statistics")

				// Reset stats
				atomic.StoreInt64(&stats.packetsProcessed, 0)
				atomic.StoreInt64(&stats.packetsDropped, 0)
				atomic.StoreInt64(&stats.errors, 0)
				stats.lastStatsReport = time.Now()
			}

		case packet, ok := <-w.PacketChan:
			if !ok {
				// Channel was closed
				logger.Debug("Packet channel closed, stopping worker")
				return
			}

			// Update last activity time
			atomic.StoreInt64(&w.lastActivity, time.Now().UnixNano())

			// Extract the connection key from the packet
			connKey, err := w.extractConnectionKey(packet)
			if err != nil {
				logger.WithError(err).Warn("Failed to extract connection key from packet")

				// Signal the error on the done channel if provided
				if packet.DoneCh != nil {
					select {
					case packet.DoneCh <- err:
						// Successfully sent the error
					default:
						// Channel might be closed or full, don't block
						logger.Debug("Could not send error to done channel, it might be closed or full")
					}
					// Don't close the channel here, let the caller close it
				}

				// Update error metrics
				atomic.AddInt64(&w.Metrics.ErrorCount, 1)
				atomic.AddInt64(&stats.errors, 1)
				atomic.AddInt64(&stats.packetsDropped, 1)
				continue
			}

			// Process the packet
			packetSize := len(packet.Data)
			packetLogger := logger.WithFields(logrus.Fields{
				"packet_size": packetSize,
				"dest_ip":     packet.DestIP,
				"conn_key":    string(connKey),
			})

			// Debug logging only when needed
			if multiWorkerLog.Logger.GetLevel() >= logrus.DebugLevel {
				packetLogger.Debug("Processing packet")
			}

			// Process the packet with resilience patterns
			err = w.processPacket(packet, connKey)
			if err != nil {
				packetLogger.WithError(err).Warn("Failed to process packet")

				// Signal the error on the done channel if provided
				if packet.DoneCh != nil {
					select {
					case packet.DoneCh <- err:
						// Successfully sent the error
					default:
						// Channel might be closed or full, don't block
						packetLogger.Debug("Could not send error to done channel, it might be closed or full")
					}
					// Don't close the channel here, let the caller close it
				}

				// Update error metrics
				atomic.AddInt64(&w.Metrics.ErrorCount, 1)
				atomic.AddInt64(&stats.errors, 1)
				atomic.AddInt64(&stats.packetsDropped, 1)

				// Update connection-specific metrics
				if value, ok := w.connections.Load(connKey); ok {
					connState := value.(*ConnectionState)
					atomic.AddInt64(&connState.ErrorCount, 1)
				}
			} else {
				// Signal success on the done channel if provided
				if packet.DoneCh != nil {
					select {
					case packet.DoneCh <- nil:
						// Successfully sent the result
					default:
						// Channel might be closed or full, don't block
						packetLogger.Debug("Could not send success to done channel, it might be closed or full")
					}
					// Don't close the channel here, let the caller close it
				}

				// Update packet metrics
				atomic.AddInt64(&w.Metrics.PacketCount, 1)
				atomic.AddInt64(&w.Metrics.BytesSent, int64(packetSize))
				atomic.AddInt64(&stats.packetsProcessed, 1)

				// Update connection-specific metrics
				if value, ok := w.connections.Load(connKey); ok {
					connState := value.(*ConnectionState)
					atomic.AddInt64(&connState.PacketCount, 1)
					atomic.AddInt64(&connState.BytesSent, int64(packetSize))
				}
			}
		}
	}
}

// extractConnectionKey extracts the connection key from a packet
func (w *MultiConnectionWorker) extractConnectionKey(packet *types.QueuedPacket) (types.ConnectionKey, error) {
	// Try to extract connection information from the packet
	packetInfo, err := types.ExtractPacketInfo(packet.Data)
	if err != nil {
		return "", err
	}

	// Create a connection key based on available information
	// If ports are missing, use default values
	srcPort := 0
	dstPort := 0

	// Use actual ports if available
	if packetInfo.SrcPort != nil {
		srcPort = *packetInfo.SrcPort
	}

	if packetInfo.DstPort != nil {
		dstPort = *packetInfo.DstPort
	}

	// Format the connection key
	return types.FormatConnectionKey(srcPort, packet.DestIP, dstPort), nil
}

// processPacket processes a packet with resilience patterns
func (w *MultiConnectionWorker) processPacket(packet *types.QueuedPacket, connKey types.ConnectionKey) error {
	// Create a logger for this operation
	logger := multiWorkerLog.WithFields(logrus.Fields{
		"worker_id": w.WorkerID,
		"peer_id":   w.PeerID.String(),
		"conn_key":  string(connKey),
		"dest_ip":   packet.DestIP,
	})

	// Create a breaker ID for this operation
	breakerId := w.ResilienceService.FormatPeerBreakerId(w.PeerID, "send_packet")

	// Record start time for performance monitoring
	startTime := time.Now()

	// Execute with resilience patterns using the original context
	err, retryAttempts := w.ResilienceService.ExecuteWithResilience(
		packet.Ctx, // Use the original context without any deadline
		breakerId,
		func() error {
			// Send the packet through the stream manager
			sendErr := w.StreamManager.SendPacket(packet.Ctx, connKey, w.PeerID, packet)
			if sendErr != nil {
				// Log detailed error information
				logger.WithError(sendErr).Debug("Error sending packet, will retry if possible")

				// Check if this is a stream-related error that should trigger a new stream
				if errors.Is(sendErr, types.ErrStreamChannelClosed) ||
					errors.Is(sendErr, types.ErrNoHealthyStreams) ||
					errors.Is(sendErr, types.ErrStreamWriteFailed) ||
					// Check for common libp2p stream errors
					strings.Contains(sendErr.Error(), "stream reset") ||
					strings.Contains(sendErr.Error(), "protocol not supported") ||
					strings.Contains(sendErr.Error(), "connection closed") ||
					strings.Contains(sendErr.Error(), "stream closed") ||
					strings.Contains(sendErr.Error(), "deadline exceeded") ||
					strings.Contains(sendErr.Error(), "EOF") {
					// Get the stream manager implementation
					if sm, ok := w.StreamManager.(*pool.StreamManager); ok {
						// Release the connection to force getting a new stream on retry
						logger.WithError(sendErr).Debug("Releasing connection due to stream error")
						sm.ReleaseConnection(connKey, false)
					}
				}
			}
			return sendErr
		},
	)

	// Calculate processing duration
	processingDuration := time.Since(startTime)

	// Log performance information
	if processingDuration > 500*time.Millisecond {
		logger.WithFields(logrus.Fields{
			"duration_ms":    processingDuration.Milliseconds(),
			"retry_attempts": retryAttempts,
		}).Warn("Packet processing took longer than expected")
	} else if retryAttempts > 1 {
		logger.WithFields(logrus.Fields{
			"duration_ms":    processingDuration.Milliseconds(),
			"retry_attempts": retryAttempts,
		}).Debug("Packet required retries")
	}

	return err
}

// IsRunning returns whether the worker is running
func (w *MultiConnectionWorker) IsRunning() bool {
	return atomic.LoadInt32(&w.running) == 1
}

// GetLastActivity returns the timestamp of the last activity
func (w *MultiConnectionWorker) GetLastActivity() time.Time {
	return time.Unix(0, atomic.LoadInt64(&w.lastActivity))
}

// GetConnectionCount returns the number of connections this worker is handling
func (w *MultiConnectionWorker) GetConnectionCount() int {
	count := 0
	w.connections.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// GetMetrics returns the worker's metrics
func (w *MultiConnectionWorker) GetMetrics() types.WorkerMetrics {
	return types.WorkerMetrics{
		PacketCount: atomic.LoadInt64(&w.Metrics.PacketCount),
		ErrorCount:  atomic.LoadInt64(&w.Metrics.ErrorCount),
		BytesSent:   atomic.LoadInt64(&w.Metrics.BytesSent),
	}
}

// GetBufferUtilization returns the current buffer utilization as a percentage (0-100)
func (w *MultiConnectionWorker) GetBufferUtilization() int {
	// Get the current length of the packet channel
	currentLen := len(w.PacketChan)

	// Get the capacity of the packet channel
	capacity := cap(w.PacketChan)

	// Calculate utilization percentage
	if capacity == 0 {
		return 0
	}

	return (currentLen * 100) / capacity
}

// GetConnectionMetrics returns metrics for all connections
func (w *MultiConnectionWorker) GetConnectionMetrics() map[string]types.WorkerMetrics {
	metrics := make(map[string]types.WorkerMetrics)

	w.connections.Range(func(key, value interface{}) bool {
		connKey := key.(types.ConnectionKey)
		connState := value.(*ConnectionState)

		metrics[string(connKey)] = types.WorkerMetrics{
			PacketCount: atomic.LoadInt64(&connState.PacketCount),
			ErrorCount:  atomic.LoadInt64(&connState.ErrorCount),
			BytesSent:   atomic.LoadInt64(&connState.BytesSent),
		}

		return true
	})

	return metrics
}

// Close implements io.Closer
func (w *MultiConnectionWorker) Close() error {
	w.Stop()
	return nil
}
