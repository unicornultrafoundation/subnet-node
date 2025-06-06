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

// SubnetDeploymentDeployment is an auto generated low-level Go binding around an user-defined struct.
type SubnetDeploymentDeployment struct {
	DockerConfig string
	Owner        common.Address
	CreatedAt    *big.Int
	UpdatedAt    *big.Int
}

// SubnetDeploymentMetaData contains all meta data concerning the SubnetDeployment contract.
var SubnetDeploymentMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeIp\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"SubnetDeployed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeIp\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"SubnetDeploymentDeleted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeIp\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"dockerConfig\",\"type\":\"string\"}],\"name\":\"SubnetDeploymentUpdated\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"appStore\",\"outputs\":[{\"internalType\":\"contractSubnetAppStore\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"uint256[]\",\"name\":\"nodeIps\",\"type\":\"uint256[]\"}],\"name\":\"batchDeleteDeployments\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nodeIp\",\"type\":\"uint256\"}],\"name\":\"deleteDeployment\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nodeIp\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"dockerConfig\",\"type\":\"string\"}],\"name\":\"deploySubnet\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"dockerConfig\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"updatedAt\",\"type\":\"uint256\"}],\"internalType\":\"structSubnetDeployment.Deployment\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nodeIp\",\"type\":\"uint256\"}],\"name\":\"deploymentExists\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nodeIp\",\"type\":\"uint256\"}],\"name\":\"getDeployment\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"dockerConfig\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"updatedAt\",\"type\":\"uint256\"}],\"internalType\":\"structSubnetDeployment.Deployment\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nodeIp\",\"type\":\"uint256\"}],\"name\":\"getDeploymentDuration\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"uint256[]\",\"name\":\"nodeIps\",\"type\":\"uint256[]\"}],\"name\":\"getDeployments\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"dockerConfig\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"updatedAt\",\"type\":\"uint256\"}],\"internalType\":\"structSubnetDeployment.Deployment[]\",\"name\":\"deployments\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nodeIp\",\"type\":\"uint256\"}],\"name\":\"getTimeSinceUpdate\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"timeSinceUpdate\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_appStore\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nodeIp\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"dockerConfig\",\"type\":\"string\"}],\"name\":\"updateDeployment\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// SubnetDeploymentABI is the input ABI used to generate the binding from.
// Deprecated: Use SubnetDeploymentMetaData.ABI instead.
var SubnetDeploymentABI = SubnetDeploymentMetaData.ABI

// SubnetDeployment is an auto generated Go binding around an Ethereum contract.
type SubnetDeployment struct {
	SubnetDeploymentCaller     // Read-only binding to the contract
	SubnetDeploymentTransactor // Write-only binding to the contract
	SubnetDeploymentFilterer   // Log filterer for contract events
}

// SubnetDeploymentCaller is an auto generated read-only Go binding around an Ethereum contract.
type SubnetDeploymentCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetDeploymentTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SubnetDeploymentTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetDeploymentFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SubnetDeploymentFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetDeploymentSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SubnetDeploymentSession struct {
	Contract     *SubnetDeployment // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SubnetDeploymentCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SubnetDeploymentCallerSession struct {
	Contract *SubnetDeploymentCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// SubnetDeploymentTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SubnetDeploymentTransactorSession struct {
	Contract     *SubnetDeploymentTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// SubnetDeploymentRaw is an auto generated low-level Go binding around an Ethereum contract.
type SubnetDeploymentRaw struct {
	Contract *SubnetDeployment // Generic contract binding to access the raw methods on
}

// SubnetDeploymentCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SubnetDeploymentCallerRaw struct {
	Contract *SubnetDeploymentCaller // Generic read-only contract binding to access the raw methods on
}

// SubnetDeploymentTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SubnetDeploymentTransactorRaw struct {
	Contract *SubnetDeploymentTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSubnetDeployment creates a new instance of SubnetDeployment, bound to a specific deployed contract.
