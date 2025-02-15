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

// SubnetVerifierVerifierInfo is an auto generated low-level Go binding around an user-defined struct.
type SubnetVerifierVerifierInfo struct {
	IsRegistered bool
	PeerIds      []string
}

// SubnetVerifierMetaData contains all meta data concerning the SubnetVerifier contract.
var SubnetVerifierMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"verifier\",\"type\":\"address\"}],\"name\":\"VerifierDeleted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"verifier\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string[]\",\"name\":\"peerIds\",\"type\":\"string[]\"}],\"name\":\"VerifierPeersUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"verifier\",\"type\":\"address\"}],\"name\":\"VerifierRegistered\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"verifier\",\"type\":\"address\"}],\"name\":\"deleteVerifier\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"verifier\",\"type\":\"address\"}],\"name\":\"getVerifierInfo\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"isRegistered\",\"type\":\"bool\"},{\"internalType\":\"string[]\",\"name\":\"peerIds\",\"type\":\"string[]\"}],\"internalType\":\"structSubnetVerifier.VerifierInfo\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_initialOwner\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"verifier\",\"type\":\"address\"},{\"internalType\":\"string[]\",\"name\":\"peerIds\",\"type\":\"string[]\"}],\"name\":\"registerVerifier\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"verifier\",\"type\":\"address\"},{\"internalType\":\"string[]\",\"name\":\"peerIds\",\"type\":\"string[]\"}],\"name\":\"updateVerifierPeers\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"verifiers\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"isRegistered\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]",
}

// SubnetVerifierABI is the input ABI used to generate the binding from.
// Deprecated: Use SubnetVerifierMetaData.ABI instead.
var SubnetVerifierABI = SubnetVerifierMetaData.ABI

// SubnetVerifier is an auto generated Go binding around an Ethereum contract.
type SubnetVerifier struct {
	SubnetVerifierCaller     // Read-only binding to the contract
	SubnetVerifierTransactor // Write-only binding to the contract
	SubnetVerifierFilterer   // Log filterer for contract events
}

