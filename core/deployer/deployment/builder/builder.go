package builder

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/pkg/errors"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// Settings configures k8s object generation such that it is customized to the
// cluster environment that is being used.
// For instance, GCP requires a different service type than minikube.
type Settings struct {
	// Service type for deployments
	// gcp:    NodePort
	// others: ClusterIP
	DeploymentServiceType corev1.ServiceType

	// Whether to use static hosts for ingress
	// gcp:    false
	// others: true
	DeploymentIngressStaticHosts bool

	// Ingress domain to map deployments to
	DeploymentIngressDomain string

	// Whether to expose load balancer host in status
	// gcp:    true
	// others: optional
	DeploymentIngressExposeLBHosts bool

	// Global hostname for arbitrary ports
	ClusterPublicHostname string

	// Whether NetworkPolicies should be installed
	NetworkPoliciesEnabled bool

	// Resource commit levels
	CPUCommitLevel     float64
	GPUCommitLevel     float64
	MemoryCommitLevel  float64
	StorageCommitLevel float64

	// Runtime class for deployments
	DeploymentRuntimeClass string

	// Name of the image pull secret to use in pod spec
	DockerImagePullSecretsName string

	// Kubernetes client
	Client kubernetes.Interface

	// Dynamic client for CRDs
	DynamicClient dynamic.Interface

	// Logger
	Logger *zap.Logger

	// Additional settings
	EnablePodSecurityAdmission     bool
	EnableResourceQuotas           bool
	EnableLimitRanges              bool
	EnableHorizontalPodAutoscaling bool
	EnableVerticalPodAutoscaling   bool
}

var ErrSettingsValidation = errors.New("settings validation")

// ValidateSettings validates the settings configuration
func ValidateSettings(settings Settings) error {
	if settings.Client == nil {
		return fmt.Errorf("%w: kubernetes client is required", ErrSettingsValidation)
	}

	if settings.Logger == nil {
		return fmt.Errorf("%w: logger is required", ErrSettingsValidation)
	}

	if settings.DeploymentIngressStaticHosts {
		if settings.DeploymentIngressDomain == "" {
			return fmt.Errorf("%w: empty ingress domain", ErrSettingsValidation)
		}

		if !isDomainName(settings.DeploymentIngressDomain) {
			return fmt.Errorf("%w: invalid domain name %q", ErrSettingsValidation, settings.DeploymentIngressDomain)
		}
	}

	// Validate resource commit levels
	if settings.CPUCommitLevel < 0 || settings.CPUCommitLevel > 1 {
		return fmt.Errorf("%w: invalid CPU commit level %f", ErrSettingsValidation, settings.CPUCommitLevel)
	}
	if settings.GPUCommitLevel < 0 || settings.GPUCommitLevel > 1 {
		return fmt.Errorf("%w: invalid GPU commit level %f", ErrSettingsValidation, settings.GPUCommitLevel)
	}
	if settings.MemoryCommitLevel < 0 || settings.MemoryCommitLevel > 1 {
		return fmt.Errorf("%w: invalid memory commit level %f", ErrSettingsValidation, settings.MemoryCommitLevel)
	}
	if settings.StorageCommitLevel < 0 || settings.StorageCommitLevel > 1 {
		return fmt.Errorf("%w: invalid storage commit level %f", ErrSettingsValidation, settings.StorageCommitLevel)
	}

	return nil
}

// NewDefaultSettings returns default settings configuration
func NewDefaultSettings() Settings {
	return Settings{
		DeploymentServiceType:          corev1.ServiceTypeClusterIP,
		DeploymentIngressStaticHosts:   false,
		DeploymentIngressExposeLBHosts: false,
		NetworkPoliciesEnabled:         false,
		EnablePodSecurityAdmission:     false,
		EnableResourceQuotas:           false,
		EnableLimitRanges:              false,
		EnableHorizontalPodAutoscaling: false,
		EnableVerticalPodAutoscaling:   false,
		CPUCommitLevel:                 1.0,
		GPUCommitLevel:                 1.0,
		MemoryCommitLevel:              1.0,
		StorageCommitLevel:             1.0,
	}
}

// BuildNamespace creates a namespace builder
func BuildNamespace(settings Settings, deployment *types.ManagedDeployment, client kubernetes.Interface) *namespaceBuilder {
	return &namespaceBuilder{
		baseBuilder: baseBuilder{
			settings:   settings,
			deployment: deployment,
		},
		client: client,
	}
}

