package builder

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

const (
	defaultVMGenerateName = "subnet-"
	defaultVMNamespace    = "default"

	defaultVMCPUCores = 1
	defaultVMMemory   = "256Mi"

	SubnetAPIGroup                                        = "subnet.unicornultrafoundation.io"
	LabelAnnotationPrefixSubnet                           = SubnetAPIGroup + "/"
	LabelKeyVirtualMachineCreator                         = LabelAnnotationPrefixSubnet + "creator"
	LabelKeyVirtualMachineName                            = LabelAnnotationPrefixSubnet + "vmName"
	AnnotationKeyVirtualMachineSSHNames                   = LabelAnnotationPrefixSubnet + "sshNames"
	AnnotationKeyVirtualMachineWaitForLeaseInterfaceNames = LabelAnnotationPrefixSubnet + "waitForLeaseInterfaceNames"
	AnnotationKeyVirtualMachineDiskNames                  = LabelAnnotationPrefixSubnet + "diskNames"
	AnnotationKeyImageID                                  = LabelAnnotationPrefixSubnet + "imageId"
)

type VMBuilder struct {
	VirtualMachine             *kubevirtv1.VirtualMachine
	SSHNames                   []string
	WaitForLeaseInterfaceNames []string
}

func NewVMBuilder(creator string) *VMBuilder {
	vmLabels := map[string]string{
		LabelKeyVirtualMachineCreator: creator,
	}
	objectMeta := metav1.ObjectMeta{
		Namespace:    defaultVMNamespace,
		GenerateName: defaultVMGenerateName,
		Labels:       vmLabels,
		Annotations:  map[string]string{},
	}
	runStrategy := kubevirtv1.RunStrategyHalted
	cpu := &kubevirtv1.CPU{
		Cores: defaultVMCPUCores,
	}
	resources := kubevirtv1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse(defaultVMMemory),
			corev1.ResourceCPU:    *resource.NewQuantity(defaultVMCPUCores, resource.DecimalSI),
		},
	}
	template := &kubevirtv1.VirtualMachineInstanceTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: vmLabels,
		},
		Spec: kubevirtv1.VirtualMachineInstanceSpec{
			Domain: kubevirtv1.DomainSpec{
				CPU: cpu,
				Devices: kubevirtv1.Devices{
					Disks:      []kubevirtv1.Disk{},
					Interfaces: []kubevirtv1.Interface{},
				},
				Resources: resources,
			},
			Affinity: &corev1.Affinity{},
			Networks: []kubevirtv1.Network{},
			Volumes:  []kubevirtv1.Volume{},
		},
	}

	vm := &kubevirtv1.VirtualMachine{
		ObjectMeta: objectMeta,
		Spec: kubevirtv1.VirtualMachineSpec{
			RunStrategy: &runStrategy,
			Template:    template,
		},
	}
	return &VMBuilder{
		VirtualMachine:             vm,
		SSHNames:                   []string{},
		WaitForLeaseInterfaceNames: []string{},
	}
}
