package hardware_detector

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

// DetectHardware detects the current system hardware
func DetectHardware() (*HardwareInfo, error) {
	info := &HardwareInfo{
		Architecture: runtime.GOARCH,
		OSType:       runtime.GOOS,
	}

	// Detect CPU information
	if err := detectCPU(info); err != nil {
		hardwareLog.Warnf("Failed to detect CPU: %v", err)
	}

	// Detect memory information
	if err := detectMemory(info); err != nil {
		hardwareLog.Warnf("Failed to detect memory: %v", err)
	}

	// Detect GPU information
	if err := detectGPU(info); err != nil {
		hardwareLog.Warnf("Failed to detect GPU: %v", err)
	}

	// Detect platform-specific information
	detectPlatform(info)

	hardwareLog.Infof("Hardware detection completed: %+v", info)
	return info, nil
}

func ValidateOSTypeCompatibility(requestedOSType string) error {

	// Detect hardware to determine appropriate OS type
	hardware, err := DetectHardware()
	if err != nil {
		hardwareLog.Warnf("Hardware detection failed, skipping OS type validation: %v", err)
		return nil // Skip validation if we can't detect hardware
	}

	// Get the appropriate OS type for the detected hardware
	appropriateOSType := determineOSType(hardware)
	hardwareLog.Infof("Detected hardware architecture: %s", hardware.Architecture)
	hardwareLog.Infof("Appropriate OS type for hardware: %s", appropriateOSType)
	hardwareLog.Infof("Requested OS type: %s", requestedOSType)

	// Check if the requested OS type is compatible with the hardware architecture
	if !isOSTypeCompatible(requestedOSType, hardware.Architecture) {
		return fmt.Errorf("OS type '%s' is not compatible with hardware architecture '%s'. Recommended OS type: '%s'",
			requestedOSType, hardware.Architecture, appropriateOSType)
	}

	hardwareLog.Infof("OS type validation passed: %s is compatible with %s architecture", requestedOSType, hardware.Architecture)
	return nil

}

// GetVirtualBoxSettings returns appropriate VirtualBox settings based on hardware
func GetVirtualBoxSettings(hardware *HardwareInfo) *VirtualBoxSettings {
	settings := &VirtualBoxSettings{}

	// Determine OS type based on architecture
	settings.OSType = determineOSType(hardware)

	// Determine chipset based on architecture
	settings.Chipset = determineChipset(hardware)

	// Determine firmware based on architecture and OS
	settings.Firmware = determineFirmware(hardware)

	// Determine graphics controller based on platform
	settings.GraphicsController = determineGraphicsController(hardware)

	// Determine VRAM based on GPU capabilities
	settings.VRAMMB = determineVRAM(hardware)

	// Determine IOAPIC settings based on architecture
	settings.IOAPICEnabled = determineIOAPIC(hardware)

	// Determine USB controller settings
	settings.USBController = determineUSBController(hardware)

	// Determine audio settings
	settings.AudioController = determineAudioController(hardware)
	settings.AudioOutput = determineAudioOutput(hardware)
	settings.AudioInput = determineAudioInput(hardware)

	hardwareLog.Infof("VirtualBox settings determined: %+v", settings)
	return settings
}

// detectCPU detects CPU information
func detectCPU(info *HardwareInfo) error {
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
func detectMemory(info *HardwareInfo) error {
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
func detectGPU(info *HardwareInfo) error {
	switch runtime.GOOS {
	case "darwin":
		return detectMacGPU(info)
	case "linux":
		return detectLinuxGPU(info)
	case "windows":
		return detectWindowsGPU(info)
	default:
		info.GPUType = "Unknown"
		info.GPUCount = 0
		return nil
	}
}

// detectMacGPU detects GPU on macOS
func detectMacGPU(info *HardwareInfo) error {
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
func detectLinuxGPU(info *HardwareInfo) error {
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
func detectWindowsGPU(info *HardwareInfo) error {
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
func detectPlatform(info *HardwareInfo) {
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
func determineOSType(hardware *HardwareInfo) string {
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

// isOSTypeCompatible checks if an OS type is compatible with a given architecture
func isOSTypeCompatible(osType, architecture string) bool {
	// Define compatibility matrix
	compatibilityMap := map[string][]string{
		"Ubuntu_ARM64": {"arm64", "aarch64"},
		"Ubuntu_64":    {"amd64", "x86_64"},
		"Ubuntu":       {"arm", "386", "i386", "amd64", "x86_64", "arm64", "aarch64"},
		"Debian_ARM64": {"arm64", "aarch64"},
		"Debian_64":    {"amd64", "x86_64"},
		"Debian":       {"arm", "386", "i386", "amd64", "x86_64", "arm64", "aarch64"},
		// "Windows_ARM64": {"arm64", "aarch64"},
		// "Windows_64":    {"amd64", "x86_64"},
		// "Windows":       {"amd64", "x86_64"},
	}

	// Check if the OS type is in our compatibility map
	supportedArchitectures, exists := compatibilityMap[osType]
	if !exists {
		hardwareLog.Warnf("Unknown OS type: %s, allowing it to pass validation", osType)
		return true // Allow unknown OS types to pass validation
	}

	// Check if the architecture is supported by this OS type
	for _, supportedArch := range supportedArchitectures {
		if supportedArch == architecture {
			return true
		}
	}

	return false
}

// determineChipset determines the appropriate chipset for VirtualBox
func determineChipset(hardware *HardwareInfo) string {
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
func determineFirmware(hardware *HardwareInfo) string {
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
func determineGraphicsController(hardware *HardwareInfo) string {
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
func determineVRAM(hardware *HardwareInfo) int {
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
func determineIOAPIC(hardware *HardwareInfo) bool {
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
func determineUSBController(hardware *HardwareInfo) string {
	// Modern systems typically use XHCI
	return "xHCI"
}

// determineAudioController determines the appropriate audio controller
func determineAudioController(hardware *HardwareInfo) string {
	// HDA is the most compatible audio controller
	return "hda"
}

// determineAudioOutput determines audio output settings
func determineAudioOutput(hardware *HardwareInfo) string {
	return "on"
}

// determineAudioInput determines audio input settings
func determineAudioInput(hardware *HardwareInfo) string {
	return "off"
}
