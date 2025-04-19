# VPN Stream Package

This package manages P2P streams and multiplexing for the VPN system.

## Components

### VPNStream

The `VPNStream` interface defines the methods required for a VPN stream, including:

- Reading and writing data
- Closing and resetting the stream
- Setting deadlines

### Service

The `Service` interface defines the methods for creating and managing streams, including:

- Creating new VPN streams to peers

### PoolService

The `PoolService` interface defines the methods for stream pooling, including:

- Getting streams from the pool
- Releasing streams back to the pool

### HealthService

The `HealthService` interface defines the methods for stream health monitoring, including:

- Starting and stopping the health checker
- Starting and stopping the stream warmer
- Getting health metrics

### MultiplexService

The `MultiplexService` interface defines the methods for stream multiplexing, including:

- Starting and stopping the multiplexer
- Sending packets using the multiplexer
- Getting multiplexer metrics

### MultiplexerMetrics

The `MultiplexerMetrics` struct contains metrics for a multiplexer, including:

- Packet counts (sent, dropped)
- Byte counts (sent)
- Stream counts (created, closed)
- Auto-scaling operations
- Latency measurements

## Usage

```go
// Create a new VPN stream
vpnStream, err := streamService.CreateNewVPNStream(ctx, peerID)
if err != nil {
    return nil, fmt.Errorf("failed to create P2P stream: %v", err)
}

// Get a stream from the pool
vpnStream, err := streamPoolService.GetStream(ctx, peerID)
if err != nil {
    return nil, fmt.Errorf("failed to get stream from pool: %v", err)
}

// Release a stream back to the pool
streamPoolService.ReleaseStream(peerID, vpnStream, true)

// Send a packet using the multiplexer
err := streamMultiplexService.SendPacketMultiplexed(ctx, peerID, packet)
if err != nil {
    return fmt.Errorf("failed to send packet: %v", err)
}
```
