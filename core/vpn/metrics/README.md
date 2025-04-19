# VPN Metrics Package

This package provides metrics collection and reporting for the VPN system.

## Components

### VPNMetrics

The `VPNMetrics` struct tracks metrics for the VPN service, including:

- Packet counts (received, sent, dropped)
- Byte counts (received, sent)
- Stream errors
- Circuit breaker drops

### StreamPoolMetrics

The `StreamPoolMetrics` struct tracks metrics for a stream pool, including:

- Stream counts (created, closed, acquired, returned)
- Acquisition failures
- Unhealthy streams

### HealthMetrics

The `HealthMetrics` struct tracks metrics for stream health, including:

- Health checks performed
- Healthy and unhealthy streams
- Stream warming operations
- Warm failures

### MultiplexerMetrics

The `MultiplexerMetrics` struct tracks metrics for a stream multiplexer, including:

- Packet counts (sent, dropped)
- Byte counts (sent)
- Stream counts (created, closed)
- Stream errors
- Auto-scaling operations
- Latency measurements

## Usage

```go
// Create a VPN metrics collector
vpnMetrics := metrics.NewVPNMetrics()

// Increment metrics
vpnMetrics.IncrementPacketsReceived(100)
vpnMetrics.IncrementPacketsSent(200)
vpnMetrics.IncrementPacketsDropped()
vpnMetrics.IncrementStreamErrors()

// Get metrics
metrics := vpnMetrics.GetMetrics()
log.WithFields(logrus.Fields{
    "packets_received": metrics["packets_received"],
    "packets_sent":     metrics["packets_sent"],
    "packets_dropped":  metrics["packets_dropped"],
    "bytes_received":   metrics["bytes_received"],
    "bytes_sent":       metrics["bytes_sent"],
    "stream_errors":    metrics["stream_errors"],
}).Info("VPN metrics")
```
