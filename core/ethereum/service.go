package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/contracts"
	"github.com/unicornultrafoundation/subnet-node/repo"
	"go.uber.org/fx"
)

var log = logrus.WithField("service", "ethereum")

type Service interface {
	GetClient() *ethclient.Client
	GetChainID() *big.Int
	Provider() *contracts.SubnetProvider
	IPRegistry() IPRegistry
	ProviderAddr() string
	IPRegistryAddr() string
}

type IPRegistry interface {
	GetPeer(opts *bind.CallOpts, tokenID *big.Int) (string, error)
}

type ServiceConfig struct {
	RPCURL               string
	ChainID              int64
	SubnetIPRegistryAddr string
	SubnetProviderAddr   string
}

func NewServiceConfig(cfg *config.C) (*ServiceConfig, error) {
	rpc := cfg.GetString("account.rpc", config.DefaultRPC)
	if rpc == "" {
		return nil, fmt.Errorf("RPC URL cannot be empty")
	}
	chainID := int64(cfg.GetInt("account.chainid", config.DefaultChainID))
	if chainID <= 0 {
		return nil, fmt.Errorf("invalid chain ID: %d", chainID)
	}
	ipReg := cfg.GetString("apps.subnet_ip_registry", config.DefaultSubnetIP)
	if ipReg == "" {
		return nil, fmt.Errorf("subnet IP registry address cannot be empty")
	}
	return &ServiceConfig{
		RPCURL:               rpc,
		ChainID:              chainID,
		SubnetIPRegistryAddr: ipReg,
		SubnetProviderAddr:   cfg.GetString("apps.subnet_provider", config.DefaultSubnetProviderAddr),
	}, nil
}

type EthereumService struct {
	client               *ethclient.Client
	chainID              *big.Int
	subnetProvider       *contracts.SubnetProvider
	subnetProviderAddr   string
	subnetIPRegistry     IPRegistry
	subnetIPRegistryAddr string
	cfg                  *config.C

	ctx    context.Context
	cancel context.CancelFunc
}

func (s *EthereumService) GetClient() *ethclient.Client        { return s.client }
func (s *EthereumService) GetChainID() *big.Int                { return s.chainID }
func (s *EthereumService) Provider() *contracts.SubnetProvider { return s.subnetProvider }
func (s *EthereumService) IPRegistry() IPRegistry              { return s.subnetIPRegistry }
func (s *EthereumService) ProviderAddr() string                { return s.subnetProviderAddr }
func (s *EthereumService) IPRegistryAddr() string              { return s.subnetIPRegistryAddr }

func NewEthereumService(cfg *config.C) (*EthereumService, error) {
	sc, err := NewServiceConfig(cfg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	client, err := ethclient.DialContext(ctx, sc.RPCURL)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to RPC endpoint %s: %w", sc.RPCURL, err)
	}

	chainID := big.NewInt(sc.ChainID)
	netChainID, err := client.ChainID(ctx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get network chain ID: %w", err)
	}
	if chainID.Cmp(netChainID) != 0 {
		log.Warnf("Configured chain ID (%s) differs from network chain ID (%s)", chainID.String(), netChainID.String())
	}

	ipReg, err := contracts.NewSubnetIPRegistry(common.HexToAddress(sc.SubnetIPRegistryAddr), client)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize IP registry contract: %w", err)
	}

	var provider *contracts.SubnetProvider
	if sc.SubnetProviderAddr != "" {
		provider, err = contracts.NewSubnetProvider(common.HexToAddress(sc.SubnetProviderAddr), client)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize provider contract: %w", err)
		}
	}

	return &EthereumService{
		client:               client,
		chainID:              chainID,
		subnetProvider:       provider,
		subnetProviderAddr:   sc.SubnetProviderAddr,
		subnetIPRegistry:     ipReg,
		subnetIPRegistryAddr: sc.SubnetIPRegistryAddr,
		cfg:                  cfg,
		ctx:                  ctx,
		cancel:               cancel,
	}, nil
}

func ServiceFx(lc fx.Lifecycle, cfg *config.C, rp repo.Repo) (*EthereumService, error) {
	// ensure repo is initialized (side-effect consistent with keystore dir creation elsewhere)
	_ = os.MkdirAll(rp.Path(), 0700)

	svc, err := NewEthereumService(cfg)
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
			svc.cancel()
			if svc.client != nil {
				svc.client.Close()
			}
			log.Info("EthereumService stopped")
			return nil
		},
	})
	return svc, nil
}
