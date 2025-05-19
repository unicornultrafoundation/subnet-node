package discovery

import (
	"encoding/binary"
	"net"
)

// ConvertVirtualIPToNumber converts a virtual IP to a token ID
func ConvertVirtualIPToNumber(virtualIP string) uint32 {
	// Parse the IP address
	ip := net.ParseIP(virtualIP)
	if ip == nil {
		// Invalid IP address format
		return 0
	}

	// Convert to IPv4 format
	ipv4 := ip.To4()
	if ipv4 == nil {
		// Not an IPv4 address
		return 0
	}

	// Check if it's in the 10.0.0.0/8 range
	if ipv4[0] != 10 {
		return 0
	}

	// Convert to uint32
	return binary.BigEndian.Uint32(ipv4)
}