// SubnetVerifierCaller is an auto generated read-only Go binding around an Ethereum contract.
type SubnetVerifierCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetVerifierTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SubnetVerifierTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetVerifierFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SubnetVerifierFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetVerifierSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SubnetVerifierSession struct {
	Contract     *SubnetVerifier   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SubnetVerifierCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SubnetVerifierCallerSession struct {
	Contract *SubnetVerifierCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// SubnetVerifierTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SubnetVerifierTransactorSession struct {
	Contract     *SubnetVerifierTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// SubnetVerifierRaw is an auto generated low-level Go binding around an Ethereum contract.
type SubnetVerifierRaw struct {
	Contract *SubnetVerifier // Generic contract binding to access the raw methods on
}

// SubnetVerifierCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SubnetVerifierCallerRaw struct {
	Contract *SubnetVerifierCaller // Generic read-only contract binding to access the raw methods on
}

// SubnetVerifierTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SubnetVerifierTransactorRaw struct {
	Contract *SubnetVerifierTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSubnetVerifier creates a new instance of SubnetVerifier, bound to a specific deployed contract.
func NewSubnetVerifier(address common.Address, backend bind.ContractBackend) (*SubnetVerifier, error) {
	contract, err := bindSubnetVerifier(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SubnetVerifier{SubnetVerifierCaller: SubnetVerifierCaller{contract: contract}, SubnetVerifierTransactor: SubnetVerifierTransactor{contract: contract}, SubnetVerifierFilterer: SubnetVerifierFilterer{contract: contract}}, nil
}

// NewSubnetVerifierCaller creates a new read-only instance of SubnetVerifier, bound to a specific deployed contract.
func NewSubnetVerifierCaller(address common.Address, caller bind.ContractCaller) (*SubnetVerifierCaller, error) {
	contract, err := bindSubnetVerifier(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SubnetVerifierCaller{contract: contract}, nil
}

// NewSubnetVerifierTransactor creates a new write-only instance of SubnetVerifier, bound to a specific deployed contract.
func NewSubnetVerifierTransactor(address common.Address, transactor bind.ContractTransactor) (*SubnetVerifierTransactor, error) {
	contract, err := bindSubnetVerifier(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SubnetVerifierTransactor{contract: contract}, nil
}

// NewSubnetVerifierFilterer creates a new log filterer instance of SubnetVerifier, bound to a specific deployed contract.
func NewSubnetVerifierFilterer(address common.Address, filterer bind.ContractFilterer) (*SubnetVerifierFilterer, error) {
	contract, err := bindSubnetVerifier(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SubnetVerifierFilterer{contract: contract}, nil
}

// bindSubnetVerifier binds a generic wrapper to an already deployed contract.
func bindSubnetVerifier(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SubnetVerifierMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SubnetVerifier *SubnetVerifierRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SubnetVerifier.Contract.SubnetVerifierCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SubnetVerifier *SubnetVerifierRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetVerifier.Contract.SubnetVerifierTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SubnetVerifier *SubnetVerifierRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SubnetVerifier.Contract.SubnetVerifierTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SubnetVerifier *SubnetVerifierCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SubnetVerifier.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SubnetVerifier *SubnetVerifierTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetVerifier.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SubnetVerifier *SubnetVerifierTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SubnetVerifier.Contract.contract.Transact(opts, method, params...)
}

// GetVerifierInfo is a free data retrieval call binding the contract method 0xa3e0851d.
//
// Solidity: function getVerifierInfo(address verifier) view returns((bool,string[]))
func (_SubnetVerifier *SubnetVerifierCaller) GetVerifierInfo(opts *bind.CallOpts, verifier common.Address) (SubnetVerifierVerifierInfo, error) {
	var out []interface{}
	err := _SubnetVerifier.contract.Call(opts, &out, "getVerifierInfo", verifier)

	if err != nil {
		return *new(SubnetVerifierVerifierInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(SubnetVerifierVerifierInfo)).(*SubnetVerifierVerifierInfo)

	return out0, err

}

// GetVerifierInfo is a free data retrieval call binding the contract method 0xa3e0851d.
//
// Solidity: function getVerifierInfo(address verifier) view returns((bool,string[]))
func (_SubnetVerifier *SubnetVerifierSession) GetVerifierInfo(verifier common.Address) (SubnetVerifierVerifierInfo, error) {
	return _SubnetVerifier.Contract.GetVerifierInfo(&_SubnetVerifier.CallOpts, verifier)
}

// GetVerifierInfo is a free data retrieval call binding the contract method 0xa3e0851d.
//
// Solidity: function getVerifierInfo(address verifier) view returns((bool,string[]))
func (_SubnetVerifier *SubnetVerifierCallerSession) GetVerifierInfo(verifier common.Address) (SubnetVerifierVerifierInfo, error) {
	return _SubnetVerifier.Contract.GetVerifierInfo(&_SubnetVerifier.CallOpts, verifier)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetVerifier *SubnetVerifierCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetVerifier.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetVerifier *SubnetVerifierSession) Owner() (common.Address, error) {
	return _SubnetVerifier.Contract.Owner(&_SubnetVerifier.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetVerifier *SubnetVerifierCallerSession) Owner() (common.Address, error) {
	return _SubnetVerifier.Contract.Owner(&_SubnetVerifier.CallOpts)
}

// Verifiers is a free data retrieval call binding the contract method 0x6c824487.
//
// Solidity: function verifiers(address ) view returns(bool isRegistered)
func (_SubnetVerifier *SubnetVerifierCaller) Verifiers(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _SubnetVerifier.contract.Call(opts, &out, "verifiers", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Verifiers is a free data retrieval call binding the contract method 0x6c824487.
//
// Solidity: function verifiers(address ) view returns(bool isRegistered)
func (_SubnetVerifier *SubnetVerifierSession) Verifiers(arg0 common.Address) (bool, error) {
	return _SubnetVerifier.Contract.Verifiers(&_SubnetVerifier.CallOpts, arg0)
}

// Verifiers is a free data retrieval call binding the contract method 0x6c824487.
//
// Solidity: function verifiers(address ) view returns(bool isRegistered)
func (_SubnetVerifier *SubnetVerifierCallerSession) Verifiers(arg0 common.Address) (bool, error) {
	return _SubnetVerifier.Contract.Verifiers(&_SubnetVerifier.CallOpts, arg0)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(string)
func (_SubnetVerifier *SubnetVerifierCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _SubnetVerifier.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(string)
func (_SubnetVerifier *SubnetVerifierSession) Version() (string, error) {
	return _SubnetVerifier.Contract.Version(&_SubnetVerifier.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(string)
func (_SubnetVerifier *SubnetVerifierCallerSession) Version() (string, error) {
	return _SubnetVerifier.Contract.Version(&_SubnetVerifier.CallOpts)
}

// DeleteVerifier is a paid mutator transaction binding the contract method 0x55f32817.
//
// Solidity: function deleteVerifier(address verifier) returns()
func (_SubnetVerifier *SubnetVerifierTransactor) DeleteVerifier(opts *bind.TransactOpts, verifier common.Address) (*types.Transaction, error) {
	return _SubnetVerifier.contract.Transact(opts, "deleteVerifier", verifier)
}

// DeleteVerifier is a paid mutator transaction binding the contract method 0x55f32817.
//
// Solidity: function deleteVerifier(address verifier) returns()
func (_SubnetVerifier *SubnetVerifierSession) DeleteVerifier(verifier common.Address) (*types.Transaction, error) {
	return _SubnetVerifier.Contract.DeleteVerifier(&_SubnetVerifier.TransactOpts, verifier)
}

// DeleteVerifier is a paid mutator transaction binding the contract method 0x55f32817.
//
// Solidity: function deleteVerifier(address verifier) returns()
func (_SubnetVerifier *SubnetVerifierTransactorSession) DeleteVerifier(verifier common.Address) (*types.Transaction, error) {
	return _SubnetVerifier.Contract.DeleteVerifier(&_SubnetVerifier.TransactOpts, verifier)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _initialOwner) returns()
func (_SubnetVerifier *SubnetVerifierTransactor) Initialize(opts *bind.TransactOpts, _initialOwner common.Address) (*types.Transaction, error) {
	return _SubnetVerifier.contract.Transact(opts, "initialize", _initialOwner)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _initialOwner) returns()
func (_SubnetVerifier *SubnetVerifierSession) Initialize(_initialOwner common.Address) (*types.Transaction, error) {
	return _SubnetVerifier.Contract.Initialize(&_SubnetVerifier.TransactOpts, _initialOwner)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _initialOwner) returns()
func (_SubnetVerifier *SubnetVerifierTransactorSession) Initialize(_initialOwner common.Address) (*types.Transaction, error) {
	return _SubnetVerifier.Contract.Initialize(&_SubnetVerifier.TransactOpts, _initialOwner)
}

// RegisterVerifier is a paid mutator transaction binding the contract method 0x0e1fb4ca.
//
// Solidity: function registerVerifier(address verifier, string[] peerIds) returns()
func (_SubnetVerifier *SubnetVerifierTransactor) RegisterVerifier(opts *bind.TransactOpts, verifier common.Address, peerIds []string) (*types.Transaction, error) {
	return _SubnetVerifier.contract.Transact(opts, "registerVerifier", verifier, peerIds)
}

// RegisterVerifier is a paid mutator transaction binding the contract method 0x0e1fb4ca.
//
// Solidity: function registerVerifier(address verifier, string[] peerIds) returns()
func (_SubnetVerifier *SubnetVerifierSession) RegisterVerifier(verifier common.Address, peerIds []string) (*types.Transaction, error) {
	return _SubnetVerifier.Contract.RegisterVerifier(&_SubnetVerifier.TransactOpts, verifier, peerIds)
}

// RegisterVerifier is a paid mutator transaction binding the contract method 0x0e1fb4ca.
//
// Solidity: function registerVerifier(address verifier, string[] peerIds) returns()
func (_SubnetVerifier *SubnetVerifierTransactorSession) RegisterVerifier(verifier common.Address, peerIds []string) (*types.Transaction, error) {
	return _SubnetVerifier.Contract.RegisterVerifier(&_SubnetVerifier.TransactOpts, verifier, peerIds)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetVerifier *SubnetVerifierTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetVerifier.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetVerifier *SubnetVerifierSession) RenounceOwnership() (*types.Transaction, error) {
	return _SubnetVerifier.Contract.RenounceOwnership(&_SubnetVerifier.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetVerifier *SubnetVerifierTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _SubnetVerifier.Contract.RenounceOwnership(&_SubnetVerifier.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetVerifier *SubnetVerifierTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _SubnetVerifier.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetVerifier *SubnetVerifierSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SubnetVerifier.Contract.TransferOwnership(&_SubnetVerifier.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetVerifier *SubnetVerifierTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SubnetVerifier.Contract.TransferOwnership(&_SubnetVerifier.TransactOpts, newOwner)
}

// UpdateVerifierPeers is a paid mutator transaction binding the contract method 0x1c4dada3.
//
// Solidity: function updateVerifierPeers(address verifier, string[] peerIds) returns()
func (_SubnetVerifier *SubnetVerifierTransactor) UpdateVerifierPeers(opts *bind.TransactOpts, verifier common.Address, peerIds []string) (*types.Transaction, error) {
	return _SubnetVerifier.contract.Transact(opts, "updateVerifierPeers", verifier, peerIds)
}

// UpdateVerifierPeers is a paid mutator transaction binding the contract method 0x1c4dada3.
//
// Solidity: function updateVerifierPeers(address verifier, string[] peerIds) returns()
func (_SubnetVerifier *SubnetVerifierSession) UpdateVerifierPeers(verifier common.Address, peerIds []string) (*types.Transaction, error) {
	return _SubnetVerifier.Contract.UpdateVerifierPeers(&_SubnetVerifier.TransactOpts, verifier, peerIds)
}

// UpdateVerifierPeers is a paid mutator transaction binding the contract method 0x1c4dada3.
//
// Solidity: function updateVerifierPeers(address verifier, string[] peerIds) returns()
func (_SubnetVerifier *SubnetVerifierTransactorSession) UpdateVerifierPeers(verifier common.Address, peerIds []string) (*types.Transaction, error) {
	return _SubnetVerifier.Contract.UpdateVerifierPeers(&_SubnetVerifier.TransactOpts, verifier, peerIds)
}

// SubnetVerifierInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the SubnetVerifier contract.
type SubnetVerifierInitializedIterator struct {
	Event *SubnetVerifierInitialized // Event containing the contract specifics and raw log

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
func (it *SubnetVerifierInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetVerifierInitialized)
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
		it.Event = new(SubnetVerifierInitialized)
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
func (it *SubnetVerifierInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetVerifierInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetVerifierInitialized represents a Initialized event raised by the SubnetVerifier contract.
type SubnetVerifierInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_SubnetVerifier *SubnetVerifierFilterer) FilterInitialized(opts *bind.FilterOpts) (*SubnetVerifierInitializedIterator, error) {

	logs, sub, err := _SubnetVerifier.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &SubnetVerifierInitializedIterator{contract: _SubnetVerifier.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_SubnetVerifier *SubnetVerifierFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *SubnetVerifierInitialized) (event.Subscription, error) {

	logs, sub, err := _SubnetVerifier.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetVerifierInitialized)
				if err := _SubnetVerifier.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_SubnetVerifier *SubnetVerifierFilterer) ParseInitialized(log types.Log) (*SubnetVerifierInitialized, error) {
	event := new(SubnetVerifierInitialized)
	if err := _SubnetVerifier.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetVerifierOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the SubnetVerifier contract.
type SubnetVerifierOwnershipTransferredIterator struct {
	Event *SubnetVerifierOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *SubnetVerifierOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetVerifierOwnershipTransferred)
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
		it.Event = new(SubnetVerifierOwnershipTransferred)
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
func (it *SubnetVerifierOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetVerifierOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetVerifierOwnershipTransferred represents a OwnershipTransferred event raised by the SubnetVerifier contract.
type SubnetVerifierOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SubnetVerifier *SubnetVerifierFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*SubnetVerifierOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SubnetVerifier.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &SubnetVerifierOwnershipTransferredIterator{contract: _SubnetVerifier.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SubnetVerifier *SubnetVerifierFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *SubnetVerifierOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SubnetVerifier.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetVerifierOwnershipTransferred)
				if err := _SubnetVerifier.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_SubnetVerifier *SubnetVerifierFilterer) ParseOwnershipTransferred(log types.Log) (*SubnetVerifierOwnershipTransferred, error) {
	event := new(SubnetVerifierOwnershipTransferred)
	if err := _SubnetVerifier.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetVerifierVerifierDeletedIterator is returned from FilterVerifierDeleted and is used to iterate over the raw logs and unpacked data for VerifierDeleted events raised by the SubnetVerifier contract.
type SubnetVerifierVerifierDeletedIterator struct {
	Event *SubnetVerifierVerifierDeleted // Event containing the contract specifics and raw log

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
func (it *SubnetVerifierVerifierDeletedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetVerifierVerifierDeleted)
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
		it.Event = new(SubnetVerifierVerifierDeleted)
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
func (it *SubnetVerifierVerifierDeletedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetVerifierVerifierDeletedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetVerifierVerifierDeleted represents a VerifierDeleted event raised by the SubnetVerifier contract.
type SubnetVerifierVerifierDeleted struct {
	Verifier common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterVerifierDeleted is a free log retrieval operation binding the contract event 0x179d784bd033412e105437d6852ec26971ad53d1d4d2c3a05c1a1809ee55bc16.
//
// Solidity: event VerifierDeleted(address indexed verifier)
func (_SubnetVerifier *SubnetVerifierFilterer) FilterVerifierDeleted(opts *bind.FilterOpts, verifier []common.Address) (*SubnetVerifierVerifierDeletedIterator, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _SubnetVerifier.contract.FilterLogs(opts, "VerifierDeleted", verifierRule)
	if err != nil {
		return nil, err
	}
	return &SubnetVerifierVerifierDeletedIterator{contract: _SubnetVerifier.contract, event: "VerifierDeleted", logs: logs, sub: sub}, nil
}

// WatchVerifierDeleted is a free log subscription operation binding the contract event 0x179d784bd033412e105437d6852ec26971ad53d1d4d2c3a05c1a1809ee55bc16.
//
// Solidity: event VerifierDeleted(address indexed verifier)
func (_SubnetVerifier *SubnetVerifierFilterer) WatchVerifierDeleted(opts *bind.WatchOpts, sink chan<- *SubnetVerifierVerifierDeleted, verifier []common.Address) (event.Subscription, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _SubnetVerifier.contract.WatchLogs(opts, "VerifierDeleted", verifierRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetVerifierVerifierDeleted)
				if err := _SubnetVerifier.contract.UnpackLog(event, "VerifierDeleted", log); err != nil {
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

// ParseVerifierDeleted is a log parse operation binding the contract event 0x179d784bd033412e105437d6852ec26971ad53d1d4d2c3a05c1a1809ee55bc16.
//
// Solidity: event VerifierDeleted(address indexed verifier)
func (_SubnetVerifier *SubnetVerifierFilterer) ParseVerifierDeleted(log types.Log) (*SubnetVerifierVerifierDeleted, error) {
	event := new(SubnetVerifierVerifierDeleted)
	if err := _SubnetVerifier.contract.UnpackLog(event, "VerifierDeleted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetVerifierVerifierPeersUpdatedIterator is returned from FilterVerifierPeersUpdated and is used to iterate over the raw logs and unpacked data for VerifierPeersUpdated events raised by the SubnetVerifier contract.
type SubnetVerifierVerifierPeersUpdatedIterator struct {
	Event *SubnetVerifierVerifierPeersUpdated // Event containing the contract specifics and raw log

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
func (it *SubnetVerifierVerifierPeersUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetVerifierVerifierPeersUpdated)
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
		it.Event = new(SubnetVerifierVerifierPeersUpdated)
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
func (it *SubnetVerifierVerifierPeersUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetVerifierVerifierPeersUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetVerifierVerifierPeersUpdated represents a VerifierPeersUpdated event raised by the SubnetVerifier contract.
type SubnetVerifierVerifierPeersUpdated struct {
	Verifier common.Address
	PeerIds  []string
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterVerifierPeersUpdated is a free log retrieval operation binding the contract event 0x0f671e18862549ec4d8019cc59f3ab1cfef01ea6b7f5f1b73bef2ed1bebc5876.
//
// Solidity: event VerifierPeersUpdated(address indexed verifier, string[] peerIds)
func (_SubnetVerifier *SubnetVerifierFilterer) FilterVerifierPeersUpdated(opts *bind.FilterOpts, verifier []common.Address) (*SubnetVerifierVerifierPeersUpdatedIterator, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _SubnetVerifier.contract.FilterLogs(opts, "VerifierPeersUpdated", verifierRule)
	if err != nil {
		return nil, err
	}
	return &SubnetVerifierVerifierPeersUpdatedIterator{contract: _SubnetVerifier.contract, event: "VerifierPeersUpdated", logs: logs, sub: sub}, nil
}

// WatchVerifierPeersUpdated is a free log subscription operation binding the contract event 0x0f671e18862549ec4d8019cc59f3ab1cfef01ea6b7f5f1b73bef2ed1bebc5876.
//
// Solidity: event VerifierPeersUpdated(address indexed verifier, string[] peerIds)
func (_SubnetVerifier *SubnetVerifierFilterer) WatchVerifierPeersUpdated(opts *bind.WatchOpts, sink chan<- *SubnetVerifierVerifierPeersUpdated, verifier []common.Address) (event.Subscription, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _SubnetVerifier.contract.WatchLogs(opts, "VerifierPeersUpdated", verifierRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetVerifierVerifierPeersUpdated)
				if err := _SubnetVerifier.contract.UnpackLog(event, "VerifierPeersUpdated", log); err != nil {
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

// ParseVerifierPeersUpdated is a log parse operation binding the contract event 0x0f671e18862549ec4d8019cc59f3ab1cfef01ea6b7f5f1b73bef2ed1bebc5876.
//
// Solidity: event VerifierPeersUpdated(address indexed verifier, string[] peerIds)
func (_SubnetVerifier *SubnetVerifierFilterer) ParseVerifierPeersUpdated(log types.Log) (*SubnetVerifierVerifierPeersUpdated, error) {
	event := new(SubnetVerifierVerifierPeersUpdated)
	if err := _SubnetVerifier.contract.UnpackLog(event, "VerifierPeersUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetVerifierVerifierRegisteredIterator is returned from FilterVerifierRegistered and is used to iterate over the raw logs and unpacked data for VerifierRegistered events raised by the SubnetVerifier contract.
type SubnetVerifierVerifierRegisteredIterator struct {
	Event *SubnetVerifierVerifierRegistered // Event containing the contract specifics and raw log

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
func (it *SubnetVerifierVerifierRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetVerifierVerifierRegistered)
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
		it.Event = new(SubnetVerifierVerifierRegistered)
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
func (it *SubnetVerifierVerifierRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetVerifierVerifierRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetVerifierVerifierRegistered represents a VerifierRegistered event raised by the SubnetVerifier contract.
type SubnetVerifierVerifierRegistered struct {
	Verifier common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterVerifierRegistered is a free log retrieval operation binding the contract event 0x85cf819da48b7e3c138cca9531adae41f25faeff5801e9e096eb2c1fb3cbbab7.
//
// Solidity: event VerifierRegistered(address indexed verifier)
func (_SubnetVerifier *SubnetVerifierFilterer) FilterVerifierRegistered(opts *bind.FilterOpts, verifier []common.Address) (*SubnetVerifierVerifierRegisteredIterator, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _SubnetVerifier.contract.FilterLogs(opts, "VerifierRegistered", verifierRule)
	if err != nil {
		return nil, err
	}
	return &SubnetVerifierVerifierRegisteredIterator{contract: _SubnetVerifier.contract, event: "VerifierRegistered", logs: logs, sub: sub}, nil
}

// WatchVerifierRegistered is a free log subscription operation binding the contract event 0x85cf819da48b7e3c138cca9531adae41f25faeff5801e9e096eb2c1fb3cbbab7.
//
// Solidity: event VerifierRegistered(address indexed verifier)
func (_SubnetVerifier *SubnetVerifierFilterer) WatchVerifierRegistered(opts *bind.WatchOpts, sink chan<- *SubnetVerifierVerifierRegistered, verifier []common.Address) (event.Subscription, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _SubnetVerifier.contract.WatchLogs(opts, "VerifierRegistered", verifierRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetVerifierVerifierRegistered)
				if err := _SubnetVerifier.contract.UnpackLog(event, "VerifierRegistered", log); err != nil {
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

// ParseVerifierRegistered is a log parse operation binding the contract event 0x85cf819da48b7e3c138cca9531adae41f25faeff5801e9e096eb2c1fb3cbbab7.
//
// Solidity: event VerifierRegistered(address indexed verifier)
func (_SubnetVerifier *SubnetVerifierFilterer) ParseVerifierRegistered(log types.Log) (*SubnetVerifierVerifierRegistered, error) {
	event := new(SubnetVerifierVerifierRegistered)
	if err := _SubnetVerifier.contract.UnpackLog(event, "VerifierRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
