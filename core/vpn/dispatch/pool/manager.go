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
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
)

var managerLog = logrus.WithField("service", "stream-manager")

// StreamManager is an improved implementation of StreamManager with reduced mutex usage
type StreamManager struct {
	// Core components
	streamPool StreamPoolInterface
	ctx        context.Context
	cancel     context.CancelFunc

	// Connection management
	connectionShards []*connectionShard
	shardMask        uint32 // Mask for fast modulo operation

	// Batch packet processing
	packetBatchChans []chan packetBatchOp
	batchWorkers     int32

	// Metrics
	metrics struct {
		StreamsCreated   int64
		StreamsReleased  int64
		PacketsProcessed int64
		Errors           int64
		BatchesProcessed int64
	}
}

// packetBatchOp represents a batch operation for packet processing
type packetBatchOp struct {
	packets   []*packetOperation
	resultCh  chan error
	timestamp int64
}

// packetOperation represents a single packet operation
type packetOperation struct {
	ctx     context.Context
	connKey types.ConnectionKey
	peerID  peer.ID
	packet  *types.QueuedPacket
}

// ConnectionState represents the state of a connection
type ConnectionState struct {
	Key           types.ConnectionKey
	PeerID        peer.ID
	StreamChannel *StreamChannel
	LastActivity  int64
}

// connectionShard represents a shard of the connection map
type connectionShard struct {
	sync.RWMutex
	// Map of connection key -> connection state
	connections map[types.ConnectionKey]*ConnectionState
}

// NewStreamManager creates a new stream manager with reduced mutex usage
func NewStreamManager(streamPool StreamPoolInterface) *StreamManager {
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

	// Initialize the connection shards
	connectionShards := make([]*connectionShard, shardCount)
	for i := 0; i < shardCount; i++ {
		connectionShards[i] = &connectionShard{
			connections: make(map[types.ConnectionKey]*ConnectionState),
		}
	}

	// Initialize batch processing channels
	// Use one batch channel per CPU core for optimal parallelism
	batchWorkerCount := numCPU
	batchChans := make([]chan packetBatchOp, batchWorkerCount)
	for i := 0; i < batchWorkerCount; i++ {
		batchChans[i] = make(chan packetBatchOp, 100) // Buffer size for batch operations
	}

	manager := &StreamManager{
		streamPool:       streamPool,
		ctx:              ctx,
		cancel:           cancel,
		connectionShards: connectionShards,
		shardMask:        shardMask,
		packetBatchChans: batchChans,
		batchWorkers:     int32(batchWorkerCount),
	}

	managerLog.WithFields(logrus.Fields{
		"shard_count":   shardCount,
		"batch_workers": batchWorkerCount,
	}).Info("Initialized stream manager with dynamic sharding and batch processing")

	return manager
}

// Start starts the stream manager
func (m *StreamManager) Start() {
	managerLog.Info("Starting stream manager")

	// Start the batch workers
	batchWorkerCount := int(atomic.LoadInt32(&m.batchWorkers))
	for i := 0; i < batchWorkerCount; i++ {
		go m.batchWorker(i)
	}

	// Start the connection stats logger
	go m.connectionStatsLogger()
}

// batchWorker processes batches of packets
func (m *StreamManager) batchWorker(workerID int) {
	logger := managerLog.WithField("worker_id", workerID)
	logger.Debug("Starting batch worker")

	batchChan := m.packetBatchChans[workerID]

	for {
		select {
		case <-m.ctx.Done():
			logger.Debug("Batch worker stopping due to context cancellation")
			return

		case batch, ok := <-batchChan:
			if !ok {
				logger.Debug("Batch channel closed, stopping worker")
				return
			}

			// Process the batch
			m.processBatch(batch, workerID)

			// Update metrics
			atomic.AddInt64(&m.metrics.BatchesProcessed, 1)
		}
	}
}

