package builder

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NS represents a Kubernetes namespace builder
type NS struct {
	baseBuilder
}

// NewNS creates a new namespace builder
func NewNS(b baseBuilder) *NS {
	return &NS{
		baseBuilder: b,
	}
}

// Create creates a new Kubernetes namespace
func (b *NS) Create() (*corev1.Namespace, error) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   b.Name(),
			Labels: b.labels(),
		},
	}

	return ns, nil
}

// Update updates an existing Kubernetes namespace
func (b *NS) Update(obj *corev1.Namespace) (*corev1.Namespace, error) {
	obj.Labels = b.labels()
	return obj, nil
}

// Name returns the namespace name
func (b *NS) Name() string {
	return b.baseBuilder.Name()
}

// NS returns the namespace
func (b *NS) NS() string {
	return b.baseBuilder.NS()
}

// labels returns the namespace labels
func (b *NS) labels() map[string]string {
	return map[string]string{
		"app": b.baseBuilder.Name(),
	}
}