func NewSubnetDeployment(address common.Address, backend bind.ContractBackend) (*SubnetDeployment, error) {
	contract, err := bindSubnetDeployment(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SubnetDeployment{SubnetDeploymentCaller: SubnetDeploymentCaller{contract: contract}, SubnetDeploymentTransactor: SubnetDeploymentTransactor{contract: contract}, SubnetDeploymentFilterer: SubnetDeploymentFilterer{contract: contract}}, nil
}

// NewSubnetDeploymentCaller creates a new read-only instance of SubnetDeployment, bound to a specific deployed contract.
func NewSubnetDeploymentCaller(address common.Address, caller bind.ContractCaller) (*SubnetDeploymentCaller, error) {
	contract, err := bindSubnetDeployment(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SubnetDeploymentCaller{contract: contract}, nil
}

// NewSubnetDeploymentTransactor creates a new write-only instance of SubnetDeployment, bound to a specific deployed contract.
func NewSubnetDeploymentTransactor(address common.Address, transactor bind.ContractTransactor) (*SubnetDeploymentTransactor, error) {
	contract, err := bindSubnetDeployment(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SubnetDeploymentTransactor{contract: contract}, nil
}

// NewSubnetDeploymentFilterer creates a new log filterer instance of SubnetDeployment, bound to a specific deployed contract.
func NewSubnetDeploymentFilterer(address common.Address, filterer bind.ContractFilterer) (*SubnetDeploymentFilterer, error) {
	contract, err := bindSubnetDeployment(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SubnetDeploymentFilterer{contract: contract}, nil
}

// bindSubnetDeployment binds a generic wrapper to an already deployed contract.
func bindSubnetDeployment(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SubnetDeploymentMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SubnetDeployment *SubnetDeploymentRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SubnetDeployment.Contract.SubnetDeploymentCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SubnetDeployment *SubnetDeploymentRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetDeployment.Contract.SubnetDeploymentTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SubnetDeployment *SubnetDeploymentRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SubnetDeployment.Contract.SubnetDeploymentTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SubnetDeployment *SubnetDeploymentCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SubnetDeployment.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SubnetDeployment *SubnetDeploymentTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetDeployment.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SubnetDeployment *SubnetDeploymentTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SubnetDeployment.Contract.contract.Transact(opts, method, params...)
}

// AppStore is a free data retrieval call binding the contract method 0x72f86f05.
//
// Solidity: function appStore() view returns(address)
func (_SubnetDeployment *SubnetDeploymentCaller) AppStore(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetDeployment.contract.Call(opts, &out, "appStore")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AppStore is a free data retrieval call binding the contract method 0x72f86f05.
//
// Solidity: function appStore() view returns(address)
func (_SubnetDeployment *SubnetDeploymentSession) AppStore() (common.Address, error) {
	return _SubnetDeployment.Contract.AppStore(&_SubnetDeployment.CallOpts)
}

// AppStore is a free data retrieval call binding the contract method 0x72f86f05.
//
// Solidity: function appStore() view returns(address)
func (_SubnetDeployment *SubnetDeploymentCallerSession) AppStore() (common.Address, error) {
	return _SubnetDeployment.Contract.AppStore(&_SubnetDeployment.CallOpts)
}

// DeploymentExists is a free data retrieval call binding the contract method 0xe5f188a1.
//
// Solidity: function deploymentExists(uint256 appId, uint256 nodeIp) view returns(bool)
func (_SubnetDeployment *SubnetDeploymentCaller) DeploymentExists(opts *bind.CallOpts, appId *big.Int, nodeIp *big.Int) (bool, error) {
	var out []interface{}
	err := _SubnetDeployment.contract.Call(opts, &out, "deploymentExists", appId, nodeIp)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// DeploymentExists is a free data retrieval call binding the contract method 0xe5f188a1.
//
// Solidity: function deploymentExists(uint256 appId, uint256 nodeIp) view returns(bool)
func (_SubnetDeployment *SubnetDeploymentSession) DeploymentExists(appId *big.Int, nodeIp *big.Int) (bool, error) {
	return _SubnetDeployment.Contract.DeploymentExists(&_SubnetDeployment.CallOpts, appId, nodeIp)
}

// DeploymentExists is a free data retrieval call binding the contract method 0xe5f188a1.
//
// Solidity: function deploymentExists(uint256 appId, uint256 nodeIp) view returns(bool)
func (_SubnetDeployment *SubnetDeploymentCallerSession) DeploymentExists(appId *big.Int, nodeIp *big.Int) (bool, error) {
	return _SubnetDeployment.Contract.DeploymentExists(&_SubnetDeployment.CallOpts, appId, nodeIp)
}

// GetDeployment is a free data retrieval call binding the contract method 0x8b10fed6.
//
// Solidity: function getDeployment(uint256 appId, uint256 nodeIp) view returns((string,address,uint256,uint256))
func (_SubnetDeployment *SubnetDeploymentCaller) GetDeployment(opts *bind.CallOpts, appId *big.Int, nodeIp *big.Int) (SubnetDeploymentDeployment, error) {
	var out []interface{}
	err := _SubnetDeployment.contract.Call(opts, &out, "getDeployment", appId, nodeIp)

	if err != nil {
		return *new(SubnetDeploymentDeployment), err
	}

	out0 := *abi.ConvertType(out[0], new(SubnetDeploymentDeployment)).(*SubnetDeploymentDeployment)

	return out0, err

}

// GetDeployment is a free data retrieval call binding the contract method 0x8b10fed6.
//
// Solidity: function getDeployment(uint256 appId, uint256 nodeIp) view returns((string,address,uint256,uint256))
func (_SubnetDeployment *SubnetDeploymentSession) GetDeployment(appId *big.Int, nodeIp *big.Int) (SubnetDeploymentDeployment, error) {
	return _SubnetDeployment.Contract.GetDeployment(&_SubnetDeployment.CallOpts, appId, nodeIp)
}

// GetDeployment is a free data retrieval call binding the contract method 0x8b10fed6.
//
// Solidity: function getDeployment(uint256 appId, uint256 nodeIp) view returns((string,address,uint256,uint256))
func (_SubnetDeployment *SubnetDeploymentCallerSession) GetDeployment(appId *big.Int, nodeIp *big.Int) (SubnetDeploymentDeployment, error) {
	return _SubnetDeployment.Contract.GetDeployment(&_SubnetDeployment.CallOpts, appId, nodeIp)
}

// GetDeploymentDuration is a free data retrieval call binding the contract method 0x5a1801e1.
//
// Solidity: function getDeploymentDuration(uint256 appId, uint256 nodeIp) view returns(uint256 duration)
func (_SubnetDeployment *SubnetDeploymentCaller) GetDeploymentDuration(opts *bind.CallOpts, appId *big.Int, nodeIp *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SubnetDeployment.contract.Call(opts, &out, "getDeploymentDuration", appId, nodeIp)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetDeploymentDuration is a free data retrieval call binding the contract method 0x5a1801e1.
//
// Solidity: function getDeploymentDuration(uint256 appId, uint256 nodeIp) view returns(uint256 duration)
func (_SubnetDeployment *SubnetDeploymentSession) GetDeploymentDuration(appId *big.Int, nodeIp *big.Int) (*big.Int, error) {
	return _SubnetDeployment.Contract.GetDeploymentDuration(&_SubnetDeployment.CallOpts, appId, nodeIp)
}

// GetDeploymentDuration is a free data retrieval call binding the contract method 0x5a1801e1.
//
// Solidity: function getDeploymentDuration(uint256 appId, uint256 nodeIp) view returns(uint256 duration)
func (_SubnetDeployment *SubnetDeploymentCallerSession) GetDeploymentDuration(appId *big.Int, nodeIp *big.Int) (*big.Int, error) {
	return _SubnetDeployment.Contract.GetDeploymentDuration(&_SubnetDeployment.CallOpts, appId, nodeIp)
}

// GetDeployments is a free data retrieval call binding the contract method 0x64129b48.
//
// Solidity: function getDeployments(uint256 appId, uint256[] nodeIps) view returns((string,address,uint256,uint256)[] deployments)
func (_SubnetDeployment *SubnetDeploymentCaller) GetDeployments(opts *bind.CallOpts, appId *big.Int, nodeIps []*big.Int) ([]SubnetDeploymentDeployment, error) {
	var out []interface{}
	err := _SubnetDeployment.contract.Call(opts, &out, "getDeployments", appId, nodeIps)

	if err != nil {
		return *new([]SubnetDeploymentDeployment), err
	}

	out0 := *abi.ConvertType(out[0], new([]SubnetDeploymentDeployment)).(*[]SubnetDeploymentDeployment)

	return out0, err

}

// GetDeployments is a free data retrieval call binding the contract method 0x64129b48.
//
// Solidity: function getDeployments(uint256 appId, uint256[] nodeIps) view returns((string,address,uint256,uint256)[] deployments)
func (_SubnetDeployment *SubnetDeploymentSession) GetDeployments(appId *big.Int, nodeIps []*big.Int) ([]SubnetDeploymentDeployment, error) {
	return _SubnetDeployment.Contract.GetDeployments(&_SubnetDeployment.CallOpts, appId, nodeIps)
}

// GetDeployments is a free data retrieval call binding the contract method 0x64129b48.
//
// Solidity: function getDeployments(uint256 appId, uint256[] nodeIps) view returns((string,address,uint256,uint256)[] deployments)
func (_SubnetDeployment *SubnetDeploymentCallerSession) GetDeployments(appId *big.Int, nodeIps []*big.Int) ([]SubnetDeploymentDeployment, error) {
	return _SubnetDeployment.Contract.GetDeployments(&_SubnetDeployment.CallOpts, appId, nodeIps)
}

// GetTimeSinceUpdate is a free data retrieval call binding the contract method 0xdabe3544.
//
// Solidity: function getTimeSinceUpdate(uint256 appId, uint256 nodeIp) view returns(uint256 timeSinceUpdate)
func (_SubnetDeployment *SubnetDeploymentCaller) GetTimeSinceUpdate(opts *bind.CallOpts, appId *big.Int, nodeIp *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SubnetDeployment.contract.Call(opts, &out, "getTimeSinceUpdate", appId, nodeIp)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTimeSinceUpdate is a free data retrieval call binding the contract method 0xdabe3544.
//
// Solidity: function getTimeSinceUpdate(uint256 appId, uint256 nodeIp) view returns(uint256 timeSinceUpdate)
func (_SubnetDeployment *SubnetDeploymentSession) GetTimeSinceUpdate(appId *big.Int, nodeIp *big.Int) (*big.Int, error) {
	return _SubnetDeployment.Contract.GetTimeSinceUpdate(&_SubnetDeployment.CallOpts, appId, nodeIp)
}

// GetTimeSinceUpdate is a free data retrieval call binding the contract method 0xdabe3544.
//
// Solidity: function getTimeSinceUpdate(uint256 appId, uint256 nodeIp) view returns(uint256 timeSinceUpdate)
func (_SubnetDeployment *SubnetDeploymentCallerSession) GetTimeSinceUpdate(appId *big.Int, nodeIp *big.Int) (*big.Int, error) {
	return _SubnetDeployment.Contract.GetTimeSinceUpdate(&_SubnetDeployment.CallOpts, appId, nodeIp)
}

// BatchDeleteDeployments is a paid mutator transaction binding the contract method 0x2416fc55.
//
// Solidity: function batchDeleteDeployments(uint256 appId, uint256[] nodeIps) returns()
func (_SubnetDeployment *SubnetDeploymentTransactor) BatchDeleteDeployments(opts *bind.TransactOpts, appId *big.Int, nodeIps []*big.Int) (*types.Transaction, error) {
	return _SubnetDeployment.contract.Transact(opts, "batchDeleteDeployments", appId, nodeIps)
}

// BatchDeleteDeployments is a paid mutator transaction binding the contract method 0x2416fc55.
//
// Solidity: function batchDeleteDeployments(uint256 appId, uint256[] nodeIps) returns()
func (_SubnetDeployment *SubnetDeploymentSession) BatchDeleteDeployments(appId *big.Int, nodeIps []*big.Int) (*types.Transaction, error) {
	return _SubnetDeployment.Contract.BatchDeleteDeployments(&_SubnetDeployment.TransactOpts, appId, nodeIps)
}

// BatchDeleteDeployments is a paid mutator transaction binding the contract method 0x2416fc55.
//
// Solidity: function batchDeleteDeployments(uint256 appId, uint256[] nodeIps) returns()
func (_SubnetDeployment *SubnetDeploymentTransactorSession) BatchDeleteDeployments(appId *big.Int, nodeIps []*big.Int) (*types.Transaction, error) {
	return _SubnetDeployment.Contract.BatchDeleteDeployments(&_SubnetDeployment.TransactOpts, appId, nodeIps)
}

// DeleteDeployment is a paid mutator transaction binding the contract method 0x016180a9.
//
// Solidity: function deleteDeployment(uint256 appId, uint256 nodeIp) returns()
func (_SubnetDeployment *SubnetDeploymentTransactor) DeleteDeployment(opts *bind.TransactOpts, appId *big.Int, nodeIp *big.Int) (*types.Transaction, error) {
	return _SubnetDeployment.contract.Transact(opts, "deleteDeployment", appId, nodeIp)
}

// DeleteDeployment is a paid mutator transaction binding the contract method 0x016180a9.
//
// Solidity: function deleteDeployment(uint256 appId, uint256 nodeIp) returns()
func (_SubnetDeployment *SubnetDeploymentSession) DeleteDeployment(appId *big.Int, nodeIp *big.Int) (*types.Transaction, error) {
	return _SubnetDeployment.Contract.DeleteDeployment(&_SubnetDeployment.TransactOpts, appId, nodeIp)
}

// DeleteDeployment is a paid mutator transaction binding the contract method 0x016180a9.
//
// Solidity: function deleteDeployment(uint256 appId, uint256 nodeIp) returns()
func (_SubnetDeployment *SubnetDeploymentTransactorSession) DeleteDeployment(appId *big.Int, nodeIp *big.Int) (*types.Transaction, error) {
	return _SubnetDeployment.Contract.DeleteDeployment(&_SubnetDeployment.TransactOpts, appId, nodeIp)
}

// DeploySubnet is a paid mutator transaction binding the contract method 0x9411654e.
//
// Solidity: function deploySubnet(uint256 appId, uint256 nodeIp, string dockerConfig) returns((string,address,uint256,uint256))
func (_SubnetDeployment *SubnetDeploymentTransactor) DeploySubnet(opts *bind.TransactOpts, appId *big.Int, nodeIp *big.Int, dockerConfig string) (*types.Transaction, error) {
	return _SubnetDeployment.contract.Transact(opts, "deploySubnet", appId, nodeIp, dockerConfig)
}

// DeploySubnet is a paid mutator transaction binding the contract method 0x9411654e.
//
// Solidity: function deploySubnet(uint256 appId, uint256 nodeIp, string dockerConfig) returns((string,address,uint256,uint256))
func (_SubnetDeployment *SubnetDeploymentSession) DeploySubnet(appId *big.Int, nodeIp *big.Int, dockerConfig string) (*types.Transaction, error) {
	return _SubnetDeployment.Contract.DeploySubnet(&_SubnetDeployment.TransactOpts, appId, nodeIp, dockerConfig)
}

// DeploySubnet is a paid mutator transaction binding the contract method 0x9411654e.
//
// Solidity: function deploySubnet(uint256 appId, uint256 nodeIp, string dockerConfig) returns((string,address,uint256,uint256))
func (_SubnetDeployment *SubnetDeploymentTransactorSession) DeploySubnet(appId *big.Int, nodeIp *big.Int, dockerConfig string) (*types.Transaction, error) {
	return _SubnetDeployment.Contract.DeploySubnet(&_SubnetDeployment.TransactOpts, appId, nodeIp, dockerConfig)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _appStore) returns()
func (_SubnetDeployment *SubnetDeploymentTransactor) Initialize(opts *bind.TransactOpts, _appStore common.Address) (*types.Transaction, error) {
	return _SubnetDeployment.contract.Transact(opts, "initialize", _appStore)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _appStore) returns()
func (_SubnetDeployment *SubnetDeploymentSession) Initialize(_appStore common.Address) (*types.Transaction, error) {
	return _SubnetDeployment.Contract.Initialize(&_SubnetDeployment.TransactOpts, _appStore)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _appStore) returns()
func (_SubnetDeployment *SubnetDeploymentTransactorSession) Initialize(_appStore common.Address) (*types.Transaction, error) {
	return _SubnetDeployment.Contract.Initialize(&_SubnetDeployment.TransactOpts, _appStore)
}

// UpdateDeployment is a paid mutator transaction binding the contract method 0x4374538e.
//
// Solidity: function updateDeployment(uint256 appId, uint256 nodeIp, string dockerConfig) returns()
func (_SubnetDeployment *SubnetDeploymentTransactor) UpdateDeployment(opts *bind.TransactOpts, appId *big.Int, nodeIp *big.Int, dockerConfig string) (*types.Transaction, error) {
	return _SubnetDeployment.contract.Transact(opts, "updateDeployment", appId, nodeIp, dockerConfig)
}

// UpdateDeployment is a paid mutator transaction binding the contract method 0x4374538e.
//
// Solidity: function updateDeployment(uint256 appId, uint256 nodeIp, string dockerConfig) returns()
func (_SubnetDeployment *SubnetDeploymentSession) UpdateDeployment(appId *big.Int, nodeIp *big.Int, dockerConfig string) (*types.Transaction, error) {
	return _SubnetDeployment.Contract.UpdateDeployment(&_SubnetDeployment.TransactOpts, appId, nodeIp, dockerConfig)
}

// UpdateDeployment is a paid mutator transaction binding the contract method 0x4374538e.
//
// Solidity: function updateDeployment(uint256 appId, uint256 nodeIp, string dockerConfig) returns()
func (_SubnetDeployment *SubnetDeploymentTransactorSession) UpdateDeployment(appId *big.Int, nodeIp *big.Int, dockerConfig string) (*types.Transaction, error) {
	return _SubnetDeployment.Contract.UpdateDeployment(&_SubnetDeployment.TransactOpts, appId, nodeIp, dockerConfig)
}

// SubnetDeploymentInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the SubnetDeployment contract.
type SubnetDeploymentInitializedIterator struct {
	Event *SubnetDeploymentInitialized // Event containing the contract specifics and raw log

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
func (it *SubnetDeploymentInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetDeploymentInitialized)
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
		it.Event = new(SubnetDeploymentInitialized)
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
func (it *SubnetDeploymentInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetDeploymentInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetDeploymentInitialized represents a Initialized event raised by the SubnetDeployment contract.
type SubnetDeploymentInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_SubnetDeployment *SubnetDeploymentFilterer) FilterInitialized(opts *bind.FilterOpts) (*SubnetDeploymentInitializedIterator, error) {

	logs, sub, err := _SubnetDeployment.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &SubnetDeploymentInitializedIterator{contract: _SubnetDeployment.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_SubnetDeployment *SubnetDeploymentFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *SubnetDeploymentInitialized) (event.Subscription, error) {

	logs, sub, err := _SubnetDeployment.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetDeploymentInitialized)
				if err := _SubnetDeployment.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_SubnetDeployment *SubnetDeploymentFilterer) ParseInitialized(log types.Log) (*SubnetDeploymentInitialized, error) {
	event := new(SubnetDeploymentInitialized)
	if err := _SubnetDeployment.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetDeploymentSubnetDeployedIterator is returned from FilterSubnetDeployed and is used to iterate over the raw logs and unpacked data for SubnetDeployed events raised by the SubnetDeployment contract.
type SubnetDeploymentSubnetDeployedIterator struct {
	Event *SubnetDeploymentSubnetDeployed // Event containing the contract specifics and raw log

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
func (it *SubnetDeploymentSubnetDeployedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetDeploymentSubnetDeployed)
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
		it.Event = new(SubnetDeploymentSubnetDeployed)
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
func (it *SubnetDeploymentSubnetDeployedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetDeploymentSubnetDeployedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetDeploymentSubnetDeployed represents a SubnetDeployed event raised by the SubnetDeployment contract.
type SubnetDeploymentSubnetDeployed struct {
	AppId  *big.Int
	NodeIp *big.Int
	Owner  common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterSubnetDeployed is a free log retrieval operation binding the contract event 0x2205f96223e007c4bf7b7e81320ac9ff0bbd8393ae72ac518c59b4814c06747d.
//
// Solidity: event SubnetDeployed(uint256 indexed appId, uint256 indexed nodeIp, address indexed owner)
func (_SubnetDeployment *SubnetDeploymentFilterer) FilterSubnetDeployed(opts *bind.FilterOpts, appId []*big.Int, nodeIp []*big.Int, owner []common.Address) (*SubnetDeploymentSubnetDeployedIterator, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}
	var nodeIpRule []interface{}
	for _, nodeIpItem := range nodeIp {
		nodeIpRule = append(nodeIpRule, nodeIpItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _SubnetDeployment.contract.FilterLogs(opts, "SubnetDeployed", appIdRule, nodeIpRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &SubnetDeploymentSubnetDeployedIterator{contract: _SubnetDeployment.contract, event: "SubnetDeployed", logs: logs, sub: sub}, nil
}

// WatchSubnetDeployed is a free log subscription operation binding the contract event 0x2205f96223e007c4bf7b7e81320ac9ff0bbd8393ae72ac518c59b4814c06747d.
//
// Solidity: event SubnetDeployed(uint256 indexed appId, uint256 indexed nodeIp, address indexed owner)
func (_SubnetDeployment *SubnetDeploymentFilterer) WatchSubnetDeployed(opts *bind.WatchOpts, sink chan<- *SubnetDeploymentSubnetDeployed, appId []*big.Int, nodeIp []*big.Int, owner []common.Address) (event.Subscription, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}
	var nodeIpRule []interface{}
	for _, nodeIpItem := range nodeIp {
		nodeIpRule = append(nodeIpRule, nodeIpItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _SubnetDeployment.contract.WatchLogs(opts, "SubnetDeployed", appIdRule, nodeIpRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetDeploymentSubnetDeployed)
				if err := _SubnetDeployment.contract.UnpackLog(event, "SubnetDeployed", log); err != nil {
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

// ParseSubnetDeployed is a log parse operation binding the contract event 0x2205f96223e007c4bf7b7e81320ac9ff0bbd8393ae72ac518c59b4814c06747d.
//
// Solidity: event SubnetDeployed(uint256 indexed appId, uint256 indexed nodeIp, address indexed owner)
func (_SubnetDeployment *SubnetDeploymentFilterer) ParseSubnetDeployed(log types.Log) (*SubnetDeploymentSubnetDeployed, error) {
	event := new(SubnetDeploymentSubnetDeployed)
	if err := _SubnetDeployment.contract.UnpackLog(event, "SubnetDeployed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetDeploymentSubnetDeploymentDeletedIterator is returned from FilterSubnetDeploymentDeleted and is used to iterate over the raw logs and unpacked data for SubnetDeploymentDeleted events raised by the SubnetDeployment contract.
type SubnetDeploymentSubnetDeploymentDeletedIterator struct {
	Event *SubnetDeploymentSubnetDeploymentDeleted // Event containing the contract specifics and raw log

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
func (it *SubnetDeploymentSubnetDeploymentDeletedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetDeploymentSubnetDeploymentDeleted)
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
		it.Event = new(SubnetDeploymentSubnetDeploymentDeleted)
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
func (it *SubnetDeploymentSubnetDeploymentDeletedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetDeploymentSubnetDeploymentDeletedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetDeploymentSubnetDeploymentDeleted represents a SubnetDeploymentDeleted event raised by the SubnetDeployment contract.
type SubnetDeploymentSubnetDeploymentDeleted struct {
	AppId  *big.Int
	NodeIp *big.Int
	Owner  common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterSubnetDeploymentDeleted is a free log retrieval operation binding the contract event 0x598adead5d1f1bd56e51ffd0b4c50c881b4fe9c77b8032e33a869e1ab75b4c4a.
//
// Solidity: event SubnetDeploymentDeleted(uint256 indexed appId, uint256 indexed nodeIp, address indexed owner)
func (_SubnetDeployment *SubnetDeploymentFilterer) FilterSubnetDeploymentDeleted(opts *bind.FilterOpts, appId []*big.Int, nodeIp []*big.Int, owner []common.Address) (*SubnetDeploymentSubnetDeploymentDeletedIterator, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}
	var nodeIpRule []interface{}
	for _, nodeIpItem := range nodeIp {
		nodeIpRule = append(nodeIpRule, nodeIpItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _SubnetDeployment.contract.FilterLogs(opts, "SubnetDeploymentDeleted", appIdRule, nodeIpRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &SubnetDeploymentSubnetDeploymentDeletedIterator{contract: _SubnetDeployment.contract, event: "SubnetDeploymentDeleted", logs: logs, sub: sub}, nil
}

// WatchSubnetDeploymentDeleted is a free log subscription operation binding the contract event 0x598adead5d1f1bd56e51ffd0b4c50c881b4fe9c77b8032e33a869e1ab75b4c4a.
//
// Solidity: event SubnetDeploymentDeleted(uint256 indexed appId, uint256 indexed nodeIp, address indexed owner)
func (_SubnetDeployment *SubnetDeploymentFilterer) WatchSubnetDeploymentDeleted(opts *bind.WatchOpts, sink chan<- *SubnetDeploymentSubnetDeploymentDeleted, appId []*big.Int, nodeIp []*big.Int, owner []common.Address) (event.Subscription, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}
	var nodeIpRule []interface{}
	for _, nodeIpItem := range nodeIp {
		nodeIpRule = append(nodeIpRule, nodeIpItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _SubnetDeployment.contract.WatchLogs(opts, "SubnetDeploymentDeleted", appIdRule, nodeIpRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetDeploymentSubnetDeploymentDeleted)
				if err := _SubnetDeployment.contract.UnpackLog(event, "SubnetDeploymentDeleted", log); err != nil {
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

// ParseSubnetDeploymentDeleted is a log parse operation binding the contract event 0x598adead5d1f1bd56e51ffd0b4c50c881b4fe9c77b8032e33a869e1ab75b4c4a.
//
// Solidity: event SubnetDeploymentDeleted(uint256 indexed appId, uint256 indexed nodeIp, address indexed owner)
func (_SubnetDeployment *SubnetDeploymentFilterer) ParseSubnetDeploymentDeleted(log types.Log) (*SubnetDeploymentSubnetDeploymentDeleted, error) {
	event := new(SubnetDeploymentSubnetDeploymentDeleted)
	if err := _SubnetDeployment.contract.UnpackLog(event, "SubnetDeploymentDeleted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetDeploymentSubnetDeploymentUpdatedIterator is returned from FilterSubnetDeploymentUpdated and is used to iterate over the raw logs and unpacked data for SubnetDeploymentUpdated events raised by the SubnetDeployment contract.
type SubnetDeploymentSubnetDeploymentUpdatedIterator struct {
	Event *SubnetDeploymentSubnetDeploymentUpdated // Event containing the contract specifics and raw log

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
func (it *SubnetDeploymentSubnetDeploymentUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetDeploymentSubnetDeploymentUpdated)
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
		it.Event = new(SubnetDeploymentSubnetDeploymentUpdated)
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
func (it *SubnetDeploymentSubnetDeploymentUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetDeploymentSubnetDeploymentUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetDeploymentSubnetDeploymentUpdated represents a SubnetDeploymentUpdated event raised by the SubnetDeployment contract.
type SubnetDeploymentSubnetDeploymentUpdated struct {
	AppId        *big.Int
	NodeIp       *big.Int
	DockerConfig string
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterSubnetDeploymentUpdated is a free log retrieval operation binding the contract event 0x8aa1bd7f5e2e3629e9c13e53e617710fe9deef21afda521c76b4e18c5d0e30cc.
//
// Solidity: event SubnetDeploymentUpdated(uint256 indexed appId, uint256 indexed nodeIp, string dockerConfig)
func (_SubnetDeployment *SubnetDeploymentFilterer) FilterSubnetDeploymentUpdated(opts *bind.FilterOpts, appId []*big.Int, nodeIp []*big.Int) (*SubnetDeploymentSubnetDeploymentUpdatedIterator, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}
	var nodeIpRule []interface{}
	for _, nodeIpItem := range nodeIp {
		nodeIpRule = append(nodeIpRule, nodeIpItem)
	}

	logs, sub, err := _SubnetDeployment.contract.FilterLogs(opts, "SubnetDeploymentUpdated", appIdRule, nodeIpRule)
	if err != nil {
		return nil, err
	}
	return &SubnetDeploymentSubnetDeploymentUpdatedIterator{contract: _SubnetDeployment.contract, event: "SubnetDeploymentUpdated", logs: logs, sub: sub}, nil
}

// WatchSubnetDeploymentUpdated is a free log subscription operation binding the contract event 0x8aa1bd7f5e2e3629e9c13e53e617710fe9deef21afda521c76b4e18c5d0e30cc.
//
// Solidity: event SubnetDeploymentUpdated(uint256 indexed appId, uint256 indexed nodeIp, string dockerConfig)
func (_SubnetDeployment *SubnetDeploymentFilterer) WatchSubnetDeploymentUpdated(opts *bind.WatchOpts, sink chan<- *SubnetDeploymentSubnetDeploymentUpdated, appId []*big.Int, nodeIp []*big.Int) (event.Subscription, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}
	var nodeIpRule []interface{}
	for _, nodeIpItem := range nodeIp {
		nodeIpRule = append(nodeIpRule, nodeIpItem)
	}

	logs, sub, err := _SubnetDeployment.contract.WatchLogs(opts, "SubnetDeploymentUpdated", appIdRule, nodeIpRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetDeploymentSubnetDeploymentUpdated)
				if err := _SubnetDeployment.contract.UnpackLog(event, "SubnetDeploymentUpdated", log); err != nil {
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

// ParseSubnetDeploymentUpdated is a log parse operation binding the contract event 0x8aa1bd7f5e2e3629e9c13e53e617710fe9deef21afda521c76b4e18c5d0e30cc.
//
// Solidity: event SubnetDeploymentUpdated(uint256 indexed appId, uint256 indexed nodeIp, string dockerConfig)
func (_SubnetDeployment *SubnetDeploymentFilterer) ParseSubnetDeploymentUpdated(log types.Log) (*SubnetDeploymentSubnetDeploymentUpdated, error) {
	event := new(SubnetDeploymentSubnetDeploymentUpdated)
	if err := _SubnetDeployment.contract.UnpackLog(event, "SubnetDeploymentUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
