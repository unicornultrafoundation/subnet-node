package packet

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

// Worker implementation is in types.go

// NewWorker creates a new packet worker with a direct stream
func NewWorker(
	syncKey string,
	destIP string,
	peerID peer.ID,
	stream types.VPNStream,
	ctx context.Context,
	cancel context.CancelFunc,
	bufferSize int,
) *Worker {
	return &Worker{
		SyncKey:      syncKey,
		DestIP:       destIP,
		PeerID:       peerID,
		Stream:       stream,
		PacketChan:   make(chan *QueuedPacket, bufferSize),
		LastActivity: time.Now(),
		Ctx:          ctx,
		Cancel:       cancel,
		Running:      true,
	}
}

// NewPooledWorker creates a new packet worker that uses stream pooling
func NewPooledWorker(
	syncKey string,
	destIP string,
	peerID peer.ID,
	poolService types.PoolService,
	ctx context.Context,
	cancel context.CancelFunc,
	bufferSize int,
) *Worker {
	return &Worker{
		SyncKey:      syncKey,
		DestIP:       destIP,
		PeerID:       peerID,
		PoolService:  poolService,
		PacketChan:   make(chan *QueuedPacket, bufferSize),
		LastActivity: time.Now(),
		Ctx:          ctx,
		Cancel:       cancel,
		Running:      true,
		UsePooling:   true,
	}
}

// NewMultiplexedWorker creates a new packet worker that uses stream multiplexing
func NewMultiplexedWorker(
	syncKey string,
	destIP string,
	peerID peer.ID,
	multiplexService types.MultiplexService,
	ctx context.Context,
	cancel context.CancelFunc,
	bufferSize int,
) *Worker {
	return &Worker{
		SyncKey:          syncKey,
		DestIP:           destIP,
		PeerID:           peerID,
		MultiplexService: multiplexService,
		PacketChan:       make(chan *QueuedPacket, bufferSize),
		LastActivity:     time.Now(),
		Ctx:              ctx,
		Cancel:           cancel,
		Running:          true,
		UseMultiplexing:  true,
	}
}

// Start starts the worker's processing loop
func (w *Worker) Start() {
	go w.run()
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.Cancel()
}

// run processes packets for the worker
func (w *Worker) run() {
	defer func() {
		w.Mu.Lock()
		w.Running = false
		w.Mu.Unlock()

		// Clean up based on the worker mode
		if w.UseMultiplexing && w.MultiplexService != nil {
			// No need to close anything for multiplexed mode
			// The multiplexer manager handles stream lifecycle
		} else if w.UsePooling && w.PoolService != nil {
			// No need to close anything for pooled mode
			// The pool manager handles stream lifecycle
		} else if w.Stream != nil {
			// Close the direct stream
			w.Stream.Close()
		}

		log.Debugf("Worker for %s stopped", w.SyncKey)
	}()

	for {
		select {
		case <-w.Ctx.Done():
			return
		case packet, ok := <-w.PacketChan:
			if !ok {
				// Channel was closed
				return
			}

			w.Mu.Lock()
			// Update last activity time
			w.LastActivity = time.Now()
			w.Mu.Unlock()

			// Process the packet
			atomic.AddInt64(&w.PacketCount, 1)
			err := w.processPacket(packet)

			if err != nil {
				atomic.AddInt64(&w.ErrorCount, 1)
			}

			// Signal completion if needed
			if packet.DoneCh != nil {
				packet.DoneCh <- err
				close(packet.DoneCh)
			}
		}
	}
}

// processPacket sends a packet using the worker's stream or service
func (w *Worker) processPacket(packet *QueuedPacket) error {
	// Handle based on the worker mode
	if w.UseMultiplexing && w.MultiplexService != nil {
		// Use multiplexing
		return w.processPacketMultiplexed(packet)
	} else if w.UsePooling && w.PoolService != nil {
		// Use stream pooling
		return w.processPacketPooled(packet)
	} else {
		// Use direct stream
		return w.processPacketDirect(packet)
	}
}

