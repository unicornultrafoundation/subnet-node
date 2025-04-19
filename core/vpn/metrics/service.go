package metrics

import (
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
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

	// Multiplexer metrics
	IncrementPacketsSentMultiplexer(peerID peer.ID, bytes int)
	IncrementPacketsDroppedMultiplexer(peerID peer.ID)
	IncrementStreamErrorsMultiplexer(peerID peer.ID)
	IncrementStreamsCreatedMultiplexer(peerID peer.ID)
	IncrementStreamsClosedMultiplexer(peerID peer.ID)
	IncrementScaleUpOperations(peerID peer.ID)
	IncrementScaleDownOperations(peerID peer.ID)
	UpdateLatency(peerID peer.ID, latency int64)

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

	// Multiplexer metrics
	multiplexerMetrics map[string]*MultiplexerMetrics
	multiplexerMu      sync.RWMutex

	// Circuit breaker metrics
	circuitOpenCount    int64
	circuitCloseCount   int64
	circuitResetCount   int64
	requestBlockCount   int64
	requestAllowCount   int64
	activeBreakers      int64
	circuitBreakerMu    sync.RWMutex
}

// NewMetricsService creates a new metrics service
func NewMetricsService() *MetricsServiceImpl {
	return &MetricsServiceImpl{
		vpnMetrics:        NewVPNMetrics(),
		streamPoolMetrics: NewStreamPoolMetrics(),
		healthMetrics:     NewHealthMetrics(),
		multiplexerMetrics: make(map[string]*MultiplexerMetrics),
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

// Multiplexer metrics

// getOrCreateMultiplexerMetrics gets or creates multiplexer metrics for a peer
func (m *MetricsServiceImpl) getOrCreateMultiplexerMetrics(peerID peer.ID) *MultiplexerMetrics {
	peerIDStr := peerID.String()

	m.multiplexerMu.RLock()
	metrics, exists := m.multiplexerMetrics[peerIDStr]
	m.multiplexerMu.RUnlock()

	if exists {
		return metrics
	}

	m.multiplexerMu.Lock()
	defer m.multiplexerMu.Unlock()

	// Check again in case another goroutine created the metrics
	metrics, exists = m.multiplexerMetrics[peerIDStr]
	if exists {
		return metrics
	}

	// Create new metrics
	metrics = NewMultiplexerMetrics()
	m.multiplexerMetrics[peerIDStr] = metrics

	return metrics
}

// IncrementPacketsSentMultiplexer increments the packets sent counter for a multiplexer
func (m *MetricsServiceImpl) IncrementPacketsSentMultiplexer(peerID peer.ID, bytes int) {
	metrics := m.getOrCreateMultiplexerMetrics(peerID)
	metrics.IncrementPacketsSent(bytes)
}

// IncrementPacketsDroppedMultiplexer increments the packets dropped counter for a multiplexer
func (m *MetricsServiceImpl) IncrementPacketsDroppedMultiplexer(peerID peer.ID) {
	metrics := m.getOrCreateMultiplexerMetrics(peerID)
	metrics.IncrementPacketsDropped()
}

// IncrementStreamErrorsMultiplexer increments the stream errors counter for a multiplexer
func (m *MetricsServiceImpl) IncrementStreamErrorsMultiplexer(peerID peer.ID) {
	metrics := m.getOrCreateMultiplexerMetrics(peerID)
	metrics.IncrementStreamErrors()
}

// IncrementStreamsCreatedMultiplexer increments the streams created counter for a multiplexer
func (m *MetricsServiceImpl) IncrementStreamsCreatedMultiplexer(peerID peer.ID) {
	metrics := m.getOrCreateMultiplexerMetrics(peerID)
	metrics.IncrementStreamsCreated()
}

// IncrementStreamsClosedMultiplexer increments the streams closed counter for a multiplexer
func (m *MetricsServiceImpl) IncrementStreamsClosedMultiplexer(peerID peer.ID) {
	metrics := m.getOrCreateMultiplexerMetrics(peerID)
	metrics.IncrementStreamsClosed()
}

// IncrementScaleUpOperations increments the scale up operations counter for a multiplexer
func (m *MetricsServiceImpl) IncrementScaleUpOperations(peerID peer.ID) {
	metrics := m.getOrCreateMultiplexerMetrics(peerID)
	metrics.IncrementScaleUpOperations()
}

// IncrementScaleDownOperations increments the scale down operations counter for a multiplexer
func (m *MetricsServiceImpl) IncrementScaleDownOperations(peerID peer.ID) {
	metrics := m.getOrCreateMultiplexerMetrics(peerID)
	metrics.IncrementScaleDownOperations()
}

// UpdateLatency updates the average latency for a multiplexer
func (m *MetricsServiceImpl) UpdateLatency(peerID peer.ID, latency int64) {
	metrics := m.getOrCreateMultiplexerMetrics(peerID)
	metrics.UpdateLatency(latency)
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

	// Add multiplexer metrics
	m.multiplexerMu.RLock()
	for peerIDStr, multiplexerMetrics := range m.multiplexerMetrics {
		for k, v := range multiplexerMetrics.GetMetrics() {
			metrics["multiplexer_"+peerIDStr+"_"+k] = v
		}
	}
	m.multiplexerMu.RUnlock()

	return metrics
}
