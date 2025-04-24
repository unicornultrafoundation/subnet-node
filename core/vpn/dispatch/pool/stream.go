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
	// Map of peer ID -> mutex for that peer's streams
	peerMutexes map[string]*sync.RWMutex
	// Mutex for the peerMutexes map
	peerMutexesMu sync.RWMutex
}

// StreamChannel represents a channel dedicated to a specific stream
type StreamChannel struct {
	// The actual stream
	Stream api.VPNStream
	// Channel for sending packets to the stream
	PacketChan chan *types.QueuedPacket
	// Last activity time (Unix nano time)
	lastActivity int64
	// Whether the stream is healthy (0 = unhealthy, 1 = healthy)
	healthy int32
	// Metrics for the stream
	Metrics types.StreamMetrics
	// Context for the stream
	ctx context.Context
	// Cancel function for the stream context
	cancel context.CancelFunc
}

// GetBufferUtilization returns the current buffer utilization as a percentage (0-100)
func (s *StreamChannel) GetBufferUtilization() int {
	// Get the current length of the packet channel
	currentLen := len(s.PacketChan)

	// Get the capacity of the packet channel
	capacity := cap(s.PacketChan)

	// Calculate utilization percentage
	if capacity == 0 {
		return 0
	}

	return (currentLen * 100) / capacity
}

// StreamPoolConfig contains configuration for the stream pool
type StreamPoolConfig struct {
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
		maxStreamsPerPeer: config.MaxStreamsPerPeer,
		streamIdleTimeout: config.StreamIdleTimeout,
		cleanupInterval:   config.CleanupInterval,
		streams:           make(map[string][]*StreamChannel),
		streamUsage:       make(map[string][]int64),
		peerMutexes:       make(map[string]*sync.RWMutex),
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

	// Get a snapshot of all peer IDs
	p.streamsMu.RLock()
	peerIDs := make([]string, 0, len(p.streams))
	for peerID := range p.streams {
		peerIDs = append(peerIDs, peerID)
	}
	p.streamsMu.RUnlock()

	// Process each peer separately with its own mutex
	for _, peerID := range peerIDs {
		peerMutex := p.getPeerMutex(peerID)
		peerMutex.Lock()
		p.streamsMu.Lock()

		// Check if the peer still exists (might have been removed by another goroutine)
		streamChannels, exists := p.streams[peerID]
		if !exists {
			p.streamsMu.Unlock()
			peerMutex.Unlock()
			continue
		}

		// Close all streams for this peer
		for _, streamChannel := range streamChannels {
			// Cancel the stream context
			streamChannel.cancel()
			// Close the packet channel
			close(streamChannel.PacketChan)
			// Close the stream
			streamChannel.Stream.Close()
		}

		// Remove the peer from the maps
		delete(p.streams, peerID)
		delete(p.streamUsage, peerID)

		p.streamsMu.Unlock()
		peerMutex.Unlock()
	}

	// Clear the peer mutexes map
	p.peerMutexesMu.Lock()
	p.peerMutexes = make(map[string]*sync.RWMutex)
	p.peerMutexesMu.Unlock()
}

