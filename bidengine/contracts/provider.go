// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// SubnetProviderMachine is an auto generated low-level Go binding around an user-defined struct.
type SubnetProviderMachine struct {
	Active               bool
	MachineType          *big.Int
	Region               *big.Int
	CpuCores             *big.Int
	GpuCores             *big.Int
	GpuMemory            *big.Int
	MemoryMB             *big.Int
	DiskGB               *big.Int
	UploadSpeed          *big.Int
	DownloadSpeed        *big.Int
	CreatedAt            *big.Int
	UpdatedAt            *big.Int
	StakeAmount          *big.Int
	RemovedAt            *big.Int
	UnlockTime           *big.Int
	WithdrawalProcessed  bool
	Metadata             string
	CpuPricePerSecond    *big.Int
	GpuPricePerSecond    *big.Int
	MemoryPricePerSecond *big.Int
	DiskPricePerSecond   *big.Int
}

// SubnetProviderProvider is an auto generated low-level Go binding around an user-defined struct.
type SubnetProviderProvider struct {
	Operator           common.Address
	Registered         bool
	Reputation         *big.Int
	MachineCount       *big.Int
	CreatedAt          *big.Int
	UpdatedAt          *big.Int
	TotalStaked        *big.Int
	PendingWithdrawals *big.Int
	SlashedAmount      *big.Int
	TokenId            *big.Int
	Metadata           string
	IsSlashed          bool
	IsActive           bool
	Verified           bool
}

// ProviderMetaData contains all meta data concerning the Provider contract.
var ProviderMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"AddressEmptyCode\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"AddressInsufficientBalance\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"ERC721IncorrectOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ERC721InsufficientApproval\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"approver\",\"type\":\"address\"}],\"name\":\"ERC721InvalidApprover\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"ERC721InvalidOperator\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"ERC721InvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"}],\"name\":\"ERC721InvalidReceiver\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"ERC721InvalidSender\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ERC721NonexistentToken\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"SafeERC20FailedOperation\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"approved\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_fromTokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_toTokenId\",\"type\":\"uint256\"}],\"name\":\"BatchMetadataUpdate\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"oldPeriod\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newPeriod\",\"type\":\"uint256\"}],\"name\":\"LockPeriodUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"machineId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"stakedAmount\",\"type\":\"uint256\"}],\"name\":\"MachineAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"machineId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"unlocktime\",\"type\":\"uint256\"}],\"name\":\"MachineRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"machineId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"cpuPricePerSecond\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"gpuPricePerSecond\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"memoryPricePerSecond\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"diskPricePerSecond\",\"type\":\"uint256\"}],\"name\":\"MachineResourcePriceUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"machineId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"additionalStake\",\"type\":\"uint256\"}],\"name\":\"MachineUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"MetadataUpdate\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newReputation\",\"type\":\"uint256\"}],\"name\":\"ProviderReputationUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"}],\"name\":\"ProviderUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"verified\",\"type\":\"bool\"}],\"name\":\"ProviderVerified\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"baseAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"cpuRate\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"gpuRate\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"memoryRate\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"diskRate\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"uploadSpeedRate\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"downloadSpeedRate\",\"type\":\"uint256\"}],\"name\":\"StakeParametersUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"machineId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"name\":\"StakeSlashed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"machineId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"StakeWithdrawn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"machineType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"region\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"cpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuMemory\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"memoryMB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"diskGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"uploadSpeed\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"downloadSpeed\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"cpuPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"memoryPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"diskPricePerSecond\",\"type\":\"uint256\"}],\"name\":\"addMachine\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"baseStakeAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"cpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"memoryMB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"diskGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"uploadSpeed\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"downloadSpeed\",\"type\":\"uint256\"}],\"name\":\"calculateRequiredStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"machineId\",\"type\":\"uint256\"}],\"name\":\"claimWithdrawal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"cpuStakeRate\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"diskStakeRate\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"downloadSpeedStakeRate\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"start\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit\",\"type\":\"uint256\"}],\"name\":\"getActiveMachinesPaginated\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"machineType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"region\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"cpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuMemory\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"memoryMB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"diskGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"uploadSpeed\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"downloadSpeed\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"updatedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stakeAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"removedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"unlockTime\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"withdrawalProcessed\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"cpuPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"memoryPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"diskPricePerSecond\",\"type\":\"uint256\"}],\"internalType\":\"structSubnetProvider.Machine[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getApproved\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"machineId\",\"type\":\"uint256\"}],\"name\":\"getMachineResourcePrice\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"cpuPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"memoryPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"diskPricePerSecond\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"}],\"name\":\"getMachines\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"machineType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"region\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"cpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuMemory\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"memoryMB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"diskGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"uploadSpeed\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"downloadSpeed\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"updatedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stakeAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"removedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"unlockTime\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"withdrawalProcessed\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"cpuPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"memoryPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"diskPricePerSecond\",\"type\":\"uint256\"}],\"internalType\":\"structSubnetProvider.Machine[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"start\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"end\",\"type\":\"uint256\"}],\"name\":\"getMachinesPaginated\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"machineType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"region\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"cpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuMemory\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"memoryMB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"diskGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"uploadSpeed\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"downloadSpeed\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"updatedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stakeAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"removedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"unlockTime\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"withdrawalProcessed\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"cpuPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"memoryPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"diskPricePerSecond\",\"type\":\"uint256\"}],\"internalType\":\"structSubnetProvider.Machine[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"}],\"name\":\"getProvider\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"registered\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"reputation\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"machineCount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"updatedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalStaked\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pendingWithdrawals\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"slashedAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"},{\"internalType\":\"bool\",\"name\":\"isSlashed\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"isActive\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"verified\",\"type\":\"bool\"}],\"internalType\":\"structSubnetProvider.Provider\",\"name\":\"provider\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"}],\"name\":\"getProviderOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"}],\"name\":\"getProviderReputation\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"}],\"name\":\"getProviderSlashedAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"gpuStakeRate\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_stakingToken\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"nftName\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"nftSymbol\",\"type\":\"string\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"machineId\",\"type\":\"uint256\"}],\"name\":\"isMachineActive\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"isProviderOperatorOrOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"}],\"name\":\"isVerified\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"lockPeriod\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"memoryStakeRate\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"providerMachines\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"machineType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"region\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"cpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuMemory\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"memoryMB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"diskGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"uploadSpeed\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"downloadSpeed\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"updatedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stakeAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"removedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"unlockTime\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"withdrawalProcessed\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"cpuPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"memoryPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"diskPricePerSecond\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"providers\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"registered\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"reputation\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"machineCount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"updatedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalStaked\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pendingWithdrawals\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"slashedAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"},{\"internalType\":\"bool\",\"name\":\"isSlashed\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"isActive\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"verified\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"}],\"name\":\"registerProvider\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"providerMetadata\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"machineType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"region\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"cpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuMemory\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"memoryMB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"diskGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"uploadSpeed\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"downloadSpeed\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"machineMetadata\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"cpuPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"memoryPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"diskPricePerSecond\",\"type\":\"uint256\"}],\"name\":\"registerProviderWithMachine\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"machineId\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"machineId\",\"type\":\"uint256\"}],\"name\":\"removeMachine\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newLockPeriod\",\"type\":\"uint256\"}],\"name\":\"setLockPeriod\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"machineId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"cpuPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"memoryPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"diskPricePerSecond\",\"type\":\"uint256\"}],\"name\":\"setMachineResourcePrice\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"newOperator\",\"type\":\"address\"}],\"name\":\"setProviderOperator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"newReputation\",\"type\":\"uint256\"}],\"name\":\"setProviderReputation\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"verified_\",\"type\":\"bool\"}],\"name\":\"setProviderVerified\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newBaseStakeAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"newCpuStakeRate\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"newGpuStakeRate\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"newMemoryStakeRate\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"newDiskStakeRate\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"newUploadSpeedStakeRate\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"newDownloadSpeedStakeRate\",\"type\":\"uint256\"}],\"name\":\"setStakeParameters\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"machineId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"name\":\"slashStake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"stakingToken\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"tokenURI\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSlashed\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"machineId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"cpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuMemory\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"memoryMB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"diskGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"uploadSpeed\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"downloadSpeed\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"cpuPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"memoryPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"diskPricePerSecond\",\"type\":\"uint256\"}],\"name\":\"updateMachine\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"}],\"name\":\"updateProviderInfo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"uploadSpeedStakeRate\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"machineType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"machineId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minCpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minMemoryMB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minDiskGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minGpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minUploadSpeed\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minDownloadSpeed\",\"type\":\"uint256\"}],\"name\":\"validateMachineRequirements\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdrawSlashedFunds\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// ProviderABI is the input ABI used to generate the binding from.
// Deprecated: Use ProviderMetaData.ABI instead.
var ProviderABI = ProviderMetaData.ABI

// Provider is an auto generated Go binding around an Ethereum contract.
type Provider struct {
	ProviderCaller     // Read-only binding to the contract
	ProviderTransactor // Write-only binding to the contract
	ProviderFilterer   // Log filterer for contract events
}

// ProviderCaller is an auto generated read-only Go binding around an Ethereum contract.
type ProviderCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProviderTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ProviderTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProviderFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ProviderFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProviderSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ProviderSession struct {
	Contract     *Provider         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ProviderCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ProviderCallerSession struct {
	Contract *ProviderCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// ProviderTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ProviderTransactorSession struct {
	Contract     *ProviderTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ProviderRaw is an auto generated low-level Go binding around an Ethereum contract.
type ProviderRaw struct {
	Contract *Provider // Generic contract binding to access the raw methods on
}

// ProviderCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ProviderCallerRaw struct {
	Contract *ProviderCaller // Generic read-only contract binding to access the raw methods on
}

// ProviderTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ProviderTransactorRaw struct {
	Contract *ProviderTransactor // Generic write-only contract binding to access the raw methods on
}

