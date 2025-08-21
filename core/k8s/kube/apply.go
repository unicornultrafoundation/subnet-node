package kube

// nolint:deadcode,golint

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	"github.com/unicornultrafoundation/subnet-node/core/k8s/kube/builder"
)

type k8sPatch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

func applyNS(ctx context.Context, kc kubernetes.Interface, b builder.NS) (*corev1.Namespace, *corev1.Namespace, *corev1.Namespace, error) {
	oobj, err := kc.CoreV1().Namespaces().Get(ctx, b.Name(), metav1.GetOptions{})

	var nobj *corev1.Namespace
	var uobj *corev1.Namespace

	switch {
	case err == nil:
		uobj, err = b.Update(oobj)

		if err == nil && (!b.IsObjectRevisionLatest(oobj.Labels) ||
			!reflect.DeepEqual(&uobj.Spec, &oobj.Spec) ||
			!reflect.DeepEqual(uobj.Labels, oobj.Labels)) {
			uobj, err = kc.CoreV1().Namespaces().Update(ctx, uobj, metav1.UpdateOptions{})
		}
	case errors.IsNotFound(err):
		nobj, err = b.Create()
		if err == nil {
			nobj, err = kc.CoreV1().Namespaces().Create(ctx, nobj, metav1.CreateOptions{})
		}
	}

	return nobj, uobj, oobj, err
}

// Apply list of Network Policies
func applyNetPolicies(ctx context.Context, kc kubernetes.Interface, b builder.NetPol) ([]netv1.NetworkPolicy, []netv1.NetworkPolicy, []netv1.NetworkPolicy, error) {
	var err error

	policies, err := b.Create()
	if err != nil {
		return nil, nil, nil, err
	}

	currPolicies, err := kc.NetworkingV1().NetworkPolicies(b.NS()).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, nil, err
	}

	var nobjs []netv1.NetworkPolicy
	var uobjs []netv1.NetworkPolicy
	oobjs := currPolicies.Items

	for _, pol := range policies {
		oobj, err := kc.NetworkingV1().NetworkPolicies(b.NS()).Get(ctx, pol.Name, metav1.GetOptions{})

		var nobj *netv1.NetworkPolicy
		var uobj *netv1.NetworkPolicy

		switch {
		case err == nil:
			uobj, err = b.Update(oobj)

			if err == nil && (!b.IsObjectRevisionLatest(uobj.Labels) ||
				!reflect.DeepEqual(&uobj.Spec, &oobj.Spec) ||
				!reflect.DeepEqual(uobj.Labels, oobj.Labels)) {
				uobj, err = kc.NetworkingV1().NetworkPolicies(b.NS()).Update(ctx, pol, metav1.UpdateOptions{})
				if err == nil {
					uobjs = append(uobjs, *uobj)
				}
			}
		case errors.IsNotFound(err):
			nobj, err = kc.NetworkingV1().NetworkPolicies(b.NS()).Create(ctx, pol, metav1.CreateOptions{})
			if err == nil {
				nobjs = append(nobjs, *nobj)
			}
		}

		if err != nil {
			break
		}
	}

	return nobjs, uobjs, oobjs, err
}

func applyServiceCredentials(ctx context.Context, kc kubernetes.Interface, b builder.ServiceCredentials) (*corev1.Secret, *corev1.Secret, *corev1.Secret, error) {
	oobj, err := kc.CoreV1().Secrets(b.NS()).Get(ctx, b.Name(), metav1.GetOptions{})

	var nobj *corev1.Secret
	var uobj *corev1.Secret

	switch {
	case err == nil:
		uobj, err = b.Update(oobj)
		if err == nil && (!b.IsObjectRevisionLatest(uobj.Labels) ||
			!reflect.DeepEqual(&uobj.Data, &oobj.Data) ||
			!reflect.DeepEqual(uobj.Labels, oobj.Labels)) {
			uobj, err = kc.CoreV1().Secrets(b.NS()).Update(ctx, uobj, metav1.UpdateOptions{})
		}
	case errors.IsNotFound(err):
		nobj, err = b.Create()
		if err == nil {
			nobj, err = kc.CoreV1().Secrets(b.NS()).Create(ctx, nobj, metav1.CreateOptions{})
		}
	}

	return nobj, uobj, oobj, err

}

