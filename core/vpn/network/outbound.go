package network

import (
	"context"
	"errors"
	"io"
	"os"
	"runtime"
	"sync/atomic"

	"github.com/sirupsen/logrus"
	vpnconfig "github.com/unicornultrafoundation/subnet-node/core/vpn/config"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatcher"
)

// OutboundConfig contains configuration for the outbound packet service
type OutboundConfig struct {
	// MTU for the TUN interface
	MTU int
}

// OutboundPacketService handles outbound packets from the TUN device to the network
type OutboundPacketService struct {
	// TUN service
	tunService *TUNService
	// Packet dispatcher
	dispatcher dispatcher.DispatcherService
	// Configuration
	config *OutboundConfig
	// Logger
	logger *logrus.Entry
	// Flag to indicate if the service is closed
	closed atomic.Bool
}

// NewOutboundPacketService creates a new outbound packet service
func NewOutboundPacketService(tunService *TUNService, dispatcher dispatcher.DispatcherService, configService vpnconfig.ConfigService) *OutboundPacketService {
	logger := logrus.WithField("service", "vpn-outbound")

	return &OutboundPacketService{
		tunService: tunService,
		dispatcher: dispatcher,
		config:     &OutboundConfig{MTU: configService.GetMTU()},
		logger:     logger,
		closed:     atomic.Bool{},
	}
}

// Start initializes and starts the outbound packet service
func (s *OutboundPacketService) Start(ctx context.Context) error {
	// Get the readers from the TUN service
	readers := s.tunService.GetReaders()

	// Start reader routines
	for i := 0; i < len(readers); i++ {
		go s.listenTUN(ctx, readers[i], i)
	}

	return nil
}

// listenTUN reads packets from the TUN device and dispatches them
func (s *OutboundPacketService) listenTUN(ctx context.Context, reader io.ReadWriteCloser, queueID int) {
	// Lock the OS thread to ensure the routine runs on the same thread
	runtime.LockOSThread()

	// Create a buffer for reading packets
	packet := make([]byte, s.config.MTU)

	// Read packets from the TUN device
	for {
		select {
		case <-ctx.Done():
			// Context canceled, exit the routine
			return
		default:
			// Read a packet from the TUN device
			n, err := reader.Read(packet)
			if err != nil {
				if errors.Is(err, os.ErrClosed) && s.closed.Load() {
					return
				}

				s.logger.WithError(err).Error("Error while reading outbound packet")
				// This only seems to happen when something fatal happens to the fd, so exit.
				os.Exit(2)
			}

			// Process the packet
			s.processOutboundPacket(ctx, packet[:n], queueID)
		}
	}
}

// processOutboundPacket processes an outbound packet from the TUN device
func (s *OutboundPacketService) processOutboundPacket(ctx context.Context, packet []byte, queueID int) {
	// Dispatch the packet with the queue ID
	err := s.dispatcher.DispatchPacket(ctx, packet, queueID)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"queueID": queueID,
		}).Error("Failed to dispatch packet")
		return
	}
}

// Close closes the outbound packet service
func (s *OutboundPacketService) Close() error {
	s.closed.Store(true)
	return nil
}
