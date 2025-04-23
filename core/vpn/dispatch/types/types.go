package types

import (
	"context"
	"net"
)

// ConnectionKey represents a unique identifier for a connection
// in the format sourcePort:destinationIP:destinationPort
type ConnectionKey string

// PacketInfo contains extracted source and destination IPs and ports from a network packet.
// This structure is used for packet analysis and routing decisions.
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

// QueuedPacket represents a packet in the sending queue, ready to be processed by a worker.
// It contains all the necessary information for packet transmission and completion signaling.
type QueuedPacket struct {
	// Ctx is the context for the packet processing operation
	Ctx context.Context
	// DestIP is the destination IP address for the packet
	DestIP string
	// Data contains the raw packet bytes to be transmitted
	Data []byte
	// DoneCh is a channel to signal when packet processing is complete
	// The error (if any) is sent on this channel
	DoneCh chan error
}

// WorkerMetrics contains metrics for a worker
type WorkerMetrics struct {
	// PacketCount is the number of packets processed by the worker
	PacketCount int64
	// ErrorCount is the number of errors encountered by the worker
	ErrorCount int64
	// BytesSent is the number of bytes sent by the worker
	BytesSent int64
}

// StreamMetrics contains metrics for a stream
type StreamMetrics struct {
	// PacketCount is the number of packets sent through the stream
	PacketCount int64
	// ErrorCount is the number of errors encountered by the stream
	ErrorCount int64
	// BytesSent is the number of bytes sent through the stream
	BytesSent int64
}

// FormatConnectionKey formats a connection key from source port, destination IP, and destination port
func FormatConnectionKey(srcPort int, destIP string, destPort int) ConnectionKey {
	return ConnectionKey(net.JoinHostPort(net.JoinHostPort(destIP, string(destPort)), string(srcPort)))
}

// ExtractPacketInfo extracts source and destination IPs and ports from a packet
func ExtractPacketInfo(packet []byte) (*PacketInfo, error) {
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

	// Extract ports for TCP/UDP
	if protocol == 6 || protocol == 17 { // TCP or UDP
		if len(packet) < headerLen+4 {
			return result, nil // Return what we have without ports
		}

		// Extract source and destination ports
		srcPort := int(uint16(packet[headerLen])<<8 | uint16(packet[headerLen+1]))
		dstPort := int(uint16(packet[headerLen+2])<<8 | uint16(packet[headerLen+3]))

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

	// Extract ports for TCP/UDP
	if protocol == 6 || protocol == 17 { // TCP or UDP
		if len(packet) < 44 {
			return result, nil // Return what we have without ports
		}

		// Extract source and destination ports
		srcPort := int(uint16(packet[40])<<8 | uint16(packet[41]))
		dstPort := int(uint16(packet[42])<<8 | uint16(packet[43]))

		result.SrcPort = &srcPort
		result.DstPort = &dstPort
	}

	return result, nil
}
