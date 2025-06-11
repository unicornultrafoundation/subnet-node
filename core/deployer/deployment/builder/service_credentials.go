package builder

import (
	"encoding/base64"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceCredentials represents credentials for accessing container registries
type ServiceCredentials struct {
	baseBuilder
	host     string
	username string
	password string
}

// NewServiceCredentials creates a new service credentials builder
func NewServiceCredentials(b baseBuilder, host, username, password string) *ServiceCredentials {
	return &ServiceCredentials{
		baseBuilder: b,
		host:        host,
		username:    username,
		password:    password,
	}
}

// labels returns the service credentials labels
func (b *ServiceCredentials) labels() map[string]string {
	return map[string]string{
		"app": b.baseBuilder.Name(),
	}
}

// Create creates a new Kubernetes secret for registry credentials
func (b *ServiceCredentials) Create() (*corev1.Secret, error) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.baseBuilder.Name(),
			Namespace: b.baseBuilder.NS(),
			Labels:    b.labels(),
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			".dockerconfigjson": []byte(b.createDockerConfig()),
		},
	}

	return secret, nil
}

// Update updates an existing Kubernetes secret for registry credentials
func (b *ServiceCredentials) Update(obj *corev1.Secret) (*corev1.Secret, error) {
	obj.Labels = b.labels()
	obj.Data = map[string][]byte{
		".dockerconfigjson": []byte(b.createDockerConfig()),
	}
	return obj, nil
}

// createDockerConfig creates a Docker config JSON string
func (b *ServiceCredentials) createDockerConfig() string {
	auth := base64Encode(b.username + ":" + b.password)
	return `{
		"auths": {
			"` + b.host + `": {
				"auth": "` + auth + `"
			}
		}
	}`
}

// base64Encode encodes a string to base64
func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
