package manifest

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"

	"k8s.io/apimachinery/pkg/api/resource"
)

// SDL represents the Stack Definition Language structure
type SDL struct {
	VersionStr string                `yaml:"version" json:"version"`
	Services   map[string]Service    `yaml:"services" json:"services"`
	Profiles   Profiles              `yaml:"profiles" json:"profiles"`
	Deployment map[string]Deployment `yaml:"deployment" json:"deployment"`
	Endpoints  map[string]Endpoint   `yaml:"endpoints" json:"endpoints"`
}

// Service defines a containerized application service
type Service struct {
	Image            string            `yaml:"image" json:"image"`
	Command          []string          `yaml:"command,omitempty" json:"command,omitempty"`
	Args             []string          `yaml:"args,omitempty" json:"args,omitempty"`
	Env              []string          `yaml:"env,omitempty" json:"env,omitempty"`
	Expose           []Expose          `yaml:"expose,omitempty" json:"expose,omitempty"`
	DependsOn        []string          `yaml:"depends-on,omitempty" json:"depends-on,omitempty"`
	Params           *ServiceParams    `yaml:"params,omitempty" json:"params,omitempty"`
	Credentials      *RegistryAuth     `yaml:"credentials,omitempty" json:"credentials,omitempty"`
	ImagePullSecrets []ImagePullSecret `yaml:"imagePullSecrets,omitempty" json:"imagePullSecrets,omitempty"`
	Count            int32             `yaml:"count" json:"count"`
	Resources        *ServiceResources `yaml:"resources,omitempty" json:"resources,omitempty"`
	SchedulerParams  *SchedulerParams  `yaml:"scheduler_params,omitempty" json:"scheduler_params,omitempty"`
}

// ServiceParams defines service-specific parameters
type ServiceParams struct {
	Storage *StorageParams `yaml:"storage,omitempty" json:"storage,omitempty"`
	Health  *HealthParams  `yaml:"health,omitempty" json:"health,omitempty"`
}

// StorageParams defines storage parameters
type StorageParams struct {
	SHM *SHMParams `yaml:"shm,omitempty" json:"shm,omitempty"`
}

// SHMParams defines shared memory parameters
type SHMParams struct {
	Mount string `yaml:"mount" json:"mount"`
}

// HealthParams defines health check parameters
type HealthParams struct {
	Readiness *HealthCheck `yaml:"readiness,omitempty" json:"readiness,omitempty"`
	Liveness  *HealthCheck `yaml:"liveness,omitempty" json:"liveness,omitempty"`
}

// HealthCheck defines a health check configuration
type HealthCheck struct {
	HTTP                *HTTPHealthCheck `yaml:"http,omitempty" json:"http,omitempty"`
	InitialDelaySeconds int32            `yaml:"initial_delay_seconds" json:"initial_delay_seconds"`
	PeriodSeconds       int32            `yaml:"period_seconds" json:"period_seconds"`
	TimeoutSeconds      int32            `yaml:"timeout_seconds" json:"timeout_seconds"`
	SuccessThreshold    int32            `yaml:"success_threshold" json:"success_threshold"`
	FailureThreshold    int32            `yaml:"failure_threshold" json:"failure_threshold"`
}

// HTTPHealthCheck defines HTTP health check parameters
type HTTPHealthCheck struct {
	Path string `yaml:"path" json:"path"`
	Port int32  `yaml:"port" json:"port"`
}

// Expose defines how a service is exposed
type Expose struct {
	Port        int32        `yaml:"port" json:"port"`
	As          int32        `yaml:"as,omitempty" json:"as,omitempty"`
	Proto       string       `yaml:"proto,omitempty" json:"proto,omitempty"`
	To          []To         `yaml:"to,omitempty" json:"to,omitempty"`
	Accept      []string     `yaml:"accept,omitempty" json:"accept,omitempty"`
	HTTPOptions *HTTPOptions `yaml:"http_options,omitempty" json:"http_options,omitempty"`
}

// HTTPOptions defines HTTP-specific configuration
type HTTPOptions struct {
	MaxBodySize int64    `yaml:"max_body_size,omitempty" json:"max_body_size,omitempty"`
	NextCases   []string `yaml:"next_cases,omitempty" json:"next_cases,omitempty"`
}

