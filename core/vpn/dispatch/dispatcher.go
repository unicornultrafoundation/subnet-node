package dispatch

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/pool"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
)

var dispatcherLog = logrus.WithField("service", "vpn-dispatcher")

// Dispatcher manages packet routing directly to stream channels
type Dispatcher struct {
	// Service references
	peerDiscovery api.PeerDiscoveryService
	streamService api.StreamService

	// Core components
	streamPool    *pool.StreamPool
	streamManager *pool.StreamManager

	// Context for the dispatcher
	ctx    context.Context
	cancel context.CancelFunc

	// Configuration
	config *Config

	// Lifecycle management
	stopChan chan struct{}
	running  bool

	// Resilience service
	resilienceService *resilience.ResilienceService

	// Metrics
	metrics struct {
		PacketsDispatched int64
		PacketsDropped    int64
		Errors            int64
	}
}

// NewDispatcher creates a new packet dispatcher using V2 components
func NewDispatcher(
	peerDiscovery api.PeerDiscoveryService,
	streamService api.StreamService,
	config *Config,
	resilienceService *resilience.ResilienceService,
) *Dispatcher {
	ctx, cancel := context.WithCancel(context.Background())

	// Use default resilience service if none provided
	if resilienceService == nil {
		resilienceService = resilience.NewResilienceService(nil)
	}

	// Create stream pool
	streamPoolConfig := &pool.StreamPoolConfig{
		MaxStreamsPerPeer:   config.MaxStreamsPerPeer,
		StreamIdleTimeout:   config.StreamIdleTimeout,
		CleanupInterval:     config.StreamCleanupInterval,
		PacketBufferSize:    config.PacketBufferSize,
		UsageCountWeight:    config.UsageCountWeight,
		BufferUtilWeight:    config.BufferUtilWeight,
		BufferUtilThreshold: config.BufferUtilThreshold,
		UsageCountThreshold: config.UsageCountThreshold,
	}
	streamPool := pool.NewStreamPool(streamService, streamPoolConfig)

	// Create stream manager
	streamManager := pool.NewStreamManager(streamPool)

	return &Dispatcher{
		peerDiscovery:     peerDiscovery,
		streamService:     streamService,
		streamPool:        streamPool,
		streamManager:     streamManager,
		ctx:               ctx,
		cancel:            cancel,
		config:            config,
		stopChan:          make(chan struct{}),
		running:           false,
		resilienceService: resilienceService,
	}
}

// Start starts the dispatcher and its components
func (d *Dispatcher) Start() {
	if d.running {
		return
	}

	d.running = true

	// Start the stream pool
	d.streamPool.Start()

	// Start the stream manager
	d.streamManager.Start()

	dispatcherLog.Info("Packet dispatcher started")
}

// Stop stops the dispatcher and its components
func (d *Dispatcher) Stop() {
	if !d.running {
		return
	}

	// Signal shutdown
	close(d.stopChan)
	d.running = false

	// Cancel the context to stop all operations
	d.cancel()

	// Stop the stream manager
	d.streamManager.Stop()

	// Stop the stream pool
	d.streamPool.Stop()

	dispatcherLog.Info("Packet dispatcher stopped")
}

// DispatchPacket dispatches a packet to the appropriate worker
func (d *Dispatcher) DispatchPacket(
	ctx context.Context,
	connKey types.ConnectionKey,
	destIP string,
	packet []byte,
) error {
	if !d.running {
		return types.ErrDispatcherStopped
	}

	// Create a packet object
	packetObj := &types.QueuedPacket{
		Ctx:    ctx,
		DestIP: destIP,
		Data:   packet,
	}

	// Dispatch the packet
	err := d.dispatchPacketInternal(ctx, connKey, destIP, packetObj)
	if err != nil {
		atomic.AddInt64(&d.metrics.PacketsDropped, 1)
		atomic.AddInt64(&d.metrics.Errors, 1)
		return err
	}

	atomic.AddInt64(&d.metrics.PacketsDispatched, 1)
	return nil
}

// DispatchPacketWithCallback dispatches a packet and provides a callback channel for the result
func (d *Dispatcher) DispatchPacketWithCallback(
	ctx context.Context,
	connKey types.ConnectionKey,
	destIP string,
	packet []byte,
	doneCh chan error,
) error {
	if !d.running {
		if doneCh != nil {
			doneCh <- types.ErrDispatcherStopped
			close(doneCh)
		}
		return types.ErrDispatcherStopped
	}

	// Create a packet object with the done channel
	packetObj := &types.QueuedPacket{
		Ctx:    ctx,
		DestIP: destIP,
		Data:   packet,
		DoneCh: doneCh,
	}

	// Dispatch the packet
	err := d.dispatchPacketInternal(ctx, connKey, destIP, packetObj)
	if err != nil {
		atomic.AddInt64(&d.metrics.PacketsDropped, 1)
		atomic.AddInt64(&d.metrics.Errors, 1)

		// Signal the error on the done channel if provided
		if doneCh != nil {
			doneCh <- err
			close(doneCh)
		}

		return err
	}

	atomic.AddInt64(&d.metrics.PacketsDispatched, 1)
	return nil
}

// dispatchPacketInternal handles the actual packet dispatching logic
func (d *Dispatcher) dispatchPacketInternal(
	ctx context.Context,
	connKey types.ConnectionKey,
	destIP string,
	packet *types.QueuedPacket,
) error {
	// Get peer ID for the destination IP
	peerID, err := d.getPeerIDForDestIP(ctx, destIP)
	if err != nil {
		return err
	}

	// Send the packet directly to the stream manager
	return d.streamManager.SendPacket(ctx, connKey, peerID, packet)
}

// getPeerIDForDestIP gets the peer ID for a destination IP
func (d *Dispatcher) getPeerIDForDestIP(ctx context.Context, destIP string) (peer.ID, error) {
	// Get peer ID for the destination IP
	peerIDStr, err := d.peerDiscovery.GetPeerID(ctx, destIP)
	if err != nil {
		return "", fmt.Errorf("no peer mapping found for IP %s: %w", destIP, err)
	}

	// Parse the peer ID
	peerID, err := peer.Decode(peerIDStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse peer ID %s for IP %s: %w", peerIDStr, destIP, err)
	}

	return peerID, nil
}

// GetMetrics returns the dispatcher's metrics
func (d *Dispatcher) GetMetrics() map[string]int64 {
	metrics := map[string]int64{
		"packets_dispatched": atomic.LoadInt64(&d.metrics.PacketsDispatched),
		"packets_dropped":    atomic.LoadInt64(&d.metrics.PacketsDropped),
		"errors":             atomic.LoadInt64(&d.metrics.Errors),
		"total_streams":      int64(d.streamPool.GetTotalStreamCount()),
		"total_connections":  int64(d.streamManager.GetConnectionCount()),
	}

	// Add stream manager metrics
	for k, v := range d.streamManager.GetMetrics() {
		metrics["stream_manager_"+k] = v
	}

	return metrics
}

// Close implements io.Closer
func (d *Dispatcher) Close() error {
	d.Stop()
	return nil
}
