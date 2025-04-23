package packet

import (
	"net"
)

// PacketInfo contains extracted source and destination IPs and ports from a network packet
type PacketInfo struct {
	// SrcIP is the source IP address of the packet
	SrcIP net.IP
	// DstIP is the destination IP address of the packet
	DstIP net.IP
	// SrcPort is the source port of the packet (pointer to allow nil for non-TCP/UDP protocols)
	SrcPort *int
	// DstPort is the destination port of the packet (pointer to allow nil for non-TCP/UDP protocols)
	DstPort *int
	// Protocol is the IP protocol number (e.g., 6 for TCP, 17 for UDP)
	Protocol uint8
}

// ExtractIPAndPorts extracts source and destination IPs and ports from a packet
func ExtractIPAndPorts(packet []byte) (*PacketInfo, error) {
	if len(packet) < 1 {
		return nil, ErrEmptyPacket
	}

	version := packet[0] >> 4
	switch version {
	case 4:
		return extractIPv4Info(packet)
	case 6:
		return extractIPv6Info(packet)
	default:
		return nil, ErrUnsupportedIPVersion
	}
}

// extractIPv4Info extracts IP and port information from an IPv4 packet
func extractIPv4Info(packet []byte) (*PacketInfo, error) {
	if len(packet) < 20 {
		return nil, ErrInvalidPacketLength
	}

	// Extract header length
	headerLen := int(packet[0]&0x0F) * 4
	if len(packet) < headerLen {
		return nil, ErrInvalidPacketLength
	}

	// Extract protocol
	protocol := packet[9]

	// Extract source and destination IPs
	srcIP := net.IP(packet[12:16])
	dstIP := net.IP(packet[16:20])

	// Initialize result
	result := &PacketInfo{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		Protocol: protocol,
	}

	// Extract ports for TCP and UDP
	if (protocol == 6 || protocol == 17) && len(packet) >= headerLen+4 {
		// Extract source and destination ports
		srcPort := int(packet[headerLen])<<8 | int(packet[headerLen+1])
		dstPort := int(packet[headerLen+2])<<8 | int(packet[headerLen+3])
		result.SrcPort = &srcPort
		result.DstPort = &dstPort
	}

	return result, nil
}

// extractIPv6Info extracts IP and port information from an IPv6 packet
func extractIPv6Info(packet []byte) (*PacketInfo, error) {
	if len(packet) < 40 {
		return nil, ErrInvalidPacketLength
	}

	// Extract next header (protocol)
	protocol := packet[6]

	// Extract source and destination IPs
	srcIP := net.IP(packet[8:24])
	dstIP := net.IP(packet[24:40])

	// Initialize result
	result := &PacketInfo{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		Protocol: protocol,
	}

	// Extract ports for TCP and UDP
	if (protocol == 6 || protocol == 17) && len(packet) >= 44 {
		// Extract source and destination ports
		srcPort := int(packet[40])<<8 | int(packet[41])
		dstPort := int(packet[42])<<8 | int(packet[43])
		result.SrcPort = &srcPort
		result.DstPort = &dstPort
	}

	return result, nil
}
