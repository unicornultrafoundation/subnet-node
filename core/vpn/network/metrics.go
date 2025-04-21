package network

import (
	"github.com/unicornultrafoundation/subnet-node/core/vpn/metrics"
)

// VPNMetricsInterface defines the interface for VPN metrics
type VPNMetricsInterface interface {
	// IncrementPacketsReceived increments the packets received counter
	IncrementPacketsReceived(bytes int)
	// IncrementPacketsSent increments the packets sent counter
	IncrementPacketsSent(bytes int)
	// IncrementPacketsDropped increments the packets dropped counter
	IncrementPacketsDropped()
}

// MetricsAdapter adapts the metrics service to the VPNMetricsInterface
type MetricsAdapter struct {
	metricsService *metrics.MetricsServiceImpl
}

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
