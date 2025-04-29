package pool

import (
	"context"
	"fmt"
	"runtime"
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
	maxStreamsPerPeer   int
	streamIdleTimeout   time.Duration
	cleanupInterval     time.Duration
	packetBufferSize    int
	usageCountWeight    float64
	bufferUtilWeight    float64
	bufferUtilThreshold int
	usageCountThreshold int

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
	peerShards []*peerShard
	shardMask  uint32 // Mask for fast modulo operation
}

// StreamPoolConfig contains configuration for the stream pool
type StreamPoolConfig struct {
	MaxStreamsPerPeer   int
	StreamIdleTimeout   time.Duration
	CleanupInterval     time.Duration
	PacketBufferSize    int
	UsageCountWeight    float64 // Weight for usage count in load score calculation (default: 0.7)
	BufferUtilWeight    float64 // Weight for buffer utilization in load score calculation (default: 0.3)
	BufferUtilThreshold int     // Buffer utilization threshold percentage (default: 70)
	UsageCountThreshold int     // Usage count threshold (default: 100)
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

	// Calculate optimal shard count based on CPU cores
	// Use a power of 2 for efficient modulo operations with bit masking
	numCPU := runtime.NumCPU()
	shardCount := 16 // Minimum shard count

	// Scale up shards based on CPU count, but keep it a power of 2
	if numCPU > 4 {
		shardCount = 32
	}
	if numCPU > 8 {
		shardCount = 64
	}
	if numCPU > 16 {
		shardCount = 128
	}

	// Calculate the bit mask for fast modulo operations
	shardMask := uint32(shardCount - 1)

	// Initialize the peer shards
	peerShards := make([]*peerShard, shardCount)
	for i := 0; i < shardCount; i++ {
		peerShards[i] = &peerShard{
			peers: make(map[string]*peerData),
		}
	}

	// Set default values for new configuration parameters if not provided
	usageCountWeight := config.UsageCountWeight
	if usageCountWeight <= 0 {
		usageCountWeight = 0.7 // Default value
	}

	bufferUtilWeight := config.BufferUtilWeight
	if bufferUtilWeight <= 0 {
		bufferUtilWeight = 0.3 // Default value
	}

	bufferUtilThreshold := config.BufferUtilThreshold
	if bufferUtilThreshold <= 0 {
		bufferUtilThreshold = 70 // Default value
	}

	usageCountThreshold := config.UsageCountThreshold
	if usageCountThreshold <= 0 {
		usageCountThreshold = 100 // Default value
	}

	pool := &StreamPool{
		streamCreator:       streamCreator,
		ctx:                 ctx,
		cancel:              cancel,
		maxStreamsPerPeer:   config.MaxStreamsPerPeer,
		streamIdleTimeout:   config.StreamIdleTimeout,
		cleanupInterval:     config.CleanupInterval,
		packetBufferSize:    config.PacketBufferSize,
		usageCountWeight:    usageCountWeight,
		bufferUtilWeight:    bufferUtilWeight,
		bufferUtilThreshold: bufferUtilThreshold,
		usageCountThreshold: usageCountThreshold,
		opChan:              make(chan streamOp, 100), // Buffer size for operations
		peerShards:          peerShards,
		shardMask:           shardMask,
	}

	streamPoolLog.WithField("shard_count", shardCount).Info("Initialized stream pool with dynamic sharding")
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

	// First try to find a healthy stream using the non-blocking algorithm
	stream := p.findLeastLoadedStreamNonBlocking(peerID)
	if stream != nil {
		return stream, nil
	}

	// If no healthy stream was found, check if we need to create a new one
	shard.RLock()
	peerData, exists := shard.peers[peerIDStr]
	var streamCount int
	var streamNearingCapacity bool

	if exists {
		streamCount = len(peerData.streams)

		// Check if any existing stream is nearing capacity
		for i, stream := range peerData.streams {
			if !stream.IsHealthy() {
				continue
			}

			// Get metrics without holding locks
			bufferUtil := stream.GetBufferUtilization()
			count := atomic.LoadInt64(&peerData.usageCounts[i])

			// Check if this stream is nearing capacity
			if bufferUtil > p.bufferUtilThreshold || count > int64(p.usageCountThreshold) {
				streamNearingCapacity = true
				break
			}
		}
	}
	shard.RUnlock()

	// If this is the first time we're seeing this peer or all streams are unhealthy
	if !exists || streamCount == 0 {
		// Create a new stream directly
		return p.createNewStreamChannel(ctx, peerID)
	}

	// Check if we should create a new stream based on resource utilization
	if streamNearingCapacity && streamCount < p.maxStreamsPerPeer && p.maxStreamsPerPeer > 1 {
		// Log the decision to create a new stream
		streamPoolLog.WithFields(logrus.Fields{
			"peer_id":         peerIDStr,
			"current_streams": streamCount,
			"max_streams":     p.maxStreamsPerPeer,
		}).Info("Creating new stream due to resource utilization")

		// Create a new stream channel
		return p.createNewStreamChannel(ctx, peerID)
	}

	// Try one more time with the non-blocking algorithm
	// This handles the case where a stream became healthy while we were checking
	stream = p.findLeastLoadedStreamNonBlocking(peerID)
	if stream != nil {
		return stream, nil
	}

	// Create a new stream channel as a last resort
	return p.createNewStreamChannel(ctx, peerID)
}

// findLeastLoadedStreamNonBlocking finds the least loaded healthy stream without locking
// Returns nil if no healthy stream is found
func (p *StreamPool) findLeastLoadedStreamNonBlocking(peerID peer.ID) *StreamChannel {
	peerIDStr := peerID.String()
	shardIdx := p.getPeerShardIndex(peerIDStr)
	shard := p.peerShards[shardIdx]

	// Take a snapshot of the current streams with minimal locking
	shard.RLock()
	peerData, exists := shard.peers[peerIDStr]
	if !exists || len(peerData.streams) == 0 {
		shard.RUnlock()
		return nil
	}

	// Create local copies to work with outside the lock
	streams := make([]*StreamChannel, len(peerData.streams))
	copy(streams, peerData.streams)
	shard.RUnlock()

	// Find the least loaded healthy stream
	var minLoadScore float64 = float64(^uint64(0) >> 1) // Max float64
	var selectedStream *StreamChannel
	var selectedIndex int = -1

	for i, stream := range streams {
		// Skip unhealthy streams
		if !stream.IsHealthy() {
			continue
		}

		// Get metrics without holding locks
		bufferUtil := stream.GetBufferUtilization()

		// Get usage count with atomic operation
		shard.RLock()
		count := atomic.LoadInt64(&peerData.usageCounts[i])
		shard.RUnlock()

		// Calculate load score
		loadScore := (float64(count) * p.usageCountWeight) + (float64(bufferUtil) * p.bufferUtilWeight)

		// Update if this is the least loaded stream
		if loadScore < minLoadScore {
			minLoadScore = loadScore
			selectedStream = stream
			selectedIndex = i
		}
	}

	// If we found a healthy stream, increment its usage count
	if selectedStream != nil && selectedIndex >= 0 {
		shard.RLock()
		if selectedIndex < len(peerData.usageCounts) {
			atomic.AddInt64(&peerData.usageCounts[selectedIndex], 1)
		}
		shard.RUnlock()
		return selectedStream
	}

	return nil
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
		overflowQueue:            NewLockFreeQueue(),
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
	// Use bit masking for faster modulo operation (works because shardCount is a power of 2)
	return int(hash & p.shardMask)
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
