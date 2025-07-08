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
}

type v1Service struct {
	Image       string               `yaml:"image" json:"image"`
	Command     []string             `yaml:"command" json:"command"`
	Args        []string             `yaml:"args" json:"args"`
	Env         []string             `yaml:"env" json:"env"`
	Volumes     v1Volumes            `yaml:"volumes" json:"volumes"`
	Resources   v1Resources          `yaml:"resources" json:"resources"`
	Expose      v1Exposes            `yaml:"expose" json:"expose"`
	Ingress     v1Ingresses          `yaml:"ingress" json:"ingress"`
	Count       int                  `yaml:"count" json:"count"`
	Credentials v1ServiceCredentials `yaml:"credentials" json:"credentials"`
}

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
	MountPath    string       `yaml:"mount_path" json:"mount_path"`
	SubPath      string       `yaml:"sub_path" json:"sub_path"`
	ReadOnly     bool         `yaml:"read_only" json:"read_only"`
	Size         byteQuantity `yaml:"size" json:"size"`
	Persistent   bool         `yaml:"persistent" json:"persistent"`
	StorageClass string       `yaml:"storage_class" json:"storage_class"`
}

type v1Volumes []v1Volume

type v1Resources struct {
	CPU    cpuQuantity    `yaml:"cpu" json:"cpu"`
	Memory memoryQuantity `yaml:"memory" json:"memory"`
	GPU    gpuVendor      `yaml:"gpu" json:"gpu"`
}

type v1Ingresses []v1Ingress

type v1Ingress struct {
	Host string `yaml:"host" json:"host"`
	Port int    `yaml:"port" json:"port"`
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
	Email    string `yaml:",omitempty"`
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