// BuildNetPol creates a network policy builder
func BuildNetPol(settings Settings, deployment *types.ManagedDeployment) *networkPolicyBuilder {
	return &networkPolicyBuilder{
		baseBuilder: baseBuilder{
			settings:   settings,
			deployment: deployment,
		},
	}
}

// BuildService creates a service builder
func BuildService(settings Settings, deployment *types.ManagedDeployment, groupIndex, serviceIndex int, isGlobal bool) *serviceBuilder {
	return &serviceBuilder{
		baseBuilder: baseBuilder{
			settings:   settings,
			deployment: deployment,
		},
		groupIndex:   groupIndex,
		serviceIndex: serviceIndex,
		isGlobal:     isGlobal,
	}
}

// BuildStatefulSet creates a statefulset builder
func BuildStatefulSet(settings Settings, deployment *types.ManagedDeployment) *statefulSetBuilder {
	return &statefulSetBuilder{
		baseBuilder: baseBuilder{
			settings:   settings,
			deployment: deployment,
		},
	}
}

// BuildDeployment creates a deployment builder
func BuildDeployment(settings Settings, deployment *types.ManagedDeployment) *deploymentBuilder {
	return &deploymentBuilder{
		baseBuilder: baseBuilder{
			settings:   settings,
			deployment: deployment,
		},
	}
}

// BuildServiceCredentials creates a service credentials builder
func BuildServiceCredentials(settings Settings, deployment *types.ManagedDeployment) *serviceCredentialsBuilder {
	return &serviceCredentialsBuilder{
		baseBuilder: baseBuilder{
			settings:   settings,
			deployment: deployment,
		},
	}
}

// baseBuilder provides common functionality for all builders
type baseBuilder struct {
	settings   Settings
	deployment *types.ManagedDeployment
	client     kubernetes.Interface
	ns         string
}

func (b *baseBuilder) Name() string {
	return b.deployment.Name
}

func (b *baseBuilder) NS() string {
	return b.deployment.Namespace
}

// namespaceBuilder implements NS interface
type namespaceBuilder struct {
	baseBuilder
	client kubernetes.Interface
}

func (b *namespaceBuilder) Create() (*corev1.Namespace, error) {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.NS(),
			Labels: map[string]string{
				"app": b.Name(),
			},
		},
	}, nil
}

func (b *namespaceBuilder) Update(ns *corev1.Namespace) (*corev1.Namespace, error) {
	if ns == nil {
		return nil, fmt.Errorf("namespace is nil")
	}
	ns.Labels = map[string]string{
		"app": b.Name(),
	}
	return ns, nil
}

// networkPolicyBuilder implements NetPol interface
type networkPolicyBuilder struct {
	baseBuilder
}

func (b *networkPolicyBuilder) Create() (*networkingv1.NetworkPolicy, error) {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.Name(),
			Namespace: b.NS(),
			Labels: map[string]string{
				"app": b.Name(),
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": b.Name(),
				},
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
		},
	}, nil
}

func (b *networkPolicyBuilder) Update(np *networkingv1.NetworkPolicy) (*networkingv1.NetworkPolicy, error) {
	if np == nil {
		return nil, fmt.Errorf("network policy is nil")
	}
	np.Labels = map[string]string{
		"app": b.Name(),
	}
	np.Spec.PodSelector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": b.Name(),
		},
	}
	return np, nil
}

// serviceBuilder implements Service interface
type serviceBuilder struct {
	baseBuilder
	groupIndex   int
	serviceIndex int
	isGlobal     bool
}

func (b *serviceBuilder) Create() (*corev1.Service, error) {
	// Get the service from the manifest
	service := b.deployment.Manifest.Groups[b.groupIndex].Services[b.serviceIndex]

	// Create ports slice
	ports := make([]corev1.ServicePort, 0)

	// Add ports from expose configuration
	for _, expose := range service.Expose {
		port := corev1.ServicePort{
			Name:       fmt.Sprintf("port-%d", expose.Port),
			Port:       int32(expose.Port),
			TargetPort: intstr.FromInt(int(expose.Port)),
			Protocol:   corev1.Protocol(expose.Proto),
		}
		ports = append(ports, port)
	}

	// If no ports were added, return error
	if len(ports) == 0 {
		return nil, fmt.Errorf("no ports configured for service")
	}

	name := fmt.Sprintf("%s-group-%d-service-%d", b.deployment.Name, b.groupIndex, b.serviceIndex)
	if b.isGlobal {
		name += "-global"
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: b.NS(),
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Type:  b.settings.DeploymentServiceType,
			Ports: ports,
		},
	}, nil
}

