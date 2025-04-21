package packet

import (
	"context"
	"net"
)

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
	PacketCount int64
	ErrorCount  int64
}
