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

					if svc.Spec.Type == corev1.ServiceTypeNodePort && port.NodePort > 0 {
						servicePort.URI = fmt.Sprintf("http://localhost:%d/", port.NodePort)
						serviceURIs = append(serviceURIs, servicePort.URI)
					} else if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
						if len(svc.Status.LoadBalancer.Ingress) > 0 {
							ingress := svc.Status.LoadBalancer.Ingress[0]
							if ingress.IP != "" {
								servicePort.URI = fmt.Sprintf("http://%s:%d/", ingress.IP, port.Port)
								serviceURIs = append(serviceURIs, servicePort.URI)
							} else if ingress.Hostname != "" {
								servicePort.URI = fmt.Sprintf("http://%s:%d/", ingress.Hostname, port.Port)
								serviceURIs = append(serviceURIs, servicePort.URI)
							}
						}
					} else if svc.Spec.Type == corev1.ServiceTypeClusterIP {
						servicePort.URI = fmt.Sprintf("http://%s:%d/", svc.Spec.ClusterIP, port.Port)
						serviceURIs = append(serviceURIs, servicePort.URI)
					}

					servicePorts = append(servicePorts, servicePort)
				}
			}
		}

		// Get resource usage from deployment spec
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

// inspectDeploymentStatusInternal provides kubectl describe-like detailed information
func (c *KubeClient) inspectDeploymentStatusInternal(ctx context.Context, deploymentID string) (*types.DetailedDeploymentStatus, error) {
	// Get basic deployment status first
	basicStatus, err := c.getDeploymentStatusInternal(ctx, deploymentID)
	if err != nil {
		return nil, err
	}

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

	// Get all events in the namespace
	events, err := c.Client.CoreV1().Events(deploymentID).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.namespace=%s", deploymentID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	// Get all configmaps in the namespace
	configMaps, err := c.Client.CoreV1().ConfigMaps(deploymentID).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list configmaps: %w", err)
	}

	// Get all secrets in the namespace
	secrets, err := c.Client.CoreV1().Secrets(deploymentID).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	// Get all persistent volume claims in the namespace
	pvcs, err := c.Client.CoreV1().PersistentVolumeClaims(deploymentID).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list persistent volume claims: %w", err)
	}

	// Get all persistent volumes
	var persistentVolumes []corev1.PersistentVolume
	if len(pvcs.Items) > 0 {
		pvs, err := c.Client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list persistent volumes: %w", err)
		}
		persistentVolumes = pvs.Items
	}

	// Get the namespace to read TTL and creation time
	namespace, err := c.Client.CoreV1().Namespaces().Get(ctx, deploymentID, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}

	// Get resource quota for the namespace
	var resourceQuota *types.ResourceQuota
	quotas, err := c.Client.CoreV1().ResourceQuotas(deploymentID).List(ctx, metav1.ListOptions{})
	if err == nil && len(quotas.Items) > 0 {
		quota := quotas.Items[0]
		resourceQuota = &types.ResourceQuota{
			Hard: make(map[string]string),
			Used: make(map[string]string),
		}
		for resource, quantity := range quota.Spec.Hard {
			resourceQuota.Hard[string(resource)] = quantity.String()
		}
		for resource, quantity := range quota.Status.Used {
			resourceQuota.Used[string(resource)] = quantity.String()
		}
	}

	// Build namespace info
	namespaceInfo := &types.NamespaceInfo{
		Name:          namespace.Name,
		Status:        string(namespace.Status.Phase),
		Labels:        namespace.Labels,
		Annotations:   namespace.Annotations,
		CreatedAt:     namespace.CreationTimestamp.Format(time.RFC3339),
		ResourceQuota: resourceQuota,
	}

	// Build detailed deployments
	var detailedDeployments []types.DetailedDeployment
	for _, deployment := range deployments.Items {
		// Build deployment conditions
		var conditions []types.DeploymentCondition
		for _, condition := range deployment.Status.Conditions {
			conditions = append(conditions, types.DeploymentCondition{
				Type:               string(condition.Type),
				Status:             string(condition.Status),
				LastUpdateTime:     condition.LastUpdateTime.Format(time.RFC3339),
				LastTransitionTime: condition.LastTransitionTime.Format(time.RFC3339),
				Reason:             condition.Reason,
				Message:            condition.Message,
			})
		}

		// Build deployment strategy
		strategy := types.DeploymentStrategy{
			Type:           string(deployment.Spec.Strategy.Type),
			MaxSurge:       deployment.Spec.Strategy.RollingUpdate.MaxSurge.String(),
			MaxUnavailable: deployment.Spec.Strategy.RollingUpdate.MaxUnavailable.String(),
		}

		// Build pod template spec
		template := types.PodTemplateSpec{
			Labels:      deployment.Spec.Template.Labels,
			Annotations: deployment.Spec.Template.Annotations,
			Containers:  []types.ContainerSpec{},
			Volumes:     []types.VolumeSpec{},
		}

		// Build container specs
		for _, container := range deployment.Spec.Template.Spec.Containers {
			containerSpec := types.ContainerSpec{
				Name:    container.Name,
				Image:   container.Image,
				Command: container.Command,
				Args:    container.Args,
				Ports:   []types.ContainerPort{},
				Env:     []types.EnvVar{},
				Resources: types.ResourceRequirements{
					Requests: make(map[string]string),
					Limits:   make(map[string]string),
				},
				VolumeMounts: []types.VolumeMount{},
			}

			// Build container ports
			for _, port := range container.Ports {
				containerSpec.Ports = append(containerSpec.Ports, types.ContainerPort{
					Name:          port.Name,
					ContainerPort: port.ContainerPort,
					Protocol:      string(port.Protocol),
					HostPort:      port.HostPort,
				})
			}

			// Build environment variables
			for _, env := range container.Env {
				containerSpec.Env = append(containerSpec.Env, types.EnvVar{
					Name:  env.Name,
					Value: env.Value,
				})
			}

			// Build resource requirements
			if container.Resources.Requests != nil {
				for resource, quantity := range container.Resources.Requests {
					containerSpec.Resources.Requests[string(resource)] = quantity.String()
				}
			}
			if container.Resources.Limits != nil {
				for resource, quantity := range container.Resources.Limits {
					containerSpec.Resources.Limits[string(resource)] = quantity.String()
				}
			}

			// Build volume mounts
			for _, mount := range container.VolumeMounts {
				containerSpec.VolumeMounts = append(containerSpec.VolumeMounts, types.VolumeMount{
					Name:      mount.Name,
					MountPath: mount.MountPath,
					ReadOnly:  mount.ReadOnly,
				})
			}

			// Build probes
			if container.LivenessProbe != nil {
				containerSpec.LivenessProbe = &types.Probe{
					InitialDelaySeconds: container.LivenessProbe.InitialDelaySeconds,
					PeriodSeconds:       container.LivenessProbe.PeriodSeconds,
					TimeoutSeconds:      container.LivenessProbe.TimeoutSeconds,
					SuccessThreshold:    container.LivenessProbe.SuccessThreshold,
					FailureThreshold:    container.LivenessProbe.FailureThreshold,
				}
			}
			if container.ReadinessProbe != nil {
				containerSpec.ReadinessProbe = &types.Probe{
					InitialDelaySeconds: container.ReadinessProbe.InitialDelaySeconds,
					PeriodSeconds:       container.ReadinessProbe.PeriodSeconds,
					TimeoutSeconds:      container.ReadinessProbe.TimeoutSeconds,
					SuccessThreshold:    container.ReadinessProbe.SuccessThreshold,
					FailureThreshold:    container.ReadinessProbe.FailureThreshold,
				}
			}

			// Build security context
			if container.SecurityContext != nil {
				containerSpec.SecurityContext = &types.SecurityContext{
					RunAsUser:    container.SecurityContext.RunAsUser,
					RunAsGroup:   container.SecurityContext.RunAsGroup,
					RunAsNonRoot: container.SecurityContext.RunAsNonRoot,
					Privileged:   container.SecurityContext.Privileged,
				}
			}

			template.Containers = append(template.Containers, containerSpec)
		}

		// Build volumes
		for _, volume := range deployment.Spec.Template.Spec.Volumes {
			volumeSpec := types.VolumeSpec{
				Name:         volume.Name,
				VolumeSource: make(map[string]interface{}),
			}
			// Note: Volume source details would need to be implemented based on specific volume types
			template.Volumes = append(template.Volumes, volumeSpec)
		}

		detailedDeployment := types.DetailedDeployment{
			Name:                deployment.Name,
			Namespace:           deployment.Namespace,
			Labels:              deployment.Labels,
			Annotations:         deployment.Annotations,
			Replicas:            *deployment.Spec.Replicas,
			ReadyReplicas:       deployment.Status.ReadyReplicas,
			AvailableReplicas:   deployment.Status.AvailableReplicas,
			UnavailableReplicas: deployment.Status.UnavailableReplicas,
			UpdatedReplicas:     deployment.Status.UpdatedReplicas,
			Strategy:            strategy,
			Conditions:          conditions,
			Selector:            deployment.Spec.Selector.MatchLabels,
			Template:            template,
			CreatedAt:           deployment.CreationTimestamp.Format(time.RFC3339),
			UpdatedAt:           fmt.Sprintf("%d", deployment.Status.UpdatedReplicas),
		}

		detailedDeployments = append(detailedDeployments, detailedDeployment)
	}

	// Build detailed services
	var detailedServices []types.DetailedService
	for _, service := range services.Items {
		// Get endpoints for this service
		endpoints, err := c.Client.CoreV1().Endpoints(deploymentID).Get(ctx, service.Name, metav1.GetOptions{})
		var endpointInfos []types.Endpoint
		if err == nil {
			for _, subset := range endpoints.Subsets {
				endpoint := types.Endpoint{
					Addresses: []string{},
					Ports:     []types.EndpointPort{},
				}
				for _, address := range subset.Addresses {
					endpoint.Addresses = append(endpoint.Addresses, address.IP)
				}
				for _, port := range subset.Ports {
					endpoint.Ports = append(endpoint.Ports, types.EndpointPort{
						Name:     port.Name,
						Port:     port.Port,
						Protocol: string(port.Protocol),
					})
				}
				endpointInfos = append(endpointInfos, endpoint)
			}
		}

		detailedService := types.DetailedService{
			Name:            service.Name,
			Namespace:       service.Namespace,
			Labels:          service.Labels,
			Annotations:     service.Annotations,
			Type:            string(service.Spec.Type),
			ClusterIP:       service.Spec.ClusterIP,
			ExternalIPs:     service.Spec.ExternalIPs,
			LoadBalancerIP:  service.Spec.LoadBalancerIP,
			Ports:           []types.ServicePort{},
			Selector:        service.Spec.Selector,
			SessionAffinity: string(service.Spec.SessionAffinity),
			CreatedAt:       service.CreationTimestamp.Format(time.RFC3339),
			Endpoints:       endpointInfos,
		}

		// Build service ports
		for _, port := range service.Spec.Ports {
			servicePort := types.ServicePort{
				Port:        port.Port,
				Protocol:    string(port.Protocol),
				ServiceType: string(service.Spec.Type),
			}
			if service.Spec.Type == corev1.ServiceTypeNodePort && port.NodePort > 0 {
				servicePort.URI = fmt.Sprintf("http://localhost:%d/", port.NodePort)
			}
			detailedService.Ports = append(detailedService.Ports, servicePort)
		}

		detailedServices = append(detailedServices, detailedService)
	}

	// Build detailed pods
	var detailedPods []types.DetailedPod
	for _, pod := range pods.Items {
		// Get basic pod status
		var basicPodStatus *types.PodStatus
		for _, status := range basicStatus.Pods {
			if status.Name == pod.Name {
				basicPodStatus = &status
				break
			}
		}

		// Build pod conditions
		var conditions []types.PodCondition
		for _, condition := range pod.Status.Conditions {
			conditions = append(conditions, types.PodCondition{
				Type:               string(condition.Type),
				Status:             string(condition.Status),
				LastProbeTime:      condition.LastProbeTime.Format(time.RFC3339),
				LastTransitionTime: condition.LastTransitionTime.Format(time.RFC3339),
				Reason:             condition.Reason,
				Message:            condition.Message,
			})
		}

		// Build volumes
		var volumes []types.VolumeInfo
		for _, volume := range pod.Spec.Volumes {
			volumeInfo := types.VolumeInfo{
				Name:         volume.Name,
				VolumeSource: make(map[string]interface{}),
			}
			// Note: Volume source details would need to be implemented based on specific volume types
			volumes = append(volumes, volumeInfo)
		}

		detailedPod := types.DetailedPod{
			PodStatus:          basicPodStatus,
			Labels:             pod.Labels,
			Annotations:        pod.Annotations,
			NodeName:           pod.Spec.NodeName,
			HostNetwork:        pod.Spec.HostNetwork,
			DNSPolicy:          string(pod.Spec.DNSPolicy),
			RestartPolicy:      string(pod.Spec.RestartPolicy),
			SchedulerName:      pod.Spec.SchedulerName,
			ServiceAccountName: pod.Spec.ServiceAccountName,
			Conditions:         conditions,
			Volumes:            volumes,
			QOSClass:           string(pod.Status.QOSClass),
			Priority:           pod.Spec.Priority,
			PriorityClassName:  pod.Spec.PriorityClassName,
		}

		detailedPods = append(detailedPods, detailedPod)
	}

	// Build events
	var eventInfos []types.Event
	for _, event := range events.Items {
		eventInfo := types.Event{
			Type:      event.Type,
			Reason:    event.Reason,
			Message:   event.Message,
			Source:    fmt.Sprintf("%s/%s", event.Source.Component, event.Source.Host),
			Count:     event.Count,
			FirstTime: event.FirstTimestamp.Format(time.RFC3339),
			LastTime:  event.LastTimestamp.Format(time.RFC3339),
			InvolvedObject: types.EventInvolvedObject{
				Kind:      event.InvolvedObject.Kind,
				Name:      event.InvolvedObject.Name,
				Namespace: event.InvolvedObject.Namespace,
				UID:       string(event.InvolvedObject.UID),
			},
		}
		eventInfos = append(eventInfos, eventInfo)
	}

	// Build configmaps
	var configMapInfos []types.ConfigMapInfo
	for _, configMap := range configMaps.Items {
		configMapInfo := types.ConfigMapInfo{
			Name:        configMap.Name,
			Namespace:   configMap.Namespace,
			Labels:      configMap.Labels,
			Annotations: configMap.Annotations,
			Data:        configMap.Data,
			BinaryData:  make(map[string]string),
			CreatedAt:   configMap.CreationTimestamp.Format(time.RFC3339),
		}
		// Convert binary data to base64 strings
		for key, value := range configMap.BinaryData {
			configMapInfo.BinaryData[key] = string(value)
		}
		configMapInfos = append(configMapInfos, configMapInfo)
	}

	// Build secrets
	var secretInfos []types.SecretInfo
	for _, secret := range secrets.Items {
		secretInfo := types.SecretInfo{
			Name:        secret.Name,
			Namespace:   secret.Namespace,
			Labels:      secret.Labels,
			Annotations: secret.Annotations,
			Type:        string(secret.Type),
			Data:        make(map[string]string),
			CreatedAt:   secret.CreationTimestamp.Format(time.RFC3339),
		}
		// Convert binary data to base64 strings
		for key, value := range secret.Data {
			secretInfo.Data[key] = string(value)
		}
		secretInfos = append(secretInfos, secretInfo)
	}

	// Build persistent volumes
	var persistentVolumeInfos []types.PersistentVolumeInfo
	for _, pvc := range pvcs.Items {
		// Find corresponding persistent volume
		for _, pv := range persistentVolumes {
			if pv.Spec.ClaimRef != nil && pv.Spec.ClaimRef.Name == pvc.Name && pv.Spec.ClaimRef.Namespace == pvc.Namespace {
				persistentVolumeInfo := types.PersistentVolumeInfo{
					Name:                   pv.Name,
					Namespace:              pvc.Namespace,
					Labels:                 pv.Labels,
					Annotations:            pv.Annotations,
					Capacity:               make(map[string]string),
					AccessModes:            []string{},
					PersistentVolumeSource: make(map[string]interface{}),
					Status:                 string(pv.Status.Phase),
					CreatedAt:              pv.CreationTimestamp.Format(time.RFC3339),
				}

				// Build capacity
				for resource, quantity := range pv.Spec.Capacity {
					persistentVolumeInfo.Capacity[string(resource)] = quantity.String()
				}

				// Build access modes
				for _, accessMode := range pv.Spec.AccessModes {
					persistentVolumeInfo.AccessModes = append(persistentVolumeInfo.AccessModes, string(accessMode))
				}

				// Build claim reference
				if pv.Spec.ClaimRef != nil {
					persistentVolumeInfo.ClaimRef = &types.ObjectReference{
						Kind:      pv.Spec.ClaimRef.Kind,
						Name:      pv.Spec.ClaimRef.Name,
						Namespace: pv.Spec.ClaimRef.Namespace,
						UID:       string(pv.Spec.ClaimRef.UID),
					}
				}

				persistentVolumeInfos = append(persistentVolumeInfos, persistentVolumeInfo)
				break
			}
		}
	}

	// Create detailed deployment status
	detailedStatus := &types.DetailedDeploymentStatus{
		DeploymentStatus:  basicStatus,
		Namespace:         namespaceInfo,
		Deployments:       detailedDeployments,
		Services:          detailedServices,
		Pods:              detailedPods,
		Events:            eventInfos,
		ConfigMaps:        configMapInfos,
		Secrets:           secretInfos,
		PersistentVolumes: persistentVolumeInfos,
	}

	return detailedStatus, nil
}

// InspectDeployment retrieves detailed deployment information similar to kubectl describe
func (c *KubeClient) InspectDeployment(ctx context.Context, deploymentID string) (*types.DetailedDeploymentStatus, error) {
	return c.inspectDeploymentStatusInternal(ctx, deploymentID)
}
