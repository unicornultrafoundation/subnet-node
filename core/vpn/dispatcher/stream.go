package dispatcher

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// extractPeerID extracts the peer ID from a stream ID (format: peerID/queueID)
func extractPeerID(streamID string) string {
	parts := strings.Split(streamID, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return streamID
}

// getOrCreateStreamWithID gets an existing stream or creates a new one with a specific ID
func (d *Dispatcher) getOrCreateStreamWithID(ctx context.Context, peerID peer.ID, streamID string) (network.Stream, error) {
	// Check if we already have a stream for this ID
	if value, ok := d.streams.Load(streamID); ok {
		stream := value.(network.Stream)

		// Check if the stream is healthy using a more reliable method
		// Set a short deadline to check if the stream is responsive
		err := stream.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))
		if err == nil {
			// Reset the deadline to avoid affecting future operations
			_ = stream.SetWriteDeadline(time.Time{})
			return stream, nil
		}

		log.WithError(err).WithField("stream", streamID).Error("Stream is unhealthy")
		// Stream is unhealthy, will create a new one
	}

	// Create a new stream
	stream, err := d.streamService.CreateNewVPNStream(ctx, peerID)
	if err != nil {
		return nil, fmt.Errorf("failed to create new VPN stream: %w", err)
	}

	// Store the stream in the map with the specific ID
	now := time.Now()
	d.streams.Store(streamID, stream)
	d.lastUsed.Store(streamID, now)

	return stream, nil
}

// cleanupStreams periodically checks for idle streams and closes them
func (d *Dispatcher) cleanupStreams() {
	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recovered from panic in cleanupStreams: %v", r)
			// Restart the goroutine after a short delay
			time.Sleep(time.Second)
			go d.cleanupStreams()
		}
	}()

	for {
		select {
		case <-d.stopChan:
			return
		case <-d.cleanupTicker.C:
			// Wrap the cleanup in another recovery block
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Errorf("Recovered from panic in performStreamCleanup: %v", r)
					}
				}()
				d.performStreamCleanup()
			}()
		}
	}
}

// performStreamCleanup checks for idle streams and closes them
func (d *Dispatcher) performStreamCleanup() {
	idleTimeout := d.configService.GetStreamIdleTimeout()
	now := time.Now()

	// Track how many streams were closed for logging
	var closedCount atomic.Uint64
	var activeStreamsCount atomic.Uint64

	// Process streams in batches to avoid memory pressure
	const batchSize = 50
	var batch int

	// Check each stream for idleness and close directly during iteration
	d.streams.Range(func(key, value interface{}) bool {
		streamID := key.(string)
		stream := value.(network.Stream)

		// Get last used time
		lastUsedValue, ok := d.lastUsed.Load(streamID)
		if !ok {
			// If we don't have a last used time, set it to now
			d.lastUsed.Store(streamID, now)
			activeStreamsCount.Add(1)
			return true // continue iteration
		}

		lastUsed := lastUsedValue.(time.Time)
		idleTime := now.Sub(lastUsed)

		// If the stream has been idle for too long, close it
		if idleTime > idleTimeout {
			// Extract peer ID from stream ID (format: peerID/queueID)
			peerInfo := extractPeerID(streamID)

			log.Infof("Closing idle stream %s (peer: %s) (idle for %s)", streamID, peerInfo, idleTime)

			// Close the stream
			if err := stream.Close(); err != nil {
				log.WithError(err).WithField("streamID", streamID).Error("Error closing idle stream")
			}

			// Remove the stream from the maps
			d.streams.Delete(streamID)
			d.lastUsed.Delete(streamID)

			closedCount.Add(1)

			// Increment batch counter
			batch++

			// If we've processed a full batch, yield to allow other goroutines to run
			if batch >= batchSize {
				batch = 0
				// Sleep for a very short time to yield
				time.Sleep(time.Millisecond)
			}
		} else {
			activeStreamsCount.Add(1)
		}

		return true // continue iteration
	})

	log.Debugf("Stream cleanup completed. Closed: %d, Active streams: %d",
		closedCount.Load(), activeStreamsCount.Load())
}
