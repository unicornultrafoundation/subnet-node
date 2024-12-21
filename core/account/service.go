package account

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"go.uber.org/fx"
)

var log = logrus.New().WithField("service", "account")

// AccountService is a service to handle Ethereum transactions
type AccountService struct {
	privateKey *ecdsa.PrivateKey
	client     *ethclient.Client
	chainID    *big.Int
}

// NewAccountService initializes a new AccountService
func NewAccountService(privateKeyHex, rpcURL string, chainID *big.Int) (*AccountService, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	return &AccountService{
		privateKey: privateKey,
		client:     client,
		chainID:    chainID,
	}, nil
}

// GetClient retrieves the ethclient instance
func (s *AccountService) GetClient() *ethclient.Client {
	return s.client
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
	service, err := NewAccountService(cfg.GetString("account.private_key", ""), cfg.GetString("account.rpc", config.DefaultRPC), big.NewInt(int64(cfg.GetInt("account.chainid", 2484))))
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
