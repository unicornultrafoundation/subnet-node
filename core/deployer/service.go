package deployer

import (
	"context"

	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/deployment"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/kube"
	"go.uber.org/zap"
)

// Service represents the deployer service
type Service struct {
	config            *ServiceConfig
	accountService    *account.AccountService
	deploymentService *deployment.DeploymentService
	logger            *zap.Logger
}

func NewService(config *config.C, acc *account.AccountService) (*Service, error) {
	serviceConfig, err := NewServiceConfigFromConfig(config)
	if err != nil {
		return nil, err
	}

	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	return &Service{
		config:         serviceConfig,
		accountService: acc,
		logger:         logger.Named("deployer"),
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	kubeClient, err := kube.NewKubeClient(ctx, s.config.KubeConfigPath, s.logger)
	if err != nil {
		return err
	}

	s.deploymentService = deployment.NewDeploymentService(kubeClient, s.logger)
	return s.deploymentService.Start(ctx)
}

func (s *Service) Stop(ctx context.Context) error {
	return s.deploymentService.Stop(ctx)
}
