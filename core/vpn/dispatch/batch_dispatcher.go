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
			continue
		}

		// Get or create a stream for this connection
		streamChannel, err := d.streamManager.GetOrCreateStreamForConnection(ctx, connKey, peerID)
		if err != nil {
			// Log the error and continue with other connections
			logrus.WithFields(logrus.Fields{
				"conn_key": string(connKey),
				"dest_ip":  destIP,
				"error":    err,
			}).Error("Failed to get or create stream for connection")
			continue
		}

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
			}
		}
	}

	// Update batch stats
	atomic.AddInt64(&d.stats.batchesProcessed, 1)

	// Return packets to the pool
	for _, packet := range packets {
		types.GlobalPacketPool.Put(packet)
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

	return metrics
}
