package k8scluster

import (
	"context"
	"fmt"
	"time"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DeploymentStatusInfo represents the current status of a deployment
type DeploymentStatusInfo struct {
	Status    types.DeploymentStatus
	UpdatedAt time.Time
}

// PodStatusInfo represents the status of a pod
type PodStatusInfo struct {
	Name   string
	Status string
	Ready  bool
}

// GetDeploymentStatus returns the current status of a deployment
func (s *Service) GetDeploymentStatus(ctx context.Context, id string) (*DeploymentStatusInfo, error) {
	dep, err := s.GetDeployment(ctx, id)
	if err != nil {
		return nil, err
	}
	return &DeploymentStatusInfo{
		Status:    dep.Status,
		UpdatedAt: dep.UpdatedAt,
	}, nil
}

// GetDeploymentPods gets the pods for a deployment
func (s *Service) GetDeploymentPods(ctx context.Context, id string) ([]PodStatusInfo, error) {
	// Get deployment
	dep, err := s.deploymentMgr.GetDeployment(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	// Get pods
	pods, err := s.client.CoreV1().Pods(dep.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	// Convert to PodStatusInfo
	podStatuses := make([]PodStatusInfo, 0, len(pods.Items))
	for _, pod := range pods.Items {
		podStatuses = append(podStatuses, PodStatusInfo{
			Name:   pod.Name,
			Status: string(pod.Status.Phase),
			Ready:  pod.Status.Phase == corev1.PodRunning,
		})
	}

	return podStatuses, nil
}

// GetServiceURL gets the URL for a service
func (s *Service) GetServiceURL(ctx context.Context, deploymentID string, serviceName string) (string, error) {
	// Get deployment
	dep, err := s.deploymentMgr.GetDeployment(deploymentID)
	if err != nil {
		return "", fmt.Errorf("failed to get deployment: %w", err)
	}

	// Get service
	svc, err := s.client.CoreV1().Services(dep.Namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get service: %w", err)
	}

	// Check if service has ports
	if len(svc.Spec.Ports) == 0 {
		return "", fmt.Errorf("service has no ports defined")
	}

	// Get the first port
	port := svc.Spec.Ports[0]

	// Get node IP
	nodes, err := s.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list nodes: %w", err)
	}

	if len(nodes.Items) == 0 {
		return "", fmt.Errorf("no nodes found")
	}

	// Get the first node's external IP
	var nodeIP string
	for _, addr := range nodes.Items[0].Status.Addresses {
		if addr.Type == corev1.NodeExternalIP {
			nodeIP = addr.Address
			break
		}
	}
	if nodeIP == "" {
		// Fallback to internal IP if external IP not found
		for _, addr := range nodes.Items[0].Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				nodeIP = addr.Address
				break
			}
		}
	}
	if nodeIP == "" {
		return "", fmt.Errorf("no valid node IP found")
	}

	// Use NodePort for NodePort services, otherwise use the service port
	var portNum int32
	if svc.Spec.Type == corev1.ServiceTypeNodePort {
		portNum = port.NodePort
	} else {
		portNum = port.Port
	}

	return fmt.Sprintf("http://%s:%d", nodeIP, portNum), nil
}

// StopDeployment stops a deployment
func (s *Service) StopDeployment(ctx context.Context, id string) error {
	// Get deployment
	dep, err := s.deploymentMgr.GetDeployment(id)
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	// Clean up resources
	if err := s.cleanupResources(ctx, dep); err != nil {
		s.logger.Error("failed to cleanup resources", zap.Error(err))
	}

	// Stop deployment
	if err := s.deploymentMgr.StopDeployment(ctx, id); err != nil {
		// Update marketplace status on failure
		s.logger.Error("failed to stop deployment", zap.Error(err))
		return fmt.Errorf("failed to stop deployment: %w", err)
	}

	// Delete namespace
	if err := s.DeleteNamespace(ctx, dep.Namespace); err != nil {
		s.logger.Error("failed to delete namespace", zap.Error(err))
	}

	// Update marketplace status
	s.logger.Info("deployment terminated", zap.String("deploymentID", id))

	return nil
}

// deleteNamespace deletes a Kubernetes namespace
func (s *Service) DeleteNamespace(ctx context.Context, name string) error {
	// Get Kubernetes client
	clientset, err := kubernetes.NewForConfig(s.config.KubeConfig)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Delete namespace
	err = clientset.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	return nil
}

// GetDeployment gets a deployment
func (s *Service) GetDeployment(ctx context.Context, id string) (*types.ManagedDeployment, error) {
	dep, err := s.deploymentMgr.GetDeployment(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}
	return dep, nil
}

// ListDeployments lists deployments
func (s *Service) ListDeployments(ctx context.Context) ([]*types.ManagedDeployment, error) {
	deps, err := s.deploymentMgr.ListDeployments()
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}
	return deps, nil
}
