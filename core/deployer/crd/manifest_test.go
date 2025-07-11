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

func TestParseManifestFromJSON(t *testing.T) {
	// Test JSON string that represents a complete manifest
	jsonStr := `{
		"apiVersion": "depinsubnet.com/v1",
		"kind": "Manifest",
		"metadata": {
			"name": "test-manifest",
			"namespace": "default"
		},
		"spec": {
			"lease": {
				"owner": "0x1234567890abcdef",
				"id": 12345
			},
			"group": {
				"name": "test-group",
				"credentials": {
					"docker-registry": {
						"host": "docker.io",
						"email": "user@example.com",
						"username": "username",
						"password": "password"
					}
				},
				"volumes": {
					"data-volume": {
						"size": "10Gi",
						"class": "ssd",
						"attributes": {
							"type": "persistent",
							"performance": "high",
							"encrypted": true,
							"iops": 3000
						}
					}
				},
				"services": {
					"web-service": {
						"image": "nginx:latest",
						"command": ["nginx", "-g", "daemon off;"],
						"args": ["--config", "/etc/nginx/nginx.conf"],
						"env": ["NODE_ENV=production"],
						"credential": "docker-registry",
						"resources": {
							"cpu": {
								"units": "1000",
								"attributes": {
									"arch": "x86_64",
									"vendor": "intel",
									"cores": 8,
									"threads": 16
								}
							},
							"memory": {
								"size": "512Mi",
								"attributes": {
									"type": "DDR4",
									"speed": "3200MHz",
									"ecc": true,
									"channels": 2
								}
							},
							"gpu": {
								"units": "1",
								"attributes": {
									"model": "RTX 3080",
									"vendor": "nvidia",
									"memory": "10GB",
									"cuda": true,
									"tensor_cores": 272
								}
							}
						},
						"volumes": [
							{
								"name": "data-volume",
								"mount": "/data",
								"read_only": false
							}
						],
						"count": 3,
						"expose": [
							{
								"ip": "0.0.0.0",
								"port": 80,
								"proto": "tcp",
								"service": "web-service",
								"global": true,
								"http_options": {
									"max_body_size": 1048576,
									"read_timeout": 30,
									"send_timeout": 30,
									"next_tries": 3,
									"next_timeout": 10,
									"next_cases": ["error", "timeout"]
								},
								"hosts": ["example.com", "www.example.com"]
							}
						]
					}
				}
			}
		}
	}`

	// Parse JSON into manifest
	var manifest Manifest
	err := json.Unmarshal([]byte(jsonStr), &manifest)
	require.NoError(t, err)

	// Verify basic structure
	assert.Equal(t, "depinsubnet.com/v1", manifest.APIVersion)
	assert.Equal(t, "Manifest", manifest.Kind)
	assert.Equal(t, "test-manifest", manifest.Name)
	assert.Equal(t, "default", manifest.Namespace)

	// Verify lease
	assert.Equal(t, "0x1234567890abcdef", manifest.Spec.Lease.Owner)
	assert.Equal(t, uint64(12345), manifest.Spec.Lease.ID)

	// Verify group
	assert.Equal(t, "test-group", manifest.Spec.Group.Name)

	// Verify credentials
	assert.Len(t, manifest.Spec.Group.Credentials, 1)
	cred := manifest.Spec.Group.Credentials["docker-registry"]
	assert.Equal(t, "docker.io", cred.Host)
	assert.Equal(t, "user@example.com", cred.Email)
	assert.Equal(t, "username", cred.Username)
	assert.Equal(t, "password", cred.Password)

	// Verify volumes
	assert.Len(t, manifest.Spec.Group.Volumes, 1)
	vol := manifest.Spec.Group.Volumes["data-volume"]
	assert.Equal(t, "10Gi", vol.Size)
	assert.Equal(t, "ssd", vol.Class)
	assert.Equal(t, "persistent", vol.Attributes["type"])
	assert.Equal(t, "high", vol.Attributes["performance"])
	assert.Equal(t, true, vol.Attributes["encrypted"])
	assert.Equal(t, float64(3000), vol.Attributes["iops"]) // JSON numbers become float64

	// Verify services
	assert.Len(t, manifest.Spec.Group.Services, 1)
	service := manifest.Spec.Group.Services["web-service"]
	assert.Equal(t, "nginx:latest", service.Image)
	assert.Equal(t, []string{"nginx", "-g", "daemon off;"}, service.Command)
	assert.Equal(t, []string{"--config", "/etc/nginx/nginx.conf"}, service.Args)
	assert.Equal(t, []string{"NODE_ENV=production"}, service.Env)
	assert.Equal(t, "docker-registry", service.Credential)

	// Verify resources
	assert.NotNil(t, service.Resources.CPU)
	assert.Equal(t, "1000", service.Resources.CPU.Units)
	assert.Equal(t, "x86_64", service.Resources.CPU.Attributes["arch"])
	assert.Equal(t, "intel", service.Resources.CPU.Attributes["vendor"])
	assert.Equal(t, float64(8), service.Resources.CPU.Attributes["cores"])
	assert.Equal(t, float64(16), service.Resources.CPU.Attributes["threads"])

	assert.NotNil(t, service.Resources.Memory)
	assert.Equal(t, "512Mi", service.Resources.Memory.Size)
	assert.Equal(t, "DDR4", service.Resources.Memory.Attributes["type"])
	assert.Equal(t, "3200MHz", service.Resources.Memory.Attributes["speed"])
	assert.Equal(t, true, service.Resources.Memory.Attributes["ecc"])
	assert.Equal(t, float64(2), service.Resources.Memory.Attributes["channels"])

	assert.NotNil(t, service.Resources.GPU)
	assert.Equal(t, "1", service.Resources.GPU.Units)
	assert.Equal(t, "RTX 3080", service.Resources.GPU.Attributes["model"])
	assert.Equal(t, "nvidia", service.Resources.GPU.Attributes["vendor"])
	assert.Equal(t, "10GB", service.Resources.GPU.Attributes["memory"])
	assert.Equal(t, true, service.Resources.GPU.Attributes["cuda"])
	assert.Equal(t, float64(272), service.Resources.GPU.Attributes["tensor_cores"])

	// Verify service volumes
	assert.Len(t, service.Volumes, 1)
	serviceVol := service.Volumes[0]
	assert.Equal(t, "data-volume", serviceVol.Name)
	assert.Equal(t, "/data", serviceVol.Mount)
	assert.Equal(t, false, serviceVol.ReadOnly)

	// Verify count
	assert.Equal(t, uint32(3), service.Count)

	// Verify expose
	assert.Len(t, service.Expose, 1)
	expose := service.Expose[0]
	assert.Equal(t, "0.0.0.0", expose.IP)
	assert.Equal(t, uint16(80), expose.Port)
	assert.Equal(t, "tcp", expose.Proto)
	assert.Equal(t, "web-service", expose.Service)
	assert.Equal(t, true, expose.Global)
	assert.Equal(t, []string{"example.com", "www.example.com"}, expose.Hosts)

	// Verify HTTP options
	assert.NotNil(t, expose.HTTPOptions)
	assert.Equal(t, 1048576, expose.HTTPOptions.MaxBodySize)
	assert.Equal(t, 30, expose.HTTPOptions.ReadTimeout)
	assert.Equal(t, 30, expose.HTTPOptions.SendTimeout)
	assert.Equal(t, 3, expose.HTTPOptions.NextTries)
	assert.Equal(t, 10, expose.HTTPOptions.NextTimeout)
	assert.Equal(t, []string{"error", "timeout"}, expose.HTTPOptions.NextCases)
}

