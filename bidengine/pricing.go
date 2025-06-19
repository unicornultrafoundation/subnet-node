package bidengine

import (
	"github.com/holiman/uint256"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/bidengine/types"
)

// calculateMachineCost calculates the cost for running a workload on a specific machine
// for a specified duration (in seconds)
func (b *Service) calculateMachineCost(requirements *types.BidRequirements, machine *types.Machine, durationSeconds uint64) *uint256.Int {
	// Calculate cost per second
	cpuCost := new(uint256.Int).Mul(requirements.MinCPUCores, machine.CpuPricePerSec)
	memCost := new(uint256.Int).Mul(requirements.MinMemoryMB, machine.MemoryPricePerSec)
	diskCost := new(uint256.Int).Mul(requirements.MinDiskGB, machine.DiskPricePerSec)
	gpuCost := new(uint256.Int).Mul(requirements.MinGPUCores, machine.GpuPricePerSec)

	// Calculate total cost per second
	costPerSecond := new(uint256.Int).Add(cpuCost, memCost)
	costPerSecond = new(uint256.Int).Add(costPerSecond, diskCost)
	costPerSecond = new(uint256.Int).Add(costPerSecond, gpuCost)

	// Multiply by duration
	durationUint256 := uint256.NewInt(durationSeconds)
	totalCost := new(uint256.Int).Mul(costPerSecond, durationUint256)

	b.log.WithFields(logrus.Fields{
		"machineId":       machine.ID,
		"durationSeconds": durationSeconds,
		"cpuCostPerSec":   machine.CpuPricePerSec,
		"memCostPerSec":   machine.MemoryPricePerSec,
		"diskCostPerSec":  machine.DiskPricePerSec,
		"gpuCostPerSec":   machine.GpuPricePerSec,
		"totalCost":       totalCost,
	}).Debug("Calculated time-based machine cost")

	return totalCost
}
