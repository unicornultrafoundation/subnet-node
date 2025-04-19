package metrics

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVPNMetrics(t *testing.T) {
	metrics := NewVPNMetrics()

	// Test initial values
	initialMetrics := metrics.GetMetrics()
	assert.Equal(t, int64(0), initialMetrics["packets_received"])
	assert.Equal(t, int64(0), initialMetrics["packets_sent"])
	assert.Equal(t, int64(0), initialMetrics["packets_dropped"])
	assert.Equal(t, int64(0), initialMetrics["bytes_received"])
	assert.Equal(t, int64(0), initialMetrics["bytes_sent"])
	assert.Equal(t, int64(0), initialMetrics["stream_errors"])
	assert.Equal(t, int64(0), initialMetrics["circuit_open_drops"])

	// Test incrementing metrics
	metrics.IncrementPacketsReceived(100)
	metrics.IncrementPacketsSent(200)
	metrics.IncrementPacketsDropped()
	metrics.IncrementStreamErrors()
	metrics.IncrementCircuitOpenDrops()

	// Check updated values
	updatedMetrics := metrics.GetMetrics()
	assert.Equal(t, int64(1), updatedMetrics["packets_received"])
	assert.Equal(t, int64(1), updatedMetrics["packets_sent"])
	assert.Equal(t, int64(1), updatedMetrics["packets_dropped"])
	assert.Equal(t, int64(100), updatedMetrics["bytes_received"])
	assert.Equal(t, int64(200), updatedMetrics["bytes_sent"])
	assert.Equal(t, int64(1), updatedMetrics["stream_errors"])
	assert.Equal(t, int64(1), updatedMetrics["circuit_open_drops"])

	// Test concurrent access
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			metrics.IncrementPacketsReceived(10)
			metrics.IncrementPacketsSent(20)
			metrics.IncrementPacketsDropped()
		}()
	}
	wg.Wait()

	// Check final values
	finalMetrics := metrics.GetMetrics()
	assert.Equal(t, int64(101), finalMetrics["packets_received"])
	assert.Equal(t, int64(101), finalMetrics["packets_sent"])
	assert.Equal(t, int64(101), finalMetrics["packets_dropped"])
	assert.Equal(t, int64(1100), finalMetrics["bytes_received"])
	assert.Equal(t, int64(2200), finalMetrics["bytes_sent"])
}

func TestStreamPoolMetrics(t *testing.T) {
	metrics := NewStreamPoolMetrics()

	// Test initial values
	initialMetrics := metrics.GetMetrics()
	assert.Equal(t, int64(0), initialMetrics["streams_created"])
	assert.Equal(t, int64(0), initialMetrics["streams_closed"])
	assert.Equal(t, int64(0), initialMetrics["streams_acquired"])
	assert.Equal(t, int64(0), initialMetrics["streams_returned"])
	assert.Equal(t, int64(0), initialMetrics["acquisition_failures"])
	assert.Equal(t, int64(0), initialMetrics["unhealthy_streams"])

	// Test incrementing metrics
	metrics.IncrementStreamsCreated()
	metrics.IncrementStreamsClosed()
	metrics.IncrementStreamsAcquired()
	metrics.IncrementStreamsReturned()
	metrics.IncrementAcquisitionFailures()
	metrics.IncrementUnhealthyStreams()

	// Check updated values
	updatedMetrics := metrics.GetMetrics()
	assert.Equal(t, int64(1), updatedMetrics["streams_created"])
	assert.Equal(t, int64(1), updatedMetrics["streams_closed"])
	assert.Equal(t, int64(1), updatedMetrics["streams_acquired"])
	assert.Equal(t, int64(1), updatedMetrics["streams_returned"])
	assert.Equal(t, int64(1), updatedMetrics["acquisition_failures"])
	assert.Equal(t, int64(1), updatedMetrics["unhealthy_streams"])

	// Test concurrent access
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			metrics.IncrementStreamsCreated()
			metrics.IncrementStreamsClosed()
		}()
	}
	wg.Wait()

	// Check final values
	finalMetrics := metrics.GetMetrics()
	assert.Equal(t, int64(101), finalMetrics["streams_created"])
	assert.Equal(t, int64(101), finalMetrics["streams_closed"])
}

