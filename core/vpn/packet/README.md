# VPN Packet Package

This package handles packet processing and routing for the VPN system.

## Components

### PacketInfo

The `PacketInfo` struct contains extracted source and destination IPs and ports from a packet.

### QueuedPacket

The `QueuedPacket` struct represents a packet in the sending queue, including:

- Context for cancellation
- Destination IP
- Packet data
- Done channel for signaling completion

### Worker

The `Worker` handles packets for a specific destination IP:Port combination, including:

- Synchronization key (IP:Port or just IP)
- Destination IP
- Peer ID
- Stream to the peer
- Packet channel
- Last activity time
- Packet and error counts

### Dispatcher

The `Dispatcher` manages workers for different destination IP:Port combinations, including:

- Worker creation and cleanup
- Packet routing to the appropriate worker
- Worker idle timeout management

### Utils

The `utils.go` file provides functions for extracting IP and port information from packets, supporting both IPv4 and IPv6.

## Usage

```go
// Create a dispatcher
dispatcher := packet.NewDispatcher(
    peerDiscovery,
    streamService,
    workerIdleTimeout,
    workerCleanupInterval,
    workerBufferSize,
)

// Start the dispatcher
dispatcher.Start()

// Dispatch a packet
dispatcher.DispatchPacket(ctx, syncKey, destIP, packetData)

// Stop the dispatcher
dispatcher.Stop()
```
