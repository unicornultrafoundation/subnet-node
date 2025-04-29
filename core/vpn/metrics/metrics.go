package metrics

import (
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/utils"
)

// VPNMetrics collects and reports metrics for the VPN system
type VPNMetrics struct {
	// Packet processing metrics
	packetsProcessed int64
	packetsDropped   int64
	bytesProcessed   int64

	// Batch processing metrics
	batchesProcessed int64
	avgBatchSize     int64
	batchSizeCount   int64
	batchSizeTotal   int64

	// Timing metrics
	processingTimeTotal int64
	processingTimeCount int64

	// Buffer pool metrics
	bufferPoolGets    int64
	bufferPoolPuts    int64
	bufferPoolMisses  int64
	bufferPoolCreates int64

	// Packet pool metrics
	packetPoolGets    int64
	packetPoolPuts    int64
	packetPoolCreates int64

	// Lock-free queue metrics
	queueEnqueues int64
	queueDequeues int64
	queueSize     int64

	// Network metrics
	networkLatencyTotal int64
	networkLatencyCount int64
	networkLatencyMin   int64
	networkLatencyMax   int64
	networkJitterTotal  int64
	networkJitterCount  int64
	networkSuccessCount int64
	networkFailureCount int64

	// Stream metrics
	activeStreams       int64
	totalStreamsCreated int64
	totalStreamsRemoved int64
	activeConnections   int64

	// Last report time
	lastReportTime time.Time

	// Report interval
	reportInterval time.Duration

	// Logger
	logger *logrus.Entry
}

// NewVPNMetrics creates a new VPN metrics collector
func NewVPNMetrics() *VPNMetrics {
	return &VPNMetrics{
		lastReportTime: time.Now(),
		reportInterval: time.Minute,
		logger:         logrus.WithField("component", "vpn_metrics"),
	}
}

// RecordPacketProcessed records a processed packet
func (m *VPNMetrics) RecordPacketProcessed(size int) {
	atomic.AddInt64(&m.packetsProcessed, 1)
	atomic.AddInt64(&m.bytesProcessed, int64(size))
	m.checkAndReport()
}

// RecordPacketDropped records a dropped packet
func (m *VPNMetrics) RecordPacketDropped() {
	atomic.AddInt64(&m.packetsDropped, 1)
	m.checkAndReport()
}

// RecordBatchProcessed records a processed batch
func (m *VPNMetrics) RecordBatchProcessed(batchSize int, processingTime time.Duration) {
	atomic.AddInt64(&m.batchesProcessed, 1)
	atomic.AddInt64(&m.batchSizeTotal, int64(batchSize))
	atomic.AddInt64(&m.batchSizeCount, 1)
	atomic.AddInt64(&m.processingTimeTotal, processingTime.Microseconds())
	atomic.AddInt64(&m.processingTimeCount, 1)
	m.checkAndReport()
}

// RecordBufferPoolStats records buffer pool statistics
func (m *VPNMetrics) RecordBufferPoolStats(stats map[string]int64) {
	atomic.StoreInt64(&m.bufferPoolGets, stats["gets"])
	atomic.StoreInt64(&m.bufferPoolPuts, stats["puts"])
	atomic.StoreInt64(&m.bufferPoolMisses, stats["misses"])
	atomic.StoreInt64(&m.bufferPoolCreates, stats["creates"])
	m.checkAndReport()
}

// RecordPacketPoolStats records packet pool statistics
func (m *VPNMetrics) RecordPacketPoolStats(stats map[string]int64) {
	atomic.StoreInt64(&m.packetPoolGets, stats["gets"])
	atomic.StoreInt64(&m.packetPoolPuts, stats["puts"])
	atomic.StoreInt64(&m.packetPoolCreates, stats["creates"])
	m.checkAndReport()
}

// RecordQueueStats records lock-free queue statistics
func (m *VPNMetrics) RecordQueueStats(enqueues, dequeues, size int64) {
	atomic.StoreInt64(&m.queueEnqueues, enqueues)
	atomic.StoreInt64(&m.queueDequeues, dequeues)
	atomic.StoreInt64(&m.queueSize, size)
	m.checkAndReport()
}

