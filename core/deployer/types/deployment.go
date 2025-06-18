package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/manifest"
)

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
	State     DeploymentState `json:"state"`
	Services  []ServiceStatus `json:"services"`
	Endpoints []EndpointInfo  `json:"endpoints"`
	CreatedAt string          `json:"createdAt"`
	UpdatedAt string          `json:"updatedAt"`
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
