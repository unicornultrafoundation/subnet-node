package kube

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/manifest"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	pathTypePrefix = networkingv1.PathTypePrefix
)

// Deploy deploys a manifest to Kubernetes
func (c *KubeClient) Deploy(ctx context.Context, deployment types.Deployment) error {
	// Validate manifest
	if err := c.validateManifest(deployment.Manifest); err != nil {
		return fmt.Errorf("manifest validation failed: %w", err)
	}

	// Calculate version hash for deterministic resource naming
	versionHash, err := deployment.Manifest.Version()
	if err != nil {
		return fmt.Errorf("failed to calculate version hash: %w", err)
	}

	// Include deployment ID in version to make it unique
	combinedData := append(versionHash, []byte(deployment.ID)...)
	combinedHash := sha256.Sum256(combinedData)
	versionStr := fmt.Sprintf("%x", combinedHash)

	// Track created resources for cleanup in case of failure
	var createdResources []struct {
		kind      string
		name      string
		namespace string
	}
	defer func() {
		if err != nil {
			c.Logger.WithField("namespace", deployment.ID).WithField("version", versionStr).Error("deployment failed, cleaning up resources", err)
			c.cleanupResources(ctx, createdResources)
		}
	}()

	// Validate namespace and create if needed
	if err := c.validateAndCreateNamespace(ctx, deployment.ID, versionStr, deployment.ID, deployment.Requester.Hex(), deployment.TTL); err != nil {
		return fmt.Errorf("namespace validation failed: %w", err)
	}

	// Validate compute profiles and resources
	if err := c.validateComputeProfiles(ctx, deployment.Manifest.Profiles); err != nil {
		return fmt.Errorf("compute profile validation failed: %w", err)
	}

	// Create registry secrets
	for _, group := range deployment.Manifest.Groups {
		for _, service := range group.Services {
			if err := c.createRegistrySecret(ctx, service, deployment.ID); err != nil {
				return fmt.Errorf("failed to create registry secret: %w", err)
			}
			createdResources = append(createdResources, struct {
				kind      string
				name      string
				namespace string
			}{
				kind:      "Secret",
				name:      fmt.Sprintf("registry-auth-%s", getServiceName(service)),
				namespace: deployment.ID,
			})
		}
	}

	// Create a map to track service deployment status
	deployedServices := make(map[string]bool)

	// Function to check if a service's dependencies are deployed
	areDependenciesDeployed := func(service manifest.Service, groupName string) bool {
		for _, dep := range service.DependsOn {
			depKey := fmt.Sprintf("%s-%s", groupName, dep)
			if !deployedServices[depKey] {
				return false
			}
		}
		return true
	}

	// Deploy services in dependency order
	for _, group := range deployment.Manifest.Groups {
		// Keep track of services that need to be deployed
		remainingServices := make(map[string]manifest.Service)
		for _, service := range group.Services {
			remainingServices[service.Image] = service
		}

		// Deploy services until all are deployed
		for len(remainingServices) > 0 {
			deployedInThisRound := false
			for serviceName, service := range remainingServices {
				if areDependenciesDeployed(service, group.Name) {
					// Get compute profile for the service
					var computeProfile *manifest.ComputeProfile
					if len(deployment.Manifest.Profiles.Compute) > 0 {
						if profile, ok := deployment.Manifest.Profiles.Compute[service.Image]; ok {
							computeProfile = &profile
						}
					}

					// Create deployment
					if err := c.createDeployment(ctx, service, group, computeProfile, deployment.ID, versionStr, deployment.ID, deployment.Requester.Hex(), deployment.TTL); err != nil {
						return fmt.Errorf("failed to create deployment for service %s: %w", serviceName, err)
					}
					createdResources = append(createdResources, struct {
						kind      string
						name      string
						namespace string
					}{
						kind:      "Deployment",
						name:      fmt.Sprintf("%s-%s-%s", group.Name, getServiceName(service), versionStr[:8]),
						namespace: deployment.ID,
					})

					// Create services for exposed ports
					for _, expose := range service.Expose {
						if len(expose.To) > 0 {
							if err := c.createService(ctx, service, group, expose, deployment.ID, versionStr, deployment.ID, deployment.Requester.Hex(), deployment.TTL); err != nil {
								return fmt.Errorf("failed to create service for port %d: %w", expose.Port, err)
							}
							createdResources = append(createdResources, struct {
								kind      string
								name      string
								namespace string
							}{
								kind:      "Service",
								name:      fmt.Sprintf("%s-%s-%d-%s", group.Name, getServiceName(service), expose.Port, versionStr[:8]),
								namespace: deployment.ID,
							})
						}
					}

					// Create endpoints if specified
					if deployment.Manifest.Endpoints != nil {
						for endpointName, endpoint := range deployment.Manifest.Endpoints {
							if err := c.createIngress(ctx, service, group, endpoint, endpointName, deployment.ID, deployment.ID, deployment.Requester.Hex(), deployment.TTL); err != nil {
								return fmt.Errorf("failed to create ingress for endpoint %s: %w", endpointName, err)
							}
							createdResources = append(createdResources, struct {
								kind      string
								name      string
								namespace string
							}{
								kind:      "Ingress",
								name:      fmt.Sprintf("%s-%s", group.Name, endpointName),
								namespace: deployment.ID,
							})
						}
					}

					// Mark service as deployed
					deployedServices[fmt.Sprintf("%s-%s", group.Name, serviceName)] = true
					delete(remainingServices, serviceName)
					deployedInThisRound = true
				}
			}

			// If no services were deployed in this round, we have a circular dependency
			if !deployedInThisRound {
				return fmt.Errorf("circular dependency detected in service dependencies")
			}
		}
	}

	return nil
}

