package virtualbox

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/common/fsutil"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/templates"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

var vboxLog = logrus.WithField("component", "vboxmanage")

// VBoxManageExecutor provides functions to execute VBoxManage commands
type VBoxManageExecutor struct {
	vmDir       string
	templateMgr *templates.TemplateManager
}

// NewVBoxManageExecutor creates a new VBoxManage executor
func NewVBoxManageExecutor(vmDir string) *VBoxManageExecutor {
	// Get template directory (relative to the project root)
	// We need to find the project root and then navigate to templates
	projectRoot := findProjectRoot()
	templateDir := filepath.Join(projectRoot, "core", "virtualbox", "templates")

	return &VBoxManageExecutor{
		vmDir:       vmDir,
		templateMgr: templates.NewTemplateManager(templateDir),
	}
}

// findProjectRoot finds the project root directory by looking for go.mod
func findProjectRoot() string {
	// Start from current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		// Fallback to a reasonable default
		return "."
	}

	// Walk up the directory tree to find go.mod
	dir := currentDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	// Fallback to current directory
	return currentDir
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

// ConfigureVMHardware configures VM hardware settings with dynamic hardware detection
func (e *VBoxManageExecutor) ConfigureVMHardware(vmName string, cpuCount int, memoryMB int) error {
	vboxLog.Infof("Configuring VM hardware for: %s", vmName)

	// Detect hardware and get appropriate settings
	detector := NewHardwareDetector()
	hardware, err := detector.DetectHardware()
	if err != nil {
		vboxLog.Warnf("Hardware detection failed, using fallback settings: %v", err)
		// Fallback to basic settings
		return e.configureVMHardwareFallback(vmName, cpuCount, memoryMB)
	}

	settings := detector.GetVirtualBoxSettings(hardware)
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

// GenerateCloudInitFiles generates cloud-init meta-data and user-data files from templates
func (e *VBoxManageExecutor) GenerateCloudInitFiles(vmName, hostname, username, password string) (metaDataPath, userDataPath, cloudInitDir string, err error) {
	cloudInitDir = filepath.Join(e.vmDir, vmName, "cloud-init")
	if err := fsutil.DirWritable(cloudInitDir); err != nil {
		return "", "", "", fmt.Errorf("failed to create cloud-init dir: %w", err)
	}

	// Prepare template data
	templateData := templates.CloudInitData{
		InstanceID: vmName,
		Hostname:   hostname,
		Username:   username,
		Password:   password,
	}

	// Generate meta-data from template
	metaData, err := e.templateMgr.GenerateMetaData(templateData)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate meta-data from template: %w", err)
	}

	metaDataPath = filepath.Join(cloudInitDir, "meta-data")
	if err := ioutil.WriteFile(metaDataPath, []byte(metaData), 0644); err != nil {
		return "", "", "", fmt.Errorf("failed to write meta-data: %w", err)
	}

	// Generate user-data from template
	userData, err := e.templateMgr.GenerateUserData(templateData)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate user-data from template: %w", err)
	}

	userDataPath = filepath.Join(cloudInitDir, "user-data")
	if err := ioutil.WriteFile(userDataPath, []byte(userData), 0644); err != nil {
		return "", "", "", fmt.Errorf("failed to write user-data: %w", err)
	}

	return metaDataPath, userDataPath, cloudInitDir, nil
}

// GenerateCloneVMCloudInitFiles generates cloud-init meta-data and user-data files from clone VM templates
func (e *VBoxManageExecutor) GenerateCloneVMCloudInitFiles(vmName, hostname, username, password string) (metaDataPath, userDataPath, cloudInitDir string, err error) {
	cloudInitDir = filepath.Join(e.vmDir, vmName, "cloud-init")
	if err := fsutil.DirWritable(cloudInitDir); err != nil {
		return "", "", "", fmt.Errorf("failed to create cloud-init dir: %w", err)
	}

	// Prepare template data
	templateData := templates.CloudInitData{
		InstanceID: vmName,
		Hostname:   hostname,
		Username:   username,
		Password:   password,
	}

	// Generate meta-data from clone VM template
	metaData, err := e.templateMgr.GenerateCloneVMMetaData(templateData)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate clone VM meta-data from template: %w", err)
	}

	metaDataPath = filepath.Join(cloudInitDir, "meta-data")
	if err := ioutil.WriteFile(metaDataPath, []byte(metaData), 0644); err != nil {
		return "", "", "", fmt.Errorf("failed to write meta-data: %w", err)
	}

	// Generate user-data from clone VM template
	userData, err := e.templateMgr.GenerateCloneVMUserData(templateData)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate clone VM user-data from template: %w", err)
	}

	userDataPath = filepath.Join(cloudInitDir, "user-data")
	if err := ioutil.WriteFile(userDataPath, []byte(userData), 0644); err != nil {
		return "", "", "", fmt.Errorf("failed to write user-data: %w", err)
	}

	return metaDataPath, userDataPath, cloudInitDir, nil
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

