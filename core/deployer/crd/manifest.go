package crd

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Manifest represents the custom resource for deployment manifests
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
	Units      string                 `json:"units" yaml:"units"`
	Attributes map[string]interface{} `json:"attributes,omitempty" yaml:"attributes,omitempty"`
}

// MemoryResource defines memory requirements
type MemoryResource struct {
	Size       string                 `json:"size" yaml:"size"`
	Attributes map[string]interface{} `json:"attributes,omitempty" yaml:"attributes,omitempty"`
}

// GPUResource defines GPU requirements
type GPUResource struct {
	Units      string                 `json:"units" yaml:"units"`
	Attributes map[string]interface{} `json:"attributes,omitempty" yaml:"attributes,omitempty"`
}

// VolumeSpec defines volume configuration
type VolumeSpec struct {
	Name       string                 `json:"name,omitempty" yaml:"name,omitempty"`
	Mount      string                 `json:"mount,omitempty" yaml:"mount,omitempty"`
	ReadOnly   bool                   `json:"read_only,omitempty" yaml:"read_only,omitempty"`
	Size       string                 `json:"size,omitempty" yaml:"size,omitempty"`
	Class      string                 `json:"class,omitempty" yaml:"class,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty" yaml:"attributes,omitempty"`
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
type ManifestList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Items           []Manifest `json:"items" yaml:"items"`
}
