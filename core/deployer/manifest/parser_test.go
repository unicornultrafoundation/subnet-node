package manifest

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	// Create test cases
	testCases := []struct {
		name     string
		filePath string
		validate func(t *testing.T, sdl *SDL)
	}{
		{
			name:     "Basic YAML parsing",
			filePath: "examples/example.yaml",
			validate: validateBasicSDL,
		},
		{
			name:     "Basic JSON parsing",
			filePath: "examples/example.json",
			validate: validateBasicSDL,
		},
		{
			name:     "Complex YAML configuration",
			filePath: "examples/complex.yaml",
			validate: validateComplexSDL,
		},
		{
			name:     "Complex JSON configuration",
			filePath: "examples/complex.json",
			validate: validateComplexSDL,
		},
		{
			name:     "Minimal YAML configuration",
			filePath: "examples/minimal.yaml",
			validate: validateMinimalSDL,
		},
		{
			name:     "Minimal JSON configuration",
			filePath: "examples/minimal.json",
			validate: validateMinimalSDL,
		},
		{
			name:     "Network YAML configuration",
			filePath: "examples/network.yaml",
			validate: validateNetworkSDL,
		},
		{
			name:     "Network JSON configuration",
			filePath: "examples/network.json",
			validate: validateNetworkSDL,
		},
		{
			name:     "Echo JSON configuration",
			filePath: "examples/echo.json",
			validate: validateEchoSDL,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create parser
			parser := NewParser()

			// Get absolute path to test file
			absPath, err := filepath.Abs(tc.filePath)
			require.NoError(t, err)

			// Parse and validate SDL
			sdl, err := parser.ValidateAndParse(absPath)
			require.NoError(t, err)
			require.NotNil(t, sdl)

			// Run specific validation
			tc.validate(t, sdl)
		})
	}
}

func validateBasicSDL(t *testing.T, sdl *SDL) {
	// Test version
	assert.Equal(t, "2.0", sdl.VersionStr)

	// Test services
	require.Contains(t, sdl.Services, "web")
	webService := sdl.Services["web"]
	assert.Equal(t, "nginx:latest", webService.Image)
	assert.Equal(t, []string{"nginx", "-g", "daemon off;"}, webService.Command)
	assert.Equal(t, []string{"NGINX_HOST=example.com"}, webService.Env)
	assert.Equal(t, int32(2), webService.Count)

	// Test expose configuration
	require.Len(t, webService.Expose, 1)
	expose := webService.Expose[0]
	assert.Equal(t, int32(5678), expose.Port)
	assert.Equal(t, int32(80), expose.As)
	assert.Equal(t, "TCP", expose.Proto)
	require.Len(t, expose.To, 1)
	assert.True(t, expose.To[0].Global)

	// Test health check configuration
	require.NotNil(t, webService.Params)
	require.NotNil(t, webService.Params.Health)
	require.NotNil(t, webService.Params.Health.Readiness)
	readiness := webService.Params.Health.Readiness
	assert.Equal(t, int32(5), readiness.InitialDelaySeconds)
	assert.Equal(t, int32(10), readiness.PeriodSeconds)
	assert.Equal(t, int32(5), readiness.TimeoutSeconds)
	assert.Equal(t, int32(1), readiness.SuccessThreshold)
	assert.Equal(t, int32(3), readiness.FailureThreshold)
	require.NotNil(t, readiness.HTTP)
	assert.Equal(t, "/", readiness.HTTP.Path)
	assert.Equal(t, int32(5678), readiness.HTTP.Port)

	// Test liveness check
	require.NotNil(t, webService.Params.Health.Liveness)
	liveness := webService.Params.Health.Liveness
	assert.Equal(t, int32(15), liveness.InitialDelaySeconds)
	assert.Equal(t, int32(10), liveness.PeriodSeconds)
	assert.Equal(t, int32(5), liveness.TimeoutSeconds)
	assert.Equal(t, int32(1), liveness.SuccessThreshold)
	assert.Equal(t, int32(3), liveness.FailureThreshold)
	require.NotNil(t, liveness.HTTP)
	assert.Equal(t, "/", liveness.HTTP.Path)
	assert.Equal(t, int32(5678), liveness.HTTP.Port)

	// Test resources
	require.NotNil(t, webService.Resources)
	require.NotNil(t, webService.Resources.CPU)
	assert.Equal(t, int64(500), webService.Resources.CPU.Units.Value)
	assert.Equal(t, "m", webService.Resources.CPU.Units.Unit)
	require.NotNil(t, webService.Resources.Memory)
	assert.Equal(t, int64(512), webService.Resources.Memory.Size.Value)
	assert.Equal(t, "Mi", webService.Resources.Memory.Size.Unit)

	// Test profiles
	require.Contains(t, sdl.Profiles.Compute, "default")
	computeProfile := sdl.Profiles.Compute["default"]
	assert.Equal(t, "500m", computeProfile.Resources.CPU.Request)
	assert.Equal(t, "1000m", computeProfile.Resources.CPU.Limit)
	assert.Equal(t, "512Mi", computeProfile.Resources.Memory.Request)
	assert.Equal(t, "1Gi", computeProfile.Resources.Memory.Limit)

	// Test storage configuration
	require.Len(t, computeProfile.Resources.Storage, 1)
	storage := computeProfile.Resources.Storage[0]
	assert.Equal(t, "data", storage.Name)
	assert.Equal(t, "10Gi", storage.Size)
	require.NotNil(t, storage.Attributes)
	assert.True(t, storage.Attributes.Persistent)
	assert.Equal(t, "standard", storage.Attributes.Class)

	// Test deployment
	require.Contains(t, sdl.Deployment, "default")
	deployment := sdl.Deployment["default"]
	assert.Equal(t, "default", deployment.Profile)
	assert.Equal(t, int32(2), deployment.Count)

	// Test endpoints
	require.Contains(t, sdl.Endpoints, "web")
	assert.Equal(t, "http", sdl.Endpoints["web"].Kind)
}

