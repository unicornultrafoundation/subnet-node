package pool

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
)

var streamPoolLog = logrus.WithField("service", "vpn-stream-pool")

// StreamPool is an improved implementation of StreamPool with reduced mutex usage
type StreamPool struct {
	// Core components
	streamCreator api.StreamService
	ctx           context.Context
	cancel        context.CancelFunc

	// Configuration
	maxStreamsPerPeer int
	streamIdleTimeout time.Duration
	cleanupInterval   time.Duration
	packetBufferSize  int

	// Lifecycle management
	active int32 // 0 = inactive, 1 = active

	// Stream operation channel
	opChan chan streamOp

	// Metrics
	metrics struct {
		StreamsCreated  int64
		StreamsRemoved  int64
		PacketsHandled  int64
		PacketsDropped  int64
		Errors          int64
		ActiveStreams   int64
		ConnectionCount int64
	}

	// Sharded peer data (to reduce contention)
	peerShards [16]*peerShard
}

// StreamPoolConfig contains configuration for the stream pool
type StreamPoolConfig struct {
	MaxStreamsPerPeer int
	StreamIdleTimeout time.Duration
	CleanupInterval   time.Duration
	PacketBufferSize  int
}

// peerShard represents a shard of the peer data
type peerShard struct {
	sync.RWMutex
	// Map of peer ID -> peer data
	peers map[string]*peerData
}

// peerData contains all data for a specific peer
type peerData struct {
	// Streams for this peer
	streams []*StreamChannel
	// Usage counts for each stream
	usageCounts []int64
}

// streamOp represents an operation on the stream pool
type streamOp struct {
	opType     string // "get_stream", "release_stream", "cleanup", "get_metrics", etc.
	peerID     peer.ID
	stream     *StreamChannel
	healthy    bool
	ctx        context.Context
	resultChan chan streamOpResult
}

// streamOpResult represents the result of a stream operation
type streamOpResult struct {
	stream  *StreamChannel
	metrics map[string]map[string]int64
	count   int
	err     error
}

// NewStreamPool creates a new stream pool with reduced mutex usage
func NewStreamPool(streamCreator api.StreamService, config *StreamPoolConfig) *StreamPool {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize the peer shards
	peerShards := [16]*peerShard{}
	for i := 0; i < 16; i++ {
		peerShards[i] = &peerShard{
			peers: make(map[string]*peerData),
		}
	}

	pool := &StreamPool{
		streamCreator:     streamCreator,
		ctx:               ctx,
		cancel:            cancel,
		maxStreamsPerPeer: config.MaxStreamsPerPeer,
		streamIdleTimeout: config.StreamIdleTimeout,
		cleanupInterval:   config.CleanupInterval,
		packetBufferSize:  config.PacketBufferSize,
		opChan:            make(chan streamOp, 100), // Buffer size for operations
		peerShards:        peerShards,
	}

	return pool
}

// Start starts the stream pool and its worker routine
func (p *StreamPool) Start() {
	// Only start if not already active
	if !atomic.CompareAndSwapInt32(&p.active, 0, 1) {
		return
	}

	streamPoolLog.Info("Starting stream pool")

	// Start the worker routine
	go p.processOperations()

	// Start the cleanup routine
	go p.cleanupIdleStreams()

	// Start the stream stats logger
	go p.streamStatsLogger()
}

// Stop stops the stream pool and its worker routine
func (p *StreamPool) Stop() {
	// Only stop if active
	if !atomic.CompareAndSwapInt32(&p.active, 1, 0) {
		return
	}

	streamPoolLog.Info("Stopping stream pool")
	p.cancel()

	// Close the operation channel
	close(p.opChan)
}

// processOperations is the main worker routine that processes stream operations
func (p *StreamPool) processOperations() {
	for op := range p.opChan {
		switch op.opType {
		case "get_stream":
			stream, err := p.getStreamChannelInternal(op.ctx, op.peerID)
			op.resultChan <- streamOpResult{stream: stream, err: err}
		case "release_stream":
			p.releaseStreamChannelInternal(op.peerID, op.stream, op.healthy)
			op.resultChan <- streamOpResult{}
		case "get_metrics":
			metrics := p.getStreamMetricsInternal()
			op.resultChan <- streamOpResult{metrics: metrics}
		case "get_stream_count":
			count := p.getStreamCountInternal(op.peerID)
			op.resultChan <- streamOpResult{count: count}
		case "get_total_stream_count":
			count := p.getTotalStreamCountInternal()
			op.resultChan <- streamOpResult{count: count}
		case "cleanup":
			p.cleanupIdleStreamsInternal()
			op.resultChan <- streamOpResult{}
		}
	}
}

