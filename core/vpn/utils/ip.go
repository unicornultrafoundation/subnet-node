package utils

import (
	"net"
)

type IPType string

const (
	STATIC_IP_TYPE      IPType = "static"
	DYNAMIC_IP_TYPE     IPType = "dynamic"
	UNSPECIFIED_IP_TYPE IPType = "unspecified"
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

	// Check if it's in 100.66.0.0/15 or 100.68.0.0/14 range (100.66.x.x through 100.71.x.x)
	if ipv4[0] != 100 {
		return 0
	}
	// Valid ranges: 100.66-100.71
	if ipv4[1] < 66 || ipv4[1] > 71 {
		return 0
	}

	// Skip unusable host addresses x.x.x.0 and x.x.x.255
	if ipv4[3] == 0 || ipv4[3] == 255 {
		return 0
	}

	// Compact mapping: token = 167772160 + 254*blockIndex + host
	// blockIndex = 0-255 for 100.66.x.x, 256-511 for 100.67.x.x, 512-767 for 100.68.x.x, etc.
	// This maps 100.66.0.1 to token ID 167772161
	base := uint32(10) << 24                               // 167772160 (token base, not IP base)
	blockIndex := uint32(ipv4[1]-66)*256 + uint32(ipv4[2]) // 0..1535 (6 subnets * 256)
	host := uint32(ipv4[3])                                // 1..254 (4th octet)
	return base + blockIndex*254 + host
}

// ConvertNumberToVirtualIP converts a token ID to a virtual IP
func ConvertNumberToVirtualIP(tokenID uint32) string {
	if tokenID == 0 {
		return ""
	}

	base := uint32(10) << 24 // token base (167772160)
	if tokenID < base+1 {
		return ""
	}

	delta := tokenID - base
	// Inverse of token = base + 254*blockIndex + host (host in 1..254)
	blockIndex := (delta - 1) / 254
	host := byte(((delta - 1) % 254) + 1) // 1..254

	if blockIndex >= 1536 {
		return ""
	}

	// Map blockIndex to IP:
	// 0-255 -> 100.66.x.x, 256-511 -> 100.67.x.x, 512-767 -> 100.68.x.x
	// 768-1023 -> 100.69.x.x, 1024-1279 -> 100.70.x.x, 1280-1535 -> 100.71.x.x
	secondOctet := byte(66 + blockIndex/256)
	thirdOctet := byte(blockIndex % 256)
	ip := net.IPv4(100, secondOctet, thirdOctet, host)
	return ip.String()
}

// GetIPType detects the type of an IP address
// Returns STATIC_IP_TYPE for 100.66.0.0/15 (100.66.x.x, 100.67.x.x)
// Returns DYNAMIC_IP_TYPE for 100.68.0.0/14 (100.68.x.x through 100.71.x.x)
// Returns UNSPECIFIED_IP_TYPE for invalid or unsupported IPs
func GetIPType(virtualIP string) IPType {
	// Parse the IP address
	ip := net.ParseIP(virtualIP)
	if ip == nil {
		return UNSPECIFIED_IP_TYPE
	}

	// Convert to IPv4 format
	ipv4 := ip.To4()
	if ipv4 == nil {
		return UNSPECIFIED_IP_TYPE
	}

	// Check if it's in our supported range
	if ipv4[0] != 100 {
		return UNSPECIFIED_IP_TYPE
	}

	// Determine type based on second octet
	secondOctet := ipv4[1]

	// Static IP: 100.66.x.x or 100.67.x.x
	if secondOctet == 66 || secondOctet == 67 {
		return STATIC_IP_TYPE
	}

	// Dynamic IP: 100.68.x.x through 100.71.x.x
	if secondOctet >= 68 && secondOctet <= 71 {
		return DYNAMIC_IP_TYPE
	}

	return UNSPECIFIED_IP_TYPE
}

// GetTokenIPType detects the type of an IP from a token ID
// Returns STATIC_IP_TYPE for token IDs corresponding to 100.66.0.0/15
// Returns DYNAMIC_IP_TYPE for token IDs corresponding to 100.68.0.0/14
// Returns UNSPECIFIED_IP_TYPE for invalid token IDs
func GetTokenIPType(tokenID uint32) IPType {
	if tokenID == 0 {
		return UNSPECIFIED_IP_TYPE
	}

	base := uint32(10) << 24 // token base (167772160)
	if tokenID < base+1 {
		return UNSPECIFIED_IP_TYPE
	}

	delta := tokenID - base
	// Inverse of token = base + 254*blockIndex + host (host in 1..254)
	blockIndex := (delta - 1) / 254

	// Check if blockIndex is valid
	if blockIndex >= 1536 {
		return UNSPECIFIED_IP_TYPE
	}

	// Static IP: blockIndex 0-511 (100.66.x.x and 100.67.x.x)
	if blockIndex < 512 {
		return STATIC_IP_TYPE
	}

	// Dynamic IP: blockIndex 512-1535 (100.68.x.x through 100.71.x.x)
	return DYNAMIC_IP_TYPE
}
