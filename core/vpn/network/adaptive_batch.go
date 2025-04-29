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
}

// NewAdaptiveBatchSizer creates a new adaptive batch sizer
func NewAdaptiveBatchSizer(initialSize, minSize, maxSize int, targetTime int64) *AdaptiveBatchSizer {
	return &AdaptiveBatchSizer{
		batchSize:           int32(initialSize),
		minBatchSize:        int32(minSize),
		maxBatchSize:        int32(maxSize),
		targetProcessingTime: targetTime,
		avgProcessingTime:   0,
		samples:             0,
		totalProcessingTime: 0,
		lastAdjustment:      time.Now(),
		adjustmentInterval:  time.Second * 5,
		logger:              logrus.WithField("component", "adaptive_batch"),
	}
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

// adjustBatchSize adjusts the batch size based on the average processing time
func (a *AdaptiveBatchSizer) adjustBatchSize() {
	avgTime := atomic.LoadInt64(&a.avgProcessingTime)
	currentSize := atomic.LoadInt32(&a.batchSize)
	
	// If the average time is too high, decrease the batch size
	if avgTime > a.targetProcessingTime {
		// Only decrease if we're above the minimum
		if currentSize > a.minBatchSize {
			// Decrease by 25%
			newSize := max(a.minBatchSize, int32(float64(currentSize)*0.75))
			atomic.StoreInt32(&a.batchSize, newSize)
			a.logger.WithFields(logrus.Fields{
				"old_size": currentSize,
				"new_size": newSize,
				"avg_time": avgTime,
				"target":   a.targetProcessingTime,
			}).Info("Decreased batch size")
		}
	} else if avgTime < a.targetProcessingTime/2 {
		// If the average time is much lower than the target, increase the batch size
		// Only increase if we're below the maximum
		if currentSize < a.maxBatchSize {
			// Increase by 25%
			newSize := min(a.maxBatchSize, int32(float64(currentSize)*1.25))
			atomic.StoreInt32(&a.batchSize, newSize)
			a.logger.WithFields(logrus.Fields{
				"old_size": currentSize,
				"new_size": newSize,
				"avg_time": avgTime,
				"target":   a.targetProcessingTime,
			}).Info("Increased batch size")
		}
	}
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
