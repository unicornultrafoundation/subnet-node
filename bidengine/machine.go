package bidengine

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/holiman/uint256"
	"github.com/sirupsen/logrus"
)

// machineCache provides in-memory caching of machine information
type machineCache struct {
	machines    map[string]*Machine // Map of providerId:machineId -> Machine
	lastUpdated time.Time
	mutex       sync.RWMutex
}

// Initialize the machine cache
var mCache = &machineCache{
	machines: make(map[string]*Machine),
}

// Get machines for a provider with caching
func (b *Service) getMachines(ctx context.Context, forceRefresh bool) ([]*Machine, error) {
	// Create a cache key prefix
	providerKey := b.providerId.String()

	// Check if we need to refresh the cache
	mCache.mutex.RLock()
	cacheExpired := time.Since(mCache.lastUpdated) > 5*time.Minute
	mCache.mutex.RUnlock()

	if !forceRefresh && !cacheExpired {
		// Return machines from cache if available for this provider
		var providerMachines []*Machine

		mCache.mutex.RLock()
		for key, machine := range mCache.machines {
			if key[:len(providerKey)] == providerKey {
				// Create a copy to avoid race conditions
				machineCopy := *machine
				providerMachines = append(providerMachines, &machineCopy)
			}
		}
		mCache.mutex.RUnlock()

		if len(providerMachines) > 0 {
			b.log.WithField("machineCount", len(providerMachines)).Debug("Retrieved machines from cache")
			return providerMachines, nil
		}
	}

	// Get provider details to see how many machines there are
	provider, err := b.provider.GetProvider(&bind.CallOpts{Context: ctx}, b.providerId)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider details: %v", err)
	}

	// Check if the provider is active
	if !provider.IsActive {
		return nil, fmt.Errorf("provider %s is not active", b.providerId.String())
	}

	if provider.MachineCount.IsZero() {
		return []*Machine{}, nil
	}

	// Get all machines for our provider
	machines, err := b.provider.GetMachinesPaginated(
		&bind.CallOpts{Context: ctx},
		b.providerId,
		uint256.NewInt(0),     // Start index
		provider.MachineCount, // End index (exclusive)
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get machines from provider contract: %v", err)
	}

	// Convert blockchain machines to our internal machine representation

	for _, machine := range machines {
		// Update machine's allocated resources based on active bids
		b.updateMachineAllocatedResources(ctx, machine)

		// Update cache
		mCache.mutex.Lock()
		cacheKey := fmt.Sprintf("%s:%s", b.providerId.String(), machine.ID.String())
		mCache.machines[cacheKey] = machine
		mCache.mutex.Unlock()
	}

	// Update cache timestamp
	mCache.mutex.Lock()
	mCache.lastUpdated = time.Now()
	mCache.mutex.Unlock()

	b.log.WithField("machineCount", len(machines)).Debug("Retrieved machines from blockchain")
	return machines, nil
}

// updateMachineAllocatedResources updates a machine's allocated resources based on active bids
func (b *Service) updateMachineAllocatedResources(ctx context.Context, machine *Machine) {
	// Reset allocated resources
	machine.AllocatedResources = &MachineResources{
		CpuCores: uint256.NewInt(0),
		MemoryMB: uint256.NewInt(0),
		DiskGB:   uint256.NewInt(0),
		GpuCores: uint256.NewInt(0),
	}

	machine.ActiveBids = make([]*uint256.Int, 0)

	// Calculate resource allocation from active bids
	b.bidsMutex.RLock()
	for _, bid := range b.activeBids {
		// Only count bids that are pending or accepted for this machine
		if (bid.Status == BidStatusPending || bid.Status == BidStatusAccepted) &&
			bid.MachineId.Eq(machine.ID) && bid.ProviderId.Eq(machine.ProviderId) {

			// Add requirements to allocated resources
			machine.AllocatedResources.CpuCores.Add(machine.AllocatedResources.CpuCores, bid.Requirements.MinCPUCores)
			machine.AllocatedResources.MemoryMB.Add(machine.AllocatedResources.MemoryMB, bid.Requirements.MinMemoryMB)
			machine.AllocatedResources.DiskGB.Add(machine.AllocatedResources.DiskGB, bid.Requirements.MinDiskGB)
			machine.AllocatedResources.GpuCores.Add(machine.AllocatedResources.GpuCores, bid.Requirements.MinGPUCores)

			// Add to active bids list
			machine.ActiveBids = append(machine.ActiveBids, bid.ID)
		}
	}
	b.bidsMutex.RUnlock()

	// Update utilization percentage
	machine.CalculateUtilization()
}

// GetMachine retrieves information about a specific machine
func (b *Service) GetMachine(ctx context.Context, providerId, machineId *uint256.Int) (*Machine, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("%s:%s", providerId.String(), machineId.String())

	mCache.mutex.RLock()
	cachedMachine, exists := mCache.machines[cacheKey]
	cacheExpired := time.Since(mCache.lastUpdated) > 5*time.Minute
	mCache.mutex.RUnlock()

	if exists && !cacheExpired {
		// Update allocated resources before returning
		b.updateMachineAllocatedResources(ctx, cachedMachine)
		return cachedMachine, nil
	}

	// Get machine details directly
	machine, err := b.provider.GetMachine(&bind.CallOpts{Context: ctx}, providerId, machineId)
	if err != nil {
		return nil, fmt.Errorf("failed to get machine details: %v", err)
	}

	// Update machine's allocated resources based on active bids
	b.updateMachineAllocatedResources(ctx, machine)

	// Update cache
	mCache.mutex.Lock()
	mCache.machines[cacheKey] = machine
	mCache.mutex.Unlock()

	return machine, nil
}

