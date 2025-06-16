package bidengine

import (
	"context"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/sirupsen/logrus"
)

// findSuitableMachine finds a suitable machine for the given requirements
func (b *BidEngine) findSuitableMachine(ctx context.Context, requirements *BidRequirements) *big.Int {
	// First get the provider ID from config or a previous initialization
	if b.providerId == nil {
		b.log.Error("Provider ID not set, cannot find suitable machine")
		return nil
	}

	// Get provider details to see how many machines there are
	provider, err := b.provider.GetProvider(&bind.CallOpts{Context: ctx}, b.providerId)
	if err != nil {
		b.log.WithError(err).Error("Failed to get provider details")
		return nil
	}

	// Check if the provider is active
	if !provider.IsActive {
		b.log.WithField("providerId", b.providerId).Warn("Provider is not active, cannot allocate machines")
		return nil
	}

	machineCount := provider.MachineCount.Int64()
	if machineCount == 0 {
		b.log.Warn("No machines found for provider")
		return nil
	}

	// Get all machines for our provider
	machines, err := b.provider.GetMachinesPaginated(
		&bind.CallOpts{Context: ctx},
		b.providerId,
		big.NewInt(0),            // Start index
		big.NewInt(machineCount), // End index (exclusive)
	)

	if err != nil {
		b.log.WithError(err).Error("Failed to get machines from provider contract")
		return nil
	}

	b.log.WithField("machineCount", len(machines)).Debug("Retrieved machines from provider contract")

	// Track resource allocation per machine
	type machineResources struct {
		usedCPU      *big.Int
		usedMemory   *big.Int
		usedDisk     *big.Int
		usedGPU      *big.Int
		usedUpload   *big.Int
		usedDownload *big.Int
	}

	machineAllocations := make(map[string]*machineResources)

	// Calculate resource allocation from active bids
	b.bidsMutex.RLock()
	for _, bid := range b.activeBids {
		// Only count bids that are pending or accepted
		if bid.Status != BidStatusRejected && bid.Status != BidStatusExpired {
			machineID := bid.MachineId.String()

			// Initialize if this is the first allocation for this machine
			if _, exists := machineAllocations[machineID]; !exists {
				machineAllocations[machineID] = &machineResources{
					usedCPU:      big.NewInt(0),
					usedMemory:   big.NewInt(0),
					usedDisk:     big.NewInt(0),
					usedGPU:      big.NewInt(0),
					usedUpload:   big.NewInt(0),
					usedDownload: big.NewInt(0),
				}
			}

			// Add the requirements of this bid to the allocated resources
			alloc := machineAllocations[machineID]
			alloc.usedCPU = new(big.Int).Add(alloc.usedCPU, bid.Requirements.MinCPUCores)
			alloc.usedMemory = new(big.Int).Add(alloc.usedMemory, bid.Requirements.MinMemoryMB)
			alloc.usedDisk = new(big.Int).Add(alloc.usedDisk, bid.Requirements.MinDiskGB)
			alloc.usedGPU = new(big.Int).Add(alloc.usedGPU, bid.Requirements.MinGPUCores)
			alloc.usedUpload = new(big.Int).Add(alloc.usedUpload, bid.Requirements.MinUploadSpeed)
			alloc.usedDownload = new(big.Int).Add(alloc.usedDownload, bid.Requirements.MinDownloadSpeed)
		}
	}
	b.bidsMutex.RUnlock()

	// Define machine candidates with their scores
	type machineCandidate struct {
		id    *big.Int
		score float64
	}
	var candidates []machineCandidate

	// Find all machines that meet requirements after accounting for allocated resources
	for i, machine := range machines {
		// Skip inactive machines
		if !machine.Active {
			continue
		}

		machineId := big.NewInt(int64(i)) // Assuming machine IDs are sequential
		machineIdStr := machineId.String()

		// Get allocated resources, or initialize to zero if none
		var allocated *machineResources
		if alloc, exists := machineAllocations[machineIdStr]; exists {
			allocated = alloc
		} else {
			allocated = &machineResources{
				usedCPU:      big.NewInt(0),
				usedMemory:   big.NewInt(0),
				usedDisk:     big.NewInt(0),
				usedGPU:      big.NewInt(0),
				usedUpload:   big.NewInt(0),
				usedDownload: big.NewInt(0),
			}
		}

		// Calculate remaining resources
		availableCPU := new(big.Int).Sub(machine.CpuCores, allocated.usedCPU)
		availableMemory := new(big.Int).Sub(machine.MemoryMB, allocated.usedMemory)
		availableDisk := new(big.Int).Sub(machine.DiskGB, allocated.usedDisk)
		availableGPU := new(big.Int).Sub(machine.GpuCores, allocated.usedGPU)
		availableUpload := new(big.Int).Sub(machine.UploadSpeed, allocated.usedUpload)
		availableDownload := new(big.Int).Sub(machine.DownloadSpeed, allocated.usedDownload)

		// Log resource availability
		b.log.WithFields(logrus.Fields{
			"machineId":       machineId,
			"availableCPU":    availableCPU,
			"availableMemory": availableMemory,
			"availableDisk":   availableDisk,
			"totalCPU":        machine.CpuCores,
			"totalMemory":     machine.MemoryMB,
			"totalDisk":       machine.DiskGB,
		}).Debug("Machine resource availability")

		// Check if machine has enough remaining resources for this order
		if availableCPU.Cmp(requirements.MinCPUCores) < 0 ||
			availableMemory.Cmp(requirements.MinMemoryMB) < 0 ||
			availableDisk.Cmp(requirements.MinDiskGB) < 0 ||
			availableGPU.Cmp(requirements.MinGPUCores) < 0 ||
			availableUpload.Cmp(requirements.MinUploadSpeed) < 0 ||
			availableDownload.Cmp(requirements.MinDownloadSpeed) < 0 ||
			machine.Region.Cmp(requirements.Region) != 0 ||
			machine.MachineType.Cmp(requirements.MachineType) != 0 {
			b.log.WithField("machineId", machineId).Debug("Machine does not have enough remaining resources")
			continue
		}

		// Calculate a score based on remaining resources - lower is better
		// We aim for machines that just meet the requirements without wasting resources
		cpuRatio := float64(availableCPU.Int64()) / float64(requirements.MinCPUCores.Int64())
		memRatio := float64(availableMemory.Int64()) / float64(requirements.MinMemoryMB.Int64())
		diskRatio := float64(availableDisk.Int64()) / float64(requirements.MinDiskGB.Int64())

		// Calculate how much each resource exceeds requirements (0 is perfect)
		cpuExcess := cpuRatio - 1.0
		memExcess := memRatio - 1.0
		diskExcess := diskRatio - 1.0

		// Calculate a combined score - the smaller the better
		// We square the values to penalize larger excesses more
		score := cpuExcess*cpuExcess + memExcess*memExcess + diskExcess*diskExcess

		candidates = append(candidates, machineCandidate{
			id:    machineId,
			score: score,
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

	b.log.WithFields(logrus.Fields{
		"machineId":       candidates[0].id,
		"score":           candidates[0].score,
		"totalCandidates": len(candidates),
	}).Info("Selected suitable machine for bid")

	return candidates[0].id
}
