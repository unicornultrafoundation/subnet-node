package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/manifest"
)

type DeploymentResponse struct {
	ID        string
	Requester common.Address
	Status    *DeploymentStatus
}

type Deployment struct {
	ID        string
	Requester common.Address
	Manifest  *manifest.Manifest
	Status    *DeploymentStatus
}

type DeploymentRequest struct {
	OrderID   string       `json:"order_id"`
	Manifest  manifest.SDL `json:"manifest"`
	Signature string       `json:"signature"`
	Requester string       `json:"requester"`
}

// DeploymentStatus represents the current status of a deployment
type DeploymentStatus struct {
	State      DeploymentState `json:"state"`
	Services   []ServiceStatus `json:"services"`
	Pods       []PodStatus     `json:"pods"`
	Endpoints  []EndpointInfo  `json:"endpoints"`
	CreatedAt  string          `json:"createdAt"`
	UpdatedAt  string          `json:"updatedAt"`
	DeployedAt string          `json:"deployedAt"`
}

// DeploymentState represents the overall state of a deployment
type DeploymentState string

const (
	DeploymentStateRunning  DeploymentState = "running"
	DeploymentStateStopped  DeploymentState = "stopped"
	DeploymentStateFailed   DeploymentState = "failed"
	DeploymentStatePending  DeploymentState = "pending"
	DeploymentStateNotFound DeploymentState = "notfound"
)

// ServiceStatus represents the status of a service within a deployment
type ServiceStatus struct {
	Name              string         `json:"name"`
	Group             string         `json:"group"`
	Image             string         `json:"image"`
	State             ServiceState   `json:"state"`
	Replicas          int32          `json:"replicas"`
	ReadyReplicas     int32          `json:"readyReplicas"`
	AvailableReplicas int32          `json:"availableReplicas"`
	URIs              []string       `json:"uris"`
	Ports             []ServicePort  `json:"ports"`
	Resources         *ResourceUsage `json:"resources"`
	LastUpdated       string         `json:"lastUpdated"`
}

// PodStatus represents the status of a pod within a deployment
type PodStatus struct {
	Name              string            `json:"name"`
	ServiceName       string            `json:"serviceName"`
	Group             string            `json:"group"`
	Image             string            `json:"image"`
	State             PodState          `json:"state"`
	Phase             string            `json:"phase"`
	Ready             bool              `json:"ready"`
	RestartCount      int32             `json:"restartCount"`
	IP                string            `json:"ip"`
	HostIP            string            `json:"hostIP"`
	Resources         *ResourceUsage    `json:"resources"`
	CreatedAt         string            `json:"createdAt"`
	StartedAt         string            `json:"startedAt"`
	LastRestartTime   string            `json:"lastRestartTime"`
	ContainerStatuses []ContainerStatus `json:"containerStatuses"`
}

// PodState represents the state of a pod
type PodState string

const (
	PodStateRunning PodState = "running"
	PodStateStopped PodState = "stopped"
	PodStateFailed  PodState = "failed"
	PodStatePending PodState = "pending"
	PodStateUnknown PodState = "unknown"
)

// ContainerStatus represents the status of a container within a pod
type ContainerStatus struct {
	Name         string `json:"name"`
	Image        string `json:"image"`
	Ready        bool   `json:"ready"`
	RestartCount int32  `json:"restartCount"`
	State        string `json:"state"`
	StartedAt    string `json:"startedAt"`
}

// ServiceState represents the state of a service
type ServiceState string

const (
	ServiceStateRunning  ServiceState = "running"
	ServiceStateStopped  ServiceState = "stopped"
	ServiceStateFailed   ServiceState = "failed"
	ServiceStatePending  ServiceState = "pending"
	ServiceStateNotFound ServiceState = "notfound"
)

// ServicePort represents an exposed port of a service
type ServicePort struct {
	Port        int32  `json:"port"`
	Protocol    string `json:"protocol"`
	ServiceType string `json:"serviceType"`
	URI         string `json:"uri"`
}

// EndpointInfo represents an endpoint configuration
type EndpointInfo struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Path     string `json:"path"`
	Protocol string `json:"protocol"`
	URI      string `json:"uri"`
}

