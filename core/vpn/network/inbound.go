package network

import (
	"encoding/binary"
	"io"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/utils"
	"github.com/unicornultrafoundation/subnet-node/firewall"
)

// InboundConfig contains configuration for the inbound packet service
type InboundConfig struct {
	// MTU for the TUN interface
	MTU int
}

// InboundPacketService handles inbound packets from the network to the TUN device
type InboundPacketService struct {
	// TUN service
	tunService *TUNService
	// Configuration
	config *InboundConfig
	// Logger
	logger *logrus.Entry
	// Firewall instance
	firewall *firewall.Firewall
}

// NewInboundPacketService creates a new inbound packet service
func NewInboundPacketService(tunService *TUNService, config *InboundConfig) *InboundPacketService {
	// Create a new logger
	logger := logrus.WithField("service", "vpn-inbound")

	return &InboundPacketService{
		tunService: tunService,
		config:     config,
		logger:     logger,
	}
}

// SetFirewall sets the firewall instance for the inbound service
func (s *InboundPacketService) SetFirewall(fw *firewall.Firewall) {
	s.firewall = fw
}

// HandleStream handles an incoming stream from a peer
func (s *InboundPacketService) HandleStream(stream network.Stream) {
	defer stream.Close()

	// Get the peer ID
	peer := stream.Conn().RemotePeer().String()
	s.logger.WithField("peer", peer).Debug("New VPN stream established")

	// Ensure the TUN device is set up
	if s.tunService.device == nil {
		s.logger.WithField("peer", peer).Error("TUN device not set up")
		return
	}

	// Pre-allocate all buffers to reduce GC pressure
	combinedBuf := make([]byte, s.config.MTU+4) // 4 bytes for length + max packet size
	fwPacket := &firewall.Packet{}

	for {
		// Read the packet length and data in potentially multiple operations
		// but minimize allocations
		_, err := io.ReadAtLeast(stream, combinedBuf[:4], 4)
		if err != nil {
			if err != io.EOF {
				s.logger.WithError(err).WithField("peer", peer).Error("Error reading packet length")
			} else {
				s.logger.WithField("peer", peer).Debug("Stream closed by peer")
			}
			return
		}

		// Parse the packet length
		packetLength := binary.BigEndian.Uint32(combinedBuf[:4])
		if packetLength == 0 || packetLength > uint32(s.config.MTU) {
			s.logger.WithField("length", packetLength).WithField("peer", peer).Error("Invalid packet length")
			return
		}

		// Read the actual packet data
		_, err = io.ReadFull(stream, combinedBuf[4:4+packetLength])
		if err != nil {
			if err != io.EOF {
				s.logger.WithError(err).WithField("peer", peer).Error("Error reading packet data")
			} else {
				s.logger.WithField("peer", peer).Debug("Unexpected EOF while reading packet")
			}
			return
		}

		// Process the packet - make a copy if needed for async processing
		// to avoid race conditions with the buffer
		packetCopy := make([]byte, packetLength)
		copy(packetCopy, combinedBuf[4:4+packetLength])

		s.processInboundPacket(packetCopy, fwPacket, peer)
	}
}

// processInboundPacket processes an inbound packet from a peer
func (s *InboundPacketService) processInboundPacket(packet []byte, fwPacket *firewall.Packet, peerID string) {
	// Parse the packet
	err := utils.ParsePacket(packet, true, fwPacket)
	if err != nil {
		s.logger.WithError(err).Error("Error while parsing inbound packet")
		return
	}

	// Check firewall rules for inbound traffic
	if s.firewall != nil {
		err := s.firewall.Drop(*fwPacket, true, nil)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"error": err,
				"peer":  peerID,
				"port":  fwPacket.RemotePort,
			}).Debug("Packet dropped by firewall")
			return
		}
	}

	// Get the readers from the TUN service
	readers := s.tunService.GetReaders()
	routines := s.tunService.GetRoutines()

	// Determine which reader to use based on peer ID hash for load balancing
	var readerIdx int
	if routines > 1 && peerID != "" {
		// Simple hash function - sum the bytes of the peer ID string
		var hash uint64
		for i := 0; i < len(peerID); i++ {
			hash += uint64(peerID[i])
		}
		readerIdx = int(hash % uint64(routines))
	}

	// Write the packet to the appropriate reader
	var writer io.Writer
	if readerIdx < len(readers) && readers[readerIdx] != nil {
		writer = readers[readerIdx]
	} else {
		// Fallback to main interface if something went wrong with reader selection
		writer = s.tunService.device
	}

	// Write the packet to the selected writer
	_, err = writer.Write(packet)
	if err != nil {
		s.logger.WithError(err).Error("Failed to write packet to TUN device")
		return
	}
}

// Close closes the inbound packet service
func (s *InboundPacketService) Close() error {
	return nil
}
