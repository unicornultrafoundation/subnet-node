package deployment

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/kube"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/store"
)

type Config struct {
	MonitorInterval time.Duration
}

type Service struct {
	kubeClient          *kube.KubeClient
	store               *store.Service
	logger              *logrus.Logger
	mu                  sync.Mutex          // Mutex to protect the deployment cache
	deploymentListCache map[string]struct{} // Set of deployment IDs running in the cluster, used for monitoring
	cfg                 *Config
}

func NewService(kubeClient *kube.KubeClient, store *store.Service, logger *logrus.Logger, cfg *Config) *Service {
	return &Service{
		kubeClient:          kubeClient,
		store:               store,
		logger:              logger,
		cfg:                 cfg,
		deploymentListCache: make(map[string]struct{}),
	}
}

func (s *Service) Start(ctx context.Context) error {
	// Load the deployments from the datastore
	deployments, err := s.store.GetDeploymentRequests(ctx)
	if err != nil {
		s.logger.Error("Failed to get deployments", err)
		return err
	}

	// Load the deployments into the cache
	for _, deployment := range deployments {
		s.mu.Lock()
		s.deploymentListCache[deployment.OrderID] = struct{}{}
		s.mu.Unlock()
	}

	// Start the monitor
	go s.MonitorDeployments(ctx)

	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	return nil
}