// GetStreamChannel gets a stream channel for a peer
// It returns the least used stream channel for the peer
func (p *StreamPool) GetStreamChannel(ctx context.Context, peerID peer.ID) (*StreamChannel, error) {
	peerIDStr := peerID.String()

	// Get the mutex for this peer
	peerMutex := p.getPeerMutex(peerIDStr)

	// Check if we have any streams for this peer
	peerMutex.RLock()
	p.streamsMu.RLock()
	streamChannels, exists := p.streams[peerIDStr]
	var usageCounts []int64
	if exists {
		usageCounts = p.streamUsage[peerIDStr]
	}
	p.streamsMu.RUnlock()
	peerMutex.RUnlock()

	// If this is the first time we're seeing this peer, initialize the minimum number of streams
	if !exists || len(streamChannels) == 0 {
		err := p.initializeMinStreams(ctx, peerID)
		if err != nil {
			return nil, err
		}

		// Refresh our view of the streams
		peerMutex.RLock()
		p.streamsMu.RLock()
		streamChannels = p.streams[peerIDStr]
		usageCounts = p.streamUsage[peerIDStr]
		p.streamsMu.RUnlock()
		peerMutex.RUnlock()
	}

	if len(streamChannels) > 0 {
		// Find the least loaded stream based on a combination of usage count and buffer utilization
		minLoadScore := float64(^uint64(0) >> 1) // Max float64
		minIndex := 0

		// Define thresholds for creating new streams
		const (
			// Buffer utilization threshold (percentage)
			bufferUtilThreshold = 70
			// Usage count threshold
			usageCountThreshold = 100
		)

		// Track if any stream is approaching capacity
		streamNearingCapacity := false

		for i, count := range usageCounts {
			// Get buffer utilization for this stream
			bufferUtil := streamChannels[i].GetBufferUtilization()

			// Calculate a load score that considers both usage count and buffer utilization
			// This is a weighted score where higher values mean more load
			loadScore := (float64(count) * 0.7) + (float64(bufferUtil) * 0.3)

			// Check if this stream is nearing capacity
			if bufferUtil > bufferUtilThreshold || count > usageCountThreshold {
				streamNearingCapacity = true

				// Log detailed metrics when a stream is nearing capacity
				streamLog.WithFields(logrus.Fields{
					"peer_id":            peerIDStr,
					"stream_index":       i,
					"buffer_utilization": bufferUtil,
					"usage_count":        count,
				}).Info("Stream nearing capacity")
			}

			// Find the stream with the lowest load score
			if loadScore < minLoadScore {
				minLoadScore = loadScore
				minIndex = i
			}
		}

		// Check if the stream is still healthy
		healthy := atomic.LoadInt32(&streamChannels[minIndex].healthy) == 1

		if healthy {
			// Check if we should create a new stream based on resource utilization
			if streamNearingCapacity &&
				len(streamChannels) < p.maxStreamsPerPeer &&
				p.maxStreamsPerPeer > 1 {
				// Log the decision to create a new stream
				streamLog.WithFields(logrus.Fields{
					"peer_id":         peerIDStr,
					"current_streams": len(streamChannels),
					"max_streams":     p.maxStreamsPerPeer,
				}).Info("Creating new stream due to resource utilization")

				// Create a new stream channel
				return p.createNewStreamChannel(ctx, peerID)
			}

			// Increment the usage count
			atomic.AddInt64(&usageCounts[minIndex], 1)
			return streamChannels[minIndex], nil
		}

		// Stream is unhealthy, remove it
		peerMutex.Lock()
		p.streamsMu.Lock()
		p.removeStreamAtIndex(peerIDStr, minIndex)
		p.streamsMu.Unlock()
		peerMutex.Unlock()
	}

	// Create a new stream channel
	return p.createNewStreamChannel(ctx, peerID)
}

