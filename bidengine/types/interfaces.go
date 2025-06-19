package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/holiman/uint256"
	"github.com/unicornultrafoundation/subnet-node/bidengine/contracts"
)

// ProviderService defines the interface for interacting with the Provider contract
type ProviderService interface {
	// IsMachineActive checks if a machine is active
	IsMachineActive(opts *bind.CallOpts, providerId, machineId *big.Int) (bool, error)

	// ValidateMachineRequirements checks if a machine meets the specified requirements
	ValidateMachineRequirements(opts *bind.CallOpts, machineType, providerId, machineId *big.Int, minCPUCores, minMemoryMB, minDiskGB, minGPUCores, minUploadSpeed, minDownloadSpeed *big.Int) (bool, error)

	// GetMachine retrieves detailed information about a specific machine
	GetMachine(opts *bind.CallOpts, providerId, machineId *uint256.Int) (*Machine, error)

	// GetProvider retrieves provider information
	GetProvider(opts *bind.CallOpts, providerId *uint256.Int) (*Provider, error)

	// GetMachinesPaginated retrieves a paginated list of machines for a provider
	GetMachinesPaginated(opts *bind.CallOpts, providerId *uint256.Int, startIndex, endIndex *uint256.Int) ([]*Machine, error)
}

// MarketService defines the interface for interacting with the BidMarket contract
type MarketService interface {
	// GetOrder retrieves order details from the blockchain
	GetOrder(opts *bind.CallOpts, orderId *uint256.Int) (*OrderInfo, error)

	// GetOrderInfo retrieves order information (alias for GetOrder)
	GetOrderInfo(opts *bind.CallOpts, orderId *uint256.Int) (*OrderInfo, error)

	// GetBids retrieves all bids for a specific order
	GetBids(opts *bind.CallOpts, orderId *uint256.Int) ([]*Bid, error)

	SubmitBid(auth *bind.TransactOpts, orderId, providerId, machineId, pricePerSecond *uint256.Int) (*types.Transaction, error)

	// WatchOrderCreated subscribes to the OrderCreated event
	WatchOrderCreated(opts *bind.WatchOpts, ch chan<- *contracts.BidMarketOrderCreated) (event.Subscription, error)

	// WatchBidAccepted subscribes to the BidAccepted event
	WatchBidAccepted(opts *bind.WatchOpts, ch chan<- *contracts.BidMarketBidAccepted, providerId *uint256.Int) (event.Subscription, error)
}
