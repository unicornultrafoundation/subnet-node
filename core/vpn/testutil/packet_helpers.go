package testutil

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net/netip"
	"sync"
	"time"

	"github.com/unicornultrafoundation/subnet-node/firewall"
)

// PacketGenerator generates test packets for VPN testing
type PacketGenerator struct {
	// Source IP for generated packets
	SourceIP string
	// Destination IP for generated packets
	DestinationIP string
	// Starting source port
	SourcePortStart uint16
	// Starting destination port
	DestPortStart uint16
	// Current source port
	currentSourcePort uint16
	// Current destination port
	currentDestPort uint16
	// Mutex for thread safety
	mu sync.Mutex
}

// NewPacketGenerator creates a new packet generator
func NewPacketGenerator(sourceIP, destIP string, sourcePortStart, destPortStart uint16) *PacketGenerator {
	return &PacketGenerator{
		SourceIP:          sourceIP,
		DestinationIP:     destIP,
		SourcePortStart:   sourcePortStart,
		DestPortStart:     destPortStart,
		currentSourcePort: sourcePortStart,
		currentDestPort:   destPortStart,
	}
}

// GenerateIPv4Packet generates a simple IPv4 packet
func (g *PacketGenerator) GenerateIPv4Packet(size int) []byte {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Increment ports for next packet
	g.currentSourcePort++
	g.currentDestPort++

	// Create a packet of the specified size
	packet := make([]byte, size)

	// Set IP version (4) and header length (5 32-bit words = 20 bytes)
	packet[0] = 0x45 // Version 4, header length 5

	// Set total length (header + data)
	binary.BigEndian.PutUint16(packet[2:4], uint16(size))

	// Set protocol (TCP = 6)
	packet[9] = 6

	// Set source IP
	srcIP := netip.MustParseAddr(g.SourceIP)
	copy(packet[12:16], srcIP.AsSlice())

	// Set destination IP
	dstIP := netip.MustParseAddr(g.DestinationIP)
	copy(packet[16:20], dstIP.AsSlice())

	// Set source port
	binary.BigEndian.PutUint16(packet[20:22], g.currentSourcePort)

	// Set destination port
	binary.BigEndian.PutUint16(packet[22:24], g.currentDestPort)

	// Fill the rest with random data
	_, _ = rand.Read(packet[24:])

	return packet
}

// GenerateFirewallPacket generates a firewall packet for testing
func (g *PacketGenerator) GenerateFirewallPacket() *firewall.Packet {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Increment ports for next packet
	g.currentSourcePort++
	g.currentDestPort++

	return &firewall.Packet{
		LocalAddr:  netip.MustParseAddr(g.SourceIP),
		RemoteAddr: netip.MustParseAddr(g.DestinationIP),
		LocalPort:  g.currentSourcePort,
		RemotePort: g.currentDestPort,
		Protocol:   6, // TCP
		Fragment:   false,
	}
}

// PacketCapture captures packets for testing
type PacketCapture struct {
	// Captured packets
	Packets [][]byte
	// Mutex for thread safety
	mu sync.Mutex
}

// NewPacketCapture creates a new packet capture
func NewPacketCapture() *PacketCapture {
	return &PacketCapture{
		Packets: make([][]byte, 0),
	}
}

// CapturePacket captures a packet
func (c *PacketCapture) CapturePacket(packet []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Make a copy of the packet to avoid issues with reused buffers
	packetCopy := make([]byte, len(packet))
	copy(packetCopy, packet)

	c.Packets = append(c.Packets, packetCopy)
}

// GetPackets returns all captured packets
func (c *PacketCapture) GetPackets() [][]byte {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Packets
}

// CountPackets returns the number of captured packets
func (c *PacketCapture) CountPackets() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.Packets)
}

// FindPacket searches for a packet matching the given criteria
func (c *PacketCapture) FindPacket(matcher func([]byte) bool) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, packet := range c.Packets {
		if matcher(packet) {
			return packet, true
		}
	}

	return nil, false
}

// WaitForPackets waits for a specific number of packets to be captured
func (c *PacketCapture) WaitForPackets(count int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		c.mu.Lock()
		currentCount := len(c.Packets)
		c.mu.Unlock()

		if currentCount >= count {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for packets: got %d, wanted %d", c.CountPackets(), count)
}
