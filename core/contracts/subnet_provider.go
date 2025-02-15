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

// SubnetProviderPeerNode is an auto generated low-level Go binding around an user-defined struct.
type SubnetProviderPeerNode struct {
	IsRegistered bool
	Metadata     string
}

// SubnetProviderProvider is an auto generated low-level Go binding around an user-defined struct.
type SubnetProviderProvider struct {
	TokenId      *big.Int
	ProviderName string
	Operator     common.Address
	Website      string
	Metadata     string
}

// SubnetProviderMetaData contains all meta data concerning the SubnetProvider contract.
var SubnetProviderMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"ERC721IncorrectOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ERC721InsufficientApproval\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"approver\",\"type\":\"address\"}],\"name\":\"ERC721InvalidApprover\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"ERC721InvalidOperator\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"ERC721InvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"}],\"name\":\"ERC721InvalidReceiver\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"ERC721InvalidSender\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ERC721NonexistentToken\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"approved\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"providerAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"NFTMinted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"OperatorUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"}],\"name\":\"PeerNodeDeleted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"}],\"name\":\"PeerNodeRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"}],\"name\":\"PeerNodeUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ProviderDeleted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"providerAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"providerName\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"website\",\"type\":\"string\"}],\"name\":\"ProviderRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"providerName\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"website\",\"type\":\"string\"}],\"name\":\"ProviderUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"website\",\"type\":\"string\"}],\"name\":\"WebsiteUpdated\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"}],\"name\":\"deletePeerNode\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"deleteProvider\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getApproved\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"}],\"name\":\"getPeerNode\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"isRegistered\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"}],\"internalType\":\"structSubnetProvider.PeerNode\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"getProvider\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"providerName\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"website\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"}],\"internalType\":\"structSubnetProvider.Provider\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"peerNodeRegistered\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"isRegistered\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"providers\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"providerName\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"website\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"}],\"name\":\"registerPeerNode\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_providerName\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_metadata\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"_operator\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_website\",\"type\":\"string\"}],\"name\":\"registerProvider\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"tokenURI\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_operator\",\"type\":\"address\"}],\"name\":\"updateOperator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"}],\"name\":\"updatePeerNode\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"_providerName\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_metadata\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"_operator\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_website\",\"type\":\"string\"}],\"name\":\"updateProvider\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"_website\",\"type\":\"string\"}],\"name\":\"updateWebsite\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]",
}

// SubnetProviderABI is the input ABI used to generate the binding from.
// Deprecated: Use SubnetProviderMetaData.ABI instead.
var SubnetProviderABI = SubnetProviderMetaData.ABI

// SubnetProvider is an auto generated Go binding around an Ethereum contract.
type SubnetProvider struct {
	SubnetProviderCaller     // Read-only binding to the contract
	SubnetProviderTransactor // Write-only binding to the contract
	SubnetProviderFilterer   // Log filterer for contract events
}

