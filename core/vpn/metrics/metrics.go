package metrics

import (
	"sync/atomic"
)

// VPNMetrics tracks metrics for the VPN service
type VPNMetrics struct {
	// Packet metrics
	PacketsReceived int64
	PacketsSent     int64
	PacketsDropped  int64
	BytesReceived   int64
	BytesSent       int64

	// Stream metrics
	StreamErrors int64

	// Circuit breaker metrics
	CircuitOpenDrops int64
}

// NewVPNMetrics creates a new VPNMetrics instance
func NewVPNMetrics() *VPNMetrics {
	return &VPNMetrics{}
}

// IncrementPacketsReceived increments the packets received counter
func (m *VPNMetrics) IncrementPacketsReceived(bytes int) {
	atomic.AddInt64(&m.PacketsReceived, 1)
	atomic.AddInt64(&m.BytesReceived, int64(bytes))
}

// IncrementPacketsSent increments the packets sent counter
func (m *VPNMetrics) IncrementPacketsSent(bytes int) {
	atomic.AddInt64(&m.PacketsSent, 1)
	atomic.AddInt64(&m.BytesSent, int64(bytes))
}

// IncrementPacketsDropped increments the packets dropped counter
func (m *VPNMetrics) IncrementPacketsDropped() {
	atomic.AddInt64(&m.PacketsDropped, 1)
}

// IncrementStreamErrors increments the stream errors counter
func (m *VPNMetrics) IncrementStreamErrors() {
	atomic.AddInt64(&m.StreamErrors, 1)
}

// IncrementCircuitOpenDrops increments the circuit open drops counter
func (m *VPNMetrics) IncrementCircuitOpenDrops() {
	atomic.AddInt64(&m.CircuitOpenDrops, 1)
}

// GetMetrics returns the current metrics as a map
func (m *VPNMetrics) GetMetrics() map[string]int64 {
	return map[string]int64{
		"packets_received":   atomic.LoadInt64(&m.PacketsReceived),
		"packets_sent":       atomic.LoadInt64(&m.PacketsSent),
		"packets_dropped":    atomic.LoadInt64(&m.PacketsDropped),
		"bytes_received":     atomic.LoadInt64(&m.BytesReceived),
		"bytes_sent":         atomic.LoadInt64(&m.BytesSent),
		"stream_errors":      atomic.LoadInt64(&m.StreamErrors),
		"circuit_open_drops": atomic.LoadInt64(&m.CircuitOpenDrops),
	}
}

// StreamPoolMetrics tracks metrics for a stream pool
type StreamPoolMetrics struct {
	StreamsCreated      int64
	StreamsClosed       int64
	StreamsAcquired     int64
	StreamsReturned     int64
	AcquisitionFailures int64
	UnhealthyStreams    int64
}

// NewStreamPoolMetrics creates a new StreamPoolMetrics instance
func NewStreamPoolMetrics() *StreamPoolMetrics {
	return &StreamPoolMetrics{}
}

// IncrementStreamsCreated increments the streams created counter
func (m *StreamPoolMetrics) IncrementStreamsCreated() {
	atomic.AddInt64(&m.StreamsCreated, 1)
}

// IncrementStreamsClosed increments the streams closed counter
func (m *StreamPoolMetrics) IncrementStreamsClosed() {
	atomic.AddInt64(&m.StreamsClosed, 1)
}

// IncrementStreamsAcquired increments the streams acquired counter
func (m *StreamPoolMetrics) IncrementStreamsAcquired() {
	atomic.AddInt64(&m.StreamsAcquired, 1)
}

// IncrementStreamsReturned increments the streams returned counter
func (m *StreamPoolMetrics) IncrementStreamsReturned() {
	atomic.AddInt64(&m.StreamsReturned, 1)
}

// IncrementAcquisitionFailures increments the acquisition failures counter
func (m *StreamPoolMetrics) IncrementAcquisitionFailures() {
	atomic.AddInt64(&m.AcquisitionFailures, 1)
}

// IncrementUnhealthyStreams increments the unhealthy streams counter
func (m *StreamPoolMetrics) IncrementUnhealthyStreams() {
	atomic.AddInt64(&m.UnhealthyStreams, 1)
}

// GetMetrics returns the current metrics as a map
func (m *StreamPoolMetrics) GetMetrics() map[string]int64 {
	return map[string]int64{
		"streams_created":      atomic.LoadInt64(&m.StreamsCreated),
		"streams_closed":       atomic.LoadInt64(&m.StreamsClosed),
		"streams_acquired":     atomic.LoadInt64(&m.StreamsAcquired),
		"streams_returned":     atomic.LoadInt64(&m.StreamsReturned),
		"acquisition_failures": atomic.LoadInt64(&m.AcquisitionFailures),
		"unhealthy_streams":    atomic.LoadInt64(&m.UnhealthyStreams),
	}
}

// HealthMetrics tracks metrics for stream health
type HealthMetrics struct {
	ChecksPerformed  int64
	HealthyStreams   int64
	UnhealthyStreams int64
	StreamsWarmed    int64
	WarmsPerformed   int64
	WarmFailures     int64
	PeerResets       int64
}

// NewHealthMetrics creates a new HealthMetrics instance
func NewHealthMetrics() *HealthMetrics {
	return &HealthMetrics{}
}

// IncrementChecksPerformed increments the checks performed counter
func (m *HealthMetrics) IncrementChecksPerformed() {
	atomic.AddInt64(&m.ChecksPerformed, 1)
}

// IncrementHealthyStreams increments the healthy streams counter
func (m *HealthMetrics) IncrementHealthyStreams() {
	atomic.AddInt64(&m.HealthyStreams, 1)
}

// IncrementUnhealthyStreams increments the unhealthy streams counter
func (m *HealthMetrics) IncrementUnhealthyStreams() {
	atomic.AddInt64(&m.UnhealthyStreams, 1)
}

// IncrementStreamsWarmed increments the streams warmed counter
func (m *HealthMetrics) IncrementStreamsWarmed() {
	atomic.AddInt64(&m.StreamsWarmed, 1)
}

// IncrementWarmsPerformed increments the warms performed counter
func (m *HealthMetrics) IncrementWarmsPerformed() {
	atomic.AddInt64(&m.WarmsPerformed, 1)
}

// IncrementWarmFailures increments the warm failures counter
func (m *HealthMetrics) IncrementWarmFailures() {
	atomic.AddInt64(&m.WarmFailures, 1)
}

// IncrementPeerResets increments the peer resets counter
func (m *HealthMetrics) IncrementPeerResets() {
	atomic.AddInt64(&m.PeerResets, 1)
}

// GetMetrics returns the current metrics as a map
func (m *HealthMetrics) GetMetrics() map[string]int64 {
	return map[string]int64{
		"checks_performed":  atomic.LoadInt64(&m.ChecksPerformed),
		"healthy_streams":   atomic.LoadInt64(&m.HealthyStreams),
		"unhealthy_streams": atomic.LoadInt64(&m.UnhealthyStreams),
		"streams_warmed":    atomic.LoadInt64(&m.StreamsWarmed),
		"warms_performed":   atomic.LoadInt64(&m.WarmsPerformed),
		"warm_failures":     atomic.LoadInt64(&m.WarmFailures),
		"peer_resets":       atomic.LoadInt64(&m.PeerResets),
	}
}