// RecordNetworkLatency records network latency
func (m *VPNMetrics) RecordNetworkLatency(latency time.Duration) {
	// Convert to microseconds
	latencyMicros := latency.Microseconds()

	// Update total and count
	atomic.AddInt64(&m.networkLatencyTotal, latencyMicros)
	atomic.AddInt64(&m.networkLatencyCount, 1)

	// Update min if needed
	for {
		currentMin := atomic.LoadInt64(&m.networkLatencyMin)
		if currentMin == 0 || latencyMicros < currentMin {
			if atomic.CompareAndSwapInt64(&m.networkLatencyMin, currentMin, latencyMicros) {
				break
			}
		} else {
			break
		}
	}

	// Update max if needed
	for {
		currentMax := atomic.LoadInt64(&m.networkLatencyMax)
		if latencyMicros > currentMax {
			if atomic.CompareAndSwapInt64(&m.networkLatencyMax, currentMax, latencyMicros) {
				break
			}
		} else {
			break
		}
	}

	// Calculate jitter (simplified as absolute difference from average)
	var jitter int64
	if count := atomic.LoadInt64(&m.networkLatencyCount); count > 1 {
		avg := atomic.LoadInt64(&m.networkLatencyTotal) / count
		jitter = latencyMicros - avg
		if jitter < 0 {
			jitter = -jitter
		}

		// Update jitter metrics
		atomic.AddInt64(&m.networkJitterTotal, jitter)
		atomic.AddInt64(&m.networkJitterCount, 1)
	}

	m.checkAndReport()
}

// RecordNetworkResult records the result of a network operation
func (m *VPNMetrics) RecordNetworkResult(success bool) {
	if success {
		atomic.AddInt64(&m.networkSuccessCount, 1)
	} else {
		atomic.AddInt64(&m.networkFailureCount, 1)
	}
	m.checkAndReport()
}

// RecordStreamCreated records a stream creation
func (m *VPNMetrics) RecordStreamCreated() {
	atomic.AddInt64(&m.totalStreamsCreated, 1)
	atomic.AddInt64(&m.activeStreams, 1)
	m.checkAndReport()
}

// RecordStreamRemoved records a stream removal
func (m *VPNMetrics) RecordStreamRemoved() {
	atomic.AddInt64(&m.totalStreamsRemoved, 1)
	atomic.AddInt64(&m.activeStreams, -1)
	m.checkAndReport()
}

// UpdateActiveStreams updates the active streams count directly
func (m *VPNMetrics) UpdateActiveStreams(count int64) {
	atomic.StoreInt64(&m.activeStreams, count)
	m.checkAndReport()
}

// UpdateActiveConnections updates the active connections count
func (m *VPNMetrics) UpdateActiveConnections(count int64) {
	atomic.StoreInt64(&m.activeConnections, count)
	m.checkAndReport()
}

// checkAndReport checks if it's time to report metrics and reports them if needed
func (m *VPNMetrics) checkAndReport() {
	if time.Since(m.lastReportTime) > m.reportInterval {
		m.reportMetrics()
		m.lastReportTime = time.Now()
	}
}

