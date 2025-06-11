package deployer

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/unicornultrafoundation/subnet-node/config"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ServiceConfig represents service configuration
type ServiceConfig struct {
	// Provider configuration
	ProviderAddress common.Address
	MaxRetries      int
	StoreDir        string // Directory for storing payment data

	// Kubernetes configuration
	KubeConfigPath string
	KubeConfig     *rest.Config

	// Blockchain configuration
	EthEndpoint     string
	IPFSURL         string
	ContractAddress string

	// Cluster resource configuration
	ClusterResources struct {
		CPU     int64  // Number of CPU cores
		Memory  string // Memory in GB (e.g., "8Gi")
		Storage string // Storage in GB (e.g., "100Gi")
		GPU     int64  // Number of GPU units
	}
}

func NewServiceConfigFromConfig(cfg *config.C) (*ServiceConfig, error) {
	serviceConfig := DefaultServiceConfig()

	// Provider configuration
	serviceConfig.MaxRetries = cfg.GetInt("deployer.max_retries", serviceConfig.MaxRetries)
	serviceConfig.ProviderAddress = common.HexToAddress(cfg.GetString("provider.address", ""))

	// Kubernetes configuration
	serviceConfig.KubeConfigPath = cfg.GetString("deployer.kubeconfig_path", "")
	serviceConfig.KubeConfig = nil
	if serviceConfig.KubeConfigPath != "" {
		var err error
		serviceConfig.KubeConfig, err = clientcmd.BuildConfigFromFlags("", serviceConfig.KubeConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
		}
	}

	// Blockchain configuration
	serviceConfig.EthEndpoint = cfg.GetString("account.rpc", "")
	serviceConfig.IPFSURL = cfg.GetString("deployer.ipfs_url", "")
	serviceConfig.ContractAddress = cfg.GetString("deployer.contract_address", "")

	// Cluster resources
	serviceConfig.ClusterResources.CPU = int64(cfg.GetInt("deployer.cluster_resources.cpu", int(serviceConfig.ClusterResources.CPU)))
	serviceConfig.ClusterResources.Memory = cfg.GetString("deployer.cluster_resources.memory", serviceConfig.ClusterResources.Memory)
	serviceConfig.ClusterResources.Storage = cfg.GetString("deployer.cluster_resources.storage", serviceConfig.ClusterResources.Storage)
	serviceConfig.ClusterResources.GPU = int64(cfg.GetInt("deployer.cluster_resources.gpu", int(serviceConfig.ClusterResources.GPU)))

	return serviceConfig, nil
}

// DefaultServiceConfig returns a default service configuration
func DefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		MaxRetries: 3,
		ClusterResources: struct {
			CPU     int64
			Memory  string
			Storage string
			GPU     int64
		}{
			CPU:     2,      // 2 CPU cores
			Memory:  "4Gi",  // 4 GB memory
			Storage: "10Gi", // 10 GB storage
			GPU:     0,      // No GPU by default
		},
	}
}

// Validate validates the service configuration
func (c *ServiceConfig) Validate() error {
	if c.KubeConfig == nil {
		return fmt.Errorf("kubeconfig is required")
	}

	if c.StoreDir == "" {
		return fmt.Errorf("store directory is required")
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries must be non-negative")
	}

	// Validate cluster resources
	if c.ClusterResources.CPU < 0 {
		return fmt.Errorf("CPU cores must be non-negative")
	}

	if c.ClusterResources.Memory != "" {
		if _, err := resource.ParseQuantity(c.ClusterResources.Memory); err != nil {
			return fmt.Errorf("invalid memory configuration: %w", err)
		}
	}

	if c.ClusterResources.Storage != "" {
		if _, err := resource.ParseQuantity(c.ClusterResources.Storage); err != nil {
			return fmt.Errorf("invalid storage configuration: %w", err)
		}
	}

	if c.ClusterResources.GPU < 0 {
		return fmt.Errorf("GPU units must be non-negative")
	}

	return nil
}