// processPacketDirect sends a packet using the worker's direct stream
func (w *Worker) processPacketDirect(packet *QueuedPacket) error {
	// If the stream is nil, try to recreate it
	if w.Stream == nil {
		log.Debugf("Stream is nil for %s, attempting to recreate", w.SyncKey)
		var err error
		w.Stream, err = w.recreateStream(packet.Ctx)
		if err != nil {
			return fmt.Errorf("failed to recreate stream: %v", err)
		}
	}

	// Write the packet to the stream
	_, err := w.Stream.Write(packet.Data)
	if err != nil {
		// Check for specific error types
		if isConnectionReset(err) || isStreamClosed(err) {
			w.Stream = nil
			return fmt.Errorf("connection error, will retry: %v", err)
		} else if isTemporaryError(err) {
			// For temporary errors, don't close the stream
			return fmt.Errorf("temporary error, will retry: %v", err)
		} else {
			// For other errors, close the stream
			w.Stream = nil
			return fmt.Errorf("error writing to P2P stream: %v", err)
		}
	}

	return nil
}

// processPacketPooled sends a packet using the worker's pool service
func (w *Worker) processPacketPooled(packet *QueuedPacket) error {
	// Get a stream from the pool
	stream, err := w.PoolService.GetStream(packet.Ctx, w.PeerID)
	if err != nil {
		return fmt.Errorf("failed to get stream from pool: %v", err)
	}

	// Write the packet to the stream
	_, err = stream.Write(packet.Data)

	// Release the stream back to the pool
	w.PoolService.ReleaseStream(w.PeerID, stream, err == nil)

	if err != nil {
		return fmt.Errorf("error writing to pooled stream: %v", err)
	}

	return nil
}

// processPacketMultiplexed sends a packet using the worker's multiplex service
func (w *Worker) processPacketMultiplexed(packet *QueuedPacket) error {
	// Send the packet using the multiplexer
	err := w.MultiplexService.SendPacketMultiplexed(packet.Ctx, w.PeerID, packet.Data)
	if err != nil {
		return fmt.Errorf("error sending packet through multiplexer: %v", err)
	}

	return nil
}

// recreateStream creates a new stream to the peer
func (w *Worker) recreateStream(ctx context.Context) (types.VPNStream, error) {
	// Create a new stream to the peer
	stream, err := w.getStreamService().CreateNewVPNStream(ctx, w.PeerID)
	if err != nil {
		return nil, fmt.Errorf("failed to create P2P stream: %v", err)
	}

	return stream, nil
}

// getStreamService returns the appropriate stream service based on the worker's configuration
func (w *Worker) getStreamService() types.Service {
	if w.UseMultiplexing && w.MultiplexService != nil {
		// Use the multiplex service as a stream service if it implements the Service interface
		if service, ok := w.MultiplexService.(types.Service); ok {
			return service
		}
	} else if w.UsePooling && w.PoolService != nil {
		// Use the pool service as a stream service if it implements the Service interface
		if service, ok := w.PoolService.(types.Service); ok {
			return service
		}
	}

	// Fall back to the direct stream service
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
	w.Mu.Lock()
	defer w.Mu.Unlock()
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

// isConnectionReset checks if the error is a connection reset
func isConnectionReset(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "connection reset") ||
		strings.Contains(err.Error(), "broken pipe")
}

// isStreamClosed checks if the error is due to a closed stream
func isStreamClosed(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "stream closed") ||
		strings.Contains(err.Error(), "stream reset")
}

// isTemporaryError checks if the error is temporary
func isTemporaryError(err error) bool {
	if err == nil {
		return false
	}
	// Check for timeout errors, which are typically temporary
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout()
	}
	// Check for other common temporary error messages
	return strings.Contains(err.Error(), "temporary") ||
		strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "would block")
}