// createNewStreamChannel creates a new stream channel for a peer
func (p *StreamPool) createNewStreamChannel(ctx context.Context, peerID peer.ID) (*StreamChannel, error) {
	peerIDStr := peerID.String()

	// Get the mutex for this peer
	peerMutex := p.getPeerMutex(peerIDStr)

	// Check if we've reached the maximum number of streams for this peer
	peerMutex.RLock()
	p.streamsMu.RLock()
	streamCount := len(p.streams[peerIDStr])
	p.streamsMu.RUnlock()
	peerMutex.RUnlock()

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

	// Create a new stream channel with configurable buffer size
	bufferSize := 100 // Default buffer size
	if config, ok := p.streamCreator.(api.ConfigurableStreamService); ok {
		if config.GetPacketBufferSize() > 0 {
			bufferSize = config.GetPacketBufferSize()
		}
	}

	streamChannel := &StreamChannel{
		Stream:       stream,
		PacketChan:   make(chan *types.QueuedPacket, bufferSize),
		lastActivity: time.Now().UnixNano(),
		healthy:      1, // 1 = healthy
		ctx:          streamCtx,
		cancel:       streamCancel,
	}

	// Start the stream processor
	go p.processStreamPackets(streamChannel, peerID)

	// Add the stream to the pool
	peerMutex.Lock()
	p.streamsMu.Lock()

	if _, exists := p.streams[peerIDStr]; !exists {
		p.streams[peerIDStr] = make([]*StreamChannel, 0)
		p.streamUsage[peerIDStr] = make([]int64, 0)
	}

	p.streams[peerIDStr] = append(p.streams[peerIDStr], streamChannel)
	p.streamUsage[peerIDStr] = append(p.streamUsage[peerIDStr], 1) // Start with usage count of 1

	p.streamsMu.Unlock()
	peerMutex.Unlock()

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
		atomic.StoreInt32(&streamChannel.healthy, 0)

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
			atomic.StoreInt64(&streamChannel.lastActivity, time.Now().UnixNano())

			// Process the packet
			err := p.writePacketToStream(streamChannel, packet)
			if err != nil {
				logger.WithError(err).Debug("Failed to write packet to stream")

				// Signal the error on the done channel if provided
				if packet.DoneCh != nil {
					// Use a non-blocking send to avoid panics if the channel is closed
					select {
					case packet.DoneCh <- err:
						// Successfully sent the error
					default:
						// Channel might be closed or full, don't panic
					}
					// Don't close the channel here, let the caller close it
				}

				// Mark the stream as unhealthy
				atomic.StoreInt32(&streamChannel.healthy, 0)

				// Increment error count
				atomic.AddInt64(&streamChannel.Metrics.ErrorCount, 1)

				return
			}

			// Signal success on the done channel if provided
			if packet.DoneCh != nil {
				select {
				case packet.DoneCh <- nil:
					// Successfully sent the result
				default:
					// Channel might be closed or full, don't panic
				}
				// Don't close the channel here, let the caller close it
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
		atomic.StoreInt32(&streamChannel.healthy, 0)

		// Cancel the stream context to stop the processor
		streamChannel.cancel()
	}
}

// removeStreamAtIndex removes a stream at the specified index
// Note: This method assumes that the streamsMu lock is already held
// Note: This method assumes that the peer-specific mutex is already held
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
			// Get a snapshot of all peer IDs
			p.streamsMu.RLock()
			peerIDs := make([]string, 0, len(p.streams))
			for peerID := range p.streams {
				peerIDs = append(peerIDs, peerID)
			}
			p.streamsMu.RUnlock()

			// Process each peer separately with its own mutex
			for _, peerID := range peerIDs {
				peerMutex := p.getPeerMutex(peerID)
				peerMutex.Lock()
				p.streamsMu.Lock()

				// Check if the peer still exists (might have been removed by another goroutine)
				streamChannels, exists := p.streams[peerID]
				if !exists {
					p.streamsMu.Unlock()
					peerMutex.Unlock()
					continue
				}

				// Check each stream
				for i := len(streamChannels) - 1; i >= 0; i-- {
					lastActivity := time.Unix(0, atomic.LoadInt64(&streamChannels[i].lastActivity))
					healthy := atomic.LoadInt32(&streamChannels[i].healthy) == 1

					// Check if the stream is idle or unhealthy
					if !healthy || time.Since(lastActivity) > p.streamIdleTimeout {
						p.removeStreamAtIndex(peerID, i)
					}
				}

				p.streamsMu.Unlock()
				peerMutex.Unlock()
			}
		}
	}
}

// GetStreamCount returns the number of streams for a peer
func (p *StreamPool) GetStreamCount(peerID peer.ID) int {
	peerIDStr := peerID.String()

	// Get the mutex for this peer
	peerMutex := p.getPeerMutex(peerIDStr)

	peerMutex.RLock()
	p.streamsMu.RLock()
	count := len(p.streams[peerIDStr])
	p.streamsMu.RUnlock()
	peerMutex.RUnlock()

	return count
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
	// Get a snapshot of all peer IDs
	p.streamsMu.RLock()
	peerIDs := make([]string, 0, len(p.streams))
	for peerID := range p.streams {
		peerIDs = append(peerIDs, peerID)
	}
	p.streamsMu.RUnlock()

	metrics := make(map[string]map[string]int64)

	// Process each peer separately with its own mutex
	for _, peerID := range peerIDs {
		peerMutex := p.getPeerMutex(peerID)
		peerMutex.RLock()
		p.streamsMu.RLock()

		// Check if the peer still exists (might have been removed by another goroutine)
		streamChannels, exists := p.streams[peerID]
		if !exists {
			p.streamsMu.RUnlock()
			peerMutex.RUnlock()
			continue
		}

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

		p.streamsMu.RUnlock()
		peerMutex.RUnlock()
	}

	return metrics
}

