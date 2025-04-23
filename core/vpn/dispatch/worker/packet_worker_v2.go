package worker

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/pool"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
)

var packetWorkerV2Log = logrus.WithField("service", "packet-worker-v2")

// PacketWorkerV2 processes packets for a specific connection using V2 components
type PacketWorkerV2 struct {
	// Core components
	peerID        peer.ID
	connKey       types.ConnectionKey
	streamManager *pool.StreamManagerV2

	// Configuration
	config *PacketWorkerConfig

	// Context for the worker
	ctx    context.Context
	cancel context.CancelFunc

	// Packet channel
	packetChan chan *types.QueuedPacket

	// Lifecycle management
	stopChan chan struct{}
	running  bool
	wg       sync.WaitGroup

	// Idle tracking
	lastActivity int64 // Unix timestamp in nanoseconds

	// Resilience service
	resilienceService *resilience.ResilienceService

	// Metrics
	metrics struct {
		PacketsProcessed int64
		PacketsDropped   int64
		Errors           int64
	}
}

// PacketWorkerConfig contains configuration for a packet worker
type PacketWorkerConfig struct {
	// BufferSize is the size of the packet buffer
	BufferSize int
	// IdleTimeout is the duration after which a worker is considered idle
	IdleTimeout time.Duration
}

// NewPacketWorkerV2 creates a new packet worker using V2 components
func NewPacketWorkerV2(
	parentCtx context.Context,
	peerID peer.ID,
	connKey types.ConnectionKey,
	streamManager *pool.StreamManagerV2,
	config *PacketWorkerConfig,
	resilienceService *resilience.ResilienceService,
) *PacketWorkerV2 {
	ctx, cancel := context.WithCancel(parentCtx)

	// Use default resilience service if none provided
	if resilienceService == nil {
		resilienceService = resilience.NewResilienceService(nil)
	}

	// Determine the buffer size
	bufferSize := config.BufferSize
	if bufferSize <= 0 {
		bufferSize = 100 // Default buffer size
	}

	return &PacketWorkerV2{
		peerID:            peerID,
		connKey:           connKey,
		streamManager:     streamManager,
		config:            config,
		ctx:               ctx,
		cancel:            cancel,
		packetChan:        make(chan *types.QueuedPacket, bufferSize),
		stopChan:          make(chan struct{}),
		running:           false,
		lastActivity:      time.Now().UnixNano(),
		resilienceService: resilienceService,
	}
}

// Start starts the worker
func (w *PacketWorkerV2) Start() {
	if w.running {
		return
	}

	w.running = true

	// Start the packet processor
	w.wg.Add(1)
	go w.processPackets()

	packetWorkerV2Log.WithFields(logrus.Fields{
		"peer_id":  w.peerID.String(),
		"conn_key": w.connKey.String(),
	}).Debug("Packet worker V2 started")
}

// Stop stops the worker
func (w *PacketWorkerV2) Stop() {
	if !w.running {
		return
	}

	// Signal shutdown
	close(w.stopChan)
	w.running = false

	// Cancel the context to stop all operations
	w.cancel()

	// Wait for the processor to finish
	w.wg.Wait()

	// Close the packet channel
	close(w.packetChan)

	packetWorkerV2Log.WithFields(logrus.Fields{
		"peer_id":  w.peerID.String(),
		"conn_key": w.connKey.String(),
	}).Debug("Packet worker V2 stopped")
}

// ProcessPacket processes a packet
func (w *PacketWorkerV2) ProcessPacket(ctx context.Context, destIP string, packet *types.QueuedPacket) error {
	if !w.running {
		return types.ErrWorkerStopped
	}

	// Update last activity time
	atomic.StoreInt64(&w.lastActivity, time.Now().UnixNano())

	// Send the packet to the processor
	select {
	case w.packetChan <- packet:
		return nil
	case <-ctx.Done():
		atomic.AddInt64(&w.metrics.PacketsDropped, 1)
		atomic.AddInt64(&w.metrics.Errors, 1)
		return ctx.Err()
	case <-w.ctx.Done():
		atomic.AddInt64(&w.metrics.PacketsDropped, 1)
		atomic.AddInt64(&w.metrics.Errors, 1)
		return types.ErrWorkerStopped
	default:
		// Channel full, packet dropped
		atomic.AddInt64(&w.metrics.PacketsDropped, 1)
		atomic.AddInt64(&w.metrics.Errors, 1)
		return types.ErrWorkerChannelFull
	}
}

// processPackets processes packets from the packet channel
func (w *PacketWorkerV2) processPackets() {
	defer w.wg.Done()

	for {
		select {
		case packet := <-w.packetChan:
			// Process the packet
			w.processPacketInternal(packet)
		case <-w.stopChan:
			return
		case <-w.ctx.Done():
			return
		}
	}
}

// processPacketInternal processes a single packet
func (w *PacketWorkerV2) processPacketInternal(packet *types.QueuedPacket) {
	// Update last activity time
	atomic.StoreInt64(&w.lastActivity, time.Now().UnixNano())

	// Create a context with timeout for the packet processing
	ctx, cancel := context.WithTimeout(w.ctx, 5*time.Second)
	defer cancel()

	// Use the resilience service to handle retries and circuit breaking
	breakerId := w.resilienceService.FormatWorkerBreakerId(w.connKey.String(), "send_packet")
	err, _ := w.resilienceService.ExecuteWithResilience(ctx, breakerId, func() error {
		// Send the packet through the stream manager
		return w.streamManager.SendPacket(ctx, w.connKey, w.peerID, packet)
	})

	// Handle the result
	if err != nil {
		atomic.AddInt64(&w.metrics.Errors, 1)

		// Signal the error on the done channel if provided
		if packet.DoneCh != nil {
			packet.DoneCh <- err
			close(packet.DoneCh)
		}

		packetWorkerV2Log.WithFields(logrus.Fields{
			"peer_id":  w.peerID.String(),
			"conn_key": w.connKey.String(),
			"error":    err.Error(),
		}).Debug("Failed to send packet")
	} else {
		atomic.AddInt64(&w.metrics.PacketsProcessed, 1)

		// Signal success on the done channel if provided
		if packet.DoneCh != nil {
			packet.DoneCh <- nil
			close(packet.DoneCh)
		}
	}
}

// IsIdle checks if the worker is idle
func (w *PacketWorkerV2) IsIdle(now time.Time) bool {
	lastActivity := time.Unix(0, atomic.LoadInt64(&w.lastActivity))
	return now.Sub(lastActivity) > w.config.IdleTimeout
}

// GetMetrics returns the worker's metrics
func (w *PacketWorkerV2) GetMetrics() map[string]int64 {
	return map[string]int64{
		"packets_processed": atomic.LoadInt64(&w.metrics.PacketsProcessed),
		"packets_dropped":   atomic.LoadInt64(&w.metrics.PacketsDropped),
		"errors":            atomic.LoadInt64(&w.metrics.Errors),
	}
}

// Close implements io.Closer
func (w *PacketWorkerV2) Close() error {
	w.Stop()
	return nil
}