// NewProvider creates a new instance of Provider, bound to a specific deployed contract.
func NewProvider(address common.Address, backend bind.ContractBackend) (*Provider, error) {
	contract, err := bindProvider(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Provider{ProviderCaller: ProviderCaller{contract: contract}, ProviderTransactor: ProviderTransactor{contract: contract}, ProviderFilterer: ProviderFilterer{contract: contract}}, nil
}

// NewProviderCaller creates a new read-only instance of Provider, bound to a specific deployed contract.
func NewProviderCaller(address common.Address, caller bind.ContractCaller) (*ProviderCaller, error) {
	contract, err := bindProvider(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ProviderCaller{contract: contract}, nil
}

// NewProviderTransactor creates a new write-only instance of Provider, bound to a specific deployed contract.
func NewProviderTransactor(address common.Address, transactor bind.ContractTransactor) (*ProviderTransactor, error) {
	contract, err := bindProvider(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ProviderTransactor{contract: contract}, nil
}

// NewProviderFilterer creates a new log filterer instance of Provider, bound to a specific deployed contract.
func NewProviderFilterer(address common.Address, filterer bind.ContractFilterer) (*ProviderFilterer, error) {
	contract, err := bindProvider(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ProviderFilterer{contract: contract}, nil
}

// bindProvider binds a generic wrapper to an already deployed contract.
func bindProvider(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ProviderMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Provider *ProviderRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Provider.Contract.ProviderCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Provider *ProviderRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Provider.Contract.ProviderTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Provider *ProviderRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Provider.Contract.ProviderTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Provider *ProviderCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Provider.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Provider *ProviderTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Provider.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Provider *ProviderTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Provider.Contract.contract.Transact(opts, method, params...)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_Provider *ProviderCaller) BalanceOf(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "balanceOf", owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_Provider *ProviderSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _Provider.Contract.BalanceOf(&_Provider.CallOpts, owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_Provider *ProviderCallerSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _Provider.Contract.BalanceOf(&_Provider.CallOpts, owner)
}

// BaseStakeAmount is a free data retrieval call binding the contract method 0x71129559.
//
// Solidity: function baseStakeAmount() view returns(uint256)
func (_Provider *ProviderCaller) BaseStakeAmount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "baseStakeAmount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BaseStakeAmount is a free data retrieval call binding the contract method 0x71129559.
//
// Solidity: function baseStakeAmount() view returns(uint256)
func (_Provider *ProviderSession) BaseStakeAmount() (*big.Int, error) {
	return _Provider.Contract.BaseStakeAmount(&_Provider.CallOpts)
}

// BaseStakeAmount is a free data retrieval call binding the contract method 0x71129559.
//
// Solidity: function baseStakeAmount() view returns(uint256)
func (_Provider *ProviderCallerSession) BaseStakeAmount() (*big.Int, error) {
	return _Provider.Contract.BaseStakeAmount(&_Provider.CallOpts)
}

// CalculateRequiredStake is a free data retrieval call binding the contract method 0x00812be7.
//
// Solidity: function calculateRequiredStake(uint256 cpuCores, uint256 gpuCores, uint256 memoryMB, uint256 diskGB, uint256 uploadSpeed, uint256 downloadSpeed) view returns(uint256)
func (_Provider *ProviderCaller) CalculateRequiredStake(opts *bind.CallOpts, cpuCores *big.Int, gpuCores *big.Int, memoryMB *big.Int, diskGB *big.Int, uploadSpeed *big.Int, downloadSpeed *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "calculateRequiredStake", cpuCores, gpuCores, memoryMB, diskGB, uploadSpeed, downloadSpeed)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CalculateRequiredStake is a free data retrieval call binding the contract method 0x00812be7.
//
// Solidity: function calculateRequiredStake(uint256 cpuCores, uint256 gpuCores, uint256 memoryMB, uint256 diskGB, uint256 uploadSpeed, uint256 downloadSpeed) view returns(uint256)
func (_Provider *ProviderSession) CalculateRequiredStake(cpuCores *big.Int, gpuCores *big.Int, memoryMB *big.Int, diskGB *big.Int, uploadSpeed *big.Int, downloadSpeed *big.Int) (*big.Int, error) {
	return _Provider.Contract.CalculateRequiredStake(&_Provider.CallOpts, cpuCores, gpuCores, memoryMB, diskGB, uploadSpeed, downloadSpeed)
}

// CalculateRequiredStake is a free data retrieval call binding the contract method 0x00812be7.
//
// Solidity: function calculateRequiredStake(uint256 cpuCores, uint256 gpuCores, uint256 memoryMB, uint256 diskGB, uint256 uploadSpeed, uint256 downloadSpeed) view returns(uint256)
func (_Provider *ProviderCallerSession) CalculateRequiredStake(cpuCores *big.Int, gpuCores *big.Int, memoryMB *big.Int, diskGB *big.Int, uploadSpeed *big.Int, downloadSpeed *big.Int) (*big.Int, error) {
	return _Provider.Contract.CalculateRequiredStake(&_Provider.CallOpts, cpuCores, gpuCores, memoryMB, diskGB, uploadSpeed, downloadSpeed)
}

// CpuStakeRate is a free data retrieval call binding the contract method 0xff3d5528.
//
// Solidity: function cpuStakeRate() view returns(uint256)
func (_Provider *ProviderCaller) CpuStakeRate(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "cpuStakeRate")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CpuStakeRate is a free data retrieval call binding the contract method 0xff3d5528.
//
// Solidity: function cpuStakeRate() view returns(uint256)
func (_Provider *ProviderSession) CpuStakeRate() (*big.Int, error) {
	return _Provider.Contract.CpuStakeRate(&_Provider.CallOpts)
}

// CpuStakeRate is a free data retrieval call binding the contract method 0xff3d5528.
//
// Solidity: function cpuStakeRate() view returns(uint256)
func (_Provider *ProviderCallerSession) CpuStakeRate() (*big.Int, error) {
	return _Provider.Contract.CpuStakeRate(&_Provider.CallOpts)
}

// DiskStakeRate is a free data retrieval call binding the contract method 0xd9486e35.
//
// Solidity: function diskStakeRate() view returns(uint256)
func (_Provider *ProviderCaller) DiskStakeRate(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "diskStakeRate")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DiskStakeRate is a free data retrieval call binding the contract method 0xd9486e35.
//
// Solidity: function diskStakeRate() view returns(uint256)
func (_Provider *ProviderSession) DiskStakeRate() (*big.Int, error) {
	return _Provider.Contract.DiskStakeRate(&_Provider.CallOpts)
}

// DiskStakeRate is a free data retrieval call binding the contract method 0xd9486e35.
//
// Solidity: function diskStakeRate() view returns(uint256)
func (_Provider *ProviderCallerSession) DiskStakeRate() (*big.Int, error) {
	return _Provider.Contract.DiskStakeRate(&_Provider.CallOpts)
}

// DownloadSpeedStakeRate is a free data retrieval call binding the contract method 0xfce9d999.
//
// Solidity: function downloadSpeedStakeRate() view returns(uint256)
func (_Provider *ProviderCaller) DownloadSpeedStakeRate(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "downloadSpeedStakeRate")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DownloadSpeedStakeRate is a free data retrieval call binding the contract method 0xfce9d999.
//
// Solidity: function downloadSpeedStakeRate() view returns(uint256)
func (_Provider *ProviderSession) DownloadSpeedStakeRate() (*big.Int, error) {
	return _Provider.Contract.DownloadSpeedStakeRate(&_Provider.CallOpts)
}

// DownloadSpeedStakeRate is a free data retrieval call binding the contract method 0xfce9d999.
//
// Solidity: function downloadSpeedStakeRate() view returns(uint256)
func (_Provider *ProviderCallerSession) DownloadSpeedStakeRate() (*big.Int, error) {
	return _Provider.Contract.DownloadSpeedStakeRate(&_Provider.CallOpts)
}

// GetActiveMachinesPaginated is a free data retrieval call binding the contract method 0x32f892d5.
//
// Solidity: function getActiveMachinesPaginated(uint256 providerId, uint256 start, uint256 limit) view returns((bool,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,bool,string,uint256,uint256,uint256,uint256)[])
func (_Provider *ProviderCaller) GetActiveMachinesPaginated(opts *bind.CallOpts, providerId *big.Int, start *big.Int, limit *big.Int) ([]SubnetProviderMachine, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "getActiveMachinesPaginated", providerId, start, limit)

	if err != nil {
		return *new([]SubnetProviderMachine), err
	}

	out0 := *abi.ConvertType(out[0], new([]SubnetProviderMachine)).(*[]SubnetProviderMachine)

	return out0, err

}

// GetActiveMachinesPaginated is a free data retrieval call binding the contract method 0x32f892d5.
//
// Solidity: function getActiveMachinesPaginated(uint256 providerId, uint256 start, uint256 limit) view returns((bool,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,bool,string,uint256,uint256,uint256,uint256)[])
func (_Provider *ProviderSession) GetActiveMachinesPaginated(providerId *big.Int, start *big.Int, limit *big.Int) ([]SubnetProviderMachine, error) {
	return _Provider.Contract.GetActiveMachinesPaginated(&_Provider.CallOpts, providerId, start, limit)
}

// GetActiveMachinesPaginated is a free data retrieval call binding the contract method 0x32f892d5.
//
// Solidity: function getActiveMachinesPaginated(uint256 providerId, uint256 start, uint256 limit) view returns((bool,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,bool,string,uint256,uint256,uint256,uint256)[])
func (_Provider *ProviderCallerSession) GetActiveMachinesPaginated(providerId *big.Int, start *big.Int, limit *big.Int) ([]SubnetProviderMachine, error) {
	return _Provider.Contract.GetActiveMachinesPaginated(&_Provider.CallOpts, providerId, start, limit)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_Provider *ProviderCaller) GetApproved(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "getApproved", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_Provider *ProviderSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _Provider.Contract.GetApproved(&_Provider.CallOpts, tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_Provider *ProviderCallerSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _Provider.Contract.GetApproved(&_Provider.CallOpts, tokenId)
}

// GetMachineResourcePrice is a free data retrieval call binding the contract method 0x1077399e.
//
// Solidity: function getMachineResourcePrice(uint256 providerId, uint256 machineId) view returns(uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond)
func (_Provider *ProviderCaller) GetMachineResourcePrice(opts *bind.CallOpts, providerId *big.Int, machineId *big.Int) (struct {
	CpuPricePerSecond    *big.Int
	GpuPricePerSecond    *big.Int
	MemoryPricePerSecond *big.Int
	DiskPricePerSecond   *big.Int
}, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "getMachineResourcePrice", providerId, machineId)

	outstruct := new(struct {
		CpuPricePerSecond    *big.Int
		GpuPricePerSecond    *big.Int
		MemoryPricePerSecond *big.Int
		DiskPricePerSecond   *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.CpuPricePerSecond = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.GpuPricePerSecond = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.MemoryPricePerSecond = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.DiskPricePerSecond = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetMachineResourcePrice is a free data retrieval call binding the contract method 0x1077399e.
//
// Solidity: function getMachineResourcePrice(uint256 providerId, uint256 machineId) view returns(uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond)
func (_Provider *ProviderSession) GetMachineResourcePrice(providerId *big.Int, machineId *big.Int) (struct {
	CpuPricePerSecond    *big.Int
	GpuPricePerSecond    *big.Int
	MemoryPricePerSecond *big.Int
	DiskPricePerSecond   *big.Int
}, error) {
	return _Provider.Contract.GetMachineResourcePrice(&_Provider.CallOpts, providerId, machineId)
}

// GetMachineResourcePrice is a free data retrieval call binding the contract method 0x1077399e.
//
// Solidity: function getMachineResourcePrice(uint256 providerId, uint256 machineId) view returns(uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond)
func (_Provider *ProviderCallerSession) GetMachineResourcePrice(providerId *big.Int, machineId *big.Int) (struct {
	CpuPricePerSecond    *big.Int
	GpuPricePerSecond    *big.Int
	MemoryPricePerSecond *big.Int
	DiskPricePerSecond   *big.Int
}, error) {
	return _Provider.Contract.GetMachineResourcePrice(&_Provider.CallOpts, providerId, machineId)
}

// GetMachines is a free data retrieval call binding the contract method 0x8d49e78d.
//
// Solidity: function getMachines(uint256 providerId) view returns((bool,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,bool,string,uint256,uint256,uint256,uint256)[])
func (_Provider *ProviderCaller) GetMachines(opts *bind.CallOpts, providerId *big.Int) ([]SubnetProviderMachine, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "getMachines", providerId)

	if err != nil {
		return *new([]SubnetProviderMachine), err
	}

	out0 := *abi.ConvertType(out[0], new([]SubnetProviderMachine)).(*[]SubnetProviderMachine)

	return out0, err

}

// GetMachines is a free data retrieval call binding the contract method 0x8d49e78d.
//
// Solidity: function getMachines(uint256 providerId) view returns((bool,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,bool,string,uint256,uint256,uint256,uint256)[])
func (_Provider *ProviderSession) GetMachines(providerId *big.Int) ([]SubnetProviderMachine, error) {
	return _Provider.Contract.GetMachines(&_Provider.CallOpts, providerId)
}

// GetMachines is a free data retrieval call binding the contract method 0x8d49e78d.
//
// Solidity: function getMachines(uint256 providerId) view returns((bool,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,bool,string,uint256,uint256,uint256,uint256)[])
func (_Provider *ProviderCallerSession) GetMachines(providerId *big.Int) ([]SubnetProviderMachine, error) {
	return _Provider.Contract.GetMachines(&_Provider.CallOpts, providerId)
}

// GetMachinesPaginated is a free data retrieval call binding the contract method 0xa1d95967.
//
// Solidity: function getMachinesPaginated(uint256 providerId, uint256 start, uint256 end) view returns((bool,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,bool,string,uint256,uint256,uint256,uint256)[])
func (_Provider *ProviderCaller) GetMachinesPaginated(opts *bind.CallOpts, providerId *big.Int, start *big.Int, end *big.Int) ([]SubnetProviderMachine, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "getMachinesPaginated", providerId, start, end)

	if err != nil {
		return *new([]SubnetProviderMachine), err
	}

	out0 := *abi.ConvertType(out[0], new([]SubnetProviderMachine)).(*[]SubnetProviderMachine)

	return out0, err

}

// GetMachinesPaginated is a free data retrieval call binding the contract method 0xa1d95967.
//
// Solidity: function getMachinesPaginated(uint256 providerId, uint256 start, uint256 end) view returns((bool,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,bool,string,uint256,uint256,uint256,uint256)[])
func (_Provider *ProviderSession) GetMachinesPaginated(providerId *big.Int, start *big.Int, end *big.Int) ([]SubnetProviderMachine, error) {
	return _Provider.Contract.GetMachinesPaginated(&_Provider.CallOpts, providerId, start, end)
}

// GetMachinesPaginated is a free data retrieval call binding the contract method 0xa1d95967.
//
// Solidity: function getMachinesPaginated(uint256 providerId, uint256 start, uint256 end) view returns((bool,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,bool,string,uint256,uint256,uint256,uint256)[])
func (_Provider *ProviderCallerSession) GetMachinesPaginated(providerId *big.Int, start *big.Int, end *big.Int) ([]SubnetProviderMachine, error) {
	return _Provider.Contract.GetMachinesPaginated(&_Provider.CallOpts, providerId, start, end)
}

// GetProvider is a free data retrieval call binding the contract method 0x5c42d079.
//
// Solidity: function getProvider(uint256 providerId) view returns((address,bool,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,string,bool,bool,bool) provider)
func (_Provider *ProviderCaller) GetProvider(opts *bind.CallOpts, providerId *big.Int) (SubnetProviderProvider, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "getProvider", providerId)

	if err != nil {
		return *new(SubnetProviderProvider), err
	}

	out0 := *abi.ConvertType(out[0], new(SubnetProviderProvider)).(*SubnetProviderProvider)

	return out0, err

}

// GetProvider is a free data retrieval call binding the contract method 0x5c42d079.
//
// Solidity: function getProvider(uint256 providerId) view returns((address,bool,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,string,bool,bool,bool) provider)
func (_Provider *ProviderSession) GetProvider(providerId *big.Int) (SubnetProviderProvider, error) {
	return _Provider.Contract.GetProvider(&_Provider.CallOpts, providerId)
}

// GetProvider is a free data retrieval call binding the contract method 0x5c42d079.
//
// Solidity: function getProvider(uint256 providerId) view returns((address,bool,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,string,bool,bool,bool) provider)
func (_Provider *ProviderCallerSession) GetProvider(providerId *big.Int) (SubnetProviderProvider, error) {
	return _Provider.Contract.GetProvider(&_Provider.CallOpts, providerId)
}

// GetProviderOwner is a free data retrieval call binding the contract method 0x1df1ec82.
//
// Solidity: function getProviderOwner(uint256 providerId) view returns(address)
func (_Provider *ProviderCaller) GetProviderOwner(opts *bind.CallOpts, providerId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "getProviderOwner", providerId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetProviderOwner is a free data retrieval call binding the contract method 0x1df1ec82.
//
// Solidity: function getProviderOwner(uint256 providerId) view returns(address)
func (_Provider *ProviderSession) GetProviderOwner(providerId *big.Int) (common.Address, error) {
	return _Provider.Contract.GetProviderOwner(&_Provider.CallOpts, providerId)
}

// GetProviderOwner is a free data retrieval call binding the contract method 0x1df1ec82.
//
// Solidity: function getProviderOwner(uint256 providerId) view returns(address)
func (_Provider *ProviderCallerSession) GetProviderOwner(providerId *big.Int) (common.Address, error) {
	return _Provider.Contract.GetProviderOwner(&_Provider.CallOpts, providerId)
}

// GetProviderReputation is a free data retrieval call binding the contract method 0x0d2bbacb.
//
// Solidity: function getProviderReputation(uint256 providerId) view returns(uint256)
func (_Provider *ProviderCaller) GetProviderReputation(opts *bind.CallOpts, providerId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "getProviderReputation", providerId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetProviderReputation is a free data retrieval call binding the contract method 0x0d2bbacb.
//
// Solidity: function getProviderReputation(uint256 providerId) view returns(uint256)
func (_Provider *ProviderSession) GetProviderReputation(providerId *big.Int) (*big.Int, error) {
	return _Provider.Contract.GetProviderReputation(&_Provider.CallOpts, providerId)
}

// GetProviderReputation is a free data retrieval call binding the contract method 0x0d2bbacb.
//
// Solidity: function getProviderReputation(uint256 providerId) view returns(uint256)
func (_Provider *ProviderCallerSession) GetProviderReputation(providerId *big.Int) (*big.Int, error) {
	return _Provider.Contract.GetProviderReputation(&_Provider.CallOpts, providerId)
}

// GetProviderSlashedAmount is a free data retrieval call binding the contract method 0x972282fb.
//
// Solidity: function getProviderSlashedAmount(uint256 providerId) view returns(uint256)
func (_Provider *ProviderCaller) GetProviderSlashedAmount(opts *bind.CallOpts, providerId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "getProviderSlashedAmount", providerId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetProviderSlashedAmount is a free data retrieval call binding the contract method 0x972282fb.
//
// Solidity: function getProviderSlashedAmount(uint256 providerId) view returns(uint256)
func (_Provider *ProviderSession) GetProviderSlashedAmount(providerId *big.Int) (*big.Int, error) {
	return _Provider.Contract.GetProviderSlashedAmount(&_Provider.CallOpts, providerId)
}

// GetProviderSlashedAmount is a free data retrieval call binding the contract method 0x972282fb.
//
// Solidity: function getProviderSlashedAmount(uint256 providerId) view returns(uint256)
func (_Provider *ProviderCallerSession) GetProviderSlashedAmount(providerId *big.Int) (*big.Int, error) {
	return _Provider.Contract.GetProviderSlashedAmount(&_Provider.CallOpts, providerId)
}

// GpuStakeRate is a free data retrieval call binding the contract method 0xbd1516ae.
//
// Solidity: function gpuStakeRate() view returns(uint256)
func (_Provider *ProviderCaller) GpuStakeRate(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "gpuStakeRate")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GpuStakeRate is a free data retrieval call binding the contract method 0xbd1516ae.
//
// Solidity: function gpuStakeRate() view returns(uint256)
func (_Provider *ProviderSession) GpuStakeRate() (*big.Int, error) {
	return _Provider.Contract.GpuStakeRate(&_Provider.CallOpts)
}

// GpuStakeRate is a free data retrieval call binding the contract method 0xbd1516ae.
//
// Solidity: function gpuStakeRate() view returns(uint256)
func (_Provider *ProviderCallerSession) GpuStakeRate() (*big.Int, error) {
	return _Provider.Contract.GpuStakeRate(&_Provider.CallOpts)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_Provider *ProviderCaller) IsApprovedForAll(opts *bind.CallOpts, owner common.Address, operator common.Address) (bool, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "isApprovedForAll", owner, operator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_Provider *ProviderSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _Provider.Contract.IsApprovedForAll(&_Provider.CallOpts, owner, operator)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_Provider *ProviderCallerSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _Provider.Contract.IsApprovedForAll(&_Provider.CallOpts, owner, operator)
}

// IsMachineActive is a free data retrieval call binding the contract method 0x38b794cc.
//
// Solidity: function isMachineActive(uint256 providerId, uint256 machineId) view returns(bool)
func (_Provider *ProviderCaller) IsMachineActive(opts *bind.CallOpts, providerId *big.Int, machineId *big.Int) (bool, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "isMachineActive", providerId, machineId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsMachineActive is a free data retrieval call binding the contract method 0x38b794cc.
//
// Solidity: function isMachineActive(uint256 providerId, uint256 machineId) view returns(bool)
func (_Provider *ProviderSession) IsMachineActive(providerId *big.Int, machineId *big.Int) (bool, error) {
	return _Provider.Contract.IsMachineActive(&_Provider.CallOpts, providerId, machineId)
}

// IsMachineActive is a free data retrieval call binding the contract method 0x38b794cc.
//
// Solidity: function isMachineActive(uint256 providerId, uint256 machineId) view returns(bool)
func (_Provider *ProviderCallerSession) IsMachineActive(providerId *big.Int, machineId *big.Int) (bool, error) {
	return _Provider.Contract.IsMachineActive(&_Provider.CallOpts, providerId, machineId)
}

// IsProviderOperatorOrOwner is a free data retrieval call binding the contract method 0xd041d0ba.
//
// Solidity: function isProviderOperatorOrOwner(uint256 providerId, address account) view returns(bool)
func (_Provider *ProviderCaller) IsProviderOperatorOrOwner(opts *bind.CallOpts, providerId *big.Int, account common.Address) (bool, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "isProviderOperatorOrOwner", providerId, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsProviderOperatorOrOwner is a free data retrieval call binding the contract method 0xd041d0ba.
//
// Solidity: function isProviderOperatorOrOwner(uint256 providerId, address account) view returns(bool)
func (_Provider *ProviderSession) IsProviderOperatorOrOwner(providerId *big.Int, account common.Address) (bool, error) {
	return _Provider.Contract.IsProviderOperatorOrOwner(&_Provider.CallOpts, providerId, account)
}

// IsProviderOperatorOrOwner is a free data retrieval call binding the contract method 0xd041d0ba.
//
// Solidity: function isProviderOperatorOrOwner(uint256 providerId, address account) view returns(bool)
func (_Provider *ProviderCallerSession) IsProviderOperatorOrOwner(providerId *big.Int, account common.Address) (bool, error) {
	return _Provider.Contract.IsProviderOperatorOrOwner(&_Provider.CallOpts, providerId, account)
}

// IsVerified is a free data retrieval call binding the contract method 0x37b6d96b.
//
// Solidity: function isVerified(uint256 providerId) view returns(bool)
func (_Provider *ProviderCaller) IsVerified(opts *bind.CallOpts, providerId *big.Int) (bool, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "isVerified", providerId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsVerified is a free data retrieval call binding the contract method 0x37b6d96b.
//
// Solidity: function isVerified(uint256 providerId) view returns(bool)
func (_Provider *ProviderSession) IsVerified(providerId *big.Int) (bool, error) {
	return _Provider.Contract.IsVerified(&_Provider.CallOpts, providerId)
}

// IsVerified is a free data retrieval call binding the contract method 0x37b6d96b.
//
// Solidity: function isVerified(uint256 providerId) view returns(bool)
func (_Provider *ProviderCallerSession) IsVerified(providerId *big.Int) (bool, error) {
	return _Provider.Contract.IsVerified(&_Provider.CallOpts, providerId)
}

// LockPeriod is a free data retrieval call binding the contract method 0x3fd8b02f.
//
// Solidity: function lockPeriod() view returns(uint256)
func (_Provider *ProviderCaller) LockPeriod(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "lockPeriod")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LockPeriod is a free data retrieval call binding the contract method 0x3fd8b02f.
//
// Solidity: function lockPeriod() view returns(uint256)
func (_Provider *ProviderSession) LockPeriod() (*big.Int, error) {
	return _Provider.Contract.LockPeriod(&_Provider.CallOpts)
}

// LockPeriod is a free data retrieval call binding the contract method 0x3fd8b02f.
//
// Solidity: function lockPeriod() view returns(uint256)
func (_Provider *ProviderCallerSession) LockPeriod() (*big.Int, error) {
	return _Provider.Contract.LockPeriod(&_Provider.CallOpts)
}

// MemoryStakeRate is a free data retrieval call binding the contract method 0x57356f06.
//
// Solidity: function memoryStakeRate() view returns(uint256)
func (_Provider *ProviderCaller) MemoryStakeRate(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "memoryStakeRate")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MemoryStakeRate is a free data retrieval call binding the contract method 0x57356f06.
//
// Solidity: function memoryStakeRate() view returns(uint256)
func (_Provider *ProviderSession) MemoryStakeRate() (*big.Int, error) {
	return _Provider.Contract.MemoryStakeRate(&_Provider.CallOpts)
}

// MemoryStakeRate is a free data retrieval call binding the contract method 0x57356f06.
//
// Solidity: function memoryStakeRate() view returns(uint256)
func (_Provider *ProviderCallerSession) MemoryStakeRate() (*big.Int, error) {
	return _Provider.Contract.MemoryStakeRate(&_Provider.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Provider *ProviderCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Provider *ProviderSession) Name() (string, error) {
	return _Provider.Contract.Name(&_Provider.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Provider *ProviderCallerSession) Name() (string, error) {
	return _Provider.Contract.Name(&_Provider.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Provider *ProviderCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Provider *ProviderSession) Owner() (common.Address, error) {
	return _Provider.Contract.Owner(&_Provider.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Provider *ProviderCallerSession) Owner() (common.Address, error) {
	return _Provider.Contract.Owner(&_Provider.CallOpts)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_Provider *ProviderCaller) OwnerOf(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "ownerOf", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_Provider *ProviderSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _Provider.Contract.OwnerOf(&_Provider.CallOpts, tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_Provider *ProviderCallerSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _Provider.Contract.OwnerOf(&_Provider.CallOpts, tokenId)
}

// ProviderMachines is a free data retrieval call binding the contract method 0x0e55e90c.
//
// Solidity: function providerMachines(uint256 , uint256 ) view returns(bool active, uint256 machineType, uint256 region, uint256 cpuCores, uint256 gpuCores, uint256 gpuMemory, uint256 memoryMB, uint256 diskGB, uint256 uploadSpeed, uint256 downloadSpeed, uint256 createdAt, uint256 updatedAt, uint256 stakeAmount, uint256 removedAt, uint256 unlockTime, bool withdrawalProcessed, string metadata, uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond)
func (_Provider *ProviderCaller) ProviderMachines(opts *bind.CallOpts, arg0 *big.Int, arg1 *big.Int) (struct {
	Active               bool
	MachineType          *big.Int
	Region               *big.Int
	CpuCores             *big.Int
	GpuCores             *big.Int
	GpuMemory            *big.Int
	MemoryMB             *big.Int
	DiskGB               *big.Int
	UploadSpeed          *big.Int
	DownloadSpeed        *big.Int
	CreatedAt            *big.Int
	UpdatedAt            *big.Int
	StakeAmount          *big.Int
	RemovedAt            *big.Int
	UnlockTime           *big.Int
	WithdrawalProcessed  bool
	Metadata             string
	CpuPricePerSecond    *big.Int
	GpuPricePerSecond    *big.Int
	MemoryPricePerSecond *big.Int
	DiskPricePerSecond   *big.Int
}, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "providerMachines", arg0, arg1)

	outstruct := new(struct {
		Active               bool
		MachineType          *big.Int
		Region               *big.Int
		CpuCores             *big.Int
		GpuCores             *big.Int
		GpuMemory            *big.Int
		MemoryMB             *big.Int
		DiskGB               *big.Int
		UploadSpeed          *big.Int
		DownloadSpeed        *big.Int
		CreatedAt            *big.Int
		UpdatedAt            *big.Int
		StakeAmount          *big.Int
		RemovedAt            *big.Int
		UnlockTime           *big.Int
		WithdrawalProcessed  bool
		Metadata             string
		CpuPricePerSecond    *big.Int
		GpuPricePerSecond    *big.Int
		MemoryPricePerSecond *big.Int
		DiskPricePerSecond   *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Active = *abi.ConvertType(out[0], new(bool)).(*bool)
	outstruct.MachineType = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Region = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.CpuCores = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.GpuCores = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.GpuMemory = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)
	outstruct.MemoryMB = *abi.ConvertType(out[6], new(*big.Int)).(**big.Int)
	outstruct.DiskGB = *abi.ConvertType(out[7], new(*big.Int)).(**big.Int)
	outstruct.UploadSpeed = *abi.ConvertType(out[8], new(*big.Int)).(**big.Int)
	outstruct.DownloadSpeed = *abi.ConvertType(out[9], new(*big.Int)).(**big.Int)
	outstruct.CreatedAt = *abi.ConvertType(out[10], new(*big.Int)).(**big.Int)
	outstruct.UpdatedAt = *abi.ConvertType(out[11], new(*big.Int)).(**big.Int)
	outstruct.StakeAmount = *abi.ConvertType(out[12], new(*big.Int)).(**big.Int)
	outstruct.RemovedAt = *abi.ConvertType(out[13], new(*big.Int)).(**big.Int)
	outstruct.UnlockTime = *abi.ConvertType(out[14], new(*big.Int)).(**big.Int)
	outstruct.WithdrawalProcessed = *abi.ConvertType(out[15], new(bool)).(*bool)
	outstruct.Metadata = *abi.ConvertType(out[16], new(string)).(*string)
	outstruct.CpuPricePerSecond = *abi.ConvertType(out[17], new(*big.Int)).(**big.Int)
	outstruct.GpuPricePerSecond = *abi.ConvertType(out[18], new(*big.Int)).(**big.Int)
	outstruct.MemoryPricePerSecond = *abi.ConvertType(out[19], new(*big.Int)).(**big.Int)
	outstruct.DiskPricePerSecond = *abi.ConvertType(out[20], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// ProviderMachines is a free data retrieval call binding the contract method 0x0e55e90c.
//
// Solidity: function providerMachines(uint256 , uint256 ) view returns(bool active, uint256 machineType, uint256 region, uint256 cpuCores, uint256 gpuCores, uint256 gpuMemory, uint256 memoryMB, uint256 diskGB, uint256 uploadSpeed, uint256 downloadSpeed, uint256 createdAt, uint256 updatedAt, uint256 stakeAmount, uint256 removedAt, uint256 unlockTime, bool withdrawalProcessed, string metadata, uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond)
func (_Provider *ProviderSession) ProviderMachines(arg0 *big.Int, arg1 *big.Int) (struct {
	Active               bool
	MachineType          *big.Int
	Region               *big.Int
	CpuCores             *big.Int
	GpuCores             *big.Int
	GpuMemory            *big.Int
	MemoryMB             *big.Int
	DiskGB               *big.Int
	UploadSpeed          *big.Int
	DownloadSpeed        *big.Int
	CreatedAt            *big.Int
	UpdatedAt            *big.Int
	StakeAmount          *big.Int
	RemovedAt            *big.Int
	UnlockTime           *big.Int
	WithdrawalProcessed  bool
	Metadata             string
	CpuPricePerSecond    *big.Int
	GpuPricePerSecond    *big.Int
	MemoryPricePerSecond *big.Int
	DiskPricePerSecond   *big.Int
}, error) {
	return _Provider.Contract.ProviderMachines(&_Provider.CallOpts, arg0, arg1)
}

// ProviderMachines is a free data retrieval call binding the contract method 0x0e55e90c.
//
// Solidity: function providerMachines(uint256 , uint256 ) view returns(bool active, uint256 machineType, uint256 region, uint256 cpuCores, uint256 gpuCores, uint256 gpuMemory, uint256 memoryMB, uint256 diskGB, uint256 uploadSpeed, uint256 downloadSpeed, uint256 createdAt, uint256 updatedAt, uint256 stakeAmount, uint256 removedAt, uint256 unlockTime, bool withdrawalProcessed, string metadata, uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond)
func (_Provider *ProviderCallerSession) ProviderMachines(arg0 *big.Int, arg1 *big.Int) (struct {
	Active               bool
	MachineType          *big.Int
	Region               *big.Int
	CpuCores             *big.Int
	GpuCores             *big.Int
	GpuMemory            *big.Int
	MemoryMB             *big.Int
	DiskGB               *big.Int
	UploadSpeed          *big.Int
	DownloadSpeed        *big.Int
	CreatedAt            *big.Int
	UpdatedAt            *big.Int
	StakeAmount          *big.Int
	RemovedAt            *big.Int
	UnlockTime           *big.Int
	WithdrawalProcessed  bool
	Metadata             string
	CpuPricePerSecond    *big.Int
	GpuPricePerSecond    *big.Int
	MemoryPricePerSecond *big.Int
	DiskPricePerSecond   *big.Int
}, error) {
	return _Provider.Contract.ProviderMachines(&_Provider.CallOpts, arg0, arg1)
}

// Providers is a free data retrieval call binding the contract method 0x50f3fc81.
//
// Solidity: function providers(uint256 ) view returns(address operator, bool registered, uint256 reputation, uint256 machineCount, uint256 createdAt, uint256 updatedAt, uint256 totalStaked, uint256 pendingWithdrawals, uint256 slashedAmount, uint256 tokenId, string metadata, bool isSlashed, bool isActive, bool verified)
func (_Provider *ProviderCaller) Providers(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Operator           common.Address
	Registered         bool
	Reputation         *big.Int
	MachineCount       *big.Int
	CreatedAt          *big.Int
	UpdatedAt          *big.Int
	TotalStaked        *big.Int
	PendingWithdrawals *big.Int
	SlashedAmount      *big.Int
	TokenId            *big.Int
	Metadata           string
	IsSlashed          bool
	IsActive           bool
	Verified           bool
}, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "providers", arg0)

	outstruct := new(struct {
		Operator           common.Address
		Registered         bool
		Reputation         *big.Int
		MachineCount       *big.Int
		CreatedAt          *big.Int
		UpdatedAt          *big.Int
		TotalStaked        *big.Int
		PendingWithdrawals *big.Int
		SlashedAmount      *big.Int
		TokenId            *big.Int
		Metadata           string
		IsSlashed          bool
		IsActive           bool
		Verified           bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Operator = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Registered = *abi.ConvertType(out[1], new(bool)).(*bool)
	outstruct.Reputation = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.MachineCount = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.CreatedAt = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.UpdatedAt = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)
	outstruct.TotalStaked = *abi.ConvertType(out[6], new(*big.Int)).(**big.Int)
	outstruct.PendingWithdrawals = *abi.ConvertType(out[7], new(*big.Int)).(**big.Int)
	outstruct.SlashedAmount = *abi.ConvertType(out[8], new(*big.Int)).(**big.Int)
	outstruct.TokenId = *abi.ConvertType(out[9], new(*big.Int)).(**big.Int)
	outstruct.Metadata = *abi.ConvertType(out[10], new(string)).(*string)
	outstruct.IsSlashed = *abi.ConvertType(out[11], new(bool)).(*bool)
	outstruct.IsActive = *abi.ConvertType(out[12], new(bool)).(*bool)
	outstruct.Verified = *abi.ConvertType(out[13], new(bool)).(*bool)

	return *outstruct, err

}

// Providers is a free data retrieval call binding the contract method 0x50f3fc81.
//
// Solidity: function providers(uint256 ) view returns(address operator, bool registered, uint256 reputation, uint256 machineCount, uint256 createdAt, uint256 updatedAt, uint256 totalStaked, uint256 pendingWithdrawals, uint256 slashedAmount, uint256 tokenId, string metadata, bool isSlashed, bool isActive, bool verified)
func (_Provider *ProviderSession) Providers(arg0 *big.Int) (struct {
	Operator           common.Address
	Registered         bool
	Reputation         *big.Int
	MachineCount       *big.Int
	CreatedAt          *big.Int
	UpdatedAt          *big.Int
	TotalStaked        *big.Int
	PendingWithdrawals *big.Int
	SlashedAmount      *big.Int
	TokenId            *big.Int
	Metadata           string
	IsSlashed          bool
	IsActive           bool
	Verified           bool
}, error) {
	return _Provider.Contract.Providers(&_Provider.CallOpts, arg0)
}

// Providers is a free data retrieval call binding the contract method 0x50f3fc81.
//
// Solidity: function providers(uint256 ) view returns(address operator, bool registered, uint256 reputation, uint256 machineCount, uint256 createdAt, uint256 updatedAt, uint256 totalStaked, uint256 pendingWithdrawals, uint256 slashedAmount, uint256 tokenId, string metadata, bool isSlashed, bool isActive, bool verified)
func (_Provider *ProviderCallerSession) Providers(arg0 *big.Int) (struct {
	Operator           common.Address
	Registered         bool
	Reputation         *big.Int
	MachineCount       *big.Int
	CreatedAt          *big.Int
	UpdatedAt          *big.Int
	TotalStaked        *big.Int
	PendingWithdrawals *big.Int
	SlashedAmount      *big.Int
	TokenId            *big.Int
	Metadata           string
	IsSlashed          bool
	IsActive           bool
	Verified           bool
}, error) {
	return _Provider.Contract.Providers(&_Provider.CallOpts, arg0)
}

// StakingToken is a free data retrieval call binding the contract method 0x72f702f3.
//
// Solidity: function stakingToken() view returns(address)
func (_Provider *ProviderCaller) StakingToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "stakingToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// StakingToken is a free data retrieval call binding the contract method 0x72f702f3.
//
// Solidity: function stakingToken() view returns(address)
func (_Provider *ProviderSession) StakingToken() (common.Address, error) {
	return _Provider.Contract.StakingToken(&_Provider.CallOpts)
}

// StakingToken is a free data retrieval call binding the contract method 0x72f702f3.
//
// Solidity: function stakingToken() view returns(address)
func (_Provider *ProviderCallerSession) StakingToken() (common.Address, error) {
	return _Provider.Contract.StakingToken(&_Provider.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Provider *ProviderCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Provider *ProviderSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Provider.Contract.SupportsInterface(&_Provider.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Provider *ProviderCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Provider.Contract.SupportsInterface(&_Provider.CallOpts, interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Provider *ProviderCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Provider *ProviderSession) Symbol() (string, error) {
	return _Provider.Contract.Symbol(&_Provider.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Provider *ProviderCallerSession) Symbol() (string, error) {
	return _Provider.Contract.Symbol(&_Provider.CallOpts)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_Provider *ProviderCaller) TokenURI(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "tokenURI", tokenId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_Provider *ProviderSession) TokenURI(tokenId *big.Int) (string, error) {
	return _Provider.Contract.TokenURI(&_Provider.CallOpts, tokenId)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_Provider *ProviderCallerSession) TokenURI(tokenId *big.Int) (string, error) {
	return _Provider.Contract.TokenURI(&_Provider.CallOpts, tokenId)
}

// TotalSlashed is a free data retrieval call binding the contract method 0xa201bbdd.
//
// Solidity: function totalSlashed() view returns(uint256)
func (_Provider *ProviderCaller) TotalSlashed(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "totalSlashed")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSlashed is a free data retrieval call binding the contract method 0xa201bbdd.
//
// Solidity: function totalSlashed() view returns(uint256)
func (_Provider *ProviderSession) TotalSlashed() (*big.Int, error) {
	return _Provider.Contract.TotalSlashed(&_Provider.CallOpts)
}

// TotalSlashed is a free data retrieval call binding the contract method 0xa201bbdd.
//
// Solidity: function totalSlashed() view returns(uint256)
func (_Provider *ProviderCallerSession) TotalSlashed() (*big.Int, error) {
	return _Provider.Contract.TotalSlashed(&_Provider.CallOpts)
}

// UploadSpeedStakeRate is a free data retrieval call binding the contract method 0x304edf5a.
//
// Solidity: function uploadSpeedStakeRate() view returns(uint256)
func (_Provider *ProviderCaller) UploadSpeedStakeRate(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "uploadSpeedStakeRate")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UploadSpeedStakeRate is a free data retrieval call binding the contract method 0x304edf5a.
//
// Solidity: function uploadSpeedStakeRate() view returns(uint256)
func (_Provider *ProviderSession) UploadSpeedStakeRate() (*big.Int, error) {
	return _Provider.Contract.UploadSpeedStakeRate(&_Provider.CallOpts)
}

// UploadSpeedStakeRate is a free data retrieval call binding the contract method 0x304edf5a.
//
// Solidity: function uploadSpeedStakeRate() view returns(uint256)
func (_Provider *ProviderCallerSession) UploadSpeedStakeRate() (*big.Int, error) {
	return _Provider.Contract.UploadSpeedStakeRate(&_Provider.CallOpts)
}

// ValidateMachineRequirements is a free data retrieval call binding the contract method 0x3ad9b58a.
//
// Solidity: function validateMachineRequirements(uint256 machineType, uint256 providerId, uint256 machineId, uint256 minCpuCores, uint256 minMemoryMB, uint256 minDiskGB, uint256 minGpuCores, uint256 minUploadSpeed, uint256 minDownloadSpeed) view returns(bool)
func (_Provider *ProviderCaller) ValidateMachineRequirements(opts *bind.CallOpts, machineType *big.Int, providerId *big.Int, machineId *big.Int, minCpuCores *big.Int, minMemoryMB *big.Int, minDiskGB *big.Int, minGpuCores *big.Int, minUploadSpeed *big.Int, minDownloadSpeed *big.Int) (bool, error) {
	var out []interface{}
	err := _Provider.contract.Call(opts, &out, "validateMachineRequirements", machineType, providerId, machineId, minCpuCores, minMemoryMB, minDiskGB, minGpuCores, minUploadSpeed, minDownloadSpeed)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ValidateMachineRequirements is a free data retrieval call binding the contract method 0x3ad9b58a.
//
// Solidity: function validateMachineRequirements(uint256 machineType, uint256 providerId, uint256 machineId, uint256 minCpuCores, uint256 minMemoryMB, uint256 minDiskGB, uint256 minGpuCores, uint256 minUploadSpeed, uint256 minDownloadSpeed) view returns(bool)
func (_Provider *ProviderSession) ValidateMachineRequirements(machineType *big.Int, providerId *big.Int, machineId *big.Int, minCpuCores *big.Int, minMemoryMB *big.Int, minDiskGB *big.Int, minGpuCores *big.Int, minUploadSpeed *big.Int, minDownloadSpeed *big.Int) (bool, error) {
	return _Provider.Contract.ValidateMachineRequirements(&_Provider.CallOpts, machineType, providerId, machineId, minCpuCores, minMemoryMB, minDiskGB, minGpuCores, minUploadSpeed, minDownloadSpeed)
}

// ValidateMachineRequirements is a free data retrieval call binding the contract method 0x3ad9b58a.
//
// Solidity: function validateMachineRequirements(uint256 machineType, uint256 providerId, uint256 machineId, uint256 minCpuCores, uint256 minMemoryMB, uint256 minDiskGB, uint256 minGpuCores, uint256 minUploadSpeed, uint256 minDownloadSpeed) view returns(bool)
func (_Provider *ProviderCallerSession) ValidateMachineRequirements(machineType *big.Int, providerId *big.Int, machineId *big.Int, minCpuCores *big.Int, minMemoryMB *big.Int, minDiskGB *big.Int, minGpuCores *big.Int, minUploadSpeed *big.Int, minDownloadSpeed *big.Int) (bool, error) {
	return _Provider.Contract.ValidateMachineRequirements(&_Provider.CallOpts, machineType, providerId, machineId, minCpuCores, minMemoryMB, minDiskGB, minGpuCores, minUploadSpeed, minDownloadSpeed)
}

// AddMachine is a paid mutator transaction binding the contract method 0xcf8a1ae7.
//
// Solidity: function addMachine(uint256 providerId, uint256 machineType, uint256 region, uint256 cpuCores, uint256 gpuCores, uint256 gpuMemory, uint256 memoryMB, uint256 diskGB, uint256 uploadSpeed, uint256 downloadSpeed, string metadata, uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond) returns(uint256)
func (_Provider *ProviderTransactor) AddMachine(opts *bind.TransactOpts, providerId *big.Int, machineType *big.Int, region *big.Int, cpuCores *big.Int, gpuCores *big.Int, gpuMemory *big.Int, memoryMB *big.Int, diskGB *big.Int, uploadSpeed *big.Int, downloadSpeed *big.Int, metadata string, cpuPricePerSecond *big.Int, gpuPricePerSecond *big.Int, memoryPricePerSecond *big.Int, diskPricePerSecond *big.Int) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "addMachine", providerId, machineType, region, cpuCores, gpuCores, gpuMemory, memoryMB, diskGB, uploadSpeed, downloadSpeed, metadata, cpuPricePerSecond, gpuPricePerSecond, memoryPricePerSecond, diskPricePerSecond)
}

// AddMachine is a paid mutator transaction binding the contract method 0xcf8a1ae7.
//
// Solidity: function addMachine(uint256 providerId, uint256 machineType, uint256 region, uint256 cpuCores, uint256 gpuCores, uint256 gpuMemory, uint256 memoryMB, uint256 diskGB, uint256 uploadSpeed, uint256 downloadSpeed, string metadata, uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond) returns(uint256)
func (_Provider *ProviderSession) AddMachine(providerId *big.Int, machineType *big.Int, region *big.Int, cpuCores *big.Int, gpuCores *big.Int, gpuMemory *big.Int, memoryMB *big.Int, diskGB *big.Int, uploadSpeed *big.Int, downloadSpeed *big.Int, metadata string, cpuPricePerSecond *big.Int, gpuPricePerSecond *big.Int, memoryPricePerSecond *big.Int, diskPricePerSecond *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.AddMachine(&_Provider.TransactOpts, providerId, machineType, region, cpuCores, gpuCores, gpuMemory, memoryMB, diskGB, uploadSpeed, downloadSpeed, metadata, cpuPricePerSecond, gpuPricePerSecond, memoryPricePerSecond, diskPricePerSecond)
}

// AddMachine is a paid mutator transaction binding the contract method 0xcf8a1ae7.
//
// Solidity: function addMachine(uint256 providerId, uint256 machineType, uint256 region, uint256 cpuCores, uint256 gpuCores, uint256 gpuMemory, uint256 memoryMB, uint256 diskGB, uint256 uploadSpeed, uint256 downloadSpeed, string metadata, uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond) returns(uint256)
func (_Provider *ProviderTransactorSession) AddMachine(providerId *big.Int, machineType *big.Int, region *big.Int, cpuCores *big.Int, gpuCores *big.Int, gpuMemory *big.Int, memoryMB *big.Int, diskGB *big.Int, uploadSpeed *big.Int, downloadSpeed *big.Int, metadata string, cpuPricePerSecond *big.Int, gpuPricePerSecond *big.Int, memoryPricePerSecond *big.Int, diskPricePerSecond *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.AddMachine(&_Provider.TransactOpts, providerId, machineType, region, cpuCores, gpuCores, gpuMemory, memoryMB, diskGB, uploadSpeed, downloadSpeed, metadata, cpuPricePerSecond, gpuPricePerSecond, memoryPricePerSecond, diskPricePerSecond)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_Provider *ProviderTransactor) Approve(opts *bind.TransactOpts, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "approve", to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_Provider *ProviderSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.Approve(&_Provider.TransactOpts, to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_Provider *ProviderTransactorSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.Approve(&_Provider.TransactOpts, to, tokenId)
}

// ClaimWithdrawal is a paid mutator transaction binding the contract method 0xf21340e4.
//
// Solidity: function claimWithdrawal(uint256 providerId, uint256 machineId) returns()
func (_Provider *ProviderTransactor) ClaimWithdrawal(opts *bind.TransactOpts, providerId *big.Int, machineId *big.Int) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "claimWithdrawal", providerId, machineId)
}

// ClaimWithdrawal is a paid mutator transaction binding the contract method 0xf21340e4.
//
// Solidity: function claimWithdrawal(uint256 providerId, uint256 machineId) returns()
func (_Provider *ProviderSession) ClaimWithdrawal(providerId *big.Int, machineId *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.ClaimWithdrawal(&_Provider.TransactOpts, providerId, machineId)
}

// ClaimWithdrawal is a paid mutator transaction binding the contract method 0xf21340e4.
//
// Solidity: function claimWithdrawal(uint256 providerId, uint256 machineId) returns()
func (_Provider *ProviderTransactorSession) ClaimWithdrawal(providerId *big.Int, machineId *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.ClaimWithdrawal(&_Provider.TransactOpts, providerId, machineId)
}

// Initialize is a paid mutator transaction binding the contract method 0x2016a0d2.
//
// Solidity: function initialize(address owner, address _stakingToken, string nftName, string nftSymbol) returns()
func (_Provider *ProviderTransactor) Initialize(opts *bind.TransactOpts, owner common.Address, _stakingToken common.Address, nftName string, nftSymbol string) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "initialize", owner, _stakingToken, nftName, nftSymbol)
}

// Initialize is a paid mutator transaction binding the contract method 0x2016a0d2.
//
// Solidity: function initialize(address owner, address _stakingToken, string nftName, string nftSymbol) returns()
func (_Provider *ProviderSession) Initialize(owner common.Address, _stakingToken common.Address, nftName string, nftSymbol string) (*types.Transaction, error) {
	return _Provider.Contract.Initialize(&_Provider.TransactOpts, owner, _stakingToken, nftName, nftSymbol)
}

// Initialize is a paid mutator transaction binding the contract method 0x2016a0d2.
//
// Solidity: function initialize(address owner, address _stakingToken, string nftName, string nftSymbol) returns()
func (_Provider *ProviderTransactorSession) Initialize(owner common.Address, _stakingToken common.Address, nftName string, nftSymbol string) (*types.Transaction, error) {
	return _Provider.Contract.Initialize(&_Provider.TransactOpts, owner, _stakingToken, nftName, nftSymbol)
}

// RegisterProvider is a paid mutator transaction binding the contract method 0x5b20cc5b.
//
// Solidity: function registerProvider(address operator, string metadata) returns(uint256)
func (_Provider *ProviderTransactor) RegisterProvider(opts *bind.TransactOpts, operator common.Address, metadata string) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "registerProvider", operator, metadata)
}

// RegisterProvider is a paid mutator transaction binding the contract method 0x5b20cc5b.
//
// Solidity: function registerProvider(address operator, string metadata) returns(uint256)
func (_Provider *ProviderSession) RegisterProvider(operator common.Address, metadata string) (*types.Transaction, error) {
	return _Provider.Contract.RegisterProvider(&_Provider.TransactOpts, operator, metadata)
}

// RegisterProvider is a paid mutator transaction binding the contract method 0x5b20cc5b.
//
// Solidity: function registerProvider(address operator, string metadata) returns(uint256)
func (_Provider *ProviderTransactorSession) RegisterProvider(operator common.Address, metadata string) (*types.Transaction, error) {
	return _Provider.Contract.RegisterProvider(&_Provider.TransactOpts, operator, metadata)
}

// RegisterProviderWithMachine is a paid mutator transaction binding the contract method 0x9cee4f69.
//
// Solidity: function registerProviderWithMachine(address operator, string providerMetadata, uint256 machineType, uint256 region, uint256 cpuCores, uint256 gpuCores, uint256 gpuMemory, uint256 memoryMB, uint256 diskGB, uint256 uploadSpeed, uint256 downloadSpeed, string machineMetadata, uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond) returns(uint256 providerId, uint256 machineId)
func (_Provider *ProviderTransactor) RegisterProviderWithMachine(opts *bind.TransactOpts, operator common.Address, providerMetadata string, machineType *big.Int, region *big.Int, cpuCores *big.Int, gpuCores *big.Int, gpuMemory *big.Int, memoryMB *big.Int, diskGB *big.Int, uploadSpeed *big.Int, downloadSpeed *big.Int, machineMetadata string, cpuPricePerSecond *big.Int, gpuPricePerSecond *big.Int, memoryPricePerSecond *big.Int, diskPricePerSecond *big.Int) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "registerProviderWithMachine", operator, providerMetadata, machineType, region, cpuCores, gpuCores, gpuMemory, memoryMB, diskGB, uploadSpeed, downloadSpeed, machineMetadata, cpuPricePerSecond, gpuPricePerSecond, memoryPricePerSecond, diskPricePerSecond)
}

// RegisterProviderWithMachine is a paid mutator transaction binding the contract method 0x9cee4f69.
//
// Solidity: function registerProviderWithMachine(address operator, string providerMetadata, uint256 machineType, uint256 region, uint256 cpuCores, uint256 gpuCores, uint256 gpuMemory, uint256 memoryMB, uint256 diskGB, uint256 uploadSpeed, uint256 downloadSpeed, string machineMetadata, uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond) returns(uint256 providerId, uint256 machineId)
func (_Provider *ProviderSession) RegisterProviderWithMachine(operator common.Address, providerMetadata string, machineType *big.Int, region *big.Int, cpuCores *big.Int, gpuCores *big.Int, gpuMemory *big.Int, memoryMB *big.Int, diskGB *big.Int, uploadSpeed *big.Int, downloadSpeed *big.Int, machineMetadata string, cpuPricePerSecond *big.Int, gpuPricePerSecond *big.Int, memoryPricePerSecond *big.Int, diskPricePerSecond *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.RegisterProviderWithMachine(&_Provider.TransactOpts, operator, providerMetadata, machineType, region, cpuCores, gpuCores, gpuMemory, memoryMB, diskGB, uploadSpeed, downloadSpeed, machineMetadata, cpuPricePerSecond, gpuPricePerSecond, memoryPricePerSecond, diskPricePerSecond)
}

// RegisterProviderWithMachine is a paid mutator transaction binding the contract method 0x9cee4f69.
//
// Solidity: function registerProviderWithMachine(address operator, string providerMetadata, uint256 machineType, uint256 region, uint256 cpuCores, uint256 gpuCores, uint256 gpuMemory, uint256 memoryMB, uint256 diskGB, uint256 uploadSpeed, uint256 downloadSpeed, string machineMetadata, uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond) returns(uint256 providerId, uint256 machineId)
func (_Provider *ProviderTransactorSession) RegisterProviderWithMachine(operator common.Address, providerMetadata string, machineType *big.Int, region *big.Int, cpuCores *big.Int, gpuCores *big.Int, gpuMemory *big.Int, memoryMB *big.Int, diskGB *big.Int, uploadSpeed *big.Int, downloadSpeed *big.Int, machineMetadata string, cpuPricePerSecond *big.Int, gpuPricePerSecond *big.Int, memoryPricePerSecond *big.Int, diskPricePerSecond *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.RegisterProviderWithMachine(&_Provider.TransactOpts, operator, providerMetadata, machineType, region, cpuCores, gpuCores, gpuMemory, memoryMB, diskGB, uploadSpeed, downloadSpeed, machineMetadata, cpuPricePerSecond, gpuPricePerSecond, memoryPricePerSecond, diskPricePerSecond)
}

// RemoveMachine is a paid mutator transaction binding the contract method 0x88eb26ee.
//
// Solidity: function removeMachine(uint256 providerId, uint256 machineId) returns()
func (_Provider *ProviderTransactor) RemoveMachine(opts *bind.TransactOpts, providerId *big.Int, machineId *big.Int) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "removeMachine", providerId, machineId)
}

// RemoveMachine is a paid mutator transaction binding the contract method 0x88eb26ee.
//
// Solidity: function removeMachine(uint256 providerId, uint256 machineId) returns()
func (_Provider *ProviderSession) RemoveMachine(providerId *big.Int, machineId *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.RemoveMachine(&_Provider.TransactOpts, providerId, machineId)
}

// RemoveMachine is a paid mutator transaction binding the contract method 0x88eb26ee.
//
// Solidity: function removeMachine(uint256 providerId, uint256 machineId) returns()
func (_Provider *ProviderTransactorSession) RemoveMachine(providerId *big.Int, machineId *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.RemoveMachine(&_Provider.TransactOpts, providerId, machineId)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Provider *ProviderTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Provider *ProviderSession) RenounceOwnership() (*types.Transaction, error) {
	return _Provider.Contract.RenounceOwnership(&_Provider.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Provider *ProviderTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Provider.Contract.RenounceOwnership(&_Provider.TransactOpts)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_Provider *ProviderTransactor) SafeTransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "safeTransferFrom", from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_Provider *ProviderSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.SafeTransferFrom(&_Provider.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_Provider *ProviderTransactorSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.SafeTransferFrom(&_Provider.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_Provider *ProviderTransactor) SafeTransferFrom0(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "safeTransferFrom0", from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_Provider *ProviderSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _Provider.Contract.SafeTransferFrom0(&_Provider.TransactOpts, from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_Provider *ProviderTransactorSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _Provider.Contract.SafeTransferFrom0(&_Provider.TransactOpts, from, to, tokenId, data)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_Provider *ProviderTransactor) SetApprovalForAll(opts *bind.TransactOpts, operator common.Address, approved bool) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "setApprovalForAll", operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_Provider *ProviderSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _Provider.Contract.SetApprovalForAll(&_Provider.TransactOpts, operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_Provider *ProviderTransactorSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _Provider.Contract.SetApprovalForAll(&_Provider.TransactOpts, operator, approved)
}

// SetLockPeriod is a paid mutator transaction binding the contract method 0x779972da.
//
// Solidity: function setLockPeriod(uint256 newLockPeriod) returns()
func (_Provider *ProviderTransactor) SetLockPeriod(opts *bind.TransactOpts, newLockPeriod *big.Int) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "setLockPeriod", newLockPeriod)
}

// SetLockPeriod is a paid mutator transaction binding the contract method 0x779972da.
//
// Solidity: function setLockPeriod(uint256 newLockPeriod) returns()
func (_Provider *ProviderSession) SetLockPeriod(newLockPeriod *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.SetLockPeriod(&_Provider.TransactOpts, newLockPeriod)
}

// SetLockPeriod is a paid mutator transaction binding the contract method 0x779972da.
//
// Solidity: function setLockPeriod(uint256 newLockPeriod) returns()
func (_Provider *ProviderTransactorSession) SetLockPeriod(newLockPeriod *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.SetLockPeriod(&_Provider.TransactOpts, newLockPeriod)
}

// SetMachineResourcePrice is a paid mutator transaction binding the contract method 0xe845a981.
//
// Solidity: function setMachineResourcePrice(uint256 providerId, uint256 machineId, uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond) returns()
func (_Provider *ProviderTransactor) SetMachineResourcePrice(opts *bind.TransactOpts, providerId *big.Int, machineId *big.Int, cpuPricePerSecond *big.Int, gpuPricePerSecond *big.Int, memoryPricePerSecond *big.Int, diskPricePerSecond *big.Int) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "setMachineResourcePrice", providerId, machineId, cpuPricePerSecond, gpuPricePerSecond, memoryPricePerSecond, diskPricePerSecond)
}

// SetMachineResourcePrice is a paid mutator transaction binding the contract method 0xe845a981.
//
// Solidity: function setMachineResourcePrice(uint256 providerId, uint256 machineId, uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond) returns()
func (_Provider *ProviderSession) SetMachineResourcePrice(providerId *big.Int, machineId *big.Int, cpuPricePerSecond *big.Int, gpuPricePerSecond *big.Int, memoryPricePerSecond *big.Int, diskPricePerSecond *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.SetMachineResourcePrice(&_Provider.TransactOpts, providerId, machineId, cpuPricePerSecond, gpuPricePerSecond, memoryPricePerSecond, diskPricePerSecond)
}

// SetMachineResourcePrice is a paid mutator transaction binding the contract method 0xe845a981.
//
// Solidity: function setMachineResourcePrice(uint256 providerId, uint256 machineId, uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond) returns()
func (_Provider *ProviderTransactorSession) SetMachineResourcePrice(providerId *big.Int, machineId *big.Int, cpuPricePerSecond *big.Int, gpuPricePerSecond *big.Int, memoryPricePerSecond *big.Int, diskPricePerSecond *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.SetMachineResourcePrice(&_Provider.TransactOpts, providerId, machineId, cpuPricePerSecond, gpuPricePerSecond, memoryPricePerSecond, diskPricePerSecond)
}

// SetProviderOperator is a paid mutator transaction binding the contract method 0x7496be68.
//
// Solidity: function setProviderOperator(uint256 providerId, address newOperator) returns()
func (_Provider *ProviderTransactor) SetProviderOperator(opts *bind.TransactOpts, providerId *big.Int, newOperator common.Address) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "setProviderOperator", providerId, newOperator)
}

// SetProviderOperator is a paid mutator transaction binding the contract method 0x7496be68.
//
// Solidity: function setProviderOperator(uint256 providerId, address newOperator) returns()
func (_Provider *ProviderSession) SetProviderOperator(providerId *big.Int, newOperator common.Address) (*types.Transaction, error) {
	return _Provider.Contract.SetProviderOperator(&_Provider.TransactOpts, providerId, newOperator)
}

// SetProviderOperator is a paid mutator transaction binding the contract method 0x7496be68.
//
// Solidity: function setProviderOperator(uint256 providerId, address newOperator) returns()
func (_Provider *ProviderTransactorSession) SetProviderOperator(providerId *big.Int, newOperator common.Address) (*types.Transaction, error) {
	return _Provider.Contract.SetProviderOperator(&_Provider.TransactOpts, providerId, newOperator)
}

// SetProviderReputation is a paid mutator transaction binding the contract method 0x6d77c256.
//
// Solidity: function setProviderReputation(uint256 providerId, uint256 newReputation) returns()
func (_Provider *ProviderTransactor) SetProviderReputation(opts *bind.TransactOpts, providerId *big.Int, newReputation *big.Int) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "setProviderReputation", providerId, newReputation)
}

// SetProviderReputation is a paid mutator transaction binding the contract method 0x6d77c256.
//
// Solidity: function setProviderReputation(uint256 providerId, uint256 newReputation) returns()
func (_Provider *ProviderSession) SetProviderReputation(providerId *big.Int, newReputation *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.SetProviderReputation(&_Provider.TransactOpts, providerId, newReputation)
}

// SetProviderReputation is a paid mutator transaction binding the contract method 0x6d77c256.
//
// Solidity: function setProviderReputation(uint256 providerId, uint256 newReputation) returns()
func (_Provider *ProviderTransactorSession) SetProviderReputation(providerId *big.Int, newReputation *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.SetProviderReputation(&_Provider.TransactOpts, providerId, newReputation)
}

// SetProviderVerified is a paid mutator transaction binding the contract method 0x3123d1ff.
//
// Solidity: function setProviderVerified(uint256 providerId, bool verified_) returns()
func (_Provider *ProviderTransactor) SetProviderVerified(opts *bind.TransactOpts, providerId *big.Int, verified_ bool) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "setProviderVerified", providerId, verified_)
}

// SetProviderVerified is a paid mutator transaction binding the contract method 0x3123d1ff.
//
// Solidity: function setProviderVerified(uint256 providerId, bool verified_) returns()
func (_Provider *ProviderSession) SetProviderVerified(providerId *big.Int, verified_ bool) (*types.Transaction, error) {
	return _Provider.Contract.SetProviderVerified(&_Provider.TransactOpts, providerId, verified_)
}

// SetProviderVerified is a paid mutator transaction binding the contract method 0x3123d1ff.
//
// Solidity: function setProviderVerified(uint256 providerId, bool verified_) returns()
func (_Provider *ProviderTransactorSession) SetProviderVerified(providerId *big.Int, verified_ bool) (*types.Transaction, error) {
	return _Provider.Contract.SetProviderVerified(&_Provider.TransactOpts, providerId, verified_)
}

// SetStakeParameters is a paid mutator transaction binding the contract method 0x03f612c4.
//
// Solidity: function setStakeParameters(uint256 newBaseStakeAmount, uint256 newCpuStakeRate, uint256 newGpuStakeRate, uint256 newMemoryStakeRate, uint256 newDiskStakeRate, uint256 newUploadSpeedStakeRate, uint256 newDownloadSpeedStakeRate) returns()
func (_Provider *ProviderTransactor) SetStakeParameters(opts *bind.TransactOpts, newBaseStakeAmount *big.Int, newCpuStakeRate *big.Int, newGpuStakeRate *big.Int, newMemoryStakeRate *big.Int, newDiskStakeRate *big.Int, newUploadSpeedStakeRate *big.Int, newDownloadSpeedStakeRate *big.Int) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "setStakeParameters", newBaseStakeAmount, newCpuStakeRate, newGpuStakeRate, newMemoryStakeRate, newDiskStakeRate, newUploadSpeedStakeRate, newDownloadSpeedStakeRate)
}

// SetStakeParameters is a paid mutator transaction binding the contract method 0x03f612c4.
//
// Solidity: function setStakeParameters(uint256 newBaseStakeAmount, uint256 newCpuStakeRate, uint256 newGpuStakeRate, uint256 newMemoryStakeRate, uint256 newDiskStakeRate, uint256 newUploadSpeedStakeRate, uint256 newDownloadSpeedStakeRate) returns()
func (_Provider *ProviderSession) SetStakeParameters(newBaseStakeAmount *big.Int, newCpuStakeRate *big.Int, newGpuStakeRate *big.Int, newMemoryStakeRate *big.Int, newDiskStakeRate *big.Int, newUploadSpeedStakeRate *big.Int, newDownloadSpeedStakeRate *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.SetStakeParameters(&_Provider.TransactOpts, newBaseStakeAmount, newCpuStakeRate, newGpuStakeRate, newMemoryStakeRate, newDiskStakeRate, newUploadSpeedStakeRate, newDownloadSpeedStakeRate)
}

// SetStakeParameters is a paid mutator transaction binding the contract method 0x03f612c4.
//
// Solidity: function setStakeParameters(uint256 newBaseStakeAmount, uint256 newCpuStakeRate, uint256 newGpuStakeRate, uint256 newMemoryStakeRate, uint256 newDiskStakeRate, uint256 newUploadSpeedStakeRate, uint256 newDownloadSpeedStakeRate) returns()
func (_Provider *ProviderTransactorSession) SetStakeParameters(newBaseStakeAmount *big.Int, newCpuStakeRate *big.Int, newGpuStakeRate *big.Int, newMemoryStakeRate *big.Int, newDiskStakeRate *big.Int, newUploadSpeedStakeRate *big.Int, newDownloadSpeedStakeRate *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.SetStakeParameters(&_Provider.TransactOpts, newBaseStakeAmount, newCpuStakeRate, newGpuStakeRate, newMemoryStakeRate, newDiskStakeRate, newUploadSpeedStakeRate, newDownloadSpeedStakeRate)
}

// SlashStake is a paid mutator transaction binding the contract method 0x0694f9a0.
//
// Solidity: function slashStake(uint256 providerId, uint256 machineId, uint256 amount, string reason) returns()
func (_Provider *ProviderTransactor) SlashStake(opts *bind.TransactOpts, providerId *big.Int, machineId *big.Int, amount *big.Int, reason string) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "slashStake", providerId, machineId, amount, reason)
}

