// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package apps

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

// SubnetRegistrySubnet is an auto generated low-level Go binding around an user-defined struct.
type SubnetRegistrySubnet struct {
	Name          string
	NftId         *big.Int
	Owner         common.Address
	PeerAddr      string
	Metadata      string
	StartTime     *big.Int
	TotalUptime   *big.Int
	ClaimedUptime *big.Int
	Active        bool
	TrustScores   *big.Int
}

// SubnetRegistryMetaData contains all meta data concerning the SubnetRegistry contract.
var SubnetRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"initialOwner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_nftContract\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_rewardPerSecond\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"subnetId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"peerAddr\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"RewardClaimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"oldRewardPerSecond\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newRewardPerSecond\",\"type\":\"uint256\"}],\"name\":\"RewardPerSecondUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"updater\",\"type\":\"address\"}],\"name\":\"ScoreUpdaterAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"updater\",\"type\":\"address\"}],\"name\":\"ScoreUpdaterRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"subnetId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"peerAddr\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"uptime\",\"type\":\"uint256\"}],\"name\":\"SubnetDeregistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"subnetId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nftId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"peerAddr\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"}],\"name\":\"SubnetRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"subnetId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newScore\",\"type\":\"uint256\"}],\"name\":\"TrustScoreUpdated\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"DEFAULT_TRUST_SCORE\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"updater\",\"type\":\"address\"}],\"name\":\"addScoreUpdater\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"subnetId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalUptime\",\"type\":\"uint256\"},{\"internalType\":\"bytes32[]\",\"name\":\"proof\",\"type\":\"bytes32[]\"}],\"name\":\"claimReward\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"subnetId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"points\",\"type\":\"uint256\"}],\"name\":\"decreaseTrustScore\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"subnetId\",\"type\":\"uint256\"}],\"name\":\"deregisterSubnet\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"subnetId\",\"type\":\"uint256\"}],\"name\":\"getSubnet\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"nftId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"peerAddr\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"startTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalUptime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"claimedUptime\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"trustScores\",\"type\":\"uint256\"}],\"internalType\":\"structSubnetRegistry.Subnet\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"subnetId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"points\",\"type\":\"uint256\"}],\"name\":\"increaseTrustScore\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"merkleRoot\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"nftContract\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"peerToSubnet\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nftId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"peerAddr\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"}],\"name\":\"registerSubnet\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"updater\",\"type\":\"address\"}],\"name\":\"removeScoreUpdater\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"rewardPerSecond\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"scoreUpdaters\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"subnetCounter\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"subnets\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"nftId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"peerAddr\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"startTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalUptime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"claimedUptime\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"trustScores\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_merkleRoot\",\"type\":\"bytes32\"}],\"name\":\"updateMerkleRoot\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_rewardPerSecond\",\"type\":\"uint256\"}],\"name\":\"updateRewardPerSecond\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// SubnetRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use SubnetRegistryMetaData.ABI instead.
var SubnetRegistryABI = SubnetRegistryMetaData.ABI

// SubnetRegistry is an auto generated Go binding around an Ethereum contract.
type SubnetRegistry struct {
	SubnetRegistryCaller     // Read-only binding to the contract
	SubnetRegistryTransactor // Write-only binding to the contract
	SubnetRegistryFilterer   // Log filterer for contract events
}

// SubnetRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type SubnetRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SubnetRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SubnetRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SubnetRegistrySession struct {
	Contract     *SubnetRegistry   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SubnetRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SubnetRegistryCallerSession struct {
	Contract *SubnetRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// SubnetRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SubnetRegistryTransactorSession struct {
	Contract     *SubnetRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// SubnetRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type SubnetRegistryRaw struct {
	Contract *SubnetRegistry // Generic contract binding to access the raw methods on
}

// SubnetRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SubnetRegistryCallerRaw struct {
	Contract *SubnetRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// SubnetRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SubnetRegistryTransactorRaw struct {
	Contract *SubnetRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSubnetRegistry creates a new instance of SubnetRegistry, bound to a specific deployed contract.
func NewSubnetRegistry(address common.Address, backend bind.ContractBackend) (*SubnetRegistry, error) {
	contract, err := bindSubnetRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SubnetRegistry{SubnetRegistryCaller: SubnetRegistryCaller{contract: contract}, SubnetRegistryTransactor: SubnetRegistryTransactor{contract: contract}, SubnetRegistryFilterer: SubnetRegistryFilterer{contract: contract}}, nil
}

// NewSubnetRegistryCaller creates a new read-only instance of SubnetRegistry, bound to a specific deployed contract.
func NewSubnetRegistryCaller(address common.Address, caller bind.ContractCaller) (*SubnetRegistryCaller, error) {
	contract, err := bindSubnetRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SubnetRegistryCaller{contract: contract}, nil
}

// NewSubnetRegistryTransactor creates a new write-only instance of SubnetRegistry, bound to a specific deployed contract.
func NewSubnetRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*SubnetRegistryTransactor, error) {
	contract, err := bindSubnetRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SubnetRegistryTransactor{contract: contract}, nil
}

// NewSubnetRegistryFilterer creates a new log filterer instance of SubnetRegistry, bound to a specific deployed contract.
func NewSubnetRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*SubnetRegistryFilterer, error) {
	contract, err := bindSubnetRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SubnetRegistryFilterer{contract: contract}, nil
}

