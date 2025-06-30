package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateGlobalResourceCap(t *testing.T) {
	manifest := &Manifest{
		Groups: []Group{
			{
				Name: "group1",
				Services: []Service{
					{
						Name:  "svc1",
						Count: 2,
						Resources: &ServiceResources{
							CPU:     &CPUResource{Units: ResourceValue{Value: 500, Unit: "m"}},     // 0.5 core per instance
							Memory:  &MemoryResource{Size: ResourceValue{Value: 1024, Unit: "Mi"}}, // 1024 MiB per instance
							Storage: &StorageResource{Size: ResourceValue{Value: 10, Unit: "Gi"}},  // 10 GiB per instance
						},
					},
					{
						Name:  "svc2",
						Count: 1,
						Resources: &ServiceResources{
							CPU:     &CPUResource{Units: ResourceValue{Value: 1, Unit: "core"}},      // 1 core per instance
							Memory:  &MemoryResource{Size: ResourceValue{Value: 2048, Unit: "Mi"}},   // 2048 MiB per instance
							Storage: &StorageResource{Size: ResourceValue{Value: 10240, Unit: "Mi"}}, // 10 GiB per instance
						},
					},
				},
			},
		},
	}

	cpuCores, gpuCores, gpuMemoryMB, memoryMB, diskGB, err := manifest.CalculateGlobalResourceCap()
	assert.NoError(t, err)
	assert.InDelta(t, 2.0, cpuCores, 0.001) // (2*0.5) + (1*1) = 2.0
	assert.Equal(t, int64(0), gpuCores)
	assert.Equal(t, int64(0), gpuMemoryMB)
	assert.Equal(t, int64(4096), memoryMB) // (2*1024) + (1*2048) = 4096
	assert.Equal(t, int64(30), diskGB)     // (2*10) + (1*10) = 30
}

func TestParseRAMString(t *testing.T) {
	mb, err := parseRAMString("8192MB")
	assert.NoError(t, err)
	assert.Equal(t, int64(8192), mb)

	mb, err = parseRAMString("8GB")
	assert.NoError(t, err)
	assert.Equal(t, int64(8192), mb)

	_, err = parseRAMString("2TB")
	assert.Error(t, err)
}

func TestCalculateGlobalResourceCap_WithGPU(t *testing.T) {
	manifest := &Manifest{
		Groups: []Group{
			{
				Name: "group1",
				Services: []Service{
					{
						Name:  "svc1",
						Count: 1,
						Resources: &ServiceResources{
							CPU:    &CPUResource{Units: ResourceValue{Value: 1000, Unit: "m"}},
							Memory: &MemoryResource{Size: ResourceValue{Value: 4096, Unit: "Mi"}},
							GPU: &GPUResource{
								Units: 2,
								Attributes: GPUAttributes{
									Vendor: map[string][]GPUModel{
										"nvidia": {{Model: "A100", RAM: "8GB"}},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	cpuCores, gpuCores, gpuMemoryMB, memoryMB, diskGB, err := manifest.CalculateGlobalResourceCap()
	assert.NoError(t, err)
	assert.InDelta(t, 1.0, cpuCores, 0.001)
	assert.Equal(t, int64(2), gpuCores)
	assert.Equal(t, int64(16384), gpuMemoryMB) // 2 units * 8GB (8192MB) = 16384MB
	assert.Equal(t, int64(4096), memoryMB)
	assert.Equal(t, int64(0), diskGB)
}
