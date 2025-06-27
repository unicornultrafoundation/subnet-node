package deployer

import (
	"context"
	"io"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	"k8s.io/client-go/tools/remotecommand"
)

func (s *Service) RequestDeployment(ctx context.Context, deploymentRequest *types.DeploymentRequest) (*types.DeploymentResponse, error) {
	return s.deploymentService.RequestDeployment(ctx, deploymentRequest)
}

func (s *Service) GetDeployment(ctx context.Context, orderID string) (*types.DeploymentResponse, error) {
	return s.deploymentService.GetDeployment(ctx, orderID)
}

func (s *Service) GetServiceStatus(ctx context.Context, orderID string, serviceName string) (*types.ServiceStatus, error) {
	return s.deploymentService.GetServiceStatus(ctx, orderID, serviceName)
}

func (s *Service) CleanupDeployment(ctx context.Context, orderID string) error {
	return s.deploymentService.CleanupDeployment(ctx, orderID)
}

func (s *Service) GetDeployments(ctx context.Context, requester string) ([]*types.DeploymentResponse, error) {
	return s.deploymentService.GetDeployments(ctx, requester)
}

func (s *Service) GetDeploymentRequest(ctx context.Context, orderID string) (*types.DeploymentRequest, error) {
	return s.storeService.GetDeploymentRequest(ctx, orderID)
}

func (s *Service) GetDeploymentLogs(ctx context.Context, orderID string) ([]*types.ServiceLog, error) {
	return s.deploymentService.GetDeploymentLogs(ctx, orderID)
}

func (s *Service) Exec(ctx context.Context, orderID string, podName string, serviceName string, cmd []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, tty bool, tsq remotecommand.TerminalSizeQueue) (types.ExecResult, error) {
	return s.deploymentService.Exec(ctx, orderID, podName, serviceName, cmd, stdin, stdout, stderr, tty, tsq)
}

func (s *Service) GetDeploymentStats(ctx context.Context, orderID string) (*types.DeploymentStats, error) {
	return s.deploymentService.GetDeploymentStats(ctx, orderID)
}