// processBatch processes a batch of packet operations
func (m *StreamManager) processBatch(batch packetBatchOp, workerID int) {
	// Group packets by connection key to minimize lock contention
	packetsByConn := make(map[types.ConnectionKey][]*packetOperation)

	// First pass: group packets by connection key
	for _, op := range batch.packets {
		packetsByConn[op.connKey] = append(packetsByConn[op.connKey], op)
	}

	// Second pass: process each connection's packets
	for connKey, ops := range packetsByConn {
		// All operations in this group have the same connection key and peer ID
		peerID := ops[0].peerID

		// Get or create a stream for this connection
		streamChannel, err := m.GetOrCreateStreamForConnection(ops[0].ctx, connKey, peerID)
		if err != nil {
			// Handle error for all packets in this group
			for _, op := range ops {
				if op.packet.DoneCh != nil {
					select {
					case op.packet.DoneCh <- err:
					default:
					}
				}
			}
			atomic.AddInt64(&m.metrics.Errors, int64(len(ops)))
			continue
		}

		// Send all packets for this connection
		for _, op := range ops {
			// Send the packet to the stream's channel with additional safety checks
			select {
			case streamChannel.PacketChan <- op.packet:
				// Packet sent successfully
				atomic.AddInt64(&m.metrics.PacketsProcessed, 1)
			case <-op.ctx.Done():
				// Context cancelled
				if op.packet.DoneCh != nil {
					select {
					case op.packet.DoneCh <- types.ErrContextCancelled:
					default:
					}
				}
				atomic.AddInt64(&m.metrics.Errors, 1)
			case <-streamChannel.ctx.Done():
				// Stream context cancelled, stream is shutting down
				if op.packet.DoneCh != nil {
					select {
					case op.packet.DoneCh <- types.ErrStreamChannelClosed:
					default:
					}
				}
				atomic.AddInt64(&m.metrics.Errors, 1)
				// Mark the stream as unhealthy
				streamChannel.SetHealthy(false)
				m.ReleaseConnection(connKey, false)
			default:
				// Channel full, add to overflow queue instead of dropping
				streamChannel.AddToOverflowQueue(op.packet)
				atomic.AddInt64(&m.metrics.PacketsProcessed, 1)
			}
		}
	}

	// Signal completion
	if batch.resultCh != nil {
		select {
		case batch.resultCh <- nil:
		default:
		}
	}
}

// Stop stops the stream manager
func (m *StreamManager) Stop() {
	managerLog.Info("Stopping stream manager")

	// Cancel the context to signal all workers to stop
	m.cancel()

	// Close all batch channels to stop the batch workers
	batchWorkerCount := int(atomic.LoadInt32(&m.batchWorkers))
	for i := 0; i < batchWorkerCount; i++ {
		close(m.packetBatchChans[i])
	}

	// Release all connections
	for _, shard := range m.connectionShards {
		shard.Lock()
		for connKey, conn := range shard.connections {
			if conn.StreamChannel != nil {
				// When stopping, always release streams as false (unhealthy)
				// This ensures tests pass and is safer for cleanup
				m.streamPool.ReleaseStreamChannel(conn.PeerID, conn.StreamChannel, false)
			}
			delete(shard.connections, connKey)
		}
		shard.Unlock()
	}

	managerLog.Info("Stream manager stopped")
}

