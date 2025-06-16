package bidengine

import (
	"math/big"

	"github.com/sirupsen/logrus"
)

// calculateResourceCost estimates the cost to fulfill the requirements based on configured prices
// and machine type
func (b *BidEngine) calculateResourceCost(requirements *BidRequirements) *big.Int {
	// Calculate base costs as before
	cpuCost := new(big.Int).Mul(requirements.MinCPUCores, b.bidConfig.CpuCorePrice)

	// Calculate memory cost (convert MB to GB for pricing)
	memGB := new(big.Int).Div(requirements.MinMemoryMB, big.NewInt(1024))
	memCost := new(big.Int).Mul(memGB, b.bidConfig.MemoryGBPrice)

	// Calculate disk cost
	diskCost := new(big.Int).Mul(requirements.MinDiskGB, b.bidConfig.DiskGBPrice)

	// Calculate GPU cost if applicable
	gpuCost := new(big.Int).Mul(requirements.MinGPUCores, b.bidConfig.GpuCorePrice)

	// Calculate network costs if applicable
	uploadCost := new(big.Int).Mul(requirements.MinUploadSpeed, b.bidConfig.UploadSpeedPrice)
	downloadCost := new(big.Int).Mul(requirements.MinDownloadSpeed, b.bidConfig.DownloadSpeedPrice)

	// Sum up all costs for base calculation
	totalCost := new(big.Int).Add(cpuCost, memCost)
	totalCost = new(big.Int).Add(totalCost, diskCost)
	totalCost = new(big.Int).Add(totalCost, gpuCost)
	totalCost = new(big.Int).Add(totalCost, uploadCost)
	totalCost = new(big.Int).Add(totalCost, downloadCost)

	// Apply machine type multiplier if specified
	// Default to 1.0 (no change) if machine type not found in config
	machineTypeId := requirements.MachineType.Int64()
	multiplier, exists := b.bidConfig.MachineTypeMultipliers[machineTypeId]
	if !exists {
		multiplier = big.NewInt(10) // Default multiplier (1.0) if machine type is not configured
	}

	// Apply the multiplier to get the final cost
	if multiplier.Cmp(big.NewInt(10)) != 0 {
		// Convert to float for multiplication with multiplier, then back to big.Int
		multiplierFloat := new(big.Float).SetInt(multiplier)
		multiplierFloat.Quo(multiplierFloat, big.NewFloat(10.0)) // Convert from integer to decimal

		costFloat := new(big.Float).SetInt(totalCost)
		costFloat.Mul(costFloat, multiplierFloat)

		// Convert back to big.Int (truncating any fractional part)
		var finalCost big.Int
		costFloat.Int(&finalCost)
		totalCost = &finalCost
	}

	// Log the cost breakdown for debugging
	b.log.WithFields(logrus.Fields{
		"cpuCost":      cpuCost,
		"memCost":      memCost,
		"diskCost":     diskCost,
		"gpuCost":      gpuCost,
		"uploadCost":   uploadCost,
		"downloadCost": downloadCost,
		"machineType":  machineTypeId,
		"multiplier":   multiplier,
		"totalCost":    totalCost,
	}).Debug("Resource cost breakdown")

	return totalCost
}
