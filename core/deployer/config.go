package k8scluster

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/rest"
)

// ServiceConfig represents service configuration
type ServiceConfig struct {
	// Provider configuration
	ProviderAddress common.Address
	DeploymentDir   string
	MaxRetries      int
	StoreDir        string // Directory for storing payment data

	// Kubernetes configuration
	KubeConfig *rest.Config

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

	if c.DeploymentDir == "" {
		return fmt.Errorf("deployment directory is required")
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
