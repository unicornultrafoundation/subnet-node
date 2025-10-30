package vbox_service

import (
	"bufio"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/hardware_detector"

	"github.com/sirupsen/logrus"

	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/cmd_exec"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

var serviceLog = logrus.WithField("service", "vbox_service")

type VBoxService struct {
	mu      sync.RWMutex
	vBoxCmd cmd_exec.Command
}

func NewVBoxService() *VBoxService {

	return &VBoxService{
		vBoxCmd: cmd_exec.GetVBoxCmd(),
		mu:      sync.RWMutex{},
	}
}

var (
	reVMNameUUID      = regexp.MustCompile(`"(.+)" {([0-9a-f-]+)}`)
	reVMInfoLine      = regexp.MustCompile(`(?:"(.+)"|(.+))=(?:"(.*)"|(.*))`)
	reColonLine       = regexp.MustCompile(`(.+):\s+(.*)`)
	reMachineNotFound = regexp.MustCompile(`Could not find a registered machine named '(.+)'`)
)

func (s *VBoxService) GetVM(vmId string) (*vbtypes.VM, error) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	args := []string{"showvminfo", vmId, "--machinereadable"}
	stdout, stderr, err := s.vBoxCmd.Run(args...)

	if err != nil {
		if reMachineNotFound.FindString(stderr) != "" {
			return nil, cmd_exec.ErrMachineNotExist
		}
		return nil, err
	}

	/* Read all VM info into a map */
	props := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		res := reVMInfoLine.FindStringSubmatch(scanner.Text())
		if res == nil {
			continue
		}
		key := res[1]
		if key == "" {
			key = res[2]
		}
		val := res[3]
		if val == "" {
			val = res[4]
		}
		props[key] = val
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading machine info: %w", err)
	}

	// error that occured during parsing
	var perr error

	sp := func(field string, def ...string) string {
		if v, exists := props[field]; exists {
			return v
		}
		if len(def) < 1 {
			return ""
		}
		return def[0]
	}

	up := func(field string, def ...uint) uint {
		if v, exists := props[field]; exists {
			n, err := strconv.ParseUint(v, 10, 32)
			if err != nil {
				perr = err
				return 0
			}
			return uint(n)
		}
		if len(def) < 1 {
			return 0
		}
		return def[0]
	}

	/* Extract basic info */
	vm := &vbtypes.VM{
		ID:       sp("UUID"),
		Name:     sp("name"),
		Status:   vbtypes.VMStatus(sp("VMState")),
		CPUCores: int(up("cpus")),
		MemoryMB: int(up("memory")),
		// IPAddress: sp("IPAddress"),
		// SSHPort: up("SSHPort"),
		VMFolder: sp("CfgFile"),
	}

	// Determine disk size since --machinereadable does not include it reliably
	if sizeGB, err := s.detectDiskSizeGB(props); err == nil {
		vm.DiskSizeGB = sizeGB
	}

	if perr != nil {
		return nil, fmt.Errorf("parsing machine props failed: %w", perr)
	}

	return vm, nil
}

func (s *VBoxService) DeleteVM(vmId string) error {
	vm, err := s.GetVM(vmId)
	if err != nil {
		return fmt.Errorf("failed to get VM: %w", err)
	}

	if err := s.powerOff(vm); err != nil {
		return fmt.Errorf("failed to power off VM: %w", err)
	}

	_, _, err = s.vBoxCmd.Run("unregistervm", vm.ID, "--delete")

	if err != nil {
		return fmt.Errorf("failed to delete VM: %w", err)
	}

	return nil
}