// validateService validates a service configuration
func (c *KubeClient) validateService(service manifest.Service, groupName string, groupServices []manifest.Service) error {
	if service.Image == "" {
		return fmt.Errorf("service %s: image is required", service.Image)
	}
	if service.Count <= 0 {
		return fmt.Errorf("service %s: count must be positive", service.Image)
	}

	// Validate expose configuration
	for _, expose := range service.Expose {
		if expose.Port <= 0 {
			return fmt.Errorf("service %s: invalid port number %d", service.Image, expose.Port)
		}
		if expose.Proto != "" && expose.Proto != "TCP" && expose.Proto != "UDP" {
			return fmt.Errorf("service %s: invalid protocol %s", service.Image, expose.Proto)
		}
	}

	// Validate health checks
	if service.Params != nil && service.Params.Health != nil {
		if err := c.validateHealthCheck(service.Params.Health.Readiness, "readiness", service.Image); err != nil {
			return err
		}
		if err := c.validateHealthCheck(service.Params.Health.Liveness, "liveness", service.Image); err != nil {
			return err
		}
	}

	// Validate dependencies
	for _, dep := range service.DependsOn {
		depFound := false
		for _, otherService := range groupServices {
			if otherService.Image == dep {
				depFound = true
				break
			}
		}
		if !depFound {
			return fmt.Errorf("service %s: dependency %s not found in group %s", service.Image, dep, groupName)
		}
	}

	return nil
}

// validateHealthCheck validates a health check configuration
func (c *KubeClient) validateHealthCheck(check *manifest.HealthCheck, checkType, serviceName string) error {
	if check == nil || check.HTTP == nil {
		return nil
	}
	if check.HTTP.Path == "" {
		return fmt.Errorf("service %s: HTTP %s check path is required", serviceName, checkType)
	}
	if check.HTTP.Port <= 0 {
		return fmt.Errorf("service %s: HTTP %s check port must be positive", serviceName, checkType)
	}
	return nil
}

// createRegistrySecret creates a registry authentication secret
func (c *KubeClient) createRegistrySecret(ctx context.Context, service manifest.Service, namespace string) error {
	if service.Credentials == nil {
		return nil
	}

	secretName := fmt.Sprintf("registry-auth-%s", getServiceName(service))
	dockerConfig := fmt.Sprintf(`{"auths":{"%s":{"auth":"%s"}}}`,
		service.Credentials.Host,
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", service.Credentials.Username, service.Credentials.Password))))

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			".dockerconfigjson": []byte(dockerConfig),
		},
	}

	_, err := c.Client.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create registry secret: %w", err)
	}

	// Add the secret to imagePullSecrets if not already present
	found := false
	for _, existingSecret := range service.ImagePullSecrets {
		if existingSecret.Name == secretName {
			found = true
			break
		}
	}
	if !found {
		service.ImagePullSecrets = append(service.ImagePullSecrets, manifest.ImagePullSecret{
			Name: secretName,
		})
	}

	return nil
}

