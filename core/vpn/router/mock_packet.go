package router

import (
	"net"

	"github.com/unicornultrafoundation/subnet-node/core/vpn/packet"
)

// MockMode controls the behavior of MockExtractIPAndPorts
var MockMode string = "WITH_PORTS"

// MockExtractIPAndPorts is a mock implementation of packet.ExtractIPAndPorts for testing
func MockExtractIPAndPorts(data []byte) (*packet.PacketInfo, error) {
	// Return fixed values for testing
	if MockMode == "WITHOUT_PORTS" {
		// Return ICMP packet info (no ports)
		return &packet.PacketInfo{
			SrcIP:    net.ParseIP("10.0.0.1"),
			DstIP:    net.ParseIP("192.168.1.1"),
			SrcPort:  nil,
			DstPort:  nil,
			Protocol: 1, // ICMP
		}, nil
	} else {
		// Return TCP packet info (with ports)
		srcPort := 12345
		dstPort := 80
		return &packet.PacketInfo{
			SrcIP:    net.ParseIP("10.0.0.1"),
			DstIP:    net.ParseIP("192.168.1.1"),
			SrcPort:  &srcPort,
			DstPort:  &dstPort,
			Protocol: 6, // TCP
		}, nil
	}
}
