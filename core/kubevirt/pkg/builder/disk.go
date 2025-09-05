package builder

import kubevirtv1 "kubevirt.io/api/core/v1"

func (v *VMBuilder) Disk(diskName, diskBus string, isCDRom bool, bootOrder uint) *VMBuilder {
	var (
		exist bool
		index int
		disks = v.VirtualMachine.Spec.Template.Spec.Domain.Devices.Disks
	)
	for i, disk := range disks {
		if disk.Name == diskName {
			exist = true
			index = i
			break
		}
	}
	diskDevice := kubevirtv1.DiskDevice{
		Disk: &kubevirtv1.DiskTarget{
			Bus: kubevirtv1.DiskBus(diskBus),
		},
	}
	if isCDRom {
		diskDevice = kubevirtv1.DiskDevice{
			CDRom: &kubevirtv1.CDRomTarget{
				Bus: kubevirtv1.DiskBus(diskBus),
			},
		}
	}
	disk := kubevirtv1.Disk{
		Name:       diskName,
		DiskDevice: diskDevice,
	}
	if bootOrder > 0 {
		disk.BootOrder = &bootOrder
	}
	if exist {
		disks[index] = disk
	} else {
		disks = append(disks, disk)
	}
	v.VirtualMachine.Spec.Template.Spec.Domain.Devices.Disks = disks
	return v
}

func (v *VMBuilder) Volume(diskName string, volume kubevirtv1.Volume) *VMBuilder {
	var (
		exist   bool
		index   int
		volumes = v.VirtualMachine.Spec.Template.Spec.Volumes
	)
	for i, e := range volumes {
		if e.Name == diskName {
			exist = true
			index = i
			break
		}
	}

	if exist {
		volumes[index] = volume
	} else {
		volumes = append(volumes, volume)
	}
	v.VirtualMachine.Spec.Template.Spec.Volumes = volumes
	return v
}