func applyDeployment(ctx context.Context, kc kubernetes.Interface, b builder.Deployment) (*appsv1.Deployment, *appsv1.Deployment, *appsv1.Deployment, error) {
	oobj, err := kc.AppsV1().Deployments(b.NS()).Get(ctx, b.Name(), metav1.GetOptions{})

	var nobj *appsv1.Deployment
	var uobj *appsv1.Deployment

	switch {
	case err == nil:
		uobj, err = b.Update(oobj)
		if err != nil {
			break
		}

		if b.IsObjectRevisionLatest(uobj.Labels) ||
			!reflect.DeepEqual(&uobj.Spec, &oobj.Spec) ||
			!reflect.DeepEqual(uobj.Labels, oobj.Labels) {
			uobj, err = kc.AppsV1().Deployments(b.NS()).Update(ctx, uobj, metav1.UpdateOptions{})
		}

		var patches []k8sPatch

		if rev := oobj.Spec.RevisionHistoryLimit; rev == nil || *rev != 10 {
			patches = append(patches, k8sPatch{
				Op:    "add",
				Path:  "/spec/revisionHistoryLimit",
				Value: int32(10),
			})
		}

		ustrategy := &oobj.Spec.Strategy
		if uobj != nil {
			ustrategy = &uobj.Spec.Strategy
		}

		maxSurge := intstr.FromInt32(0)
		maxUnavailable := intstr.FromInt32(1)

		strategy := appsv1.DeploymentStrategy{
			Type: appsv1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &appsv1.RollingUpdateDeployment{
				MaxUnavailable: &maxUnavailable,
				MaxSurge:       &maxSurge,
			},
		}

		if !reflect.DeepEqual(&strategy, &ustrategy) {
			patches = append(patches, k8sPatch{
				Op:    "replace",
				Path:  "/spec/strategy",
				Value: strategy,
			})
		}

		if len(patches) > 0 {
			data, _ := json.Marshal(patches)

			oobj, err = kc.AppsV1().Deployments(b.NS()).Patch(ctx, oobj.Name, k8stypes.JSONPatchType, data, metav1.PatchOptions{})
		}
	case errors.IsNotFound(err):
		nobj, err = b.Create()
		if err == nil {
			nobj, err = kc.AppsV1().Deployments(b.NS()).Create(ctx, nobj, metav1.CreateOptions{})
		}
	}

	return nobj, uobj, oobj, err
}

func applyPVC(ctx context.Context, kc kubernetes.Interface, pvc corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, *corev1.PersistentVolumeClaim, *corev1.PersistentVolumeClaim, error) {
	// Ensure PVC has a namespace set
	if pvc.Namespace == "" {
		return nil, nil, nil, fmt.Errorf("PVC %s has no namespace set", pvc.Name)
	}

	oobj, err := kc.CoreV1().PersistentVolumeClaims(pvc.Namespace).Get(ctx, pvc.Name, metav1.GetOptions{})

	var nobj *corev1.PersistentVolumeClaim
	var uobj *corev1.PersistentVolumeClaim

	switch {
	case err == nil:
		// PVC exists, update it
		uobj = oobj.DeepCopy()
		uobj.Spec = pvc.Spec
		uobj.Labels = pvc.Labels
		uobj.Annotations = pvc.Annotations

		if !reflect.DeepEqual(&uobj.Spec, &oobj.Spec) ||
			!reflect.DeepEqual(uobj.Labels, oobj.Labels) ||
			!reflect.DeepEqual(uobj.Annotations, oobj.Annotations) {
			uobj, err = kc.CoreV1().PersistentVolumeClaims(pvc.Namespace).Update(ctx, uobj, metav1.UpdateOptions{})
		}
	case errors.IsNotFound(err):
		// PVC doesn't exist, create it
		nobj = &pvc
		nobj.Namespace = pvc.Namespace
		nobj, err = kc.CoreV1().PersistentVolumeClaims(pvc.Namespace).Create(ctx, nobj, metav1.CreateOptions{})
	}

	return nobj, uobj, oobj, err
}

func applyStatefulSet(ctx context.Context, kc kubernetes.Interface, b builder.StatefulSet) (*appsv1.StatefulSet, *appsv1.StatefulSet, *appsv1.StatefulSet, error) {
	oobj, err := kc.AppsV1().StatefulSets(b.NS()).Get(ctx, b.Name(), metav1.GetOptions{})

	var nobj *appsv1.StatefulSet
	var uobj *appsv1.StatefulSet

	switch {
	case err == nil:
		uobj, err = b.Update(oobj)
		if err != nil {
			break
		}

		if b.IsObjectRevisionLatest(uobj.Labels) ||
			!reflect.DeepEqual(&uobj.Spec, &oobj.Spec) ||
			!reflect.DeepEqual(uobj.Labels, oobj.Labels) {
			uobj, err = kc.AppsV1().StatefulSets(b.NS()).Update(ctx, uobj, metav1.UpdateOptions{})
		}
	case errors.IsNotFound(err):
		nobj, err = b.Create()
		if err == nil {
			nobj, err = kc.AppsV1().StatefulSets(b.NS()).Create(ctx, nobj, metav1.CreateOptions{})
		}
	}

	return nobj, uobj, oobj, err
}

func applyService(ctx context.Context, kc kubernetes.Interface, b builder.Service) (*corev1.Service, *corev1.Service, *corev1.Service, error) {
	oobj, err := kc.CoreV1().Services(b.NS()).Get(ctx, b.Name(), metav1.GetOptions{})
	var nobj *corev1.Service
	var uobj *corev1.Service

	switch {
	case err == nil:
		uobj, err = b.Update(oobj)
		if err == nil && (b.IsObjectRevisionLatest(uobj.Labels) ||
			!reflect.DeepEqual(&uobj.Spec, &oobj.Spec) ||
			!reflect.DeepEqual(uobj.Labels, oobj.Labels)) {
			uobj, err = kc.CoreV1().Services(b.NS()).Update(ctx, uobj, metav1.UpdateOptions{})
		}
	case errors.IsNotFound(err):
		nobj, err = b.Create()
		if err == nil {
			nobj, err = kc.CoreV1().Services(b.NS()).Create(ctx, nobj, metav1.CreateOptions{})
		}
	}

	return nobj, uobj, oobj, err
}