// findSuitableMachine finds a suitable machine for the given requirements
func (b *Service) findSuitableMachine(ctx context.Context, requirements *BidRequirements) *Machine {
	// First get the provider ID from config or a previous initialization
	if b.providerId == nil {
		b.log.Error("Provider ID not set, cannot find suitable machine")
		return nil
	}

	// Check if we should force refresh the machine data
	// We force refresh if this is an important bid or we haven't updated in a while
	mCache.mutex.RLock()
	forceRefresh := time.Since(mCache.lastUpdated) > time.Minute*2
	mCache.mutex.RUnlock()

	// Get all machines for the provider - force refresh if needed
	machines, err := b.getMachines(ctx, forceRefresh)
	if err != nil {
		b.log.WithError(err).Error("Failed to get machines")
		return nil
	}

	if len(machines) == 0 {
		b.log.Warn("No machines found for provider")
		return nil
	}

	b.log.WithField("requirementsCpu", requirements.MinCPUCores.String()).
		WithField("requirementsMemory", requirements.MinMemoryMB.String()).
		WithField("requirementsDisk", requirements.MinDiskGB.String()).
		WithField("region", requirements.Region.String()).
		WithField("machineType", requirements.MachineType.String()).
		Debug("Looking for machine with requirements")

	// Define machine candidates with their scores
	type machineCandidate struct {
		machine *Machine
		score   float64
	}
	var candidates []machineCandidate

	// Find all machines that meet requirements after accounting for allocated resources
	for _, machine := range machines {
		// Skip inactive machines
		if !machine.Active {
			continue
		}

		// Check if this machine has the right region and type
		if !machine.Region.Eq(requirements.Region) || !machine.MachineType.Eq(requirements.MachineType) {
			b.log.WithFields(logrus.Fields{
				"machineId":           machine.ID,
				"machineRegion":       machine.Region,
				"machineType":         machine.MachineType,
				"requiredRegion":      requirements.Region,
				"requiredMachineType": requirements.MachineType,
			}).Debug("Machine region or type mismatch")
			continue
		}

		// For each machine, ensure allocated resources are up-to-date
		b.updateMachineAllocatedResources(ctx, machine)

		// Get remaining resources
		remaining := machine.CalculateRemainingResources()

		// Log resource availability
		b.log.WithFields(logrus.Fields{
			"machineId":       machine.ID.String(),
			"availableCPU":    remaining.CpuCores.String(),
			"availableMemory": remaining.MemoryMB.String(),
			"availableDisk":   remaining.DiskGB.String(),
			"utilization":     machine.Utilization,
		}).Debug("Machine resource availability")

		// Check if machine has enough remaining resources for this order
		if remaining.CpuCores.Lt(requirements.MinCPUCores) ||
			remaining.MemoryMB.Lt(requirements.MinMemoryMB) ||
			remaining.DiskGB.Lt(requirements.MinDiskGB) ||
			remaining.GpuCores.Lt(requirements.MinGPUCores) {
			b.log.WithField("machineId", machine.ID).Debug("Machine does not have enough remaining resources")
			continue
		}

		// Calculate a score based on remaining resources - lower is better
		// We aim for machines that just meet the requirements without wasting resources
		cpuRatio := float64(remaining.CpuCores.Uint64()) / float64(requirements.MinCPUCores.Uint64())
		memRatio := float64(remaining.MemoryMB.Uint64()) / float64(requirements.MinMemoryMB.Uint64())
		diskRatio := float64(remaining.DiskGB.Uint64()) / float64(requirements.MinDiskGB.Uint64())

		// Calculate how much each resource exceeds requirements (0 is perfect)
		cpuExcess := cpuRatio - 1.0
		memExcess := memRatio - 1.0
		diskExcess := diskRatio - 1.0

		// Calculate a combined score - the smaller the better
		// We square the values to penalize larger excesses more
		// Also factor in current utilization - prefer less utilized machines
		utilizationFactor := machine.Utilization / 100.0 // Convert to 0-1 range
		resourceScore := cpuExcess*cpuExcess + memExcess*memExcess + diskExcess*diskExcess

		// Final score combines resource fit and utilization
		// Weight resource fit more heavily (80%) than utilization (20%)
		score := resourceScore*0.8 + utilizationFactor*0.2

		candidates = append(candidates, machineCandidate{
			machine: machine,
			score:   score,
		})
	}

	if len(candidates) == 0 {
		b.log.Warn("No suitable machines found for requirements")
		return nil
	}

	// Sort candidates by score (lower is better)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score < candidates[j].score
	})

	selectedMachine := candidates[0].machine

	b.log.WithFields(logrus.Fields{
		"machineId":       selectedMachine.ID,
		"score":           candidates[0].score,
		"utilization":     selectedMachine.Utilization,
		"totalCandidates": len(candidates),
		"cpuCores":        selectedMachine.CpuCores,
		"memoryMB":        selectedMachine.MemoryMB,
		"diskGB":          selectedMachine.DiskGB,
	}).Info("Selected suitable machine for bid")

	return selectedMachine
}
