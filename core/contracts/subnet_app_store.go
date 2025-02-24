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

// SubnetAppStoreApp is an auto generated low-level Go binding around an user-defined struct.
type SubnetAppStoreApp struct {
	PeerId              string
	Owner               common.Address
	Operator            common.Address
	Verifier            common.Address
	Name                string
	Symbol              string
	Budget              *big.Int
	SpentBudget         *big.Int
	PricePerCpu         *big.Int
	PricePerGpu         *big.Int
	PricePerMemoryGB    *big.Int
	PricePerStorageGB   *big.Int
	PricePerBandwidthGB *big.Int
	Metadata            string
	PaymentToken        common.Address
}

// SubnetAppStoreDeployment is an auto generated low-level Go binding around an user-defined struct.
type SubnetAppStoreDeployment struct {
	IsRegistered  bool
	LastClaimTime *big.Int
	PendingReward *big.Int
}

// SubnetAppStoreMetaData contains all meta data concerning the SubnetAppStore contract.
var SubnetAppStoreMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"AddressEmptyCode\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"AddressInsufficientBalance\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ECDSAInvalidSignature\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"length\",\"type\":\"uint256\"}],\"name\":\"ECDSAInvalidSignatureLength\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"ECDSAInvalidSignatureS\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"SafeERC20FailedOperation\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"budget\",\"type\":\"uint256\"}],\"name\":\"AppCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"}],\"name\":\"AppUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"BudgetDeposited\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"BudgetRefunded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"}],\"name\":\"DeploymentClosed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"}],\"name\":\"DeploymentCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EIP712DomainChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ProviderRefunded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"reward\",\"type\":\"uint256\"}],\"name\":\"RewardClaimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"usedCpu\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"usedGpu\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"usedMemory\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"usedStorage\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"usedUploadBytes\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"usedDownloadBytes\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"reward\",\"type\":\"uint256\"}],\"name\":\"UsageReported\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"verifier\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"reward\",\"type\":\"uint256\"}],\"name\":\"VerifierRewardClaimed\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"appCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"apps\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"verifier\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"budget\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"spentBudget\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerMemoryGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerStorageGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerBandwidthGB\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"paymentToken\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedMemory\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedStorage\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedUploadBytes\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedDownloadBytes\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"name\":\"calculateReward\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"reward\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"}],\"name\":\"claimReward\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"budget\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerMemoryGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerStorageGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerBandwidthGB\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"verifier\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"paymentToken\",\"type\":\"address\"}],\"name\":\"createApp\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"deployments\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"isRegistered\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"lastClaimTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pendingReward\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"eip712Domain\",\"outputs\":[{\"internalType\":\"bytes1\",\"name\":\"fields\",\"type\":\"bytes1\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"verifyingContract\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"salt\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"extensions\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"feeRate\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"}],\"name\":\"getApp\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"verifier\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"budget\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"spentBudget\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerMemoryGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerStorageGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerBandwidthGB\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"paymentToken\",\"type\":\"address\"}],\"internalType\":\"structSubnetAppStore.App\",\"name\":\"app\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"}],\"name\":\"getDeployment\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"isRegistered\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"lastClaimTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pendingReward\",\"type\":\"uint256\"}],\"internalType\":\"structSubnetAppStore.Deployment\",\"name\":\"deployment\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_subnetProvider\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"initialOwner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_treasury\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_feeRate\",\"type\":\"uint256\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"start\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"end\",\"type\":\"uint256\"}],\"name\":\"listApps\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"verifier\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"budget\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"spentBudget\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerMemoryGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerStorageGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerBandwidthGB\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"paymentToken\",\"type\":\"address\"}],\"internalType\":\"structSubnetAppStore.App[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"}],\"name\":\"refund\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"}],\"name\":\"refundProvider\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"providerId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"usedCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedMemory\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedStorage\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedUploadBytes\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"usedDownloadBytes\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"reportUsage\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_feeRate\",\"type\":\"uint256\"}],\"name\":\"setFeeRate\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_treasury\",\"type\":\"address\"}],\"name\":\"setTreasury\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_verifierRewardRate\",\"type\":\"uint256\"}],\"name\":\"setVerifierRewardRate\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"subnetProvider\",\"outputs\":[{\"internalType\":\"contractSubnetProvider\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"symbolToAppId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"treasury\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerCpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerGpu\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerMemoryGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerStorageGB\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerBandwidthGB\",\"type\":\"uint256\"}],\"name\":\"updateApp\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"metadata\",\"type\":\"string\"}],\"name\":\"updateMetadata\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"}],\"name\":\"updateName\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"updateOperator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"peerId\",\"type\":\"string\"}],\"name\":\"updatePeerId\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"appId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"verifier\",\"type\":\"address\"}],\"name\":\"updateVerifier\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"usedMessageHashes\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"verifierRewardRate\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]",
}

// SubnetAppStoreABI is the input ABI used to generate the binding from.
// Deprecated: Use SubnetAppStoreMetaData.ABI instead.
var SubnetAppStoreABI = SubnetAppStoreMetaData.ABI

// SubnetAppStore is an auto generated Go binding around an Ethereum contract.
type SubnetAppStore struct {
	SubnetAppStoreCaller     // Read-only binding to the contract
	SubnetAppStoreTransactor // Write-only binding to the contract
	SubnetAppStoreFilterer   // Log filterer for contract events
}

