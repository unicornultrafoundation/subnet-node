package kvm

import (
	"fmt"
	"math/big"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/node/resource"
)

// ResourceCheckerImpl implements ResourceChecker interface
type ResourceCheckerImpl struct {
	resourceService *resource.Service
	kvmConfig       *KVMConfig
	registry        VMRegistry
	logger          *logrus.Logger
}

// NewResourceCheckerImpl creates a new resource checker instance
func NewResourceCheckerImpl(resourceService *resource.Service, kvmConfig *KVMConfig, registry VMRegistry, logger *logrus.Logger) (ResourceChecker, error) {
	return &ResourceCheckerImpl{
		resourceService: resourceService,
		kvmConfig:       kvmConfig,
		registry:        registry,
		logger:          logger,
	}, nil
}

// CheckAvailableResources validates if requested resources are available
func (rc *ResourceCheckerImpl) CheckAvailableResources(req *ResourceRequirement) error {
	capabilities, err := rc.GetSystemCapabilities()
	if err != nil {
		return fmt.Errorf("failed to get system capabilities: %w", err)
	}

	usage, err := rc.GetResourceUsage()
	if err != nil {
		return fmt.Errorf("failed to get resource usage: %w", err)
	}

	// Check CPU availability
	availableCPU := big.NewInt(int64(capabilities.AvailableCPU))
	usedCPU := big.NewInt(int64(usage.UsedCPU))
	remainingCPU := new(big.Int).Sub(availableCPU, usedCPU)

	if req.CPU.Cmp(remainingCPU) > 0 {
		return fmt.Errorf("insufficient CPU: required %s, available %s", req.CPU.String(), remainingCPU.String())
	}

	// Check memory availability
	remainingMemory := new(big.Int).Sub(capabilities.AvailableMemory, usage.UsedMemory)
	if req.Memory.Cmp(remainingMemory) > 0 {
		return fmt.Errorf("insufficient memory: required %s bytes, available %s bytes", req.Memory.String(), remainingMemory.String())
	}

	// Check storage availability
	remainingStorage := new(big.Int).Sub(capabilities.AvailableStorage, usage.UsedStorage)
	if req.Disk.Cmp(remainingStorage) > 0 {
		return fmt.Errorf("insufficient storage: required %s bytes, available %s bytes", req.Disk.String(), remainingStorage.String())
	}

	// Check VM count limit
	if usage.TotalVMs >= rc.kvmConfig.MaxVMs {
		return fmt.Errorf("maximum VM limit reached: %d/%d", usage.TotalVMs, rc.kvmConfig.MaxVMs)
	}

	return nil
}

// GetSystemCapabilities returns current system capabilities
func (rc *ResourceCheckerImpl) GetSystemCapabilities() (*SystemCapabilities, error) {
	if rc.resourceService == nil {
		return rc.getFallbackCapabilities(), nil
	}

	// Get resource information from the existing service
	resourceInfo, err := rc.resourceService.GetResource()
	if err != nil {
		rc.logger.WithError(err).Warn("Failed to get resource info, using fallback")
		return rc.getFallbackCapabilities(), nil
	}

	// Calculate available resources (subtract reserved amounts)
	totalCPU := resourceInfo.CPU.Count
	availableCPU := int(float64(totalCPU) * (1.0 - rc.kvmConfig.ReservedCPU))

	totalMemory := big.NewInt(int64(resourceInfo.Memory.Total))
	reservedMemoryBytes := new(big.Int).Mul(totalMemory, big.NewInt(int64(rc.kvmConfig.ReservedMemory*100)))
	reservedMemoryBytes.Div(reservedMemoryBytes, big.NewInt(100))
	availableMemory := new(big.Int).Sub(totalMemory, reservedMemoryBytes)

	totalStorage := big.NewInt(int64(resourceInfo.Storage.Total))
	availableStorage := new(big.Int).Set(totalStorage) // Use full storage for now

	return &SystemCapabilities{
		TotalCPU:         totalCPU,
		TotalMemory:      totalMemory,
		TotalStorage:     totalStorage,
		AvailableCPU:     availableCPU,
		AvailableMemory:  availableMemory,
		AvailableStorage: availableStorage,
		MaxVMs:           rc.kvmConfig.MaxVMs,
		SupportedArch:    []string{runtime.GOARCH},
	}, nil
}

// ValidateVMResources validates VM resource requirements
func (rc *ResourceCheckerImpl) ValidateVMResources(cpu, memory, disk *big.Int) error {
	req := &ResourceRequirement{
		CPU:    cpu,
		Memory: memory,
		Disk:   disk,
	}
	return rc.CheckAvailableResources(req)
}

// GetResourceUsage returns current resource usage
func (rc *ResourceCheckerImpl) GetResourceUsage() (*ResourceUsage, error) {
	// Get all VMs from registry
	allVMs, err := rc.registry.ListVMs(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs: %w", err)
	}

	usage := &ResourceUsage{
		UsedCPU:     0,
		UsedMemory:  big.NewInt(0),
		UsedStorage: big.NewInt(0),
		TotalVMs:    len(allVMs),
		RunningVMs:  0,
	}

	// Calculate resource usage from all VMs
	for _, vm := range allVMs {
		// Add CPU usage
		if vm.Resources != nil && vm.Resources.CPU != nil {
			usage.UsedCPU += int(vm.Resources.CPU.Int64())
		}

		// Add memory usage
		if vm.Resources != nil && vm.Resources.Memory != nil {
			usage.UsedMemory.Add(usage.UsedMemory, vm.Resources.Memory)
		}

		// Add storage usage
		if vm.Resources != nil && vm.Resources.Disk != nil {
			usage.UsedStorage.Add(usage.UsedStorage, vm.Resources.Disk)
		}

		// Count running VMs
		if vm.Status == VMStatusRunning {
			usage.RunningVMs++
		}
	}

	return usage, nil
}

// getFallbackCapabilities returns fallback capabilities when resource service is not available
func (rc *ResourceCheckerImpl) getFallbackCapabilities() *SystemCapabilities {
	// Use runtime information as fallback
	cpuCount := runtime.NumCPU()
	availableCPU := int(float64(cpuCount) * (1.0 - rc.kvmConfig.ReservedCPU))

	// Default memory and storage values (these should be configured properly)
	defaultMemory := big.NewInt(8 * 1024 * 1024 * 1024)    // 8GB default
	defaultStorage := big.NewInt(100 * 1024 * 1024 * 1024) // 100GB default

	reservedMemory := new(big.Int).Mul(defaultMemory, big.NewInt(int64(rc.kvmConfig.ReservedMemory*100)))
	reservedMemory.Div(reservedMemory, big.NewInt(100))
	availableMemory := new(big.Int).Sub(defaultMemory, reservedMemory)

	return &SystemCapabilities{
		TotalCPU:         cpuCount,
		TotalMemory:      defaultMemory,
		TotalStorage:     defaultStorage,
		AvailableCPU:     availableCPU,
		AvailableMemory:  availableMemory,
		AvailableStorage: defaultStorage,
		MaxVMs:           rc.kvmConfig.MaxVMs,
		SupportedArch:    []string{runtime.GOARCH},
	}
}
