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

// SubnetProviderUptimeMetaData contains all meta data concerning the SubnetProviderUptime contract.
var SubnetProviderUptimeMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_subnetProvider\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_rewardToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_rewardPerSecond\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"_verifierPeerId\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"_operator\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"SafeERC20FailedOperation\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"oldMerkleRoot\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"newMerkleRoot\",\"type\":\"bytes32\"}],\"name\":\"MerkleRootUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"oldOperator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newOperator\",\"type\":\"address\"}],\"name\":\"OperatorUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"RewardClaimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"oldRewardPerSecond\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newRewardPerSecond\",\"type\":\"uint256\"}],\"name\":\"RewardPerSecondUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"pendingReward\",\"type\":\"uint256\"}],\"name\":\"RewardReported\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"totalUptime\",\"type\":\"uint256\"}],\"name\":\"UptimeUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"oldVerifierPeerId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"newVerifierPeerId\",\"type\":\"string\"}],\"name\":\"VerifierPeerIdUpdated\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"claimReward\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"depositRewards\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getTotalUptime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"merkleRoot\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"operator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalUptime\",\"type\":\"uint256\"},{\"internalType\":\"bytes32[]\",\"name\":\"proof\",\"type\":\"bytes32[]\"}],\"name\":\"reportUptime\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"rewardPerSecond\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"rewardToken\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"subnetProvider\",\"outputs\":[{\"internalType\":\"contractSubnetProvider\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_merkleRoot\",\"type\":\"bytes32\"}],\"name\":\"updateMerkleRoot\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_operator\",\"type\":\"address\"}],\"name\":\"updateOperator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_rewardPerSecond\",\"type\":\"uint256\"}],\"name\":\"updateRewardPerSecond\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_verifierPeerId\",\"type\":\"string\"}],\"name\":\"updateVerifierPeerId\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"uptimes\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"totalUptime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lastUpdate\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"claimedUptime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pendingReward\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lastClaimTime\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"verifierPeerId\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// SubnetProviderUptimeABI is the input ABI used to generate the binding from.
// Deprecated: Use SubnetProviderUptimeMetaData.ABI instead.
var SubnetProviderUptimeABI = SubnetProviderUptimeMetaData.ABI

// SubnetProviderUptime is an auto generated Go binding around an Ethereum contract.
type SubnetProviderUptime struct {
	SubnetProviderUptimeCaller     // Read-only binding to the contract
	SubnetProviderUptimeTransactor // Write-only binding to the contract
	SubnetProviderUptimeFilterer   // Log filterer for contract events
}

// SubnetProviderUptimeCaller is an auto generated read-only Go binding around an Ethereum contract.
type SubnetProviderUptimeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetProviderUptimeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SubnetProviderUptimeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetProviderUptimeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SubnetProviderUptimeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetProviderUptimeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SubnetProviderUptimeSession struct {
	Contract     *SubnetProviderUptime // Generic contract binding to set the session for
	CallOpts     bind.CallOpts         // Call options to use throughout this session
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// SubnetProviderUptimeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SubnetProviderUptimeCallerSession struct {
	Contract *SubnetProviderUptimeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts               // Call options to use throughout this session
}

// SubnetProviderUptimeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SubnetProviderUptimeTransactorSession struct {
	Contract     *SubnetProviderUptimeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts               // Transaction auth options to use throughout this session
}

// SubnetProviderUptimeRaw is an auto generated low-level Go binding around an Ethereum contract.
type SubnetProviderUptimeRaw struct {
	Contract *SubnetProviderUptime // Generic contract binding to access the raw methods on
}

// SubnetProviderUptimeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SubnetProviderUptimeCallerRaw struct {
	Contract *SubnetProviderUptimeCaller // Generic read-only contract binding to access the raw methods on
}

// SubnetProviderUptimeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SubnetProviderUptimeTransactorRaw struct {
	Contract *SubnetProviderUptimeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSubnetProviderUptime creates a new instance of SubnetProviderUptime, bound to a specific deployed contract.
