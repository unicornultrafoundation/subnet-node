package dispatcher

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/utils"
	"github.com/unicornultrafoundation/subnet-node/firewall"
)

// DispatchPacket dispatches a packet to the appropriate stream
func (d *Dispatcher) DispatchPacket(ctx context.Context, packet []byte, queueID int) error {
	// Parse the packet to get the destination IP
	fwPacket := &firewall.Packet{}
	if err := utils.ParsePacket(packet, false, fwPacket); err != nil {
		// Track packet parse errors using atomic
		d.packetParseErrors.Add(1)

		log.WithError(err).WithField("packet_size", len(packet)).Error("Failed to parse packet")
		return fmt.Errorf("failed to parse packet: %w", err)
	}

	// Get the remote address
	remote := fwPacket.RemoteAddr.String()

	// Ignore non-10.x.x.x addresses
	if !strings.HasPrefix(remote, "10.") {
		return nil
	}

	// Get the peer ID for the destination IP
	peerIDStr, err := d.peerDiscovery.GetPeerID(ctx, remote)
	if err != nil {
		// Track peer discovery errors using atomic
		d.peerDiscoveryErrors.Add(1)

		log.WithError(err).WithField("remote_ip", remote).Error("Failed to get peer ID")
		return fmt.Errorf("failed to get peer ID for %s: %w", remote, err)
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

		// Track stream errors - atomic for counter, sync.Map for per-peer tracking
		d.streamErrors.Add(1)

		// Update stream errors by peer using sync.Map
		peerStr := peerID.String()
		val, _ := d.streamErrorsByPeer.LoadOrStore(peerStr, &atomic.Uint64{})
		counter := val.(*atomic.Uint64)
		counter.Add(1)

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

	// Track successful packet dispatch - atomic for counters
	d.packetsDispatched.Add(1)
	d.bytesDispatched.Add(uint64(len(packet)))

	// Update per-peer metrics using sync.Map
	peerStr := peerID.String()

	// Update packets count by peer
	packetVal, _ := d.packetsDispatchedByPeer.LoadOrStore(peerStr, &atomic.Uint64{})
	packetCounter := packetVal.(*atomic.Uint64)
	packetCounter.Add(1)

	// Update bytes count by peer
	byteVal, _ := d.bytesDispatchedByPeer.LoadOrStore(peerStr, &atomic.Uint64{})
	byteCounter := byteVal.(*atomic.Uint64)
	byteCounter.Add(uint64(len(packet)))

	return nil
}
