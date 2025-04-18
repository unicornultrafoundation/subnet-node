package vpn

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/patrickmn/go-cache"
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

			packet := make([]byte, n)
			copy(packet, buf[:n])
			packetInfo, err := ExtractIPAndPorts(packet)
			if err != nil {
				log.Debugf("failed to parse the packet info: %v", err)
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

			// Get the packet queue for this destination
			queue := s.getOrCreatePacketQueue(syncKey)

			// Create a packet and add it to the queue (non-blocking)
			packetObj := &QueuedPacket{
				Ctx:    ctx,
				DestIP: destIP,
				Data:   packet,
				// No DoneCh since we don't need to wait for completion
			}

			// Try to add the packet to the queue without blocking
			select {
			case queue <- packetObj:
				// Successfully added to queue
			default:
				// Queue is full, log and continue
				log.Debugf("Packet queue full for %s, dropping packet", syncKey)
			}
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

// getOrCreatePacketQueue gets or creates a packet queue for the given synchronization key
func (s *Service) getOrCreatePacketQueue(syncKey string) chan *QueuedPacket {
	s.packetQueuesMu.Lock()
	defer s.packetQueuesMu.Unlock()

	queueVal, exists := s.packetQueues.Load(syncKey)
	if exists {
		return queueVal.(chan *QueuedPacket)
	}

	// Create a new queue with buffer size 100
	queue := make(chan *QueuedPacket, 100)
	s.packetQueues.Store(syncKey, queue)

	// Start a goroutine to process this queue
	go s.processPacketQueue(syncKey, queue)

	return queue
}

// processPacketQueue processes packets from the queue in sequential order
func (s *Service) processPacketQueue(syncKey string, queue chan *QueuedPacket) {
	for {
		select {
		case <-s.stopChan:
			return
		case packet, ok := <-queue:
			if !ok {
				// Channel was closed
				return
			}

			// Update activity cache for this synchronization key
			s.ipActivityMu.Lock()
			if s.ipActivityCache != nil {
				s.ipActivityCache.Set(syncKey, time.Now(), cache.DefaultExpiration)
			}
			s.ipActivityMu.Unlock()

			// Process the packet
			operation := func() error {
				return s.SendTrafficToPeer(packet.Ctx, packet.DestIP, packet.Data)
			}

			bo := backoff.NewExponentialBackOff()
			bo.MaxElapsedTime = 30 * time.Second

			err := backoff.Retry(operation, backoff.WithContext(bo, packet.Ctx))

			// Signal completion
			if packet.DoneCh != nil {
				packet.DoneCh <- err
				close(packet.DoneCh)
			}
		}
	}
}

func (s *Service) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (network.Stream, error) {
	stream, err := s.PeerHost.NewStream(ctx, peerID, VPNProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to create new VPN P2P stream: %v", err)
	}
	s.streamCache[peerID.String()] = stream

	return stream, nil
}
