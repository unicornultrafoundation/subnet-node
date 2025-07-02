package deployer

import (
	"os"
	"path/filepath"
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

	// Deployment configuration
	DeploymentWaitTimeout time.Duration // Time to wait for the deployment to be ready after the deployment is created
}

func NewServiceConfigFromConfig(cfg *config.C) (*ServiceConfig, error) {
	serviceConfig := DefaultServiceConfig()

	// Get user's home directory for default kubeconfig path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "" // Fallback to empty string if home directory cannot be determined
	}
	defaultKubeConfigPath := ""
	if homeDir != "" {
		defaultKubeConfigPath = filepath.Join(homeDir, ".kube", "config")
	}

	// Kubernetes configuration
	serviceConfig.KubeConfigPath = cfg.GetString("deployer.kubeconfig_path", defaultKubeConfigPath)

	// Monitor interval
	serviceConfig.MonitorInterval = cfg.GetDuration("deployer.monitor_interval", 30*time.Second)

	// Service configuration
	serviceConfig.DefaultServiceType = cfg.GetString("deployer.default_service_type", "NodePort")
	serviceConfig.LocalhostEnabled = cfg.GetBool("deployer.localhost_enabled", true)

	// Deployment configuration
	serviceConfig.DeploymentWaitTimeout = cfg.GetDuration("deployer.deployment_wait_timeout", 30*time.Second)

	if err := serviceConfig.Validate(); err != nil {
		return nil, err
	}

	return serviceConfig, nil
}

// DefaultServiceConfig returns a default service configuration
func DefaultServiceConfig() *ServiceConfig {
	// Get user's home directory for default kubeconfig path
	homeDir, err := os.UserHomeDir()
	defaultKubeConfigPath := ""
	if err == nil && homeDir != "" {
		defaultKubeConfigPath = filepath.Join(homeDir, ".kube", "config")
	}

	return &ServiceConfig{
		KubeConfigPath:        defaultKubeConfigPath,
		MonitorInterval:       30 * time.Second,
		DefaultServiceType:    "NodePort",
		LocalhostEnabled:      true,
		DeploymentWaitTimeout: 30 * time.Second,
	}
}

// Validate validates the service configuration
func (c *ServiceConfig) Validate() error {
	return nil
}
