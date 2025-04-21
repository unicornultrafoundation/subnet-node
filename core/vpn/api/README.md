# VPN API Package

The `api` package defines the interfaces for the VPN service. These interfaces provide a clean separation between components and enable easier testing and mocking.

## Overview

The API package follows the Interface Segregation Principle (ISP) to create small, focused interfaces that define the contracts between different components of the VPN service.

## Key Interfaces

### VPNService

The main interface for the VPN system, including:

- Starting and stopping the VPN service
- Getting metrics

### PeerDiscoveryService

Handles peer discovery and mapping, including:

- Getting the peer ID for a destination IP
- Syncing the peer ID to the DHT

### StreamService

Handles stream creation and management, including:

- Creating new VPN streams to peers

### StreamPoolService

Handles stream pooling, including:

- Getting streams from the pool
- Releasing streams back to the pool

### CircuitBreakerService

Handles circuit breaker operations, including:

- Executing operations with circuit breaker protection
- Resetting circuit breakers
- Getting circuit breaker states

### StreamHealthService

Handles stream health monitoring, including:

- Starting and stopping the health checker
- Starting and stopping the stream warmer
- Getting health metrics

### RetryService

Handles retry operations with backoff, including:

- Retrying operations with exponential backoff

### MetricsService

Handles metrics collection, including:

- Incrementing stream errors
- Incrementing packets sent
- Incrementing packets dropped

### ConfigService

Handles configuration retrieval, including:

- Getting worker settings
- Getting stream pool settings
- Getting circuit breaker settings
- Getting health check settings
- Getting retry settings

## Usage

The interfaces in the API package are used by the various components of the VPN service to interact with each other without direct dependencies. For example, the `PeerDiscovery` component implements the `PeerDiscoveryService` interface, and other components that need peer discovery functionality depend on this interface rather than the concrete implementation.

This approach has several benefits:

1. **Decoupling**: Components are decoupled from each other, making the system more modular and easier to maintain.
2. **Testability**: Components can be tested in isolation using mock implementations of their dependencies.
3. **Flexibility**: Implementations can be swapped out without affecting the rest of the system.

## Example

Here's an example of how the interfaces are used in the VPN service:

```go
// The packet dispatcher depends on the PeerDiscoveryService interface
type Dispatcher struct {
    peerDiscovery api.PeerDiscoveryService
    // ...
}

// The dispatcher uses the interface to get peer IDs for destination IPs
func (d *Dispatcher) dispatchPacket(packet []byte, destIP string) error {
    peerID, err := d.peerDiscovery.GetPeerID(ctx, destIP)
    if err != nil {
        return err
    }

    // Use the peer ID to route the packet
    // ...
}
```

## Best Practices

When working with the API package, follow these best practices:

1. **Keep Interfaces Small**: Follow the Interface Segregation Principle and keep interfaces focused on a single responsibility.
2. **Use Interfaces for Dependencies**: Components should depend on interfaces, not concrete implementations.
3. **Document Interface Contracts**: Clearly document the expected behavior of each interface method.
4. **Provide Mock Implementations**: Create mock implementations of interfaces for testing.
5. **Avoid Interface Bloat**: Don't add methods to interfaces unless they are needed by multiple components.