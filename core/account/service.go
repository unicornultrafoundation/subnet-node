package account

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
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
	mu                   sync.RWMutex
	ks                   *keystore.KeyStore
	client               *ethclient.Client
	chainID              *big.Int
	subnetProvider       *contracts.SubnetProvider
	subnetProviderAddr   string
	subnetIPRegistry     IPRegistry
	subnetIPRegistryAddr string
	cfg                  *config.C

	// Cached values for performance
	cachedAccount *accounts.Account
	cachedAddress *common.Address
	cachedSigner  types.Signer

	// Connection pool settings
	clientTimeout time.Duration
	maxRetries    int

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// ServiceConfig holds configuration for the AccountService
type ServiceConfig struct {
	RPCURL               string
	ChainID              int64
	SubnetIPRegistryAddr string
	SubnetProviderAddr   string
	Password             string
	ClientTimeout        time.Duration
	MaxRetries           int
}

// readPasswordFromFile reads password from a file if the path is provided
func readPasswordFromFile(passwordPath string) (string, error) {
	if passwordPath == "" {
		return "", nil
	}

	// Check if it's a file path
	if _, err := os.Stat(passwordPath); err == nil {
		// File exists, read password from it
		passwordBytes, err := os.ReadFile(passwordPath)
		if err != nil {
			return "", fmt.Errorf("failed to read password file %s: %w", passwordPath, err)
		}
		// Trim whitespace and newlines from password
		return strings.TrimSpace(string(passwordBytes)), nil
	}

	// Not a file path, return as is
	return passwordPath, nil
}

// NewServiceConfig creates a new ServiceConfig from the application config
func NewServiceConfig(cfg *config.C) (*ServiceConfig, error) {
	passwordPath := cfg.GetString("password", os.Getenv("ACCOUNT_PASSWORD"))
	password, err := readPasswordFromFile(passwordPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read password: %w", err)
	}

	sc := &ServiceConfig{
		RPCURL:               cfg.GetString("account.rpc", config.DefaultRPC),
		ChainID:              int64(cfg.GetInt("account.chainid", config.DefaultChainID)),
		SubnetIPRegistryAddr: cfg.GetString("apps.subnet_ip_registry", config.DefaultSubnetIP),
		SubnetProviderAddr:   cfg.GetString("apps.subnet_provider", config.DefaultSubnetProviderAddr),
		Password:             password,
		ClientTimeout:        cfg.GetDuration("account.client_timeout", 30*time.Second),
		MaxRetries:           cfg.GetInt("account.max_retries", 3),
	}

	// Validate configuration
	if sc.RPCURL == "" {
		return nil, fmt.Errorf("RPC URL cannot be empty")
	}
	if sc.ChainID <= 0 {
		return nil, fmt.Errorf("invalid chain ID: %d", sc.ChainID)
	}
	if sc.SubnetIPRegistryAddr == "" {
		return nil, fmt.Errorf("subnet IP registry address cannot be empty")
	}
	if sc.Password == "" {
		return nil, fmt.Errorf("account password cannot be empty")
	}

	return sc, nil
}

// NewAccountService initializes a new AccountService
func NewAccountService(cfg *config.C, ks *keystore.KeyStore) (*AccountService, error) {
	serviceConfig, err := NewServiceConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid service configuration: %w", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Create ethclient with timeout
	client, err := ethclient.DialContext(ctx, serviceConfig.RPCURL)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to RPC endpoint %s: %w", serviceConfig.RPCURL, err)
	}

	// Validate chain ID
	chainID := big.NewInt(serviceConfig.ChainID)
	networkChainID, err := client.ChainID(ctx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get network chain ID: %w", err)
	}
	if chainID.Cmp(networkChainID) != 0 {
		log.Warnf("Configured chain ID (%s) differs from network chain ID (%s)", chainID.String(), networkChainID.String())
	}

	// Initialize IP registry contract
	subnetIPRegistry, err := contracts.NewSubnetIPRegistry(
		common.HexToAddress(serviceConfig.SubnetIPRegistryAddr),
		client,
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize IP registry contract: %w", err)
	}

	// Initialize provider contract if address is provided
	var subnetProvider *contracts.SubnetProvider
	if serviceConfig.SubnetProviderAddr != "" {
		subnetProvider, err = contracts.NewSubnetProvider(
			common.HexToAddress(serviceConfig.SubnetProviderAddr),
			client,
		)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize provider contract: %w", err)
		}
	}

	// Ensure keystore has at least one account
	if len(ks.Accounts()) == 0 {
		_, _, err := sinit.CreateKeystoreAccount(ks, serviceConfig.Password, "Enter keystore password: ")
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to create keystore account: %w", err)
		}
	}

	s := &AccountService{
		ks:                   ks,
		client:               client,
		chainID:              chainID,
		subnetProvider:       subnetProvider,
		subnetProviderAddr:   serviceConfig.SubnetProviderAddr,
		subnetIPRegistry:     subnetIPRegistry,
		subnetIPRegistryAddr: serviceConfig.SubnetIPRegistryAddr,
		cfg:                  cfg,
		clientTimeout:        serviceConfig.ClientTimeout,
		maxRetries:           serviceConfig.MaxRetries,
		ctx:                  ctx,
		cancel:               cancel,
	}

	// Initialize cached signer
	s.cachedSigner = types.LatestSignerForChainID(chainID)

	// Get and validate current account
	currentAcc, err := s.GetCurrentAccount()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get current account: %w", err)
	}

	// Unlock the account
	if err := ks.Unlock(currentAcc, serviceConfig.Password); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to unlock keystore: %w", err)
	}

	// Cache the account and address
	s.cachedAccount = &currentAcc
	addr := currentAcc.Address
	s.cachedAddress = &addr

	s.registerReloadCallback(cfg)
	return s, nil
}

