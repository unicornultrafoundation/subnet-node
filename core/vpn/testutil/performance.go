package testutil

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/unicornultrafoundation/subnet-node/core/vpn/packet"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream"
)

// PerformanceResult contains the results of a performance test
type PerformanceResult struct {
	TotalPackets      int64
	SuccessfulPackets int64
	FailedPackets     int64
	TotalBytes        int64
	Duration          time.Duration
	PacketsPerSecond  float64
	BytesPerSecond    float64
	AverageLatency    time.Duration
	P50Latency        time.Duration
	P90Latency        time.Duration
	P99Latency        time.Duration
}

// PerformanceConfig contains the configuration for a performance test
type PerformanceConfig struct {
	PacketSize       int
	PacketsPerSecond int
	Duration         time.Duration
	Concurrency      int
	WarmupDuration   time.Duration
	CooldownDuration time.Duration
}

// DefaultPerformanceConfig returns a default performance test configuration
func DefaultPerformanceConfig() *PerformanceConfig {
	return &PerformanceConfig{
		PacketSize:       1024,
		PacketsPerSecond: 100,
		Duration:         5 * time.Second,
		Concurrency:      1,
		WarmupDuration:   1 * time.Second,
		CooldownDuration: 1 * time.Second,
	}
}

// RunStreamPerformanceTest runs a performance test on a stream service
func RunStreamPerformanceTest(t *testing.T, streamService *stream.StreamService, peerID peer.ID, config *PerformanceConfig) *PerformanceResult {
	if config == nil {
		config = DefaultPerformanceConfig()
	}

	// Create a fixed-size packet
	packet := CreateTestPacket(config.PacketSize)

	// Calculate total packets to send
	totalPackets := int(float64(config.PacketsPerSecond) * config.Duration.Seconds())
	packetsPerGoroutine := totalPackets / config.Concurrency

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.Duration+config.WarmupDuration+config.CooldownDuration)
	defer cancel()

	// Create a wait group for goroutines
	var wg sync.WaitGroup
	wg.Add(config.Concurrency)

	// Create atomic counters for metrics
	var successCount int64
	var failureCount int64
	var totalBytes int64
	var totalLatency int64
	var latencies []time.Duration

	// Create a mutex for latencies
	latenciesMutex := &sync.Mutex{}

	// Calculate the interval between packets
	interval := time.Duration(1000000000/config.PacketsPerSecond) * time.Nanosecond

	// Warm up
	time.Sleep(config.WarmupDuration)

	// Start time
	startTime := time.Now()

	// Start goroutines
	for i := 0; i < config.Concurrency; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < packetsPerGoroutine; j++ {
				// Check if context is done
				select {
				case <-ctx.Done():
					return
				default:
					// Continue
				}

				// Get a stream from the pool
				start := time.Now()
				stream, err := streamService.GetStream(ctx, peerID)
				if err == nil {
					// Write to the stream
					n, err := stream.Write(packet)
					// Release the stream back to the pool
					streamService.ReleaseStream(peerID, stream, err == nil)

					// Record metrics
					if err == nil {
						atomic.AddInt64(&successCount, 1)
						atomic.AddInt64(&totalBytes, int64(n))

						// Record latency
						latency := time.Since(start)
						atomic.AddInt64(&totalLatency, int64(latency))

						latenciesMutex.Lock()
						latencies = append(latencies, latency)
						latenciesMutex.Unlock()
					} else {
						atomic.AddInt64(&failureCount, 1)
					}
				} else {
					atomic.AddInt64(&failureCount, 1)
				}

				// Sleep to maintain the desired rate
				time.Sleep(interval)
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// End time
	duration := time.Since(startTime)

	// Cool down
	time.Sleep(config.CooldownDuration)

	// Calculate metrics
	totalPacketsSent := atomic.LoadInt64(&successCount) + atomic.LoadInt64(&failureCount)
	packetsPerSecond := float64(totalPacketsSent) / duration.Seconds()
	bytesPerSecond := float64(atomic.LoadInt64(&totalBytes)) / duration.Seconds()

	// Calculate latency percentiles
	var p50Latency, p90Latency, p99Latency time.Duration
	latenciesMutex.Lock()
	if len(latencies) > 0 {
		// Sort latencies
		sortLatencies(latencies)

		// Calculate percentiles
		p50Latency = latencies[len(latencies)*50/100]
		p90Latency = latencies[len(latencies)*90/100]
		p99Latency = latencies[len(latencies)*99/100]
	}
	latenciesMutex.Unlock()

	// Create result
	result := &PerformanceResult{
		TotalPackets:      totalPacketsSent,
		SuccessfulPackets: atomic.LoadInt64(&successCount),
		FailedPackets:     atomic.LoadInt64(&failureCount),
		TotalBytes:        atomic.LoadInt64(&totalBytes),
		Duration:          duration,
		PacketsPerSecond:  packetsPerSecond,
		BytesPerSecond:    bytesPerSecond,
		AverageLatency:    time.Duration(atomic.LoadInt64(&totalLatency) / totalPacketsSent),
		P50Latency:        p50Latency,
		P90Latency:        p90Latency,
		P99Latency:        p99Latency,
	}

	// Log results
	t.Logf("Performance test results:")
	t.Logf("  Total packets: %d", result.TotalPackets)
	t.Logf("  Successful packets: %d (%.2f%%)", result.SuccessfulPackets, float64(result.SuccessfulPackets)/float64(result.TotalPackets)*100)
	t.Logf("  Failed packets: %d (%.2f%%)", result.FailedPackets, float64(result.FailedPackets)/float64(result.TotalPackets)*100)
	t.Logf("  Total bytes: %d", result.TotalBytes)
	t.Logf("  Duration: %v", result.Duration)
	t.Logf("  Packets per second: %.2f", result.PacketsPerSecond)
	t.Logf("  Bytes per second: %.2f (%.2f MB/s)", result.BytesPerSecond, result.BytesPerSecond/1024/1024)
	t.Logf("  Average latency: %v", result.AverageLatency)
	t.Logf("  P50 latency: %v", result.P50Latency)
	t.Logf("  P90 latency: %v", result.P90Latency)
	t.Logf("  P99 latency: %v", result.P99Latency)

	return result
}

