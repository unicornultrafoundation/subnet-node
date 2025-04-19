package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

// ExtractIPAndPorts extracts source and destination IPs and ports from a packet
func ExtractIPAndPorts(packet []byte) (*PacketInfo, error) {
	if len(packet) < 1 {
		return nil, errors.New("empty packet")
	}

	version := packet[0] >> 4
	switch version {
	case 4:
		return extractIPv4Info(packet)
	case 6:
		return extractIPv6Info(packet)
	default:
		return nil, fmt.Errorf("unsupported IP version: %d", version)
	}
}

// extractIPv4Info extracts information from an IPv4 packet
func extractIPv4Info(packet []byte) (*PacketInfo, error) {
	if len(packet) < 20 {
		return nil, errors.New("packet too short for IPv4")
	}

	// Extract header length
	ihl := int(packet[0]&0x0F) * 4
	if len(packet) < ihl {
		return nil, errors.New("packet too short for IPv4 header")
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

	// Extract ports for TCP or UDP
	if protocol == 6 || protocol == 17 { // TCP or UDP
		if len(packet) < ihl+4 {
			return result, nil // Not enough data for ports, return what we have
		}

		// Extract source and destination ports
		srcPort := int(binary.BigEndian.Uint16(packet[ihl : ihl+2]))
		dstPort := int(binary.BigEndian.Uint16(packet[ihl+2 : ihl+4]))

		result.SrcPort = &srcPort
		result.DstPort = &dstPort
	}

	return result, nil
}

// extractIPv6Info extracts information from an IPv6 packet
func extractIPv6Info(packet []byte) (*PacketInfo, error) {
	if len(packet) < 40 {
		return nil, errors.New("packet too short for IPv6")
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

	// Extract ports for TCP or UDP
	if protocol == 6 || protocol == 17 { // TCP or UDP
		if len(packet) < 44 {
			return result, nil // Not enough data for ports, return what we have
		}

		// Extract source and destination ports
		srcPort := int(binary.BigEndian.Uint16(packet[40:42]))
		dstPort := int(binary.BigEndian.Uint16(packet[42:44]))

		result.SrcPort = &srcPort
		result.DstPort = &dstPort
	}

	return result, nil
}
