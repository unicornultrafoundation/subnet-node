package builder

import (
	"fmt"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Service represents a Kubernetes Service configuration
type Service struct {
	baseBuilder
	serviceType  corev1.ServiceType
	ports        []corev1.ServicePort
	selector     map[string]string
	annotations  map[string]string
	isGlobal     bool
	serviceIndex int
	groupIndex   int
}

// NewService creates a new Service configuration
func NewService(b baseBuilder, name string, serviceType corev1.ServiceType, ports []corev1.ServicePort, selector map[string]string, annotations map[string]string, isGlobal bool, serviceIndex int, groupIndex int) *Service {
	return &Service{
		baseBuilder:  b,
		serviceType:  serviceType,
		ports:        ports,
		selector:     selector,
		annotations:  annotations,
		isGlobal:     isGlobal,
		serviceIndex: serviceIndex,
		groupIndex:   groupIndex,
	}
}

// NewServiceBuilder creates a service builder
func NewServiceBuilder(settings Settings, deployment *types.ManagedDeployment, serviceIndex int, isGlobal bool, groupIndex int) *Service {
	return &Service{
		baseBuilder: baseBuilder{
			settings:   settings,
			deployment: deployment,
		},
		serviceType:  settings.DeploymentServiceType,
		ports:        make([]corev1.ServicePort, 0),
		selector:     make(map[string]string),
		annotations:  make(map[string]string),
		isGlobal:     isGlobal,
		serviceIndex: serviceIndex,
		groupIndex:   groupIndex,
	}
}

// Name returns the name of the Service
func (b *Service) Name() string {
	if b.isGlobal {
		return fmt.Sprintf("%s-service-%d-global", b.deployment.Name, b.serviceIndex)
	}
	return fmt.Sprintf("%s-service-%d", b.deployment.Name, b.serviceIndex)
}

// NS returns the namespace of the Service
func (b *Service) NS() string {
	return b.deployment.Namespace
}

// SetType sets the type of the Service
func (b *Service) SetType(serviceType corev1.ServiceType) {
	b.serviceType = serviceType
}

// AddPort adds a port to the Service
func (b *Service) AddPort(name string, port int32, targetPort int32, protocol corev1.Protocol) {
	b.ports = append(b.ports, corev1.ServicePort{
		Name:       name,
		Port:       port,
		TargetPort: intstr.FromInt(int(targetPort)),
		Protocol:   protocol,
	})
}

// SetSelector sets the selector for the Service
func (b *Service) SetSelector(key, value string) {
	b.selector[key] = value
}

// AddAnnotation adds an annotation to the Service
func (b *Service) AddAnnotation(key, value string) {
	b.annotations[key] = value
}

// SetGlobal sets whether the Service is global
func (b *Service) SetGlobal(isGlobal bool) {
	b.isGlobal = isGlobal
}

// Create creates a new Service
func (b *Service) Create() *corev1.Service {
	service := b.deployment.Manifest.Groups[0].Services[b.serviceIndex]
	ports := make([]corev1.ServicePort, 0)

	for _, expose := range service.Expose {
		port := corev1.ServicePort{
			Name:       fmt.Sprintf("port-%d", expose.Port),
			Protocol:   corev1.ProtocolTCP,
			Port:       int32(expose.As),
			TargetPort: intstr.FromInt(int(expose.Port)),
		}
		ports = append(ports, port)
	}

	// For global services, we need to use NodePort type
	serviceType := b.serviceType
	if b.isGlobal {
		serviceType = corev1.ServiceTypeNodePort
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.Name(),
			Namespace: b.NS(),
			Labels:    b.labels(),
		},
		Spec: corev1.ServiceSpec{
			Type: serviceType,
			Selector: map[string]string{
				"app":                  fmt.Sprintf("%s-group-%d-service-%d", b.deployment.Name, b.groupIndex, b.serviceIndex),
				"subnet-node/lease-id": b.deployment.ID,
			},
			Ports: ports,
		},
	}

	// Set external traffic policy for global services
	if b.isGlobal && serviceType == corev1.ServiceTypeNodePort {
		svc.Spec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyTypeLocal
	}

	return svc
}

// Update updates an existing Service
func (b *Service) Update(obj *corev1.Service) (*corev1.Service, error) {
	obj.Labels = b.labels()
	obj.Annotations = b.annotations
	obj.Spec.Ports = b.ports
	obj.Spec.Selector = map[string]string{
		"app":                  fmt.Sprintf("%s-group-%d-service-%d", b.deployment.Name, b.groupIndex, b.serviceIndex),
		"subnet-node/lease-id": b.deployment.ID,
	}

	// For global services, we need to use NodePort type
	serviceType := b.serviceType
	if b.isGlobal {
		serviceType = corev1.ServiceTypeNodePort
	}
	obj.Spec.Type = serviceType

	if b.isGlobal && serviceType == corev1.ServiceTypeNodePort {
		obj.Spec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyTypeLocal
	} else {
		obj.Spec.ExternalTrafficPolicy = ""
	}

	return obj, nil
}

// labels returns the labels for the Service
func (b *Service) labels() map[string]string {
	return map[string]string{
		"app": b.deployment.Name,
	}
}

// Validate validates the Service configuration
func (b *Service) Validate() error {
	if len(b.ports) == 0 {
		return fmt.Errorf("service must have at least one port")
	}

	if len(b.selector) == 0 {
		return fmt.Errorf("service must have at least one selector")
	}

	return nil
}

// Any returns whether the Service has any configuration
func (b *Service) Any() bool {
	return len(b.ports) > 0
}
