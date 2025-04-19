package multiplex

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

var log = logrus.WithField("service", "vpn-stream-multiplex")

// StreamMultiplexer multiplexes packets over multiple streams
type StreamMultiplexer struct {
	// Peer ID for this multiplexer
	peerID peer.ID
	// Streams for this multiplexer
	streams []types.VPNStream
	// Mutex to protect access to streams
	mu sync.RWMutex
	// Stream service for creating new streams
	streamService types.Service
	// Stream pool service for managing streams
	streamPoolService types.PoolService
	// Maximum streams per multiplexer
	maxStreamsPerMultiplexer int
	// Minimum streams per multiplexer
	minStreamsPerMultiplexer int
	// Current stream index for round-robin
	currentStreamIndex int32
	// Metrics for this multiplexer
	metrics *types.MultiplexerMetrics
}

// NewStreamMultiplexer creates a new stream multiplexer
func NewStreamMultiplexer(
	peerID peer.ID,
	streamService types.Service,
	streamPoolService types.PoolService,
	maxStreamsPerMultiplexer int,
	minStreamsPerMultiplexer int,
) *StreamMultiplexer {
	return &StreamMultiplexer{
		peerID:                   peerID,
		streams:                  make([]types.VPNStream, 0, maxStreamsPerMultiplexer),
		streamService:            streamService,
		streamPoolService:        streamPoolService,
		maxStreamsPerMultiplexer: maxStreamsPerMultiplexer,
		minStreamsPerMultiplexer: minStreamsPerMultiplexer,
		currentStreamIndex:       0,
		metrics:                  &types.MultiplexerMetrics{},
	}
}

// Initialize initializes the multiplexer with streams
func (m *StreamMultiplexer) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create initial streams
	var createdStreams int
	var lastError error

	// Try to create the minimum number of streams
	for i := 0; i < m.minStreamsPerMultiplexer; i++ {
		// Create a new stream
		s, err := m.streamService.CreateNewVPNStream(ctx, m.peerID)
		if err != nil {
			// Update metrics
			atomic.AddInt64(&m.metrics.StreamErrors, 1)

			// Save the error
			lastError = err

			// Log the error
			log.Warnf("Failed to create stream %d for peer %s: %v",
				i, m.peerID.String(), err)

			// Continue trying to create more streams
			continue
		}

		// Add the stream to the list
		m.streams = append(m.streams, s)

		// Update metrics
		atomic.AddInt64(&m.metrics.StreamsCreated, 1)

		// Increment the counter
		createdStreams++
	}

	// Check if we created any streams
	if createdStreams == 0 {
		return fmt.Errorf("failed to create any streams: %v", lastError)
	}

	// Check if we created fewer than the minimum
	if createdStreams < m.minStreamsPerMultiplexer {
		log.Warnf("Created fewer streams (%d) than minimum (%d) for peer %s",
			createdStreams, m.minStreamsPerMultiplexer, m.peerID.String())
	}

	return nil
}

// SendPacket sends a packet using the multiplexer
func (m *StreamMultiplexer) SendPacket(ctx context.Context, packet []byte) error {
	// Get a stream using round-robin
	stream, err := m.getNextStream()
	if err != nil {
		// Update metrics
		atomic.AddInt64(&m.metrics.PacketsDropped, 1)

		// Try to recover by creating a new stream
		err = m.recoverStreams(ctx)
		if err != nil {
			return fmt.Errorf("no streams available and failed to recover: %v", err)
		}

		// Try again with the new stream
		stream, err = m.getNextStream()
		if err != nil {
			// Still no streams available
			atomic.AddInt64(&m.metrics.PacketsDropped, 1)
			return fmt.Errorf("failed to get stream after recovery: %v", err)
		}
	}

	// Record the start time for latency measurement
	startTime := time.Now()

	// Write the packet to the stream
	_, err = stream.Write(packet)
	if err != nil {
		// Update metrics
		atomic.AddInt64(&m.metrics.StreamErrors, 1)
		atomic.AddInt64(&m.metrics.PacketsDropped, 1)

		// Try to replace the failed stream
		m.mu.Lock()
		defer m.mu.Unlock()

		// Find the stream in the list
		for i, s := range m.streams {
			if s == stream {
				// Close the stream
				s.Close()

				// Remove the stream from the list
				m.streams = append(m.streams[:i], m.streams[i+1:]...)

				// Update metrics
				atomic.AddInt64(&m.metrics.StreamsClosed, 1)

				// Try to create a new stream
				newStream, err := m.streamService.CreateNewVPNStream(ctx, m.peerID)
				if err != nil {
					// Update metrics
					atomic.AddInt64(&m.metrics.StreamErrors, 1)

					// If we're below the minimum, try to recover
					if len(m.streams) < m.minStreamsPerMultiplexer {
						log.Warnf("Stream count (%d) below minimum (%d), attempting recovery",
							len(m.streams), m.minStreamsPerMultiplexer)

						// Try to create at least the minimum number of streams
						for len(m.streams) < m.minStreamsPerMultiplexer {
							recoveryStream, recoveryErr := m.streamService.CreateNewVPNStream(ctx, m.peerID)
							if recoveryErr != nil {
								// Failed to recover, but we'll continue with what we have
								log.Errorf("Failed to recover streams: %v", recoveryErr)
								break
							}

							// Add the new stream
							m.streams = append(m.streams, recoveryStream)
							atomic.AddInt64(&m.metrics.StreamsCreated, 1)
						}

						// If we still don't have any streams, return an error
						if len(m.streams) == 0 {
							return fmt.Errorf("failed to create any streams: %v", err)
						}

						// Otherwise, continue with what we have
						return fmt.Errorf("failed to send packet, but recovered with %d streams: %v",
							len(m.streams), err)
					}

					// Otherwise, continue with fewer streams
					return fmt.Errorf("failed to send packet: %v", err)
				}

				// Add the new stream to the list
				m.streams = append(m.streams, newStream)

				// Update metrics
				atomic.AddInt64(&m.metrics.StreamsCreated, 1)

				break
			}
		}

		return fmt.Errorf("failed to send packet: %v", err)
	}

	// Record the end time for latency measurement
	endTime := time.Now()
	latency := endTime.Sub(startTime).Milliseconds()

	// Update metrics
	atomic.AddInt64(&m.metrics.PacketsSent, 1)
	atomic.AddInt64(&m.metrics.BytesSent, int64(len(packet)))

	// Update latency metrics
	oldAvg := atomic.LoadInt64(&m.metrics.AvgLatency)
	oldCount := atomic.LoadInt64(&m.metrics.LatencyMeasurements)
	newCount := oldCount + 1
	newAvg := (oldAvg*oldCount + latency) / newCount
	atomic.StoreInt64(&m.metrics.AvgLatency, newAvg)
	atomic.StoreInt64(&m.metrics.LatencyMeasurements, newCount)

	return nil
}

