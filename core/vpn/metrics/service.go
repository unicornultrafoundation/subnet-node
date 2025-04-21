package metrics

import (
	"sync"
)

// MetricsService is the interface for the centralized metrics service
type MetricsService interface {
	// VPN metrics
	IncrementPacketsReceived(bytes int)
	IncrementPacketsSent(bytes int)
	IncrementPacketsDropped()
	IncrementStreamErrors()
	IncrementCircuitOpenDrops()

	// Stream pool metrics
	IncrementStreamsCreated()
	IncrementStreamsClosed()
	IncrementStreamsAcquired()
	IncrementStreamsReturned()
	IncrementAcquisitionFailures()
	IncrementUnhealthyStreams()

	// Health metrics
	IncrementChecksPerformed()
	IncrementHealthyStreams()
	IncrementUnhealthyStreamsHealth()
	IncrementStreamsWarmed()
	IncrementWarmsPerformed()
	IncrementWarmFailures()
	IncrementPeerResets()

	// Circuit breaker metrics
	IncrementCircuitOpenCount()
	IncrementCircuitCloseCount()
	IncrementCircuitResetCount()
	IncrementRequestBlockCount()
	IncrementRequestAllowCount()
	SetActiveBreakers(count int)

	// Get all metrics
	GetAllMetrics() map[string]int64
}

// MetricsServiceImpl is the implementation of the MetricsService interface
type MetricsServiceImpl struct {
	// VPN metrics
	vpnMetrics *VPNMetrics

	// Stream pool metrics
	streamPoolMetrics *StreamPoolMetrics

	// Health metrics
	healthMetrics *HealthMetrics

	// Circuit breaker metrics
	circuitOpenCount  int64
	circuitCloseCount int64
	circuitResetCount int64
	requestBlockCount int64
	requestAllowCount int64
	activeBreakers    int64
	circuitBreakerMu  sync.RWMutex
}

// NewMetricsService creates a new metrics service
func NewMetricsService() *MetricsServiceImpl {
	return &MetricsServiceImpl{
		vpnMetrics:        NewVPNMetrics(),
		streamPoolMetrics: NewStreamPoolMetrics(),
		healthMetrics:     NewHealthMetrics(),
	}
}

// VPN metrics

// IncrementPacketsReceived increments the packets received counter
func (m *MetricsServiceImpl) IncrementPacketsReceived(bytes int) {
	m.vpnMetrics.IncrementPacketsReceived(bytes)
}

// IncrementPacketsSent increments the packets sent counter
func (m *MetricsServiceImpl) IncrementPacketsSent(bytes int) {
	m.vpnMetrics.IncrementPacketsSent(bytes)
}

// IncrementPacketsDropped increments the packets dropped counter
func (m *MetricsServiceImpl) IncrementPacketsDropped() {
	m.vpnMetrics.IncrementPacketsDropped()
}

// IncrementStreamErrors increments the stream errors counter
func (m *MetricsServiceImpl) IncrementStreamErrors() {
	m.vpnMetrics.IncrementStreamErrors()
}

// IncrementCircuitOpenDrops increments the circuit open drops counter
func (m *MetricsServiceImpl) IncrementCircuitOpenDrops() {
	m.vpnMetrics.IncrementCircuitOpenDrops()
}

// Stream pool metrics

// IncrementStreamsCreated increments the streams created counter
func (m *MetricsServiceImpl) IncrementStreamsCreated() {
	m.streamPoolMetrics.IncrementStreamsCreated()
}

// IncrementStreamsClosed increments the streams closed counter
func (m *MetricsServiceImpl) IncrementStreamsClosed() {
	m.streamPoolMetrics.IncrementStreamsClosed()
}

// IncrementStreamsAcquired increments the streams acquired counter
func (m *MetricsServiceImpl) IncrementStreamsAcquired() {
	m.streamPoolMetrics.IncrementStreamsAcquired()
}

// IncrementStreamsReturned increments the streams returned counter
func (m *MetricsServiceImpl) IncrementStreamsReturned() {
	m.streamPoolMetrics.IncrementStreamsReturned()
}

// IncrementAcquisitionFailures increments the acquisition failures counter
func (m *MetricsServiceImpl) IncrementAcquisitionFailures() {
	m.streamPoolMetrics.IncrementAcquisitionFailures()
}

