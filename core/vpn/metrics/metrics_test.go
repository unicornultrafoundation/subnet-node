package metrics

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMetricsIncrement tests that all metrics increment correctly
func TestMetricsIncrement(t *testing.T) {
	// Test VPN metrics
	vpnMetrics := NewVPNMetrics()
	vpnMetrics.IncrementPacketsReceived(100)
	vpnMetrics.IncrementPacketsSent(200)
	vpnMetrics.IncrementPacketsDropped()
	vpnMetrics.IncrementStreamErrors()
	vpnMetrics.IncrementCircuitOpenDrops()

	vpnResult := vpnMetrics.GetMetrics()
	assert.Equal(t, int64(1), vpnResult["packets_received"])
	assert.Equal(t, int64(1), vpnResult["packets_sent"])
	assert.Equal(t, int64(1), vpnResult["packets_dropped"])
	assert.Equal(t, int64(100), vpnResult["bytes_received"])
	assert.Equal(t, int64(200), vpnResult["bytes_sent"])
	assert.Equal(t, int64(1), vpnResult["stream_errors"])
	assert.Equal(t, int64(1), vpnResult["circuit_open_drops"])

	// Test stream pool metrics
	poolMetrics := NewStreamPoolMetrics()
	poolMetrics.IncrementStreamsCreated()
	poolMetrics.IncrementStreamsClosed()
	poolMetrics.IncrementStreamsAcquired()
	poolMetrics.IncrementStreamsReturned()
	poolMetrics.IncrementAcquisitionFailures()
	poolMetrics.IncrementUnhealthyStreams()

	poolResult := poolMetrics.GetMetrics()
	assert.Equal(t, int64(1), poolResult["streams_created"])
	assert.Equal(t, int64(1), poolResult["streams_closed"])
	assert.Equal(t, int64(1), poolResult["streams_acquired"])
	assert.Equal(t, int64(1), poolResult["streams_returned"])
	assert.Equal(t, int64(1), poolResult["acquisition_failures"])
	assert.Equal(t, int64(1), poolResult["unhealthy_streams"])

	// Test health metrics
	healthMetrics := NewHealthMetrics()
	healthMetrics.IncrementChecksPerformed()
	healthMetrics.IncrementHealthyStreams()
	healthMetrics.IncrementUnhealthyStreams()
	healthMetrics.IncrementStreamsWarmed()
	healthMetrics.IncrementWarmsPerformed()
	healthMetrics.IncrementWarmFailures()
	healthMetrics.IncrementPeerResets()

	healthResult := healthMetrics.GetMetrics()
	assert.Equal(t, int64(1), healthResult["checks_performed"])
	assert.Equal(t, int64(1), healthResult["healthy_streams"])
	assert.Equal(t, int64(1), healthResult["unhealthy_streams"])
	assert.Equal(t, int64(1), healthResult["streams_warmed"])
	assert.Equal(t, int64(1), healthResult["warms_performed"])
	assert.Equal(t, int64(1), healthResult["warm_failures"])
	assert.Equal(t, int64(1), healthResult["peer_resets"])
}

// TestConcurrentAccess tests that metrics can be safely accessed concurrently
func TestConcurrentAccess(t *testing.T) {
	// We'll just test one type of metrics since they all use atomic operations
	metrics := NewVPNMetrics()

	// Test concurrent access
	var wg sync.WaitGroup
	for range make([]struct{}, 100) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			metrics.IncrementPacketsReceived(10)
			metrics.IncrementPacketsSent(20)
			metrics.IncrementPacketsDropped()
			// Also read metrics while writing
			_ = metrics.GetMetrics()
		}()
	}
	wg.Wait()

	// Check final values
	finalMetrics := metrics.GetMetrics()
	assert.Equal(t, int64(100), finalMetrics["packets_received"])
	assert.Equal(t, int64(100), finalMetrics["packets_sent"])
	assert.Equal(t, int64(100), finalMetrics["packets_dropped"])
	assert.Equal(t, int64(1000), finalMetrics["bytes_received"])
	assert.Equal(t, int64(2000), finalMetrics["bytes_sent"])
}

// TestMetricsService tests the MetricsServiceImpl
func TestMetricsService(t *testing.T) {
	// Create a metrics service
	service := NewMetricsService()

	// Test VPN metrics
	service.IncrementPacketsReceived(100)
	service.IncrementPacketsSent(200)
	service.IncrementPacketsDropped()
	service.IncrementStreamErrors()
	service.IncrementCircuitOpenDrops()

	// Test stream pool metrics
	service.IncrementStreamsCreated()
	service.IncrementStreamsClosed()
	service.IncrementStreamsAcquired()
	service.IncrementStreamsReturned()
	service.IncrementAcquisitionFailures()
	service.IncrementUnhealthyStreams()

	// Test health metrics
	service.IncrementChecksPerformed()
	service.IncrementHealthyStreams()
	service.IncrementUnhealthyStreamsHealth()
	service.IncrementStreamsWarmed()
	service.IncrementWarmsPerformed()
	service.IncrementWarmFailures()
	service.IncrementPeerResets()

	// Test circuit breaker metrics
	service.IncrementCircuitOpenCount()
	service.IncrementCircuitCloseCount()
	service.IncrementCircuitResetCount()
	service.IncrementRequestBlockCount()
	service.IncrementRequestAllowCount()
	service.SetActiveBreakers(5)

	// Get all metrics
	allMetrics := service.GetAllMetrics()

	// Check VPN metrics
	assert.Equal(t, int64(1), allMetrics["packets_received"])
	assert.Equal(t, int64(1), allMetrics["packets_sent"])
	assert.Equal(t, int64(1), allMetrics["packets_dropped"])
	assert.Equal(t, int64(100), allMetrics["bytes_received"])
	assert.Equal(t, int64(200), allMetrics["bytes_sent"])
	assert.Equal(t, int64(1), allMetrics["stream_errors"])
	assert.Equal(t, int64(1), allMetrics["circuit_open_drops"])

	// Check stream pool metrics
	assert.Equal(t, int64(1), allMetrics["stream_pool_streams_created"])
	assert.Equal(t, int64(1), allMetrics["stream_pool_streams_closed"])
	assert.Equal(t, int64(1), allMetrics["stream_pool_streams_acquired"])
	assert.Equal(t, int64(1), allMetrics["stream_pool_streams_returned"])
	assert.Equal(t, int64(1), allMetrics["stream_pool_acquisition_failures"])
	assert.Equal(t, int64(1), allMetrics["stream_pool_unhealthy_streams"])

	// Check health metrics
	assert.Equal(t, int64(1), allMetrics["health_checks_performed"])
	assert.Equal(t, int64(1), allMetrics["health_healthy_streams"])
	assert.Equal(t, int64(1), allMetrics["health_unhealthy_streams"])
	assert.Equal(t, int64(1), allMetrics["health_streams_warmed"])
	assert.Equal(t, int64(1), allMetrics["health_warms_performed"])
	assert.Equal(t, int64(1), allMetrics["health_warm_failures"])
	assert.Equal(t, int64(1), allMetrics["health_peer_resets"])

	// Check circuit breaker metrics
	assert.Equal(t, int64(1), allMetrics["circuit_open_count"])
	assert.Equal(t, int64(1), allMetrics["circuit_close_count"])
	assert.Equal(t, int64(1), allMetrics["circuit_reset_count"])
	assert.Equal(t, int64(1), allMetrics["request_block_count"])
	assert.Equal(t, int64(1), allMetrics["request_allow_count"])
	assert.Equal(t, int64(5), allMetrics["active_breakers"])
}
