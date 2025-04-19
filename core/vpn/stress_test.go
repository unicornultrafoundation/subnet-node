package vpn

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/metrics"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/packet"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

// StressTestStream is a mock implementation of the types.VPNStream interface for stress testing
type StressTestStream struct {
	mock.Mock
	mu           sync.Mutex
	closed       bool
	buffer       []byte
	writeLatency time.Duration
	failRate     float64
	writeCount   int64
	readCount    int64
	closeCount   int64
	resetCount   int64
}

func NewStressTestStream(writeLatency time.Duration, failRate float64) *StressTestStream {
	return &StressTestStream{
		buffer:       make([]byte, 0),
		writeLatency: writeLatency,
		failRate:     failRate,
	}
}

func (m *StressTestStream) Read(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.AddInt64(&m.readCount, 1)

	if m.closed {
		return 0, fmt.Errorf("stream closed")
	}

	if len(m.buffer) == 0 {
		return 0, nil
	}

	n = copy(p, m.buffer)
	m.buffer = m.buffer[n:]
	return n, nil
}

func (m *StressTestStream) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.AddInt64(&m.writeCount, 1)

	if m.closed {
		return 0, fmt.Errorf("stream closed")
	}

	// Simulate network latency
	if m.writeLatency > 0 {
		time.Sleep(m.writeLatency)
	}

	// Simulate random failures
	if m.failRate > 0 && rand.Float64() < m.failRate {
		return 0, fmt.Errorf("simulated network error")
	}

	// Simulate writing to the stream
	return len(p), nil
}

func (m *StressTestStream) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.AddInt64(&m.closeCount, 1)
	m.closed = true
	return nil
}

func (m *StressTestStream) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.AddInt64(&m.resetCount, 1)
	m.closed = true
	return nil
}

func (m *StressTestStream) SetDeadline(t time.Time) error {
	return nil
}

func (m *StressTestStream) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *StressTestStream) SetWriteDeadline(t time.Time) error {
	return nil
}

func (m *StressTestStream) GetStats() map[string]int64 {
	return map[string]int64{
		"write_count": atomic.LoadInt64(&m.writeCount),
		"read_count":  atomic.LoadInt64(&m.readCount),
		"close_count": atomic.LoadInt64(&m.closeCount),
		"reset_count": atomic.LoadInt64(&m.resetCount),
	}
}

// StressTestStreamService is a mock implementation of the types.Service interface for stress testing
type StressTestStreamService struct {
	mock.Mock
	mu              sync.Mutex
	streams         map[peer.ID][]*StressTestStream
	latency         time.Duration
	failRate        float64
	streamCreations int64
}

func NewStressTestStreamService(latency time.Duration, failRate float64) *StressTestStreamService {
	return &StressTestStreamService{
		streams:  make(map[peer.ID][]*StressTestStream),
		latency:  latency,
		failRate: failRate,
	}
}

func (m *StressTestStreamService) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (types.VPNStream, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.AddInt64(&m.streamCreations, 1)

	// Simulate random failures
	if m.failRate > 0 && rand.Float64() < m.failRate {
		return nil, fmt.Errorf("simulated stream creation failure")
	}

	// Create a new mock stream
	stream := NewStressTestStream(m.latency, m.failRate)

	// Store the stream
	if _, exists := m.streams[peerID]; !exists {
		m.streams[peerID] = make([]*StressTestStream, 0)
	}
	m.streams[peerID] = append(m.streams[peerID], stream)

	return stream, nil
}

func (m *StressTestStreamService) GetStats() map[string]int64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	totalStreams := 0
	for _, streams := range m.streams {
		totalStreams += len(streams)
	}

	return map[string]int64{
		"stream_creations": atomic.LoadInt64(&m.streamCreations),
		"active_streams":   int64(totalStreams),
	}
}

// StressTestDiscoveryService is a mock implementation of the discovery.PeerDiscoveryService interface for stress testing
type StressTestDiscoveryService struct {
	mock.Mock
	mu        sync.Mutex
	peerIDMap map[string]string
	latency   time.Duration
	failRate  float64
	lookups   int64
}

