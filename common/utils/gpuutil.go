package utils

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// GetGpuUsage returns the GPU usage for all processes.
func GetGpuUsage() (map[int32]uint64, error) {
	// Check if the operating system supports NVIDIA
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("NVIDIA GPU usage is only supported on Linux")
	}

	// Check if nvidia-smi command is available
	if _, err := exec.LookPath("nvidia-smi"); err != nil {
		return nil, fmt.Errorf("nvidia-smi command not found: %v", err)
	}

	// Execute nvidia-smi command to get GPU processes
	cmd := exec.Command("nvidia-smi", "--query-compute-apps=pid,used_memory", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute nvidia-smi: %v", err)
	}

	// Parse the output to get GPU usage for each process
	lines := strings.Split(string(output), "\n")
	gpuUsage := make(map[int32]uint64)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) != 2 {
			continue
		}
		gpuPid, err := strconv.Atoi(strings.TrimSpace(fields[0]))
		if err != nil {
			return nil, fmt.Errorf("failed to parse PID from nvidia-smi output: %v", err)
		}
		usedMemory, err := strconv.ParseUint(strings.TrimSpace(fields[1]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse used memory from nvidia-smi output: %v", err)
		}
		gpuUsage[int32(gpuPid)] += usedMemory
	}

	return gpuUsage, nil
}
