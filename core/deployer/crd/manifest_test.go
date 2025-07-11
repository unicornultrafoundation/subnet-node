package crd

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestManifestStructs(t *testing.T) {
	// Test creating a basic manifest
	manifest := &Manifest{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "depinsubnet.com/v1",
			Kind:       "Manifest",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-manifest",
			Namespace: "default",
		},
		Spec: ManifestSpec{
			Lease: LeaseSpec{
				Owner: "0x1234567890abcdef",
				ID:    12345,
			},
			Group: GroupSpec{
				Name: "test-group",
				Services: []ServiceSpec{
					{
						Name:  "test-service",
						Image: "nginx:latest",
						Resources: ResourceSpec{
							CPU: &CPUResource{
								Units: 1000,
								Attributes: []Attribute{
									{Name: "arch", Value: "x86_64"},
								},
							},
							Memory: &MemoryResource{
								Size: 512,
								Attributes: []Attribute{
									{Name: "type", Value: "RAM"},
								},
							},
						},
						Volumes: []VolumeSpec{
							{
								Name:     "data",
								Mount:    "/data",
								ReadOnly: false,
							},
						},
						Count: 1,
						Expose: []ExposeSpec{
							{
								IP:      "0.0.0.0",
								Port:    80,
								Proto:   "tcp",
								Service: "test-service",
								Global:  true,
							},
						},
					},
				},
			},
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(manifest)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Test JSON unmarshaling
	var unmarshaledManifest Manifest
	err = json.Unmarshal(jsonData, &unmarshaledManifest)
	require.NoError(t, err)

	// Verify the data is preserved
	assert.Equal(t, manifest.Spec.Lease.Owner, unmarshaledManifest.Spec.Lease.Owner)
	assert.Equal(t, manifest.Spec.Lease.ID, unmarshaledManifest.Spec.Lease.ID)
	assert.Equal(t, manifest.Spec.Group.Name, unmarshaledManifest.Spec.Group.Name)
	assert.Equal(t, len(manifest.Spec.Group.Services), len(unmarshaledManifest.Spec.Group.Services))
	assert.Equal(t, manifest.Spec.Group.Services[0].Count, unmarshaledManifest.Spec.Group.Services[0].Count)
	assert.Equal(t, len(manifest.Spec.Group.Services[0].Volumes), len(unmarshaledManifest.Spec.Group.Services[0].Volumes))
	assert.Equal(t, len(manifest.Spec.Group.Services[0].Expose), len(unmarshaledManifest.Spec.Group.Services[0].Expose))
}

func TestManifestValidation(t *testing.T) {
	// Test valid manifest
	validManifest := ExampleManifest()
	err := ValidateManifest(validManifest)
	assert.NoError(t, err)

	// Test invalid manifest - missing lease owner
	invalidManifest := &Manifest{
		Spec: ManifestSpec{
			Lease: LeaseSpec{
				Owner: "", // Missing owner
				ID:    12345,
			},
			Group: GroupSpec{
				Name:     "test-group",
				Services: []ServiceSpec{},
			},
		},
	}
	err = ValidateManifest(invalidManifest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lease owner is required")

	// Test invalid manifest - missing group name
	invalidManifest2 := &Manifest{
		Spec: ManifestSpec{
			Lease: LeaseSpec{
				Owner: "0x1234567890abcdef",
				ID:    12345,
			},
			Group: GroupSpec{
				Name:     "", // Missing group name
				Services: []ServiceSpec{},
			},
		},
	}
	err = ValidateManifest(invalidManifest2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "group name is required")

	// Test invalid manifest - no services
	invalidManifest3 := &Manifest{
		Spec: ManifestSpec{
			Lease: LeaseSpec{
				Owner: "0x1234567890abcdef",
				ID:    12345,
			},
			Group: GroupSpec{
				Name:     "test-group",
				Services: []ServiceSpec{}, // No services
			},
		},
	}
	err = ValidateManifest(invalidManifest3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one service is required")

	// Test invalid manifest - service count is 0
	invalidManifest4 := &Manifest{
		Spec: ManifestSpec{
			Lease: LeaseSpec{
				Owner: "0x1234567890abcdef",
				ID:    12345,
			},
			Group: GroupSpec{
				Name: "test-group",
				Services: []ServiceSpec{
					{
						Name:  "test-service",
						Image: "nginx:latest",
						Count: 0, // Invalid count
					},
				},
			},
		},
	}
	err = ValidateManifest(invalidManifest4)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service 0 count must be greater than 0")
}

func TestResourceSpecs(t *testing.T) {
	// Test CPU resource
	cpuResource := &CPUResource{
		Units: 2000,
		Attributes: []Attribute{
			{Name: "arch", Value: "x86_64"},
			{Name: "vendor", Value: "intel"},
		},
	}
	assert.Equal(t, uint32(2000), cpuResource.Units)
	assert.Len(t, cpuResource.Attributes, 2)

	// Test Memory resource
	memoryResource := &MemoryResource{
		Size: 1024,
		Attributes: []Attribute{
			{Name: "type", Value: "RAM"},
		},
	}
	assert.Equal(t, uint32(1024), memoryResource.Size)
	assert.Len(t, memoryResource.Attributes, 1)

	// Test GPU resource
	gpuResource := &GPUResource{
		Units: 1,
		Attributes: []Attribute{
			{Name: "vendor", Value: "nvidia"},
			{Name: "model", Value: "RTX 3080"},
		},
	}
	assert.Equal(t, uint32(1), gpuResource.Units)
	assert.Len(t, gpuResource.Attributes, 2)
}

func TestExposeSpec(t *testing.T) {
	exposeSpec := ExposeSpec{
		IP:      "0.0.0.0",
		Port:    80,
		Proto:   "tcp",
		Service: "web-service",
		Global:  true,
		HTTPOptions: &HTTPOptions{
			MaxBodySize: 1048576,
			ReadTimeout: 30,
			SendTimeout: 30,
			NextTries:   3,
			NextTimeout: 10,
			NextCases:   []string{"error", "timeout"},
			Hosts:       []string{"example.com", "www.example.com"},
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(exposeSpec)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Test JSON unmarshaling
	var unmarshaledExpose ExposeSpec
	err = json.Unmarshal(jsonData, &unmarshaledExpose)
	require.NoError(t, err)

	// Verify the data is preserved
	assert.Equal(t, exposeSpec.IP, unmarshaledExpose.IP)
	assert.Equal(t, exposeSpec.Port, unmarshaledExpose.Port)
	assert.Equal(t, exposeSpec.Proto, unmarshaledExpose.Proto)
	assert.Equal(t, exposeSpec.Service, unmarshaledExpose.Service)
	assert.Equal(t, exposeSpec.Global, unmarshaledExpose.Global)
	assert.NotNil(t, unmarshaledExpose.HTTPOptions)
	assert.Equal(t, exposeSpec.HTTPOptions.MaxBodySize, unmarshaledExpose.HTTPOptions.MaxBodySize)
	assert.Equal(t, exposeSpec.HTTPOptions.Hosts, unmarshaledExpose.HTTPOptions.Hosts)
}

func TestManifestList(t *testing.T) {
	manifestList := &ManifestList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "depinsubnet.com/v1",
			Kind:       "ManifestList",
		},
		Items: []Manifest{
			*ExampleManifest(),
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(manifestList)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Test JSON unmarshaling
	var unmarshaledList ManifestList
	err = json.Unmarshal(jsonData, &unmarshaledList)
	require.NoError(t, err)

	// Verify the data is preserved
	assert.Len(t, unmarshaledList.Items, 1)
	assert.Equal(t, manifestList.Items[0].Spec.Lease.Owner, unmarshaledList.Items[0].Spec.Lease.Owner)
}

func TestServiceSpecStructure(t *testing.T) {
	// Test that ServiceSpec has the correct structure with volumes, count, expose outside resources
	service := ServiceSpec{
		Name:  "test-service",
		Image: "nginx:latest",
		Resources: ResourceSpec{
			CPU: &CPUResource{Units: 1000},
		},
		Volumes: []VolumeSpec{
			{Name: "data", Mount: "/data", ReadOnly: false},
		},
		Count: 2,
		Expose: []ExposeSpec{
			{IP: "0.0.0.0", Port: 80, Proto: "tcp", Service: "test-service"},
		},
	}

	// Verify structure
	assert.Equal(t, "test-service", service.Name)
	assert.Equal(t, "nginx:latest", service.Image)
	assert.NotNil(t, service.Resources.CPU)
	assert.Len(t, service.Volumes, 1)
	assert.Equal(t, uint32(2), service.Count)
	assert.Len(t, service.Expose, 1)
}
