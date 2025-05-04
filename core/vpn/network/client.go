package network

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/metrics"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/overlay"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/utils"
)

// ClientService handles client-side VPN operations
type ClientService struct {
	// TUN service for managing the TUN interface
	tunService *TUNService
	// Packet dispatcher for routing packets
	dispatcher dispatch.DispatcherService
	// Buffer pool for packet processing
	bufferPool utils.BufferPoolInterface
	// Stop channel for graceful shutdown
	stopChan chan struct{}
}

// NewClientService creates a new client service
func NewClientService(
	tunService *TUNService,
	dispatcher dispatch.DispatcherService,
	bufferPool utils.BufferPoolInterface,
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
func (s *ClientService) Start(ctx context.Context, existingDevice overlay.Device) error {
	var device overlay.Device
	var err error

	logger := logrus.WithField("service", "client")
	if existingDevice != nil {
		// Use the existing TUN interface
		logger.Debug("Using existing TUN interface")
		device = existingDevice
	} else {
		// Set up a new TUN interface
		logger.Debug("Creating new TUN interface")
		device, err = s.tunService.SetupTUN()
		if err != nil {
			return fmt.Errorf("failed to setup TUN interface: %v", err)
		}
	}

	// Start metrics collection
	metrics.StartMetricsCollection()

	// Start listening for packets from the TUN interface
	go s.listenFromTUN(ctx, device)

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
func (s *ClientService) listenFromTUN(ctx context.Context, device overlay.Device) {
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

	// Start the first reader with the main device
	go func() {
		defer wg.Done()
		s.packetReader(ctx, device, 0)
	}()

	// Start additional readers with multi-queue readers if supported
	for i := 1; i < numReaders; i++ {
		// Create a new reader for each goroutine
		reader, err := device.NewMultiQueueReader()
		if err != nil {
			logger.WithError(err).Errorf("Failed to create multi-queue reader %d, continuing with fewer readers", i)
			wg.Add(-1) // Reduce the wait count
			continue
		}

		go func(readerID int, reader io.ReadWriteCloser) {
			defer wg.Done()
			s.packetReaderRWC(ctx, reader, readerID)
		}(i, reader)
	}

	// Wait for all readers to finish
	wg.Wait()
}

// packetReader is a worker that reads packets from the TUN interface using overlay.Device
func (s *ClientService) packetReader(ctx context.Context, device overlay.Device, readerID int) {
	logger := logrus.WithField("reader_id", readerID)
	logger.Debug("Starting TUN packet reader")

	// Each reader has its own buffer
	buf := s.bufferPool.Get()
	defer s.bufferPool.Put(buf)

	// Check if the dispatcher supports batch processing
	batchDispatcher, supportsBatch := s.dispatcher.(types.BatchDispatcher)

	// Create an adaptive batch sizer
	adaptiveBatchSizer := NewAdaptiveBatchSizer(16, 4, 64, 500) // Initial: 16, Min: 4, Max: 64, Target: 500µs

	// Create batch processing variables if supported
	var (
		batchInterval = 5 * time.Millisecond
		packets       []*types.QueuedPacket
		connKeys      []types.ConnectionKey
		destIPs       []string
		contexts      []context.Context
	)

	if supportsBatch {
		logger.Info("Batch processing enabled with adaptive sizing")
		initialBatchSize := adaptiveBatchSizer.GetBatchSize()
		packets = make([]*types.QueuedPacket, 0, initialBatchSize)
		connKeys = make([]types.ConnectionKey, 0, initialBatchSize)
		destIPs = make([]string, 0, initialBatchSize)
		contexts = make([]context.Context, 0, initialBatchSize)
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
					// Record the start time for adaptive sizing
					startTime := time.Now()

					// Dispatch the batch
					err := batchDispatcher.DispatchPacketBatch(packets, connKeys, destIPs, contexts)

					// Record the processing time for adaptive sizing
					processingTime := time.Since(startTime)
					adaptiveBatchSizer.RecordProcessingTime(processingTime, len(packets))

					// Record network conditions
					if err == nil {
						// Success - record latency and success
						adaptiveBatchSizer.RecordNetworkLatency(processingTime)
						adaptiveBatchSizer.RecordPacketResult(true)
					} else {
						// Failure - record failure
						adaptiveBatchSizer.RecordPacketResult(false)
					}

					// Record metrics
					metrics.GlobalMetrics.RecordBatchProcessed(len(packets), processingTime)

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

		// Read a packet from the overlay device
		n, err := device.Read(buf)
		if err != nil {
			logger.WithError(err).Error("Error reading from TUN interface")
			// Add a short sleep to prevent tight loop in case of persistent errors
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Record metrics
		metrics.GlobalMetrics.RecordPacketProcessed(n)

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
			// Record dropped packet
			metrics.GlobalMetrics.RecordPacketDropped()
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
			// Get a packet from the pool and initialize it
			packet := types.GlobalPacketPool.GetWithData(ctx, destIP, packetData)

			// Add to batch
			packets = append(packets, packet)
			connKeys = append(connKeys, connKey)
			destIPs = append(destIPs, destIP)
			contexts = append(contexts, ctx)

			// If batch is full, send it
			currentBatchSize := adaptiveBatchSizer.GetBatchSize()
			if len(packets) >= currentBatchSize {
				// Record the start time for adaptive sizing
				startTime := time.Now()

				// Dispatch the batch
				err := batchDispatcher.DispatchPacketBatch(packets, connKeys, destIPs, contexts)

				// Record the processing time for adaptive sizing
				processingTime := time.Since(startTime)
				adaptiveBatchSizer.RecordProcessingTime(processingTime, len(packets))

				// Record network conditions
				if err == nil {
					// Success - record latency and success
					adaptiveBatchSizer.RecordNetworkLatency(processingTime)
					adaptiveBatchSizer.RecordPacketResult(true)
				} else {
					// Failure - record failure
					adaptiveBatchSizer.RecordPacketResult(false)
				}

				// Record metrics
				metrics.GlobalMetrics.RecordBatchProcessed(len(packets), processingTime)

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
			// Use regular dispatch for single packets with buffer tracking
			err = s.dispatcher.DispatchPacketWithFuncCallback(ctx, connKey, destIP, packetData,
				func(err error) {
					// Always return the buffer to the pool, whether success or error
					s.bufferPool.Put(packetBuf)

					if err != nil {
						logger.WithError(err).Debug("Failed to dispatch packet")
						// Record dropped packet
						metrics.GlobalMetrics.RecordPacketDropped()
					}
				})

			// If there's an immediate error (before the callback), handle it here
			if err != nil {
				logger.WithError(err).Debug("Immediate failure dispatching packet")
				// Return the buffer to the pool on immediate error
				s.bufferPool.Put(packetBuf)
				// Record dropped packet
				metrics.GlobalMetrics.RecordPacketDropped()
			}
		}
	}
}

// packetReaderRWC is a worker that reads packets from a ReadWriteCloser
// This is used for multi-queue readers from the overlay device
func (s *ClientService) packetReaderRWC(ctx context.Context, rwc io.ReadWriteCloser, readerID int) {
	logger := logrus.WithField("reader_id", readerID)
	logger.Debug("Starting TUN packet reader (RWC)")

	// Each reader has its own buffer
	buf := s.bufferPool.Get()
	defer s.bufferPool.Put(buf)

	// Check if the dispatcher supports batch processing
	batchDispatcher, supportsBatch := s.dispatcher.(types.BatchDispatcher)

	// Create an adaptive batch sizer
	adaptiveBatchSizer := NewAdaptiveBatchSizer(16, 4, 64, 500) // Initial: 16, Min: 4, Max: 64, Target: 500µs

	// Create batch processing variables if supported
	var (
		batchInterval = 5 * time.Millisecond
		packets       []*types.QueuedPacket
		connKeys      []types.ConnectionKey
		destIPs       []string
		contexts      []context.Context
	)

	if supportsBatch {
		logger.Info("Batch processing enabled with adaptive sizing")
		initialBatchSize := adaptiveBatchSizer.GetBatchSize()
		packets = make([]*types.QueuedPacket, 0, initialBatchSize)
		connKeys = make([]types.ConnectionKey, 0, initialBatchSize)
		destIPs = make([]string, 0, initialBatchSize)
		contexts = make([]context.Context, 0, initialBatchSize)
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
					// Record the start time for adaptive sizing
					startTime := time.Now()

					// Dispatch the batch
					err := batchDispatcher.DispatchPacketBatch(packets, connKeys, destIPs, contexts)

					// Record the processing time for adaptive sizing
					processingTime := time.Since(startTime)
					adaptiveBatchSizer.RecordProcessingTime(processingTime, len(packets))

					// Record network conditions
					if err == nil {
						// Success - record latency and success
						adaptiveBatchSizer.RecordNetworkLatency(processingTime)
						adaptiveBatchSizer.RecordPacketResult(true)
					} else {
						// Failure - record failure
						adaptiveBatchSizer.RecordPacketResult(false)
					}

					// Record metrics
					metrics.GlobalMetrics.RecordBatchProcessed(len(packets), processingTime)

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

		// Read a packet from the ReadWriteCloser
		n, err := rwc.Read(buf)
		if err != nil {
			logger.WithError(err).Error("Error reading from TUN interface")
			// Add a short sleep to prevent tight loop in case of persistent errors
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Record metrics
		metrics.GlobalMetrics.RecordPacketProcessed(n)

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
			// Record dropped packet
			metrics.GlobalMetrics.RecordPacketDropped()
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
			// Get a packet from the pool and initialize it
			packet := types.GlobalPacketPool.GetWithData(ctx, destIP, packetData)

			// Add to batch
			packets = append(packets, packet)
			connKeys = append(connKeys, connKey)
			destIPs = append(destIPs, destIP)
			contexts = append(contexts, ctx)

			// If batch is full, send it
			currentBatchSize := adaptiveBatchSizer.GetBatchSize()
			if len(packets) >= currentBatchSize {
				// Record the start time for adaptive sizing
				startTime := time.Now()

				// Dispatch the batch
				err := batchDispatcher.DispatchPacketBatch(packets, connKeys, destIPs, contexts)

				// Record the processing time for adaptive sizing
				processingTime := time.Since(startTime)
				adaptiveBatchSizer.RecordProcessingTime(processingTime, len(packets))

				// Record network conditions
				if err == nil {
					// Success - record latency and success
					adaptiveBatchSizer.RecordNetworkLatency(processingTime)
					adaptiveBatchSizer.RecordPacketResult(true)
				} else {
					// Failure - record failure
					adaptiveBatchSizer.RecordPacketResult(false)
				}

				// Record metrics
				metrics.GlobalMetrics.RecordBatchProcessed(len(packets), processingTime)

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
			// Use regular dispatch for single packets with buffer tracking
			err = s.dispatcher.DispatchPacketWithFuncCallback(ctx, connKey, destIP, packetData,
				func(err error) {
					// Always return the buffer to the pool, whether success or error
					s.bufferPool.Put(packetBuf)

					if err != nil {
						logger.WithError(err).Debug("Failed to dispatch packet")
						// Record dropped packet
						metrics.GlobalMetrics.RecordPacketDropped()
					}
				})

			// If there's an immediate error (before the callback), handle it here
			if err != nil {
				logger.WithError(err).Debug("Immediate failure dispatching packet")
				// Return the buffer to the pool on immediate error
				s.bufferPool.Put(packetBuf)
				// Record dropped packet
				metrics.GlobalMetrics.RecordPacketDropped()
			}
		}
	}
}