// Update SetupStorage to accept cloudInitISO
func (e *VBoxManageExecutor) SetupStorage(vmName string, req vbtypes.VMCreateRequest, isoPath string, cloudInitISO string) error {
	vboxLog.Infof("Setting up storage for VM: %s", vmName)

	// Create virtual disk
	diskPath := filepath.Join(e.vmDir, vmName, fmt.Sprintf("%s.vdi", vmName))
	diskSizeGB := req.DiskSizeGB
	if _, err := e.createVirtualDisk(vmName, diskPath, diskSizeGB); err != nil {
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
func (e *VBoxManageExecutor) StartVM(vmName string, headless bool) error {
	vboxLog.Infof("Starting VM: %s (headless: %v)", vmName, headless)

	var cmd *exec.Cmd
	if headless {
		cmd = exec.Command("VBoxManage", "startvm", vmName, "--type", "headless")
	} else {
		cmd = exec.Command("VBoxManage", "startvm", vmName)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start VM: %w, output: %s", err, string(output))
	}

	vboxLog.Infof("VM started successfully: %s", vmName)
	return nil
}

// StopVM stops a VM
func (e *VBoxManageExecutor) StopVM(vmName string) error {
	vboxLog.Infof("Stopping VM: %s", vmName)

	cmd := exec.Command("VBoxManage", "controlvm", vmName, "poweroff")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop VM: %w, output: %s", err, string(output))
	}

	vboxLog.Infof("VM stopped successfully: %s", vmName)
	return nil
}

// PauseVM pauses a VM
func (e *VBoxManageExecutor) PauseVM(vmName string) error {
	vboxLog.Infof("Pausing VM: %s", vmName)

	cmd := exec.Command("VBoxManage", "controlvm", vmName, "pause")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to pause VM: %w, output: %s", err, string(output))
	}

	vboxLog.Infof("VM paused successfully: %s", vmName)
	return nil
}

// ResumeVM resumes a VM
func (e *VBoxManageExecutor) ResumeVM(vmName string) error {
	vboxLog.Infof("Resuming VM: %s", vmName)

	cmd := exec.Command("VBoxManage", "controlvm", vmName, "resume")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to resume VM: %w, output: %s", err, string(output))
	}

	vboxLog.Infof("VM resumed successfully: %s", vmName)
	return nil
}

// ResetVM resets a VM
func (e *VBoxManageExecutor) ResetVM(vmName string) error {
	vboxLog.Infof("Resetting VM: %s", vmName)

	cmd := exec.Command("VBoxManage", "controlvm", vmName, "reset")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to reset VM: %w, output: %s", err, string(output))
	}

	vboxLog.Infof("VM reset successfully: %s", vmName)
	return nil
}

// DeleteVM deletes a VM and its associated files
func (e *VBoxManageExecutor) DeleteVM(vmName string) error {
	vboxLog.Infof("Deleting VM: %s", vmName)

	cmd := exec.Command("VBoxManage", "unregistervm", vmName, "--delete")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete VM: %w, output: %s", err, string(output))
	}

	vboxLog.Infof("VM deleted successfully: %s", vmName)
	return nil
}

// GetVMStatus gets the current status of a VM
func (e *VBoxManageExecutor) GetVMStatus(vmName string) (string, error) {
	cmd := exec.Command("VBoxManage", "showvminfo", vmName, "--machinereadable")
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

func (e *VBoxManageExecutor) setBootOrder(vmName string) error {
	vboxLog.Infof("Setting boot order")
	_, err := e.executeCommand("modifyvm", vmName, "--boot1", "dvd", "--boot2", "disk", "--boot3", "none", "--boot4", "none")
	return err
}

func (e *VBoxManageExecutor) configureInputDevices(vmName string) error {
	vboxLog.Infof("Configuring input devices")
	_, err := e.executeCommand("modifyvm", vmName, "--mouse", "usbtablet", "--keyboard", "usb")
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
func (e *VBoxManageExecutor) createVirtualDisk(vmName string, diskPath string, diskSizeGB int) (string, error) {
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

// CreateTemplateVM creates a template VM with ubuntu/ubuntu credentials
func (e *VBoxManageExecutor) CreateTemplateVM(templateName string, osType string, cpuCount int, memoryMB int, diskSizeGB int) error {
	vboxLog.Infof("Creating template VM: %s", templateName)

	// Create the base VM
	if err := e.CreateVM(templateName, osType); err != nil {
		return fmt.Errorf("failed to create template VM: %w", err)
	}

	// Configure VM hardware
	if err := e.ConfigureVMHardware(templateName, cpuCount, memoryMB); err != nil {
		return fmt.Errorf("failed to configure template VM hardware: %w", err)
	}

	// Configure network adapter
	if err := e.ConfigureNetwork(templateName, "nat"); err != nil {
		return fmt.Errorf("failed to configure template VM network: %w", err)
	}

	// Generate cloud-init files with ubuntu/ubuntu credentials
	_, _, cloudInitDir, err := e.GenerateCloudInitFiles(templateName, templateName, "ubuntu", "ubuntu")
	if err != nil {
		return fmt.Errorf("failed to generate cloud-init files for template: %w", err)
	}

	// Generate cloud-init ISO
	cloudInitISO, err := e.GenerateCloudInitISO(cloudInitDir)
	if err != nil {
		return fmt.Errorf("failed to generate cloud-init ISO for template: %w", err)
	}

	// Create template VM request
	req := vbtypes.VMCreateRequest{
		Name:       templateName,
		CPUCores:   cpuCount,
		MemoryMB:   memoryMB,
		DiskSizeGB: diskSizeGB,
		OSType:     osType,
		Username:   "ubuntu",
		Password:   "ubuntu",
	}

	// Setup storage with cloud-init ISO
	if err := e.SetupStorage(templateName, req, "", cloudInitISO); err != nil {
		return fmt.Errorf("failed to setup template VM storage: %w", err)
	}

	vboxLog.Infof("Template VM created successfully: %s with ubuntu/ubuntu credentials", templateName)
	return nil
}

// CloneTemplateVM creates a new VM by cloning an existing template VM
func (e *VBoxManageExecutor) CloneTemplateVM(templateName, newVMName string) error {
	vboxLog.Infof("Cloning template VM %s to %s", templateName, newVMName)

	// Clone the VM using VBoxManage
	_, err := e.executeCommand("clonevm", templateName, "--name", newVMName, "--register")
	if err != nil {
		return fmt.Errorf("failed to clone template VM: %w", err)
	}

	return nil
}
