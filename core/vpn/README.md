# VPN Package

## Overview

The VPN package provides a secure, peer-to-peer Virtual Private Network implementation for the Subnet Node. It creates an encrypted overlay network that allows nodes to communicate securely regardless of their physical location or network configuration.

The VPN service leverages libp2p for peer-to-peer communication and establishes TUN interfaces on participating nodes to route traffic through the secure overlay network.

## Architecture

The VPN package is organized into several subpackages, each responsible for a specific aspect of the VPN functionality:

### Core Components

1. **Service (`service.go`)**: The main VPN service that coordinates all VPN functionality. It manages the lifecycle of all VPN components and provides the main interface for starting, stopping, and monitoring the VPN service.

2. **Network**: Handles packet processing and TUN interface management
   - `TUNService`: Manages the TUN interface for sending/receiving packets
   - `OutboundPacketService`: Processes packets from the TUN device to the network
   - `InboundPacketService`: Processes packets from the network to the TUN device

3. **Dispatcher**: Routes packets to the appropriate peer streams
   - `Dispatcher`: Implements the `DispatcherService` interface
   - Manages stream creation and packet routing

4. **Discovery**: Handles peer discovery and virtual IP mapping
   - `PeerDiscovery`: Maps virtual IPs to peer IDs
   - Uses DHT for distributed peer discovery

5. **Resilience**: Provides fault tolerance mechanisms
   - `ResilienceService`: Coordinates retry and circuit breaker patterns
   - `RetryManager`: Implements exponential backoff retry logic
   - `CircuitBreakerManager`: Implements circuit breaker pattern

6. **Config**: Manages VPN configuration
   - `ConfigService`: Provides access to VPN configuration settings
   - `VPNConfig`: Encapsulates all configuration settings

7. **Overlay**: Implements platform-specific TUN device functionality
   - Platform-specific implementations (Linux, Windows, macOS, etc.)
   - Handles routing and IP configuration

8. **Utils**: Provides utility functions
   - `ParsePacket`: Parses IP packets for firewall and routing
   - `ResourceManager`: Manages VPN resources

## Flow of Operation

1. **Initialization**:
   - The VPN service is created with configuration and dependencies
   - Core components are initialized (discovery, TUN, dispatcher, etc.)

2. **Starting the Service**:
   - Wait for peer connections
   - Set up the TUN interface with retry mechanism
   - Register stream handlers for incoming connections
   - Start the outbound packet service

3. **Packet Processing**:
   - **Outbound**: Packets from the local system are read from the TUN interface, processed, and dispatched to the appropriate peer
   - **Inbound**: Packets from peers are received via libp2p streams, processed, and written to the TUN interface

4. **Peer Discovery**:
   - Virtual IPs are mapped to peer IDs using DHT and registry lookups
   - Mappings are cached for performance

5. **Resilience**:
   - Circuit breaker pattern prevents cascading failures
   - Retry mechanisms handle transient errors
   - Resource management ensures proper cleanup

## Configuration

The VPN service can be configured through the following settings:

### Basic Settings
- `Enable`: Enable/disable the VPN service
- `MTU`: Maximum Transmission Unit for the TUN interface
- `VirtualIP`: Virtual IP address for this node
- `Subnet`: Subnet mask length
- `Routes`: Additional routes to add
- `Protocol`: Protocol ID for libp2p streams
- `Routines`: Number of reader routines

### Security Settings
- `UnallowedPorts`: Map of ports that are not allowed for inbound connections

### Stream Settings
- `StreamIdleTimeout`: Timeout for idle streams
- `StreamCleanupInterval`: Interval for cleaning up idle streams

### Circuit Breaker Settings
- `CircuitBreakerFailureThreshold`: Number of failures before opening the circuit
- `CircuitBreakerResetTimeout`: Timeout before attempting to close the circuit
- `CircuitBreakerSuccessThreshold`: Number of successes required to close the circuit

### Retry Settings
- `RetryMaxAttempts`: Maximum number of retry attempts
- `RetryInitialInterval`: Initial interval between retries
- `RetryMaxInterval`: Maximum interval between retries

## Platform Support

The VPN package supports multiple platforms through platform-specific TUN implementations:

- Linux
- Windows
- macOS (Darwin)
- FreeBSD
- NetBSD
- OpenBSD

Each platform has its own implementation in the `overlay` package.

## Error Handling and Resilience

The VPN package implements several resilience patterns:

1. **Circuit Breaker**: Prevents cascading failures by failing fast when a service is unavailable
2. **Retry with Exponential Backoff**: Handles transient errors by retrying operations with increasing delays
3. **Resource Management**: Ensures proper cleanup of resources

Example of using resilience patterns:

```go
// Use ExecuteWithResilience for better fault tolerance and metrics
breakerId := "tun_setup"
err, attempts := s.resilienceService.ExecuteWithResilience(ctx, breakerId, func() error {
    // Operation that might fail
    return s.tunService.SetupTUN()
})
```

## Security Considerations

- The VPN service filters inbound packets to prevent access to unauthorized ports
- Communication between peers is encrypted using libp2p's transport encryption
- Virtual IP mappings are verified to prevent spoofing

## Testing

The VPN package includes various tests:

- Unit tests for individual components
- Integration tests for end-to-end functionality
- Platform-specific tests for TUN implementations

## Known Issues and TODOs

- Multi-queue support is not implemented for all platforms
