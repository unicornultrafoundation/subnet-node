package vpn

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// NewPacketWorker creates a new packet worker
func NewPacketWorker(
	syncKey string,
	destIP string,
	peerID peer.ID,
	stream VPNStream,
	ctx context.Context,
	cancel context.CancelFunc,
	bufferSize int,
) *PacketWorker {
	return &PacketWorker{
		SyncKey:      syncKey,
		DestIP:       destIP,
		PeerID:       peerID,
		Stream:       stream,
		PacketChan:   make(chan *QueuedPacket, bufferSize),
		LastActivity: time.Now(),
		Ctx:          ctx,
		Cancel:       cancel,
		Running:      true,
		PacketCount:  0,
		ErrorCount:   0,
	}
}

// Start starts the worker's processing loop
func (w *PacketWorker) Start(service VPNService) {
	go w.run(service)
}

// run processes packets for the worker
func (w *PacketWorker) run(service VPNService) {
	defer func() {
		w.Mu.Lock()
		w.Running = false
		w.Mu.Unlock()

		// Close the stream
		if w.Stream != nil {
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

			// Update last activity time
			w.Mu.Lock()
			w.LastActivity = time.Now()
			w.Mu.Unlock()

			// Process the packet
			w.Mu.Lock()
			w.PacketCount++
			w.Mu.Unlock()

			err := w.processPacket(service, packet)

			if err != nil {
				w.Mu.Lock()
				w.ErrorCount++
				w.Mu.Unlock()
			}

			// Signal completion if needed
			if packet.DoneCh != nil {
				packet.DoneCh <- err
				close(packet.DoneCh)
			}
		}
	}
}

// processPacket sends a packet using the worker's stream
func (w *PacketWorker) processPacket(service VPNService, packet *QueuedPacket) error {
	// Extract the specific service interfaces we need
	stream := service.(StreamService)
	metrics := service.(MetricsService)
	retry := service.(RetryService)

	// Use retry mechanism with exponential backoff
	operation := func() error {
		// Check if stream is valid
		if w.Stream == nil {
			// Try to create a new stream
			newStream, err := stream.CreateNewVPNStream(packet.Ctx, w.PeerID)
			if err != nil {
				// Increment stream error metric
				metrics.IncrementStreamErrors()
				return fmt.Errorf("failed to create new stream: %v", err)
			}
			w.Stream = newStream
		}

		// Write the packet to the stream
		n, err := w.Stream.Write(packet.Data)
		if err != nil {
			// Increment stream error metric
			metrics.IncrementStreamErrors()

			// Check for specific error types
			if isConnectionReset(err) || isStreamClosed(err) {
				// Close and reset for connection-level errors
				w.Stream.Close()
				w.Stream = nil
				return fmt.Errorf("connection error, will retry: %v", err)
			} else if isTemporaryError(err) {
				// For temporary errors, don't close the stream
				return fmt.Errorf("temporary error, will retry: %v", err)
			} else {
				// For other errors, close and reset
				w.Stream.Close()
				w.Stream = nil
				return fmt.Errorf("error writing to P2P stream: %v", err)
			}
		}

		// Update metrics
		metrics.IncrementPacketsSent(n)

		return nil
	}

	return retry.RetryOperation(packet.Ctx, operation)
}

// Helper functions to categorize errors
func isConnectionReset(err error) bool {
	return strings.Contains(err.Error(), "connection reset")
}

func isStreamClosed(err error) bool {
	return strings.Contains(err.Error(), "stream closed")
}

func isTemporaryError(err error) bool {
	if netErr, ok := err.(net.Error); ok {
		return netErr.Temporary()
	}
	return false
}

// Stop stops the worker
func (w *PacketWorker) Stop() {
	// Cancel the worker context
	w.Cancel()

	// Close the packet channel
	close(w.PacketChan)
}

// UpdateLastActivity updates the worker's last activity time
func (w *PacketWorker) UpdateLastActivity() {
	w.Mu.Lock()
	defer w.Mu.Unlock()
	w.LastActivity = time.Now()
}

// GetLastActivity returns the worker's last activity time
func (w *PacketWorker) GetLastActivity() time.Time {
	w.Mu.Lock()
	defer w.Mu.Unlock()
	return w.LastActivity
}

// EnqueuePacket adds a packet to the worker's queue
func (w *PacketWorker) EnqueuePacket(packet *QueuedPacket) bool {
	select {
	case w.PacketChan <- packet:
		// Successfully added to worker's channel
		return true
	default:
		// Channel is full
		return false
	}
}

// IsIdle checks if the worker has been idle for longer than the given duration
func (w *PacketWorker) IsIdle(timeout time.Duration) bool {
	lastActivity := w.GetLastActivity()
	return time.Since(lastActivity) > timeout
}