// To defines the exposure target
type To struct {
	Service string `yaml:"service,omitempty" json:"service,omitempty"`
	Global  bool   `yaml:"global,omitempty" json:"global,omitempty"`
}

// Profiles defines compute and placement profiles
type Profiles struct {
	Compute   map[string]ComputeProfile   `yaml:"compute" json:"compute"`
	Placement map[string]PlacementProfile `yaml:"placement" json:"placement"`
}

// ComputeProfile defines a compute profile
type ComputeProfile struct {
	Resources ResourceRequirements `yaml:"resources" json:"resources"`
}

// ResourceRequirements defines compute resources
type ResourceRequirements struct {
	CPU     ResourceLimit    `yaml:"cpu" json:"cpu"`
	Memory  ResourceLimit    `yaml:"memory" json:"memory"`
	Storage []StorageVolume  `yaml:"storage,omitempty" json:"storage,omitempty"`
	GPU     *GPUResource     `yaml:"gpu,omitempty" json:"gpu,omitempty"`
	Network *NetworkResource `yaml:"network,omitempty" json:"network,omitempty"`
}

// ResourceLimit defines a resource limit
type ResourceLimit struct {
	Request string `yaml:"request" json:"request"`
	Limit   string `yaml:"limit" json:"limit"`
}

// StorageVolume defines a storage volume
type StorageVolume struct {
	Name       string             `yaml:"name,omitempty" json:"name,omitempty"`
	Size       string             `yaml:"size" json:"size"`
	Attributes *StorageAttributes `yaml:"attributes,omitempty" json:"attributes,omitempty"`
}

// StorageAttributes defines storage attributes
type StorageAttributes struct {
	Persistent bool   `yaml:"persistent" json:"persistent"`
	Class      string `yaml:"class,omitempty" json:"class,omitempty"`
}

// GPUResource defines GPU requirements
type GPUResource struct {
	Units      int32         `yaml:"units" json:"units"`
	Attributes GPUAttributes `yaml:"attributes" json:"attributes"`
}

// GPUAttributes defines GPU attributes
type GPUAttributes struct {
	Vendor map[string][]GPUModel `yaml:"vendor" json:"vendor"`
}

// GPUModel defines a specific GPU model
type GPUModel struct {
	Model     string `yaml:"model,omitempty" json:"model,omitempty"`
	RAM       string `yaml:"ram,omitempty" json:"ram,omitempty"`
	Interface string `yaml:"interface,omitempty" json:"interface,omitempty"`
}

// NetworkResource defines network requirements
type NetworkResource struct {
	Bandwidth string `yaml:"bandwidth,omitempty" json:"bandwidth,omitempty"`
}

// PlacementProfile defines a placement profile
type PlacementProfile struct {
	Attributes map[string]string        `yaml:"attributes" json:"attributes"`
	Pricing    map[string]PricingConfig `yaml:"pricing" json:"pricing"`
	SignedBy   *SignatureRequirements   `yaml:"signedBy,omitempty" json:"signedBy,omitempty"`
}

// PricingConfig defines pricing configuration
type PricingConfig struct {
	Denom  string `yaml:"denom" json:"denom"`
	Amount string `yaml:"amount" json:"amount"`
}

// SignatureRequirements defines signature requirements
type SignatureRequirements struct {
	AllOf []string `yaml:"allOf,omitempty" json:"allOf,omitempty"`
	AnyOf []string `yaml:"anyOf,omitempty" json:"anyOf,omitempty"`
}

// Deployment defines deployment settings
type Deployment struct {
	Profile string `yaml:"profile" json:"profile"`
	Count   int32  `yaml:"count" json:"count"`
}

