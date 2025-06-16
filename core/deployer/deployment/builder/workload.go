package builder

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"go.uber.org/zap"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
)

const (
	ResourceGPUNvidia = corev1.ResourceName("nvidia.com/gpu")
	ResourceGPUAMD    = corev1.ResourceName("amd.com/gpu")
	GPUVendorNvidia   = "nvidia"
	GPUVendorAMD      = "amd"
)

// Workload represents a Kubernetes workload builder
type Workload struct {
	baseBuilder
	groupIndex   int
	serviceIndex int
	log          *zap.Logger
}

// NewWorkloadBuilder creates a new workload builder
func NewWorkloadBuilder(logger *zap.Logger, settings Settings, deployment *types.ManagedDeployment, groupIndex, serviceIndex int) (*Workload, error) {
	if groupIndex < 0 || groupIndex >= len(deployment.Manifest.Groups) {
		return nil, fmt.Errorf("invalid group index: %d", groupIndex)
	}
	if serviceIndex < 0 || serviceIndex >= len(deployment.Manifest.Groups[groupIndex].Services) {
		return nil, fmt.Errorf("invalid service index: %d", serviceIndex)
	}

	return &Workload{
		baseBuilder: baseBuilder{
			settings:   settings,
			client:     settings.Client,
			ns:         deployment.Namespace,
			deployment: deployment,
		},
		groupIndex:   groupIndex,
		serviceIndex: serviceIndex,
		log:          logger.With(zap.String("module", "kube-builder")),
	}, nil
}

// Name returns the name of the workload
func (b *Workload) Name() string {
	return fmt.Sprintf("%s-group-%d-service-%d", b.deployment.Name, b.groupIndex, b.serviceIndex)
}

// NS returns the namespace of the workload
func (b *Workload) NS() string {
	return b.ns
}

