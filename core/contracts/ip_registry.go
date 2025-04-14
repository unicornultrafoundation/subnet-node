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

// SubnetIPRegistryMetaData contains all meta data concerning the SubnetIPRegistry contract.
var SubnetIPRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"AddressEmptyCode\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"AddressInsufficientBalance\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"ERC721IncorrectOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ERC721InsufficientApproval\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"approver\",\"type\":\"address\"}],\"name\":\"ERC721InvalidApprover\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"ERC721InvalidOperator\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"ERC721InvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"}],\"name\":\"ERC721InvalidReceiver\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"ERC721InvalidSender\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ERC721NonexistentToken\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"SafeERC20FailedOperation\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"approved\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"}],\"name\":\"PeerBound\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"}],\"name\":\"bindPeer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getApproved\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getPeer\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"initialOwner\",\"type\":\"address\"},{\"internalType\":\"contractIERC20\",\"name\":\"_paymentToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_treasury\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_purchaseFee\",\"type\":\"uint256\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"ipPeer\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"paymentToken\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"purchase\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"purchaseFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"tokenURI\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"treasury\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newPurchaseFee\",\"type\":\"uint256\"}],\"name\":\"updatePurchaseFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newTreasury\",\"type\":\"address\"}],\"name\":\"updateTreasury\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// SubnetIPRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use SubnetIPRegistryMetaData.ABI instead.
var SubnetIPRegistryABI = SubnetIPRegistryMetaData.ABI

// SubnetIPRegistry is an auto generated Go binding around an Ethereum contract.
type SubnetIPRegistry struct {
	SubnetIPRegistryCaller     // Read-only binding to the contract
	SubnetIPRegistryTransactor // Write-only binding to the contract
	SubnetIPRegistryFilterer   // Log filterer for contract events
}

// SubnetIPRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type SubnetIPRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetIPRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SubnetIPRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetIPRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SubnetIPRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetIPRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SubnetIPRegistrySession struct {
	Contract     *SubnetIPRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SubnetIPRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SubnetIPRegistryCallerSession struct {
	Contract *SubnetIPRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// SubnetIPRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SubnetIPRegistryTransactorSession struct {
	Contract     *SubnetIPRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// SubnetIPRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type SubnetIPRegistryRaw struct {
	Contract *SubnetIPRegistry // Generic contract binding to access the raw methods on
}

// SubnetIPRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SubnetIPRegistryCallerRaw struct {
	Contract *SubnetIPRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// SubnetIPRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SubnetIPRegistryTransactorRaw struct {
	Contract *SubnetIPRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSubnetIPRegistry creates a new instance of SubnetIPRegistry, bound to a specific deployed contract.
func NewSubnetIPRegistry(address common.Address, backend bind.ContractBackend) (*SubnetIPRegistry, error) {
	contract, err := bindSubnetIPRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SubnetIPRegistry{SubnetIPRegistryCaller: SubnetIPRegistryCaller{contract: contract}, SubnetIPRegistryTransactor: SubnetIPRegistryTransactor{contract: contract}, SubnetIPRegistryFilterer: SubnetIPRegistryFilterer{contract: contract}}, nil
}

// NewSubnetIPRegistryCaller creates a new read-only instance of SubnetIPRegistry, bound to a specific deployed contract.
func NewSubnetIPRegistryCaller(address common.Address, caller bind.ContractCaller) (*SubnetIPRegistryCaller, error) {
	contract, err := bindSubnetIPRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SubnetIPRegistryCaller{contract: contract}, nil
}

// NewSubnetIPRegistryTransactor creates a new write-only instance of SubnetIPRegistry, bound to a specific deployed contract.
func NewSubnetIPRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*SubnetIPRegistryTransactor, error) {
	contract, err := bindSubnetIPRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SubnetIPRegistryTransactor{contract: contract}, nil
}

// NewSubnetIPRegistryFilterer creates a new log filterer instance of SubnetIPRegistry, bound to a specific deployed contract.
func NewSubnetIPRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*SubnetIPRegistryFilterer, error) {
	contract, err := bindSubnetIPRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SubnetIPRegistryFilterer{contract: contract}, nil
}

