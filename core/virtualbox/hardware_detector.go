package virtualbox

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/sirupsen/logrus"
)

var hardwareLog = logrus.WithField("component", "hardware-detector")

// HardwareInfo contains detected hardware information
type HardwareInfo struct {
	Architecture      string `json:"architecture"`
	OSType            string `json:"os_type"`
	CPUCount          int    `json:"cpu_count"`
	TotalMemoryMB     int    `json:"total_memory_mb"`
	AvailableMemoryMB int    `json:"available_memory_mb"`
	GPUType           string `json:"gpu_type"`
	GPUCount          int    `json:"gpu_count"`
	IsAppleSilicon    bool   `json:"is_apple_silicon"`
	IsIntelMac        bool   `json:"is_intel_mac"`
	IsLinux           bool   `json:"is_linux"`
	IsWindows         bool   `json:"is_windows"`
}

// VirtualBoxSettings contains the recommended VirtualBox settings based on hardware
type VirtualBoxSettings struct {
	OSType             string `json:"os_type"`
	Chipset            string `json:"chipset"`
	Firmware           string `json:"firmware"`
	GraphicsController string `json:"graphics_controller"`
	VRAMMB             int    `json:"vram_mb"`
	IOAPICEnabled      bool   `json:"ioapic_enabled"`
	USBController      string `json:"usb_controller"`
	AudioController    string `json:"audio_controller"`
	AudioOutput        string `json:"audio_output"`
	AudioInput         string `json:"audio_input"`
}

// HardwareDetector provides hardware detection capabilities
type HardwareDetector struct{}

// NewHardwareDetector creates a new hardware detector
func NewHardwareDetector() *HardwareDetector {
	return &HardwareDetector{}
}

// DetectHardware detects the current system hardware
func (d *HardwareDetector) DetectHardware() (*HardwareInfo, error) {
	info := &HardwareInfo{
		Architecture: runtime.GOARCH,
		OSType:       runtime.GOOS,
	}

	// Detect CPU information
	if err := d.detectCPU(info); err != nil {
		hardwareLog.Warnf("Failed to detect CPU: %v", err)
	}

	// Detect memory information
	if err := d.detectMemory(info); err != nil {
		hardwareLog.Warnf("Failed to detect memory: %v", err)
	}

	// Detect GPU information
	if err := d.detectGPU(info); err != nil {
		hardwareLog.Warnf("Failed to detect GPU: %v", err)
	}

	// Detect platform-specific information
	d.detectPlatform(info)

	hardwareLog.Infof("Hardware detection completed: %+v", info)
	return info, nil
}

// GetVirtualBoxSettings returns appropriate VirtualBox settings based on hardware
func (d *HardwareDetector) GetVirtualBoxSettings(hardware *HardwareInfo) *VirtualBoxSettings {
	settings := &VirtualBoxSettings{}

	// Determine OS type based on architecture
	settings.OSType = d.determineOSType(hardware)

	// Determine chipset based on architecture
	settings.Chipset = d.determineChipset(hardware)

	// Determine firmware based on architecture and OS
	settings.Firmware = d.determineFirmware(hardware)

	// Determine graphics controller based on platform
	settings.GraphicsController = d.determineGraphicsController(hardware)

	// Determine VRAM based on GPU capabilities
	settings.VRAMMB = d.determineVRAM(hardware)

	// Determine IOAPIC settings based on architecture
	settings.IOAPICEnabled = d.determineIOAPIC(hardware)

	// Determine USB controller settings
	settings.USBController = d.determineUSBController(hardware)

	// Determine audio settings
	settings.AudioController = d.determineAudioController(hardware)
	settings.AudioOutput = d.determineAudioOutput(hardware)
	settings.AudioInput = d.determineAudioInput(hardware)

	hardwareLog.Infof("VirtualBox settings determined: %+v", settings)
	return settings
}

// detectCPU detects CPU information
func (d *HardwareDetector) detectCPU(info *HardwareInfo) error {
	// Get CPU count
	cpuCount, err := cpu.Counts(false)
	if err != nil {
		return fmt.Errorf("failed to get CPU count: %w", err)
	}
	info.CPUCount = cpuCount

	// Get detailed CPU info
	cpuInfo, err := cpu.Info()
	if err != nil {
		return fmt.Errorf("failed to get CPU info: %w", err)
	}

	if len(cpuInfo) > 0 {
		hardwareLog.Infof("CPU detected: %s with %d cores", cpuInfo[0].ModelName, cpuCount)
	}

	return nil
}