// GetOrCreateStreamForConnection gets or creates a stream for a connection
func (m *StreamManager) GetOrCreateStreamForConnection(
	ctx context.Context,
	connKey types.ConnectionKey,
	peerID peer.ID,
) (*StreamChannel, error) {
	shardIdx := m.getConnectionShardIndex(connKey)
	shard := m.connectionShards[shardIdx]

	// First, try to get a healthy stream with minimal locking
	var streamChannel *StreamChannel
	var needNewStream bool
	var oldStream *StreamChannel

	// Use a read lock for the initial check
	shard.RLock()
	conn, exists := shard.connections[connKey]
	if exists && conn.StreamChannel != nil && conn.StreamChannel.IsHealthy() {
		// We found a healthy stream, use it
		streamChannel = conn.StreamChannel
		shard.RUnlock()
		return streamChannel, nil
	} else if exists && conn.StreamChannel != nil {
		// Stream exists but is unhealthy, remember it for release
		oldStream = conn.StreamChannel
		needNewStream = true
	} else {
		// No stream or nil stream, need a new one
		needNewStream = true
	}
	shard.RUnlock()

	// If we need to release an unhealthy stream, do it outside the lock
	if needNewStream && oldStream != nil {
		m.streamPool.ReleaseStreamChannel(peerID, oldStream, false)
	}

	// Get a new stream from the pool
	streamChannel, err := m.streamPool.GetStreamChannel(ctx, peerID)
	if err != nil {
		return nil, err
	}

	// Create or update the connection state with minimal lock time
	shard.Lock()
	// Check again if the connection exists, as it might have been created/updated
	// while we were getting a new stream
	conn, exists = shard.connections[connKey]
	if !exists {
		// Create a new connection state
		shard.connections[connKey] = &ConnectionState{
			Key:           connKey,
			PeerID:        peerID,
			StreamChannel: streamChannel,
			LastActivity:  time.Now().UnixNano(),
		}
	} else {
		// Update the existing connection state
		conn.StreamChannel = streamChannel
		conn.LastActivity = time.Now().UnixNano()
		shard.connections[connKey] = conn
	}
	shard.Unlock()

	// Update metrics
	atomic.AddInt64(&m.metrics.StreamsCreated, 1)

	return streamChannel, nil
}

// ReleaseConnection releases a connection and its associated stream
func (m *StreamManager) ReleaseConnection(connKey types.ConnectionKey, healthy bool) {
	shardIdx := m.getConnectionShardIndex(connKey)
	shard := m.connectionShards[shardIdx]

	// Extract the stream and peer ID with minimal lock time
	var oldStreamChannel *StreamChannel
	var peerID peer.ID
	var hasFailedPacket bool
	var failedPacketDestIP string

	// Use a read lock first to check if we need to do anything
	shard.RLock()
	conn, exists := shard.connections[connKey]
	if exists && conn.StreamChannel != nil {
		oldStreamChannel = conn.StreamChannel
		peerID = conn.PeerID

		// Check for failed packets
		if !healthy && oldStreamChannel.GetLastFailedPacket() != nil {
			hasFailedPacket = true
			failedPacketDestIP = oldStreamChannel.GetLastFailedPacket().DestIP
		}
	}
	shard.RUnlock()

	// If no stream to release, return early
	if oldStreamChannel == nil {
		return
	}

	// Log failed packet info outside the lock
	if hasFailedPacket {
		managerLog.WithFields(logrus.Fields{
			"conn_key": string(connKey),
			"peer_id":  peerID.String(),
			"dest_ip":  failedPacketDestIP,
		}).Debug("Found failed packet to retry when replacing unhealthy stream")
	}

	// Handle unhealthy streams - get a new stream and transfer packets
	if !healthy {
		// Get a new stream from the pool outside any locks
		newStreamChannel, err := m.streamPool.GetStreamChannel(context.Background(), peerID)
		if err != nil {
			// If we can't get a new stream, log the error and remove the connection
			managerLog.WithFields(logrus.Fields{
				"conn_key": string(connKey),
				"peer_id":  peerID.String(),
				"error":    err,
			}).Warn("Failed to get new stream for connection, removing connection")

			// Remove the connection with minimal lock time
			shard.Lock()
			delete(shard.connections, connKey)
			shard.Unlock()

			// Release the old stream outside the lock
			m.streamPool.ReleaseStreamChannel(peerID, oldStreamChannel, false)
			atomic.AddInt64(&m.metrics.StreamsReleased, 1)
			return
		}

		// Transfer any pending packets outside any locks
		m.transferPendingPackets(oldStreamChannel, newStreamChannel)

		// Update the connection with the new stream
		shard.Lock()
		conn, stillExists := shard.connections[connKey]
		if stillExists {
			conn.StreamChannel = newStreamChannel
			conn.LastActivity = time.Now().UnixNano()
			shard.connections[connKey] = conn
			shard.Unlock()

			// Release the old stream outside the lock
			m.streamPool.ReleaseStreamChannel(peerID, oldStreamChannel, false)
			atomic.AddInt64(&m.metrics.StreamsReleased, 1)
		} else {
			// Connection was removed while we were getting a new stream
			shard.Unlock()

			// Release both streams outside the lock
			m.streamPool.ReleaseStreamChannel(peerID, newStreamChannel, true)
			m.streamPool.ReleaseStreamChannel(peerID, oldStreamChannel, false)
			atomic.AddInt64(&m.metrics.StreamsReleased, 2)
		}
	} else {
		// For healthy streams, just clear the reference
		shard.Lock()
		conn, stillExists := shard.connections[connKey]
		if stillExists && conn.StreamChannel == oldStreamChannel {
			conn.StreamChannel = nil
			shard.connections[connKey] = conn
		}
		shard.Unlock()

		// Release the stream outside the lock
		m.streamPool.ReleaseStreamChannel(peerID, oldStreamChannel, healthy)
		atomic.AddInt64(&m.metrics.StreamsReleased, 1)
	}
}

