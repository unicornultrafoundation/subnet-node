package virtualbox

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/hardware_detector"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

var vboxLog = logrus.WithField("component", "vboxmanage")

// VBoxManageExecutor provides functions to execute VBoxManage commands
type VBoxManageExecutor struct {
	vmDir string
}

// NewVBoxManageExecutor creates a new VBoxManage executor
func NewVBoxManageExecutor(vmDir string) *VBoxManageExecutor {
	return &VBoxManageExecutor{
		vmDir: vmDir,
	}
}

// executeCommand executes a VBoxManage command and returns the output
func (e *VBoxManageExecutor) executeCommand(args ...string) (string, error) {
	cmd := exec.Command("VBoxManage", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("VBoxManage command failed: %w, output: %s", err, string(output))
	}
	return string(output), nil
}

// CreateVM creates a new VM using VBoxManage
func (e *VBoxManageExecutor) CreateVM(vmName string, osType string) error {
	vboxLog.Infof("Creating VM: %s", vmName)

	cmd := exec.Command("VBoxManage", "createvm", "--name", vmName, "--ostype", osType, "--register")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create VM: %w, output: %s", err, string(output))
	}

	vboxLog.Infof("VM created successfully: %s", vmName)
	return nil
}

// CreateVMWithoutOS creates a new VM without specifying OS type
func (e *VBoxManageExecutor) CreateVMWithoutOS(vmName string) error {
	vboxLog.Infof("Creating VM without OS type: %s", vmName)

	cmd := exec.Command("VBoxManage", "createvm", "--name", vmName, "--register")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create VM without OS: %w, output: %s", err, string(output))
	}

	vboxLog.Infof("VM created successfully without OS: %s", vmName)
	return nil
}

// ConfigureVMHardware configures VM hardware settings with dynamic hardware detection
func (e *VBoxManageExecutor) ConfigureVMHardware(vmName string, cpuCount int, memoryMB int) error {
	vboxLog.Infof("Configuring VM hardware for: %s", vmName)

	// Detect hardware and get appropriate settings
	hardware, err := hardware_detector.DetectHardware()
	if err != nil {
		vboxLog.Warnf("Hardware detection failed, using fallback settings: %v", err)
		// Fallback to basic settings
		return e.configureVMHardwareFallback(vmName, cpuCount, memoryMB)
	}

	settings := hardware_detector.GetVirtualBoxSettings(hardware)
	vboxLog.Infof("Detected hardware: %+v", hardware)
	vboxLog.Infof("Using VirtualBox settings: %+v", settings)

	// Set CPU count
	if err := e.setCPUs(vmName, cpuCount); err != nil {
		return fmt.Errorf("failed to set CPU count: %w", err)
	}

	// Set memory
	if err := e.setMemory(vmName, memoryMB); err != nil {
		return fmt.Errorf("failed to set memory: %w", err)
	}

	// Set VRAM based on detected hardware
	if err := e.setVRAM(vmName, settings.VRAMMB); err != nil {
		return fmt.Errorf("failed to set VRAM: %w", err)
	}

	// Set chipset based on detected hardware
	if err := e.setChipset(vmName, settings.Chipset); err != nil {
		return fmt.Errorf("failed to set chipset: %w", err)
	}

	// Set firmware based on detected hardware
	if err := e.setFirmware(vmName, settings.Firmware); err != nil {
		return fmt.Errorf("failed to set firmware: %w", err)
	}

	// Set graphics controller based on detected hardware
	if err := e.setGraphicsController(vmName, settings.GraphicsController); err != nil {
		return fmt.Errorf("failed to set graphics controller: %w", err)
	}

	// Set IOAPIC based on detected hardware
	if err := e.setIOAPIC(vmName, settings.IOAPICEnabled); err != nil {
		return fmt.Errorf("failed to set IOAPIC: %w", err)
	}

	// Set boot order
	if err := e.setBootOrder(vmName); err != nil {
		return fmt.Errorf("failed to set boot order: %w", err)
	}

	// Configure input devices
	if err := e.configureInputDevices(vmName); err != nil {
		return fmt.Errorf("failed to configure input devices: %w", err)
	}

	// Configure USB based on detected hardware
	if err := e.configureUSBWithSettings(vmName, settings.USBController); err != nil {
		return fmt.Errorf("failed to configure USB: %w", err)
	}

	// Configure audio based on detected hardware
	if err := e.configureAudioWithSettings(vmName, settings.AudioController, settings.AudioOutput, settings.AudioInput); err != nil {
		return fmt.Errorf("failed to configure audio: %w", err)
	}

	vboxLog.Infof("VM hardware configuration completed for: %s", vmName)
	return nil
}

