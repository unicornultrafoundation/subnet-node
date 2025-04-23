# VPN Dispatch Package

This package implements an efficient packet forwarding system using libp2p, designed to handle high-throughput VPN traffic with optimal concurrency and resource utilization.

## Design Principles

1. **Stream-Assigned Channels**: Each stream has its own dedicated channel for packet writing, eliminating contention between workers.

2. **PeerID-Specific Worker Pools**: Workers are organized by PeerID, allowing concurrent processing of packets destined for different peers.

3. **Connection Key-Based Packet Assignment**: Packets with the same connection key (sourcePort:destinationIP:destinationPort) are processed by the same worker, ensuring sequential processing when needed.

4. **No Mutex for Stream Access**: Stream access is managed through dedicated channels, eliminating the need for mutex locking and improving concurrency.

5. **Dynamic Worker and Stream Scaling**: The system can scale workers and streams based on traffic load.

## Package Structure

### Root Package (`dispatch`)

- `dispatcher.go`: Main dispatcher that routes packets to appropriate worker pools
- `interfaces.go`: Defines interfaces for the dispatch components

### Subpackages

- **`types`**: Contains common data structures and error types
- **`pool`**: Manages stream pools and connection-to-stream mapping
- **`worker`**: Handles packet processing workers and worker pools

## Key Components

### Dispatcher

The `Dispatcher` is the entry point for packet processing:

- Routes packets to the appropriate worker pool based on destination peer ID
- Manages worker pools for different peers
- Provides metrics and monitoring capabilities

### Stream Pool

The `StreamPool` manages streams for communication with peers:

- Maintains a pool of streams for each peer
- Provides load balancing across streams
- Handles stream health monitoring and cleanup

### Stream Manager

The `StreamManager` manages the mapping between connections and streams:

- Assigns streams to connection keys
- Ensures packets with the same connection key use the same stream
- Provides metrics for connection and stream usage

### Worker Pool

The `WorkerPool` manages workers for a specific peer:

- Creates and manages workers for different connection keys
- Routes packets to the appropriate worker
- Cleans up idle workers to prevent resource leaks

### Worker

The `Worker` processes packets for a specific connection key:

- Ensures sequential processing of packets with the same connection key
- Handles packet sending through the assigned stream
- Implements resilience patterns for error handling

## Flow Overview

1. A packet arrives at the `Dispatcher`
2. The dispatcher determines the destination peer ID and routes the packet to the appropriate `WorkerPool`
3. The worker pool identifies or creates a `Worker` for the connection key
4. The worker gets a stream from the `StreamManager` and sends the packet
5. The stream manager ensures packets with the same connection key use the same stream

## Usage Example

```go
// Create a dispatcher configuration
config := &dispatch.DispatcherConfig{
    MaxStreamsPerPeer:     10,
    StreamIdleTimeout:     5 * time.Minute,
    StreamCleanupInterval: 1 * time.Minute,
    WorkerIdleTimeout:     300, // seconds
    WorkerCleanupInterval: 1 * time.Minute,
    WorkerBufferSize:      100,
    PacketBufferSize:      100,
}

// Create a dispatcher
dispatcher := dispatch.NewDispatcher(
    peerDiscovery,
    streamService,
    config,
    resilienceService,
)

// Start the dispatcher
dispatcher.Start()

// Dispatch a packet
connKey := types.FormatConnectionKey(sourcePort, destIP, destPort)
err := dispatcher.DispatchPacket(ctx, connKey, destIP, packetData)

// Dispatch with callback for completion notification
doneCh := make(chan error, 1)
dispatcher.DispatchPacketWithCallback(ctx, connKey, destIP, packetData, doneCh)
err := <-doneCh

// Get metrics
metrics := dispatcher.GetMetrics()

// Stop the dispatcher
dispatcher.Stop()
```

## Performance Considerations

- The system is designed to maximize concurrency while ensuring sequential processing when needed
- Each worker operates independently, reducing contention
- Stream channels provide a buffer for packet processing, preventing backpressure
- Idle workers and streams are automatically cleaned up to conserve resources
- Load balancing ensures even distribution of traffic across streams
