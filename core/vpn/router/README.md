# Stream Router

The Stream Router is a high-performance packet routing system designed for VPN applications. It efficiently manages streams to peers while ensuring proper packet ordering and optimal resource utilization.

## Key Features

- **Minimal Stream Usage**: Starts with just 1 stream per peer and scales based on throughput demands
- **Dynamic Worker Scaling**: Automatically adjusts the number of workers based on load
- **Connection Affinity**: Ensures packets from the same connection are processed sequentially
- **Efficient Connection Management**: Uses sharded TTL cache with short expiration times for one-time connections
- **Throughput-Based Stream Scaling**: Adds or removes streams based on actual throughput needs

## Configuration

The Stream Router can be configured with the following parameters:

```go
type StreamRouterConfig struct {
    // Stream configuration
    MinStreamsPerPeer    int           // Default: 1
    MaxStreamsPerPeer    int           // Default: 10
    ThroughputThreshold  int64         // Packets per second per stream
    ScaleUpThreshold     float64       // Default: 0.8
    ScaleDownThreshold   float64       // Default: 0.3
    ScalingInterval      time.Duration // Default: 5 seconds

    // Worker pool configuration
    MinWorkers              int           // Default: 4
    MaxWorkers              int           // Default: 32
    InitialWorkers          int           // Default: 8
    WorkerQueueSize         int           // Default: 1000
    WorkerScaleInterval     time.Duration // Default: 5 seconds
    WorkerScaleUpThreshold  float64       // Default: 0.75
    WorkerScaleDownThreshold float64      // Default: 0.25

    // Connection cache configuration
    ConnectionTTL        time.Duration // Default: 5 seconds
    CleanupInterval      time.Duration // Default: 5 seconds
    CacheShardCount      int           // Default: 16
}
```

## Usage

```go
// Create dependencies
streamPool := createStreamPool()
peerDiscovery := createPeerDiscovery()

// Create router with default config
config := router.DefaultStreamRouterConfig()
streamRouter := router.NewStreamRouter(config, streamPool, peerDiscovery)
defer streamRouter.Shutdown()

// Dispatch a packet
err := streamRouter.DispatchPacket(ctx, packetData)
if err != nil {
    // Handle error
}
```

## Architecture

The Stream Router consists of several key components:

1. **Connection Management**: Uses a sharded TTL cache to efficiently manage connection routes
2. **Worker Pool**: Dynamically scales workers based on load
3. **Stream Management**: Maintains minimal streams per peer and scales based on throughput
4. **Metrics Collection**: Tracks throughput and utilization for scaling decisions

### Packet Flow

1. Packet arrives from TUN device
2. Connection key is created from source and destination information
3. Connection is assigned to a specific worker (consistent hashing)
4. Worker processes the packet using the appropriate stream
5. Metrics are updated for scaling decisions

### Dynamic Scaling

- **Worker Scaling**: Based on queue utilization
- **Stream Scaling**: Based on throughput per stream

## Performance Considerations

- **Connection Cache**: Uses short TTL (30 seconds) and frequent cleanup (10 seconds) to handle one-time connections
- **Sharded Cache**: Reduces lock contention for high-throughput scenarios
- **Worker Pool**: Efficiently distributes processing across available CPU cores
- **Stream Reuse**: Minimizes the number of streams to stay under libp2p limits

## Integration

The Stream Router integrates with:

- **Stream Pool**: For managing streams to peers
- **Peer Discovery**: For resolving destination IPs to peer IDs

See the example directory for integration examples.
