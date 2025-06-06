package types

import (
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

// K8sDeploymentManifest represents the deployment configuration
type K8sDeploymentManifest struct {
	Services []K8sService `json:"services"`
}

// K8sService represents a service in the deployment
type K8sService struct {
	Name      string        `json:"name"`
	Image     string        `json:"image"`
	Resources *K8sResources `json:"resources"`
	Ports     []K8sPort     `json:"ports"`
}

// K8sResources represents the resource requirements for a service
type K8sResources struct {
	CPU     *K8sCPU     `json:"cpu"`
	Memory  *K8sMemory  `json:"memory"`
	Storage *K8sStorage `json:"storage"`
}

// K8sCPU represents CPU resource requirements
type K8sCPU struct {
	Units float64 `json:"units"`
}

// K8sMemory represents memory resource requirements
type K8sMemory struct {
	Quantity resource.Quantity `json:"quantity"`
}

// K8sStorage represents storage resource requirements
type K8sStorage struct {
	Quantity resource.Quantity `json:"quantity"`
}

// K8sPort represents a service port
type K8sPort struct {
	Number int32 `json:"number"`
}

// K8sDeploymentStatus represents the status of a deployment
type K8sDeploymentStatus string

const (
	K8sDeploymentStatusPending   K8sDeploymentStatus = "pending"
	K8sDeploymentStatusRunning   K8sDeploymentStatus = "running"
	K8sDeploymentStatusFailed    K8sDeploymentStatus = "failed"
	K8sDeploymentStatusCompleted K8sDeploymentStatus = "completed"
)

// K8sBid represents a bid for a deployment
type K8sBid struct {
	ID        string    `json:"id"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}