// createStorageVolume creates a storage volume for a service
func (c *KubeClient) createStorageVolume(ctx context.Context, service manifest.Service, computeProfile *manifest.ComputeProfile, namespace, versionStr string) (*corev1.Volume, []corev1.VolumeMount, error) {
	if service.Resources == nil || service.Resources.Storage == nil {
		return nil, nil, nil
	}

	volumeName := fmt.Sprintf("storage-%s", getServiceName(service))
	var storageSize string
	var isPersistent bool
	if computeProfile != nil && len(computeProfile.Resources.Storage) > 0 {
		for _, storage := range computeProfile.Resources.Storage {
			if storage.Name != "" {
				volumeName = storage.Name
			}
			if storage.Size != "" {
				storageSize = storage.Size
			}
			if storage.Attributes != nil {
				isPersistent = storage.Attributes.Persistent
			}
		}
	}

	volumeMount := corev1.VolumeMount{
		Name:      volumeName,
		MountPath: fmt.Sprintf("/data/%s", volumeName),
	}

	volume := corev1.Volume{
		Name: volumeName,
	}

	if isPersistent && (service.Resources.Storage.Size.Value > 0 || storageSize != "") {
		if storageSize == "" {
			storageSize = fmt.Sprintf("%d%s", service.Resources.Storage.Size.Value, service.Resources.Storage.Size.Unit)
			if service.Resources.Storage.Size.Unit == "" {
				storageSize = fmt.Sprintf("%dGi", service.Resources.Storage.Size.Value)
			}
		}

		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      volumeName,
				Namespace: namespace,
				Labels: map[string]string{
					"version": versionStr,
				},
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(storageSize),
					},
				},
			},
		}

		if computeProfile != nil && len(computeProfile.Resources.Storage) > 0 {
			for _, storage := range computeProfile.Resources.Storage {
				if storage.Attributes != nil && storage.Attributes.Class != "" {
					pvc.Spec.StorageClassName = &storage.Attributes.Class
					break
				}
			}
		}

		_, err := c.Client.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create PVC for volume %s: %w", volumeName, err)
		}

		volume.VolumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: volumeName,
			},
		}
	} else {
		volume.VolumeSource = corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		}
	}

	return &volume, []corev1.VolumeMount{volumeMount}, nil
}

// createHealthCheckProbes creates health check probes for a service
func (c *KubeClient) createHealthCheckProbes(service manifest.Service) (*corev1.Probe, *corev1.Probe) {
	if service.Params == nil || service.Params.Health == nil {
		return nil, nil
	}

	var readinessProbe, livenessProbe *corev1.Probe

	if service.Params.Health.Readiness != nil {
		readinessProbe = &corev1.Probe{
			InitialDelaySeconds: service.Params.Health.Readiness.InitialDelaySeconds,
			PeriodSeconds:       service.Params.Health.Readiness.PeriodSeconds,
			TimeoutSeconds:      service.Params.Health.Readiness.TimeoutSeconds,
			SuccessThreshold:    service.Params.Health.Readiness.SuccessThreshold,
			FailureThreshold:    service.Params.Health.Readiness.FailureThreshold,
		}
		if service.Params.Health.Readiness.HTTP != nil {
			readinessProbe.HTTPGet = &corev1.HTTPGetAction{
				Path: service.Params.Health.Readiness.HTTP.Path,
				Port: intstr.FromInt(int(service.Params.Health.Readiness.HTTP.Port)),
			}
		}
	}

	if service.Params.Health.Liveness != nil {
		livenessProbe = &corev1.Probe{
			InitialDelaySeconds: service.Params.Health.Liveness.InitialDelaySeconds,
			PeriodSeconds:       service.Params.Health.Liveness.PeriodSeconds,
			TimeoutSeconds:      service.Params.Health.Liveness.TimeoutSeconds,
			SuccessThreshold:    service.Params.Health.Liveness.SuccessThreshold,
			FailureThreshold:    service.Params.Health.Liveness.FailureThreshold,
		}
		if service.Params.Health.Liveness.HTTP != nil {
			livenessProbe.HTTPGet = &corev1.HTTPGetAction{
				Path: service.Params.Health.Liveness.HTTP.Path,
				Port: intstr.FromInt(int(service.Params.Health.Liveness.HTTP.Port)),
			}
		}
	}

	return readinessProbe, livenessProbe
}

