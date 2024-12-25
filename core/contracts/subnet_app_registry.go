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

// SubnetAppRegistryApp is an auto generated low-level Go binding around an user-defined struct.
type SubnetAppRegistryApp struct {
	PeerId               string
	Owner                common.Address
	Name                 string
	Symbol               string
	Budget               *big.Int
	SpentBudget          *big.Int
	MaxNodes             *big.Int
	MinCpu               *big.Int
	MinGpu               *big.Int
	MinMemory            *big.Int
	MinUploadBandwidth   *big.Int
	MinDownloadBandwidth *big.Int
	NodeCount            *big.Int
	PricePerCpu          *big.Int
	PricePerGpu          *big.Int
	PricePerMemoryGB     *big.Int
	PricePerStorageGB    *big.Int
	PricePerBandwidthGB  *big.Int
	Metadata             string
}

// SubnetAppRegistryAppNode is an auto generated low-level Go binding around an user-defined struct.
type SubnetAppRegistryAppNode struct {
	Duration          *big.Int
	UsedCpu           *big.Int
	UsedGpu           *big.Int
	UsedMemory        *big.Int
	UsedStorage       *big.Int
	UsedDownloadBytes *big.Int
	UsedUploadBytes   *big.Int
	IsRegistered      bool
}

// SubnetAppRegistryMetaData contains all meta data concerning the SubnetAppRegistry contract.
var SubnetAppRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_subnetRegistry\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"initialOwner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_treasury\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_feeRate\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"ECDSAInvalidSignature\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"length\",\"type\":\"uint256\"}],\"name\":\"ECDSAInvalidSignatureLength\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"ECDSAInvalidSignatureS\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidShortString\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"str\",\"type\":\"string\"}],\"name\":\"StringTooLong\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"budget\",\"type\":\"uint256\"}],\"name\":\"AppCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"}],\"name\":\"AppUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EIP712DomainChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"subnetId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"NodeRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"subnetId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"node\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"reward\",\"type\":\"uint256\"}],\"name\":\"RewardClaimed\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"appCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"appNodes\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedMemory\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedStorage\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedDownloadBytes\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedUploadBytes\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isRegistered\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"apps\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"budget\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"spentBudget\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxNodes\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minMemory\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minUploadBandwidth\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minDownloadBandwidth\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nodeCount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerMemoryGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerStorageGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerBandwidthGB\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"subnetId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedMemory\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedStorage\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedUploadBytes\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedDownloadBytes\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"claimReward\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"budget\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxNodes\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minMemory\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minUploadBandwidth\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minDownloadBandwidth\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerMemoryGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerStorageGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerBandwidthGB\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"}],\"name\":\"createApp\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"eip712Domain\",\"outputs\":[{\"internalType\":\"bytes1\",\"name\":\"fields\",\"type\":\"bytes1\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"verifyingContract\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"salt\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"extensions\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"feeRate\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"}],\"name\":\"getApp\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"budget\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"spentBudget\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxNodes\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minMemory\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minUploadBandwidth\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minDownloadBandwidth\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nodeCount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerMemoryGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerStorageGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerBandwidthGB\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"}],\"internalType\":\"structSubnetAppRegistry.App\",\"name\":\"app\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"subnetId\",\"type\":\"uint256\"}],\"name\":\"getAppNode\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedMemory\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedStorage\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedDownloadBytes\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedUploadBytes\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isRegistered\",\"type\":\"bool\"}],\"internalType\":\"structSubnetAppRegistry.AppNode\",\"name\":\"appNode\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"start\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"end\",\"type\":\"uint256\"}],\"name\":\"listApps\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"budget\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"spentBudget\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxNodes\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minMemory\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minUploadBandwidth\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minDownloadBandwidth\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nodeCount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerMemoryGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerStorageGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerBandwidthGB\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"}],\"internalType\":\"structSubnetAppRegistry.App[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"nodeToAppId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"subnetId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"}],\"name\":\"registerNode\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_feeRate\",\"type\":\"uint256\"}],\"name\":\"setFeeRate\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_treasury\",\"type\":\"address\"}],\"name\":\"setTreasury\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"subnetRegistry\",\"outputs\":[{\"internalType\":\"contractISubnetRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"symbolToAppId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"treasury\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"maxNodes\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minMemory\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minUploadBandwidth\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minDownloadBandwidth\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerMemoryGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerStorageGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerBandwidthGB\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"}],\"name\":\"updateApp\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"usedMessageHashes\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]",
}

// SubnetAppRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use SubnetAppRegistryMetaData.ABI instead.
var SubnetAppRegistryABI = SubnetAppRegistryMetaData.ABI

// SubnetAppRegistry is an auto generated Go binding around an Ethereum contract.
type SubnetAppRegistry struct {
	SubnetAppRegistryCaller     // Read-only binding to the contract
	SubnetAppRegistryTransactor // Write-only binding to the contract
	SubnetAppRegistryFilterer   // Log filterer for contract events
}

// SubnetAppRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type SubnetAppRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetAppRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SubnetAppRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetAppRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SubnetAppRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetAppRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SubnetAppRegistrySession struct {
	Contract     *SubnetAppRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// SubnetAppRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SubnetAppRegistryCallerSession struct {
	Contract *SubnetAppRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// SubnetAppRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SubnetAppRegistryTransactorSession struct {
	Contract     *SubnetAppRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// SubnetAppRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type SubnetAppRegistryRaw struct {
	Contract *SubnetAppRegistry // Generic contract binding to access the raw methods on
}

// SubnetAppRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SubnetAppRegistryCallerRaw struct {
	Contract *SubnetAppRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// SubnetAppRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SubnetAppRegistryTransactorRaw struct {
	Contract *SubnetAppRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSubnetAppRegistry creates a new instance of SubnetAppRegistry, bound to a specific deployed contract.
func NewSubnetAppRegistry(address common.Address, backend bind.ContractBackend) (*SubnetAppRegistry, error) {
	contract, err := bindSubnetAppRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SubnetAppRegistry{SubnetAppRegistryCaller: SubnetAppRegistryCaller{contract: contract}, SubnetAppRegistryTransactor: SubnetAppRegistryTransactor{contract: contract}, SubnetAppRegistryFilterer: SubnetAppRegistryFilterer{contract: contract}}, nil
}

// NewSubnetAppRegistryCaller creates a new read-only instance of SubnetAppRegistry, bound to a specific deployed contract.
func NewSubnetAppRegistryCaller(address common.Address, caller bind.ContractCaller) (*SubnetAppRegistryCaller, error) {
	contract, err := bindSubnetAppRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SubnetAppRegistryCaller{contract: contract}, nil
}

// NewSubnetAppRegistryTransactor creates a new write-only instance of SubnetAppRegistry, bound to a specific deployed contract.
func NewSubnetAppRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*SubnetAppRegistryTransactor, error) {
	contract, err := bindSubnetAppRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SubnetAppRegistryTransactor{contract: contract}, nil
}

// NewSubnetAppRegistryFilterer creates a new log filterer instance of SubnetAppRegistry, bound to a specific deployed contract.
func NewSubnetAppRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*SubnetAppRegistryFilterer, error) {
	contract, err := bindSubnetAppRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SubnetAppRegistryFilterer{contract: contract}, nil
}

// bindSubnetAppRegistry binds a generic wrapper to an already deployed contract.
func bindSubnetAppRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SubnetAppRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SubnetAppRegistry *SubnetAppRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SubnetAppRegistry.Contract.SubnetAppRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SubnetAppRegistry *SubnetAppRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.SubnetAppRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SubnetAppRegistry *SubnetAppRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.SubnetAppRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SubnetAppRegistry *SubnetAppRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SubnetAppRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SubnetAppRegistry *SubnetAppRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SubnetAppRegistry *SubnetAppRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.contract.Transact(opts, method, params...)
}

