package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

var metricsLog = logrus.WithField("service", "vpn-metrics")

// EnhancedMetricsCollector is an enhanced metrics collector for the VPN system
type EnhancedMetricsCollector struct {
	// Basic metrics
	PacketsProcessed int64
	PacketsDropped   int64
	BytesSent        int64
	BytesReceived    int64
	Errors           int64

	// Stream metrics
	ActiveStreams    int64
	StreamsCreated   int64
	StreamsReleased  int64
	StreamErrors     int64
	OverflowPackets  int64
	
	// Connection metrics
	ActiveConnections int64
	ConnectionErrors  int64
	
	// Performance metrics
	ProcessingLatency struct {
		sync.RWMutex
		samples []time.Duration
		sum     time.Duration
		count   int
	}
	
	// Resource metrics
	BufferUtilization struct {
		sync.RWMutex
		max       int
		avg       float64
		samples   int
		overflows int
	}
	
	// Health metrics
	CircuitBreakerTrips int64
	RetryAttempts       int64
	RetrySuccesses      int64
	
	// Alerting thresholds
	alertThresholds struct {
		packetDropRate       float64 // Percentage of packets dropped
		errorRate            float64 // Percentage of errors
		highLatencyThreshold time.Duration // Threshold for high latency
		bufferUtilThreshold  int // Threshold for high buffer utilization
	}
	
	// Alert state
	alerts struct {
		sync.RWMutex
		active map[string]time.Time // Map of active alerts and when they were triggered
	}
	
	// Alert callback
	alertCallback func(string, map[string]interface{})
}

// NewEnhancedMetricsCollector creates a new enhanced metrics collector
func NewEnhancedMetricsCollector() *EnhancedMetricsCollector {
	collector := &EnhancedMetricsCollector{
		alertThresholds: struct {
			packetDropRate       float64
			errorRate            float64
			highLatencyThreshold time.Duration
			bufferUtilThreshold  int
		}{
			packetDropRate:       0.05, // 5% packet drop rate
			errorRate:            0.05, // 5% error rate
			highLatencyThreshold: 100 * time.Millisecond, // 100ms latency
			bufferUtilThreshold:  80, // 80% buffer utilization
		},
	}
	
	// Initialize the alerts map
	collector.alerts.active = make(map[string]time.Time)
	
	return collector
}

// SetAlertCallback sets the callback function for alerts
func (c *EnhancedMetricsCollector) SetAlertCallback(callback func(string, map[string]interface{})) {
	c.alertCallback = callback
}

// SetAlertThresholds sets the thresholds for alerts
func (c *EnhancedMetricsCollector) SetAlertThresholds(
	packetDropRate float64,
	errorRate float64,
	highLatencyThreshold time.Duration,
	bufferUtilThreshold int,
) {
	c.alertThresholds.packetDropRate = packetDropRate
	c.alertThresholds.errorRate = errorRate
	c.alertThresholds.highLatencyThreshold = highLatencyThreshold
	c.alertThresholds.bufferUtilThreshold = bufferUtilThreshold
}

// AddPacketProcessed increments the packets processed counter
func (c *EnhancedMetricsCollector) AddPacketProcessed() {
	atomic.AddInt64(&c.PacketsProcessed, 1)
}

// AddPacketsProcessed adds to the packets processed counter
func (c *EnhancedMetricsCollector) AddPacketsProcessed(count int64) {
	atomic.AddInt64(&c.PacketsProcessed, count)
}

// AddPacketDropped increments the packets dropped counter
func (c *EnhancedMetricsCollector) AddPacketDropped() {
	atomic.AddInt64(&c.PacketsDropped, 1)
	
	// Check for high packet drop rate
	c.checkPacketDropRate()
}

// AddPacketsDropped adds to the packets dropped counter
func (c *EnhancedMetricsCollector) AddPacketsDropped(count int64) {
	atomic.AddInt64(&c.PacketsDropped, count)
	
	// Check for high packet drop rate
	c.checkPacketDropRate()
}

// AddBytesSent adds to the bytes sent counter
func (c *EnhancedMetricsCollector) AddBytesSent(bytes int64) {
	atomic.AddInt64(&c.BytesSent, bytes)
}

// AddBytesReceived adds to the bytes received counter
func (c *EnhancedMetricsCollector) AddBytesReceived(bytes int64) {
	atomic.AddInt64(&c.BytesReceived, bytes)
}

// AddError increments the errors counter
func (c *EnhancedMetricsCollector) AddError() {
	atomic.AddInt64(&c.Errors, 1)
	
	// Check for high error rate
	c.checkErrorRate()
}

