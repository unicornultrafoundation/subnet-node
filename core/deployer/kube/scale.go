package kube

import (
	"context"
	"fmt"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *KubeClient) ScaleDeployment(ctx context.Context, deploymentID string, replicas int32) error {
	// Validate replicas count
	if replicas < 0 {
		return fmt.Errorf("replicas count cannot be negative")
	}

	// Get all deployments in the namespace
	deployments, err := c.Client.AppsV1().Deployments(deploymentID).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	if len(deployments.Items) == 0 {
		return fmt.Errorf("no deployments found for deployment ID: %s", deploymentID)
	}

	// Scale each deployment to the specified number of replicas
	for _, deployment := range deployments.Items {
		// Set the replicas count
		deployment.Spec.Replicas = &replicas

		// Update the deployment
		_, err := c.Client.AppsV1().Deployments(deploymentID).Update(ctx, &deployment, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to scale deployment %s: %w", deployment.Name, err)
		}

		c.Logger.WithFields(map[string]interface{}{
			"deploymentID": deploymentID,
			"deployment":   deployment.Name,
			"replicas":     replicas,
		}).Info("scaled deployment")
	}

	return nil
}

func (c *KubeClient) ScaleDeploymentToOriginal(ctx context.Context, deploymentRequest *types.DeploymentRequest) error {
	// Validate deployment request
	if deploymentRequest == nil {
		return fmt.Errorf("deployment request cannot be nil")
	}

	if deploymentRequest.OrderID == "" {
		return fmt.Errorf("order ID cannot be empty")
	}

	// Note: Manifest is of type manifest.SDL, not a pointer, so we don't need to check for nil

	// Convert SDL to manifest to get the original service configurations
	manifest, err := deploymentRequest.Manifest.ToManifest()
	if err != nil {
		return fmt.Errorf("failed to convert manifest: %w", err)
	}

	// Get all deployments in the namespace
	deployments, err := c.Client.AppsV1().Deployments(deploymentRequest.OrderID).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentRequest.OrderID),
	})
	if err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	if len(deployments.Items) == 0 {
		return fmt.Errorf("no deployments found for deployment ID: %s", deploymentRequest.OrderID)
	}

	// Create a map of service names to their original replica counts
	originalReplicas := make(map[string]int32)
	if len(manifest.Groups) > 0 {
		for _, group := range manifest.Groups {
			for _, service := range group.Services {
				originalReplicas[service.Name] = service.Count
			}
		}
	} else {
		// fallback for test SDLs that only set Services
		for name, service := range deploymentRequest.Manifest.Services {
			svcName := service.Name
			if svcName == "" {
				svcName = name
			}
			originalReplicas[svcName] = service.Count
		}
	}

	// Scale each deployment to its original replica count
	for _, deployment := range deployments.Items {
		// Extract service name from deployment labels
		serviceName := deployment.Labels["service"]
		if serviceName == "" {
			c.Logger.WithFields(map[string]interface{}{
				"deploymentID": deploymentRequest.OrderID,
				"deployment":   deployment.Name,
			}).Warn("deployment missing service label, skipping")
			continue
		}

		// Get original replica count for this service
		originalCount, exists := originalReplicas[serviceName]
		if !exists {
			c.Logger.WithFields(map[string]interface{}{
				"deploymentID": deploymentRequest.OrderID,
				"deployment":   deployment.Name,
				"service":      serviceName,
			}).Warn("service not found in original manifest, skipping")
			continue
		}

		// Set the replicas count to original
		deployment.Spec.Replicas = &originalCount

		// Update the deployment
		_, err := c.Client.AppsV1().Deployments(deploymentRequest.OrderID).Update(ctx, &deployment, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to scale deployment %s to original count: %w", deployment.Name, err)
		}

		c.Logger.WithFields(map[string]interface{}{
			"deploymentID": deploymentRequest.OrderID,
			"deployment":   deployment.Name,
			"service":      serviceName,
			"replicas":     originalCount,
		}).Info("scaled deployment to original replica count")
	}

	return nil
}
