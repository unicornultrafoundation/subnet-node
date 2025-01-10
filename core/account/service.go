package account

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
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
	privateKey               *ecdsa.PrivateKey
	client                   *ethclient.Client
	chainID                  *big.Int
	subnetRegistry           *contracts.SubnetRegistry
	subnetRegistryAddress    string
	subnetAppRegistry        *contracts.SubnetAppRegistry
	subnetAppRegistryAddress string
	nftLicense               *contracts.ERC721
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

	subnetAppRegistryAddress := cfg.GetString("apps.subnet_app_registry_contract", config.DefaultSubnetAppRegistryContract)
	subnetAppRegistry, err := contracts.NewSubnetAppRegistry(
		common.HexToAddress(subnetAppRegistryAddress),
		client,
	)

	if err != nil {
		return nil, err
	}

	subnetRegistryAddress := cfg.GetString("apps.subnet_registry_contract", config.DefaultSubnetRegistryCOntract)
	subnetRegistry, err := contracts.NewSubnetRegistry(
		common.HexToAddress(subnetRegistryAddress),
		client,
	)

	if err != nil {
		return nil, err
	}

	nftAddr, err := subnetRegistry.NftContract(nil)

	if err != nil {
		return nil, err
	}

	nftLicense, err := contracts.NewERC721(nftAddr, client)

	if err != nil {
		return nil, err
	}

	s := &AccountService{
		privateKey:               privateKey,
		client:                   client,
		chainID:                  chainID,
		subnetRegistry:           subnetRegistry,
		subnetRegistryAddress:    subnetRegistryAddress,
		subnetAppRegistry:        subnetAppRegistry,
		subnetAppRegistryAddress: subnetAppRegistryAddress,
		nftLicense:               nftLicense,
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

func (s *AccountService) NftLicense() *contracts.ERC721 {
	return s.nftLicense
}

func (s *AccountService) GetChainID() *big.Int {
	return s.chainID
}

func (s *AccountService) GetSubnetRegistryAddress() string {
	return s.subnetRegistryAddress
}

func (s *AccountService) GetSubnetAppRegistryAddress() string {
	return s.subnetAppRegistryAddress
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

// Sign the hash using ECDSA
func (account *AccountService) Sign(hash []byte) ([]byte, error) {
	r, s, err := ecdsa.Sign(rand.Reader, account.privateKey, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %v", err)
	}

	// Ensure `s` is in the lower half-order
	halfOrder := new(big.Int).Rsh(crypto.S256().Params().N, 1)
	v := byte(27)
	if s.Cmp(halfOrder) > 0 {
		s.Sub(crypto.S256().Params().N, s)
		v = 28
	}

	// Format signature: r (32 bytes) || s (32 bytes) || v (1 byte)
	signature := append(r.Bytes(), append(s.Bytes(), v)...)

	return signature, nil
}
