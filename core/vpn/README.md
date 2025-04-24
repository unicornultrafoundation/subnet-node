# VPN Service

The VPN (Virtual Private Network) service provides secure, peer-to-peer networking capabilities between nodes in the Subnet Network. It creates an encrypted overlay network that allows nodes to communicate securely regardless of their physical location or network configuration.

## Overview

The VPN service leverages libp2p for peer-to-peer communication and establishes TUN interfaces on participating nodes to route traffic through the secure overlay network.

## Features

- **Secure P2P Communication**: End-to-end encrypted communication between nodes
- **Virtual IP Addressing**: Assigns virtual IP addresses to nodes for routing
- **Automatic Peer Discovery**: Uses DHT for peer discovery and mapping
- **Resilient Connections**: Implements circuit breakers and retries for reliability
- **Performance Metrics**: Comprehensive metrics for monitoring
- **Packet Synchronization**: Ensures strict sequential processing of packets
- **Stream Pooling**: Efficient reuse of P2P streams with automatic scaling
- **Stream Health Monitoring**: Detection and replacement of unhealthy streams
- **Connection-Based Routing**: Ensures packets with the same connection are processed sequentially

## Package Structure

The VPN package is organized into several subpackages:

- `api`: Defines the interfaces for the VPN system
- `config`: Contains configuration structures and validation
- `discovery`: Handles peer discovery and mapping
- `dispatch`: Manages packet dispatching and stream pooling
- `metrics`: Provides metrics collection and reporting
- `network`: Manages network interfaces and connections
- `resilience`: Implements resilience patterns like circuit breakers and retries
- `utils`: Provides common utilities used across the VPN service
- `validator`: Validates VPN configuration and settings
- `testutil`: Provides testing utilities for the VPN system

## Architecture

The VPN service follows a modular architecture with several key components:

```
┌─────────────────────────────────────────────────────────────────┐
│                         VPN Service                             │
├─────────────┬─────────────┬─────────────┬─────────────┬─────────┤
│ PeerDiscovery│  TUNService │ClientService│ServerService│ Metrics │
├─────────────┼─────────────┼─────────────┼─────────────┼─────────┤
│  Dispatcher │ StreamPool  │CircuitBreaker│RetryManager │ Utils   │
└─────────────┴─────────────┴─────────────┴─────────────┴─────────┘
```

### Core Components

- **Service**: Main entry point that coordinates all VPN functionality
- **PeerDiscovery**: Handles peer discovery and mapping between virtual IPs and peer IDs
- **TUNService**: Manages the TUN interface for capturing and injecting network packets
- **ClientService**: Handles outgoing VPN traffic from the local node
- **ServerService**: Processes incoming VPN traffic from remote peers
- **Dispatcher**: Routes packets to the appropriate stream based on connection key
- **StreamPool**: Manages pools of libp2p streams for efficient communication between peers
- **CircuitBreakerManager**: Provides fault tolerance through circuit breaker pattern
- **RetryManager**: Handles retries for failed operations with exponential backoff
- **MetricsService**: Collects and reports performance metrics

### Configuration

The VPN configuration is defined in the `config` package. It includes settings for:

- Network configuration (MTU, virtual IP, subnet, routes)
- Stream pool settings (max streams per peer, idle timeout, cleanup interval)
- Packet buffer settings (buffer size for stream channels)
- Circuit breaker settings (failure threshold, reset timeout, success threshold)
- Retry settings (max attempts, initial interval, max interval)

### Packet Processing

Packets are processed by stream channels that handle specific connection keys (IP:Port combinations). The dispatcher routes packets to the appropriate stream channel based on the connection key, ensuring that packets with the same connection key are processed sequentially.

```
TUN Interface -> ClientService -> Dispatcher -> Stream Channel -> P2P Stream -> Peer
```

### Stream Management

Streams are managed by a pool that maintains connections to peers. The pool dynamically scales the number of streams per peer based on traffic load, up to a configurable maximum. Idle streams are automatically cleaned up after a configurable timeout period. Each stream is monitored for health and replaced if it becomes unhealthy.

### Resilience

The VPN system includes several resilience patterns:

