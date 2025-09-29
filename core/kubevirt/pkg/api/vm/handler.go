package vm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/rancher/apiserver/pkg/apierror"
	ctlcorev1 "github.com/rancher/wrangler/v3/pkg/generated/controllers/core/v1"
	volumeapi "github.com/unicornultrafoundation/subnet-node/core/kubevirt/pkg/api/volume"
	ctlkubevirtv1 "github.com/unicornultrafoundation/subnet-node/core/kubevirt/pkg/generated/controllers/kubevirt.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// Return VM status instead of 204 No Content
	if err := h.returnVMStatus(rw, req); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write([]byte(fmt.Sprintf("Failed to get VM status: %v", err)))
		return
	}
}

func (h vmActionHandler) doAction(rw http.ResponseWriter, r *http.Request) error {
	action := r.URL.Query().Get("action")

	// Extract namespace and name from URL path
	// Expected path format: /kubevirt/kubevirt.io.virtualmachine/{namespace}/{name}?action={action}
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	name := vars["name"]

	if namespace == "" || name == "" {
		return fmt.Errorf("namespace and name must be provided in the URL path")
	}

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

	// Use the direct client instead of cache to avoid type conversion issues
	vm, err := h.vms.Get(namespace, name, metav1.GetOptions{})
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
		// Use direct client to start VM by updating runStrategy
		return h.startVM(namespace, name)
	case restartVM:
		// For restart, we need to stop first, then start
		if err := h.stopVM(namespace, name); err != nil {
			return err
		}
		return h.startVM(namespace, name)
	case pauseVM, unpauseVM, softReboot:
		// These operations require VMI subresources
		if h.virtSubresourceRestClient == nil {
			return fmt.Errorf("VMI operations (pause, unpause, softReboot) are not available: KubeVirt subresources not installed")
		}
		err := h.virtSubresourceRestClient.Put().Namespace(namespace).Resource(resource).SubResource(subresourece).Name(name).Do(ctx).Error()
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (h *vmActionHandler) startVM(namespace, name string) error {
	// Use the direct client to start VM by updating runStrategy
	vm, err := h.vms.Get(namespace, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get virtual machine %s/%s: %v", namespace, name, err)
	}

	vmCopy := vm.DeepCopy()
	runStrategy := kubevirtv1.RunStrategyAlways
	vmCopy.Spec.RunStrategy = &runStrategy
	if !reflect.DeepEqual(vm, vmCopy) {
		_, err = h.vms.Update(vmCopy)
		return err
	}
	return nil
}

func (h *vmActionHandler) stopVM(namespace, name string) error {

	// Use the direct client instead of cache to avoid type conversion issues
	vm, err := h.vms.Get(namespace, name, metav1.GetOptions{})
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

// VMStatusResponse represents the response structure for VM status
type VMStatusResponse struct {
	Action    string `json:"action"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Ready     bool   `json:"ready"`
	Message   string `json:"message,omitempty"`
	VMIStatus string `json:"vmi_status,omitempty"`
	VMIReady  bool   `json:"vmi_ready,omitempty"`
}

func (h *vmActionHandler) returnVMStatus(rw http.ResponseWriter, req *http.Request) error {
	// Extract namespace and name from URL path
	vars := mux.Vars(req)
	namespace := vars["namespace"]
	name := vars["name"]
	action := req.URL.Query().Get("action")

	// Get VM status
	vm, err := h.vms.Get(namespace, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get VM: %v", err)
	}

	// Get VMI status if it exists
	var vmiStatus string
	var vmiReady bool
	vmi, err := h.vmis.Get(namespace, name, metav1.GetOptions{})
	if err == nil {
		vmiStatus = string(vmi.Status.Phase)
		// Check if VMI is ready by looking at conditions
		vmiReady = false
		for _, condition := range vmi.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				vmiReady = true
				break
			}
		}
	}

	// Create response
	response := VMStatusResponse{
		Action:    action,
		Namespace: namespace,
		Name:      name,
		Status:    string(vm.Status.PrintableStatus),
		Ready:     vm.Status.Ready,
		VMIStatus: vmiStatus,
		VMIReady:  vmiReady,
	}

	// Add status message based on VM conditions
	if len(vm.Status.Conditions) > 0 {
		for _, condition := range vm.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "False" {
				response.Message = condition.Message
				break
			}
		}
	}

	// Set response headers
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	// Encode and send response
	encoder := json.NewEncoder(rw)
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}
