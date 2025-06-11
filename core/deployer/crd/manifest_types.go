package crd

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// K8sManifest represents a manifest resource
type K8sManifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManifestSpec   `json:"spec,omitempty"`
	Status ManifestStatus `json:"status,omitempty"`

	// Unstructured is used for dynamic client operations
	Unstructured unstructured.Unstructured `json:"-"`
	// Context is not serialized
	Context context.Context `json:"-"`
}

// DeepCopyObject implements runtime.Object interface
func (m *K8sManifest) DeepCopyObject() runtime.Object {
	if m == nil {
		return nil
	}
	return m.DeepCopy()
}

// DeepCopy creates a deep copy of the K8sManifest
func (m *K8sManifest) DeepCopy() *K8sManifest {
	if m == nil {
		return nil
	}
	out := new(K8sManifest)
	m.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies the receiver into the given out
func (m *K8sManifest) DeepCopyInto(out *K8sManifest) {
	*out = *m
	out.TypeMeta = m.TypeMeta
	m.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	m.Spec.DeepCopyInto(&out.Spec)
	m.Status.DeepCopyInto(&out.Status)
}

// K8sManifestList represents a list of manifest resources
type K8sManifestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []K8sManifest `json:"items"`

	// UnstructuredList is used for dynamic client operations
	UnstructuredList unstructured.UnstructuredList `json:"-"`
}

// DeepCopyObject implements runtime.Object interface
func (l *K8sManifestList) DeepCopyObject() runtime.Object {
	if l == nil {
		return nil
	}
	return l.DeepCopy()
}

