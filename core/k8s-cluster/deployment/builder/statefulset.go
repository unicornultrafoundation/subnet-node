package builder

import (
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatefulSet represents a Kubernetes statefulset builder
type StatefulSet struct {
	baseBuilder
	name           string
	replicas       int32
	image          string
	containerPorts []int32
	volumeClaims   []corev1.PersistentVolumeClaim
}

// NewStatefulSet creates a new statefulset builder
func NewStatefulSet(b baseBuilder, name string, replicas int32, image string, ports []int32, claims []corev1.PersistentVolumeClaim) *StatefulSet {
	return &StatefulSet{
		baseBuilder:    b,
		name:           name,
		replicas:       replicas,
		image:          image,
		containerPorts: ports,
		volumeClaims:   claims,
	}
}

// Create creates a new Kubernetes statefulset
func (b *StatefulSet) Create() (*appsv1.StatefulSet, error) {
	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.Name(),
			Namespace: b.NS(),
			Labels:    b.labels(),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &b.replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": b.name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": b.name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  b.name,
							Image: b.image,
							Ports: b.buildContainerPorts(),
						},
					},
				},
			},
			VolumeClaimTemplates: b.volumeClaims,
			ServiceName:          b.name,
		},
	}

	return statefulset, nil
}

// Update updates an existing Kubernetes statefulset
func (b *StatefulSet) Update(obj *appsv1.StatefulSet) (*appsv1.StatefulSet, error) {
	obj.Labels = b.labels()
	obj.Spec = appsv1.StatefulSetSpec{
		Replicas: &b.replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": b.name,
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": b.name,
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  b.name,
						Image: b.image,
						Ports: b.buildContainerPorts(),
					},
				},
			},
		},
		VolumeClaimTemplates: b.volumeClaims,
		ServiceName:          b.name,
	}
	return obj, nil
}

// buildContainerPorts builds the container ports
func (b *StatefulSet) buildContainerPorts() []corev1.ContainerPort {
	ports := make([]corev1.ContainerPort, 0, len(b.containerPorts))
	for _, port := range b.containerPorts {
		ports = append(ports, corev1.ContainerPort{
			Name:          "port-" + strconv.Itoa(int(port)),
			ContainerPort: port,
			Protocol:      corev1.ProtocolTCP,
		})
	}
	return ports
}

// Name returns the statefulset name
func (b *StatefulSet) Name() string {
	return b.name
}

// NS returns the namespace
func (b *StatefulSet) NS() string {
	return b.baseBuilder.NS()
}

// labels returns the statefulset labels
func (b *StatefulSet) labels() map[string]string {
	return map[string]string{
		"app": b.name,
	}
}
