package ifce

import (
	"context"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/firewall"
)

func (f *Interface) getOrCreateStream(peerId peer.ID, i int) (network.Stream, error) {
	f.streamLock.Lock()
	defer f.streamLock.Unlock()

	streamId := fmt.Sprintf("%s/%d", peerId, i)

	stream, exists := f.streams[streamId]
	if exists {
		if _, err := stream.Write(nil); err != nil {
			f.l.WithError(err).WithField("stream", stream).Error("Stream is closed")
			exists = false
		} else {
			return stream, nil
		}
	}

	stream, err := f.p2phost.NewStream(context.Background(), peerId, "/vpn/1.0.0")
	if err != nil {
		return nil, err
	}

	f.streams[streamId] = stream
	return stream, nil
}

func (f *Interface) consumeInsidePacket(packet []byte, fwPacket *firewall.Packet, q int) {
	err := newPacket(packet, false, fwPacket)
	if err != nil {
		f.l.WithField("packet", packet).Errorf("Error while validating outbound packet: %s", err)
		return
	}

	remote := fwPacket.RemoteAddr.String()

	// Ignore non-10.x.x.x addresses
	if !strings.HasPrefix(remote, "10.") {
		return
	}

	peerIdStr, err := f.peerDiscovery.GetPeerID(context.Background(), remote)
	if err != nil {
		f.l.WithField("remoteAddr", remote).Error("Failed to get peer ID")
		return
	}

	peerId, err := peer.Decode(peerIdStr)
	if err != nil {
		f.l.WithField("remoteAddr", remote).Error("Failed to decode peer ID")
		return
	}

	// Prepend a 4-byte header indicating the packet size
	packetSize := len(packet)
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(packetSize))
	packetWithHeader := append(header, packet...)

	stream, err := f.getOrCreateStream(peerId, q)
	if err != nil {
		f.l.WithError(err).WithField("remoteAddr", remote).Error("Failed to get stream")
		return
	}

	// Write the packet with the header to the stream
	if _, err := stream.Write(packetWithHeader); err != nil {
		f.l.WithError(err).WithField("remoteAddr", remote).Error("Failed to write to stream")
		return
	}
}
