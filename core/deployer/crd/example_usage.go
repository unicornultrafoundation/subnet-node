package crd

import (
	"encoding/json"
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ExampleManifest demonstrates how to create a Manifest resource
func ExampleManifest() *Manifest {
	// Create a new manifest
	manifest := &Manifest{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "depinsubnet.com/v1",
			Kind:       "Manifest",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-manifest",
			Namespace: "default",
		},
		Spec: ManifestSpec{
			Lease: LeaseSpec{
				Owner: "user123",
				ID:    12345,
			},
			Group: GroupSpec{
				Name: "example-group",
				Credentials: map[string]CredentialsSpec{
					"registry1": {
						Host:     "registry.example.com",
						Email:    "user@example.com",
						Username: "username",
						Password: "password",
					},
				},
				Volumes: map[string]VolumeSpec{
					"data-volume": {
						Size:  "10Gi",
						Class: "ssd",
						Attributes: map[string]interface{}{
							"type":        "persistent",
							"performance": "high",
							"encrypted":   true,
						},
					},
				},
				Services: map[string]ServiceSpec{
					"web-service": {
						Image:      "nginx:latest",
						Command:    []string{"nginx", "-g", "daemon off;"},
						Args:       []string{"--config", "/etc/nginx/nginx.conf"},
						Env:        []string{"NODE_ENV=production"},
						Credential: "registry1",
						Resources: ResourceSpec{
							CPU: &CPUResource{
								Units: "1000",
								Attributes: map[string]interface{}{
									"arch":    "x86_64",
									"vendor":  "intel",
									"cores":   8,
									"threads": 16,
								},
							},
							Memory: &MemoryResource{
								Size: "512Mi",
								Attributes: map[string]interface{}{
									"type":     "DDR4",
									"speed":    "3200MHz",
									"ecc":      true,
									"channels": 2,
								},
							},
							GPU: &GPUResource{
								Units: "1",
								Attributes: map[string]interface{}{
									"model":        "RTX 3080",
									"vendor":       "nvidia",
									"memory":       "10GB",
									"cuda":         true,
									"tensor_cores": 272,
								},
							},
						},
						Volumes: []VolumeSpec{
							{
								Name:     "data-volume",
								Mount:    "/data",
								ReadOnly: false,
							},
						},
						Count: 3,
						Expose: []ExposeSpec{
							{
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
							},
						},
					},
				},
			},
		},
	}

	return manifest
}

// DemonstrateManifestUsage shows how to use the Manifest struct
func DemonstrateManifestUsage() {
	manifest := ExampleManifest()

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Manifest JSON:\n%s\n", string(jsonData))

	// Access specific fields
	fmt.Printf("Owner: %s\n", manifest.Spec.Lease.Owner)
	fmt.Printf("Group Name: %s\n", manifest.Spec.Group.Name)

	// Access service by name
	if webService, exists := manifest.Spec.Group.Services["web-service"]; exists {
		fmt.Printf("Web Service Image: %s\n", webService.Image)
		fmt.Printf("Web Service Count: %d\n", webService.Count)

		// Access resource attributes
		if webService.Resources.CPU != nil {
			fmt.Printf("CPU Cores: %v\n", webService.Resources.CPU.Attributes["cores"])
		}
		if webService.Resources.Memory != nil {
			fmt.Printf("Memory Speed: %v\n", webService.Resources.Memory.Attributes["speed"])
		}
		if webService.Resources.GPU != nil {
			fmt.Printf("GPU Model: %v\n", webService.Resources.GPU.Attributes["model"])
		}

		// Access volumes and expose
		fmt.Printf("Number of Volumes: %d\n", len(webService.Volumes))
		fmt.Printf("Number of Expose Rules: %d\n", len(webService.Expose))
	}
}

// PrintManifestExample demonstrates how to serialize and print a manifest
func PrintManifestExample() {
	manifest := ExampleManifest()

	// Convert to JSON
	jsonData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling manifest: %v\n", err)
		return
	}

	fmt.Printf("Example Manifest JSON:\n%s\n", string(jsonData))
}

// ValidateManifest performs basic validation on a manifest
func ValidateManifest(manifest *Manifest) error {
	if manifest.Spec.Lease.Owner == "" {
		return fmt.Errorf("lease owner is required")
	}

	if manifest.Spec.Group.Name == "" {
		return fmt.Errorf("group name is required")
	}

	if len(manifest.Spec.Group.Services) == 0 {
		return fmt.Errorf("at least one service is required")
	}

	for serviceName, service := range manifest.Spec.Group.Services {
		if service.Image == "" {
			return fmt.Errorf("service %s image is required", serviceName)
		}

		if service.Count == 0 {
			return fmt.Errorf("service %s count must be greater than 0", serviceName)
		}
	}

	return nil
}
