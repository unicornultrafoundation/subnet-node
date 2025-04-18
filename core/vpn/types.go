package vpn

import (
	"context"
	"net"
)

// PacketInfo contains extracted source and destination IPs and ports
type PacketInfo struct {
	SrcIP    net.IP
	DstIP    net.IP
	SrcPort  *int // pointer to allow nil
	DstPort  *int
	Protocol uint8
}

// QueuedPacket represents a packet in the sending queue
type QueuedPacket struct {
	Ctx    context.Context
	DestIP string
	Data   []byte
	DoneCh chan error // Channel to signal when packet processing is complete
}
