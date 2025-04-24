# VPN Network Package

This package manages network interfaces and connections for the VPN system.

## Components

### TUNService

The `TUNService` handles TUN interface operations, including:

- Setting up the TUN interface
- Configuring the interface (MTU, IP, routes)
- Closing the interface

### ClientService

The `ClientService` handles client-side VPN operations, including:

- Reading packets from the TUN interface
- Dispatching packets to the appropriate stream
- Managing the buffer pool

### ServerService

The `ServerService` handles server-side VPN operations, including:

- Handling incoming P2P streams
- Writing packets to the TUN interface
- Filtering packets based on allowed ports

### VirtualIPManager

The `VirtualIPManager` manages virtual IP addresses, including:

- Allocating IPs for peers
- Releasing IPs
- Finding available IPs in the subnet

### BufferPool

The `BufferPool` manages a pool of byte buffers for packet processing, including:

- Getting buffers from the pool
- Returning buffers to the pool
- Creating new buffers when needed

## Usage

```go
// Create a TUN service
tunConfig := &network.TUNConfig{
    MTU:       1500,
    VirtualIP: "10.0.0.1",
    Subnet:    24,
    Routes:    []string{"10.0.0.0/24"},
}
tunService := network.NewTUNService(tunConfig)

// Set up the TUN interface
iface, err := tunService.SetupTUN()
if err != nil {
    return fmt.Errorf("failed to setup TUN interface: %v", err)
}

// Create a client service
clientService := network.NewClientService(
    tunService,
    dispatcher,
    metrics,
    bufferPool,
)

// Start the client service
clientService.Start(ctx)

// Create a server service
serverConfig := &network.ServerConfig{
    MTU:            1500,
    UnallowedPorts: map[string]bool{"22": true},
}
serverService := network.NewServerService(serverConfig, metrics)

// Handle an incoming stream
serverService.HandleStream(stream, iface)
```