func validateComplexSDL(t *testing.T, sdl *SDL) {
	// Test multiple services
	require.Contains(t, sdl.Services, "web")
	require.Contains(t, sdl.Services, "api")
	require.Contains(t, sdl.Services, "db")
	require.Contains(t, sdl.Services, "cache")
	require.Contains(t, sdl.Services, "ml-service")

	// Test service dependencies
	webService := sdl.Services["web"]
	require.Contains(t, webService.DependsOn, "db")
	require.Contains(t, webService.DependsOn, "cache")

	// Test multiple expose ports
	require.Len(t, webService.Expose, 2)
	assert.Equal(t, int32(80), webService.Expose[0].Port)
	assert.Equal(t, int32(443), webService.Expose[1].Port)

	// Test HTTP options
	require.NotNil(t, webService.Expose[0].HTTPOptions)
	assert.Equal(t, int64(10485760), webService.Expose[0].HTTPOptions.MaxBodySize)
	assert.Equal(t, []string{"api"}, webService.Expose[0].HTTPOptions.NextCases)

	// Test credentials
	require.NotNil(t, webService.Credentials)
	assert.Equal(t, "docker.io", webService.Credentials.Host)
	assert.Equal(t, "user", webService.Credentials.Username)
	assert.Equal(t, "pass", webService.Credentials.Password)

	// Test image pull secrets
	require.Len(t, webService.ImagePullSecrets, 1)
	assert.Equal(t, "docker-registry-secret", webService.ImagePullSecrets[0].Name)

	// Test GPU resources
	mlService := sdl.Services["ml-service"]
	require.NotNil(t, mlService.Resources.GPU)
	assert.Equal(t, int32(1), mlService.Resources.GPU.Units)
	require.Contains(t, mlService.Resources.GPU.Attributes.Vendor, "nvidia")
	nvidiaGPU := mlService.Resources.GPU.Attributes.Vendor["nvidia"][0]
	assert.Equal(t, "A100", nvidiaGPU.Model)
	assert.Equal(t, "40Gi", nvidiaGPU.RAM)
	assert.Equal(t, "PCIe", nvidiaGPU.Interface)

	// Test multiple compute profiles
	require.Contains(t, sdl.Profiles.Compute, "default")
	require.Contains(t, sdl.Profiles.Compute, "high-memory")
	require.Contains(t, sdl.Profiles.Compute, "gpu")

	// Test multiple deployments
	require.Contains(t, sdl.Deployment, "web")
	require.Contains(t, sdl.Deployment, "api")
	require.Contains(t, sdl.Deployment, "db")
	require.Contains(t, sdl.Deployment, "cache")
	require.Contains(t, sdl.Deployment, "ml")
}

func validateMinimalSDL(t *testing.T, sdl *SDL) {
	// Test minimal service
	require.Contains(t, sdl.Services, "minimal")
	minimalService := sdl.Services["minimal"]
	assert.Equal(t, "busybox:latest", minimalService.Image)
	assert.Equal(t, int32(1), minimalService.Count)
	assert.Nil(t, minimalService.Command)
	assert.Nil(t, minimalService.Env)
	assert.Nil(t, minimalService.Expose)
	assert.Nil(t, minimalService.Params)

	// Test minimal compute profile
	require.Contains(t, sdl.Profiles.Compute, "default")
	computeProfile := sdl.Profiles.Compute["default"]
	assert.Equal(t, "100m", computeProfile.Resources.CPU.Request)
	assert.Equal(t, "200m", computeProfile.Resources.CPU.Limit)
	assert.Equal(t, "128Mi", computeProfile.Resources.Memory.Request)
	assert.Equal(t, "256Mi", computeProfile.Resources.Memory.Limit)
	assert.Nil(t, computeProfile.Resources.Storage)

	// Test minimal deployment
	require.Contains(t, sdl.Deployment, "default")
	deployment := sdl.Deployment["default"]
	assert.Equal(t, "default", deployment.Profile)
	assert.Equal(t, int32(1), deployment.Count)
}

