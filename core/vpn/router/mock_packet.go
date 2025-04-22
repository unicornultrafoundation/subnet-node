package router

import (
	"net"

	"github.com/unicornultrafoundation/subnet-node/core/vpn/packet"
)

// MockExtractIPAndPorts is a mock implementation of packet.ExtractIPAndPorts for testing
func MockExtractIPAndPorts(data []byte) (*packet.PacketInfo, error) {
	// Return fixed values for testing
	srcPort := 12345
	dstPort := 80
	return &packet.PacketInfo{
		SrcIP:   net.ParseIP("10.0.0.1"),
		DstIP:   net.ParseIP("192.168.1.1"),
		SrcPort: &srcPort,
		DstPort: &dstPort,
	}, nil
}