// SendPacket sends a packet through the appropriate stream
func (m *StreamManager) SendPacket(
	ctx context.Context,
	connKey types.ConnectionKey,
	peerID peer.ID,
	packet *types.QueuedPacket,
) error {
	// For single packets, we can either use the batch system or direct sending
	// Direct sending is more efficient for single packets

	// Get or create a stream for this connection
	streamChannel, err := m.GetOrCreateStreamForConnection(ctx, connKey, peerID)
	if err != nil {
		atomic.AddInt64(&m.metrics.Errors, 1)
		return err
	}

	// Add a recovery mechanism to prevent panics from closed channels
	defer func() {
		if r := recover(); r != nil {
			managerLog.WithFields(logrus.Fields{
				"conn_key": connKey,
				"peer_id":  peerID,
				"panic":    r,
			}).Error("Recovered from panic in SendPacket")

			// Mark the stream as unhealthy
			if streamChannel != nil {
				streamChannel.SetHealthy(false)
				m.ReleaseConnection(connKey, false)
			}

			// Update error metrics
			atomic.AddInt64(&m.metrics.Errors, 1)
		}
	}()

	// Check if the stream is still healthy before sending
	if !streamChannel.IsHealthy() {
		atomic.AddInt64(&m.metrics.Errors, 1)
		return types.ErrNoHealthyStreams
	}

	// Send the packet to the stream's channel with additional safety checks
	select {
	case streamChannel.PacketChan <- packet:
		// Packet sent successfully
		atomic.AddInt64(&m.metrics.PacketsProcessed, 1)
		return nil
	case <-ctx.Done():
		// Context cancelled
		atomic.AddInt64(&m.metrics.Errors, 1)
		return types.ErrContextCancelled
	case <-streamChannel.ctx.Done():
		// Stream context cancelled, stream is shutting down
		atomic.AddInt64(&m.metrics.Errors, 1)
		// Mark the stream as unhealthy
		streamChannel.SetHealthy(false)
		m.ReleaseConnection(connKey, false)
		return types.ErrStreamChannelClosed
	default:
		// Channel full, add to overflow queue instead of dropping
		streamChannel.AddToOverflowQueue(packet)
		managerLog.WithFields(logrus.Fields{
			"conn_key":      string(connKey),
			"peer_id":       peerID.String(),
			"overflow_size": streamChannel.GetOverflowQueueSize(),
		}).Debug("Added packet to overflow queue")
		atomic.AddInt64(&m.metrics.PacketsProcessed, 1)
		return nil
	}
}

