# VPN Service Documentation

## Overview

The VPN (Virtual Private Network) service in the Subnet Node provides secure, peer-to-peer networking capabilities between nodes in the network. It creates an encrypted overlay network that allows nodes to communicate securely regardless of their physical location or network configuration.

The VPN service leverages libp2p for peer-to-peer communication and establishes TUN interfaces on participating nodes to route traffic through the secure overlay network.

## Architecture

The VPN service follows a modular architecture with several key components:

### Core Components

1. **Service**: The main entry point that coordinates all VPN functionality.
2. **PeerDiscovery**: Handles peer discovery and mapping between virtual IPs and peer IDs.
3. **TUNService**: Manages the TUN interface for capturing and injecting network packets.
4. **ClientService**: Handles outgoing VPN traffic from the local node.
5. **ServerService**: Processes incoming VPN traffic from remote peers.
6. **Dispatcher**: Routes packets to the appropriate worker based on destination.
7. **StreamService**: Manages libp2p streams for communication between peers.
8. **CircuitBreakerManager**: Provides fault tolerance through circuit breaker pattern.
9. **RetryManager**: Handles retries for failed operations with exponential backoff.
10. **MetricsService**: Collects and reports performance metrics.

### Data Flow

1. **Outgoing Traffic**:
   - Packets are captured by the TUN interface
   - ClientService processes the packets
   - Dispatcher routes packets to the appropriate peer
   - StreamService establishes and manages connections to peers
   - Packets are sent over encrypted libp2p streams

2. **Incoming Traffic**:
   - Packets are received over libp2p streams
   - ServerService processes the packets
   - Packets are injected into the TUN interface
   - Operating system routes packets to the appropriate local application

## Configuration

The VPN service is highly configurable through the configuration service. Key configuration options include:

### Basic Settings

- `vpn.enable`: Enable or disable the VPN service
- `vpn.mtu`: Maximum Transmission Unit for the VPN
- `vpn.virtual_ip`: Virtual IP address for this node
- `vpn.subnet`: Subnet mask for the VPN network
- `vpn.routes`: Routes to be added to the routing table

### Security Settings

- `vpn.unallowed_ports`: Ports that are blocked from VPN traffic

### Performance Settings

- `vpn.worker_idle_timeout`: Timeout for idle workers
- `vpn.worker_buffer_size`: Buffer size for packet workers
- `vpn.max_workers`: Maximum number of packet workers
- `vpn.worker_cleanup_interval`: Interval for cleaning up idle workers

### Stream Settings


- `vpn.min_streams_per_peer`: Minimum number of streams per peer
- `vpn.stream_idle_timeout`: Timeout for idle streams
- `vpn.cleanup_interval`: Interval for cleaning up idle streams

### Resilience Settings

- `vpn.circuit_breaker_failure_threshold`: Number of failures before circuit opens
- `vpn.circuit_breaker_reset_timeout`: Time before attempting to close circuit
- `vpn.circuit_breaker_success_threshold`: Number of successes to close circuit
- `vpn.retry_max_attempts`: Maximum retry attempts
- `vpn.retry_initial_interval`: Initial retry interval
- `vpn.retry_max_interval`: Maximum retry interval

## Usage

### Starting the VPN Service

The VPN service is typically started as part of the node initialization process:

```go
// Create a new VPN service
vpnService := vpn.New(cfg, peerHost, dht, accountService)

// Start the VPN service
err := vpnService.Start(ctx)
if err != nil {
    log.Errorf("Failed to start VPN service: %v", err)
    return err
}
```

### Stopping the VPN Service

The VPN service should be stopped gracefully when the node is shutting down:

```go
// Stop the VPN service
err := vpnService.Stop()
if err != nil {
    log.Errorf("Failed to stop VPN service: %v", err)
    return err
}
```

### Monitoring VPN Metrics

The VPN service provides metrics for monitoring performance:

```go
// Get VPN metrics
metrics := vpnService.GetMetrics()

// Log metrics
log.WithFields(logrus.Fields{
    "packets_received": metrics["packets_received"],
    "packets_sent":     metrics["packets_sent"],
    "packets_dropped":  metrics["packets_dropped"],
    "bytes_received":   metrics["bytes_received"],
    "bytes_sent":       metrics["bytes_sent"],
    "stream_errors":    metrics["stream_errors"],
}).Info("VPN metrics")
```

## Component Details

### PeerDiscovery

The PeerDiscovery component is responsible for:
- Mapping between virtual IPs and peer IDs
- Storing and retrieving peer information from DHT
- Verifying virtual IP registrations
- Providing peer lookup services

### TUNService

The TUNService component:
- Creates and manages the TUN interface
- Configures the interface with the virtual IP and routes
- Provides read/write access to the TUN interface

### ClientService

