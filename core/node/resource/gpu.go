//go:build !windows
// +build !windows

package resource

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func getAppleSiliconGPU() (GpuInfo, error) {
	// Get chip info
	cmd := exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
	output, err := cmd.Output()
	if err != nil {
		return GpuInfo{}, fmt.Errorf("failed to get CPU info: %v", err)
	}
	chipInfo := strings.TrimSpace(string(output))

	// Get GPU cores count using ioreg
	cmd = exec.Command("sh", "-c", "ioreg -l | grep gpu-core-count")
	output, err = cmd.Output()
	if err != nil {
		return GpuInfo{}, fmt.Errorf("failed to get GPU core count: %v", err)
	}

	// Extract the number from output like: "gpu-core-count" = 10
	parts := strings.Split(string(output), "=")
	if len(parts) != 2 {
		return GpuInfo{}, fmt.Errorf("unexpected ioreg output format")
	}

	// Clean and parse the number
	coreCount := strings.TrimSpace(parts[1])
	gpuCores, err := strconv.Atoi(coreCount)
	if err != nil {
		return GpuInfo{}, fmt.Errorf("failed to parse GPU core count: %v", err)
	}

	return GpuInfo{
		Count: gpuCores,
		Name:  fmt.Sprintf("Apple Silicon GPU (%s)", chipInfo),
	}, nil
}

func getGpu() (GpuInfo, error) {
	// Check if running on macOS
	if runtime.GOOS == "darwin" {
		// Check for arm64 architecture (Apple Silicon)
		if runtime.GOARCH == "arm64" {
			return getAppleSiliconGPU()
		}
		// For Intel Macs, return integrated graphics
		return GpuInfo{
			Count: 1,
			Name:  "Apple Integrated Graphics",
		}, nil
	}

	// // For non-macOS systems, try NVIDIA GPU detection
	// ret := nvml.Init()
	// if ret != nvml.SUCCESS {
	// 	return GpuInfo{}, nil
	// }
	// defer func() {
	// 	if ret := nvml.Shutdown(); ret != nvml.SUCCESS {
	// 		log.Errorf("Unable to shutdown NVML: %v", nvml.ErrorString(ret))
	// 	}
	// }()

	// count, ret := nvml.DeviceGetCount()
	// if ret != nvml.SUCCESS {
	// 	return GpuInfo{}, fmt.Errorf("unable to get device count: %v", nvml.ErrorString(ret))
	// }

	// gpuNames := make([]string, count)
	// for i := 0; i < count; i++ {
	// 	device, ret := nvml.DeviceGetHandleByIndex(i)
	// 	if ret != nvml.SUCCESS {
	// 		return GpuInfo{}, fmt.Errorf("unable to get device at index %d: %v", i, nvml.ErrorString(ret))
	// 	}
	// 	name, ret := device.GetName()
	// 	if ret != nvml.SUCCESS {
	// 		return GpuInfo{}, fmt.Errorf("unable to get name of device at index %d: %v", i, nvml.ErrorString(ret))
	// 	}
	// 	gpuNames[i] = name
	// }

	return GpuInfo{}, nil
}
