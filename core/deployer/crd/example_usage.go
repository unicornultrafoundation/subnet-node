package crd

import (
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ExampleManifest demonstrates how to create a Manifest resource
func ExampleManifest() *Manifest {
	return &Manifest{
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
				Owner: "0x1234567890abcdef",
				ID:    12345,
			},
			Group: GroupSpec{
				Name: "example-group",
				Services: []ServiceSpec{
					{
						Name:  "web-service",
						Image: "nginx:latest",
						Command: []string{
							"nginx",
							"-g",
							"daemon off;",
						},
						Env: []string{
							"NGINX_HOST=localhost",
							"NGINX_PORT=80",
						},
						Credentials: &CredentialsSpec{
							Host:     "docker.io",
							Email:    "user@example.com",
							Username: "username",
							Password: "password",
						},
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
							GPU: &GPUResource{
								Units: 1,
								Attributes: []Attribute{
									{Name: "vendor", Value: "nvidia"},
									{Name: "model", Value: "RTX 3080"},
								},
							},
							Storage: []StorageResource{
								{
									Name: "data",
									Size: 1024,
									Attributes: []Attribute{
										{Name: "type", Value: "SSD"},
									},
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
						Count: 2,
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
									Hosts:       []string{"example.com", "www.example.com"},
								},
							},
						},
					},
				},
			},
		},
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

	for i, service := range manifest.Spec.Group.Services {
		if service.Name == "" {
			return fmt.Errorf("service %d name is required", i)
		}

		if service.Image == "" {
			return fmt.Errorf("service %d image is required", i)
		}

		if service.Count == 0 {
			return fmt.Errorf("service %d count must be greater than 0", i)
		}
	}

	return nil
}