// createDeployment creates a Kubernetes deployment for a service
func (c *KubeClient) createDeployment(ctx context.Context, service manifest.Service, group manifest.Group, computeProfile *manifest.ComputeProfile, namespace, versionStr string, deploymentID string, requester string, ttl int64) error {
	resources, err := c.createResourceRequirements(service, computeProfile)
	if err != nil {
		return fmt.Errorf("failed to create resource requirements: %w", err)
	}

	containerPorts := c.createContainerPorts(service)
	envVars := c.createEnvironmentVariables(service)
	volume, volumeMounts, err := c.createStorageVolume(ctx, service, computeProfile, namespace, versionStr)
	if err != nil {
		return fmt.Errorf("failed to create storage volume: %w", err)
	}

	readinessProbe, livenessProbe := c.createHealthCheckProbes(service)

	volumes := []corev1.Volume{}
	if volume != nil {
		volumes = append(volumes, *volume)
	}

	deploymentName := fmt.Sprintf("%s-%s-%s", group.Name, getServiceName(service), versionStr[:8])
	k8sDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      truncateLabelValue(deploymentName),
			Namespace: namespace,
			Labels: map[string]string{
				"app":          group.Name,
				"service":      getServiceName(service),
				"version":      truncateLabelValue(versionStr),
				"deploymentID": truncateLabelValue(deploymentID),
				"requester":    requester,
			},
			Annotations: map[string]string{
				"deploymentID": deploymentID,
				"ttl":          strconv.FormatInt(ttl, 10),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &service.Count,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":          group.Name,
					"service":      getServiceName(service),
					"version":      truncateLabelValue(versionStr),
					"deploymentID": truncateLabelValue(deploymentID),
					"requester":    requester,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":          group.Name,
						"service":      getServiceName(service),
						"version":      truncateLabelValue(versionStr),
						"deploymentID": truncateLabelValue(deploymentID),
						"requester":    requester,
					},
					Annotations: map[string]string{
						"deploymentID": deploymentID,
						"ttl":          strconv.FormatInt(ttl, 10),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            getServiceName(service),
							Image:           service.Image,
							Command:         service.Command,
							Args:            service.Args,
							Ports:           containerPorts,
							Env:             envVars,
							Resources:       resources,
							ImagePullPolicy: corev1.PullIfNotPresent,
							VolumeMounts:    volumeMounts,
							ReadinessProbe:  readinessProbe,
							LivenessProbe:   livenessProbe,
						},
					},
					Volumes: volumes,
				},
			},
		},
	}

	// Add network bandwidth configuration if specified
	if service.Resources != nil && service.Resources.Network != nil && service.Resources.Network.Bandwidth != "" {
		k8sDeployment.Spec.Template.Annotations["k8s.v1.cni.cncf.io/networks"] = fmt.Sprintf(`[{
			"name": "default",
			"bandwidth": "%s"
		}]`, service.Resources.Network.Bandwidth)
	}

	if service.SchedulerParams != nil && service.SchedulerParams.RuntimeClass != "" {
		k8sDeployment.Spec.Template.Spec.RuntimeClassName = &service.SchedulerParams.RuntimeClass
	}

	if len(service.ImagePullSecrets) > 0 {
		secretRefs := make([]corev1.LocalObjectReference, len(service.ImagePullSecrets))
		for i, secret := range service.ImagePullSecrets {
			secretRefs[i] = corev1.LocalObjectReference{Name: secret.Name}
		}
		k8sDeployment.Spec.Template.Spec.ImagePullSecrets = secretRefs
	}

	_, err = c.Client.AppsV1().Deployments(namespace).Create(ctx, k8sDeployment, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create deployment for service %s: %w", service.Image, err)
	}

	return nil
}

