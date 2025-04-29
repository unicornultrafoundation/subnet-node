package network

import (
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// AdaptiveBatchSizer dynamically adjusts batch sizes based on performance metrics
type AdaptiveBatchSizer struct {
	// Current batch size
	batchSize int32
	// Minimum batch size
	minBatchSize int32
	// Maximum batch size
	maxBatchSize int32
	// Target processing time in microseconds
	targetProcessingTime int64
	// Current average processing time in microseconds
	avgProcessingTime int64
	// Number of samples for the average
	samples int64
	// Total processing time for the samples
	totalProcessingTime int64
	// Last adjustment time
	lastAdjustment time.Time
	// Adjustment interval
	adjustmentInterval time.Duration
	// Logger
	logger *logrus.Entry

	// Network condition metrics
	packetSuccessRate float64
	networkLatency    int64
	networkJitter     int64

	// Error tracking
	errorCount       int64
	successCount     int64
	lastErrorCount   int64
	lastSuccessCount int64
}

// NewAdaptiveBatchSizer creates a new adaptive batch sizer
func NewAdaptiveBatchSizer(initialSize, minSize, maxSize int, targetTime int64) *AdaptiveBatchSizer {
	return &AdaptiveBatchSizer{
		batchSize:            int32(initialSize),
		minBatchSize:         int32(minSize),
		maxBatchSize:         int32(maxSize),
		targetProcessingTime: targetTime,
		avgProcessingTime:    0,
		samples:              0,
		totalProcessingTime:  0,
		lastAdjustment:       time.Now(),
		adjustmentInterval:   time.Second * 5,
		logger:               logrus.WithField("component", "adaptive_batch"),
		packetSuccessRate:    1.0, // Start with assumption of 100% success
		networkLatency:       0,
		networkJitter:        0,
		errorCount:           0,
		successCount:         0,
		lastErrorCount:       0,
		lastSuccessCount:     0,
	}
}

// RecordNetworkLatency records a network latency measurement
func (a *AdaptiveBatchSizer) RecordNetworkLatency(latency time.Duration) {
	// Convert to microseconds for consistency
	latencyMicros := latency.Microseconds()

	// Use exponential moving average with alpha=0.2
	if a.networkLatency == 0 {
		a.networkLatency = latencyMicros
	} else {
		a.networkLatency = int64(0.8*float64(a.networkLatency) + 0.2*float64(latencyMicros))
	}

	// Calculate jitter (simplified as absolute difference from average)
	jitter := latencyMicros - a.networkLatency
	if jitter < 0 {
		jitter = -jitter
	}

	// Update jitter with exponential moving average
	if a.networkJitter == 0 {
		a.networkJitter = jitter
	} else {
		a.networkJitter = int64(0.8*float64(a.networkJitter) + 0.2*float64(jitter))
	}
}

// RecordPacketResult records the result of a packet transmission
func (a *AdaptiveBatchSizer) RecordPacketResult(success bool) {
	if success {
		atomic.AddInt64(&a.successCount, 1)
	} else {
		atomic.AddInt64(&a.errorCount, 1)
	}

	// Update success rate periodically
	if time.Since(a.lastAdjustment) > a.adjustmentInterval {
		a.updateSuccessRate()
	}
}

// updateSuccessRate calculates the current packet success rate
func (a *AdaptiveBatchSizer) updateSuccessRate() {
	currentSuccess := atomic.LoadInt64(&a.successCount)
	currentErrors := atomic.LoadInt64(&a.errorCount)

	// Calculate success rate for the period since last update
	periodSuccess := currentSuccess - a.lastSuccessCount
	periodErrors := currentErrors - a.lastErrorCount
	periodTotal := periodSuccess + periodErrors

	if periodTotal > 0 {
		newRate := float64(periodSuccess) / float64(periodTotal)

		// Use exponential moving average with alpha=0.3
		a.packetSuccessRate = 0.7*a.packetSuccessRate + 0.3*newRate
	}

	// Update last counts
	a.lastSuccessCount = currentSuccess
	a.lastErrorCount = currentErrors
}

// GetBatchSize returns the current batch size
func (a *AdaptiveBatchSizer) GetBatchSize() int {
	return int(atomic.LoadInt32(&a.batchSize))
}

// RecordProcessingTime records a batch processing time and adjusts the batch size if needed
func (a *AdaptiveBatchSizer) RecordProcessingTime(processingTime time.Duration, batchSize int) {
	// Convert to microseconds
	processingTimeMicros := processingTime.Microseconds()

	// Calculate processing time per packet
	perPacketTime := processingTimeMicros / int64(batchSize)

	// Update the average
	atomic.AddInt64(&a.totalProcessingTime, perPacketTime)
	newSamples := atomic.AddInt64(&a.samples, 1)

	// Update the average processing time
	atomic.StoreInt64(&a.avgProcessingTime, a.totalProcessingTime/newSamples)

	// Check if it's time to adjust the batch size
	if time.Since(a.lastAdjustment) > a.adjustmentInterval {
		a.adjustBatchSize()
		a.lastAdjustment = time.Now()

		// Reset the samples
		atomic.StoreInt64(&a.samples, 0)
		atomic.StoreInt64(&a.totalProcessingTime, 0)
	}
}

// adjustBatchSize adjusts the batch size based on multiple factors:
// - Average processing time
// - Network latency and jitter
// - Packet success rate
func (a *AdaptiveBatchSizer) adjustBatchSize() {
	avgTime := atomic.LoadInt64(&a.avgProcessingTime)
	currentSize := atomic.LoadInt32(&a.batchSize)

	// Calculate a network health score (0.0 to 1.0)
	// Lower score means worse network conditions
	networkHealthScore := a.calculateNetworkHealthScore()

	// If network conditions are poor, be more aggressive in reducing batch size
	if networkHealthScore < 0.5 {
		// Network conditions are poor, reduce batch size more aggressively
		if currentSize > a.minBatchSize {
			// Decrease by 40-50% depending on how bad the network is
			reductionFactor := 0.5 + (networkHealthScore * 0.2) // 0.5 to 0.7
			newSize := max(a.minBatchSize, int32(float64(currentSize)*reductionFactor))
			atomic.StoreInt32(&a.batchSize, newSize)
			a.logger.WithFields(logrus.Fields{
				"old_size":            currentSize,
				"new_size":            newSize,
				"avg_time":            avgTime,
				"target":              a.targetProcessingTime,
				"network_health":      networkHealthScore,
				"packet_success_rate": a.packetSuccessRate,
				"network_latency_us":  a.networkLatency,
				"network_jitter_us":   a.networkJitter,
			}).Info("Decreased batch size due to poor network conditions")
		}
		return
	}

	// Normal adjustment based on processing time, but modulated by network health
	if avgTime > a.targetProcessingTime {
		// Only decrease if we're above the minimum
		if currentSize > a.minBatchSize {
			// Decrease by 15-35% depending on network health
			reductionFactor := 0.65 + (networkHealthScore * 0.2) // 0.65 to 0.85
			newSize := max(a.minBatchSize, int32(float64(currentSize)*reductionFactor))
			atomic.StoreInt32(&a.batchSize, newSize)
			a.logger.WithFields(logrus.Fields{
				"old_size":            currentSize,
				"new_size":            newSize,
				"avg_time":            avgTime,
				"target":              a.targetProcessingTime,
				"network_health":      networkHealthScore,
				"packet_success_rate": a.packetSuccessRate,
			}).Info("Decreased batch size")
		}
	} else if avgTime < a.targetProcessingTime/2 && networkHealthScore > 0.8 {
		// Only increase if network conditions are good and we're below the maximum
		if currentSize < a.maxBatchSize {
			// Increase by 10-25% depending on network health
			growthFactor := 1.1 + (networkHealthScore * 0.15) // 1.1 to 1.25
			newSize := min(a.maxBatchSize, int32(float64(currentSize)*growthFactor))
			atomic.StoreInt32(&a.batchSize, newSize)
			a.logger.WithFields(logrus.Fields{
				"old_size":            currentSize,
				"new_size":            newSize,
				"avg_time":            avgTime,
				"target":              a.targetProcessingTime,
				"network_health":      networkHealthScore,
				"packet_success_rate": a.packetSuccessRate,
			}).Info("Increased batch size")
		}
	}
}

// calculateNetworkHealthScore calculates a score from 0.0 to 1.0 representing network health
// 0.0 means very poor network conditions, 1.0 means excellent conditions
func (a *AdaptiveBatchSizer) calculateNetworkHealthScore() float64 {
	// Weight factors for different metrics
	const (
		successRateWeight = 0.5 // Success rate is most important
		latencyWeight     = 0.3 // Latency is next most important
		jitterWeight      = 0.2 // Jitter is least important
	)

	// Calculate success rate score (0.0 to 1.0)
	successRateScore := a.packetSuccessRate

	// Calculate latency score (0.0 to 1.0)
	// We consider latency up to 200ms (200,000 μs) as the range
	var latencyScore float64 = 1.0
	if a.networkLatency > 0 {
		// Lower is better, so we invert the score
		latencyRatio := float64(a.networkLatency) / 200000.0
		if latencyRatio > 1.0 {
			latencyRatio = 1.0
		}
		latencyScore = 1.0 - latencyRatio
	}

	// Calculate jitter score (0.0 to 1.0)
	// We consider jitter up to 50ms (50,000 μs) as the range
	var jitterScore float64 = 1.0
	if a.networkJitter > 0 {
		// Lower is better, so we invert the score
		jitterRatio := float64(a.networkJitter) / 50000.0
		if jitterRatio > 1.0 {
			jitterRatio = 1.0
		}
		jitterScore = 1.0 - jitterRatio
	}

	// Calculate weighted average
	networkHealthScore := (successRateScore * successRateWeight) +
		(latencyScore * latencyWeight) +
		(jitterScore * jitterWeight)

	return networkHealthScore
}

// min returns the minimum of two int32 values
func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two int32 values
func max(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

// minFloat64 returns the minimum of two float64 values
func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// maxFloat64 returns the maximum of two float64 values
func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