// IncrementUnhealthyStreams increments the unhealthy streams counter
func (m *MetricsServiceImpl) IncrementUnhealthyStreams() {
	m.streamPoolMetrics.IncrementUnhealthyStreams()
}

// Health metrics

// IncrementChecksPerformed increments the checks performed counter
func (m *MetricsServiceImpl) IncrementChecksPerformed() {
	m.healthMetrics.IncrementChecksPerformed()
}

// IncrementHealthyStreams increments the healthy streams counter
func (m *MetricsServiceImpl) IncrementHealthyStreams() {
	m.healthMetrics.IncrementHealthyStreams()
}

// IncrementUnhealthyStreamsHealth increments the unhealthy streams counter for health
func (m *MetricsServiceImpl) IncrementUnhealthyStreamsHealth() {
	m.healthMetrics.IncrementUnhealthyStreams()
}

// IncrementStreamsWarmed increments the streams warmed counter
func (m *MetricsServiceImpl) IncrementStreamsWarmed() {
	m.healthMetrics.IncrementStreamsWarmed()
}

// IncrementWarmsPerformed increments the warms performed counter
func (m *MetricsServiceImpl) IncrementWarmsPerformed() {
	m.healthMetrics.IncrementWarmsPerformed()
}

// IncrementWarmFailures increments the warm failures counter
func (m *MetricsServiceImpl) IncrementWarmFailures() {
	m.healthMetrics.IncrementWarmFailures()
}

// IncrementPeerResets increments the peer resets counter
func (m *MetricsServiceImpl) IncrementPeerResets() {
	m.healthMetrics.IncrementPeerResets()
}

// Circuit breaker metrics

// IncrementCircuitOpenCount increments the circuit open count
func (m *MetricsServiceImpl) IncrementCircuitOpenCount() {
	m.circuitBreakerMu.Lock()
	defer m.circuitBreakerMu.Unlock()
	m.circuitOpenCount++
}

// IncrementCircuitCloseCount increments the circuit close count
func (m *MetricsServiceImpl) IncrementCircuitCloseCount() {
	m.circuitBreakerMu.Lock()
	defer m.circuitBreakerMu.Unlock()
	m.circuitCloseCount++
}

// IncrementCircuitResetCount increments the circuit reset count
func (m *MetricsServiceImpl) IncrementCircuitResetCount() {
	m.circuitBreakerMu.Lock()
	defer m.circuitBreakerMu.Unlock()
	m.circuitResetCount++
}

// IncrementRequestBlockCount increments the request block count
func (m *MetricsServiceImpl) IncrementRequestBlockCount() {
	m.circuitBreakerMu.Lock()
	defer m.circuitBreakerMu.Unlock()
	m.requestBlockCount++
}

// IncrementRequestAllowCount increments the request allow count
func (m *MetricsServiceImpl) IncrementRequestAllowCount() {
	m.circuitBreakerMu.Lock()
	defer m.circuitBreakerMu.Unlock()
	m.requestAllowCount++
}

// SetActiveBreakers sets the active breakers count
func (m *MetricsServiceImpl) SetActiveBreakers(count int) {
	m.circuitBreakerMu.Lock()
	defer m.circuitBreakerMu.Unlock()
	m.activeBreakers = int64(count)
}

// GetAllMetrics returns all metrics as a map
func (m *MetricsServiceImpl) GetAllMetrics() map[string]int64 {
	metrics := make(map[string]int64)

	// Add VPN metrics
	for k, v := range m.vpnMetrics.GetMetrics() {
		metrics[k] = v
	}

	// Add stream pool metrics
	for k, v := range m.streamPoolMetrics.GetMetrics() {
		metrics["stream_pool_"+k] = v
	}

	// Add health metrics
	for k, v := range m.healthMetrics.GetMetrics() {
		metrics["health_"+k] = v
	}

	// Add circuit breaker metrics
	m.circuitBreakerMu.RLock()
	metrics["circuit_open_count"] = m.circuitOpenCount
	metrics["circuit_close_count"] = m.circuitCloseCount
	metrics["circuit_reset_count"] = m.circuitResetCount
	metrics["request_block_count"] = m.requestBlockCount
	metrics["request_allow_count"] = m.requestAllowCount
	metrics["active_breakers"] = m.activeBreakers
	m.circuitBreakerMu.RUnlock()

	return metrics
}