// createResourceRequirements creates Kubernetes resource requirements
func (c *KubeClient) createResourceRequirements(service manifest.Service, computeProfile *manifest.ComputeProfile) (corev1.ResourceRequirements, error) {
	resources := corev1.ResourceRequirements{}
	if service.Resources != nil {
		k8sResources, err := service.Resources.ToK8sResources()
		if err != nil {
			return resources, fmt.Errorf("failed to convert resources: %w", err)
		}

		if len(k8sResources) > 0 {
			resources.Requests = make(corev1.ResourceList)
			for name, quantity := range k8sResources {
				resources.Requests[corev1.ResourceName(name)] = quantity
			}
		}

		if computeProfile != nil {
			profileResources, err := computeProfile.Resources.ToK8sResources()
			if err != nil {
				return resources, fmt.Errorf("failed to convert compute profile resources: %w", err)
			}

			for resourceName, quantity := range profileResources {
				if resources.Requests == nil {
					resources.Requests = corev1.ResourceList{}
				}
				resources.Requests[corev1.ResourceName(resourceName)] = quantity
			}

			if computeProfile.Resources.CPU.Limit != "" {
				cpuLimit, err := resource.ParseQuantity(computeProfile.Resources.CPU.Limit)
				if err != nil {
					return resources, fmt.Errorf("invalid CPU limit in compute profile: %w", err)
				}
				if resources.Limits == nil {
					resources.Limits = make(corev1.ResourceList)
				}
				resources.Limits[corev1.ResourceCPU] = cpuLimit
			}

			if computeProfile.Resources.Memory.Limit != "" {
				memLimit, err := resource.ParseQuantity(computeProfile.Resources.Memory.Limit)
				if err != nil {
					return resources, fmt.Errorf("invalid memory limit in compute profile: %w", err)
				}
				if resources.Limits == nil {
					resources.Limits = make(corev1.ResourceList)
				}
				resources.Limits[corev1.ResourceMemory] = memLimit
			}
		}
	}
	return resources, nil
}

// createContainerPorts creates container ports from expose configuration
func (c *KubeClient) createContainerPorts(service manifest.Service) []corev1.ContainerPort {
	var containerPorts []corev1.ContainerPort
	for _, expose := range service.Expose {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			ContainerPort: expose.Port,
			Protocol:      corev1.Protocol(expose.Proto),
		})
	}
	return containerPorts
}

// createEnvironmentVariables creates environment variables from service configuration
func (c *KubeClient) createEnvironmentVariables(service manifest.Service) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	for _, env := range service.Env {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envVars = append(envVars, corev1.EnvVar{
				Name:  parts[0],
				Value: parts[1],
			})
		}
	}
	return envVars
}

// createService creates a Kubernetes service for an exposed port
func (c *KubeClient) createService(ctx context.Context, service manifest.Service, group manifest.Group, expose manifest.Expose, namespace, versionStr string, deploymentID string, requester string, ttl int64) error {
	serviceName := fmt.Sprintf("%s-%s-%d-%s", group.Name, getServiceName(service), expose.Port, versionStr[:8])
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      truncateLabelValue(serviceName),
			Namespace: namespace,
			Labels: map[string]string{
				"app":          group.Name,
				"service":      getServiceName(service),
				"version":      truncateLabelValue(versionStr),
				"deploymentID": truncateLabelValue(deploymentID),
				"requester":    requester,
			},
			Annotations: map[string]string{
				"deploymentID": deploymentID,
				"ttl":          strconv.FormatInt(ttl, 10),
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":          group.Name,
				"service":      getServiceName(service),
				"version":      truncateLabelValue(versionStr),
				"deploymentID": truncateLabelValue(deploymentID),
				"requester":    requester,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       expose.As,
					TargetPort: intstr.FromInt(int(expose.Port)),
					Protocol:   corev1.Protocol(expose.Proto),
				},
			},
		},
	}

	// Add accept rules as annotations
	if len(expose.Accept) > 0 {
		svc.ObjectMeta.Annotations["ingress.kubernetes.io/whitelist-source-range"] = strings.Join(expose.Accept, ",")
	}

	// Add HTTP options as annotations if specified
	if expose.HTTPOptions != nil {
		if expose.HTTPOptions.MaxBodySize > 0 {
			svc.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/proxy-body-size"] = fmt.Sprintf("%d", expose.HTTPOptions.MaxBodySize)
		}
		if len(expose.HTTPOptions.NextCases) > 0 {
			svc.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/configuration-snippet"] = fmt.Sprintf("set $next_cases %s;", strings.Join(expose.HTTPOptions.NextCases, " "))
		}
	}

	// Set service type based on exposure and config
	for _, to := range expose.To {
		if to.Global {
			// Use the configured service type
			switch c.DefaultServiceType {
			case "NodePort":
				svc.Spec.Type = corev1.ServiceTypeNodePort
				// Add localhost binding annotation if enabled
				if c.LocalhostEnabled {
					svc.ObjectMeta.Annotations["service.beta.kubernetes.io/aws-load-balancer-internal"] = "true"
					svc.ObjectMeta.Annotations["service.beta.kubernetes.io/aws-load-balancer-scheme"] = "internal"
				}
			case "ClusterIP":
				svc.Spec.Type = corev1.ServiceTypeClusterIP
			default:
				svc.Spec.Type = corev1.ServiceTypeLoadBalancer
			}
			break
		} else if to.Service != "" {
			svc.ObjectMeta.Labels["target-service"] = to.Service
		}
	}

	_, err := c.Client.CoreV1().Services(namespace).Create(ctx, svc, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create service for port %d: %w", expose.Port, err)
	}

	return nil
}