// initializeMinStreams creates the minimum number of streams for a peer
func (p *StreamPool) initializeMinStreams(ctx context.Context, peerID peer.ID) error {
	peerIDStr := peerID.String()
	streamLog.WithFields(logrus.Fields{
		"peer_id":     peerIDStr,
		"min_streams": 1,
	}).Debug("Initializing minimum streams for peer")

	// Get the mutex for this peer
	peerMutex := p.getPeerMutex(peerIDStr)

	// Lock to prevent concurrent initialization
	peerMutex.Lock()
	p.streamsMu.Lock()

	// Check again in case another goroutine initialized the streams while we were waiting for the lock
	streamChannels, exists := p.streams[peerIDStr]
	if exists && len(streamChannels) > 0 {
		p.streamsMu.Unlock()
		peerMutex.Unlock()
		return nil
	}

	// Initialize the maps if they don't exist
	if !exists {
		p.streams[peerIDStr] = make([]*StreamChannel, 0, 1)
		p.streamUsage[peerIDStr] = make([]int64, 0, 1)
	}

	// Create a new stream
	stream, err := p.streamCreator.CreateNewVPNStream(ctx, peerID)
	if err != nil {
		streamLog.WithFields(logrus.Fields{
			"peer_id": peerIDStr,
			"error":   err,
		}).Warn("Failed to create stream during initialization")
	} else {
		// Create a new stream context
		streamCtx, streamCancel := context.WithCancel(p.ctx)

		// Create a new stream channel with configurable buffer size
		bufferSize := 100 // Default buffer size
		if config, ok := p.streamCreator.(api.ConfigurableStreamService); ok {
			if config.GetPacketBufferSize() > 0 {
				bufferSize = config.GetPacketBufferSize()
			}
		}

		streamChannel := &StreamChannel{
			Stream:       stream,
			PacketChan:   make(chan *types.QueuedPacket, bufferSize),
			lastActivity: time.Now().UnixNano(),
			healthy:      1, // 1 = healthy
			ctx:          streamCtx,
			cancel:       streamCancel,
		}

		// Start the stream processor
		go p.processStreamPackets(streamChannel, peerID)

		// Add the stream to the pool
		p.streams[peerIDStr] = append(p.streams[peerIDStr], streamChannel)
		p.streamUsage[peerIDStr] = append(p.streamUsage[peerIDStr], 0) // Start with usage count of 0
	}

	streamCount := len(p.streams[peerIDStr])

	p.streamsMu.Unlock()
	peerMutex.Unlock()

	streamLog.WithFields(logrus.Fields{
		"peer_id":      peerIDStr,
		"stream_count": streamCount,
	}).Debug("Initialized streams for peer")

	return nil
}

// getPeerMutex gets or creates a mutex for a peer
func (p *StreamPool) getPeerMutex(peerIDStr string) *sync.RWMutex {
	// First try with a read lock
	p.peerMutexesMu.RLock()
	mutex, exists := p.peerMutexes[peerIDStr]
	p.peerMutexesMu.RUnlock()

	if exists {
		return mutex
	}

	// If not found, acquire write lock and create
	p.peerMutexesMu.Lock()
	defer p.peerMutexesMu.Unlock()

	// Check again in case another goroutine created it while we were waiting
	mutex, exists = p.peerMutexes[peerIDStr]
	if exists {
		return mutex
	}

	// Create a new mutex
	mutex = &sync.RWMutex{}
	p.peerMutexes[peerIDStr] = mutex
	return mutex
}

// Close implements io.Closer
func (p *StreamPool) Close() error {
	p.Stop()
	return nil
}

// GetPacketBufferSize returns the packet buffer size for streams
func (p *StreamPool) GetPacketBufferSize() int {
	// If the stream creator is configurable, use its buffer size
	if config, ok := p.streamCreator.(api.ConfigurableStreamService); ok {
		return config.GetPacketBufferSize()
	}

	// Otherwise, return a default value
	return 100
}