// ConfigureNetwork configures network adapter settings
func (e *VBoxManageExecutor) ConfigureNetwork(vmName string, networkType string) error {
	vboxLog.Infof("Configuring network adapter for: %s", vmName)

	cmd := exec.Command("VBoxManage", "modifyvm", vmName, "--nic1", networkType, "--cableconnected1", "on")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to configure network adapter: %w, output: %s", err, string(output))
	}

	vboxLog.Infof("Network adapter configured successfully for: %s", vmName)
	return nil
}

// GenerateCloudInitISO creates a cloud-init ISO from the meta-data and user-data files
func (e *VBoxManageExecutor) GenerateCloudInitISO(cloudInitDir string) (string, error) {
	isoPath := filepath.Join(cloudInitDir, "cloud-init.iso")
	metaDataPath := filepath.Join(cloudInitDir, "meta-data")
	userDataPath := filepath.Join(cloudInitDir, "user-data")

	// Try mkisofs, fallback to genisoimage if not found
	cmd := exec.Command("mkisofs", "-o", isoPath, "-V", "cidata", "-r", "-J", userDataPath, metaDataPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Try genisoimage as fallback
		cmd = exec.Command("genisoimage", "-output", isoPath, "-volid", "cidata", "-joliet", "-rock", userDataPath, metaDataPath)
		if output2, err2 := cmd.CombinedOutput(); err2 != nil {
			return "", fmt.Errorf("failed to create cloud-init ISO: %w, output: %s, fallback output: %s", err, string(output), string(output2))
		}
	}
	return isoPath, nil
}

// AttachCloudInitISO attaches the cloud-init ISO to the VM
func (e *VBoxManageExecutor) AttachCloudInitISO(vmName, isoPath string) error {
	vboxLog.Infof("Attaching cloud-init ISO to VM: %s", vmName)
	_, err := e.executeCommand(
		"storageattach", vmName,
		"--storagectl", "VirtioSCSI",
		"--port", "2", "--device", "0",
		"--type", "dvddrive", "--medium", isoPath,
	)
	return err
}

// AttachCloudInitISOVirtioSCSI attaches the cloud-init ISO to the VM using VirtioSCSI controller
func (e *VBoxManageExecutor) AttachCloudInitISOVirtioSCSI(vmName, isoPath string) error {
	vboxLog.Infof("Attaching cloud-init ISO to VM using VirtioSCSI: %s", vmName)
	_, err := e.executeCommand(
		"storageattach", vmName,
		"--storagectl", "VirtioSCSI",
		"--port", "1", "--device", "0",
		"--type", "dvddrive", "--medium", isoPath,
	)
	return err
}

// Update SetupStorage to accept cloudInitISO
func (e *VBoxManageExecutor) SetupStorage(vmName string, req types.VMCreateRequest, isoPath string, cloudInitISO string) error {
	vboxLog.Infof("Setting up storage for VM: %s", vmName)

	// Create virtual disk
	diskPath := filepath.Join(e.vmDir, vmName, fmt.Sprintf("%s.vdi", vmName))
	diskSizeGB := req.DiskSizeGB
	if _, err := e.createVirtualDisk(diskPath, diskSizeGB); err != nil {
		return fmt.Errorf("failed to create virtual disk: %w", err)
	}

	// Add VirtioSCSI controller
	if err := e.addVirtioSCSIController(vmName); err != nil {
		return fmt.Errorf("failed to add VirtioSCSI controller: %w", err)
	}

	// Attach disk to controller
	if err := e.attachDisk(vmName, diskPath); err != nil {
		return fmt.Errorf("failed to attach disk: %w", err)
	}

	// Attach ISO if available
	if isoPath != "" {
		if err := e.attachISO(vmName, isoPath); err != nil {
			return fmt.Errorf("failed to attach ISO: %w", err)
		}
	}
	// Attach cloud-init ISO if provided
	if cloudInitISO != "" {
		if err := e.AttachCloudInitISO(vmName, cloudInitISO); err != nil {
			return fmt.Errorf("failed to attach cloud-init ISO: %w", err)
		}
	}

	vboxLog.Infof("Storage setup completed for VM: %s", vmName)
	return nil
}