// detectMemory detects memory information
func (d *HardwareDetector) detectMemory(info *HardwareInfo) error {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("failed to get memory info: %w", err)
	}

	info.TotalMemoryMB = int(vmStat.Total / (1024 * 1024))
	info.AvailableMemoryMB = int(vmStat.Available / (1024 * 1024))

	hardwareLog.Infof("Memory detected: %d MB total, %d MB available", info.TotalMemoryMB, info.AvailableMemoryMB)
	return nil
}

// detectGPU detects GPU information
func (d *HardwareDetector) detectGPU(info *HardwareInfo) error {
	switch runtime.GOOS {
	case "darwin":
		return d.detectMacGPU(info)
	case "linux":
		return d.detectLinuxGPU(info)
	case "windows":
		return d.detectWindowsGPU(info)
	default:
		info.GPUType = "Unknown"
		info.GPUCount = 0
		return nil
	}
}

// detectMacGPU detects GPU on macOS
func (d *HardwareDetector) detectMacGPU(info *HardwareInfo) error {
	if runtime.GOARCH == "arm64" {
		// Apple Silicon
		info.IsAppleSilicon = true
		info.GPUType = "Apple Silicon GPU"

		// Try to get GPU core count
		cmd := exec.Command("sh", "-c", "ioreg -l | grep gpu-core-count")
		output, err := cmd.Output()
		if err == nil {
			parts := strings.Split(string(output), "=")
			if len(parts) == 2 {
				coreCount := strings.TrimSpace(parts[1])
				if count, err := strconv.Atoi(coreCount); err == nil {
					info.GPUCount = count
				}
			}
		}

		// Fallback to reasonable default for Apple Silicon
		if info.GPUCount == 0 {
			info.GPUCount = 8 // Typical for M1/M2 series
		}
	} else {
		// Intel Mac
		info.IsIntelMac = true
		info.GPUType = "Intel Integrated Graphics"
		info.GPUCount = 1
	}

	hardwareLog.Infof("Mac GPU detected: %s with %d cores", info.GPUType, info.GPUCount)
	return nil
}

// detectLinuxGPU detects GPU on Linux
func (d *HardwareDetector) detectLinuxGPU(info *HardwareInfo) error {
	info.IsLinux = true

	// Try to detect NVIDIA GPU
	if _, err := exec.LookPath("nvidia-smi"); err == nil {
		cmd := exec.Command("nvidia-smi", "--query-gpu=name", "--format=csv,noheader")
		output, err := cmd.Output()
		if err == nil {
			gpuNames := strings.Split(strings.TrimSpace(string(output)), "\n")
			info.GPUType = "NVIDIA " + gpuNames[0]
			info.GPUCount = len(gpuNames)
			hardwareLog.Infof("Linux NVIDIA GPU detected: %s with %d GPUs", info.GPUType, info.GPUCount)
			return nil
		}
	}

	// Try to detect AMD GPU
	if _, err := exec.LookPath("rocm-smi"); err == nil {
		cmd := exec.Command("rocm-smi", "--showproductname")
		output, err := cmd.Output()
		if err == nil {
			info.GPUType = "AMD " + strings.TrimSpace(string(output))
			info.GPUCount = 1
			hardwareLog.Infof("Linux AMD GPU detected: %s", info.GPUType)
			return nil
		}
	}

	// Fallback to integrated graphics
	info.GPUType = "Integrated Graphics"
	info.GPUCount = 1
	hardwareLog.Infof("Linux GPU detection: using integrated graphics")
	return nil
}

// detectWindowsGPU detects GPU on Windows
func (d *HardwareDetector) detectWindowsGPU(info *HardwareInfo) error {
	info.IsWindows = true

	// Try to detect NVIDIA GPU using nvidia-smi
	if _, err := exec.LookPath("nvidia-smi"); err == nil {
		cmd := exec.Command("nvidia-smi", "--query-gpu=name", "--format=csv,noheader")
		output, err := cmd.Output()
		if err == nil {
			gpuNames := strings.Split(strings.TrimSpace(string(output)), "\n")
			info.GPUType = "NVIDIA " + gpuNames[0]
			info.GPUCount = len(gpuNames)
			hardwareLog.Infof("Windows NVIDIA GPU detected: %s with %d GPUs", info.GPUType, info.GPUCount)
			return nil
		}
	}

	// Fallback to integrated graphics
	info.GPUType = "Integrated Graphics"
	info.GPUCount = 1
	hardwareLog.Infof("Windows GPU detection: using integrated graphics")
	return nil
}

