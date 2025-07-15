package v1alpha1

import (
	"fmt"
	"math/big"

	deployertypes "github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	"github.com/unicornultrafoundation/subnet-node/core/types"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Manifest represents the custom resource for deployment manifests
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Manifest struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Spec              ManifestSpec `json:"spec" yaml:"spec"`
}

// ManifestSpec defines the desired state of Manifest
type ManifestSpec struct {
	Lease LeaseSpec `json:"lease" yaml:"lease"`
	Group GroupSpec `json:"group" yaml:"group"`
}

// LeaseSpec defines the lease information
type LeaseSpec struct {
	Owner string `json:"owner" yaml:"owner"`
	ID    uint64 `json:"id" yaml:"id"`
}

// GroupSpec defines the group configuration
type GroupSpec struct {
	Name        string                     `json:"name" yaml:"name"`
	Credentials map[string]CredentialsSpec `json:"credentials,omitempty" yaml:"credentials,omitempty"`
	Volumes     map[string]VolumeSpec      `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	Services    map[string]ServiceSpec     `json:"services" yaml:"services"`
}

// ServiceSpec defines a service within a group
type ServiceSpec struct {
	Image      string       `json:"image" yaml:"image"`
	Command    []string     `json:"command,omitempty" yaml:"command,omitempty"`
	Args       []string     `json:"args,omitempty" yaml:"args,omitempty"`
	Env        []string     `json:"env,omitempty" yaml:"env,omitempty"`
	Credential string       `json:"credential,omitempty" yaml:"credential,omitempty"`
	Resources  ResourceSpec `json:"resources" yaml:"resources"`
	Volumes    []VolumeSpec `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	Count      uint32       `json:"count" yaml:"count"`
	Expose     []ExposeSpec `json:"expose,omitempty" yaml:"expose,omitempty"`
}

// CredentialsSpec defines container registry credentials
type CredentialsSpec struct {
	Host     string `json:"host" yaml:"host"`
	Email    string `json:"email" yaml:"email"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

// ResourceSpec defines resource requirements
type ResourceSpec struct {
	CPU    *CPUResource    `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Memory *MemoryResource `json:"memory,omitempty" yaml:"memory,omitempty"`
	GPU    *GPUResource    `json:"gpu,omitempty" yaml:"gpu,omitempty"`
}

// CPUResource defines CPU requirements
type CPUResource struct {
	Units      string           `json:"units" yaml:"units"`
	Attributes types.Attributes `json:"attributes,omitempty" yaml:"attributes,omitempty"`
}

// MemoryResource defines memory requirements
type MemoryResource struct {
	Size       string           `json:"size" yaml:"size"`
	Attributes types.Attributes `json:"attributes,omitempty" yaml:"attributes,omitempty"`
}

// GPUResource defines GPU requirements
type GPUResource struct {
	Units      string           `json:"units" yaml:"units"`
	Attributes types.Attributes `json:"attributes,omitempty" yaml:"attributes,omitempty"`
}

// VolumeSpec defines volume configuration
type VolumeSpec struct {
	Name       string           `json:"name,omitempty" yaml:"name,omitempty"`
	Mount      string           `json:"mount,omitempty" yaml:"mount,omitempty"`
	ReadOnly   bool             `json:"read_only,omitempty" yaml:"read_only,omitempty"`
	Size       string           `json:"size,omitempty" yaml:"size,omitempty"`
	Class      string           `json:"class,omitempty" yaml:"class,omitempty"`
	Attributes types.Attributes `json:"attributes,omitempty" yaml:"attributes,omitempty"`
}

// ExposeSpec defines service exposure configuration
type ExposeSpec struct {
	IP          string       `json:"ip" yaml:"ip"`
	Port        uint16       `json:"port" yaml:"port"`
	Proto       string       `json:"proto" yaml:"proto"`
	Service     string       `json:"service" yaml:"service"`
	Global      bool         `json:"global" yaml:"global"`
	HTTPOptions *HTTPOptions `json:"http_options,omitempty" yaml:"http_options,omitempty"`
	Hosts       []string     `json:"hosts,omitempty" yaml:"hosts,omitempty"`
}

// HTTPOptions defines HTTP-specific configuration
type HTTPOptions struct {
	MaxBodySize int      `json:"max_body_size,omitempty" yaml:"max_body_size,omitempty"`
	ReadTimeout int      `json:"read_timeout,omitempty" yaml:"read_timeout,omitempty"`
	SendTimeout int      `json:"send_timeout,omitempty" yaml:"send_timeout,omitempty"`
	NextTries   int      `json:"next_tries,omitempty" yaml:"next_tries,omitempty"`
	NextTimeout int      `json:"next_timeout,omitempty" yaml:"next_timeout,omitempty"`
	NextCases   []string `json:"next_cases,omitempty" yaml:"next_cases,omitempty"`
}

// Attribute represents a key-value attribute (kept for backward compatibility)
type Attribute struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