func NewSubnetProviderUptime(address common.Address, backend bind.ContractBackend) (*SubnetProviderUptime, error) {
	contract, err := bindSubnetProviderUptime(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SubnetProviderUptime{SubnetProviderUptimeCaller: SubnetProviderUptimeCaller{contract: contract}, SubnetProviderUptimeTransactor: SubnetProviderUptimeTransactor{contract: contract}, SubnetProviderUptimeFilterer: SubnetProviderUptimeFilterer{contract: contract}}, nil
}

// NewSubnetProviderUptimeCaller creates a new read-only instance of SubnetProviderUptime, bound to a specific deployed contract.
func NewSubnetProviderUptimeCaller(address common.Address, caller bind.ContractCaller) (*SubnetProviderUptimeCaller, error) {
	contract, err := bindSubnetProviderUptime(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SubnetProviderUptimeCaller{contract: contract}, nil
}

// NewSubnetProviderUptimeTransactor creates a new write-only instance of SubnetProviderUptime, bound to a specific deployed contract.
func NewSubnetProviderUptimeTransactor(address common.Address, transactor bind.ContractTransactor) (*SubnetProviderUptimeTransactor, error) {
	contract, err := bindSubnetProviderUptime(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SubnetProviderUptimeTransactor{contract: contract}, nil
}

// NewSubnetProviderUptimeFilterer creates a new log filterer instance of SubnetProviderUptime, bound to a specific deployed contract.
func NewSubnetProviderUptimeFilterer(address common.Address, filterer bind.ContractFilterer) (*SubnetProviderUptimeFilterer, error) {
	contract, err := bindSubnetProviderUptime(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SubnetProviderUptimeFilterer{contract: contract}, nil
}

// bindSubnetProviderUptime binds a generic wrapper to an already deployed contract.
func bindSubnetProviderUptime(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SubnetProviderUptimeMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SubnetProviderUptime *SubnetProviderUptimeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SubnetProviderUptime.Contract.SubnetProviderUptimeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SubnetProviderUptime *SubnetProviderUptimeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.SubnetProviderUptimeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SubnetProviderUptime *SubnetProviderUptimeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.SubnetProviderUptimeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SubnetProviderUptime *SubnetProviderUptimeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SubnetProviderUptime.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SubnetProviderUptime *SubnetProviderUptimeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SubnetProviderUptime *SubnetProviderUptimeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.contract.Transact(opts, method, params...)
}

// GetTotalUptime is a free data retrieval call binding the contract method 0x8ad84945.
//
// Solidity: function getTotalUptime(uint256 tokenId) view returns(uint256)
func (_SubnetProviderUptime *SubnetProviderUptimeCaller) GetTotalUptime(opts *bind.CallOpts, tokenId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SubnetProviderUptime.contract.Call(opts, &out, "getTotalUptime", tokenId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalUptime is a free data retrieval call binding the contract method 0x8ad84945.
//
// Solidity: function getTotalUptime(uint256 tokenId) view returns(uint256)
func (_SubnetProviderUptime *SubnetProviderUptimeSession) GetTotalUptime(tokenId *big.Int) (*big.Int, error) {
	return _SubnetProviderUptime.Contract.GetTotalUptime(&_SubnetProviderUptime.CallOpts, tokenId)
}

// GetTotalUptime is a free data retrieval call binding the contract method 0x8ad84945.
//
// Solidity: function getTotalUptime(uint256 tokenId) view returns(uint256)
func (_SubnetProviderUptime *SubnetProviderUptimeCallerSession) GetTotalUptime(tokenId *big.Int) (*big.Int, error) {
	return _SubnetProviderUptime.Contract.GetTotalUptime(&_SubnetProviderUptime.CallOpts, tokenId)
}

// MerkleRoot is a free data retrieval call binding the contract method 0x2eb4a7ab.
//
// Solidity: function merkleRoot() view returns(bytes32)
func (_SubnetProviderUptime *SubnetProviderUptimeCaller) MerkleRoot(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SubnetProviderUptime.contract.Call(opts, &out, "merkleRoot")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MerkleRoot is a free data retrieval call binding the contract method 0x2eb4a7ab.
//
// Solidity: function merkleRoot() view returns(bytes32)
func (_SubnetProviderUptime *SubnetProviderUptimeSession) MerkleRoot() ([32]byte, error) {
	return _SubnetProviderUptime.Contract.MerkleRoot(&_SubnetProviderUptime.CallOpts)
}

// MerkleRoot is a free data retrieval call binding the contract method 0x2eb4a7ab.
//
// Solidity: function merkleRoot() view returns(bytes32)
func (_SubnetProviderUptime *SubnetProviderUptimeCallerSession) MerkleRoot() ([32]byte, error) {
	return _SubnetProviderUptime.Contract.MerkleRoot(&_SubnetProviderUptime.CallOpts)
}

// Operator is a free data retrieval call binding the contract method 0x570ca735.
//
// Solidity: function operator() view returns(address)
func (_SubnetProviderUptime *SubnetProviderUptimeCaller) Operator(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetProviderUptime.contract.Call(opts, &out, "operator")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Operator is a free data retrieval call binding the contract method 0x570ca735.
//
// Solidity: function operator() view returns(address)
func (_SubnetProviderUptime *SubnetProviderUptimeSession) Operator() (common.Address, error) {
	return _SubnetProviderUptime.Contract.Operator(&_SubnetProviderUptime.CallOpts)
}

// Operator is a free data retrieval call binding the contract method 0x570ca735.
//
// Solidity: function operator() view returns(address)
func (_SubnetProviderUptime *SubnetProviderUptimeCallerSession) Operator() (common.Address, error) {
	return _SubnetProviderUptime.Contract.Operator(&_SubnetProviderUptime.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetProviderUptime *SubnetProviderUptimeCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetProviderUptime.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetProviderUptime *SubnetProviderUptimeSession) Owner() (common.Address, error) {
	return _SubnetProviderUptime.Contract.Owner(&_SubnetProviderUptime.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetProviderUptime *SubnetProviderUptimeCallerSession) Owner() (common.Address, error) {
	return _SubnetProviderUptime.Contract.Owner(&_SubnetProviderUptime.CallOpts)
}

// RewardPerSecond is a free data retrieval call binding the contract method 0x8f10369a.
//
// Solidity: function rewardPerSecond() view returns(uint256)
func (_SubnetProviderUptime *SubnetProviderUptimeCaller) RewardPerSecond(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SubnetProviderUptime.contract.Call(opts, &out, "rewardPerSecond")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RewardPerSecond is a free data retrieval call binding the contract method 0x8f10369a.
//
// Solidity: function rewardPerSecond() view returns(uint256)
func (_SubnetProviderUptime *SubnetProviderUptimeSession) RewardPerSecond() (*big.Int, error) {
	return _SubnetProviderUptime.Contract.RewardPerSecond(&_SubnetProviderUptime.CallOpts)
}

// RewardPerSecond is a free data retrieval call binding the contract method 0x8f10369a.
//
// Solidity: function rewardPerSecond() view returns(uint256)
func (_SubnetProviderUptime *SubnetProviderUptimeCallerSession) RewardPerSecond() (*big.Int, error) {
	return _SubnetProviderUptime.Contract.RewardPerSecond(&_SubnetProviderUptime.CallOpts)
}

// RewardToken is a free data retrieval call binding the contract method 0xf7c618c1.
//
// Solidity: function rewardToken() view returns(address)
func (_SubnetProviderUptime *SubnetProviderUptimeCaller) RewardToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetProviderUptime.contract.Call(opts, &out, "rewardToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RewardToken is a free data retrieval call binding the contract method 0xf7c618c1.
//
// Solidity: function rewardToken() view returns(address)
func (_SubnetProviderUptime *SubnetProviderUptimeSession) RewardToken() (common.Address, error) {
	return _SubnetProviderUptime.Contract.RewardToken(&_SubnetProviderUptime.CallOpts)
}

// RewardToken is a free data retrieval call binding the contract method 0xf7c618c1.
//
// Solidity: function rewardToken() view returns(address)
func (_SubnetProviderUptime *SubnetProviderUptimeCallerSession) RewardToken() (common.Address, error) {
	return _SubnetProviderUptime.Contract.RewardToken(&_SubnetProviderUptime.CallOpts)
}

// SubnetProvider is a free data retrieval call binding the contract method 0x6a46dc8c.
//
// Solidity: function subnetProvider() view returns(address)
func (_SubnetProviderUptime *SubnetProviderUptimeCaller) SubnetProvider(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetProviderUptime.contract.Call(opts, &out, "subnetProvider")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SubnetProvider is a free data retrieval call binding the contract method 0x6a46dc8c.
//
// Solidity: function subnetProvider() view returns(address)
func (_SubnetProviderUptime *SubnetProviderUptimeSession) SubnetProvider() (common.Address, error) {
	return _SubnetProviderUptime.Contract.SubnetProvider(&_SubnetProviderUptime.CallOpts)
}

// SubnetProvider is a free data retrieval call binding the contract method 0x6a46dc8c.
//
// Solidity: function subnetProvider() view returns(address)
func (_SubnetProviderUptime *SubnetProviderUptimeCallerSession) SubnetProvider() (common.Address, error) {
	return _SubnetProviderUptime.Contract.SubnetProvider(&_SubnetProviderUptime.CallOpts)
}

// Uptimes is a free data retrieval call binding the contract method 0x663e4a70.
//
// Solidity: function uptimes(uint256 ) view returns(uint256 totalUptime, uint256 lastUpdate, uint256 claimedUptime, uint256 pendingReward, uint256 lastClaimTime)
func (_SubnetProviderUptime *SubnetProviderUptimeCaller) Uptimes(opts *bind.CallOpts, arg0 *big.Int) (struct {
	TotalUptime   *big.Int
	LastUpdate    *big.Int
	ClaimedUptime *big.Int
	PendingReward *big.Int
	LastClaimTime *big.Int
}, error) {
	var out []interface{}
	err := _SubnetProviderUptime.contract.Call(opts, &out, "uptimes", arg0)

	outstruct := new(struct {
		TotalUptime   *big.Int
		LastUpdate    *big.Int
		ClaimedUptime *big.Int
		PendingReward *big.Int
		LastClaimTime *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.TotalUptime = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.LastUpdate = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.ClaimedUptime = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.PendingReward = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.LastClaimTime = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Uptimes is a free data retrieval call binding the contract method 0x663e4a70.
//
// Solidity: function uptimes(uint256 ) view returns(uint256 totalUptime, uint256 lastUpdate, uint256 claimedUptime, uint256 pendingReward, uint256 lastClaimTime)
func (_SubnetProviderUptime *SubnetProviderUptimeSession) Uptimes(arg0 *big.Int) (struct {
	TotalUptime   *big.Int
	LastUpdate    *big.Int
	ClaimedUptime *big.Int
	PendingReward *big.Int
	LastClaimTime *big.Int
}, error) {
	return _SubnetProviderUptime.Contract.Uptimes(&_SubnetProviderUptime.CallOpts, arg0)
}

// Uptimes is a free data retrieval call binding the contract method 0x663e4a70.
//
// Solidity: function uptimes(uint256 ) view returns(uint256 totalUptime, uint256 lastUpdate, uint256 claimedUptime, uint256 pendingReward, uint256 lastClaimTime)
func (_SubnetProviderUptime *SubnetProviderUptimeCallerSession) Uptimes(arg0 *big.Int) (struct {
	TotalUptime   *big.Int
	LastUpdate    *big.Int
	ClaimedUptime *big.Int
	PendingReward *big.Int
	LastClaimTime *big.Int
}, error) {
	return _SubnetProviderUptime.Contract.Uptimes(&_SubnetProviderUptime.CallOpts, arg0)
}

// VerifierPeerId is a free data retrieval call binding the contract method 0x58f81e3a.
//
// Solidity: function verifierPeerId() view returns(string)
func (_SubnetProviderUptime *SubnetProviderUptimeCaller) VerifierPeerId(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _SubnetProviderUptime.contract.Call(opts, &out, "verifierPeerId")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// VerifierPeerId is a free data retrieval call binding the contract method 0x58f81e3a.
//
// Solidity: function verifierPeerId() view returns(string)
func (_SubnetProviderUptime *SubnetProviderUptimeSession) VerifierPeerId() (string, error) {
	return _SubnetProviderUptime.Contract.VerifierPeerId(&_SubnetProviderUptime.CallOpts)
}

// VerifierPeerId is a free data retrieval call binding the contract method 0x58f81e3a.
//
// Solidity: function verifierPeerId() view returns(string)
func (_SubnetProviderUptime *SubnetProviderUptimeCallerSession) VerifierPeerId() (string, error) {
	return _SubnetProviderUptime.Contract.VerifierPeerId(&_SubnetProviderUptime.CallOpts)
}

// ClaimReward is a paid mutator transaction binding the contract method 0xae169a50.
//
// Solidity: function claimReward(uint256 tokenId) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeTransactor) ClaimReward(opts *bind.TransactOpts, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetProviderUptime.contract.Transact(opts, "claimReward", tokenId)
}

// ClaimReward is a paid mutator transaction binding the contract method 0xae169a50.
//
// Solidity: function claimReward(uint256 tokenId) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeSession) ClaimReward(tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.ClaimReward(&_SubnetProviderUptime.TransactOpts, tokenId)
}

// ClaimReward is a paid mutator transaction binding the contract method 0xae169a50.
//
// Solidity: function claimReward(uint256 tokenId) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeTransactorSession) ClaimReward(tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.ClaimReward(&_SubnetProviderUptime.TransactOpts, tokenId)
}

// DepositRewards is a paid mutator transaction binding the contract method 0x8bdf67f2.
//
// Solidity: function depositRewards(uint256 amount) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeTransactor) DepositRewards(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _SubnetProviderUptime.contract.Transact(opts, "depositRewards", amount)
}

// DepositRewards is a paid mutator transaction binding the contract method 0x8bdf67f2.
//
// Solidity: function depositRewards(uint256 amount) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeSession) DepositRewards(amount *big.Int) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.DepositRewards(&_SubnetProviderUptime.TransactOpts, amount)
}

// DepositRewards is a paid mutator transaction binding the contract method 0x8bdf67f2.
//
// Solidity: function depositRewards(uint256 amount) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeTransactorSession) DepositRewards(amount *big.Int) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.DepositRewards(&_SubnetProviderUptime.TransactOpts, amount)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetProviderUptime *SubnetProviderUptimeTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetProviderUptime.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetProviderUptime *SubnetProviderUptimeSession) RenounceOwnership() (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.RenounceOwnership(&_SubnetProviderUptime.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetProviderUptime *SubnetProviderUptimeTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.RenounceOwnership(&_SubnetProviderUptime.TransactOpts)
}

// ReportUptime is a paid mutator transaction binding the contract method 0xcb999ab5.
//
// Solidity: function reportUptime(uint256 tokenId, uint256 totalUptime, bytes32[] proof) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeTransactor) ReportUptime(opts *bind.TransactOpts, tokenId *big.Int, totalUptime *big.Int, proof [][32]byte) (*types.Transaction, error) {
	return _SubnetProviderUptime.contract.Transact(opts, "reportUptime", tokenId, totalUptime, proof)
}

// ReportUptime is a paid mutator transaction binding the contract method 0xcb999ab5.
//
// Solidity: function reportUptime(uint256 tokenId, uint256 totalUptime, bytes32[] proof) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeSession) ReportUptime(tokenId *big.Int, totalUptime *big.Int, proof [][32]byte) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.ReportUptime(&_SubnetProviderUptime.TransactOpts, tokenId, totalUptime, proof)
}

// ReportUptime is a paid mutator transaction binding the contract method 0xcb999ab5.
//
// Solidity: function reportUptime(uint256 tokenId, uint256 totalUptime, bytes32[] proof) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeTransactorSession) ReportUptime(tokenId *big.Int, totalUptime *big.Int, proof [][32]byte) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.ReportUptime(&_SubnetProviderUptime.TransactOpts, tokenId, totalUptime, proof)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _SubnetProviderUptime.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.TransferOwnership(&_SubnetProviderUptime.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.TransferOwnership(&_SubnetProviderUptime.TransactOpts, newOwner)
}

// UpdateMerkleRoot is a paid mutator transaction binding the contract method 0x4783f0ef.
//
// Solidity: function updateMerkleRoot(bytes32 _merkleRoot) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeTransactor) UpdateMerkleRoot(opts *bind.TransactOpts, _merkleRoot [32]byte) (*types.Transaction, error) {
	return _SubnetProviderUptime.contract.Transact(opts, "updateMerkleRoot", _merkleRoot)
}

// UpdateMerkleRoot is a paid mutator transaction binding the contract method 0x4783f0ef.
//
// Solidity: function updateMerkleRoot(bytes32 _merkleRoot) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeSession) UpdateMerkleRoot(_merkleRoot [32]byte) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.UpdateMerkleRoot(&_SubnetProviderUptime.TransactOpts, _merkleRoot)
}

// UpdateMerkleRoot is a paid mutator transaction binding the contract method 0x4783f0ef.
//
// Solidity: function updateMerkleRoot(bytes32 _merkleRoot) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeTransactorSession) UpdateMerkleRoot(_merkleRoot [32]byte) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.UpdateMerkleRoot(&_SubnetProviderUptime.TransactOpts, _merkleRoot)
}

// UpdateOperator is a paid mutator transaction binding the contract method 0xac7475ed.
//
// Solidity: function updateOperator(address _operator) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeTransactor) UpdateOperator(opts *bind.TransactOpts, _operator common.Address) (*types.Transaction, error) {
	return _SubnetProviderUptime.contract.Transact(opts, "updateOperator", _operator)
}

// UpdateOperator is a paid mutator transaction binding the contract method 0xac7475ed.
//
// Solidity: function updateOperator(address _operator) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeSession) UpdateOperator(_operator common.Address) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.UpdateOperator(&_SubnetProviderUptime.TransactOpts, _operator)
}

// UpdateOperator is a paid mutator transaction binding the contract method 0xac7475ed.
//
// Solidity: function updateOperator(address _operator) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeTransactorSession) UpdateOperator(_operator common.Address) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.UpdateOperator(&_SubnetProviderUptime.TransactOpts, _operator)
}

// UpdateRewardPerSecond is a paid mutator transaction binding the contract method 0x4004c8e7.
//
// Solidity: function updateRewardPerSecond(uint256 _rewardPerSecond) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeTransactor) UpdateRewardPerSecond(opts *bind.TransactOpts, _rewardPerSecond *big.Int) (*types.Transaction, error) {
	return _SubnetProviderUptime.contract.Transact(opts, "updateRewardPerSecond", _rewardPerSecond)
}

// UpdateRewardPerSecond is a paid mutator transaction binding the contract method 0x4004c8e7.
//
// Solidity: function updateRewardPerSecond(uint256 _rewardPerSecond) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeSession) UpdateRewardPerSecond(_rewardPerSecond *big.Int) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.UpdateRewardPerSecond(&_SubnetProviderUptime.TransactOpts, _rewardPerSecond)
}