// DeepCopy creates a deep copy of the K8sManifestList
func (l *K8sManifestList) DeepCopy() *K8sManifestList {
	if l == nil {
		return nil
	}
	out := new(K8sManifestList)
	l.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies the receiver into the given out
func (l *K8sManifestList) DeepCopyInto(out *K8sManifestList) {
	*out = *l
	out.TypeMeta = l.TypeMeta
	l.ListMeta.DeepCopyInto(&out.ListMeta)
	if l.Items != nil {
		out.Items = make([]K8sManifest, len(l.Items))
		for i := range l.Items {
			l.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

// ManifestSpec represents the specification of a manifest
type ManifestSpec struct {
	Services []ServiceSpec `json:"services,omitempty"`
}

// DeepCopy creates a deep copy of the ManifestSpec
func (s *ManifestSpec) DeepCopy() *ManifestSpec {
	if s == nil {
		return nil
	}
	out := new(ManifestSpec)
	s.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies the receiver into the given out
func (s *ManifestSpec) DeepCopyInto(out *ManifestSpec) {
	*out = *s
	if s.Services != nil {
		out.Services = make([]ServiceSpec, len(s.Services))
		for i := range s.Services {
			s.Services[i].DeepCopyInto(&out.Services[i])
		}
	}
}

// ServiceSpec represents a service in the manifest
type ServiceSpec struct {
	Name      string       `json:"name"`
	Image     string       `json:"image"`
	Ports     []PortSpec   `json:"ports,omitempty"`
	Resources ResourceSpec `json:"resources,omitempty"`
}

// DeepCopy creates a deep copy of the ServiceSpec
func (s *ServiceSpec) DeepCopy() *ServiceSpec {
	if s == nil {
		return nil
	}
	out := new(ServiceSpec)
	s.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies the receiver into the given out
func (s *ServiceSpec) DeepCopyInto(out *ServiceSpec) {
	*out = *s
	if s.Ports != nil {
		out.Ports = make([]PortSpec, len(s.Ports))
		copy(out.Ports, s.Ports)
	}
	s.Resources.DeepCopyInto(&out.Resources)
}

// PortSpec represents a port configuration
type PortSpec struct {
	Name     string `json:"name"`
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"`
}

// ResourceSpec represents resource requirements
type ResourceSpec struct {
	CPU     ResourceValue `json:"cpu,omitempty"`
	Memory  ResourceValue `json:"memory,omitempty"`
	Storage ResourceValue `json:"storage,omitempty"`
}

// DeepCopy creates a deep copy of the ResourceSpec
func (r *ResourceSpec) DeepCopy() *ResourceSpec {
	if r == nil {
		return nil
	}
	out := new(ResourceSpec)
	r.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies the receiver into the given out
func (r *ResourceSpec) DeepCopyInto(out *ResourceSpec) {
	*out = *r
	r.CPU.DeepCopyInto(&out.CPU)
	r.Memory.DeepCopyInto(&out.Memory)
	r.Storage.DeepCopyInto(&out.Storage)
}

// ResourceValue represents a resource value with units
type ResourceValue struct {
	Units string `json:"units,omitempty"`
	Size  string `json:"size,omitempty"`
}

// DeepCopy creates a deep copy of the ResourceValue
func (r *ResourceValue) DeepCopy() *ResourceValue {
	if r == nil {
		return nil
	}
	out := new(ResourceValue)
	r.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies the receiver into the given out
func (r *ResourceValue) DeepCopyInto(out *ResourceValue) {
	*out = *r
}

// ManifestStatus represents the status of a manifest
type ManifestStatus struct {
	Phase       string `json:"phase,omitempty"`
	Message     string `json:"message,omitempty"`
	LastUpdated string `json:"lastUpdated,omitempty"`
}

// DeepCopy creates a deep copy of the ManifestStatus
func (s *ManifestStatus) DeepCopy() *ManifestStatus {
	if s == nil {
		return nil
	}
	out := new(ManifestStatus)
	s.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies the receiver into the given out
func (s *ManifestStatus) DeepCopyInto(out *ManifestStatus) {
	*out = *s
}

// K8sResource represents a Kubernetes resource
type K8sResource struct {
	// Kind is the Kubernetes resource kind
	Kind string `json:"kind"`
	// Name is the name of the resource
	Name string `json:"name"`
	// Namespace is the namespace of the resource
	Namespace string `json:"namespace"`
	// Spec is the resource specification
	Spec map[string]interface{} `json:"spec"`
}

// DeepCopyInto copies the receiver into the given out
func (r *K8sResource) DeepCopyInto(out *K8sResource) {
	*out = *r
	if r.Spec != nil {
		out.Spec = make(map[string]interface{})
		for k, v := range r.Spec {
			out.Spec[k] = v
		}
	}
}

// Service defines a containerized application service
type Service struct {
	Image            string            `json:"image"`
	Command          []string          `json:"command,omitempty"`
	Args             []string          `json:"args,omitempty"`
	Env              []string          `json:"env,omitempty"`
	Expose           []Expose          `json:"expose,omitempty"`
	DependsOn        []string          `json:"depends_on,omitempty"`
	Params           *ServiceParams    `json:"params,omitempty"`
	Credentials      *RegistryAuth     `json:"credentials,omitempty"`
	ImagePullSecrets []ImagePullSecret `json:"image_pull_secrets,omitempty"`
	Count            int32             `json:"count"`
	Resources        *ServiceResources `json:"resources,omitempty"`
	SchedulerParams  *SchedulerParams  `json:"scheduler_params,omitempty"`
}

// DeepCopyInto copies the receiver into the given out
func (s *Service) DeepCopyInto(out *Service) {
	*out = *s
	if s.Command != nil {
		out.Command = make([]string, len(s.Command))
		copy(out.Command, s.Command)
	}
	if s.Args != nil {
		out.Args = make([]string, len(s.Args))
		copy(out.Args, s.Args)
	}
	if s.Env != nil {
		out.Env = make([]string, len(s.Env))
		copy(out.Env, s.Env)
	}
	if s.Expose != nil {
		out.Expose = make([]Expose, len(s.Expose))
		for i := range s.Expose {
			s.Expose[i].DeepCopyInto(&out.Expose[i])
		}
	}
	if s.DependsOn != nil {
		out.DependsOn = make([]string, len(s.DependsOn))
		copy(out.DependsOn, s.DependsOn)
	}
	if s.Params != nil {
		out.Params = new(ServiceParams)
		s.Params.DeepCopyInto(out.Params)
	}
	if s.Credentials != nil {
		out.Credentials = new(RegistryAuth)
		*out.Credentials = *s.Credentials
	}
	if s.ImagePullSecrets != nil {
		out.ImagePullSecrets = make([]ImagePullSecret, len(s.ImagePullSecrets))
		copy(out.ImagePullSecrets, s.ImagePullSecrets)
	}
	if s.Resources != nil {
		out.Resources = new(ServiceResources)
		s.Resources.DeepCopyInto(out.Resources)
	}
	if s.SchedulerParams != nil {
		out.SchedulerParams = new(SchedulerParams)
		*out.SchedulerParams = *s.SchedulerParams
	}
}

// Expose defines service exposure configuration
type Expose struct {
	Port  int32  `json:"port"`
	Proto string `json:"proto"`
	To    []To   `json:"to"`
}

// DeepCopyInto copies the receiver into the given out
func (e *Expose) DeepCopyInto(out *Expose) {
	*out = *e
	if e.To != nil {
		out.To = make([]To, len(e.To))
		copy(out.To, e.To)
	}
}

// To defines service exposure target
type To struct {
	Service string `json:"service,omitempty"`
	Global  bool   `json:"global,omitempty"`
}

// ServiceParams defines service-specific parameters
type ServiceParams struct {
	Storage *StorageParams `json:"storage,omitempty"`
	Health  *HealthParams  `json:"health,omitempty"`
}

// DeepCopyInto copies the receiver into the given out
func (p *ServiceParams) DeepCopyInto(out *ServiceParams) {
	*out = *p
	if p.Storage != nil {
		out.Storage = new(StorageParams)
		*out.Storage = *p.Storage
	}
	if p.Health != nil {
		out.Health = new(HealthParams)
		p.Health.DeepCopyInto(out.Health)
	}
}

// StorageParams defines storage parameters
type StorageParams struct {
	Name       string             `json:"name"`
	Size       string             `json:"size"`
	Attributes *StorageAttributes `json:"attributes,omitempty"`
}

// StorageAttributes defines storage attributes
type StorageAttributes struct {
	Persistent bool   `json:"persistent"`
	Class      string `json:"class,omitempty"`
}

// HealthParams defines health check parameters
type HealthParams struct {
	Readiness *Probe `json:"readiness,omitempty"`
	Liveness  *Probe `json:"liveness,omitempty"`
}

// DeepCopyInto copies the receiver into the given out
func (h *HealthParams) DeepCopyInto(out *HealthParams) {
	*out = *h
	if h.Readiness != nil {
		out.Readiness = new(Probe)
		*out.Readiness = *h.Readiness
	}
	if h.Liveness != nil {
		out.Liveness = new(Probe)
		*out.Liveness = *h.Liveness
	}
}

// Probe defines a health check probe
type Probe struct {
	InitialDelaySeconds int32 `json:"initial_delay_seconds"`
	PeriodSeconds       int32 `json:"period_seconds"`
	TimeoutSeconds      int32 `json:"timeout_seconds"`
	SuccessThreshold    int32 `json:"success_threshold"`
	FailureThreshold    int32 `json:"failure_threshold"`
	HTTP                *HTTP `json:"http,omitempty"`
}

// HTTP defines HTTP probe configuration
type HTTP struct {
	Path string `json:"path"`
	Port int32  `json:"port"`
}

// RegistryAuth defines authentication for private container registries
type RegistryAuth struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// ImagePullSecret defines an image pull secret
type ImagePullSecret struct {
	Name string `json:"name"`
}

// ServiceResources defines the resource requirements for a service
type ServiceResources struct {
	CPU     *CPUResource     `json:"cpu,omitempty"`
	Memory  *MemoryResource  `json:"memory,omitempty"`
	Storage *StorageResource `json:"storage,omitempty"`
	GPU     *GPUResource     `json:"gpu,omitempty"`
}

// DeepCopyInto copies the receiver into the given out
func (r *ServiceResources) DeepCopyInto(out *ServiceResources) {
	*out = *r
	if r.CPU != nil {
		out.CPU = new(CPUResource)
		*out.CPU = *r.CPU
	}
	if r.Memory != nil {
		out.Memory = new(MemoryResource)
		*out.Memory = *r.Memory
	}
	if r.Storage != nil {
		out.Storage = new(StorageResource)
		*out.Storage = *r.Storage
	}
	if r.GPU != nil {
		out.GPU = new(GPUResource)
		r.GPU.DeepCopyInto(out.GPU)
	}
}

// CPUResource defines CPU resource requirements
type CPUResource struct {
	Units ResourceValue `json:"units"`
}

// MemoryResource defines memory resource requirements
type MemoryResource struct {
	Size ResourceValue `json:"size"`
}

// StorageResource defines storage resource requirements
type StorageResource struct {
	Size ResourceValue `json:"size"`
}

// GPUResource defines GPU requirements
type GPUResource struct {
	Units      int32         `json:"units"`
	Attributes GPUAttributes `json:"attributes"`
}

// DeepCopyInto copies the receiver into the given out
func (g *GPUResource) DeepCopyInto(out *GPUResource) {
	*out = *g
	out.Attributes = g.Attributes
}

// GPUAttributes defines GPU attributes
type GPUAttributes struct {
	Vendor     string `json:"vendor"`
	Model      string `json:"model"`
	Interface  string `json:"interface"`
	MemorySize string `json:"memory_size"`
}

// SchedulerParams defines scheduler-specific parameters
type SchedulerParams struct {
	RuntimeClass string `json:"runtime_class,omitempty"`
}

// Profiles defines compute and placement profiles
type Profiles struct {
	Compute   map[string]ComputeProfile   `json:"compute"`
	Placement map[string]PlacementProfile `json:"placement"`
}

// DeepCopyInto copies the receiver into the given out
func (p *Profiles) DeepCopyInto(out *Profiles) {
	*out = *p
	if p.Compute != nil {
		out.Compute = make(map[string]ComputeProfile)
		for k, v := range p.Compute {
			out.Compute[k] = v
		}
	}
	if p.Placement != nil {
		out.Placement = make(map[string]PlacementProfile)
		for k, v := range p.Placement {
			out.Placement[k] = v
		}
	}
}

// ComputeProfile defines a compute profile
type ComputeProfile struct {
	Resources ResourceRequirements `json:"resources"`
}

// DeepCopyInto copies the receiver into the given out
func (c *ComputeProfile) DeepCopyInto(out *ComputeProfile) {
	*out = *c
	c.Resources.DeepCopyInto(&out.Resources)
}

// ResourceRequirements defines resource requirements
type ResourceRequirements struct {
	CPU     ResourceLimit    `json:"cpu"`
	Memory  ResourceLimit    `json:"memory"`
	Storage []StorageVolume  `json:"storage,omitempty"`
	GPU     *GPUResource     `json:"gpu,omitempty"`
	Network *NetworkResource `json:"network,omitempty"`
}

// DeepCopyInto copies the receiver into the given out
func (r *ResourceRequirements) DeepCopyInto(out *ResourceRequirements) {
	*out = *r
	if r.Storage != nil {
		out.Storage = make([]StorageVolume, len(r.Storage))
		copy(out.Storage, r.Storage)
	}
	if r.GPU != nil {
		out.GPU = new(GPUResource)
		r.GPU.DeepCopyInto(out.GPU)
	}
	if r.Network != nil {
		out.Network = new(NetworkResource)
		*out.Network = *r.Network
	}
}

// ResourceLimit defines resource limits
type ResourceLimit struct {
	Request string `json:"request"`
	Limit   string `json:"limit"`
}

// StorageVolume defines a storage volume
type StorageVolume struct {
	Name       string             `json:"name,omitempty"`
	Size       string             `json:"size"`
	Attributes *StorageAttributes `json:"attributes,omitempty"`
}

// NetworkResource defines network resource requirements
type NetworkResource struct {
	Bandwidth string `json:"bandwidth,omitempty"`
}

// PlacementProfile defines a placement profile
type PlacementProfile struct {
	Attributes map[string]string        `json:"attributes"`
	Pricing    map[string]PricingConfig `json:"pricing"`
	SignedBy   *SignatureRequirements   `json:"signed_by,omitempty"`
}

// DeepCopyInto copies the receiver into the given out
func (p *PlacementProfile) DeepCopyInto(out *PlacementProfile) {
	*out = *p
	if p.Attributes != nil {
		out.Attributes = make(map[string]string)
		for k, v := range p.Attributes {
			out.Attributes[k] = v
		}
	}
	if p.Pricing != nil {
		out.Pricing = make(map[string]PricingConfig)
		for k, v := range p.Pricing {
			out.Pricing[k] = v
		}
	}
	if p.SignedBy != nil {
		out.SignedBy = new(SignatureRequirements)
		*out.SignedBy = *p.SignedBy
	}
}

// PricingConfig defines pricing configuration
type PricingConfig struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// SignatureRequirements defines signature requirements
type SignatureRequirements struct {
	AllOf []string `json:"all_of,omitempty"`
	AnyOf []string `json:"any_of,omitempty"`
}

// DeepCopyInto copies the receiver into the given out
func (s *SignatureRequirements) DeepCopyInto(out *SignatureRequirements) {
	*out = *s
	if s.AllOf != nil {
		out.AllOf = make([]string, len(s.AllOf))
		copy(out.AllOf, s.AllOf)
	}
	if s.AnyOf != nil {
		out.AnyOf = make([]string, len(s.AnyOf))
		copy(out.AnyOf, s.AnyOf)
	}
}

// K8sDeploymentSpec defines a Kubernetes deployment specification
type K8sDeploymentSpec struct {
	Profile string `json:"profile"`
	Count   int32  `json:"count"`
}

// Endpoint defines an endpoint
type Endpoint struct {
	Kind string `json:"kind"`
}
