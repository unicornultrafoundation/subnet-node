package deployment

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/kube"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/store"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
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

// GetDeployments returns a list of deployment IDs for a specific requester
func (s *Service) GetDeployments(ctx context.Context, requester string) ([]*types.Deployment, error) {
	// Get deployment requests for the specific requester from the datastore
	deploymentRequests, err := s.store.GetDeploymentRequests(ctx)
	if err != nil {
		s.logger.WithField("requester", requester).Error("Failed to get deployment requests by requester", err)
		return nil, err
	}

	var deploymentIDs []string
	for _, req := range deploymentRequests {
		deploymentIDs = append(deploymentIDs, req.OrderID)
	}

	var deployments []*types.Deployment
	for _, deploymentID := range deploymentIDs {
		deployment, err := s.kubeClient.GetDeployment(ctx, deploymentID)
		if err != nil {
			s.logger.WithField("deploymentID", deploymentID).Error("Failed to get deployment", err)
		}
		if deployment != nil {
			deployments = append(deployments, deployment)
		}
	}

	return deployments, nil
}