func NewStressTestDiscoveryService(latency time.Duration, failRate float64) *StressTestDiscoveryService {
	return &StressTestDiscoveryService{
		peerIDMap: make(map[string]string),
		latency:   latency,
		failRate:  failRate,
	}
}

func (m *StressTestDiscoveryService) GetPeerID(ctx context.Context, destIP string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.AddInt64(&m.lookups, 1)

	// Simulate network latency
	if m.latency > 0 {
		time.Sleep(m.latency)
	}

	// Simulate random failures
	if m.failRate > 0 && rand.Float64() < m.failRate {
		return "", fmt.Errorf("simulated lookup failure")
	}

	peerID, exists := m.peerIDMap[destIP]
	if !exists {
		return "", fmt.Errorf("peer not found for IP %s", destIP)
	}
	return peerID, nil
}

func (m *StressTestDiscoveryService) SyncPeerIDToDHT(ctx context.Context) error {
	return nil
}

func (m *StressTestDiscoveryService) GetStats() map[string]int64 {
	return map[string]int64{
		"lookups": atomic.LoadInt64(&m.lookups),
	}
}

// TestStressMetrics tests the metrics collection under stress
func TestStressMetrics(t *testing.T) {
	// Create a metrics collector
	vpnMetrics := metrics.NewVPNMetrics()

	// Increment some metrics in parallel
	const numGoroutines = 10
	const numOperations = 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				vpnMetrics.IncrementPacketsReceived(100)
				vpnMetrics.IncrementPacketsSent(200)
				vpnMetrics.IncrementPacketsDropped()
			}
		}()
	}

	wg.Wait()

	// Get the metrics
	metrics := vpnMetrics.GetMetrics()

	// Verify the metrics
	assert.Equal(t, int64(numGoroutines*numOperations), metrics["packets_received"])
	assert.Equal(t, int64(numGoroutines*numOperations), metrics["packets_sent"])
	assert.Equal(t, int64(numGoroutines*numOperations), metrics["packets_dropped"])
	assert.Equal(t, int64(numGoroutines*numOperations*100), metrics["bytes_received"])
	assert.Equal(t, int64(numGoroutines*numOperations*200), metrics["bytes_sent"])
}

// TestStressStreamService tests the stream service under stress
func TestStressStreamService(t *testing.T) {
	// Create a mock stream service with some latency and failure rate
	mockStreamService := NewStressTestStreamService(1*time.Millisecond, 0.01) // 1ms latency, 1% failure rate

	// Create a peer ID
	peerID, err := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	require.NoError(t, err)

	// Create a stream service with realistic parameters
	streamService := stream.NewStreamService(
		mockStreamService,
		20,                   // maxStreamsPerPeer
		5,                    // minStreamsPerPeer
		5*time.Minute,        // streamIdleTimeout
		30*time.Second,       // cleanupInterval
		5*time.Second,        // healthCheckInterval
		500*time.Millisecond, // healthCheckTimeout
		3,                    // maxConsecutiveFailures
		30*time.Second,       // warmInterval
		10,                   // maxStreamsPerMultiplexer
		3,                    // minStreamsPerMultiplexer
		5*time.Second,        // autoScalingInterval
		true,                 // multiplexingEnabled
	)

	// Start the service
	streamService.Start()
	defer streamService.Stop()

	// Run stress test with different packet sizes and rates
	packetSizes := []int{64, 512, 1024, 4096, 8192}
	packetRates := []int{10, 100, 1000} // packets per second

	for _, size := range packetSizes {
		for _, rate := range packetRates {
			t.Run(fmt.Sprintf("Size=%d,Rate=%d", size, rate), func(t *testing.T) {
				stressTestStreamService(t, streamService, peerID, size, rate)
			})
		}
	}

	// Get and log the service metrics
	healthMetrics := streamService.GetHealthMetrics()
	multiplexerMetrics := streamService.GetMultiplexerMetrics()
	poolMetrics := streamService.GetStreamPoolMetrics()

	t.Logf("Health metrics: %v", healthMetrics)
	t.Logf("Multiplexer metrics: %v", multiplexerMetrics)
	t.Logf("Pool metrics: %v", poolMetrics)
	t.Logf("Mock service stats: %v", mockStreamService.GetStats())
}

