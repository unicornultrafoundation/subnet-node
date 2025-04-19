# VPN API Package

This package defines the interfaces for the VPN system, providing a clean separation between components.

## Interfaces

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

### StreamMultiplexService

Handles stream multiplexing, including:

- Starting and stopping the multiplexer
- Sending packets using the multiplexer
- Getting multiplexer metrics

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
- Getting multiplexer settings