// bindSubnetRegistry binds a generic wrapper to an already deployed contract.
func bindSubnetRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SubnetRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SubnetRegistry *SubnetRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SubnetRegistry.Contract.SubnetRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SubnetRegistry *SubnetRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.SubnetRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SubnetRegistry *SubnetRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.SubnetRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SubnetRegistry *SubnetRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SubnetRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SubnetRegistry *SubnetRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SubnetRegistry *SubnetRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTTRUSTSCORE is a free data retrieval call binding the contract method 0x354f63f4.
//
// Solidity: function DEFAULT_TRUST_SCORE() view returns(uint256)
func (_SubnetRegistry *SubnetRegistryCaller) DEFAULTTRUSTSCORE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SubnetRegistry.contract.Call(opts, &out, "DEFAULT_TRUST_SCORE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DEFAULTTRUSTSCORE is a free data retrieval call binding the contract method 0x354f63f4.
//
// Solidity: function DEFAULT_TRUST_SCORE() view returns(uint256)
func (_SubnetRegistry *SubnetRegistrySession) DEFAULTTRUSTSCORE() (*big.Int, error) {
	return _SubnetRegistry.Contract.DEFAULTTRUSTSCORE(&_SubnetRegistry.CallOpts)
}

// DEFAULTTRUSTSCORE is a free data retrieval call binding the contract method 0x354f63f4.
//
// Solidity: function DEFAULT_TRUST_SCORE() view returns(uint256)
func (_SubnetRegistry *SubnetRegistryCallerSession) DEFAULTTRUSTSCORE() (*big.Int, error) {
	return _SubnetRegistry.Contract.DEFAULTTRUSTSCORE(&_SubnetRegistry.CallOpts)
}

// GetSubnet is a free data retrieval call binding the contract method 0x58ca7504.
//
// Solidity: function getSubnet(uint256 subnetId) view returns((string,uint256,address,string,string,uint256,uint256,uint256,bool,uint256))
func (_SubnetRegistry *SubnetRegistryCaller) GetSubnet(opts *bind.CallOpts, subnetId *big.Int) (SubnetRegistrySubnet, error) {
	var out []interface{}
	err := _SubnetRegistry.contract.Call(opts, &out, "getSubnet", subnetId)

	if err != nil {
		return *new(SubnetRegistrySubnet), err
	}

	out0 := *abi.ConvertType(out[0], new(SubnetRegistrySubnet)).(*SubnetRegistrySubnet)

	return out0, err

}

// GetSubnet is a free data retrieval call binding the contract method 0x58ca7504.
//
// Solidity: function getSubnet(uint256 subnetId) view returns((string,uint256,address,string,string,uint256,uint256,uint256,bool,uint256))
func (_SubnetRegistry *SubnetRegistrySession) GetSubnet(subnetId *big.Int) (SubnetRegistrySubnet, error) {
	return _SubnetRegistry.Contract.GetSubnet(&_SubnetRegistry.CallOpts, subnetId)
}

// GetSubnet is a free data retrieval call binding the contract method 0x58ca7504.
//
// Solidity: function getSubnet(uint256 subnetId) view returns((string,uint256,address,string,string,uint256,uint256,uint256,bool,uint256))
func (_SubnetRegistry *SubnetRegistryCallerSession) GetSubnet(subnetId *big.Int) (SubnetRegistrySubnet, error) {
	return _SubnetRegistry.Contract.GetSubnet(&_SubnetRegistry.CallOpts, subnetId)
}

// MerkleRoot is a free data retrieval call binding the contract method 0x2eb4a7ab.
//
// Solidity: function merkleRoot() view returns(bytes32)
func (_SubnetRegistry *SubnetRegistryCaller) MerkleRoot(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SubnetRegistry.contract.Call(opts, &out, "merkleRoot")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MerkleRoot is a free data retrieval call binding the contract method 0x2eb4a7ab.
//
// Solidity: function merkleRoot() view returns(bytes32)
func (_SubnetRegistry *SubnetRegistrySession) MerkleRoot() ([32]byte, error) {
	return _SubnetRegistry.Contract.MerkleRoot(&_SubnetRegistry.CallOpts)
}

// MerkleRoot is a free data retrieval call binding the contract method 0x2eb4a7ab.
//
// Solidity: function merkleRoot() view returns(bytes32)
func (_SubnetRegistry *SubnetRegistryCallerSession) MerkleRoot() ([32]byte, error) {
	return _SubnetRegistry.Contract.MerkleRoot(&_SubnetRegistry.CallOpts)
}

// NftContract is a free data retrieval call binding the contract method 0xd56d229d.
//
// Solidity: function nftContract() view returns(address)
func (_SubnetRegistry *SubnetRegistryCaller) NftContract(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetRegistry.contract.Call(opts, &out, "nftContract")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// NftContract is a free data retrieval call binding the contract method 0xd56d229d.
//
// Solidity: function nftContract() view returns(address)
func (_SubnetRegistry *SubnetRegistrySession) NftContract() (common.Address, error) {
	return _SubnetRegistry.Contract.NftContract(&_SubnetRegistry.CallOpts)
}

// NftContract is a free data retrieval call binding the contract method 0xd56d229d.
//
// Solidity: function nftContract() view returns(address)
func (_SubnetRegistry *SubnetRegistryCallerSession) NftContract() (common.Address, error) {
	return _SubnetRegistry.Contract.NftContract(&_SubnetRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetRegistry *SubnetRegistryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetRegistry.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetRegistry *SubnetRegistrySession) Owner() (common.Address, error) {
	return _SubnetRegistry.Contract.Owner(&_SubnetRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetRegistry *SubnetRegistryCallerSession) Owner() (common.Address, error) {
	return _SubnetRegistry.Contract.Owner(&_SubnetRegistry.CallOpts)
}

// PeerToSubnet is a free data retrieval call binding the contract method 0xd52a45f1.
//
// Solidity: function peerToSubnet(string ) view returns(uint256)
func (_SubnetRegistry *SubnetRegistryCaller) PeerToSubnet(opts *bind.CallOpts, arg0 string) (*big.Int, error) {
	var out []interface{}
	err := _SubnetRegistry.contract.Call(opts, &out, "peerToSubnet", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PeerToSubnet is a free data retrieval call binding the contract method 0xd52a45f1.
//
// Solidity: function peerToSubnet(string ) view returns(uint256)
func (_SubnetRegistry *SubnetRegistrySession) PeerToSubnet(arg0 string) (*big.Int, error) {
	return _SubnetRegistry.Contract.PeerToSubnet(&_SubnetRegistry.CallOpts, arg0)
}

// PeerToSubnet is a free data retrieval call binding the contract method 0xd52a45f1.
//
// Solidity: function peerToSubnet(string ) view returns(uint256)
func (_SubnetRegistry *SubnetRegistryCallerSession) PeerToSubnet(arg0 string) (*big.Int, error) {
	return _SubnetRegistry.Contract.PeerToSubnet(&_SubnetRegistry.CallOpts, arg0)
}

// RewardPerSecond is a free data retrieval call binding the contract method 0x8f10369a.
//
// Solidity: function rewardPerSecond() view returns(uint256)
func (_SubnetRegistry *SubnetRegistryCaller) RewardPerSecond(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SubnetRegistry.contract.Call(opts, &out, "rewardPerSecond")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RewardPerSecond is a free data retrieval call binding the contract method 0x8f10369a.
//
// Solidity: function rewardPerSecond() view returns(uint256)
func (_SubnetRegistry *SubnetRegistrySession) RewardPerSecond() (*big.Int, error) {
	return _SubnetRegistry.Contract.RewardPerSecond(&_SubnetRegistry.CallOpts)
}

// RewardPerSecond is a free data retrieval call binding the contract method 0x8f10369a.
//
// Solidity: function rewardPerSecond() view returns(uint256)
func (_SubnetRegistry *SubnetRegistryCallerSession) RewardPerSecond() (*big.Int, error) {
	return _SubnetRegistry.Contract.RewardPerSecond(&_SubnetRegistry.CallOpts)
}

// ScoreUpdaters is a free data retrieval call binding the contract method 0x867951d8.
//
// Solidity: function scoreUpdaters(address ) view returns(bool)
func (_SubnetRegistry *SubnetRegistryCaller) ScoreUpdaters(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _SubnetRegistry.contract.Call(opts, &out, "scoreUpdaters", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ScoreUpdaters is a free data retrieval call binding the contract method 0x867951d8.
//
// Solidity: function scoreUpdaters(address ) view returns(bool)
func (_SubnetRegistry *SubnetRegistrySession) ScoreUpdaters(arg0 common.Address) (bool, error) {
	return _SubnetRegistry.Contract.ScoreUpdaters(&_SubnetRegistry.CallOpts, arg0)
}

// ScoreUpdaters is a free data retrieval call binding the contract method 0x867951d8.
//
// Solidity: function scoreUpdaters(address ) view returns(bool)
func (_SubnetRegistry *SubnetRegistryCallerSession) ScoreUpdaters(arg0 common.Address) (bool, error) {
	return _SubnetRegistry.Contract.ScoreUpdaters(&_SubnetRegistry.CallOpts, arg0)
}

// SubnetCounter is a free data retrieval call binding the contract method 0x1b2f926e.
//
// Solidity: function subnetCounter() view returns(uint256)
func (_SubnetRegistry *SubnetRegistryCaller) SubnetCounter(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SubnetRegistry.contract.Call(opts, &out, "subnetCounter")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SubnetCounter is a free data retrieval call binding the contract method 0x1b2f926e.
//
// Solidity: function subnetCounter() view returns(uint256)
func (_SubnetRegistry *SubnetRegistrySession) SubnetCounter() (*big.Int, error) {
	return _SubnetRegistry.Contract.SubnetCounter(&_SubnetRegistry.CallOpts)
}

// SubnetCounter is a free data retrieval call binding the contract method 0x1b2f926e.
//
// Solidity: function subnetCounter() view returns(uint256)
func (_SubnetRegistry *SubnetRegistryCallerSession) SubnetCounter() (*big.Int, error) {
	return _SubnetRegistry.Contract.SubnetCounter(&_SubnetRegistry.CallOpts)
}

// Subnets is a free data retrieval call binding the contract method 0x475726f7.
//
// Solidity: function subnets(uint256 ) view returns(string name, uint256 nftId, address owner, string peerAddr, string metadata, uint256 startTime, uint256 totalUptime, uint256 claimedUptime, bool active, uint256 trustScores)
func (_SubnetRegistry *SubnetRegistryCaller) Subnets(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Name          string
	NftId         *big.Int
	Owner         common.Address
	PeerAddr      string
	Metadata      string
	StartTime     *big.Int
	TotalUptime   *big.Int
	ClaimedUptime *big.Int
	Active        bool
	TrustScores   *big.Int
}, error) {
	var out []interface{}
	err := _SubnetRegistry.contract.Call(opts, &out, "subnets", arg0)

	outstruct := new(struct {
		Name          string
		NftId         *big.Int
		Owner         common.Address
		PeerAddr      string
		Metadata      string
		StartTime     *big.Int
		TotalUptime   *big.Int
		ClaimedUptime *big.Int
		Active        bool
		TrustScores   *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Name = *abi.ConvertType(out[0], new(string)).(*string)
	outstruct.NftId = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Owner = *abi.ConvertType(out[2], new(common.Address)).(*common.Address)
	outstruct.PeerAddr = *abi.ConvertType(out[3], new(string)).(*string)
	outstruct.Metadata = *abi.ConvertType(out[4], new(string)).(*string)
	outstruct.StartTime = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)
	outstruct.TotalUptime = *abi.ConvertType(out[6], new(*big.Int)).(**big.Int)
	outstruct.ClaimedUptime = *abi.ConvertType(out[7], new(*big.Int)).(**big.Int)
	outstruct.Active = *abi.ConvertType(out[8], new(bool)).(*bool)
	outstruct.TrustScores = *abi.ConvertType(out[9], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Subnets is a free data retrieval call binding the contract method 0x475726f7.
//
// Solidity: function subnets(uint256 ) view returns(string name, uint256 nftId, address owner, string peerAddr, string metadata, uint256 startTime, uint256 totalUptime, uint256 claimedUptime, bool active, uint256 trustScores)
func (_SubnetRegistry *SubnetRegistrySession) Subnets(arg0 *big.Int) (struct {
	Name          string
	NftId         *big.Int
	Owner         common.Address
	PeerAddr      string
	Metadata      string
	StartTime     *big.Int
	TotalUptime   *big.Int
	ClaimedUptime *big.Int
	Active        bool
	TrustScores   *big.Int
}, error) {
	return _SubnetRegistry.Contract.Subnets(&_SubnetRegistry.CallOpts, arg0)
}

// Subnets is a free data retrieval call binding the contract method 0x475726f7.
//
// Solidity: function subnets(uint256 ) view returns(string name, uint256 nftId, address owner, string peerAddr, string metadata, uint256 startTime, uint256 totalUptime, uint256 claimedUptime, bool active, uint256 trustScores)
func (_SubnetRegistry *SubnetRegistryCallerSession) Subnets(arg0 *big.Int) (struct {
	Name          string
	NftId         *big.Int
	Owner         common.Address
	PeerAddr      string
	Metadata      string
	StartTime     *big.Int
	TotalUptime   *big.Int
	ClaimedUptime *big.Int
	Active        bool
	TrustScores   *big.Int
}, error) {
	return _SubnetRegistry.Contract.Subnets(&_SubnetRegistry.CallOpts, arg0)
}

// AddScoreUpdater is a paid mutator transaction binding the contract method 0xdc3812a2.
//
// Solidity: function addScoreUpdater(address updater) returns()
func (_SubnetRegistry *SubnetRegistryTransactor) AddScoreUpdater(opts *bind.TransactOpts, updater common.Address) (*types.Transaction, error) {
	return _SubnetRegistry.contract.Transact(opts, "addScoreUpdater", updater)
}

// AddScoreUpdater is a paid mutator transaction binding the contract method 0xdc3812a2.
//
// Solidity: function addScoreUpdater(address updater) returns()
func (_SubnetRegistry *SubnetRegistrySession) AddScoreUpdater(updater common.Address) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.AddScoreUpdater(&_SubnetRegistry.TransactOpts, updater)
}

// AddScoreUpdater is a paid mutator transaction binding the contract method 0xdc3812a2.
//
// Solidity: function addScoreUpdater(address updater) returns()
func (_SubnetRegistry *SubnetRegistryTransactorSession) AddScoreUpdater(updater common.Address) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.AddScoreUpdater(&_SubnetRegistry.TransactOpts, updater)
}

// ClaimReward is a paid mutator transaction binding the contract method 0x9821f5f2.
//
// Solidity: function claimReward(uint256 subnetId, uint256 totalUptime, bytes32[] proof) returns()
func (_SubnetRegistry *SubnetRegistryTransactor) ClaimReward(opts *bind.TransactOpts, subnetId *big.Int, totalUptime *big.Int, proof [][32]byte) (*types.Transaction, error) {
	return _SubnetRegistry.contract.Transact(opts, "claimReward", subnetId, totalUptime, proof)
}

// ClaimReward is a paid mutator transaction binding the contract method 0x9821f5f2.
//
// Solidity: function claimReward(uint256 subnetId, uint256 totalUptime, bytes32[] proof) returns()
func (_SubnetRegistry *SubnetRegistrySession) ClaimReward(subnetId *big.Int, totalUptime *big.Int, proof [][32]byte) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.ClaimReward(&_SubnetRegistry.TransactOpts, subnetId, totalUptime, proof)
}

// ClaimReward is a paid mutator transaction binding the contract method 0x9821f5f2.
//
// Solidity: function claimReward(uint256 subnetId, uint256 totalUptime, bytes32[] proof) returns()
func (_SubnetRegistry *SubnetRegistryTransactorSession) ClaimReward(subnetId *big.Int, totalUptime *big.Int, proof [][32]byte) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.ClaimReward(&_SubnetRegistry.TransactOpts, subnetId, totalUptime, proof)
}

// DecreaseTrustScore is a paid mutator transaction binding the contract method 0xd2ef5b4e.
//
// Solidity: function decreaseTrustScore(uint256 subnetId, uint256 points) returns()
func (_SubnetRegistry *SubnetRegistryTransactor) DecreaseTrustScore(opts *bind.TransactOpts, subnetId *big.Int, points *big.Int) (*types.Transaction, error) {
	return _SubnetRegistry.contract.Transact(opts, "decreaseTrustScore", subnetId, points)
}

// DecreaseTrustScore is a paid mutator transaction binding the contract method 0xd2ef5b4e.
//
// Solidity: function decreaseTrustScore(uint256 subnetId, uint256 points) returns()
func (_SubnetRegistry *SubnetRegistrySession) DecreaseTrustScore(subnetId *big.Int, points *big.Int) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.DecreaseTrustScore(&_SubnetRegistry.TransactOpts, subnetId, points)
}

// DecreaseTrustScore is a paid mutator transaction binding the contract method 0xd2ef5b4e.
//
// Solidity: function decreaseTrustScore(uint256 subnetId, uint256 points) returns()
func (_SubnetRegistry *SubnetRegistryTransactorSession) DecreaseTrustScore(subnetId *big.Int, points *big.Int) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.DecreaseTrustScore(&_SubnetRegistry.TransactOpts, subnetId, points)
}

// Deposit is a paid mutator transaction binding the contract method 0xd0e30db0.
//
// Solidity: function deposit() payable returns()
func (_SubnetRegistry *SubnetRegistryTransactor) Deposit(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetRegistry.contract.Transact(opts, "deposit")
}

// Deposit is a paid mutator transaction binding the contract method 0xd0e30db0.
//
// Solidity: function deposit() payable returns()
func (_SubnetRegistry *SubnetRegistrySession) Deposit() (*types.Transaction, error) {
	return _SubnetRegistry.Contract.Deposit(&_SubnetRegistry.TransactOpts)
}

// Deposit is a paid mutator transaction binding the contract method 0xd0e30db0.
//
// Solidity: function deposit() payable returns()
func (_SubnetRegistry *SubnetRegistryTransactorSession) Deposit() (*types.Transaction, error) {
	return _SubnetRegistry.Contract.Deposit(&_SubnetRegistry.TransactOpts)
}

// DeregisterSubnet is a paid mutator transaction binding the contract method 0x0cf02c5e.
//
// Solidity: function deregisterSubnet(uint256 subnetId) returns()
func (_SubnetRegistry *SubnetRegistryTransactor) DeregisterSubnet(opts *bind.TransactOpts, subnetId *big.Int) (*types.Transaction, error) {
	return _SubnetRegistry.contract.Transact(opts, "deregisterSubnet", subnetId)
}

// DeregisterSubnet is a paid mutator transaction binding the contract method 0x0cf02c5e.
//
// Solidity: function deregisterSubnet(uint256 subnetId) returns()
func (_SubnetRegistry *SubnetRegistrySession) DeregisterSubnet(subnetId *big.Int) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.DeregisterSubnet(&_SubnetRegistry.TransactOpts, subnetId)
}

// DeregisterSubnet is a paid mutator transaction binding the contract method 0x0cf02c5e.
//
// Solidity: function deregisterSubnet(uint256 subnetId) returns()
func (_SubnetRegistry *SubnetRegistryTransactorSession) DeregisterSubnet(subnetId *big.Int) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.DeregisterSubnet(&_SubnetRegistry.TransactOpts, subnetId)
}

// IncreaseTrustScore is a paid mutator transaction binding the contract method 0xd05eff83.
//
// Solidity: function increaseTrustScore(uint256 subnetId, uint256 points) returns()
func (_SubnetRegistry *SubnetRegistryTransactor) IncreaseTrustScore(opts *bind.TransactOpts, subnetId *big.Int, points *big.Int) (*types.Transaction, error) {
	return _SubnetRegistry.contract.Transact(opts, "increaseTrustScore", subnetId, points)
}

// IncreaseTrustScore is a paid mutator transaction binding the contract method 0xd05eff83.
//
// Solidity: function increaseTrustScore(uint256 subnetId, uint256 points) returns()
func (_SubnetRegistry *SubnetRegistrySession) IncreaseTrustScore(subnetId *big.Int, points *big.Int) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.IncreaseTrustScore(&_SubnetRegistry.TransactOpts, subnetId, points)
}

// IncreaseTrustScore is a paid mutator transaction binding the contract method 0xd05eff83.
//
// Solidity: function increaseTrustScore(uint256 subnetId, uint256 points) returns()
func (_SubnetRegistry *SubnetRegistryTransactorSession) IncreaseTrustScore(subnetId *big.Int, points *big.Int) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.IncreaseTrustScore(&_SubnetRegistry.TransactOpts, subnetId, points)
}

// RegisterSubnet is a paid mutator transaction binding the contract method 0x5ef791f7.
//
// Solidity: function registerSubnet(uint256 nftId, string peerAddr, string name, string metadata) returns()
func (_SubnetRegistry *SubnetRegistryTransactor) RegisterSubnet(opts *bind.TransactOpts, nftId *big.Int, peerAddr string, name string, metadata string) (*types.Transaction, error) {
	return _SubnetRegistry.contract.Transact(opts, "registerSubnet", nftId, peerAddr, name, metadata)
}

// RegisterSubnet is a paid mutator transaction binding the contract method 0x5ef791f7.
//
// Solidity: function registerSubnet(uint256 nftId, string peerAddr, string name, string metadata) returns()
func (_SubnetRegistry *SubnetRegistrySession) RegisterSubnet(nftId *big.Int, peerAddr string, name string, metadata string) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.RegisterSubnet(&_SubnetRegistry.TransactOpts, nftId, peerAddr, name, metadata)
}

// RegisterSubnet is a paid mutator transaction binding the contract method 0x5ef791f7.
//
// Solidity: function registerSubnet(uint256 nftId, string peerAddr, string name, string metadata) returns()
func (_SubnetRegistry *SubnetRegistryTransactorSession) RegisterSubnet(nftId *big.Int, peerAddr string, name string, metadata string) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.RegisterSubnet(&_SubnetRegistry.TransactOpts, nftId, peerAddr, name, metadata)
}

// RemoveScoreUpdater is a paid mutator transaction binding the contract method 0x3416aca0.
//
// Solidity: function removeScoreUpdater(address updater) returns()
func (_SubnetRegistry *SubnetRegistryTransactor) RemoveScoreUpdater(opts *bind.TransactOpts, updater common.Address) (*types.Transaction, error) {
	return _SubnetRegistry.contract.Transact(opts, "removeScoreUpdater", updater)
}

// RemoveScoreUpdater is a paid mutator transaction binding the contract method 0x3416aca0.
//
// Solidity: function removeScoreUpdater(address updater) returns()
func (_SubnetRegistry *SubnetRegistrySession) RemoveScoreUpdater(updater common.Address) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.RemoveScoreUpdater(&_SubnetRegistry.TransactOpts, updater)
}

