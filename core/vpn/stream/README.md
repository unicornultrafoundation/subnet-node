# VPN Stream Package

This package manages P2P streams for the VPN system, focusing on stream pooling and health monitoring.

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

### MetricsService

The `MetricsService` interface defines the methods for collecting metrics, including:

- Getting stream pool metrics
- Getting health metrics

### LifecycleService

The `LifecycleService` interface defines the methods for managing service lifecycle, including:

- Starting the service
- Stopping the service

## Architecture

The stream package uses a layered architecture:

1. **StreamService**: The main service that coordinates all stream-related operations
2. **StreamPoolManager**: Manages pools of streams for different peers
3. **HealthChecker**: Monitors stream health and closes unhealthy streams
4. **StreamWarmer**: Ensures a minimum number of streams are available for each peer

## Usage

```go
// Create a stream service with configuration
streamService := stream.CreateStreamService(baseStreamService, config)

// Start the stream service
streamService.Start()

// Create a new VPN stream directly
vpnStream, err := streamService.CreateNewVPNStream(ctx, peerID)
if err != nil {
    return nil, fmt.Errorf("failed to create P2P stream: %v", err)
}

// Get a stream from the pool
vpnStream, err := streamService.GetStream(ctx, peerID)
if err != nil {
    return nil, fmt.Errorf("failed to get stream from pool: %v", err)
}

// Release a stream back to the pool
streamService.ReleaseStream(peerID, vpnStream, true)

// Get metrics
healthMetrics := streamService.GetHealthMetrics()
poolMetrics := streamService.GetStreamPoolMetrics()

// Stop the stream service when done
streamService.Stop()
```
