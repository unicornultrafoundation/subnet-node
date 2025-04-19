package packet

import (
	"encoding/binary"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractIPAndPorts(t *testing.T) {
	// Create a test IPv4 packet
	ipv4Packet := createTestIPv4Packet("192.168.1.1", "10.0.0.1", 12345, 80, 6, []byte("test"))

	// Test extracting information from the IPv4 packet
	info, err := ExtractIPAndPorts(ipv4Packet)
	assert.NoError(t, err)
	assert.Equal(t, net.ParseIP("192.168.1.1").To4(), info.SrcIP)
	assert.Equal(t, net.ParseIP("10.0.0.1").To4(), info.DstIP)
	assert.Equal(t, 12345, *info.SrcPort)
	assert.Equal(t, 80, *info.DstPort)
	assert.Equal(t, uint8(6), info.Protocol) // TCP

	// Create a test IPv6 packet
	ipv6Packet := createTestIPv6Packet("2001:db8::1", "2001:db8::2", 12345, 80, 6, []byte("test"))

	// Test extracting information from the IPv6 packet
	info, err = ExtractIPAndPorts(ipv6Packet)
	assert.NoError(t, err)
	assert.Equal(t, net.ParseIP("2001:db8::1"), info.SrcIP)
	assert.Equal(t, net.ParseIP("2001:db8::2"), info.DstIP)
	assert.Equal(t, 12345, *info.SrcPort)
	assert.Equal(t, 80, *info.DstPort)
	assert.Equal(t, uint8(6), info.Protocol) // TCP

	// Test with an empty packet
	_, err = ExtractIPAndPorts([]byte{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty packet")

	// Test with an invalid IP version
	invalidPacket := []byte{0x00, 0x00, 0x00, 0x00} // First 4 bits are version, set to 0
	_, err = ExtractIPAndPorts(invalidPacket)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported IP version")

	// Test with a packet that's too short for IPv4
	shortPacket := []byte{0x45} // IPv4, but too short
	_, err = ExtractIPAndPorts(shortPacket)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "packet too short")

	// Test with a packet that's too short for IPv6
	shortIPv6Packet := []byte{0x60} // IPv6, but too short
	_, err = ExtractIPAndPorts(shortIPv6Packet)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "packet too short")
}

// Helper function to create a test IPv4 packet
func createTestIPv4Packet(srcIP, dstIP string, srcPort, dstPort int, protocol uint8, payload []byte) []byte {
	// IPv4 header is at least 20 bytes
	headerSize := 20
	// TCP/UDP header is at least 8 bytes
	transportHeaderSize := 8
	// Total packet size
	packetSize := headerSize + transportHeaderSize + len(payload)
	
	packet := make([]byte, packetSize)
	
	// Set IP version (4) and header length (5 32-bit words = 20 bytes)
	packet[0] = 0x45
	
	// Set total length
	binary.BigEndian.PutUint16(packet[2:4], uint16(packetSize))
	
	// Set protocol (TCP=6, UDP=17)
	packet[9] = protocol
	
	// Set source IP
	copy(packet[12:16], net.ParseIP(srcIP).To4())
	
	// Set destination IP
	copy(packet[16:20], net.ParseIP(dstIP).To4())
	
	// Set source port
	binary.BigEndian.PutUint16(packet[headerSize:headerSize+2], uint16(srcPort))
	
	// Set destination port
	binary.BigEndian.PutUint16(packet[headerSize+2:headerSize+4], uint16(dstPort))
	
	// Add payload
	copy(packet[headerSize+transportHeaderSize:], payload)
	
	return packet
}

// Helper function to create a test IPv6 packet
func createTestIPv6Packet(srcIP, dstIP string, srcPort, dstPort int, protocol uint8, payload []byte) []byte {
	// IPv6 header is 40 bytes
	headerSize := 40
	// TCP/UDP header is at least 8 bytes
	transportHeaderSize := 8
	// Total packet size
	packetSize := headerSize + transportHeaderSize + len(payload)
	
	packet := make([]byte, packetSize)
	
	// Set IP version (6) and traffic class/flow label
	packet[0] = 0x60
	
	// Set payload length (excluding IPv6 header)
	binary.BigEndian.PutUint16(packet[4:6], uint16(transportHeaderSize+len(payload)))
	
	// Set next header (protocol)
	packet[6] = protocol
	
	// Set source IP
	copy(packet[8:24], net.ParseIP(srcIP).To16())
	
	// Set destination IP
	copy(packet[24:40], net.ParseIP(dstIP).To16())
	
	// Set source port
	binary.BigEndian.PutUint16(packet[headerSize:headerSize+2], uint16(srcPort))
	
	// Set destination port
	binary.BigEndian.PutUint16(packet[headerSize+2:headerSize+4], uint16(dstPort))
	
	// Add payload
	copy(packet[headerSize+transportHeaderSize:], payload)
	
	return packet
}