// ResourceUsage represents resource usage information
type ResourceUsage struct {
	CPU     string `json:"cpu"`
	Memory  string `json:"memory"`
	Storage string `json:"storage"`
}

type DeploymentStats struct {
	UsedUploadBytes   uint64 `json:"usedUploadBytes"`
	UsedDownloadBytes uint64 `json:"usedDownloadBytes"`
	UsedGpu           uint64 `json:"usedGpu"`
	UsedCpu           uint64 `json:"usedCpu"`
	UsedMemory        uint64 `json:"usedMemory"`
	UsedStorage       uint64 `json:"usedStorage"`
	Duration          int64  `json:"duration"`
}

// DetailedDeploymentStatus extends DeploymentStatus with kubectl describe-like information
type DetailedDeploymentStatus struct {
	*DeploymentStatus
	Namespace         *NamespaceInfo         `json:"namespace"`
	Deployments       []DetailedDeployment   `json:"deployments"`
	Services          []DetailedService      `json:"services"`
	Pods              []DetailedPod          `json:"pods"`
	Events            []Event                `json:"events"`
	ConfigMaps        []ConfigMapInfo        `json:"configMaps"`
	Secrets           []SecretInfo           `json:"secrets"`
	PersistentVolumes []PersistentVolumeInfo `json:"persistentVolumes"`
}

// NamespaceInfo contains detailed namespace information
type NamespaceInfo struct {
	Name          string            `json:"name"`
	Status        string            `json:"status"`
	Labels        map[string]string `json:"labels"`
	Annotations   map[string]string `json:"annotations"`
	CreatedAt     string            `json:"createdAt"`
	ResourceQuota *ResourceQuota    `json:"resourceQuota"`
}

// DetailedDeployment contains detailed deployment information
type DetailedDeployment struct {
	Name                string                `json:"name"`
	Namespace           string                `json:"namespace"`
	Labels              map[string]string     `json:"labels"`
	Annotations         map[string]string     `json:"annotations"`
	Replicas            int32                 `json:"replicas"`
	ReadyReplicas       int32                 `json:"readyReplicas"`
	AvailableReplicas   int32                 `json:"availableReplicas"`
	UnavailableReplicas int32                 `json:"unavailableReplicas"`
	UpdatedReplicas     int32                 `json:"updatedReplicas"`
	Strategy            DeploymentStrategy    `json:"strategy"`
	Conditions          []DeploymentCondition `json:"conditions"`
	Selector            map[string]string     `json:"selector"`
	Template            PodTemplateSpec       `json:"template"`
	CreatedAt           string                `json:"createdAt"`
	UpdatedAt           string                `json:"updatedAt"`
}

// DeploymentStrategy contains deployment strategy information
type DeploymentStrategy struct {
	Type           string `json:"type"`
	MaxSurge       string `json:"maxSurge"`
	MaxUnavailable string `json:"maxUnavailable"`
}

// DeploymentCondition contains deployment condition information
type DeploymentCondition struct {
	Type               string `json:"type"`
	Status             string `json:"status"`
	LastUpdateTime     string `json:"lastUpdateTime"`
	LastTransitionTime string `json:"lastTransitionTime"`
	Reason             string `json:"reason"`
	Message            string `json:"message"`
}

// PodTemplateSpec contains pod template specification
type PodTemplateSpec struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Containers  []ContainerSpec   `json:"containers"`
	Volumes     []VolumeSpec      `json:"volumes"`
	Affinity    *AffinitySpec     `json:"affinity"`
	Tolerations []TolerationSpec  `json:"tolerations"`
}

// ContainerSpec contains container specification
type ContainerSpec struct {
	Name            string               `json:"name"`
	Image           string               `json:"image"`
	Command         []string             `json:"command"`
	Args            []string             `json:"args"`
	Ports           []ContainerPort      `json:"ports"`
	Env             []EnvVar             `json:"env"`
	Resources       ResourceRequirements `json:"resources"`
	VolumeMounts    []VolumeMount        `json:"volumeMounts"`
	LivenessProbe   *Probe               `json:"livenessProbe"`
	ReadinessProbe  *Probe               `json:"readinessProbe"`
	SecurityContext *SecurityContext     `json:"securityContext"`
}

// ContainerPort contains container port information
type ContainerPort struct {
	Name          string `json:"name"`
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol"`
	HostPort      int32  `json:"hostPort"`
}

