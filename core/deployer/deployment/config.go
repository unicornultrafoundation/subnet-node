package deployment

import (
	"k8s.io/client-go/rest"
)

// DeploymentManagerConfig represents deployment manager configuration
type DeploymentManagerConfig struct {
	KubeConfig    *rest.Config
	StoreDir      string
	MaxRetries    int
	DeploymentDir string
}