// RunDispatcherPerformanceTest runs a performance test on a packet dispatcher
func RunDispatcherPerformanceTest(t *testing.T, dispatcher *packet.Dispatcher, destIP string, config *PerformanceConfig) *PerformanceResult {
	if config == nil {
		config = DefaultPerformanceConfig()
	}

	// Create a fixed-size packet
	packet := CreateTestPacket(config.PacketSize)

	// Calculate total packets to send
	totalPackets := int(float64(config.PacketsPerSecond) * config.Duration.Seconds())
	packetsPerGoroutine := totalPackets / config.Concurrency

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.Duration+config.WarmupDuration+config.CooldownDuration)
	defer cancel()

	// Create a wait group for goroutines
	var wg sync.WaitGroup
	wg.Add(config.Concurrency)

	// Create atomic counters for metrics
	var successCount int64
	var failureCount int64
	var totalBytes int64
	var totalLatency int64
	var latencies []time.Duration

	// Create a mutex for latencies
	latenciesMutex := &sync.Mutex{}

	// Calculate the interval between packets
	interval := time.Duration(1000000000/config.PacketsPerSecond) * time.Nanosecond

	// Create sync keys for each goroutine
	syncKeys := make([]string, config.Concurrency)
	for i := 0; i < config.Concurrency; i++ {
		syncKeys[i] = fmt.Sprintf("%s:%d", destIP, 1000+i)
	}

	// Warm up
	time.Sleep(config.WarmupDuration)

	// Start time
	startTime := time.Now()

	// Start goroutines
	for i := 0; i < config.Concurrency; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			syncKey := syncKeys[goroutineID]

			for j := 0; j < packetsPerGoroutine; j++ {
				// Check if context is done
				select {
				case <-ctx.Done():
					return
				default:
					// Continue
				}

				// Create a done channel to get the result
				doneCh := make(chan error, 1)

				// Dispatch the packet with a callback
				start := time.Now()
				dispatcher.DispatchPacketWithCallback(ctx, syncKey, destIP, packet, doneCh)

				// Wait for the result or timeout
				select {
				case err := <-doneCh:
					// Record metrics
					if err == nil {
						atomic.AddInt64(&successCount, 1)
						atomic.AddInt64(&totalBytes, int64(len(packet)))

						// Record latency
						latency := time.Since(start)
						atomic.AddInt64(&totalLatency, int64(latency))

						latenciesMutex.Lock()
						latencies = append(latencies, latency)
						latenciesMutex.Unlock()
					} else {
						atomic.AddInt64(&failureCount, 1)
					}
				case <-time.After(100 * time.Millisecond):
					// Timeout - packet was not processed in time
					atomic.AddInt64(&failureCount, 1)
				}

				// Sleep to maintain the desired rate
				time.Sleep(interval)
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// End time
	duration := time.Since(startTime)

	// Cool down
	time.Sleep(config.CooldownDuration)

	// Calculate metrics
	totalPacketsSent := atomic.LoadInt64(&successCount) + atomic.LoadInt64(&failureCount)
	packetsPerSecond := float64(totalPacketsSent) / duration.Seconds()
	bytesPerSecond := float64(atomic.LoadInt64(&totalBytes)) / duration.Seconds()

	// Calculate latency percentiles
	var p50Latency, p90Latency, p99Latency time.Duration
	latenciesMutex.Lock()
	if len(latencies) > 0 {
		// Sort latencies
		sortLatencies(latencies)

		// Calculate percentiles
		p50Latency = latencies[len(latencies)*50/100]
		p90Latency = latencies[len(latencies)*90/100]
		p99Latency = latencies[len(latencies)*99/100]
	}
	latenciesMutex.Unlock()

	// Create result
	result := &PerformanceResult{
		TotalPackets:      totalPacketsSent,
		SuccessfulPackets: atomic.LoadInt64(&successCount),
		FailedPackets:     atomic.LoadInt64(&failureCount),
		TotalBytes:        atomic.LoadInt64(&totalBytes),
		Duration:          duration,
		PacketsPerSecond:  packetsPerSecond,
		BytesPerSecond:    bytesPerSecond,
		AverageLatency:    time.Duration(atomic.LoadInt64(&totalLatency) / totalPacketsSent),
		P50Latency:        p50Latency,
		P90Latency:        p90Latency,
		P99Latency:        p99Latency,
	}

	// Log results
	t.Logf("Performance test results:")
	t.Logf("  Total packets: %d", result.TotalPackets)
	t.Logf("  Successful packets: %d (%.2f%%)", result.SuccessfulPackets, float64(result.SuccessfulPackets)/float64(result.TotalPackets)*100)
	t.Logf("  Failed packets: %d (%.2f%%)", result.FailedPackets, float64(result.FailedPackets)/float64(result.TotalPackets)*100)
	t.Logf("  Total bytes: %d", result.TotalBytes)
	t.Logf("  Duration: %v", result.Duration)
	t.Logf("  Packets per second: %.2f", result.PacketsPerSecond)
	t.Logf("  Bytes per second: %.2f (%.2f MB/s)", result.BytesPerSecond, result.BytesPerSecond/1024/1024)
	t.Logf("  Average latency: %v", result.AverageLatency)
	t.Logf("  P50 latency: %v", result.P50Latency)
	t.Logf("  P90 latency: %v", result.P90Latency)
	t.Logf("  P99 latency: %v", result.P99Latency)

	return result
}

// VerifyPerformanceRequirements logs performance metrics for informational purposes
func VerifyPerformanceRequirements(t *testing.T, result *PerformanceResult, minPacketsPerSecond float64, maxLatency time.Duration) {
	t.Logf("Performance metrics: %.2f packets/sec (target: %.2f), %v P90 latency (target: %v)",
		result.PacketsPerSecond, minPacketsPerSecond, result.P90Latency, maxLatency)
}

// Helper function to sort latencies
func sortLatencies(latencies []time.Duration) {
	for i := 0; i < len(latencies); i++ {
		for j := i + 1; j < len(latencies); j++ {
			if latencies[i] > latencies[j] {
				latencies[i], latencies[j] = latencies[j], latencies[i]
			}
		}
	}
}
