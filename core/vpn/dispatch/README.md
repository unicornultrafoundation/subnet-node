# VPN Dispatch Package

This package implements an efficient packet forwarding system using libp2p, designed to handle high-throughput VPN traffic with optimal concurrency and resource utilization.

## Design Principles

1. **Stream-Assigned Channels**: Each stream has its own dedicated channel for packet writing, eliminating contention between connections.

2. **PeerID-Specific Stream Pools**: Streams are organized by PeerID, allowing concurrent processing of packets destined for different peers.

3. **Connection Key-Based Packet Assignment**: Packets with the same connection key (sourcePort:destinationIP:destinationPort) are processed sequentially, ensuring proper packet ordering.

4. **No Mutex for Stream Access**: Stream access is managed through dedicated channels, eliminating the need for mutex locking and improving concurrency.

5. **Dynamic Stream Scaling**: The system can scale streams based on traffic load, adding more streams when needed and removing idle ones.

## Package Structure

### Root Package (`dispatch`)

- `dispatcher.go`: Main dispatcher that routes packets to appropriate streams
- `interfaces.go`: Defines interfaces for the dispatch components

### Subpackages

- **`types`**: Contains common data structures and error types
- **`pool`**: Manages stream pools and connection-to-stream mapping

## Key Components

### Dispatcher

The `Dispatcher` is the entry point for packet processing:

- Routes packets to the appropriate stream based on destination peer ID and connection key
- Manages stream pools for different peers
- Provides metrics and monitoring capabilities

### StreamPool

The `StreamPool` manages streams for communication with peers:

- Maintains a pool of streams for each peer
- Provides load balancing across streams
- Handles stream health monitoring and cleanup
- Dynamically scales the number of streams based on traffic load

### StreamChannel

The `StreamChannel` manages the communication channel for a stream:

- Buffers packets for sending to ensure non-blocking operation
- Handles packet sending through the assigned stream
- Implements resilience patterns for error handling
- Ensures sequential processing of packets

## Flow Overview

1. A packet arrives at the `Dispatcher`
2. The dispatcher determines the destination peer ID and connection key
3. The dispatcher gets a stream from the `StreamPool` for the peer ID
4. The dispatcher sends the packet to the stream's channel
5. The `StreamChannel` processes the packet and sends it through the stream
6. The stream pool ensures packets with the same connection key use the same stream

## Usage Example

```go
// Create a dispatcher configuration
config := &dispatch.Config{
    MaxStreamsPerPeer:     10,
    StreamIdleTimeout:     5 * time.Minute,
    StreamCleanupInterval: 1 * time.Minute,
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
connKey := types.ConnectionKey(fmt.Sprintf("%s:%s:%s", sourcePort, destIP, destPort))
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
- Each stream operates independently, reducing contention
- Stream channels provide a buffer for packet processing, preventing backpressure
- Idle streams are automatically cleaned up to conserve resources
- Load balancing ensures even distribution of traffic across streams
- Dynamic stream scaling ensures optimal resource utilization based on traffic load
