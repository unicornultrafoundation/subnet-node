package account

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/contracts"
	"go.uber.org/fx"
)

var log = logrus.New().WithField("service", "account")

// AccountService is a service to handle Ethereum transactions
type AccountService struct {
	privateKey        *ecdsa.PrivateKey
	client            *ethclient.Client
	chainID           *big.Int
	subnetRegistry    *contracts.SubnetRegistry
	subnetAppRegistry *contracts.SubnetAppRegistry
}

// NewAccountService initializes a new AccountService
func NewAccountService(cfg *config.C) (*AccountService, error) {
	privateKeyHex := cfg.GetString("account.private_key", "")
	rpcURL := cfg.GetString("account.rpc", config.DefaultRPC)
	chainID := big.NewInt(int64(cfg.GetInt("account.chainid", 2484)))
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	subnetAppRegistry, err := contracts.NewSubnetAppRegistry(
		common.HexToAddress(cfg.GetString("apps.subnet_app_registry_contract", config.DefaultSubnetAppRegistryContract)),
		client,
	)

	if err != nil {
		return nil, err
	}

	subnetRegistry, err := contracts.NewSubnetRegistry(
		common.HexToAddress(cfg.GetString("apps.subnet_registry_contract", config.DefaultSubnetRegistryCOntract)),
		client,
	)

	if err != nil {
		return nil, err
	}

	s := &AccountService{
		privateKey:        privateKey,
		client:            client,
		chainID:           chainID,
		subnetRegistry:    subnetRegistry,
		subnetAppRegistry: subnetAppRegistry,
	}
	s.registerReloadCallback(cfg)
	return s, nil
}

func (s *AccountService) registerReloadCallback(cfg *config.C) {
	cfg.RegisterReloadCallback(func(cfg *config.C) {
		if cfg.HasChanged("account.private_key") {
			privateKeyHex := cfg.GetString("account.private_key", "")
			privateKey, err := crypto.HexToECDSA(privateKeyHex)
			if err != nil {
				log.WithError(err).Error("failed to reload private key")
				return
			}
			s.privateKey = privateKey
		}
	})
}

// GetClient retrieves the ethclient instance
func (s *AccountService) GetClient() *ethclient.Client {
	return s.client
}

func (s *AccountService) SubnetRegistry() *contracts.SubnetRegistry {
	return s.subnetRegistry
}

func (s *AccountService) SubnetAppRegistry() *contracts.SubnetAppRegistry {
	return s.subnetAppRegistry
}

// GetAddress retrieves the Ethereum address from the private key
func (s *AccountService) GetAddress() common.Address {
	publicKey := s.privateKey.Public().(*ecdsa.PublicKey)
	return crypto.PubkeyToAddress(*publicKey)
}

// GetBalance retrieves the Ether balance of the account
func (s *AccountService) GetBalance(address common.Address) (*big.Int, error) {
	balance, err := s.client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (s *AccountService) NewKeyedTransactor() (*bind.TransactOpts, error) {
	return bind.NewKeyedTransactorWithChainID(s.privateKey, s.chainID)
}

// SignAndSendTransaction creates, signs, and sends a transaction
func (s *AccountService) SignAndSendTransaction(toAddress string, value *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) (string, error) {
	// Derive the sender address from the private key
	publicKey := s.privateKey.Public().(*ecdsa.PublicKey)
	senderAddress := crypto.PubkeyToAddress(*publicKey)

	nonce, err := s.client.PendingNonceAt(context.Background(), senderAddress)
	if err != nil {
		return "", err
	}

	// Create a new transaction
	tx := types.NewTransaction(nonce, common.HexToAddress(toAddress), value, gasLimit, gasPrice, data)

	// Sign the transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(s.chainID), s.privateKey)
	if err != nil {
		return "", err
	}

	// Send the transaction
	err = s.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	return signedTx.Hash().Hex(), nil
}

// EthereumService provides a lifecycle-managed Ethereum service
func EthereumService(lc fx.Lifecycle, cfg *config.C) (*AccountService, error) {
	service, err := NewAccountService(cfg)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println("EthereumService started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("EthereumService stopped")
			return nil
		},
	})

	return service, nil
}