// SendPacketBatch sends a batch of packets through the appropriate streams
// This is more efficient than sending packets individually when there are many packets
func (m *StreamManager) SendPacketBatch(
	packets []*types.QueuedPacket,
	connKeys []types.ConnectionKey,
	peerIDs []peer.ID,
	contexts []context.Context,
) error {
	if len(packets) == 0 {
		return nil
	}

	// Validate input lengths
	if len(packets) != len(connKeys) || len(packets) != len(peerIDs) || len(packets) != len(contexts) {
		return fmt.Errorf("mismatched input lengths: packets=%d, connKeys=%d, peerIDs=%d, contexts=%d",
			len(packets), len(connKeys), len(peerIDs), len(contexts))
	}

	// Create packet operations
	ops := make([]*packetOperation, len(packets))
	for i := range packets {
		ops[i] = &packetOperation{
			ctx:     contexts[i],
			connKey: connKeys[i],
			peerID:  peerIDs[i],
			packet:  packets[i],
		}
	}

	// Create a batch operation
	batch := packetBatchOp{
		packets:   ops,
		resultCh:  make(chan error, 1),
		timestamp: time.Now().UnixNano(),
	}

	// Select a batch channel based on a simple hash of the first connection key
	// This helps ensure packets for the same connection tend to go to the same worker
	var hash uint32
	for i := 0; i < len(connKeys[0]); i++ {
		hash = hash*31 + uint32(connKeys[0][i])
	}
	workerIdx := int(hash % uint32(atomic.LoadInt32(&m.batchWorkers)))

	// Send the batch to the selected worker
	select {
	case m.packetBatchChans[workerIdx] <- batch:
		// Wait for the result with a timeout
		select {
		case err := <-batch.resultCh:
			return err
		case <-time.After(100 * time.Millisecond):
			// Timeout waiting for result, but the batch is still being processed
			return nil
		}
	case <-time.After(50 * time.Millisecond):
		// Batch channel is full or slow, fall back to processing the batch directly
		m.processBatch(batch, -1) // -1 indicates direct processing
		return nil
	}
}

// GetMetrics returns the manager's metrics
func (m *StreamManager) GetMetrics() map[string]int64 {
	return map[string]int64{
		"streams_created":   atomic.LoadInt64(&m.metrics.StreamsCreated),
		"streams_released":  atomic.LoadInt64(&m.metrics.StreamsReleased),
		"packets_processed": atomic.LoadInt64(&m.metrics.PacketsProcessed),
		"batches_processed": atomic.LoadInt64(&m.metrics.BatchesProcessed),
		"errors":            atomic.LoadInt64(&m.metrics.Errors),
		"batch_workers":     int64(atomic.LoadInt32(&m.batchWorkers)),
	}
}

// GetConnectionCount returns the number of active connections
func (m *StreamManager) GetConnectionCount() int {
	count := 0
	for _, shard := range m.connectionShards {
		shard.RLock()
		count += len(shard.connections)
		shard.RUnlock()
	}
	return count
}

// getConnectionShardIndex returns the shard index for a connection key
func (m *StreamManager) getConnectionShardIndex(connKey types.ConnectionKey) int {
	// Simple hash function to distribute connections across shards
	var hash uint32
	for i := 0; i < len(connKey); i++ {
		hash = hash*31 + uint32(connKey[i])
	}
	// Use bit masking for faster modulo operation (works because shardCount is a power of 2)
	return int(hash & m.shardMask)
}

// Close implements io.Closer
func (m *StreamManager) Close() error {
	m.Stop()
	return nil
}

// connectionStatsLogger periodically logs statistics about connection-to-stream mappings
func (m *StreamManager) connectionStatsLogger() {
	// Create a ticker for periodic logging (every 30 seconds)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			// Log connection stats
			m.logConnectionStats()
		}
	}
}

