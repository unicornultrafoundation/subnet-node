package builder

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"go.uber.org/zap"

	manifest "github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/manifest"
	types "github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
)

type PVC struct {
	baseBuilder
	serviceName string
	storage     *manifest.StorageParams
	volumeName  string
	size        string
	class       string
}

func NewPVCBuilder(
	logger *zap.Logger,
	settings Settings,
	deployment *types.ManagedDeployment,
	serviceName string,
	storage *manifest.StorageParams,
) (*PVC, error) {
	return &PVC{
		baseBuilder: baseBuilder{
			settings:   settings,
			deployment: deployment,
		},
		serviceName: serviceName,
		storage:     storage,
		volumeName:  fmt.Sprintf("%s-shm", serviceName),
		size:        "64Mi", // Default SHM size
	}, nil
}

func NewPersistentVolumeClaimBuilder(
	logger *zap.Logger,
	settings Settings,
	deployment *types.ManagedDeployment,
	serviceName string,
	volumeName string,
	size string,
	class string,
) (*PVC, error) {
	return &PVC{
		baseBuilder: baseBuilder{
			settings:   settings,
			deployment: deployment,
		},
		serviceName: serviceName,
		volumeName:  volumeName,
		size:        size,
		class:       class,
	}, nil
}

func (b *PVC) Name() string {
	return b.volumeName
}

func (b *PVC) NS() string {
	return b.deployment.Namespace
}

func (b *PVC) Create() (*corev1.PersistentVolumeClaim, error) {
	volumeMode := corev1.PersistentVolumeFilesystem
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.Name(),
			Namespace: b.NS(),
			Labels:    b.labels(),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(b.size),
				},
			},
			VolumeMode: &volumeMode,
		},
	}

	if b.class != "" {
		pvc.Spec.StorageClassName = &b.class
	}

	return pvc, nil
}

func (b *PVC) Update(obj *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	volumeMode := corev1.PersistentVolumeFilesystem
	pvc := obj.DeepCopy()
	pvc.Spec = corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
		Resources: corev1.VolumeResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(b.size),
			},
		},
		VolumeMode: &volumeMode,
	}

	if b.class != "" {
		pvc.Spec.StorageClassName = &b.class
	}

	return pvc, nil
}

func (b *PVC) labels() map[string]string {
	return map[string]string{
		"app": b.serviceName,
	}
}