// reportMetrics reports all metrics
func (m *VPNMetrics) reportMetrics() {
	// Calculate derived metrics
	var avgBatchSize float64
	if batchSizeCount := atomic.LoadInt64(&m.batchSizeCount); batchSizeCount > 0 {
		avgBatchSize = float64(atomic.LoadInt64(&m.batchSizeTotal)) / float64(batchSizeCount)
	}

	var avgProcessingTime float64
	if processingTimeCount := atomic.LoadInt64(&m.processingTimeCount); processingTimeCount > 0 {
		avgProcessingTime = float64(atomic.LoadInt64(&m.processingTimeTotal)) / float64(processingTimeCount)
	}

	// Calculate network metrics
	var avgNetworkLatency float64
	var avgNetworkJitter float64
	var networkSuccessRate float64

	if latencyCount := atomic.LoadInt64(&m.networkLatencyCount); latencyCount > 0 {
		avgNetworkLatency = float64(atomic.LoadInt64(&m.networkLatencyTotal)) / float64(latencyCount)
	}

	if jitterCount := atomic.LoadInt64(&m.networkJitterCount); jitterCount > 0 {
		avgNetworkJitter = float64(atomic.LoadInt64(&m.networkJitterTotal)) / float64(jitterCount)
	}

	successCount := atomic.LoadInt64(&m.networkSuccessCount)
	failureCount := atomic.LoadInt64(&m.networkFailureCount)
	totalCount := successCount + failureCount

	if totalCount > 0 {
		networkSuccessRate = float64(successCount) / float64(totalCount) * 100.0
	}

	// Report metrics
	m.logger.WithFields(logrus.Fields{
		"packets_processed":      atomic.LoadInt64(&m.packetsProcessed),
		"packets_dropped":        atomic.LoadInt64(&m.packetsDropped),
		"bytes_processed":        atomic.LoadInt64(&m.bytesProcessed),
		"batches_processed":      atomic.LoadInt64(&m.batchesProcessed),
		"avg_batch_size":         avgBatchSize,
		"avg_processing_time_us": avgProcessingTime,
		"buffer_pool_gets":       atomic.LoadInt64(&m.bufferPoolGets),
		"buffer_pool_puts":       atomic.LoadInt64(&m.bufferPoolPuts),
		"buffer_pool_misses":     atomic.LoadInt64(&m.bufferPoolMisses),
		"buffer_pool_creates":    atomic.LoadInt64(&m.bufferPoolCreates),
		"packet_pool_gets":       atomic.LoadInt64(&m.packetPoolGets),
		"packet_pool_puts":       atomic.LoadInt64(&m.packetPoolPuts),
		"packet_pool_creates":    atomic.LoadInt64(&m.packetPoolCreates),
		"queue_enqueues":         atomic.LoadInt64(&m.queueEnqueues),
		"queue_dequeues":         atomic.LoadInt64(&m.queueDequeues),
		"queue_size":             atomic.LoadInt64(&m.queueSize),
		"network_latency_avg_us": avgNetworkLatency,
		"network_latency_min_us": atomic.LoadInt64(&m.networkLatencyMin),
		"network_latency_max_us": atomic.LoadInt64(&m.networkLatencyMax),
		"network_jitter_avg_us":  avgNetworkJitter,
		"network_success_rate":   networkSuccessRate,
		"network_success_count":  successCount,
		"network_failure_count":  failureCount,
		"active_streams":         atomic.LoadInt64(&m.activeStreams),
		"total_streams_created":  atomic.LoadInt64(&m.totalStreamsCreated),
		"total_streams_removed":  atomic.LoadInt64(&m.totalStreamsRemoved),
		"active_connections":     atomic.LoadInt64(&m.activeConnections),
	}).Info("VPN metrics")

	// Reset counters for rate calculations
	atomic.StoreInt64(&m.packetsProcessed, 0)
	atomic.StoreInt64(&m.packetsDropped, 0)
	atomic.StoreInt64(&m.bytesProcessed, 0)
	atomic.StoreInt64(&m.batchesProcessed, 0)
	atomic.StoreInt64(&m.batchSizeTotal, 0)
	atomic.StoreInt64(&m.batchSizeCount, 0)
	atomic.StoreInt64(&m.processingTimeTotal, 0)
	atomic.StoreInt64(&m.processingTimeCount, 0)

	// Reset network counters but keep min/max values
	atomic.StoreInt64(&m.networkLatencyTotal, 0)
	atomic.StoreInt64(&m.networkLatencyCount, 0)
	atomic.StoreInt64(&m.networkJitterTotal, 0)
	atomic.StoreInt64(&m.networkJitterCount, 0)
	atomic.StoreInt64(&m.networkSuccessCount, 0)
	atomic.StoreInt64(&m.networkFailureCount, 0)
}