// GetStreamChannel gets a stream channel for a peer
func (p *StreamPool) GetStreamChannel(ctx context.Context, peerID peer.ID) (*StreamChannel, error) {
	// Send the operation to the worker routine
	resultChan := make(chan streamOpResult, 1)
	p.opChan <- streamOp{
		opType:     "get_stream",
		peerID:     peerID,
		ctx:        ctx,
		resultChan: resultChan,
	}

	// Wait for the result
	result := <-resultChan
	return result.stream, result.err
}

// getStreamChannelInternal is the internal implementation of GetStreamChannel
func (p *StreamPool) getStreamChannelInternal(ctx context.Context, peerID peer.ID) (*StreamChannel, error) {
	peerIDStr := peerID.String()
	shardIdx := p.getPeerShardIndex(peerIDStr)
	shard := p.peerShards[shardIdx]

	// Check if we have any streams for this peer
	shard.RLock()
	peerData, exists := shard.peers[peerIDStr]
	var streams []*StreamChannel
	var usageCounts []int64
	if exists {
		streams = peerData.streams
		usageCounts = peerData.usageCounts
	}
	shard.RUnlock()

	// If this is the first time we're seeing this peer, initialize the peer data
	if !exists || len(streams) == 0 {
		// Create a new stream directly instead of initializing minimum streams
		return p.createNewStreamChannel(ctx, peerID)
	}

	if len(streams) > 0 {
		// Define thresholds for creating new streams
		const (
			// Buffer utilization threshold (percentage)
			bufferUtilThreshold = 70
			// Usage count threshold
			usageCountThreshold = 100
		)

		// Track if any stream is approaching capacity
		streamNearingCapacity := false

		// Find the least loaded stream based on a combination of usage count and buffer utilization
		minLoadScore := float64(^uint64(0) >> 1) // Max float64
		minIndex := 0

		for i, count := range usageCounts {
			// Get buffer utilization for this stream
			bufferUtil := streams[i].GetBufferUtilization()

			// Calculate a load score that considers both usage count and buffer utilization
			// This is a weighted score where higher values mean more load
			loadScore := (float64(count) * 0.7) + (float64(bufferUtil) * 0.3)

			// Check if this stream is nearing capacity
			if bufferUtil > bufferUtilThreshold || count > usageCountThreshold {
				streamNearingCapacity = true

				// Log detailed metrics when a stream is nearing capacity
				streamPoolLog.WithFields(logrus.Fields{
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
		if streams[minIndex].IsHealthy() {
			// Check if we should create a new stream based on resource utilization
			if streamNearingCapacity &&
				len(streams) < p.maxStreamsPerPeer &&
				p.maxStreamsPerPeer > 1 {
				// Log the decision to create a new stream
				streamPoolLog.WithFields(logrus.Fields{
					"peer_id":         peerIDStr,
					"current_streams": len(streams),
					"max_streams":     p.maxStreamsPerPeer,
				}).Info("Creating new stream due to resource utilization")

				// Create a new stream channel
				return p.createNewStreamChannel(ctx, peerID)
			}

			// Increment the usage count
			atomic.AddInt64(&usageCounts[minIndex], 1)
			return streams[minIndex], nil
		}

		// Stream is unhealthy, remove it
		shard.Lock()
		p.removeStreamAtIndex(peerIDStr, minIndex, shard)
		shard.Unlock()
	}

	// Create a new stream channel
	return p.createNewStreamChannel(ctx, peerID)
}

// ReleaseStreamChannel marks a stream channel as unhealthy if it's not healthy
func (p *StreamPool) ReleaseStreamChannel(peerID peer.ID, streamChannel *StreamChannel, healthy bool) {
	// Send the operation to the worker routine
	resultChan := make(chan streamOpResult, 1)
	p.opChan <- streamOp{
		opType:     "release_stream",
		peerID:     peerID,
		stream:     streamChannel,
		healthy:    healthy,
		resultChan: resultChan,
	}

	// Wait for the operation to complete
	<-resultChan
}

// releaseStreamChannelInternal is the internal implementation of ReleaseStreamChannel
func (p *StreamPool) releaseStreamChannelInternal(peerID peer.ID, streamChannel *StreamChannel, healthy bool) {
	// Check if the stream channel is nil
	if streamChannel == nil {
		return
	}

	if !healthy {
		// Mark the stream as unhealthy
		streamChannel.SetHealthy(false)

		// Cancel the stream context to stop the processor
		streamChannel.cancel()
	}
}

// createNewStreamChannel creates a new stream channel for a peer
func (p *StreamPool) createNewStreamChannel(ctx context.Context, peerID peer.ID) (*StreamChannel, error) {
	peerIDStr := peerID.String()
	shardIdx := p.getPeerShardIndex(peerIDStr)
	shard := p.peerShards[shardIdx]

	// Check if we've reached the maximum number of streams for this peer
	shard.RLock()
	peerDataValue, exists := shard.peers[peerIDStr]
	var streamCount int
	if exists {
		streamCount = len(peerDataValue.streams)
	}
	shard.RUnlock()

	if streamCount >= p.maxStreamsPerPeer {
		return nil, types.ErrStreamPoolExhausted
	}

	// Create a new stream with a context that has no deadline
	// This allows us to wait as long as needed for the stream to be created
	streamCtx, cancel := context.WithCancel(context.Background())

	// Set up cancellation if the original context is canceled
	go func() {
		select {
		case <-ctx.Done():
			// Original context was canceled, cancel our stream context too
			cancel()
		case <-streamCtx.Done():
			// Our context was canceled, nothing to do
			return
		}
	}()

	// Create the stream with the no-deadline context
	stream, err := p.streamCreator.CreateNewVPNStream(streamCtx, peerID)

	// Cancel the stream context since we're done with it
	cancel()

	if err != nil {
		return nil, err
	}

	// Create a new stream context
	streamCtx, streamCancel := context.WithCancel(p.ctx)

	// Determine the buffer size
	bufferSize := p.packetBufferSize
	if bufferSize <= 0 {
		bufferSize = 100 // Default buffer size
	}

	// Create a new stream channel
	streamChannel := &StreamChannel{
		Stream:                   stream,
		PacketChan:               make(chan *types.QueuedPacket, bufferSize),
		lastActivity:             time.Now().UnixNano(),
		healthy:                  1, // 1 = healthy
		ctx:                      streamCtx,
		cancel:                   streamCancel,
		overflowQueue:            make([]*types.QueuedPacket, 0),
		overflowSignal:           make(chan struct{}, 1),
		overflowProcessorRunning: 0,
	}

	// Start the stream processor
	go streamChannel.ProcessPackets(peerIDStr)

	// Add the stream to the pool
	shard.Lock()
	if !exists {
		shard.peers[peerIDStr] = &peerData{
			streams:     make([]*StreamChannel, 0, 1), // Start with capacity for 1 stream
			usageCounts: make([]int64, 0, 1),
		}
	}

	peerDataValue = shard.peers[peerIDStr]
	peerDataValue.streams = append(peerDataValue.streams, streamChannel)
	peerDataValue.usageCounts = append(peerDataValue.usageCounts, 1) // Start with usage count of 1
	shard.Unlock()

	// Update metrics
	atomic.AddInt64(&p.metrics.StreamsCreated, 1)
	atomic.AddInt64(&p.metrics.ActiveStreams, 1)

	return streamChannel, nil
}

// removeStreamAtIndex removes a stream at the specified index
func (p *StreamPool) removeStreamAtIndex(peerIDStr string, index int, shard *peerShard) {
	peerData, exists := shard.peers[peerIDStr]
	if !exists || index >= len(peerData.streams) {
		return
	}

	// Close the stream
	peerData.streams[index].Close()

	// Remove the stream from the slice
	peerData.streams = append(peerData.streams[:index], peerData.streams[index+1:]...)
	peerData.usageCounts = append(peerData.usageCounts[:index], peerData.usageCounts[index+1:]...)

	// Update metrics
	atomic.AddInt64(&p.metrics.StreamsRemoved, 1)
	atomic.AddInt64(&p.metrics.ActiveStreams, -1)

	// If there are no more streams, remove the peer from the map
	if len(peerData.streams) == 0 {
		delete(shard.peers, peerIDStr)
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
			// Check if the pool is still active
			if atomic.LoadInt32(&p.active) == 0 {
				return
			}

			// Send the cleanup operation to the worker routine
			resultChan := make(chan streamOpResult, 1)
			select {
			case p.opChan <- streamOp{
				opType:     "cleanup",
				resultChan: resultChan,
			}:
				// Wait for the result
				select {
				case <-resultChan:
					// Cleanup completed
				case <-p.ctx.Done():
					// Context cancelled
					return
				}
			case <-p.ctx.Done():
				// Context cancelled
				return
			}
		}
	}
}

// cleanupIdleStreamsInternal is the internal implementation of cleanupIdleStreams
func (p *StreamPool) cleanupIdleStreamsInternal() {
	now := time.Now()

	// Process each shard
	for _, shard := range p.peerShards {
		// Get a snapshot of all peer IDs in this shard
		shard.RLock()
		peerIDs := make([]string, 0, len(shard.peers))
		for peerID := range shard.peers {
			peerIDs = append(peerIDs, peerID)
		}
		shard.RUnlock()

		// Process each peer
		for _, peerID := range peerIDs {
			shard.Lock()
			peerData, exists := shard.peers[peerID]
			if !exists {
				shard.Unlock()
				continue
			}

			// Check each stream
			for i := len(peerData.streams) - 1; i >= 0; i-- {
				stream := peerData.streams[i]
				lastActivity := stream.GetLastActivity()
				healthy := stream.IsHealthy()

				// Check if the stream is idle or unhealthy
				if !healthy || now.Sub(lastActivity) > p.streamIdleTimeout {
					p.removeStreamAtIndex(peerID, i, shard)
				}
			}
			shard.Unlock()
		}
	}
}

// GetStreamMetrics returns metrics for all streams
func (p *StreamPool) GetStreamMetrics() map[string]map[string]int64 {
	// Send the operation to the worker routine
	resultChan := make(chan streamOpResult, 1)
	p.opChan <- streamOp{
		opType:     "get_metrics",
		resultChan: resultChan,
	}

	// Wait for the result
	result := <-resultChan
	return result.metrics
}

// getStreamMetricsInternal is the internal implementation of GetStreamMetrics
func (p *StreamPool) getStreamMetricsInternal() map[string]map[string]int64 {
	metrics := make(map[string]map[string]int64)

	// Process each shard
	for _, shard := range p.peerShards {
		// Get a snapshot of all peer IDs in this shard
		shard.RLock()
		peerIDs := make([]string, 0, len(shard.peers))
		for peerID := range shard.peers {
			peerIDs = append(peerIDs, peerID)
		}
		shard.RUnlock()

		// Process each peer
		for _, peerID := range peerIDs {
			shard.RLock()
			peerData, exists := shard.peers[peerID]
			if !exists {
				shard.RUnlock()
				continue
			}

			peerMetrics := make(map[string]int64)
			var totalPackets, totalErrors, totalBytes int64

			for _, stream := range peerData.streams {
				totalPackets += atomic.LoadInt64(&stream.Metrics.PacketCount)
				totalErrors += atomic.LoadInt64(&stream.Metrics.ErrorCount)
				totalBytes += atomic.LoadInt64(&stream.Metrics.BytesSent)
			}

			peerMetrics["stream_count"] = int64(len(peerData.streams))
			peerMetrics["packet_count"] = totalPackets
			peerMetrics["error_count"] = totalErrors
			peerMetrics["bytes_sent"] = totalBytes

			metrics[peerID] = peerMetrics
			shard.RUnlock()
		}
	}

	return metrics
}

// GetStreamCount returns the number of streams for a peer
func (p *StreamPool) GetStreamCount(peerID peer.ID) int {
	// Send the operation to the worker routine
	resultChan := make(chan streamOpResult, 1)
	p.opChan <- streamOp{
		opType:     "get_stream_count",
		peerID:     peerID,
		resultChan: resultChan,
	}

	// Wait for the result
	result := <-resultChan
	return result.count
}

// getStreamCountInternal is the internal implementation of GetStreamCount
func (p *StreamPool) getStreamCountInternal(peerID peer.ID) int {
	peerIDStr := peerID.String()
	shardIdx := p.getPeerShardIndex(peerIDStr)
	shard := p.peerShards[shardIdx]

	shard.RLock()
	defer shard.RUnlock()

	peerData, exists := shard.peers[peerIDStr]
	if !exists {
		return 0
	}

	return len(peerData.streams)
}

// GetTotalStreamCount returns the total number of streams
func (p *StreamPool) GetTotalStreamCount() int {
	// Send the operation to the worker routine
	resultChan := make(chan streamOpResult, 1)
	p.opChan <- streamOp{
		opType:     "get_total_stream_count",
		resultChan: resultChan,
	}

	// Wait for the result
	result := <-resultChan
	return result.count
}

// getTotalStreamCountInternal is the internal implementation of GetTotalStreamCount
func (p *StreamPool) getTotalStreamCountInternal() int {
	count := 0

	// Process each shard
	for _, shard := range p.peerShards {
		shard.RLock()
		for _, peerData := range shard.peers {
			count += len(peerData.streams)
		}
		shard.RUnlock()
	}

	return count
}

// GetPacketBufferSize returns the packet buffer size for streams
func (p *StreamPool) GetPacketBufferSize() int {
	if p.packetBufferSize > 0 {
		return p.packetBufferSize
	}
	return 100 // Default buffer size
}

// getPeerShardIndex returns the shard index for a peer ID
func (p *StreamPool) getPeerShardIndex(peerID string) int {
	// Simple hash function to distribute peers across shards
	var hash uint32
	for i := 0; i < len(peerID); i++ {
		hash = hash*31 + uint32(peerID[i])
	}
	return int(hash % 16)
}

// streamStatsLogger periodically logs statistics about streams per peer
func (p *StreamPool) streamStatsLogger() {
	// Create a ticker for periodic logging (every 30 seconds)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			// Check if the pool is still active
			if atomic.LoadInt32(&p.active) == 0 {
				return
			}

			// Log stream stats for each peer
			p.logStreamStats()
		}
	}
}

// logStreamStats logs detailed statistics about streams per peer
func (p *StreamPool) logStreamStats() {
	// Get metrics for all peers
	metrics := p.getStreamMetricsInternal()

	// Log overall stats
	totalStreams := p.getTotalStreamCountInternal()
	streamPoolLog.WithFields(logrus.Fields{
		"total_streams":   totalStreams,
		"active_streams":  atomic.LoadInt64(&p.metrics.ActiveStreams),
		"streams_created": atomic.LoadInt64(&p.metrics.StreamsCreated),
		"streams_removed": atomic.LoadInt64(&p.metrics.StreamsRemoved),
	}).Info("Stream pool statistics")

	// Log detailed stats for each peer
	for peerID, peerMetrics := range metrics {
		// Get detailed stream information
		shardIdx := p.getPeerShardIndex(peerID)
		shard := p.peerShards[shardIdx]

		shard.RLock()
		peerData, exists := shard.peers[peerID]
		if !exists {
			shard.RUnlock()
			continue
		}

		// Create a list of stream details
		streamDetails := make([]map[string]interface{}, 0, len(peerData.streams))
		for i, stream := range peerData.streams {
			// Get stream metrics
			packetCount := atomic.LoadInt64(&stream.Metrics.PacketCount)
			errorCount := atomic.LoadInt64(&stream.Metrics.ErrorCount)
			bytesSent := atomic.LoadInt64(&stream.Metrics.BytesSent)
			bufferUtil := stream.GetBufferUtilization()
			usageCount := atomic.LoadInt64(&peerData.usageCounts[i])
			lastActivity := stream.GetLastActivity()

			// Add stream details
			streamDetails = append(streamDetails, map[string]interface{}{
				"stream_id":         fmt.Sprintf("%p", stream),
				"healthy":           stream.IsHealthy(),
				"buffer_util":       bufferUtil,
				"usage_count":       usageCount,
				"packet_count":      packetCount,
				"error_count":       errorCount,
				"bytes_sent":        bytesSent,
				"last_activity_ago": time.Since(lastActivity).String(),
			})
		}
		shard.RUnlock()

		// Log peer stream stats
		streamPoolLog.WithFields(logrus.Fields{
			"peer_id":        peerID,
			"stream_count":   peerMetrics["stream_count"],
			"packet_count":   peerMetrics["packet_count"],
			"error_count":    peerMetrics["error_count"],
			"bytes_sent":     peerMetrics["bytes_sent"],
			"stream_details": streamDetails,
		}).Info("Peer stream statistics")
	}
}

// Close implements io.Closer
func (p *StreamPool) Close() error {
	p.Stop()
	return nil
}

// Ensure StreamPool implements StreamPoolInterface
var _ StreamPoolInterface = (*StreamPool)(nil)