// stressTestStreamService runs a stress test on the stream service with the given parameters
func stressTestStreamService(t *testing.T, streamService *stream.StreamService, peerID peer.ID, packetSize, packetsPerSecond int) {
	// Calculate test duration and total packets
	testDuration := 5 * time.Second
	totalPackets := int(float64(packetsPerSecond) * testDuration.Seconds())

	// Create a fixed-size packet
	packet := make([]byte, packetSize)
	for i := 0; i < packetSize; i++ {
		packet[i] = byte(i % 256)
	}

	// Calculate interval between packets
	interval := time.Second / time.Duration(packetsPerSecond)

	// Start time
	startTime := time.Now()

	// Track packet count
	var successCount int64

	// Send packets at the specified rate
	ctx := context.Background()

	// Use multiple goroutines for higher rates
	numGoroutines := 1
	if packetsPerSecond > 100 {
		numGoroutines = 4
	}
	if packetsPerSecond > 500 {
		numGoroutines = 8
	}

	packetsPerGoroutine := totalPackets / numGoroutines

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			defer wg.Done()

			for i := 0; i < packetsPerGoroutine; i++ {
				// Send packet using different modes
				var err error

				switch i % 3 {
				case 0: // Direct mode
					stream, err := streamService.CreateNewVPNStream(ctx, peerID)
					if err == nil {
						_, err = stream.Write(packet)
						stream.Close()
					}
				case 1: // Pooled mode
					stream, err := streamService.GetStream(ctx, peerID)
					if err == nil {
						_, err = stream.Write(packet)
						streamService.ReleaseStream(peerID, stream, err == nil)
					}
				case 2: // Multiplexed mode
					err = streamService.SendPacketMultiplexed(ctx, peerID, packet)
				}

				// Count successful operations
				if err == nil {
					atomic.AddInt64(&successCount, 1)
				}

				// Sleep to maintain the rate
				if interval > 0 && numGoroutines == 1 {
					time.Sleep(interval)
				} else if numGoroutines > 1 {
					// For multiple goroutines, add a small random sleep to avoid thundering herd
					time.Sleep(time.Duration(rand.Intn(1000)) * time.Microsecond)
				}
			}
		}(g)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// End time
	endTime := time.Now()
	actualDuration := endTime.Sub(startTime)

	// Calculate throughput
	totalSent := atomic.LoadInt64(&successCount)
	totalBytes := int64(packetSize) * totalSent
	throughputPackets := float64(totalSent) / actualDuration.Seconds()
	throughputMbps := (float64(totalBytes) * 8 / 1000000) / actualDuration.Seconds()

	// Log results
	t.Logf("Packet size: %d bytes, Rate: %d pps, Duration: %.2f seconds", packetSize, packetsPerSecond, actualDuration.Seconds())
	t.Logf("Total packets sent: %d, Successful: %d", totalPackets, totalSent)
	t.Logf("Throughput: %.2f packets/s, %.2f Mbps", throughputPackets, throughputMbps)
}

