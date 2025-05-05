package dispatcher

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
)

// getOrCreateStreamWithID gets an existing stream or creates a new one with a specific ID
func (d *Dispatcher) getOrCreateStreamWithID(ctx context.Context, peerID peer.ID, streamID string) (api.VPNStream, error) {
	// Check if we already have a stream for this ID
	d.mu.RLock()
	stream, exists := d.streams[streamID]
	d.mu.RUnlock()

	if exists {
		if _, err := stream.Write(nil); err != nil {
			log.WithError(err).WithField("stream", stream).Error("Stream is closed")
			exists = false
		} else {
			return stream, nil
		}
	}

	// Create a new stream
	stream, err := d.streamService.CreateNewVPNStream(ctx, peerID)
	if err != nil {
		return nil, fmt.Errorf("failed to create new VPN stream: %w", err)
	}

	// Store the stream in the map with the specific ID
	now := time.Now()
	d.mu.Lock()
	d.streams[streamID] = stream
	d.lastUsed[streamID] = now
	d.mu.Unlock()

	// Track stream creation metrics - atomic for counter
	d.streamsCreated.Add(1)

	// Still need mutex for maps
	d.metricsMu.Lock()

	// Extract peer ID from stream ID
	parts := strings.Split(streamID, "/")
	peerStr := streamID
	if len(parts) > 0 {
		peerStr = parts[0]
	}

	// Update active streams count by peer using sync.Map
	activeVal, _ := d.activeStreamsByPeer.LoadOrStore(peerStr, &atomic.Uint64{})
	activeCounter := activeVal.(*atomic.Uint64)
	activeCounter.Add(1)

	// Track stream creation time using sync.Map
	d.streamCreationTimes.Store(streamID, now)
	if d.oldestStreamCreationTime.After(now) {
		d.oldestStreamCreationTime = now
	}
	d.newestStreamCreationTime = now

	d.metricsMu.Unlock()

	return stream, nil
}

// cleanupStreams periodically checks for idle streams and closes them
func (d *Dispatcher) cleanupStreams() {
	for {
		select {
		case <-d.stopChan:
			return
		case <-d.cleanupTicker.C:
			d.performStreamCleanup()
		}
	}
}

// performStreamCleanup checks for idle streams and closes them
func (d *Dispatcher) performStreamCleanup() {
	idleTimeout := d.configService.GetStreamIdleTimeout()
	now := time.Now()

	// Lock the streams map for writing
	d.mu.Lock()
	defer d.mu.Unlock()

	// Check each stream for idleness
	for streamID, stream := range d.streams {
		lastUsed, ok := d.lastUsed[streamID]
		if !ok {
			// If we don't have a last used time, set it to now
			d.lastUsed[streamID] = now
			continue
		}

		// If the stream has been idle for too long, close it
		if now.Sub(lastUsed) > idleTimeout {
			// Extract peer ID from stream ID (format: peerID/queueID)
			parts := strings.Split(streamID, "/")
			peerInfo := streamID
			if len(parts) > 0 {
				peerInfo = parts[0]
			}

			log.Infof("Closing idle stream %s (peer: %s) (idle for %s)", streamID, peerInfo, now.Sub(lastUsed))

			// Close the stream
			if err := stream.Close(); err != nil {
				log.WithError(err).WithField("streamID", streamID).Error("Error closing idle stream")
			}

			// Remove the stream from the maps
			delete(d.streams, streamID)
			delete(d.lastUsed, streamID)

			// Update metrics for stream closure - atomic for counter
			d.streamsClosed.Add(1)

			// Still need mutex for maps
			d.metricsMu.Lock()

			// Extract peer ID from stream ID
			streamParts := strings.Split(streamID, "/")
			peerStr := streamID
			if len(streamParts) > 0 {
				peerStr = streamParts[0]
			}

			// Update active streams count by peer using sync.Map
			if val, ok := d.activeStreamsByPeer.Load(peerStr); ok {
				counter := val.(*atomic.Uint64)
				if counter.Load() > 0 {
					counter.Add(^uint64(0)) // Subtract 1 (using bitwise complement of 0)
				}
			}

			// Remove stream creation time using sync.Map
			d.streamCreationTimes.Delete(streamID)

			d.metricsMu.Unlock()
		}
	}

	log.Debugf("Stream cleanup completed. Active streams: %d", len(d.streams))
}
