package network

import (
	"context"
	"fmt"
	"time"

	"github.com/songgao/water"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/utils"
)

// ClientService handles client-side VPN operations
type ClientService struct {
	// TUN service for managing the TUN interface
	tunService *TUNService
	// Packet dispatcher for routing packets
	dispatcher dispatch.DispatcherService
	// Buffer pool for packet processing
	bufferPool *utils.BufferPool
	// Stop channel for graceful shutdown
	stopChan chan struct{}
}

// NewClientService creates a new client service
func NewClientService(
	tunService *TUNService,
	dispatcher dispatch.DispatcherService,
	bufferPool *utils.BufferPool,
) *ClientService {
	return &ClientService{
		tunService: tunService,
		dispatcher: dispatcher,
		bufferPool: bufferPool,
		stopChan:   make(chan struct{}),
	}
}

// Start starts the client service
// If an existing TUN interface is provided, it will be used instead of creating a new one
func (s *ClientService) Start(ctx context.Context, existingIface *water.Interface) error {
	var iface *water.Interface
	var err error

	if existingIface != nil {
		// Use the existing TUN interface
		log.Debug("Using existing TUN interface")
		iface = existingIface
	} else {
		// Set up a new TUN interface
		log.Debug("Creating new TUN interface")
		iface, err = s.tunService.SetupTUN()
		if err != nil {
			return fmt.Errorf("failed to setup TUN interface: %v", err)
		}
	}

	// Start listening for packets from the TUN interface
	go s.listenFromTUN(ctx, iface)

	return nil
}

// Stop stops the client service
func (s *ClientService) Stop() error {
	close(s.stopChan)
	return s.tunService.Close()
}

// Close implements the io.Closer interface
func (s *ClientService) Close() error {
	return s.Stop()
}

// listenFromTUN listens for packets from the TUN interface
func (s *ClientService) listenFromTUN(ctx context.Context, iface *water.Interface) {
	// Get a buffer from the pool
	buf := s.bufferPool.Get()
	defer s.bufferPool.Put(buf)

	for {
		select {
		case <-s.stopChan:
			return
		default:
			n, err := iface.Read(buf)
			if err != nil {
				log.Errorf("Error reading from TUN interface: %v", err)
				// Add a short sleep to prevent tight loop in case of persistent errors
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Create a copy that will persist beyond this function
			packetData := make([]byte, n)
			copy(packetData, buf[:n])

			// Extract packet information
			packetInfo, err := types.ExtractPacketInfo(packetData)
			if err != nil {
				log.Debugf("failed to parse the packet info: %v", err)
				continue
			}

			// Get destination IP for synchronization
			destIP := packetInfo.DstIP.String()

			// Create a synchronization key based on destination IP and destination port
			syncKey := destIP

			// Add destination port to the key if available
			if packetInfo.DstPort != nil {
				syncKey = fmt.Sprintf("%s:%d", syncKey, *packetInfo.DstPort)
			}

			// Dispatch the packet to the appropriate worker
			s.dispatcher.DispatchPacket(ctx, types.ConnectionKey(syncKey), destIP, packetData)
		}
	}
}