// detectDiskSizeGB finds the primary disk medium from VM properties and queries its size via VBoxManage showmediuminfo.
func (s *VBoxService) detectDiskSizeGB(props map[string]string) (int, error) {
	// Try to find any attached HDD medium path from props like "SATA-0-0"="/path/to/disk.vdi"
	var mediumPath string
	for k, v := range props {
		// Ignore empty and non-medium values
		if v == "none" || v == "" {
			continue
		}
		lv := strings.ToLower(v)
		if strings.HasSuffix(lv, ".vdi") || strings.HasSuffix(lv, ".vmdk") || strings.HasSuffix(lv, ".vhd") || strings.HasSuffix(lv, ".vhdx") {
			mediumPath = v
			break
		}
		// Some props may contain quoted paths; strip quotes if present
		if strings.Contains(lv, ".vdi") || strings.Contains(lv, ".vmdk") || strings.Contains(lv, ".vhd") || strings.Contains(lv, ".vhdx") {
			mediumPath = strings.Trim(v, "\"")
			break
		}
		// Alternatively, keys like "*-* -*" usually denote attachments; keep heuristic key filter
		_ = k
	}

	if mediumPath == "" {
		return 0, fmt.Errorf("no disk medium found")
	}

	// Normalize any surrounding quotes
	mediumPath = strings.TrimSpace(strings.Trim(mediumPath, "\""))

	// If VM folder is present and mediumPath is relative, make it absolute
	if !filepath.IsAbs(mediumPath) {
		if cfg := props["CfgFile"]; cfg != "" {
			// CfgFile is an absolute path to the .vbox; use its dir as base
			base := filepath.Dir(cfg)
			mediumPath = filepath.Join(base, mediumPath)
		}
	}

	// Query VirtualBox for medium info
	stdout, _, err := s.vBoxCmd.Run("showmediuminfo", mediumPath)
	if err != nil {
		return 0, err
	}

	// Parse size; VirtualBox outputs either "Capacity:" or "Logical size:" lines
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Capacity:") || strings.HasPrefix(line, "Logical size:") {
			// Examples:
			// Capacity:      102400 MBytes
			// Logical size:  10.00 GB
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				// last numeric before unit might be fields[len-2], but safer: find first token that parses as float
				var numStr string
				for _, f := range fields {
					if _, err := strconv.ParseFloat(strings.TrimRight(f, ","), 64); err == nil {
						numStr = strings.TrimRight(f, ",")
						break
					}
				}
				if numStr != "" {
					val, _ := strconv.ParseFloat(numStr, 64)
					unit := ""
					if len(fields) > 0 {
						unit = fields[len(fields)-1]
					}
					sizeGB := toGB(val, unit)
					if sizeGB > 0 {
						return sizeGB, nil
					}
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return 0, fmt.Errorf("disk size not found")
}

func toGB(val float64, unit string) int {
	switch strings.ToUpper(strings.TrimSuffix(unit, ",")) {
	case "B", "BYTES":
		return int(val / (1024.0 * 1024.0 * 1024.0))
	case "KB", "KBYTES":
		return int(val / (1024.0 * 1024.0))
	case "MB", "MBYTES", "MIB":
		return int(val / 1024.0)
	case "GB", "GBYTES", "GIB":
		return int(val)
	case "TB", "TBYTES", "TIB":
		return int(val * 1024.0)
	default:
		return int(val)
	}
}

func (s *VBoxService) ListVMs() ([]*vbtypes.VM, error) {

	stdout, _, err := s.vBoxCmd.Run("list", "vms")
	if err != nil {
		return nil, fmt.Errorf("unable to list vms: %w", err)
	}
	vms := []*vbtypes.VM{}
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		res := reVMNameUUID.FindStringSubmatch(scanner.Text())
		if res == nil {
			continue
		}
		m, err := s.GetVM(res[1])
		if err != nil {
			// Sometimes a VM is listed but not available, so we need to handle this.
			if errors.Is(err, cmd_exec.ErrMachineNotExist) {
				continue
			} else {
				return nil, fmt.Errorf("unable to get machine info: %w", err)
			}
		}
		vms = append(vms, m)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading machine list: %w", err)
	}
	return vms, nil
}

func (s *VBoxService) CreateVMWithoutOS(name string) (*vbtypes.VM, error) {
	return s.createVM(name)
}

func (s *VBoxService) createVM(name string) (*vbtypes.VM, error) {

	if name == "" {
		return nil, fmt.Errorf("machine name is empty")
	}

	// Check if a machine with the given name already exists.
	ms, err := s.ListVMs()
	if err != nil {
		return nil, err
	}
	for _, m := range ms {
		if m.Name == name {
			return nil, cmd_exec.ErrMachineExist
		}
	}

	// Create and register the machine.
	args := []string{"createvm", "--name", name, "--register"}

	if _, _, err = s.vBoxCmd.Run(args...); err != nil {
		return nil, err
	}

	return s.GetVM(name)
}

