package vpn

import (
	"context"
	"fmt"
	"net"
	"time"

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
	if err := netlink.LinkSetMTU(link, s.config.MTU); err != nil {
		return nil, fmt.Errorf("failed to set MTU: %w", err)
	}

	// Assign IP address
	addr, err := netlink.ParseAddr(fmt.Sprintf("%s/%d", s.config.VirtualIP, s.config.Subnet))
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
	for _, r := range s.config.Routes {
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
	// Get a buffer from the pool
	bufPtr := s.bufferPool.Get().(*[]byte)
	buf := *bufPtr

	defer s.bufferPool.Put(bufPtr)

	for {
		select {
		case <-s.stopChan:
			return nil
		default:
			n, err := iface.Read(buf)
			if err != nil {
				log.Errorf("Error reading from TUN interface: %v", err)
				// Add a short sleep to prevent tight loop in case of persistent errors
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Update metrics
			s.metrics.mutex.Lock()
			s.metrics.packetsReceived++
			s.metrics.bytesReceived += int64(n)
			s.metrics.mutex.Unlock()

			// Create a copy that will persist beyond this function
			packet := make([]byte, n)
			copy(packet, buf[:n])
			packetInfo, err := ExtractIPAndPorts(packet)
			if err != nil {
				log.Debugf("failed to parse the packet info: %v", err)
				s.IncrementPacketsDropped()
				continue
			}

			// Get destination IP for synchronization
			destIP := packetInfo.DstIP.String()

			// Create a synchronization key based on IP and port
			syncKey := destIP
			if packetInfo.DstPort != nil {
				// If port is available, use IP:Port as the key
				syncKey = fmt.Sprintf("%s:%d", destIP, *packetInfo.DstPort)
			}

			// Dispatch the packet to the appropriate worker
			s.dispatcher.DispatchPacket(ctx, syncKey, destIP, packet)
		}
	}
}

func (s *Service) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (VPNStream, error) {
	stream, err := s.PeerHost.NewStream(ctx, peerID, VPNProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to create new VPN P2P stream: %v", err)
	}

	return stream, nil
}