// createIngress creates a Kubernetes ingress for an endpoint
func (c *KubeClient) createIngress(ctx context.Context, service manifest.Service, group manifest.Group, endpoint manifest.Endpoint, endpointName, namespace string, deploymentID string, requester string, ttl int64) error {
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      truncateLabelValue(fmt.Sprintf("%s-%s", group.Name, endpointName)),
			Namespace: namespace,
			Labels: map[string]string{
				"app":          group.Name,
				"service":      getServiceName(service),
				"endpoint":     endpointName,
				"deploymentID": truncateLabelValue(deploymentID),
				"requester":    requester,
			},
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": endpoint.Kind,
				"deploymentID":                truncateLabelValue(deploymentID),
				"ttl":                         strconv.FormatInt(ttl, 10),
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: fmt.Sprintf("%s.%s", endpointName, namespace),
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathTypePrefix,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: fmt.Sprintf("%s-%s-%d", group.Name, getServiceName(service), service.Expose[0].Port),
											Port: networkingv1.ServiceBackendPort{
												Number: service.Expose[0].Port,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Add HTTP options as annotations if specified
	if service.Expose[0].HTTPOptions != nil {
		if service.Expose[0].HTTPOptions.MaxBodySize > 0 {
			ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/proxy-body-size"] = fmt.Sprintf("%d", service.Expose[0].HTTPOptions.MaxBodySize)
		}
		if len(service.Expose[0].HTTPOptions.NextCases) > 0 {
			ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/configuration-snippet"] = fmt.Sprintf("set $next_cases %s;", strings.Join(service.Expose[0].HTTPOptions.NextCases, " "))
		}
	}

	_, err := c.Client.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create ingress for endpoint %s: %w", endpointName, err)
	}

	return nil
}

// createNamespace creates a Kubernetes namespace if it doesn't exist
func (c *KubeClient) createNamespace(ctx context.Context, namespace, versionStr string, deploymentID string, requester string, ttl int64) error {
	_, err := c.Client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"version":      truncateLabelValue(versionStr),
				"deploymentID": truncateLabelValue(deploymentID),
				"requester":    requester,
			},
			Annotations: map[string]string{
				"deploymentID": deploymentID,
				"ttl":          strconv.FormatInt(ttl, 10),
			},
		},
	}, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create namespace: %w", err)
	}
	return nil
}

// validateManifest validates the deployment manifest
func (c *KubeClient) validateManifest(ma *manifest.Manifest) error {
	if ma == nil {
		return fmt.Errorf("manifest is required")
	}

	// Validate services
	for _, group := range ma.Groups {
		if !manifest.ValidateServiceName(group.Name) {
			return fmt.Errorf("invalid group name %s", group.Name)
		}
		for _, service := range group.Services {
			if err := c.validateService(service, group.Name, group.Services); err != nil {
				return err
			}
		}
	}

	// Validate endpoints
	if ma.Endpoints != nil {
		for endpointName := range ma.Endpoints {
			if !manifest.ValidateEndpointName(endpointName) {
				return fmt.Errorf("invalid endpoint name %s", endpointName)
			}
		}
	}

	return nil
}

