package account

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/unicornultrafoundation/subnet-node/core/contracts"
)

// Service is the interface for the account service
type Service interface {
	// GetClient retrieves the ethclient instance
	GetClient() *ethclient.Client
	// Provider returns the subnet provider contract
	Provider() *contracts.SubnetProvider
	// AppStore returns the subnet app store contract
	AppStore() *contracts.SubnetAppStore
	// IPRegistry returns the subnet IP registry contract
	IPRegistry() IPRegistry
	// GetChainID returns the chain ID
	GetChainID() *big.Int
	// AppStoreAddr returns the subnet app store address
	AppStoreAddr() string
	// ProviderAddr returns the subnet provider address
	ProviderAddr() string
	// IPRegistryAddr returns the subnet IP registry address
	IPRegistryAddr() string
	// GetAddress retrieves the Ethereum address from the private key
	GetAddress() common.Address
	// GetBalance retrieves the Ether balance of the account
	GetBalance(address common.Address) (*big.Int, error)
	// NewKeyedTransactor returns a new keyed transactor
	NewKeyedTransactor() (*bind.TransactOpts, error)
	// ProviderID returns the provider ID
	ProviderID() int64
	// SignAndSendTransaction creates, signs, and sends a transaction
	SignAndSendTransaction(toAddress string, value *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) (string, error)
	// Sign signs the hash using ECDSA
	Sign(hash []byte) ([]byte, error)
}

// IPRegistry is the interface for the IP registry contract
type IPRegistry interface {
	// GetPeer gets the peer ID for a token ID
	GetPeer(opts *bind.CallOpts, tokenID *big.Int) (string, error)
}