// detectPlatform detects platform-specific information
func (d *HardwareDetector) detectPlatform(info *HardwareInfo) {
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			info.IsAppleSilicon = true
		} else {
			info.IsIntelMac = true
		}
	case "linux":
		info.IsLinux = true
	case "windows":
		info.IsWindows = true
	}
}

// determineOSType determines the appropriate OS type for VirtualBox
func (d *HardwareDetector) determineOSType(hardware *HardwareInfo) string {
	switch hardware.Architecture {
	case "arm64", "aarch64":
		return "Ubuntu_ARM64"
	case "amd64", "x86_64":
		return "Ubuntu_64"
	case "arm":
		return "Ubuntu"
	case "386", "i386":
		return "Ubuntu"
	default:
		return "Ubuntu_64"
	}
}

// determineChipset determines the appropriate chipset for VirtualBox
func (d *HardwareDetector) determineChipset(hardware *HardwareInfo) string {
	switch hardware.Architecture {
	case "arm64", "aarch64":
		return "armv8virtual"
	case "amd64", "x86_64":
		return "ich9"
	case "arm":
		return "armv8virtual"
	case "386", "i386":
		return "ich9"
	default:
		return "ich9"
	}
}

// determineFirmware determines the appropriate firmware for VirtualBox
func (d *HardwareDetector) determineFirmware(hardware *HardwareInfo) string {
	switch hardware.Architecture {
	case "arm64", "aarch64":
		return "efi"
	case "amd64", "x86_64":
		return "efi"
	case "arm":
		return "efi"
	case "386", "i386":
		return "bios"
	default:
		return "efi"
	}
}

// determineGraphicsController determines the appropriate graphics controller
func (d *HardwareDetector) determineGraphicsController(hardware *HardwareInfo) string {
	if hardware.IsAppleSilicon {
		return "vmsvga"
	} else if hardware.IsIntelMac {
		return "vmsvga"
	} else if hardware.IsLinux {
		return "vmsvga"
	} else if hardware.IsWindows {
		return "vboxsvga"
	}

	// Default based on architecture
	switch hardware.Architecture {
	case "arm64", "aarch64":
		return "vmsvga"
	case "amd64", "x86_64":
		return "vboxsvga"
	default:
		return "vmsvga"
	}
}

// determineVRAM determines the appropriate VRAM size
func (d *HardwareDetector) determineVRAM(hardware *HardwareInfo) int {
	// Base VRAM on GPU capabilities
	if hardware.IsAppleSilicon {
		// Apple Silicon has good integrated graphics
		switch hardware.GPUCount {
		case 0, 1:
			return 32
		case 2, 3, 4:
			return 64
		case 5, 6, 7, 8:
			return 128
		default:
			return 256
		}
	} else if strings.Contains(hardware.GPUType, "NVIDIA") {
		// NVIDIA GPUs typically have more VRAM
		return 128
	} else if strings.Contains(hardware.GPUType, "AMD") {
		// AMD GPUs
		return 64
	} else {
		// Integrated graphics
		return 32
	}
}

// determineIOAPIC determines whether IOAPIC should be enabled
func (d *HardwareDetector) determineIOAPIC(hardware *HardwareInfo) bool {
	// IOAPIC is typically disabled for ARM64 architectures
	switch hardware.Architecture {
	case "arm64", "aarch64", "arm":
		return false
	case "amd64", "x86_64", "386", "i386":
		return true
	default:
		return true
	}
}

// determineUSBController determines the appropriate USB controller
func (d *HardwareDetector) determineUSBController(hardware *HardwareInfo) string {
	// Modern systems typically use XHCI
	return "xHCI"
}

// determineAudioController determines the appropriate audio controller
func (d *HardwareDetector) determineAudioController(hardware *HardwareInfo) string {
	// HDA is the most compatible audio controller
	return "hda"
}

// determineAudioOutput determines audio output settings
func (d *HardwareDetector) determineAudioOutput(hardware *HardwareInfo) string {
	return "on"
}

// determineAudioInput determines audio input settings
func (d *HardwareDetector) determineAudioInput(hardware *HardwareInfo) string {
	return "off"
}