// UpdateRewardPerSecond is a paid mutator transaction binding the contract method 0x4004c8e7.
//
// Solidity: function updateRewardPerSecond(uint256 _rewardPerSecond) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeTransactorSession) UpdateRewardPerSecond(_rewardPerSecond *big.Int) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.UpdateRewardPerSecond(&_SubnetProviderUptime.TransactOpts, _rewardPerSecond)
}

// UpdateVerifierPeerId is a paid mutator transaction binding the contract method 0xd14da7ad.
//
// Solidity: function updateVerifierPeerId(string _verifierPeerId) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeTransactor) UpdateVerifierPeerId(opts *bind.TransactOpts, _verifierPeerId string) (*types.Transaction, error) {
	return _SubnetProviderUptime.contract.Transact(opts, "updateVerifierPeerId", _verifierPeerId)
}

// UpdateVerifierPeerId is a paid mutator transaction binding the contract method 0xd14da7ad.
//
// Solidity: function updateVerifierPeerId(string _verifierPeerId) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeSession) UpdateVerifierPeerId(_verifierPeerId string) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.UpdateVerifierPeerId(&_SubnetProviderUptime.TransactOpts, _verifierPeerId)
}

// UpdateVerifierPeerId is a paid mutator transaction binding the contract method 0xd14da7ad.
//
// Solidity: function updateVerifierPeerId(string _verifierPeerId) returns()
func (_SubnetProviderUptime *SubnetProviderUptimeTransactorSession) UpdateVerifierPeerId(_verifierPeerId string) (*types.Transaction, error) {
	return _SubnetProviderUptime.Contract.UpdateVerifierPeerId(&_SubnetProviderUptime.TransactOpts, _verifierPeerId)
}