func (s *VBoxService) ConfigureVMHardware(vmName string, cpuCount int, memoryMB int) error {

	hardware, err := hardware_detector.DetectHardware()
	if err != nil {
		serviceLog.Warnf("Hardware detection failed, using fallback settings: %v", err)
		return s.configureVMHardwareFallback(vmName, cpuCount, memoryMB)
	}

	if err := s.setCPUs(vmName, cpuCount); err != nil {
		return fmt.Errorf("failed to set CPU count: %w", err)
	}

	if err := s.setCPUs(vmName, cpuCount); err != nil {
		return fmt.Errorf("failed to set CPU count: %w", err)
	}

	if err := s.setMemory(vmName, memoryMB); err != nil {
		return fmt.Errorf("failed to set memory: %w", err)
	}

	// Set info based on hardware
	settings := hardware_detector.GetVirtualBoxSettings(hardware)

	if err := s.setVRAM(vmName, settings.VRAMMB); err != nil {
		return fmt.Errorf("failed to set VRAM: %w", err)
	}

	if err := s.setChipset(vmName, settings.Chipset); err != nil {
		return fmt.Errorf("failed to set chipset: %w", err)
	}

	if err := s.setFirmware(vmName, settings.Firmware); err != nil {
		return fmt.Errorf("failed to set firmware: %w", err)
	}

	if err := s.setGraphicsController(vmName, settings.GraphicsController); err != nil {
		return fmt.Errorf("failed to set graphics controller: %w", err)
	}

	if err := s.setIOAPIC(vmName, settings.IOAPICEnabled); err != nil {
		return fmt.Errorf("failed to set IOAPIC: %w", err)
	}

	if err := s.setBootOrder(vmName); err != nil {
		return fmt.Errorf("failed to set boot order: %w", err)
	}

	if err := s.configureInputDevices(vmName); err != nil {
		return fmt.Errorf("failed to configure input devices: %w", err)
	}

	if err := s.configureUSBWithSettings(vmName, settings.USBController); err != nil {
		return fmt.Errorf("failed to configure USB: %w", err)
	}

	if err := s.configureAudioWithSettings(vmName, settings.AudioController, settings.AudioOutput, settings.AudioInput); err != nil {
		return fmt.Errorf("failed to configure audio: %w", err)
	}

	return nil
}

func (s *VBoxService) ConfigureNetwork(vmName string, networkType string) error {
	_, _, err := s.vBoxCmd.Run("modifyvm", vmName, "--nic1", networkType, "--cableconnected1", "on")
	return err
}

func (s *VBoxService) configureVMHardwareFallback(vmName string, cpuCount int, memoryMB int) error {
	if err := s.setCPUs(vmName, cpuCount); err != nil {
		return fmt.Errorf("failed to set CPU count: %w", err)
	}
	if err := s.setMemory(vmName, memoryMB); err != nil {
		return fmt.Errorf("failed to set memory: %w", err)
	}
	if err := s.setVRAM(vmName, 16); err != nil {
		return fmt.Errorf("failed to set VRAM: %w", err)
	}

	arch := runtime.GOARCH

	chipset := "ich9"
	if arch == "arm64" || arch == "aarch64" {
		chipset = "armv8virtual"
	}

	if err := s.setChipset(vmName, chipset); err != nil {
		return fmt.Errorf("failed to set chipset: %w", err)
	}

	firmware := "efi"
	if arch == "386" || arch == "i386" {
		firmware = "bios"
	}
	if err := s.setFirmware(vmName, firmware); err != nil {
		return fmt.Errorf("failed to set firmware: %w", err)
	}

	// Set graphics controller - fallback to vmsvga
	if err := s.setGraphicsController(vmName, "vmsvga"); err != nil {
		return fmt.Errorf("failed to set graphics controller: %w", err)
	}

	ioapicEnabled := true
	if arch == "arm64" || arch == "aarch64" || arch == "arm" {
		ioapicEnabled = false
	}
	if err := s.setIOAPIC(vmName, ioapicEnabled); err != nil {
		return fmt.Errorf("failed to set IOAPIC: %w", err)
	}

	if err := s.setBootOrder(vmName); err != nil {
		return fmt.Errorf("failed to set boot order: %w", err)
	}

	if err := s.configureInputDevices(vmName); err != nil {
		return fmt.Errorf("failed to configure input devices: %w", err)
	}

	if err := s.configureUSB(vmName); err != nil {
		return fmt.Errorf("failed to configure USB: %w", err)
	}

	if err := s.configureAudio(vmName); err != nil {
		return fmt.Errorf("failed to configure audio: %w", err)
	}

	return nil
}

func (s *VBoxService) setOSType(vmName, osType string) error {
	_, _, err := s.vBoxCmd.Run("modifyvm", vmName, "--ostype", osType)
	return err
}

func (s *VBoxService) setCPUs(vmName string, cpuCount int) error {
	_, _, err := s.vBoxCmd.Run("modifyvm", vmName, "--cpus", fmt.Sprintf("%d", cpuCount))
	return err
}

func (s *VBoxService) setMemory(vmName string, memoryMB int) error {
	_, _, err := s.vBoxCmd.Run("modifyvm", vmName, "--memory", fmt.Sprintf("%d", memoryMB))
	return err
}

func (s *VBoxService) setVRAM(vmName string, vramMB int) error {
	_, _, err := s.vBoxCmd.Run("modifyvm", vmName, "--vram", fmt.Sprintf("%d", vramMB))
	return err
}

