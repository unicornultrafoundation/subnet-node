package builder

import (
	"fmt"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
)

// ServiceCredentialsBuilder represents a Kubernetes service credentials builder
type ServiceCredentialsBuilder struct {
	baseBuilder
	groupIndex   int
	serviceIndex int
	log          *zap.Logger
}

// NewServiceCredentialsBuilder creates a new service credentials builder
func NewServiceCredentialsBuilder(logger *zap.Logger, settings Settings, deployment *types.ManagedDeployment, groupIndex, serviceIndex int) (*ServiceCredentialsBuilder, error) {
	if groupIndex < 0 || groupIndex >= len(deployment.Manifest.Groups) {
		return nil, fmt.Errorf("invalid group index: %d", groupIndex)
	}
	if serviceIndex < 0 || serviceIndex >= len(deployment.Manifest.Groups[groupIndex].Services) {
		return nil, fmt.Errorf("invalid service index: %d", serviceIndex)
	}
	return &ServiceCredentialsBuilder{
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

// Name returns the name of the service credentials
func (b *ServiceCredentialsBuilder) Name() string {
	return fmt.Sprintf("%s-group-%d-service-%d", b.deployment.Name, b.groupIndex, b.serviceIndex)
}

// NS returns the namespace of the service credentials
func (b *ServiceCredentialsBuilder) NS() string {
	return b.ns
}

// labels returns the labels for the service credentials
func (b *ServiceCredentialsBuilder) labels() map[string]string {
	return map[string]string{
		"app":                  b.Name(),
		"subnet-node/lease-id": b.deployment.ID,
		"subnet-node/owner":    b.deployment.Requester.Hex(),
	}
}

// Create creates a new service credentials secret
func (b *ServiceCredentialsBuilder) Create() (*corev1.Secret, error) {
	service := b.deployment.Manifest.Groups[b.groupIndex].Services[b.serviceIndex]

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-credentials", b.Name()),
			Namespace: b.ns,
			Labels:    b.labels(),
		},
		Type: corev1.SecretTypeOpaque,
		Data: make(map[string][]byte),
	}

	// Add service credentials
	if service.Credentials != nil {
		if service.Credentials.Username != "" {
			secret.Data["username"] = []byte(service.Credentials.Username)
		}
		if service.Credentials.Password != "" {
			secret.Data["password"] = []byte(service.Credentials.Password)
		}
	}

	return secret, nil
}

// Update updates an existing service credentials secret
func (b *ServiceCredentialsBuilder) Update(existing *corev1.Secret) (*corev1.Secret, error) {
	service := b.deployment.Manifest.Groups[b.groupIndex].Services[b.serviceIndex]

	// Update labels
	existing.Labels = b.labels()

	// Update credentials
	if service.Credentials != nil {
		if service.Credentials.Username != "" {
			existing.Data["username"] = []byte(service.Credentials.Username)
		}
		if service.Credentials.Password != "" {
			existing.Data["password"] = []byte(service.Credentials.Password)
		}
	}

	return existing, nil
}
