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
				Credentials: map[string]CredentialsSpec{
					"docker-registry": {
						Host:     "docker.io",
						Username: "username",
						Password: "password",
					},
				},
				Volumes: map[string]VolumeSpec{
					"data-volume": {
						Size:  "10Gi",
						Class: "ssd",
						Attributes: map[string]interface{}{
							"type":        "SSD",
							"performance": "high",
						},
					},
				},
				Services: map[string]ServiceSpec{
					"test-service": {
						Image:      "nginx:latest",
						Credential: "docker-registry",
						Resources: ResourceSpec{
							CPU: &CPUResource{
								Units: "1000",
								Attributes: map[string]interface{}{
									"arch":   "x86_64",
									"vendor": "intel",
									"cores":  4,
								},
							},
							Memory: &MemoryResource{
								Size: "512Mi",
								Attributes: map[string]interface{}{
									"type":  "RAM",
									"speed": "3200MHz",
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

	// Access service by key
	testService := manifest.Spec.Group.Services["test-service"]
	unmarshaledTestService := unmarshaledManifest.Spec.Group.Services["test-service"]
	assert.Equal(t, testService.Count, unmarshaledTestService.Count)
	assert.Equal(t, len(testService.Volumes), len(unmarshaledTestService.Volumes))
	assert.Equal(t, len(testService.Expose), len(unmarshaledTestService.Expose))
}

func TestManifestValidation(t *testing.T) {
	// Test valid manifest
	validManifest := createValidManifest()
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
				Services: map[string]ServiceSpec{},
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
				Services: map[string]ServiceSpec{},
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
				Services: map[string]ServiceSpec{}, // No services
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
				Services: map[string]ServiceSpec{
					"test-service": {
						Image: "nginx:latest",
						Count: 0, // Invalid count
					},
				},
			},
		},
	}
	err = ValidateManifest(invalidManifest4)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service test-service count must be greater than 0")
}

func TestResourceSpecs(t *testing.T) {
	// Test CPU resource
	cpuResource := &CPUResource{
		Units: "2000",
		Attributes: map[string]interface{}{
			"arch":    "x86_64",
			"vendor":  "intel",
			"cores":   8,
			"threads": 16,
		},
	}
	assert.Equal(t, "2000", cpuResource.Units)
	assert.Len(t, cpuResource.Attributes, 4)
	assert.Equal(t, "x86_64", cpuResource.Attributes["arch"])
	assert.Equal(t, 8, cpuResource.Attributes["cores"])

	// Test Memory resource
	memoryResource := &MemoryResource{
		Size: "1024Mi",
		Attributes: map[string]interface{}{
			"type":     "RAM",
			"speed":    "3200MHz",
			"ecc":      true,
			"channels": 2,
		},
	}
	assert.Equal(t, "1024Mi", memoryResource.Size)
	assert.Len(t, memoryResource.Attributes, 4)
	assert.Equal(t, "RAM", memoryResource.Attributes["type"])
	assert.Equal(t, true, memoryResource.Attributes["ecc"])

	// Test GPU resource
	gpuResource := &GPUResource{
		Units: "1",
		Attributes: map[string]interface{}{
			"vendor":       "nvidia",
			"model":        "RTX 3080",
			"memory":       "10GB",
			"cuda":         true,
			"tensor_cores": 272,
		},
	}
	assert.Equal(t, "1", gpuResource.Units)
	assert.Len(t, gpuResource.Attributes, 5)
	assert.Equal(t, "nvidia", gpuResource.Attributes["vendor"])
	assert.Equal(t, 272, gpuResource.Attributes["tensor_cores"])
}

func TestVolumeSpec(t *testing.T) {
	// Test VolumeSpec with new attributes structure
	volumeSpec := VolumeSpec{
		Name:  "test-volume",
		Size:  "20Gi",
		Class: "ssd",
		Attributes: map[string]interface{}{
			"type":        "persistent",
			"performance": "high",
			"encrypted":   true,
			"iops":        3000,
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(volumeSpec)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Test JSON unmarshaling
	var unmarshaledVolume VolumeSpec
	err = json.Unmarshal(jsonData, &unmarshaledVolume)
	require.NoError(t, err)

	// Verify the data is preserved
	assert.Equal(t, volumeSpec.Name, unmarshaledVolume.Name)
	assert.Equal(t, volumeSpec.Size, unmarshaledVolume.Size)
	assert.Equal(t, volumeSpec.Class, unmarshaledVolume.Class)
	assert.Equal(t, volumeSpec.Attributes["type"], unmarshaledVolume.Attributes["type"])
	assert.Equal(t, volumeSpec.Attributes["performance"], unmarshaledVolume.Attributes["performance"])
	assert.Equal(t, volumeSpec.Attributes["encrypted"], unmarshaledVolume.Attributes["encrypted"])
	// JSON unmarshaling converts numbers to float64, so we need to compare as float64
	assert.Equal(t, float64(3000), unmarshaledVolume.Attributes["iops"])
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
		},
		Hosts: []string{"example.com", "www.example.com"},
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
	assert.Equal(t, exposeSpec.Hosts, unmarshaledExpose.Hosts)
}

func TestManifestList(t *testing.T) {
	manifestList := &ManifestList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "depinsubnet.com/v1",
			Kind:       "ManifestList",
		},
		Items: []Manifest{
			*createValidManifest(),
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
	// Test that ServiceSpec has the correct structure with volumes, count, expose as direct fields
	service := ServiceSpec{
		Image:      "nginx:latest",
		Credential: "docker-registry",
		Resources: ResourceSpec{
			CPU: &CPUResource{Units: "1000"},
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
	assert.Equal(t, "nginx:latest", service.Image)
	assert.Equal(t, "docker-registry", service.Credential)
	assert.NotNil(t, service.Resources.CPU)
	assert.Len(t, service.Volumes, 1)
	assert.Equal(t, uint32(2), service.Count)
	assert.Len(t, service.Expose, 1)
}

// createValidManifest creates a valid manifest for testing
func createValidManifest() *Manifest {
	return &Manifest{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "depinsubnet.com/v1",
			Kind:       "Manifest",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "valid-manifest",
			Namespace: "default",
		},
		Spec: ManifestSpec{
			Lease: LeaseSpec{
				Owner: "0x1234567890abcdef",
				ID:    12345,
			},
			Group: GroupSpec{
				Name: "valid-group",
				Services: map[string]ServiceSpec{
					"valid-service": {
						Image: "nginx:latest",
						Count: 1,
					},
				},
			},
		},
	}
}
