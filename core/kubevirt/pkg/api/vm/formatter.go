package vm

import (
	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/norman/types/convert"
	ctlcorev1 "github.com/rancher/wrangler/v3/pkg/generated/controllers/core/v1"
	ctlstoragev1 "github.com/rancher/wrangler/v3/pkg/generated/controllers/storage/v1"
	ctlkubevirtv1 "github.com/unicornultrafoundation/subnet-node/core/kubevirt/pkg/generated/controllers/kubevirt.io/v1"
	"k8s.io/client-go/kubernetes"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

const (
	startVM    = "start"
	stopVM     = "stop"
	restartVM  = "restart"
	pauseVM    = "pause"
	unpauseVM  = "unpause"
	softReboot = "softReboot"
)

type vmformatter struct {
	vmiCache  ctlkubevirtv1.VirtualMachineInstanceCache
	pvcCache  ctlcorev1.PersistentVolumeClaimCache
	nodeCache ctlcorev1.NodeCache
	scCache   ctlstoragev1.StorageClassCache
	clientSet kubernetes.Clientset
}

func (vf *vmformatter) formatter(request *types.APIRequest, resource *types.RawResource) {
	// reset resource actions, because action map already be set when add actions handler,
	// but current framework can't support use formatter to remove key from action map
	resource.Actions = make(map[string]string, 1)
	if request.AccessControl.CanUpdate(request, resource.APIObject, resource.Schema) != nil {
		return
	}

	vm := &kubevirtv1.VirtualMachine{}
	err := convert.ToObj(resource.APIObject.Data(), vm)
	if err != nil {
		return
	}

}
