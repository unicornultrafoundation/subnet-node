package sdl

import (
	"errors"
	"net/url"

	"gopkg.in/yaml.v3"
)

var (
	errCannotSpecifyOffAndOtherCases = errors.New("if 'off' is specified, no other cases may be specified")
	errUnknownNextCase               = errors.New("next case is unknown")
	errHTTPOptionNotAllowed          = errors.New("http option not allowed")
	errSDLInvalid                    = errors.New("SDL invalid")
	errCredentialNoHost              = errors.New("Service Credentials missing Host")
	errCredentialNoUsername          = errors.New("Service Credentials missing Username")
	errCredentialNoPassword          = errors.New("Service Credentials missing Password")
)

type v1SDL struct {
	Services map[string]v1Service `yaml:"services" json:"services"`
	Volumes  map[string]v1Volume  `yaml:"volumes" json:"volumes"`
}

type v1Service struct {
	Image       string               `yaml:"image" json:"image"`
	Command     []string             `yaml:"command" json:"command"`
	Args        []string             `yaml:"args" json:"args"`
	Env         []string             `yaml:"env" json:"env"`
	Volumes     v1ServiceVolumes     `yaml:"volumes" json:"volumes"`
	Resources   v1Resources          `yaml:"resources" json:"resources"`
	Expose      v1Exposes            `yaml:"expose" json:"expose"`
	Count       int                  `yaml:"count" json:"count"`
	Credentials v1ServiceCredentials `yaml:"credentials" json:"credentials"`
}

type v1ServiceVolume struct {
	Name     string `yaml:"name" json:"name"`
	Mount    string `yaml:"mount" json:"mount"`
	ReadOnly bool   `yaml:"read_only" json:"read_only"`
}

type v1ServiceVolumes []v1ServiceVolume

type v1Exposes []v1Expose

type v1ExposeTo struct {
	Service     string        `yaml:"service,omitempty"`
	Global      bool          `yaml:"global,omitempty"`
	HTTPOptions v1HTTPOptions `yaml:"http_options"`
	IP          string        `yaml:"ip"`
}

type v1HTTPOptions struct {
	MaxBodySize uint32   `yaml:"max_body_size"`
	ReadTimeout uint32   `yaml:"read_timeout"`
	SendTimeout uint32   `yaml:"send_timeout"`
	NextTries   uint32   `yaml:"next_tries"`
	NextTimeout uint32   `yaml:"next_timeout"`
	NextCases   []string `yaml:"next_cases"`
}

type v1Volume struct {
	Name         string       `yaml:"name" json:"name"`
	Size         byteQuantity `yaml:"size" json:"size"`
	Persistent   bool         `yaml:"persistent" json:"persistent"`
	StorageClass string       `yaml:"storage_class" json:"storage_class"`
}

type v1Resources struct {
	CPU    v1Cpu    `yaml:"cpu" json:"cpu"`
	Memory v1Memory `yaml:"memory" json:"memory"`
	GPU    v1GPU    `yaml:"gpu" json:"gpu"`
}

type v1ServiceIngress struct {
	Name string
	Path string
}

type v1ServiceIngresses []v1ServiceIngress

type v1Cpu struct {
	Units cpuQuantity    `yaml:"units" json:"units"`
	Model string         `yaml:"model" json:"model"`
	Arch  string         `yaml:"arch" json:"arch"`
	Clock string         `yaml:"clock" json:"clock"`
	Cache memoryQuantity `yaml:"cache" json:"cache"`
}

type v1Memory struct {
	Size  memoryQuantity `yaml:"size" json:"size"`
	Type  string         `yaml:"type" json:"type"`
	Speed uint64         `yaml:"speed" json:"speed"`
}

type v1Expose struct {
	Port        uint32
	As          uint32
	Proto       string        `yaml:"proto,omitempty"`
	To          []v1ExposeTo  `yaml:"to,omitempty"`
	Accept      v1Accept      `yaml:"accept"`
	HTTPOptions v1HTTPOptions `yaml:"http_options"`
}

type v1Health struct {
	Path string `yaml:"path" json:"path"`
	Port int    `yaml:"port" json:"port"`
}

type v1ServiceCredentials struct {
	Host     string `yaml:",omitempty"`
	Username string `yaml:",omitempty"`
	Password string `yaml:",omitempty"`
}

type v1Accept struct {
	Items []string `yaml:"items,omitempty"`
}

func (p *v1Accept) UnmarshalYAML(node *yaml.Node) error {
	var accept []string
	if err := node.Decode(&accept); err != nil {
		return err
	}

	for _, item := range accept {
		if _, err := url.ParseRequestURI("http://" + item); err != nil {
			return err
		}
	}

	p.Items = accept

	return nil
}

// TotalResources represents the total resources needed for all services
type TotalResources struct {
	CPU     cpuQuantity
	Memory  memoryQuantity
	GPU     gpuQuantity
	Storage byteQuantity
}

// GetTotalResources calculates the total resources required for all services in the SDL
func (s *v1SDL) GetTotalResources() TotalResources {
	var total TotalResources

	for _, service := range s.Services {
		// Multiply resources by service count
		count := uint64(service.Count)
		if count == 0 {
			count = 1 // Default to 1 if count is not specified
		}

		// Add CPU resources
		total.CPU += service.Resources.CPU.Units * cpuQuantity(count)

		// Add Memory resources
		total.Memory += service.Resources.Memory.Size * memoryQuantity(count)

		// Add Storage resources from volumes
		for _, volume := range service.Volumes {
			// Only count persistent volumes for storage calculation
			if s.Volumes[volume.Name].Persistent {
				total.Storage += s.Volumes[volume.Name].Size * byteQuantity(count)
			}
		}

		// Add GPU resources (for now just count NVIDIA GPUs)
		total.GPU += service.Resources.GPU.Units * gpuQuantity(count)
	}

	return total
}

// GetServiceResources returns the resources for a specific service
func (s *v1SDL) GetServiceResources(serviceName string) (v1Resources, bool) {
	service, exists := s.Services[serviceName]
	if !exists {
		return v1Resources{}, false
	}
	return service.Resources, true
}

// GetServiceCount returns the total number of service instances across all services
func (s *v1SDL) GetServiceCount() int {
	total := 0
	for _, service := range s.Services {
		if service.Count > 0 {
			total += service.Count
		} else {
			total += 1 // Default to 1 if count is not specified
		}
	}
	return total
}
