package pool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
)

var streamLog = logrus.WithField("service", "vpn-stream-pool")

// StreamPool manages a pool of streams for each peer
type StreamPool struct {
	// Core components
	streamCreator api.StreamService
	ctx           context.Context
	cancel        context.CancelFunc

	// Configuration
	minStreamsPerPeer int
	maxStreamsPerPeer int
	streamIdleTimeout time.Duration
	cleanupInterval   time.Duration

	// Lifecycle management
	mu     sync.RWMutex
	active bool

	// Stream management
	streamsMu sync.RWMutex
	// Map of peer ID -> slice of stream channels
	streams map[string][]*StreamChannel
	// Map of peer ID -> stream usage counts
	streamUsage map[string][]int64
}

// StreamChannel represents a channel dedicated to a specific stream
type StreamChannel struct {
	// The actual stream
	Stream api.VPNStream
	// Channel for sending packets to the stream
	PacketChan chan *types.QueuedPacket
	// Last activity time
	LastActivity time.Time
	// Mutex for protecting access to the stream
	mu sync.RWMutex
	// Whether the stream is healthy
	Healthy bool
	// Metrics for the stream
	Metrics types.StreamMetrics
	// Context for the stream
	ctx context.Context
	// Cancel function for the stream context
	cancel context.CancelFunc
}

// StreamPoolConfig contains configuration for the stream pool
type StreamPoolConfig struct {
	MinStreamsPerPeer int
	MaxStreamsPerPeer int
	StreamIdleTimeout time.Duration
	CleanupInterval   time.Duration
	PacketBufferSize  int
}

// NewStreamPool creates a new stream pool
func NewStreamPool(streamCreator api.StreamService, config *StreamPoolConfig) *StreamPool {
	ctx, cancel := context.WithCancel(context.Background())

	return &StreamPool{
		streamCreator:     streamCreator,
		ctx:               ctx,
		cancel:            cancel,
		minStreamsPerPeer: config.MinStreamsPerPeer,
		maxStreamsPerPeer: config.MaxStreamsPerPeer,
		streamIdleTimeout: config.StreamIdleTimeout,
		cleanupInterval:   config.CleanupInterval,
		streams:           make(map[string][]*StreamChannel),
		streamUsage:       make(map[string][]int64),
	}
}

// Start starts the stream pool and its cleanup routine
func (p *StreamPool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.active {
		return
	}

	streamLog.Info("Starting stream pool")
	p.active = true

	// Start the cleanup routine
	go p.cleanupIdleStreams()
}

// Stop stops the stream pool and its cleanup routine
func (p *StreamPool) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.active {
		return
	}

	streamLog.Info("Stopping stream pool")
	p.cancel()
	p.active = false

	// Close all streams
	p.streamsMu.Lock()
	defer p.streamsMu.Unlock()

	for peerID, streamChannels := range p.streams {
		for _, streamChannel := range streamChannels {
			// Cancel the stream context
			streamChannel.cancel()
			// Close the packet channel
			close(streamChannel.PacketChan)
			// Close the stream
			streamChannel.Stream.Close()
		}
		delete(p.streams, peerID)
		delete(p.streamUsage, peerID)
	}
}

// GetStreamChannel gets a stream channel for a peer
// It returns the least used stream channel for the peer
func (p *StreamPool) GetStreamChannel(ctx context.Context, peerID peer.ID) (*StreamChannel, error) {
	peerIDStr := peerID.String()

	// Check if we have any streams for this peer
	p.streamsMu.RLock()
	streamChannels, exists := p.streams[peerIDStr]
	var usageCounts []int64
	if exists {
		usageCounts = p.streamUsage[peerIDStr]
	}
	p.streamsMu.RUnlock()

	// If this is the first time we're seeing this peer, initialize the minimum number of streams
	if !exists || len(streamChannels) < p.minStreamsPerPeer {
		err := p.initializeMinStreams(ctx, peerID)
		if err != nil {
			return nil, err
		}

		// Refresh our view of the streams
		p.streamsMu.RLock()
		streamChannels = p.streams[peerIDStr]
		usageCounts = p.streamUsage[peerIDStr]
		p.streamsMu.RUnlock()
	}

	if len(streamChannels) > 0 {
		// Find the least used stream
		minUsage := int64(^uint64(0) >> 1) // Max int64
		minIndex := 0

		for i, count := range usageCounts {
			if count < minUsage {
				minUsage = count
				minIndex = i
			}
		}

		// Check if the stream is still healthy
		streamChannels[minIndex].mu.RLock()
		healthy := streamChannels[minIndex].Healthy
		streamChannels[minIndex].mu.RUnlock()

		if healthy {
			// Increment the usage count
			atomic.AddInt64(&usageCounts[minIndex], 1)
			return streamChannels[minIndex], nil
		}

		// Stream is unhealthy, remove it
		p.streamsMu.Lock()
		p.removeStreamAtIndex(peerIDStr, minIndex)
		p.streamsMu.Unlock()
	}

	// Create a new stream channel
	return p.createNewStreamChannel(ctx, peerID)
}

