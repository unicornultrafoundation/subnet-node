package vpn

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

// CreateTUN initializes a TUN device using water package
func (s *Service) CreateTUN(ip string, subnet string) (*water.Interface, error) {
	config := water.Config{DeviceType: water.TUN}
	iface, err := water.New(config)
	if err != nil {
		return nil, err
	}

	link, err := netlink.LinkByName(iface.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to find interface: %w", err)
	}

	// Assign IP address
	addr, err := netlink.ParseAddr(ip + "/" + subnet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse IP: %w", err)
	}
	if err := netlink.AddrAdd(link, addr); err != nil {
		return nil, fmt.Errorf("failed to assign IP: %w", err)
	}

	// Bring the interface up
	if err := netlink.LinkSetUp(link); err != nil {
		return nil, fmt.Errorf("failed to bring up interface: %w", err)
	}

	fmt.Println("TUN interface created:", iface.Name())

	return iface, nil
}

// Listen from TUN Interface & redirect requests/packets to Peer via p2p stream
func (s *Service) listenFromTUN(iface *water.Interface) error {
	buf := make([]byte, 1500)

	for {
		select {
		case <-s.stopChan:
			return nil
		default:
			n, err := iface.Read(buf)
			if err != nil {
				log.Fatalf("Error reading from TUN: %v", err)
			}

			destIP, srcIP, err := ExtractDestAndSrcIP(buf[:n])
			log.Println("Received packet for IP: ", destIP)
			if err != nil {
				log.Errorf("failed to parse the addresses: %v", err)
				continue
			}

			if destPeerID, exists := s.requestMap[srcIP]; exists {
				s.SendTrafficToPeer(destPeerID, buf[:n])
			} else if len(s.ConnectingPeerID) != 0 {
				s.SendTrafficToPeer(s.ConnectingPeerID, buf[:n])
			}
		}
	}
}

// SendTrafficToPeer forwards packets over P2P
func (s *Service) SendTrafficToPeer(destPeer peer.ID, data []byte) error {
	// peerID, err := s.GetPeerID(destIP)
	// if err != nil {
	// 	return fmt.Errorf("no peer mapping found for IP %s: %v", destIP, err)
	// }
	// parsedPeerID, err := peer.Decode(peerID)
	// if err != nil {
	// 	return fmt.Errorf("failed to parse peerid %s: %v", peerID, err)
	// }

	parsedPeerID := destPeer

	ctx := context.Background()
	stream, err := s.PeerHost.NewStream(ctx, parsedPeerID, VPNProtocol)
	if err != nil {
		return fmt.Errorf("failed to create P2P stream: %v", err)
	}
	defer stream.Close()

	_, err = stream.Write(data)
	if err != nil {
		log.Println("Error writing to P2P stream:", err)
	}
	fmt.Println("Sent P2P traffic to:", parsedPeerID.String())

	return nil
}