// EnvVar contains environment variable information
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ResourceRequirements contains resource requirements
type ResourceRequirements struct {
	Requests map[string]string `json:"requests"`
	Limits   map[string]string `json:"limits"`
}

// VolumeMount contains volume mount information
type VolumeMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
	ReadOnly  bool   `json:"readOnly"`
}

// Probe contains probe information
type Probe struct {
	Type                string `json:"type"`
	InitialDelaySeconds int32  `json:"initialDelaySeconds"`
	PeriodSeconds       int32  `json:"periodSeconds"`
	TimeoutSeconds      int32  `json:"timeoutSeconds"`
	SuccessThreshold    int32  `json:"successThreshold"`
	FailureThreshold    int32  `json:"failureThreshold"`
}

// SecurityContext contains security context information
type SecurityContext struct {
	RunAsUser    *int64 `json:"runAsUser"`
	RunAsGroup   *int64 `json:"runAsGroup"`
	RunAsNonRoot *bool  `json:"runAsNonRoot"`
	Privileged   *bool  `json:"privileged"`
}

// VolumeSpec contains volume specification
type VolumeSpec struct {
	Name         string                 `json:"name"`
	VolumeSource map[string]interface{} `json:"volumeSource"`
}

// AffinitySpec contains affinity specification
type AffinitySpec struct {
	NodeAffinity    *NodeAffinity    `json:"nodeAffinity"`
	PodAffinity     *PodAffinity     `json:"podAffinity"`
	PodAntiAffinity *PodAntiAffinity `json:"podAntiAffinity"`
}

// NodeAffinity contains node affinity information
type NodeAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution  *NodeSelector             `json:"requiredDuringSchedulingIgnoredDuringExecution"`
	PreferredDuringSchedulingIgnoredDuringExecution []PreferredSchedulingTerm `json:"preferredDuringSchedulingIgnoredDuringExecution"`
}

// NodeSelector contains node selector information
type NodeSelector struct {
	NodeSelectorTerms []NodeSelectorTerm `json:"nodeSelectorTerms"`
}

// NodeSelectorTerm contains node selector term information
type NodeSelectorTerm struct {
	MatchExpressions []NodeSelectorRequirement `json:"matchExpressions"`
}

// NodeSelectorRequirement contains node selector requirement information
type NodeSelectorRequirement struct {
	Key      string   `json:"key"`
	Operator string   `json:"operator"`
	Values   []string `json:"values"`
}

// PreferredSchedulingTerm contains preferred scheduling term information
type PreferredSchedulingTerm struct {
	Weight     int32            `json:"weight"`
	Preference NodeSelectorTerm `json:"preference"`
}

// PodAffinity contains pod affinity information
type PodAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution  []PodAffinityTerm         `json:"requiredDuringSchedulingIgnoredDuringExecution"`
	PreferredDuringSchedulingIgnoredDuringExecution []WeightedPodAffinityTerm `json:"preferredDuringSchedulingIgnoredDuringExecution"`
}

// PodAntiAffinity contains pod anti-affinity information
type PodAntiAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution  []PodAffinityTerm         `json:"requiredDuringSchedulingIgnoredDuringExecution"`
	PreferredDuringSchedulingIgnoredDuringExecution []WeightedPodAffinityTerm `json:"preferredDuringSchedulingIgnoredDuringExecution"`
}

// PodAffinityTerm contains pod affinity term information
type PodAffinityTerm struct {
	LabelSelector *LabelSelector `json:"labelSelector"`
	Namespaces    []string       `json:"namespaces"`
	TopologyKey   string         `json:"topologyKey"`
}

// WeightedPodAffinityTerm contains weighted pod affinity term information
type WeightedPodAffinityTerm struct {
	Weight          int32           `json:"weight"`
	PodAffinityTerm PodAffinityTerm `json:"podAffinityTerm"`
}

// LabelSelector contains label selector information
type LabelSelector struct {
	MatchLabels      map[string]string          `json:"matchLabels"`
	MatchExpressions []LabelSelectorRequirement `json:"matchExpressions"`
}

// LabelSelectorRequirement contains label selector requirement information
type LabelSelectorRequirement struct {
	Key      string   `json:"key"`
	Operator string   `json:"operator"`
	Values   []string `json:"values"`
}