// createNewStreamChannel creates a new stream channel for a peer
func (p *StreamPool) createNewStreamChannel(ctx context.Context, peerID peer.ID) (*StreamChannel, error) {
	peerIDStr := peerID.String()

	// Check if we've reached the maximum number of streams for this peer
	p.streamsMu.RLock()
	streamCount := len(p.streams[peerIDStr])
	p.streamsMu.RUnlock()

	if streamCount >= p.maxStreamsPerPeer {
		return nil, types.ErrStreamPoolExhausted
	}

	// Create a new stream
	stream, err := p.streamCreator.CreateNewVPNStream(ctx, peerID)
	if err != nil {
		return nil, err
	}

	// Create a new stream context
	streamCtx, streamCancel := context.WithCancel(p.ctx)

	// Create a new stream channel
	streamChannel := &StreamChannel{
		Stream:       stream,
		PacketChan:   make(chan *types.QueuedPacket, 100), // Buffer size
		LastActivity: time.Now(),
		Healthy:      true,
		ctx:          streamCtx,
		cancel:       streamCancel,
	}

	// Start the stream processor
	go p.processStreamPackets(streamChannel, peerID)

	// Add the stream to the pool
	p.streamsMu.Lock()
	defer p.streamsMu.Unlock()

	if _, exists := p.streams[peerIDStr]; !exists {
		p.streams[peerIDStr] = make([]*StreamChannel, 0)
		p.streamUsage[peerIDStr] = make([]int64, 0)
	}

	p.streams[peerIDStr] = append(p.streams[peerIDStr], streamChannel)
	p.streamUsage[peerIDStr] = append(p.streamUsage[peerIDStr], 1) // Start with usage count of 1

	return streamChannel, nil
}

// processStreamPackets processes packets for a stream
func (p *StreamPool) processStreamPackets(streamChannel *StreamChannel, peerID peer.ID) {
	peerIDStr := peerID.String()
	logger := streamLog.WithFields(logrus.Fields{
		"peer_id": peerIDStr,
	})

	defer func() {
		// Mark the stream as unhealthy
		streamChannel.mu.Lock()
		streamChannel.Healthy = false
		streamChannel.mu.Unlock()

		// Close the stream
		streamChannel.Stream.Close()

		logger.Debug("Stream processor stopped")
	}()

	for {
		select {
		case <-streamChannel.ctx.Done():
			logger.Debug("Stream context cancelled")
			return
		case packet, ok := <-streamChannel.PacketChan:
			if !ok {
				logger.Debug("Stream packet channel closed")
				return
			}

			// Update last activity time
			streamChannel.mu.Lock()
			streamChannel.LastActivity = time.Now()
			streamChannel.mu.Unlock()

			// Process the packet
			err := p.writePacketToStream(streamChannel, packet)
			if err != nil {
				logger.WithError(err).Debug("Failed to write packet to stream")

				// Signal the error on the done channel if provided
				if packet.DoneCh != nil {
					packet.DoneCh <- err
					close(packet.DoneCh)
				}

				// Mark the stream as unhealthy
				streamChannel.mu.Lock()
				streamChannel.Healthy = false
				streamChannel.mu.Unlock()

				// Increment error count
				atomic.AddInt64(&streamChannel.Metrics.ErrorCount, 1)

				return
			}

			// Signal success on the done channel if provided
			if packet.DoneCh != nil {
				packet.DoneCh <- nil
				close(packet.DoneCh)
			}

			// Update metrics
			atomic.AddInt64(&streamChannel.Metrics.PacketCount, 1)
			atomic.AddInt64(&streamChannel.Metrics.BytesSent, int64(len(packet.Data)))
		}
	}
}

// writePacketToStream writes a packet to a stream
func (p *StreamPool) writePacketToStream(streamChannel *StreamChannel, packet *types.QueuedPacket) error {
	// Write the packet to the stream
	_, err := streamChannel.Stream.Write(packet.Data)
	if err != nil {
		return types.NewNetworkError(err, "write", packet.DestIP, "")
	}

	return nil
}

// ReleaseStreamChannel marks a stream channel as unhealthy if it's not healthy
func (p *StreamPool) ReleaseStreamChannel(peerID peer.ID, streamChannel *StreamChannel, healthy bool) {
	if !healthy {
		// Mark the stream as unhealthy
		streamChannel.mu.Lock()
		streamChannel.Healthy = false
		streamChannel.mu.Unlock()

		// Cancel the stream context to stop the processor
		streamChannel.cancel()
	}
}

// removeStreamAtIndex removes a stream at the specified index
func (p *StreamPool) removeStreamAtIndex(peerIDStr string, index int) {
	streams := p.streams[peerIDStr]
	usageCounts := p.streamUsage[peerIDStr]

	if index >= len(streams) {
		return
	}

	// Cancel the stream context
	streams[index].cancel()
	// Close the packet channel
	close(streams[index].PacketChan)
	// Close the stream
	streams[index].Stream.Close()

	// Remove the stream from the slice
	p.streams[peerIDStr] = append(streams[:index], streams[index+1:]...)
	p.streamUsage[peerIDStr] = append(usageCounts[:index], usageCounts[index+1:]...)

	// If there are no more streams, remove the peer from the maps
	if len(p.streams[peerIDStr]) == 0 {
		delete(p.streams, peerIDStr)
		delete(p.streamUsage, peerIDStr)
	}
}