The ClientService component:
- Reads packets from the TUN interface
- Processes outgoing packets
- Dispatches packets to the appropriate peer

### ServerService

The ServerService component:
- Handles incoming packets from peers
- Processes packets and injects them into the TUN interface
- Enforces security policies (e.g., blocked ports)

### Dispatcher

The Dispatcher component:
- Routes packets to the appropriate worker based on destination
- Manages worker lifecycle
- Provides packet buffering and flow control

### StreamService

The StreamService component:
- Creates and manages libp2p streams
- Provides stream pooling for performance
- Implements stream health checking
- Supports stream multiplexing

### CircuitBreakerManager

The CircuitBreakerManager component:
- Implements the circuit breaker pattern for fault tolerance
- Prevents cascading failures
- Provides automatic recovery

### RetryManager

The RetryManager component:
- Handles retries for failed operations
- Implements exponential backoff
- Limits retry attempts

### MetricsService

The MetricsService component:
- Collects performance metrics
- Provides centralized metrics reporting
- Supports multiple metric types (counters, gauges, etc.)

## Troubleshooting

### Common Issues

1. **VPN Service Not Starting**
   - Check if the VPN service is enabled in the configuration
   - Verify that the node has at least one peer connected
   - Check for errors in the logs

2. **No Connectivity Through VPN**
   - Verify that the TUN interface is properly configured
   - Check if the routes are correctly set up
   - Ensure that the peer discovery is working correctly

3. **Poor Performance**
   - Adjust the MTU setting
   - Increase the worker buffer size
   - Tune the number of streams per peer

4. **High Packet Loss**
   - Check for circuit breaker activations
   - Verify network connectivity between peers
   - Adjust retry settings

### Logging

The VPN service uses structured logging with the following fields:
- `service`: The service name (e.g., "vpn", "vpn-discovery", "vpn-packet")
- `packets_received`: Number of packets received
- `packets_sent`: Number of packets sent
- `packets_dropped`: Number of packets dropped
- `bytes_received`: Number of bytes received
- `bytes_sent`: Number of bytes sent
- `stream_errors`: Number of stream errors

### Metrics

The VPN service provides the following metrics:
- `packets_received`: Number of packets received
- `packets_sent`: Number of packets sent
- `packets_dropped`: Number of packets dropped
- `bytes_received`: Number of bytes received
- `bytes_sent`: Number of bytes sent
- `stream_errors`: Number of stream errors
- `circuit_open_count`: Number of times circuits opened
- `circuit_close_count`: Number of times circuits closed
- `circuit_reset_count`: Number of times circuits reset
- `request_block_count`: Number of requests blocked by circuit breakers
- `request_allow_count`: Number of requests allowed by circuit breakers
- `active_breakers`: Number of active circuit breakers

## Security Considerations

The VPN service provides secure communication between peers, but there are several security considerations to keep in mind:

1. **Virtual IP Assignment**: Virtual IPs should be unique and properly registered to prevent IP conflicts.

2. **Port Blocking**: Use the `vpn.unallowed_ports` configuration to block sensitive ports.

3. **Peer Authentication**: The VPN service relies on libp2p for peer authentication. Ensure that proper peer authentication is configured.

4. **Traffic Encryption**: All VPN traffic is encrypted using libp2p's transport encryption.

5. **Access Control**: The VPN service does not implement application-level access control. Additional access control mechanisms should be implemented if needed.

## Performance Tuning

To optimize VPN performance, consider the following tuning options:

1. **MTU Optimization**: Adjust the MTU setting to match your network conditions.

2. **Stream Pooling**: Configure the minimum streams per peer (`min_streams_per_peer`) based on expected traffic patterns.

3. **Worker Settings**: Adjust worker buffer size and maximum workers based on system resources.

4. **Circuit Breaker Tuning**: Configure circuit breaker thresholds based on network reliability.

5. **Retry Settings**: Adjust retry settings based on network latency and reliability.

## Implementation Details

### Adapter Pattern

The VPN service uses the adapter pattern extensively to:
- Break circular dependencies
- Provide clean interfaces between components
- Enable easier testing and mocking

### Resilience Patterns

The VPN service implements several resilience patterns:
- Circuit Breaker: Prevents cascading failures
- Retry with Exponential Backoff: Handles transient failures
- Stream Health Checking: Detects and recovers from unhealthy streams
- Stream Pooling: Provides connection reuse and failover

### Concurrency Model

The VPN service uses a combination of:
- Goroutines for concurrent processing
- Channels for communication between components
- Mutexes for protecting shared state
- Atomic operations for high-performance counters

## Conclusion

The VPN service provides a secure, peer-to-peer networking layer for the Subnet Node. It leverages libp2p for peer-to-peer communication and implements several resilience patterns to ensure reliable operation. The service is highly configurable and provides comprehensive metrics for monitoring performance.