func (b *Workload) container() corev1.Container {
	falseValue := false

	service := b.deployment.Manifest.Groups[b.groupIndex].Services[b.serviceIndex]

	kcontainer := corev1.Container{
		Name:    b.Name(),
		Image:   service.Image,
		Command: service.Command,
		Args:    service.Args,
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("50m"),
				corev1.ResourceMemory: resource.MustParse("128Mi"),
			},
		},
		ImagePullPolicy: corev1.PullIfNotPresent,
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot:             &falseValue,
			Privileged:               &falseValue,
			AllowPrivilegeEscalation: &falseValue,
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{
					"ALL",
				},
			},
		},
	}

	// Configure health checks if specified in the SDL
	if service.Params != nil && service.Params.Health != nil {
		// Configure readiness probe
		if service.Params.Health.Readiness != nil {
			kcontainer.ReadinessProbe = &corev1.Probe{
				InitialDelaySeconds: service.Params.Health.Readiness.InitialDelaySeconds,
				PeriodSeconds:       service.Params.Health.Readiness.PeriodSeconds,
				TimeoutSeconds:      service.Params.Health.Readiness.TimeoutSeconds,
				SuccessThreshold:    service.Params.Health.Readiness.SuccessThreshold,
				FailureThreshold:    service.Params.Health.Readiness.FailureThreshold,
			}

			if service.Params.Health.Readiness.HTTP != nil {
				kcontainer.ReadinessProbe.ProbeHandler = corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: service.Params.Health.Readiness.HTTP.Path,
						Port: intstr.FromInt(int(service.Params.Health.Readiness.HTTP.Port)),
					},
				}
			}
		}

		// Configure liveness probe
		if service.Params.Health.Liveness != nil {
			kcontainer.LivenessProbe = &corev1.Probe{
				InitialDelaySeconds: service.Params.Health.Liveness.InitialDelaySeconds,
				PeriodSeconds:       service.Params.Health.Liveness.PeriodSeconds,
				TimeoutSeconds:      service.Params.Health.Liveness.TimeoutSeconds,
				SuccessThreshold:    service.Params.Health.Liveness.SuccessThreshold,
				FailureThreshold:    service.Params.Health.Liveness.FailureThreshold,
			}

			if service.Params.Health.Liveness.HTTP != nil {
				kcontainer.LivenessProbe.ProbeHandler = corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: service.Params.Health.Liveness.HTTP.Path,
						Port: intstr.FromInt(int(service.Params.Health.Liveness.HTTP.Port)),
					},
				}
			}
		}
	} else {
		// Default health checks if not specified
		// Only set up health checks if there are exposed ports
		if len(service.Expose) > 0 {
			healthCheckPort := service.Expose[0].Port
			kcontainer.ReadinessProbe = &corev1.Probe{
				InitialDelaySeconds: 10,
				PeriodSeconds:       10,
				TimeoutSeconds:      5,
				SuccessThreshold:    1,
				FailureThreshold:    3,
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/health", // More standard health check path
						Port: intstr.FromInt(int(healthCheckPort)),
					},
				},
			}
			kcontainer.LivenessProbe = &corev1.Probe{
				InitialDelaySeconds: 20,
				PeriodSeconds:       10,
				TimeoutSeconds:      5,
				SuccessThreshold:    1,
				FailureThreshold:    3,
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/health", // More standard health check path
						Port: intstr.FromInt(int(healthCheckPort)),
					},
				},
			}
		}
	}

	// Handle GPU resources
	if service.Resources.GPU != nil {
		var resourceName corev1.ResourceName
		for vendor := range service.Resources.GPU.Attributes.Vendor {
			switch vendor {
			case GPUVendorNvidia:
				resourceName = ResourceGPUNvidia
			case GPUVendorAMD:
				resourceName = ResourceGPUAMD
			default:
				b.log.Warn("unsupported GPU vendor requested", zap.String("vendor", vendor))
				continue
			}
			kcontainer.Resources.Requests[resourceName] = resource.NewQuantity(int64(service.Resources.GPU.Units), resource.DecimalSI).DeepCopy()
			kcontainer.Resources.Limits[resourceName] = resource.NewQuantity(int64(service.Resources.GPU.Units), resource.DecimalSI).DeepCopy()
		}
	}

	// Handle storage resources
	if service.Resources.Storage != nil && service.Resources.Storage.Size.Value > 0 {
		// Handle ephemeral storage
		requestedStorage := service.Resources.Storage.Size.Value
		// Convert to bytes if the unit is Mi
		if service.Resources.Storage.Size.Unit == "Mi" {
			requestedStorage = requestedStorage * 1024 * 1024
		}
		kcontainer.Resources.Requests[corev1.ResourceEphemeralStorage] = resource.NewQuantity(requestedStorage, resource.DecimalSI).DeepCopy()
		kcontainer.Resources.Limits[corev1.ResourceEphemeralStorage] = resource.NewQuantity(requestedStorage, resource.DecimalSI).DeepCopy()

		// Add volume mount for persistent storage
		kcontainer.VolumeMounts = append(kcontainer.VolumeMounts, corev1.VolumeMount{
			Name:      fmt.Sprintf("%s-storage", b.Name()),
			MountPath: "/data",
			ReadOnly:  false,
		})
	}

	// Handle storage mounts
	if service.Params != nil && service.Params.Storage != nil {
		// Handle SHM storage if present
		if service.Params.Storage.SHM != nil {
			kcontainer.VolumeMounts = append(kcontainer.VolumeMounts, corev1.VolumeMount{
				Name:      fmt.Sprintf("%s-shm", b.Name()),
				ReadOnly:  false,
				MountPath: service.Params.Storage.SHM.Mount,
			})
		}
	}

	// Handle environment variables
	envVarsAdded := make(map[string]int)
	for _, env := range service.Env {
		parts := strings.SplitN(env, "=", 2)
		switch len(parts) {
		case 2:
			kcontainer.Env = append(kcontainer.Env, corev1.EnvVar{Name: parts[0], Value: parts[1]})
		case 1:
			kcontainer.Env = append(kcontainer.Env, corev1.EnvVar{Name: parts[0]})
		}
		envVarsAdded[parts[0]] = 0
	}
	kcontainer.Env = b.addEnvVarsForDeployment(envVarsAdded, kcontainer.Env)

	// Handle ports
	for _, expose := range service.Expose {
		kcontainer.Ports = append(kcontainer.Ports, corev1.ContainerPort{
			ContainerPort: expose.Port,
			Name:          fmt.Sprintf("port-%d", expose.Port),
			Protocol:      corev1.ProtocolTCP,
		})
	}

	return kcontainer
}

