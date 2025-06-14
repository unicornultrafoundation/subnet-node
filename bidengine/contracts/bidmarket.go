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

// SubnetBidMarketplaceBid is an auto generated low-level Go binding around an user-defined struct.
type SubnetBidMarketplaceBid struct {
	Provider       common.Address
	PricePerSecond *big.Int
	Status         uint8
	CreatedAt      *big.Int
	ProviderId     *big.Int
	MachineId      *big.Int
}

// BidMarketMetaData contains all meta data concerning the BidMarket contract.
var BidMarketMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"AddressEmptyCode\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"AddressInsufficientBalance\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"SafeERC20FailedOperation\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"}],\"name\":\"BidAccepted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"bidIndex\",\"type\":\"uint256\"}],\"name\":\"BidCancelled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"machineId\",\"type\":\"uint256\"}],\"name\":\"BidSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"}],\"name\":\"BidTimeExpired\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"oldLimit\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newLimit\",\"type\":\"uint256\"}],\"name\":\"BidTimeLimitUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"wallet\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"FeesWithdrawn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OrderCancelled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"refundAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"name\":\"OrderClosed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"name\":\"OrderCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"additionalDuration\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newExpiry\",\"type\":\"uint256\"}],\"name\":\"OrderExtended\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"oldFee\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newFee\",\"type\":\"uint256\"}],\"name\":\"PlatformFeeUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"oldWallet\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newWallet\",\"type\":\"address\"}],\"name\":\"PlatformWalletUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"newSpecs\",\"type\":\"string\"}],\"name\":\"SpecsUpdated\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"bidIndex\",\"type\":\"uint256\"}],\"name\":\"acceptBid\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"bidTimeLimit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"bidIndex\",\"type\":\"uint256\"}],\"name\":\"cancelBid\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"}],\"name\":\"cancelOrder\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"}],\"name\":\"claimPayment\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"name\":\"closeOrder\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"machineType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minBidPrice\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxBidPrice\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"region\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"cpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuMemory\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"memoryMB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"diskGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"uploadMbps\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"downloadMbps\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"specs\",\"type\":\"string\"}],\"name\":\"createOrder\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"}],\"name\":\"extend\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"}],\"name\":\"getBids\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"enumSubnetBidMarketplace.BidStatus\",\"name\":\"status\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"machineId\",\"type\":\"uint256\"}],\"internalType\":\"structSubnetBidMarketplace.Bid[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"}],\"name\":\"getRemainingBidTime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_paymentToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_subnetProviderContract\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"}],\"name\":\"isBiddingOpen\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"orderBids\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"enumSubnetBidMarketplace.BidStatus\",\"name\":\"status\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"machineId\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"orderCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"orders\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"machineType\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"enumSubnetBidMarketplace.OrderStatus\",\"name\":\"status\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minBidPrice\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxBidPrice\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"acceptedBidPricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"parentOrderId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"paymentToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"cpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuCores\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gpuMemory\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"memoryMB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"diskGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"uploadMbps\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"downloadMbps\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"region\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"specs\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"acceptedProviderId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"acceptedMachineId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"startAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"expiredAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lastPaidAt\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"paymentToken\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"platformFeePercentage\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"platformWallet\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newBidTimeLimit\",\"type\":\"uint256\"}],\"name\":\"setBidTimeLimit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_paymentToken\",\"type\":\"address\"}],\"name\":\"setPaymentConfig\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newFeePercentage\",\"type\":\"uint256\"}],\"name\":\"setPlatformFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newWallet\",\"type\":\"address\"}],\"name\":\"setPlatformWallet\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_subnetProviderContract\",\"type\":\"address\"}],\"name\":\"setSubnetProviderContract\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"machineId\",\"type\":\"uint256\"}],\"name\":\"submitBid\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"subnetProviderContract\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalAccumulatedFees\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdrawFees\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// BidMarketABI is the input ABI used to generate the binding from.
// Deprecated: Use BidMarketMetaData.ABI instead.
var BidMarketABI = BidMarketMetaData.ABI

// BidMarket is an auto generated Go binding around an Ethereum contract.
type BidMarket struct {
	BidMarketCaller     // Read-only binding to the contract
	BidMarketTransactor // Write-only binding to the contract
	BidMarketFilterer   // Log filterer for contract events
}

