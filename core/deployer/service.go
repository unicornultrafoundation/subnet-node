package deployer

import (
	"context"

	"github.com/ipfs/go-datastore"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/deployment"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/kube"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/store"
)

// Service represents the deployer service
type Service struct {
	config            *ServiceConfig
	deploymentService *deployment.Service
	storeService      *store.Service
	logger            *logrus.Logger
}

func NewService(config *config.C, ds datastore.Datastore) (*Service, error) {
	serviceConfig, err := NewServiceConfigFromConfig(config)
	if err != nil {
		return nil, err
	}

	logger := logrus.New().WithField("service", "deployer").Logger

	storeService := store.NewService(ds, logger)

	return &Service{
		config:       serviceConfig,
		storeService: storeService,
		logger:       logger,
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Starting deployer service")

	if err := s.storeService.Start(ctx); err != nil {
		return err
	}

	kubeClient, err := kube.NewKubeClient(ctx, s.config.KubeConfigPath, s.logger)
	if err != nil {
		return err
	}

	deploymentConfig := &deployment.Config{
		MonitorInterval: s.config.MonitorInterval,
	}

	s.deploymentService = deployment.NewService(kubeClient, s.storeService, s.logger, deploymentConfig)
	if err := s.deploymentService.Start(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	return s.deploymentService.Stop(ctx)
}
