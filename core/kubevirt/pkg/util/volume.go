package util

import (
	"encoding/json"

	ctlstoragev1 "github.com/rancher/wrangler/v3/pkg/generated/controllers/storage/v1"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/kubevirt/pkg/settings"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	AnnStorageProvisioner     = "volume.kubernetes.io/storage-provisioner"
	AnnBetaStorageProvisioner = "volume.beta.kubernetes.io/storage-provisioner"
	LonghornDataLocality      = "dataLocality"
	IndexPodByPVC             = "indexPodByPVC"
)

var (
	PersistentVolumeClaimsKind = schema.GroupVersionKind{Group: "", Version: "v1", Kind: "PersistentVolumeClaim"}
)

func GetCSIProvisionerSnapshotCapability(provisioner string) bool {
	csiDriverConfig := make(map[string]settings.CSIDriverInfo)
	if err := json.Unmarshal([]byte(settings.CSIDriverConfig.Get()), &csiDriverConfig); err != nil {
		logrus.Warnf("Failed to unmarshal CSIDriverConfig setting, err: %v", err)
		return false
	}
	CSIConfigs, find := csiDriverConfig[provisioner]
	logrus.Debugf("Provisioner: %v, CSIConfigs: %+v", provisioner, CSIConfigs)
	if find && CSIConfigs.VolumeSnapshotClassName != "" {
		return true
	}

	return false
}

// GetProvisionedPVCProvisioner do not use this function when the PVC is just created
func GetProvisionedPVCProvisioner(pvc *corev1.PersistentVolumeClaim, scCache ctlstoragev1.StorageClassCache) string {
	provisioner, ok := pvc.Annotations[AnnBetaStorageProvisioner]
	if !ok {
		provisioner = pvc.Annotations[AnnStorageProvisioner]
	}
	if provisioner == "" {
		// fallback, fetch provisioner from storage class
		if pvc.Spec.StorageClassName == nil {
			logrus.Warnf("PVC %s/%s does not have storage class name, maybe just created", pvc.Namespace, pvc.Name)
			return provisioner
		}
		targetSCName := *pvc.Spec.StorageClassName
		sc, err := scCache.Get(targetSCName)
		if err != nil {
			logrus.Errorf("failed to get storage class %s, %v", targetSCName, err)
			return provisioner
		}
		provisioner = sc.Provisioner
	}
	return provisioner
}