// transferPendingPackets transfers any pending packets from the old stream to the new one
func (m *StreamManager) transferPendingPackets(oldStream, newStream *StreamChannel) {
	// Create a temporary buffer to hold packets
	var packets []*types.QueuedPacket

	// Get any packets from the overflow queue
	overflowPackets := oldStream.DrainOverflowQueue()
	if len(overflowPackets) > 0 {
		managerLog.WithField("packet_count", len(overflowPackets)).Debug("Transferred overflow packets to new stream")
		packets = append(packets, overflowPackets...)
	}

	// Try to drain the old stream's packet channel without blocking
	drainCount := 0
	for {
		select {
		case packet, ok := <-oldStream.PacketChan:
			if !ok {
				// Channel is closed
				break
			}
			packets = append(packets, packet)
			drainCount++
		default:
			// No more packets available without blocking
			goto doneDraining
		}
	}
doneDraining:

	// Log how many packets were transferred
	if drainCount > 0 {
		managerLog.WithField("packet_count", drainCount).Debug("Transferred pending packets from channel to new stream")
	}

	// Check if there's a failed packet that needs to be retried
	if oldStream.lastFailedPacket != nil {
		// Add the failed packet to the beginning of our packet list to prioritize it
		managerLog.Debug("Found failed packet to retry, adding to front of queue")
		packets = append([]*types.QueuedPacket{oldStream.lastFailedPacket}, packets...)
	}

	// Send the packets to the new stream's channel
	// We do this in reverse order to maintain the original order as much as possible
	for i := len(packets) - 1; i >= 0; i-- {
		select {
		case newStream.PacketChan <- packets[i]:
			// Successfully transferred
		default:
			// New stream's channel is full, add to overflow queue
			newStream.AddToOverflowQueue(packets[i])
			managerLog.WithField("dest_ip", packets[i].DestIP).Debug("Added packet to new stream's overflow queue")
		}
	}
}

// logConnectionStats logs detailed statistics about connection-to-stream mappings
func (m *StreamManager) logConnectionStats() {
	// Log overall stats
	connCount := m.GetConnectionCount()
	managerLog.WithFields(logrus.Fields{
		"connection_count":  connCount,
		"streams_created":   atomic.LoadInt64(&m.metrics.StreamsCreated),
		"streams_released":  atomic.LoadInt64(&m.metrics.StreamsReleased),
		"packets_processed": atomic.LoadInt64(&m.metrics.PacketsProcessed),
		"errors":            atomic.LoadInt64(&m.metrics.Errors),
	}).Info("Stream manager statistics")

	// Log detailed connection-to-stream mappings
	for shardIdx, shard := range m.connectionShards {
		shard.RLock()

		// Skip empty shards
		if len(shard.connections) == 0 {
			shard.RUnlock()
			continue
		}

		// Create a list of connection details
		connDetails := make([]map[string]interface{}, 0, len(shard.connections))
		for connKey, conn := range shard.connections {
			// Create connection detail
			detail := map[string]interface{}{
				"conn_key": string(connKey),
				"peer_id":  conn.PeerID.String(),
			}

			// Add stream information if available
			if conn.StreamChannel != nil {
				detail["stream_id"] = fmt.Sprintf("%p", conn.StreamChannel)
				detail["stream_healthy"] = conn.StreamChannel.IsHealthy()
				detail["buffer_util"] = conn.StreamChannel.GetBufferUtilization()
				detail["overflow_size"] = conn.StreamChannel.GetOverflowQueueSize()
				detail["total_buffer_util"] = conn.StreamChannel.GetTotalBufferUtilization()
				detail["packet_count"] = atomic.LoadInt64(&conn.StreamChannel.Metrics.PacketCount)
				detail["error_count"] = atomic.LoadInt64(&conn.StreamChannel.Metrics.ErrorCount)
			} else {
				detail["stream_id"] = "none"
			}

			connDetails = append(connDetails, detail)
		}
		shard.RUnlock()

		// Log shard connection stats
		managerLog.WithFields(logrus.Fields{
			"shard_index":      shardIdx,
			"connection_count": len(connDetails),
			"connections":      connDetails,
		}).Info("Connection-to-stream mappings")
	}
}
