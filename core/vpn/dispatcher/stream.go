package dispatcher

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
)

// streamTimeMutex is used to protect access to stream creation time fields
var streamTimeMutex sync.Mutex

// updateStreamCreationTimes updates the oldest and newest stream creation times atomically
func updateStreamCreationTimes(d *Dispatcher, now time.Time) {
	streamTimeMutex.Lock()
	defer streamTimeMutex.Unlock()

	if d.oldestStreamCreationTime.After(now) {
		d.oldestStreamCreationTime = now
	}
	d.newestStreamCreationTime = now
}

// extractPeerID extracts the peer ID from a stream ID (format: peerID/queueID)
func extractPeerID(streamID string) string {
	parts := strings.Split(streamID, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return streamID
}

// getOrCreateStreamWithID gets an existing stream or creates a new one with a specific ID
func (d *Dispatcher) getOrCreateStreamWithID(ctx context.Context, peerID peer.ID, streamID string) (api.VPNStream, error) {
	// Check if we already have a stream for this ID
	if value, ok := d.streams.Load(streamID); ok {
		stream := value.(api.VPNStream)

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

	// Track stream creation metrics - atomic for counter
	d.streamsCreated.Add(1)

	// Extract peer ID from stream ID
	peerStr := extractPeerID(streamID)

	// Update active streams count by peer using sync.Map
	activeVal, _ := d.activeStreamsByPeer.LoadOrStore(peerStr, &atomic.Uint64{})
	activeCounter := activeVal.(*atomic.Uint64)
	activeCounter.Add(1)

	// Track stream creation time using sync.Map
	d.streamCreationTimes.Store(streamID, now)

	// Update oldest and newest stream creation times atomically
	updateStreamCreationTimes(d, now)

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
		stream := value.(api.VPNStream)

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

			// Update metrics for stream closure - atomic for counter
			d.streamsClosed.Add(1)
			closedCount.Add(1)

			// Update active streams count by peer using sync.Map
			if val, ok := d.activeStreamsByPeer.Load(peerInfo); ok {
				counter := val.(*atomic.Uint64)
				if counter.Load() > 0 {
					counter.Add(^uint64(0)) // Subtract 1 (using bitwise complement of 0)
				}
			}

			// Remove stream creation time using sync.Map
			d.streamCreationTimes.Delete(streamID)

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
