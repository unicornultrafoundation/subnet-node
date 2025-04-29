package network

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
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

	logger := logrus.WithField("service", "client")
	if existingIface != nil {
		// Use the existing TUN interface
		logger.Debug("Using existing TUN interface")
		iface = existingIface
	} else {
		// Set up a new TUN interface
		logger.Debug("Creating new TUN interface")
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
// This is an optimized version that uses multiple readers and batch processing
func (s *ClientService) listenFromTUN(ctx context.Context, iface *water.Interface) {
	// Determine the number of reader goroutines based on CPU cores
	numReaders := runtime.NumCPU()
	if numReaders > 8 {
		// Cap at 8 readers to avoid excessive overhead
		numReaders = 8
	}

	logger := logrus.WithField("service", "client")
	logger.Infof("Starting %d TUN reader goroutines", numReaders)

	// Create a wait group to wait for all readers to finish
	var wg sync.WaitGroup
	wg.Add(numReaders)

	// Start multiple reader goroutines
	for i := 0; i < numReaders; i++ {
		go func(readerID int) {
			defer wg.Done()
			s.packetReader(ctx, iface, readerID)
		}(i)
	}

	// Wait for all readers to finish
	wg.Wait()
}

// packetReader is a worker that reads packets from the TUN interface
func (s *ClientService) packetReader(ctx context.Context, iface *water.Interface, readerID int) {
	logger := logrus.WithField("reader_id", readerID)
	logger.Debug("Starting TUN packet reader")

	// Each reader has its own buffer
	buf := s.bufferPool.Get()
	defer s.bufferPool.Put(buf)

	// Check if the dispatcher supports batch processing
	batchDispatcher, supportsBatch := s.dispatcher.(types.BatchDispatcher)

	// Create batch processing variables if supported
	var (
		batchSize     = 16 // Default batch size
		batchInterval = 5 * time.Millisecond
		packets       []*types.QueuedPacket
		connKeys      []types.ConnectionKey
		destIPs       []string
		contexts      []context.Context
	)

	if supportsBatch {
		logger.Info("Batch processing enabled")
		packets = make([]*types.QueuedPacket, 0, batchSize)
		connKeys = make([]types.ConnectionKey, 0, batchSize)
		destIPs = make([]string, 0, batchSize)
		contexts = make([]context.Context, 0, batchSize)
	}

	// Create a ticker for batch flushing if batch processing is supported
	var flushTicker *time.Ticker
	if supportsBatch {
		flushTicker = time.NewTicker(batchInterval)
		defer flushTicker.Stop()
	}

	// Process packets
	for {
		// Check if we should stop
		select {
		case <-s.stopChan:
			logger.Debug("Stopping TUN packet reader due to stop signal")
			return
		case <-ctx.Done():
			logger.Debug("Stopping TUN packet reader due to context cancellation")
			return
		default:
			// Continue processing
		}

		// If batch processing is enabled, check if we need to flush the batch
		if supportsBatch && flushTicker != nil {
			select {
			case <-flushTicker.C:
				// Flush any pending packets
				if len(packets) > 0 {
					err := batchDispatcher.DispatchPacketBatch(packets, connKeys, destIPs, contexts)
					if err != nil {
						logger.WithError(err).Warn("Failed to dispatch packet batch")
					}

					// Clear the batch
					packets = packets[:0]
					connKeys = connKeys[:0]
					destIPs = destIPs[:0]
					contexts = contexts[:0]
				}
				continue
			default:
				// No need to flush yet
			}
		}

		// Read a packet - we can't set a deadline on the TUN interface
		// so we'll use a non-blocking approach with select
		n, err := iface.Read(buf)
		if err != nil {
			logger.WithError(err).Error("Error reading from TUN interface")
			// Add a short sleep to prevent tight loop in case of persistent errors
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Use zero-copy approach with buffer pool
		packetBuf := s.bufferPool.Get()
		copy(packetBuf[:n], buf[:n])
		packetData := packetBuf[:n]

		// Extract packet information
		packetInfo, err := types.ExtractPacketInfo(packetData)
		if err != nil {
			logger.WithError(err).Debug("Failed to parse packet info")
			// Return the buffer to the pool
			s.bufferPool.Put(packetBuf)
			continue
		}

		// Get destination IP for synchronization
		destIP := packetInfo.DstIP.String()

		// Create a synchronization key based on destination IP and ports
		syncKey := destIP

		// Add source port to the key if available
		if packetInfo.SrcPort != nil {
			syncKey = fmt.Sprintf("%d:%s", *packetInfo.SrcPort, syncKey)
		}

		// Add destination port to the key if available
		if packetInfo.DstPort != nil {
			syncKey = fmt.Sprintf("%s:%d", syncKey, *packetInfo.DstPort)
		}

		// Create connection key
		connKey := types.ConnectionKey(syncKey)

		// If batch processing is supported, add to batch
		if supportsBatch {
			// Create a queued packet
			packet := &types.QueuedPacket{
				Ctx:    ctx,
				DestIP: destIP,
				Data:   packetData,
			}

			// Add to batch
			packets = append(packets, packet)
			connKeys = append(connKeys, connKey)
			destIPs = append(destIPs, destIP)
			contexts = append(contexts, ctx)

			// If batch is full, send it
			if len(packets) >= batchSize {
				err := batchDispatcher.DispatchPacketBatch(packets, connKeys, destIPs, contexts)
				if err != nil {
					logger.WithError(err).Warn("Failed to dispatch packet batch")
				}

				// Clear the batch
				packets = packets[:0]
				connKeys = connKeys[:0]
				destIPs = destIPs[:0]
				contexts = contexts[:0]
			}
		} else {
			// Use regular dispatch for single packets
			err = s.dispatcher.DispatchPacket(ctx, connKey, destIP, packetData)
			if err != nil {
				logger.WithError(err).Debug("Failed to dispatch packet")
				// Return the buffer to the pool on error
				s.bufferPool.Put(packetBuf)
			}
		}
	}
}
