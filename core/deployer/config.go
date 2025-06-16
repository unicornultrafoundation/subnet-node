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

	// Pricing configuration
	Pricing struct {
		MemPriceMin      int64   // Minimum memory price in wei
		MemPriceMax      int64   // Maximum memory price in wei
		BidPriceStrategy string  // Pricing strategy (e.g., "dynamic", "fixed")
		BidCPUScale      float64 // CPU price scaling factor
		BidStorageScale  float64 // Storage price scaling factor
		ProcessLimit     int     // Maximum number of concurrent processes
		ProcessTimeout   int     // Process timeout in seconds
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

	// Pricing configuration
	serviceConfig.Pricing.MemPriceMin = int64(cfg.GetInt("deployer.pricing.mem_price_min", int(serviceConfig.Pricing.MemPriceMin)))
	serviceConfig.Pricing.MemPriceMax = int64(cfg.GetInt("deployer.pricing.mem_price_max", int(serviceConfig.Pricing.MemPriceMax)))
	serviceConfig.Pricing.BidPriceStrategy = cfg.GetString("deployer.pricing.bid_price_strategy", serviceConfig.Pricing.BidPriceStrategy)
	serviceConfig.Pricing.BidCPUScale = float64(cfg.GetInt("deployer.pricing.bid_cpu_scale", int(serviceConfig.Pricing.BidCPUScale*100))) / 100
	serviceConfig.Pricing.BidStorageScale = float64(cfg.GetInt("deployer.pricing.bid_storage_scale", int(serviceConfig.Pricing.BidStorageScale*100))) / 100
	serviceConfig.Pricing.ProcessLimit = cfg.GetInt("deployer.pricing.process_limit", serviceConfig.Pricing.ProcessLimit)
	serviceConfig.Pricing.ProcessTimeout = cfg.GetInt("deployer.pricing.process_timeout", serviceConfig.Pricing.ProcessTimeout)

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
		Pricing: struct {
			MemPriceMin      int64
			MemPriceMax      int64
			BidPriceStrategy string
			BidCPUScale      float64
			BidStorageScale  float64
			ProcessLimit     int
			ProcessTimeout   int
		}{
			MemPriceMin:      1000000000000000,  // 0.001 U2U
			MemPriceMax:      10000000000000000, // 0.01 U2U
			BidPriceStrategy: "dynamic",
			BidCPUScale:      1.5,
			BidStorageScale:  1.2,
			ProcessLimit:     10,
			ProcessTimeout:   30,
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

	// Validate pricing configuration
	if c.Pricing.MemPriceMin < 0 {
		return fmt.Errorf("minimum memory price must be non-negative")
	}

	if c.Pricing.MemPriceMax < c.Pricing.MemPriceMin {
		return fmt.Errorf("maximum memory price must be greater than minimum memory price")
	}

	if c.Pricing.BidCPUScale <= 0 {
		return fmt.Errorf("CPU price scale must be positive")
	}

	if c.Pricing.BidStorageScale <= 0 {
		return fmt.Errorf("storage price scale must be positive")
	}

	if c.Pricing.ProcessLimit <= 0 {
		return fmt.Errorf("process limit must be positive")
	}

	if c.Pricing.ProcessTimeout <= 0 {
		return fmt.Errorf("process timeout must be positive")
	}

	return nil
}
