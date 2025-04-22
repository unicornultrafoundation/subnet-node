package main

import (
	"log"
	"time"

	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/router"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/pool"
)

func main() {
	// This is a simplified example showing how to integrate the StreamRouter
	// In a real application, you would use actual implementations of these services

	// Create dependencies
	streamPool := createStreamPool()
	peerDiscovery := createPeerDiscovery()

	// Create router config with short TTL for frequent cleanup
	config := &router.StreamRouterConfig{
		MinStreamsPerPeer:        1,
		MaxStreamsPerPeer:        5,
		ThroughputThreshold:      1000,
		ScaleUpThreshold:         0.8,
		ScaleDownThreshold:       0.3,
		ScalingInterval:          5 * time.Second,
		MinWorkers:               4,
		MaxWorkers:               16,
		InitialWorkers:           8,
		WorkerQueueSize:          1000,
		WorkerScaleInterval:      10 * time.Second,
		WorkerScaleUpThreshold:   0.75,
		WorkerScaleDownThreshold: 0.25,
		ConnectionTTL:            30 * time.Second, // Short TTL for frequent cleanup
		CleanupInterval:          10 * time.Second, // Short interval for frequent cleanup
		CacheShardCount:          16,
	}

	// Create router
	streamRouter := router.NewStreamRouter(config, streamPool, peerDiscovery)
	defer streamRouter.Shutdown()

	// In a real application, you would:
	// 1. Read packets from TUN device
	// 2. Dispatch them using streamRouter.DispatchPacket
	// 3. Handle any errors

	// For this example, we'll just sleep to keep the program running
	log.Println("StreamRouter started. Press Ctrl+C to exit.")
	select {}
}

// Helper functions to create dependencies
// In a real application, you would use actual implementations

func createStreamPool() pool.PoolServiceExtension {
	// In a real application, you would create a real PoolService
	// For this example, we'll return nil
	return nil
}

func createPeerDiscovery() api.PeerDiscoveryService {
	// In a real application, you would create a real PeerDiscoveryService
	// For this example, we'll return nil
	return nil
}