// SlashStake is a paid mutator transaction binding the contract method 0x0694f9a0.
//
// Solidity: function slashStake(uint256 providerId, uint256 machineId, uint256 amount, string reason) returns()
func (_Provider *ProviderSession) SlashStake(providerId *big.Int, machineId *big.Int, amount *big.Int, reason string) (*types.Transaction, error) {
	return _Provider.Contract.SlashStake(&_Provider.TransactOpts, providerId, machineId, amount, reason)
}

// SlashStake is a paid mutator transaction binding the contract method 0x0694f9a0.
//
// Solidity: function slashStake(uint256 providerId, uint256 machineId, uint256 amount, string reason) returns()
func (_Provider *ProviderTransactorSession) SlashStake(providerId *big.Int, machineId *big.Int, amount *big.Int, reason string) (*types.Transaction, error) {
	return _Provider.Contract.SlashStake(&_Provider.TransactOpts, providerId, machineId, amount, reason)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_Provider *ProviderTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "transferFrom", from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_Provider *ProviderSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.TransferFrom(&_Provider.TransactOpts, from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_Provider *ProviderTransactorSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.TransferFrom(&_Provider.TransactOpts, from, to, tokenId)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Provider *ProviderTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Provider *ProviderSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Provider.Contract.TransferOwnership(&_Provider.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Provider *ProviderTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Provider.Contract.TransferOwnership(&_Provider.TransactOpts, newOwner)
}

// UpdateMachine is a paid mutator transaction binding the contract method 0xbc59863c.
//
// Solidity: function updateMachine(uint256 providerId, uint256 machineId, uint256 cpuCores, uint256 gpuCores, uint256 gpuMemory, uint256 memoryMB, uint256 diskGB, uint256 uploadSpeed, uint256 downloadSpeed, string metadata, uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond) returns()
func (_Provider *ProviderTransactor) UpdateMachine(opts *bind.TransactOpts, providerId *big.Int, machineId *big.Int, cpuCores *big.Int, gpuCores *big.Int, gpuMemory *big.Int, memoryMB *big.Int, diskGB *big.Int, uploadSpeed *big.Int, downloadSpeed *big.Int, metadata string, cpuPricePerSecond *big.Int, gpuPricePerSecond *big.Int, memoryPricePerSecond *big.Int, diskPricePerSecond *big.Int) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "updateMachine", providerId, machineId, cpuCores, gpuCores, gpuMemory, memoryMB, diskGB, uploadSpeed, downloadSpeed, metadata, cpuPricePerSecond, gpuPricePerSecond, memoryPricePerSecond, diskPricePerSecond)
}