func TestHealthMetrics(t *testing.T) {
	metrics := NewHealthMetrics()

	// Test initial values
	initialMetrics := metrics.GetMetrics()
	assert.Equal(t, int64(0), initialMetrics["checks_performed"])
	assert.Equal(t, int64(0), initialMetrics["healthy_streams"])
	assert.Equal(t, int64(0), initialMetrics["unhealthy_streams"])
	assert.Equal(t, int64(0), initialMetrics["streams_warmed"])
	assert.Equal(t, int64(0), initialMetrics["warms_performed"])
	assert.Equal(t, int64(0), initialMetrics["warm_failures"])

	// Test incrementing metrics
	metrics.IncrementChecksPerformed()
	metrics.IncrementHealthyStreams()
	metrics.IncrementUnhealthyStreams()
	metrics.IncrementStreamsWarmed()
	metrics.IncrementWarmsPerformed()
	metrics.IncrementWarmFailures()

	// Check updated values
	updatedMetrics := metrics.GetMetrics()
	assert.Equal(t, int64(1), updatedMetrics["checks_performed"])
	assert.Equal(t, int64(1), updatedMetrics["healthy_streams"])
	assert.Equal(t, int64(1), updatedMetrics["unhealthy_streams"])
	assert.Equal(t, int64(1), updatedMetrics["streams_warmed"])
	assert.Equal(t, int64(1), updatedMetrics["warms_performed"])
	assert.Equal(t, int64(1), updatedMetrics["warm_failures"])
}

func TestMultiplexerMetrics(t *testing.T) {
	metrics := NewMultiplexerMetrics()

	// Test initial values
	initialMetrics := metrics.GetMetrics()
	assert.Equal(t, int64(0), initialMetrics["packets_sent"])
	assert.Equal(t, int64(0), initialMetrics["packets_dropped"])
	assert.Equal(t, int64(0), initialMetrics["bytes_sent"])
	assert.Equal(t, int64(0), initialMetrics["stream_errors"])
	assert.Equal(t, int64(0), initialMetrics["streams_created"])
	assert.Equal(t, int64(0), initialMetrics["streams_closed"])
	assert.Equal(t, int64(0), initialMetrics["scale_up_operations"])
	assert.Equal(t, int64(0), initialMetrics["scale_down_operations"])
	assert.Equal(t, int64(0), initialMetrics["avg_latency"])
	assert.Equal(t, int64(0), initialMetrics["latency_measurements"])

	// Test incrementing metrics
	metrics.IncrementPacketsSent(100)
	metrics.IncrementPacketsDropped()
	metrics.IncrementStreamErrors()
	metrics.IncrementStreamsCreated()
	metrics.IncrementStreamsClosed()
	metrics.IncrementScaleUpOperations()
	metrics.IncrementScaleDownOperations()
	metrics.UpdateLatency(50)

	// Check updated values
	updatedMetrics := metrics.GetMetrics()
	assert.Equal(t, int64(1), updatedMetrics["packets_sent"])
	assert.Equal(t, int64(1), updatedMetrics["packets_dropped"])
	assert.Equal(t, int64(100), updatedMetrics["bytes_sent"])
	assert.Equal(t, int64(1), updatedMetrics["stream_errors"])
	assert.Equal(t, int64(1), updatedMetrics["streams_created"])
	assert.Equal(t, int64(1), updatedMetrics["streams_closed"])
	assert.Equal(t, int64(1), updatedMetrics["scale_up_operations"])
	assert.Equal(t, int64(1), updatedMetrics["scale_down_operations"])
	assert.Equal(t, int64(50), updatedMetrics["avg_latency"])
	assert.Equal(t, int64(1), updatedMetrics["latency_measurements"])

	// Test latency averaging
	metrics.UpdateLatency(150)
	finalMetrics := metrics.GetMetrics()
	assert.Equal(t, int64(100), finalMetrics["avg_latency"]) // (50 + 150) / 2 = 100
	assert.Equal(t, int64(2), finalMetrics["latency_measurements"])
}
