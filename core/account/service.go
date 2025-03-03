package account

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	signer "github.com/unicornultrafoundation/subnet-node/common/signer"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/contracts"
	"go.uber.org/fx"
)

var log = logrus.WithField("service", "account")

// AccountService is a service to handle Ethereum transactions
type AccountService struct {
	privateKey         *ecdsa.PrivateKey
	client             *ethclient.Client
	chainID            *big.Int
	subnetProvider     *contracts.SubnetProvider
	subnetProviderAddr string
	subnetAppStore     *contracts.SubnetAppStore
	subnetAppStoreAddr string
	providerID         int64
}

// NewAccountService initializes a new AccountService
func NewAccountService(cfg *config.C) (*AccountService, error) {
	privateKeyHex := cfg.GetString("account.private_key", "")
	rpcURL := cfg.GetString("account.rpc", config.DefaultRPC)
	chainID := big.NewInt(int64(cfg.GetInt("account.chainid", config.DefaultChainID)))
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	subnetAppStoreAddr := cfg.GetString("apps.subnet_app_store", config.DefaultSubnetAppStoreAddr)
	subnetAppStore, err := contracts.NewSubnetAppStore(
		common.HexToAddress(subnetAppStoreAddr),
		client,
	)

	if err != nil {
		return nil, err
	}

	subnetProviderAddr := cfg.GetString("apps.subnet_provider", config.DefaultSubnetProviderAddr)
	subnetRegistry, err := contracts.NewSubnetProvider(
		common.HexToAddress(subnetProviderAddr),
		client,
	)

	if err != nil {
		return nil, err
	}

	s := &AccountService{
		privateKey:         privateKey,
		client:             client,
		chainID:            chainID,
		subnetProvider:     subnetRegistry,
		subnetProviderAddr: subnetProviderAddr,
		subnetAppStore:     subnetAppStore,
		subnetAppStoreAddr: subnetAppStoreAddr,
	}
	s.updateProviderID(cfg)
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

		if cfg.HasChanged("provider.id") {
			s.updateProviderID(cfg)
		}
	})
}

func (s *AccountService) updateProviderID(cfg *config.C) {
	providerIdHex := cfg.GetString("provider.id", "")
	if providerIdHex != "" {
		s.providerID = int64(hexutil.MustDecodeUint64(providerIdHex))
	}
}

// GetClient retrieves the ethclient instance
func (s *AccountService) GetClient() *ethclient.Client {
	return s.client
}

func (s *AccountService) Provider() *contracts.SubnetProvider {
	return s.subnetProvider
}

func (s *AccountService) AppStore() *contracts.SubnetAppStore {
	return s.subnetAppStore
}

func (s *AccountService) GetChainID() *big.Int {
	return s.chainID
}

func (s *AccountService) AppStoreAddr() string {
	return s.subnetAppStoreAddr
}

func (s *AccountService) ProviderAddr() string {
	return s.subnetProviderAddr
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

func (s *AccountService) ProviderID() int64 {
	return s.providerID
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

// Sign the hash using ECDSA
func (account *AccountService) Sign(hash []byte) ([]byte, error) {
	signature, err := crypto.Sign(hash, account.privateKey)
	if err != nil {
		return nil, err
	}
	signature[64] += 27
	return signature, nil
}

func (account *AccountService) SignTypedData(typedData *signer.TypedData) ([]byte, []byte, error) {
	typedDataHash, _, err := signer.TypedDataAndHash(*typedData)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("failed to hash typed data: %v", err)
	}
	s, e := account.Sign(typedDataHash)
	return typedDataHash, s, e
}
