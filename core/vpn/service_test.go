package vpn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/metrics"
)

// TestMetrics tests the metrics functionality
func TestMetrics(t *testing.T) {
	// Create a metrics collector
	vpnMetrics := metrics.NewVPNMetrics()

	// Increment some metrics
	vpnMetrics.IncrementPacketsReceived(100)
	vpnMetrics.IncrementPacketsSent(200)
	vpnMetrics.IncrementPacketsDropped()

	// Get the metrics
	metrics := vpnMetrics.GetMetrics()

	// Verify the metrics
	assert.Equal(t, int64(1), metrics["packets_received"])
	assert.Equal(t, int64(1), metrics["packets_sent"])
	assert.Equal(t, int64(1), metrics["packets_dropped"])
	assert.Equal(t, int64(100), metrics["bytes_received"])
	assert.Equal(t, int64(200), metrics["bytes_sent"])
}
