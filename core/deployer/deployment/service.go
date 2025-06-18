package deployment

import (
	"context"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/kube"
	"go.uber.org/zap"
)

type DeploymentService struct {
	KubeClient *kube.KubeClient
	Logger     *zap.Logger
}

func NewDeploymentService(kubeClient *kube.KubeClient, logger *zap.Logger) *DeploymentService {
	return &DeploymentService{
		KubeClient: kubeClient,
		Logger:     logger,
	}
}

func (s *DeploymentService) Start(ctx context.Context) error {
	return nil
}

func (s *DeploymentService) Stop(ctx context.Context) error {
	return nil
}
