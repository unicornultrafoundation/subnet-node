package dispatch

import (
	"context"
	"sync/atomic"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
)

// BatchDispatcher is an implementation of the Dispatcher interface that supports batch processing
type BatchDispatcher struct {
	// Embed the standard dispatcher
	*Dispatcher

	// Stats
	stats struct {
		batchesProcessed int64
		packetsProcessed int64
		packetsDropped   int64
		peerIDErrors     int64
		streamErrors     int64
		channelFullCount int64
	}
}

// NewBatchDispatcher creates a new batch dispatcher
func NewBatchDispatcher(peerDiscovery api.PeerDiscoveryService, streamService api.StreamService, config *Config, resilienceService *resilience.ResilienceService) *BatchDispatcher {
	return &BatchDispatcher{
		Dispatcher: NewDispatcher(peerDiscovery, streamService, config, resilienceService),
	}
}

// DispatchPacketBatch dispatches a batch of packets
func (d *BatchDispatcher) DispatchPacketBatch(
	packets []*types.QueuedPacket,
	connKeys []types.ConnectionKey,
	destIPs []string,
	contexts []context.Context,
) error {
	if len(packets) == 0 {
		return nil
	}

	// Validate input lengths
	if len(packets) != len(connKeys) || len(packets) != len(destIPs) || len(packets) != len(contexts) {
		return types.ErrInvalidBatchInput
	}

	// Group packets by connection key to minimize lock contention
	packetsByConn := make(map[types.ConnectionKey][]*types.QueuedPacket)
	contextsByConn := make(map[types.ConnectionKey]context.Context)
	destIPsByConn := make(map[types.ConnectionKey]string)

	// Track which connections were successfully processed
	processedConnections := make(map[types.ConnectionKey]bool)

	// First pass: group packets by connection key
	for i, packet := range packets {
		connKey := connKeys[i]
		packetsByConn[connKey] = append(packetsByConn[connKey], packet)
		contextsByConn[connKey] = contexts[i]
		destIPsByConn[connKey] = destIPs[i]
	}

	// Second pass: process each connection's packets
	for connKey, connPackets := range packetsByConn {
		ctx := contextsByConn[connKey]
		destIP := destIPsByConn[connKey]

		// Get peer ID for the destination IP
		peerID, err := d.getPeerIDForDestIP(ctx, destIP)
		if err != nil {
			// Log the error and continue with other connections
			logrus.WithFields(logrus.Fields{
				"conn_key": string(connKey),
				"dest_ip":  destIP,
				"error":    err,
			}).Error("Failed to get peer ID for destination IP")

			// Update error metrics
			atomic.AddInt64(&d.stats.peerIDErrors, 1)
			atomic.AddInt64(&d.stats.packetsDropped, int64(len(connPackets)))

			// Don't mark this connection as processed
			continue
		}

		// Get or create a stream for this connection
		streamChannel, err := d.streamManager.GetOrCreateStreamForConnection(ctx, connKey, peerID)
		if err != nil {
			// Log the error and continue with other connections
			logrus.WithFields(logrus.Fields{
				"conn_key": string(connKey),
				"dest_ip":  destIP,
				"peer_id":  peerID.String(),
				"error":    err,
			}).Error("Failed to get or create stream for connection")

			// Update error metrics
			atomic.AddInt64(&d.stats.streamErrors, 1)
			atomic.AddInt64(&d.stats.packetsDropped, int64(len(connPackets)))

			// Don't mark this connection as processed
			continue
		}

		// Mark this connection as successfully processed
		processedConnections[connKey] = true

		// Send all packets for this connection
		for _, packet := range connPackets {
			select {
			case streamChannel.PacketChan <- packet:
				// Packet sent successfully
				atomic.AddInt64(&d.stats.packetsProcessed, 1)
			default:
				// Channel full, add to overflow queue
				streamChannel.AddToOverflowQueue(packet)
				atomic.AddInt64(&d.stats.packetsProcessed, 1)
				atomic.AddInt64(&d.stats.channelFullCount, 1)

				// Log if overflow queue is getting large
				overflowSize := streamChannel.GetOverflowQueueSize()
				if overflowSize > 100 {
					logrus.WithFields(logrus.Fields{
						"conn_key":      string(connKey),
						"dest_ip":       destIP,
						"peer_id":       peerID.String(),
						"overflow_size": overflowSize,
					}).Warn("Stream channel overflow queue is large")
				}
			}
		}
	}

	// Update batch stats
	atomic.AddInt64(&d.stats.batchesProcessed, 1)

	// Track which packets were successfully processed
	successfulPackets := make(map[*types.QueuedPacket]bool)

	// Mark all packets as successful initially
	for _, packet := range packets {
		successfulPackets[packet] = true
	}

	// Mark packets as failed if they were in a failed connection
	for connKey, connPackets := range packetsByConn {
		// If we failed to get a peer ID or stream for this connection,
		// mark all its packets as failed
		if _, ok := processedConnections[connKey]; !ok {
			for _, packet := range connPackets {
				successfulPackets[packet] = false
			}
		}
	}

	// Return packets to the pool based on their processing status
	for _, packet := range packets {
		if successfulPackets[packet] {
			// Successfully processed, return to pool
			types.GlobalPacketPool.Put(packet)
		} else {
			// Failed to process, log and return to pool
			logrus.WithFields(logrus.Fields{
				"dest_ip": packet.DestIP,
			}).Debug("Failed to process packet in batch, returning to pool")
			types.GlobalPacketPool.Put(packet)
		}
	}

	return nil
}

// GetMetrics returns the dispatcher's metrics
func (d *BatchDispatcher) GetMetrics() map[string]int64 {
	// Get base metrics
	metrics := d.Dispatcher.GetMetrics()

	// Add batch metrics
	metrics["batches_processed"] = atomic.LoadInt64(&d.stats.batchesProcessed)
	metrics["batch_packets_processed"] = atomic.LoadInt64(&d.stats.packetsProcessed)
	metrics["batch_packets_dropped"] = atomic.LoadInt64(&d.stats.packetsDropped)
	metrics["batch_peer_id_errors"] = atomic.LoadInt64(&d.stats.peerIDErrors)
	metrics["batch_stream_errors"] = atomic.LoadInt64(&d.stats.streamErrors)
	metrics["batch_channel_full_count"] = atomic.LoadInt64(&d.stats.channelFullCount)

	return metrics
}
