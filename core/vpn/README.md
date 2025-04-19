# VPN Package

This package provides a virtual private network (VPN) implementation for the subnet node.

## Package Structure

The VPN package is organized into several subpackages:

- `api`: Defines the interfaces for the VPN system
- `adapter`: Provides adapters between different implementations
- `config`: Contains configuration structures and validation
- `discovery`: Handles peer discovery and mapping
- `metrics`: Provides metrics collection and reporting
- `network`: Manages network interfaces and connections
- `packet`: Handles packet processing and routing
- `resilience`: Implements resilience patterns like circuit breakers and retries
- `stream`: Manages P2P streams and multiplexing

## Main Components

### Configuration

The VPN configuration is defined in the `config` package. It includes settings for:

- Network configuration (MTU, virtual IP, subnet, routes)
- Worker settings (idle timeout, buffer size, cleanup interval)
- Stream pool settings (max/min streams per peer, idle timeout)
- Circuit breaker settings (failure threshold, reset timeout, success threshold)
- Health check settings (interval, timeout, max consecutive failures)
- Multiplexer settings (max/min streams per multiplexer, auto-scaling interval)

### Packet Processing

Packets are processed by workers that handle specific destination IP:Port combinations. The workers are managed by a dispatcher that routes packets to the appropriate worker.

### Stream Management

Streams are managed by a pool that maintains connections to peers. The pool can be configured to maintain a minimum number of streams per peer and to automatically clean up idle streams.

### Resilience

The VPN system includes several resilience patterns:

- Circuit breakers to prevent cascading failures
- Retries with exponential backoff for transient failures
- Stream health checks to detect and replace unhealthy streams
- Stream multiplexing to improve throughput and reliability

### Metrics

The VPN system collects metrics for monitoring and debugging, including:

- Packet counts (received, sent, dropped)
- Byte counts (received, sent)
- Stream counts (created, closed, errors)
- Circuit breaker states and counts
- Worker counts and states

## Usage

To use the VPN system, create a new VPN service and start it:

```go
service := vpn.New(cfg, peerHost, dht, accountService)
service.Start(ctx)
```

The service will set up a TUN interface, handle P2P traffic, and listen for packets from the TUN interface.

## Development

When developing new features for the VPN system, consider the following:

- Use the Interface Segregation Principle to create small, focused interfaces
- Move hardcoded time values into configuration settings
- Implement packet synchronization to ensure strict sequential processing
- Include stress tests and metrics checks in test suites
- Keep files small and focused on a single responsibility
- Provide clear explanations of complex logic before restructuring