// TolerationSpec contains toleration specification
type TolerationSpec struct {
	Key      string `json:"key"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
	Effect   string `json:"effect"`
}

// DetailedService contains detailed service information
type DetailedService struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
	Type            string            `json:"type"`
	ClusterIP       string            `json:"clusterIP"`
	ExternalIPs     []string          `json:"externalIPs"`
	LoadBalancerIP  string            `json:"loadBalancerIP"`
	Ports           []ServicePort     `json:"ports"`
	Selector        map[string]string `json:"selector"`
	SessionAffinity string            `json:"sessionAffinity"`
	CreatedAt       string            `json:"createdAt"`
	Endpoints       []Endpoint        `json:"endpoints"`
}

// Endpoint contains endpoint information
type Endpoint struct {
	Addresses []string       `json:"addresses"`
	Ports     []EndpointPort `json:"ports"`
}

// EndpointPort contains endpoint port information
type EndpointPort struct {
	Name     string `json:"name"`
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"`
}

// DetailedPod contains detailed pod information
type DetailedPod struct {
	*PodStatus
	Labels             map[string]string `json:"labels"`
	Annotations        map[string]string `json:"annotations"`
	NodeName           string            `json:"nodeName"`
	HostNetwork        bool              `json:"hostNetwork"`
	DNSPolicy          string            `json:"dnsPolicy"`
	RestartPolicy      string            `json:"restartPolicy"`
	SchedulerName      string            `json:"schedulerName"`
	ServiceAccountName string            `json:"serviceAccountName"`
	Conditions         []PodCondition    `json:"conditions"`
	Volumes            []VolumeInfo      `json:"volumes"`
	QOSClass           string            `json:"qosClass"`
	Priority           *int32            `json:"priority"`
	PriorityClassName  string            `json:"priorityClassName"`
}

// PodCondition contains pod condition information
type PodCondition struct {
	Type               string `json:"type"`
	Status             string `json:"status"`
	LastProbeTime      string `json:"lastProbeTime"`
	LastTransitionTime string `json:"lastTransitionTime"`
	Reason             string `json:"reason"`
	Message            string `json:"message"`
}

// VolumeInfo contains volume information
type VolumeInfo struct {
	Name         string                 `json:"name"`
	VolumeSource map[string]interface{} `json:"volumeSource"`
}

// Event contains event information
type Event struct {
	Type           string              `json:"type"`
	Reason         string              `json:"reason"`
	Message        string              `json:"message"`
	Source         string              `json:"source"`
	Count          int32               `json:"count"`
	FirstTime      string              `json:"firstTime"`
	LastTime       string              `json:"lastTime"`
	InvolvedObject EventInvolvedObject `json:"involvedObject"`
}

// EventInvolvedObject contains event involved object information
type EventInvolvedObject struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	UID       string `json:"uid"`
}

// ConfigMapInfo contains configmap information
type ConfigMapInfo struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Data        map[string]string `json:"data"`
	BinaryData  map[string]string `json:"binaryData"`
	CreatedAt   string            `json:"createdAt"`
}

// SecretInfo contains secret information
type SecretInfo struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Type        string            `json:"type"`
	Data        map[string]string `json:"data"`
	CreatedAt   string            `json:"createdAt"`
}

// PersistentVolumeInfo contains persistent volume information
type PersistentVolumeInfo struct {
	Name                   string                 `json:"name"`
	Namespace              string                 `json:"namespace"`
	Labels                 map[string]string      `json:"labels"`
	Annotations            map[string]string      `json:"annotations"`
	Capacity               map[string]string      `json:"capacity"`
	AccessModes            []string               `json:"accessModes"`
	PersistentVolumeSource map[string]interface{} `json:"persistentVolumeSource"`
	Status                 string                 `json:"status"`
	ClaimRef               *ObjectReference       `json:"claimRef"`
	CreatedAt              string                 `json:"createdAt"`
}

// ObjectReference contains object reference information
type ObjectReference struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	UID       string `json:"uid"`
}

// ResourceQuota contains resource quota information
type ResourceQuota struct {
	Hard map[string]string `json:"hard"`
	Used map[string]string `json:"used"`
}
