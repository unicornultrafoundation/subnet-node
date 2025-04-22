package router

import (
	"time"
)

// ConfigService provides configuration for the StreamRouter
type ConfigService struct {
	cfg interface {
		GetBool(key string, defaultValue bool) bool
		GetInt(key string, defaultValue int) int
		GetString(key string, defaultValue string) string
		GetDuration(key string, defaultValue time.Duration) time.Duration
	}
}

// NewConfigService creates a new ConfigService
func NewConfigService(cfg interface {
	GetBool(key string, defaultValue bool) bool
	GetInt(key string, defaultValue int) int
	GetString(key string, defaultValue string) string
	GetDuration(key string, defaultValue time.Duration) time.Duration
}) *ConfigService {
	return &ConfigService{
		cfg: cfg,
	}
}

// GetStreamRouterConfig returns the StreamRouter configuration
func (s *ConfigService) GetStreamRouterConfig() *StreamRouterConfig {
	return &StreamRouterConfig{
		MinStreamsPerPeer:        s.GetMinStreamsPerPeer(),
		MaxStreamsPerPeer:        s.GetMaxStreamsPerPeer(),
		ThroughputThreshold:      s.GetThroughputThreshold(),
		ScaleUpThreshold:         s.GetScaleUpThreshold(),
		ScaleDownThreshold:       s.GetScaleDownThreshold(),
		ScalingInterval:          s.GetScalingInterval(),
		MinWorkers:               s.GetMinWorkers(),
		MaxWorkers:               s.GetMaxWorkers(),
		InitialWorkers:           s.GetInitialWorkers(),
		WorkerQueueSize:          s.GetWorkerQueueSize(),
		WorkerScaleInterval:      s.GetWorkerScaleInterval(),
		WorkerScaleUpThreshold:   s.GetWorkerScaleUpThreshold(),
		WorkerScaleDownThreshold: s.GetWorkerScaleDownThreshold(),
		ConnectionTTL:            s.GetConnectionTTL(),
		CleanupInterval:          s.GetCleanupInterval(),
		CacheShardCount:          s.GetCacheShardCount(),
	}
}

// GetMinStreamsPerPeer returns the minimum number of streams per peer
func (s *ConfigService) GetMinStreamsPerPeer() int {
	return s.cfg.GetInt("vpn.router.min_streams_per_peer", 1)
}

// GetMaxStreamsPerPeer returns the maximum number of streams per peer
func (s *ConfigService) GetMaxStreamsPerPeer() int {
	return s.cfg.GetInt("vpn.router.max_streams_per_peer", 10)
}

// GetThroughputThreshold returns the throughput threshold for scaling
func (s *ConfigService) GetThroughputThreshold() int64 {
	return int64(s.cfg.GetInt("vpn.router.throughput_threshold", 1000))
}

// GetScaleUpThreshold returns the scale up threshold
func (s *ConfigService) GetScaleUpThreshold() float64 {
	return float64(s.cfg.GetInt("vpn.router.scale_up_threshold_percent", 80)) / 100.0
}

// GetScaleDownThreshold returns the scale down threshold
func (s *ConfigService) GetScaleDownThreshold() float64 {
	return float64(s.cfg.GetInt("vpn.router.scale_down_threshold_percent", 30)) / 100.0
}

// GetScalingInterval returns the scaling interval
func (s *ConfigService) GetScalingInterval() time.Duration {
	return s.cfg.GetDuration("vpn.router.scaling_interval", 5*time.Second)
}

// GetMinWorkers returns the minimum number of workers
func (s *ConfigService) GetMinWorkers() int {
	return s.cfg.GetInt("vpn.router.min_workers", 4)
}

// GetMaxWorkers returns the maximum number of workers
func (s *ConfigService) GetMaxWorkers() int {
	return s.cfg.GetInt("vpn.router.max_workers", 32)
}

// GetInitialWorkers returns the initial number of workers
func (s *ConfigService) GetInitialWorkers() int {
	return s.cfg.GetInt("vpn.router.initial_workers", 8)
}

// GetWorkerQueueSize returns the worker queue size
func (s *ConfigService) GetWorkerQueueSize() int {
	return s.cfg.GetInt("vpn.router.worker_queue_size", 1000)
}

// GetWorkerScaleInterval returns the worker scale interval
func (s *ConfigService) GetWorkerScaleInterval() time.Duration {
	return s.cfg.GetDuration("vpn.router.worker_scale_interval", 10*time.Second)
}

// GetWorkerScaleUpThreshold returns the worker scale up threshold
func (s *ConfigService) GetWorkerScaleUpThreshold() float64 {
	return float64(s.cfg.GetInt("vpn.router.worker_scale_up_threshold_percent", 75)) / 100.0
}

// GetWorkerScaleDownThreshold returns the worker scale down threshold
func (s *ConfigService) GetWorkerScaleDownThreshold() float64 {
	return float64(s.cfg.GetInt("vpn.router.worker_scale_down_threshold_percent", 25)) / 100.0
}

// GetConnectionTTL returns the connection TTL
func (s *ConfigService) GetConnectionTTL() time.Duration {
	return s.cfg.GetDuration("vpn.router.connection_ttl", 30*time.Second)
}

// GetCleanupInterval returns the cleanup interval
func (s *ConfigService) GetCleanupInterval() time.Duration {
	return s.cfg.GetDuration("vpn.router.cleanup_interval", 10*time.Second)
}

// GetCacheShardCount returns the cache shard count
func (s *ConfigService) GetCacheShardCount() int {
	return s.cfg.GetInt("vpn.router.cache_shard_count", 16)
}