// RegistryAuth defines authentication for private container registries
type RegistryAuth struct {
	Host     string `yaml:"host" json:"host"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

// ImagePullSecret defines an image pull secret
type ImagePullSecret struct {
	Name string `yaml:"name" json:"name"`
}

// Manifest represents the deployment manifest
type Manifest struct {
	Groups []Group `json:"groups"`
}

// Group represents a deployment group
type Group struct {
	Name     string    `json:"name"`
	Services []Service `json:"services"`
}

// Endpoint defines an endpoint configuration
type Endpoint struct {
	Kind string `yaml:"kind" json:"kind"`
}

// ServiceResources defines the resource requirements for a service
type ServiceResources struct {
	CPU     *CPUResource     `yaml:"cpu,omitempty" json:"cpu,omitempty"`
	Memory  *MemoryResource  `yaml:"memory,omitempty" json:"memory,omitempty"`
	Storage *StorageResource `yaml:"storage,omitempty" json:"storage,omitempty"`
	GPU     *GPUResource     `yaml:"gpu,omitempty" json:"gpu,omitempty"`
}

// CPUResource defines CPU resource requirements
type CPUResource struct {
	Units ResourceValue `yaml:"units" json:"units"`
}

// MemoryResource defines memory resource requirements
type MemoryResource struct {
	Size ResourceValue `yaml:"size" json:"size"`
}

// StorageResource defines storage resource requirements
type StorageResource struct {
	Size ResourceValue `yaml:"size" json:"size"`
}

// ResourceValue represents a resource value with units
type ResourceValue struct {
	Value int64  `yaml:"value" json:"value"`
	Unit  string `yaml:"unit" json:"unit"`
}

// SchedulerParams defines scheduler-specific parameters
type SchedulerParams struct {
	RuntimeClass string `yaml:"runtime_class,omitempty" json:"runtime_class,omitempty"`
}

var (
	serviceNameValidationRegex  = regexp.MustCompile(`^[a-z]([-a-z0-9]*[a-z0-9])?$`)
	endpointNameValidationRegex = regexp.MustCompile(`^[a-z]([-a-z0-9]*[a-z0-9])?$`)
)

// ToManifest converts SDL to a deployment manifest
func (s *SDL) ToManifest() (*Manifest, error) {
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("invalid SDL: %w", err)
	}

	groups := make([]Group, 0)
	for name, deployment := range s.Deployment {
		computeProfile, ok := s.Profiles.Compute[deployment.Profile]
		if !ok {
			return nil, fmt.Errorf("compute profile %s not found", deployment.Profile)
		}

		group := Group{
			Name:     name,
			Services: make([]Service, 0, len(s.Services)),
		}

		for _, svc := range s.Services {
			// Create a copy of the service with deployment count
			manifestSvc := svc
			manifestSvc.Count = deployment.Count

			// Apply compute profile resources
			if manifestSvc.Resources == nil {
				manifestSvc.Resources = &ServiceResources{}
			}

			// Parse CPU limit
			cpuLimit, err := resource.ParseQuantity(computeProfile.Resources.CPU.Limit)
			if err != nil {
				return nil, fmt.Errorf("invalid CPU limit: %w", err)
			}
			manifestSvc.Resources.CPU = &CPUResource{
				Units: ResourceValue{
					Value: cpuLimit.Value(),
					Unit:  "m", // Default to millicores
				},
			}

			// Parse memory limit
			memLimit, err := resource.ParseQuantity(computeProfile.Resources.Memory.Limit)
			if err != nil {
				return nil, fmt.Errorf("invalid memory limit: %w", err)
			}
			manifestSvc.Resources.Memory = &MemoryResource{
				Size: ResourceValue{
					Value: memLimit.Value(),
					Unit:  "Mi", // Default to MiB
				},
			}

			group.Services = append(group.Services, manifestSvc)
		}

		groups = append(groups, group)
	}

	return &Manifest{Groups: groups}, nil
}

// Version calculates the identifying deterministic hash for an SDL
func (s *SDL) Version() ([]byte, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	// Sort the JSON bytes for deterministic hashing
	sortedBytes, err := sortJSON(data)
	if err != nil {
		return nil, err
	}

	sum := sha256.Sum256(sortedBytes)
	return sum[:], nil
}

// Validate validates the SDL configuration
func (s *SDL) Validate() error {
	if s.VersionStr == "" {
		return fmt.Errorf("version is required")
	}

	if len(s.Services) == 0 {
		return fmt.Errorf("at least one service is required")
	}

	if len(s.Profiles.Compute) == 0 {
		return fmt.Errorf("at least one compute profile is required")
	}

	// Validate endpoints
	for endpointName, endpoint := range s.Endpoints {
		if !validateEndpointName(endpointName) {
			return fmt.Errorf("invalid endpoint name %s", endpointName)
		}
		if endpoint.Kind == "" {
			return fmt.Errorf("endpoint %s: kind is required", endpointName)
		}
	}

	// Validate services
	for name, service := range s.Services {
		if !serviceNameValidationRegex.MatchString(name) {
			return fmt.Errorf("invalid service name %s", name)
		}

		if service.Image == "" {
			return fmt.Errorf("service %s: image is required", name)
		}

		// Validate expose configuration
		for _, expose := range service.Expose {
			if expose.Port <= 0 {
				return fmt.Errorf("service %s: invalid port number %d", name, expose.Port)
			}
			if expose.Proto != "" && expose.Proto != "TCP" && expose.Proto != "UDP" {
				return fmt.Errorf("service %s: invalid protocol %s", name, expose.Proto)
			}
			if expose.HTTPOptions != nil {
				if expose.HTTPOptions.MaxBodySize < 0 {
					return fmt.Errorf("service %s: invalid max body size %d", name, expose.HTTPOptions.MaxBodySize)
				}
			}
		}

		// Validate health checks
		if service.Params != nil && service.Params.Health != nil {
			if err := validateHealthCheck(service.Params.Health.Readiness); err != nil {
				return fmt.Errorf("service %s: invalid readiness check: %w", name, err)
			}
			if err := validateHealthCheck(service.Params.Health.Liveness); err != nil {
				return fmt.Errorf("service %s: invalid liveness check: %w", name, err)
			}
		}
	}

	return nil
}

// validateHealthCheck validates a health check configuration
func validateHealthCheck(check *HealthCheck) error {
	if check == nil {
		return nil
	}

	if check.HTTP != nil {
		if check.HTTP.Path == "" {
			return fmt.Errorf("HTTP health check path is required")
		}
		if check.HTTP.Port <= 0 {
			return fmt.Errorf("HTTP health check port must be positive")
		}
	}

	if check.InitialDelaySeconds < 0 {
		return fmt.Errorf("initial delay seconds must be non-negative")
	}
	if check.PeriodSeconds <= 0 {
		return fmt.Errorf("period seconds must be positive")
	}
	if check.TimeoutSeconds <= 0 {
		return fmt.Errorf("timeout seconds must be positive")
	}
	if check.SuccessThreshold <= 0 {
		return fmt.Errorf("success threshold must be positive")
	}
	if check.FailureThreshold <= 0 {
		return fmt.Errorf("failure threshold must be positive")
	}

	return nil
}

// validateEndpointName validates an endpoint name
func validateEndpointName(name string) bool {
	return endpointNameValidationRegex.MatchString(name)
}

// ToK8sResources converts SDL resources to Kubernetes resources
func (r *ResourceRequirements) ToK8sResources() (map[string]resource.Quantity, error) {
	resources := make(map[string]resource.Quantity)

	// Convert CPU
	if r.CPU.Request != "" {
		cpu, err := resource.ParseQuantity(r.CPU.Request)
		if err != nil {
			return nil, fmt.Errorf("invalid CPU request: %w", err)
		}
		resources["cpu"] = cpu
	}

	// Convert Memory
	if r.Memory.Request != "" {
		memory, err := resource.ParseQuantity(r.Memory.Request)
		if err != nil {
			return nil, fmt.Errorf("invalid memory request: %w", err)
		}
		resources["memory"] = memory
	}

	// Convert Storage
	for _, storage := range r.Storage {
		if storage.Size != "" {
			size, err := resource.ParseQuantity(storage.Size)
			if err != nil {
				return nil, fmt.Errorf("invalid storage size: %w", err)
			}
			resources["storage"] = size
		}
	}

	// Convert GPU
	if r.GPU != nil {
		gpu, err := resource.ParseQuantity(fmt.Sprintf("%d", r.GPU.Units))
		if err != nil {
			return nil, fmt.Errorf("invalid GPU units: %w", err)
		}
		resources["nvidia.com/gpu"] = gpu
	}

	return resources, nil
}

// sortJSON sorts the keys in a JSON object for deterministic hashing
func sortJSON(data []byte) ([]byte, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return json.Marshal(v)
}