func (b *serviceBuilder) Update(svc *corev1.Service) (*corev1.Service, error) {
	if svc == nil {
		return nil, fmt.Errorf("service is nil")
	}

	service := b.deployment.Manifest.Groups[b.groupIndex].Services[b.serviceIndex]
	ports := make([]corev1.ServicePort, 0)
	for _, expose := range service.Expose {
		port := corev1.ServicePort{
			Name:       fmt.Sprintf("port-%d", expose.Port),
			Port:       int32(expose.Port),
			TargetPort: intstr.FromInt(int(expose.Port)),
			Protocol:   corev1.Protocol(expose.Proto),
		}
		ports = append(ports, port)
	}
	if len(ports) == 0 {
		return nil, fmt.Errorf("no ports configured for service")
	}
	name := fmt.Sprintf("%s-group-%d-service-%d", b.deployment.Name, b.groupIndex, b.serviceIndex)
	if b.isGlobal {
		name += "-global"
	}
	svc.Labels = map[string]string{
		"app": name,
	}
	svc.Spec.Selector = map[string]string{
		"app": name,
	}
	svc.Spec.Type = b.settings.DeploymentServiceType
	svc.Spec.Ports = ports
	return svc, nil
}

func (b *serviceBuilder) Any() bool {
	// Check if there are any exposed ports
	service := b.deployment.Manifest.Groups[b.groupIndex].Services[b.serviceIndex]
	return len(service.Expose) > 0
}

// statefulSetBuilder implements StatefulSet interface
type statefulSetBuilder struct {
	baseBuilder
}

func (b *statefulSetBuilder) Create() (*appsv1.StatefulSet, error) {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.Name(),
			Namespace: b.NS(),
			Labels: map[string]string{
				"app": b.Name(),
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": b.Name(),
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": b.Name(),
					},
				},
			},
		},
	}, nil
}

func (b *statefulSetBuilder) Update(sts *appsv1.StatefulSet) (*appsv1.StatefulSet, error) {
	if sts == nil {
		return nil, fmt.Errorf("statefulset is nil")
	}
	sts.Labels = map[string]string{
		"app": b.Name(),
	}
	sts.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": b.Name(),
		},
	}
	sts.Spec.Template.Labels = map[string]string{
		"app": b.Name(),
	}
	return sts, nil
}

// deploymentBuilder implements Deployment interface
type deploymentBuilder struct {
	baseBuilder
}

func (b *deploymentBuilder) Create() (*appsv1.Deployment, error) {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.Name(),
			Namespace: b.NS(),
			Labels: map[string]string{
				"app": b.Name(),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": b.Name(),
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": b.Name(),
					},
				},
			},
		},
	}, nil
}

func (b *deploymentBuilder) Update(dep *appsv1.Deployment) (*appsv1.Deployment, error) {
	if dep == nil {
		return nil, fmt.Errorf("deployment is nil")
	}
	dep.Labels = map[string]string{
		"app": b.Name(),
	}
	dep.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": b.Name(),
		},
	}
	dep.Spec.Template.Labels = map[string]string{
		"app": b.Name(),
	}
	return dep, nil
}

// serviceCredentialsBuilder implements ServiceCredentials interface
type serviceCredentialsBuilder struct {
	baseBuilder
}

func (b *serviceCredentialsBuilder) Create() (*corev1.Secret, error) {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.Name() + "-sa",
			Namespace: b.NS(),
			Labels: map[string]string{
				"app": b.Name(),
			},
		},
		Type: corev1.SecretTypeServiceAccountToken,
	}, nil
}

func (b *serviceCredentialsBuilder) Update(secret *corev1.Secret) (*corev1.Secret, error) {
	if secret == nil {
		return nil, fmt.Errorf("secret is nil")
	}
	secret.Labels = map[string]string{
		"app": b.Name(),
	}
	return secret, nil
}

// isDomainName checks if the given string is a valid domain name
func isDomainName(s string) bool {
	// See RFC 1035, RFC 3696.
	if len(s) == 0 {
		return false
	}
	if len(s) > 255 {
		return false
	}

	last := byte('.')
	ok := false // Ok once we've seen a letter.
	partlen := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		default:
			return false
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_':
			ok = true
			partlen++
		case '0' <= c && c <= '9':
			// fine
			partlen++
		case c == '-':
			// Byte before dash cannot be dot.
			if last == '.' {
				return false
			}
			partlen++
		case c == '.':
			// Byte before dot cannot be dot, dash.
			if last == '.' || last == '-' {
				return false
			}
			if partlen > 63 || partlen == 0 {
				return false
			}
			partlen = 0
		}
		last = c
	}
	if last == '-' || partlen > 63 {
		return false
	}

	return ok
}
