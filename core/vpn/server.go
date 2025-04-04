package vpn

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"net"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/songgao/water"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

// HandleP2PTraffic listens for incoming P2P streams
func (s *Service) HandleP2PTraffic(iface *water.Interface) {
	s.PeerHost.SetStreamHandler(VPNProtocol, func(stream network.Stream) {
		buf := make([]byte, 1500)
		n, err := stream.Read(buf)
		if err != nil {
			log.Println("Error reading P2P stream:", err)
			return
		}
		log.Printf("Received P2P traffic: %x\n", buf[:n])

		packet := buf[:n]

		if s.IsProvider {
			destIP, srcIP, err := ExtractDestAndSrcIP(buf[:n])
			if err != nil {
				log.Errorf("failed to destIP and srcIP: %v", err)
				return
			}

			// Save the peerID of the sender for sending response later
			s.requestMap[srcIP] = stream.Conn().RemotePeer()

			appIdInt, err := GetAppIDFromVirtualIP(destIP)
			if err != nil {
				log.Errorf("failed to get appID from virtual ID: %v", err)
				return
			}

			appId := new(big.Int).SetInt64(appIdInt)
			containerId := atypes.GetContainerIdFromAppId(appId)

			// Get container IP
			container, err := s.dockerClient.ContainerInspect(context.Background(), containerId)
			if err != nil {
				log.Errorf("failed to inspect container: %v", err)
				return
			}

			containerIP := container.NetworkSettings.IPAddress
			// Save the container IP for sending response later
			s.responseMap[containerIP] = destIP

			// Rewrite destination IP to container IP
			err = UpdatePacketDestination(packet, net.ParseIP(containerIP))
			if err != nil {
				log.Errorf("failed to update packet destination: %v", err)
			}
		}
		destIPNew, srcIPNew, err := ExtractDestAndSrcIP(packet)
		if err == nil {
			// Save the peerID of the sender for sending response later
			s.requestMap[srcIPNew] = stream.Conn().RemotePeer()
			log.Println("src", srcIPNew)
			log.Println("dest", destIPNew)
		}
		_, err = iface.Write(packet)
		if err != nil {
			log.Errorf("failed to write to iface: %v", err)
		}
	})
}

// ExtractDestAndSrcIP extracts the destination and source IP addresses from an IPv4 packet.
func ExtractDestAndSrcIP(packet []byte) (destIP, srcIP string, err error) {
	// Ensure packet is at least as long as the IPv4 header (20 bytes)
	if len(packet) < 20 {
		return "", "", errors.New("packet too short to contain an IPv4 header")
	}

	// Check if it's an IPv4 packet (first 4 bits of first byte)
	if (packet[0] >> 4) != 4 {
		return "", "", errors.New("not an IPv4 packet")
	}

	// Extract source and destination IPs from the IPv4 header
	srcIP = net.IPv4(packet[12], packet[13], packet[14], packet[15]).String()
	destIP = net.IPv4(packet[16], packet[17], packet[18], packet[19]).String()

	return destIP, srcIP, nil
}

// UpdatePacketDestination updates the destination IP of an IP packet.
func UpdatePacketDestination(packet []byte, newDestIP net.IP) error {
	// Ensure the packet is at least 20 bytes long (minimum IP header size)
	if len(packet) < 20 {
		return fmt.Errorf("packet too short to be a valid IP packet")
	}

	// Extract IP version (first 4 bits)
	ipVersion := (packet[0] >> 4) & 0xF
	if ipVersion != 4 {
		return fmt.Errorf("only IPv4 is supported")
	}

	// Update destination IP in the packet (bytes 16-19 in IPv4 header)
	copy(packet[16:20], newDestIP.To4())

	// Recalculate checksum (IP header checksum at bytes 10-11)
	binary.BigEndian.PutUint16(packet[10:12], 0) // Reset checksum field
	checksum := calculateIPChecksum(packet[:20])
	binary.BigEndian.PutUint16(packet[10:12], checksum)

	return nil
}

// calculateIPChecksum calculates the checksum for the IP header.
func calculateIPChecksum(header []byte) uint16 {
	var sum uint32
	for i := 0; i < len(header); i += 2 {
		if i+1 < len(header) {
			sum += uint32(binary.BigEndian.Uint16(header[i : i+2]))
		} else {
			sum += uint32(header[i]) // Handle odd length case
		}
	}

	// Fold sum to 16 bits
	for sum > 0xFFFF {
		sum = (sum >> 16) + (sum & 0xFFFF)
	}
	return ^uint16(sum)
}