// GetMetrics returns all metrics as a map
func (m *VPNMetrics) GetMetrics() map[string]interface{} {
	// Calculate derived metrics
	var avgBatchSize float64
	if batchSizeCount := atomic.LoadInt64(&m.batchSizeCount); batchSizeCount > 0 {
		avgBatchSize = float64(atomic.LoadInt64(&m.batchSizeTotal)) / float64(batchSizeCount)
	}

	var avgProcessingTime float64
	if processingTimeCount := atomic.LoadInt64(&m.processingTimeCount); processingTimeCount > 0 {
		avgProcessingTime = float64(atomic.LoadInt64(&m.processingTimeTotal)) / float64(processingTimeCount)
	}

	// Calculate network metrics
	var avgNetworkLatency float64
	var avgNetworkJitter float64
	var networkSuccessRate float64

	if latencyCount := atomic.LoadInt64(&m.networkLatencyCount); latencyCount > 0 {
		avgNetworkLatency = float64(atomic.LoadInt64(&m.networkLatencyTotal)) / float64(latencyCount)
	}

	if jitterCount := atomic.LoadInt64(&m.networkJitterCount); jitterCount > 0 {
		avgNetworkJitter = float64(atomic.LoadInt64(&m.networkJitterTotal)) / float64(jitterCount)
	}

	successCount := atomic.LoadInt64(&m.networkSuccessCount)
	failureCount := atomic.LoadInt64(&m.networkFailureCount)
	totalCount := successCount + failureCount

	if totalCount > 0 {
		networkSuccessRate = float64(successCount) / float64(totalCount) * 100.0
	}

	return map[string]interface{}{
		"packets_processed":      atomic.LoadInt64(&m.packetsProcessed),
		"packets_dropped":        atomic.LoadInt64(&m.packetsDropped),
		"bytes_processed":        atomic.LoadInt64(&m.bytesProcessed),
		"batches_processed":      atomic.LoadInt64(&m.batchesProcessed),
		"avg_batch_size":         avgBatchSize,
		"avg_processing_time_us": avgProcessingTime,
		"buffer_pool_gets":       atomic.LoadInt64(&m.bufferPoolGets),
		"buffer_pool_puts":       atomic.LoadInt64(&m.bufferPoolPuts),
		"buffer_pool_misses":     atomic.LoadInt64(&m.bufferPoolMisses),
		"buffer_pool_creates":    atomic.LoadInt64(&m.bufferPoolCreates),
		"packet_pool_gets":       atomic.LoadInt64(&m.packetPoolGets),
		"packet_pool_puts":       atomic.LoadInt64(&m.packetPoolPuts),
		"packet_pool_creates":    atomic.LoadInt64(&m.packetPoolCreates),
		"queue_enqueues":         atomic.LoadInt64(&m.queueEnqueues),
		"queue_dequeues":         atomic.LoadInt64(&m.queueDequeues),
		"queue_size":             atomic.LoadInt64(&m.queueSize),
		"network_latency_avg_us": avgNetworkLatency,
		"network_latency_min_us": atomic.LoadInt64(&m.networkLatencyMin),
		"network_latency_max_us": atomic.LoadInt64(&m.networkLatencyMax),
		"network_jitter_avg_us":  avgNetworkJitter,
		"network_success_rate":   networkSuccessRate,
		"network_success_count":  successCount,
		"network_failure_count":  failureCount,
		"active_streams":         atomic.LoadInt64(&m.activeStreams),
		"total_streams_created":  atomic.LoadInt64(&m.totalStreamsCreated),
		"total_streams_removed":  atomic.LoadInt64(&m.totalStreamsRemoved),
		"active_connections":     atomic.LoadInt64(&m.activeConnections),
	}
}

// Global metrics instance
var GlobalMetrics = NewVPNMetrics()

// StartMetricsCollection starts a background goroutine to collect metrics
func StartMetricsCollection() {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			// Collect buffer pool stats
			if pool, ok := utils.GetEnhancedBufferPool(); ok {
				GlobalMetrics.RecordBufferPoolStats(pool.Stats())
			}

			// Collect packet pool stats
			GlobalMetrics.RecordPacketPoolStats(types.GlobalPacketPool.Stats())
		}
	}()
}

// UpdateStreamMetrics updates the stream metrics from the stream pool and manager
func UpdateStreamMetrics(activeStreams, totalCreated, totalRemoved, activeConnections int64) {
	GlobalMetrics.UpdateActiveStreams(activeStreams)
	atomic.StoreInt64(&GlobalMetrics.totalStreamsCreated, totalCreated)
	atomic.StoreInt64(&GlobalMetrics.totalStreamsRemoved, totalRemoved)
	GlobalMetrics.UpdateActiveConnections(activeConnections)
}