// bindSubnetIPRegistry binds a generic wrapper to an already deployed contract.
func bindSubnetIPRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SubnetIPRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SubnetIPRegistry *SubnetIPRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SubnetIPRegistry.Contract.SubnetIPRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SubnetIPRegistry *SubnetIPRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.SubnetIPRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SubnetIPRegistry *SubnetIPRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.SubnetIPRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SubnetIPRegistry *SubnetIPRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SubnetIPRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SubnetIPRegistry *SubnetIPRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SubnetIPRegistry *SubnetIPRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.contract.Transact(opts, method, params...)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_SubnetIPRegistry *SubnetIPRegistryCaller) BalanceOf(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _SubnetIPRegistry.contract.Call(opts, &out, "balanceOf", owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_SubnetIPRegistry *SubnetIPRegistrySession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _SubnetIPRegistry.Contract.BalanceOf(&_SubnetIPRegistry.CallOpts, owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_SubnetIPRegistry *SubnetIPRegistryCallerSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _SubnetIPRegistry.Contract.BalanceOf(&_SubnetIPRegistry.CallOpts, owner)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_SubnetIPRegistry *SubnetIPRegistryCaller) GetApproved(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _SubnetIPRegistry.contract.Call(opts, &out, "getApproved", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_SubnetIPRegistry *SubnetIPRegistrySession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _SubnetIPRegistry.Contract.GetApproved(&_SubnetIPRegistry.CallOpts, tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_SubnetIPRegistry *SubnetIPRegistryCallerSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _SubnetIPRegistry.Contract.GetApproved(&_SubnetIPRegistry.CallOpts, tokenId)
}

// GetPeer is a free data retrieval call binding the contract method 0x67ebb6b2.
//
// Solidity: function getPeer(uint256 tokenId) view returns(string)
func (_SubnetIPRegistry *SubnetIPRegistryCaller) GetPeer(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var out []interface{}
	err := _SubnetIPRegistry.contract.Call(opts, &out, "getPeer", tokenId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetPeer is a free data retrieval call binding the contract method 0x67ebb6b2.
//
// Solidity: function getPeer(uint256 tokenId) view returns(string)
func (_SubnetIPRegistry *SubnetIPRegistrySession) GetPeer(tokenId *big.Int) (string, error) {
	return _SubnetIPRegistry.Contract.GetPeer(&_SubnetIPRegistry.CallOpts, tokenId)
}

// GetPeer is a free data retrieval call binding the contract method 0x67ebb6b2.
//
// Solidity: function getPeer(uint256 tokenId) view returns(string)
func (_SubnetIPRegistry *SubnetIPRegistryCallerSession) GetPeer(tokenId *big.Int) (string, error) {
	return _SubnetIPRegistry.Contract.GetPeer(&_SubnetIPRegistry.CallOpts, tokenId)
}

// IpPeer is a free data retrieval call binding the contract method 0xc5835b08.
//
// Solidity: function ipPeer(uint256 ) view returns(string)
func (_SubnetIPRegistry *SubnetIPRegistryCaller) IpPeer(opts *bind.CallOpts, arg0 *big.Int) (string, error) {
	var out []interface{}
	err := _SubnetIPRegistry.contract.Call(opts, &out, "ipPeer", arg0)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// IpPeer is a free data retrieval call binding the contract method 0xc5835b08.
//
// Solidity: function ipPeer(uint256 ) view returns(string)
func (_SubnetIPRegistry *SubnetIPRegistrySession) IpPeer(arg0 *big.Int) (string, error) {
	return _SubnetIPRegistry.Contract.IpPeer(&_SubnetIPRegistry.CallOpts, arg0)
}

// IpPeer is a free data retrieval call binding the contract method 0xc5835b08.
//
// Solidity: function ipPeer(uint256 ) view returns(string)
func (_SubnetIPRegistry *SubnetIPRegistryCallerSession) IpPeer(arg0 *big.Int) (string, error) {
	return _SubnetIPRegistry.Contract.IpPeer(&_SubnetIPRegistry.CallOpts, arg0)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_SubnetIPRegistry *SubnetIPRegistryCaller) IsApprovedForAll(opts *bind.CallOpts, owner common.Address, operator common.Address) (bool, error) {
	var out []interface{}
	err := _SubnetIPRegistry.contract.Call(opts, &out, "isApprovedForAll", owner, operator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_SubnetIPRegistry *SubnetIPRegistrySession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _SubnetIPRegistry.Contract.IsApprovedForAll(&_SubnetIPRegistry.CallOpts, owner, operator)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_SubnetIPRegistry *SubnetIPRegistryCallerSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _SubnetIPRegistry.Contract.IsApprovedForAll(&_SubnetIPRegistry.CallOpts, owner, operator)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_SubnetIPRegistry *SubnetIPRegistryCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _SubnetIPRegistry.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_SubnetIPRegistry *SubnetIPRegistrySession) Name() (string, error) {
	return _SubnetIPRegistry.Contract.Name(&_SubnetIPRegistry.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_SubnetIPRegistry *SubnetIPRegistryCallerSession) Name() (string, error) {
	return _SubnetIPRegistry.Contract.Name(&_SubnetIPRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetIPRegistry *SubnetIPRegistryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetIPRegistry.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetIPRegistry *SubnetIPRegistrySession) Owner() (common.Address, error) {
	return _SubnetIPRegistry.Contract.Owner(&_SubnetIPRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetIPRegistry *SubnetIPRegistryCallerSession) Owner() (common.Address, error) {
	return _SubnetIPRegistry.Contract.Owner(&_SubnetIPRegistry.CallOpts)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_SubnetIPRegistry *SubnetIPRegistryCaller) OwnerOf(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _SubnetIPRegistry.contract.Call(opts, &out, "ownerOf", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_SubnetIPRegistry *SubnetIPRegistrySession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _SubnetIPRegistry.Contract.OwnerOf(&_SubnetIPRegistry.CallOpts, tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_SubnetIPRegistry *SubnetIPRegistryCallerSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _SubnetIPRegistry.Contract.OwnerOf(&_SubnetIPRegistry.CallOpts, tokenId)
}

// PaymentToken is a free data retrieval call binding the contract method 0x3013ce29.
//
// Solidity: function paymentToken() view returns(address)
func (_SubnetIPRegistry *SubnetIPRegistryCaller) PaymentToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetIPRegistry.contract.Call(opts, &out, "paymentToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PaymentToken is a free data retrieval call binding the contract method 0x3013ce29.
//
// Solidity: function paymentToken() view returns(address)
func (_SubnetIPRegistry *SubnetIPRegistrySession) PaymentToken() (common.Address, error) {
	return _SubnetIPRegistry.Contract.PaymentToken(&_SubnetIPRegistry.CallOpts)
}

// PaymentToken is a free data retrieval call binding the contract method 0x3013ce29.
//
// Solidity: function paymentToken() view returns(address)
func (_SubnetIPRegistry *SubnetIPRegistryCallerSession) PaymentToken() (common.Address, error) {
	return _SubnetIPRegistry.Contract.PaymentToken(&_SubnetIPRegistry.CallOpts)
}

// PurchaseFee is a free data retrieval call binding the contract method 0x14b5e981.
//
// Solidity: function purchaseFee() view returns(uint256)
func (_SubnetIPRegistry *SubnetIPRegistryCaller) PurchaseFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SubnetIPRegistry.contract.Call(opts, &out, "purchaseFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PurchaseFee is a free data retrieval call binding the contract method 0x14b5e981.
//
// Solidity: function purchaseFee() view returns(uint256)
func (_SubnetIPRegistry *SubnetIPRegistrySession) PurchaseFee() (*big.Int, error) {
	return _SubnetIPRegistry.Contract.PurchaseFee(&_SubnetIPRegistry.CallOpts)
}

// PurchaseFee is a free data retrieval call binding the contract method 0x14b5e981.
//
// Solidity: function purchaseFee() view returns(uint256)
func (_SubnetIPRegistry *SubnetIPRegistryCallerSession) PurchaseFee() (*big.Int, error) {
	return _SubnetIPRegistry.Contract.PurchaseFee(&_SubnetIPRegistry.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_SubnetIPRegistry *SubnetIPRegistryCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _SubnetIPRegistry.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_SubnetIPRegistry *SubnetIPRegistrySession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _SubnetIPRegistry.Contract.SupportsInterface(&_SubnetIPRegistry.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_SubnetIPRegistry *SubnetIPRegistryCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _SubnetIPRegistry.Contract.SupportsInterface(&_SubnetIPRegistry.CallOpts, interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_SubnetIPRegistry *SubnetIPRegistryCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _SubnetIPRegistry.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_SubnetIPRegistry *SubnetIPRegistrySession) Symbol() (string, error) {
	return _SubnetIPRegistry.Contract.Symbol(&_SubnetIPRegistry.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_SubnetIPRegistry *SubnetIPRegistryCallerSession) Symbol() (string, error) {
	return _SubnetIPRegistry.Contract.Symbol(&_SubnetIPRegistry.CallOpts)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_SubnetIPRegistry *SubnetIPRegistryCaller) TokenURI(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var out []interface{}
	err := _SubnetIPRegistry.contract.Call(opts, &out, "tokenURI", tokenId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_SubnetIPRegistry *SubnetIPRegistrySession) TokenURI(tokenId *big.Int) (string, error) {
	return _SubnetIPRegistry.Contract.TokenURI(&_SubnetIPRegistry.CallOpts, tokenId)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_SubnetIPRegistry *SubnetIPRegistryCallerSession) TokenURI(tokenId *big.Int) (string, error) {
	return _SubnetIPRegistry.Contract.TokenURI(&_SubnetIPRegistry.CallOpts, tokenId)
}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_SubnetIPRegistry *SubnetIPRegistryCaller) Treasury(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetIPRegistry.contract.Call(opts, &out, "treasury")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_SubnetIPRegistry *SubnetIPRegistrySession) Treasury() (common.Address, error) {
	return _SubnetIPRegistry.Contract.Treasury(&_SubnetIPRegistry.CallOpts)
}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_SubnetIPRegistry *SubnetIPRegistryCallerSession) Treasury() (common.Address, error) {
	return _SubnetIPRegistry.Contract.Treasury(&_SubnetIPRegistry.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactor) Approve(opts *bind.TransactOpts, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetIPRegistry.contract.Transact(opts, "approve", to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_SubnetIPRegistry *SubnetIPRegistrySession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.Approve(&_SubnetIPRegistry.TransactOpts, to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactorSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.Approve(&_SubnetIPRegistry.TransactOpts, to, tokenId)
}

// BindPeer is a paid mutator transaction binding the contract method 0x0e03720e.
//
// Solidity: function bindPeer(uint256 tokenId, string peerId) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactor) BindPeer(opts *bind.TransactOpts, tokenId *big.Int, peerId string) (*types.Transaction, error) {
	return _SubnetIPRegistry.contract.Transact(opts, "bindPeer", tokenId, peerId)
}

// BindPeer is a paid mutator transaction binding the contract method 0x0e03720e.
//
// Solidity: function bindPeer(uint256 tokenId, string peerId) returns()
func (_SubnetIPRegistry *SubnetIPRegistrySession) BindPeer(tokenId *big.Int, peerId string) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.BindPeer(&_SubnetIPRegistry.TransactOpts, tokenId, peerId)
}

// BindPeer is a paid mutator transaction binding the contract method 0x0e03720e.
//
// Solidity: function bindPeer(uint256 tokenId, string peerId) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactorSession) BindPeer(tokenId *big.Int, peerId string) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.BindPeer(&_SubnetIPRegistry.TransactOpts, tokenId, peerId)
}

// Initialize is a paid mutator transaction binding the contract method 0xcf756fdf.
//
// Solidity: function initialize(address initialOwner, address _paymentToken, address _treasury, uint256 _purchaseFee) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactor) Initialize(opts *bind.TransactOpts, initialOwner common.Address, _paymentToken common.Address, _treasury common.Address, _purchaseFee *big.Int) (*types.Transaction, error) {
	return _SubnetIPRegistry.contract.Transact(opts, "initialize", initialOwner, _paymentToken, _treasury, _purchaseFee)
}

// Initialize is a paid mutator transaction binding the contract method 0xcf756fdf.
//
// Solidity: function initialize(address initialOwner, address _paymentToken, address _treasury, uint256 _purchaseFee) returns()
func (_SubnetIPRegistry *SubnetIPRegistrySession) Initialize(initialOwner common.Address, _paymentToken common.Address, _treasury common.Address, _purchaseFee *big.Int) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.Initialize(&_SubnetIPRegistry.TransactOpts, initialOwner, _paymentToken, _treasury, _purchaseFee)
}

// Initialize is a paid mutator transaction binding the contract method 0xcf756fdf.
//
// Solidity: function initialize(address initialOwner, address _paymentToken, address _treasury, uint256 _purchaseFee) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactorSession) Initialize(initialOwner common.Address, _paymentToken common.Address, _treasury common.Address, _purchaseFee *big.Int) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.Initialize(&_SubnetIPRegistry.TransactOpts, initialOwner, _paymentToken, _treasury, _purchaseFee)
}

// Purchase is a paid mutator transaction binding the contract method 0x25b31a97.
//
// Solidity: function purchase(address to) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactor) Purchase(opts *bind.TransactOpts, to common.Address) (*types.Transaction, error) {
	return _SubnetIPRegistry.contract.Transact(opts, "purchase", to)
}

// Purchase is a paid mutator transaction binding the contract method 0x25b31a97.
//
// Solidity: function purchase(address to) returns()
func (_SubnetIPRegistry *SubnetIPRegistrySession) Purchase(to common.Address) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.Purchase(&_SubnetIPRegistry.TransactOpts, to)
}

// Purchase is a paid mutator transaction binding the contract method 0x25b31a97.
//
// Solidity: function purchase(address to) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactorSession) Purchase(to common.Address) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.Purchase(&_SubnetIPRegistry.TransactOpts, to)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetIPRegistry.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetIPRegistry *SubnetIPRegistrySession) RenounceOwnership() (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.RenounceOwnership(&_SubnetIPRegistry.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.RenounceOwnership(&_SubnetIPRegistry.TransactOpts)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactor) SafeTransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetIPRegistry.contract.Transact(opts, "safeTransferFrom", from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_SubnetIPRegistry *SubnetIPRegistrySession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.SafeTransferFrom(&_SubnetIPRegistry.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactorSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.SafeTransferFrom(&_SubnetIPRegistry.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactor) SafeTransferFrom0(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _SubnetIPRegistry.contract.Transact(opts, "safeTransferFrom0", from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_SubnetIPRegistry *SubnetIPRegistrySession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.SafeTransferFrom0(&_SubnetIPRegistry.TransactOpts, from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactorSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.SafeTransferFrom0(&_SubnetIPRegistry.TransactOpts, from, to, tokenId, data)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactor) SetApprovalForAll(opts *bind.TransactOpts, operator common.Address, approved bool) (*types.Transaction, error) {
	return _SubnetIPRegistry.contract.Transact(opts, "setApprovalForAll", operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_SubnetIPRegistry *SubnetIPRegistrySession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.SetApprovalForAll(&_SubnetIPRegistry.TransactOpts, operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactorSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.SetApprovalForAll(&_SubnetIPRegistry.TransactOpts, operator, approved)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetIPRegistry.contract.Transact(opts, "transferFrom", from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_SubnetIPRegistry *SubnetIPRegistrySession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.TransferFrom(&_SubnetIPRegistry.TransactOpts, from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactorSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.TransferFrom(&_SubnetIPRegistry.TransactOpts, from, to, tokenId)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _SubnetIPRegistry.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetIPRegistry *SubnetIPRegistrySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.TransferOwnership(&_SubnetIPRegistry.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.TransferOwnership(&_SubnetIPRegistry.TransactOpts, newOwner)
}

// UpdatePurchaseFee is a paid mutator transaction binding the contract method 0x9ca1f363.
//
// Solidity: function updatePurchaseFee(uint256 newPurchaseFee) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactor) UpdatePurchaseFee(opts *bind.TransactOpts, newPurchaseFee *big.Int) (*types.Transaction, error) {
	return _SubnetIPRegistry.contract.Transact(opts, "updatePurchaseFee", newPurchaseFee)
}

// UpdatePurchaseFee is a paid mutator transaction binding the contract method 0x9ca1f363.
//
// Solidity: function updatePurchaseFee(uint256 newPurchaseFee) returns()
func (_SubnetIPRegistry *SubnetIPRegistrySession) UpdatePurchaseFee(newPurchaseFee *big.Int) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.UpdatePurchaseFee(&_SubnetIPRegistry.TransactOpts, newPurchaseFee)
}

// UpdatePurchaseFee is a paid mutator transaction binding the contract method 0x9ca1f363.
//
// Solidity: function updatePurchaseFee(uint256 newPurchaseFee) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactorSession) UpdatePurchaseFee(newPurchaseFee *big.Int) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.UpdatePurchaseFee(&_SubnetIPRegistry.TransactOpts, newPurchaseFee)
}

// UpdateTreasury is a paid mutator transaction binding the contract method 0x7f51bb1f.
//
// Solidity: function updateTreasury(address newTreasury) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactor) UpdateTreasury(opts *bind.TransactOpts, newTreasury common.Address) (*types.Transaction, error) {
	return _SubnetIPRegistry.contract.Transact(opts, "updateTreasury", newTreasury)
}

// UpdateTreasury is a paid mutator transaction binding the contract method 0x7f51bb1f.
//
// Solidity: function updateTreasury(address newTreasury) returns()
func (_SubnetIPRegistry *SubnetIPRegistrySession) UpdateTreasury(newTreasury common.Address) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.UpdateTreasury(&_SubnetIPRegistry.TransactOpts, newTreasury)
}

// UpdateTreasury is a paid mutator transaction binding the contract method 0x7f51bb1f.
//
// Solidity: function updateTreasury(address newTreasury) returns()
func (_SubnetIPRegistry *SubnetIPRegistryTransactorSession) UpdateTreasury(newTreasury common.Address) (*types.Transaction, error) {
	return _SubnetIPRegistry.Contract.UpdateTreasury(&_SubnetIPRegistry.TransactOpts, newTreasury)
}

// SubnetIPRegistryApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the SubnetIPRegistry contract.
type SubnetIPRegistryApprovalIterator struct {
	Event *SubnetIPRegistryApproval // Event containing the contract specifics and raw log

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
func (it *SubnetIPRegistryApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetIPRegistryApproval)
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
		it.Event = new(SubnetIPRegistryApproval)
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
func (it *SubnetIPRegistryApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetIPRegistryApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetIPRegistryApproval represents a Approval event raised by the SubnetIPRegistry contract.
type SubnetIPRegistryApproval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_SubnetIPRegistry *SubnetIPRegistryFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, approved []common.Address, tokenId []*big.Int) (*SubnetIPRegistryApprovalIterator, error) {

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

	logs, sub, err := _SubnetIPRegistry.contract.FilterLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &SubnetIPRegistryApprovalIterator{contract: _SubnetIPRegistry.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_SubnetIPRegistry *SubnetIPRegistryFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *SubnetIPRegistryApproval, owner []common.Address, approved []common.Address, tokenId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _SubnetIPRegistry.contract.WatchLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetIPRegistryApproval)
				if err := _SubnetIPRegistry.contract.UnpackLog(event, "Approval", log); err != nil {
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
func (_SubnetIPRegistry *SubnetIPRegistryFilterer) ParseApproval(log types.Log) (*SubnetIPRegistryApproval, error) {
	event := new(SubnetIPRegistryApproval)
	if err := _SubnetIPRegistry.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetIPRegistryApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the SubnetIPRegistry contract.
type SubnetIPRegistryApprovalForAllIterator struct {
	Event *SubnetIPRegistryApprovalForAll // Event containing the contract specifics and raw log

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
func (it *SubnetIPRegistryApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetIPRegistryApprovalForAll)
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
		it.Event = new(SubnetIPRegistryApprovalForAll)
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
func (it *SubnetIPRegistryApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetIPRegistryApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetIPRegistryApprovalForAll represents a ApprovalForAll event raised by the SubnetIPRegistry contract.
type SubnetIPRegistryApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_SubnetIPRegistry *SubnetIPRegistryFilterer) FilterApprovalForAll(opts *bind.FilterOpts, owner []common.Address, operator []common.Address) (*SubnetIPRegistryApprovalForAllIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _SubnetIPRegistry.contract.FilterLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return &SubnetIPRegistryApprovalForAllIterator{contract: _SubnetIPRegistry.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_SubnetIPRegistry *SubnetIPRegistryFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *SubnetIPRegistryApprovalForAll, owner []common.Address, operator []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _SubnetIPRegistry.contract.WatchLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetIPRegistryApprovalForAll)
				if err := _SubnetIPRegistry.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
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
func (_SubnetIPRegistry *SubnetIPRegistryFilterer) ParseApprovalForAll(log types.Log) (*SubnetIPRegistryApprovalForAll, error) {
	event := new(SubnetIPRegistryApprovalForAll)
	if err := _SubnetIPRegistry.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetIPRegistryInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the SubnetIPRegistry contract.
type SubnetIPRegistryInitializedIterator struct {
	Event *SubnetIPRegistryInitialized // Event containing the contract specifics and raw log

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
func (it *SubnetIPRegistryInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetIPRegistryInitialized)
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
		it.Event = new(SubnetIPRegistryInitialized)
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
func (it *SubnetIPRegistryInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetIPRegistryInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetIPRegistryInitialized represents a Initialized event raised by the SubnetIPRegistry contract.
type SubnetIPRegistryInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_SubnetIPRegistry *SubnetIPRegistryFilterer) FilterInitialized(opts *bind.FilterOpts) (*SubnetIPRegistryInitializedIterator, error) {

	logs, sub, err := _SubnetIPRegistry.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &SubnetIPRegistryInitializedIterator{contract: _SubnetIPRegistry.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_SubnetIPRegistry *SubnetIPRegistryFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *SubnetIPRegistryInitialized) (event.Subscription, error) {

	logs, sub, err := _SubnetIPRegistry.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetIPRegistryInitialized)
				if err := _SubnetIPRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_SubnetIPRegistry *SubnetIPRegistryFilterer) ParseInitialized(log types.Log) (*SubnetIPRegistryInitialized, error) {
	event := new(SubnetIPRegistryInitialized)
	if err := _SubnetIPRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetIPRegistryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the SubnetIPRegistry contract.
type SubnetIPRegistryOwnershipTransferredIterator struct {
	Event *SubnetIPRegistryOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *SubnetIPRegistryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetIPRegistryOwnershipTransferred)
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
		it.Event = new(SubnetIPRegistryOwnershipTransferred)
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
func (it *SubnetIPRegistryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetIPRegistryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetIPRegistryOwnershipTransferred represents a OwnershipTransferred event raised by the SubnetIPRegistry contract.
type SubnetIPRegistryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SubnetIPRegistry *SubnetIPRegistryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*SubnetIPRegistryOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SubnetIPRegistry.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &SubnetIPRegistryOwnershipTransferredIterator{contract: _SubnetIPRegistry.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SubnetIPRegistry *SubnetIPRegistryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *SubnetIPRegistryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SubnetIPRegistry.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetIPRegistryOwnershipTransferred)
				if err := _SubnetIPRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_SubnetIPRegistry *SubnetIPRegistryFilterer) ParseOwnershipTransferred(log types.Log) (*SubnetIPRegistryOwnershipTransferred, error) {
	event := new(SubnetIPRegistryOwnershipTransferred)
	if err := _SubnetIPRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetIPRegistryPeerBoundIterator is returned from FilterPeerBound and is used to iterate over the raw logs and unpacked data for PeerBound events raised by the SubnetIPRegistry contract.
type SubnetIPRegistryPeerBoundIterator struct {
	Event *SubnetIPRegistryPeerBound // Event containing the contract specifics and raw log

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
func (it *SubnetIPRegistryPeerBoundIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetIPRegistryPeerBound)
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
		it.Event = new(SubnetIPRegistryPeerBound)
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
func (it *SubnetIPRegistryPeerBoundIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetIPRegistryPeerBoundIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetIPRegistryPeerBound represents a PeerBound event raised by the SubnetIPRegistry contract.
type SubnetIPRegistryPeerBound struct {
	TokenId *big.Int
	PeerId  string
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPeerBound is a free log retrieval operation binding the contract event 0x7826251d6dd3b3f6b469df5d8b38e0055e2985d85c5080879821acd5fa226536.
//
// Solidity: event PeerBound(uint256 indexed tokenId, string peerId)
func (_SubnetIPRegistry *SubnetIPRegistryFilterer) FilterPeerBound(opts *bind.FilterOpts, tokenId []*big.Int) (*SubnetIPRegistryPeerBoundIterator, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _SubnetIPRegistry.contract.FilterLogs(opts, "PeerBound", tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &SubnetIPRegistryPeerBoundIterator{contract: _SubnetIPRegistry.contract, event: "PeerBound", logs: logs, sub: sub}, nil
}

// WatchPeerBound is a free log subscription operation binding the contract event 0x7826251d6dd3b3f6b469df5d8b38e0055e2985d85c5080879821acd5fa226536.
//
// Solidity: event PeerBound(uint256 indexed tokenId, string peerId)
func (_SubnetIPRegistry *SubnetIPRegistryFilterer) WatchPeerBound(opts *bind.WatchOpts, sink chan<- *SubnetIPRegistryPeerBound, tokenId []*big.Int) (event.Subscription, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _SubnetIPRegistry.contract.WatchLogs(opts, "PeerBound", tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetIPRegistryPeerBound)
				if err := _SubnetIPRegistry.contract.UnpackLog(event, "PeerBound", log); err != nil {
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

// ParsePeerBound is a log parse operation binding the contract event 0x7826251d6dd3b3f6b469df5d8b38e0055e2985d85c5080879821acd5fa226536.
//
// Solidity: event PeerBound(uint256 indexed tokenId, string peerId)
func (_SubnetIPRegistry *SubnetIPRegistryFilterer) ParsePeerBound(log types.Log) (*SubnetIPRegistryPeerBound, error) {
	event := new(SubnetIPRegistryPeerBound)
	if err := _SubnetIPRegistry.contract.UnpackLog(event, "PeerBound", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetIPRegistryTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the SubnetIPRegistry contract.
type SubnetIPRegistryTransferIterator struct {
	Event *SubnetIPRegistryTransfer // Event containing the contract specifics and raw log

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
func (it *SubnetIPRegistryTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetIPRegistryTransfer)
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
		it.Event = new(SubnetIPRegistryTransfer)
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
func (it *SubnetIPRegistryTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetIPRegistryTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetIPRegistryTransfer represents a Transfer event raised by the SubnetIPRegistry contract.
type SubnetIPRegistryTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_SubnetIPRegistry *SubnetIPRegistryFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address, tokenId []*big.Int) (*SubnetIPRegistryTransferIterator, error) {

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

	logs, sub, err := _SubnetIPRegistry.contract.FilterLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &SubnetIPRegistryTransferIterator{contract: _SubnetIPRegistry.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_SubnetIPRegistry *SubnetIPRegistryFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *SubnetIPRegistryTransfer, from []common.Address, to []common.Address, tokenId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _SubnetIPRegistry.contract.WatchLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetIPRegistryTransfer)
				if err := _SubnetIPRegistry.contract.UnpackLog(event, "Transfer", log); err != nil {
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
func (_SubnetIPRegistry *SubnetIPRegistryFilterer) ParseTransfer(log types.Log) (*SubnetIPRegistryTransfer, error) {
	event := new(SubnetIPRegistryTransfer)
	if err := _SubnetIPRegistry.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
