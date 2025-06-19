package kube

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetDeployment retrieves deployment information including status and service URIs
func (c *KubeClient) GetDeployment(ctx context.Context, deploymentID string) (*types.DeploymentResponse, error) {
	// Check if namespace exists
	namespace, err := c.Client.CoreV1().Namespaces().Get(ctx, deploymentID, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Return deployment with not found status
			return &types.DeploymentResponse{
				ID: deploymentID,
				Status: &types.DeploymentStatus{
					State: types.DeploymentStateNotFound,
				},
			}, nil
		}
		return nil, fmt.Errorf("failed to check namespace existence: %w", err)
	}

	// Get deployment status using shared helper
	deploymentStatus, err := c.getDeploymentStatusInternal(ctx, deploymentID)
	if err != nil {
		return nil, err
	}

	// Get requester from namespace labels
	requester := namespace.Labels["requester"]

	// Create deployment response
	deployment := &types.DeploymentResponse{
		ID:        deploymentID,
		Requester: common.HexToAddress(requester),
		Status:    deploymentStatus,
		// Note: Manifest would need to be retrieved from storage/database
		// as it's not stored in Kubernetes resources
	}

	return deployment, nil
}

func (c *KubeClient) GetDeployments(ctx context.Context, requester string) ([]*types.DeploymentResponse, error) {
	// Query all namespaces to find deployments with the specified requester
	namespaces, err := c.Client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("requester=%s", requester),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces for requester %s: %w", requester, err)
	}

	var deployments []*types.DeploymentResponse
	for _, namespace := range namespaces.Items {
		// The namespace name is the deployment ID
		deployment, err := c.GetDeployment(ctx, namespace.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get deployment %s: %w", namespace.Name, err)
		}
		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

// getDeploymentStatusInternal is a shared helper method for getting deployment status
func (c *KubeClient) getDeploymentStatusInternal(ctx context.Context, deploymentID string) (*types.DeploymentStatus, error) {
	// Get all deployments in the namespace
	deployments, err := c.Client.AppsV1().Deployments(deploymentID).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	// Get all services in the namespace
	services, err := c.Client.CoreV1().Services(deploymentID).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	// Get all ingresses in the namespace
	ingresses, err := c.Client.NetworkingV1().Ingresses(deploymentID).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list ingresses: %w", err)
	}

	// Build service statuses
	var serviceStatuses []types.ServiceStatus
	deploymentState := types.DeploymentStateRunning

	for _, deployment := range deployments.Items {
		// Extract service information from deployment labels
		serviceName := deployment.Labels["service"]
		groupName := deployment.Labels["app"]

		// Determine service state based on deployment status
		serviceState := types.ServiceStatePending
		if deployment.Status.ReadyReplicas > 0 {
			serviceState = types.ServiceStateRunning
		} else if deployment.Status.Replicas == 0 {
			serviceState = types.ServiceStateStopped
		} else if deployment.Status.UnavailableReplicas > 0 {
			serviceState = types.ServiceStateFailed
		}

		// Update overall deployment state
		if serviceState == types.ServiceStateFailed {
			deploymentState = types.DeploymentStateFailed
		} else if serviceState == types.ServiceStatePending && deploymentState != types.DeploymentStateFailed {
			deploymentState = types.DeploymentStatePending
		}

		// Find corresponding service for URI generation
		var serviceURIs []string
		var servicePorts []types.ServicePort

		for _, svc := range services.Items {
			if svc.Labels["service"] == serviceName && svc.Labels["app"] == groupName {
				// Generate URIs for the service
				for _, port := range svc.Spec.Ports {
					servicePort := types.ServicePort{
						Port:        port.Port,
						Protocol:    string(port.Protocol),
						ServiceType: string(svc.Spec.Type),
					}

					// Generate URI based on service type
					if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
						if len(svc.Status.LoadBalancer.Ingress) > 0 {
							lb := svc.Status.LoadBalancer.Ingress[0]
							if lb.IP != "" {
								uri := fmt.Sprintf("http://%s:%d", lb.IP, port.Port)
								servicePort.URI = uri
								serviceURIs = append(serviceURIs, uri)
							} else if lb.Hostname != "" {
								uri := fmt.Sprintf("http://%s:%d", lb.Hostname, port.Port)
								servicePort.URI = uri
								serviceURIs = append(serviceURIs, uri)
							} else {
								// LoadBalancer is pending
								uri := fmt.Sprintf("LoadBalancer pending for service %s", svc.Name)
								servicePort.URI = uri
								serviceURIs = append(serviceURIs, uri)
							}
						} else {
							// LoadBalancer is still being provisioned
							uri := fmt.Sprintf("LoadBalancer provisioning for service %s", svc.Name)
							servicePort.URI = uri
							serviceURIs = append(serviceURIs, uri)
						}
					} else if svc.Spec.Type == corev1.ServiceTypeNodePort {
						// For NodePort, use localhost with the NodePort
						if port.NodePort > 0 {
							uri := fmt.Sprintf("http://localhost:%d", port.NodePort)
							servicePort.URI = uri
							serviceURIs = append(serviceURIs, uri)
						} else {
							// NodePort not assigned yet
							uri := fmt.Sprintf("NodePort pending for service %s", svc.Name)
							servicePort.URI = uri
							serviceURIs = append(serviceURIs, uri)
						}
					} else if svc.Spec.Type == corev1.ServiceTypeClusterIP {
						// For cluster IP, use internal service name
						uri := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", svc.Name, deploymentID, port.Port)
						servicePort.URI = uri
						serviceURIs = append(serviceURIs, uri)
					}

					servicePorts = append(servicePorts, servicePort)
				}
				break
			}
		}

		// Get resource usage from deployment
		var resourceUsage *types.ResourceUsage
		if len(deployment.Spec.Template.Spec.Containers) > 0 {
			container := deployment.Spec.Template.Spec.Containers[0]
			if container.Resources.Requests != nil {
				resourceUsage = &types.ResourceUsage{
					CPU:    container.Resources.Requests.Cpu().String(),
					Memory: container.Resources.Requests.Memory().String(),
				}
			}
		}

		serviceStatus := types.ServiceStatus{
			Name:              serviceName,
			Group:             groupName,
			Image:             deployment.Spec.Template.Spec.Containers[0].Image, // Use actual image name
			State:             serviceState,
			Replicas:          *deployment.Spec.Replicas,
			ReadyReplicas:     deployment.Status.ReadyReplicas,
			AvailableReplicas: deployment.Status.AvailableReplicas,
			URIs:              serviceURIs,
			Ports:             servicePorts,
			Resources:         resourceUsage,
			LastUpdated:       fmt.Sprintf("%d", deployment.Status.UpdatedReplicas),
		}

		serviceStatuses = append(serviceStatuses, serviceStatus)
	}

	// Build endpoint information
	var endpointInfos []types.EndpointInfo
	for _, ingress := range ingresses.Items {
		for _, rule := range ingress.Spec.Rules {
			for _, path := range rule.HTTP.Paths {
				// Try to find a better URI for external access
				var externalURI string

				// Check if we have a LoadBalancer service that can provide external access
				for _, svc := range services.Items {
					if svc.Spec.Type == corev1.ServiceTypeLoadBalancer && len(svc.Status.LoadBalancer.Ingress) > 0 {
						lb := svc.Status.LoadBalancer.Ingress[0]
						if lb.IP != "" {
							externalURI = fmt.Sprintf("http://%s:%d%s", lb.IP, svc.Spec.Ports[0].Port, path.Path)
							break
						} else if lb.Hostname != "" {
							externalURI = fmt.Sprintf("http://%s:%d%s", lb.Hostname, svc.Spec.Ports[0].Port, path.Path)
							break
						}
					} else if svc.Spec.Type == corev1.ServiceTypeNodePort && svc.Spec.Ports[0].NodePort > 0 {
						// For NodePort services, use localhost
						externalURI = fmt.Sprintf("http://localhost:%d%s", svc.Spec.Ports[0].NodePort, path.Path)
						break
					}
				}

				// If no external URI found, use the ingress host (might be accessible via ingress controller)
				if externalURI == "" {
					externalURI = fmt.Sprintf("http://%s%s", rule.Host, path.Path)
				}

				endpointInfo := types.EndpointInfo{
					Name:     fmt.Sprintf("%s-%s", ingress.Labels["app"], ingress.Labels["endpoint"]),
					Host:     rule.Host,
					Path:     path.Path,
					Protocol: "http",
					URI:      externalURI,
				}
				endpointInfos = append(endpointInfos, endpointInfo)
			}
		}
	}

	// If no deployments found, mark as not found
	if len(deployments.Items) == 0 {
		deploymentState = types.DeploymentStateNotFound
	}

	// Create deployment status
	deploymentStatus := &types.DeploymentStatus{
		State:     deploymentState,
		Services:  serviceStatuses,
		Endpoints: endpointInfos,
		CreatedAt: time.Now().Format(time.RFC3339), // This should come from deployment metadata
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	return deploymentStatus, nil
}