// BidMarketCaller is an auto generated read-only Go binding around an Ethereum contract.
type BidMarketCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BidMarketTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BidMarketTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BidMarketFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BidMarketFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BidMarketSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BidMarketSession struct {
	Contract     *BidMarket        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BidMarketCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BidMarketCallerSession struct {
	Contract *BidMarketCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// BidMarketTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BidMarketTransactorSession struct {
	Contract     *BidMarketTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// BidMarketRaw is an auto generated low-level Go binding around an Ethereum contract.
type BidMarketRaw struct {
	Contract *BidMarket // Generic contract binding to access the raw methods on
}

// BidMarketCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BidMarketCallerRaw struct {
	Contract *BidMarketCaller // Generic read-only contract binding to access the raw methods on
}

// BidMarketTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BidMarketTransactorRaw struct {
	Contract *BidMarketTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBidMarket creates a new instance of BidMarket, bound to a specific deployed contract.
func NewBidMarket(address common.Address, backend bind.ContractBackend) (*BidMarket, error) {
	contract, err := bindBidMarket(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BidMarket{BidMarketCaller: BidMarketCaller{contract: contract}, BidMarketTransactor: BidMarketTransactor{contract: contract}, BidMarketFilterer: BidMarketFilterer{contract: contract}}, nil
}

// NewBidMarketCaller creates a new read-only instance of BidMarket, bound to a specific deployed contract.
func NewBidMarketCaller(address common.Address, caller bind.ContractCaller) (*BidMarketCaller, error) {
	contract, err := bindBidMarket(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BidMarketCaller{contract: contract}, nil
}

// NewBidMarketTransactor creates a new write-only instance of BidMarket, bound to a specific deployed contract.
func NewBidMarketTransactor(address common.Address, transactor bind.ContractTransactor) (*BidMarketTransactor, error) {
	contract, err := bindBidMarket(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BidMarketTransactor{contract: contract}, nil
}

// NewBidMarketFilterer creates a new log filterer instance of BidMarket, bound to a specific deployed contract.
func NewBidMarketFilterer(address common.Address, filterer bind.ContractFilterer) (*BidMarketFilterer, error) {
	contract, err := bindBidMarket(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BidMarketFilterer{contract: contract}, nil
}

// bindBidMarket binds a generic wrapper to an already deployed contract.
func bindBidMarket(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BidMarketMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BidMarket *BidMarketRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BidMarket.Contract.BidMarketCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BidMarket *BidMarketRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BidMarket.Contract.BidMarketTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BidMarket *BidMarketRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BidMarket.Contract.BidMarketTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BidMarket *BidMarketCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BidMarket.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BidMarket *BidMarketTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BidMarket.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BidMarket *BidMarketTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BidMarket.Contract.contract.Transact(opts, method, params...)
}

// BidTimeLimit is a free data retrieval call binding the contract method 0x05993f06.
//
// Solidity: function bidTimeLimit() view returns(uint256)
func (_BidMarket *BidMarketCaller) BidTimeLimit(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BidMarket.contract.Call(opts, &out, "bidTimeLimit")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BidTimeLimit is a free data retrieval call binding the contract method 0x05993f06.
//
// Solidity: function bidTimeLimit() view returns(uint256)
func (_BidMarket *BidMarketSession) BidTimeLimit() (*big.Int, error) {
	return _BidMarket.Contract.BidTimeLimit(&_BidMarket.CallOpts)
}

// BidTimeLimit is a free data retrieval call binding the contract method 0x05993f06.
//
// Solidity: function bidTimeLimit() view returns(uint256)
func (_BidMarket *BidMarketCallerSession) BidTimeLimit() (*big.Int, error) {
	return _BidMarket.Contract.BidTimeLimit(&_BidMarket.CallOpts)
}

// GetBids is a free data retrieval call binding the contract method 0x131d9a27.
//
// Solidity: function getBids(uint256 orderId) view returns((address,uint256,uint8,uint256,uint256,uint256)[])
func (_BidMarket *BidMarketCaller) GetBids(opts *bind.CallOpts, orderId *big.Int) ([]SubnetBidMarketplaceBid, error) {
	var out []interface{}
	err := _BidMarket.contract.Call(opts, &out, "getBids", orderId)

	if err != nil {
		return *new([]SubnetBidMarketplaceBid), err
	}

	out0 := *abi.ConvertType(out[0], new([]SubnetBidMarketplaceBid)).(*[]SubnetBidMarketplaceBid)

	return out0, err

}

// GetBids is a free data retrieval call binding the contract method 0x131d9a27.
//
// Solidity: function getBids(uint256 orderId) view returns((address,uint256,uint8,uint256,uint256,uint256)[])
func (_BidMarket *BidMarketSession) GetBids(orderId *big.Int) ([]SubnetBidMarketplaceBid, error) {
	return _BidMarket.Contract.GetBids(&_BidMarket.CallOpts, orderId)
}

// GetBids is a free data retrieval call binding the contract method 0x131d9a27.
//
// Solidity: function getBids(uint256 orderId) view returns((address,uint256,uint8,uint256,uint256,uint256)[])
func (_BidMarket *BidMarketCallerSession) GetBids(orderId *big.Int) ([]SubnetBidMarketplaceBid, error) {
	return _BidMarket.Contract.GetBids(&_BidMarket.CallOpts, orderId)
}

// GetRemainingBidTime is a free data retrieval call binding the contract method 0x3567e47e.
//
// Solidity: function getRemainingBidTime(uint256 orderId) view returns(uint256)
func (_BidMarket *BidMarketCaller) GetRemainingBidTime(opts *bind.CallOpts, orderId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _BidMarket.contract.Call(opts, &out, "getRemainingBidTime", orderId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRemainingBidTime is a free data retrieval call binding the contract method 0x3567e47e.
//
// Solidity: function getRemainingBidTime(uint256 orderId) view returns(uint256)
func (_BidMarket *BidMarketSession) GetRemainingBidTime(orderId *big.Int) (*big.Int, error) {
	return _BidMarket.Contract.GetRemainingBidTime(&_BidMarket.CallOpts, orderId)
}

// GetRemainingBidTime is a free data retrieval call binding the contract method 0x3567e47e.
//
// Solidity: function getRemainingBidTime(uint256 orderId) view returns(uint256)
func (_BidMarket *BidMarketCallerSession) GetRemainingBidTime(orderId *big.Int) (*big.Int, error) {
	return _BidMarket.Contract.GetRemainingBidTime(&_BidMarket.CallOpts, orderId)
}

// IsBiddingOpen is a free data retrieval call binding the contract method 0xe5abcee6.
//
// Solidity: function isBiddingOpen(uint256 orderId) view returns(bool)
func (_BidMarket *BidMarketCaller) IsBiddingOpen(opts *bind.CallOpts, orderId *big.Int) (bool, error) {
	var out []interface{}
	err := _BidMarket.contract.Call(opts, &out, "isBiddingOpen", orderId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsBiddingOpen is a free data retrieval call binding the contract method 0xe5abcee6.
//
// Solidity: function isBiddingOpen(uint256 orderId) view returns(bool)
func (_BidMarket *BidMarketSession) IsBiddingOpen(orderId *big.Int) (bool, error) {
	return _BidMarket.Contract.IsBiddingOpen(&_BidMarket.CallOpts, orderId)
}

// IsBiddingOpen is a free data retrieval call binding the contract method 0xe5abcee6.
//
// Solidity: function isBiddingOpen(uint256 orderId) view returns(bool)
func (_BidMarket *BidMarketCallerSession) IsBiddingOpen(orderId *big.Int) (bool, error) {
	return _BidMarket.Contract.IsBiddingOpen(&_BidMarket.CallOpts, orderId)
}

// OrderBids is a free data retrieval call binding the contract method 0xef5a59c8.
//
// Solidity: function orderBids(uint256 , uint256 ) view returns(address provider, uint256 pricePerSecond, uint8 status, uint256 createdAt, uint256 providerId, uint256 machineId)
func (_BidMarket *BidMarketCaller) OrderBids(opts *bind.CallOpts, arg0 *big.Int, arg1 *big.Int) (struct {
	Provider       common.Address
	PricePerSecond *big.Int
	Status         uint8
	CreatedAt      *big.Int
	ProviderId     *big.Int
	MachineId      *big.Int
}, error) {
	var out []interface{}
	err := _BidMarket.contract.Call(opts, &out, "orderBids", arg0, arg1)

	outstruct := new(struct {
		Provider       common.Address
		PricePerSecond *big.Int
		Status         uint8
		CreatedAt      *big.Int
		ProviderId     *big.Int
		MachineId      *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Provider = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.PricePerSecond = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Status = *abi.ConvertType(out[2], new(uint8)).(*uint8)
	outstruct.CreatedAt = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.ProviderId = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.MachineId = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// OrderBids is a free data retrieval call binding the contract method 0xef5a59c8.
//
// Solidity: function orderBids(uint256 , uint256 ) view returns(address provider, uint256 pricePerSecond, uint8 status, uint256 createdAt, uint256 providerId, uint256 machineId)
func (_BidMarket *BidMarketSession) OrderBids(arg0 *big.Int, arg1 *big.Int) (struct {
	Provider       common.Address
	PricePerSecond *big.Int
	Status         uint8
	CreatedAt      *big.Int
	ProviderId     *big.Int
	MachineId      *big.Int
}, error) {
	return _BidMarket.Contract.OrderBids(&_BidMarket.CallOpts, arg0, arg1)
}

// OrderBids is a free data retrieval call binding the contract method 0xef5a59c8.
//
// Solidity: function orderBids(uint256 , uint256 ) view returns(address provider, uint256 pricePerSecond, uint8 status, uint256 createdAt, uint256 providerId, uint256 machineId)
func (_BidMarket *BidMarketCallerSession) OrderBids(arg0 *big.Int, arg1 *big.Int) (struct {
	Provider       common.Address
	PricePerSecond *big.Int
	Status         uint8
	CreatedAt      *big.Int
	ProviderId     *big.Int
	MachineId      *big.Int
}, error) {
	return _BidMarket.Contract.OrderBids(&_BidMarket.CallOpts, arg0, arg1)
}

// OrderCount is a free data retrieval call binding the contract method 0x2453ffa8.
//
// Solidity: function orderCount() view returns(uint256)
func (_BidMarket *BidMarketCaller) OrderCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BidMarket.contract.Call(opts, &out, "orderCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// OrderCount is a free data retrieval call binding the contract method 0x2453ffa8.
//
// Solidity: function orderCount() view returns(uint256)
func (_BidMarket *BidMarketSession) OrderCount() (*big.Int, error) {
	return _BidMarket.Contract.OrderCount(&_BidMarket.CallOpts)
}

// OrderCount is a free data retrieval call binding the contract method 0x2453ffa8.
//
// Solidity: function orderCount() view returns(uint256)
func (_BidMarket *BidMarketCallerSession) OrderCount() (*big.Int, error) {
	return _BidMarket.Contract.OrderCount(&_BidMarket.CallOpts)
}

// Orders is a free data retrieval call binding the contract method 0xa85c38ef.
//
// Solidity: function orders(uint256 ) view returns(uint256 machineType, address owner, uint8 status, uint256 createdAt, uint256 duration, uint256 minBidPrice, uint256 maxBidPrice, uint256 acceptedBidPricePerSecond, uint256 parentOrderId, address paymentToken, uint256 cpuCores, uint256 gpuCores, uint256 gpuMemory, uint256 memoryMB, uint256 diskGB, uint256 uploadMbps, uint256 downloadMbps, uint256 region, string specs, uint256 acceptedProviderId, uint256 acceptedMachineId, uint256 startAt, uint256 expiredAt, uint256 lastPaidAt)
func (_BidMarket *BidMarketCaller) Orders(opts *bind.CallOpts, arg0 *big.Int) (struct {
	MachineType               *big.Int
	Owner                     common.Address
	Status                    uint8
	CreatedAt                 *big.Int
	Duration                  *big.Int
	MinBidPrice               *big.Int
	MaxBidPrice               *big.Int
	AcceptedBidPricePerSecond *big.Int
	ParentOrderId             *big.Int
	PaymentToken              common.Address
	CpuCores                  *big.Int
	GpuCores                  *big.Int
	GpuMemory                 *big.Int
	MemoryMB                  *big.Int
	DiskGB                    *big.Int
	UploadMbps                *big.Int
	DownloadMbps              *big.Int
	Region                    *big.Int
	Specs                     string
	AcceptedProviderId        *big.Int
	AcceptedMachineId         *big.Int
	StartAt                   *big.Int
	ExpiredAt                 *big.Int
	LastPaidAt                *big.Int
}, error) {
	var out []interface{}
	err := _BidMarket.contract.Call(opts, &out, "orders", arg0)

	outstruct := new(struct {
		MachineType               *big.Int
		Owner                     common.Address
		Status                    uint8
		CreatedAt                 *big.Int
		Duration                  *big.Int
		MinBidPrice               *big.Int
		MaxBidPrice               *big.Int
		AcceptedBidPricePerSecond *big.Int
		ParentOrderId             *big.Int
		PaymentToken              common.Address
		CpuCores                  *big.Int
		GpuCores                  *big.Int
		GpuMemory                 *big.Int
		MemoryMB                  *big.Int
		DiskGB                    *big.Int
		UploadMbps                *big.Int
		DownloadMbps              *big.Int
		Region                    *big.Int
		Specs                     string
		AcceptedProviderId        *big.Int
		AcceptedMachineId         *big.Int
		StartAt                   *big.Int
		ExpiredAt                 *big.Int
		LastPaidAt                *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.MachineType = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Owner = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	outstruct.Status = *abi.ConvertType(out[2], new(uint8)).(*uint8)
	outstruct.CreatedAt = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.Duration = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.MinBidPrice = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)
	outstruct.MaxBidPrice = *abi.ConvertType(out[6], new(*big.Int)).(**big.Int)
	outstruct.AcceptedBidPricePerSecond = *abi.ConvertType(out[7], new(*big.Int)).(**big.Int)
	outstruct.ParentOrderId = *abi.ConvertType(out[8], new(*big.Int)).(**big.Int)
	outstruct.PaymentToken = *abi.ConvertType(out[9], new(common.Address)).(*common.Address)
	outstruct.CpuCores = *abi.ConvertType(out[10], new(*big.Int)).(**big.Int)
	outstruct.GpuCores = *abi.ConvertType(out[11], new(*big.Int)).(**big.Int)
	outstruct.GpuMemory = *abi.ConvertType(out[12], new(*big.Int)).(**big.Int)
	outstruct.MemoryMB = *abi.ConvertType(out[13], new(*big.Int)).(**big.Int)
	outstruct.DiskGB = *abi.ConvertType(out[14], new(*big.Int)).(**big.Int)
	outstruct.UploadMbps = *abi.ConvertType(out[15], new(*big.Int)).(**big.Int)
	outstruct.DownloadMbps = *abi.ConvertType(out[16], new(*big.Int)).(**big.Int)
	outstruct.Region = *abi.ConvertType(out[17], new(*big.Int)).(**big.Int)
	outstruct.Specs = *abi.ConvertType(out[18], new(string)).(*string)
	outstruct.AcceptedProviderId = *abi.ConvertType(out[19], new(*big.Int)).(**big.Int)
	outstruct.AcceptedMachineId = *abi.ConvertType(out[20], new(*big.Int)).(**big.Int)
	outstruct.StartAt = *abi.ConvertType(out[21], new(*big.Int)).(**big.Int)
	outstruct.ExpiredAt = *abi.ConvertType(out[22], new(*big.Int)).(**big.Int)
	outstruct.LastPaidAt = *abi.ConvertType(out[23], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Orders is a free data retrieval call binding the contract method 0xa85c38ef.
//
// Solidity: function orders(uint256 ) view returns(uint256 machineType, address owner, uint8 status, uint256 createdAt, uint256 duration, uint256 minBidPrice, uint256 maxBidPrice, uint256 acceptedBidPricePerSecond, uint256 parentOrderId, address paymentToken, uint256 cpuCores, uint256 gpuCores, uint256 gpuMemory, uint256 memoryMB, uint256 diskGB, uint256 uploadMbps, uint256 downloadMbps, uint256 region, string specs, uint256 acceptedProviderId, uint256 acceptedMachineId, uint256 startAt, uint256 expiredAt, uint256 lastPaidAt)
func (_BidMarket *BidMarketSession) Orders(arg0 *big.Int) (struct {
	MachineType               *big.Int
	Owner                     common.Address
	Status                    uint8
	CreatedAt                 *big.Int
	Duration                  *big.Int
	MinBidPrice               *big.Int
	MaxBidPrice               *big.Int
	AcceptedBidPricePerSecond *big.Int
	ParentOrderId             *big.Int
	PaymentToken              common.Address
	CpuCores                  *big.Int
	GpuCores                  *big.Int
	GpuMemory                 *big.Int
	MemoryMB                  *big.Int
	DiskGB                    *big.Int
	UploadMbps                *big.Int
	DownloadMbps              *big.Int
	Region                    *big.Int
	Specs                     string
	AcceptedProviderId        *big.Int
	AcceptedMachineId         *big.Int
	StartAt                   *big.Int
	ExpiredAt                 *big.Int
	LastPaidAt                *big.Int
}, error) {
	return _BidMarket.Contract.Orders(&_BidMarket.CallOpts, arg0)
}

// Orders is a free data retrieval call binding the contract method 0xa85c38ef.
//
// Solidity: function orders(uint256 ) view returns(uint256 machineType, address owner, uint8 status, uint256 createdAt, uint256 duration, uint256 minBidPrice, uint256 maxBidPrice, uint256 acceptedBidPricePerSecond, uint256 parentOrderId, address paymentToken, uint256 cpuCores, uint256 gpuCores, uint256 gpuMemory, uint256 memoryMB, uint256 diskGB, uint256 uploadMbps, uint256 downloadMbps, uint256 region, string specs, uint256 acceptedProviderId, uint256 acceptedMachineId, uint256 startAt, uint256 expiredAt, uint256 lastPaidAt)
func (_BidMarket *BidMarketCallerSession) Orders(arg0 *big.Int) (struct {
	MachineType               *big.Int
	Owner                     common.Address
	Status                    uint8
	CreatedAt                 *big.Int
	Duration                  *big.Int
	MinBidPrice               *big.Int
	MaxBidPrice               *big.Int
	AcceptedBidPricePerSecond *big.Int
	ParentOrderId             *big.Int
	PaymentToken              common.Address
	CpuCores                  *big.Int
	GpuCores                  *big.Int
	GpuMemory                 *big.Int
	MemoryMB                  *big.Int
	DiskGB                    *big.Int
	UploadMbps                *big.Int
	DownloadMbps              *big.Int
	Region                    *big.Int
	Specs                     string
	AcceptedProviderId        *big.Int
	AcceptedMachineId         *big.Int
	StartAt                   *big.Int
	ExpiredAt                 *big.Int
	LastPaidAt                *big.Int
}, error) {
	return _BidMarket.Contract.Orders(&_BidMarket.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BidMarket *BidMarketCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BidMarket.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BidMarket *BidMarketSession) Owner() (common.Address, error) {
	return _BidMarket.Contract.Owner(&_BidMarket.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BidMarket *BidMarketCallerSession) Owner() (common.Address, error) {
	return _BidMarket.Contract.Owner(&_BidMarket.CallOpts)
}

// PaymentToken is a free data retrieval call binding the contract method 0x3013ce29.
//
// Solidity: function paymentToken() view returns(address)
func (_BidMarket *BidMarketCaller) PaymentToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BidMarket.contract.Call(opts, &out, "paymentToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PaymentToken is a free data retrieval call binding the contract method 0x3013ce29.
//
// Solidity: function paymentToken() view returns(address)
func (_BidMarket *BidMarketSession) PaymentToken() (common.Address, error) {
	return _BidMarket.Contract.PaymentToken(&_BidMarket.CallOpts)
}

// PaymentToken is a free data retrieval call binding the contract method 0x3013ce29.
//
// Solidity: function paymentToken() view returns(address)
func (_BidMarket *BidMarketCallerSession) PaymentToken() (common.Address, error) {
	return _BidMarket.Contract.PaymentToken(&_BidMarket.CallOpts)
}

// PlatformFeePercentage is a free data retrieval call binding the contract method 0xcdd78cfc.
//
// Solidity: function platformFeePercentage() view returns(uint256)
func (_BidMarket *BidMarketCaller) PlatformFeePercentage(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BidMarket.contract.Call(opts, &out, "platformFeePercentage")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PlatformFeePercentage is a free data retrieval call binding the contract method 0xcdd78cfc.
//
// Solidity: function platformFeePercentage() view returns(uint256)
func (_BidMarket *BidMarketSession) PlatformFeePercentage() (*big.Int, error) {
	return _BidMarket.Contract.PlatformFeePercentage(&_BidMarket.CallOpts)
}

// PlatformFeePercentage is a free data retrieval call binding the contract method 0xcdd78cfc.
//
// Solidity: function platformFeePercentage() view returns(uint256)
func (_BidMarket *BidMarketCallerSession) PlatformFeePercentage() (*big.Int, error) {
	return _BidMarket.Contract.PlatformFeePercentage(&_BidMarket.CallOpts)
}

// PlatformWallet is a free data retrieval call binding the contract method 0xfa2af9da.
//
// Solidity: function platformWallet() view returns(address)
func (_BidMarket *BidMarketCaller) PlatformWallet(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BidMarket.contract.Call(opts, &out, "platformWallet")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PlatformWallet is a free data retrieval call binding the contract method 0xfa2af9da.
//
// Solidity: function platformWallet() view returns(address)
func (_BidMarket *BidMarketSession) PlatformWallet() (common.Address, error) {
	return _BidMarket.Contract.PlatformWallet(&_BidMarket.CallOpts)
}

// PlatformWallet is a free data retrieval call binding the contract method 0xfa2af9da.
//
// Solidity: function platformWallet() view returns(address)
func (_BidMarket *BidMarketCallerSession) PlatformWallet() (common.Address, error) {
	return _BidMarket.Contract.PlatformWallet(&_BidMarket.CallOpts)
}

// SubnetProviderContract is a free data retrieval call binding the contract method 0xcb43dedd.
//
// Solidity: function subnetProviderContract() view returns(address)
func (_BidMarket *BidMarketCaller) SubnetProviderContract(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BidMarket.contract.Call(opts, &out, "subnetProviderContract")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SubnetProviderContract is a free data retrieval call binding the contract method 0xcb43dedd.
//
// Solidity: function subnetProviderContract() view returns(address)
func (_BidMarket *BidMarketSession) SubnetProviderContract() (common.Address, error) {
	return _BidMarket.Contract.SubnetProviderContract(&_BidMarket.CallOpts)
}

// SubnetProviderContract is a free data retrieval call binding the contract method 0xcb43dedd.
//
// Solidity: function subnetProviderContract() view returns(address)
func (_BidMarket *BidMarketCallerSession) SubnetProviderContract() (common.Address, error) {
	return _BidMarket.Contract.SubnetProviderContract(&_BidMarket.CallOpts)
}

// TotalAccumulatedFees is a free data retrieval call binding the contract method 0x40a8546c.
//
// Solidity: function totalAccumulatedFees() view returns(uint256)
func (_BidMarket *BidMarketCaller) TotalAccumulatedFees(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BidMarket.contract.Call(opts, &out, "totalAccumulatedFees")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalAccumulatedFees is a free data retrieval call binding the contract method 0x40a8546c.
//
// Solidity: function totalAccumulatedFees() view returns(uint256)
func (_BidMarket *BidMarketSession) TotalAccumulatedFees() (*big.Int, error) {
	return _BidMarket.Contract.TotalAccumulatedFees(&_BidMarket.CallOpts)
}

// TotalAccumulatedFees is a free data retrieval call binding the contract method 0x40a8546c.
//
// Solidity: function totalAccumulatedFees() view returns(uint256)
func (_BidMarket *BidMarketCallerSession) TotalAccumulatedFees() (*big.Int, error) {
	return _BidMarket.Contract.TotalAccumulatedFees(&_BidMarket.CallOpts)
}

// AcceptBid is a paid mutator transaction binding the contract method 0x02e9d5e4.
//
// Solidity: function acceptBid(uint256 orderId, uint256 bidIndex) returns()
func (_BidMarket *BidMarketTransactor) AcceptBid(opts *bind.TransactOpts, orderId *big.Int, bidIndex *big.Int) (*types.Transaction, error) {
	return _BidMarket.contract.Transact(opts, "acceptBid", orderId, bidIndex)
}

// AcceptBid is a paid mutator transaction binding the contract method 0x02e9d5e4.
//
// Solidity: function acceptBid(uint256 orderId, uint256 bidIndex) returns()
func (_BidMarket *BidMarketSession) AcceptBid(orderId *big.Int, bidIndex *big.Int) (*types.Transaction, error) {
	return _BidMarket.Contract.AcceptBid(&_BidMarket.TransactOpts, orderId, bidIndex)
}

// AcceptBid is a paid mutator transaction binding the contract method 0x02e9d5e4.
//
// Solidity: function acceptBid(uint256 orderId, uint256 bidIndex) returns()
func (_BidMarket *BidMarketTransactorSession) AcceptBid(orderId *big.Int, bidIndex *big.Int) (*types.Transaction, error) {
	return _BidMarket.Contract.AcceptBid(&_BidMarket.TransactOpts, orderId, bidIndex)
}

// CancelBid is a paid mutator transaction binding the contract method 0x4b393605.
//
// Solidity: function cancelBid(uint256 orderId, uint256 bidIndex) returns()
func (_BidMarket *BidMarketTransactor) CancelBid(opts *bind.TransactOpts, orderId *big.Int, bidIndex *big.Int) (*types.Transaction, error) {
	return _BidMarket.contract.Transact(opts, "cancelBid", orderId, bidIndex)
}

// CancelBid is a paid mutator transaction binding the contract method 0x4b393605.
//
// Solidity: function cancelBid(uint256 orderId, uint256 bidIndex) returns()
func (_BidMarket *BidMarketSession) CancelBid(orderId *big.Int, bidIndex *big.Int) (*types.Transaction, error) {
	return _BidMarket.Contract.CancelBid(&_BidMarket.TransactOpts, orderId, bidIndex)
}

// CancelBid is a paid mutator transaction binding the contract method 0x4b393605.
//
// Solidity: function cancelBid(uint256 orderId, uint256 bidIndex) returns()
func (_BidMarket *BidMarketTransactorSession) CancelBid(orderId *big.Int, bidIndex *big.Int) (*types.Transaction, error) {
	return _BidMarket.Contract.CancelBid(&_BidMarket.TransactOpts, orderId, bidIndex)
}

// CancelOrder is a paid mutator transaction binding the contract method 0x514fcac7.
//
// Solidity: function cancelOrder(uint256 orderId) returns()
func (_BidMarket *BidMarketTransactor) CancelOrder(opts *bind.TransactOpts, orderId *big.Int) (*types.Transaction, error) {
	return _BidMarket.contract.Transact(opts, "cancelOrder", orderId)
}

// CancelOrder is a paid mutator transaction binding the contract method 0x514fcac7.
//
// Solidity: function cancelOrder(uint256 orderId) returns()
func (_BidMarket *BidMarketSession) CancelOrder(orderId *big.Int) (*types.Transaction, error) {
	return _BidMarket.Contract.CancelOrder(&_BidMarket.TransactOpts, orderId)
}

// CancelOrder is a paid mutator transaction binding the contract method 0x514fcac7.
//
// Solidity: function cancelOrder(uint256 orderId) returns()
func (_BidMarket *BidMarketTransactorSession) CancelOrder(orderId *big.Int) (*types.Transaction, error) {
	return _BidMarket.Contract.CancelOrder(&_BidMarket.TransactOpts, orderId)
}

// ClaimPayment is a paid mutator transaction binding the contract method 0xc63fdcc7.
//
// Solidity: function claimPayment(uint256 orderId) returns()
func (_BidMarket *BidMarketTransactor) ClaimPayment(opts *bind.TransactOpts, orderId *big.Int) (*types.Transaction, error) {
	return _BidMarket.contract.Transact(opts, "claimPayment", orderId)
}

// ClaimPayment is a paid mutator transaction binding the contract method 0xc63fdcc7.
//
// Solidity: function claimPayment(uint256 orderId) returns()
func (_BidMarket *BidMarketSession) ClaimPayment(orderId *big.Int) (*types.Transaction, error) {
	return _BidMarket.Contract.ClaimPayment(&_BidMarket.TransactOpts, orderId)
}

// ClaimPayment is a paid mutator transaction binding the contract method 0xc63fdcc7.
//
// Solidity: function claimPayment(uint256 orderId) returns()
func (_BidMarket *BidMarketTransactorSession) ClaimPayment(orderId *big.Int) (*types.Transaction, error) {
	return _BidMarket.Contract.ClaimPayment(&_BidMarket.TransactOpts, orderId)
}

// CloseOrder is a paid mutator transaction binding the contract method 0xbca5a3a8.
//
// Solidity: function closeOrder(uint256 orderId, string reason) returns()
func (_BidMarket *BidMarketTransactor) CloseOrder(opts *bind.TransactOpts, orderId *big.Int, reason string) (*types.Transaction, error) {
	return _BidMarket.contract.Transact(opts, "closeOrder", orderId, reason)
}

// CloseOrder is a paid mutator transaction binding the contract method 0xbca5a3a8.
//
// Solidity: function closeOrder(uint256 orderId, string reason) returns()
func (_BidMarket *BidMarketSession) CloseOrder(orderId *big.Int, reason string) (*types.Transaction, error) {
	return _BidMarket.Contract.CloseOrder(&_BidMarket.TransactOpts, orderId, reason)
}

// CloseOrder is a paid mutator transaction binding the contract method 0xbca5a3a8.
//
// Solidity: function closeOrder(uint256 orderId, string reason) returns()
func (_BidMarket *BidMarketTransactorSession) CloseOrder(orderId *big.Int, reason string) (*types.Transaction, error) {
	return _BidMarket.Contract.CloseOrder(&_BidMarket.TransactOpts, orderId, reason)
}

// CreateOrder is a paid mutator transaction binding the contract method 0x53cb7dbc.
//
// Solidity: function createOrder(uint256 machineType, uint256 duration, uint256 minBidPrice, uint256 maxBidPrice, uint256 region, uint256 cpuCores, uint256 gpuCores, uint256 gpuMemory, uint256 memoryMB, uint256 diskGB, uint256 uploadMbps, uint256 downloadMbps, string specs) returns(uint256)
func (_BidMarket *BidMarketTransactor) CreateOrder(opts *bind.TransactOpts, machineType *big.Int, duration *big.Int, minBidPrice *big.Int, maxBidPrice *big.Int, region *big.Int, cpuCores *big.Int, gpuCores *big.Int, gpuMemory *big.Int, memoryMB *big.Int, diskGB *big.Int, uploadMbps *big.Int, downloadMbps *big.Int, specs string) (*types.Transaction, error) {
	return _BidMarket.contract.Transact(opts, "createOrder", machineType, duration, minBidPrice, maxBidPrice, region, cpuCores, gpuCores, gpuMemory, memoryMB, diskGB, uploadMbps, downloadMbps, specs)
}

// CreateOrder is a paid mutator transaction binding the contract method 0x53cb7dbc.
//
// Solidity: function createOrder(uint256 machineType, uint256 duration, uint256 minBidPrice, uint256 maxBidPrice, uint256 region, uint256 cpuCores, uint256 gpuCores, uint256 gpuMemory, uint256 memoryMB, uint256 diskGB, uint256 uploadMbps, uint256 downloadMbps, string specs) returns(uint256)
func (_BidMarket *BidMarketSession) CreateOrder(machineType *big.Int, duration *big.Int, minBidPrice *big.Int, maxBidPrice *big.Int, region *big.Int, cpuCores *big.Int, gpuCores *big.Int, gpuMemory *big.Int, memoryMB *big.Int, diskGB *big.Int, uploadMbps *big.Int, downloadMbps *big.Int, specs string) (*types.Transaction, error) {
	return _BidMarket.Contract.CreateOrder(&_BidMarket.TransactOpts, machineType, duration, minBidPrice, maxBidPrice, region, cpuCores, gpuCores, gpuMemory, memoryMB, diskGB, uploadMbps, downloadMbps, specs)
}

// CreateOrder is a paid mutator transaction binding the contract method 0x53cb7dbc.
//
// Solidity: function createOrder(uint256 machineType, uint256 duration, uint256 minBidPrice, uint256 maxBidPrice, uint256 region, uint256 cpuCores, uint256 gpuCores, uint256 gpuMemory, uint256 memoryMB, uint256 diskGB, uint256 uploadMbps, uint256 downloadMbps, string specs) returns(uint256)
func (_BidMarket *BidMarketTransactorSession) CreateOrder(machineType *big.Int, duration *big.Int, minBidPrice *big.Int, maxBidPrice *big.Int, region *big.Int, cpuCores *big.Int, gpuCores *big.Int, gpuMemory *big.Int, memoryMB *big.Int, diskGB *big.Int, uploadMbps *big.Int, downloadMbps *big.Int, specs string) (*types.Transaction, error) {
	return _BidMarket.Contract.CreateOrder(&_BidMarket.TransactOpts, machineType, duration, minBidPrice, maxBidPrice, region, cpuCores, gpuCores, gpuMemory, memoryMB, diskGB, uploadMbps, downloadMbps, specs)
}

// Extend is a paid mutator transaction binding the contract method 0x9714378c.
//
// Solidity: function extend(uint256 orderId) returns()
func (_BidMarket *BidMarketTransactor) Extend(opts *bind.TransactOpts, orderId *big.Int) (*types.Transaction, error) {
	return _BidMarket.contract.Transact(opts, "extend", orderId)
}

// Extend is a paid mutator transaction binding the contract method 0x9714378c.
//
// Solidity: function extend(uint256 orderId) returns()
func (_BidMarket *BidMarketSession) Extend(orderId *big.Int) (*types.Transaction, error) {
	return _BidMarket.Contract.Extend(&_BidMarket.TransactOpts, orderId)
}

// Extend is a paid mutator transaction binding the contract method 0x9714378c.
//
// Solidity: function extend(uint256 orderId) returns()
func (_BidMarket *BidMarketTransactorSession) Extend(orderId *big.Int) (*types.Transaction, error) {
	return _BidMarket.Contract.Extend(&_BidMarket.TransactOpts, orderId)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address owner, address _paymentToken, address _subnetProviderContract) returns()
func (_BidMarket *BidMarketTransactor) Initialize(opts *bind.TransactOpts, owner common.Address, _paymentToken common.Address, _subnetProviderContract common.Address) (*types.Transaction, error) {
	return _BidMarket.contract.Transact(opts, "initialize", owner, _paymentToken, _subnetProviderContract)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address owner, address _paymentToken, address _subnetProviderContract) returns()
func (_BidMarket *BidMarketSession) Initialize(owner common.Address, _paymentToken common.Address, _subnetProviderContract common.Address) (*types.Transaction, error) {
	return _BidMarket.Contract.Initialize(&_BidMarket.TransactOpts, owner, _paymentToken, _subnetProviderContract)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address owner, address _paymentToken, address _subnetProviderContract) returns()
func (_BidMarket *BidMarketTransactorSession) Initialize(owner common.Address, _paymentToken common.Address, _subnetProviderContract common.Address) (*types.Transaction, error) {
	return _BidMarket.Contract.Initialize(&_BidMarket.TransactOpts, owner, _paymentToken, _subnetProviderContract)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BidMarket *BidMarketTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BidMarket.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BidMarket *BidMarketSession) RenounceOwnership() (*types.Transaction, error) {
	return _BidMarket.Contract.RenounceOwnership(&_BidMarket.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BidMarket *BidMarketTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _BidMarket.Contract.RenounceOwnership(&_BidMarket.TransactOpts)
}

// SetBidTimeLimit is a paid mutator transaction binding the contract method 0x722af14f.
//
// Solidity: function setBidTimeLimit(uint256 newBidTimeLimit) returns()
func (_BidMarket *BidMarketTransactor) SetBidTimeLimit(opts *bind.TransactOpts, newBidTimeLimit *big.Int) (*types.Transaction, error) {
	return _BidMarket.contract.Transact(opts, "setBidTimeLimit", newBidTimeLimit)
}

// SetBidTimeLimit is a paid mutator transaction binding the contract method 0x722af14f.
//
// Solidity: function setBidTimeLimit(uint256 newBidTimeLimit) returns()
func (_BidMarket *BidMarketSession) SetBidTimeLimit(newBidTimeLimit *big.Int) (*types.Transaction, error) {
	return _BidMarket.Contract.SetBidTimeLimit(&_BidMarket.TransactOpts, newBidTimeLimit)
}

// SetBidTimeLimit is a paid mutator transaction binding the contract method 0x722af14f.
//
// Solidity: function setBidTimeLimit(uint256 newBidTimeLimit) returns()
func (_BidMarket *BidMarketTransactorSession) SetBidTimeLimit(newBidTimeLimit *big.Int) (*types.Transaction, error) {
	return _BidMarket.Contract.SetBidTimeLimit(&_BidMarket.TransactOpts, newBidTimeLimit)
}

// SetPaymentConfig is a paid mutator transaction binding the contract method 0x62e65603.
//
// Solidity: function setPaymentConfig(address _paymentToken) returns()
func (_BidMarket *BidMarketTransactor) SetPaymentConfig(opts *bind.TransactOpts, _paymentToken common.Address) (*types.Transaction, error) {
	return _BidMarket.contract.Transact(opts, "setPaymentConfig", _paymentToken)
}

// SetPaymentConfig is a paid mutator transaction binding the contract method 0x62e65603.
//
// Solidity: function setPaymentConfig(address _paymentToken) returns()
func (_BidMarket *BidMarketSession) SetPaymentConfig(_paymentToken common.Address) (*types.Transaction, error) {
	return _BidMarket.Contract.SetPaymentConfig(&_BidMarket.TransactOpts, _paymentToken)
}

// SetPaymentConfig is a paid mutator transaction binding the contract method 0x62e65603.
//
// Solidity: function setPaymentConfig(address _paymentToken) returns()
func (_BidMarket *BidMarketTransactorSession) SetPaymentConfig(_paymentToken common.Address) (*types.Transaction, error) {
	return _BidMarket.Contract.SetPaymentConfig(&_BidMarket.TransactOpts, _paymentToken)
}

// SetPlatformFee is a paid mutator transaction binding the contract method 0x12e8e2c3.
//
// Solidity: function setPlatformFee(uint256 newFeePercentage) returns()
func (_BidMarket *BidMarketTransactor) SetPlatformFee(opts *bind.TransactOpts, newFeePercentage *big.Int) (*types.Transaction, error) {
	return _BidMarket.contract.Transact(opts, "setPlatformFee", newFeePercentage)
}

// SetPlatformFee is a paid mutator transaction binding the contract method 0x12e8e2c3.
//
// Solidity: function setPlatformFee(uint256 newFeePercentage) returns()
func (_BidMarket *BidMarketSession) SetPlatformFee(newFeePercentage *big.Int) (*types.Transaction, error) {
	return _BidMarket.Contract.SetPlatformFee(&_BidMarket.TransactOpts, newFeePercentage)
}

// SetPlatformFee is a paid mutator transaction binding the contract method 0x12e8e2c3.
//
// Solidity: function setPlatformFee(uint256 newFeePercentage) returns()
func (_BidMarket *BidMarketTransactorSession) SetPlatformFee(newFeePercentage *big.Int) (*types.Transaction, error) {
	return _BidMarket.Contract.SetPlatformFee(&_BidMarket.TransactOpts, newFeePercentage)
}

// SetPlatformWallet is a paid mutator transaction binding the contract method 0x8831e9cf.
//
// Solidity: function setPlatformWallet(address newWallet) returns()
func (_BidMarket *BidMarketTransactor) SetPlatformWallet(opts *bind.TransactOpts, newWallet common.Address) (*types.Transaction, error) {
	return _BidMarket.contract.Transact(opts, "setPlatformWallet", newWallet)
}

// SetPlatformWallet is a paid mutator transaction binding the contract method 0x8831e9cf.
//
// Solidity: function setPlatformWallet(address newWallet) returns()
func (_BidMarket *BidMarketSession) SetPlatformWallet(newWallet common.Address) (*types.Transaction, error) {
	return _BidMarket.Contract.SetPlatformWallet(&_BidMarket.TransactOpts, newWallet)
}

// SetPlatformWallet is a paid mutator transaction binding the contract method 0x8831e9cf.
//
// Solidity: function setPlatformWallet(address newWallet) returns()
func (_BidMarket *BidMarketTransactorSession) SetPlatformWallet(newWallet common.Address) (*types.Transaction, error) {
	return _BidMarket.Contract.SetPlatformWallet(&_BidMarket.TransactOpts, newWallet)
}

// SetSubnetProviderContract is a paid mutator transaction binding the contract method 0x334d0236.
//
// Solidity: function setSubnetProviderContract(address _subnetProviderContract) returns()
func (_BidMarket *BidMarketTransactor) SetSubnetProviderContract(opts *bind.TransactOpts, _subnetProviderContract common.Address) (*types.Transaction, error) {
	return _BidMarket.contract.Transact(opts, "setSubnetProviderContract", _subnetProviderContract)
}

// SetSubnetProviderContract is a paid mutator transaction binding the contract method 0x334d0236.
//
// Solidity: function setSubnetProviderContract(address _subnetProviderContract) returns()
func (_BidMarket *BidMarketSession) SetSubnetProviderContract(_subnetProviderContract common.Address) (*types.Transaction, error) {
	return _BidMarket.Contract.SetSubnetProviderContract(&_BidMarket.TransactOpts, _subnetProviderContract)
}

// SetSubnetProviderContract is a paid mutator transaction binding the contract method 0x334d0236.
//
// Solidity: function setSubnetProviderContract(address _subnetProviderContract) returns()
func (_BidMarket *BidMarketTransactorSession) SetSubnetProviderContract(_subnetProviderContract common.Address) (*types.Transaction, error) {
	return _BidMarket.Contract.SetSubnetProviderContract(&_BidMarket.TransactOpts, _subnetProviderContract)
}

// SubmitBid is a paid mutator transaction binding the contract method 0xd793139e.
//
// Solidity: function submitBid(uint256 orderId, uint256 pricePerSecond, uint256 providerId, uint256 machineId) returns()
func (_BidMarket *BidMarketTransactor) SubmitBid(opts *bind.TransactOpts, orderId *big.Int, pricePerSecond *big.Int, providerId *big.Int, machineId *big.Int) (*types.Transaction, error) {
	return _BidMarket.contract.Transact(opts, "submitBid", orderId, pricePerSecond, providerId, machineId)
}

// SubmitBid is a paid mutator transaction binding the contract method 0xd793139e.
//
// Solidity: function submitBid(uint256 orderId, uint256 pricePerSecond, uint256 providerId, uint256 machineId) returns()
func (_BidMarket *BidMarketSession) SubmitBid(orderId *big.Int, pricePerSecond *big.Int, providerId *big.Int, machineId *big.Int) (*types.Transaction, error) {
	return _BidMarket.Contract.SubmitBid(&_BidMarket.TransactOpts, orderId, pricePerSecond, providerId, machineId)
}

// SubmitBid is a paid mutator transaction binding the contract method 0xd793139e.
//
// Solidity: function submitBid(uint256 orderId, uint256 pricePerSecond, uint256 providerId, uint256 machineId) returns()
func (_BidMarket *BidMarketTransactorSession) SubmitBid(orderId *big.Int, pricePerSecond *big.Int, providerId *big.Int, machineId *big.Int) (*types.Transaction, error) {
	return _BidMarket.Contract.SubmitBid(&_BidMarket.TransactOpts, orderId, pricePerSecond, providerId, machineId)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BidMarket *BidMarketTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _BidMarket.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BidMarket *BidMarketSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BidMarket.Contract.TransferOwnership(&_BidMarket.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BidMarket *BidMarketTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BidMarket.Contract.TransferOwnership(&_BidMarket.TransactOpts, newOwner)
}

// WithdrawFees is a paid mutator transaction binding the contract method 0x5e318e07.
//
// Solidity: function withdrawFees(uint256 amount) returns()
func (_BidMarket *BidMarketTransactor) WithdrawFees(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _BidMarket.contract.Transact(opts, "withdrawFees", amount)
}

// WithdrawFees is a paid mutator transaction binding the contract method 0x5e318e07.
//
// Solidity: function withdrawFees(uint256 amount) returns()
func (_BidMarket *BidMarketSession) WithdrawFees(amount *big.Int) (*types.Transaction, error) {
	return _BidMarket.Contract.WithdrawFees(&_BidMarket.TransactOpts, amount)
}

// WithdrawFees is a paid mutator transaction binding the contract method 0x5e318e07.
//
// Solidity: function withdrawFees(uint256 amount) returns()
func (_BidMarket *BidMarketTransactorSession) WithdrawFees(amount *big.Int) (*types.Transaction, error) {
	return _BidMarket.Contract.WithdrawFees(&_BidMarket.TransactOpts, amount)
}

// BidMarketBidAcceptedIterator is returned from FilterBidAccepted and is used to iterate over the raw logs and unpacked data for BidAccepted events raised by the BidMarket contract.
type BidMarketBidAcceptedIterator struct {
	Event *BidMarketBidAccepted // Event containing the contract specifics and raw log

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
func (it *BidMarketBidAcceptedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BidMarketBidAccepted)
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
		it.Event = new(BidMarketBidAccepted)
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
func (it *BidMarketBidAcceptedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BidMarketBidAcceptedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BidMarketBidAccepted represents a BidAccepted event raised by the BidMarket contract.
type BidMarketBidAccepted struct {
	OrderId  *big.Int
	Provider common.Address
	Price    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterBidAccepted is a free log retrieval operation binding the contract event 0x91a721693afff25f233ac308edbd2fa15ad7d4dfacdd164dd4e8c57ab72f4ae9.
//
// Solidity: event BidAccepted(uint256 indexed orderId, address indexed provider, uint256 price)
func (_BidMarket *BidMarketFilterer) FilterBidAccepted(opts *bind.FilterOpts, orderId []*big.Int, provider []common.Address) (*BidMarketBidAcceptedIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BidMarket.contract.FilterLogs(opts, "BidAccepted", orderIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &BidMarketBidAcceptedIterator{contract: _BidMarket.contract, event: "BidAccepted", logs: logs, sub: sub}, nil
}

// WatchBidAccepted is a free log subscription operation binding the contract event 0x91a721693afff25f233ac308edbd2fa15ad7d4dfacdd164dd4e8c57ab72f4ae9.
//
// Solidity: event BidAccepted(uint256 indexed orderId, address indexed provider, uint256 price)
func (_BidMarket *BidMarketFilterer) WatchBidAccepted(opts *bind.WatchOpts, sink chan<- *BidMarketBidAccepted, orderId []*big.Int, provider []common.Address) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BidMarket.contract.WatchLogs(opts, "BidAccepted", orderIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BidMarketBidAccepted)
				if err := _BidMarket.contract.UnpackLog(event, "BidAccepted", log); err != nil {
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

// ParseBidAccepted is a log parse operation binding the contract event 0x91a721693afff25f233ac308edbd2fa15ad7d4dfacdd164dd4e8c57ab72f4ae9.
//
// Solidity: event BidAccepted(uint256 indexed orderId, address indexed provider, uint256 price)
func (_BidMarket *BidMarketFilterer) ParseBidAccepted(log types.Log) (*BidMarketBidAccepted, error) {
	event := new(BidMarketBidAccepted)
	if err := _BidMarket.contract.UnpackLog(event, "BidAccepted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BidMarketBidCancelledIterator is returned from FilterBidCancelled and is used to iterate over the raw logs and unpacked data for BidCancelled events raised by the BidMarket contract.
type BidMarketBidCancelledIterator struct {
	Event *BidMarketBidCancelled // Event containing the contract specifics and raw log

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
func (it *BidMarketBidCancelledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BidMarketBidCancelled)
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
		it.Event = new(BidMarketBidCancelled)
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
func (it *BidMarketBidCancelledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BidMarketBidCancelledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BidMarketBidCancelled represents a BidCancelled event raised by the BidMarket contract.
type BidMarketBidCancelled struct {
	OrderId  *big.Int
	Provider common.Address
	BidIndex *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterBidCancelled is a free log retrieval operation binding the contract event 0xaf2f02012ea1953fe9293dd4a8bb5194c96792324f9385297c991c2cfeade3d9.
//
// Solidity: event BidCancelled(uint256 indexed orderId, address indexed provider, uint256 bidIndex)
func (_BidMarket *BidMarketFilterer) FilterBidCancelled(opts *bind.FilterOpts, orderId []*big.Int, provider []common.Address) (*BidMarketBidCancelledIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BidMarket.contract.FilterLogs(opts, "BidCancelled", orderIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &BidMarketBidCancelledIterator{contract: _BidMarket.contract, event: "BidCancelled", logs: logs, sub: sub}, nil
}

// WatchBidCancelled is a free log subscription operation binding the contract event 0xaf2f02012ea1953fe9293dd4a8bb5194c96792324f9385297c991c2cfeade3d9.
//
// Solidity: event BidCancelled(uint256 indexed orderId, address indexed provider, uint256 bidIndex)
func (_BidMarket *BidMarketFilterer) WatchBidCancelled(opts *bind.WatchOpts, sink chan<- *BidMarketBidCancelled, orderId []*big.Int, provider []common.Address) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BidMarket.contract.WatchLogs(opts, "BidCancelled", orderIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BidMarketBidCancelled)
				if err := _BidMarket.contract.UnpackLog(event, "BidCancelled", log); err != nil {
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

// ParseBidCancelled is a log parse operation binding the contract event 0xaf2f02012ea1953fe9293dd4a8bb5194c96792324f9385297c991c2cfeade3d9.
//
// Solidity: event BidCancelled(uint256 indexed orderId, address indexed provider, uint256 bidIndex)
func (_BidMarket *BidMarketFilterer) ParseBidCancelled(log types.Log) (*BidMarketBidCancelled, error) {
	event := new(BidMarketBidCancelled)
	if err := _BidMarket.contract.UnpackLog(event, "BidCancelled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BidMarketBidSubmittedIterator is returned from FilterBidSubmitted and is used to iterate over the raw logs and unpacked data for BidSubmitted events raised by the BidMarket contract.
type BidMarketBidSubmittedIterator struct {
	Event *BidMarketBidSubmitted // Event containing the contract specifics and raw log

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
func (it *BidMarketBidSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BidMarketBidSubmitted)
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
		it.Event = new(BidMarketBidSubmitted)
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
func (it *BidMarketBidSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BidMarketBidSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BidMarketBidSubmitted represents a BidSubmitted event raised by the BidMarket contract.
type BidMarketBidSubmitted struct {
	OrderId    *big.Int
	Provider   common.Address
	Price      *big.Int
	ProviderId *big.Int
	MachineId  *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterBidSubmitted is a free log retrieval operation binding the contract event 0x1b3d307c757cc471e124508de8fb270e98d9c23f6220af3df3e152cbefc05045.
//
// Solidity: event BidSubmitted(uint256 indexed orderId, address indexed provider, uint256 price, uint256 providerId, uint256 machineId)
func (_BidMarket *BidMarketFilterer) FilterBidSubmitted(opts *bind.FilterOpts, orderId []*big.Int, provider []common.Address) (*BidMarketBidSubmittedIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BidMarket.contract.FilterLogs(opts, "BidSubmitted", orderIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &BidMarketBidSubmittedIterator{contract: _BidMarket.contract, event: "BidSubmitted", logs: logs, sub: sub}, nil
}

// WatchBidSubmitted is a free log subscription operation binding the contract event 0x1b3d307c757cc471e124508de8fb270e98d9c23f6220af3df3e152cbefc05045.
//
// Solidity: event BidSubmitted(uint256 indexed orderId, address indexed provider, uint256 price, uint256 providerId, uint256 machineId)
func (_BidMarket *BidMarketFilterer) WatchBidSubmitted(opts *bind.WatchOpts, sink chan<- *BidMarketBidSubmitted, orderId []*big.Int, provider []common.Address) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BidMarket.contract.WatchLogs(opts, "BidSubmitted", orderIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BidMarketBidSubmitted)
				if err := _BidMarket.contract.UnpackLog(event, "BidSubmitted", log); err != nil {
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

// ParseBidSubmitted is a log parse operation binding the contract event 0x1b3d307c757cc471e124508de8fb270e98d9c23f6220af3df3e152cbefc05045.
//
// Solidity: event BidSubmitted(uint256 indexed orderId, address indexed provider, uint256 price, uint256 providerId, uint256 machineId)
func (_BidMarket *BidMarketFilterer) ParseBidSubmitted(log types.Log) (*BidMarketBidSubmitted, error) {
	event := new(BidMarketBidSubmitted)
	if err := _BidMarket.contract.UnpackLog(event, "BidSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BidMarketBidTimeExpiredIterator is returned from FilterBidTimeExpired and is used to iterate over the raw logs and unpacked data for BidTimeExpired events raised by the BidMarket contract.
type BidMarketBidTimeExpiredIterator struct {
	Event *BidMarketBidTimeExpired // Event containing the contract specifics and raw log

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
func (it *BidMarketBidTimeExpiredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BidMarketBidTimeExpired)
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
		it.Event = new(BidMarketBidTimeExpired)
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
func (it *BidMarketBidTimeExpiredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BidMarketBidTimeExpiredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BidMarketBidTimeExpired represents a BidTimeExpired event raised by the BidMarket contract.
type BidMarketBidTimeExpired struct {
	OrderId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterBidTimeExpired is a free log retrieval operation binding the contract event 0x10eced0c84b15a7afa2d00aa09f0d7cc62ca90d4cf6a5897034209cb6904aaa0.
//
// Solidity: event BidTimeExpired(uint256 indexed orderId)
func (_BidMarket *BidMarketFilterer) FilterBidTimeExpired(opts *bind.FilterOpts, orderId []*big.Int) (*BidMarketBidTimeExpiredIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _BidMarket.contract.FilterLogs(opts, "BidTimeExpired", orderIdRule)
	if err != nil {
		return nil, err
	}
	return &BidMarketBidTimeExpiredIterator{contract: _BidMarket.contract, event: "BidTimeExpired", logs: logs, sub: sub}, nil
}

// WatchBidTimeExpired is a free log subscription operation binding the contract event 0x10eced0c84b15a7afa2d00aa09f0d7cc62ca90d4cf6a5897034209cb6904aaa0.
//
// Solidity: event BidTimeExpired(uint256 indexed orderId)
func (_BidMarket *BidMarketFilterer) WatchBidTimeExpired(opts *bind.WatchOpts, sink chan<- *BidMarketBidTimeExpired, orderId []*big.Int) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _BidMarket.contract.WatchLogs(opts, "BidTimeExpired", orderIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BidMarketBidTimeExpired)
				if err := _BidMarket.contract.UnpackLog(event, "BidTimeExpired", log); err != nil {
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

// ParseBidTimeExpired is a log parse operation binding the contract event 0x10eced0c84b15a7afa2d00aa09f0d7cc62ca90d4cf6a5897034209cb6904aaa0.
//
// Solidity: event BidTimeExpired(uint256 indexed orderId)
func (_BidMarket *BidMarketFilterer) ParseBidTimeExpired(log types.Log) (*BidMarketBidTimeExpired, error) {
	event := new(BidMarketBidTimeExpired)
	if err := _BidMarket.contract.UnpackLog(event, "BidTimeExpired", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BidMarketBidTimeLimitUpdatedIterator is returned from FilterBidTimeLimitUpdated and is used to iterate over the raw logs and unpacked data for BidTimeLimitUpdated events raised by the BidMarket contract.
type BidMarketBidTimeLimitUpdatedIterator struct {
	Event *BidMarketBidTimeLimitUpdated // Event containing the contract specifics and raw log

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
func (it *BidMarketBidTimeLimitUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BidMarketBidTimeLimitUpdated)
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
		it.Event = new(BidMarketBidTimeLimitUpdated)
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
func (it *BidMarketBidTimeLimitUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BidMarketBidTimeLimitUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BidMarketBidTimeLimitUpdated represents a BidTimeLimitUpdated event raised by the BidMarket contract.
type BidMarketBidTimeLimitUpdated struct {
	OldLimit *big.Int
	NewLimit *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterBidTimeLimitUpdated is a free log retrieval operation binding the contract event 0x9639a82e292ab5f28f7cb06ef7bd2259222c9468822650cf1798e015df671b3e.
//
// Solidity: event BidTimeLimitUpdated(uint256 oldLimit, uint256 newLimit)
func (_BidMarket *BidMarketFilterer) FilterBidTimeLimitUpdated(opts *bind.FilterOpts) (*BidMarketBidTimeLimitUpdatedIterator, error) {

	logs, sub, err := _BidMarket.contract.FilterLogs(opts, "BidTimeLimitUpdated")
	if err != nil {
		return nil, err
	}
	return &BidMarketBidTimeLimitUpdatedIterator{contract: _BidMarket.contract, event: "BidTimeLimitUpdated", logs: logs, sub: sub}, nil
}

// WatchBidTimeLimitUpdated is a free log subscription operation binding the contract event 0x9639a82e292ab5f28f7cb06ef7bd2259222c9468822650cf1798e015df671b3e.
//
// Solidity: event BidTimeLimitUpdated(uint256 oldLimit, uint256 newLimit)
func (_BidMarket *BidMarketFilterer) WatchBidTimeLimitUpdated(opts *bind.WatchOpts, sink chan<- *BidMarketBidTimeLimitUpdated) (event.Subscription, error) {

	logs, sub, err := _BidMarket.contract.WatchLogs(opts, "BidTimeLimitUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BidMarketBidTimeLimitUpdated)
				if err := _BidMarket.contract.UnpackLog(event, "BidTimeLimitUpdated", log); err != nil {
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

// ParseBidTimeLimitUpdated is a log parse operation binding the contract event 0x9639a82e292ab5f28f7cb06ef7bd2259222c9468822650cf1798e015df671b3e.
//
// Solidity: event BidTimeLimitUpdated(uint256 oldLimit, uint256 newLimit)
func (_BidMarket *BidMarketFilterer) ParseBidTimeLimitUpdated(log types.Log) (*BidMarketBidTimeLimitUpdated, error) {
	event := new(BidMarketBidTimeLimitUpdated)
	if err := _BidMarket.contract.UnpackLog(event, "BidTimeLimitUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BidMarketFeesWithdrawnIterator is returned from FilterFeesWithdrawn and is used to iterate over the raw logs and unpacked data for FeesWithdrawn events raised by the BidMarket contract.
type BidMarketFeesWithdrawnIterator struct {
	Event *BidMarketFeesWithdrawn // Event containing the contract specifics and raw log

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
func (it *BidMarketFeesWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BidMarketFeesWithdrawn)
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
		it.Event = new(BidMarketFeesWithdrawn)
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
func (it *BidMarketFeesWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BidMarketFeesWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BidMarketFeesWithdrawn represents a FeesWithdrawn event raised by the BidMarket contract.
type BidMarketFeesWithdrawn struct {
	Wallet common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterFeesWithdrawn is a free log retrieval operation binding the contract event 0xc0819c13be868895eb93e40eaceb96de976442fa1d404e5c55f14bb65a8c489a.
//
// Solidity: event FeesWithdrawn(address wallet, uint256 amount)
func (_BidMarket *BidMarketFilterer) FilterFeesWithdrawn(opts *bind.FilterOpts) (*BidMarketFeesWithdrawnIterator, error) {

	logs, sub, err := _BidMarket.contract.FilterLogs(opts, "FeesWithdrawn")
	if err != nil {
		return nil, err
	}
	return &BidMarketFeesWithdrawnIterator{contract: _BidMarket.contract, event: "FeesWithdrawn", logs: logs, sub: sub}, nil
}

// WatchFeesWithdrawn is a free log subscription operation binding the contract event 0xc0819c13be868895eb93e40eaceb96de976442fa1d404e5c55f14bb65a8c489a.
//
// Solidity: event FeesWithdrawn(address wallet, uint256 amount)
func (_BidMarket *BidMarketFilterer) WatchFeesWithdrawn(opts *bind.WatchOpts, sink chan<- *BidMarketFeesWithdrawn) (event.Subscription, error) {

	logs, sub, err := _BidMarket.contract.WatchLogs(opts, "FeesWithdrawn")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BidMarketFeesWithdrawn)
				if err := _BidMarket.contract.UnpackLog(event, "FeesWithdrawn", log); err != nil {
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

// ParseFeesWithdrawn is a log parse operation binding the contract event 0xc0819c13be868895eb93e40eaceb96de976442fa1d404e5c55f14bb65a8c489a.
//
// Solidity: event FeesWithdrawn(address wallet, uint256 amount)
func (_BidMarket *BidMarketFilterer) ParseFeesWithdrawn(log types.Log) (*BidMarketFeesWithdrawn, error) {
	event := new(BidMarketFeesWithdrawn)
	if err := _BidMarket.contract.UnpackLog(event, "FeesWithdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BidMarketInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the BidMarket contract.
type BidMarketInitializedIterator struct {
	Event *BidMarketInitialized // Event containing the contract specifics and raw log

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
func (it *BidMarketInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BidMarketInitialized)
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
		it.Event = new(BidMarketInitialized)
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
func (it *BidMarketInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BidMarketInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BidMarketInitialized represents a Initialized event raised by the BidMarket contract.
type BidMarketInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_BidMarket *BidMarketFilterer) FilterInitialized(opts *bind.FilterOpts) (*BidMarketInitializedIterator, error) {

	logs, sub, err := _BidMarket.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &BidMarketInitializedIterator{contract: _BidMarket.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_BidMarket *BidMarketFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *BidMarketInitialized) (event.Subscription, error) {

	logs, sub, err := _BidMarket.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BidMarketInitialized)
				if err := _BidMarket.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_BidMarket *BidMarketFilterer) ParseInitialized(log types.Log) (*BidMarketInitialized, error) {
	event := new(BidMarketInitialized)
	if err := _BidMarket.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BidMarketOrderCancelledIterator is returned from FilterOrderCancelled and is used to iterate over the raw logs and unpacked data for OrderCancelled events raised by the BidMarket contract.
type BidMarketOrderCancelledIterator struct {
	Event *BidMarketOrderCancelled // Event containing the contract specifics and raw log

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
func (it *BidMarketOrderCancelledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BidMarketOrderCancelled)
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
		it.Event = new(BidMarketOrderCancelled)
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
func (it *BidMarketOrderCancelledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BidMarketOrderCancelledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BidMarketOrderCancelled represents a OrderCancelled event raised by the BidMarket contract.
type BidMarketOrderCancelled struct {
	OrderId *big.Int
	Owner   common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterOrderCancelled is a free log retrieval operation binding the contract event 0xc0362da6f2ff36b382b34aec0814f6b3cdf89f5ef282a1d1f114d0c0b036d596.
//
// Solidity: event OrderCancelled(uint256 indexed orderId, address owner)
func (_BidMarket *BidMarketFilterer) FilterOrderCancelled(opts *bind.FilterOpts, orderId []*big.Int) (*BidMarketOrderCancelledIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _BidMarket.contract.FilterLogs(opts, "OrderCancelled", orderIdRule)
	if err != nil {
		return nil, err
	}
	return &BidMarketOrderCancelledIterator{contract: _BidMarket.contract, event: "OrderCancelled", logs: logs, sub: sub}, nil
}

// WatchOrderCancelled is a free log subscription operation binding the contract event 0xc0362da6f2ff36b382b34aec0814f6b3cdf89f5ef282a1d1f114d0c0b036d596.
//
// Solidity: event OrderCancelled(uint256 indexed orderId, address owner)
func (_BidMarket *BidMarketFilterer) WatchOrderCancelled(opts *bind.WatchOpts, sink chan<- *BidMarketOrderCancelled, orderId []*big.Int) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _BidMarket.contract.WatchLogs(opts, "OrderCancelled", orderIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BidMarketOrderCancelled)
				if err := _BidMarket.contract.UnpackLog(event, "OrderCancelled", log); err != nil {
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

// ParseOrderCancelled is a log parse operation binding the contract event 0xc0362da6f2ff36b382b34aec0814f6b3cdf89f5ef282a1d1f114d0c0b036d596.
//
// Solidity: event OrderCancelled(uint256 indexed orderId, address owner)
func (_BidMarket *BidMarketFilterer) ParseOrderCancelled(log types.Log) (*BidMarketOrderCancelled, error) {
	event := new(BidMarketOrderCancelled)
	if err := _BidMarket.contract.UnpackLog(event, "OrderCancelled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BidMarketOrderClosedIterator is returned from FilterOrderClosed and is used to iterate over the raw logs and unpacked data for OrderClosed events raised by the BidMarket contract.
type BidMarketOrderClosedIterator struct {
	Event *BidMarketOrderClosed // Event containing the contract specifics and raw log

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
func (it *BidMarketOrderClosedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BidMarketOrderClosed)
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
		it.Event = new(BidMarketOrderClosed)
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
func (it *BidMarketOrderClosedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BidMarketOrderClosedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BidMarketOrderClosed represents a OrderClosed event raised by the BidMarket contract.
type BidMarketOrderClosed struct {
	OrderId      *big.Int
	RefundAmount *big.Int
	Reason       string
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterOrderClosed is a free log retrieval operation binding the contract event 0xc2ebf8195090973176020e1d05b4a31dd5e03116437282453e16ce78d9318e7e.
//
// Solidity: event OrderClosed(uint256 indexed orderId, uint256 refundAmount, string reason)
func (_BidMarket *BidMarketFilterer) FilterOrderClosed(opts *bind.FilterOpts, orderId []*big.Int) (*BidMarketOrderClosedIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _BidMarket.contract.FilterLogs(opts, "OrderClosed", orderIdRule)
	if err != nil {
		return nil, err
	}
	return &BidMarketOrderClosedIterator{contract: _BidMarket.contract, event: "OrderClosed", logs: logs, sub: sub}, nil
}

// WatchOrderClosed is a free log subscription operation binding the contract event 0xc2ebf8195090973176020e1d05b4a31dd5e03116437282453e16ce78d9318e7e.
//
// Solidity: event OrderClosed(uint256 indexed orderId, uint256 refundAmount, string reason)
func (_BidMarket *BidMarketFilterer) WatchOrderClosed(opts *bind.WatchOpts, sink chan<- *BidMarketOrderClosed, orderId []*big.Int) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _BidMarket.contract.WatchLogs(opts, "OrderClosed", orderIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BidMarketOrderClosed)
				if err := _BidMarket.contract.UnpackLog(event, "OrderClosed", log); err != nil {
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

// ParseOrderClosed is a log parse operation binding the contract event 0xc2ebf8195090973176020e1d05b4a31dd5e03116437282453e16ce78d9318e7e.
//
// Solidity: event OrderClosed(uint256 indexed orderId, uint256 refundAmount, string reason)
func (_BidMarket *BidMarketFilterer) ParseOrderClosed(log types.Log) (*BidMarketOrderClosed, error) {
	event := new(BidMarketOrderClosed)
	if err := _BidMarket.contract.UnpackLog(event, "OrderClosed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BidMarketOrderCreatedIterator is returned from FilterOrderCreated and is used to iterate over the raw logs and unpacked data for OrderCreated events raised by the BidMarket contract.
type BidMarketOrderCreatedIterator struct {
	Event *BidMarketOrderCreated // Event containing the contract specifics and raw log

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
func (it *BidMarketOrderCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BidMarketOrderCreated)
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
		it.Event = new(BidMarketOrderCreated)
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
func (it *BidMarketOrderCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BidMarketOrderCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BidMarketOrderCreated represents a OrderCreated event raised by the BidMarket contract.
type BidMarketOrderCreated struct {
	OrderId  *big.Int
	Owner    common.Address
	Duration *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterOrderCreated is a free log retrieval operation binding the contract event 0x28eb86fd03c9bd96a8a83e4a46ab2ad257b70cea893d98ad6d482e96aa25fad0.
//
// Solidity: event OrderCreated(uint256 indexed orderId, address owner, uint256 duration)
func (_BidMarket *BidMarketFilterer) FilterOrderCreated(opts *bind.FilterOpts, orderId []*big.Int) (*BidMarketOrderCreatedIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _BidMarket.contract.FilterLogs(opts, "OrderCreated", orderIdRule)
	if err != nil {
		return nil, err
	}
	return &BidMarketOrderCreatedIterator{contract: _BidMarket.contract, event: "OrderCreated", logs: logs, sub: sub}, nil
}

// WatchOrderCreated is a free log subscription operation binding the contract event 0x28eb86fd03c9bd96a8a83e4a46ab2ad257b70cea893d98ad6d482e96aa25fad0.
//
// Solidity: event OrderCreated(uint256 indexed orderId, address owner, uint256 duration)
func (_BidMarket *BidMarketFilterer) WatchOrderCreated(opts *bind.WatchOpts, sink chan<- *BidMarketOrderCreated, orderId []*big.Int) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _BidMarket.contract.WatchLogs(opts, "OrderCreated", orderIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BidMarketOrderCreated)
				if err := _BidMarket.contract.UnpackLog(event, "OrderCreated", log); err != nil {
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

// ParseOrderCreated is a log parse operation binding the contract event 0x28eb86fd03c9bd96a8a83e4a46ab2ad257b70cea893d98ad6d482e96aa25fad0.
//
// Solidity: event OrderCreated(uint256 indexed orderId, address owner, uint256 duration)
func (_BidMarket *BidMarketFilterer) ParseOrderCreated(log types.Log) (*BidMarketOrderCreated, error) {
	event := new(BidMarketOrderCreated)
	if err := _BidMarket.contract.UnpackLog(event, "OrderCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BidMarketOrderExtendedIterator is returned from FilterOrderExtended and is used to iterate over the raw logs and unpacked data for OrderExtended events raised by the BidMarket contract.
type BidMarketOrderExtendedIterator struct {
	Event *BidMarketOrderExtended // Event containing the contract specifics and raw log

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
func (it *BidMarketOrderExtendedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BidMarketOrderExtended)
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
		it.Event = new(BidMarketOrderExtended)
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
func (it *BidMarketOrderExtendedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BidMarketOrderExtendedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BidMarketOrderExtended represents a OrderExtended event raised by the BidMarket contract.
type BidMarketOrderExtended struct {
	OrderId            *big.Int
	AdditionalDuration *big.Int
	NewExpiry          *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterOrderExtended is a free log retrieval operation binding the contract event 0xec5915f6de124dc9e88220ea0c9da53a73b64a575107c9999a716735dda7c3aa.
//
// Solidity: event OrderExtended(uint256 indexed orderId, uint256 additionalDuration, uint256 newExpiry)
func (_BidMarket *BidMarketFilterer) FilterOrderExtended(opts *bind.FilterOpts, orderId []*big.Int) (*BidMarketOrderExtendedIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _BidMarket.contract.FilterLogs(opts, "OrderExtended", orderIdRule)
	if err != nil {
		return nil, err
	}
	return &BidMarketOrderExtendedIterator{contract: _BidMarket.contract, event: "OrderExtended", logs: logs, sub: sub}, nil
}

// WatchOrderExtended is a free log subscription operation binding the contract event 0xec5915f6de124dc9e88220ea0c9da53a73b64a575107c9999a716735dda7c3aa.
//
// Solidity: event OrderExtended(uint256 indexed orderId, uint256 additionalDuration, uint256 newExpiry)
func (_BidMarket *BidMarketFilterer) WatchOrderExtended(opts *bind.WatchOpts, sink chan<- *BidMarketOrderExtended, orderId []*big.Int) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _BidMarket.contract.WatchLogs(opts, "OrderExtended", orderIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BidMarketOrderExtended)
				if err := _BidMarket.contract.UnpackLog(event, "OrderExtended", log); err != nil {
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

// ParseOrderExtended is a log parse operation binding the contract event 0xec5915f6de124dc9e88220ea0c9da53a73b64a575107c9999a716735dda7c3aa.
//
// Solidity: event OrderExtended(uint256 indexed orderId, uint256 additionalDuration, uint256 newExpiry)
func (_BidMarket *BidMarketFilterer) ParseOrderExtended(log types.Log) (*BidMarketOrderExtended, error) {
	event := new(BidMarketOrderExtended)
	if err := _BidMarket.contract.UnpackLog(event, "OrderExtended", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BidMarketOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the BidMarket contract.
type BidMarketOwnershipTransferredIterator struct {
	Event *BidMarketOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BidMarketOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BidMarketOwnershipTransferred)
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
		it.Event = new(BidMarketOwnershipTransferred)
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
func (it *BidMarketOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BidMarketOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BidMarketOwnershipTransferred represents a OwnershipTransferred event raised by the BidMarket contract.
type BidMarketOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BidMarket *BidMarketFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BidMarketOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BidMarket.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BidMarketOwnershipTransferredIterator{contract: _BidMarket.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BidMarket *BidMarketFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BidMarketOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BidMarket.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BidMarketOwnershipTransferred)
				if err := _BidMarket.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_BidMarket *BidMarketFilterer) ParseOwnershipTransferred(log types.Log) (*BidMarketOwnershipTransferred, error) {
	event := new(BidMarketOwnershipTransferred)
	if err := _BidMarket.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BidMarketPlatformFeeUpdatedIterator is returned from FilterPlatformFeeUpdated and is used to iterate over the raw logs and unpacked data for PlatformFeeUpdated events raised by the BidMarket contract.
type BidMarketPlatformFeeUpdatedIterator struct {
	Event *BidMarketPlatformFeeUpdated // Event containing the contract specifics and raw log

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
func (it *BidMarketPlatformFeeUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BidMarketPlatformFeeUpdated)
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
		it.Event = new(BidMarketPlatformFeeUpdated)
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
func (it *BidMarketPlatformFeeUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BidMarketPlatformFeeUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BidMarketPlatformFeeUpdated represents a PlatformFeeUpdated event raised by the BidMarket contract.
type BidMarketPlatformFeeUpdated struct {
	OldFee *big.Int
	NewFee *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterPlatformFeeUpdated is a free log retrieval operation binding the contract event 0xd347e206f25a89b917fc9482f1a2d294d749baa4dc9bde7fb495ee11fe491643.
//
// Solidity: event PlatformFeeUpdated(uint256 oldFee, uint256 newFee)
func (_BidMarket *BidMarketFilterer) FilterPlatformFeeUpdated(opts *bind.FilterOpts) (*BidMarketPlatformFeeUpdatedIterator, error) {

	logs, sub, err := _BidMarket.contract.FilterLogs(opts, "PlatformFeeUpdated")
	if err != nil {
		return nil, err
	}
	return &BidMarketPlatformFeeUpdatedIterator{contract: _BidMarket.contract, event: "PlatformFeeUpdated", logs: logs, sub: sub}, nil
}

// WatchPlatformFeeUpdated is a free log subscription operation binding the contract event 0xd347e206f25a89b917fc9482f1a2d294d749baa4dc9bde7fb495ee11fe491643.
//
// Solidity: event PlatformFeeUpdated(uint256 oldFee, uint256 newFee)
func (_BidMarket *BidMarketFilterer) WatchPlatformFeeUpdated(opts *bind.WatchOpts, sink chan<- *BidMarketPlatformFeeUpdated) (event.Subscription, error) {

	logs, sub, err := _BidMarket.contract.WatchLogs(opts, "PlatformFeeUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BidMarketPlatformFeeUpdated)
				if err := _BidMarket.contract.UnpackLog(event, "PlatformFeeUpdated", log); err != nil {
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

// ParsePlatformFeeUpdated is a log parse operation binding the contract event 0xd347e206f25a89b917fc9482f1a2d294d749baa4dc9bde7fb495ee11fe491643.
//
// Solidity: event PlatformFeeUpdated(uint256 oldFee, uint256 newFee)
func (_BidMarket *BidMarketFilterer) ParsePlatformFeeUpdated(log types.Log) (*BidMarketPlatformFeeUpdated, error) {
	event := new(BidMarketPlatformFeeUpdated)
	if err := _BidMarket.contract.UnpackLog(event, "PlatformFeeUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BidMarketPlatformWalletUpdatedIterator is returned from FilterPlatformWalletUpdated and is used to iterate over the raw logs and unpacked data for PlatformWalletUpdated events raised by the BidMarket contract.
type BidMarketPlatformWalletUpdatedIterator struct {
	Event *BidMarketPlatformWalletUpdated // Event containing the contract specifics and raw log

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
func (it *BidMarketPlatformWalletUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BidMarketPlatformWalletUpdated)
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
		it.Event = new(BidMarketPlatformWalletUpdated)
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
func (it *BidMarketPlatformWalletUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BidMarketPlatformWalletUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BidMarketPlatformWalletUpdated represents a PlatformWalletUpdated event raised by the BidMarket contract.
type BidMarketPlatformWalletUpdated struct {
	OldWallet common.Address
	NewWallet common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterPlatformWalletUpdated is a free log retrieval operation binding the contract event 0xf84f6b525c89a18c80bbbc0c62a7eb5390956d6b8432ab067b49963a98f8c38f.
//
// Solidity: event PlatformWalletUpdated(address oldWallet, address newWallet)
func (_BidMarket *BidMarketFilterer) FilterPlatformWalletUpdated(opts *bind.FilterOpts) (*BidMarketPlatformWalletUpdatedIterator, error) {

	logs, sub, err := _BidMarket.contract.FilterLogs(opts, "PlatformWalletUpdated")
	if err != nil {
		return nil, err
	}
	return &BidMarketPlatformWalletUpdatedIterator{contract: _BidMarket.contract, event: "PlatformWalletUpdated", logs: logs, sub: sub}, nil
}

// WatchPlatformWalletUpdated is a free log subscription operation binding the contract event 0xf84f6b525c89a18c80bbbc0c62a7eb5390956d6b8432ab067b49963a98f8c38f.
//
// Solidity: event PlatformWalletUpdated(address oldWallet, address newWallet)
func (_BidMarket *BidMarketFilterer) WatchPlatformWalletUpdated(opts *bind.WatchOpts, sink chan<- *BidMarketPlatformWalletUpdated) (event.Subscription, error) {

	logs, sub, err := _BidMarket.contract.WatchLogs(opts, "PlatformWalletUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BidMarketPlatformWalletUpdated)
				if err := _BidMarket.contract.UnpackLog(event, "PlatformWalletUpdated", log); err != nil {
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

// ParsePlatformWalletUpdated is a log parse operation binding the contract event 0xf84f6b525c89a18c80bbbc0c62a7eb5390956d6b8432ab067b49963a98f8c38f.
//
// Solidity: event PlatformWalletUpdated(address oldWallet, address newWallet)
func (_BidMarket *BidMarketFilterer) ParsePlatformWalletUpdated(log types.Log) (*BidMarketPlatformWalletUpdated, error) {
	event := new(BidMarketPlatformWalletUpdated)
	if err := _BidMarket.contract.UnpackLog(event, "PlatformWalletUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BidMarketSpecsUpdatedIterator is returned from FilterSpecsUpdated and is used to iterate over the raw logs and unpacked data for SpecsUpdated events raised by the BidMarket contract.
type BidMarketSpecsUpdatedIterator struct {
	Event *BidMarketSpecsUpdated // Event containing the contract specifics and raw log

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
func (it *BidMarketSpecsUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BidMarketSpecsUpdated)
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
		it.Event = new(BidMarketSpecsUpdated)
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
func (it *BidMarketSpecsUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BidMarketSpecsUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BidMarketSpecsUpdated represents a SpecsUpdated event raised by the BidMarket contract.
type BidMarketSpecsUpdated struct {
	OrderId  *big.Int
	NewSpecs string
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterSpecsUpdated is a free log retrieval operation binding the contract event 0x0138855fd02b911f24548d98d5a7548540eb70c0e0b823a8637a5ade8ad0afb0.
//
// Solidity: event SpecsUpdated(uint256 indexed orderId, string newSpecs)
func (_BidMarket *BidMarketFilterer) FilterSpecsUpdated(opts *bind.FilterOpts, orderId []*big.Int) (*BidMarketSpecsUpdatedIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _BidMarket.contract.FilterLogs(opts, "SpecsUpdated", orderIdRule)
	if err != nil {
		return nil, err
	}
	return &BidMarketSpecsUpdatedIterator{contract: _BidMarket.contract, event: "SpecsUpdated", logs: logs, sub: sub}, nil
}

// WatchSpecsUpdated is a free log subscription operation binding the contract event 0x0138855fd02b911f24548d98d5a7548540eb70c0e0b823a8637a5ade8ad0afb0.
//
// Solidity: event SpecsUpdated(uint256 indexed orderId, string newSpecs)
func (_BidMarket *BidMarketFilterer) WatchSpecsUpdated(opts *bind.WatchOpts, sink chan<- *BidMarketSpecsUpdated, orderId []*big.Int) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _BidMarket.contract.WatchLogs(opts, "SpecsUpdated", orderIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BidMarketSpecsUpdated)
				if err := _BidMarket.contract.UnpackLog(event, "SpecsUpdated", log); err != nil {
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

// ParseSpecsUpdated is a log parse operation binding the contract event 0x0138855fd02b911f24548d98d5a7548540eb70c0e0b823a8637a5ade8ad0afb0.
//
// Solidity: event SpecsUpdated(uint256 indexed orderId, string newSpecs)
func (_BidMarket *BidMarketFilterer) ParseSpecsUpdated(log types.Log) (*BidMarketSpecsUpdated, error) {
	event := new(BidMarketSpecsUpdated)
	if err := _BidMarket.contract.UnpackLog(event, "SpecsUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