// SubnetProviderUptimeMerkleRootUpdatedIterator is returned from FilterMerkleRootUpdated and is used to iterate over the raw logs and unpacked data for MerkleRootUpdated events raised by the SubnetProviderUptime contract.
type SubnetProviderUptimeMerkleRootUpdatedIterator struct {
	Event *SubnetProviderUptimeMerkleRootUpdated // Event containing the contract specifics and raw log

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
func (it *SubnetProviderUptimeMerkleRootUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderUptimeMerkleRootUpdated)
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
		it.Event = new(SubnetProviderUptimeMerkleRootUpdated)
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
func (it *SubnetProviderUptimeMerkleRootUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderUptimeMerkleRootUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderUptimeMerkleRootUpdated represents a MerkleRootUpdated event raised by the SubnetProviderUptime contract.
type SubnetProviderUptimeMerkleRootUpdated struct {
	OldMerkleRoot [32]byte
	NewMerkleRoot [32]byte
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterMerkleRootUpdated is a free log retrieval operation binding the contract event 0xfd69edeceaf1d6832d935be1fba54ca93bf17e71520c6c9ffc08d6e9529f8757.
//
// Solidity: event MerkleRootUpdated(bytes32 oldMerkleRoot, bytes32 newMerkleRoot)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) FilterMerkleRootUpdated(opts *bind.FilterOpts) (*SubnetProviderUptimeMerkleRootUpdatedIterator, error) {

	logs, sub, err := _SubnetProviderUptime.contract.FilterLogs(opts, "MerkleRootUpdated")
	if err != nil {
		return nil, err
	}
	return &SubnetProviderUptimeMerkleRootUpdatedIterator{contract: _SubnetProviderUptime.contract, event: "MerkleRootUpdated", logs: logs, sub: sub}, nil
}

// WatchMerkleRootUpdated is a free log subscription operation binding the contract event 0xfd69edeceaf1d6832d935be1fba54ca93bf17e71520c6c9ffc08d6e9529f8757.
//
// Solidity: event MerkleRootUpdated(bytes32 oldMerkleRoot, bytes32 newMerkleRoot)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) WatchMerkleRootUpdated(opts *bind.WatchOpts, sink chan<- *SubnetProviderUptimeMerkleRootUpdated) (event.Subscription, error) {

	logs, sub, err := _SubnetProviderUptime.contract.WatchLogs(opts, "MerkleRootUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderUptimeMerkleRootUpdated)
				if err := _SubnetProviderUptime.contract.UnpackLog(event, "MerkleRootUpdated", log); err != nil {
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

// ParseMerkleRootUpdated is a log parse operation binding the contract event 0xfd69edeceaf1d6832d935be1fba54ca93bf17e71520c6c9ffc08d6e9529f8757.
//
// Solidity: event MerkleRootUpdated(bytes32 oldMerkleRoot, bytes32 newMerkleRoot)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) ParseMerkleRootUpdated(log types.Log) (*SubnetProviderUptimeMerkleRootUpdated, error) {
	event := new(SubnetProviderUptimeMerkleRootUpdated)
	if err := _SubnetProviderUptime.contract.UnpackLog(event, "MerkleRootUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetProviderUptimeOperatorUpdatedIterator is returned from FilterOperatorUpdated and is used to iterate over the raw logs and unpacked data for OperatorUpdated events raised by the SubnetProviderUptime contract.
type SubnetProviderUptimeOperatorUpdatedIterator struct {
	Event *SubnetProviderUptimeOperatorUpdated // Event containing the contract specifics and raw log

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
func (it *SubnetProviderUptimeOperatorUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderUptimeOperatorUpdated)
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
		it.Event = new(SubnetProviderUptimeOperatorUpdated)
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
func (it *SubnetProviderUptimeOperatorUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderUptimeOperatorUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderUptimeOperatorUpdated represents a OperatorUpdated event raised by the SubnetProviderUptime contract.
type SubnetProviderUptimeOperatorUpdated struct {
	OldOperator common.Address
	NewOperator common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterOperatorUpdated is a free log retrieval operation binding the contract event 0xfbe5b6cbafb274f445d7fed869dc77a838d8243a22c460de156560e8857cad03.
//
// Solidity: event OperatorUpdated(address oldOperator, address newOperator)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) FilterOperatorUpdated(opts *bind.FilterOpts) (*SubnetProviderUptimeOperatorUpdatedIterator, error) {

	logs, sub, err := _SubnetProviderUptime.contract.FilterLogs(opts, "OperatorUpdated")
	if err != nil {
		return nil, err
	}
	return &SubnetProviderUptimeOperatorUpdatedIterator{contract: _SubnetProviderUptime.contract, event: "OperatorUpdated", logs: logs, sub: sub}, nil
}

// WatchOperatorUpdated is a free log subscription operation binding the contract event 0xfbe5b6cbafb274f445d7fed869dc77a838d8243a22c460de156560e8857cad03.
//
// Solidity: event OperatorUpdated(address oldOperator, address newOperator)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) WatchOperatorUpdated(opts *bind.WatchOpts, sink chan<- *SubnetProviderUptimeOperatorUpdated) (event.Subscription, error) {

	logs, sub, err := _SubnetProviderUptime.contract.WatchLogs(opts, "OperatorUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderUptimeOperatorUpdated)
				if err := _SubnetProviderUptime.contract.UnpackLog(event, "OperatorUpdated", log); err != nil {
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

// ParseOperatorUpdated is a log parse operation binding the contract event 0xfbe5b6cbafb274f445d7fed869dc77a838d8243a22c460de156560e8857cad03.
//
// Solidity: event OperatorUpdated(address oldOperator, address newOperator)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) ParseOperatorUpdated(log types.Log) (*SubnetProviderUptimeOperatorUpdated, error) {
	event := new(SubnetProviderUptimeOperatorUpdated)
	if err := _SubnetProviderUptime.contract.UnpackLog(event, "OperatorUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetProviderUptimeOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the SubnetProviderUptime contract.
type SubnetProviderUptimeOwnershipTransferredIterator struct {
	Event *SubnetProviderUptimeOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *SubnetProviderUptimeOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderUptimeOwnershipTransferred)
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
		it.Event = new(SubnetProviderUptimeOwnershipTransferred)
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
func (it *SubnetProviderUptimeOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderUptimeOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderUptimeOwnershipTransferred represents a OwnershipTransferred event raised by the SubnetProviderUptime contract.
type SubnetProviderUptimeOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*SubnetProviderUptimeOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SubnetProviderUptime.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &SubnetProviderUptimeOwnershipTransferredIterator{contract: _SubnetProviderUptime.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *SubnetProviderUptimeOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SubnetProviderUptime.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderUptimeOwnershipTransferred)
				if err := _SubnetProviderUptime.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) ParseOwnershipTransferred(log types.Log) (*SubnetProviderUptimeOwnershipTransferred, error) {
	event := new(SubnetProviderUptimeOwnershipTransferred)
	if err := _SubnetProviderUptime.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetProviderUptimeRewardClaimedIterator is returned from FilterRewardClaimed and is used to iterate over the raw logs and unpacked data for RewardClaimed events raised by the SubnetProviderUptime contract.
type SubnetProviderUptimeRewardClaimedIterator struct {
	Event *SubnetProviderUptimeRewardClaimed // Event containing the contract specifics and raw log

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
func (it *SubnetProviderUptimeRewardClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderUptimeRewardClaimed)
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
		it.Event = new(SubnetProviderUptimeRewardClaimed)
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
func (it *SubnetProviderUptimeRewardClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderUptimeRewardClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderUptimeRewardClaimed represents a RewardClaimed event raised by the SubnetProviderUptime contract.
type SubnetProviderUptimeRewardClaimed struct {
	TokenId *big.Int
	Owner   common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRewardClaimed is a free log retrieval operation binding the contract event 0x24b5efa61dd1cfc659205a97fb8ed868f3cb8c81922bab2b96423e5de1de2cb7.
//
// Solidity: event RewardClaimed(uint256 indexed tokenId, address indexed owner, uint256 amount)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) FilterRewardClaimed(opts *bind.FilterOpts, tokenId []*big.Int, owner []common.Address) (*SubnetProviderUptimeRewardClaimedIterator, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _SubnetProviderUptime.contract.FilterLogs(opts, "RewardClaimed", tokenIdRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &SubnetProviderUptimeRewardClaimedIterator{contract: _SubnetProviderUptime.contract, event: "RewardClaimed", logs: logs, sub: sub}, nil
}

// WatchRewardClaimed is a free log subscription operation binding the contract event 0x24b5efa61dd1cfc659205a97fb8ed868f3cb8c81922bab2b96423e5de1de2cb7.
//
// Solidity: event RewardClaimed(uint256 indexed tokenId, address indexed owner, uint256 amount)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) WatchRewardClaimed(opts *bind.WatchOpts, sink chan<- *SubnetProviderUptimeRewardClaimed, tokenId []*big.Int, owner []common.Address) (event.Subscription, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _SubnetProviderUptime.contract.WatchLogs(opts, "RewardClaimed", tokenIdRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderUptimeRewardClaimed)
				if err := _SubnetProviderUptime.contract.UnpackLog(event, "RewardClaimed", log); err != nil {
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

// ParseRewardClaimed is a log parse operation binding the contract event 0x24b5efa61dd1cfc659205a97fb8ed868f3cb8c81922bab2b96423e5de1de2cb7.
//
// Solidity: event RewardClaimed(uint256 indexed tokenId, address indexed owner, uint256 amount)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) ParseRewardClaimed(log types.Log) (*SubnetProviderUptimeRewardClaimed, error) {
	event := new(SubnetProviderUptimeRewardClaimed)
	if err := _SubnetProviderUptime.contract.UnpackLog(event, "RewardClaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetProviderUptimeRewardPerSecondUpdatedIterator is returned from FilterRewardPerSecondUpdated and is used to iterate over the raw logs and unpacked data for RewardPerSecondUpdated events raised by the SubnetProviderUptime contract.
type SubnetProviderUptimeRewardPerSecondUpdatedIterator struct {
	Event *SubnetProviderUptimeRewardPerSecondUpdated // Event containing the contract specifics and raw log

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
func (it *SubnetProviderUptimeRewardPerSecondUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderUptimeRewardPerSecondUpdated)
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
		it.Event = new(SubnetProviderUptimeRewardPerSecondUpdated)
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
func (it *SubnetProviderUptimeRewardPerSecondUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderUptimeRewardPerSecondUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderUptimeRewardPerSecondUpdated represents a RewardPerSecondUpdated event raised by the SubnetProviderUptime contract.
type SubnetProviderUptimeRewardPerSecondUpdated struct {
	OldRewardPerSecond *big.Int
	NewRewardPerSecond *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterRewardPerSecondUpdated is a free log retrieval operation binding the contract event 0xad9b65d94abd7b01bc5e6f82634fc2860427247e02b5b06e1fefc524c6af512f.
//
// Solidity: event RewardPerSecondUpdated(uint256 oldRewardPerSecond, uint256 newRewardPerSecond)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) FilterRewardPerSecondUpdated(opts *bind.FilterOpts) (*SubnetProviderUptimeRewardPerSecondUpdatedIterator, error) {

	logs, sub, err := _SubnetProviderUptime.contract.FilterLogs(opts, "RewardPerSecondUpdated")
	if err != nil {
		return nil, err
	}
	return &SubnetProviderUptimeRewardPerSecondUpdatedIterator{contract: _SubnetProviderUptime.contract, event: "RewardPerSecondUpdated", logs: logs, sub: sub}, nil
}

// WatchRewardPerSecondUpdated is a free log subscription operation binding the contract event 0xad9b65d94abd7b01bc5e6f82634fc2860427247e02b5b06e1fefc524c6af512f.
//
// Solidity: event RewardPerSecondUpdated(uint256 oldRewardPerSecond, uint256 newRewardPerSecond)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) WatchRewardPerSecondUpdated(opts *bind.WatchOpts, sink chan<- *SubnetProviderUptimeRewardPerSecondUpdated) (event.Subscription, error) {

	logs, sub, err := _SubnetProviderUptime.contract.WatchLogs(opts, "RewardPerSecondUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderUptimeRewardPerSecondUpdated)
				if err := _SubnetProviderUptime.contract.UnpackLog(event, "RewardPerSecondUpdated", log); err != nil {
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
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) ParseRewardPerSecondUpdated(log types.Log) (*SubnetProviderUptimeRewardPerSecondUpdated, error) {
	event := new(SubnetProviderUptimeRewardPerSecondUpdated)
	if err := _SubnetProviderUptime.contract.UnpackLog(event, "RewardPerSecondUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetProviderUptimeRewardReportedIterator is returned from FilterRewardReported and is used to iterate over the raw logs and unpacked data for RewardReported events raised by the SubnetProviderUptime contract.
type SubnetProviderUptimeRewardReportedIterator struct {
	Event *SubnetProviderUptimeRewardReported // Event containing the contract specifics and raw log

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
func (it *SubnetProviderUptimeRewardReportedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderUptimeRewardReported)
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
		it.Event = new(SubnetProviderUptimeRewardReported)
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
func (it *SubnetProviderUptimeRewardReportedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderUptimeRewardReportedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderUptimeRewardReported represents a RewardReported event raised by the SubnetProviderUptime contract.
type SubnetProviderUptimeRewardReported struct {
	TokenId       *big.Int
	PendingReward *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterRewardReported is a free log retrieval operation binding the contract event 0x99ca4546844cad32b055fa342831c6dd7df3886cd85edfc0a95977b840e8f339.
//
// Solidity: event RewardReported(uint256 indexed tokenId, uint256 pendingReward)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) FilterRewardReported(opts *bind.FilterOpts, tokenId []*big.Int) (*SubnetProviderUptimeRewardReportedIterator, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _SubnetProviderUptime.contract.FilterLogs(opts, "RewardReported", tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &SubnetProviderUptimeRewardReportedIterator{contract: _SubnetProviderUptime.contract, event: "RewardReported", logs: logs, sub: sub}, nil
}

// WatchRewardReported is a free log subscription operation binding the contract event 0x99ca4546844cad32b055fa342831c6dd7df3886cd85edfc0a95977b840e8f339.
//
// Solidity: event RewardReported(uint256 indexed tokenId, uint256 pendingReward)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) WatchRewardReported(opts *bind.WatchOpts, sink chan<- *SubnetProviderUptimeRewardReported, tokenId []*big.Int) (event.Subscription, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _SubnetProviderUptime.contract.WatchLogs(opts, "RewardReported", tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderUptimeRewardReported)
				if err := _SubnetProviderUptime.contract.UnpackLog(event, "RewardReported", log); err != nil {
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

// ParseRewardReported is a log parse operation binding the contract event 0x99ca4546844cad32b055fa342831c6dd7df3886cd85edfc0a95977b840e8f339.
//
// Solidity: event RewardReported(uint256 indexed tokenId, uint256 pendingReward)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) ParseRewardReported(log types.Log) (*SubnetProviderUptimeRewardReported, error) {
	event := new(SubnetProviderUptimeRewardReported)
	if err := _SubnetProviderUptime.contract.UnpackLog(event, "RewardReported", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetProviderUptimeUptimeUpdatedIterator is returned from FilterUptimeUpdated and is used to iterate over the raw logs and unpacked data for UptimeUpdated events raised by the SubnetProviderUptime contract.
type SubnetProviderUptimeUptimeUpdatedIterator struct {
	Event *SubnetProviderUptimeUptimeUpdated // Event containing the contract specifics and raw log

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
func (it *SubnetProviderUptimeUptimeUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderUptimeUptimeUpdated)
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
		it.Event = new(SubnetProviderUptimeUptimeUpdated)
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
func (it *SubnetProviderUptimeUptimeUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderUptimeUptimeUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderUptimeUptimeUpdated represents a UptimeUpdated event raised by the SubnetProviderUptime contract.
type SubnetProviderUptimeUptimeUpdated struct {
	TokenId     *big.Int
	TotalUptime *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterUptimeUpdated is a free log retrieval operation binding the contract event 0xf62a3c7d93025c93c8a5fb8c136eb68e1c5e720e63ec4a85bd9f1b20677eef71.
//
// Solidity: event UptimeUpdated(uint256 indexed tokenId, uint256 totalUptime)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) FilterUptimeUpdated(opts *bind.FilterOpts, tokenId []*big.Int) (*SubnetProviderUptimeUptimeUpdatedIterator, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _SubnetProviderUptime.contract.FilterLogs(opts, "UptimeUpdated", tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &SubnetProviderUptimeUptimeUpdatedIterator{contract: _SubnetProviderUptime.contract, event: "UptimeUpdated", logs: logs, sub: sub}, nil
}

// WatchUptimeUpdated is a free log subscription operation binding the contract event 0xf62a3c7d93025c93c8a5fb8c136eb68e1c5e720e63ec4a85bd9f1b20677eef71.
//
// Solidity: event UptimeUpdated(uint256 indexed tokenId, uint256 totalUptime)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) WatchUptimeUpdated(opts *bind.WatchOpts, sink chan<- *SubnetProviderUptimeUptimeUpdated, tokenId []*big.Int) (event.Subscription, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _SubnetProviderUptime.contract.WatchLogs(opts, "UptimeUpdated", tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderUptimeUptimeUpdated)
				if err := _SubnetProviderUptime.contract.UnpackLog(event, "UptimeUpdated", log); err != nil {
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

// ParseUptimeUpdated is a log parse operation binding the contract event 0xf62a3c7d93025c93c8a5fb8c136eb68e1c5e720e63ec4a85bd9f1b20677eef71.
//
// Solidity: event UptimeUpdated(uint256 indexed tokenId, uint256 totalUptime)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) ParseUptimeUpdated(log types.Log) (*SubnetProviderUptimeUptimeUpdated, error) {
	event := new(SubnetProviderUptimeUptimeUpdated)
	if err := _SubnetProviderUptime.contract.UnpackLog(event, "UptimeUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetProviderUptimeVerifierPeerIdUpdatedIterator is returned from FilterVerifierPeerIdUpdated and is used to iterate over the raw logs and unpacked data for VerifierPeerIdUpdated events raised by the SubnetProviderUptime contract.
type SubnetProviderUptimeVerifierPeerIdUpdatedIterator struct {
	Event *SubnetProviderUptimeVerifierPeerIdUpdated // Event containing the contract specifics and raw log

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
func (it *SubnetProviderUptimeVerifierPeerIdUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderUptimeVerifierPeerIdUpdated)
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
		it.Event = new(SubnetProviderUptimeVerifierPeerIdUpdated)
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
func (it *SubnetProviderUptimeVerifierPeerIdUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderUptimeVerifierPeerIdUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderUptimeVerifierPeerIdUpdated represents a VerifierPeerIdUpdated event raised by the SubnetProviderUptime contract.
type SubnetProviderUptimeVerifierPeerIdUpdated struct {
	OldVerifierPeerId string
	NewVerifierPeerId string
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterVerifierPeerIdUpdated is a free log retrieval operation binding the contract event 0x7f1ec80300d7812b29d17bbaab67b231a25f3da6b00f5ae6df703ede6f61e47f.
//
// Solidity: event VerifierPeerIdUpdated(string oldVerifierPeerId, string newVerifierPeerId)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) FilterVerifierPeerIdUpdated(opts *bind.FilterOpts) (*SubnetProviderUptimeVerifierPeerIdUpdatedIterator, error) {

	logs, sub, err := _SubnetProviderUptime.contract.FilterLogs(opts, "VerifierPeerIdUpdated")
	if err != nil {
		return nil, err
	}
	return &SubnetProviderUptimeVerifierPeerIdUpdatedIterator{contract: _SubnetProviderUptime.contract, event: "VerifierPeerIdUpdated", logs: logs, sub: sub}, nil
}

// WatchVerifierPeerIdUpdated is a free log subscription operation binding the contract event 0x7f1ec80300d7812b29d17bbaab67b231a25f3da6b00f5ae6df703ede6f61e47f.
//
// Solidity: event VerifierPeerIdUpdated(string oldVerifierPeerId, string newVerifierPeerId)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) WatchVerifierPeerIdUpdated(opts *bind.WatchOpts, sink chan<- *SubnetProviderUptimeVerifierPeerIdUpdated) (event.Subscription, error) {

	logs, sub, err := _SubnetProviderUptime.contract.WatchLogs(opts, "VerifierPeerIdUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderUptimeVerifierPeerIdUpdated)
				if err := _SubnetProviderUptime.contract.UnpackLog(event, "VerifierPeerIdUpdated", log); err != nil {
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

// ParseVerifierPeerIdUpdated is a log parse operation binding the contract event 0x7f1ec80300d7812b29d17bbaab67b231a25f3da6b00f5ae6df703ede6f61e47f.
//
// Solidity: event VerifierPeerIdUpdated(string oldVerifierPeerId, string newVerifierPeerId)
func (_SubnetProviderUptime *SubnetProviderUptimeFilterer) ParseVerifierPeerIdUpdated(log types.Log) (*SubnetProviderUptimeVerifierPeerIdUpdated, error) {
	event := new(SubnetProviderUptimeVerifierPeerIdUpdated)
	if err := _SubnetProviderUptime.contract.UnpackLog(event, "VerifierPeerIdUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