// cleanupResources cleans up resources created during a failed deployment
func (c *KubeClient) cleanupResources(ctx context.Context, resources []struct {
	kind      string
	name      string
	namespace string
}) {
	for _, resource := range resources {
		switch resource.kind {
		case "Deployment":
			err := c.Client.AppsV1().Deployments(resource.namespace).Delete(ctx, resource.name, metav1.DeleteOptions{})
			if err != nil {
				c.Logger.WithField("name", resource.name).WithField("namespace", resource.namespace).Error("failed to cleanup deployment", err)
			}
		case "Service":
			err := c.Client.CoreV1().Services(resource.namespace).Delete(ctx, resource.name, metav1.DeleteOptions{})
			if err != nil {
				c.Logger.WithField("name", resource.name).WithField("namespace", resource.namespace).Error("failed to cleanup service", err)
			}
		case "Ingress":
			err := c.Client.NetworkingV1().Ingresses(resource.namespace).Delete(ctx, resource.name, metav1.DeleteOptions{})
			if err != nil {
				c.Logger.WithField("name", resource.name).WithField("namespace", resource.namespace).Error("failed to cleanup ingress", err)
			}
		case "Secret":
			err := c.Client.CoreV1().Secrets(resource.namespace).Delete(ctx, resource.name, metav1.DeleteOptions{})
			if err != nil {
				c.Logger.WithField("name", resource.name).WithField("namespace", resource.namespace).Error("failed to cleanup secret", err)
			}
		}
	}
}

// validateAndCreateNamespace validates and creates a namespace if needed
func (c *KubeClient) validateAndCreateNamespace(ctx context.Context, namespace, versionStr string, deploymentID string, requester string, ttl int64) error {
	// Check if namespace exists
	_, err := c.Client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Create namespace
			if err := c.createNamespace(ctx, namespace, versionStr, deploymentID, requester, ttl); err != nil {
				return fmt.Errorf("failed to create namespace: %w", err)
			}
		} else {
			return fmt.Errorf("failed to check namespace existence: %w", err)
		}
	}
	return nil
}

// validateComputeProfiles validates compute profiles and their resources
func (c *KubeClient) validateComputeProfiles(ctx context.Context, profiles manifest.Profiles) error {
	for name, profile := range profiles.Compute {
		// Validate CPU resources
		if profile.Resources.CPU.Request == "" {
			return fmt.Errorf("compute profile %s: CPU request is required", name)
		}
		if profile.Resources.CPU.Limit == "" {
			return fmt.Errorf("compute profile %s: CPU limit is required", name)
		}

		// Validate memory resources
		if profile.Resources.Memory.Request == "" {
			return fmt.Errorf("compute profile %s: memory request is required", name)
		}
		if profile.Resources.Memory.Limit == "" {
			return fmt.Errorf("compute profile %s: memory limit is required", name)
		}

		// Validate storage configuration
		for _, storage := range profile.Resources.Storage {
			if storage.Attributes != nil && storage.Attributes.Class != "" {
				// Check if storage class exists
				_, err := c.Client.StorageV1().StorageClasses().Get(ctx, storage.Attributes.Class, metav1.GetOptions{})
				if err != nil {
					return fmt.Errorf("compute profile %s: storage class %s not found", name, storage.Attributes.Class)
				}
			}
		}
	}
	return nil
}

// Helper to truncate label values to 63 chars
func truncateLabelValue(val string) string {
	if len(val) > 63 {
		return val[:63]
	}
	return val
}

// sanitizeK8sName sanitizes a name for use in Kubernetes resources
func sanitizeK8sName(name string) string {
	// If name is empty, return a default
	if name == "" {
		return "service"
	}

	// Replace invalid characters with hyphens
	re := regexp.MustCompile(`[^a-z0-9-]`)
	sanitized := re.ReplaceAllString(strings.ToLower(name), "-")

	// Remove leading/trailing hyphens
	sanitized = strings.Trim(sanitized, "-")

	// Ensure it starts with a letter or number
	if len(sanitized) > 0 && !regexp.MustCompile(`^[a-z0-9]`).MatchString(sanitized) {
		sanitized = "s-" + sanitized
	}

	// Truncate to 63 characters
	if len(sanitized) > 63 {
		sanitized = sanitized[:63]
		// Ensure it doesn't end with a hyphen
		sanitized = strings.TrimRight(sanitized, "-")
	}

	return sanitized
}

// getServiceName returns a valid service name for Kubernetes resources
func getServiceName(service manifest.Service) string {
	if service.Name != "" {
		return sanitizeK8sName(service.Name)
	}

	// Fallback to image name if service name is not set
	// Extract the last part of the image name (after the last slash)
	parts := strings.Split(service.Image, "/")
	imageName := parts[len(parts)-1]

	// Remove tag if present
	if strings.Contains(imageName, ":") {
		imageName = strings.Split(imageName, ":")[0]
	}

	return sanitizeK8sName(imageName)
}
