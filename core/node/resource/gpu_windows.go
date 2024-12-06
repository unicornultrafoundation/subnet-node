//go:build windows
// +build windows

package resource

// getGpu returns a placeholder for Windows as NVML is not supported
func getGpu() (GpuInfo, error) {
	return GpuInfo{
		Count: 0,
		Name:  "Windows platform - GPU detection not supported",
	}, nil
}
