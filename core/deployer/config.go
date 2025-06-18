package deployer

import (
	"github.com/unicornultrafoundation/subnet-node/config"
)

// ServiceConfig represents service configuration
type ServiceConfig struct {
	// Kubernetes configuration
	KubeConfigPath string
}

func NewServiceConfigFromConfig(cfg *config.C) (*ServiceConfig, error) {
	serviceConfig := DefaultServiceConfig()

	// Kubernetes configuration
	serviceConfig.KubeConfigPath = cfg.GetString("deployer.kubeconfig_path", "")

	if err := serviceConfig.Validate(); err != nil {
		return nil, err
	}

	return serviceConfig, nil
}

// DefaultServiceConfig returns a default service configuration
func DefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		KubeConfigPath: "~/.kube/config",
	}
}

// Validate validates the service configuration
func (c *ServiceConfig) Validate() error {
	return nil
}
