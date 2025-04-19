# VPN Discovery Package

This package handles peer discovery and mapping for the VPN system.

## Components

### PeerDiscovery

The `PeerDiscovery` implements the `PeerDiscoveryService` interface, including:

- Getting the peer ID for a destination IP
- Syncing the peer ID to the DHT
- Caching peer ID lookups for performance

## Usage

```go
// Create a peer discovery service
peerDiscovery := discovery.NewPeerDiscovery(host, dht, virtualIP)

// Sync the peer ID to the DHT
err := peerDiscovery.SyncPeerIDToDHT(ctx)
if err != nil {
    return fmt.Errorf("failed to sync peer id to DHT: %v", err)
}

// Get the peer ID for a destination IP
peerID, err := peerDiscovery.GetPeerID(ctx, destIP)
if err != nil {
    return fmt.Errorf("no peer mapping found for IP %s: %v", destIP, err)
}

// Parse the peer ID
parsedPeerID, err := discovery.ParsePeerID(peerID)
if err != nil {
    return fmt.Errorf("failed to parse peerid %s: %v", peerID, err)
}
```
