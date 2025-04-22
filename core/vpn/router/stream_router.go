package router

import (
	"context"
	"fmt"
	"hash/fnv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/packet"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/pool"
	streamTypes "github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

var log = logrus.WithField("service", "vpn-stream-router")

// StreamRouter manages packet routing and stream allocation
type StreamRouter struct {
	// Core components
	streamPool    pool.PoolServiceExtension
	peerDiscovery api.PeerDiscoveryService

	// Connection management
	connectionCache    *ShardedCache
	connectionToWorker map[string]int // Maps connection keys to worker IDs
	connectionMapMu    sync.RWMutex

	// Stream management
	streamsPerPeer map[peer.ID]int
	streamsMu      sync.RWMutex

	// Scaling configuration
	minStreamsPerPeer   int     // Default: 1
	maxStreamsPerPeer   int     // Default: 10
	throughputThreshold int64   // Packets per second per stream
	scaleUpThreshold    float64 // e.g., 0.8 (80% of throughputThreshold)
	scaleDownThreshold  float64 // e.g., 0.3 (30% of throughputThreshold)
	scalingInterval     time.Duration

	// Metrics for scaling decisions
	metrics *RouterMetrics

	// Workers for packet processing
	workerPool *WorkerPool

	// Context for lifecycle management
	ctx    context.Context
	cancel context.CancelFunc

	// Configuration
	config *StreamRouterConfig
}

// ConnectionRoute represents routing information for a specific connection
type ConnectionRoute struct {
	connKey      string
	peerID       peer.ID
	streamIndex  int
	lastActivity time.Time
	packetCount  int64
	workerID     int // Assigned worker for processing
}

// PacketTask represents a packet to be processed
type PacketTask struct {
	packet []byte
	route  *ConnectionRoute
	ctx    context.Context
	doneCh chan error
}

// StreamRouterConfig with worker pool configuration
type StreamRouterConfig struct {
	// Stream configuration
	MinStreamsPerPeer   int           // Default: 1
	MaxStreamsPerPeer   int           // Default: 10
	ThroughputThreshold int64         // Packets per second per stream
	ScaleUpThreshold    float64       // Default: 0.8
	ScaleDownThreshold  float64       // Default: 0.3
	ScalingInterval     time.Duration // Default: 5 seconds

	// Worker pool configuration
	MinWorkers               int           // Default: 4
	MaxWorkers               int           // Default: 32
	InitialWorkers           int           // Default: 8
	WorkerQueueSize          int           // Default: 1000
	WorkerScaleInterval      time.Duration // Default: 10 seconds
	WorkerScaleUpThreshold   float64       // Default: 0.75
	WorkerScaleDownThreshold float64       // Default: 0.25

	// Connection cache configuration
	ConnectionTTL   time.Duration // Default: 30 seconds (shortened)
	CleanupInterval time.Duration // Default: 10 seconds (shortened)
	CacheShardCount int           // Default: 16
}

// DefaultStreamRouterConfig returns a default configuration
func DefaultStreamRouterConfig() *StreamRouterConfig {
	return &StreamRouterConfig{
		MinStreamsPerPeer:        1,
		MaxStreamsPerPeer:        10,
		ThroughputThreshold:      1000, // 1000 packets per second per stream
		ScaleUpThreshold:         0.8,
		ScaleDownThreshold:       0.3,
		ScalingInterval:          5 * time.Second,
		MinWorkers:               4,
		MaxWorkers:               32,
		InitialWorkers:           8,
		WorkerQueueSize:          1000,
		WorkerScaleInterval:      10 * time.Second,
		WorkerScaleUpThreshold:   0.75,
		WorkerScaleDownThreshold: 0.25,
		ConnectionTTL:            30 * time.Second, // Shortened for frequent cleanup
		CleanupInterval:          10 * time.Second, // Shortened for frequent cleanup
		CacheShardCount:          16,
	}
}

// NewStreamRouter creates a new stream router
func NewStreamRouter(config *StreamRouterConfig, streamPool pool.PoolServiceExtension, peerDiscovery api.PeerDiscoveryService) *StreamRouter {
	if config == nil {
		config = DefaultStreamRouterConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create worker pool configuration
	workerPoolConfig := &WorkerPoolConfig{
		MinWorkers:         config.MinWorkers,
		MaxWorkers:         config.MaxWorkers,
		InitialWorkers:     config.InitialWorkers,
		QueueSize:          config.WorkerQueueSize,
		ScaleInterval:      config.WorkerScaleInterval,
		ScaleUpThreshold:   config.WorkerScaleUpThreshold,
		ScaleDownThreshold: config.WorkerScaleDownThreshold,
	}

	// Create worker pool
	workerPool := NewWorkerPool(workerPoolConfig)

	// Create sharded cache for connections
	connectionCache := NewShardedCache(
		config.CacheShardCount,
		config.ConnectionTTL,
		config.CleanupInterval,
	)

	router := &StreamRouter{
		streamPool:          streamPool,
		peerDiscovery:       peerDiscovery,
		connectionCache:     connectionCache,
		connectionToWorker:  make(map[string]int),
		streamsPerPeer:      make(map[peer.ID]int),
		minStreamsPerPeer:   config.MinStreamsPerPeer,
		maxStreamsPerPeer:   config.MaxStreamsPerPeer,
		throughputThreshold: config.ThroughputThreshold,
		scaleUpThreshold:    config.ScaleUpThreshold,
		scaleDownThreshold:  config.ScaleDownThreshold,
		scalingInterval:     config.ScalingInterval,
		metrics:             NewRouterMetrics(),
		workerPool:          workerPool,
		ctx:                 ctx,
		cancel:              cancel,
		config:              config,
	}

	// Set up cache eviction callback
	router.connectionCache.OnEvicted(func(connKey string, value interface{}) {
		router.connectionMapMu.Lock()
		delete(router.connectionToWorker, connKey)
		router.connectionMapMu.Unlock()
	})

	// Initialize worker pool with reference to router
	workerPool.Initialize(router)

	// Start background routines
	go router.startThroughputMonitoring()

	log.Info("StreamRouter initialized with config: ", config)
	return router
}

// DispatchPacket dispatches a packet to the appropriate worker
func (r *StreamRouter) DispatchPacket(ctx context.Context, packetData []byte) error {
	// Extract packet info
	packetInfo, err := packet.ExtractIPAndPorts(packetData)
	if err != nil {
		return err
	}

	// Get destination IP and create connection key
	destIP := packetInfo.DstIP.String()

	// Create a connection key that includes the full tuple
	var connKey string
	if packetInfo.SrcPort == nil || packetInfo.DstPort == nil {
		// For protocols without ports (like ICMP), use protocol number instead
		connKey = fmt.Sprintf("%s:p%d:%s:p%d",
			packetInfo.SrcIP.String(),
			packetInfo.Protocol,
			destIP,
			packetInfo.Protocol)
	} else {
		// For protocols with ports (like TCP/UDP), use the standard format
		connKey = fmt.Sprintf("%s:%d:%s:%d",
			packetInfo.SrcIP.String(),
			*packetInfo.SrcPort,
			destIP,
			*packetInfo.DstPort)
	}

	// Get peer ID for destination IP
	peerID, err := r.peerDiscovery.GetPeerID(ctx, destIP)
	if err != nil {
		return err
	}

	parsedPeerID, err := peer.Decode(peerID)
	if err != nil {
		return err
	}

	// Get or create route for this connection
	route, err := r.getOrCreateRoute(connKey, parsedPeerID)
	if err != nil {
		return err
	}

	// Create packet task
	task := &PacketTask{
		packet: packetData,
		route:  route,
		ctx:    ctx,
		doneCh: make(chan error, 1),
	}

	// Get the assigned worker - this will NEVER change for a given connKey
	worker := r.workerPool.GetWorker(route.workerID)

	// Try to send to worker queue with timeout
	select {
	case worker.packetChan <- task:
		// Successfully queued
	case <-time.After(50 * time.Millisecond):
		// Worker queue full - implement backpressure
		return fmt.Errorf("worker queue full for connection %s, packet dropped", connKey)
	}

	// Wait for completion or timeout
	select {
	case err := <-task.doneCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// getOrCreateRoute gets or creates a route for a connection
func (r *StreamRouter) getOrCreateRoute(connKey string, peerID peer.ID) (*ConnectionRoute, error) {
	// Try to get from cache
	if routeObj, found := r.connectionCache.Get(connKey); found {
		route := routeObj.(*ConnectionRoute)
		route.lastActivity = time.Now()
		return route, nil
	}

	// Check if we already have a worker assignment for this connection
	r.connectionMapMu.RLock()
	workerID, exists := r.connectionToWorker[connKey]
	r.connectionMapMu.RUnlock()

	if !exists {
		// Create new worker assignment using consistent hashing
		h := fnv.New32a()
		h.Write([]byte(connKey))
		workerID = int(h.Sum32() % uint32(r.workerPool.GetWorkerCount()))

		// Store the assignment
		r.connectionMapMu.Lock()
		r.connectionToWorker[connKey] = workerID
		r.connectionMapMu.Unlock()
	}

	// Initialize peer stream count if needed
	r.streamsMu.RLock()
	streamCount := r.streamsPerPeer[peerID]
	if streamCount == 0 {
		r.streamsMu.RUnlock()
		r.streamsMu.Lock()
		r.streamsPerPeer[peerID] = r.minStreamsPerPeer
		streamCount = r.minStreamsPerPeer
		r.streamsMu.Unlock()
	} else {
		r.streamsMu.RUnlock()
	}

	// Determine stream index using consistent hash
	streamIndex := r.getStreamIndexForConnection(connKey, streamCount)

	// Create new route with the permanent worker ID
	route := &ConnectionRoute{
		connKey:      connKey,
		peerID:       peerID,
		streamIndex:  streamIndex,
		lastActivity: time.Now(),
		packetCount:  0,
		workerID:     workerID,
	}

	// Store with TTL
	r.connectionCache.Set(connKey, route, cache.DefaultExpiration)

	return route, nil
}

// getStreamIndexForConnection determines which stream to use for a connection
func (r *StreamRouter) getStreamIndexForConnection(connKey string, streamCount int) int {
	// Use consistent hash to determine stream index
	h := fnv.New32a()
	h.Write([]byte(connKey))
	return int(h.Sum32() % uint32(streamCount))
}

// getStreamForRoute gets the appropriate stream for a route
func (r *StreamRouter) getStreamForRoute(route *ConnectionRoute) (streamTypes.VPNStream, error) {
	// Get current stream count for this peer
	r.streamsMu.RLock()
	streamCount := r.streamsPerPeer[route.peerID]
	if streamCount == 0 {
		streamCount = r.minStreamsPerPeer
	}
	r.streamsMu.RUnlock()

	// Ensure stream index is valid
	if route.streamIndex >= streamCount {
		route.streamIndex = route.streamIndex % streamCount
	}

	// Get stream from pool
	stream, err := r.streamPool.GetStreamByIndex(r.ctx, route.peerID, route.streamIndex)
	if err != nil {
		return nil, err
	}

	return stream, nil
}

// releaseStream releases a stream back to the pool
func (r *StreamRouter) releaseStream(peerID peer.ID, streamIndex int) {
	r.streamPool.ReleaseStreamByIndex(peerID, streamIndex, false)
}

// recordPacket records a packet for metrics
func (r *StreamRouter) recordPacket(route *ConnectionRoute) {
	// Update route stats
	atomic.AddInt64(&route.packetCount, 1)

	// Update metrics
	r.metrics.recordPacket(route.peerID, route.streamIndex)
}

// startThroughputMonitoring starts the throughput monitoring routine
func (r *StreamRouter) startThroughputMonitoring() {
	ticker := time.NewTicker(r.scalingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			r.updateThroughputMetrics()
			r.evaluateScaling()
		}
	}
}

// updateThroughputMetrics updates throughput metrics
func (r *StreamRouter) updateThroughputMetrics() {
	r.metrics.updateThroughput()
}

// evaluateScaling evaluates whether to scale streams up or down
func (r *StreamRouter) evaluateScaling() {
	r.metrics.mu.RLock()
	peerMetrics := make(map[peer.ID]struct {
		totalThroughput        float64
		streamCount            int
		avgThroughputPerStream float64
	})

	// Calculate metrics for each peer
	for peerID, throughput := range r.metrics.peerThroughput {
		r.streamsMu.RLock()
		streamCount := r.streamsPerPeer[peerID]
		if streamCount == 0 {
			streamCount = r.minStreamsPerPeer
		}
		r.streamsMu.RUnlock()

		avgThroughput := throughput / float64(streamCount)

		peerMetrics[peerID] = struct {
			totalThroughput        float64
			streamCount            int
			avgThroughputPerStream float64
		}{
			totalThroughput:        throughput,
			streamCount:            streamCount,
			avgThroughputPerStream: avgThroughput,
		}
	}
	r.metrics.mu.RUnlock()

	// Evaluate scaling for each peer
	for peerID, metrics := range peerMetrics {
		// Scale up if average throughput per stream is above threshold
		if metrics.avgThroughputPerStream > float64(r.throughputThreshold)*r.scaleUpThreshold &&
			metrics.streamCount < r.maxStreamsPerPeer {
			r.scaleStreamsForPeer(peerID, metrics.streamCount+1)
			log.Infof("Scaling up streams for peer %s from %d to %d (avg throughput: %.2f packets/sec/stream)",
				peerID.String(), metrics.streamCount, metrics.streamCount+1, metrics.avgThroughputPerStream)
		}

		// Scale down if average throughput per stream is below threshold
		// and we have more than minimum streams
		if metrics.avgThroughputPerStream < float64(r.throughputThreshold)*r.scaleDownThreshold &&
			metrics.streamCount > r.minStreamsPerPeer {
			r.scaleStreamsForPeer(peerID, metrics.streamCount-1)
			log.Infof("Scaling down streams for peer %s from %d to %d (avg throughput: %.2f packets/sec/stream)",
				peerID.String(), metrics.streamCount, metrics.streamCount-1, metrics.avgThroughputPerStream)
		}
	}
}

// scaleStreamsForPeer scales the number of streams for a peer
func (r *StreamRouter) scaleStreamsForPeer(peerID peer.ID, newCount int) {
	// Ensure new count is within bounds
	if newCount < r.minStreamsPerPeer {
		newCount = r.minStreamsPerPeer
	}
	if newCount > r.maxStreamsPerPeer {
		newCount = r.maxStreamsPerPeer
	}

	r.streamsMu.Lock()
	oldCount := r.streamsPerPeer[peerID]
	if oldCount == 0 {
		oldCount = r.minStreamsPerPeer
	}
	r.streamsPerPeer[peerID] = newCount
	r.streamsMu.Unlock()

	// If scaling down, we need to rebalance connections
	if newCount < oldCount {
		r.rebalanceConnectionsForPeer(peerID, oldCount, newCount)
	}

	// Notify the stream pool of the new target
	r.streamPool.SetTargetStreamsForPeer(peerID, newCount)
}

// rebalanceConnectionsForPeer rebalances connections when scaling down streams
func (r *StreamRouter) rebalanceConnectionsForPeer(peerID peer.ID, oldCount, newCount int) {
	// Find all connections for this peer that need rebalancing
	r.connectionMapMu.RLock()
	affectedConnKeys := make([]string, 0)

	// Get all connection keys from the cache
	for _, shard := range r.connectionCache.shards {
		shard.mu.RLock()
		for key, item := range shard.cache.Items() {
			if route, ok := item.Object.(*ConnectionRoute); ok && route.peerID == peerID && route.streamIndex >= newCount {
				affectedConnKeys = append(affectedConnKeys, key)
			}
		}
		shard.mu.RUnlock()
	}
	r.connectionMapMu.RUnlock()

	// Rebalance affected connections
	for _, connKey := range affectedConnKeys {
		if routeObj, found := r.connectionCache.Get(connKey); found {
			route := routeObj.(*ConnectionRoute)

			// Recalculate stream index
			route.streamIndex = route.streamIndex % newCount

			// Update in cache
			r.connectionCache.Set(connKey, route, cache.DefaultExpiration)
		}
	}

	if len(affectedConnKeys) > 0 {
		log.Infof("Rebalanced %d connections for peer %s after scaling down streams from %d to %d",
			len(affectedConnKeys), peerID.String(), oldCount, newCount)
	}
}

// Shutdown gracefully shuts down the router and all workers
func (r *StreamRouter) Shutdown() {
	log.Info("Shutting down StreamRouter")

	// Cancel context to signal all goroutines
	r.cancel()

	// Wait for workers to finish
	r.workerPool.Shutdown()

	log.Info("StreamRouter shutdown complete")
}
