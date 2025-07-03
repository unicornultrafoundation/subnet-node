package deployment

import (
	"context"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	"k8s.io/client-go/tools/remotecommand"
)

func (s *Service) RequestDeployment(ctx context.Context, deploymentRequest *types.DeploymentRequest) (*types.DeploymentResponse, error) {
	if deploymentRequest.OrderID == "" {
		return nil, fmt.Errorf("orderID is required")
	}

	if s.HasDeploymentExpiryListCache(deploymentRequest.OrderID) {
		return nil, fmt.Errorf("the order is expired, please extend the order before requesting a new deployment")
	}

	if err := deploymentRequest.Manifest.Validate(); err != nil {
		s.logger.Error("Invalid deployment request", err)
		return nil, err
	}

	manifest, err := deploymentRequest.Manifest.ToManifest()
	if err != nil {
		s.logger.Error("Failed to convert manifest to kubernetes manifest", err)
		return nil, err
	}

	requester := common.HexToAddress(deploymentRequest.Requester)

	deployment := types.Deployment{
		ID:        deploymentRequest.OrderID,
		Manifest:  manifest,
		Requester: requester,
	}

	// Deploy the deployment request
	if err := s.kubeClient.Deploy(ctx, deployment); err != nil {
		s.logger.Error("Failed to deploy deployment request", err)
		return nil, err
	}

	// Wait for the deployment to be ready
	if err := s.kubeClient.WaitForDeployment(ctx, deploymentRequest.OrderID, s.cfg.DeploymentWaitTimeout); err != nil {
		s.logger.Error("Failed to wait for deployment to be ready", err)
		return nil, err
	}

	// Store the deployment request
	if err := s.store.StoreDeploymentRequest(ctx, deploymentRequest); err != nil {
		s.logger.Error("Failed to store deployment request", err)
		return nil, err
	}

	// Store the deployment in the cache
	s.AddDeploymentListCache(deploymentRequest.OrderID)

	s.logger.WithField("order_id", deploymentRequest.OrderID).Info("Deployment request deployed")

	return s.kubeClient.GetDeployment(ctx, deploymentRequest.OrderID)
}

// GetDeployments returns a list of deployment IDs for a specific requester
func (s *Service) GetDeployments(ctx context.Context, requester string) ([]*types.DeploymentResponse, error) {
	// Get deployment IDs for the specific requester from Kubernetes
	deployments, err := s.kubeClient.GetDeployments(ctx, requester)
	if err != nil {
		s.logger.WithField("requester", requester).Error("Failed to get deployments by requester", err)
		return nil, err
	}

	return deployments, nil
}

func (s *Service) GetDeployment(ctx context.Context, orderID string) (*types.DeploymentResponse, error) {
	return s.kubeClient.GetDeployment(ctx, orderID)
}

func (s *Service) GetServiceStatus(ctx context.Context, orderID string, serviceName string) (*types.ServiceStatus, error) {
	return s.kubeClient.GetServiceStatus(ctx, orderID, serviceName)
}

func (s *Service) CleanupDeployment(ctx context.Context, orderID string) error {
	err := s.kubeClient.CleanupResources(ctx, orderID)
	if err != nil {
		s.logger.WithField("orderID", orderID).Error("Failed to cleanup deployment", err)
		return err
	}

	// Wait for the deployment to be terminated
	if err := s.kubeClient.WaitForNamespaceTermination(ctx, orderID, s.cfg.DeploymentWaitTimeout); err != nil {
		s.logger.WithField("orderID", orderID).Error("Failed to wait for deployment to be terminated", err)
		return err
	}

	// Delete the deployment request
	if err := s.store.DeleteDeploymentRequest(ctx, orderID); err != nil {
		s.logger.WithField("orderID", orderID).Error("Failed to delete deployment request from datastore", err)
		return err
	}
	s.DeleteDeploymentListCache(orderID)
	if err := s.DeleteDeploymentExpiryListCache(ctx, orderID); err != nil {
		s.logger.WithField("orderID", orderID).Error("Failed to delete deployment expiry list from datastore", err)
		return err
	}

	s.logger.WithField("orderID", orderID).Info("Deployment request deleted")

	return nil
}

func (s *Service) GetDeploymentLogs(ctx context.Context, orderID string) ([]*types.ServiceLog, error) {
	return s.kubeClient.GetDeploymentLogs(ctx, orderID, nil)
}

func (s *Service) Exec(ctx context.Context, orderID string, podName string, serviceName string, cmd []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, tty bool, tsq remotecommand.TerminalSizeQueue) (types.ExecResult, error) {
	return s.kubeClient.Exec(ctx, orderID, podName, serviceName, cmd, stdin, stdout, stderr, tty, tsq)
}

func (s *Service) GetDeploymentStats(ctx context.Context, orderID string) (*types.DeploymentStats, error) {
	return s.kubeClient.GetDeploymentStats(ctx, orderID)
}

// InspectDeployment retrieves detailed deployment information similar to kubectl describe
func (s *Service) InspectDeployment(ctx context.Context, orderID string) (*types.DetailedDeploymentStatus, error) {
	return s.kubeClient.InspectDeployment(ctx, orderID)
}
