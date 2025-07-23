package account

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	sinit "github.com/unicornultrafoundation/subnet-node/cmd/init"
	signer "github.com/unicornultrafoundation/subnet-node/common/signer"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/contracts"
	"github.com/unicornultrafoundation/subnet-node/repo"
	"go.uber.org/fx"
)

var log = logrus.WithField("service", "account")

// AccountService is a service to handle Ethereum transactions
type AccountService struct {
	ks                   *keystore.KeyStore
	client               *ethclient.Client
	chainID              *big.Int
	subnetProvider       *contracts.SubnetProvider
	subnetProviderAddr   string
	subnetIPRegistry     IPRegistry
	subnetIPRegistryAddr string
}

// NewAccountService initializes a new AccountService
func NewAccountService(cfg *config.C, ks *keystore.KeyStore) (*AccountService, error) {
	rpcURL := cfg.GetString("account.rpc", config.DefaultRPC)
	chainID := big.NewInt(int64(cfg.GetInt("account.chainid", config.DefaultChainID)))

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	subnetIPRegistryAddr := cfg.GetString("apps.subnet_ip_registry", config.DefaultSubnetIP)
	subnetIPRegistry, err := contracts.NewSubnetIPRegistry(
		common.HexToAddress(subnetIPRegistryAddr),
		client,
	)
	if err != nil {
		return nil, err
	}

	var ipRegistry IPRegistry = subnetIPRegistry

	accounts := ks.Accounts()
	password := cfg.GetString("account.password", os.Getenv("ACCOUNT_PASSWORD"))
	if len(accounts) == 0 {
		account, _, _ := sinit.CreateKeystoreAccount(ks, password, "Enter keystore password: ")
		accounts = append(accounts, account)
	}
	if err := ks.Unlock(accounts[0], password); err != nil {
		return nil, fmt.Errorf("failed to unlock keystore: %w", err)
	}

	s := &AccountService{
		ks:                   ks,
		client:               client,
		chainID:              chainID,
		subnetIPRegistry:     ipRegistry,
		subnetIPRegistryAddr: subnetIPRegistryAddr,
	}
	s.registerReloadCallback(cfg)
	return s, nil
}

func (s *AccountService) registerReloadCallback(cfg *config.C) {
	cfg.RegisterReloadCallback(func(cfg *config.C) {
		// No reload for keystore password from config; password is from env
	})
}

// GetClient retrieves the ethclient instance
func (s *AccountService) GetClient() *ethclient.Client {
	return s.client
}

func (s *AccountService) Provider() *contracts.SubnetProvider {
	return s.subnetProvider
}

func (s *AccountService) IPRegistry() IPRegistry {
	return s.subnetIPRegistry
}

func (s *AccountService) GetChainID() *big.Int {
	return s.chainID
}

func (s *AccountService) ProviderAddr() string {
	return s.subnetProviderAddr
}

func (s *AccountService) IPRegistryAddr() string {
	return s.subnetIPRegistryAddr
}

// GetAddress retrieves the Ethereum address from the private key
func (s *AccountService) GetAddress() common.Address {
	accounts := s.ks.Accounts()
	if len(accounts) == 0 {
		return common.Address{}
	}

	return accounts[0].Address
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
	accounts := s.ks.Accounts()
	if len(accounts) == 0 {
		return nil, fmt.Errorf("no accounts found")
	}

	keyAddr := accounts[0].Address

	signer := types.LatestSignerForChainID(s.chainID)
	return &bind.TransactOpts{
		From: accounts[0].Address,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != keyAddr {
				return nil, fmt.Errorf("not authorized to sign transaction from address %s", address.Hex())
			}
			signature, err := s.ks.SignHash(accounts[0], signer.Hash(tx).Bytes())
			if err != nil {
				return nil, err
			}
			return tx.WithSignature(signer, signature)
		},
		Context: context.Background(),
	}, nil
}

// SignAndSendTransaction creates, signs, and sends a transaction
func (s *AccountService) SignAndSendTransaction(toAddress string, value *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) (string, error) {
	accounts := s.ks.Accounts()
	if len(accounts) == 0 {
		return "", fmt.Errorf("no accounts found")
	}

	account := accounts[0]

	nonce, err := s.client.PendingNonceAt(context.Background(), account.Address)
	if err != nil {
		return "", err
	}

	// Create a new transaction
	tx := types.NewTransaction(nonce, common.HexToAddress(toAddress), value, gasLimit, gasPrice, data)

	signedTx, err := s.ks.SignTx(account, tx, s.chainID)

	// Send the transaction
	err = s.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	return signedTx.Hash().Hex(), nil
}

// EthereumService provides a lifecycle-managed Ethereum service
func EthereumService(lc fx.Lifecycle, cfg *config.C, rp repo.Repo) (*AccountService, error) {
	repoPath := filepath.Clean(rp.Path())
	ksDir := filepath.Join(repoPath, "keystore")
	ks := keystore.NewKeyStore(ksDir, keystore.StandardScryptN, keystore.StandardScryptP)

	service, err := NewAccountService(cfg, ks)
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

	accounts := account.ks.Accounts()
	if len(accounts) == 0 {
		return nil, fmt.Errorf("no accounts found")
	}
	acc := accounts[0]

	signature, err := account.ks.SignHash(acc, hash)
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
