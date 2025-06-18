package bidengine

import (
	"github.com/holiman/uint256"
	"github.com/unicornultrafoundation/subnet-node/config"
)

// ConfigKeys defines constants for configuration keys
const (
	// Bidding configuration keys
	ConfigKeyMinBidPercent = "bidengine.min_bid_percent"
	ConfigKeyMaxBidPercent = "bidengine.max_bid_percent"
	ConfigKeyPriceFactor   = "bidengine.price_factor"
	ConfigKeyMaxProfit     = "bidengine.max_profit"
	ConfigKeyMinProfit     = "bidengine.min_profit"

	// Resource requirements keys
	ConfigKeyMinCPUCores      = "bidengine.min_requirements.cpu_cores"
	ConfigKeyMinMemoryMB      = "bidengine.min_requirements.memory_mb"
	ConfigKeyMinDiskGB        = "bidengine.min_requirements.disk_gb"
	ConfigKeyMinGPUCores      = "bidengine.min_requirements.gpu_cores"
	ConfigKeyMinUploadSpeed   = "bidengine.min_requirements.upload_speed"
	ConfigKeyMinDownloadSpeed = "bidengine.min_requirements.download_speed"

	// Resource cost keys
	ConfigKeyCPUCorePrice       = "bidengine.resource_cost.cpu_core"
	ConfigKeyMemoryGBPrice      = "bidengine.resource_cost.memory_gb"
	ConfigKeyDiskGBPrice        = "bidengine.resource_cost.disk_gb"
	ConfigKeyGPUCorePrice       = "bidengine.resource_cost.gpu_core"
	ConfigKeyUploadSpeedPrice   = "bidengine.resource_cost.upload_speed"
	ConfigKeyDownloadSpeedPrice = "bidengine.resource_cost.download_speed"

	// Machine type multiplier keys
	ConfigKeyMachineTypeShared    = "bidengine.machine_type_multipliers.shared"
	ConfigKeyMachineTypeVM        = "bidengine.machine_type_multipliers.vm"
	ConfigKeyMachineTypeBareMetal = "bidengine.machine_type_multipliers.bare_metal"
)

// LoadBidConfig loads the bid configuration from the config file
func LoadBidConfig(cfg *config.C) BidConfig {
	bidConfig := BidConfig{
		MinBidPercent: cfg.GetInt(ConfigKeyMinBidPercent, 70),
		MaxBidPercent: cfg.GetInt(ConfigKeyMaxBidPercent, 95),
		PriceFactor:   float64(cfg.GetInt(ConfigKeyPriceFactor, 1)),
		MinRequirements: &BidRequirements{
			MinCPUCores:      uint256.NewInt(uint64(cfg.GetInt(ConfigKeyMinCPUCores, 1))),
			MinMemoryMB:      uint256.NewInt(uint64(cfg.GetInt(ConfigKeyMinMemoryMB, 1024))),
			MinDiskGB:        uint256.NewInt(uint64(cfg.GetInt(ConfigKeyMinDiskGB, 10))),
			MinGPUCores:      uint256.NewInt(uint64(cfg.GetInt(ConfigKeyMinGPUCores, 0))),
			MinUploadSpeed:   uint256.NewInt(uint64(cfg.GetInt(ConfigKeyMinUploadSpeed, 0))),
			MinDownloadSpeed: uint256.NewInt(uint64(cfg.GetInt(ConfigKeyMinDownloadSpeed, 0))),
		},
		MaxProfit: uint256.NewInt(uint64(cfg.GetInt(ConfigKeyMaxProfit, 1000000000000000000))), // Default 1 ETH
	}

	return bidConfig
}