// SubnetAppStoreCaller is an auto generated read-only Go binding around an Ethereum contract.
type SubnetAppStoreCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetAppStoreTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SubnetAppStoreTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetAppStoreFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SubnetAppStoreFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetAppStoreSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SubnetAppStoreSession struct {
	Contract     *SubnetAppStore   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SubnetAppStoreCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SubnetAppStoreCallerSession struct {
	Contract *SubnetAppStoreCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// SubnetAppStoreTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SubnetAppStoreTransactorSession struct {
	Contract     *SubnetAppStoreTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// SubnetAppStoreRaw is an auto generated low-level Go binding around an Ethereum contract.
type SubnetAppStoreRaw struct {
	Contract *SubnetAppStore // Generic contract binding to access the raw methods on
}

// SubnetAppStoreCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SubnetAppStoreCallerRaw struct {
	Contract *SubnetAppStoreCaller // Generic read-only contract binding to access the raw methods on
}

// SubnetAppStoreTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SubnetAppStoreTransactorRaw struct {
	Contract *SubnetAppStoreTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSubnetAppStore creates a new instance of SubnetAppStore, bound to a specific deployed contract.
func NewSubnetAppStore(address common.Address, backend bind.ContractBackend) (*SubnetAppStore, error) {
	contract, err := bindSubnetAppStore(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SubnetAppStore{SubnetAppStoreCaller: SubnetAppStoreCaller{contract: contract}, SubnetAppStoreTransactor: SubnetAppStoreTransactor{contract: contract}, SubnetAppStoreFilterer: SubnetAppStoreFilterer{contract: contract}}, nil
}

// NewSubnetAppStoreCaller creates a new read-only instance of SubnetAppStore, bound to a specific deployed contract.
func NewSubnetAppStoreCaller(address common.Address, caller bind.ContractCaller) (*SubnetAppStoreCaller, error) {
	contract, err := bindSubnetAppStore(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SubnetAppStoreCaller{contract: contract}, nil
}

// NewSubnetAppStoreTransactor creates a new write-only instance of SubnetAppStore, bound to a specific deployed contract.
func NewSubnetAppStoreTransactor(address common.Address, transactor bind.ContractTransactor) (*SubnetAppStoreTransactor, error) {
	contract, err := bindSubnetAppStore(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SubnetAppStoreTransactor{contract: contract}, nil
}

// NewSubnetAppStoreFilterer creates a new log filterer instance of SubnetAppStore, bound to a specific deployed contract.
func NewSubnetAppStoreFilterer(address common.Address, filterer bind.ContractFilterer) (*SubnetAppStoreFilterer, error) {
	contract, err := bindSubnetAppStore(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SubnetAppStoreFilterer{contract: contract}, nil
}

// bindSubnetAppStore binds a generic wrapper to an already deployed contract.
func bindSubnetAppStore(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SubnetAppStoreMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SubnetAppStore *SubnetAppStoreRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SubnetAppStore.Contract.SubnetAppStoreCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SubnetAppStore *SubnetAppStoreRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.SubnetAppStoreTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SubnetAppStore *SubnetAppStoreRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.SubnetAppStoreTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SubnetAppStore *SubnetAppStoreCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SubnetAppStore.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SubnetAppStore *SubnetAppStoreTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SubnetAppStore *SubnetAppStoreTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.contract.Transact(opts, method, params...)
}

// AppCount is a free data retrieval call binding the contract method 0xb55ca2c3.
//
// Solidity: function appCount() view returns(uint256)
func (_SubnetAppStore *SubnetAppStoreCaller) AppCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SubnetAppStore.contract.Call(opts, &out, "appCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// AppCount is a free data retrieval call binding the contract method 0xb55ca2c3.
//
// Solidity: function appCount() view returns(uint256)
func (_SubnetAppStore *SubnetAppStoreSession) AppCount() (*big.Int, error) {
	return _SubnetAppStore.Contract.AppCount(&_SubnetAppStore.CallOpts)
}

// AppCount is a free data retrieval call binding the contract method 0xb55ca2c3.
//
// Solidity: function appCount() view returns(uint256)
func (_SubnetAppStore *SubnetAppStoreCallerSession) AppCount() (*big.Int, error) {
	return _SubnetAppStore.Contract.AppCount(&_SubnetAppStore.CallOpts)
}

// Apps is a free data retrieval call binding the contract method 0x61acc37e.
//
// Solidity: function apps(uint256 ) view returns(string peerId, address owner, address operator, address verifier, string name, string symbol, uint256 budget, uint256 spentBudget, uint256 pricePerCpu, uint256 pricePerGpu, uint256 pricePerMemoryGB, uint256 pricePerStorageGB, uint256 pricePerBandwidthGB, string metadata, address paymentToken)
func (_SubnetAppStore *SubnetAppStoreCaller) Apps(opts *bind.CallOpts, arg0 *big.Int) (struct {
	PeerId              string
	Owner               common.Address
	Operator            common.Address
	Verifier            common.Address
	Name                string
	Symbol              string
	Budget              *big.Int
	SpentBudget         *big.Int
	PricePerCpu         *big.Int
	PricePerGpu         *big.Int
	PricePerMemoryGB    *big.Int
	PricePerStorageGB   *big.Int
	PricePerBandwidthGB *big.Int
	Metadata            string
	PaymentToken        common.Address
}, error) {
	var out []interface{}
	err := _SubnetAppStore.contract.Call(opts, &out, "apps", arg0)

	outstruct := new(struct {
		PeerId              string
		Owner               common.Address
		Operator            common.Address
		Verifier            common.Address
		Name                string
		Symbol              string
		Budget              *big.Int
		SpentBudget         *big.Int
		PricePerCpu         *big.Int
		PricePerGpu         *big.Int
		PricePerMemoryGB    *big.Int
		PricePerStorageGB   *big.Int
		PricePerBandwidthGB *big.Int
		Metadata            string
		PaymentToken        common.Address
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.PeerId = *abi.ConvertType(out[0], new(string)).(*string)
	outstruct.Owner = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	outstruct.Operator = *abi.ConvertType(out[2], new(common.Address)).(*common.Address)
	outstruct.Verifier = *abi.ConvertType(out[3], new(common.Address)).(*common.Address)
	outstruct.Name = *abi.ConvertType(out[4], new(string)).(*string)
	outstruct.Symbol = *abi.ConvertType(out[5], new(string)).(*string)
	outstruct.Budget = *abi.ConvertType(out[6], new(*big.Int)).(**big.Int)
	outstruct.SpentBudget = *abi.ConvertType(out[7], new(*big.Int)).(**big.Int)
	outstruct.PricePerCpu = *abi.ConvertType(out[8], new(*big.Int)).(**big.Int)
	outstruct.PricePerGpu = *abi.ConvertType(out[9], new(*big.Int)).(**big.Int)
	outstruct.PricePerMemoryGB = *abi.ConvertType(out[10], new(*big.Int)).(**big.Int)
	outstruct.PricePerStorageGB = *abi.ConvertType(out[11], new(*big.Int)).(**big.Int)
	outstruct.PricePerBandwidthGB = *abi.ConvertType(out[12], new(*big.Int)).(**big.Int)
	outstruct.Metadata = *abi.ConvertType(out[13], new(string)).(*string)
	outstruct.PaymentToken = *abi.ConvertType(out[14], new(common.Address)).(*common.Address)

	return *outstruct, err

}

// Apps is a free data retrieval call binding the contract method 0x61acc37e.
//
// Solidity: function apps(uint256 ) view returns(string peerId, address owner, address operator, address verifier, string name, string symbol, uint256 budget, uint256 spentBudget, uint256 pricePerCpu, uint256 pricePerGpu, uint256 pricePerMemoryGB, uint256 pricePerStorageGB, uint256 pricePerBandwidthGB, string metadata, address paymentToken)
func (_SubnetAppStore *SubnetAppStoreSession) Apps(arg0 *big.Int) (struct {
	PeerId              string
	Owner               common.Address
	Operator            common.Address
	Verifier            common.Address
	Name                string
	Symbol              string
	Budget              *big.Int
	SpentBudget         *big.Int
	PricePerCpu         *big.Int
	PricePerGpu         *big.Int
	PricePerMemoryGB    *big.Int
	PricePerStorageGB   *big.Int
	PricePerBandwidthGB *big.Int
	Metadata            string
	PaymentToken        common.Address
}, error) {
	return _SubnetAppStore.Contract.Apps(&_SubnetAppStore.CallOpts, arg0)
}

// Apps is a free data retrieval call binding the contract method 0x61acc37e.
//
// Solidity: function apps(uint256 ) view returns(string peerId, address owner, address operator, address verifier, string name, string symbol, uint256 budget, uint256 spentBudget, uint256 pricePerCpu, uint256 pricePerGpu, uint256 pricePerMemoryGB, uint256 pricePerStorageGB, uint256 pricePerBandwidthGB, string metadata, address paymentToken)
func (_SubnetAppStore *SubnetAppStoreCallerSession) Apps(arg0 *big.Int) (struct {
	PeerId              string
	Owner               common.Address
	Operator            common.Address
	Verifier            common.Address
	Name                string
	Symbol              string
	Budget              *big.Int
	SpentBudget         *big.Int
	PricePerCpu         *big.Int
	PricePerGpu         *big.Int
	PricePerMemoryGB    *big.Int
	PricePerStorageGB   *big.Int
	PricePerBandwidthGB *big.Int
	Metadata            string
	PaymentToken        common.Address
}, error) {
	return _SubnetAppStore.Contract.Apps(&_SubnetAppStore.CallOpts, arg0)
}

// CalculateReward is a free data retrieval call binding the contract method 0x86f66018.
//
// Solidity: function calculateReward(uint256 appId, uint256 usedCpu, uint256 usedGpu, uint256 usedMemory, uint256 usedStorage, uint256 usedUploadBytes, uint256 usedDownloadBytes, uint256 duration) view returns(uint256 reward)
func (_SubnetAppStore *SubnetAppStoreCaller) CalculateReward(opts *bind.CallOpts, appId *big.Int, usedCpu *big.Int, usedGpu *big.Int, usedMemory *big.Int, usedStorage *big.Int, usedUploadBytes *big.Int, usedDownloadBytes *big.Int, duration *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SubnetAppStore.contract.Call(opts, &out, "calculateReward", appId, usedCpu, usedGpu, usedMemory, usedStorage, usedUploadBytes, usedDownloadBytes, duration)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CalculateReward is a free data retrieval call binding the contract method 0x86f66018.
//
// Solidity: function calculateReward(uint256 appId, uint256 usedCpu, uint256 usedGpu, uint256 usedMemory, uint256 usedStorage, uint256 usedUploadBytes, uint256 usedDownloadBytes, uint256 duration) view returns(uint256 reward)
func (_SubnetAppStore *SubnetAppStoreSession) CalculateReward(appId *big.Int, usedCpu *big.Int, usedGpu *big.Int, usedMemory *big.Int, usedStorage *big.Int, usedUploadBytes *big.Int, usedDownloadBytes *big.Int, duration *big.Int) (*big.Int, error) {
	return _SubnetAppStore.Contract.CalculateReward(&_SubnetAppStore.CallOpts, appId, usedCpu, usedGpu, usedMemory, usedStorage, usedUploadBytes, usedDownloadBytes, duration)
}

// CalculateReward is a free data retrieval call binding the contract method 0x86f66018.
//
// Solidity: function calculateReward(uint256 appId, uint256 usedCpu, uint256 usedGpu, uint256 usedMemory, uint256 usedStorage, uint256 usedUploadBytes, uint256 usedDownloadBytes, uint256 duration) view returns(uint256 reward)
func (_SubnetAppStore *SubnetAppStoreCallerSession) CalculateReward(appId *big.Int, usedCpu *big.Int, usedGpu *big.Int, usedMemory *big.Int, usedStorage *big.Int, usedUploadBytes *big.Int, usedDownloadBytes *big.Int, duration *big.Int) (*big.Int, error) {
	return _SubnetAppStore.Contract.CalculateReward(&_SubnetAppStore.CallOpts, appId, usedCpu, usedGpu, usedMemory, usedStorage, usedUploadBytes, usedDownloadBytes, duration)
}

// Deployments is a free data retrieval call binding the contract method 0x20df8a23.
//
// Solidity: function deployments(uint256 , uint256 ) view returns(bool isRegistered, uint256 lastClaimTime, uint256 pendingReward)
func (_SubnetAppStore *SubnetAppStoreCaller) Deployments(opts *bind.CallOpts, arg0 *big.Int, arg1 *big.Int) (struct {
	IsRegistered  bool
	LastClaimTime *big.Int
	PendingReward *big.Int
}, error) {
	var out []interface{}
	err := _SubnetAppStore.contract.Call(opts, &out, "deployments", arg0, arg1)

	outstruct := new(struct {
		IsRegistered  bool
		LastClaimTime *big.Int
		PendingReward *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.IsRegistered = *abi.ConvertType(out[0], new(bool)).(*bool)
	outstruct.LastClaimTime = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.PendingReward = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Deployments is a free data retrieval call binding the contract method 0x20df8a23.
//
// Solidity: function deployments(uint256 , uint256 ) view returns(bool isRegistered, uint256 lastClaimTime, uint256 pendingReward)
func (_SubnetAppStore *SubnetAppStoreSession) Deployments(arg0 *big.Int, arg1 *big.Int) (struct {
	IsRegistered  bool
	LastClaimTime *big.Int
	PendingReward *big.Int
}, error) {
	return _SubnetAppStore.Contract.Deployments(&_SubnetAppStore.CallOpts, arg0, arg1)
}

// Deployments is a free data retrieval call binding the contract method 0x20df8a23.
//
// Solidity: function deployments(uint256 , uint256 ) view returns(bool isRegistered, uint256 lastClaimTime, uint256 pendingReward)
func (_SubnetAppStore *SubnetAppStoreCallerSession) Deployments(arg0 *big.Int, arg1 *big.Int) (struct {
	IsRegistered  bool
	LastClaimTime *big.Int
	PendingReward *big.Int
}, error) {
	return _SubnetAppStore.Contract.Deployments(&_SubnetAppStore.CallOpts, arg0, arg1)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_SubnetAppStore *SubnetAppStoreCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _SubnetAppStore.contract.Call(opts, &out, "eip712Domain")

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
func (_SubnetAppStore *SubnetAppStoreSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _SubnetAppStore.Contract.Eip712Domain(&_SubnetAppStore.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_SubnetAppStore *SubnetAppStoreCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _SubnetAppStore.Contract.Eip712Domain(&_SubnetAppStore.CallOpts)
}

// FeeRate is a free data retrieval call binding the contract method 0x978bbdb9.
//
// Solidity: function feeRate() view returns(uint256)
func (_SubnetAppStore *SubnetAppStoreCaller) FeeRate(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SubnetAppStore.contract.Call(opts, &out, "feeRate")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FeeRate is a free data retrieval call binding the contract method 0x978bbdb9.
//
// Solidity: function feeRate() view returns(uint256)
func (_SubnetAppStore *SubnetAppStoreSession) FeeRate() (*big.Int, error) {
	return _SubnetAppStore.Contract.FeeRate(&_SubnetAppStore.CallOpts)
}

// FeeRate is a free data retrieval call binding the contract method 0x978bbdb9.
//
// Solidity: function feeRate() view returns(uint256)
func (_SubnetAppStore *SubnetAppStoreCallerSession) FeeRate() (*big.Int, error) {
	return _SubnetAppStore.Contract.FeeRate(&_SubnetAppStore.CallOpts)
}

// GetApp is a free data retrieval call binding the contract method 0x24f3a51b.
//
// Solidity: function getApp(uint256 appId) view returns((string,address,address,address,string,string,uint256,uint256,uint256,uint256,uint256,uint256,uint256,string,address) app)
func (_SubnetAppStore *SubnetAppStoreCaller) GetApp(opts *bind.CallOpts, appId *big.Int) (SubnetAppStoreApp, error) {
	var out []interface{}
	err := _SubnetAppStore.contract.Call(opts, &out, "getApp", appId)

	if err != nil {
		return *new(SubnetAppStoreApp), err
	}

	out0 := *abi.ConvertType(out[0], new(SubnetAppStoreApp)).(*SubnetAppStoreApp)

	return out0, err

}

// GetApp is a free data retrieval call binding the contract method 0x24f3a51b.
//
// Solidity: function getApp(uint256 appId) view returns((string,address,address,address,string,string,uint256,uint256,uint256,uint256,uint256,uint256,uint256,string,address) app)
func (_SubnetAppStore *SubnetAppStoreSession) GetApp(appId *big.Int) (SubnetAppStoreApp, error) {
	return _SubnetAppStore.Contract.GetApp(&_SubnetAppStore.CallOpts, appId)
}

// GetApp is a free data retrieval call binding the contract method 0x24f3a51b.
//
// Solidity: function getApp(uint256 appId) view returns((string,address,address,address,string,string,uint256,uint256,uint256,uint256,uint256,uint256,uint256,string,address) app)
func (_SubnetAppStore *SubnetAppStoreCallerSession) GetApp(appId *big.Int) (SubnetAppStoreApp, error) {
	return _SubnetAppStore.Contract.GetApp(&_SubnetAppStore.CallOpts, appId)
}

// GetDeployment is a free data retrieval call binding the contract method 0x8b10fed6.
//
// Solidity: function getDeployment(uint256 appId, uint256 providerId) view returns((bool,uint256,uint256) deployment)
func (_SubnetAppStore *SubnetAppStoreCaller) GetDeployment(opts *bind.CallOpts, appId *big.Int, providerId *big.Int) (SubnetAppStoreDeployment, error) {
	var out []interface{}
	err := _SubnetAppStore.contract.Call(opts, &out, "getDeployment", appId, providerId)

	if err != nil {
		return *new(SubnetAppStoreDeployment), err
	}

	out0 := *abi.ConvertType(out[0], new(SubnetAppStoreDeployment)).(*SubnetAppStoreDeployment)

	return out0, err

}

// GetDeployment is a free data retrieval call binding the contract method 0x8b10fed6.
//
// Solidity: function getDeployment(uint256 appId, uint256 providerId) view returns((bool,uint256,uint256) deployment)
func (_SubnetAppStore *SubnetAppStoreSession) GetDeployment(appId *big.Int, providerId *big.Int) (SubnetAppStoreDeployment, error) {
	return _SubnetAppStore.Contract.GetDeployment(&_SubnetAppStore.CallOpts, appId, providerId)
}

// GetDeployment is a free data retrieval call binding the contract method 0x8b10fed6.
//
// Solidity: function getDeployment(uint256 appId, uint256 providerId) view returns((bool,uint256,uint256) deployment)
func (_SubnetAppStore *SubnetAppStoreCallerSession) GetDeployment(appId *big.Int, providerId *big.Int) (SubnetAppStoreDeployment, error) {
	return _SubnetAppStore.Contract.GetDeployment(&_SubnetAppStore.CallOpts, appId, providerId)
}

// ListApps is a free data retrieval call binding the contract method 0x3d205559.
//
// Solidity: function listApps(uint256 start, uint256 end) view returns((string,address,address,address,string,string,uint256,uint256,uint256,uint256,uint256,uint256,uint256,string,address)[])
func (_SubnetAppStore *SubnetAppStoreCaller) ListApps(opts *bind.CallOpts, start *big.Int, end *big.Int) ([]SubnetAppStoreApp, error) {
	var out []interface{}
	err := _SubnetAppStore.contract.Call(opts, &out, "listApps", start, end)

	if err != nil {
		return *new([]SubnetAppStoreApp), err
	}

	out0 := *abi.ConvertType(out[0], new([]SubnetAppStoreApp)).(*[]SubnetAppStoreApp)

	return out0, err

}

// ListApps is a free data retrieval call binding the contract method 0x3d205559.
//
// Solidity: function listApps(uint256 start, uint256 end) view returns((string,address,address,address,string,string,uint256,uint256,uint256,uint256,uint256,uint256,uint256,string,address)[])
func (_SubnetAppStore *SubnetAppStoreSession) ListApps(start *big.Int, end *big.Int) ([]SubnetAppStoreApp, error) {
	return _SubnetAppStore.Contract.ListApps(&_SubnetAppStore.CallOpts, start, end)
}

// ListApps is a free data retrieval call binding the contract method 0x3d205559.
//
// Solidity: function listApps(uint256 start, uint256 end) view returns((string,address,address,address,string,string,uint256,uint256,uint256,uint256,uint256,uint256,uint256,string,address)[])
func (_SubnetAppStore *SubnetAppStoreCallerSession) ListApps(start *big.Int, end *big.Int) ([]SubnetAppStoreApp, error) {
	return _SubnetAppStore.Contract.ListApps(&_SubnetAppStore.CallOpts, start, end)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetAppStore *SubnetAppStoreCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetAppStore.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetAppStore *SubnetAppStoreSession) Owner() (common.Address, error) {
	return _SubnetAppStore.Contract.Owner(&_SubnetAppStore.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetAppStore *SubnetAppStoreCallerSession) Owner() (common.Address, error) {
	return _SubnetAppStore.Contract.Owner(&_SubnetAppStore.CallOpts)
}

// SubnetProvider is a free data retrieval call binding the contract method 0x6a46dc8c.
//
// Solidity: function subnetProvider() view returns(address)
func (_SubnetAppStore *SubnetAppStoreCaller) SubnetProvider(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetAppStore.contract.Call(opts, &out, "subnetProvider")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SubnetProvider is a free data retrieval call binding the contract method 0x6a46dc8c.
//
// Solidity: function subnetProvider() view returns(address)
func (_SubnetAppStore *SubnetAppStoreSession) SubnetProvider() (common.Address, error) {
	return _SubnetAppStore.Contract.SubnetProvider(&_SubnetAppStore.CallOpts)
}

// SubnetProvider is a free data retrieval call binding the contract method 0x6a46dc8c.
//
// Solidity: function subnetProvider() view returns(address)
func (_SubnetAppStore *SubnetAppStoreCallerSession) SubnetProvider() (common.Address, error) {
	return _SubnetAppStore.Contract.SubnetProvider(&_SubnetAppStore.CallOpts)
}

// SymbolToAppId is a free data retrieval call binding the contract method 0xcb4b2360.
//
// Solidity: function symbolToAppId(string ) view returns(uint256)
func (_SubnetAppStore *SubnetAppStoreCaller) SymbolToAppId(opts *bind.CallOpts, arg0 string) (*big.Int, error) {
	var out []interface{}
	err := _SubnetAppStore.contract.Call(opts, &out, "symbolToAppId", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SymbolToAppId is a free data retrieval call binding the contract method 0xcb4b2360.
//
// Solidity: function symbolToAppId(string ) view returns(uint256)
func (_SubnetAppStore *SubnetAppStoreSession) SymbolToAppId(arg0 string) (*big.Int, error) {
	return _SubnetAppStore.Contract.SymbolToAppId(&_SubnetAppStore.CallOpts, arg0)
}

// SymbolToAppId is a free data retrieval call binding the contract method 0xcb4b2360.
//
// Solidity: function symbolToAppId(string ) view returns(uint256)
func (_SubnetAppStore *SubnetAppStoreCallerSession) SymbolToAppId(arg0 string) (*big.Int, error) {
	return _SubnetAppStore.Contract.SymbolToAppId(&_SubnetAppStore.CallOpts, arg0)
}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_SubnetAppStore *SubnetAppStoreCaller) Treasury(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetAppStore.contract.Call(opts, &out, "treasury")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_SubnetAppStore *SubnetAppStoreSession) Treasury() (common.Address, error) {
	return _SubnetAppStore.Contract.Treasury(&_SubnetAppStore.CallOpts)
}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_SubnetAppStore *SubnetAppStoreCallerSession) Treasury() (common.Address, error) {
	return _SubnetAppStore.Contract.Treasury(&_SubnetAppStore.CallOpts)
}

// UsedMessageHashes is a free data retrieval call binding the contract method 0x87d69207.
//
// Solidity: function usedMessageHashes(bytes32 ) view returns(bool)
func (_SubnetAppStore *SubnetAppStoreCaller) UsedMessageHashes(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var out []interface{}
	err := _SubnetAppStore.contract.Call(opts, &out, "usedMessageHashes", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// UsedMessageHashes is a free data retrieval call binding the contract method 0x87d69207.
//
// Solidity: function usedMessageHashes(bytes32 ) view returns(bool)
func (_SubnetAppStore *SubnetAppStoreSession) UsedMessageHashes(arg0 [32]byte) (bool, error) {
	return _SubnetAppStore.Contract.UsedMessageHashes(&_SubnetAppStore.CallOpts, arg0)
}

// UsedMessageHashes is a free data retrieval call binding the contract method 0x87d69207.
//
// Solidity: function usedMessageHashes(bytes32 ) view returns(bool)
func (_SubnetAppStore *SubnetAppStoreCallerSession) UsedMessageHashes(arg0 [32]byte) (bool, error) {
	return _SubnetAppStore.Contract.UsedMessageHashes(&_SubnetAppStore.CallOpts, arg0)
}

// VerifierRewardRate is a free data retrieval call binding the contract method 0xc93381f0.
//
// Solidity: function verifierRewardRate() view returns(uint256)
func (_SubnetAppStore *SubnetAppStoreCaller) VerifierRewardRate(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SubnetAppStore.contract.Call(opts, &out, "verifierRewardRate")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VerifierRewardRate is a free data retrieval call binding the contract method 0xc93381f0.
//
// Solidity: function verifierRewardRate() view returns(uint256)
func (_SubnetAppStore *SubnetAppStoreSession) VerifierRewardRate() (*big.Int, error) {
	return _SubnetAppStore.Contract.VerifierRewardRate(&_SubnetAppStore.CallOpts)
}

// VerifierRewardRate is a free data retrieval call binding the contract method 0xc93381f0.
//
// Solidity: function verifierRewardRate() view returns(uint256)
func (_SubnetAppStore *SubnetAppStoreCallerSession) VerifierRewardRate() (*big.Int, error) {
	return _SubnetAppStore.Contract.VerifierRewardRate(&_SubnetAppStore.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(string)
func (_SubnetAppStore *SubnetAppStoreCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _SubnetAppStore.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(string)
func (_SubnetAppStore *SubnetAppStoreSession) Version() (string, error) {
	return _SubnetAppStore.Contract.Version(&_SubnetAppStore.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(string)
func (_SubnetAppStore *SubnetAppStoreCallerSession) Version() (string, error) {
	return _SubnetAppStore.Contract.Version(&_SubnetAppStore.CallOpts)
}

// ClaimReward is a paid mutator transaction binding the contract method 0x86bb8f37.
//
// Solidity: function claimReward(uint256 providerId, uint256 appId) returns()
func (_SubnetAppStore *SubnetAppStoreTransactor) ClaimReward(opts *bind.TransactOpts, providerId *big.Int, appId *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.contract.Transact(opts, "claimReward", providerId, appId)
}

// ClaimReward is a paid mutator transaction binding the contract method 0x86bb8f37.
//
// Solidity: function claimReward(uint256 providerId, uint256 appId) returns()
func (_SubnetAppStore *SubnetAppStoreSession) ClaimReward(providerId *big.Int, appId *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.ClaimReward(&_SubnetAppStore.TransactOpts, providerId, appId)
}

// ClaimReward is a paid mutator transaction binding the contract method 0x86bb8f37.
//
// Solidity: function claimReward(uint256 providerId, uint256 appId) returns()
func (_SubnetAppStore *SubnetAppStoreTransactorSession) ClaimReward(providerId *big.Int, appId *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.ClaimReward(&_SubnetAppStore.TransactOpts, providerId, appId)
}

// CreateApp is a paid mutator transaction binding the contract method 0xf93579cc.
//
// Solidity: function createApp(string name, string symbol, string peerId, uint256 budget, uint256 pricePerCpu, uint256 pricePerGpu, uint256 pricePerMemoryGB, uint256 pricePerStorageGB, uint256 pricePerBandwidthGB, string metadata, address operator, address verifier, address paymentToken) returns(uint256)
func (_SubnetAppStore *SubnetAppStoreTransactor) CreateApp(opts *bind.TransactOpts, name string, symbol string, peerId string, budget *big.Int, pricePerCpu *big.Int, pricePerGpu *big.Int, pricePerMemoryGB *big.Int, pricePerStorageGB *big.Int, pricePerBandwidthGB *big.Int, metadata string, operator common.Address, verifier common.Address, paymentToken common.Address) (*types.Transaction, error) {
	return _SubnetAppStore.contract.Transact(opts, "createApp", name, symbol, peerId, budget, pricePerCpu, pricePerGpu, pricePerMemoryGB, pricePerStorageGB, pricePerBandwidthGB, metadata, operator, verifier, paymentToken)
}

// CreateApp is a paid mutator transaction binding the contract method 0xf93579cc.
//
// Solidity: function createApp(string name, string symbol, string peerId, uint256 budget, uint256 pricePerCpu, uint256 pricePerGpu, uint256 pricePerMemoryGB, uint256 pricePerStorageGB, uint256 pricePerBandwidthGB, string metadata, address operator, address verifier, address paymentToken) returns(uint256)
func (_SubnetAppStore *SubnetAppStoreSession) CreateApp(name string, symbol string, peerId string, budget *big.Int, pricePerCpu *big.Int, pricePerGpu *big.Int, pricePerMemoryGB *big.Int, pricePerStorageGB *big.Int, pricePerBandwidthGB *big.Int, metadata string, operator common.Address, verifier common.Address, paymentToken common.Address) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.CreateApp(&_SubnetAppStore.TransactOpts, name, symbol, peerId, budget, pricePerCpu, pricePerGpu, pricePerMemoryGB, pricePerStorageGB, pricePerBandwidthGB, metadata, operator, verifier, paymentToken)
}

// CreateApp is a paid mutator transaction binding the contract method 0xf93579cc.
//
// Solidity: function createApp(string name, string symbol, string peerId, uint256 budget, uint256 pricePerCpu, uint256 pricePerGpu, uint256 pricePerMemoryGB, uint256 pricePerStorageGB, uint256 pricePerBandwidthGB, string metadata, address operator, address verifier, address paymentToken) returns(uint256)
func (_SubnetAppStore *SubnetAppStoreTransactorSession) CreateApp(name string, symbol string, peerId string, budget *big.Int, pricePerCpu *big.Int, pricePerGpu *big.Int, pricePerMemoryGB *big.Int, pricePerStorageGB *big.Int, pricePerBandwidthGB *big.Int, metadata string, operator common.Address, verifier common.Address, paymentToken common.Address) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.CreateApp(&_SubnetAppStore.TransactOpts, name, symbol, peerId, budget, pricePerCpu, pricePerGpu, pricePerMemoryGB, pricePerStorageGB, pricePerBandwidthGB, metadata, operator, verifier, paymentToken)
}

// Deposit is a paid mutator transaction binding the contract method 0xe2bbb158.
//
// Solidity: function deposit(uint256 appId, uint256 amount) returns()
func (_SubnetAppStore *SubnetAppStoreTransactor) Deposit(opts *bind.TransactOpts, appId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.contract.Transact(opts, "deposit", appId, amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xe2bbb158.
//
// Solidity: function deposit(uint256 appId, uint256 amount) returns()
func (_SubnetAppStore *SubnetAppStoreSession) Deposit(appId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.Deposit(&_SubnetAppStore.TransactOpts, appId, amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xe2bbb158.
//
// Solidity: function deposit(uint256 appId, uint256 amount) returns()
func (_SubnetAppStore *SubnetAppStoreTransactorSession) Deposit(appId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.Deposit(&_SubnetAppStore.TransactOpts, appId, amount)
}

// Initialize is a paid mutator transaction binding the contract method 0xcf756fdf.
//
// Solidity: function initialize(address _subnetProvider, address initialOwner, address _treasury, uint256 _feeRate) returns()
func (_SubnetAppStore *SubnetAppStoreTransactor) Initialize(opts *bind.TransactOpts, _subnetProvider common.Address, initialOwner common.Address, _treasury common.Address, _feeRate *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.contract.Transact(opts, "initialize", _subnetProvider, initialOwner, _treasury, _feeRate)
}

// Initialize is a paid mutator transaction binding the contract method 0xcf756fdf.
//
// Solidity: function initialize(address _subnetProvider, address initialOwner, address _treasury, uint256 _feeRate) returns()
func (_SubnetAppStore *SubnetAppStoreSession) Initialize(_subnetProvider common.Address, initialOwner common.Address, _treasury common.Address, _feeRate *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.Initialize(&_SubnetAppStore.TransactOpts, _subnetProvider, initialOwner, _treasury, _feeRate)
}

// Initialize is a paid mutator transaction binding the contract method 0xcf756fdf.
//
// Solidity: function initialize(address _subnetProvider, address initialOwner, address _treasury, uint256 _feeRate) returns()
func (_SubnetAppStore *SubnetAppStoreTransactorSession) Initialize(_subnetProvider common.Address, initialOwner common.Address, _treasury common.Address, _feeRate *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.Initialize(&_SubnetAppStore.TransactOpts, _subnetProvider, initialOwner, _treasury, _feeRate)
}

// Refund is a paid mutator transaction binding the contract method 0x278ecde1.
//
// Solidity: function refund(uint256 appId) returns()
func (_SubnetAppStore *SubnetAppStoreTransactor) Refund(opts *bind.TransactOpts, appId *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.contract.Transact(opts, "refund", appId)
}

// Refund is a paid mutator transaction binding the contract method 0x278ecde1.
//
// Solidity: function refund(uint256 appId) returns()
func (_SubnetAppStore *SubnetAppStoreSession) Refund(appId *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.Refund(&_SubnetAppStore.TransactOpts, appId)
}

// Refund is a paid mutator transaction binding the contract method 0x278ecde1.
//
// Solidity: function refund(uint256 appId) returns()
func (_SubnetAppStore *SubnetAppStoreTransactorSession) Refund(appId *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.Refund(&_SubnetAppStore.TransactOpts, appId)
}

// RefundProvider is a paid mutator transaction binding the contract method 0x6252bda2.
//
// Solidity: function refundProvider(uint256 appId, uint256 providerId) returns()
func (_SubnetAppStore *SubnetAppStoreTransactor) RefundProvider(opts *bind.TransactOpts, appId *big.Int, providerId *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.contract.Transact(opts, "refundProvider", appId, providerId)
}

// RefundProvider is a paid mutator transaction binding the contract method 0x6252bda2.
//
// Solidity: function refundProvider(uint256 appId, uint256 providerId) returns()
func (_SubnetAppStore *SubnetAppStoreSession) RefundProvider(appId *big.Int, providerId *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.RefundProvider(&_SubnetAppStore.TransactOpts, appId, providerId)
}

// RefundProvider is a paid mutator transaction binding the contract method 0x6252bda2.
//
// Solidity: function refundProvider(uint256 appId, uint256 providerId) returns()
func (_SubnetAppStore *SubnetAppStoreTransactorSession) RefundProvider(appId *big.Int, providerId *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.RefundProvider(&_SubnetAppStore.TransactOpts, appId, providerId)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetAppStore *SubnetAppStoreTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetAppStore.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetAppStore *SubnetAppStoreSession) RenounceOwnership() (*types.Transaction, error) {
	return _SubnetAppStore.Contract.RenounceOwnership(&_SubnetAppStore.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetAppStore *SubnetAppStoreTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _SubnetAppStore.Contract.RenounceOwnership(&_SubnetAppStore.TransactOpts)
}

// ReportUsage is a paid mutator transaction binding the contract method 0x8f4e980d.
//
// Solidity: function reportUsage(uint256 appId, uint256 providerId, string peerId, uint256 usedCpu, uint256 usedGpu, uint256 usedMemory, uint256 usedStorage, uint256 usedUploadBytes, uint256 usedDownloadBytes, uint256 duration, uint256 timestamp, bytes signature) returns()
func (_SubnetAppStore *SubnetAppStoreTransactor) ReportUsage(opts *bind.TransactOpts, appId *big.Int, providerId *big.Int, peerId string, usedCpu *big.Int, usedGpu *big.Int, usedMemory *big.Int, usedStorage *big.Int, usedUploadBytes *big.Int, usedDownloadBytes *big.Int, duration *big.Int, timestamp *big.Int, signature []byte) (*types.Transaction, error) {
	return _SubnetAppStore.contract.Transact(opts, "reportUsage", appId, providerId, peerId, usedCpu, usedGpu, usedMemory, usedStorage, usedUploadBytes, usedDownloadBytes, duration, timestamp, signature)
}

// ReportUsage is a paid mutator transaction binding the contract method 0x8f4e980d.
//
// Solidity: function reportUsage(uint256 appId, uint256 providerId, string peerId, uint256 usedCpu, uint256 usedGpu, uint256 usedMemory, uint256 usedStorage, uint256 usedUploadBytes, uint256 usedDownloadBytes, uint256 duration, uint256 timestamp, bytes signature) returns()
func (_SubnetAppStore *SubnetAppStoreSession) ReportUsage(appId *big.Int, providerId *big.Int, peerId string, usedCpu *big.Int, usedGpu *big.Int, usedMemory *big.Int, usedStorage *big.Int, usedUploadBytes *big.Int, usedDownloadBytes *big.Int, duration *big.Int, timestamp *big.Int, signature []byte) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.ReportUsage(&_SubnetAppStore.TransactOpts, appId, providerId, peerId, usedCpu, usedGpu, usedMemory, usedStorage, usedUploadBytes, usedDownloadBytes, duration, timestamp, signature)
}

// ReportUsage is a paid mutator transaction binding the contract method 0x8f4e980d.
//
// Solidity: function reportUsage(uint256 appId, uint256 providerId, string peerId, uint256 usedCpu, uint256 usedGpu, uint256 usedMemory, uint256 usedStorage, uint256 usedUploadBytes, uint256 usedDownloadBytes, uint256 duration, uint256 timestamp, bytes signature) returns()
func (_SubnetAppStore *SubnetAppStoreTransactorSession) ReportUsage(appId *big.Int, providerId *big.Int, peerId string, usedCpu *big.Int, usedGpu *big.Int, usedMemory *big.Int, usedStorage *big.Int, usedUploadBytes *big.Int, usedDownloadBytes *big.Int, duration *big.Int, timestamp *big.Int, signature []byte) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.ReportUsage(&_SubnetAppStore.TransactOpts, appId, providerId, peerId, usedCpu, usedGpu, usedMemory, usedStorage, usedUploadBytes, usedDownloadBytes, duration, timestamp, signature)
}

// SetFeeRate is a paid mutator transaction binding the contract method 0x45596e2e.
//
// Solidity: function setFeeRate(uint256 _feeRate) returns()
func (_SubnetAppStore *SubnetAppStoreTransactor) SetFeeRate(opts *bind.TransactOpts, _feeRate *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.contract.Transact(opts, "setFeeRate", _feeRate)
}

// SetFeeRate is a paid mutator transaction binding the contract method 0x45596e2e.
//
// Solidity: function setFeeRate(uint256 _feeRate) returns()
func (_SubnetAppStore *SubnetAppStoreSession) SetFeeRate(_feeRate *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.SetFeeRate(&_SubnetAppStore.TransactOpts, _feeRate)
}

// SetFeeRate is a paid mutator transaction binding the contract method 0x45596e2e.
//
// Solidity: function setFeeRate(uint256 _feeRate) returns()
func (_SubnetAppStore *SubnetAppStoreTransactorSession) SetFeeRate(_feeRate *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.SetFeeRate(&_SubnetAppStore.TransactOpts, _feeRate)
}

// SetTreasury is a paid mutator transaction binding the contract method 0xf0f44260.
//
// Solidity: function setTreasury(address _treasury) returns()
func (_SubnetAppStore *SubnetAppStoreTransactor) SetTreasury(opts *bind.TransactOpts, _treasury common.Address) (*types.Transaction, error) {
	return _SubnetAppStore.contract.Transact(opts, "setTreasury", _treasury)
}

// SetTreasury is a paid mutator transaction binding the contract method 0xf0f44260.
//
// Solidity: function setTreasury(address _treasury) returns()
func (_SubnetAppStore *SubnetAppStoreSession) SetTreasury(_treasury common.Address) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.SetTreasury(&_SubnetAppStore.TransactOpts, _treasury)
}

// SetTreasury is a paid mutator transaction binding the contract method 0xf0f44260.
//
// Solidity: function setTreasury(address _treasury) returns()
func (_SubnetAppStore *SubnetAppStoreTransactorSession) SetTreasury(_treasury common.Address) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.SetTreasury(&_SubnetAppStore.TransactOpts, _treasury)
}

// SetVerifierRewardRate is a paid mutator transaction binding the contract method 0xd2f55802.
//
// Solidity: function setVerifierRewardRate(uint256 _verifierRewardRate) returns()
func (_SubnetAppStore *SubnetAppStoreTransactor) SetVerifierRewardRate(opts *bind.TransactOpts, _verifierRewardRate *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.contract.Transact(opts, "setVerifierRewardRate", _verifierRewardRate)
}

// SetVerifierRewardRate is a paid mutator transaction binding the contract method 0xd2f55802.
//
// Solidity: function setVerifierRewardRate(uint256 _verifierRewardRate) returns()
func (_SubnetAppStore *SubnetAppStoreSession) SetVerifierRewardRate(_verifierRewardRate *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.SetVerifierRewardRate(&_SubnetAppStore.TransactOpts, _verifierRewardRate)
}

// SetVerifierRewardRate is a paid mutator transaction binding the contract method 0xd2f55802.
//
// Solidity: function setVerifierRewardRate(uint256 _verifierRewardRate) returns()
func (_SubnetAppStore *SubnetAppStoreTransactorSession) SetVerifierRewardRate(_verifierRewardRate *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.SetVerifierRewardRate(&_SubnetAppStore.TransactOpts, _verifierRewardRate)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetAppStore *SubnetAppStoreTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _SubnetAppStore.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetAppStore *SubnetAppStoreSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.TransferOwnership(&_SubnetAppStore.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetAppStore *SubnetAppStoreTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.TransferOwnership(&_SubnetAppStore.TransactOpts, newOwner)
}

// UpdateApp is a paid mutator transaction binding the contract method 0xf28ef0e8.
//
// Solidity: function updateApp(uint256 appId, uint256 pricePerCpu, uint256 pricePerGpu, uint256 pricePerMemoryGB, uint256 pricePerStorageGB, uint256 pricePerBandwidthGB) returns()
func (_SubnetAppStore *SubnetAppStoreTransactor) UpdateApp(opts *bind.TransactOpts, appId *big.Int, pricePerCpu *big.Int, pricePerGpu *big.Int, pricePerMemoryGB *big.Int, pricePerStorageGB *big.Int, pricePerBandwidthGB *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.contract.Transact(opts, "updateApp", appId, pricePerCpu, pricePerGpu, pricePerMemoryGB, pricePerStorageGB, pricePerBandwidthGB)
}

// UpdateApp is a paid mutator transaction binding the contract method 0xf28ef0e8.
//
// Solidity: function updateApp(uint256 appId, uint256 pricePerCpu, uint256 pricePerGpu, uint256 pricePerMemoryGB, uint256 pricePerStorageGB, uint256 pricePerBandwidthGB) returns()
func (_SubnetAppStore *SubnetAppStoreSession) UpdateApp(appId *big.Int, pricePerCpu *big.Int, pricePerGpu *big.Int, pricePerMemoryGB *big.Int, pricePerStorageGB *big.Int, pricePerBandwidthGB *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.UpdateApp(&_SubnetAppStore.TransactOpts, appId, pricePerCpu, pricePerGpu, pricePerMemoryGB, pricePerStorageGB, pricePerBandwidthGB)
}

// UpdateApp is a paid mutator transaction binding the contract method 0xf28ef0e8.
//
// Solidity: function updateApp(uint256 appId, uint256 pricePerCpu, uint256 pricePerGpu, uint256 pricePerMemoryGB, uint256 pricePerStorageGB, uint256 pricePerBandwidthGB) returns()
func (_SubnetAppStore *SubnetAppStoreTransactorSession) UpdateApp(appId *big.Int, pricePerCpu *big.Int, pricePerGpu *big.Int, pricePerMemoryGB *big.Int, pricePerStorageGB *big.Int, pricePerBandwidthGB *big.Int) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.UpdateApp(&_SubnetAppStore.TransactOpts, appId, pricePerCpu, pricePerGpu, pricePerMemoryGB, pricePerStorageGB, pricePerBandwidthGB)
}

// UpdateMetadata is a paid mutator transaction binding the contract method 0x53c8388e.
//
// Solidity: function updateMetadata(uint256 appId, string metadata) returns()
func (_SubnetAppStore *SubnetAppStoreTransactor) UpdateMetadata(opts *bind.TransactOpts, appId *big.Int, metadata string) (*types.Transaction, error) {
	return _SubnetAppStore.contract.Transact(opts, "updateMetadata", appId, metadata)
}

// UpdateMetadata is a paid mutator transaction binding the contract method 0x53c8388e.
//
// Solidity: function updateMetadata(uint256 appId, string metadata) returns()
func (_SubnetAppStore *SubnetAppStoreSession) UpdateMetadata(appId *big.Int, metadata string) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.UpdateMetadata(&_SubnetAppStore.TransactOpts, appId, metadata)
}

// UpdateMetadata is a paid mutator transaction binding the contract method 0x53c8388e.
//
// Solidity: function updateMetadata(uint256 appId, string metadata) returns()
func (_SubnetAppStore *SubnetAppStoreTransactorSession) UpdateMetadata(appId *big.Int, metadata string) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.UpdateMetadata(&_SubnetAppStore.TransactOpts, appId, metadata)
}

// UpdateName is a paid mutator transaction binding the contract method 0x53e76f2c.
//
// Solidity: function updateName(uint256 appId, string name) returns()
func (_SubnetAppStore *SubnetAppStoreTransactor) UpdateName(opts *bind.TransactOpts, appId *big.Int, name string) (*types.Transaction, error) {
	return _SubnetAppStore.contract.Transact(opts, "updateName", appId, name)
}

// UpdateName is a paid mutator transaction binding the contract method 0x53e76f2c.
//
// Solidity: function updateName(uint256 appId, string name) returns()
func (_SubnetAppStore *SubnetAppStoreSession) UpdateName(appId *big.Int, name string) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.UpdateName(&_SubnetAppStore.TransactOpts, appId, name)
}

// UpdateName is a paid mutator transaction binding the contract method 0x53e76f2c.
//
// Solidity: function updateName(uint256 appId, string name) returns()
func (_SubnetAppStore *SubnetAppStoreTransactorSession) UpdateName(appId *big.Int, name string) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.UpdateName(&_SubnetAppStore.TransactOpts, appId, name)
}

// UpdateOperator is a paid mutator transaction binding the contract method 0xc65a391d.
//
// Solidity: function updateOperator(uint256 appId, address operator) returns()
func (_SubnetAppStore *SubnetAppStoreTransactor) UpdateOperator(opts *bind.TransactOpts, appId *big.Int, operator common.Address) (*types.Transaction, error) {
	return _SubnetAppStore.contract.Transact(opts, "updateOperator", appId, operator)
}

// UpdateOperator is a paid mutator transaction binding the contract method 0xc65a391d.
//
// Solidity: function updateOperator(uint256 appId, address operator) returns()
func (_SubnetAppStore *SubnetAppStoreSession) UpdateOperator(appId *big.Int, operator common.Address) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.UpdateOperator(&_SubnetAppStore.TransactOpts, appId, operator)
}

// UpdateOperator is a paid mutator transaction binding the contract method 0xc65a391d.
//
// Solidity: function updateOperator(uint256 appId, address operator) returns()
func (_SubnetAppStore *SubnetAppStoreTransactorSession) UpdateOperator(appId *big.Int, operator common.Address) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.UpdateOperator(&_SubnetAppStore.TransactOpts, appId, operator)
}

// UpdatePeerId is a paid mutator transaction binding the contract method 0xe89c7408.
//
// Solidity: function updatePeerId(uint256 appId, string peerId) returns()
func (_SubnetAppStore *SubnetAppStoreTransactor) UpdatePeerId(opts *bind.TransactOpts, appId *big.Int, peerId string) (*types.Transaction, error) {
	return _SubnetAppStore.contract.Transact(opts, "updatePeerId", appId, peerId)
}

// UpdatePeerId is a paid mutator transaction binding the contract method 0xe89c7408.
//
// Solidity: function updatePeerId(uint256 appId, string peerId) returns()
func (_SubnetAppStore *SubnetAppStoreSession) UpdatePeerId(appId *big.Int, peerId string) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.UpdatePeerId(&_SubnetAppStore.TransactOpts, appId, peerId)
}

// UpdatePeerId is a paid mutator transaction binding the contract method 0xe89c7408.
//
// Solidity: function updatePeerId(uint256 appId, string peerId) returns()
func (_SubnetAppStore *SubnetAppStoreTransactorSession) UpdatePeerId(appId *big.Int, peerId string) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.UpdatePeerId(&_SubnetAppStore.TransactOpts, appId, peerId)
}

// UpdateVerifier is a paid mutator transaction binding the contract method 0x293de354.
//
// Solidity: function updateVerifier(uint256 appId, address verifier) returns()
func (_SubnetAppStore *SubnetAppStoreTransactor) UpdateVerifier(opts *bind.TransactOpts, appId *big.Int, verifier common.Address) (*types.Transaction, error) {
	return _SubnetAppStore.contract.Transact(opts, "updateVerifier", appId, verifier)
}

// UpdateVerifier is a paid mutator transaction binding the contract method 0x293de354.
//
// Solidity: function updateVerifier(uint256 appId, address verifier) returns()
func (_SubnetAppStore *SubnetAppStoreSession) UpdateVerifier(appId *big.Int, verifier common.Address) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.UpdateVerifier(&_SubnetAppStore.TransactOpts, appId, verifier)
}

// UpdateVerifier is a paid mutator transaction binding the contract method 0x293de354.
//
// Solidity: function updateVerifier(uint256 appId, address verifier) returns()
func (_SubnetAppStore *SubnetAppStoreTransactorSession) UpdateVerifier(appId *big.Int, verifier common.Address) (*types.Transaction, error) {
	return _SubnetAppStore.Contract.UpdateVerifier(&_SubnetAppStore.TransactOpts, appId, verifier)
}

// SubnetAppStoreAppCreatedIterator is returned from FilterAppCreated and is used to iterate over the raw logs and unpacked data for AppCreated events raised by the SubnetAppStore contract.
type SubnetAppStoreAppCreatedIterator struct {
	Event *SubnetAppStoreAppCreated // Event containing the contract specifics and raw log

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
func (it *SubnetAppStoreAppCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetAppStoreAppCreated)
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
		it.Event = new(SubnetAppStoreAppCreated)
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
func (it *SubnetAppStoreAppCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetAppStoreAppCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetAppStoreAppCreated represents a AppCreated event raised by the SubnetAppStore contract.
type SubnetAppStoreAppCreated struct {
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
func (_SubnetAppStore *SubnetAppStoreFilterer) FilterAppCreated(opts *bind.FilterOpts, appId []*big.Int, owner []common.Address) (*SubnetAppStoreAppCreatedIterator, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _SubnetAppStore.contract.FilterLogs(opts, "AppCreated", appIdRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &SubnetAppStoreAppCreatedIterator{contract: _SubnetAppStore.contract, event: "AppCreated", logs: logs, sub: sub}, nil
}

// WatchAppCreated is a free log subscription operation binding the contract event 0x9b96599010d3232696eb9dcbd3fa0a37ce73e2ffccd85f705a63a90afc65e420.
//
// Solidity: event AppCreated(uint256 indexed appId, string name, string symbol, address indexed owner, uint256 budget)
func (_SubnetAppStore *SubnetAppStoreFilterer) WatchAppCreated(opts *bind.WatchOpts, sink chan<- *SubnetAppStoreAppCreated, appId []*big.Int, owner []common.Address) (event.Subscription, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _SubnetAppStore.contract.WatchLogs(opts, "AppCreated", appIdRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetAppStoreAppCreated)
				if err := _SubnetAppStore.contract.UnpackLog(event, "AppCreated", log); err != nil {
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
func (_SubnetAppStore *SubnetAppStoreFilterer) ParseAppCreated(log types.Log) (*SubnetAppStoreAppCreated, error) {
	event := new(SubnetAppStoreAppCreated)
	if err := _SubnetAppStore.contract.UnpackLog(event, "AppCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetAppStoreAppUpdatedIterator is returned from FilterAppUpdated and is used to iterate over the raw logs and unpacked data for AppUpdated events raised by the SubnetAppStore contract.
type SubnetAppStoreAppUpdatedIterator struct {
	Event *SubnetAppStoreAppUpdated // Event containing the contract specifics and raw log

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
func (it *SubnetAppStoreAppUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetAppStoreAppUpdated)
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
		it.Event = new(SubnetAppStoreAppUpdated)
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
func (it *SubnetAppStoreAppUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetAppStoreAppUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetAppStoreAppUpdated represents a AppUpdated event raised by the SubnetAppStore contract.
type SubnetAppStoreAppUpdated struct {
	AppId *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterAppUpdated is a free log retrieval operation binding the contract event 0x5f9e31e4fc57fabac502d848e3a5d8ee121194d5ed670c12476c4fe260924157.
//
// Solidity: event AppUpdated(uint256 appId)
func (_SubnetAppStore *SubnetAppStoreFilterer) FilterAppUpdated(opts *bind.FilterOpts) (*SubnetAppStoreAppUpdatedIterator, error) {

	logs, sub, err := _SubnetAppStore.contract.FilterLogs(opts, "AppUpdated")
	if err != nil {
		return nil, err
	}
	return &SubnetAppStoreAppUpdatedIterator{contract: _SubnetAppStore.contract, event: "AppUpdated", logs: logs, sub: sub}, nil
}

// WatchAppUpdated is a free log subscription operation binding the contract event 0x5f9e31e4fc57fabac502d848e3a5d8ee121194d5ed670c12476c4fe260924157.
//
// Solidity: event AppUpdated(uint256 appId)
func (_SubnetAppStore *SubnetAppStoreFilterer) WatchAppUpdated(opts *bind.WatchOpts, sink chan<- *SubnetAppStoreAppUpdated) (event.Subscription, error) {

	logs, sub, err := _SubnetAppStore.contract.WatchLogs(opts, "AppUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetAppStoreAppUpdated)
				if err := _SubnetAppStore.contract.UnpackLog(event, "AppUpdated", log); err != nil {
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
func (_SubnetAppStore *SubnetAppStoreFilterer) ParseAppUpdated(log types.Log) (*SubnetAppStoreAppUpdated, error) {
	event := new(SubnetAppStoreAppUpdated)
	if err := _SubnetAppStore.contract.UnpackLog(event, "AppUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetAppStoreBudgetDepositedIterator is returned from FilterBudgetDeposited and is used to iterate over the raw logs and unpacked data for BudgetDeposited events raised by the SubnetAppStore contract.
type SubnetAppStoreBudgetDepositedIterator struct {
	Event *SubnetAppStoreBudgetDeposited // Event containing the contract specifics and raw log

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
func (it *SubnetAppStoreBudgetDepositedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetAppStoreBudgetDeposited)
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
		it.Event = new(SubnetAppStoreBudgetDeposited)
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
func (it *SubnetAppStoreBudgetDepositedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetAppStoreBudgetDepositedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetAppStoreBudgetDeposited represents a BudgetDeposited event raised by the SubnetAppStore contract.
type SubnetAppStoreBudgetDeposited struct {
	AppId  *big.Int
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterBudgetDeposited is a free log retrieval operation binding the contract event 0x7fa8170314b2ec131ba6754af9b3e7428a0bd93e1f6f31356422785f25cebedd.
//
// Solidity: event BudgetDeposited(uint256 indexed appId, uint256 amount)
func (_SubnetAppStore *SubnetAppStoreFilterer) FilterBudgetDeposited(opts *bind.FilterOpts, appId []*big.Int) (*SubnetAppStoreBudgetDepositedIterator, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}

	logs, sub, err := _SubnetAppStore.contract.FilterLogs(opts, "BudgetDeposited", appIdRule)
	if err != nil {
		return nil, err
	}
	return &SubnetAppStoreBudgetDepositedIterator{contract: _SubnetAppStore.contract, event: "BudgetDeposited", logs: logs, sub: sub}, nil
}

// WatchBudgetDeposited is a free log subscription operation binding the contract event 0x7fa8170314b2ec131ba6754af9b3e7428a0bd93e1f6f31356422785f25cebedd.
//
// Solidity: event BudgetDeposited(uint256 indexed appId, uint256 amount)
func (_SubnetAppStore *SubnetAppStoreFilterer) WatchBudgetDeposited(opts *bind.WatchOpts, sink chan<- *SubnetAppStoreBudgetDeposited, appId []*big.Int) (event.Subscription, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}

	logs, sub, err := _SubnetAppStore.contract.WatchLogs(opts, "BudgetDeposited", appIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetAppStoreBudgetDeposited)
				if err := _SubnetAppStore.contract.UnpackLog(event, "BudgetDeposited", log); err != nil {
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

// ParseBudgetDeposited is a log parse operation binding the contract event 0x7fa8170314b2ec131ba6754af9b3e7428a0bd93e1f6f31356422785f25cebedd.
//
// Solidity: event BudgetDeposited(uint256 indexed appId, uint256 amount)
func (_SubnetAppStore *SubnetAppStoreFilterer) ParseBudgetDeposited(log types.Log) (*SubnetAppStoreBudgetDeposited, error) {
	event := new(SubnetAppStoreBudgetDeposited)
	if err := _SubnetAppStore.contract.UnpackLog(event, "BudgetDeposited", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetAppStoreBudgetRefundedIterator is returned from FilterBudgetRefunded and is used to iterate over the raw logs and unpacked data for BudgetRefunded events raised by the SubnetAppStore contract.
type SubnetAppStoreBudgetRefundedIterator struct {
	Event *SubnetAppStoreBudgetRefunded // Event containing the contract specifics and raw log

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
func (it *SubnetAppStoreBudgetRefundedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetAppStoreBudgetRefunded)
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
		it.Event = new(SubnetAppStoreBudgetRefunded)
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
func (it *SubnetAppStoreBudgetRefundedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetAppStoreBudgetRefundedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetAppStoreBudgetRefunded represents a BudgetRefunded event raised by the SubnetAppStore contract.
type SubnetAppStoreBudgetRefunded struct {
	AppId  *big.Int
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterBudgetRefunded is a free log retrieval operation binding the contract event 0x95da8a4f660594f653bc780752dadf78f4f089254077e243f1318f946465fd09.
//
// Solidity: event BudgetRefunded(uint256 indexed appId, uint256 amount)
func (_SubnetAppStore *SubnetAppStoreFilterer) FilterBudgetRefunded(opts *bind.FilterOpts, appId []*big.Int) (*SubnetAppStoreBudgetRefundedIterator, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}

	logs, sub, err := _SubnetAppStore.contract.FilterLogs(opts, "BudgetRefunded", appIdRule)
	if err != nil {
		return nil, err
	}
	return &SubnetAppStoreBudgetRefundedIterator{contract: _SubnetAppStore.contract, event: "BudgetRefunded", logs: logs, sub: sub}, nil
}

// WatchBudgetRefunded is a free log subscription operation binding the contract event 0x95da8a4f660594f653bc780752dadf78f4f089254077e243f1318f946465fd09.
//
// Solidity: event BudgetRefunded(uint256 indexed appId, uint256 amount)
func (_SubnetAppStore *SubnetAppStoreFilterer) WatchBudgetRefunded(opts *bind.WatchOpts, sink chan<- *SubnetAppStoreBudgetRefunded, appId []*big.Int) (event.Subscription, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}

	logs, sub, err := _SubnetAppStore.contract.WatchLogs(opts, "BudgetRefunded", appIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetAppStoreBudgetRefunded)
				if err := _SubnetAppStore.contract.UnpackLog(event, "BudgetRefunded", log); err != nil {
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

// ParseBudgetRefunded is a log parse operation binding the contract event 0x95da8a4f660594f653bc780752dadf78f4f089254077e243f1318f946465fd09.
//
// Solidity: event BudgetRefunded(uint256 indexed appId, uint256 amount)
func (_SubnetAppStore *SubnetAppStoreFilterer) ParseBudgetRefunded(log types.Log) (*SubnetAppStoreBudgetRefunded, error) {
	event := new(SubnetAppStoreBudgetRefunded)
	if err := _SubnetAppStore.contract.UnpackLog(event, "BudgetRefunded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetAppStoreDeploymentClosedIterator is returned from FilterDeploymentClosed and is used to iterate over the raw logs and unpacked data for DeploymentClosed events raised by the SubnetAppStore contract.
type SubnetAppStoreDeploymentClosedIterator struct {
	Event *SubnetAppStoreDeploymentClosed // Event containing the contract specifics and raw log

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
func (it *SubnetAppStoreDeploymentClosedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetAppStoreDeploymentClosed)
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
		it.Event = new(SubnetAppStoreDeploymentClosed)
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
func (it *SubnetAppStoreDeploymentClosedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetAppStoreDeploymentClosedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetAppStoreDeploymentClosed represents a DeploymentClosed event raised by the SubnetAppStore contract.
type SubnetAppStoreDeploymentClosed struct {
	ProviderId *big.Int
	AppId      *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterDeploymentClosed is a free log retrieval operation binding the contract event 0x8a42b8ff47c7c473ba08417a63c2a1660c2e6165052333e1d28d20687038c1c7.
//
// Solidity: event DeploymentClosed(uint256 indexed providerId, uint256 indexed appId)
func (_SubnetAppStore *SubnetAppStoreFilterer) FilterDeploymentClosed(opts *bind.FilterOpts, providerId []*big.Int, appId []*big.Int) (*SubnetAppStoreDeploymentClosedIterator, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}
	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}

	logs, sub, err := _SubnetAppStore.contract.FilterLogs(opts, "DeploymentClosed", providerIdRule, appIdRule)
	if err != nil {
		return nil, err
	}
	return &SubnetAppStoreDeploymentClosedIterator{contract: _SubnetAppStore.contract, event: "DeploymentClosed", logs: logs, sub: sub}, nil
}

// WatchDeploymentClosed is a free log subscription operation binding the contract event 0x8a42b8ff47c7c473ba08417a63c2a1660c2e6165052333e1d28d20687038c1c7.
//
// Solidity: event DeploymentClosed(uint256 indexed providerId, uint256 indexed appId)
func (_SubnetAppStore *SubnetAppStoreFilterer) WatchDeploymentClosed(opts *bind.WatchOpts, sink chan<- *SubnetAppStoreDeploymentClosed, providerId []*big.Int, appId []*big.Int) (event.Subscription, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}
	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}

	logs, sub, err := _SubnetAppStore.contract.WatchLogs(opts, "DeploymentClosed", providerIdRule, appIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetAppStoreDeploymentClosed)
				if err := _SubnetAppStore.contract.UnpackLog(event, "DeploymentClosed", log); err != nil {
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

// ParseDeploymentClosed is a log parse operation binding the contract event 0x8a42b8ff47c7c473ba08417a63c2a1660c2e6165052333e1d28d20687038c1c7.
//
// Solidity: event DeploymentClosed(uint256 indexed providerId, uint256 indexed appId)
func (_SubnetAppStore *SubnetAppStoreFilterer) ParseDeploymentClosed(log types.Log) (*SubnetAppStoreDeploymentClosed, error) {
	event := new(SubnetAppStoreDeploymentClosed)
	if err := _SubnetAppStore.contract.UnpackLog(event, "DeploymentClosed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetAppStoreDeploymentCreatedIterator is returned from FilterDeploymentCreated and is used to iterate over the raw logs and unpacked data for DeploymentCreated events raised by the SubnetAppStore contract.
type SubnetAppStoreDeploymentCreatedIterator struct {
	Event *SubnetAppStoreDeploymentCreated // Event containing the contract specifics and raw log

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
func (it *SubnetAppStoreDeploymentCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetAppStoreDeploymentCreated)
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
		it.Event = new(SubnetAppStoreDeploymentCreated)
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
func (it *SubnetAppStoreDeploymentCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetAppStoreDeploymentCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetAppStoreDeploymentCreated represents a DeploymentCreated event raised by the SubnetAppStore contract.
type SubnetAppStoreDeploymentCreated struct {
	ProviderId *big.Int
	AppId      *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterDeploymentCreated is a free log retrieval operation binding the contract event 0x074e70ca54f5e72d7ce248c31c6f8c17c6435d0e993cac6db3e345ca59c72488.
//
// Solidity: event DeploymentCreated(uint256 indexed providerId, uint256 indexed appId)
func (_SubnetAppStore *SubnetAppStoreFilterer) FilterDeploymentCreated(opts *bind.FilterOpts, providerId []*big.Int, appId []*big.Int) (*SubnetAppStoreDeploymentCreatedIterator, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}
	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}

	logs, sub, err := _SubnetAppStore.contract.FilterLogs(opts, "DeploymentCreated", providerIdRule, appIdRule)
	if err != nil {
		return nil, err
	}
	return &SubnetAppStoreDeploymentCreatedIterator{contract: _SubnetAppStore.contract, event: "DeploymentCreated", logs: logs, sub: sub}, nil
}

// WatchDeploymentCreated is a free log subscription operation binding the contract event 0x074e70ca54f5e72d7ce248c31c6f8c17c6435d0e993cac6db3e345ca59c72488.
//
// Solidity: event DeploymentCreated(uint256 indexed providerId, uint256 indexed appId)
func (_SubnetAppStore *SubnetAppStoreFilterer) WatchDeploymentCreated(opts *bind.WatchOpts, sink chan<- *SubnetAppStoreDeploymentCreated, providerId []*big.Int, appId []*big.Int) (event.Subscription, error) {

	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}
	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}

	logs, sub, err := _SubnetAppStore.contract.WatchLogs(opts, "DeploymentCreated", providerIdRule, appIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetAppStoreDeploymentCreated)
				if err := _SubnetAppStore.contract.UnpackLog(event, "DeploymentCreated", log); err != nil {
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

// ParseDeploymentCreated is a log parse operation binding the contract event 0x074e70ca54f5e72d7ce248c31c6f8c17c6435d0e993cac6db3e345ca59c72488.
//
// Solidity: event DeploymentCreated(uint256 indexed providerId, uint256 indexed appId)
func (_SubnetAppStore *SubnetAppStoreFilterer) ParseDeploymentCreated(log types.Log) (*SubnetAppStoreDeploymentCreated, error) {
	event := new(SubnetAppStoreDeploymentCreated)
	if err := _SubnetAppStore.contract.UnpackLog(event, "DeploymentCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetAppStoreEIP712DomainChangedIterator is returned from FilterEIP712DomainChanged and is used to iterate over the raw logs and unpacked data for EIP712DomainChanged events raised by the SubnetAppStore contract.
type SubnetAppStoreEIP712DomainChangedIterator struct {
	Event *SubnetAppStoreEIP712DomainChanged // Event containing the contract specifics and raw log

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
func (it *SubnetAppStoreEIP712DomainChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetAppStoreEIP712DomainChanged)
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
		it.Event = new(SubnetAppStoreEIP712DomainChanged)
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
func (it *SubnetAppStoreEIP712DomainChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetAppStoreEIP712DomainChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetAppStoreEIP712DomainChanged represents a EIP712DomainChanged event raised by the SubnetAppStore contract.
type SubnetAppStoreEIP712DomainChanged struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEIP712DomainChanged is a free log retrieval operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_SubnetAppStore *SubnetAppStoreFilterer) FilterEIP712DomainChanged(opts *bind.FilterOpts) (*SubnetAppStoreEIP712DomainChangedIterator, error) {

	logs, sub, err := _SubnetAppStore.contract.FilterLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return &SubnetAppStoreEIP712DomainChangedIterator{contract: _SubnetAppStore.contract, event: "EIP712DomainChanged", logs: logs, sub: sub}, nil
}

// WatchEIP712DomainChanged is a free log subscription operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_SubnetAppStore *SubnetAppStoreFilterer) WatchEIP712DomainChanged(opts *bind.WatchOpts, sink chan<- *SubnetAppStoreEIP712DomainChanged) (event.Subscription, error) {

	logs, sub, err := _SubnetAppStore.contract.WatchLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetAppStoreEIP712DomainChanged)
				if err := _SubnetAppStore.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
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
func (_SubnetAppStore *SubnetAppStoreFilterer) ParseEIP712DomainChanged(log types.Log) (*SubnetAppStoreEIP712DomainChanged, error) {
	event := new(SubnetAppStoreEIP712DomainChanged)
	if err := _SubnetAppStore.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetAppStoreInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the SubnetAppStore contract.
type SubnetAppStoreInitializedIterator struct {
	Event *SubnetAppStoreInitialized // Event containing the contract specifics and raw log

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
func (it *SubnetAppStoreInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetAppStoreInitialized)
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
		it.Event = new(SubnetAppStoreInitialized)
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
func (it *SubnetAppStoreInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetAppStoreInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetAppStoreInitialized represents a Initialized event raised by the SubnetAppStore contract.
type SubnetAppStoreInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_SubnetAppStore *SubnetAppStoreFilterer) FilterInitialized(opts *bind.FilterOpts) (*SubnetAppStoreInitializedIterator, error) {

	logs, sub, err := _SubnetAppStore.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &SubnetAppStoreInitializedIterator{contract: _SubnetAppStore.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_SubnetAppStore *SubnetAppStoreFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *SubnetAppStoreInitialized) (event.Subscription, error) {

	logs, sub, err := _SubnetAppStore.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetAppStoreInitialized)
				if err := _SubnetAppStore.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_SubnetAppStore *SubnetAppStoreFilterer) ParseInitialized(log types.Log) (*SubnetAppStoreInitialized, error) {
	event := new(SubnetAppStoreInitialized)
	if err := _SubnetAppStore.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetAppStoreOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the SubnetAppStore contract.
type SubnetAppStoreOwnershipTransferredIterator struct {
	Event *SubnetAppStoreOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *SubnetAppStoreOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetAppStoreOwnershipTransferred)
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
		it.Event = new(SubnetAppStoreOwnershipTransferred)
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
func (it *SubnetAppStoreOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetAppStoreOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetAppStoreOwnershipTransferred represents a OwnershipTransferred event raised by the SubnetAppStore contract.
type SubnetAppStoreOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SubnetAppStore *SubnetAppStoreFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*SubnetAppStoreOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SubnetAppStore.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &SubnetAppStoreOwnershipTransferredIterator{contract: _SubnetAppStore.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SubnetAppStore *SubnetAppStoreFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *SubnetAppStoreOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SubnetAppStore.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetAppStoreOwnershipTransferred)
				if err := _SubnetAppStore.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_SubnetAppStore *SubnetAppStoreFilterer) ParseOwnershipTransferred(log types.Log) (*SubnetAppStoreOwnershipTransferred, error) {
	event := new(SubnetAppStoreOwnershipTransferred)
	if err := _SubnetAppStore.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetAppStoreProviderRefundedIterator is returned from FilterProviderRefunded and is used to iterate over the raw logs and unpacked data for ProviderRefunded events raised by the SubnetAppStore contract.
type SubnetAppStoreProviderRefundedIterator struct {
	Event *SubnetAppStoreProviderRefunded // Event containing the contract specifics and raw log

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
func (it *SubnetAppStoreProviderRefundedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetAppStoreProviderRefunded)
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
		it.Event = new(SubnetAppStoreProviderRefunded)
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
func (it *SubnetAppStoreProviderRefundedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetAppStoreProviderRefundedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetAppStoreProviderRefunded represents a ProviderRefunded event raised by the SubnetAppStore contract.
type SubnetAppStoreProviderRefunded struct {
	AppId      *big.Int
	ProviderId *big.Int
	Amount     *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterProviderRefunded is a free log retrieval operation binding the contract event 0x968c359f1f36580744aa6aa2e36c2168a431decc06bb64cf7c4ad01b0a2477f6.
//
// Solidity: event ProviderRefunded(uint256 indexed appId, uint256 indexed providerId, uint256 amount)
func (_SubnetAppStore *SubnetAppStoreFilterer) FilterProviderRefunded(opts *bind.FilterOpts, appId []*big.Int, providerId []*big.Int) (*SubnetAppStoreProviderRefundedIterator, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _SubnetAppStore.contract.FilterLogs(opts, "ProviderRefunded", appIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return &SubnetAppStoreProviderRefundedIterator{contract: _SubnetAppStore.contract, event: "ProviderRefunded", logs: logs, sub: sub}, nil
}

// WatchProviderRefunded is a free log subscription operation binding the contract event 0x968c359f1f36580744aa6aa2e36c2168a431decc06bb64cf7c4ad01b0a2477f6.
//
// Solidity: event ProviderRefunded(uint256 indexed appId, uint256 indexed providerId, uint256 amount)
func (_SubnetAppStore *SubnetAppStoreFilterer) WatchProviderRefunded(opts *bind.WatchOpts, sink chan<- *SubnetAppStoreProviderRefunded, appId []*big.Int, providerId []*big.Int) (event.Subscription, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _SubnetAppStore.contract.WatchLogs(opts, "ProviderRefunded", appIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetAppStoreProviderRefunded)
				if err := _SubnetAppStore.contract.UnpackLog(event, "ProviderRefunded", log); err != nil {
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

// ParseProviderRefunded is a log parse operation binding the contract event 0x968c359f1f36580744aa6aa2e36c2168a431decc06bb64cf7c4ad01b0a2477f6.
//
// Solidity: event ProviderRefunded(uint256 indexed appId, uint256 indexed providerId, uint256 amount)
func (_SubnetAppStore *SubnetAppStoreFilterer) ParseProviderRefunded(log types.Log) (*SubnetAppStoreProviderRefunded, error) {
	event := new(SubnetAppStoreProviderRefunded)
	if err := _SubnetAppStore.contract.UnpackLog(event, "ProviderRefunded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetAppStoreRewardClaimedIterator is returned from FilterRewardClaimed and is used to iterate over the raw logs and unpacked data for RewardClaimed events raised by the SubnetAppStore contract.
type SubnetAppStoreRewardClaimedIterator struct {
	Event *SubnetAppStoreRewardClaimed // Event containing the contract specifics and raw log

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
func (it *SubnetAppStoreRewardClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetAppStoreRewardClaimed)
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
		it.Event = new(SubnetAppStoreRewardClaimed)
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
func (it *SubnetAppStoreRewardClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetAppStoreRewardClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetAppStoreRewardClaimed represents a RewardClaimed event raised by the SubnetAppStore contract.
type SubnetAppStoreRewardClaimed struct {
	AppId      *big.Int
	ProviderId *big.Int
	Reward     *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterRewardClaimed is a free log retrieval operation binding the contract event 0x64bc8e516eb66c9bd33c0d0f80b0e16a5bac9eb2bce20d6fdf7698a0d7ab6eba.
//
// Solidity: event RewardClaimed(uint256 indexed appId, uint256 indexed providerId, uint256 reward)
func (_SubnetAppStore *SubnetAppStoreFilterer) FilterRewardClaimed(opts *bind.FilterOpts, appId []*big.Int, providerId []*big.Int) (*SubnetAppStoreRewardClaimedIterator, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _SubnetAppStore.contract.FilterLogs(opts, "RewardClaimed", appIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return &SubnetAppStoreRewardClaimedIterator{contract: _SubnetAppStore.contract, event: "RewardClaimed", logs: logs, sub: sub}, nil
}

// WatchRewardClaimed is a free log subscription operation binding the contract event 0x64bc8e516eb66c9bd33c0d0f80b0e16a5bac9eb2bce20d6fdf7698a0d7ab6eba.
//
// Solidity: event RewardClaimed(uint256 indexed appId, uint256 indexed providerId, uint256 reward)
func (_SubnetAppStore *SubnetAppStoreFilterer) WatchRewardClaimed(opts *bind.WatchOpts, sink chan<- *SubnetAppStoreRewardClaimed, appId []*big.Int, providerId []*big.Int) (event.Subscription, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _SubnetAppStore.contract.WatchLogs(opts, "RewardClaimed", appIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetAppStoreRewardClaimed)
				if err := _SubnetAppStore.contract.UnpackLog(event, "RewardClaimed", log); err != nil {
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

// ParseRewardClaimed is a log parse operation binding the contract event 0x64bc8e516eb66c9bd33c0d0f80b0e16a5bac9eb2bce20d6fdf7698a0d7ab6eba.
//
// Solidity: event RewardClaimed(uint256 indexed appId, uint256 indexed providerId, uint256 reward)
func (_SubnetAppStore *SubnetAppStoreFilterer) ParseRewardClaimed(log types.Log) (*SubnetAppStoreRewardClaimed, error) {
	event := new(SubnetAppStoreRewardClaimed)
	if err := _SubnetAppStore.contract.UnpackLog(event, "RewardClaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetAppStoreUsageReportedIterator is returned from FilterUsageReported and is used to iterate over the raw logs and unpacked data for UsageReported events raised by the SubnetAppStore contract.
type SubnetAppStoreUsageReportedIterator struct {
	Event *SubnetAppStoreUsageReported // Event containing the contract specifics and raw log

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
func (it *SubnetAppStoreUsageReportedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetAppStoreUsageReported)
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
		it.Event = new(SubnetAppStoreUsageReported)
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
func (it *SubnetAppStoreUsageReportedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetAppStoreUsageReportedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetAppStoreUsageReported represents a UsageReported event raised by the SubnetAppStore contract.
type SubnetAppStoreUsageReported struct {
	AppId             *big.Int
	ProviderId        *big.Int
	PeerId            string
	UsedCpu           *big.Int
	UsedGpu           *big.Int
	UsedMemory        *big.Int
	UsedStorage       *big.Int
	UsedUploadBytes   *big.Int
	UsedDownloadBytes *big.Int
	Duration          *big.Int
	Timestamp         *big.Int
	Reward            *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterUsageReported is a free log retrieval operation binding the contract event 0xfe8684dcefddcc03bc46e198108fa6b7b15a09efe6e1935975111e9bfdf5deec.
//
// Solidity: event UsageReported(uint256 indexed appId, uint256 indexed providerId, string peerId, uint256 usedCpu, uint256 usedGpu, uint256 usedMemory, uint256 usedStorage, uint256 usedUploadBytes, uint256 usedDownloadBytes, uint256 duration, uint256 timestamp, uint256 reward)
func (_SubnetAppStore *SubnetAppStoreFilterer) FilterUsageReported(opts *bind.FilterOpts, appId []*big.Int, providerId []*big.Int) (*SubnetAppStoreUsageReportedIterator, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _SubnetAppStore.contract.FilterLogs(opts, "UsageReported", appIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return &SubnetAppStoreUsageReportedIterator{contract: _SubnetAppStore.contract, event: "UsageReported", logs: logs, sub: sub}, nil
}

// WatchUsageReported is a free log subscription operation binding the contract event 0xfe8684dcefddcc03bc46e198108fa6b7b15a09efe6e1935975111e9bfdf5deec.
//
// Solidity: event UsageReported(uint256 indexed appId, uint256 indexed providerId, string peerId, uint256 usedCpu, uint256 usedGpu, uint256 usedMemory, uint256 usedStorage, uint256 usedUploadBytes, uint256 usedDownloadBytes, uint256 duration, uint256 timestamp, uint256 reward)
func (_SubnetAppStore *SubnetAppStoreFilterer) WatchUsageReported(opts *bind.WatchOpts, sink chan<- *SubnetAppStoreUsageReported, appId []*big.Int, providerId []*big.Int) (event.Subscription, error) {

	var appIdRule []interface{}
	for _, appIdItem := range appId {
		appIdRule = append(appIdRule, appIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _SubnetAppStore.contract.WatchLogs(opts, "UsageReported", appIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetAppStoreUsageReported)
				if err := _SubnetAppStore.contract.UnpackLog(event, "UsageReported", log); err != nil {
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

// ParseUsageReported is a log parse operation binding the contract event 0xfe8684dcefddcc03bc46e198108fa6b7b15a09efe6e1935975111e9bfdf5deec.
//
// Solidity: event UsageReported(uint256 indexed appId, uint256 indexed providerId, string peerId, uint256 usedCpu, uint256 usedGpu, uint256 usedMemory, uint256 usedStorage, uint256 usedUploadBytes, uint256 usedDownloadBytes, uint256 duration, uint256 timestamp, uint256 reward)
func (_SubnetAppStore *SubnetAppStoreFilterer) ParseUsageReported(log types.Log) (*SubnetAppStoreUsageReported, error) {
	event := new(SubnetAppStoreUsageReported)
	if err := _SubnetAppStore.contract.UnpackLog(event, "UsageReported", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetAppStoreVerifierRewardClaimedIterator is returned from FilterVerifierRewardClaimed and is used to iterate over the raw logs and unpacked data for VerifierRewardClaimed events raised by the SubnetAppStore contract.
type SubnetAppStoreVerifierRewardClaimedIterator struct {
	Event *SubnetAppStoreVerifierRewardClaimed // Event containing the contract specifics and raw log

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
func (it *SubnetAppStoreVerifierRewardClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetAppStoreVerifierRewardClaimed)
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
		it.Event = new(SubnetAppStoreVerifierRewardClaimed)
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
func (it *SubnetAppStoreVerifierRewardClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetAppStoreVerifierRewardClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetAppStoreVerifierRewardClaimed represents a VerifierRewardClaimed event raised by the SubnetAppStore contract.
type SubnetAppStoreVerifierRewardClaimed struct {
	Verifier common.Address
	Reward   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterVerifierRewardClaimed is a free log retrieval operation binding the contract event 0x22a6f57628ecf1ae33f45f9014c841f197327963742bbe8fa04b104b41a655c2.
//
// Solidity: event VerifierRewardClaimed(address indexed verifier, uint256 reward)
func (_SubnetAppStore *SubnetAppStoreFilterer) FilterVerifierRewardClaimed(opts *bind.FilterOpts, verifier []common.Address) (*SubnetAppStoreVerifierRewardClaimedIterator, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _SubnetAppStore.contract.FilterLogs(opts, "VerifierRewardClaimed", verifierRule)
	if err != nil {
		return nil, err
	}
	return &SubnetAppStoreVerifierRewardClaimedIterator{contract: _SubnetAppStore.contract, event: "VerifierRewardClaimed", logs: logs, sub: sub}, nil
}

// WatchVerifierRewardClaimed is a free log subscription operation binding the contract event 0x22a6f57628ecf1ae33f45f9014c841f197327963742bbe8fa04b104b41a655c2.
//
// Solidity: event VerifierRewardClaimed(address indexed verifier, uint256 reward)
func (_SubnetAppStore *SubnetAppStoreFilterer) WatchVerifierRewardClaimed(opts *bind.WatchOpts, sink chan<- *SubnetAppStoreVerifierRewardClaimed, verifier []common.Address) (event.Subscription, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _SubnetAppStore.contract.WatchLogs(opts, "VerifierRewardClaimed", verifierRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetAppStoreVerifierRewardClaimed)
				if err := _SubnetAppStore.contract.UnpackLog(event, "VerifierRewardClaimed", log); err != nil {
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

// ParseVerifierRewardClaimed is a log parse operation binding the contract event 0x22a6f57628ecf1ae33f45f9014c841f197327963742bbe8fa04b104b41a655c2.
//
// Solidity: event VerifierRewardClaimed(address indexed verifier, uint256 reward)
func (_SubnetAppStore *SubnetAppStoreFilterer) ParseVerifierRewardClaimed(log types.Log) (*SubnetAppStoreVerifierRewardClaimed, error) {
	event := new(SubnetAppStoreVerifierRewardClaimed)
	if err := _SubnetAppStore.contract.UnpackLog(event, "VerifierRewardClaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