// UpdateMachine is a paid mutator transaction binding the contract method 0xbc59863c.
//
// Solidity: function updateMachine(uint256 providerId, uint256 machineId, uint256 cpuCores, uint256 gpuCores, uint256 gpuMemory, uint256 memoryMB, uint256 diskGB, uint256 uploadSpeed, uint256 downloadSpeed, string metadata, uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond) returns()
func (_Provider *ProviderSession) UpdateMachine(providerId *big.Int, machineId *big.Int, cpuCores *big.Int, gpuCores *big.Int, gpuMemory *big.Int, memoryMB *big.Int, diskGB *big.Int, uploadSpeed *big.Int, downloadSpeed *big.Int, metadata string, cpuPricePerSecond *big.Int, gpuPricePerSecond *big.Int, memoryPricePerSecond *big.Int, diskPricePerSecond *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.UpdateMachine(&_Provider.TransactOpts, providerId, machineId, cpuCores, gpuCores, gpuMemory, memoryMB, diskGB, uploadSpeed, downloadSpeed, metadata, cpuPricePerSecond, gpuPricePerSecond, memoryPricePerSecond, diskPricePerSecond)
}

// UpdateMachine is a paid mutator transaction binding the contract method 0xbc59863c.
//
// Solidity: function updateMachine(uint256 providerId, uint256 machineId, uint256 cpuCores, uint256 gpuCores, uint256 gpuMemory, uint256 memoryMB, uint256 diskGB, uint256 uploadSpeed, uint256 downloadSpeed, string metadata, uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond) returns()
func (_Provider *ProviderTransactorSession) UpdateMachine(providerId *big.Int, machineId *big.Int, cpuCores *big.Int, gpuCores *big.Int, gpuMemory *big.Int, memoryMB *big.Int, diskGB *big.Int, uploadSpeed *big.Int, downloadSpeed *big.Int, metadata string, cpuPricePerSecond *big.Int, gpuPricePerSecond *big.Int, memoryPricePerSecond *big.Int, diskPricePerSecond *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.UpdateMachine(&_Provider.TransactOpts, providerId, machineId, cpuCores, gpuCores, gpuMemory, memoryMB, diskGB, uploadSpeed, downloadSpeed, metadata, cpuPricePerSecond, gpuPricePerSecond, memoryPricePerSecond, diskPricePerSecond)
}

