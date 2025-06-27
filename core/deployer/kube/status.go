package kube

import (
	"context"
	"fmt"
	"strconv"
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
			return nil, fmt.Errorf("deployment not found")
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

	// Get all pods in the namespace
	pods, err := c.Client.CoreV1().Pods(deploymentID).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	// Get the namespace to read TTL and creation time
	namespace, err := c.Client.CoreV1().Namespaces().Get(ctx, deploymentID, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}

	var ttlMinutes int64 = 0
	var timeLeft int64 = -1
	if ttlStr, ok := namespace.Annotations["ttl"]; ok {
		ttlParsed, err := parseTTL(ttlStr)
		if err == nil {
			ttlMinutes = ttlParsed
			if namespace.CreationTimestamp.Time != (time.Time{}) {
				expiry := namespace.CreationTimestamp.Time.Add(time.Duration(ttlMinutes) * time.Minute)
				secondsLeft := int64(expiry.Sub(time.Now()).Seconds())
				if secondsLeft < 0 {
					timeLeft = 0
				} else {
					timeLeft = (secondsLeft + 59) / 60 // round up to next minute
				}
			}
		}
	}

	// Build service statuses
	var serviceStatuses []types.ServiceStatus
	var podStatuses []types.PodStatus
	deploymentState := types.DeploymentStateRunning

	// Process pods first to build pod statuses
	for _, pod := range pods.Items {
		// Extract service information from pod labels
		serviceName := pod.Labels["service"]
		groupName := pod.Labels["app"]

		// Determine pod state based on pod phase
		var podState types.PodState
		switch pod.Status.Phase {
		case corev1.PodRunning:
			podState = types.PodStateRunning
		case corev1.PodPending:
			podState = types.PodStatePending
		case corev1.PodFailed:
			podState = types.PodStateFailed
		case corev1.PodSucceeded:
			podState = types.PodStateStopped
		default:
			podState = types.PodStateUnknown
		}

		// Get container statuses
		var containerStatuses []types.ContainerStatus
		for _, container := range pod.Status.ContainerStatuses {
			// Determine container state
			var containerState string
			if container.State.Running != nil {
				containerState = "running"
			} else if container.State.Waiting != nil {
				containerState = "waiting"
			} else if container.State.Terminated != nil {
				containerState = "terminated"
			} else {
				containerState = "unknown"
			}

			containerStatus := types.ContainerStatus{
				Name:         container.Name,
				Image:        container.Image,
				Ready:        container.Ready,
				RestartCount: container.RestartCount,
				State:        containerState,
			}
			if container.State.Running != nil {
				containerStatus.StartedAt = container.State.Running.StartedAt.Format(time.RFC3339)
			}
			containerStatuses = append(containerStatuses, containerStatus)
		}

		// Get resource usage from pod spec
		var resourceUsage *types.ResourceUsage
		if len(pod.Spec.Containers) > 0 {
			container := pod.Spec.Containers[0]
			if container.Resources.Requests != nil {
				resourceUsage = &types.ResourceUsage{
					CPU:    container.Resources.Requests.Cpu().String(),
					Memory: container.Resources.Requests.Memory().String(),
				}
			}
		}

		// Get restart count (max of all containers)
		var maxRestartCount int32
		for _, container := range pod.Status.ContainerStatuses {
			if container.RestartCount > maxRestartCount {
				maxRestartCount = container.RestartCount
			}
		}

		// Get last restart time
		var lastRestartTime string
		for _, container := range pod.Status.ContainerStatuses {
			if container.LastTerminationState.Terminated != nil {
				restartTime := container.LastTerminationState.Terminated.FinishedAt.Format(time.RFC3339)
				if lastRestartTime == "" || restartTime > lastRestartTime {
					lastRestartTime = restartTime
				}
			}
		}

		// Get image from pod spec
		var image string
		if len(pod.Spec.Containers) > 0 {
			image = pod.Spec.Containers[0].Image
		}

		// Determine if pod is ready by checking conditions
		var podReady bool
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady {
				podReady = condition.Status == corev1.ConditionTrue
				break
			}
		}

		podStatus := types.PodStatus{
			Name:              pod.Name,
			ServiceName:       serviceName,
			Group:             groupName,
			Image:             image,
			State:             podState,
			Phase:             string(pod.Status.Phase),
			Ready:             podReady,
			RestartCount:      maxRestartCount,
			IP:                pod.Status.PodIP,
			HostIP:            pod.Status.HostIP,
			Resources:         resourceUsage,
			CreatedAt:         pod.CreationTimestamp.Format(time.RFC3339),
			StartedAt:         pod.Status.StartTime.Format(time.RFC3339),
			LastRestartTime:   lastRestartTime,
			ContainerStatuses: containerStatuses,
		}

		podStatuses = append(podStatuses, podStatus)
	}

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
								uri := fmt.Sprintf("LoadBalancer pending for service %s", svc.Name)
								servicePort.URI = uri
								serviceURIs = append(serviceURIs, uri)
							}
						} else {
							uri := fmt.Sprintf("LoadBalancer provisioning for service %s", svc.Name)
							servicePort.URI = uri
							serviceURIs = append(serviceURIs, uri)
						}
					} else if svc.Spec.Type == corev1.ServiceTypeNodePort {
						if port.NodePort > 0 {
							uri := fmt.Sprintf("http://localhost:%d", port.NodePort)
							servicePort.URI = uri
							serviceURIs = append(serviceURIs, uri)
						} else {
							uri := fmt.Sprintf("NodePort pending for service %s", svc.Name)
							servicePort.URI = uri
							serviceURIs = append(serviceURIs, uri)
						}
					} else if svc.Spec.Type == corev1.ServiceTypeClusterIP {
						uri := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", svc.Name, deploymentID, port.Port)
						servicePort.URI = uri
						serviceURIs = append(serviceURIs, uri)
					}
					servicePorts = append(servicePorts, servicePort)
				}
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
	for _, deployment := range deployments.Items {
		serviceName := deployment.Labels["service"]
		groupName := deployment.Labels["app"]
		for _, svc := range services.Items {
			if svc.Labels["service"] == serviceName && svc.Labels["app"] == groupName {
				for _, port := range svc.Spec.Ports {
					if svc.Spec.Type == corev1.ServiceTypeNodePort && port.NodePort > 0 {
						endpointInfo := types.EndpointInfo{
							Name:     svc.Name,
							Host:     "localhost",
							Path:     "/",
							Protocol: "http",
							URI:      fmt.Sprintf("http://localhost:%d/", port.NodePort),
						}
						endpointInfos = append(endpointInfos, endpointInfo)
					}
				}
			}
		}
	}

	// If no deployments found, mark as not found
	if len(deployments.Items) == 0 {
		deploymentState = types.DeploymentStateNotFound
	}

	// Create deployment status
	deploymentStatus := &types.DeploymentStatus{
		State:      deploymentState,
		Services:   serviceStatuses,
		Pods:       podStatuses,
		Endpoints:  endpointInfos,
		CreatedAt:  time.Now().Format(time.RFC3339), // This should come from deployment metadata
		UpdatedAt:  time.Now().Format(time.RFC3339),
		DeployedAt: namespace.CreationTimestamp.Time.Format(time.RFC3339),
		TTL:        ttlMinutes,
		TimeLeft:   timeLeft,
	}

	return deploymentStatus, nil
}

func (c *KubeClient) GetServiceStatus(ctx context.Context, deploymentID string, serviceName string) (*types.ServiceStatus, error) {
	deploymentStatus, err := c.getDeploymentStatusInternal(ctx, deploymentID)
	if err != nil {
		return nil, err
	}

	for _, svc := range deploymentStatus.Services {
		if svc.Name == serviceName {
			return &svc, nil
		}
	}

	// If not found, return a not found status
	return &types.ServiceStatus{
		Name:  serviceName,
		State: types.ServiceStateNotFound,
	}, nil
}

func (c *KubeClient) GetPodStatus(ctx context.Context, deploymentID string, podName string) (*types.PodStatus, error) {
	deploymentStatus, err := c.getDeploymentStatusInternal(ctx, deploymentID)
	if err != nil {
		return nil, err
	}

	for _, pod := range deploymentStatus.Pods {
		if pod.Name == podName {
			return &pod, nil
		}
	}

	// If not found, return a not found status
	return &types.PodStatus{
		Name:  podName,
		State: types.PodStateUnknown,
	}, nil
}

func parseTTL(ttlStr string) (int64, error) {
	return strconv.ParseInt(ttlStr, 10, 64)
}
