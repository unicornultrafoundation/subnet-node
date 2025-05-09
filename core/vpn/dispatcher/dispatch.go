package dispatcher

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/utils"
)

// DispatchPacket dispatches a packet to the appropriate stream
func (d *Dispatcher) DispatchPacket(ctx context.Context, packet []byte, remoteAddr string, queueID int) error {
	// Ignore non-10.x.x.x addresses
	if !strings.HasPrefix(remoteAddr, "10.") {
		return nil
	}

	// Get the peer ID for the destination IP
	peerIDStr, err := d.peerDiscovery.GetPeerID(ctx, remoteAddr)
	if err != nil {
		log.WithError(err).WithField("remote_ip", remoteAddr).Error("Failed to get peer ID")
		return fmt.Errorf("failed to get peer ID for %s: %w", remoteAddr, err)
	}

	// Decode the peer ID
	peerID, err := peer.Decode(peerIDStr)
	if err != nil {
		log.WithError(err).WithField("peer_id_str", peerIDStr).Error("Failed to decode peer ID")
		return fmt.Errorf("failed to decode peer ID %s: %w", peerIDStr, err)
	}

	// Create a stream ID that includes the queue ID
	streamID := fmt.Sprintf("%s/%d", peerID.String(), queueID)

	// Get or create a stream for the peer with the specific queue ID
	stream, err := d.getOrCreateStreamWithID(ctx, peerID, streamID)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"peer_id":   peerID.String(),
			"queue_id":  queueID,
			"stream_id": streamID,
		}).Error("Failed to get or create stream")
		return fmt.Errorf("failed to get or create stream for peer %s with queue %d: %w", peerID.String(), queueID, err)
	}

	// Prepend a 4-byte header indicating the packet size
	packetWithHeader := utils.PrependPacketHeader(packet)

	// Write the packet to the stream
	if _, err := stream.Write(packetWithHeader); err != nil {
		// Remove the stream from the map if there's an error
		d.streams.Delete(streamID)
		d.lastUsed.Delete(streamID)

		log.WithError(err).WithFields(logrus.Fields{
			"peer_id":     peerID.String(),
			"queue_id":    queueID,
			"stream_id":   streamID,
			"packet_size": len(packetWithHeader),
		}).Error("Failed to write packet to stream")

		return fmt.Errorf("failed to write packet to stream for peer %s with queue %d: %w", peerID.String(), queueID, err)
	}

	// Update the last used time for this stream
	d.lastUsed.Store(streamID, time.Now())

	return nil
}
