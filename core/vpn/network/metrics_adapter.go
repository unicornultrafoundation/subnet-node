package network

import (
	"github.com/unicornultrafoundation/subnet-node/core/vpn/metrics"
)

// MetricsAdapter adapts the MetricsService to the VPNMetricsInterface
type MetricsAdapter struct {
	metricsService *metrics.MetricsServiceImpl
}

// Ensure MetricsAdapter implements VPNMetricsInterface
var _ VPNMetricsInterface = (*MetricsAdapter)(nil)

// NewMetricsAdapter creates a new metrics adapter
func NewMetricsAdapter(metricsService *metrics.MetricsServiceImpl) *MetricsAdapter {
	return &MetricsAdapter{
		metricsService: metricsService,
	}
}

// IncrementPacketsReceived increments the packets received counter
func (a *MetricsAdapter) IncrementPacketsReceived(bytes int) {
	a.metricsService.IncrementPacketsReceived(bytes)
}

// IncrementPacketsSent increments the packets sent counter
func (a *MetricsAdapter) IncrementPacketsSent(bytes int) {
	a.metricsService.IncrementPacketsSent(bytes)
}

// IncrementPacketsDropped increments the packets dropped counter
func (a *MetricsAdapter) IncrementPacketsDropped() {
	a.metricsService.IncrementPacketsDropped()
}

// IncrementStreamErrors increments the stream errors counter
func (a *MetricsAdapter) IncrementStreamErrors() {
	a.metricsService.IncrementStreamErrors()
}

// IncrementCircuitOpenDrops increments the circuit open drops counter
func (a *MetricsAdapter) IncrementCircuitOpenDrops() {
	a.metricsService.IncrementCircuitOpenDrops()
}

// GetMetrics returns the current metrics as a map
func (a *MetricsAdapter) GetMetrics() map[string]int64 {
	return a.metricsService.GetAllMetrics()
}