// ManifestList represents a list of Manifest resources
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ManifestList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Items           []Manifest `json:"items" yaml:"items"`
}

// ToGroupSpec converts the manifest's resources to a GroupSpec for requirement checking
func (m *Manifest) ToGroupSpec() (*deployertypes.GroupSpec, error) {
	// Aggregate resources from all services
	totalCPU := big.NewInt(0)
	totalMemory := big.NewInt(0)
	totalGPU := big.NewInt(0)

	// Collect all attributes from resources
	cpuAttributes := make(types.Attributes)
	memoryAttributes := make(types.Attributes)
	gpuAttributes := make(types.Attributes)

	// Collect volumes
	volumes := make(types.Volumes, 0)

	// Collect endpoints
	endpoints := make(types.Endpoints, 0)

	for _, service := range m.Spec.Group.Services {
		// Convert CPU
		if service.Resources.CPU != nil {
			cpuUnits, err := parseResourceSize(service.Resources.CPU.Units)
			if err != nil {
				return nil, err
			}
			cpuValue := new(big.Int).Mul(cpuUnits, big.NewInt(int64(service.Count)))
			totalCPU.Add(totalCPU, cpuValue)

			// Merge CPU attributes
			for k, v := range service.Resources.CPU.Attributes {
				cpuAttributes[k] = v
			}
		}

		// Convert Memory
		if service.Resources.Memory != nil {
			memorySize, err := parseResourceSize(service.Resources.Memory.Size)
			if err != nil {
				return nil, err
			}
			memoryValue := new(big.Int).Mul(memorySize, big.NewInt(int64(service.Count)))
			totalMemory.Add(totalMemory, memoryValue)

			// Merge memory attributes
			for k, v := range service.Resources.Memory.Attributes {
				memoryAttributes[k] = v
			}
		}

		// Convert GPU
		if service.Resources.GPU != nil {
			gpuUnits, err := parseResourceSize(service.Resources.GPU.Units)
			if err != nil {
				return nil, err
			}
			gpuValue := new(big.Int).Mul(gpuUnits, big.NewInt(int64(service.Count)))
			totalGPU.Add(totalGPU, gpuValue)

			// Merge GPU attributes
			for k, v := range service.Resources.GPU.Attributes {
				gpuAttributes[k] = v
			}
		}

		// Convert volumes
		for _, volume := range service.Volumes {
			volSize, err := parseResourceSize(volume.Size)
			if err != nil {
				return nil, err
			}

			vol := types.Volume{
				Name:       volume.Name,
				Size:       volSize,
				Attributes: volume.Attributes,
			}
			volumes = append(volumes, vol)
		}

		// Convert expose specs to endpoints
		for _, expose := range service.Expose {
			endpoint := types.Endpoint{
				Kind: getEndpointKind(expose),
			}
			endpoints = append(endpoints, endpoint)
		}
	}

	// Create the GroupSpec
	groupSpec := &deployertypes.GroupSpec{
		Name: m.Spec.Group.Name,
		Resources: types.Resources{
			Volumes:   volumes,
			Endpoints: endpoints,
		},
	}

	// Set CPU if there are CPU resources
	if totalCPU.Cmp(big.NewInt(0)) > 0 {
		groupSpec.Resources.CPU = &types.CPU{
			Units:      totalCPU,
			Attributes: cpuAttributes,
		}
	}

	// Set Memory if there are memory resources
	if totalMemory.Cmp(big.NewInt(0)) > 0 {
		groupSpec.Resources.Memory = &types.Memory{
			Size:       totalMemory,
			Attributes: memoryAttributes,
		}
	}

	// Set GPU if there are GPU resources
	if totalGPU.Cmp(big.NewInt(0)) > 0 {
		groupSpec.Resources.GPU = &types.GPU{
			Units:      totalGPU,
			Attributes: gpuAttributes,
		}
	}

	return groupSpec, nil
}

// parseResourceSize parses resource size string to big.Int (in bytes) using Kubernetes resource parsing
func parseResourceSize(size string) (*big.Int, error) {
	if size == "" {
		return big.NewInt(0), nil
	}

	// Use Kubernetes resource parsing
	quantity, err := resource.ParseQuantity(size)
	if err != nil {
		return nil, fmt.Errorf("invalid resource size format '%s': %w", size, err)
	}

	// Convert to bytes
	bytes := quantity.Value()
	return big.NewInt(bytes), nil
}

// getEndpointKind determines the endpoint kind based on expose configuration
func getEndpointKind(expose ExposeSpec) types.Endpoint_Kind {
	// If there are hosts specified, it's shared HTTP
	if len(expose.Hosts) > 0 {
		return types.Endpoint_SHARED_HTTP
	}

	// If Global is true, it's a leased IP
	if expose.IP != "" {
		return types.Endpoint_LEASED_IP
	}

	// Default is random port
	return types.Endpoint_RANDOM_PORT
}