// AddErrors adds to the errors counter
func (c *EnhancedMetricsCollector) AddErrors(count int64) {
	atomic.AddInt64(&c.Errors, count)
	
	// Check for high error rate
	c.checkErrorRate()
}

// UpdateActiveStreams updates the active streams counter
func (c *EnhancedMetricsCollector) UpdateActiveStreams(count int64) {
	atomic.StoreInt64(&c.ActiveStreams, count)
}

// AddStreamCreated increments the streams created counter
func (c *EnhancedMetricsCollector) AddStreamCreated() {
	atomic.AddInt64(&c.StreamsCreated, 1)
}

// AddStreamReleased increments the streams released counter
func (c *EnhancedMetricsCollector) AddStreamReleased() {
	atomic.AddInt64(&c.StreamsReleased, 1)
}

// AddStreamError increments the stream errors counter
func (c *EnhancedMetricsCollector) AddStreamError() {
	atomic.AddInt64(&c.StreamErrors, 1)
}

// AddOverflowPacket increments the overflow packets counter
func (c *EnhancedMetricsCollector) AddOverflowPacket() {
	atomic.AddInt64(&c.OverflowPackets, 1)
}

// AddOverflowPackets adds to the overflow packets counter
func (c *EnhancedMetricsCollector) AddOverflowPackets(count int64) {
	atomic.AddInt64(&c.OverflowPackets, count)
}

// UpdateActiveConnections updates the active connections counter
func (c *EnhancedMetricsCollector) UpdateActiveConnections(count int64) {
	atomic.StoreInt64(&c.ActiveConnections, count)
}

// AddConnectionError increments the connection errors counter
func (c *EnhancedMetricsCollector) AddConnectionError() {
	atomic.AddInt64(&c.ConnectionErrors, 1)
}

// AddProcessingLatency adds a processing latency sample
func (c *EnhancedMetricsCollector) AddProcessingLatency(latency time.Duration) {
	c.ProcessingLatency.Lock()
	defer c.ProcessingLatency.Unlock()
	
	// Add the sample
	c.ProcessingLatency.samples = append(c.ProcessingLatency.samples, latency)
	c.ProcessingLatency.sum += latency
	c.ProcessingLatency.count++
	
	// Keep only the last 1000 samples
	if len(c.ProcessingLatency.samples) > 1000 {
		removed := c.ProcessingLatency.samples[0]
		c.ProcessingLatency.samples = c.ProcessingLatency.samples[1:]
		c.ProcessingLatency.sum -= removed
		c.ProcessingLatency.count--
	}
	
	// Check for high latency
	if latency > c.alertThresholds.highLatencyThreshold {
		c.triggerAlert("high_latency", map[string]interface{}{
			"latency_ms": latency.Milliseconds(),
			"threshold_ms": c.alertThresholds.highLatencyThreshold.Milliseconds(),
		})
	}
}

// GetAverageProcessingLatency returns the average processing latency
func (c *EnhancedMetricsCollector) GetAverageProcessingLatency() time.Duration {
	c.ProcessingLatency.RLock()
	defer c.ProcessingLatency.RUnlock()
	
	if c.ProcessingLatency.count == 0 {
		return 0
	}
	
	return c.ProcessingLatency.sum / time.Duration(c.ProcessingLatency.count)
}

// UpdateBufferUtilization updates the buffer utilization metrics
func (c *EnhancedMetricsCollector) UpdateBufferUtilization(utilization int) {
	c.BufferUtilization.Lock()
	defer c.BufferUtilization.Unlock()
	
	// Update max utilization
	if utilization > c.BufferUtilization.max {
		c.BufferUtilization.max = utilization
	}
	
	// Update average utilization
	c.BufferUtilization.avg = (c.BufferUtilization.avg*float64(c.BufferUtilization.samples) + float64(utilization)) / 
		float64(c.BufferUtilization.samples+1)
	c.BufferUtilization.samples++
	
	// Count overflows
	if utilization >= 100 {
		c.BufferUtilization.overflows++
	}
	
	// Check for high buffer utilization
	if utilization > c.alertThresholds.bufferUtilThreshold {
		c.triggerAlert("high_buffer_utilization", map[string]interface{}{
			"utilization": utilization,
			"threshold": c.alertThresholds.bufferUtilThreshold,
		})
	}
}

// GetMaxBufferUtilization returns the maximum buffer utilization
func (c *EnhancedMetricsCollector) GetMaxBufferUtilization() int {
	c.BufferUtilization.RLock()
	defer c.BufferUtilization.RUnlock()
	
	return c.BufferUtilization.max
}

// GetAverageBufferUtilization returns the average buffer utilization
func (c *EnhancedMetricsCollector) GetAverageBufferUtilization() float64 {
	c.BufferUtilization.RLock()
	defer c.BufferUtilization.RUnlock()
	
	return c.BufferUtilization.avg
}

