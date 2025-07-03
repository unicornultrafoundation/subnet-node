package deployment

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	bidengineTypes "github.com/unicornultrafoundation/subnet-node/bidengine/types"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/kube"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/store"
)

type Config struct {
	MonitorInterval       time.Duration
	DeploymentWaitTimeout time.Duration
}

type Service struct {
	cfg                       *Config
	kubeClient                *kube.KubeClient
	store                     *store.Service
	logger                    *logrus.Logger
	bidMarket                 bidengineTypes.BidMarketContract
	orderGracePeriod          time.Duration
	ethClient                 *ethclient.Client
	mu                        sync.Mutex          // Mutex to protect the deployment cache
	deploymentListCache       map[string]struct{} // Set of deployment IDs running in the cluster, used for monitoring
	muExpiry                  sync.Mutex          // Mutex to protect the deployment expiry cache
	deploymentExpiryListCache map[string]struct{} // Set of deployment IDs that have expired, used for monitoring
}

func NewService(kubeClient *kube.KubeClient, store *store.Service, logger *logrus.Logger, cfg *Config, bidMarket bidengineTypes.BidMarketContract, ethClient *ethclient.Client) *Service {
	return &Service{
		kubeClient:                kubeClient,
		store:                     store,
		logger:                    logger,
		cfg:                       cfg,
		bidMarket:                 bidMarket,
		orderGracePeriod:          time.Hour * 24, // 24 hours. TODO: get from Smart Contract
		deploymentListCache:       make(map[string]struct{}),
		deploymentExpiryListCache: make(map[string]struct{}),
		ethClient:                 ethClient,
	}
}

func (s *Service) Start(ctx context.Context) error {
	err := s.SyncDeploymentList(ctx)
	if err != nil {
		s.logger.Error("Failed to sync deployment list", err)
		return err
	}

	// Start the monitor
	go s.MonitorDeployments(ctx)

	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	return nil
}

func (s *Service) SyncDeploymentList(ctx context.Context) error {
	// Load the deployments from the datastore
	deployments, err := s.store.GetDeploymentRequests(ctx)
	if err != nil {
		s.logger.Error("Failed to get deployments", err)
		return err
	}

	// Load the expired deployments from the datastore
	expiredDeployments, err := s.store.GetDeploymentExpiredList(ctx)
	if err != nil {
		s.logger.Error("Failed to get expired deployments", err)
		return err
	}

	// Sync the deployment list
	s.mu.Lock()
	for _, deployment := range deployments {
		s.deploymentListCache[deployment.OrderID] = struct{}{}
	}
	s.mu.Unlock()

	// Sync the deployment expiry list
	s.muExpiry.Lock()
	for _, deployment := range expiredDeployments {
		s.deploymentExpiryListCache[deployment] = struct{}{}
	}
	s.muExpiry.Unlock()

	return nil
}
