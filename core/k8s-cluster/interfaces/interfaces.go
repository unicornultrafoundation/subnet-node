package interfaces

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"

	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/events"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/manifest"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/session"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
)

// DeploymentManagerInterface defines the interface for deployment management
type DeploymentManagerInterface interface {
	Start(ctx context.Context) error
	Stop()
	CreateDeployment(ctx context.Context, id string, requester common.Address, manifest *manifest.SDL) error
	StopDeployment(ctx context.Context, id string) error
	GetDeployment(id string) (*types.ManagedDeployment, error)
	ListDeployments() ([]*types.ManagedDeployment, error)
	IsHealthy() bool
	GetErrorCount() int64
	GetWarningCount() int64
	GetCriticalCount() int64
	GetResponseTime() time.Duration
	GetResourceUsage() map[string]float64
	GetDeploymentVersion(id string) (int64, error)
	UpdateDeployment(ctx context.Context, deployment *types.ManagedDeployment) error
}

// ServiceInterface defines the interface for service operations
type ServiceInterface interface {
	GetLogger() *zap.Logger
	GetClient() *kubernetes.Clientset
	GetConfig() *types.DeploymentManagerConfig
	GetEventBus() *events.DefaultEventBus[types.MarketplaceEvent]
	GetSession() *session.Session
	GetManagerChannel() chan DeploymentManagerInterface
}