func (s *VBoxService) setChipset(vmName, chipset string) error {
	_, _, err := s.vBoxCmd.Run("modifyvm", vmName, "--chipset", chipset)
	return err
}

func (s *VBoxService) setFirmware(vmName, firmware string) error {
	_, _, err := s.vBoxCmd.Run("modifyvm", vmName, "--firmware", firmware)
	return err
}

func (s *VBoxService) setGraphicsController(vmName, controller string) error {
	_, _, err := s.vBoxCmd.Run("modifyvm", vmName, "--graphicscontroller", controller)
	return err
}

func (s *VBoxService) setIOAPIC(vmName string, enabled bool) error {
	status := "off"
	if enabled {
		status = "on"
	}
	_, _, err := s.vBoxCmd.Run("modifyvm", vmName, "--ioapic", status)
	return err
}

func (s *VBoxService) setBootOrder(vmName string) error {
	_, _, err := s.vBoxCmd.Run("modifyvm", vmName, "--boot1", "dvd", "--boot2", "disk", "--boot3", "none", "--boot4", "none")
	return err
}

func (s *VBoxService) configureInputDevices(vmName string) error {
	_, _, err := s.vBoxCmd.Run("modifyvm", vmName, "--mouse", "usbtablet", "--keyboard", "usb")
	return err
}

func (s *VBoxService) configureUSB(vmName string) error {
	_, _, err := s.vBoxCmd.Run("modifyvm", vmName, "--usbohci", "off", "--usbehci", "off", "--usbxhci", "on")
	return err
}

func (s *VBoxService) configureUSBWithSettings(vmName, usbController string) error {

	switch usbController {

	case "xHCI":
		_, _, err := s.vBoxCmd.Run("modifyvm", vmName, "--usbohci", "off", "--usbehci", "off", "--usbxhci", "on")
		return err
	case "EHCI":
		_, _, err := s.vBoxCmd.Run("modifyvm", vmName, "--usbohci", "off", "--usbehci", "on", "--usbxhci", "off")
		return err
	case "OHCI":
		_, _, err := s.vBoxCmd.Run("modifyvm", vmName, "--usbohci", "on", "--usbehci", "off", "--usbxhci", "off")
		return err
	default:
		// Default to xHCI
		_, _, err := s.vBoxCmd.Run("modifyvm", vmName, "--usbohci", "off", "--usbehci", "off", "--usbxhci", "on")
		return err
	}
}

func (s *VBoxService) configureAudio(vmName string) error {
	_, _, err := s.vBoxCmd.Run("modifyvm", vmName, "--audio-controller", "hda", "--audio-out", "on", "--audio-in", "off")
	return err
}

func (s *VBoxService) configureAudioWithSettings(vmName, controller, output, input string) error {
	_, _, err := s.vBoxCmd.Run("modifyvm", vmName, "--audio-controller", controller, "--audio-out", output, "--audio-in", input)
	return err
}

func (s *VBoxService) ResizeVDI(vdiPath string, diskSizeGB int) error {
	sizeMB := diskSizeGB * 1024
	_, _, err := s.vBoxCmd.Run("modifymedium", "disk", vdiPath, "--resize", fmt.Sprintf("%d", sizeMB))
	return err
}

func (s *VBoxService) AttachExistingVDI(vmName, vdiPath string) error {

	// add virtioSCSI controller if not exists
	if err := s.addVirtioSCSIController(vmName); err != nil {
		return fmt.Errorf("failed to add VirtioSCSI controller: %w", err)
	}

	// attach vdi to virtioSCSI controller
	_, _, err := s.vBoxCmd.Run("storageattach", vmName, "--storagectl", "VirtioSCSI", "--port", "0", "--device", "0", "--type", "hdd", "--medium", vdiPath)
	return err
}

// AttachCloudInitISOVirtioSCSI attaches the cloud-init ISO to the VM using VirtioSCSI controller
func (s *VBoxService) AttachCloudInitISOVirtioSCSI(vmName, isoPath string) error {
	_, _, err := s.vBoxCmd.Run("storageattach", vmName, "--storagectl", "VirtioSCSI", "--port", "1", "--device", "0", "--type", "dvddrive", "--medium", isoPath)
	return err
}

func (s *VBoxService) addVirtioSCSIController(vmName string) error {
	_, _, err := s.vBoxCmd.Run("storagectl", vmName, "--name", "VirtioSCSI", "--add", "virtio-scsi", "--bootable", "on")
	return err
}

func (s *VBoxService) powerOff(vm *vbtypes.VM) error {

	switch vm.Status {
	case vbtypes.Poweroff, vbtypes.Aborted, vbtypes.Saved:
		return nil
	}
	_, _, err := s.vBoxCmd.Run("controlvm", vm.Name, "poweroff")
	return err
}
