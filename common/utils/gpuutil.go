package utils

import (
	"fmt"
	"runtime"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

// GetGpuUsageByPid returns the GPU usage for a given PID.
func GetGpuUsageByPid(pid int32) (uint64, error) {
	// Check if the operating system supports NVIDIA
	if runtime.GOOS != "linux" {
		return 0, fmt.Errorf("NVIDIA GPU usage is only supported on Linux")
	}

	// Initialize NVML library for GPU stats
	if ret := nvml.Init(); ret != nvml.SUCCESS {
		return 0, fmt.Errorf("failed to initialize NVML: %v", nvml.ErrorString(ret))
	}
	defer nvml.Shutdown()

	// Get GPU usage
	deviceCount, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return 0, fmt.Errorf("failed to get GPU device count: %v", nvml.ErrorString(ret))
	}

	var usedGpu uint64
	for i := 0; i < deviceCount; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			return 0, fmt.Errorf("failed to get GPU device handle: %v", nvml.ErrorString(ret))
		}

		// Get the processes running on the GPU
		processes, ret := device.GetComputeRunningProcesses()
		if ret != nvml.SUCCESS {
			return 0, fmt.Errorf("failed to get GPU processes: %v", nvml.ErrorString(ret))
		}

		// Check if the given PID is using the GPU
		for _, process := range processes {
			if process.Pid == uint32(pid) {
				usedGpu++
			}
		}
	}

	return usedGpu, nil
}