// RemoveScoreUpdater is a paid mutator transaction binding the contract method 0x3416aca0.
//
// Solidity: function removeScoreUpdater(address updater) returns()
func (_SubnetRegistry *SubnetRegistryTransactorSession) RemoveScoreUpdater(updater common.Address) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.RemoveScoreUpdater(&_SubnetRegistry.TransactOpts, updater)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetRegistry *SubnetRegistryTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetRegistry.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetRegistry *SubnetRegistrySession) RenounceOwnership() (*types.Transaction, error) {
	return _SubnetRegistry.Contract.RenounceOwnership(&_SubnetRegistry.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetRegistry *SubnetRegistryTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _SubnetRegistry.Contract.RenounceOwnership(&_SubnetRegistry.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetRegistry *SubnetRegistryTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _SubnetRegistry.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetRegistry *SubnetRegistrySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.TransferOwnership(&_SubnetRegistry.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetRegistry *SubnetRegistryTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.TransferOwnership(&_SubnetRegistry.TransactOpts, newOwner)
}

// UpdateMerkleRoot is a paid mutator transaction binding the contract method 0x4783f0ef.
//
// Solidity: function updateMerkleRoot(bytes32 _merkleRoot) returns()
func (_SubnetRegistry *SubnetRegistryTransactor) UpdateMerkleRoot(opts *bind.TransactOpts, _merkleRoot [32]byte) (*types.Transaction, error) {
	return _SubnetRegistry.contract.Transact(opts, "updateMerkleRoot", _merkleRoot)
}

// UpdateMerkleRoot is a paid mutator transaction binding the contract method 0x4783f0ef.
//
// Solidity: function updateMerkleRoot(bytes32 _merkleRoot) returns()
func (_SubnetRegistry *SubnetRegistrySession) UpdateMerkleRoot(_merkleRoot [32]byte) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.UpdateMerkleRoot(&_SubnetRegistry.TransactOpts, _merkleRoot)
}

// UpdateMerkleRoot is a paid mutator transaction binding the contract method 0x4783f0ef.
//
// Solidity: function updateMerkleRoot(bytes32 _merkleRoot) returns()
func (_SubnetRegistry *SubnetRegistryTransactorSession) UpdateMerkleRoot(_merkleRoot [32]byte) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.UpdateMerkleRoot(&_SubnetRegistry.TransactOpts, _merkleRoot)
}

// UpdateRewardPerSecond is a paid mutator transaction binding the contract method 0x4004c8e7.
//
// Solidity: function updateRewardPerSecond(uint256 _rewardPerSecond) returns()
func (_SubnetRegistry *SubnetRegistryTransactor) UpdateRewardPerSecond(opts *bind.TransactOpts, _rewardPerSecond *big.Int) (*types.Transaction, error) {
	return _SubnetRegistry.contract.Transact(opts, "updateRewardPerSecond", _rewardPerSecond)
}

// UpdateRewardPerSecond is a paid mutator transaction binding the contract method 0x4004c8e7.
//
// Solidity: function updateRewardPerSecond(uint256 _rewardPerSecond) returns()
func (_SubnetRegistry *SubnetRegistrySession) UpdateRewardPerSecond(_rewardPerSecond *big.Int) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.UpdateRewardPerSecond(&_SubnetRegistry.TransactOpts, _rewardPerSecond)
}

