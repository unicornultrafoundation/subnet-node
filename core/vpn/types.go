package vpn

import "net"

// PacketInfo contains extracted source and destination IPs and ports
type PacketInfo struct {
	SrcIP    net.IP
	DstIP    net.IP
	SrcPort  *int // pointer to allow nil
	DstPort  *int
	Protocol uint8
}
