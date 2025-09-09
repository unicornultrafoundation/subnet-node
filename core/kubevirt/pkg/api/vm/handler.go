package vm

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/rancher/apiserver/pkg/apierror"
	ctlcorev1 "github.com/rancher/wrangler/v3/pkg/generated/controllers/core/v1"
	volumeapi "github.com/unicornultrafoundation/subnet-node/core/kubevirt/pkg/api/volume"
	ctlkubevirtv1 "github.com/unicornultrafoundation/subnet-node/core/kubevirt/pkg/generated/controllers/kubevirt.io/v1"
	"k8s.io/client-go/rest"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

const (
	vmResource  = "virtualmachines"
	vmiResource = "virtualmachineinstances"
)

type vmActionHandler struct {
	namespace     string
	kubevirtCache ctlkubevirtv1.KubeVirtCache
	vms           ctlkubevirtv1.VirtualMachineClient
	vmis          ctlkubevirtv1.VirtualMachineInstanceClient
	vmCache       ctlkubevirtv1.VirtualMachineCache

	vmims  ctlkubevirtv1.VirtualMachineInstanceMigrationClient
	vmimsc ctlkubevirtv1.VirtualMachineInstanceMigrationCache

	// k8s core
	pvcCache ctlcorev1.PersistentVolumeClaimCache
	pvCache  ctlcorev1.PersistentVolumeCache

	virtSubresourceRestClient rest.Interface
}

func (h vmActionHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if err := h.doAction(rw, req); err != nil {
		status := http.StatusInternalServerError
		if e, ok := err.(*apierror.APIError); ok {
			status = e.Code.Status
		}
		rw.WriteHeader(status)
		_, _ = rw.Write([]byte(err.Error()))
		return
	}
	rw.WriteHeader(http.StatusNoContent)
}

func (h vmActionHandler) doAction(rw http.ResponseWriter, r *http.Request) error {
	action := r.URL.Query().Get("action")
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	switch action {
	// TODO: add other actions later
	case startVM, restartVM:
		if err := h.subresourceOperate(r.Context(), vmResource, namespace, name, action); err != nil {
			return fmt.Errorf("%s virtual machine %s/%s failed, %v", action, namespace, name, err)
		}
	case stopVM:
		// To align behavior with kubevirt v1.1.1, we set runStrategy to Halted when stopping a VM.
		if err := h.stopVM(namespace, name); err != nil {
			return fmt.Errorf("%s virtual machine %s/%s failed, %v", action, namespace, name, err)
		}
	case pauseVM, unpauseVM, softReboot:
		if err := h.subresourceOperate(r.Context(), vmiResource, namespace, name, action); err != nil {
			return fmt.Errorf("%s virtual machine %s/%s failed, %v", action, namespace, name, err)
		}

	}

	return nil
}

func (h *vmActionHandler) startPreCheck(namespace, name string) error {
	vm, err := h.vmCache.Get(namespace, name)
	if err != nil {
		return err
	}

	for _, volume := range vm.Spec.Template.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			pvcName := volume.PersistentVolumeClaim.PersistentVolumeClaimVolumeSource.ClaimName
			pvcNamespace := vm.Namespace
			pvc, err := h.pvcCache.Get(pvcNamespace, pvcName)
			if err != nil {
				return err
			}
			if volumeapi.IsResizing(pvc) {
				return fmt.Errorf("can not start the VM %s/%s which has a resizing volume %s/%s", vm.Namespace, vm.Name, pvcNamespace, pvcName)
			}
		}
	}

	return nil
}

func (h *vmActionHandler) subresourceOperate(ctx context.Context, resource, namespace, name, subresourece string) error {
	switch subresourece {
	case startVM:
		if err := h.startPreCheck(namespace, name); err != nil {
			return err
		}
	}

	return h.virtSubresourceRestClient.Put().Namespace(namespace).Resource(resource).SubResource(subresourece).Name(name).Do(ctx).Error()
}

func (h *vmActionHandler) stopVM(namespace, name string) error {
	vm, err := h.vmCache.Get(namespace, name)
	if err != nil {
		return fmt.Errorf("failed to get virtual machine %s/%s: %v", namespace, name, err)
	}

	vmCopy := vm.DeepCopy()
	runStrategy := kubevirtv1.RunStrategyHalted
	vmCopy.Spec.RunStrategy = &runStrategy
	if !reflect.DeepEqual(vm, vmCopy) {
		_, err = h.vms.Update(vmCopy)
		return err
	}
	return nil
}
