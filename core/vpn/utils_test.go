package vpn_test

import (
	"encoding/binary"
	"net"
	"testing"

	"github.com/unicornultrafoundation/subnet-node/core/vpn"
)

func TestExtractIPAndPorts_TCP(t *testing.T) {
	// Build minimal TCP IPv4 packet (20 byte IP + 4 byte TCP header)
	packet := make([]byte, 24)

	// IPv4 header
	packet[0] = 0x45                                    // Version=4, IHL=5
	packet[9] = 6                                       // Protocol TCP
	copy(packet[12:16], net.IPv4(192, 168, 1, 1).To4()) // Src IP
	copy(packet[16:20], net.IPv4(10, 0, 0, 1).To4())    // Dst IP

	// TCP header (just first 4 bytes for ports)
	binary.BigEndian.PutUint16(packet[20:22], 12345) // Src port
	binary.BigEndian.PutUint16(packet[22:24], 80)    // Dst port

	info, err := vpn.ExtractIPAndPorts(packet)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Protocol != 6 {
		t.Errorf("expected protocol 6 (TCP), got %d", info.Protocol)
	}
	if info.SrcIP.String() != "192.168.1.1" || info.DstIP.String() != "10.0.0.1" {
		t.Errorf("IP mismatch: got src %s dst %s", info.SrcIP, info.DstIP)
	}
	if info.SrcPort == nil || *info.SrcPort != 12345 || info.DstPort == nil || *info.DstPort != 80 {
		t.Errorf("Port mismatch: got src %v dst %v", info.SrcPort, info.DstPort)
	}
}

func TestExtractIPAndPorts_UDP(t *testing.T) {
	packet := make([]byte, 24)
	packet[0] = 0x45 // IPv4
	packet[9] = 17   // Protocol UDP
	copy(packet[12:16], net.IPv4(10, 0, 0, 5).To4())
	copy(packet[16:20], net.IPv4(10, 0, 0, 10).To4())
	binary.BigEndian.PutUint16(packet[20:22], 53)
	binary.BigEndian.PutUint16(packet[22:24], 5353)

	info, err := vpn.ExtractIPAndPorts(packet)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Protocol != 17 {
		t.Errorf("expected UDP protocol (17), got %d", info.Protocol)
	}
	if *info.SrcPort != 53 || *info.DstPort != 5353 {
		t.Errorf("expected ports 53 -> 5353, got %d -> %d", *info.SrcPort, *info.DstPort)
	}
}

func TestExtractIPAndPorts_ICMP(t *testing.T) {
	packet := make([]byte, 20)
	packet[0] = 0x45 // IPv4
	packet[9] = 1    // Protocol ICMP
	copy(packet[12:16], net.IPv4(1, 2, 3, 4).To4())
	copy(packet[16:20], net.IPv4(5, 6, 7, 8).To4())

	info, err := vpn.ExtractIPAndPorts(packet)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Protocol != 1 {
		t.Errorf("expected ICMP protocol (1), got %d", info.Protocol)
	}
	if info.SrcPort != nil || info.DstPort != nil {
		t.Errorf("expected no ports, got src=%v dst=%v", info.SrcPort, info.DstPort)
	}
}

func TestExtractIPAndPorts_TooShort(t *testing.T) {
	packet := make([]byte, 10)

	_, err := vpn.ExtractIPAndPorts(packet)
	if err == nil {
		t.Errorf("expected error on short packet, got none")
	}
}

func TestExtractIPv6TCPPacket(t *testing.T) {
	// Build a minimal IPv6 + TCP packet
	ipv6Header := []byte{
		0x60, 0x00, 0x00, 0x00, // Version (6), traffic class, flow label
		0x00, 0x14, // Payload length (20 bytes of TCP)
		0x06, // Next header: TCP
		0x40, // Hop limit

		// Source IP (16 bytes)
		0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,

		// Destination IP (16 bytes)
		0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02,
	}

	tcpHeader := make([]byte, 20)
	binary.BigEndian.PutUint16(tcpHeader[0:2], 12345) // Source port
	binary.BigEndian.PutUint16(tcpHeader[2:4], 80)    // Destination port

	packet := append(ipv6Header, tcpHeader...)

	info, err := vpn.ExtractIPAndPorts(packet)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	wantSrc := net.ParseIP("2001:db8::1")
	wantDst := net.ParseIP("2001:db8::2")

	if !info.SrcIP.Equal(wantSrc) {
		t.Errorf("expected SrcIP %v, got %v", wantSrc, info.SrcIP)
	}
	if !info.DstIP.Equal(wantDst) {
		t.Errorf("expected DstIP %v, got %v", wantDst, info.DstIP)
	}
	if info.SrcPort == nil || *info.SrcPort != 12345 {
		t.Errorf("expected SrcPort 12345, got %v", info.SrcPort)
	}
	if info.DstPort == nil || *info.DstPort != 80 {
		t.Errorf("expected DstPort 80, got %v", info.DstPort)
	}
}

func TestExtractIPv6UnsupportedProto(t *testing.T) {
	// IPv6 + ICMPv6 (next header = 58)
	ipv6Header := []byte{
		0x60, 0x00, 0x00, 0x00,
		0x00, 0x08,
		0x3A, // ICMPv6
		0x40,
		// Source IP
		0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 1,
		// Destination IP
		0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 2,
	}
	icmpData := make([]byte, 8)
	packet := append(ipv6Header, icmpData...)

	info, err := vpn.ExtractIPAndPorts(packet)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if info.SrcPort != nil || info.DstPort != nil {
		t.Errorf("expected no ports for ICMPv6, got src: %v, dst: %v", info.SrcPort, info.DstPort)
	}
}

func TestConvertVirtualIPToNumber(t *testing.T) {
	ip := "10.0.0.1"
	expected := uint32(167772161)
	result := vpn.ConvertVirtualIPToNumber(ip)
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}

func TestConvertVirtualIPToNumberInvalidIP(t *testing.T) {
	// Test with invalid IP format
	ip := "invalid"
	result := vpn.ConvertVirtualIPToNumber(ip)
	if result != 0 {
		t.Errorf("Expected 0 for invalid IP, got %d", result)
	}

	// Test with empty string
	ip = ""
	result = vpn.ConvertVirtualIPToNumber(ip)
	if result != 0 {
		t.Errorf("Expected 0 for empty string, got %d", result)
	}

	// Test with IPv6 address (should return 0 as we only support IPv4)
	ip = "2001:db8::1"
	result = vpn.ConvertVirtualIPToNumber(ip)
	if result != 0 {
		t.Errorf("Expected 0 for IPv6 address, got %d", result)
	}
}