// UpdateProviderInfo is a paid mutator transaction binding the contract method 0x3f3e8f24.
//
// Solidity: function updateProviderInfo(uint256 providerId, string metadata) returns()
func (_Provider *ProviderTransactor) UpdateProviderInfo(opts *bind.TransactOpts, providerId *big.Int, metadata string) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "updateProviderInfo", providerId, metadata)
}

// UpdateProviderInfo is a paid mutator transaction binding the contract method 0x3f3e8f24.
//
// Solidity: function updateProviderInfo(uint256 providerId, string metadata) returns()
func (_Provider *ProviderSession) UpdateProviderInfo(providerId *big.Int, metadata string) (*types.Transaction, error) {
	return _Provider.Contract.UpdateProviderInfo(&_Provider.TransactOpts, providerId, metadata)
}

// UpdateProviderInfo is a paid mutator transaction binding the contract method 0x3f3e8f24.
//
// Solidity: function updateProviderInfo(uint256 providerId, string metadata) returns()
func (_Provider *ProviderTransactorSession) UpdateProviderInfo(providerId *big.Int, metadata string) (*types.Transaction, error) {
	return _Provider.Contract.UpdateProviderInfo(&_Provider.TransactOpts, providerId, metadata)
}

// WithdrawSlashedFunds is a paid mutator transaction binding the contract method 0x125eb844.
//
// Solidity: function withdrawSlashedFunds(address recipient, uint256 amount) returns()
func (_Provider *ProviderTransactor) WithdrawSlashedFunds(opts *bind.TransactOpts, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "withdrawSlashedFunds", recipient, amount)
}

