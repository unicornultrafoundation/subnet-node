package k8s

import (
	"time"

	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/k8s/kube/builder"
)

type Config struct {
	InventoryResourcePollPeriod     time.Duration
	InventoryResourceDebugFrequency uint
	InventoryExternalPortQuantity   uint
	CPUCommitLevel                  float64
	GPUCommitLevel                  float64
	MemoryCommitLevel               float64
	StorageCommitLevel              float64
	MonitorMaxRetries               uint
	MonitorRetryPeriod              time.Duration
	MonitorRetryPeriodJitter        time.Duration
	MonitorHealthcheckPeriod        time.Duration
	MonitorHealthcheckPeriodJitter  time.Duration
	MonitorExpiryCheckPeriod        time.Duration
	MonitorExpiryCheckPeriodJitter  time.Duration
	ClusterSettings                 map[interface{}]interface{}
}

func NewConfig(cfg *config.C) Config {
	config := Config{}

	kubeSettings := builder.NewDefaultSettings()
	config.ClusterSettings = map[interface{}]interface{}{
		builder.SettingsKey: kubeSettings,
	}

	config.InventoryExternalPortQuantity = uint(cfg.GetUint32("deployer.inventory_external_port_quantity", 10000))
	config.InventoryResourcePollPeriod = cfg.GetDuration("deployer.inventory_resource_poll_period", time.Second*5)
	config.InventoryResourceDebugFrequency = uint(cfg.GetUint32("deployer.inventory_resource_debug_frequency", 10))
	config.MonitorMaxRetries = uint(cfg.GetUint32("deployer.monitor_max_retries", 40))
	config.MonitorRetryPeriod = cfg.GetDuration("deployer.monitor_retry_period", time.Second*4)
	config.MonitorRetryPeriodJitter = cfg.GetDuration("deployer.monitor_retry_period_jitter", time.Second*15)
	config.MonitorHealthcheckPeriod = cfg.GetDuration("deployer.monitor_healthcheck_period", time.Second*10)
	config.MonitorHealthcheckPeriodJitter = cfg.GetDuration("deployer.monitor_healthcheck_period_jitter", time.Second*5)
	config.MonitorExpiryCheckPeriod = cfg.GetDuration("deployer.monitor_expiry_check_period", time.Second*30)
	config.MonitorExpiryCheckPeriodJitter = cfg.GetDuration("deployer.monitor_expiry_check_period_jitter", time.Second*5)

	return config
}
