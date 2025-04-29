# VPN Discovery Package

The `discovery` package handles peer discovery and mapping between virtual IPs and peer IDs in the VPN service. It provides mechanisms for storing and retrieving peer information from the DHT and the IP registry contract.

## Overview

The discovery package implements the `PeerDiscoveryService` interface defined in the `api` package. It provides methods for:

1. **Peer Lookup**: Finding the peer ID associated with a virtual IP address
2. **DHT Synchronization**: Storing and retrieving peer information in the DHT
3. **Registry Verification**: Verifying that virtual IPs are properly registered
4. **Virtual IP Mapping**: Converting between virtual IPs and peer IDs

## Key Components

### PeerDiscovery

The `PeerDiscovery` struct is the main component of the discovery package. It implements the `PeerDiscoveryService` interface and coordinates all peer discovery operations.

```go
// PeerDiscovery handles peer discovery and mapping
type PeerDiscovery struct {
    hostService    api.HostService
    dhtService     api.DHTService
    virtualIP      string
    accountService api.AccountService
}
```

### Service Adapters

The discovery package includes several adapters that implement the interfaces defined in the `api` package:

- **HostServiceAdapter**: Adapts the libp2p host to the `api.HostService` interface
- **DHTServiceAdapter**: Adapts the DHT to the `api.DHTService` interface
- **AccountServiceAdapter**: Adapts the account service to the `api.AccountService` interface
- **IPRegistryAdapter**: Adapts the IP registry contract to the `api.IPRegistry` interface

## Peer Discovery Process

The peer discovery process follows these steps:

1. **Cache Lookup**: Check if the peer ID is already cached locally
2. **Registry Lookup**: If not found in cache, look up the peer ID in the IP registry contract
3. **DHT Lookup**: If not found in the registry, look up the peer ID in the DHT
4. **Cache Update**: Update the local cache with the found peer ID

## DHT Synchronization

The `SyncPeerIDToDHT` method synchronizes the node's peer ID and virtual IP to the DHT:

1. Create a hash of the virtual IP
2. Sign the hash with the node's private key
3. Store the mapping in the DHT with the hash as the key and the peer ID as the value

## Virtual IP Verification

The `VerifyVirtualIPHasRegistered` method verifies that a virtual IP has been properly registered:

1. Convert the virtual IP to a token ID
2. Look up the peer ID in the IP registry contract using the token ID
3. Return an error if the peer ID is not found

## Usage

The discovery package is used by the VPN service to find peers and establish connections:

```go
// Create the peer discovery service
peerDiscovery := discovery.NewPeerDiscovery(
    hostServiceAdapter,
    dhtServiceAdapter,
    virtualIP,
    accountServiceAdapter,
)

// Sync the peer ID to the DHT
err := peerDiscovery.SyncPeerIDToDHT(ctx)
if err != nil {
    return fmt.Errorf("failed to sync peer id to DHT: %v", err)
}

// Look up a peer ID for a destination IP
peerID, err := peerDiscovery.GetPeerID(ctx, destIP)
if err != nil {
    return fmt.Errorf("failed to get peer ID for %s: %v", destIP, err)
}
```

## Error Handling

The discovery package handles several error conditions:

- **Peer Not Found**: When a peer ID cannot be found for a virtual IP
- **DHT Errors**: When DHT operations fail
- **Registry Errors**: When registry operations fail
- **Invalid Virtual IP**: When a virtual IP is not in the correct format

## Performance Considerations

The discovery package includes several optimizations for performance:

1. **Local Caching**: Peer IDs are cached locally to avoid repeated lookups
2. **Prioritized Lookup**: Lookups are performed in order of increasing cost (cache -> registry -> DHT)
3. **Batch Operations**: DHT operations can be batched for efficiency

## Security Considerations

The discovery package includes several security features:

1. **Signed Mappings**: DHT mappings are signed with the node's private key
2. **Registry Verification**: Virtual IPs are verified against the registry
3. **Input Validation**: Virtual IPs are validated before use
