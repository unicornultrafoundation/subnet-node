package ifce

import (
	"encoding/binary"
	"errors"
	"io"
	"net/netip"

	"github.com/google/gopacket/layers"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/firewall"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

var (
	ErrPacketTooShort          = errors.New("packet is too short")
	ErrUnknownIPVersion        = errors.New("packet is an unknown ip version")
	ErrIPv4InvalidHeaderLength = errors.New("invalid ipv4 header length")
	ErrIPv4PacketTooShort      = errors.New("ipv4 packet is too short")
	ErrIPv6PacketTooShort      = errors.New("ipv6 packet is too short")
	ErrIPv6CouldNotFindPayload = errors.New("could not find payload in ipv6 packet")
)

const (
	minFwPacketLen = 4
)

// newPacket validates and parses the interesting bits for the firewall out of the ip and sub protocol headers
func newPacket(data []byte, incoming bool, fp *firewall.Packet) error {
	if len(data) < 1 {
		return ErrPacketTooShort
	}

	version := int((data[0] >> 4) & 0x0f)
	switch version {
	case ipv4.Version:
		return parseV4(data, incoming, fp)
	case ipv6.Version:
		return parseV6(data, incoming, fp)
	}
	return ErrUnknownIPVersion
}

func parseV6(data []byte, incoming bool, fp *firewall.Packet) error {
	dataLen := len(data)
	if dataLen < ipv6.HeaderLen {
		return ErrIPv6PacketTooShort
	}

	if incoming {
		fp.RemoteAddr, _ = netip.AddrFromSlice(data[8:24])
		fp.LocalAddr, _ = netip.AddrFromSlice(data[24:40])
	} else {
		fp.LocalAddr, _ = netip.AddrFromSlice(data[8:24])
		fp.RemoteAddr, _ = netip.AddrFromSlice(data[24:40])
	}

	protoAt := 6             // NextHeader is at 6 bytes into the ipv6 header
	offset := ipv6.HeaderLen // Start at the end of the ipv6 header
	next := 0
	for {
		if dataLen < offset {
			break
		}

		proto := layers.IPProtocol(data[protoAt])
		//fmt.Println(proto, protoAt)
		switch proto {
		case layers.IPProtocolICMPv6, layers.IPProtocolESP, layers.IPProtocolNoNextHeader:
			fp.Protocol = uint8(proto)
			fp.RemotePort = 0
			fp.LocalPort = 0
			fp.Fragment = false
			return nil

		case layers.IPProtocolTCP, layers.IPProtocolUDP:
			if dataLen < offset+4 {
				return ErrIPv6PacketTooShort
			}

			fp.Protocol = uint8(proto)
			if incoming {
				fp.RemotePort = binary.BigEndian.Uint16(data[offset : offset+2])
				fp.LocalPort = binary.BigEndian.Uint16(data[offset+2 : offset+4])
			} else {
				fp.LocalPort = binary.BigEndian.Uint16(data[offset : offset+2])
				fp.RemotePort = binary.BigEndian.Uint16(data[offset+2 : offset+4])
			}

			fp.Fragment = false
			return nil

		case layers.IPProtocolIPv6Fragment:
			// Fragment header is 8 bytes, need at least offset+4 to read the offset field
			if dataLen < offset+8 {
				return ErrIPv6PacketTooShort
			}

			// Check if this is the first fragment
			fragmentOffset := binary.BigEndian.Uint16(data[offset+2:offset+4]) &^ uint16(0x7) // Remove the reserved and M flag bits
			if fragmentOffset != 0 {
				// Non-first fragment, use what we have now and stop processing
				fp.Protocol = data[offset]
				fp.Fragment = true
				fp.RemotePort = 0
				fp.LocalPort = 0
				return nil
			}

			// The next loop should be the transport layer since we are the first fragment
			next = 8 // Fragment headers are always 8 bytes

		case layers.IPProtocolAH:
			// Auth headers, used by IPSec, have a different meaning for header length
			if dataLen < offset+1 {
				break
			}

			next = int(data[offset+1]+2) << 2

		default:
			// Normal ipv6 header length processing
			if dataLen < offset+1 {
				break
			}

			next = int(data[offset+1]+1) << 3
		}

		if next <= 0 {
			// Safety check, each ipv6 header has to be at least 8 bytes
			next = 8
		}

		protoAt = offset
		offset = offset + next
	}

	return ErrIPv6CouldNotFindPayload
}

func parseV4(data []byte, incoming bool, fp *firewall.Packet) error {
	// Do we at least have an ipv4 header worth of data?
	if len(data) < ipv4.HeaderLen {
		return ErrIPv4PacketTooShort
	}

	// Adjust our start position based on the advertised ip header length
	ihl := int(data[0]&0x0f) << 2

	// Well-formed ip header length?
	if ihl < ipv4.HeaderLen {
		return ErrIPv4InvalidHeaderLength
	}

	// Check if this is the second or further fragment of a fragmented packet.
	flagsfrags := binary.BigEndian.Uint16(data[6:8])
	fp.Fragment = (flagsfrags & 0x1FFF) != 0

	// Firewall handles protocol checks
	fp.Protocol = data[9]

	// Accounting for a variable header length, do we have enough data for our src/dst tuples?
	minLen := ihl
	if !fp.Fragment && fp.Protocol != firewall.ProtoICMP {
		minLen += minFwPacketLen
	}
	if len(data) < minLen {
		return ErrIPv4InvalidHeaderLength
	}

	// Firewall packets are locally oriented
	if incoming {
		fp.RemoteAddr, _ = netip.AddrFromSlice(data[12:16])
		fp.LocalAddr, _ = netip.AddrFromSlice(data[16:20])
		if fp.Fragment || fp.Protocol == firewall.ProtoICMP {
			fp.RemotePort = 0
			fp.LocalPort = 0
		} else {
			fp.RemotePort = binary.BigEndian.Uint16(data[ihl : ihl+2])
			fp.LocalPort = binary.BigEndian.Uint16(data[ihl+2 : ihl+4])
		}
	} else {
		fp.LocalAddr, _ = netip.AddrFromSlice(data[12:16])
		fp.RemoteAddr, _ = netip.AddrFromSlice(data[16:20])
		if fp.Fragment || fp.Protocol == firewall.ProtoICMP {
			fp.RemotePort = 0
			fp.LocalPort = 0
		} else {
			fp.LocalPort = binary.BigEndian.Uint16(data[ihl : ihl+2])
			fp.RemotePort = binary.BigEndian.Uint16(data[ihl+2 : ihl+4])
		}
	}

	return nil
}

// readOutsidePackets processes incoming packets from peers and writes them to the TUN device
// using load balancing across available readers based on peer ID
func (f *Interface) readOutsidePackets(packet []byte, fwPacket *firewall.Packet, peerIdStr string) {
	// Parse the packet
	err := newPacket(packet, true, fwPacket)
	if err != nil {
		f.l.WithError(err).Error("Error while parsing inbound packet")
		return
	}

	//Reject all requests that are to ports outside the allowed range
	if fwPacket.RemotePort > 0 && fwPacket.RemotePort < 30000 {
		return
	}

	// Determine which reader to use based on peer ID hash for load balancing
	var readerIdx int
	if f.routines > 1 && peerIdStr != "" {
		// Simple hash function - sum the bytes of the peer ID string
		var hash uint64
		for i := 0; i < len(peerIdStr); i++ {
			hash += uint64(peerIdStr[i])
		}
		readerIdx = int(hash % uint64(f.routines))
	}

	// Write the packet to the appropriate reader
	var writer io.Writer
	if readerIdx < len(f.readers) && f.readers[readerIdx] != nil {
		writer = f.readers[readerIdx]
		f.l.WithFields(logrus.Fields{
			"peerId": peerIdStr,
			"queue":  readerIdx,
		}).Debug("Writing packet to specific TUN queue")
	} else {
		// Fallback to main interface if something went wrong with reader selection
		writer = f.inside
	}

	_, err = writer.Write(packet)
	if err != nil {
		f.l.WithError(err).Error("Failed to write packet to tun device")
		return
	}

	f.l.Debug("Packet written to tun device successfully")
}