// GetBufferOverflows returns the number of buffer overflows
func (c *EnhancedMetricsCollector) GetBufferOverflows() int {
	c.BufferUtilization.RLock()
	defer c.BufferUtilization.RUnlock()
	
	return c.BufferUtilization.overflows
}

// AddCircuitBreakerTrip increments the circuit breaker trips counter
func (c *EnhancedMetricsCollector) AddCircuitBreakerTrip() {
	atomic.AddInt64(&c.CircuitBreakerTrips, 1)
	
	// Trigger an alert for circuit breaker trip
	c.triggerAlert("circuit_breaker_trip", map[string]interface{}{
		"total_trips": atomic.LoadInt64(&c.CircuitBreakerTrips),
	})
}

// AddRetryAttempt increments the retry attempts counter
func (c *EnhancedMetricsCollector) AddRetryAttempt() {
	atomic.AddInt64(&c.RetryAttempts, 1)
}

// AddRetrySuccess increments the retry successes counter
func (c *EnhancedMetricsCollector) AddRetrySuccess() {
	atomic.AddInt64(&c.RetrySuccesses, 1)
}

// GetMetrics returns all metrics as a map
func (c *EnhancedMetricsCollector) GetMetrics() map[string]interface{} {
	// Get processing latency metrics
	avgLatency := c.GetAverageProcessingLatency()
	
	// Get buffer utilization metrics
	maxBufferUtil := c.GetMaxBufferUtilization()
	avgBufferUtil := c.GetAverageBufferUtilization()
	bufferOverflows := c.GetBufferOverflows()
	
	// Calculate derived metrics
	var packetDropRate float64
	totalPackets := atomic.LoadInt64(&c.PacketsProcessed) + atomic.LoadInt64(&c.PacketsDropped)
	if totalPackets > 0 {
		packetDropRate = float64(atomic.LoadInt64(&c.PacketsDropped)) / float64(totalPackets)
	}
	
	var errorRate float64
	totalOperations := totalPackets
	if totalOperations > 0 {
		errorRate = float64(atomic.LoadInt64(&c.Errors)) / float64(totalOperations)
	}
	
	var retrySuccessRate float64
	totalRetries := atomic.LoadInt64(&c.RetryAttempts)
	if totalRetries > 0 {
		retrySuccessRate = float64(atomic.LoadInt64(&c.RetrySuccesses)) / float64(totalRetries)
	}
	
	return map[string]interface{}{
		// Basic metrics
		"packets_processed": atomic.LoadInt64(&c.PacketsProcessed),
		"packets_dropped":   atomic.LoadInt64(&c.PacketsDropped),
		"bytes_sent":        atomic.LoadInt64(&c.BytesSent),
		"bytes_received":    atomic.LoadInt64(&c.BytesReceived),
		"errors":            atomic.LoadInt64(&c.Errors),
		
		// Stream metrics
		"active_streams":    atomic.LoadInt64(&c.ActiveStreams),
		"streams_created":   atomic.LoadInt64(&c.StreamsCreated),
		"streams_released":  atomic.LoadInt64(&c.StreamsReleased),
		"stream_errors":     atomic.LoadInt64(&c.StreamErrors),
		"overflow_packets":  atomic.LoadInt64(&c.OverflowPackets),
		
		// Connection metrics
		"active_connections": atomic.LoadInt64(&c.ActiveConnections),
		"connection_errors":  atomic.LoadInt64(&c.ConnectionErrors),
		
		// Performance metrics
		"avg_processing_latency_ms": avgLatency.Milliseconds(),
		
		// Resource metrics
		"max_buffer_utilization": maxBufferUtil,
		"avg_buffer_utilization": avgBufferUtil,
		"buffer_overflows":       bufferOverflows,
		
		// Health metrics
		"circuit_breaker_trips": atomic.LoadInt64(&c.CircuitBreakerTrips),
		"retry_attempts":        atomic.LoadInt64(&c.RetryAttempts),
		"retry_successes":       atomic.LoadInt64(&c.RetrySuccesses),
		
		// Derived metrics
		"packet_drop_rate":   packetDropRate,
		"error_rate":         errorRate,
		"retry_success_rate": retrySuccessRate,
	}
}

