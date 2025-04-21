# VPN Packet Package

This package handles packet processing and routing for the VPN system. It is organized into several subpackages to improve modularity and maintainability while maintaining backward compatibility through re-exports in the root package.

## Package Structure

### Root Package (`packet`)

The root package provides backward compatibility by re-exporting types and functions from subpackages. New code should import from the specific subpackages directly.

### Subpackages

- **`dispatcher`**: Manages packet routing to appropriate workers
- **`worker`**: Handles packet processing for specific destinations
- **`errors`**: Defines error types and error handling utilities
- **`utils`**: Provides utility functions for packet analysis
- **`types`**: Contains common data structures used across packages
- **`testutil`**: Contains testing utilities and mocks

## Key Components

### Types (`types` package)

#### PacketInfo

The `PacketInfo` struct contains extracted source and destination IPs and ports from a packet:

```go
type PacketInfo struct {
    SrcIP    net.IP
    DstIP    net.IP
    SrcPort  *int    // Pointer to allow nil for non-TCP/UDP protocols
    DstPort  *int    // Pointer to allow nil for non-TCP/UDP protocols
    Protocol uint8   // IP protocol number (e.g., 6 for TCP, 17 for UDP)
}
```

#### QueuedPacket

The `QueuedPacket` struct represents a packet in the sending queue:

```go
type QueuedPacket struct {
    Ctx    context.Context  // Context for cancellation
    DestIP string           // Destination IP address
    Data   []byte           // Raw packet data
    DoneCh chan error       // Channel for completion signaling
}
```

### Dispatcher (`dispatcher` package)

The `Dispatcher` manages workers for different destination IP:Port combinations:

- Creates and manages workers for each unique destination
- Routes packets to the appropriate worker
- Cleans up idle workers to prevent resource leaks
- Integrates with resilience patterns for error handling

### Worker (`worker` package)

The `Worker` handles packets for a specific destination IP:Port combination:

- Processes packets sequentially for a single destination
- Manages connections to peers via the stream pool
- Tracks metrics like packet counts and error rates
- Implements retry and circuit breaking patterns

### Error Handling (`errors` package)

The package provides specialized error types:

- `NetworkError`: For network-related failures with context information
- `ServiceError`: For service-related failures with operation details
- Helper functions to categorize and handle different error types

### Utilities (`utils` package)

Provides functions for packet analysis:

- `ExtractIPAndPorts`: Extracts source and destination IPs and ports from packets
- Support for both IPv4 and IPv6 packet formats

## Usage Examples

### Using the Dispatcher

```go
// Create a dispatcher
dispatcher := packet.NewDispatcher(
    peerDiscovery,
    streamService,
    poolService,
    workerIdleTimeout,
    workerCleanupInterval,
    workerBufferSize,
    resilienceService,
)

// Start the dispatcher
dispatcher.Start()

// Dispatch a packet
dispatcher.DispatchPacket(ctx, syncKey, destIP, packetData)

// Dispatch with callback for completion notification
doneCh := make(chan error, 1)
dispatcher.DispatchPacketWithCallback(ctx, syncKey, destIP, packetData, doneCh)
err := <-doneCh

// Stop the dispatcher
dispatcher.Stop()
```

### Extracting Packet Information

```go
// Extract IP and port information from a packet
packetInfo, err := packet.ExtractIPAndPorts(packetData)
if err != nil {
    log.Errorf("Failed to extract packet info: %v", err)
    return
}

// Use the extracted information
destIP := packetInfo.DstIP.String()
if packetInfo.DstPort != nil {
    destPort := *packetInfo.DstPort
    log.Debugf("Packet to %s:%d", destIP, destPort)
}
```

### Error Handling

```go
// Check if an error is a network error
if packet.IsNetworkError(err) {
    // Handle network error
}

// Check if an error is temporary and can be retried
if packet.IsTemporaryError(err) {
    // Retry the operation
}

// Check if an error is due to a connection reset
if packet.IsConnectionResetError(err) {
    // Handle connection reset
}
```
