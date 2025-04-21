package packet

import (
	"encoding/binary"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper function to create a pointer to an int
func intPtr(i int) *int {
	return &i
}

func TestExtractIPAndPorts(t *testing.T) {
	// Define test cases using a table-driven approach
	tests := []struct {
		name           string
		packet         []byte
		expectedResult *PacketInfo
		expectedError  string
	}{
		{
			name:   "Valid IPv4 TCP packet",
			packet: createTestIPv4Packet("192.168.1.1", "10.0.0.1", 12345, 80, 6, []byte("test")),
			expectedResult: &PacketInfo{
				SrcIP:    net.ParseIP("192.168.1.1").To4(),
				DstIP:    net.ParseIP("10.0.0.1").To4(),
				SrcPort:  intPtr(12345),
				DstPort:  intPtr(80),
				Protocol: 6, // TCP
			},
			expectedError: "",
		},
		{
			name:   "Valid IPv4 UDP packet",
			packet: createTestIPv4Packet("192.168.1.1", "10.0.0.1", 53, 12345, 17, []byte("test")),
			expectedResult: &PacketInfo{
				SrcIP:    net.ParseIP("192.168.1.1").To4(),
				DstIP:    net.ParseIP("10.0.0.1").To4(),
				SrcPort:  intPtr(53),
				DstPort:  intPtr(12345),
				Protocol: 17, // UDP
			},
			expectedError: "",
		},
		{
			name:   "Valid IPv4 ICMP packet",
			packet: createTestIPv4Packet("192.168.1.1", "10.0.0.1", 0, 0, 1, []byte("test")),
			expectedResult: &PacketInfo{
				SrcIP:    net.ParseIP("192.168.1.1").To4(),
				DstIP:    net.ParseIP("10.0.0.1").To4(),
				SrcPort:  nil,
				DstPort:  nil,
				Protocol: 1, // ICMP
			},
			expectedError: "",
		},
		{
			name:   "Valid IPv6 TCP packet",
			packet: createTestIPv6Packet("2001:db8::1", "2001:db8::2", 12345, 80, 6, []byte("test")),
			expectedResult: &PacketInfo{
				SrcIP:    net.ParseIP("2001:db8::1"),
				DstIP:    net.ParseIP("2001:db8::2"),
				SrcPort:  intPtr(12345),
				DstPort:  intPtr(80),
				Protocol: 6, // TCP
			},
			expectedError: "",
		},
		{
			name:   "Valid IPv6 UDP packet",
			packet: createTestIPv6Packet("2001:db8::1", "2001:db8::2", 53, 12345, 17, []byte("test")),
			expectedResult: &PacketInfo{
				SrcIP:    net.ParseIP("2001:db8::1"),
				DstIP:    net.ParseIP("2001:db8::2"),
				SrcPort:  intPtr(53),
				DstPort:  intPtr(12345),
				Protocol: 17, // UDP
			},
			expectedError: "",
		},
		{
			name:   "Valid IPv6 ICMPv6 packet",
			packet: createTestIPv6Packet("2001:db8::1", "2001:db8::2", 0, 0, 58, []byte("test")),
			expectedResult: &PacketInfo{
				SrcIP:    net.ParseIP("2001:db8::1"),
				DstIP:    net.ParseIP("2001:db8::2"),
				SrcPort:  nil,
				DstPort:  nil,
				Protocol: 58, // ICMPv6
			},
			expectedError: "",
		},
		{
			name:           "Empty packet",
			packet:         []byte{},
			expectedResult: nil,
			expectedError:  "empty packet",
		},
		{
			name:           "Invalid IP version",
			packet:         []byte{0x00, 0x00, 0x00, 0x00}, // First 4 bits are version, set to 0
			expectedResult: nil,
			expectedError:  "unsupported IP version",
		},
		{
			name:           "Short IPv4 packet",
			packet:         []byte{0x45}, // IPv4, but too short
			expectedResult: nil,
			expectedError:  "packet too short",
		},
		{
			name:           "Short IPv6 packet",
			packet:         []byte{0x60}, // IPv6, but too short
			expectedResult: nil,
			expectedError:  "packet too short",
		},
		{
			name:           "IPv4 packet with invalid header length",
			packet:         []byte{0x4f, 0x00, 0x00, 0x14, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06, 0x00, 0x00, 0xc0, 0xa8, 0x01, 0x01, 0x0a, 0x00, 0x00, 0x01},
			expectedResult: nil,
			expectedError:  "packet too short for IPv4 header",
		},
		{
			name:   "IPv4 packet with insufficient data for ports",
			packet: createTestIPv4PacketWithoutPorts("192.168.1.1", "10.0.0.1", 6),
			expectedResult: &PacketInfo{
				SrcIP:    net.ParseIP("192.168.1.1").To4(),
				DstIP:    net.ParseIP("10.0.0.1").To4(),
				SrcPort:  nil,
				DstPort:  nil,
				Protocol: 6, // TCP
			},
			expectedError: "",
		},
		{
			name:   "IPv6 packet with insufficient data for ports",
			packet: createTestIPv6PacketWithoutPorts("2001:db8::1", "2001:db8::2", 6),
			expectedResult: &PacketInfo{
				SrcIP:    net.ParseIP("2001:db8::1"),
				DstIP:    net.ParseIP("2001:db8::2"),
				SrcPort:  nil,
				DstPort:  nil,
				Protocol: 6, // TCP
			},
			expectedError: "",
		},
	}

	// Run all test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			info, err := ExtractIPAndPorts(tc.packet)

			// Check error
			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Check result
			if tc.expectedResult != nil {
				assert.NotNil(t, info)
				assert.Equal(t, tc.expectedResult.SrcIP.String(), info.SrcIP.String())
				assert.Equal(t, tc.expectedResult.DstIP.String(), info.DstIP.String())
				assert.Equal(t, tc.expectedResult.Protocol, info.Protocol)

				// Check ports if expected
				if tc.expectedResult.SrcPort != nil {
					assert.NotNil(t, info.SrcPort)
					assert.Equal(t, *tc.expectedResult.SrcPort, *info.SrcPort)
				} else {
					assert.Nil(t, info.SrcPort)
				}

				if tc.expectedResult.DstPort != nil {
					assert.NotNil(t, info.DstPort)
					assert.Equal(t, *tc.expectedResult.DstPort, *info.DstPort)
				} else {
					assert.Nil(t, info.DstPort)
				}
			} else {
				assert.Nil(t, info)
			}
		})
	}
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

// Helper function to create a test IPv4 packet without port data
func createTestIPv4PacketWithoutPorts(srcIP, dstIP string, protocol uint8) []byte {
	// IPv4 header is at least 20 bytes
	headerSize := 20
	// Total packet size (no transport header)
	packetSize := headerSize

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

	return packet
}

// Helper function to create a test IPv6 packet without port data
func createTestIPv6PacketWithoutPorts(srcIP, dstIP string, protocol uint8) []byte {
	// IPv6 header is 40 bytes
	headerSize := 40
	// Total packet size (no transport header)
	packetSize := headerSize

	packet := make([]byte, packetSize)

	// Set IP version (6) and traffic class/flow label
	packet[0] = 0x60

	// Set payload length (excluding IPv6 header)
	binary.BigEndian.PutUint16(packet[4:6], uint16(0))

	// Set next header (protocol)
	packet[6] = protocol

	// Set source IP
	copy(packet[8:24], net.ParseIP(srcIP).To16())

	// Set destination IP
	copy(packet[24:40], net.ParseIP(dstIP).To16())

	return packet
}