- Circuit breakers to prevent cascading failures
- Retries with exponential backoff for transient failures
- Stream health checks to detect and replace unhealthy streams

### Metrics

The VPN system collects metrics for monitoring and debugging, including:

- Packet counts (received, sent, dropped)
- Byte counts (received, sent)
- Stream counts (created, closed, errors)
- Stream pool metrics (streams per peer, active streams)
- Circuit breaker states and counts
- Retry operation metrics (attempts, successes, failures)

## Usage

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

// Later, when shutting down:
vpnService.Stop()
```

The service will set up a TUN interface, handle P2P traffic, and listen for packets from the TUN interface.

## Configuration Example

The VPN service is configured through the node's configuration file. Key configuration options include:

```yaml
vpn:
  enable: true
  mtu: 1400
  virtual_ip: "10.0.0.1"
  subnet: "10.0.0.0/8"
  routes:
    - "10.0.0.0/8"
  unallowed_ports:
    - 22
    - 3306

  # Stream settings
  max_streams_per_peer: 10
  stream_idle_timeout: 300      # 5 minutes
  cleanup_interval: 60          # 1 minute
  packet_buffer_size: 100       # Size of packet buffer for each stream

  # Circuit breaker settings
  circuit_breaker_failure_threshold: 5
  circuit_breaker_reset_timeout: 60    # 1 minute
  circuit_breaker_success_threshold: 2

  # Retry settings
  retry_max_attempts: 5
  retry_initial_interval: 1      # 1 second
  retry_max_interval: 30         # 30 seconds
```

## Testing

The VPN service includes comprehensive unit and integration tests. Each package includes tests to ensure correctness and reliability. The tests cover:

- Configuration validation
- Packet processing
- Stream pool management
- Connection-based routing
- Circuit breaker behavior
- Retry logic
- Metrics collection
- Resilience under network failures

```bash
# Run all VPN tests
go test -v ./core/vpn/...

# Run specific component tests
go test -v ./core/vpn/discovery/...
go test -v ./core/vpn/dispatch/...
```

For more detailed information about testing, see the [TESTING.md](./TESTING.md) file.

## Performance Tuning

To optimize VPN performance, consider the following tuning options:

1. **MTU Optimization**: Adjust the MTU setting to match your network conditions
2. **Stream Pooling**: Configure the maximum streams per peer based on expected traffic patterns
3. **Packet Buffer Size**: Adjust the packet buffer size based on expected packet rates
4. **Retry Settings**: Configure retry parameters based on network reliability
5. **Circuit Breaker Settings**: Tune circuit breaker parameters for your failure patterns

## Security Considerations

The VPN service provides secure communication between peers, but there are several security considerations to keep in mind:

1. **Virtual IP Assignment**: Virtual IPs should be unique and properly registered
2. **Port Blocking**: Use the `unallowed_ports` configuration to block sensitive ports
3. **Peer Authentication**: Ensure proper peer authentication is configured
4. **Traffic Encryption**: All VPN traffic is encrypted using libp2p's transport encryption

## Development Guidelines

When developing new features for the VPN system, consider the following:

- Use the Interface Segregation Principle to create small, focused interfaces
- Move hardcoded time values into configuration settings
- Implement packet synchronization to ensure strict sequential processing
- Include stress tests and metrics checks in test suites
- Keep files small and focused on a single responsibility
- Provide clear explanations of complex logic before restructuring

## Implementation Summary

This implementation provides a modular, maintainable VPN system for the subnet node. The code is organized into several packages, each with a specific responsibility, following the Interface Segregation Principle.

Key implementation principles include:

- **Modular Design**: Each component has a clear responsibility and interfaces
- **Resilience Patterns**: Circuit breakers and retries for handling failures
- **Metrics Collection**: Comprehensive metrics for monitoring and debugging
- **Connection-Based Routing**: Ensures packets with the same connection are processed sequentially
- **Dynamic Stream Scaling**: Automatically scales streams based on traffic load
- **Configuration Management**: Centralized configuration with validation
- **Stream Health Monitoring**: Automatic detection and replacement of unhealthy streams

## Detailed Documentation

For more detailed documentation, see the [VPN Documentation](../../docs/vpn.md) file.
