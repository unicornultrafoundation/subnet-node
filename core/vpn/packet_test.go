package vpn

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestExtractIPAndPortsIPv4TCP tests extracting IP and port information from an IPv4 TCP packet
func TestExtractIPAndPortsIPv4TCP(t *testing.T) {
	// Create a mock IPv4 TCP packet
	packet := []byte{
		// IPv4 header (20 bytes)
		0x45, 0x00, 0x00, 0x3c, // Version, IHL, TOS, Total Length
		0x00, 0x00, 0x40, 0x00, // Identification, Flags, Fragment Offset
		0x40, 0x06, 0x00, 0x00, // TTL, Protocol (6 = TCP), Header Checksum
		0x0a, 0x00, 0x00, 0x01, // Source IP: 10.0.0.1
		0xc0, 0xa8, 0x01, 0x01, // Destination IP: 192.168.1.1

		// TCP header (20 bytes)
		0x30, 0x39, 0x00, 0x50, // Source Port: 12345, Destination Port: 80
		0x00, 0x00, 0x00, 0x00, // Sequence Number
		0x00, 0x00, 0x00, 0x00, // Acknowledgment Number
		0x50, 0x02, 0x20, 0x00, // Data Offset, Flags, Window Size
		0x00, 0x00, 0x00, 0x00, // Checksum, Urgent Pointer

		// Payload (4 bytes)
		0x01, 0x02, 0x03, 0x04,
	}

	// Extract IP and port information
	packetInfo, err := ExtractIPAndPorts(packet)
	assert.NoError(t, err)

	// Verify the extracted information
	assert.Equal(t, "10.0.0.1", packetInfo.SrcIP.String())
	assert.Equal(t, "192.168.1.1", packetInfo.DstIP.String())
	assert.NotNil(t, packetInfo.SrcPort)
	assert.NotNil(t, packetInfo.DstPort)
	assert.Equal(t, 12345, *packetInfo.SrcPort)
	assert.Equal(t, 80, *packetInfo.DstPort)
	assert.Equal(t, uint8(6), packetInfo.Protocol) // TCP
}

// TestExtractIPAndPortsIPv4UDP tests extracting IP and port information from an IPv4 UDP packet
func TestExtractIPAndPortsIPv4UDP(t *testing.T) {
	// Create a mock IPv4 UDP packet
	packet := []byte{
		// IPv4 header (20 bytes)
		0x45, 0x00, 0x00, 0x1c, // Version, IHL, TOS, Total Length
		0x00, 0x00, 0x40, 0x00, // Identification, Flags, Fragment Offset
		0x40, 0x11, 0x00, 0x00, // TTL, Protocol (17 = UDP), Header Checksum
		0x0a, 0x00, 0x00, 0x01, // Source IP: 10.0.0.1
		0xc0, 0xa8, 0x01, 0x01, // Destination IP: 192.168.1.1

		// UDP header (8 bytes)
		0x30, 0x39, 0x00, 0x50, // Source Port: 12345, Destination Port: 80
		0x00, 0x08, 0x00, 0x00, // Length, Checksum
	}

	// Extract IP and port information
	packetInfo, err := ExtractIPAndPorts(packet)
	assert.NoError(t, err)

	// Verify the extracted information
	assert.Equal(t, "10.0.0.1", packetInfo.SrcIP.String())
	assert.Equal(t, "192.168.1.1", packetInfo.DstIP.String())
	assert.NotNil(t, packetInfo.SrcPort)
	assert.NotNil(t, packetInfo.DstPort)
	assert.Equal(t, 12345, *packetInfo.SrcPort)
	assert.Equal(t, 80, *packetInfo.DstPort)
	assert.Equal(t, uint8(17), packetInfo.Protocol) // UDP
}

// TestExtractIPAndPortsIPv4ICMP tests extracting IP information from an IPv4 ICMP packet
func TestExtractIPAndPortsIPv4ICMP(t *testing.T) {
	// Create a mock IPv4 ICMP packet
	packet := []byte{
		// IPv4 header (20 bytes)
		0x45, 0x00, 0x00, 0x1c, // Version, IHL, TOS, Total Length
		0x00, 0x00, 0x40, 0x00, // Identification, Flags, Fragment Offset
		0x40, 0x01, 0x00, 0x00, // TTL, Protocol (1 = ICMP), Header Checksum
		0x0a, 0x00, 0x00, 0x01, // Source IP: 10.0.0.1
		0xc0, 0xa8, 0x01, 0x01, // Destination IP: 192.168.1.1

		// ICMP header (8 bytes)
		0x08, 0x00, 0x00, 0x00, // Type, Code, Checksum
		0x00, 0x00, 0x00, 0x00, // Rest of Header
	}

	// Extract IP and port information
	packetInfo, err := ExtractIPAndPorts(packet)
	assert.NoError(t, err)

	// Verify the extracted information
	assert.Equal(t, "10.0.0.1", packetInfo.SrcIP.String())
	assert.Equal(t, "192.168.1.1", packetInfo.DstIP.String())
	assert.Nil(t, packetInfo.SrcPort)              // ICMP doesn't have ports
	assert.Nil(t, packetInfo.DstPort)              // ICMP doesn't have ports
	assert.Equal(t, uint8(1), packetInfo.Protocol) // ICMP
}

// TestExtractIPAndPortsTooShortPacket tests extracting information from an invalid packet
func TestExtractIPAndPortsTooShortPacket(t *testing.T) {
	// Create an invalid packet (too short)
	packet := []byte{0x45, 0x00, 0x00, 0x3c}

	// Extract IP and port information
	_, err := ExtractIPAndPorts(packet)
	assert.Error(t, err)
}

// TestExtractIPAndPortsNonIPv4Packet tests extracting information from a non-IPv4 packet
func TestExtractIPAndPortsNonIPv4Packet(t *testing.T) {
	// Create a non-IPv4 packet (IPv6)
	packet := []byte{
		0x60, 0x00, 0x00, 0x00, // Version (6), Traffic Class, Flow Label
		0x00, 0x08, 0x11, 0x40, // Payload Length, Next Header, Hop Limit
		// Source IP (16 bytes)
		0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
		// Destination IP (16 bytes)
		0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02,
	}

	// Extract IP and port information
	_, err := ExtractIPAndPorts(packet)
	assert.Error(t, err)
}

// TestCreateSyncKey tests creating a synchronization key from IP and port
func TestCreateSyncKey(t *testing.T) {
	// Test with IP and port
	ip := net.ParseIP("192.168.1.1")
	port := 80
	syncKey := createSyncKey(ip, &port)
	assert.Equal(t, "192.168.1.1:80", syncKey)

	// Test with IP only
	syncKey = createSyncKey(ip, nil)
	assert.Equal(t, "192.168.1.1", syncKey)
}

// Helper function to create a synchronization key from IP and port
func createSyncKey(ip net.IP, port *int) string {
	if port != nil {
		return net.JoinHostPort(ip.String(), fmt.Sprintf("%d", *port))
	}
	return ip.String()
}
