package sdl

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestV1SDL(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    v1SDL
		wantErr bool
	}{
		{
			name: "valid SDL with single service",
			yaml: `
services:
  web:
    image: nginx:latest
    command: ["nginx", "-g", "daemon off;"]
    count: 2
`,
			want: v1SDL{
				Services: map[string]v1Service{
					"web": {
						Image:   "nginx:latest",
						Command: []string{"nginx", "-g", "daemon off;"},
						Count:   2,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid SDL with multiple services",
			yaml: `
services:
  web:
    image: nginx:latest
    count: 1
  api:
    image: node:16
    command: ["node", "app.js"]
    count: 3
`,
			want: v1SDL{
				Services: map[string]v1Service{
					"web": {
						Image: "nginx:latest",
						Count: 1,
					},
					"api": {
						Image:   "node:16",
						Command: []string{"node", "app.js"},
						Count:   3,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty SDL",
			yaml: `services: {}`,
			want: v1SDL{
				Services: map[string]v1Service{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got v1SDL
			err := yaml.Unmarshal([]byte(tt.yaml), &got)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestV1Service(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    v1Service
		wantErr bool
	}{
		{
			name: "complete service definition",
			yaml: `
image: nginx:latest
command: ["nginx", "-g", "daemon off;"]
args: ["--config", "/etc/nginx/nginx.conf"]
env:
  - NGINX_PORT=80
  - NGINX_HOST=0.0.0.0
volumes:
  - name: config
    mount: /etc/nginx
resources:
  cpu:
    units: 0.5
  memory:
    size: 512Mi
  gpu:
    units: 1
expose:
  - port: 80
    as: 80
    proto: tcp
count: 2
credentials:
  host: registry.example.com
  username: user
  password: pass
`,
			want: v1Service{
				Image:   "nginx:latest",
				Command: []string{"nginx", "-g", "daemon off;"},
				Args:    []string{"--config", "/etc/nginx/nginx.conf"},
				Env:     []string{"NGINX_PORT=80", "NGINX_HOST=0.0.0.0"},
				Volumes: v1ServiceVolumes{
					{
						Name:     "config",
						Mount:    "/etc/nginx",
						ReadOnly: false,
					},
				},
				Resources: v1Resources{
					CPU:    v1Cpu{Units: cpuQuantity(500)},                    // 0.5 * 1000
					Memory: v1Memory{Size: memoryQuantity(512 * 1024 * 1024)}, // 512Mi
					GPU:    v1GPU{Units: gpuQuantity(1)},
				},
				Expose: v1Exposes{
					{
						Port:  80,
						As:    80,
						Proto: "tcp",
					},
				},
				Count: 2,
				Credentials: v1ServiceCredentials{
					Host:     "registry.example.com",
					Username: "user",
					Password: "pass",
				},
			},
			wantErr: false,
		},
		{
			name: "minimal service definition",
			yaml: `
image: alpine:latest
count: 1
`,
			want: v1Service{
				Image: "alpine:latest",
				Count: 1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got v1Service
			err := yaml.Unmarshal([]byte(tt.yaml), &got)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestV1Expose(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    v1Expose
		wantErr bool
	}{
		{
			name: "complete expose definition",
			yaml: `
port: 8080
as: 80
proto: tcp
to:
  - service: web
    global: true
    ip: 10.0.0.1
    http_options:
      max_body_size: 1048576
      read_timeout: 30
      send_timeout: 30
      next_tries: 3
      next_timeout: 10
      next_cases: ["error", "timeout"]
accept: ["10.0.0.0/8", "192.168.0.0/16"]
http_options:
  max_body_size: 524288
  read_timeout: 60
  send_timeout: 60
  next_tries: 2
  next_timeout: 5
  next_cases: ["error"]
`,
			want: v1Expose{
				Port:  8080,
				As:    80,
				Proto: "tcp",
				To: []v1ExposeTo{
					{
						Service: "web",
						Global:  true,
						IP:      "10.0.0.1",
						HTTPOptions: v1HTTPOptions{
							MaxBodySize: 1048576,
							ReadTimeout: 30,
							SendTimeout: 30,
							NextTries:   3,
							NextTimeout: 10,
							NextCases:   []string{"error", "timeout"},
						},
					},
				},
				Accept: v1Accept{
					Items: []string{"10.0.0.0/8", "192.168.0.0/16"},
				},
				HTTPOptions: v1HTTPOptions{
					MaxBodySize: 524288,
					ReadTimeout: 60,
					SendTimeout: 60,
					NextTries:   2,
					NextTimeout: 5,
					NextCases:   []string{"error"},
				},
			},
			wantErr: false,
		},
		{
			name: "minimal expose definition",
			yaml: `
port: 80
as: 80
`,
			want: v1Expose{
				Port: 80,
				As:   80,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got v1Expose
			err := yaml.Unmarshal([]byte(tt.yaml), &got)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestV1Accept(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    v1Accept
		wantErr bool
	}{
		{
			name: "valid accept items",
			yaml: `["10.0.0.0/8", "192.168.0.0/16", "172.16.0.0/12"]`,
			want: v1Accept{
				Items: []string{"10.0.0.0/8", "192.168.0.0/16", "172.16.0.0/12"},
			},
			wantErr: false,
		},
		{
			name: "single valid item",
			yaml: `["0.0.0.0/0"]`,
			want: v1Accept{
				Items: []string{"0.0.0.0/0"},
			},
			wantErr: false,
		},
		{
			name:    "invalid URL format",
			yaml:    `["invalid:url:with:colons"]`,
			want:    v1Accept{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got v1Accept
			err := yaml.Unmarshal([]byte(tt.yaml), &got)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestV1Volume(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    v1Volume
		wantErr bool
	}{
		{
			name: "complete volume definition",
			yaml: `
name: data
mount_path: /app/data
sub_path: logs
read_only: false
size: 10Gi
persistent: true
storage_class: fast-ssd
`,
			want: v1Volume{
				Name:         "data",
				Size:         byteQuantity(10 * 1024 * 1024 * 1024), // 10Gi
				Persistent:   true,
				StorageClass: "fast-ssd",
			},
			wantErr: false,
		},
		{
			name: "read-only volume",
			yaml: `
name: config
mount_path: /etc/app
read_only: true
size: 100Mi
persistent: false
`,
			want: v1Volume{
				Name:       "config",
				Size:       byteQuantity(100 * 1024 * 1024), // 100Mi
				Persistent: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got v1Volume
			err := yaml.Unmarshal([]byte(tt.yaml), &got)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestV1Resources(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    v1Resources
		wantErr bool
	}{
		{
			name: "complete resources definition",
			yaml: `
cpu:
  units: 2.5
memory:
  size: 4Gi
gpu:
  units: 1
`,
			want: v1Resources{
				CPU:    v1Cpu{Units: cpuQuantity(2500)},                        // 2.5 * 1000
				Memory: v1Memory{Size: memoryQuantity(4 * 1024 * 1024 * 1024)}, // 4Gi
				GPU:    v1GPU{Units: gpuQuantity(1)},
			},
			wantErr: false,
		},
		{
			name: "minimal resources",
			yaml: `
cpu:
  units: 0.1
memory:
  size: 128Mi
gpu:
  units: 1
`,
			want: v1Resources{
				CPU:    v1Cpu{Units: cpuQuantity(100)},                    // 0.1 * 1000
				Memory: v1Memory{Size: memoryQuantity(128 * 1024 * 1024)}, // 128Mi
				GPU:    v1GPU{Units: gpuQuantity(1)},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got v1Resources
			err := yaml.Unmarshal([]byte(tt.yaml), &got)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestV1ServiceCredentials(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    v1ServiceCredentials
		wantErr bool
	}{
		{
			name: "complete credentials",
			yaml: `
host: registry.example.com
email: user@example.com
username: myuser
password: mypass
`,
			want: v1ServiceCredentials{
				Host:     "registry.example.com",
				Username: "myuser",
				Password: "mypass",
			},
			wantErr: false,
		},
		{
			name: "partial credentials",
			yaml: `
username: readonly
password: readonly123
`,
			want: v1ServiceCredentials{
				Username: "readonly",
				Password: "readonly123",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got v1ServiceCredentials
			err := yaml.Unmarshal([]byte(tt.yaml), &got)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGPUQuantity(t *testing.T) {
	type vtype struct {
		Val gpuQuantity `yaml:"val"`
	}

	tests := []struct {
		text  string
		value uint64
		err   bool
	}{
		{`val: 1`, 1, false},
		{`val: 0`, 0, false},
		{`val: 8`, 8, false},
		{`val: "1"`, 1, false},
		{`val: "0"`, 0, false},
		{`val: ""`, 0, true},
		{`val: "invalid"`, 0, true},
	}

	for idx, test := range tests {
		buf := []byte(test.text)
		obj := &vtype{}

		err := yaml.Unmarshal(buf, obj)

		if test.err {
			assert.Error(t, err, "idx:%v text:`%v`", idx, test.text)
			continue
		}

		if !assert.NoError(t, err, "idx:%v text:`%v`", idx, test.text) {
			continue
		}

		assert.Equal(t, gpuQuantity(test.value), obj.Val, "idx:%v text:`%v`", idx, test.text)
	}
}

func TestV1GPU(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    v1GPU
		wantErr bool
	}{
		{
			name: "GPU with units only",
			yaml: `
units: 2
`,
			want: v1GPU{
				Units: gpuQuantity(2),
			},
			wantErr: false,
		},
		{
			name: "GPU with units and vendor",
			yaml: `
units: 1
vendor:
  nvidia:
    - model: RTX 4090
      ram: 24Gi
      interface: pcie
`,
			want: func() v1GPU {
				ram := memoryQuantity(24 * 1024 * 1024 * 1024) // 24Gi
				iface := gpuInterface("pcie")
				return v1GPU{
					Units: gpuQuantity(1),
					Vendor: gpuVendor{
						Nvidia: v1GPUsNvidia{
							{
								Model:     "RTX 4090",
								RAM:       &ram,
								Interface: &iface,
							},
						},
					},
				}
			}(),
			wantErr: false,
		},
		{
			name: "GPU with multiple vendor cards",
			yaml: `
units: 4
vendor:
  nvidia:
    - model: RTX 3080
      ram: 10Gi
    - model: RTX 3090
      ram: 24Gi
`,
			want: func() v1GPU {
				ram1 := memoryQuantity(10 * 1024 * 1024 * 1024) // 10Gi
				ram2 := memoryQuantity(24 * 1024 * 1024 * 1024) // 24Gi
				return v1GPU{
					Units: gpuQuantity(4),
					Vendor: gpuVendor{
						Nvidia: v1GPUsNvidia{
							{
								Model: "RTX 3080",
								RAM:   &ram1,
							},
							{
								Model: "RTX 3090",
								RAM:   &ram2,
							},
						},
					},
				}
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got v1GPU
			err := yaml.Unmarshal([]byte(tt.yaml), &got)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemoryQuantity(t *testing.T) {
	type vtype struct {
		Val memoryQuantity `yaml:"val"`
	}

	tests := []struct {
		text  string
		value uint64
		err   bool
	}{
		{`val: 1`, 1, false},
		{`val: -1`, 1, true},
		{`val: "1Mi"`, 1024 * 1024, false},
		{`val: "-1Mi"`, 0, true},
		{`val: "0.5Mi"`, 1024 * 1024 / 2, false},
		{`val: "-0.5Mi"`, 0, true},
		{`val: "3Mi"`, 3 * 1024 * 1024, false},
		{`val: "3Gi"`, 3 * 1024 * 1024 * 1024, false},
		{`val: "3Ti"`, 3 * 1024 * 1024 * 1024 * 1024, false},
		{`val: "3Pi"`, 3 * 1024 * 1024 * 1024 * 1024 * 1024, false},
		{`val: "3Ei"`, 3 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024, false},
		{`val: ""`, 0, true},
	}

	for idx, test := range tests {
		buf := []byte(test.text)
		obj := &vtype{}

		err := yaml.Unmarshal(buf, obj)

		if test.err {
			assert.Error(t, err, "idx:%v text:`%v`", idx, test.text)
			continue
		}

		if !assert.NoError(t, err, "idx:%v text:`%v`", idx, test.text) {
			continue
		}

		assert.Equal(t, memoryQuantity(test.value), obj.Val, "idx:%v text:`%v`", idx, test.text)
	}
}

func TestMemoryQuantityStringWithSuffix(t *testing.T) {
	tests := []struct {
		quantity memoryQuantity
		suffix   string
		expected string
	}{
		{memoryQuantity(1024), "Ki", "1Ki"},
		{memoryQuantity(1024 * 1024), "Mi", "1Mi"},
		{memoryQuantity(1024 * 1024 * 1024), "Gi", "1Gi"},
		{memoryQuantity(2 * 1024 * 1024), "Mi", "2Mi"},
		{memoryQuantity(1024 * 1024), "Gi", "0Gi"}, // 1Mi / 1Gi = 0Gi
		{memoryQuantity(0), "Mi", "0Mi"},
	}

	for _, test := range tests {
		result := test.quantity.StringWithSuffix(test.suffix)
		assert.Equal(t, test.expected, result)
	}
}

func TestExampleSDL(t *testing.T) {
	// Test that the example SDL file can be parsed correctly
	exampleSDL := `
services:
  web:
    image: nginx:latest
    command: ["nginx", "-g", "daemon off;"]
    count: 2
    resources:
      cpu:
        units: 0.5
      memory:
        size: 512Mi
      gpu:
        units: 1
    expose:
      - port: 80
        as: 80
        proto: tcp
        accept: ["10.0.0.0/8"]
  api:
    image: node:16-alpine
    command: ["node", "app.js"]
    count: 1
    resources:
      cpu:
        units: 1.0
      memory:
        size: 1Gi
      gpu:
        units: 1
`

	var sdl v1SDL
	err := yaml.Unmarshal([]byte(exampleSDL), &sdl)
	assert.NoError(t, err)

	// Verify the parsed structure
	assert.Len(t, sdl.Services, 2)

	// Check web service
	web, exists := sdl.Services["web"]
	assert.True(t, exists)
	assert.Equal(t, "nginx:latest", web.Image)
	assert.Equal(t, []string{"nginx", "-g", "daemon off;"}, web.Command)
	assert.Equal(t, 2, web.Count)
	assert.Equal(t, v1Cpu{Units: cpuQuantity(500)}, web.Resources.CPU)
	assert.Equal(t, v1Memory{Size: memoryQuantity(512 * 1024 * 1024)}, web.Resources.Memory)
	assert.Equal(t, v1GPU{Units: gpuQuantity(1)}, web.Resources.GPU)
	assert.Len(t, web.Expose, 1)
	assert.Equal(t, uint32(80), web.Expose[0].Port)
	assert.Equal(t, uint32(80), web.Expose[0].As)
	assert.Equal(t, "tcp", web.Expose[0].Proto)
	assert.Len(t, web.Expose[0].Accept.Items, 1)
	assert.Equal(t, "10.0.0.0/8", web.Expose[0].Accept.Items[0])

	// Check api service
	api, exists := sdl.Services["api"]
	assert.True(t, exists)
	assert.Equal(t, "node:16-alpine", api.Image)
	assert.Equal(t, []string{"node", "app.js"}, api.Command)
	assert.Equal(t, 1, api.Count)
	assert.Equal(t, v1Cpu{Units: cpuQuantity(1000)}, api.Resources.CPU)
	assert.Equal(t, v1Memory{Size: memoryQuantity(1024 * 1024 * 1024)}, api.Resources.Memory)
	assert.Equal(t, v1GPU{Units: gpuQuantity(1)}, api.Resources.GPU)
}

func TestSDLRead(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "valid SDL v1.0.0",
			yaml: `
version: 1.0.0
services:
  web:
    image: nginx:latest
    count: 1
`,
			wantErr: false,
		},
		{
			name: "valid SDL v1.0.0 with multiple services",
			yaml: `
version: 1.0.0
services:
  web:
    image: nginx:latest
    count: 1
  api:
    image: node:16
    count: 2
`,
			wantErr: false,
		},
		{
			name: "missing version",
			yaml: `
services:
  web:
    image: nginx:latest
    count: 1
`,
			wantErr: true,
		},
		{
			name: "unsupported version",
			yaml: `
version: 2.0.0
services:
  web:
    image: nginx:latest
    count: 1
`,
			wantErr: true,
		},
		{
			name: "invalid version format",
			yaml: `
version: invalid-version
services:
  web:
    image: nginx:latest
    count: 1
`,
			wantErr: true,
		},
		{
			name: "empty SDL",
			yaml: `
version: 1.0.0
services: {}
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sdl, err := Read([]byte(tt.yaml))

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, sdl)
		})
	}
}

func TestSDLReadFile(t *testing.T) {
	// Test with a temporary file
	validSDL := `
version: 1.0.0
services:
  web:
    image: nginx:latest
    count: 1
`

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "sdl_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write valid SDL to temp file
	if _, err := tmpFile.WriteString(validSDL); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Test reading the file
	sdl, err := ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sdl)

	// Test reading non-existent file
	_, err = ReadFile("/non/existent/file.yaml")
	assert.Error(t, err)
}

func TestGetTotalResources(t *testing.T) {
	sdl := &v1SDL{
		Services: map[string]v1Service{
			"web": {
				Count: 2,
				Resources: v1Resources{
					CPU:    v1Cpu{Units: cpuQuantity(500)},                    // 0.5 CPU
					Memory: v1Memory{Size: memoryQuantity(512 * 1024 * 1024)}, // 512Mi
					GPU:    v1GPU{Units: gpuQuantity(1)},
				},
				Volumes: v1ServiceVolumes{
					{
						Name:     "data",
						Mount:    "/app/data",
						ReadOnly: false,
					},
					{
						Name:     "logs",
						Mount:    "/app/logs",
						ReadOnly: false,
					},
				},
			},
			"api": {
				Count: 3,
				Resources: v1Resources{
					CPU:    v1Cpu{Units: cpuQuantity(1000)},                    // 1.0 CPU
					Memory: v1Memory{Size: memoryQuantity(1024 * 1024 * 1024)}, // 1Gi
					GPU:    v1GPU{Units: gpuQuantity(1)},
				},
				Volumes: v1ServiceVolumes{
					{
						Name:     "database",
						Mount:    "/var/lib/postgresql/data",
						ReadOnly: false,
					},
				},
			},
		},
		Volumes: map[string]v1Volume{
			"data": {
				Name:       "data",
				Size:       byteQuantity(10 * 1024 * 1024 * 1024), // 10Gi
				Persistent: true,
			},
			"logs": {
				Name:       "logs",
				Size:       byteQuantity(1 * 1024 * 1024 * 1024), // 1Gi
				Persistent: false,
			},
			"database": {
				Name:       "database",
				Size:       byteQuantity(50 * 1024 * 1024 * 1024), // 50Gi
				Persistent: true,
			},
		},
	}

	total := sdl.GetTotalResources()

	// Expected: web (2 instances) + api (3 instances)
	// CPU: (0.5 * 2) + (1.0 * 3) = 1.0 + 3.0 = 4.0 CPU
	// Memory: (512Mi * 2) + (1Gi * 3) = 1Gi + 3Gi = 4Gi
	// Storage: (10Gi * 2) + (50Gi * 3) = 20Gi + 150Gi = 170Gi
	expectedCPU := cpuQuantity(4000)                          // 4.0 * 1000
	expectedMemory := memoryQuantity(4 * 1024 * 1024 * 1024)  // 4Gi
	expectedStorage := byteQuantity(170 * 1024 * 1024 * 1024) // 170Gi

	assert.Equal(t, expectedCPU, total.CPU)
	assert.Equal(t, expectedMemory, total.Memory)
	assert.Equal(t, expectedStorage, total.Storage)
	assert.Equal(t, gpuQuantity(5), total.GPU)
}

func TestGetServiceResources(t *testing.T) {
	sdl := &v1SDL{
		Services: map[string]v1Service{
			"web": {
				Resources: v1Resources{
					CPU:    v1Cpu{Units: cpuQuantity(500)},
					Memory: v1Memory{Size: memoryQuantity(512 * 1024 * 1024)},
					GPU:    v1GPU{Units: gpuQuantity(1)},
				},
			},
		},
	}

	// Test existing service
	resources, exists := sdl.GetServiceResources("web")
	assert.True(t, exists)
	assert.Equal(t, v1Cpu{Units: cpuQuantity(500)}, resources.CPU)
	assert.Equal(t, v1Memory{Size: memoryQuantity(512 * 1024 * 1024)}, resources.Memory)

	// Test non-existing service
	_, exists = sdl.GetServiceResources("nonexistent")
	assert.False(t, exists)
}

func TestGetServiceCount(t *testing.T) {
	sdl := &v1SDL{
		Services: map[string]v1Service{
			"web": {Count: 2},
			"api": {Count: 3},
			"db":  {Count: 0}, // Should default to 1
		},
	}

	total := sdl.GetServiceCount()
	// Expected: 2 + 3 + 1 = 6
	assert.Equal(t, 6, total)
}
