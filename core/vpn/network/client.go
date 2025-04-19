package network

import (
	"context"
	"fmt"
	"time"

	"github.com/songgao/water"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/packet"
)

// ClientService handles client-side VPN operations
type ClientService struct {
	// TUN service for managing the TUN interface
	tunService *TUNService
	// Packet dispatcher for routing packets
	dispatcher packet.DispatcherService
	// Metrics for monitoring
	metrics VPNMetricsInterface
	// Buffer pool for packet processing
	bufferPool *BufferPool
	// Stop channel for graceful shutdown
	stopChan chan struct{}
}

// NewClientService creates a new client service
func NewClientService(
	tunService *TUNService,
	dispatcher packet.DispatcherService,
	metrics VPNMetricsInterface,
	bufferPool *BufferPool,
) *ClientService {
	return &ClientService{
		tunService: tunService,
		dispatcher: dispatcher,
		metrics:    metrics,
		bufferPool: bufferPool,
		stopChan:   make(chan struct{}),
	}
}

// Start starts the client service
func (s *ClientService) Start(ctx context.Context) error {
	// Set up the TUN interface
	iface, err := s.tunService.SetupTUN()
	if err != nil {
		return fmt.Errorf("failed to setup TUN interface: %v", err)
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

			// Update metrics
			s.metrics.IncrementPacketsReceived(n)

			// Create a copy that will persist beyond this function
			packetData := make([]byte, n)
			copy(packetData, buf[:n])

			// Extract packet information
			packetInfo, err := packet.ExtractIPAndPorts(packetData)
			if err != nil {
				log.Debugf("failed to parse the packet info: %v", err)
				s.metrics.IncrementPacketsDropped()
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
			s.dispatcher.DispatchPacket(ctx, syncKey, destIP, packetData)
		}
	}
}

// BufferPool manages a pool of byte buffers
type BufferPool struct {
	// Size of each buffer
	bufferSize int
	// Channel for storing buffers
	pool chan []byte
}

// NewBufferPool creates a new buffer pool
func NewBufferPool(bufferSize, poolSize int) *BufferPool {
	return &BufferPool{
		bufferSize: bufferSize,
		pool:       make(chan []byte, poolSize),
	}
}

// Get gets a buffer from the pool or creates a new one
func (p *BufferPool) Get() []byte {
	select {
	case buf := <-p.pool:
		return buf
	default:
		return make([]byte, p.bufferSize)
	}
}

// Put returns a buffer to the pool
func (p *BufferPool) Put(buf []byte) {
	select {
	case p.pool <- buf:
	default:
		// Pool is full, discard the buffer
	}
}
