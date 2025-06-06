package builder

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodSecurity represents a Pod Security configuration
type PodSecurity struct {
	baseBuilder
}

// NewPodSecurity creates a new Pod Security configuration
func NewPodSecurity(b baseBuilder) *PodSecurity {
	return &PodSecurity{
		baseBuilder: b,
	}
}

// Name returns the name of the Pod Security configuration
func (b *PodSecurity) Name() string {
	return fmt.Sprintf("%s-ps", b.deployment.Name)
}

// NS returns the namespace of the Pod Security configuration
func (b *PodSecurity) NS() string {
	return b.deployment.Namespace
}

// Create creates a new Pod Security configuration
func (b *PodSecurity) Create() (*corev1.Namespace, error) {
	if !b.settings.EnablePodSecurityAdmission {
		return nil, nil
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.NS(),
			Labels: map[string]string{
				"pod-security.kubernetes.io/enforce": "restricted",
				"pod-security.kubernetes.io/audit":   "restricted",
				"pod-security.kubernetes.io/warn":    "restricted",
			},
		},
	}

	return ns, nil
}

// Update updates an existing Pod Security configuration
func (b *PodSecurity) Update(obj *corev1.Namespace) (*corev1.Namespace, error) {
	if !b.settings.EnablePodSecurityAdmission {
		return nil, nil
	}

	obj.Labels = map[string]string{
		"pod-security.kubernetes.io/enforce": "restricted",
		"pod-security.kubernetes.io/audit":   "restricted",
		"pod-security.kubernetes.io/warn":    "restricted",
	}

	return obj, nil
}

// labels returns the labels for the Pod Security configuration
func (b *PodSecurity) labels() map[string]string {
	return map[string]string{
		"app": b.deployment.Name,
	}
}
