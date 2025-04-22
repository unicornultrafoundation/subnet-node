package router

import (
	"time"
)

// DefaultConfig returns a default configuration for the StreamRouter
func DefaultConfig() *StreamRouterConfig {
	return &StreamRouterConfig{
		MinStreamsPerPeer:        1,
		MaxStreamsPerPeer:        10,
		ThroughputThreshold:      1000, // 1000 packets per second per stream
		ScaleUpThreshold:         0.8,
		ScaleDownThreshold:       0.3,
		ScalingInterval:          5 * time.Second,
		MinWorkers:               4,
		MaxWorkers:               32,
		InitialWorkers:           8,
		WorkerQueueSize:          1000,
		WorkerScaleInterval:      5 * time.Second,
		WorkerScaleUpThreshold:   0.75,
		WorkerScaleDownThreshold: 0.25,
		ConnectionTTL:            5 * time.Second,
		CleanupInterval:          5 * time.Second,
		CacheShardCount:          16,
	}
}