func validateNetworkSDL(t *testing.T, sdl *SDL) {
	// Test network service
	require.Contains(t, sdl.Services, "network-service")
	networkService := sdl.Services["network-service"]
	require.NotNil(t, networkService.Resources.Network)
	assert.Equal(t, "100M", networkService.Resources.Network.Bandwidth)

	// Test network expose configuration
	require.Len(t, networkService.Expose, 2)
	assert.Equal(t, int32(80), networkService.Expose[0].Port)
	assert.Equal(t, int32(443), networkService.Expose[1].Port)
	require.Len(t, networkService.Expose[0].Accept, 3)
	assert.Equal(t, "10.0.0.0/8", networkService.Expose[0].Accept[0])
	assert.Equal(t, "0.0.0.0/0", networkService.Expose[1].Accept[0])

	// Test network compute profile
	require.Contains(t, sdl.Profiles.Compute, "default")
	computeProfile := sdl.Profiles.Compute["default"]
	require.NotNil(t, computeProfile.Resources.Network)
	assert.Equal(t, "100M", computeProfile.Resources.Network.Bandwidth)
}

func validateEchoSDL(t *testing.T, sdl *SDL) {
	// Test version
	assert.Equal(t, "2.0", sdl.VersionStr)

	// Test web service
	require.Contains(t, sdl.Services, "web")
	webService := sdl.Services["web"]
	assert.Equal(t, "hashicorp/http-echo", webService.Image)
	assert.Equal(t, []string{"/http-echo", "-text", "Hello from http-echo!"}, webService.Command)
	assert.Equal(t, int32(1), webService.Count)

	// Test expose configuration
	require.Len(t, webService.Expose, 1)
	expose := webService.Expose[0]
	assert.Equal(t, int32(5678), expose.Port)
	assert.Equal(t, int32(80), expose.As)
	assert.Equal(t, "TCP", expose.Proto)
	require.Len(t, expose.To, 1)
	assert.True(t, expose.To[0].Global)

	// Test health check configuration
	require.NotNil(t, webService.Params)
	require.NotNil(t, webService.Params.Health)
	require.NotNil(t, webService.Params.Health.Readiness)
	readiness := webService.Params.Health.Readiness
	assert.Equal(t, int32(5), readiness.InitialDelaySeconds)
	assert.Equal(t, int32(10), readiness.PeriodSeconds)
	assert.Equal(t, int32(5), readiness.TimeoutSeconds)
	assert.Equal(t, int32(1), readiness.SuccessThreshold)
	assert.Equal(t, int32(3), readiness.FailureThreshold)
	require.NotNil(t, readiness.HTTP)
	assert.Equal(t, "/", readiness.HTTP.Path)
	assert.Equal(t, int32(5678), readiness.HTTP.Port)

	// Test liveness check
	require.NotNil(t, webService.Params.Health.Liveness)
	liveness := webService.Params.Health.Liveness
	assert.Equal(t, int32(15), liveness.InitialDelaySeconds)
	assert.Equal(t, int32(10), liveness.PeriodSeconds)
	assert.Equal(t, int32(5), liveness.TimeoutSeconds)
	assert.Equal(t, int32(1), liveness.SuccessThreshold)
	assert.Equal(t, int32(3), liveness.FailureThreshold)
	require.NotNil(t, liveness.HTTP)
	assert.Equal(t, "/", liveness.HTTP.Path)
	assert.Equal(t, int32(5678), liveness.HTTP.Port)

	// Test resources
	require.NotNil(t, webService.Resources)
	require.NotNil(t, webService.Resources.CPU)
	assert.Equal(t, int64(100), webService.Resources.CPU.Units.Value)
	assert.Equal(t, "m", webService.Resources.CPU.Units.Unit)
	require.NotNil(t, webService.Resources.Memory)
	assert.Equal(t, int64(128), webService.Resources.Memory.Size.Value)
	assert.Equal(t, "Mi", webService.Resources.Memory.Size.Unit)

	// Test compute profile
	require.Contains(t, sdl.Profiles.Compute, "default")
	computeProfile := sdl.Profiles.Compute["default"]
	assert.Equal(t, "100m", computeProfile.Resources.CPU.Request)
	assert.Equal(t, "200m", computeProfile.Resources.CPU.Limit)
	assert.Equal(t, "128Mi", computeProfile.Resources.Memory.Request)
	assert.Equal(t, "256Mi", computeProfile.Resources.Memory.Limit)

	// Test storage configuration
	require.Len(t, computeProfile.Resources.Storage, 1)
	storage := computeProfile.Resources.Storage[0]
	assert.Equal(t, "data", storage.Name)
	assert.Equal(t, "1Gi", storage.Size)
	require.NotNil(t, storage.Attributes)
	assert.False(t, storage.Attributes.Persistent)

	// Test deployment
	require.Contains(t, sdl.Deployment, "default")
	deployment := sdl.Deployment["default"]
	assert.Equal(t, "default", deployment.Profile)
	assert.Equal(t, int32(1), deployment.Count)

	// Test endpoints
	require.Contains(t, sdl.Endpoints, "echo")
	assert.Equal(t, "http", sdl.Endpoints["echo"].Kind)
}

