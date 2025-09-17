package discovery

import (
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

	// Skip unusable host addresses x.x.x.0 and x.x.x.255
	if ipv4[3] == 0 || ipv4[3] == 255 {
		return 0
	}

	// Compact mapping: token = base + 254*blockIndex + host
	// base = 10.0.0.0 = 10<<24 = 167772160
	base := uint32(10) << 24
	blockIndex := uint32(ipv4[1])<<8 | uint32(ipv4[2]) // 0..65535
	host := uint32(ipv4[3])                            // 1..254
	return base + blockIndex*254 + host
}

// ConvertNumberToVirtualIP converts a token ID to a virtual IP
func ConvertNumberToVirtualIP(tokenID uint32) string {
	if tokenID == 0 {
		return ""
	}

	base := uint32(10) << 24 // 10.0.0.0
	if tokenID < base+1 {
		return ""
	}

	delta := tokenID - base
	// Inverse of token = base + 254*blockIndex + host (host in 1..254)
	blockIndex := (delta - 1) / 254
	host := byte(((delta - 1) % 254) + 1) // 1..254

	if blockIndex >= 65536 {
		return ""
	}

	b1 := byte((blockIndex >> 8) & 0xFF)
	b2 := byte(blockIndex & 0xFF)
	ip := net.IPv4(10, b1, b2, host)
	return ip.String()
}
