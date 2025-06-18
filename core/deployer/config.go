package deployer

import (
	"time"

	"github.com/unicornultrafoundation/subnet-node/config"
)

// ServiceConfig represents service configuration
type ServiceConfig struct {
	// Kubernetes configuration
	KubeConfigPath string

	// Monitor interval
	MonitorInterval time.Duration

	// Service configuration
	DefaultServiceType string
	LocalhostEnabled   bool
}

func NewServiceConfigFromConfig(cfg *config.C) (*ServiceConfig, error) {
	serviceConfig := DefaultServiceConfig()

	// Kubernetes configuration
	serviceConfig.KubeConfigPath = cfg.GetString("deployer.kubeconfig_path", "")

	// Monitor interval
	serviceConfig.MonitorInterval = cfg.GetDuration("deployer.monitor_interval", 30*time.Second)

	// Service configuration
	serviceConfig.DefaultServiceType = cfg.GetString("deployer.default_service_type", "NodePort")
	serviceConfig.LocalhostEnabled = cfg.GetBool("deployer.localhost_enabled", true)

	if err := serviceConfig.Validate(); err != nil {
		return nil, err
	}

	return serviceConfig, nil
}

// DefaultServiceConfig returns a default service configuration
func DefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		MonitorInterval:    30 * time.Second,
		DefaultServiceType: "NodePort",
		LocalhostEnabled:   true,
	}
}

// Validate validates the service configuration
func (c *ServiceConfig) Validate() error {
	return nil
}