// cleanupIdleStreams periodically checks for and removes idle streams
func (p *StreamPool) cleanupIdleStreams() {
	ticker := time.NewTicker(p.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.streamsMu.Lock()

			for peerID, streamChannels := range p.streams {
				for i := len(streamChannels) - 1; i >= 0; i-- {
					streamChannels[i].mu.RLock()
					lastActivity := streamChannels[i].LastActivity
					healthy := streamChannels[i].Healthy
					streamChannels[i].mu.RUnlock()

					// Check if the stream is idle or unhealthy
					if !healthy || time.Since(lastActivity) > p.streamIdleTimeout {
						p.removeStreamAtIndex(peerID, i)
					}
				}
			}

			p.streamsMu.Unlock()
		}
	}
}

// GetStreamCount returns the number of streams for a peer
func (p *StreamPool) GetStreamCount(peerID peer.ID) int {
	peerIDStr := peerID.String()

	p.streamsMu.RLock()
	defer p.streamsMu.RUnlock()

	return len(p.streams[peerIDStr])
}

// GetTotalStreamCount returns the total number of streams
func (p *StreamPool) GetTotalStreamCount() int {
	p.streamsMu.RLock()
	defer p.streamsMu.RUnlock()

	count := 0
	for _, streamChannels := range p.streams {
		count += len(streamChannels)
	}

	return count
}

// GetStreamMetrics returns metrics for all streams
func (p *StreamPool) GetStreamMetrics() map[string]map[string]int64 {
	p.streamsMu.RLock()
	defer p.streamsMu.RUnlock()

	metrics := make(map[string]map[string]int64)

	for peerID, streamChannels := range p.streams {
		peerMetrics := make(map[string]int64)

		var totalPackets, totalErrors, totalBytes int64

		for _, streamChannel := range streamChannels {
			totalPackets += atomic.LoadInt64(&streamChannel.Metrics.PacketCount)
			totalErrors += atomic.LoadInt64(&streamChannel.Metrics.ErrorCount)
			totalBytes += atomic.LoadInt64(&streamChannel.Metrics.BytesSent)
		}

		peerMetrics["stream_count"] = int64(len(streamChannels))
		peerMetrics["packet_count"] = totalPackets
		peerMetrics["error_count"] = totalErrors
		peerMetrics["bytes_sent"] = totalBytes

		metrics[peerID] = peerMetrics
	}

	return metrics
}

// initializeMinStreams creates the minimum number of streams for a peer
func (p *StreamPool) initializeMinStreams(ctx context.Context, peerID peer.ID) error {
	peerIDStr := peerID.String()
	streamLog.WithFields(logrus.Fields{
		"peer_id":     peerIDStr,
		"min_streams": p.minStreamsPerPeer,
	}).Debug("Initializing minimum streams for peer")

	// Lock to prevent concurrent initialization
	p.streamsMu.Lock()
	defer p.streamsMu.Unlock()

	// Check again in case another goroutine initialized the streams while we were waiting for the lock
	streamChannels, exists := p.streams[peerIDStr]
	if exists && len(streamChannels) >= p.minStreamsPerPeer {
		return nil
	}

	// Initialize the maps if they don't exist
	if !exists {
		p.streams[peerIDStr] = make([]*StreamChannel, 0, p.minStreamsPerPeer)
		p.streamUsage[peerIDStr] = make([]int64, 0, p.minStreamsPerPeer)
	}

	// Create the minimum number of streams
	for i := len(p.streams[peerIDStr]); i < p.minStreamsPerPeer; i++ {
		// Create a new stream
		stream, err := p.streamCreator.CreateNewVPNStream(ctx, peerID)
		if err != nil {
			streamLog.WithFields(logrus.Fields{
				"peer_id": peerIDStr,
				"error":   err,
			}).Warn("Failed to create stream during initialization")
			// Continue with the streams we've created so far
			break
		}

		// Create a new stream context
		streamCtx, streamCancel := context.WithCancel(p.ctx)

		// Create a new stream channel
		streamChannel := &StreamChannel{
			Stream:       stream,
			PacketChan:   make(chan *types.QueuedPacket, 100), // Buffer size
			LastActivity: time.Now(),
			Healthy:      true,
			ctx:          streamCtx,
			cancel:       streamCancel,
		}

		// Start the stream processor
		go p.processStreamPackets(streamChannel, peerID)

		// Add the stream to the pool
		p.streams[peerIDStr] = append(p.streams[peerIDStr], streamChannel)
		p.streamUsage[peerIDStr] = append(p.streamUsage[peerIDStr], 0) // Start with usage count of 0
	}

	streamLog.WithFields(logrus.Fields{
		"peer_id":      peerIDStr,
		"stream_count": len(p.streams[peerIDStr]),
	}).Debug("Initialized streams for peer")

	return nil
}

// Close implements io.Closer
func (p *StreamPool) Close() error {
	p.Stop()
	return nil
}