func TestParserErrors(t *testing.T) {
	parser := NewParser()

	// Test non-existent file
	_, err := parser.ParseFile("non_existent.yaml")
	assert.Error(t, err, "Expected error for non-existent file")

	// Test invalid YAML
	invalidYAML := []byte("invalid: yaml: content: [")
	_, err = parser.ParseYAML(invalidYAML)
	assert.Error(t, err, "Expected error for invalid YAML")

	// Test invalid JSON
	invalidJSON := []byte("invalid json content")
	_, err = parser.ParseJSON(invalidJSON)
	assert.Error(t, err, "Expected error for invalid JSON")

	// Test missing required fields
	missingFields := []byte(`
version: "2.0"
services: {}
profiles:
  compute: {}
deployment: {}
`)
	_, err = parser.ParseYAML(missingFields)
	assert.Error(t, err, "Expected error for missing required fields")

	// Test invalid service name
	invalidServiceName := []byte(`
version: "2.0"
services:
  Invalid_Name:
    image: nginx:latest
    count: 1
profiles:
  compute:
    default:
      resources:
        cpu:
          request: "100m"
          limit: "200m"
        memory:
          request: "128Mi"
          limit: "256Mi"
deployment:
  default:
    profile: default
    count: 1
`)
	_, err = parser.ParseYAML(invalidServiceName)
	assert.Error(t, err, "Expected error for invalid service name")

	// Test invalid resource values
	invalidResources := []byte(`
version: "2.0"
services:
  test:
    image: nginx:latest
    count: 1
    resources:
      cpu:
        units:
          value: -1
          unit: m
profiles:
  compute:
    default:
      resources:
        cpu:
          request: "100m"
          limit: "200m"
        memory:
          request: "128Mi"
          limit: "256Mi"
deployment:
  default:
    profile: default
    count: 1
`)
	_, err = parser.ParseYAML(invalidResources)
	assert.Error(t, err, "Expected error for invalid resource values")

	// Test invalid health check configuration
	invalidHealthCheck := []byte(`
version: "2.0"
services:
  test:
    image: nginx:latest
    count: 1
    params:
      health:
        readiness:
          http:
            path: ""
            port: -1
profiles:
  compute:
    default:
      resources:
        cpu:
          request: "100m"
          limit: "200m"
        memory:
          request: "128Mi"
          limit: "256Mi"
deployment:
  default:
    profile: default
    count: 1
`)
	_, err = parser.ParseYAML(invalidHealthCheck)
	assert.Error(t, err, "Expected error for invalid health check configuration")
}

func TestResourceConversion(t *testing.T) {
	// Create a test resource requirements
	req := &ResourceRequirements{
		CPU: ResourceLimit{
			Request: "500m",
			Limit:   "1000m",
		},
		Memory: ResourceLimit{
			Request: "512Mi",
			Limit:   "1Gi",
		},
		Storage: []StorageVolume{
			{
				Name: "data",
				Size: "10Gi",
				Attributes: &StorageAttributes{
					Persistent: true,
					Class:      "standard",
				},
			},
		},
		GPU: &GPUResource{
			Units: 1,
			Attributes: GPUAttributes{
				Vendor: map[string][]GPUModel{
					"nvidia": {
						{
							Model:     "A100",
							RAM:       "40Gi",
							Interface: "PCIe",
						},
					},
				},
			},
		},
		Network: &NetworkResource{
			Bandwidth: "100M",
		},
	}

	// Test conversion to K8s resources
	resources, err := req.ToK8sResources()
	require.NoError(t, err)
	assert.NotEmpty(t, resources)
	assert.Contains(t, resources, "cpu")
	assert.Contains(t, resources, "memory")
	assert.Contains(t, resources, "storage")
	assert.Contains(t, resources, "nvidia.com/gpu")
}