// WithdrawSlashedFunds is a paid mutator transaction binding the contract method 0x125eb844.
//
// Solidity: function withdrawSlashedFunds(address recipient, uint256 amount) returns()
func (_Provider *ProviderSession) WithdrawSlashedFunds(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.WithdrawSlashedFunds(&_Provider.TransactOpts, recipient, amount)
}

// WithdrawSlashedFunds is a paid mutator transaction binding the contract method 0x125eb844.
//
// Solidity: function withdrawSlashedFunds(address recipient, uint256 amount) returns()
func (_Provider *ProviderTransactorSession) WithdrawSlashedFunds(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.WithdrawSlashedFunds(&_Provider.TransactOpts, recipient, amount)
}

// ProviderApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the Provider contract.
type ProviderApprovalIterator struct {
	Event *ProviderApproval // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderApproval)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderApproval)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderApproval represents a Approval event raised by the Provider contract.
type ProviderApproval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_Provider *ProviderFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, approved []common.Address, tokenId []*big.Int) (*ProviderApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var approvedRule []interface{}
	for _, approvedItem := range approved {
		approvedRule = append(approvedRule, approvedItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _Provider.contract.FilterLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &ProviderApprovalIterator{contract: _Provider.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_Provider *ProviderFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *ProviderApproval, owner []common.Address, approved []common.Address, tokenId []*big.Int) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var approvedRule []interface{}
	for _, approvedItem := range approved {
		approvedRule = append(approvedRule, approvedItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _Provider.contract.WatchLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderApproval)
				if err := _Provider.contract.UnpackLog(event, "Approval", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_Provider *ProviderFilterer) ParseApproval(log types.Log) (*ProviderApproval, error) {
	event := new(ProviderApproval)
	if err := _Provider.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the Provider contract.
type ProviderApprovalForAllIterator struct {
	Event *ProviderApprovalForAll // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderApprovalForAll)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderApprovalForAll)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderApprovalForAll represents a ApprovalForAll event raised by the Provider contract.
type ProviderApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_Provider *ProviderFilterer) FilterApprovalForAll(opts *bind.FilterOpts, owner []common.Address, operator []common.Address) (*ProviderApprovalForAllIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _Provider.contract.FilterLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return &ProviderApprovalForAllIterator{contract: _Provider.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_Provider *ProviderFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *ProviderApprovalForAll, owner []common.Address, operator []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _Provider.contract.WatchLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderApprovalForAll)
				if err := _Provider.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseApprovalForAll is a log parse operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_Provider *ProviderFilterer) ParseApprovalForAll(log types.Log) (*ProviderApprovalForAll, error) {
	event := new(ProviderApprovalForAll)
	if err := _Provider.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderBatchMetadataUpdateIterator is returned from FilterBatchMetadataUpdate and is used to iterate over the raw logs and unpacked data for BatchMetadataUpdate events raised by the Provider contract.
type ProviderBatchMetadataUpdateIterator struct {
	Event *ProviderBatchMetadataUpdate // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderBatchMetadataUpdateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderBatchMetadataUpdate)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderBatchMetadataUpdate)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderBatchMetadataUpdateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderBatchMetadataUpdateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderBatchMetadataUpdate represents a BatchMetadataUpdate event raised by the Provider contract.
type ProviderBatchMetadataUpdate struct {
	FromTokenId *big.Int
	ToTokenId   *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterBatchMetadataUpdate is a free log retrieval operation binding the contract event 0x6bd5c950a8d8df17f772f5af37cb3655737899cbf903264b9795592da439661c.
//
// Solidity: event BatchMetadataUpdate(uint256 _fromTokenId, uint256 _toTokenId)
func (_Provider *ProviderFilterer) FilterBatchMetadataUpdate(opts *bind.FilterOpts) (*ProviderBatchMetadataUpdateIterator, error) {

	logs, sub, err := _Provider.contract.FilterLogs(opts, "BatchMetadataUpdate")
	if err != nil {
		return nil, err
	}
	return &ProviderBatchMetadataUpdateIterator{contract: _Provider.contract, event: "BatchMetadataUpdate", logs: logs, sub: sub}, nil
}

// WatchBatchMetadataUpdate is a free log subscription operation binding the contract event 0x6bd5c950a8d8df17f772f5af37cb3655737899cbf903264b9795592da439661c.
//
// Solidity: event BatchMetadataUpdate(uint256 _fromTokenId, uint256 _toTokenId)
func (_Provider *ProviderFilterer) WatchBatchMetadataUpdate(opts *bind.WatchOpts, sink chan<- *ProviderBatchMetadataUpdate) (event.Subscription, error) {

	logs, sub, err := _Provider.contract.WatchLogs(opts, "BatchMetadataUpdate")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderBatchMetadataUpdate)
				if err := _Provider.contract.UnpackLog(event, "BatchMetadataUpdate", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseBatchMetadataUpdate is a log parse operation binding the contract event 0x6bd5c950a8d8df17f772f5af37cb3655737899cbf903264b9795592da439661c.
//
// Solidity: event BatchMetadataUpdate(uint256 _fromTokenId, uint256 _toTokenId)
func (_Provider *ProviderFilterer) ParseBatchMetadataUpdate(log types.Log) (*ProviderBatchMetadataUpdate, error) {
	event := new(ProviderBatchMetadataUpdate)
	if err := _Provider.contract.UnpackLog(event, "BatchMetadataUpdate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Provider contract.
type ProviderInitializedIterator struct {
	Event *ProviderInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderInitialized represents a Initialized event raised by the Provider contract.
type ProviderInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Provider *ProviderFilterer) FilterInitialized(opts *bind.FilterOpts) (*ProviderInitializedIterator, error) {

	logs, sub, err := _Provider.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &ProviderInitializedIterator{contract: _Provider.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Provider *ProviderFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *ProviderInitialized) (event.Subscription, error) {

	logs, sub, err := _Provider.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderInitialized)
				if err := _Provider.contract.UnpackLog(event, "Initialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitialized is a log parse operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Provider *ProviderFilterer) ParseInitialized(log types.Log) (*ProviderInitialized, error) {
	event := new(ProviderInitialized)
	if err := _Provider.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderLockPeriodUpdatedIterator is returned from FilterLockPeriodUpdated and is used to iterate over the raw logs and unpacked data for LockPeriodUpdated events raised by the Provider contract.
type ProviderLockPeriodUpdatedIterator struct {
	Event *ProviderLockPeriodUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderLockPeriodUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderLockPeriodUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderLockPeriodUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderLockPeriodUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderLockPeriodUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderLockPeriodUpdated represents a LockPeriodUpdated event raised by the Provider contract.
type ProviderLockPeriodUpdated struct {
	OldPeriod *big.Int
	NewPeriod *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterLockPeriodUpdated is a free log retrieval operation binding the contract event 0x5dd7b309b7bc5010d9c96159ee535a428121d6803cb847792402fffccaf1569a.
//
// Solidity: event LockPeriodUpdated(uint256 oldPeriod, uint256 newPeriod)
func (_Provider *ProviderFilterer) FilterLockPeriodUpdated(opts *bind.FilterOpts) (*ProviderLockPeriodUpdatedIterator, error) {

	logs, sub, err := _Provider.contract.FilterLogs(opts, "LockPeriodUpdated")
	if err != nil {
		return nil, err
	}
	return &ProviderLockPeriodUpdatedIterator{contract: _Provider.contract, event: "LockPeriodUpdated", logs: logs, sub: sub}, nil
}

// WatchLockPeriodUpdated is a free log subscription operation binding the contract event 0x5dd7b309b7bc5010d9c96159ee535a428121d6803cb847792402fffccaf1569a.
//
// Solidity: event LockPeriodUpdated(uint256 oldPeriod, uint256 newPeriod)
func (_Provider *ProviderFilterer) WatchLockPeriodUpdated(opts *bind.WatchOpts, sink chan<- *ProviderLockPeriodUpdated) (event.Subscription, error) {

	logs, sub, err := _Provider.contract.WatchLogs(opts, "LockPeriodUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderLockPeriodUpdated)
				if err := _Provider.contract.UnpackLog(event, "LockPeriodUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseLockPeriodUpdated is a log parse operation binding the contract event 0x5dd7b309b7bc5010d9c96159ee535a428121d6803cb847792402fffccaf1569a.
//
// Solidity: event LockPeriodUpdated(uint256 oldPeriod, uint256 newPeriod)
func (_Provider *ProviderFilterer) ParseLockPeriodUpdated(log types.Log) (*ProviderLockPeriodUpdated, error) {
	event := new(ProviderLockPeriodUpdated)
	if err := _Provider.contract.UnpackLog(event, "LockPeriodUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderMachineAddedIterator is returned from FilterMachineAdded and is used to iterate over the raw logs and unpacked data for MachineAdded events raised by the Provider contract.
type ProviderMachineAddedIterator struct {
	Event *ProviderMachineAdded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderMachineAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderMachineAdded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderMachineAdded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderMachineAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderMachineAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderMachineAdded represents a MachineAdded event raised by the Provider contract.
type ProviderMachineAdded struct {
	ProviderId   *big.Int
	MachineId    *big.Int
	StakedAmount *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterMachineAdded is a free log retrieval operation binding the contract event 0x0394bc4be749714e8af65a6117ae27affa1aedd69805edd2a533c74eb14963a8.
//
// Solidity: event MachineAdded(uint256 indexed providerId, uint256 machineId, uint256 stakedAmount)
func (_Provider *ProviderFilterer) FilterMachineAdded(opts *bind.FilterOpts, providerId []*big.Int) (*ProviderMachineAddedIterator, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _Provider.contract.FilterLogs(opts, "MachineAdded", providerIdRule)
	if err != nil {
		return nil, err
	}
	return &ProviderMachineAddedIterator{contract: _Provider.contract, event: "MachineAdded", logs: logs, sub: sub}, nil
}

// WatchMachineAdded is a free log subscription operation binding the contract event 0x0394bc4be749714e8af65a6117ae27affa1aedd69805edd2a533c74eb14963a8.
//
// Solidity: event MachineAdded(uint256 indexed providerId, uint256 machineId, uint256 stakedAmount)
func (_Provider *ProviderFilterer) WatchMachineAdded(opts *bind.WatchOpts, sink chan<- *ProviderMachineAdded, providerId []*big.Int) (event.Subscription, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _Provider.contract.WatchLogs(opts, "MachineAdded", providerIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderMachineAdded)
				if err := _Provider.contract.UnpackLog(event, "MachineAdded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMachineAdded is a log parse operation binding the contract event 0x0394bc4be749714e8af65a6117ae27affa1aedd69805edd2a533c74eb14963a8.
//
// Solidity: event MachineAdded(uint256 indexed providerId, uint256 machineId, uint256 stakedAmount)
func (_Provider *ProviderFilterer) ParseMachineAdded(log types.Log) (*ProviderMachineAdded, error) {
	event := new(ProviderMachineAdded)
	if err := _Provider.contract.UnpackLog(event, "MachineAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderMachineRemovedIterator is returned from FilterMachineRemoved and is used to iterate over the raw logs and unpacked data for MachineRemoved events raised by the Provider contract.
type ProviderMachineRemovedIterator struct {
	Event *ProviderMachineRemoved // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderMachineRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderMachineRemoved)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderMachineRemoved)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderMachineRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderMachineRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderMachineRemoved represents a MachineRemoved event raised by the Provider contract.
type ProviderMachineRemoved struct {
	ProviderId *big.Int
	MachineId  *big.Int
	Unlocktime *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterMachineRemoved is a free log retrieval operation binding the contract event 0xf90af2eb19bb153aba9ed20eb295d13b68337404e2561b6076c1c76c55af1098.
//
// Solidity: event MachineRemoved(uint256 indexed providerId, uint256 machineId, uint256 unlocktime)
func (_Provider *ProviderFilterer) FilterMachineRemoved(opts *bind.FilterOpts, providerId []*big.Int) (*ProviderMachineRemovedIterator, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _Provider.contract.FilterLogs(opts, "MachineRemoved", providerIdRule)
	if err != nil {
		return nil, err
	}
	return &ProviderMachineRemovedIterator{contract: _Provider.contract, event: "MachineRemoved", logs: logs, sub: sub}, nil
}

// WatchMachineRemoved is a free log subscription operation binding the contract event 0xf90af2eb19bb153aba9ed20eb295d13b68337404e2561b6076c1c76c55af1098.
//
// Solidity: event MachineRemoved(uint256 indexed providerId, uint256 machineId, uint256 unlocktime)
func (_Provider *ProviderFilterer) WatchMachineRemoved(opts *bind.WatchOpts, sink chan<- *ProviderMachineRemoved, providerId []*big.Int) (event.Subscription, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _Provider.contract.WatchLogs(opts, "MachineRemoved", providerIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderMachineRemoved)
				if err := _Provider.contract.UnpackLog(event, "MachineRemoved", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMachineRemoved is a log parse operation binding the contract event 0xf90af2eb19bb153aba9ed20eb295d13b68337404e2561b6076c1c76c55af1098.
//
// Solidity: event MachineRemoved(uint256 indexed providerId, uint256 machineId, uint256 unlocktime)
func (_Provider *ProviderFilterer) ParseMachineRemoved(log types.Log) (*ProviderMachineRemoved, error) {
	event := new(ProviderMachineRemoved)
	if err := _Provider.contract.UnpackLog(event, "MachineRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderMachineResourcePriceUpdatedIterator is returned from FilterMachineResourcePriceUpdated and is used to iterate over the raw logs and unpacked data for MachineResourcePriceUpdated events raised by the Provider contract.
type ProviderMachineResourcePriceUpdatedIterator struct {
	Event *ProviderMachineResourcePriceUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderMachineResourcePriceUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderMachineResourcePriceUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderMachineResourcePriceUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderMachineResourcePriceUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderMachineResourcePriceUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderMachineResourcePriceUpdated represents a MachineResourcePriceUpdated event raised by the Provider contract.
type ProviderMachineResourcePriceUpdated struct {
	ProviderId           *big.Int
	MachineId            *big.Int
	CpuPricePerSecond    *big.Int
	GpuPricePerSecond    *big.Int
	MemoryPricePerSecond *big.Int
	DiskPricePerSecond   *big.Int
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterMachineResourcePriceUpdated is a free log retrieval operation binding the contract event 0x87c6e61ce2351914ceb4928d7104755f640d3a7ad062a1228110fc0767b03521.
//
// Solidity: event MachineResourcePriceUpdated(uint256 indexed providerId, uint256 indexed machineId, uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond)
func (_Provider *ProviderFilterer) FilterMachineResourcePriceUpdated(opts *bind.FilterOpts, providerId []*big.Int, machineId []*big.Int) (*ProviderMachineResourcePriceUpdatedIterator, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}
	var machineIdRule []interface{}
	for _, machineIdItem := range machineId {
		machineIdRule = append(machineIdRule, machineIdItem)
	}

	logs, sub, err := _Provider.contract.FilterLogs(opts, "MachineResourcePriceUpdated", providerIdRule, machineIdRule)
	if err != nil {
		return nil, err
	}
	return &ProviderMachineResourcePriceUpdatedIterator{contract: _Provider.contract, event: "MachineResourcePriceUpdated", logs: logs, sub: sub}, nil
}

// WatchMachineResourcePriceUpdated is a free log subscription operation binding the contract event 0x87c6e61ce2351914ceb4928d7104755f640d3a7ad062a1228110fc0767b03521.
//
// Solidity: event MachineResourcePriceUpdated(uint256 indexed providerId, uint256 indexed machineId, uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond)
func (_Provider *ProviderFilterer) WatchMachineResourcePriceUpdated(opts *bind.WatchOpts, sink chan<- *ProviderMachineResourcePriceUpdated, providerId []*big.Int, machineId []*big.Int) (event.Subscription, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}
	var machineIdRule []interface{}
	for _, machineIdItem := range machineId {
		machineIdRule = append(machineIdRule, machineIdItem)
	}

	logs, sub, err := _Provider.contract.WatchLogs(opts, "MachineResourcePriceUpdated", providerIdRule, machineIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderMachineResourcePriceUpdated)
				if err := _Provider.contract.UnpackLog(event, "MachineResourcePriceUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMachineResourcePriceUpdated is a log parse operation binding the contract event 0x87c6e61ce2351914ceb4928d7104755f640d3a7ad062a1228110fc0767b03521.
//
// Solidity: event MachineResourcePriceUpdated(uint256 indexed providerId, uint256 indexed machineId, uint256 cpuPricePerSecond, uint256 gpuPricePerSecond, uint256 memoryPricePerSecond, uint256 diskPricePerSecond)
func (_Provider *ProviderFilterer) ParseMachineResourcePriceUpdated(log types.Log) (*ProviderMachineResourcePriceUpdated, error) {
	event := new(ProviderMachineResourcePriceUpdated)
	if err := _Provider.contract.UnpackLog(event, "MachineResourcePriceUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderMachineUpdatedIterator is returned from FilterMachineUpdated and is used to iterate over the raw logs and unpacked data for MachineUpdated events raised by the Provider contract.
type ProviderMachineUpdatedIterator struct {
	Event *ProviderMachineUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderMachineUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderMachineUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderMachineUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderMachineUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderMachineUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderMachineUpdated represents a MachineUpdated event raised by the Provider contract.
type ProviderMachineUpdated struct {
	ProviderId      *big.Int
	MachineId       *big.Int
	AdditionalStake *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterMachineUpdated is a free log retrieval operation binding the contract event 0xae3cd44fcb506ab91f575f2f4406dfe7e64a89d903efe251237bdd2b09dd0778.
//
// Solidity: event MachineUpdated(uint256 indexed providerId, uint256 machineId, uint256 additionalStake)
func (_Provider *ProviderFilterer) FilterMachineUpdated(opts *bind.FilterOpts, providerId []*big.Int) (*ProviderMachineUpdatedIterator, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _Provider.contract.FilterLogs(opts, "MachineUpdated", providerIdRule)
	if err != nil {
		return nil, err
	}
	return &ProviderMachineUpdatedIterator{contract: _Provider.contract, event: "MachineUpdated", logs: logs, sub: sub}, nil
}

// WatchMachineUpdated is a free log subscription operation binding the contract event 0xae3cd44fcb506ab91f575f2f4406dfe7e64a89d903efe251237bdd2b09dd0778.
//
// Solidity: event MachineUpdated(uint256 indexed providerId, uint256 machineId, uint256 additionalStake)
func (_Provider *ProviderFilterer) WatchMachineUpdated(opts *bind.WatchOpts, sink chan<- *ProviderMachineUpdated, providerId []*big.Int) (event.Subscription, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _Provider.contract.WatchLogs(opts, "MachineUpdated", providerIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderMachineUpdated)
				if err := _Provider.contract.UnpackLog(event, "MachineUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMachineUpdated is a log parse operation binding the contract event 0xae3cd44fcb506ab91f575f2f4406dfe7e64a89d903efe251237bdd2b09dd0778.
//
// Solidity: event MachineUpdated(uint256 indexed providerId, uint256 machineId, uint256 additionalStake)
func (_Provider *ProviderFilterer) ParseMachineUpdated(log types.Log) (*ProviderMachineUpdated, error) {
	event := new(ProviderMachineUpdated)
	if err := _Provider.contract.UnpackLog(event, "MachineUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderMetadataUpdateIterator is returned from FilterMetadataUpdate and is used to iterate over the raw logs and unpacked data for MetadataUpdate events raised by the Provider contract.
type ProviderMetadataUpdateIterator struct {
	Event *ProviderMetadataUpdate // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderMetadataUpdateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderMetadataUpdate)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderMetadataUpdate)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderMetadataUpdateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderMetadataUpdateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderMetadataUpdate represents a MetadataUpdate event raised by the Provider contract.
type ProviderMetadataUpdate struct {
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterMetadataUpdate is a free log retrieval operation binding the contract event 0xf8e1a15aba9398e019f0b49df1a4fde98ee17ae345cb5f6b5e2c27f5033e8ce7.
//
// Solidity: event MetadataUpdate(uint256 _tokenId)
func (_Provider *ProviderFilterer) FilterMetadataUpdate(opts *bind.FilterOpts) (*ProviderMetadataUpdateIterator, error) {

	logs, sub, err := _Provider.contract.FilterLogs(opts, "MetadataUpdate")
	if err != nil {
		return nil, err
	}
	return &ProviderMetadataUpdateIterator{contract: _Provider.contract, event: "MetadataUpdate", logs: logs, sub: sub}, nil
}

// WatchMetadataUpdate is a free log subscription operation binding the contract event 0xf8e1a15aba9398e019f0b49df1a4fde98ee17ae345cb5f6b5e2c27f5033e8ce7.
//
// Solidity: event MetadataUpdate(uint256 _tokenId)
func (_Provider *ProviderFilterer) WatchMetadataUpdate(opts *bind.WatchOpts, sink chan<- *ProviderMetadataUpdate) (event.Subscription, error) {

	logs, sub, err := _Provider.contract.WatchLogs(opts, "MetadataUpdate")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderMetadataUpdate)
				if err := _Provider.contract.UnpackLog(event, "MetadataUpdate", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMetadataUpdate is a log parse operation binding the contract event 0xf8e1a15aba9398e019f0b49df1a4fde98ee17ae345cb5f6b5e2c27f5033e8ce7.
//
// Solidity: event MetadataUpdate(uint256 _tokenId)
func (_Provider *ProviderFilterer) ParseMetadataUpdate(log types.Log) (*ProviderMetadataUpdate, error) {
	event := new(ProviderMetadataUpdate)
	if err := _Provider.contract.UnpackLog(event, "MetadataUpdate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Provider contract.
type ProviderOwnershipTransferredIterator struct {
	Event *ProviderOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderOwnershipTransferred represents a OwnershipTransferred event raised by the Provider contract.
type ProviderOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Provider *ProviderFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ProviderOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Provider.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ProviderOwnershipTransferredIterator{contract: _Provider.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Provider *ProviderFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ProviderOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Provider.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderOwnershipTransferred)
				if err := _Provider.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Provider *ProviderFilterer) ParseOwnershipTransferred(log types.Log) (*ProviderOwnershipTransferred, error) {
	event := new(ProviderOwnershipTransferred)
	if err := _Provider.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderProviderReputationUpdatedIterator is returned from FilterProviderReputationUpdated and is used to iterate over the raw logs and unpacked data for ProviderReputationUpdated events raised by the Provider contract.
type ProviderProviderReputationUpdatedIterator struct {
	Event *ProviderProviderReputationUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderProviderReputationUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderProviderReputationUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderProviderReputationUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderProviderReputationUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderProviderReputationUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderProviderReputationUpdated represents a ProviderReputationUpdated event raised by the Provider contract.
type ProviderProviderReputationUpdated struct {
	ProviderId    *big.Int
	NewReputation *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterProviderReputationUpdated is a free log retrieval operation binding the contract event 0xda7de51a3c193d4a1046929774737666fb7a005e8e5e1a98e146a99952ad223e.
//
// Solidity: event ProviderReputationUpdated(uint256 indexed providerId, uint256 newReputation)
func (_Provider *ProviderFilterer) FilterProviderReputationUpdated(opts *bind.FilterOpts, providerId []*big.Int) (*ProviderProviderReputationUpdatedIterator, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _Provider.contract.FilterLogs(opts, "ProviderReputationUpdated", providerIdRule)
	if err != nil {
		return nil, err
	}
	return &ProviderProviderReputationUpdatedIterator{contract: _Provider.contract, event: "ProviderReputationUpdated", logs: logs, sub: sub}, nil
}

// WatchProviderReputationUpdated is a free log subscription operation binding the contract event 0xda7de51a3c193d4a1046929774737666fb7a005e8e5e1a98e146a99952ad223e.
//
// Solidity: event ProviderReputationUpdated(uint256 indexed providerId, uint256 newReputation)
func (_Provider *ProviderFilterer) WatchProviderReputationUpdated(opts *bind.WatchOpts, sink chan<- *ProviderProviderReputationUpdated, providerId []*big.Int) (event.Subscription, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _Provider.contract.WatchLogs(opts, "ProviderReputationUpdated", providerIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderProviderReputationUpdated)
				if err := _Provider.contract.UnpackLog(event, "ProviderReputationUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseProviderReputationUpdated is a log parse operation binding the contract event 0xda7de51a3c193d4a1046929774737666fb7a005e8e5e1a98e146a99952ad223e.
//
// Solidity: event ProviderReputationUpdated(uint256 indexed providerId, uint256 newReputation)
func (_Provider *ProviderFilterer) ParseProviderReputationUpdated(log types.Log) (*ProviderProviderReputationUpdated, error) {
	event := new(ProviderProviderReputationUpdated)
	if err := _Provider.contract.UnpackLog(event, "ProviderReputationUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderProviderUpdatedIterator is returned from FilterProviderUpdated and is used to iterate over the raw logs and unpacked data for ProviderUpdated events raised by the Provider contract.
type ProviderProviderUpdatedIterator struct {
	Event *ProviderProviderUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderProviderUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderProviderUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderProviderUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderProviderUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderProviderUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderProviderUpdated represents a ProviderUpdated event raised by the Provider contract.
type ProviderProviderUpdated struct {
	ProviderId *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterProviderUpdated is a free log retrieval operation binding the contract event 0xd78fa84b17fc71eb590722ba7d04627e81f434575b079bd714d83f03911ca7e8.
//
// Solidity: event ProviderUpdated(uint256 indexed providerId)
func (_Provider *ProviderFilterer) FilterProviderUpdated(opts *bind.FilterOpts, providerId []*big.Int) (*ProviderProviderUpdatedIterator, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _Provider.contract.FilterLogs(opts, "ProviderUpdated", providerIdRule)
	if err != nil {
		return nil, err
	}
	return &ProviderProviderUpdatedIterator{contract: _Provider.contract, event: "ProviderUpdated", logs: logs, sub: sub}, nil
}

// WatchProviderUpdated is a free log subscription operation binding the contract event 0xd78fa84b17fc71eb590722ba7d04627e81f434575b079bd714d83f03911ca7e8.
//
// Solidity: event ProviderUpdated(uint256 indexed providerId)
func (_Provider *ProviderFilterer) WatchProviderUpdated(opts *bind.WatchOpts, sink chan<- *ProviderProviderUpdated, providerId []*big.Int) (event.Subscription, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _Provider.contract.WatchLogs(opts, "ProviderUpdated", providerIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderProviderUpdated)
				if err := _Provider.contract.UnpackLog(event, "ProviderUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseProviderUpdated is a log parse operation binding the contract event 0xd78fa84b17fc71eb590722ba7d04627e81f434575b079bd714d83f03911ca7e8.
//
// Solidity: event ProviderUpdated(uint256 indexed providerId)
func (_Provider *ProviderFilterer) ParseProviderUpdated(log types.Log) (*ProviderProviderUpdated, error) {
	event := new(ProviderProviderUpdated)
	if err := _Provider.contract.UnpackLog(event, "ProviderUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderProviderVerifiedIterator is returned from FilterProviderVerified and is used to iterate over the raw logs and unpacked data for ProviderVerified events raised by the Provider contract.
type ProviderProviderVerifiedIterator struct {
	Event *ProviderProviderVerified // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderProviderVerifiedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderProviderVerified)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderProviderVerified)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderProviderVerifiedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderProviderVerifiedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderProviderVerified represents a ProviderVerified event raised by the Provider contract.
type ProviderProviderVerified struct {
	ProviderId *big.Int
	Verified   bool
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterProviderVerified is a free log retrieval operation binding the contract event 0x73c0906694486d74b07388321c8dee7f4c2d58982c9c4810d24c0b0d0ebf3bbd.
//
// Solidity: event ProviderVerified(uint256 indexed providerId, bool verified)
func (_Provider *ProviderFilterer) FilterProviderVerified(opts *bind.FilterOpts, providerId []*big.Int) (*ProviderProviderVerifiedIterator, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _Provider.contract.FilterLogs(opts, "ProviderVerified", providerIdRule)
	if err != nil {
		return nil, err
	}
	return &ProviderProviderVerifiedIterator{contract: _Provider.contract, event: "ProviderVerified", logs: logs, sub: sub}, nil
}

// WatchProviderVerified is a free log subscription operation binding the contract event 0x73c0906694486d74b07388321c8dee7f4c2d58982c9c4810d24c0b0d0ebf3bbd.
//
// Solidity: event ProviderVerified(uint256 indexed providerId, bool verified)
func (_Provider *ProviderFilterer) WatchProviderVerified(opts *bind.WatchOpts, sink chan<- *ProviderProviderVerified, providerId []*big.Int) (event.Subscription, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _Provider.contract.WatchLogs(opts, "ProviderVerified", providerIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderProviderVerified)
				if err := _Provider.contract.UnpackLog(event, "ProviderVerified", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseProviderVerified is a log parse operation binding the contract event 0x73c0906694486d74b07388321c8dee7f4c2d58982c9c4810d24c0b0d0ebf3bbd.
//
// Solidity: event ProviderVerified(uint256 indexed providerId, bool verified)
func (_Provider *ProviderFilterer) ParseProviderVerified(log types.Log) (*ProviderProviderVerified, error) {
	event := new(ProviderProviderVerified)
	if err := _Provider.contract.UnpackLog(event, "ProviderVerified", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderStakeParametersUpdatedIterator is returned from FilterStakeParametersUpdated and is used to iterate over the raw logs and unpacked data for StakeParametersUpdated events raised by the Provider contract.
type ProviderStakeParametersUpdatedIterator struct {
	Event *ProviderStakeParametersUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderStakeParametersUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderStakeParametersUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderStakeParametersUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderStakeParametersUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderStakeParametersUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderStakeParametersUpdated represents a StakeParametersUpdated event raised by the Provider contract.
type ProviderStakeParametersUpdated struct {
	BaseAmount        *big.Int
	CpuRate           *big.Int
	GpuRate           *big.Int
	MemoryRate        *big.Int
	DiskRate          *big.Int
	UploadSpeedRate   *big.Int
	DownloadSpeedRate *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterStakeParametersUpdated is a free log retrieval operation binding the contract event 0x3a9006f72d836e00b5cb8cb5e53557a819f629eb3d93bd037f16590b15be9aad.
//
// Solidity: event StakeParametersUpdated(uint256 baseAmount, uint256 cpuRate, uint256 gpuRate, uint256 memoryRate, uint256 diskRate, uint256 uploadSpeedRate, uint256 downloadSpeedRate)
func (_Provider *ProviderFilterer) FilterStakeParametersUpdated(opts *bind.FilterOpts) (*ProviderStakeParametersUpdatedIterator, error) {

	logs, sub, err := _Provider.contract.FilterLogs(opts, "StakeParametersUpdated")
	if err != nil {
		return nil, err
	}
	return &ProviderStakeParametersUpdatedIterator{contract: _Provider.contract, event: "StakeParametersUpdated", logs: logs, sub: sub}, nil
}

// WatchStakeParametersUpdated is a free log subscription operation binding the contract event 0x3a9006f72d836e00b5cb8cb5e53557a819f629eb3d93bd037f16590b15be9aad.
//
// Solidity: event StakeParametersUpdated(uint256 baseAmount, uint256 cpuRate, uint256 gpuRate, uint256 memoryRate, uint256 diskRate, uint256 uploadSpeedRate, uint256 downloadSpeedRate)
func (_Provider *ProviderFilterer) WatchStakeParametersUpdated(opts *bind.WatchOpts, sink chan<- *ProviderStakeParametersUpdated) (event.Subscription, error) {

	logs, sub, err := _Provider.contract.WatchLogs(opts, "StakeParametersUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderStakeParametersUpdated)
				if err := _Provider.contract.UnpackLog(event, "StakeParametersUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseStakeParametersUpdated is a log parse operation binding the contract event 0x3a9006f72d836e00b5cb8cb5e53557a819f629eb3d93bd037f16590b15be9aad.
//
// Solidity: event StakeParametersUpdated(uint256 baseAmount, uint256 cpuRate, uint256 gpuRate, uint256 memoryRate, uint256 diskRate, uint256 uploadSpeedRate, uint256 downloadSpeedRate)
func (_Provider *ProviderFilterer) ParseStakeParametersUpdated(log types.Log) (*ProviderStakeParametersUpdated, error) {
	event := new(ProviderStakeParametersUpdated)
	if err := _Provider.contract.UnpackLog(event, "StakeParametersUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderStakeSlashedIterator is returned from FilterStakeSlashed and is used to iterate over the raw logs and unpacked data for StakeSlashed events raised by the Provider contract.
type ProviderStakeSlashedIterator struct {
	Event *ProviderStakeSlashed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderStakeSlashedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderStakeSlashed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderStakeSlashed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderStakeSlashedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderStakeSlashedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderStakeSlashed represents a StakeSlashed event raised by the Provider contract.
type ProviderStakeSlashed struct {
	ProviderId *big.Int
	MachineId  *big.Int
	Amount     *big.Int
	Reason     string
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterStakeSlashed is a free log retrieval operation binding the contract event 0xc6fde9f3f4d68e725ce8eb50a65dbedf1a312d0578e7bd93ab6ac14066f8fbd5.
//
// Solidity: event StakeSlashed(uint256 indexed providerId, uint256 machineId, uint256 amount, string reason)
func (_Provider *ProviderFilterer) FilterStakeSlashed(opts *bind.FilterOpts, providerId []*big.Int) (*ProviderStakeSlashedIterator, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _Provider.contract.FilterLogs(opts, "StakeSlashed", providerIdRule)
	if err != nil {
		return nil, err
	}
	return &ProviderStakeSlashedIterator{contract: _Provider.contract, event: "StakeSlashed", logs: logs, sub: sub}, nil
}

// WatchStakeSlashed is a free log subscription operation binding the contract event 0xc6fde9f3f4d68e725ce8eb50a65dbedf1a312d0578e7bd93ab6ac14066f8fbd5.
//
// Solidity: event StakeSlashed(uint256 indexed providerId, uint256 machineId, uint256 amount, string reason)
func (_Provider *ProviderFilterer) WatchStakeSlashed(opts *bind.WatchOpts, sink chan<- *ProviderStakeSlashed, providerId []*big.Int) (event.Subscription, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _Provider.contract.WatchLogs(opts, "StakeSlashed", providerIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderStakeSlashed)
				if err := _Provider.contract.UnpackLog(event, "StakeSlashed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseStakeSlashed is a log parse operation binding the contract event 0xc6fde9f3f4d68e725ce8eb50a65dbedf1a312d0578e7bd93ab6ac14066f8fbd5.
//
// Solidity: event StakeSlashed(uint256 indexed providerId, uint256 machineId, uint256 amount, string reason)
func (_Provider *ProviderFilterer) ParseStakeSlashed(log types.Log) (*ProviderStakeSlashed, error) {
	event := new(ProviderStakeSlashed)
	if err := _Provider.contract.UnpackLog(event, "StakeSlashed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderStakeWithdrawnIterator is returned from FilterStakeWithdrawn and is used to iterate over the raw logs and unpacked data for StakeWithdrawn events raised by the Provider contract.
type ProviderStakeWithdrawnIterator struct {
	Event *ProviderStakeWithdrawn // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderStakeWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderStakeWithdrawn)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderStakeWithdrawn)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderStakeWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderStakeWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderStakeWithdrawn represents a StakeWithdrawn event raised by the Provider contract.
type ProviderStakeWithdrawn struct {
	ProviderId *big.Int
	MachineId  *big.Int
	Amount     *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterStakeWithdrawn is a free log retrieval operation binding the contract event 0xcda86d0ec3f513f43f5512de1982307d90747e4da0550585ca8744cbc7d8619f.
//
// Solidity: event StakeWithdrawn(uint256 indexed providerId, uint256 machineId, uint256 amount)
func (_Provider *ProviderFilterer) FilterStakeWithdrawn(opts *bind.FilterOpts, providerId []*big.Int) (*ProviderStakeWithdrawnIterator, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _Provider.contract.FilterLogs(opts, "StakeWithdrawn", providerIdRule)
	if err != nil {
		return nil, err
	}
	return &ProviderStakeWithdrawnIterator{contract: _Provider.contract, event: "StakeWithdrawn", logs: logs, sub: sub}, nil
}

// WatchStakeWithdrawn is a free log subscription operation binding the contract event 0xcda86d0ec3f513f43f5512de1982307d90747e4da0550585ca8744cbc7d8619f.
//
// Solidity: event StakeWithdrawn(uint256 indexed providerId, uint256 machineId, uint256 amount)
func (_Provider *ProviderFilterer) WatchStakeWithdrawn(opts *bind.WatchOpts, sink chan<- *ProviderStakeWithdrawn, providerId []*big.Int) (event.Subscription, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _Provider.contract.WatchLogs(opts, "StakeWithdrawn", providerIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderStakeWithdrawn)
				if err := _Provider.contract.UnpackLog(event, "StakeWithdrawn", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseStakeWithdrawn is a log parse operation binding the contract event 0xcda86d0ec3f513f43f5512de1982307d90747e4da0550585ca8744cbc7d8619f.
//
// Solidity: event StakeWithdrawn(uint256 indexed providerId, uint256 machineId, uint256 amount)
func (_Provider *ProviderFilterer) ParseStakeWithdrawn(log types.Log) (*ProviderStakeWithdrawn, error) {
	event := new(ProviderStakeWithdrawn)
	if err := _Provider.contract.UnpackLog(event, "StakeWithdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the Provider contract.
type ProviderTransferIterator struct {
	Event *ProviderTransfer // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProviderTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderTransfer)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProviderTransfer)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProviderTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderTransfer represents a Transfer event raised by the Provider contract.
type ProviderTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_Provider *ProviderFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address, tokenId []*big.Int) (*ProviderTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _Provider.contract.FilterLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &ProviderTransferIterator{contract: _Provider.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_Provider *ProviderFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *ProviderTransfer, from []common.Address, to []common.Address, tokenId []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _Provider.contract.WatchLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderTransfer)
				if err := _Provider.contract.UnpackLog(event, "Transfer", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_Provider *ProviderFilterer) ParseTransfer(log types.Log) (*ProviderTransfer, error) {
	event := new(ProviderTransfer)
	if err := _Provider.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
