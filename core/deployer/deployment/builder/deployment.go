package builder

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"go.uber.org/zap"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
)

// Deployment represents a Kubernetes deployment builder
type Deployment struct {
	*Workload
	settings    Settings
	deployment  *types.ManagedDeployment
	secretsRefs []corev1.LocalObjectReference
	volumesObjs []corev1.Volume
}

// NewDeploymentBuilder creates a new deployment builder
func NewDeploymentBuilder(
	logger *zap.Logger,
	settings Settings,
	deployment *types.ManagedDeployment,
	groupIdx int,
	serviceIdx int,
) (*Deployment, error) {
	workload, err := NewWorkloadBuilder(logger, settings, deployment, groupIdx, serviceIdx)
	if err != nil {
		return nil, err
	}

	return &Deployment{
		Workload:    workload,
		settings:    settings,
		deployment:  deployment,
		secretsRefs: workload.imagePullSecrets(),
		volumesObjs: workload.volumes(),
	}, nil
}

// Create creates a new deployment
func (b *Deployment) Create() (*appsv1.Deployment, error) {
	falseValue := false
	revisionHistoryLimit := int32(10)

	replicas := int32(1)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.Name(),
			Namespace: b.NS(),
			Labels:    b.labels(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: b.selectorLabels(),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: b.labels(),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						b.container(),
					},
					Volumes: b.volumes(),
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &falseValue,
					},
					ImagePullSecrets: b.secretsRefs,
					Affinity:         b.affinity(),
					RuntimeClassName: b.runtimeClass(),
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 0},
					MaxSurge:       &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
				},
			},
			RevisionHistoryLimit: &revisionHistoryLimit,
		},
	}, nil
}

// Update updates an existing deployment
func (b *Deployment) Update(obj *appsv1.Deployment) (*appsv1.Deployment, error) {
	uobj := obj.DeepCopy()

	uobj.Labels = b.labels()
	uobj.Spec.Selector.MatchLabels = b.selectorLabels()
	uobj.Spec.Template.Labels = b.labels()
	uobj.Spec.Template.Spec.Containers = []corev1.Container{b.container()}
	uobj.Spec.Template.Spec.ImagePullSecrets = b.secretsRefs
	uobj.Spec.Template.Spec.Volumes = b.volumesObjs

	return uobj, nil
}