// AppCount is a free data retrieval call binding the contract method 0xb55ca2c3.
//
// Solidity: function appCount() view returns(uint256)
func (_SubnetAppRegistry *SubnetAppRegistryCaller) AppCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SubnetAppRegistry.contract.Call(opts, &out, "appCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// AppCount is a free data retrieval call binding the contract method 0xb55ca2c3.
//
// Solidity: function appCount() view returns(uint256)
func (_SubnetAppRegistry *SubnetAppRegistrySession) AppCount() (*big.Int, error) {
	return _SubnetAppRegistry.Contract.AppCount(&_SubnetAppRegistry.CallOpts)
}

// AppCount is a free data retrieval call binding the contract method 0xb55ca2c3.
//
// Solidity: function appCount() view returns(uint256)
func (_SubnetAppRegistry *SubnetAppRegistryCallerSession) AppCount() (*big.Int, error) {
	return _SubnetAppRegistry.Contract.AppCount(&_SubnetAppRegistry.CallOpts)
}

// AppNodes is a free data retrieval call binding the contract method 0xd7f7faf4.
//
// Solidity: function appNodes(uint256 , uint256 ) view returns(uint256 duration, uint256 usedCpu, uint256 usedGpu, uint256 usedMemory, uint256 usedStorage, uint256 usedDownloadBytes, uint256 usedUploadBytes, bool isRegistered)
func (_SubnetAppRegistry *SubnetAppRegistryCaller) AppNodes(opts *bind.CallOpts, arg0 *big.Int, arg1 *big.Int) (struct {
	Duration          *big.Int
	UsedCpu           *big.Int
	UsedGpu           *big.Int
	UsedMemory        *big.Int
	UsedStorage       *big.Int
	UsedDownloadBytes *big.Int
	UsedUploadBytes   *big.Int
	IsRegistered      bool
}, error) {
	var out []interface{}
	err := _SubnetAppRegistry.contract.Call(opts, &out, "appNodes", arg0, arg1)

	outstruct := new(struct {
		Duration          *big.Int
		UsedCpu           *big.Int
		UsedGpu           *big.Int
		UsedMemory        *big.Int
		UsedStorage       *big.Int
		UsedDownloadBytes *big.Int
		UsedUploadBytes   *big.Int
		IsRegistered      bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Duration = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.UsedCpu = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.UsedGpu = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.UsedMemory = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.UsedStorage = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.UsedDownloadBytes = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)
	outstruct.UsedUploadBytes = *abi.ConvertType(out[6], new(*big.Int)).(**big.Int)
	outstruct.IsRegistered = *abi.ConvertType(out[7], new(bool)).(*bool)

	return *outstruct, err

}

// AppNodes is a free data retrieval call binding the contract method 0xd7f7faf4.
//
// Solidity: function appNodes(uint256 , uint256 ) view returns(uint256 duration, uint256 usedCpu, uint256 usedGpu, uint256 usedMemory, uint256 usedStorage, uint256 usedDownloadBytes, uint256 usedUploadBytes, bool isRegistered)
func (_SubnetAppRegistry *SubnetAppRegistrySession) AppNodes(arg0 *big.Int, arg1 *big.Int) (struct {
	Duration          *big.Int
	UsedCpu           *big.Int
	UsedGpu           *big.Int
	UsedMemory        *big.Int
	UsedStorage       *big.Int
	UsedDownloadBytes *big.Int
	UsedUploadBytes   *big.Int
	IsRegistered      bool
}, error) {
	return _SubnetAppRegistry.Contract.AppNodes(&_SubnetAppRegistry.CallOpts, arg0, arg1)
}

// AppNodes is a free data retrieval call binding the contract method 0xd7f7faf4.
//
// Solidity: function appNodes(uint256 , uint256 ) view returns(uint256 duration, uint256 usedCpu, uint256 usedGpu, uint256 usedMemory, uint256 usedStorage, uint256 usedDownloadBytes, uint256 usedUploadBytes, bool isRegistered)
func (_SubnetAppRegistry *SubnetAppRegistryCallerSession) AppNodes(arg0 *big.Int, arg1 *big.Int) (struct {
	Duration          *big.Int
	UsedCpu           *big.Int
	UsedGpu           *big.Int
	UsedMemory        *big.Int
	UsedStorage       *big.Int
	UsedDownloadBytes *big.Int
	UsedUploadBytes   *big.Int
	IsRegistered      bool
}, error) {
	return _SubnetAppRegistry.Contract.AppNodes(&_SubnetAppRegistry.CallOpts, arg0, arg1)
}

// Apps is a free data retrieval call binding the contract method 0x61acc37e.
//
// Solidity: function apps(uint256 ) view returns(string peerId, address owner, string name, string symbol, uint256 budget, uint256 spentBudget, uint256 maxNodes, uint256 minCpu, uint256 minGpu, uint256 minMemory, uint256 minUploadBandwidth, uint256 minDownloadBandwidth, uint256 nodeCount, uint256 pricePerCpu, uint256 pricePerGpu, uint256 pricePerMemoryGB, uint256 pricePerStorageGB, uint256 pricePerBandwidthGB, string metadata)
func (_SubnetAppRegistry *SubnetAppRegistryCaller) Apps(opts *bind.CallOpts, arg0 *big.Int) (struct {
	PeerId               string
	Owner                common.Address
	Name                 string
	Symbol               string
	Budget               *big.Int
	SpentBudget          *big.Int
	MaxNodes             *big.Int
	MinCpu               *big.Int
	MinGpu               *big.Int
	MinMemory            *big.Int
	MinUploadBandwidth   *big.Int
	MinDownloadBandwidth *big.Int
	NodeCount            *big.Int
	PricePerCpu          *big.Int
	PricePerGpu          *big.Int
	PricePerMemoryGB     *big.Int
	PricePerStorageGB    *big.Int
	PricePerBandwidthGB  *big.Int
	Metadata             string
}, error) {
	var out []interface{}
	err := _SubnetAppRegistry.contract.Call(opts, &out, "apps", arg0)

	outstruct := new(struct {
		PeerId               string
		Owner                common.Address
		Name                 string
		Symbol               string
		Budget               *big.Int
		SpentBudget          *big.Int
		MaxNodes             *big.Int
		MinCpu               *big.Int
		MinGpu               *big.Int
		MinMemory            *big.Int
		MinUploadBandwidth   *big.Int
		MinDownloadBandwidth *big.Int
		NodeCount            *big.Int
		PricePerCpu          *big.Int
		PricePerGpu          *big.Int
		PricePerMemoryGB     *big.Int
		PricePerStorageGB    *big.Int
		PricePerBandwidthGB  *big.Int
		Metadata             string
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.PeerId = *abi.ConvertType(out[0], new(string)).(*string)
	outstruct.Owner = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	outstruct.Name = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.Symbol = *abi.ConvertType(out[3], new(string)).(*string)
	outstruct.Budget = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.SpentBudget = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)
	outstruct.MaxNodes = *abi.ConvertType(out[6], new(*big.Int)).(**big.Int)
	outstruct.MinCpu = *abi.ConvertType(out[7], new(*big.Int)).(**big.Int)
	outstruct.MinGpu = *abi.ConvertType(out[8], new(*big.Int)).(**big.Int)
	outstruct.MinMemory = *abi.ConvertType(out[9], new(*big.Int)).(**big.Int)
	outstruct.MinUploadBandwidth = *abi.ConvertType(out[10], new(*big.Int)).(**big.Int)
	outstruct.MinDownloadBandwidth = *abi.ConvertType(out[11], new(*big.Int)).(**big.Int)
	outstruct.NodeCount = *abi.ConvertType(out[12], new(*big.Int)).(**big.Int)
	outstruct.PricePerCpu = *abi.ConvertType(out[13], new(*big.Int)).(**big.Int)
	outstruct.PricePerGpu = *abi.ConvertType(out[14], new(*big.Int)).(**big.Int)
	outstruct.PricePerMemoryGB = *abi.ConvertType(out[15], new(*big.Int)).(**big.Int)
	outstruct.PricePerStorageGB = *abi.ConvertType(out[16], new(*big.Int)).(**big.Int)
	outstruct.PricePerBandwidthGB = *abi.ConvertType(out[17], new(*big.Int)).(**big.Int)
	outstruct.Metadata = *abi.ConvertType(out[18], new(string)).(*string)

	return *outstruct, err

}

// Apps is a free data retrieval call binding the contract method 0x61acc37e.
//
// Solidity: function apps(uint256 ) view returns(string peerId, address owner, string name, string symbol, uint256 budget, uint256 spentBudget, uint256 maxNodes, uint256 minCpu, uint256 minGpu, uint256 minMemory, uint256 minUploadBandwidth, uint256 minDownloadBandwidth, uint256 nodeCount, uint256 pricePerCpu, uint256 pricePerGpu, uint256 pricePerMemoryGB, uint256 pricePerStorageGB, uint256 pricePerBandwidthGB, string metadata)
func (_SubnetAppRegistry *SubnetAppRegistrySession) Apps(arg0 *big.Int) (struct {
	PeerId               string
	Owner                common.Address
	Name                 string
	Symbol               string
	Budget               *big.Int
	SpentBudget          *big.Int
	MaxNodes             *big.Int
	MinCpu               *big.Int
	MinGpu               *big.Int
	MinMemory            *big.Int
	MinUploadBandwidth   *big.Int
	MinDownloadBandwidth *big.Int
	NodeCount            *big.Int
	PricePerCpu          *big.Int
	PricePerGpu          *big.Int
	PricePerMemoryGB     *big.Int
	PricePerStorageGB    *big.Int
	PricePerBandwidthGB  *big.Int
	Metadata             string
}, error) {
	return _SubnetAppRegistry.Contract.Apps(&_SubnetAppRegistry.CallOpts, arg0)
}

// Apps is a free data retrieval call binding the contract method 0x61acc37e.
//
// Solidity: function apps(uint256 ) view returns(string peerId, address owner, string name, string symbol, uint256 budget, uint256 spentBudget, uint256 maxNodes, uint256 minCpu, uint256 minGpu, uint256 minMemory, uint256 minUploadBandwidth, uint256 minDownloadBandwidth, uint256 nodeCount, uint256 pricePerCpu, uint256 pricePerGpu, uint256 pricePerMemoryGB, uint256 pricePerStorageGB, uint256 pricePerBandwidthGB, string metadata)
func (_SubnetAppRegistry *SubnetAppRegistryCallerSession) Apps(arg0 *big.Int) (struct {
	PeerId               string
	Owner                common.Address
	Name                 string
	Symbol               string
	Budget               *big.Int
	SpentBudget          *big.Int
	MaxNodes             *big.Int
	MinCpu               *big.Int
	MinGpu               *big.Int
	MinMemory            *big.Int
	MinUploadBandwidth   *big.Int
	MinDownloadBandwidth *big.Int
	NodeCount            *big.Int
	PricePerCpu          *big.Int
	PricePerGpu          *big.Int
	PricePerMemoryGB     *big.Int
	PricePerStorageGB    *big.Int
	PricePerBandwidthGB  *big.Int
	Metadata             string
}, error) {
	return _SubnetAppRegistry.Contract.Apps(&_SubnetAppRegistry.CallOpts, arg0)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_SubnetAppRegistry *SubnetAppRegistryCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _SubnetAppRegistry.contract.Call(opts, &out, "eip712Domain")

	outstruct := new(struct {
		Fields            [1]byte
		Name              string
		Version           string
		ChainId           *big.Int
		VerifyingContract common.Address
		Salt              [32]byte
		Extensions        []*big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Fields = *abi.ConvertType(out[0], new([1]byte)).(*[1]byte)
	outstruct.Name = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.Version = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.ChainId = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.VerifyingContract = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	outstruct.Salt = *abi.ConvertType(out[5], new([32]byte)).(*[32]byte)
	outstruct.Extensions = *abi.ConvertType(out[6], new([]*big.Int)).(*[]*big.Int)

	return *outstruct, err

}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_SubnetAppRegistry *SubnetAppRegistrySession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _SubnetAppRegistry.Contract.Eip712Domain(&_SubnetAppRegistry.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_SubnetAppRegistry *SubnetAppRegistryCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _SubnetAppRegistry.Contract.Eip712Domain(&_SubnetAppRegistry.CallOpts)
}

// FeeRate is a free data retrieval call binding the contract method 0x978bbdb9.
//
// Solidity: function feeRate() view returns(uint256)
func (_SubnetAppRegistry *SubnetAppRegistryCaller) FeeRate(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SubnetAppRegistry.contract.Call(opts, &out, "feeRate")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FeeRate is a free data retrieval call binding the contract method 0x978bbdb9.
//
// Solidity: function feeRate() view returns(uint256)
func (_SubnetAppRegistry *SubnetAppRegistrySession) FeeRate() (*big.Int, error) {
	return _SubnetAppRegistry.Contract.FeeRate(&_SubnetAppRegistry.CallOpts)
}

// FeeRate is a free data retrieval call binding the contract method 0x978bbdb9.
//
// Solidity: function feeRate() view returns(uint256)
func (_SubnetAppRegistry *SubnetAppRegistryCallerSession) FeeRate() (*big.Int, error) {
	return _SubnetAppRegistry.Contract.FeeRate(&_SubnetAppRegistry.CallOpts)
}

// GetApp is a free data retrieval call binding the contract method 0x24f3a51b.
//
// Solidity: function getApp(uint256 appId) view returns((string,address,string,string,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,string) app)
func (_SubnetAppRegistry *SubnetAppRegistryCaller) GetApp(opts *bind.CallOpts, appId *big.Int) (SubnetAppRegistryApp, error) {
	var out []interface{}
	err := _SubnetAppRegistry.contract.Call(opts, &out, "getApp", appId)

	if err != nil {
		return *new(SubnetAppRegistryApp), err
	}

	out0 := *abi.ConvertType(out[0], new(SubnetAppRegistryApp)).(*SubnetAppRegistryApp)

	return out0, err

}

// GetApp is a free data retrieval call binding the contract method 0x24f3a51b.
//
// Solidity: function getApp(uint256 appId) view returns((string,address,string,string,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,string) app)
func (_SubnetAppRegistry *SubnetAppRegistrySession) GetApp(appId *big.Int) (SubnetAppRegistryApp, error) {
	return _SubnetAppRegistry.Contract.GetApp(&_SubnetAppRegistry.CallOpts, appId)
}

// GetApp is a free data retrieval call binding the contract method 0x24f3a51b.
//
// Solidity: function getApp(uint256 appId) view returns((string,address,string,string,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,string) app)
func (_SubnetAppRegistry *SubnetAppRegistryCallerSession) GetApp(appId *big.Int) (SubnetAppRegistryApp, error) {
	return _SubnetAppRegistry.Contract.GetApp(&_SubnetAppRegistry.CallOpts, appId)
}

// GetAppNode is a free data retrieval call binding the contract method 0xe98edf26.
//
// Solidity: function getAppNode(uint256 appId, uint256 subnetId) view returns((uint256,uint256,uint256,uint256,uint256,uint256,uint256,bool) appNode)
func (_SubnetAppRegistry *SubnetAppRegistryCaller) GetAppNode(opts *bind.CallOpts, appId *big.Int, subnetId *big.Int) (SubnetAppRegistryAppNode, error) {
	var out []interface{}
	err := _SubnetAppRegistry.contract.Call(opts, &out, "getAppNode", appId, subnetId)

	if err != nil {
		return *new(SubnetAppRegistryAppNode), err
	}

	out0 := *abi.ConvertType(out[0], new(SubnetAppRegistryAppNode)).(*SubnetAppRegistryAppNode)

	return out0, err

}

// GetAppNode is a free data retrieval call binding the contract method 0xe98edf26.
//
// Solidity: function getAppNode(uint256 appId, uint256 subnetId) view returns((uint256,uint256,uint256,uint256,uint256,uint256,uint256,bool) appNode)
func (_SubnetAppRegistry *SubnetAppRegistrySession) GetAppNode(appId *big.Int, subnetId *big.Int) (SubnetAppRegistryAppNode, error) {
	return _SubnetAppRegistry.Contract.GetAppNode(&_SubnetAppRegistry.CallOpts, appId, subnetId)
}

// GetAppNode is a free data retrieval call binding the contract method 0xe98edf26.
//
// Solidity: function getAppNode(uint256 appId, uint256 subnetId) view returns((uint256,uint256,uint256,uint256,uint256,uint256,uint256,bool) appNode)
func (_SubnetAppRegistry *SubnetAppRegistryCallerSession) GetAppNode(appId *big.Int, subnetId *big.Int) (SubnetAppRegistryAppNode, error) {
	return _SubnetAppRegistry.Contract.GetAppNode(&_SubnetAppRegistry.CallOpts, appId, subnetId)
}

// ListApps is a free data retrieval call binding the contract method 0x3d205559.
//
// Solidity: function listApps(uint256 start, uint256 end) view returns((string,address,string,string,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,string)[])
func (_SubnetAppRegistry *SubnetAppRegistryCaller) ListApps(opts *bind.CallOpts, start *big.Int, end *big.Int) ([]SubnetAppRegistryApp, error) {
	var out []interface{}
	err := _SubnetAppRegistry.contract.Call(opts, &out, "listApps", start, end)

	if err != nil {
		return *new([]SubnetAppRegistryApp), err
	}

	out0 := *abi.ConvertType(out[0], new([]SubnetAppRegistryApp)).(*[]SubnetAppRegistryApp)

	return out0, err

}

// ListApps is a free data retrieval call binding the contract method 0x3d205559.
//
// Solidity: function listApps(uint256 start, uint256 end) view returns((string,address,string,string,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,string)[])
func (_SubnetAppRegistry *SubnetAppRegistrySession) ListApps(start *big.Int, end *big.Int) ([]SubnetAppRegistryApp, error) {
	return _SubnetAppRegistry.Contract.ListApps(&_SubnetAppRegistry.CallOpts, start, end)
}

// ListApps is a free data retrieval call binding the contract method 0x3d205559.
//
// Solidity: function listApps(uint256 start, uint256 end) view returns((string,address,string,string,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,string)[])
func (_SubnetAppRegistry *SubnetAppRegistryCallerSession) ListApps(start *big.Int, end *big.Int) ([]SubnetAppRegistryApp, error) {
	return _SubnetAppRegistry.Contract.ListApps(&_SubnetAppRegistry.CallOpts, start, end)
}

// NodeToAppId is a free data retrieval call binding the contract method 0xc4d11ccb.
//
// Solidity: function nodeToAppId(uint256 ) view returns(uint256)
func (_SubnetAppRegistry *SubnetAppRegistryCaller) NodeToAppId(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SubnetAppRegistry.contract.Call(opts, &out, "nodeToAppId", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NodeToAppId is a free data retrieval call binding the contract method 0xc4d11ccb.
//
// Solidity: function nodeToAppId(uint256 ) view returns(uint256)
func (_SubnetAppRegistry *SubnetAppRegistrySession) NodeToAppId(arg0 *big.Int) (*big.Int, error) {
	return _SubnetAppRegistry.Contract.NodeToAppId(&_SubnetAppRegistry.CallOpts, arg0)
}

// NodeToAppId is a free data retrieval call binding the contract method 0xc4d11ccb.
//
// Solidity: function nodeToAppId(uint256 ) view returns(uint256)
func (_SubnetAppRegistry *SubnetAppRegistryCallerSession) NodeToAppId(arg0 *big.Int) (*big.Int, error) {
	return _SubnetAppRegistry.Contract.NodeToAppId(&_SubnetAppRegistry.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetAppRegistry *SubnetAppRegistryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetAppRegistry.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetAppRegistry *SubnetAppRegistrySession) Owner() (common.Address, error) {
	return _SubnetAppRegistry.Contract.Owner(&_SubnetAppRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetAppRegistry *SubnetAppRegistryCallerSession) Owner() (common.Address, error) {
	return _SubnetAppRegistry.Contract.Owner(&_SubnetAppRegistry.CallOpts)
}

// SubnetRegistry is a free data retrieval call binding the contract method 0x3e0beca2.
//
// Solidity: function subnetRegistry() view returns(address)
func (_SubnetAppRegistry *SubnetAppRegistryCaller) SubnetRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetAppRegistry.contract.Call(opts, &out, "subnetRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SubnetRegistry is a free data retrieval call binding the contract method 0x3e0beca2.
//
// Solidity: function subnetRegistry() view returns(address)
func (_SubnetAppRegistry *SubnetAppRegistrySession) SubnetRegistry() (common.Address, error) {
	return _SubnetAppRegistry.Contract.SubnetRegistry(&_SubnetAppRegistry.CallOpts)
}

// SubnetRegistry is a free data retrieval call binding the contract method 0x3e0beca2.
//
// Solidity: function subnetRegistry() view returns(address)
func (_SubnetAppRegistry *SubnetAppRegistryCallerSession) SubnetRegistry() (common.Address, error) {
	return _SubnetAppRegistry.Contract.SubnetRegistry(&_SubnetAppRegistry.CallOpts)
}

// SymbolToAppId is a free data retrieval call binding the contract method 0xcb4b2360.
//
// Solidity: function symbolToAppId(string ) view returns(uint256)
func (_SubnetAppRegistry *SubnetAppRegistryCaller) SymbolToAppId(opts *bind.CallOpts, arg0 string) (*big.Int, error) {
	var out []interface{}
	err := _SubnetAppRegistry.contract.Call(opts, &out, "symbolToAppId", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SymbolToAppId is a free data retrieval call binding the contract method 0xcb4b2360.
//
// Solidity: function symbolToAppId(string ) view returns(uint256)
func (_SubnetAppRegistry *SubnetAppRegistrySession) SymbolToAppId(arg0 string) (*big.Int, error) {
	return _SubnetAppRegistry.Contract.SymbolToAppId(&_SubnetAppRegistry.CallOpts, arg0)
}

// SymbolToAppId is a free data retrieval call binding the contract method 0xcb4b2360.
//
// Solidity: function symbolToAppId(string ) view returns(uint256)
func (_SubnetAppRegistry *SubnetAppRegistryCallerSession) SymbolToAppId(arg0 string) (*big.Int, error) {
	return _SubnetAppRegistry.Contract.SymbolToAppId(&_SubnetAppRegistry.CallOpts, arg0)
}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_SubnetAppRegistry *SubnetAppRegistryCaller) Treasury(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetAppRegistry.contract.Call(opts, &out, "treasury")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_SubnetAppRegistry *SubnetAppRegistrySession) Treasury() (common.Address, error) {
	return _SubnetAppRegistry.Contract.Treasury(&_SubnetAppRegistry.CallOpts)
}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_SubnetAppRegistry *SubnetAppRegistryCallerSession) Treasury() (common.Address, error) {
	return _SubnetAppRegistry.Contract.Treasury(&_SubnetAppRegistry.CallOpts)
}

// UsedMessageHashes is a free data retrieval call binding the contract method 0x87d69207.
//
// Solidity: function usedMessageHashes(bytes32 ) view returns(bool)
func (_SubnetAppRegistry *SubnetAppRegistryCaller) UsedMessageHashes(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var out []interface{}
	err := _SubnetAppRegistry.contract.Call(opts, &out, "usedMessageHashes", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// UsedMessageHashes is a free data retrieval call binding the contract method 0x87d69207.
//
// Solidity: function usedMessageHashes(bytes32 ) view returns(bool)
func (_SubnetAppRegistry *SubnetAppRegistrySession) UsedMessageHashes(arg0 [32]byte) (bool, error) {
	return _SubnetAppRegistry.Contract.UsedMessageHashes(&_SubnetAppRegistry.CallOpts, arg0)
}

// UsedMessageHashes is a free data retrieval call binding the contract method 0x87d69207.
//
// Solidity: function usedMessageHashes(bytes32 ) view returns(bool)
func (_SubnetAppRegistry *SubnetAppRegistryCallerSession) UsedMessageHashes(arg0 [32]byte) (bool, error) {
	return _SubnetAppRegistry.Contract.UsedMessageHashes(&_SubnetAppRegistry.CallOpts, arg0)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(string)
func (_SubnetAppRegistry *SubnetAppRegistryCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _SubnetAppRegistry.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(string)
func (_SubnetAppRegistry *SubnetAppRegistrySession) Version() (string, error) {
	return _SubnetAppRegistry.Contract.Version(&_SubnetAppRegistry.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(string)
func (_SubnetAppRegistry *SubnetAppRegistryCallerSession) Version() (string, error) {
	return _SubnetAppRegistry.Contract.Version(&_SubnetAppRegistry.CallOpts)
}

// ClaimReward is a paid mutator transaction binding the contract method 0x1bdba780.
//
// Solidity: function claimReward(uint256 subnetId, uint256 appId, uint256 usedCpu, uint256 usedGpu, uint256 usedMemory, uint256 usedStorage, uint256 usedUploadBytes, uint256 usedDownloadBytes, uint256 duration, bytes signature) returns()
func (_SubnetAppRegistry *SubnetAppRegistryTransactor) ClaimReward(opts *bind.TransactOpts, subnetId *big.Int, appId *big.Int, usedCpu *big.Int, usedGpu *big.Int, usedMemory *big.Int, usedStorage *big.Int, usedUploadBytes *big.Int, usedDownloadBytes *big.Int, duration *big.Int, signature []byte) (*types.Transaction, error) {
	return _SubnetAppRegistry.contract.Transact(opts, "claimReward", subnetId, appId, usedCpu, usedGpu, usedMemory, usedStorage, usedUploadBytes, usedDownloadBytes, duration, signature)
}

// ClaimReward is a paid mutator transaction binding the contract method 0x1bdba780.
//
// Solidity: function claimReward(uint256 subnetId, uint256 appId, uint256 usedCpu, uint256 usedGpu, uint256 usedMemory, uint256 usedStorage, uint256 usedUploadBytes, uint256 usedDownloadBytes, uint256 duration, bytes signature) returns()
func (_SubnetAppRegistry *SubnetAppRegistrySession) ClaimReward(subnetId *big.Int, appId *big.Int, usedCpu *big.Int, usedGpu *big.Int, usedMemory *big.Int, usedStorage *big.Int, usedUploadBytes *big.Int, usedDownloadBytes *big.Int, duration *big.Int, signature []byte) (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.ClaimReward(&_SubnetAppRegistry.TransactOpts, subnetId, appId, usedCpu, usedGpu, usedMemory, usedStorage, usedUploadBytes, usedDownloadBytes, duration, signature)
}

// ClaimReward is a paid mutator transaction binding the contract method 0x1bdba780.
//
// Solidity: function claimReward(uint256 subnetId, uint256 appId, uint256 usedCpu, uint256 usedGpu, uint256 usedMemory, uint256 usedStorage, uint256 usedUploadBytes, uint256 usedDownloadBytes, uint256 duration, bytes signature) returns()
func (_SubnetAppRegistry *SubnetAppRegistryTransactorSession) ClaimReward(subnetId *big.Int, appId *big.Int, usedCpu *big.Int, usedGpu *big.Int, usedMemory *big.Int, usedStorage *big.Int, usedUploadBytes *big.Int, usedDownloadBytes *big.Int, duration *big.Int, signature []byte) (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.ClaimReward(&_SubnetAppRegistry.TransactOpts, subnetId, appId, usedCpu, usedGpu, usedMemory, usedStorage, usedUploadBytes, usedDownloadBytes, duration, signature)
}

// CreateApp is a paid mutator transaction binding the contract method 0xedce259d.
//
// Solidity: function createApp(string name, string symbol, string peerId, uint256 budget, uint256 maxNodes, uint256 minCpu, uint256 minGpu, uint256 minMemory, uint256 minUploadBandwidth, uint256 minDownloadBandwidth, uint256 pricePerCpu, uint256 pricePerGpu, uint256 pricePerMemoryGB, uint256 pricePerStorageGB, uint256 pricePerBandwidthGB, string metadata) payable returns()
func (_SubnetAppRegistry *SubnetAppRegistryTransactor) CreateApp(opts *bind.TransactOpts, name string, symbol string, peerId string, budget *big.Int, maxNodes *big.Int, minCpu *big.Int, minGpu *big.Int, minMemory *big.Int, minUploadBandwidth *big.Int, minDownloadBandwidth *big.Int, pricePerCpu *big.Int, pricePerGpu *big.Int, pricePerMemoryGB *big.Int, pricePerStorageGB *big.Int, pricePerBandwidthGB *big.Int, metadata string) (*types.Transaction, error) {
	return _SubnetAppRegistry.contract.Transact(opts, "createApp", name, symbol, peerId, budget, maxNodes, minCpu, minGpu, minMemory, minUploadBandwidth, minDownloadBandwidth, pricePerCpu, pricePerGpu, pricePerMemoryGB, pricePerStorageGB, pricePerBandwidthGB, metadata)
}

// CreateApp is a paid mutator transaction binding the contract method 0xedce259d.
//
// Solidity: function createApp(string name, string symbol, string peerId, uint256 budget, uint256 maxNodes, uint256 minCpu, uint256 minGpu, uint256 minMemory, uint256 minUploadBandwidth, uint256 minDownloadBandwidth, uint256 pricePerCpu, uint256 pricePerGpu, uint256 pricePerMemoryGB, uint256 pricePerStorageGB, uint256 pricePerBandwidthGB, string metadata) payable returns()
func (_SubnetAppRegistry *SubnetAppRegistrySession) CreateApp(name string, symbol string, peerId string, budget *big.Int, maxNodes *big.Int, minCpu *big.Int, minGpu *big.Int, minMemory *big.Int, minUploadBandwidth *big.Int, minDownloadBandwidth *big.Int, pricePerCpu *big.Int, pricePerGpu *big.Int, pricePerMemoryGB *big.Int, pricePerStorageGB *big.Int, pricePerBandwidthGB *big.Int, metadata string) (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.CreateApp(&_SubnetAppRegistry.TransactOpts, name, symbol, peerId, budget, maxNodes, minCpu, minGpu, minMemory, minUploadBandwidth, minDownloadBandwidth, pricePerCpu, pricePerGpu, pricePerMemoryGB, pricePerStorageGB, pricePerBandwidthGB, metadata)
}

// CreateApp is a paid mutator transaction binding the contract method 0xedce259d.
//
// Solidity: function createApp(string name, string symbol, string peerId, uint256 budget, uint256 maxNodes, uint256 minCpu, uint256 minGpu, uint256 minMemory, uint256 minUploadBandwidth, uint256 minDownloadBandwidth, uint256 pricePerCpu, uint256 pricePerGpu, uint256 pricePerMemoryGB, uint256 pricePerStorageGB, uint256 pricePerBandwidthGB, string metadata) payable returns()
func (_SubnetAppRegistry *SubnetAppRegistryTransactorSession) CreateApp(name string, symbol string, peerId string, budget *big.Int, maxNodes *big.Int, minCpu *big.Int, minGpu *big.Int, minMemory *big.Int, minUploadBandwidth *big.Int, minDownloadBandwidth *big.Int, pricePerCpu *big.Int, pricePerGpu *big.Int, pricePerMemoryGB *big.Int, pricePerStorageGB *big.Int, pricePerBandwidthGB *big.Int, metadata string) (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.CreateApp(&_SubnetAppRegistry.TransactOpts, name, symbol, peerId, budget, maxNodes, minCpu, minGpu, minMemory, minUploadBandwidth, minDownloadBandwidth, pricePerCpu, pricePerGpu, pricePerMemoryGB, pricePerStorageGB, pricePerBandwidthGB, metadata)
}

// RegisterNode is a paid mutator transaction binding the contract method 0x236a8d9d.
//
// Solidity: function registerNode(uint256 subnetId, uint256 appId) returns()
func (_SubnetAppRegistry *SubnetAppRegistryTransactor) RegisterNode(opts *bind.TransactOpts, subnetId *big.Int, appId *big.Int) (*types.Transaction, error) {
	return _SubnetAppRegistry.contract.Transact(opts, "registerNode", subnetId, appId)
}

// RegisterNode is a paid mutator transaction binding the contract method 0x236a8d9d.
//
// Solidity: function registerNode(uint256 subnetId, uint256 appId) returns()
func (_SubnetAppRegistry *SubnetAppRegistrySession) RegisterNode(subnetId *big.Int, appId *big.Int) (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.RegisterNode(&_SubnetAppRegistry.TransactOpts, subnetId, appId)
}

// RegisterNode is a paid mutator transaction binding the contract method 0x236a8d9d.
//
// Solidity: function registerNode(uint256 subnetId, uint256 appId) returns()
func (_SubnetAppRegistry *SubnetAppRegistryTransactorSession) RegisterNode(subnetId *big.Int, appId *big.Int) (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.RegisterNode(&_SubnetAppRegistry.TransactOpts, subnetId, appId)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetAppRegistry *SubnetAppRegistryTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetAppRegistry.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetAppRegistry *SubnetAppRegistrySession) RenounceOwnership() (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.RenounceOwnership(&_SubnetAppRegistry.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetAppRegistry *SubnetAppRegistryTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.RenounceOwnership(&_SubnetAppRegistry.TransactOpts)
}

// SetFeeRate is a paid mutator transaction binding the contract method 0x45596e2e.
//
// Solidity: function setFeeRate(uint256 _feeRate) returns()
func (_SubnetAppRegistry *SubnetAppRegistryTransactor) SetFeeRate(opts *bind.TransactOpts, _feeRate *big.Int) (*types.Transaction, error) {
	return _SubnetAppRegistry.contract.Transact(opts, "setFeeRate", _feeRate)
}

// SetFeeRate is a paid mutator transaction binding the contract method 0x45596e2e.
//
// Solidity: function setFeeRate(uint256 _feeRate) returns()
func (_SubnetAppRegistry *SubnetAppRegistrySession) SetFeeRate(_feeRate *big.Int) (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.SetFeeRate(&_SubnetAppRegistry.TransactOpts, _feeRate)
}

// SetFeeRate is a paid mutator transaction binding the contract method 0x45596e2e.
//
// Solidity: function setFeeRate(uint256 _feeRate) returns()
func (_SubnetAppRegistry *SubnetAppRegistryTransactorSession) SetFeeRate(_feeRate *big.Int) (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.SetFeeRate(&_SubnetAppRegistry.TransactOpts, _feeRate)
}

// SetTreasury is a paid mutator transaction binding the contract method 0xf0f44260.
//
// Solidity: function setTreasury(address _treasury) returns()
func (_SubnetAppRegistry *SubnetAppRegistryTransactor) SetTreasury(opts *bind.TransactOpts, _treasury common.Address) (*types.Transaction, error) {
	return _SubnetAppRegistry.contract.Transact(opts, "setTreasury", _treasury)
}

// SetTreasury is a paid mutator transaction binding the contract method 0xf0f44260.
//
// Solidity: function setTreasury(address _treasury) returns()
func (_SubnetAppRegistry *SubnetAppRegistrySession) SetTreasury(_treasury common.Address) (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.SetTreasury(&_SubnetAppRegistry.TransactOpts, _treasury)
}

// SetTreasury is a paid mutator transaction binding the contract method 0xf0f44260.
//
// Solidity: function setTreasury(address _treasury) returns()
func (_SubnetAppRegistry *SubnetAppRegistryTransactorSession) SetTreasury(_treasury common.Address) (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.SetTreasury(&_SubnetAppRegistry.TransactOpts, _treasury)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetAppRegistry *SubnetAppRegistryTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _SubnetAppRegistry.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetAppRegistry *SubnetAppRegistrySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.TransferOwnership(&_SubnetAppRegistry.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetAppRegistry *SubnetAppRegistryTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.TransferOwnership(&_SubnetAppRegistry.TransactOpts, newOwner)
}

// UpdateApp is a paid mutator transaction binding the contract method 0xc1fe593c.
//
// Solidity: function updateApp(uint256 appId, string name, string peerId, uint256 maxNodes, uint256 minCpu, uint256 minGpu, uint256 minMemory, uint256 minUploadBandwidth, uint256 minDownloadBandwidth, uint256 pricePerCpu, uint256 pricePerGpu, uint256 pricePerMemoryGB, uint256 pricePerStorageGB, uint256 pricePerBandwidthGB, string metadata) returns()
func (_SubnetAppRegistry *SubnetAppRegistryTransactor) UpdateApp(opts *bind.TransactOpts, appId *big.Int, name string, peerId string, maxNodes *big.Int, minCpu *big.Int, minGpu *big.Int, minMemory *big.Int, minUploadBandwidth *big.Int, minDownloadBandwidth *big.Int, pricePerCpu *big.Int, pricePerGpu *big.Int, pricePerMemoryGB *big.Int, pricePerStorageGB *big.Int, pricePerBandwidthGB *big.Int, metadata string) (*types.Transaction, error) {
	return _SubnetAppRegistry.contract.Transact(opts, "updateApp", appId, name, peerId, maxNodes, minCpu, minGpu, minMemory, minUploadBandwidth, minDownloadBandwidth, pricePerCpu, pricePerGpu, pricePerMemoryGB, pricePerStorageGB, pricePerBandwidthGB, metadata)
}

// UpdateApp is a paid mutator transaction binding the contract method 0xc1fe593c.
//
// Solidity: function updateApp(uint256 appId, string name, string peerId, uint256 maxNodes, uint256 minCpu, uint256 minGpu, uint256 minMemory, uint256 minUploadBandwidth, uint256 minDownloadBandwidth, uint256 pricePerCpu, uint256 pricePerGpu, uint256 pricePerMemoryGB, uint256 pricePerStorageGB, uint256 pricePerBandwidthGB, string metadata) returns()
func (_SubnetAppRegistry *SubnetAppRegistrySession) UpdateApp(appId *big.Int, name string, peerId string, maxNodes *big.Int, minCpu *big.Int, minGpu *big.Int, minMemory *big.Int, minUploadBandwidth *big.Int, minDownloadBandwidth *big.Int, pricePerCpu *big.Int, pricePerGpu *big.Int, pricePerMemoryGB *big.Int, pricePerStorageGB *big.Int, pricePerBandwidthGB *big.Int, metadata string) (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.UpdateApp(&_SubnetAppRegistry.TransactOpts, appId, name, peerId, maxNodes, minCpu, minGpu, minMemory, minUploadBandwidth, minDownloadBandwidth, pricePerCpu, pricePerGpu, pricePerMemoryGB, pricePerStorageGB, pricePerBandwidthGB, metadata)
}

// UpdateApp is a paid mutator transaction binding the contract method 0xc1fe593c.
//
// Solidity: function updateApp(uint256 appId, string name, string peerId, uint256 maxNodes, uint256 minCpu, uint256 minGpu, uint256 minMemory, uint256 minUploadBandwidth, uint256 minDownloadBandwidth, uint256 pricePerCpu, uint256 pricePerGpu, uint256 pricePerMemoryGB, uint256 pricePerStorageGB, uint256 pricePerBandwidthGB, string metadata) returns()
func (_SubnetAppRegistry *SubnetAppRegistryTransactorSession) UpdateApp(appId *big.Int, name string, peerId string, maxNodes *big.Int, minCpu *big.Int, minGpu *big.Int, minMemory *big.Int, minUploadBandwidth *big.Int, minDownloadBandwidth *big.Int, pricePerCpu *big.Int, pricePerGpu *big.Int, pricePerMemoryGB *big.Int, pricePerStorageGB *big.Int, pricePerBandwidthGB *big.Int, metadata string) (*types.Transaction, error) {
	return _SubnetAppRegistry.Contract.UpdateApp(&_SubnetAppRegistry.TransactOpts, appId, name, peerId, maxNodes, minCpu, minGpu, minMemory, minUploadBandwidth, minDownloadBandwidth, pricePerCpu, pricePerGpu, pricePerMemoryGB, pricePerStorageGB, pricePerBandwidthGB, metadata)
}

// SubnetAppRegistryAppCreatedIterator is returned from FilterAppCreated and is used to iterate over the raw logs and unpacked data for AppCreated events raised by the SubnetAppRegistry contract.
type SubnetAppRegistryAppCreatedIterator struct {
	Event *SubnetAppRegistryAppCreated // Event containing the contract specifics and raw log

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
func (it *SubnetAppRegistryAppCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetAppRegistryAppCreated)
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
		it.Event = new(SubnetAppRegistryAppCreated)
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
func (it *SubnetAppRegistryAppCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetAppRegistryAppCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetAppRegistryAppCreated represents a AppCreated event raised by the SubnetAppRegistry contract.
type SubnetAppRegistryAppCreated struct {
	AppId  *big.Int
	Name   string
	Symbol string
	Owner  common.Address
	Budget *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterAppCreated is a free log retrieval operation binding the contract event 0x9b96599010d3232696eb9dcbd3fa0a37ce73e2ffccd85f705a63a90afc65e420.
//
// Solidity: event AppCreated(uint256 indexed appId, string name, string symbol, address indexed owner, uint256 budget)
func (_SubnetAppRegistry *SubnetAppRegistryFilterer) FilterAppCreated(opts *bind.FilterOpts, appId []*big.Int, owner []common.Address) (*SubnetAppRegistryAppCreatedIterator, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _SubnetAppRegistry.contract.FilterLogs(opts, "AppCreated", appIdRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &SubnetAppRegistryAppCreatedIterator{contract: _SubnetAppRegistry.contract, event: "AppCreated", logs: logs, sub: sub}, nil
}

// WatchAppCreated is a free log subscription operation binding the contract event 0x9b96599010d3232696eb9dcbd3fa0a37ce73e2ffccd85f705a63a90afc65e420.
//
// Solidity: event AppCreated(uint256 indexed appId, string name, string symbol, address indexed owner, uint256 budget)
func (_SubnetAppRegistry *SubnetAppRegistryFilterer) WatchAppCreated(opts *bind.WatchOpts, sink chan<- *SubnetAppRegistryAppCreated, appId []*big.Int, owner []common.Address) (event.Subscription, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _SubnetAppRegistry.contract.WatchLogs(opts, "AppCreated", appIdRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetAppRegistryAppCreated)
				if err := _SubnetAppRegistry.contract.UnpackLog(event, "AppCreated", log); err != nil {
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

// ParseAppCreated is a log parse operation binding the contract event 0x9b96599010d3232696eb9dcbd3fa0a37ce73e2ffccd85f705a63a90afc65e420.
//
// Solidity: event AppCreated(uint256 indexed appId, string name, string symbol, address indexed owner, uint256 budget)
func (_SubnetAppRegistry *SubnetAppRegistryFilterer) ParseAppCreated(log types.Log) (*SubnetAppRegistryAppCreated, error) {
	event := new(SubnetAppRegistryAppCreated)
	if err := _SubnetAppRegistry.contract.UnpackLog(event, "AppCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetAppRegistryAppUpdatedIterator is returned from FilterAppUpdated and is used to iterate over the raw logs and unpacked data for AppUpdated events raised by the SubnetAppRegistry contract.
type SubnetAppRegistryAppUpdatedIterator struct {
	Event *SubnetAppRegistryAppUpdated // Event containing the contract specifics and raw log

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
func (it *SubnetAppRegistryAppUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetAppRegistryAppUpdated)
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
		it.Event = new(SubnetAppRegistryAppUpdated)
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
func (it *SubnetAppRegistryAppUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetAppRegistryAppUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetAppRegistryAppUpdated represents a AppUpdated event raised by the SubnetAppRegistry contract.
type SubnetAppRegistryAppUpdated struct {
	AppId *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterAppUpdated is a free log retrieval operation binding the contract event 0x5f9e31e4fc57fabac502d848e3a5d8ee121194d5ed670c12476c4fe260924157.
//
// Solidity: event AppUpdated(uint256 appId)
func (_SubnetAppRegistry *SubnetAppRegistryFilterer) FilterAppUpdated(opts *bind.FilterOpts) (*SubnetAppRegistryAppUpdatedIterator, error) {

	logs, sub, err := _SubnetAppRegistry.contract.FilterLogs(opts, "AppUpdated")
	if err != nil {
		return nil, err
	}
	return &SubnetAppRegistryAppUpdatedIterator{contract: _SubnetAppRegistry.contract, event: "AppUpdated", logs: logs, sub: sub}, nil
}

// WatchAppUpdated is a free log subscription operation binding the contract event 0x5f9e31e4fc57fabac502d848e3a5d8ee121194d5ed670c12476c4fe260924157.
//
// Solidity: event AppUpdated(uint256 appId)
func (_SubnetAppRegistry *SubnetAppRegistryFilterer) WatchAppUpdated(opts *bind.WatchOpts, sink chan<- *SubnetAppRegistryAppUpdated) (event.Subscription, error) {

	logs, sub, err := _SubnetAppRegistry.contract.WatchLogs(opts, "AppUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetAppRegistryAppUpdated)
				if err := _SubnetAppRegistry.contract.UnpackLog(event, "AppUpdated", log); err != nil {
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

// ParseAppUpdated is a log parse operation binding the contract event 0x5f9e31e4fc57fabac502d848e3a5d8ee121194d5ed670c12476c4fe260924157.
//
// Solidity: event AppUpdated(uint256 appId)
func (_SubnetAppRegistry *SubnetAppRegistryFilterer) ParseAppUpdated(log types.Log) (*SubnetAppRegistryAppUpdated, error) {
	event := new(SubnetAppRegistryAppUpdated)
	if err := _SubnetAppRegistry.contract.UnpackLog(event, "AppUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetAppRegistryEIP712DomainChangedIterator is returned from FilterEIP712DomainChanged and is used to iterate over the raw logs and unpacked data for EIP712DomainChanged events raised by the SubnetAppRegistry contract.
type SubnetAppRegistryEIP712DomainChangedIterator struct {
	Event *SubnetAppRegistryEIP712DomainChanged // Event containing the contract specifics and raw log

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
func (it *SubnetAppRegistryEIP712DomainChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetAppRegistryEIP712DomainChanged)
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
		it.Event = new(SubnetAppRegistryEIP712DomainChanged)
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
func (it *SubnetAppRegistryEIP712DomainChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetAppRegistryEIP712DomainChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetAppRegistryEIP712DomainChanged represents a EIP712DomainChanged event raised by the SubnetAppRegistry contract.
type SubnetAppRegistryEIP712DomainChanged struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEIP712DomainChanged is a free log retrieval operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_SubnetAppRegistry *SubnetAppRegistryFilterer) FilterEIP712DomainChanged(opts *bind.FilterOpts) (*SubnetAppRegistryEIP712DomainChangedIterator, error) {

	logs, sub, err := _SubnetAppRegistry.contract.FilterLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return &SubnetAppRegistryEIP712DomainChangedIterator{contract: _SubnetAppRegistry.contract, event: "EIP712DomainChanged", logs: logs, sub: sub}, nil
}

// WatchEIP712DomainChanged is a free log subscription operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_SubnetAppRegistry *SubnetAppRegistryFilterer) WatchEIP712DomainChanged(opts *bind.WatchOpts, sink chan<- *SubnetAppRegistryEIP712DomainChanged) (event.Subscription, error) {

	logs, sub, err := _SubnetAppRegistry.contract.WatchLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetAppRegistryEIP712DomainChanged)
				if err := _SubnetAppRegistry.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
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

// ParseEIP712DomainChanged is a log parse operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_SubnetAppRegistry *SubnetAppRegistryFilterer) ParseEIP712DomainChanged(log types.Log) (*SubnetAppRegistryEIP712DomainChanged, error) {
	event := new(SubnetAppRegistryEIP712DomainChanged)
	if err := _SubnetAppRegistry.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetAppRegistryNodeRegisteredIterator is returned from FilterNodeRegistered and is used to iterate over the raw logs and unpacked data for NodeRegistered events raised by the SubnetAppRegistry contract.
type SubnetAppRegistryNodeRegisteredIterator struct {
	Event *SubnetAppRegistryNodeRegistered // Event containing the contract specifics and raw log

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
func (it *SubnetAppRegistryNodeRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetAppRegistryNodeRegistered)
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
		it.Event = new(SubnetAppRegistryNodeRegistered)
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
func (it *SubnetAppRegistryNodeRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetAppRegistryNodeRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetAppRegistryNodeRegistered represents a NodeRegistered event raised by the SubnetAppRegistry contract.
type SubnetAppRegistryNodeRegistered struct {
	SubnetId *big.Int
	AppId    *big.Int
	Owner    common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterNodeRegistered is a free log retrieval operation binding the contract event 0x7f88516c140a892db548417fd31a878be51fb78bcf8cff5ea47514b13445103a.
//
// Solidity: event NodeRegistered(uint256 indexed subnetId, uint256 indexed appId, address indexed owner)
func (_SubnetAppRegistry *SubnetAppRegistryFilterer) FilterNodeRegistered(opts *bind.FilterOpts, subnetId []*big.Int, appId []*big.Int, owner []common.Address) (*SubnetAppRegistryNodeRegisteredIterator, error) {

	var subnetIdRule []interface{}
	for _, subnetIdItem := range subnetId {
		subnetIdRule = append(subnetIdRule, subnetIdItem)
	}
	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _SubnetAppRegistry.contract.FilterLogs(opts, "NodeRegistered", subnetIdRule, appIdRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &SubnetAppRegistryNodeRegisteredIterator{contract: _SubnetAppRegistry.contract, event: "NodeRegistered", logs: logs, sub: sub}, nil
}

// WatchNodeRegistered is a free log subscription operation binding the contract event 0x7f88516c140a892db548417fd31a878be51fb78bcf8cff5ea47514b13445103a.
//
// Solidity: event NodeRegistered(uint256 indexed subnetId, uint256 indexed appId, address indexed owner)
func (_SubnetAppRegistry *SubnetAppRegistryFilterer) WatchNodeRegistered(opts *bind.WatchOpts, sink chan<- *SubnetAppRegistryNodeRegistered, subnetId []*big.Int, appId []*big.Int, owner []common.Address) (event.Subscription, error) {

	var subnetIdRule []interface{}
	for _, subnetIdItem := range subnetId {
		subnetIdRule = append(subnetIdRule, subnetIdItem)
	}
	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _SubnetAppRegistry.contract.WatchLogs(opts, "NodeRegistered", subnetIdRule, appIdRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetAppRegistryNodeRegistered)
				if err := _SubnetAppRegistry.contract.UnpackLog(event, "NodeRegistered", log); err != nil {
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

// ParseNodeRegistered is a log parse operation binding the contract event 0x7f88516c140a892db548417fd31a878be51fb78bcf8cff5ea47514b13445103a.
//
// Solidity: event NodeRegistered(uint256 indexed subnetId, uint256 indexed appId, address indexed owner)
func (_SubnetAppRegistry *SubnetAppRegistryFilterer) ParseNodeRegistered(log types.Log) (*SubnetAppRegistryNodeRegistered, error) {
	event := new(SubnetAppRegistryNodeRegistered)
	if err := _SubnetAppRegistry.contract.UnpackLog(event, "NodeRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetAppRegistryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the SubnetAppRegistry contract.
type SubnetAppRegistryOwnershipTransferredIterator struct {
	Event *SubnetAppRegistryOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *SubnetAppRegistryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetAppRegistryOwnershipTransferred)
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
		it.Event = new(SubnetAppRegistryOwnershipTransferred)
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
func (it *SubnetAppRegistryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetAppRegistryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetAppRegistryOwnershipTransferred represents a OwnershipTransferred event raised by the SubnetAppRegistry contract.
type SubnetAppRegistryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SubnetAppRegistry *SubnetAppRegistryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*SubnetAppRegistryOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SubnetAppRegistry.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &SubnetAppRegistryOwnershipTransferredIterator{contract: _SubnetAppRegistry.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SubnetAppRegistry *SubnetAppRegistryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *SubnetAppRegistryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SubnetAppRegistry.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetAppRegistryOwnershipTransferred)
				if err := _SubnetAppRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_SubnetAppRegistry *SubnetAppRegistryFilterer) ParseOwnershipTransferred(log types.Log) (*SubnetAppRegistryOwnershipTransferred, error) {
	event := new(SubnetAppRegistryOwnershipTransferred)
	if err := _SubnetAppRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetAppRegistryRewardClaimedIterator is returned from FilterRewardClaimed and is used to iterate over the raw logs and unpacked data for RewardClaimed events raised by the SubnetAppRegistry contract.
type SubnetAppRegistryRewardClaimedIterator struct {
	Event *SubnetAppRegistryRewardClaimed // Event containing the contract specifics and raw log

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
func (it *SubnetAppRegistryRewardClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetAppRegistryRewardClaimed)
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
		it.Event = new(SubnetAppRegistryRewardClaimed)
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
func (it *SubnetAppRegistryRewardClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetAppRegistryRewardClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetAppRegistryRewardClaimed represents a RewardClaimed event raised by the SubnetAppRegistry contract.
type SubnetAppRegistryRewardClaimed struct {
	AppId    *big.Int
	SubnetId *big.Int
	Node     common.Address
	Reward   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterRewardClaimed is a free log retrieval operation binding the contract event 0xa8325f46ec9623d7ced8f695ff820511a4acef4f5d5c4933bd5d992141a2b905.
//
// Solidity: event RewardClaimed(uint256 indexed appId, uint256 indexed subnetId, address indexed node, uint256 reward)
func (_SubnetAppRegistry *SubnetAppRegistryFilterer) FilterRewardClaimed(opts *bind.FilterOpts, appId []*big.Int, subnetId []*big.Int, node []common.Address) (*SubnetAppRegistryRewardClaimedIterator, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}
	var subnetIdRule []interface{}
	for _, subnetIdItem := range subnetId {
		subnetIdRule = append(subnetIdRule, subnetIdItem)
	}
	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _SubnetAppRegistry.contract.FilterLogs(opts, "RewardClaimed", appIdRule, subnetIdRule, nodeRule)
	if err != nil {
		return nil, err
	}
	return &SubnetAppRegistryRewardClaimedIterator{contract: _SubnetAppRegistry.contract, event: "RewardClaimed", logs: logs, sub: sub}, nil
}

// WatchRewardClaimed is a free log subscription operation binding the contract event 0xa8325f46ec9623d7ced8f695ff820511a4acef4f5d5c4933bd5d992141a2b905.
//
// Solidity: event RewardClaimed(uint256 indexed appId, uint256 indexed subnetId, address indexed node, uint256 reward)
func (_SubnetAppRegistry *SubnetAppRegistryFilterer) WatchRewardClaimed(opts *bind.WatchOpts, sink chan<- *SubnetAppRegistryRewardClaimed, appId []*big.Int, subnetId []*big.Int, node []common.Address) (event.Subscription, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}
	var subnetIdRule []interface{}
	for _, subnetIdItem := range subnetId {
		subnetIdRule = append(subnetIdRule, subnetIdItem)
	}
	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _SubnetAppRegistry.contract.WatchLogs(opts, "RewardClaimed", appIdRule, subnetIdRule, nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetAppRegistryRewardClaimed)
				if err := _SubnetAppRegistry.contract.UnpackLog(event, "RewardClaimed", log); err != nil {
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

// ParseRewardClaimed is a log parse operation binding the contract event 0xa8325f46ec9623d7ced8f695ff820511a4acef4f5d5c4933bd5d992141a2b905.
//
// Solidity: event RewardClaimed(uint256 indexed appId, uint256 indexed subnetId, address indexed node, uint256 reward)
func (_SubnetAppRegistry *SubnetAppRegistryFilterer) ParseRewardClaimed(log types.Log) (*SubnetAppRegistryRewardClaimed, error) {
	event := new(SubnetAppRegistryRewardClaimed)
	if err := _SubnetAppRegistry.contract.UnpackLog(event, "RewardClaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
