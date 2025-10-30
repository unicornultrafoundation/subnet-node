package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/client/dynamic"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/client/static"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/utils"
)

type DiscoveryService interface {
	// GetPeerID gets the peer ID for a destination IP
	GetPeerID(ctx context.Context, destIP string) (string, error)

	// GetIP gets the IP for a Peer ID
	GetIP(ctx context.Context, peerID string) (string, error)
}

type DiscoveryServiceImpl struct {
	staticIPClient  static.StaticIPClient
	dynamicIPClient dynamic.DynamicIPClient
	peerIDCache     *cache.Cache
	ipCache         *cache.Cache
	mu              sync.RWMutex
}

func NewDiscoveryService(staticIPClient static.StaticIPClient, dynamicIPClient dynamic.DynamicIPClient) DiscoveryService {
	return &DiscoveryServiceImpl{
		staticIPClient:  staticIPClient,
		dynamicIPClient: dynamicIPClient,
		peerIDCache:     cache.New(1*time.Minute, 30*time.Second),
		ipCache:         cache.New(1*time.Minute, 30*time.Second),
	}
}

func (d *DiscoveryServiceImpl) GetPeerID(ctx context.Context, destIP string) (string, error) {
	peerID, err := d.getPeerIDFromCache(destIP)
	if err == nil && peerID != "" {
		return peerID, nil
	}

	peerID, err = d.staticIPClient.GetPeerID(ctx, destIP)
	if err == nil && peerID != "" {
		d.cachePeerID(destIP, peerID)
		d.cacheIP(peerID, destIP)
		return peerID, nil
	}

	lease, err := d.dynamicIPClient.GetLeaseByIP(ctx, destIP)
	if err == nil && lease.PeerID != "" {
		d.cachePeerID(destIP, lease.PeerID)
		d.cacheIP(lease.PeerID, destIP)
		return lease.PeerID, nil
	}

	return "", fmt.Errorf("no peer ID found for IP %s", destIP)
}

func (d *DiscoveryServiceImpl) GetIP(ctx context.Context, peerID string) (string, error) {
	ip, err := d.getIPFromCache(peerID)
	if err == nil && ip != "" {
		return ip, nil
	}

	lease, err := d.dynamicIPClient.GetLeaseByPeerID(ctx, peerID)
	if err != nil || lease.TokenID == 0 {
		return "", fmt.Errorf("failed to get lease by peer ID: %w", err)
	}

	// Compare the peer ID with the lease peer ID
	if lease.PeerID != peerID {
		return "", fmt.Errorf("peer ID %s is not the same as the lease peer ID %s", peerID, lease.PeerID)
	}

	ip = utils.ConvertNumberToVirtualIP(uint32(lease.TokenID))
	// Verify the IP is owned by the peer ID
	peerIDFromIp, err := d.GetPeerID(ctx, ip)
	if err != nil || peerIDFromIp != peerID {
		return "", fmt.Errorf("peer ID %s is not the same as the lease peer ID %s", peerID, peerIDFromIp)
	}

	d.cacheIP(peerID, ip)

	return ip, nil
}

func (d *DiscoveryServiceImpl) getIPFromCache(peerID string) (string, error) {
	d.mu.RLock()
	ip, found := d.ipCache.Get(peerID)
	d.mu.RUnlock()

	if found {
		return ip.(string), nil
	}

	return "", fmt.Errorf("no IP found for peer ID %s", peerID)
}

func (d *DiscoveryServiceImpl) cacheIP(peerID string, ip string) {
	d.mu.Lock()
	d.ipCache.Set(peerID, ip, cache.DefaultExpiration)
	d.mu.Unlock()
}

func (d *DiscoveryServiceImpl) getPeerIDFromCache(destIP string) (string, error) {
	d.mu.RLock()
	peerID, found := d.peerIDCache.Get(destIP)
	d.mu.RUnlock()

	if found {
		return peerID.(string), nil
	}

	return "", fmt.Errorf("no peer ID found for IP %s", destIP)
}

func (d *DiscoveryServiceImpl) cachePeerID(destIP string, peerID string) {
	d.mu.Lock()
	d.peerIDCache.Set(destIP, peerID, cache.DefaultExpiration)
	d.mu.Unlock()
}