func (s *AccountService) registerReloadCallback(cfg *config.C) {
	cfg.RegisterReloadCallback(func(cfg *config.C) {
		// Handle config reload - could update timeouts, retry counts, etc.
		log.Info("Account service config reloaded")
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
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.cachedAddress != nil {
		return *s.cachedAddress
	}

	// Fallback to getting from account
	if s.cachedAccount != nil {
		return s.cachedAccount.Address
	}

	// Last resort - get from keystore
	accounts := s.ks.Accounts()
	if len(accounts) > 0 {
		return accounts[0].Address
	}

	return common.Address{}
}

// GetBalance retrieves the Ether balance of the account with retry logic
func (s *AccountService) GetBalance(address common.Address) (*big.Int, error) {
	var balance *big.Int
	var err error

	ctx, cancel := context.WithTimeout(s.ctx, s.clientTimeout)
	defer cancel()

	// Retry logic for network issues
	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		balance, err = s.client.BalanceAt(ctx, address, nil)
		if err == nil {
			return balance, nil
		}

		if attempt < s.maxRetries {
			log.Warnf("Failed to get balance (attempt %d/%d): %v, retrying...", attempt+1, s.maxRetries+1, err)
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	return nil, fmt.Errorf("failed to get balance after %d attempts: %w", s.maxRetries+1, err)
}

// NewKeyedTransactor returns a new keyed transactor with optimized signing
func (s *AccountService) NewKeyedTransactor() (*bind.TransactOpts, error) {
	s.mu.RLock()
	account := s.cachedAccount
	s.mu.RUnlock()

	if account == nil {
		var err error
		acc, err := s.GetCurrentAccount()
		if err != nil {
			return nil, fmt.Errorf("failed to get current account: %w", err)
		}
		account = &acc
	}

	return &bind.TransactOpts{
		From: account.Address,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != account.Address {
				return nil, fmt.Errorf("not authorized to sign transaction from address %s", address.Hex())
			}

			// Use cached signer for better performance
			signature, err := s.ks.SignHash(*account, s.cachedSigner.Hash(tx).Bytes())
			if err != nil {
				return nil, fmt.Errorf("failed to sign transaction: %w", err)
			}

			return tx.WithSignature(s.cachedSigner, signature)
		},
		Context: s.ctx,
	}, nil
}

// SignAndSendTransaction creates, signs, and sends a transaction with retry logic
func (s *AccountService) SignAndSendTransaction(toAddress string, value *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) (string, error) {
	// Validate input parameters
	if toAddress == "" {
		return "", fmt.Errorf("to address cannot be empty")
	}
	if !common.IsHexAddress(toAddress) {
		return "", fmt.Errorf("invalid to address format: %s", toAddress)
	}
	if value == nil {
		value = big.NewInt(0)
	}
	if gasPrice == nil {
		return "", fmt.Errorf("gas price cannot be nil")
	}

	s.mu.RLock()
	account := s.cachedAccount
	s.mu.RUnlock()

	if account == nil {
		var err error
		acc, err := s.GetCurrentAccount()
		if err != nil {
			return "", fmt.Errorf("failed to get current account: %w", err)
		}
		account = &acc
	}

	// Get nonce with retry logic
	var nonce uint64
	var err error

	ctx, cancel := context.WithTimeout(s.ctx, s.clientTimeout)
	defer cancel()

	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		nonce, err = s.client.PendingNonceAt(ctx, account.Address)
		if err == nil {
			break
		}

		if attempt < s.maxRetries {
			log.Warnf("Failed to get nonce (attempt %d/%d): %v, retrying...", attempt+1, s.maxRetries+1, err)
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	if err != nil {
		return "", fmt.Errorf("failed to get nonce after %d attempts: %w", s.maxRetries+1, err)
	}

	// Create transaction
	tx := types.NewTransaction(nonce, common.HexToAddress(toAddress), value, gasLimit, gasPrice, data)

	// Sign transaction
	signedTx, err := s.ks.SignTx(*account, tx, s.chainID)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction with retry logic
	ctx, cancel = context.WithTimeout(s.ctx, s.clientTimeout)
	defer cancel()

	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		err = s.client.SendTransaction(ctx, signedTx)
		if err == nil {
			return signedTx.Hash().Hex(), nil
		}

		if attempt < s.maxRetries {
			log.Warnf("Failed to send transaction (attempt %d/%d): %v, retrying...", attempt+1, s.maxRetries+1, err)
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	return "", fmt.Errorf("failed to send transaction after %d attempts: %w", s.maxRetries+1, err)
}

// EthereumService provides a lifecycle-managed Ethereum service
func EthereumService(lc fx.Lifecycle, cfg *config.C, rp repo.Repo) (*AccountService, error) {
	repoPath := filepath.Clean(rp.Path())
	ksDir := filepath.Join(repoPath, "keystore")

	// Ensure keystore directory exists
	if err := os.MkdirAll(ksDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create keystore directory: %w", err)
	}

	ks := keystore.NewKeyStore(ksDir, keystore.StandardScryptN, keystore.StandardScryptP)

	service, err := NewAccountService(cfg, ks)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("EthereumService started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("EthereumService stopping...")
			service.cancel()
			if service.client != nil {
				service.client.Close()
			}
			log.Info("EthereumService stopped")
			return nil
		},
	})

	return service, nil
}

// GetCurrentAccount returns the account matching config.account.address, or the first account if not set or not found
func (s *AccountService) GetCurrentAccount() (accounts.Account, error) {
	s.mu.RLock()
	if s.cachedAccount != nil {
		account := *s.cachedAccount
		s.mu.RUnlock()
		return account, nil
	}
	s.mu.RUnlock()

	addressHex := s.cfg.GetString("account.address", "")
	keystoreAccounts := s.ks.Accounts()

	if len(keystoreAccounts) == 0 {
		return accounts.Account{}, fmt.Errorf("no accounts found in keystore")
	}

	// If no specific address is configured, return the first account
	if addressHex == "" {
		s.mu.Lock()
		s.cachedAccount = &keystoreAccounts[0]
		account := keystoreAccounts[0]
		s.mu.Unlock()
		return account, nil
	}

	// Find account by address
	targetAddr := common.HexToAddress(addressHex)
	for _, acc := range keystoreAccounts {
		if acc.Address == targetAddr {
			s.mu.Lock()
			s.cachedAccount = &acc
			s.mu.Unlock()
			return acc, nil
		}
	}

	// If not found, fallback to first account with warning
	log.Warnf("Account with address %s not found, using first available account", addressHex)
	s.mu.Lock()
	s.cachedAccount = &keystoreAccounts[0]
	account := keystoreAccounts[0]
	s.mu.Unlock()
	return account, nil
}

// Sign the hash using ECDSA
func (s *AccountService) Sign(hash []byte) ([]byte, error) {
	if len(hash) == 0 {
		return nil, fmt.Errorf("hash cannot be empty")
	}

	s.mu.RLock()
	account := s.cachedAccount
	s.mu.RUnlock()

	if account == nil {
		var err error
		acc, err := s.GetCurrentAccount()
		if err != nil {
			return nil, fmt.Errorf("failed to get current account: %w", err)
		}
		account = &acc
	}

	signature, err := s.ks.SignHash(*account, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign hash: %w", err)
	}

	return signature, nil
}

// SignTypedData signs typed data using EIP-712
func (s *AccountService) SignTypedData(typedData *signer.TypedData) ([]byte, []byte, error) {
	if typedData == nil {
		return nil, nil, fmt.Errorf("typed data cannot be nil")
	}

	typedDataHash, _, err := signer.TypedDataAndHash(*typedData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash typed data: %w", err)
	}

	signature, err := s.Sign(typedDataHash)
	if err != nil {
		return nil, nil, err
	}

	return typedDataHash, signature, nil
}

// Close gracefully shuts down the service
func (s *AccountService) Close() error {
	s.cancel()
	if s.client != nil {
		s.client.Close()
	}
	return nil
}