// getNextStream gets the next stream using round-robin
func (m *StreamMultiplexer) getNextStream() (types.VPNStream, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if we have any streams
	if len(m.streams) == 0 {
		return nil, fmt.Errorf("no streams available")
	}

	// Get the current index
	index := atomic.AddInt32(&m.currentStreamIndex, 1) % int32(len(m.streams))

	// Get the stream
	return m.streams[index], nil
}

// ScaleUp adds a new stream to the multiplexer
func (m *StreamMultiplexer) ScaleUp(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if we've reached the maximum
	if len(m.streams) >= m.maxStreamsPerMultiplexer {
		return fmt.Errorf("maximum number of streams reached")
	}

	// Create a new stream
	s, err := m.streamService.CreateNewVPNStream(ctx, m.peerID)
	if err != nil {
		// Update metrics
		atomic.AddInt64(&m.metrics.StreamErrors, 1)

		return fmt.Errorf("failed to create new stream: %v", err)
	}

	// Add the stream to the list
	m.streams = append(m.streams, s)

	// Update metrics
	atomic.AddInt64(&m.metrics.StreamsCreated, 1)
	atomic.AddInt64(&m.metrics.ScaleUpOperations, 1)

	return nil
}

// ScaleDown removes a stream from the multiplexer
func (m *StreamMultiplexer) ScaleDown() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if we've reached the minimum
	if len(m.streams) <= m.minStreamsPerMultiplexer {
		return fmt.Errorf("minimum number of streams reached")
	}

	// Get the last stream
	lastIndex := len(m.streams) - 1
	lastStream := m.streams[lastIndex]

	// Close the stream
	lastStream.Close()

	// Remove the stream from the list
	m.streams = m.streams[:lastIndex]

	// Update metrics
	atomic.AddInt64(&m.metrics.StreamsClosed, 1)
	atomic.AddInt64(&m.metrics.ScaleDownOperations, 1)

	return nil
}

// GetStreamCount returns the number of streams in the multiplexer
func (m *StreamMultiplexer) GetStreamCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.streams)
}

// GetMetrics returns the metrics for this multiplexer
func (m *StreamMultiplexer) GetMetrics() *types.MultiplexerMetrics {
	return m.metrics
}

// recoverStreams attempts to recover by creating new streams
func (m *StreamMultiplexer) recoverStreams(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If we already have streams, nothing to do
	if len(m.streams) > 0 {
		return nil
	}

	// Log the recovery attempt
	log.Warnf("Attempting to recover multiplexer for peer %s", m.peerID.String())

	// Try to create at least the minimum number of streams
	for i := 0; i < m.minStreamsPerMultiplexer; i++ {
		newStream, err := m.streamService.CreateNewVPNStream(ctx, m.peerID)
		if err != nil {
			// If we've created at least one stream, we can continue
			if len(m.streams) > 0 {
				log.Warnf("Partial recovery: created %d/%d streams", len(m.streams), m.minStreamsPerMultiplexer)
				return nil
			}

			// Otherwise, return the error
			return fmt.Errorf("failed to create any streams during recovery: %v", err)
		}

		// Add the new stream to the list
		m.streams = append(m.streams, newStream)

		// Update metrics
		atomic.AddInt64(&m.metrics.StreamsCreated, 1)
	}

	log.Infof("Successfully recovered multiplexer for peer %s with %d streams",
		m.peerID.String(), len(m.streams))

	return nil
}

// Close closes all streams in the multiplexer
func (m *StreamMultiplexer) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Close all streams
	for _, s := range m.streams {
		s.Close()

		// Update metrics
		atomic.AddInt64(&m.metrics.StreamsClosed, 1)
	}

	// Clear the streams list
	m.streams = make([]types.VPNStream, 0)
}