// UpdateRewardPerSecond is a paid mutator transaction binding the contract method 0x4004c8e7.
//
// Solidity: function updateRewardPerSecond(uint256 _rewardPerSecond) returns()
func (_SubnetRegistry *SubnetRegistryTransactorSession) UpdateRewardPerSecond(_rewardPerSecond *big.Int) (*types.Transaction, error) {
	return _SubnetRegistry.Contract.UpdateRewardPerSecond(&_SubnetRegistry.TransactOpts, _rewardPerSecond)
}

// SubnetRegistryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the SubnetRegistry contract.
type SubnetRegistryOwnershipTransferredIterator struct {
	Event *SubnetRegistryOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *SubnetRegistryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetRegistryOwnershipTransferred)
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
		it.Event = new(SubnetRegistryOwnershipTransferred)
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
func (it *SubnetRegistryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetRegistryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetRegistryOwnershipTransferred represents a OwnershipTransferred event raised by the SubnetRegistry contract.
type SubnetRegistryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SubnetRegistry *SubnetRegistryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*SubnetRegistryOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SubnetRegistry.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &SubnetRegistryOwnershipTransferredIterator{contract: _SubnetRegistry.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SubnetRegistry *SubnetRegistryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *SubnetRegistryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SubnetRegistry.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetRegistryOwnershipTransferred)
				if err := _SubnetRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_SubnetRegistry *SubnetRegistryFilterer) ParseOwnershipTransferred(log types.Log) (*SubnetRegistryOwnershipTransferred, error) {
	event := new(SubnetRegistryOwnershipTransferred)
	if err := _SubnetRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetRegistryRewardClaimedIterator is returned from FilterRewardClaimed and is used to iterate over the raw logs and unpacked data for RewardClaimed events raised by the SubnetRegistry contract.
type SubnetRegistryRewardClaimedIterator struct {
	Event *SubnetRegistryRewardClaimed // Event containing the contract specifics and raw log

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
func (it *SubnetRegistryRewardClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetRegistryRewardClaimed)
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
		it.Event = new(SubnetRegistryRewardClaimed)
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
func (it *SubnetRegistryRewardClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetRegistryRewardClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetRegistryRewardClaimed represents a RewardClaimed event raised by the SubnetRegistry contract.
type SubnetRegistryRewardClaimed struct {
	SubnetId *big.Int
	Owner    common.Address
	PeerAddr string
	Amount   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterRewardClaimed is a free log retrieval operation binding the contract event 0xec9aff1b12453a86c3a301e35a3077c89762449edc1330f20e4840903cbc06b3.
//
// Solidity: event RewardClaimed(uint256 indexed subnetId, address indexed owner, string peerAddr, uint256 amount)
func (_SubnetRegistry *SubnetRegistryFilterer) FilterRewardClaimed(opts *bind.FilterOpts, subnetId []*big.Int, owner []common.Address) (*SubnetRegistryRewardClaimedIterator, error) {

	var subnetIdRule []interface{}
	for _, subnetIdItem := range subnetId {
		subnetIdRule = append(subnetIdRule, subnetIdItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _SubnetRegistry.contract.FilterLogs(opts, "RewardClaimed", subnetIdRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &SubnetRegistryRewardClaimedIterator{contract: _SubnetRegistry.contract, event: "RewardClaimed", logs: logs, sub: sub}, nil
}

// WatchRewardClaimed is a free log subscription operation binding the contract event 0xec9aff1b12453a86c3a301e35a3077c89762449edc1330f20e4840903cbc06b3.
//
// Solidity: event RewardClaimed(uint256 indexed subnetId, address indexed owner, string peerAddr, uint256 amount)
func (_SubnetRegistry *SubnetRegistryFilterer) WatchRewardClaimed(opts *bind.WatchOpts, sink chan<- *SubnetRegistryRewardClaimed, subnetId []*big.Int, owner []common.Address) (event.Subscription, error) {

	var subnetIdRule []interface{}
	for _, subnetIdItem := range subnetId {
		subnetIdRule = append(subnetIdRule, subnetIdItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _SubnetRegistry.contract.WatchLogs(opts, "RewardClaimed", subnetIdRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetRegistryRewardClaimed)
				if err := _SubnetRegistry.contract.UnpackLog(event, "RewardClaimed", log); err != nil {
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

// ParseRewardClaimed is a log parse operation binding the contract event 0xec9aff1b12453a86c3a301e35a3077c89762449edc1330f20e4840903cbc06b3.
//
// Solidity: event RewardClaimed(uint256 indexed subnetId, address indexed owner, string peerAddr, uint256 amount)
func (_SubnetRegistry *SubnetRegistryFilterer) ParseRewardClaimed(log types.Log) (*SubnetRegistryRewardClaimed, error) {
	event := new(SubnetRegistryRewardClaimed)
	if err := _SubnetRegistry.contract.UnpackLog(event, "RewardClaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetRegistryRewardPerSecondUpdatedIterator is returned from FilterRewardPerSecondUpdated and is used to iterate over the raw logs and unpacked data for RewardPerSecondUpdated events raised by the SubnetRegistry contract.
type SubnetRegistryRewardPerSecondUpdatedIterator struct {
	Event *SubnetRegistryRewardPerSecondUpdated // Event containing the contract specifics and raw log

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
func (it *SubnetRegistryRewardPerSecondUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetRegistryRewardPerSecondUpdated)
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
		it.Event = new(SubnetRegistryRewardPerSecondUpdated)
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
func (it *SubnetRegistryRewardPerSecondUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetRegistryRewardPerSecondUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetRegistryRewardPerSecondUpdated represents a RewardPerSecondUpdated event raised by the SubnetRegistry contract.
type SubnetRegistryRewardPerSecondUpdated struct {
	OldRewardPerSecond *big.Int
	NewRewardPerSecond *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterRewardPerSecondUpdated is a free log retrieval operation binding the contract event 0xad9b65d94abd7b01bc5e6f82634fc2860427247e02b5b06e1fefc524c6af512f.
//
// Solidity: event RewardPerSecondUpdated(uint256 oldRewardPerSecond, uint256 newRewardPerSecond)
func (_SubnetRegistry *SubnetRegistryFilterer) FilterRewardPerSecondUpdated(opts *bind.FilterOpts) (*SubnetRegistryRewardPerSecondUpdatedIterator, error) {

	logs, sub, err := _SubnetRegistry.contract.FilterLogs(opts, "RewardPerSecondUpdated")
	if err != nil {
		return nil, err
	}
	return &SubnetRegistryRewardPerSecondUpdatedIterator{contract: _SubnetRegistry.contract, event: "RewardPerSecondUpdated", logs: logs, sub: sub}, nil
}

// WatchRewardPerSecondUpdated is a free log subscription operation binding the contract event 0xad9b65d94abd7b01bc5e6f82634fc2860427247e02b5b06e1fefc524c6af512f.
//
// Solidity: event RewardPerSecondUpdated(uint256 oldRewardPerSecond, uint256 newRewardPerSecond)
func (_SubnetRegistry *SubnetRegistryFilterer) WatchRewardPerSecondUpdated(opts *bind.WatchOpts, sink chan<- *SubnetRegistryRewardPerSecondUpdated) (event.Subscription, error) {

	logs, sub, err := _SubnetRegistry.contract.WatchLogs(opts, "RewardPerSecondUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetRegistryRewardPerSecondUpdated)
				if err := _SubnetRegistry.contract.UnpackLog(event, "RewardPerSecondUpdated", log); err != nil {
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

// ParseRewardPerSecondUpdated is a log parse operation binding the contract event 0xad9b65d94abd7b01bc5e6f82634fc2860427247e02b5b06e1fefc524c6af512f.
//
// Solidity: event RewardPerSecondUpdated(uint256 oldRewardPerSecond, uint256 newRewardPerSecond)
func (_SubnetRegistry *SubnetRegistryFilterer) ParseRewardPerSecondUpdated(log types.Log) (*SubnetRegistryRewardPerSecondUpdated, error) {
	event := new(SubnetRegistryRewardPerSecondUpdated)
	if err := _SubnetRegistry.contract.UnpackLog(event, "RewardPerSecondUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetRegistryScoreUpdaterAddedIterator is returned from FilterScoreUpdaterAdded and is used to iterate over the raw logs and unpacked data for ScoreUpdaterAdded events raised by the SubnetRegistry contract.
type SubnetRegistryScoreUpdaterAddedIterator struct {
	Event *SubnetRegistryScoreUpdaterAdded // Event containing the contract specifics and raw log

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
func (it *SubnetRegistryScoreUpdaterAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetRegistryScoreUpdaterAdded)
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
		it.Event = new(SubnetRegistryScoreUpdaterAdded)
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
func (it *SubnetRegistryScoreUpdaterAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetRegistryScoreUpdaterAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetRegistryScoreUpdaterAdded represents a ScoreUpdaterAdded event raised by the SubnetRegistry contract.
type SubnetRegistryScoreUpdaterAdded struct {
	Updater common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterScoreUpdaterAdded is a free log retrieval operation binding the contract event 0xad25f623e98acc6298062eb2e16242b2bc20e36bfc9c10ab8130c5e01aa0e1a0.
//
// Solidity: event ScoreUpdaterAdded(address indexed updater)
func (_SubnetRegistry *SubnetRegistryFilterer) FilterScoreUpdaterAdded(opts *bind.FilterOpts, updater []common.Address) (*SubnetRegistryScoreUpdaterAddedIterator, error) {

	var updaterRule []interface{}
	for _, updaterItem := range updater {
		updaterRule = append(updaterRule, updaterItem)
	}

	logs, sub, err := _SubnetRegistry.contract.FilterLogs(opts, "ScoreUpdaterAdded", updaterRule)
	if err != nil {
		return nil, err
	}
	return &SubnetRegistryScoreUpdaterAddedIterator{contract: _SubnetRegistry.contract, event: "ScoreUpdaterAdded", logs: logs, sub: sub}, nil
}

// WatchScoreUpdaterAdded is a free log subscription operation binding the contract event 0xad25f623e98acc6298062eb2e16242b2bc20e36bfc9c10ab8130c5e01aa0e1a0.
//
// Solidity: event ScoreUpdaterAdded(address indexed updater)
func (_SubnetRegistry *SubnetRegistryFilterer) WatchScoreUpdaterAdded(opts *bind.WatchOpts, sink chan<- *SubnetRegistryScoreUpdaterAdded, updater []common.Address) (event.Subscription, error) {

	var updaterRule []interface{}
	for _, updaterItem := range updater {
		updaterRule = append(updaterRule, updaterItem)
	}

	logs, sub, err := _SubnetRegistry.contract.WatchLogs(opts, "ScoreUpdaterAdded", updaterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetRegistryScoreUpdaterAdded)
				if err := _SubnetRegistry.contract.UnpackLog(event, "ScoreUpdaterAdded", log); err != nil {
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

// ParseScoreUpdaterAdded is a log parse operation binding the contract event 0xad25f623e98acc6298062eb2e16242b2bc20e36bfc9c10ab8130c5e01aa0e1a0.
//
// Solidity: event ScoreUpdaterAdded(address indexed updater)
func (_SubnetRegistry *SubnetRegistryFilterer) ParseScoreUpdaterAdded(log types.Log) (*SubnetRegistryScoreUpdaterAdded, error) {
	event := new(SubnetRegistryScoreUpdaterAdded)
	if err := _SubnetRegistry.contract.UnpackLog(event, "ScoreUpdaterAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetRegistryScoreUpdaterRemovedIterator is returned from FilterScoreUpdaterRemoved and is used to iterate over the raw logs and unpacked data for ScoreUpdaterRemoved events raised by the SubnetRegistry contract.
type SubnetRegistryScoreUpdaterRemovedIterator struct {
	Event *SubnetRegistryScoreUpdaterRemoved // Event containing the contract specifics and raw log

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
func (it *SubnetRegistryScoreUpdaterRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetRegistryScoreUpdaterRemoved)
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
		it.Event = new(SubnetRegistryScoreUpdaterRemoved)
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
func (it *SubnetRegistryScoreUpdaterRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetRegistryScoreUpdaterRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetRegistryScoreUpdaterRemoved represents a ScoreUpdaterRemoved event raised by the SubnetRegistry contract.
type SubnetRegistryScoreUpdaterRemoved struct {
	Updater common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterScoreUpdaterRemoved is a free log retrieval operation binding the contract event 0x54a65354214bb79aabb3de61c730235696f4bac79de915936d30a3bb6eef36dc.
//
// Solidity: event ScoreUpdaterRemoved(address indexed updater)
func (_SubnetRegistry *SubnetRegistryFilterer) FilterScoreUpdaterRemoved(opts *bind.FilterOpts, updater []common.Address) (*SubnetRegistryScoreUpdaterRemovedIterator, error) {

	var updaterRule []interface{}
	for _, updaterItem := range updater {
		updaterRule = append(updaterRule, updaterItem)
	}

	logs, sub, err := _SubnetRegistry.contract.FilterLogs(opts, "ScoreUpdaterRemoved", updaterRule)
	if err != nil {
		return nil, err
	}
	return &SubnetRegistryScoreUpdaterRemovedIterator{contract: _SubnetRegistry.contract, event: "ScoreUpdaterRemoved", logs: logs, sub: sub}, nil
}

// WatchScoreUpdaterRemoved is a free log subscription operation binding the contract event 0x54a65354214bb79aabb3de61c730235696f4bac79de915936d30a3bb6eef36dc.
//
// Solidity: event ScoreUpdaterRemoved(address indexed updater)
func (_SubnetRegistry *SubnetRegistryFilterer) WatchScoreUpdaterRemoved(opts *bind.WatchOpts, sink chan<- *SubnetRegistryScoreUpdaterRemoved, updater []common.Address) (event.Subscription, error) {

	var updaterRule []interface{}
	for _, updaterItem := range updater {
		updaterRule = append(updaterRule, updaterItem)
	}

	logs, sub, err := _SubnetRegistry.contract.WatchLogs(opts, "ScoreUpdaterRemoved", updaterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetRegistryScoreUpdaterRemoved)
				if err := _SubnetRegistry.contract.UnpackLog(event, "ScoreUpdaterRemoved", log); err != nil {
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

// ParseScoreUpdaterRemoved is a log parse operation binding the contract event 0x54a65354214bb79aabb3de61c730235696f4bac79de915936d30a3bb6eef36dc.
//
// Solidity: event ScoreUpdaterRemoved(address indexed updater)
func (_SubnetRegistry *SubnetRegistryFilterer) ParseScoreUpdaterRemoved(log types.Log) (*SubnetRegistryScoreUpdaterRemoved, error) {
	event := new(SubnetRegistryScoreUpdaterRemoved)
	if err := _SubnetRegistry.contract.UnpackLog(event, "ScoreUpdaterRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetRegistrySubnetDeregisteredIterator is returned from FilterSubnetDeregistered and is used to iterate over the raw logs and unpacked data for SubnetDeregistered events raised by the SubnetRegistry contract.
type SubnetRegistrySubnetDeregisteredIterator struct {
	Event *SubnetRegistrySubnetDeregistered // Event containing the contract specifics and raw log

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
func (it *SubnetRegistrySubnetDeregisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetRegistrySubnetDeregistered)
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
		it.Event = new(SubnetRegistrySubnetDeregistered)
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
func (it *SubnetRegistrySubnetDeregisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetRegistrySubnetDeregisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetRegistrySubnetDeregistered represents a SubnetDeregistered event raised by the SubnetRegistry contract.
type SubnetRegistrySubnetDeregistered struct {
	SubnetId *big.Int
	Owner    common.Address
	PeerAddr string
	Uptime   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterSubnetDeregistered is a free log retrieval operation binding the contract event 0x80718449865340bb37807ea694c29d58331a6b00108db801731c5b7e97c99454.
//
// Solidity: event SubnetDeregistered(uint256 indexed subnetId, address indexed owner, string peerAddr, uint256 uptime)
func (_SubnetRegistry *SubnetRegistryFilterer) FilterSubnetDeregistered(opts *bind.FilterOpts, subnetId []*big.Int, owner []common.Address) (*SubnetRegistrySubnetDeregisteredIterator, error) {

	var subnetIdRule []interface{}
	for _, subnetIdItem := range subnetId {
		subnetIdRule = append(subnetIdRule, subnetIdItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _SubnetRegistry.contract.FilterLogs(opts, "SubnetDeregistered", subnetIdRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &SubnetRegistrySubnetDeregisteredIterator{contract: _SubnetRegistry.contract, event: "SubnetDeregistered", logs: logs, sub: sub}, nil
}

// WatchSubnetDeregistered is a free log subscription operation binding the contract event 0x80718449865340bb37807ea694c29d58331a6b00108db801731c5b7e97c99454.
//
// Solidity: event SubnetDeregistered(uint256 indexed subnetId, address indexed owner, string peerAddr, uint256 uptime)
func (_SubnetRegistry *SubnetRegistryFilterer) WatchSubnetDeregistered(opts *bind.WatchOpts, sink chan<- *SubnetRegistrySubnetDeregistered, subnetId []*big.Int, owner []common.Address) (event.Subscription, error) {

	var subnetIdRule []interface{}
	for _, subnetIdItem := range subnetId {
		subnetIdRule = append(subnetIdRule, subnetIdItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _SubnetRegistry.contract.WatchLogs(opts, "SubnetDeregistered", subnetIdRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetRegistrySubnetDeregistered)
				if err := _SubnetRegistry.contract.UnpackLog(event, "SubnetDeregistered", log); err != nil {
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

// ParseSubnetDeregistered is a log parse operation binding the contract event 0x80718449865340bb37807ea694c29d58331a6b00108db801731c5b7e97c99454.
//
// Solidity: event SubnetDeregistered(uint256 indexed subnetId, address indexed owner, string peerAddr, uint256 uptime)
func (_SubnetRegistry *SubnetRegistryFilterer) ParseSubnetDeregistered(log types.Log) (*SubnetRegistrySubnetDeregistered, error) {
	event := new(SubnetRegistrySubnetDeregistered)
	if err := _SubnetRegistry.contract.UnpackLog(event, "SubnetDeregistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetRegistrySubnetRegisteredIterator is returned from FilterSubnetRegistered and is used to iterate over the raw logs and unpacked data for SubnetRegistered events raised by the SubnetRegistry contract.
type SubnetRegistrySubnetRegisteredIterator struct {
	Event *SubnetRegistrySubnetRegistered // Event containing the contract specifics and raw log

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
func (it *SubnetRegistrySubnetRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetRegistrySubnetRegistered)
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
		it.Event = new(SubnetRegistrySubnetRegistered)
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
func (it *SubnetRegistrySubnetRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetRegistrySubnetRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetRegistrySubnetRegistered represents a SubnetRegistered event raised by the SubnetRegistry contract.
type SubnetRegistrySubnetRegistered struct {
	SubnetId *big.Int
	Owner    common.Address
	NftId    *big.Int
	PeerAddr string
	Metadata string
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterSubnetRegistered is a free log retrieval operation binding the contract event 0xdcd3345995bfd9b7d91a439fa7f7c932121c850a92a64932c10cdce85db18316.
//
// Solidity: event SubnetRegistered(uint256 indexed subnetId, address indexed owner, uint256 nftId, string peerAddr, string metadata)
func (_SubnetRegistry *SubnetRegistryFilterer) FilterSubnetRegistered(opts *bind.FilterOpts, subnetId []*big.Int, owner []common.Address) (*SubnetRegistrySubnetRegisteredIterator, error) {

	var subnetIdRule []interface{}
	for _, subnetIdItem := range subnetId {
		subnetIdRule = append(subnetIdRule, subnetIdItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _SubnetRegistry.contract.FilterLogs(opts, "SubnetRegistered", subnetIdRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &SubnetRegistrySubnetRegisteredIterator{contract: _SubnetRegistry.contract, event: "SubnetRegistered", logs: logs, sub: sub}, nil
}

// WatchSubnetRegistered is a free log subscription operation binding the contract event 0xdcd3345995bfd9b7d91a439fa7f7c932121c850a92a64932c10cdce85db18316.
//
// Solidity: event SubnetRegistered(uint256 indexed subnetId, address indexed owner, uint256 nftId, string peerAddr, string metadata)
func (_SubnetRegistry *SubnetRegistryFilterer) WatchSubnetRegistered(opts *bind.WatchOpts, sink chan<- *SubnetRegistrySubnetRegistered, subnetId []*big.Int, owner []common.Address) (event.Subscription, error) {

	var subnetIdRule []interface{}
	for _, subnetIdItem := range subnetId {
		subnetIdRule = append(subnetIdRule, subnetIdItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _SubnetRegistry.contract.WatchLogs(opts, "SubnetRegistered", subnetIdRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetRegistrySubnetRegistered)
				if err := _SubnetRegistry.contract.UnpackLog(event, "SubnetRegistered", log); err != nil {
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

// ParseSubnetRegistered is a log parse operation binding the contract event 0xdcd3345995bfd9b7d91a439fa7f7c932121c850a92a64932c10cdce85db18316.
//
// Solidity: event SubnetRegistered(uint256 indexed subnetId, address indexed owner, uint256 nftId, string peerAddr, string metadata)
func (_SubnetRegistry *SubnetRegistryFilterer) ParseSubnetRegistered(log types.Log) (*SubnetRegistrySubnetRegistered, error) {
	event := new(SubnetRegistrySubnetRegistered)
	if err := _SubnetRegistry.contract.UnpackLog(event, "SubnetRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetRegistryTrustScoreUpdatedIterator is returned from FilterTrustScoreUpdated and is used to iterate over the raw logs and unpacked data for TrustScoreUpdated events raised by the SubnetRegistry contract.
type SubnetRegistryTrustScoreUpdatedIterator struct {
	Event *SubnetRegistryTrustScoreUpdated // Event containing the contract specifics and raw log

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
func (it *SubnetRegistryTrustScoreUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetRegistryTrustScoreUpdated)
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
		it.Event = new(SubnetRegistryTrustScoreUpdated)
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
func (it *SubnetRegistryTrustScoreUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetRegistryTrustScoreUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetRegistryTrustScoreUpdated represents a TrustScoreUpdated event raised by the SubnetRegistry contract.
type SubnetRegistryTrustScoreUpdated struct {
	SubnetId *big.Int
	NewScore *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterTrustScoreUpdated is a free log retrieval operation binding the contract event 0x20d456995c74d5114f87879b10c5214b9c28e11ea422eeb922cc2e755095669f.
//
// Solidity: event TrustScoreUpdated(uint256 indexed subnetId, uint256 newScore)
func (_SubnetRegistry *SubnetRegistryFilterer) FilterTrustScoreUpdated(opts *bind.FilterOpts, subnetId []*big.Int) (*SubnetRegistryTrustScoreUpdatedIterator, error) {

	var subnetIdRule []interface{}
	for _, subnetIdItem := range subnetId {
		subnetIdRule = append(subnetIdRule, subnetIdItem)
	}

	logs, sub, err := _SubnetRegistry.contract.FilterLogs(opts, "TrustScoreUpdated", subnetIdRule)
	if err != nil {
		return nil, err
	}
	return &SubnetRegistryTrustScoreUpdatedIterator{contract: _SubnetRegistry.contract, event: "TrustScoreUpdated", logs: logs, sub: sub}, nil
}

// WatchTrustScoreUpdated is a free log subscription operation binding the contract event 0x20d456995c74d5114f87879b10c5214b9c28e11ea422eeb922cc2e755095669f.
//
// Solidity: event TrustScoreUpdated(uint256 indexed subnetId, uint256 newScore)
func (_SubnetRegistry *SubnetRegistryFilterer) WatchTrustScoreUpdated(opts *bind.WatchOpts, sink chan<- *SubnetRegistryTrustScoreUpdated, subnetId []*big.Int) (event.Subscription, error) {

	var subnetIdRule []interface{}
	for _, subnetIdItem := range subnetId {
		subnetIdRule = append(subnetIdRule, subnetIdItem)
	}

	logs, sub, err := _SubnetRegistry.contract.WatchLogs(opts, "TrustScoreUpdated", subnetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetRegistryTrustScoreUpdated)
				if err := _SubnetRegistry.contract.UnpackLog(event, "TrustScoreUpdated", log); err != nil {
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

// ParseTrustScoreUpdated is a log parse operation binding the contract event 0x20d456995c74d5114f87879b10c5214b9c28e11ea422eeb922cc2e755095669f.
//
// Solidity: event TrustScoreUpdated(uint256 indexed subnetId, uint256 newScore)
func (_SubnetRegistry *SubnetRegistryFilterer) ParseTrustScoreUpdated(log types.Log) (*SubnetRegistryTrustScoreUpdated, error) {
	event := new(SubnetRegistryTrustScoreUpdated)
	if err := _SubnetRegistry.contract.UnpackLog(event, "TrustScoreUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