// StartVM starts a VM
func (e *VBoxManageExecutor) StartVM(vmId string, headless bool) error {
	vboxLog.Infof("Starting VM: %s (headless: %v)", vmId, headless)

	var cmd *exec.Cmd
	if headless {
		cmd = exec.Command("VBoxManage", "startvm", vmId, "--type", "headless")
	} else {
		cmd = exec.Command("VBoxManage", "startvm", vmId)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start VM: %w, output: %s", err, string(output))
	}

	vboxLog.Infof("VM started successfully: %s", vmId)
	return nil
}

// StopVM stops a VM
func (e *VBoxManageExecutor) StopVM(vmId string) error {
	vboxLog.Infof("Stopping VM: %s", vmId)

	cmd := exec.Command("VBoxManage", "controlvm", vmId, "poweroff")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop VM: %w, output: %s", err, string(output))
	}

	vboxLog.Infof("VM stopped successfully: %s", vmId)
	return nil
}

// PauseVM pauses a VM
func (e *VBoxManageExecutor) PauseVM(vmId string) error {
	vboxLog.Infof("Pausing VM: %s", vmId)

	cmd := exec.Command("VBoxManage", "controlvm", vmId, "pause")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to pause VM: %w, output: %s", err, string(output))
	}

	vboxLog.Infof("VM paused successfully: %s", vmId)
	return nil
}

// ResumeVM resumes a VM
func (e *VBoxManageExecutor) ResumeVM(vmId string) error {
	vboxLog.Infof("Resuming VM: %s", vmId)

	cmd := exec.Command("VBoxManage", "controlvm", vmId, "resume")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to resume VM: %w, output: %s", err, string(output))
	}

	vboxLog.Infof("VM resumed successfully: %s", vmId)
	return nil
}

// ResetVM resets a VM
func (e *VBoxManageExecutor) ResetVM(vmId string) error {
	vboxLog.Infof("Resetting VM: %s", vmId)

	cmd := exec.Command("VBoxManage", "controlvm", vmId, "reset")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to reset VM: %w, output: %s", err, string(output))
	}

	vboxLog.Infof("VM reset successfully: %s", vmId)
	return nil
}

// DeleteVM deletes a VM and its associated files
func (e *VBoxManageExecutor) DeleteVM(vmId string) error {
	vboxLog.Infof("Deleting VM: %s", vmId)

	cmd := exec.Command("VBoxManage", "unregistervm", vmId, "--delete")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete VM: %w, output: %s", err, string(output))
	}

	vboxLog.Infof("VM deleted successfully: %s", vmId)
	return nil
}

// GetVMStatus gets the current status of a VM
func (e *VBoxManageExecutor) GetVMStatus(vmId string) (string, error) {
	cmd := exec.Command("VBoxManage", "showvminfo", vmId, "--machinereadable")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get VM status: %w", err)
	}

	// Parse the output to find VMState
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "VMState=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.Trim(parts[1], `"`), nil
			}
		}
	}

	return "unknown", nil
}

// ListVMs lists all registered VMs
func (e *VBoxManageExecutor) ListVMs() ([]string, error) {
	cmd := exec.Command("VBoxManage", "list", "vms")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs: %w", err)
	}

	var vms []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse "VMName" {UUID} format
		if strings.Contains(line, `"`) {
			parts := strings.SplitN(line, `"`, 3)
			if len(parts) >= 2 {
				vms = append(vms, parts[1])
			}
		}
	}

	return vms, nil
}