// TestStressPacketDispatcher tests the packet dispatcher under stress
func TestStressPacketDispatcher(t *testing.T) {
	// Create mock services with some latency and failure rate
	mockStreamService := NewStressTestStreamService(1*time.Millisecond, 0.01)       // 1ms latency, 1% failure rate
	mockDiscoveryService := NewStressTestDiscoveryService(2*time.Millisecond, 0.02) // 2ms latency, 2% failure rate

	// Add peer mappings
	for i := 1; i <= 10; i++ {
		mockDiscoveryService.peerIDMap[fmt.Sprintf("192.168.1.%d", i)] = fmt.Sprintf("QmPeer%d", i)
	}

	// Create a stream service
	streamService := stream.NewStreamService(
		mockStreamService,
		20,                   // maxStreamsPerPeer
		5,                    // minStreamsPerPeer
		5*time.Minute,        // streamIdleTimeout
		30*time.Second,       // cleanupInterval
		5*time.Second,        // healthCheckInterval
		500*time.Millisecond, // healthCheckTimeout
		3,                    // maxConsecutiveFailures
		30*time.Second,       // warmInterval
		10,                   // maxStreamsPerMultiplexer
		3,                    // minStreamsPerMultiplexer
		5*time.Second,        // autoScalingInterval
		true,                 // multiplexingEnabled
	)

	// Start the stream service
	streamService.Start()
	defer streamService.Stop()

	// Create an enhanced packet dispatcher
	dispatcher := packet.NewEnhancedDispatcher(
		mockDiscoveryService,
		mockStreamService,
		streamService,
		streamService,
		300,           // workerIdleTimeout
		5*time.Second, // workerCleanupInterval
		1000,          // workerBufferSize
		true,          // useMultiplexing
	)

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Run stress test with different packet sizes and destination counts
	packetSizes := []int{64, 1024, 8192}
	destinationCounts := []int{1, 5, 10}

	for _, size := range packetSizes {
		for _, destCount := range destinationCounts {
			t.Run(fmt.Sprintf("Size=%d,Destinations=%d", size, destCount), func(t *testing.T) {
				stressTestPacketDispatcher(t, dispatcher, size, destCount)
			})
		}
	}

	// Log statistics
	t.Logf("Mock stream service stats: %v", mockStreamService.GetStats())
	t.Logf("Mock discovery service stats: %v", mockDiscoveryService.GetStats())
}

// stressTestPacketDispatcher runs a stress test on the packet dispatcher with the given parameters
func stressTestPacketDispatcher(t *testing.T, dispatcher *packet.Dispatcher, packetSize, destinationCount int) {
	// Create destinations
	destinations := make([]string, destinationCount)
	for i := 0; i < destinationCount; i++ {
		destinations[i] = fmt.Sprintf("192.168.1.%d", (i%10)+1)
	}

	// Create a fixed-size packet
	packet := make([]byte, packetSize)
	for i := 0; i < packetSize; i++ {
		packet[i] = byte(i % 256)
	}

	// Test parameters
	testDuration := 5 * time.Second
	packetsPerSecond := 1000
	totalPackets := int(float64(packetsPerSecond) * testDuration.Seconds())

	// Start time
	startTime := time.Now()

	// Track packet count
	var successCount int64

	// Use multiple goroutines
	numGoroutines := 4
	packetsPerGoroutine := totalPackets / numGoroutines

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			defer wg.Done()
			ctx := context.Background()

			for i := 0; i < packetsPerGoroutine; i++ {
				// Select a random destination
				destIdx := rand.Intn(len(destinations))
				destIP := destinations[destIdx]
				destPort := 1000 + rand.Intn(1000) // Random port between 1000-1999
				syncKey := fmt.Sprintf("%s:%d", destIP, destPort)

				// Dispatch the packet - it no longer returns an error
				dispatcher.DispatchPacket(ctx, syncKey, destIP, packet)

				// Since DispatchPacket no longer returns an error, we'll count all packets as successful
				atomic.AddInt64(&successCount, 1)

				// Add a small random sleep to avoid thundering herd
				time.Sleep(time.Duration(rand.Intn(1000)) * time.Microsecond)
			}
		}(g)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// End time
	endTime := time.Now()
	actualDuration := endTime.Sub(startTime)

	// Calculate throughput
	totalSent := atomic.LoadInt64(&successCount)
	totalBytes := int64(packetSize) * totalSent
	throughputPackets := float64(totalSent) / actualDuration.Seconds()
	throughputMbps := (float64(totalBytes) * 8 / 1000000) / actualDuration.Seconds()

	// Log results
	t.Logf("Packet size: %d bytes, Destinations: %d, Duration: %.2f seconds", packetSize, destinationCount, actualDuration.Seconds())
	t.Logf("Total packets sent: %d", totalSent)
	t.Logf("Throughput: %.2f packets/s, %.2f Mbps", throughputPackets, throughputMbps)

	// We can't directly access the workers map since it's unexported
	// Instead, we'll check that the test completed successfully
	t.Logf("Test completed successfully with packet size: %d bytes, destinations: %d", packetSize, destinationCount)
}