func TestParseMinimalManifestFromJSON(t *testing.T) {
	// Test minimal JSON string with only required fields
	jsonStr := `{
		"apiVersion": "depinsubnet.com/v1",
		"kind": "Manifest",
		"metadata": {
			"name": "minimal-manifest"
		},
		"spec": {
			"lease": {
				"owner": "0x1234567890abcdef",
				"id": 12345
			},
			"group": {
				"name": "minimal-group",
				"services": {
					"minimal-service": {
						"image": "nginx:latest",
						"count": 1
					}
				}
			}
		}
	}`

	// Parse JSON into manifest
	var manifest Manifest
	err := json.Unmarshal([]byte(jsonStr), &manifest)
	require.NoError(t, err)

	// Verify basic structure
	assert.Equal(t, "depinsubnet.com/v1", manifest.APIVersion)
	assert.Equal(t, "Manifest", manifest.Kind)
	assert.Equal(t, "minimal-manifest", manifest.Name)

	// Verify lease
	assert.Equal(t, "0x1234567890abcdef", manifest.Spec.Lease.Owner)
	assert.Equal(t, uint64(12345), manifest.Spec.Lease.ID)

	// Verify group
	assert.Equal(t, "minimal-group", manifest.Spec.Group.Name)

	// Verify services
	assert.Len(t, manifest.Spec.Group.Services, 1)
	service := manifest.Spec.Group.Services["minimal-service"]
	assert.Equal(t, "nginx:latest", service.Image)
	assert.Equal(t, uint32(1), service.Count)

	// Verify optional fields are empty/nil
	assert.Empty(t, manifest.Spec.Group.Credentials)
	assert.Empty(t, manifest.Spec.Group.Volumes)
	assert.Nil(t, service.Resources.CPU)
	assert.Nil(t, service.Resources.Memory)
	assert.Nil(t, service.Resources.GPU)
	assert.Empty(t, service.Volumes)
	assert.Empty(t, service.Expose)
}
