package vpn

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

// CreateTUN initializes a TUN device using water package
func (s *Service) SetupTUN() (*water.Interface, error) {
	config := water.Config{DeviceType: water.TUN}
	iface, err := water.New(config)
	if err != nil {
		return nil, err
	}

	link, err := netlink.LinkByName(iface.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to find interface: %w", err)
	}

	// Set MTU
	if err := netlink.LinkSetMTU(link, s.mtu); err != nil {
		return nil, fmt.Errorf("failed to set MTU: %w", err)
	}

	// Assign IP address
	addr, err := netlink.ParseAddr(fmt.Sprintf("%s/%d", s.virtualIP, s.subnet))
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

	// Add after iface creation
	for _, r := range s.routes {
		_, dst, err := net.ParseCIDR(r)
		if err != nil {
			log.Errorf("Invalid route: %s", r)
			continue
		}

		route := &netlink.Route{
			LinkIndex: link.Attrs().Index,
			Dst:       dst,
		}

		if err := netlink.RouteReplace(route); err != nil {
			log.Errorf("Failed to add route %s: %v", dst, err)
		}
	}

	log.Infof("TUN interface created with name %s, address %s", iface.Name(), addr.String())

	return iface, nil
}

// Listen from TUN Interface & redirect requests/packets to Peer via p2p stream
func (s *Service) listenFromTUN(ctx context.Context, iface *water.Interface) error {
	buf := make([]byte, s.mtu)

	for {
		select {
		case <-s.stopChan:
			return nil
		default:
			n, err := iface.Read(buf)
			if err != nil {
				log.Fatalf("error reading from TUN interface: %v", err)
			}

			packet := buf[:n]
			packetInfo, err := ExtractIPAndPorts(packet)
			if err != nil {
				log.Debugf("failed to parse the packet info: %v", err)
				continue
			}

			go func() {
				err = s.RetrySendTrafficToPeer(ctx, packetInfo.DstIP.String(), packet)
				if err != nil {
					log.Debugf("failed to send traffic to peer: %v", err)
				}
			}()
		}
	}
}

// SendTrafficToPeer forwards packets over P2P
func (s *Service) SendTrafficToPeer(ctx context.Context, destIP string, data []byte) error {
	peerID, err := s.GetPeerID(ctx, destIP)
	if err != nil {
		return fmt.Errorf("no peer mapping found for IP %s: %v", destIP, err)
	}

	parsedPeerID, err := peer.Decode(peerID)
	if err != nil {
		return fmt.Errorf("failed to parse peerid %s: %v", peerID, err)
	}

	stream, exist := s.streamCache[peerID]
	if !exist {
		stream, err = s.CreateNewVPNStream(ctx, parsedPeerID)
		if err != nil {
			return fmt.Errorf("failed to create P2P stream: %v", err)
		}
	}

	_, err = stream.Write(data)
	if err != nil {
		stream.Close()
		delete(s.streamCache, peerID)
		return fmt.Errorf("error writing to P2P stream: %v", err)
	}

	return nil
}

func (s *Service) RetrySendTrafficToPeer(ctx context.Context, destIP string, data []byte) error {
	operation := func() error {
		return s.SendTrafficToPeer(ctx, destIP, data)
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 30 * time.Second

	return backoff.Retry(operation, backoff.WithContext(bo, ctx))
}

func (s *Service) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (network.Stream, error) {
	stream, err := s.PeerHost.NewStream(ctx, peerID, VPNProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to create new VPN P2P stream: %v", err)
	}
	s.streamCache[peerID.String()] = stream

	return stream, nil
}