// SubnetProviderCaller is an auto generated read-only Go binding around an Ethereum contract.
type SubnetProviderCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetProviderTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SubnetProviderTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetProviderFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SubnetProviderFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetProviderSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SubnetProviderSession struct {
	Contract     *SubnetProvider   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SubnetProviderCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SubnetProviderCallerSession struct {
	Contract *SubnetProviderCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// SubnetProviderTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SubnetProviderTransactorSession struct {
	Contract     *SubnetProviderTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// SubnetProviderRaw is an auto generated low-level Go binding around an Ethereum contract.
type SubnetProviderRaw struct {
	Contract *SubnetProvider // Generic contract binding to access the raw methods on
}

// SubnetProviderCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SubnetProviderCallerRaw struct {
	Contract *SubnetProviderCaller // Generic read-only contract binding to access the raw methods on
}

// SubnetProviderTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SubnetProviderTransactorRaw struct {
	Contract *SubnetProviderTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSubnetProvider creates a new instance of SubnetProvider, bound to a specific deployed contract.
func NewSubnetProvider(address common.Address, backend bind.ContractBackend) (*SubnetProvider, error) {
	contract, err := bindSubnetProvider(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SubnetProvider{SubnetProviderCaller: SubnetProviderCaller{contract: contract}, SubnetProviderTransactor: SubnetProviderTransactor{contract: contract}, SubnetProviderFilterer: SubnetProviderFilterer{contract: contract}}, nil
}

// NewSubnetProviderCaller creates a new read-only instance of SubnetProvider, bound to a specific deployed contract.
func NewSubnetProviderCaller(address common.Address, caller bind.ContractCaller) (*SubnetProviderCaller, error) {
	contract, err := bindSubnetProvider(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SubnetProviderCaller{contract: contract}, nil
}

// NewSubnetProviderTransactor creates a new write-only instance of SubnetProvider, bound to a specific deployed contract.
func NewSubnetProviderTransactor(address common.Address, transactor bind.ContractTransactor) (*SubnetProviderTransactor, error) {
	contract, err := bindSubnetProvider(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SubnetProviderTransactor{contract: contract}, nil
}

// NewSubnetProviderFilterer creates a new log filterer instance of SubnetProvider, bound to a specific deployed contract.
func NewSubnetProviderFilterer(address common.Address, filterer bind.ContractFilterer) (*SubnetProviderFilterer, error) {
	contract, err := bindSubnetProvider(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SubnetProviderFilterer{contract: contract}, nil
}

// bindSubnetProvider binds a generic wrapper to an already deployed contract.
func bindSubnetProvider(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SubnetProviderMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SubnetProvider *SubnetProviderRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SubnetProvider.Contract.SubnetProviderCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SubnetProvider *SubnetProviderRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetProvider.Contract.SubnetProviderTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SubnetProvider *SubnetProviderRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SubnetProvider.Contract.SubnetProviderTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SubnetProvider *SubnetProviderCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SubnetProvider.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SubnetProvider *SubnetProviderTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetProvider.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SubnetProvider *SubnetProviderTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SubnetProvider.Contract.contract.Transact(opts, method, params...)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_SubnetProvider *SubnetProviderCaller) BalanceOf(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _SubnetProvider.contract.Call(opts, &out, "balanceOf", owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_SubnetProvider *SubnetProviderSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _SubnetProvider.Contract.BalanceOf(&_SubnetProvider.CallOpts, owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_SubnetProvider *SubnetProviderCallerSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _SubnetProvider.Contract.BalanceOf(&_SubnetProvider.CallOpts, owner)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_SubnetProvider *SubnetProviderCaller) GetApproved(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _SubnetProvider.contract.Call(opts, &out, "getApproved", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_SubnetProvider *SubnetProviderSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _SubnetProvider.Contract.GetApproved(&_SubnetProvider.CallOpts, tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_SubnetProvider *SubnetProviderCallerSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _SubnetProvider.Contract.GetApproved(&_SubnetProvider.CallOpts, tokenId)
}

// GetPeerNode is a free data retrieval call binding the contract method 0x65ffb2ec.
//
// Solidity: function getPeerNode(uint256 tokenId, string peerId) view returns((bool,string))
func (_SubnetProvider *SubnetProviderCaller) GetPeerNode(opts *bind.CallOpts, tokenId *big.Int, peerId string) (SubnetProviderPeerNode, error) {
	var out []interface{}
	err := _SubnetProvider.contract.Call(opts, &out, "getPeerNode", tokenId, peerId)

	if err != nil {
		return *new(SubnetProviderPeerNode), err
	}

	out0 := *abi.ConvertType(out[0], new(SubnetProviderPeerNode)).(*SubnetProviderPeerNode)

	return out0, err

}

// GetPeerNode is a free data retrieval call binding the contract method 0x65ffb2ec.
//
// Solidity: function getPeerNode(uint256 tokenId, string peerId) view returns((bool,string))
func (_SubnetProvider *SubnetProviderSession) GetPeerNode(tokenId *big.Int, peerId string) (SubnetProviderPeerNode, error) {
	return _SubnetProvider.Contract.GetPeerNode(&_SubnetProvider.CallOpts, tokenId, peerId)
}

// GetPeerNode is a free data retrieval call binding the contract method 0x65ffb2ec.
//
// Solidity: function getPeerNode(uint256 tokenId, string peerId) view returns((bool,string))
func (_SubnetProvider *SubnetProviderCallerSession) GetPeerNode(tokenId *big.Int, peerId string) (SubnetProviderPeerNode, error) {
	return _SubnetProvider.Contract.GetPeerNode(&_SubnetProvider.CallOpts, tokenId, peerId)
}

// GetProvider is a free data retrieval call binding the contract method 0x5c42d079.
//
// Solidity: function getProvider(uint256 _tokenId) view returns((uint256,string,address,string,string))
func (_SubnetProvider *SubnetProviderCaller) GetProvider(opts *bind.CallOpts, _tokenId *big.Int) (SubnetProviderProvider, error) {
	var out []interface{}
	err := _SubnetProvider.contract.Call(opts, &out, "getProvider", _tokenId)

	if err != nil {
		return *new(SubnetProviderProvider), err
	}

	out0 := *abi.ConvertType(out[0], new(SubnetProviderProvider)).(*SubnetProviderProvider)

	return out0, err

}

// GetProvider is a free data retrieval call binding the contract method 0x5c42d079.
//
// Solidity: function getProvider(uint256 _tokenId) view returns((uint256,string,address,string,string))
func (_SubnetProvider *SubnetProviderSession) GetProvider(_tokenId *big.Int) (SubnetProviderProvider, error) {
	return _SubnetProvider.Contract.GetProvider(&_SubnetProvider.CallOpts, _tokenId)
}

// GetProvider is a free data retrieval call binding the contract method 0x5c42d079.
//
// Solidity: function getProvider(uint256 _tokenId) view returns((uint256,string,address,string,string))
func (_SubnetProvider *SubnetProviderCallerSession) GetProvider(_tokenId *big.Int) (SubnetProviderProvider, error) {
	return _SubnetProvider.Contract.GetProvider(&_SubnetProvider.CallOpts, _tokenId)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_SubnetProvider *SubnetProviderCaller) IsApprovedForAll(opts *bind.CallOpts, owner common.Address, operator common.Address) (bool, error) {
	var out []interface{}
	err := _SubnetProvider.contract.Call(opts, &out, "isApprovedForAll", owner, operator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_SubnetProvider *SubnetProviderSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _SubnetProvider.Contract.IsApprovedForAll(&_SubnetProvider.CallOpts, owner, operator)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_SubnetProvider *SubnetProviderCallerSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _SubnetProvider.Contract.IsApprovedForAll(&_SubnetProvider.CallOpts, owner, operator)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_SubnetProvider *SubnetProviderCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _SubnetProvider.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_SubnetProvider *SubnetProviderSession) Name() (string, error) {
	return _SubnetProvider.Contract.Name(&_SubnetProvider.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_SubnetProvider *SubnetProviderCallerSession) Name() (string, error) {
	return _SubnetProvider.Contract.Name(&_SubnetProvider.CallOpts)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_SubnetProvider *SubnetProviderCaller) OwnerOf(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _SubnetProvider.contract.Call(opts, &out, "ownerOf", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_SubnetProvider *SubnetProviderSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _SubnetProvider.Contract.OwnerOf(&_SubnetProvider.CallOpts, tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_SubnetProvider *SubnetProviderCallerSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _SubnetProvider.Contract.OwnerOf(&_SubnetProvider.CallOpts, tokenId)
}

// PeerNodeRegistered is a free data retrieval call binding the contract method 0xb88923bc.
//
// Solidity: function peerNodeRegistered(uint256 , string ) view returns(bool isRegistered, string metadata)
func (_SubnetProvider *SubnetProviderCaller) PeerNodeRegistered(opts *bind.CallOpts, arg0 *big.Int, arg1 string) (struct {
	IsRegistered bool
	Metadata     string
}, error) {
	var out []interface{}
	err := _SubnetProvider.contract.Call(opts, &out, "peerNodeRegistered", arg0, arg1)

	outstruct := new(struct {
		IsRegistered bool
		Metadata     string
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.IsRegistered = *abi.ConvertType(out[0], new(bool)).(*bool)
	outstruct.Metadata = *abi.ConvertType(out[1], new(string)).(*string)

	return *outstruct, err

}

// PeerNodeRegistered is a free data retrieval call binding the contract method 0xb88923bc.
//
// Solidity: function peerNodeRegistered(uint256 , string ) view returns(bool isRegistered, string metadata)
func (_SubnetProvider *SubnetProviderSession) PeerNodeRegistered(arg0 *big.Int, arg1 string) (struct {
	IsRegistered bool
	Metadata     string
}, error) {
	return _SubnetProvider.Contract.PeerNodeRegistered(&_SubnetProvider.CallOpts, arg0, arg1)
}

// PeerNodeRegistered is a free data retrieval call binding the contract method 0xb88923bc.
//
// Solidity: function peerNodeRegistered(uint256 , string ) view returns(bool isRegistered, string metadata)
func (_SubnetProvider *SubnetProviderCallerSession) PeerNodeRegistered(arg0 *big.Int, arg1 string) (struct {
	IsRegistered bool
	Metadata     string
}, error) {
	return _SubnetProvider.Contract.PeerNodeRegistered(&_SubnetProvider.CallOpts, arg0, arg1)
}

// Providers is a free data retrieval call binding the contract method 0x50f3fc81.
//
// Solidity: function providers(uint256 ) view returns(uint256 tokenId, string providerName, address operator, string website, string metadata)
func (_SubnetProvider *SubnetProviderCaller) Providers(opts *bind.CallOpts, arg0 *big.Int) (struct {
	TokenId      *big.Int
	ProviderName string
	Operator     common.Address
	Website      string
	Metadata     string
}, error) {
	var out []interface{}
	err := _SubnetProvider.contract.Call(opts, &out, "providers", arg0)

	outstruct := new(struct {
		TokenId      *big.Int
		ProviderName string
		Operator     common.Address
		Website      string
		Metadata     string
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.TokenId = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.ProviderName = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.Operator = *abi.ConvertType(out[2], new(common.Address)).(*common.Address)
	outstruct.Website = *abi.ConvertType(out[3], new(string)).(*string)
	outstruct.Metadata = *abi.ConvertType(out[4], new(string)).(*string)

	return *outstruct, err

}

// Providers is a free data retrieval call binding the contract method 0x50f3fc81.
//
// Solidity: function providers(uint256 ) view returns(uint256 tokenId, string providerName, address operator, string website, string metadata)
func (_SubnetProvider *SubnetProviderSession) Providers(arg0 *big.Int) (struct {
	TokenId      *big.Int
	ProviderName string
	Operator     common.Address
	Website      string
	Metadata     string
}, error) {
	return _SubnetProvider.Contract.Providers(&_SubnetProvider.CallOpts, arg0)
}

// Providers is a free data retrieval call binding the contract method 0x50f3fc81.
//
// Solidity: function providers(uint256 ) view returns(uint256 tokenId, string providerName, address operator, string website, string metadata)
func (_SubnetProvider *SubnetProviderCallerSession) Providers(arg0 *big.Int) (struct {
	TokenId      *big.Int
	ProviderName string
	Operator     common.Address
	Website      string
	Metadata     string
}, error) {
	return _SubnetProvider.Contract.Providers(&_SubnetProvider.CallOpts, arg0)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_SubnetProvider *SubnetProviderCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _SubnetProvider.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_SubnetProvider *SubnetProviderSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _SubnetProvider.Contract.SupportsInterface(&_SubnetProvider.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_SubnetProvider *SubnetProviderCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _SubnetProvider.Contract.SupportsInterface(&_SubnetProvider.CallOpts, interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_SubnetProvider *SubnetProviderCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _SubnetProvider.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_SubnetProvider *SubnetProviderSession) Symbol() (string, error) {
	return _SubnetProvider.Contract.Symbol(&_SubnetProvider.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_SubnetProvider *SubnetProviderCallerSession) Symbol() (string, error) {
	return _SubnetProvider.Contract.Symbol(&_SubnetProvider.CallOpts)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_SubnetProvider *SubnetProviderCaller) TokenURI(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var out []interface{}
	err := _SubnetProvider.contract.Call(opts, &out, "tokenURI", tokenId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_SubnetProvider *SubnetProviderSession) TokenURI(tokenId *big.Int) (string, error) {
	return _SubnetProvider.Contract.TokenURI(&_SubnetProvider.CallOpts, tokenId)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_SubnetProvider *SubnetProviderCallerSession) TokenURI(tokenId *big.Int) (string, error) {
	return _SubnetProvider.Contract.TokenURI(&_SubnetProvider.CallOpts, tokenId)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(string)
func (_SubnetProvider *SubnetProviderCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _SubnetProvider.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(string)
func (_SubnetProvider *SubnetProviderSession) Version() (string, error) {
	return _SubnetProvider.Contract.Version(&_SubnetProvider.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(string)
func (_SubnetProvider *SubnetProviderCallerSession) Version() (string, error) {
	return _SubnetProvider.Contract.Version(&_SubnetProvider.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_SubnetProvider *SubnetProviderTransactor) Approve(opts *bind.TransactOpts, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetProvider.contract.Transact(opts, "approve", to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_SubnetProvider *SubnetProviderSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetProvider.Contract.Approve(&_SubnetProvider.TransactOpts, to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_SubnetProvider *SubnetProviderTransactorSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetProvider.Contract.Approve(&_SubnetProvider.TransactOpts, to, tokenId)
}

// DeletePeerNode is a paid mutator transaction binding the contract method 0x680bf85c.
//
// Solidity: function deletePeerNode(uint256 tokenId, string peerId) returns()
func (_SubnetProvider *SubnetProviderTransactor) DeletePeerNode(opts *bind.TransactOpts, tokenId *big.Int, peerId string) (*types.Transaction, error) {
	return _SubnetProvider.contract.Transact(opts, "deletePeerNode", tokenId, peerId)
}

// DeletePeerNode is a paid mutator transaction binding the contract method 0x680bf85c.
//
// Solidity: function deletePeerNode(uint256 tokenId, string peerId) returns()
func (_SubnetProvider *SubnetProviderSession) DeletePeerNode(tokenId *big.Int, peerId string) (*types.Transaction, error) {
	return _SubnetProvider.Contract.DeletePeerNode(&_SubnetProvider.TransactOpts, tokenId, peerId)
}

// DeletePeerNode is a paid mutator transaction binding the contract method 0x680bf85c.
//
// Solidity: function deletePeerNode(uint256 tokenId, string peerId) returns()
func (_SubnetProvider *SubnetProviderTransactorSession) DeletePeerNode(tokenId *big.Int, peerId string) (*types.Transaction, error) {
	return _SubnetProvider.Contract.DeletePeerNode(&_SubnetProvider.TransactOpts, tokenId, peerId)
}

// DeleteProvider is a paid mutator transaction binding the contract method 0xac953e8d.
//
// Solidity: function deleteProvider(uint256 tokenId) returns()
func (_SubnetProvider *SubnetProviderTransactor) DeleteProvider(opts *bind.TransactOpts, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetProvider.contract.Transact(opts, "deleteProvider", tokenId)
}

// DeleteProvider is a paid mutator transaction binding the contract method 0xac953e8d.
//
// Solidity: function deleteProvider(uint256 tokenId) returns()
func (_SubnetProvider *SubnetProviderSession) DeleteProvider(tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetProvider.Contract.DeleteProvider(&_SubnetProvider.TransactOpts, tokenId)
}

// DeleteProvider is a paid mutator transaction binding the contract method 0xac953e8d.
//
// Solidity: function deleteProvider(uint256 tokenId) returns()
func (_SubnetProvider *SubnetProviderTransactorSession) DeleteProvider(tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetProvider.Contract.DeleteProvider(&_SubnetProvider.TransactOpts, tokenId)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_SubnetProvider *SubnetProviderTransactor) Initialize(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetProvider.contract.Transact(opts, "initialize")
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_SubnetProvider *SubnetProviderSession) Initialize() (*types.Transaction, error) {
	return _SubnetProvider.Contract.Initialize(&_SubnetProvider.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_SubnetProvider *SubnetProviderTransactorSession) Initialize() (*types.Transaction, error) {
	return _SubnetProvider.Contract.Initialize(&_SubnetProvider.TransactOpts)
}

// RegisterPeerNode is a paid mutator transaction binding the contract method 0x54884320.
//
// Solidity: function registerPeerNode(uint256 tokenId, string peerId, string metadata) returns()
func (_SubnetProvider *SubnetProviderTransactor) RegisterPeerNode(opts *bind.TransactOpts, tokenId *big.Int, peerId string, metadata string) (*types.Transaction, error) {
	return _SubnetProvider.contract.Transact(opts, "registerPeerNode", tokenId, peerId, metadata)
}

// RegisterPeerNode is a paid mutator transaction binding the contract method 0x54884320.
//
// Solidity: function registerPeerNode(uint256 tokenId, string peerId, string metadata) returns()
func (_SubnetProvider *SubnetProviderSession) RegisterPeerNode(tokenId *big.Int, peerId string, metadata string) (*types.Transaction, error) {
	return _SubnetProvider.Contract.RegisterPeerNode(&_SubnetProvider.TransactOpts, tokenId, peerId, metadata)
}

// RegisterPeerNode is a paid mutator transaction binding the contract method 0x54884320.
//
// Solidity: function registerPeerNode(uint256 tokenId, string peerId, string metadata) returns()
func (_SubnetProvider *SubnetProviderTransactorSession) RegisterPeerNode(tokenId *big.Int, peerId string, metadata string) (*types.Transaction, error) {
	return _SubnetProvider.Contract.RegisterPeerNode(&_SubnetProvider.TransactOpts, tokenId, peerId, metadata)
}

// RegisterProvider is a paid mutator transaction binding the contract method 0xea677b77.
//
// Solidity: function registerProvider(string _providerName, string _metadata, address _operator, string _website) returns(uint256)
func (_SubnetProvider *SubnetProviderTransactor) RegisterProvider(opts *bind.TransactOpts, _providerName string, _metadata string, _operator common.Address, _website string) (*types.Transaction, error) {
	return _SubnetProvider.contract.Transact(opts, "registerProvider", _providerName, _metadata, _operator, _website)
}

// RegisterProvider is a paid mutator transaction binding the contract method 0xea677b77.
//
// Solidity: function registerProvider(string _providerName, string _metadata, address _operator, string _website) returns(uint256)
func (_SubnetProvider *SubnetProviderSession) RegisterProvider(_providerName string, _metadata string, _operator common.Address, _website string) (*types.Transaction, error) {
	return _SubnetProvider.Contract.RegisterProvider(&_SubnetProvider.TransactOpts, _providerName, _metadata, _operator, _website)
}

// RegisterProvider is a paid mutator transaction binding the contract method 0xea677b77.
//
// Solidity: function registerProvider(string _providerName, string _metadata, address _operator, string _website) returns(uint256)
func (_SubnetProvider *SubnetProviderTransactorSession) RegisterProvider(_providerName string, _metadata string, _operator common.Address, _website string) (*types.Transaction, error) {
	return _SubnetProvider.Contract.RegisterProvider(&_SubnetProvider.TransactOpts, _providerName, _metadata, _operator, _website)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_SubnetProvider *SubnetProviderTransactor) SafeTransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetProvider.contract.Transact(opts, "safeTransferFrom", from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_SubnetProvider *SubnetProviderSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetProvider.Contract.SafeTransferFrom(&_SubnetProvider.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_SubnetProvider *SubnetProviderTransactorSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetProvider.Contract.SafeTransferFrom(&_SubnetProvider.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_SubnetProvider *SubnetProviderTransactor) SafeTransferFrom0(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _SubnetProvider.contract.Transact(opts, "safeTransferFrom0", from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_SubnetProvider *SubnetProviderSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _SubnetProvider.Contract.SafeTransferFrom0(&_SubnetProvider.TransactOpts, from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_SubnetProvider *SubnetProviderTransactorSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _SubnetProvider.Contract.SafeTransferFrom0(&_SubnetProvider.TransactOpts, from, to, tokenId, data)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_SubnetProvider *SubnetProviderTransactor) SetApprovalForAll(opts *bind.TransactOpts, operator common.Address, approved bool) (*types.Transaction, error) {
	return _SubnetProvider.contract.Transact(opts, "setApprovalForAll", operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_SubnetProvider *SubnetProviderSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _SubnetProvider.Contract.SetApprovalForAll(&_SubnetProvider.TransactOpts, operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_SubnetProvider *SubnetProviderTransactorSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _SubnetProvider.Contract.SetApprovalForAll(&_SubnetProvider.TransactOpts, operator, approved)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_SubnetProvider *SubnetProviderTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetProvider.contract.Transact(opts, "transferFrom", from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_SubnetProvider *SubnetProviderSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetProvider.Contract.TransferFrom(&_SubnetProvider.TransactOpts, from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_SubnetProvider *SubnetProviderTransactorSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _SubnetProvider.Contract.TransferFrom(&_SubnetProvider.TransactOpts, from, to, tokenId)
}

// UpdateOperator is a paid mutator transaction binding the contract method 0xc65a391d.
//
// Solidity: function updateOperator(uint256 tokenId, address _operator) returns()
func (_SubnetProvider *SubnetProviderTransactor) UpdateOperator(opts *bind.TransactOpts, tokenId *big.Int, _operator common.Address) (*types.Transaction, error) {
	return _SubnetProvider.contract.Transact(opts, "updateOperator", tokenId, _operator)
}

// UpdateOperator is a paid mutator transaction binding the contract method 0xc65a391d.
//
// Solidity: function updateOperator(uint256 tokenId, address _operator) returns()
func (_SubnetProvider *SubnetProviderSession) UpdateOperator(tokenId *big.Int, _operator common.Address) (*types.Transaction, error) {
	return _SubnetProvider.Contract.UpdateOperator(&_SubnetProvider.TransactOpts, tokenId, _operator)
}

// UpdateOperator is a paid mutator transaction binding the contract method 0xc65a391d.
//
// Solidity: function updateOperator(uint256 tokenId, address _operator) returns()
func (_SubnetProvider *SubnetProviderTransactorSession) UpdateOperator(tokenId *big.Int, _operator common.Address) (*types.Transaction, error) {
	return _SubnetProvider.Contract.UpdateOperator(&_SubnetProvider.TransactOpts, tokenId, _operator)
}

// UpdatePeerNode is a paid mutator transaction binding the contract method 0x53fad0fd.
//
// Solidity: function updatePeerNode(uint256 tokenId, string peerId, string metadata) returns()
func (_SubnetProvider *SubnetProviderTransactor) UpdatePeerNode(opts *bind.TransactOpts, tokenId *big.Int, peerId string, metadata string) (*types.Transaction, error) {
	return _SubnetProvider.contract.Transact(opts, "updatePeerNode", tokenId, peerId, metadata)
}

// UpdatePeerNode is a paid mutator transaction binding the contract method 0x53fad0fd.
//
// Solidity: function updatePeerNode(uint256 tokenId, string peerId, string metadata) returns()
func (_SubnetProvider *SubnetProviderSession) UpdatePeerNode(tokenId *big.Int, peerId string, metadata string) (*types.Transaction, error) {
	return _SubnetProvider.Contract.UpdatePeerNode(&_SubnetProvider.TransactOpts, tokenId, peerId, metadata)
}

// UpdatePeerNode is a paid mutator transaction binding the contract method 0x53fad0fd.
//
// Solidity: function updatePeerNode(uint256 tokenId, string peerId, string metadata) returns()
func (_SubnetProvider *SubnetProviderTransactorSession) UpdatePeerNode(tokenId *big.Int, peerId string, metadata string) (*types.Transaction, error) {
	return _SubnetProvider.Contract.UpdatePeerNode(&_SubnetProvider.TransactOpts, tokenId, peerId, metadata)
}

// UpdateProvider is a paid mutator transaction binding the contract method 0xef17ed7e.
//
// Solidity: function updateProvider(uint256 tokenId, string _providerName, string _metadata, address _operator, string _website) returns()
func (_SubnetProvider *SubnetProviderTransactor) UpdateProvider(opts *bind.TransactOpts, tokenId *big.Int, _providerName string, _metadata string, _operator common.Address, _website string) (*types.Transaction, error) {
	return _SubnetProvider.contract.Transact(opts, "updateProvider", tokenId, _providerName, _metadata, _operator, _website)
}

// UpdateProvider is a paid mutator transaction binding the contract method 0xef17ed7e.
//
// Solidity: function updateProvider(uint256 tokenId, string _providerName, string _metadata, address _operator, string _website) returns()
func (_SubnetProvider *SubnetProviderSession) UpdateProvider(tokenId *big.Int, _providerName string, _metadata string, _operator common.Address, _website string) (*types.Transaction, error) {
	return _SubnetProvider.Contract.UpdateProvider(&_SubnetProvider.TransactOpts, tokenId, _providerName, _metadata, _operator, _website)
}

// UpdateProvider is a paid mutator transaction binding the contract method 0xef17ed7e.
//
// Solidity: function updateProvider(uint256 tokenId, string _providerName, string _metadata, address _operator, string _website) returns()
func (_SubnetProvider *SubnetProviderTransactorSession) UpdateProvider(tokenId *big.Int, _providerName string, _metadata string, _operator common.Address, _website string) (*types.Transaction, error) {
	return _SubnetProvider.Contract.UpdateProvider(&_SubnetProvider.TransactOpts, tokenId, _providerName, _metadata, _operator, _website)
}

// UpdateWebsite is a paid mutator transaction binding the contract method 0x9488a32f.
//
// Solidity: function updateWebsite(uint256 tokenId, string _website) returns()
func (_SubnetProvider *SubnetProviderTransactor) UpdateWebsite(opts *bind.TransactOpts, tokenId *big.Int, _website string) (*types.Transaction, error) {
	return _SubnetProvider.contract.Transact(opts, "updateWebsite", tokenId, _website)
}

// UpdateWebsite is a paid mutator transaction binding the contract method 0x9488a32f.
//
// Solidity: function updateWebsite(uint256 tokenId, string _website) returns()
func (_SubnetProvider *SubnetProviderSession) UpdateWebsite(tokenId *big.Int, _website string) (*types.Transaction, error) {
	return _SubnetProvider.Contract.UpdateWebsite(&_SubnetProvider.TransactOpts, tokenId, _website)
}

// UpdateWebsite is a paid mutator transaction binding the contract method 0x9488a32f.
//
// Solidity: function updateWebsite(uint256 tokenId, string _website) returns()
func (_SubnetProvider *SubnetProviderTransactorSession) UpdateWebsite(tokenId *big.Int, _website string) (*types.Transaction, error) {
	return _SubnetProvider.Contract.UpdateWebsite(&_SubnetProvider.TransactOpts, tokenId, _website)
}

// SubnetProviderApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the SubnetProvider contract.
type SubnetProviderApprovalIterator struct {
	Event *SubnetProviderApproval // Event containing the contract specifics and raw log

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
func (it *SubnetProviderApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderApproval)
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
		it.Event = new(SubnetProviderApproval)
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
func (it *SubnetProviderApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderApproval represents a Approval event raised by the SubnetProvider contract.
type SubnetProviderApproval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_SubnetProvider *SubnetProviderFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, approved []common.Address, tokenId []*big.Int) (*SubnetProviderApprovalIterator, error) {

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

	logs, sub, err := _SubnetProvider.contract.FilterLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &SubnetProviderApprovalIterator{contract: _SubnetProvider.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_SubnetProvider *SubnetProviderFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *SubnetProviderApproval, owner []common.Address, approved []common.Address, tokenId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _SubnetProvider.contract.WatchLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderApproval)
				if err := _SubnetProvider.contract.UnpackLog(event, "Approval", log); err != nil {
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
func (_SubnetProvider *SubnetProviderFilterer) ParseApproval(log types.Log) (*SubnetProviderApproval, error) {
	event := new(SubnetProviderApproval)
	if err := _SubnetProvider.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetProviderApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the SubnetProvider contract.
type SubnetProviderApprovalForAllIterator struct {
	Event *SubnetProviderApprovalForAll // Event containing the contract specifics and raw log

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
func (it *SubnetProviderApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderApprovalForAll)
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
		it.Event = new(SubnetProviderApprovalForAll)
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
func (it *SubnetProviderApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderApprovalForAll represents a ApprovalForAll event raised by the SubnetProvider contract.
type SubnetProviderApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_SubnetProvider *SubnetProviderFilterer) FilterApprovalForAll(opts *bind.FilterOpts, owner []common.Address, operator []common.Address) (*SubnetProviderApprovalForAllIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _SubnetProvider.contract.FilterLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return &SubnetProviderApprovalForAllIterator{contract: _SubnetProvider.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_SubnetProvider *SubnetProviderFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *SubnetProviderApprovalForAll, owner []common.Address, operator []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _SubnetProvider.contract.WatchLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderApprovalForAll)
				if err := _SubnetProvider.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
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
func (_SubnetProvider *SubnetProviderFilterer) ParseApprovalForAll(log types.Log) (*SubnetProviderApprovalForAll, error) {
	event := new(SubnetProviderApprovalForAll)
	if err := _SubnetProvider.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetProviderInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the SubnetProvider contract.
type SubnetProviderInitializedIterator struct {
	Event *SubnetProviderInitialized // Event containing the contract specifics and raw log

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
func (it *SubnetProviderInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderInitialized)
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
		it.Event = new(SubnetProviderInitialized)
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
func (it *SubnetProviderInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderInitialized represents a Initialized event raised by the SubnetProvider contract.
type SubnetProviderInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_SubnetProvider *SubnetProviderFilterer) FilterInitialized(opts *bind.FilterOpts) (*SubnetProviderInitializedIterator, error) {

	logs, sub, err := _SubnetProvider.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &SubnetProviderInitializedIterator{contract: _SubnetProvider.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_SubnetProvider *SubnetProviderFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *SubnetProviderInitialized) (event.Subscription, error) {

	logs, sub, err := _SubnetProvider.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderInitialized)
				if err := _SubnetProvider.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_SubnetProvider *SubnetProviderFilterer) ParseInitialized(log types.Log) (*SubnetProviderInitialized, error) {
	event := new(SubnetProviderInitialized)
	if err := _SubnetProvider.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetProviderNFTMintedIterator is returned from FilterNFTMinted and is used to iterate over the raw logs and unpacked data for NFTMinted events raised by the SubnetProvider contract.
type SubnetProviderNFTMintedIterator struct {
	Event *SubnetProviderNFTMinted // Event containing the contract specifics and raw log

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
func (it *SubnetProviderNFTMintedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderNFTMinted)
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
		it.Event = new(SubnetProviderNFTMinted)
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
func (it *SubnetProviderNFTMintedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderNFTMintedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderNFTMinted represents a NFTMinted event raised by the SubnetProvider contract.
type SubnetProviderNFTMinted struct {
	ProviderAddress common.Address
	TokenId         *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterNFTMinted is a free log retrieval operation binding the contract event 0x4cc0a9c4a99ddc700de1af2c9f916a7cbfdb71f14801ccff94061ad1ef8a8040.
//
// Solidity: event NFTMinted(address providerAddress, uint256 tokenId)
func (_SubnetProvider *SubnetProviderFilterer) FilterNFTMinted(opts *bind.FilterOpts) (*SubnetProviderNFTMintedIterator, error) {

	logs, sub, err := _SubnetProvider.contract.FilterLogs(opts, "NFTMinted")
	if err != nil {
		return nil, err
	}
	return &SubnetProviderNFTMintedIterator{contract: _SubnetProvider.contract, event: "NFTMinted", logs: logs, sub: sub}, nil
}

// WatchNFTMinted is a free log subscription operation binding the contract event 0x4cc0a9c4a99ddc700de1af2c9f916a7cbfdb71f14801ccff94061ad1ef8a8040.
//
// Solidity: event NFTMinted(address providerAddress, uint256 tokenId)
func (_SubnetProvider *SubnetProviderFilterer) WatchNFTMinted(opts *bind.WatchOpts, sink chan<- *SubnetProviderNFTMinted) (event.Subscription, error) {

	logs, sub, err := _SubnetProvider.contract.WatchLogs(opts, "NFTMinted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderNFTMinted)
				if err := _SubnetProvider.contract.UnpackLog(event, "NFTMinted", log); err != nil {
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

// ParseNFTMinted is a log parse operation binding the contract event 0x4cc0a9c4a99ddc700de1af2c9f916a7cbfdb71f14801ccff94061ad1ef8a8040.
//
// Solidity: event NFTMinted(address providerAddress, uint256 tokenId)
func (_SubnetProvider *SubnetProviderFilterer) ParseNFTMinted(log types.Log) (*SubnetProviderNFTMinted, error) {
	event := new(SubnetProviderNFTMinted)
	if err := _SubnetProvider.contract.UnpackLog(event, "NFTMinted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetProviderOperatorUpdatedIterator is returned from FilterOperatorUpdated and is used to iterate over the raw logs and unpacked data for OperatorUpdated events raised by the SubnetProvider contract.
type SubnetProviderOperatorUpdatedIterator struct {
	Event *SubnetProviderOperatorUpdated // Event containing the contract specifics and raw log

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
func (it *SubnetProviderOperatorUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderOperatorUpdated)
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
		it.Event = new(SubnetProviderOperatorUpdated)
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
func (it *SubnetProviderOperatorUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderOperatorUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderOperatorUpdated represents a OperatorUpdated event raised by the SubnetProvider contract.
type SubnetProviderOperatorUpdated struct {
	TokenId  *big.Int
	Operator common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterOperatorUpdated is a free log retrieval operation binding the contract event 0xbc3f928a2240d921e006f13446712fe7588e04553a6ddc13da2143e0ed93395b.
//
// Solidity: event OperatorUpdated(uint256 tokenId, address operator)
func (_SubnetProvider *SubnetProviderFilterer) FilterOperatorUpdated(opts *bind.FilterOpts) (*SubnetProviderOperatorUpdatedIterator, error) {

	logs, sub, err := _SubnetProvider.contract.FilterLogs(opts, "OperatorUpdated")
	if err != nil {
		return nil, err
	}
	return &SubnetProviderOperatorUpdatedIterator{contract: _SubnetProvider.contract, event: "OperatorUpdated", logs: logs, sub: sub}, nil
}

// WatchOperatorUpdated is a free log subscription operation binding the contract event 0xbc3f928a2240d921e006f13446712fe7588e04553a6ddc13da2143e0ed93395b.
//
// Solidity: event OperatorUpdated(uint256 tokenId, address operator)
func (_SubnetProvider *SubnetProviderFilterer) WatchOperatorUpdated(opts *bind.WatchOpts, sink chan<- *SubnetProviderOperatorUpdated) (event.Subscription, error) {

	logs, sub, err := _SubnetProvider.contract.WatchLogs(opts, "OperatorUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderOperatorUpdated)
				if err := _SubnetProvider.contract.UnpackLog(event, "OperatorUpdated", log); err != nil {
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

// ParseOperatorUpdated is a log parse operation binding the contract event 0xbc3f928a2240d921e006f13446712fe7588e04553a6ddc13da2143e0ed93395b.
//
// Solidity: event OperatorUpdated(uint256 tokenId, address operator)
func (_SubnetProvider *SubnetProviderFilterer) ParseOperatorUpdated(log types.Log) (*SubnetProviderOperatorUpdated, error) {
	event := new(SubnetProviderOperatorUpdated)
	if err := _SubnetProvider.contract.UnpackLog(event, "OperatorUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetProviderPeerNodeDeletedIterator is returned from FilterPeerNodeDeleted and is used to iterate over the raw logs and unpacked data for PeerNodeDeleted events raised by the SubnetProvider contract.
type SubnetProviderPeerNodeDeletedIterator struct {
	Event *SubnetProviderPeerNodeDeleted // Event containing the contract specifics and raw log

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
func (it *SubnetProviderPeerNodeDeletedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderPeerNodeDeleted)
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
		it.Event = new(SubnetProviderPeerNodeDeleted)
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
func (it *SubnetProviderPeerNodeDeletedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderPeerNodeDeletedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderPeerNodeDeleted represents a PeerNodeDeleted event raised by the SubnetProvider contract.
type SubnetProviderPeerNodeDeleted struct {
	TokenId *big.Int
	PeerId  string
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPeerNodeDeleted is a free log retrieval operation binding the contract event 0x3146c234fc656b2d6bd448c8a0a8deef61fc0efd2c579531e2f4843624249ec7.
//
// Solidity: event PeerNodeDeleted(uint256 indexed tokenId, string peerId)
func (_SubnetProvider *SubnetProviderFilterer) FilterPeerNodeDeleted(opts *bind.FilterOpts, tokenId []*big.Int) (*SubnetProviderPeerNodeDeletedIterator, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _SubnetProvider.contract.FilterLogs(opts, "PeerNodeDeleted", tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &SubnetProviderPeerNodeDeletedIterator{contract: _SubnetProvider.contract, event: "PeerNodeDeleted", logs: logs, sub: sub}, nil
}

// WatchPeerNodeDeleted is a free log subscription operation binding the contract event 0x3146c234fc656b2d6bd448c8a0a8deef61fc0efd2c579531e2f4843624249ec7.
//
// Solidity: event PeerNodeDeleted(uint256 indexed tokenId, string peerId)
func (_SubnetProvider *SubnetProviderFilterer) WatchPeerNodeDeleted(opts *bind.WatchOpts, sink chan<- *SubnetProviderPeerNodeDeleted, tokenId []*big.Int) (event.Subscription, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _SubnetProvider.contract.WatchLogs(opts, "PeerNodeDeleted", tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderPeerNodeDeleted)
				if err := _SubnetProvider.contract.UnpackLog(event, "PeerNodeDeleted", log); err != nil {
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

// ParsePeerNodeDeleted is a log parse operation binding the contract event 0x3146c234fc656b2d6bd448c8a0a8deef61fc0efd2c579531e2f4843624249ec7.
//
// Solidity: event PeerNodeDeleted(uint256 indexed tokenId, string peerId)
func (_SubnetProvider *SubnetProviderFilterer) ParsePeerNodeDeleted(log types.Log) (*SubnetProviderPeerNodeDeleted, error) {
	event := new(SubnetProviderPeerNodeDeleted)
	if err := _SubnetProvider.contract.UnpackLog(event, "PeerNodeDeleted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetProviderPeerNodeRegisteredIterator is returned from FilterPeerNodeRegistered and is used to iterate over the raw logs and unpacked data for PeerNodeRegistered events raised by the SubnetProvider contract.
type SubnetProviderPeerNodeRegisteredIterator struct {
	Event *SubnetProviderPeerNodeRegistered // Event containing the contract specifics and raw log

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
func (it *SubnetProviderPeerNodeRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderPeerNodeRegistered)
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
		it.Event = new(SubnetProviderPeerNodeRegistered)
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
func (it *SubnetProviderPeerNodeRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderPeerNodeRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderPeerNodeRegistered represents a PeerNodeRegistered event raised by the SubnetProvider contract.
type SubnetProviderPeerNodeRegistered struct {
	TokenId  *big.Int
	PeerId   string
	Metadata string
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterPeerNodeRegistered is a free log retrieval operation binding the contract event 0x0871c769e3d0dd7570972a54898f3d91c96d0f36f068b0074dc418d2a0621274.
//
// Solidity: event PeerNodeRegistered(uint256 indexed tokenId, string peerId, string metadata)
func (_SubnetProvider *SubnetProviderFilterer) FilterPeerNodeRegistered(opts *bind.FilterOpts, tokenId []*big.Int) (*SubnetProviderPeerNodeRegisteredIterator, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _SubnetProvider.contract.FilterLogs(opts, "PeerNodeRegistered", tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &SubnetProviderPeerNodeRegisteredIterator{contract: _SubnetProvider.contract, event: "PeerNodeRegistered", logs: logs, sub: sub}, nil
}

// WatchPeerNodeRegistered is a free log subscription operation binding the contract event 0x0871c769e3d0dd7570972a54898f3d91c96d0f36f068b0074dc418d2a0621274.
//
// Solidity: event PeerNodeRegistered(uint256 indexed tokenId, string peerId, string metadata)
func (_SubnetProvider *SubnetProviderFilterer) WatchPeerNodeRegistered(opts *bind.WatchOpts, sink chan<- *SubnetProviderPeerNodeRegistered, tokenId []*big.Int) (event.Subscription, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _SubnetProvider.contract.WatchLogs(opts, "PeerNodeRegistered", tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderPeerNodeRegistered)
				if err := _SubnetProvider.contract.UnpackLog(event, "PeerNodeRegistered", log); err != nil {
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

// ParsePeerNodeRegistered is a log parse operation binding the contract event 0x0871c769e3d0dd7570972a54898f3d91c96d0f36f068b0074dc418d2a0621274.
//
// Solidity: event PeerNodeRegistered(uint256 indexed tokenId, string peerId, string metadata)
func (_SubnetProvider *SubnetProviderFilterer) ParsePeerNodeRegistered(log types.Log) (*SubnetProviderPeerNodeRegistered, error) {
	event := new(SubnetProviderPeerNodeRegistered)
	if err := _SubnetProvider.contract.UnpackLog(event, "PeerNodeRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetProviderPeerNodeUpdatedIterator is returned from FilterPeerNodeUpdated and is used to iterate over the raw logs and unpacked data for PeerNodeUpdated events raised by the SubnetProvider contract.
type SubnetProviderPeerNodeUpdatedIterator struct {
	Event *SubnetProviderPeerNodeUpdated // Event containing the contract specifics and raw log

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
func (it *SubnetProviderPeerNodeUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderPeerNodeUpdated)
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
		it.Event = new(SubnetProviderPeerNodeUpdated)
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
func (it *SubnetProviderPeerNodeUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderPeerNodeUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderPeerNodeUpdated represents a PeerNodeUpdated event raised by the SubnetProvider contract.
type SubnetProviderPeerNodeUpdated struct {
	TokenId  *big.Int
	PeerId   string
	Metadata string
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterPeerNodeUpdated is a free log retrieval operation binding the contract event 0xaac9f0167af86d3540388db7322c095eb7bf138b2ee88ffee321073dac6e90eb.
//
// Solidity: event PeerNodeUpdated(uint256 indexed tokenId, string peerId, string metadata)
func (_SubnetProvider *SubnetProviderFilterer) FilterPeerNodeUpdated(opts *bind.FilterOpts, tokenId []*big.Int) (*SubnetProviderPeerNodeUpdatedIterator, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _SubnetProvider.contract.FilterLogs(opts, "PeerNodeUpdated", tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &SubnetProviderPeerNodeUpdatedIterator{contract: _SubnetProvider.contract, event: "PeerNodeUpdated", logs: logs, sub: sub}, nil
}

// WatchPeerNodeUpdated is a free log subscription operation binding the contract event 0xaac9f0167af86d3540388db7322c095eb7bf138b2ee88ffee321073dac6e90eb.
//
// Solidity: event PeerNodeUpdated(uint256 indexed tokenId, string peerId, string metadata)
func (_SubnetProvider *SubnetProviderFilterer) WatchPeerNodeUpdated(opts *bind.WatchOpts, sink chan<- *SubnetProviderPeerNodeUpdated, tokenId []*big.Int) (event.Subscription, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _SubnetProvider.contract.WatchLogs(opts, "PeerNodeUpdated", tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderPeerNodeUpdated)
				if err := _SubnetProvider.contract.UnpackLog(event, "PeerNodeUpdated", log); err != nil {
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

// ParsePeerNodeUpdated is a log parse operation binding the contract event 0xaac9f0167af86d3540388db7322c095eb7bf138b2ee88ffee321073dac6e90eb.
//
// Solidity: event PeerNodeUpdated(uint256 indexed tokenId, string peerId, string metadata)
func (_SubnetProvider *SubnetProviderFilterer) ParsePeerNodeUpdated(log types.Log) (*SubnetProviderPeerNodeUpdated, error) {
	event := new(SubnetProviderPeerNodeUpdated)
	if err := _SubnetProvider.contract.UnpackLog(event, "PeerNodeUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetProviderProviderDeletedIterator is returned from FilterProviderDeleted and is used to iterate over the raw logs and unpacked data for ProviderDeleted events raised by the SubnetProvider contract.
type SubnetProviderProviderDeletedIterator struct {
	Event *SubnetProviderProviderDeleted // Event containing the contract specifics and raw log

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
func (it *SubnetProviderProviderDeletedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderProviderDeleted)
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
		it.Event = new(SubnetProviderProviderDeleted)
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
func (it *SubnetProviderProviderDeletedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderProviderDeletedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderProviderDeleted represents a ProviderDeleted event raised by the SubnetProvider contract.
type SubnetProviderProviderDeleted struct {
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterProviderDeleted is a free log retrieval operation binding the contract event 0x99921e5d7939d41d3bfabe71f0087e38fccb316796abb58a905cd3dbf6821943.
//
// Solidity: event ProviderDeleted(uint256 tokenId)
func (_SubnetProvider *SubnetProviderFilterer) FilterProviderDeleted(opts *bind.FilterOpts) (*SubnetProviderProviderDeletedIterator, error) {

	logs, sub, err := _SubnetProvider.contract.FilterLogs(opts, "ProviderDeleted")
	if err != nil {
		return nil, err
	}
	return &SubnetProviderProviderDeletedIterator{contract: _SubnetProvider.contract, event: "ProviderDeleted", logs: logs, sub: sub}, nil
}

// WatchProviderDeleted is a free log subscription operation binding the contract event 0x99921e5d7939d41d3bfabe71f0087e38fccb316796abb58a905cd3dbf6821943.
//
// Solidity: event ProviderDeleted(uint256 tokenId)
func (_SubnetProvider *SubnetProviderFilterer) WatchProviderDeleted(opts *bind.WatchOpts, sink chan<- *SubnetProviderProviderDeleted) (event.Subscription, error) {

	logs, sub, err := _SubnetProvider.contract.WatchLogs(opts, "ProviderDeleted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderProviderDeleted)
				if err := _SubnetProvider.contract.UnpackLog(event, "ProviderDeleted", log); err != nil {
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

// ParseProviderDeleted is a log parse operation binding the contract event 0x99921e5d7939d41d3bfabe71f0087e38fccb316796abb58a905cd3dbf6821943.
//
// Solidity: event ProviderDeleted(uint256 tokenId)
func (_SubnetProvider *SubnetProviderFilterer) ParseProviderDeleted(log types.Log) (*SubnetProviderProviderDeleted, error) {
	event := new(SubnetProviderProviderDeleted)
	if err := _SubnetProvider.contract.UnpackLog(event, "ProviderDeleted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetProviderProviderRegisteredIterator is returned from FilterProviderRegistered and is used to iterate over the raw logs and unpacked data for ProviderRegistered events raised by the SubnetProvider contract.
type SubnetProviderProviderRegisteredIterator struct {
	Event *SubnetProviderProviderRegistered // Event containing the contract specifics and raw log

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
func (it *SubnetProviderProviderRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderProviderRegistered)
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
		it.Event = new(SubnetProviderProviderRegistered)
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
func (it *SubnetProviderProviderRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderProviderRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderProviderRegistered represents a ProviderRegistered event raised by the SubnetProvider contract.
type SubnetProviderProviderRegistered struct {
	ProviderAddress common.Address
	TokenId         *big.Int
	ProviderName    string
	Metadata        string
	Operator        common.Address
	Website         string
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterProviderRegistered is a free log retrieval operation binding the contract event 0x0c9a46e79f429968ccc567ee367b6795187112eb50e9788d15a8cd955a4bfe94.
//
// Solidity: event ProviderRegistered(address providerAddress, uint256 tokenId, string providerName, string metadata, address operator, string website)
func (_SubnetProvider *SubnetProviderFilterer) FilterProviderRegistered(opts *bind.FilterOpts) (*SubnetProviderProviderRegisteredIterator, error) {

	logs, sub, err := _SubnetProvider.contract.FilterLogs(opts, "ProviderRegistered")
	if err != nil {
		return nil, err
	}
	return &SubnetProviderProviderRegisteredIterator{contract: _SubnetProvider.contract, event: "ProviderRegistered", logs: logs, sub: sub}, nil
}

// WatchProviderRegistered is a free log subscription operation binding the contract event 0x0c9a46e79f429968ccc567ee367b6795187112eb50e9788d15a8cd955a4bfe94.
//
// Solidity: event ProviderRegistered(address providerAddress, uint256 tokenId, string providerName, string metadata, address operator, string website)
func (_SubnetProvider *SubnetProviderFilterer) WatchProviderRegistered(opts *bind.WatchOpts, sink chan<- *SubnetProviderProviderRegistered) (event.Subscription, error) {

	logs, sub, err := _SubnetProvider.contract.WatchLogs(opts, "ProviderRegistered")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderProviderRegistered)
				if err := _SubnetProvider.contract.UnpackLog(event, "ProviderRegistered", log); err != nil {
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

// ParseProviderRegistered is a log parse operation binding the contract event 0x0c9a46e79f429968ccc567ee367b6795187112eb50e9788d15a8cd955a4bfe94.
//
// Solidity: event ProviderRegistered(address providerAddress, uint256 tokenId, string providerName, string metadata, address operator, string website)
func (_SubnetProvider *SubnetProviderFilterer) ParseProviderRegistered(log types.Log) (*SubnetProviderProviderRegistered, error) {
	event := new(SubnetProviderProviderRegistered)
	if err := _SubnetProvider.contract.UnpackLog(event, "ProviderRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetProviderProviderUpdatedIterator is returned from FilterProviderUpdated and is used to iterate over the raw logs and unpacked data for ProviderUpdated events raised by the SubnetProvider contract.
type SubnetProviderProviderUpdatedIterator struct {
	Event *SubnetProviderProviderUpdated // Event containing the contract specifics and raw log

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
func (it *SubnetProviderProviderUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderProviderUpdated)
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
		it.Event = new(SubnetProviderProviderUpdated)
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
func (it *SubnetProviderProviderUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderProviderUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderProviderUpdated represents a ProviderUpdated event raised by the SubnetProvider contract.
type SubnetProviderProviderUpdated struct {
	TokenId      *big.Int
	ProviderName string
	Metadata     string
	Operator     common.Address
	Website      string
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterProviderUpdated is a free log retrieval operation binding the contract event 0xec62d12ef065e64d6ca9503068ed3bdffb1ed5294a1429d7482c6c6e776fea5e.
//
// Solidity: event ProviderUpdated(uint256 tokenId, string providerName, string metadata, address operator, string website)
func (_SubnetProvider *SubnetProviderFilterer) FilterProviderUpdated(opts *bind.FilterOpts) (*SubnetProviderProviderUpdatedIterator, error) {

	logs, sub, err := _SubnetProvider.contract.FilterLogs(opts, "ProviderUpdated")
	if err != nil {
		return nil, err
	}
	return &SubnetProviderProviderUpdatedIterator{contract: _SubnetProvider.contract, event: "ProviderUpdated", logs: logs, sub: sub}, nil
}

// WatchProviderUpdated is a free log subscription operation binding the contract event 0xec62d12ef065e64d6ca9503068ed3bdffb1ed5294a1429d7482c6c6e776fea5e.
//
// Solidity: event ProviderUpdated(uint256 tokenId, string providerName, string metadata, address operator, string website)
func (_SubnetProvider *SubnetProviderFilterer) WatchProviderUpdated(opts *bind.WatchOpts, sink chan<- *SubnetProviderProviderUpdated) (event.Subscription, error) {

	logs, sub, err := _SubnetProvider.contract.WatchLogs(opts, "ProviderUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderProviderUpdated)
				if err := _SubnetProvider.contract.UnpackLog(event, "ProviderUpdated", log); err != nil {
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

// ParseProviderUpdated is a log parse operation binding the contract event 0xec62d12ef065e64d6ca9503068ed3bdffb1ed5294a1429d7482c6c6e776fea5e.
//
// Solidity: event ProviderUpdated(uint256 tokenId, string providerName, string metadata, address operator, string website)
func (_SubnetProvider *SubnetProviderFilterer) ParseProviderUpdated(log types.Log) (*SubnetProviderProviderUpdated, error) {
	event := new(SubnetProviderProviderUpdated)
	if err := _SubnetProvider.contract.UnpackLog(event, "ProviderUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetProviderTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the SubnetProvider contract.
type SubnetProviderTransferIterator struct {
	Event *SubnetProviderTransfer // Event containing the contract specifics and raw log

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
func (it *SubnetProviderTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderTransfer)
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
		it.Event = new(SubnetProviderTransfer)
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
func (it *SubnetProviderTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderTransfer represents a Transfer event raised by the SubnetProvider contract.
type SubnetProviderTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_SubnetProvider *SubnetProviderFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address, tokenId []*big.Int) (*SubnetProviderTransferIterator, error) {

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

	logs, sub, err := _SubnetProvider.contract.FilterLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &SubnetProviderTransferIterator{contract: _SubnetProvider.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_SubnetProvider *SubnetProviderFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *SubnetProviderTransfer, from []common.Address, to []common.Address, tokenId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _SubnetProvider.contract.WatchLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderTransfer)
				if err := _SubnetProvider.contract.UnpackLog(event, "Transfer", log); err != nil {
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
func (_SubnetProvider *SubnetProviderFilterer) ParseTransfer(log types.Log) (*SubnetProviderTransfer, error) {
	event := new(SubnetProviderTransfer)
	if err := _SubnetProvider.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetProviderWebsiteUpdatedIterator is returned from FilterWebsiteUpdated and is used to iterate over the raw logs and unpacked data for WebsiteUpdated events raised by the SubnetProvider contract.
type SubnetProviderWebsiteUpdatedIterator struct {
	Event *SubnetProviderWebsiteUpdated // Event containing the contract specifics and raw log

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
func (it *SubnetProviderWebsiteUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetProviderWebsiteUpdated)
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
		it.Event = new(SubnetProviderWebsiteUpdated)
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
func (it *SubnetProviderWebsiteUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetProviderWebsiteUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetProviderWebsiteUpdated represents a WebsiteUpdated event raised by the SubnetProvider contract.
type SubnetProviderWebsiteUpdated struct {
	TokenId *big.Int
	Website string
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterWebsiteUpdated is a free log retrieval operation binding the contract event 0xdc8326ca64efd094910ca07722944733f570f363270e574c6febb5c3ab5a58d0.
//
// Solidity: event WebsiteUpdated(uint256 tokenId, string website)
func (_SubnetProvider *SubnetProviderFilterer) FilterWebsiteUpdated(opts *bind.FilterOpts) (*SubnetProviderWebsiteUpdatedIterator, error) {

	logs, sub, err := _SubnetProvider.contract.FilterLogs(opts, "WebsiteUpdated")
	if err != nil {
		return nil, err
	}
	return &SubnetProviderWebsiteUpdatedIterator{contract: _SubnetProvider.contract, event: "WebsiteUpdated", logs: logs, sub: sub}, nil
}

// WatchWebsiteUpdated is a free log subscription operation binding the contract event 0xdc8326ca64efd094910ca07722944733f570f363270e574c6febb5c3ab5a58d0.
//
// Solidity: event WebsiteUpdated(uint256 tokenId, string website)
func (_SubnetProvider *SubnetProviderFilterer) WatchWebsiteUpdated(opts *bind.WatchOpts, sink chan<- *SubnetProviderWebsiteUpdated) (event.Subscription, error) {

	logs, sub, err := _SubnetProvider.contract.WatchLogs(opts, "WebsiteUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetProviderWebsiteUpdated)
				if err := _SubnetProvider.contract.UnpackLog(event, "WebsiteUpdated", log); err != nil {
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

// ParseWebsiteUpdated is a log parse operation binding the contract event 0xdc8326ca64efd094910ca07722944733f570f363270e574c6febb5c3ab5a58d0.
//
// Solidity: event WebsiteUpdated(uint256 tokenId, string website)
func (_SubnetProvider *SubnetProviderFilterer) ParseWebsiteUpdated(log types.Log) (*SubnetProviderWebsiteUpdated, error) {
	event := new(SubnetProviderWebsiteUpdated)
	if err := _SubnetProvider.contract.UnpackLog(event, "WebsiteUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
