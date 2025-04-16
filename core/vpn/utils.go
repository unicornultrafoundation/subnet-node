package vpn

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/libp2p/go-libp2p/core/host"
)

func ExtractIPAndPorts(packet []byte) (*PacketInfo, error) {
	if len(packet) < 1 {
		return nil, errors.New("empty packet")
	}

	version := packet[0] >> 4
	switch version {
	case 4:
		return extractIPv4Info(packet)
	case 6:
		return extractIPv6Info(packet)
	default:
		return nil, fmt.Errorf("unsupported IP version: %d", version)
	}
}

func extractIPv6Info(packet []byte) (*PacketInfo, error) {
	if len(packet) < 40 {
		return nil, errors.New("packet too short for IPv6")
	}

	nextHeader := packet[6]
	srcIP := net.IP(packet[8:24])
	dstIP := net.IP(packet[24:40])

	var srcPort, dstPort int
	switch nextHeader {
	case 6, 17: // TCP or UDP
		if len(packet) < 44 {
			return nil, errors.New("packet too short for transport header")
		}
		srcPort = int(binary.BigEndian.Uint16(packet[40:42]))
		dstPort = int(binary.BigEndian.Uint16(packet[42:44]))
	default:
		return &PacketInfo{
			SrcIP:    srcIP,
			DstIP:    dstIP,
			Protocol: nextHeader,
		}, nil
	}

	return &PacketInfo{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		SrcPort:  &srcPort,
		DstPort:  &dstPort,
		Protocol: nextHeader,
	}, nil
}

func extractIPv4Info(packet []byte) (*PacketInfo, error) {
	if len(packet) < 20 {
		return nil, errors.New("packet too short to contain IPv4 header")
	}

	version := packet[0] >> 4
	if version != 4 {
		return nil, fmt.Errorf("unsupported IP version: %d", version)
	}

	ipHeaderLen := int(packet[0]&0x0F) * 4
	if len(packet) < ipHeaderLen {
		return nil, errors.New("invalid IP header length")
	}

	protocol := packet[9]
	srcIP := net.IPv4(packet[12], packet[13], packet[14], packet[15])
	dstIP := net.IPv4(packet[16], packet[17], packet[18], packet[19])

	var srcPort, dstPort *int

	switch protocol {
	case 6, 17: // TCP or UDP
		if len(packet) < ipHeaderLen+4 {
			return nil, errors.New("packet too short for transport header")
		}
		sport := int(binary.BigEndian.Uint16(packet[ipHeaderLen : ipHeaderLen+2]))
		dport := int(binary.BigEndian.Uint16(packet[ipHeaderLen+2 : ipHeaderLen+4]))
		srcPort = &sport
		dstPort = &dport
	}

	return &PacketInfo{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		SrcPort:  srcPort,
		DstPort:  dstPort,
		Protocol: protocol,
	}, nil
}

func WaitUntilPeerConnected(ctx context.Context, h host.Host) error {
	operation := func() error {
		if len(h.Network().Peers()) == 0 {
			return fmt.Errorf("no peers connected")
		}
		return nil
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 30 * time.Second // or whatever timeout you want

	return backoff.Retry(operation, backoff.WithContext(bo, ctx))
}

// GenerateVirtualIP returns a random IP within the given CIDR range (excluding network and broadcast addresses)
func GenerateVirtualIP(cidr string) (net.IP, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %w", err)
	}

	// Convert IPNet to big.Int
	ipInt := big.NewInt(0).SetBytes(ip.To4())
	maskSize, totalBits := ipnet.Mask.Size()

	// Calculate number of possible IPs in the subnet
	numIPs := big.NewInt(1)
	numIPs.Lsh(numIPs, uint(totalBits-maskSize)) // 2^(32 - mask)

	// Exclude network (first) and broadcast (last) address
	if numIPs.Cmp(big.NewInt(2)) <= 0 {
		return nil, fmt.Errorf("CIDR block too small to generate IPs")
	}
	numIPs.Sub(numIPs, big.NewInt(2)) // exclude first and last

	// Seed randomness
	rand.Seed(time.Now().UnixNano())

	// Pick random offset from the first usable IP
	offset := big.NewInt(rand.Int63n(numIPs.Int64()))
	offset.Add(offset, big.NewInt(1)) // +1 to skip network IP

	// Calculate new IP
	virtualIP := big.NewInt(0).Add(ipInt, offset)
	return net.IP(virtualIP.Bytes()).To4(), nil
}

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

	// Convert to uint32
	return binary.BigEndian.Uint32(ipv4)
}