// CheckVBoxManageVersion checks if VBoxManage is available and returns its version
func (e *VBoxManageExecutor) CheckVBoxManageVersion() (string, error) {
	cmd := exec.Command("VBoxManage", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("VBoxManage not found or not accessible: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// Helper functions for individual VBoxManage operations

func (e *VBoxManageExecutor) setOSType(vmName, osType string) error {
	vboxLog.Infof("Setting OS type: %s", osType)
	_, err := e.executeCommand("modifyvm", vmName, "--ostype", osType)
	return err
}

func (e *VBoxManageExecutor) setCPUs(vmName string, cpuCount int) error {
	vboxLog.Infof("Setting CPU count to %d", cpuCount)
	_, err := e.executeCommand("modifyvm", vmName, "--cpus", fmt.Sprintf("%d", cpuCount))
	return err
}

func (e *VBoxManageExecutor) setMemory(vmName string, memoryMB int) error {
	vboxLog.Infof("Setting memory to %d MB", memoryMB)
	_, err := e.executeCommand("modifyvm", vmName, "--memory", fmt.Sprintf("%d", memoryMB))
	return err
}

func (e *VBoxManageExecutor) setVRAM(vmName string, vramMB int) error {
	vboxLog.Infof("Setting VRAM to %d MB", vramMB)
	_, err := e.executeCommand("modifyvm", vmName, "--vram", fmt.Sprintf("%d", vramMB))
	return err
}

func (e *VBoxManageExecutor) setChipset(vmName, chipset string) error {
	vboxLog.Infof("Setting chipset to %s", chipset)
	_, err := e.executeCommand("modifyvm", vmName, "--chipset", chipset)
	return err
}

func (e *VBoxManageExecutor) setFirmware(vmName, firmware string) error {
	vboxLog.Infof("Setting firmware to %s", firmware)
	_, err := e.executeCommand("modifyvm", vmName, "--firmware", firmware)
	return err
}

func (e *VBoxManageExecutor) setGraphicsController(vmName, controller string) error {
	vboxLog.Infof("Setting graphics controller to %s", controller)
	_, err := e.executeCommand("modifyvm", vmName, "--graphicscontroller", controller)
	return err
}

func (e *VBoxManageExecutor) setIOAPIC(vmName string, enabled bool) error {
	status := "off"
	if enabled {
		status = "on"
	}
	vboxLog.Infof("Setting IOAPIC to %s", status)
	_, err := e.executeCommand("modifyvm", vmName, "--ioapic", status)
	return err
}

func (e *VBoxManageExecutor) setBootOrder(vmId string) error {
	vboxLog.Infof("Setting boot order")
	_, err := e.executeCommand("modifyvm", vmId, "--boot1", "dvd", "--boot2", "disk", "--boot3", "none", "--boot4", "none")
	return err
}

func (e *VBoxManageExecutor) configureInputDevices(vmId string) error {
	vboxLog.Infof("Configuring input devices")
	_, err := e.executeCommand("modifyvm", vmId, "--mouse", "usbtablet", "--keyboard", "usb")
	return err
}

func (e *VBoxManageExecutor) configureUSB(vmName string) error {
	vboxLog.Infof("Configuring USB")
	_, err := e.executeCommand("modifyvm", vmName, "--usbohci", "off", "--usbehci", "off", "--usbxhci", "on")
	return err
}

func (e *VBoxManageExecutor) configureAudio(vmName string) error {
	vboxLog.Infof("Configuring audio")
	_, err := e.executeCommand("modifyvm", vmName, "--audio-controller", "hda", "--audio-out", "on", "--audio-in", "off")
	return err
}

// configureVMHardwareFallback provides fallback hardware configuration when detection fails
func (e *VBoxManageExecutor) configureVMHardwareFallback(vmName string, cpuCount int, memoryMB int) error {
	vboxLog.Infof("Using fallback hardware configuration for: %s", vmName)

	// Set CPU count
	if err := e.setCPUs(vmName, cpuCount); err != nil {
		return fmt.Errorf("failed to set CPU count: %w", err)
	}

	// Set memory
	if err := e.setMemory(vmName, memoryMB); err != nil {
		return fmt.Errorf("failed to set memory: %w", err)
	}

	// Set VRAM (video memory) - fallback to 16MB
	if err := e.setVRAM(vmName, 16); err != nil {
		return fmt.Errorf("failed to set VRAM: %w", err)
	}

	// Set chipset based on runtime architecture
	chipset := "ich9"
	if runtime.GOARCH == "arm64" || runtime.GOARCH == "aarch64" {
		chipset = "armv8virtual"
	}
	if err := e.setChipset(vmName, chipset); err != nil {
		return fmt.Errorf("failed to set chipset: %w", err)
	}

	// Set firmware - fallback to EFI for most architectures
	firmware := "efi"
	if runtime.GOARCH == "386" || runtime.GOARCH == "i386" {
		firmware = "bios"
	}
	if err := e.setFirmware(vmName, firmware); err != nil {
		return fmt.Errorf("failed to set firmware: %w", err)
	}

	// Set graphics controller - fallback to vmsvga
	if err := e.setGraphicsController(vmName, "vmsvga"); err != nil {
		return fmt.Errorf("failed to set graphics controller: %w", err)
	}

	// Set IOAPIC based on architecture
	ioapicEnabled := true
	if runtime.GOARCH == "arm64" || runtime.GOARCH == "aarch64" || runtime.GOARCH == "arm" {
		ioapicEnabled = false
	}
	if err := e.setIOAPIC(vmName, ioapicEnabled); err != nil {
		return fmt.Errorf("failed to set IOAPIC: %w", err)
	}

	// Set boot order
	if err := e.setBootOrder(vmName); err != nil {
		return fmt.Errorf("failed to set boot order: %w", err)
	}

	// Configure input devices
	if err := e.configureInputDevices(vmName); err != nil {
		return fmt.Errorf("failed to configure input devices: %w", err)
	}

	// Configure USB
	if err := e.configureUSB(vmName); err != nil {
		return fmt.Errorf("failed to configure USB: %w", err)
	}

	// Configure audio
	if err := e.configureAudio(vmName); err != nil {
		return fmt.Errorf("failed to configure audio: %w", err)
	}

	vboxLog.Infof("Fallback VM hardware configuration completed for: %s", vmName)
	return nil
}

// configureUSBWithSettings configures USB with specific controller settings
func (e *VBoxManageExecutor) configureUSBWithSettings(vmName, usbController string) error {
	vboxLog.Infof("Configuring USB with controller: %s", usbController)

	switch usbController {
	case "xHCI":
		_, err := e.executeCommand("modifyvm", vmName, "--usbohci", "off", "--usbehci", "off", "--usbxhci", "on")
		return err
	case "EHCI":
		_, err := e.executeCommand("modifyvm", vmName, "--usbohci", "off", "--usbehci", "on", "--usbxhci", "off")
		return err
	case "OHCI":
		_, err := e.executeCommand("modifyvm", vmName, "--usbohci", "on", "--usbehci", "off", "--usbxhci", "off")
		return err
	default:
		// Default to xHCI
		_, err := e.executeCommand("modifyvm", vmName, "--usbohci", "off", "--usbehci", "off", "--usbxhci", "on")
		return err
	}
}

// configureAudioWithSettings configures audio with specific settings
func (e *VBoxManageExecutor) configureAudioWithSettings(vmName, controller, output, input string) error {
	vboxLog.Infof("Configuring audio with controller: %s, output: %s, input: %s", controller, output, input)
	_, err := e.executeCommand("modifyvm", vmName, "--audio-controller", controller, "--audio-out", output, "--audio-in", input)
	return err
}

// createVirtualDisk creates a virtual disk for a VM
func (e *VBoxManageExecutor) createVirtualDisk(diskPath string, diskSizeGB int) (string, error) {
	vboxLog.Infof("Creating virtual disk")

	_, err := e.executeCommand("createhd", "--filename", diskPath, "--size", fmt.Sprintf("%d", diskSizeGB*1024), "--format", "VDI")
	if err != nil {
		return "", err
	}

	vboxLog.Infof("Virtual disk created: %s", diskPath)
	return diskPath, nil
}

func (e *VBoxManageExecutor) addVirtioSCSIController(vmName string) error {
	vboxLog.Infof("Adding VirtioSCSI controller")
	_, err := e.executeCommand("storagectl", vmName, "--name", "VirtioSCSI", "--add", "virtio-scsi", "--bootable", "on")
	return err
}

func (e *VBoxManageExecutor) attachDisk(vmName, diskPath string) error {
	vboxLog.Infof("Attaching disk to VirtioSCSI controller")
	_, err := e.executeCommand("storageattach", vmName, "--storagectl", "VirtioSCSI", "--port", "0", "--device", "0", "--type", "hdd", "--medium", diskPath)
	return err
}

// AttachExistingVDI attaches an existing VDI file to a VM using VirtioSCSI
func (e *VBoxManageExecutor) AttachExistingVDI(vmName, vdiPath string) error {
	vboxLog.Infof("Attaching existing VDI to VM using VirtioSCSI: %s", vmName)

	// First, add a VirtioSCSI controller if it doesn't exist
	if err := e.addVirtioSCSIController(vmName); err != nil {
		return fmt.Errorf("failed to add VirtioSCSI controller: %w", err)
	}

	// Attach the VDI to the VirtioSCSI controller
	_, err := e.executeCommand("storageattach", vmName, "--storagectl", "VirtioSCSI", "--port", "0", "--device", "0", "--type", "hdd", "--medium", vdiPath)
	return err
}

// ResizeVDI resizes a VDI file to the specified size in GB
func (e *VBoxManageExecutor) ResizeVDI(vdiPath string, diskSizeGB int) error {
	vboxLog.Infof("Resizing VDI file: %s to %d GB", vdiPath, diskSizeGB)

	// VBoxManage modifymedium disk <filename> --resize <size in MB>
	sizeMB := diskSizeGB * 1024
	_, err := e.executeCommand("modifymedium", "disk", vdiPath, "--resize", fmt.Sprintf("%d", sizeMB))
	if err != nil {
		return fmt.Errorf("failed to resize VDI file: %w", err)
	}

	vboxLog.Infof("VDI file resized successfully: %s to %d GB", vdiPath, diskSizeGB)
	return nil
}

func (e *VBoxManageExecutor) attachISO(vmName, isoPath string) error {
	vboxLog.Infof("Attaching ISO to VirtioSCSI controller")
	_, err := e.executeCommand("storageattach", vmName, "--storagectl", "VirtioSCSI", "--port", "1", "--device", "0", "--type", "dvddrive", "--medium", isoPath)
	return err
}

// UpdateCPUCores updates the CPU cores of a VM
func (e *VBoxManageExecutor) UpdateCPUCores(vmName string, cpuCores int) error {
	vboxLog.Infof("Updating CPU cores for VM %s to %d", vmName, cpuCores)
	cmd := exec.Command("VBoxManage", "modifyvm", vmName, "--cpus", fmt.Sprintf("%d", cpuCores))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update CPU cores: %w, output: %s", err, string(output))
	}
	return nil
}

// UpdateMemory updates the memory of a VM
func (e *VBoxManageExecutor) UpdateMemory(vmName string, memoryMB int) error {
	vboxLog.Infof("Updating memory for VM %s to %d MB", vmName, memoryMB)
	cmd := exec.Command("VBoxManage", "modifyvm", vmName, "--memory", fmt.Sprintf("%d", memoryMB))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update memory: %w, output: %s", err, string(output))
	}
	return nil
}

// SetupSSHPortForward sets up port forwarding for SSH from a host port to the VM's port 22
func (e *VBoxManageExecutor) SetupSSHPortForward(vmName string, hostPort int, guestPort int) error {
	vboxLog.Infof("Setting up SSH port forwarding: host %d -> guest %d for VM: %s", hostPort, guestPort, vmName)
	// Remove any existing rule named "ssh"
	_, _ = e.executeCommand("modifyvm", vmName, "--natpf1", "delete", "ssh")
	// Add new rule
	rule := fmt.Sprintf("ssh,tcp,,%d,,%d", hostPort, guestPort)
	_, err := e.executeCommand("modifyvm", vmName, "--natpf1", rule)
	return err
}