// volumes returns the volumes for the workload
func (b *Workload) volumes() []corev1.Volume {
	var volumes []corev1.Volume

	service := b.deployment.Manifest.Groups[b.groupIndex].Services[b.serviceIndex]

	// Add SHM volume if configured
	if service.Params != nil && service.Params.Storage != nil && service.Params.Storage.SHM != nil {
		volumes = append(volumes, corev1.Volume{
			Name: fmt.Sprintf("%s-shm", b.Name()),
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	// Add volumes for persistent storage
	if service.Resources.Storage != nil && service.Resources.Storage.Size.Value > 0 {
		volumes = append(volumes, corev1.Volume{
			Name: fmt.Sprintf("%s-storage", b.Name()),
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: fmt.Sprintf("%s-storage", b.Name()),
				},
			},
		})
	}

	return volumes
}

// PersistentVolumeClaims returns the persistent volume claims for the workload
func (b *Workload) PersistentVolumeClaims() []corev1.PersistentVolumeClaim {
	service := b.deployment.Manifest.Groups[b.groupIndex].Services[b.serviceIndex]
	if service.Resources.Storage == nil || service.Resources.Storage.Size.Value == 0 {
		return nil
	}

	storageSize := service.Resources.Storage.Size.Value
	storageClass := "local-path" // Use local-path storage class for persistent storage
	storageQuantity := *resource.NewQuantity(storageSize, resource.DecimalSI)

	return []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-storage", b.Name()),
				Namespace: b.ns,
				Labels: map[string]string{
					"app": b.Name(),
				},
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				StorageClassName: &storageClass,
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: storageQuantity,
					},
				},
			},
		},
	}
}

// labels returns the labels for the workload
func (b *Workload) labels() map[string]string {
	return map[string]string{
		"app":                  b.Name(),
		"subnet-node/lease-id": b.deployment.ID,
		"subnet-node/owner":    b.deployment.Requester.Hex(),
	}
}

// selectorLabels returns the selector labels for the workload
func (b *Workload) selectorLabels() map[string]string {
	return map[string]string{
		"app": b.Name(),
	}
}

// imagePullSecrets returns the image pull secrets for the workload
func (b *Workload) imagePullSecrets() []corev1.LocalObjectReference {
	service := b.deployment.Manifest.Groups[b.groupIndex].Services[b.serviceIndex]
	if service.ImagePullSecrets != nil {
		refs := make([]corev1.LocalObjectReference, len(service.ImagePullSecrets))
		for i, secret := range service.ImagePullSecrets {
			refs[i] = corev1.LocalObjectReference{
				Name: secret.Name,
			}
		}
		return refs
	}
	return nil
}

// addEnvVarsForDeployment adds environment variables for the deployment
func (b *Workload) addEnvVarsForDeployment(envVarsAlreadyAdded map[string]int, env []corev1.EnvVar) []corev1.EnvVar {
	return env
}

// replicas returns the number of replicas for the workload
func (b *Workload) replicas() *int32 {
	service := b.deployment.Manifest.Groups[b.groupIndex].Services[b.serviceIndex]
	return &service.Count
}

// affinity returns the affinity for the workload
func (b *Workload) affinity() *corev1.Affinity {
	return nil
}

// runtimeClass returns the runtime class for the workload
func (b *Workload) runtimeClass() *string {
	return nil
}
