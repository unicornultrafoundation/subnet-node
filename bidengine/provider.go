package bidengine

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/holiman/uint256"
	"github.com/unicornultrafoundation/subnet-node/bidengine/contracts"
)

// Ensure ProviderService implements ProviderServiceInterface
var _ ProviderServiceInterface = (*ProviderService)(nil)

type ProviderService struct {
	providerC *contracts.Provider
}

// NewProviderService creates a new instance of ProviderService
func NewProviderService(provider *contracts.Provider) *ProviderService {
	return &ProviderService{
		providerC: provider,
	}
}

// IsMachineActive checks if a machine is active
func (p *ProviderService) IsMachineActive(opts *bind.CallOpts, providerId, machineId *big.Int) (bool, error) {
	active, err := p.providerC.IsMachineActive(opts, providerId, machineId)
	if err != nil {
		return false, err
	}
	return active, nil
}

// ValidateMachineRequirements checks if a machine meets the specified requirements
func (p *ProviderService) ValidateMachineRequirements(opts *bind.CallOpts, machineType, providerId, machineId *big.Int, minCPUCores, minMemoryMB, minDiskGB, minGPUCores, minUploadSpeed, minDownloadSpeed *big.Int) (bool, error) {
	valid, err := p.providerC.ValidateMachineRequirements(opts, machineType, providerId, machineId, minCPUCores, minMemoryMB, minDiskGB, minGPUCores, minUploadSpeed, minDownloadSpeed)
	if err != nil {
		return false, err
	}
	return valid, nil
}

func (p *ProviderService) GetMachine(opts *bind.CallOpts, providerId, machineId *uint256.Int) (*Machine, error) {
	bcMachine, err := p.providerC.ProviderMachines(opts, providerId.ToBig(), machineId.ToBig())

	if err != nil {
		return nil, err
	}

	machine := &Machine{
		ID:                machineId,
		ProviderId:        providerId,
		CpuCores:          uint256.MustFromBig(bcMachine.CpuCores),
		MemoryMB:          uint256.MustFromBig(bcMachine.MemoryMB),
		DiskGB:            uint256.MustFromBig(bcMachine.DiskGB),
		GpuCores:          uint256.MustFromBig(bcMachine.GpuCores),
		UploadSpeed:       uint256.MustFromBig(bcMachine.UploadSpeed),
		DownloadSpeed:     uint256.MustFromBig(bcMachine.DownloadSpeed),
		Active:            bcMachine.Active,
		Region:            uint256.MustFromBig(bcMachine.Region),
		MachineType:       uint256.MustFromBig(bcMachine.MachineType),
		CpuPricePerSec:    uint256.MustFromBig(bcMachine.CpuPricePerSecond),
		MemoryPricePerSec: uint256.MustFromBig(bcMachine.MemoryPricePerSecond),
		DiskPricePerSec:   uint256.MustFromBig(bcMachine.DiskPricePerSecond),
		GpuPricePerSec:    uint256.MustFromBig(bcMachine.GpuPricePerSecond),

		// Initialize tracking fields
		AllocatedResources: &MachineResources{
			CpuCores: uint256.NewInt(0),
			MemoryMB: uint256.NewInt(0),
			DiskGB:   uint256.NewInt(0),
			GpuCores: uint256.NewInt(0),
		},
		ActiveBids:  make([]*uint256.Int, 0),
		Utilization: 0,
	}

	return machine, nil
}

func (p *ProviderService) GetProvider(opts *bind.CallOpts, providerId *uint256.Int) (*Provider, error) {
	bcProvider, err := p.providerC.GetProvider(opts, providerId.ToBig())
	if err != nil {
		return nil, err
	}

	// Map fields from bcProvider (contracts.SubnetProviderProvider) to your local Provider struct
	provider := &Provider{
		ID:                 providerId,
		Name:               "",
		Description:        "",
		Operator:           bcProvider.Operator,
		Registered:         bcProvider.Registered,
		Reputation:         uint256.MustFromBig(bcProvider.Reputation),
		MachineCount:       uint256.MustFromBig(bcProvider.MachineCount),
		CreatedAt:          uint256.MustFromBig(bcProvider.CreatedAt),
		UpdatedAt:          uint256.MustFromBig(bcProvider.UpdatedAt),
		TotalStaked:        uint256.MustFromBig(bcProvider.TotalStaked),
		PendingWithdrawals: uint256.MustFromBig(bcProvider.PendingWithdrawals),
		SlashedAmount:      uint256.MustFromBig(bcProvider.SlashedAmount),
		TokenId:            uint256.MustFromBig(bcProvider.TokenId),
		Metadata:           bcProvider.Metadata,
		IsSlashed:          bcProvider.IsSlashed,
		IsActive:           bcProvider.IsActive,
	}

	return provider, nil
}

func (p *ProviderService) GetMachinesPaginated(opts *bind.CallOpts, providerId *uint256.Int, startIndex, endIndex *uint256.Int) ([]*Machine, error) {
	bcMachines, err := p.providerC.GetMachinesPaginated(opts, providerId.ToBig(), startIndex.ToBig(), endIndex.ToBig())
	if err != nil {
		return nil, err
	}

	var machines []*Machine
	for idx, bcMachine := range bcMachines {
		machine := &Machine{
			ID:            uint256.NewInt(uint64(idx)), // Use index as ID for simplicity
			ProviderId:    providerId,
			CpuCores:      uint256.MustFromBig(bcMachine.CpuCores),
			MemoryMB:      uint256.MustFromBig(bcMachine.MemoryMB),
			DiskGB:        uint256.MustFromBig(bcMachine.DiskGB),
			GpuCores:      uint256.MustFromBig(bcMachine.GpuCores),
			UploadSpeed:   uint256.MustFromBig(bcMachine.UploadSpeed),
			DownloadSpeed: uint256.MustFromBig(bcMachine.DownloadSpeed),
			Active:        bcMachine.Active,
			Region:        uint256.MustFromBig(bcMachine.Region),
			MachineType:   uint256.MustFromBig(bcMachine.MachineType),

			// Initialize tracking fields
			AllocatedResources: &MachineResources{
				CpuCores: uint256.NewInt(0),
				MemoryMB: uint256.NewInt(0),
				DiskGB:   uint256.NewInt(0),
				GpuCores: uint256.NewInt(0),
			},
			ActiveBids:  make([]*uint256.Int, 0),
			Utilization: 0,
		}
		machines = append(machines, machine)
	}

	return machines, nil
}