// ResetMetrics resets all metrics
func (c *EnhancedMetricsCollector) ResetMetrics() {
	// Reset basic metrics
	atomic.StoreInt64(&c.PacketsProcessed, 0)
	atomic.StoreInt64(&c.PacketsDropped, 0)
	atomic.StoreInt64(&c.BytesSent, 0)
	atomic.StoreInt64(&c.BytesReceived, 0)
	atomic.StoreInt64(&c.Errors, 0)
	
	// Reset stream metrics
	atomic.StoreInt64(&c.StreamsCreated, 0)
	atomic.StoreInt64(&c.StreamsReleased, 0)
	atomic.StoreInt64(&c.StreamErrors, 0)
	atomic.StoreInt64(&c.OverflowPackets, 0)
	
	// Reset connection metrics
	atomic.StoreInt64(&c.ConnectionErrors, 0)
	
	// Reset performance metrics
	c.ProcessingLatency.Lock()
	c.ProcessingLatency.samples = nil
	c.ProcessingLatency.sum = 0
	c.ProcessingLatency.count = 0
	c.ProcessingLatency.Unlock()
	
	// Reset resource metrics
	c.BufferUtilization.Lock()
	c.BufferUtilization.max = 0
	c.BufferUtilization.avg = 0
	c.BufferUtilization.samples = 0
	c.BufferUtilization.overflows = 0
	c.BufferUtilization.Unlock()
	
	// Reset health metrics
	atomic.StoreInt64(&c.CircuitBreakerTrips, 0)
	atomic.StoreInt64(&c.RetryAttempts, 0)
	atomic.StoreInt64(&c.RetrySuccesses, 0)
	
	// Reset alerts
	c.alerts.Lock()
	c.alerts.active = make(map[string]time.Time)
	c.alerts.Unlock()
}

// checkPacketDropRate checks if the packet drop rate is above the threshold
func (c *EnhancedMetricsCollector) checkPacketDropRate() {
	totalPackets := atomic.LoadInt64(&c.PacketsProcessed) + atomic.LoadInt64(&c.PacketsDropped)
	if totalPackets < 100 {
		// Not enough data to make a meaningful calculation
		return
	}
	
	packetDropRate := float64(atomic.LoadInt64(&c.PacketsDropped)) / float64(totalPackets)
	if packetDropRate > c.alertThresholds.packetDropRate {
		c.triggerAlert("high_packet_drop_rate", map[string]interface{}{
			"drop_rate": packetDropRate,
			"threshold": c.alertThresholds.packetDropRate,
			"dropped":   atomic.LoadInt64(&c.PacketsDropped),
			"total":     totalPackets,
		})
	}
}

// checkErrorRate checks if the error rate is above the threshold
func (c *EnhancedMetricsCollector) checkErrorRate() {
	totalOperations := atomic.LoadInt64(&c.PacketsProcessed) + atomic.LoadInt64(&c.PacketsDropped)
	if totalOperations < 100 {
		// Not enough data to make a meaningful calculation
		return
	}
	
	errorRate := float64(atomic.LoadInt64(&c.Errors)) / float64(totalOperations)
	if errorRate > c.alertThresholds.errorRate {
		c.triggerAlert("high_error_rate", map[string]interface{}{
			"error_rate": errorRate,
			"threshold":  c.alertThresholds.errorRate,
			"errors":     atomic.LoadInt64(&c.Errors),
			"total":      totalOperations,
		})
	}
}

// triggerAlert triggers an alert if it hasn't been triggered recently
func (c *EnhancedMetricsCollector) triggerAlert(alertType string, data map[string]interface{}) {
	// Check if we have a callback
	if c.alertCallback == nil {
		return
	}
	
	// Check if this alert has been triggered recently
	c.alerts.Lock()
	defer c.alerts.Unlock()
	
	lastTriggered, exists := c.alerts.active[alertType]
	if exists && time.Since(lastTriggered) < 5*time.Minute {
		// Don't trigger the same alert more than once every 5 minutes
		return
	}
	
	// Update the last triggered time
	c.alerts.active[alertType] = time.Now()
	
	// Call the callback
	go c.alertCallback(alertType, data)
	
	// Log the alert
	metricsLog.WithFields(logrus.Fields(data)).Warnf("Alert triggered: %s", alertType)
}

// GetActiveAlerts returns the active alerts
func (c *EnhancedMetricsCollector) GetActiveAlerts() map[string]time.Time {
	c.alerts.RLock()
	defer c.alerts.RUnlock()
	
	// Create a copy of the alerts map
	alerts := make(map[string]time.Time, len(c.alerts.active))
	for k, v := range c.alerts.active {
		alerts[k] = v
	}
	
	return alerts
}

// ClearAlert clears a specific alert
func (c *EnhancedMetricsCollector) ClearAlert(alertType string) {
	c.alerts.Lock()
	defer c.alerts.Unlock()
	
	delete(c.alerts.active, alertType)
}

// ClearAllAlerts clears all alerts
func (c *EnhancedMetricsCollector) ClearAllAlerts() {
	c.alerts.Lock()
	defer c.alerts.Unlock()
	
	c.alerts.active = make(map[string]time.Time)
}
