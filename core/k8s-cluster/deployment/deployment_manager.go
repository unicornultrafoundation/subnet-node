package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/crd"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/deployment/builder"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/events"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/interfaces"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/manifest"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/session"
	clusterTypes "github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
	networkingv1 "k8s.io/api/networking/v1"
)

var (
	manifestGVR = schema.GroupVersionResource{
		Group:    "provider.local",
		Version:  "v1",
		Resource: "manifests",
	}
)

// k8sPatch represents a JSON patch operation
type k8sPatch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

// previousObj tracks objects for rollback
type previousObj struct {
	nns             *corev1.Namespace
	omani           *crd.K8sManifest
	umani           *crd.K8sManifest
	nmani           *crd.K8sManifest
	nNetPolicies    []networkingv1.NetworkPolicy
	nDeployments    []appsv1.Deployment
	uDeployments    []appsv1.Deployment
	oDeployments    []appsv1.Deployment
	nStatefulSets   []appsv1.StatefulSet
	uStatefulSets   []appsv1.StatefulSet
	oStatefulSets   []appsv1.StatefulSet
	nLocalServices  []corev1.Service
	uLocalServices  []corev1.Service
	oLocalServices  []corev1.Service
	nGlobalServices []corev1.Service
	uGlobalServices []corev1.Service
	oGlobalServices []corev1.Service
	nServiceCreds   []corev1.Secret
	uServiceCreds   []corev1.Secret
	oServiceCreds   []corev1.Secret
}

// deployObjNames tracks resource versions
type deployObjNames struct {
	deployments  map[string]string
	statefulSets map[string]string
	services     map[string]string
}

// DeploymentManager manages Kubernetes deployments
type DeploymentManager struct {
	logger         *zap.Logger
	client         *kubernetes.Clientset
	config         *clusterTypes.DeploymentManagerConfig
	eventBus       *events.DefaultEventBus[clusterTypes.MarketplaceEvent]
	managerChannel chan interfaces.DeploymentManagerInterface
	deployments    map[string]*clusterTypes.ManagedDeployment
	errorCount     int64
	warningCount   int64
	criticalCount  int64
	responseTime   time.Duration
	resourceUsage  map[string]float64
	mu             sync.RWMutex
	stopCh         chan struct{}
	ac             *crd.Client
	ns             string
	crdClient      *crd.Client
	session        *session.Session
}

// NewDeploymentManager creates a new deployment manager
func NewDeploymentManager(
	logger *zap.Logger,
	client *kubernetes.Clientset,
	config *clusterTypes.DeploymentManagerConfig,
	eventBus *events.DefaultEventBus[clusterTypes.MarketplaceEvent],
	crdClient *crd.Client,
	session *session.Session,
) *DeploymentManager {
	return &DeploymentManager{
		logger:         logger,
		client:         client,
		config:         config,
		eventBus:       eventBus,
		managerChannel: make(chan interfaces.DeploymentManagerInterface),
		deployments:    make(map[string]*clusterTypes.ManagedDeployment),
		resourceUsage:  make(map[string]float64),
		stopCh:         make(chan struct{}),
		ac:             crdClient,
		ns:             "deployment-default",
		crdClient:      crdClient,
		session:        session,
	}
}

// Start starts the deployment manager
func (m *DeploymentManager) Start(ctx context.Context) error {
	m.logger.Info("Starting deployment manager")
	return nil
}

// Stop stops the deployment manager
func (m *DeploymentManager) Stop() {
	m.logger.Info("Stopping deployment manager")
	close(m.stopCh)
}

// CreateDeployment creates a new deployment
func (m *DeploymentManager) CreateDeployment(ctx context.Context, id string, requester common.Address, sdl *manifest.SDL) error {
	// Convert SDL to Manifest
	mani, err := sdl.ToManifest()
	if err != nil {
		return fmt.Errorf("failed to convert SDL to Manifest: %w", err)
	}

	// Create managed deployment
	dep := &clusterTypes.ManagedDeployment{
		ID:        id,
		LeaseID:   id,
		Requester: requester,
		Manifest:  mani,
		SDL:       sdl,
		Status:    clusterTypes.DeploymentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Namespace: m.ns,
		Name:      id,
	}

	// Store deployment in memory
	m.mu.Lock()
	m.deployments[id] = dep
	m.mu.Unlock()

	// Clean up any existing resources first
	if err := m.cleanupStaleResources(ctx, dep); err != nil {
		m.logger.Error("Failed to clean up stale resources", zap.Error(err))
		// Continue with deployment even if cleanup fails
	}

	// Create Kubernetes resources
	if err := m.createKubernetesResources(ctx, dep); err != nil {
		m.logger.Error("Failed to create Kubernetes resources", zap.Error(err))
		dep.Status = clusterTypes.DeploymentStatusFailed
		dep.Error = err
		return fmt.Errorf("failed to create Kubernetes resources: %w", err)
	}

	// Update deployment status
	dep.Status = clusterTypes.DeploymentStatusRunning
	dep.UpdatedAt = time.Now()

	// Publish event
	event := &clusterTypes.DeploymentCompletedEvent{
		BaseEvent:    clusterTypes.BaseEvent{Timestamp: time.Now()},
		DeploymentID: id,
		Provider:     m.session.GetProviderAddress(),
		Status:       string(dep.Status),
	}
	if err := m.eventBus.Publish(ctx, string(clusterTypes.MarketplaceEventTypeDeploymentCompleted), event); err != nil {
		m.logger.Error("Failed to publish deployment event", zap.String("id", id), zap.Error(err))
	}

	// Start monitoring the deployment
	go m.monitorDeployments(ctx)

	return nil
}

// StopDeployment stops a deployment
func (m *DeploymentManager) StopDeployment(ctx context.Context, id string) error {
	m.logger.Info("Stopping deployment", zap.String("id", id))

	// Check if deployment exists
	m.mu.RLock()
	deployment, exists := m.deployments[id]
	m.mu.RUnlock()
	if !exists {
		return fmt.Errorf("deployment %s not found", id)
	}

	// Update deployment status
	deployment.Status = clusterTypes.DeploymentStatusTerminated
	deployment.UpdatedAt = time.Now()

	// Create and publish event before deleting resources
	event := &clusterTypes.DeploymentTerminatedEvent{
		BaseEvent: clusterTypes.BaseEvent{
			Timestamp: time.Now(),
		},
		DeploymentID: deployment.ID,
		Provider:     m.session.GetProviderAddress(),
		Requester:    deployment.Requester,
	}

	// Try to publish event before resource deletion
	if err := m.eventBus.Publish(ctx, string(clusterTypes.MarketplaceEventTypeDeploymentTerminated), event); err != nil {
		// Log warning instead of error since this is expected during shutdown
		m.logger.Warn("Failed to publish deployment event",
			zap.String("id", id),
			zap.Error(err))
	}

	// Delete Kubernetes resources
	if err := m.deleteKubernetesResources(ctx, deployment); err != nil {
		m.logger.Error("Failed to delete Kubernetes resources",
			zap.String("id", id),
			zap.Error(err))
		// Continue with cleanup even if resource deletion fails
	}

	// Remove deployment from manager
	m.mu.Lock()
	delete(m.deployments, id)
	m.mu.Unlock()

	return nil
}

// GetDeployment gets a deployment
func (m *DeploymentManager) GetDeployment(id string) (*clusterTypes.ManagedDeployment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	deployment, exists := m.deployments[id]
	if !exists {
		return nil, fmt.Errorf("deployment %s not found", id)
	}
	return deployment, nil
}

// ListDeployments lists all deployments
func (m *DeploymentManager) ListDeployments() ([]*clusterTypes.ManagedDeployment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	deployments := make([]*clusterTypes.ManagedDeployment, 0, len(m.deployments))
	for _, deployment := range m.deployments {
		deployments = append(deployments, deployment)
	}
	return deployments, nil
}

// IsHealthy checks if the deployment manager is healthy
func (m *DeploymentManager) IsHealthy() bool {
	return m.errorCount == 0 && m.warningCount == 0 && m.criticalCount == 0
}

// GetErrorCount gets the error count
func (m *DeploymentManager) GetErrorCount() int64 {
	return m.errorCount
}

// GetWarningCount gets the warning count
func (m *DeploymentManager) GetWarningCount() int64 {
	return m.warningCount
}

// GetCriticalCount gets the critical count
func (m *DeploymentManager) GetCriticalCount() int64 {
	return m.criticalCount
}

// GetResponseTime gets the response time
func (m *DeploymentManager) GetResponseTime() time.Duration {
	return m.responseTime
}

// GetResourceUsage gets the resource usage
func (m *DeploymentManager) GetResourceUsage() map[string]float64 {
	return m.resourceUsage
}

// GetDeploymentVersion gets the deployment version
func (m *DeploymentManager) GetDeploymentVersion(id string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	deployment, exists := m.deployments[id]
	if !exists {
		return 0, fmt.Errorf("deployment %s not found", id)
	}
	return deployment.Version, nil
}

// UpdateDeployment updates a deployment
func (m *DeploymentManager) UpdateDeployment(ctx context.Context, deployment *clusterTypes.ManagedDeployment) error {
	m.logger.Info("Updating deployment", zap.String("id", deployment.ID))

	// Check if deployment exists
	m.mu.Lock()
	defer m.mu.Unlock()

	existingDeployment, exists := m.deployments[deployment.ID]
	if !exists {
		return fmt.Errorf("deployment %s not found", deployment.ID)
	}

	// Update deployment
	existingDeployment.Status = deployment.Status
	existingDeployment.HealthStatus = deployment.HealthStatus
	existingDeployment.Version = deployment.Version
	existingDeployment.Resources = deployment.Resources
	existingDeployment.UpdatedAt = time.Now()
	existingDeployment.LastHealth = deployment.LastHealth
	existingDeployment.Error = deployment.Error

	return nil
}

// monitorDeployments monitors all deployments managed by this manager
func (m *DeploymentManager) monitorDeployments(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.mu.RLock()
			deployments := make([]*clusterTypes.ManagedDeployment, 0, len(m.deployments))
			for _, dep := range m.deployments {
				deployments = append(deployments, dep)
			}
			m.mu.RUnlock()

			for _, dep := range deployments {
				if err := m.checkDeploymentHealth(ctx, dep); err != nil {
					m.logger.Error("Deployment health check failed",
						zap.String("id", dep.ID),
						zap.Error(err))
					dep.Status = clusterTypes.DeploymentStatusFailed
					dep.Error = err
				}
			}
		}
	}
}

// checkDeploymentHealth checks the health of a deployment
func (m *DeploymentManager) checkDeploymentHealth(ctx context.Context, dep *clusterTypes.ManagedDeployment) error {
	// Check if namespace exists
	_, err := m.client.CoreV1().Namespaces().Get(ctx, dep.Namespace, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("namespace not found: %w", err)
	}

	// Check deployment status
	deployments, err := m.client.AppsV1().Deployments(dep.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	if len(deployments.Items) == 0 {
		return fmt.Errorf("no deployments found in namespace")
	}

	// Check if all deployments are ready
	for _, deployment := range deployments.Items {
		if deployment.Status.ReadyReplicas != *deployment.Spec.Replicas {
			// Get pod events for more details
			pods, err := m.client.CoreV1().Pods(dep.Namespace).List(ctx, metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", deployment.Name),
			})
			if err != nil {
				m.logger.Error("Failed to list pods", zap.Error(err))
			} else {
				for _, pod := range pods.Items {
					m.logger.Info("Pod details",
						zap.String("name", pod.Name),
						zap.String("phase", string(pod.Status.Phase)),
						zap.Any("conditions", pod.Status.Conditions),
						zap.Any("containerStatuses", pod.Status.ContainerStatuses))
				}
			}
			return fmt.Errorf("deployment %s not ready: %d/%d replicas ready",
				deployment.Name,
				deployment.Status.ReadyReplicas,
				*deployment.Spec.Replicas)
		}
	}

	// Check pod status
	pods, err := m.client.CoreV1().Pods(dep.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodRunning {
			m.logger.Info("Pod not running",
				zap.String("name", pod.Name),
				zap.String("phase", string(pod.Status.Phase)),
				zap.Any("conditions", pod.Status.Conditions),
				zap.Any("containerStatuses", pod.Status.ContainerStatuses))
			return fmt.Errorf("pod %s not running: %s", pod.Name, pod.Status.Phase)
		}
	}

	return nil
}

// applyDeployment applies a deployment to the cluster
func (m *DeploymentManager) applyDeployment(ctx context.Context, b *builder.Deployment) (*appsv1.Deployment, *appsv1.Deployment, *appsv1.Deployment, error) {
	oobj, err := m.client.AppsV1().Deployments(b.NS()).Get(ctx, b.Name(), metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, nil, nil, err
	}

	var nobj *appsv1.Deployment
	var uobj *appsv1.Deployment

	switch {
	case err == nil:
		uobj, err = b.Update(oobj)
		if err != nil {
			break
		}

		if !reflect.DeepEqual(&uobj.Spec, &oobj.Spec) ||
			!reflect.DeepEqual(uobj.Labels, oobj.Labels) {
			uobj, err = m.client.AppsV1().Deployments(b.NS()).Update(ctx, uobj, metav1.UpdateOptions{})
		}

		var patches []k8sPatch

		if rev := oobj.Spec.RevisionHistoryLimit; rev == nil || *rev != 10 {
			patches = append(patches, k8sPatch{
				Op:    "add",
				Path:  "/spec/revisionHistoryLimit",
				Value: int32(10),
			})
		}

		maxSurge := intstr.FromInt32(0)
		maxUnavailable := intstr.FromInt32(1)

		strategy := appsv1.DeploymentStrategy{
			Type: appsv1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &appsv1.RollingUpdateDeployment{
				MaxUnavailable: &maxUnavailable,
				MaxSurge:       &maxSurge,
			},
		}

		if !reflect.DeepEqual(&strategy, &oobj.Spec.Strategy) {
			patches = append(patches, k8sPatch{
				Op:    "replace",
				Path:  "/spec/strategy",
				Value: strategy,
			})
		}

		if len(patches) > 0 {
			data, _ := json.Marshal(patches)
			oobj, err = m.client.AppsV1().Deployments(b.NS()).Patch(ctx, oobj.Name, types.JSONPatchType, data, metav1.PatchOptions{})
		}
	case errors.IsNotFound(err):
		nobj, err = b.Create()
		if err == nil {
			nobj, err = m.client.AppsV1().Deployments(b.NS()).Create(ctx, nobj, metav1.CreateOptions{})
		}
	}

	return nobj, uobj, oobj, err
}

// applyService applies a service to the cluster
func (m *DeploymentManager) applyService(ctx context.Context, b *builder.Service) (*corev1.Service, *corev1.Service, *corev1.Service, error) {
	oobj, err := m.client.CoreV1().Services(b.NS()).Get(ctx, b.Name(), metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, nil, nil, err
	}

	var nobj *corev1.Service
	var uobj *corev1.Service

	switch {
	case err == nil:
		uobj, err = b.Update(oobj)
		if err == nil && (!reflect.DeepEqual(&uobj.Spec, &oobj.Spec) ||
			!reflect.DeepEqual(uobj.Labels, oobj.Labels)) {
			uobj, err = m.client.CoreV1().Services(b.NS()).Update(ctx, uobj, metav1.UpdateOptions{})
		}
	case errors.IsNotFound(err):
		nobj = b.Create()
		nobj, err = m.client.CoreV1().Services(b.NS()).Create(ctx, nobj, metav1.CreateOptions{})
	}

	return nobj, uobj, oobj, err
}

// createKubernetesResources creates Kubernetes resources for a deployment
func (m *DeploymentManager) createKubernetesResources(ctx context.Context, deployment *clusterTypes.ManagedDeployment) error {
	m.logger.Info("Creating Kubernetes resources", zap.String("id", deployment.ID))

	// Wait for namespace to be fully terminated if it exists
	nsName := fmt.Sprintf("deployment-%s", deployment.ID)
	_, err := m.client.CoreV1().Namespaces().Get(ctx, nsName, metav1.GetOptions{})
	if err == nil {
		// Namespace exists, wait for it to be terminated
		for {
			ns, err := m.client.CoreV1().Namespaces().Get(ctx, nsName, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					break // Namespace is gone, we can proceed
				}
				return fmt.Errorf("failed to check namespace status: %w", err)
			}
			if ns.Status.Phase == corev1.NamespaceTerminating {
				m.logger.Info("Waiting for namespace to terminate", zap.String("namespace", nsName))
				time.Sleep(time.Second)
				continue
			}
			break
		}
	}

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nsName,
		},
	}
	_, err = m.client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create namespace: %w", err)
	}
	m.logger.Info("Created namespace", zap.String("namespace", ns.Name))

	// Wait for namespace to be active
	for {
		ns, err := m.client.CoreV1().Namespaces().Get(ctx, nsName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to check namespace status: %w", err)
		}
		if ns.Status.Phase == corev1.NamespaceActive {
			break
		}
		m.logger.Info("Waiting for namespace to be active", zap.String("namespace", nsName))
		time.Sleep(time.Second)
	}

	// Create manifest
	manifestBuilder := builder.BuildManifest(m.logger, builder.Settings{
		Client:        m.client,
		DynamicClient: m.crdClient.DynamicClient,
		Logger:        m.logger,
	}, ns.Name, deployment)

	// Create manifest
	if err := manifestBuilder.Create(); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create manifest: %w", err)
	}
	m.logger.Info("Created manifest in namespace", zap.String("namespace", ns.Name), zap.String("manifestName", manifestBuilder.Name()))

	// Create network policy
	netPolBuilder := builder.BuildNetPol(builder.Settings{
		Client:        m.client,
		DynamicClient: m.crdClient.DynamicClient,
		Logger:        m.logger,
	}, deployment)

	// Create network policy object
	netPolObj, err := netPolBuilder.Create()
	if err != nil {
		return fmt.Errorf("failed to create network policy object: %w", err)
	}

	// Apply network policy if enabled
	if netPolObj != nil {
		_, err = m.client.NetworkingV1().NetworkPolicies(ns.Name).Create(ctx, netPolObj, metav1.CreateOptions{})
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to apply network policy: %w", err)
		}
	}

	// Create deployments and services
	for groupIdx, group := range deployment.Manifest.Groups {
		for serviceIdx := range group.Services {
			// Create workload
			workloadBuilder, err := builder.NewWorkloadBuilder(m.logger, builder.Settings{
				Client:        m.client,
				DynamicClient: m.crdClient.DynamicClient,
				Logger:        m.logger,
			}, deployment, groupIdx, serviceIdx)
			if err != nil {
				return fmt.Errorf("failed to create workload builder: %w", err)
			}

			// Create PVCs first
			pvcs := workloadBuilder.PersistentVolumeClaims()
			for _, pvc := range pvcs {
				_, err := m.client.CoreV1().PersistentVolumeClaims(ns.Name).Create(ctx, &pvc, metav1.CreateOptions{})
				if err != nil && !errors.IsAlreadyExists(err) {
					return fmt.Errorf("failed to create PVC: %w", err)
				}
			}

			// Create deployment
			deploymentBuilder, err := builder.NewDeploymentBuilder(m.logger, builder.Settings{
				Client:        m.client,
				DynamicClient: m.crdClient.DynamicClient,
				Logger:        m.logger,
			}, deployment, groupIdx, serviceIdx)
			if err != nil {
				return fmt.Errorf("failed to create deployment builder: %w", err)
			}

			// Create deployment object
			deploymentObj, err := deploymentBuilder.Create()
			if err != nil {
				return fmt.Errorf("failed to create deployment object: %w", err)
			}

			// Apply deployment
			_, err = m.client.AppsV1().Deployments(ns.Name).Create(ctx, deploymentObj, metav1.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				return fmt.Errorf("failed to apply deployment: %w", err)
			}

			// Create service
			serviceBuilder := builder.NewServiceBuilder(builder.Settings{
				Client:        m.client,
				DynamicClient: m.crdClient.DynamicClient,
				Logger:        m.logger,
			}, deployment, serviceIdx, isServiceGlobal(&group.Services[serviceIdx]), groupIdx)

			// Create service object
			serviceObj := serviceBuilder.Create()

			// Apply service
			_, err = m.client.CoreV1().Services(ns.Name).Create(ctx, serviceObj, metav1.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				return fmt.Errorf("failed to apply service: %w", err)
			}
		}
	}

	return nil
}

// isServiceGlobal checks if a service is configured as global
func isServiceGlobal(service *manifest.Service) bool {
	for _, expose := range service.Expose {
		for _, to := range expose.To {
			if to.Global {
				return true
			}
		}
	}
	return false
}

// cleanupStaleResources cleans up stale resources for a deployment
func (m *DeploymentManager) cleanupStaleResources(ctx context.Context, dep *clusterTypes.ManagedDeployment) error {
	// Create sets of expected resource names
	expectedDeployments := make(map[string]struct{})
	expectedStatefulSets := make(map[string]struct{})
	expectedServices := make(map[string]struct{})

	for groupIdx, group := range dep.Manifest.Groups {
		for serviceIdx := range group.Services {
			serviceName := fmt.Sprintf("%s-group-%d-service-%d", dep.Name, groupIdx, serviceIdx)
			expectedDeployments[serviceName] = struct{}{}
			expectedStatefulSets[serviceName] = struct{}{}
			expectedServices[serviceName] = struct{}{}
			expectedServices[serviceName+"-global"] = struct{}{}
		}
	}

	// Delete deployments
	for groupIdx, group := range dep.Manifest.Groups {
		for serviceIdx := range group.Services {
			serviceName := fmt.Sprintf("%s-group-%d-service-%d", dep.Name, groupIdx, serviceIdx)
			m.logger.Info("Checking deployment status before deletion",
				zap.String("deploymentID", dep.ID),
				zap.String("service", serviceName))

			// Check if deployment exists before trying to delete
			deployment, err := m.client.AppsV1().Deployments(dep.Namespace).Get(ctx, serviceName, metav1.GetOptions{})
			if err != nil {
				if !errors.IsNotFound(err) {
					m.logger.Error("Failed to check deployment status",
						zap.String("deploymentID", dep.ID),
						zap.String("service", serviceName),
						zap.Error(err))
					return fmt.Errorf("failed to check deployment %s: %w", serviceName, err)
				}
				m.logger.Info("Deployment not found, skipping deletion",
					zap.String("deploymentID", dep.ID),
					zap.String("service", serviceName))
				continue
			}

			m.logger.Info("Deleting deployment",
				zap.String("deploymentID", dep.ID),
				zap.String("service", serviceName),
				zap.Int32("replicas", *deployment.Spec.Replicas),
				zap.String("resourceVersion", deployment.ResourceVersion))

			propagationPolicy := metav1.DeletePropagationForeground
			err = m.client.AppsV1().Deployments(dep.Namespace).Delete(ctx, serviceName, metav1.DeleteOptions{
				PropagationPolicy: &propagationPolicy,
			})
			if err != nil && !errors.IsNotFound(err) {
				m.logger.Error("Failed to delete deployment",
					zap.String("deploymentID", dep.ID),
					zap.String("service", serviceName),
					zap.Error(err))
				return fmt.Errorf("failed to delete deployment for service %s: %w", serviceName, err)
			}
			m.logger.Info("Deployment deletion initiated",
				zap.String("deploymentID", dep.ID),
				zap.String("service", serviceName))
		}
	}

	// Delete services
	for groupIdx, group := range dep.Manifest.Groups {
		for serviceIdx := range group.Services {
			serviceName := fmt.Sprintf("%s-group-%d-service-%d", dep.Name, groupIdx, serviceIdx)
			globalServiceName := serviceName + "-global"

			// Delete local service
			m.logger.Info("Checking service status before deletion",
				zap.String("deploymentID", dep.ID),
				zap.String("service", serviceName))

			// Check if service exists before trying to delete
			svc, err := m.client.CoreV1().Services(dep.Namespace).Get(ctx, serviceName, metav1.GetOptions{})
			if err != nil {
				if !errors.IsNotFound(err) {
					m.logger.Error("Failed to check service status",
						zap.String("deploymentID", dep.ID),
						zap.String("service", serviceName),
						zap.Error(err))
					return fmt.Errorf("failed to check service %s: %w", serviceName, err)
				}
				m.logger.Info("Service not found, skipping deletion",
					zap.String("deploymentID", dep.ID),
					zap.String("service", serviceName))
			} else {
				m.logger.Info("Deleting service",
					zap.String("deploymentID", dep.ID),
					zap.String("service", serviceName),
					zap.String("type", string(svc.Spec.Type)),
					zap.String("clusterIP", svc.Spec.ClusterIP))

				err = m.client.CoreV1().Services(dep.Namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
				if err != nil && !errors.IsNotFound(err) {
					m.logger.Error("Failed to delete service",
						zap.String("deploymentID", dep.ID),
						zap.String("service", serviceName),
						zap.Error(err))
					return fmt.Errorf("failed to delete service for service %s: %w", serviceName, err)
				}
				m.logger.Info("Service deletion completed",
					zap.String("deploymentID", dep.ID),
					zap.String("service", serviceName))
			}

			// Delete global service
			m.logger.Info("Checking global service status before deletion",
				zap.String("deploymentID", dep.ID),
				zap.String("service", globalServiceName))

			// Check if global service exists before trying to delete
			globalSvc, err := m.client.CoreV1().Services(dep.Namespace).Get(ctx, globalServiceName, metav1.GetOptions{})
			if err != nil {
				if !errors.IsNotFound(err) {
					m.logger.Error("Failed to check global service status",
						zap.String("deploymentID", dep.ID),
						zap.String("service", globalServiceName),
						zap.Error(err))
					return fmt.Errorf("failed to check global service %s: %w", globalServiceName, err)
				}
				m.logger.Info("Global service not found, skipping deletion",
					zap.String("deploymentID", dep.ID),
					zap.String("service", globalServiceName))
			} else {
				m.logger.Info("Deleting global service",
					zap.String("deploymentID", dep.ID),
					zap.String("service", globalServiceName),
					zap.String("type", string(globalSvc.Spec.Type)),
					zap.String("clusterIP", globalSvc.Spec.ClusterIP))

				err = m.client.CoreV1().Services(dep.Namespace).Delete(ctx, globalServiceName, metav1.DeleteOptions{})
				if err != nil && !errors.IsNotFound(err) {
					m.logger.Error("Failed to delete global service",
						zap.String("deploymentID", dep.ID),
						zap.String("service", globalServiceName),
						zap.Error(err))
					return fmt.Errorf("failed to delete global service for service %s: %w", globalServiceName, err)
				}
				m.logger.Info("Global service deletion completed",
					zap.String("deploymentID", dep.ID),
					zap.String("service", globalServiceName))
			}

			// Delete service account
			serviceAccountName := fmt.Sprintf("%s-sa", serviceName)
			m.logger.Info("Checking service account status before deletion",
				zap.String("deploymentID", dep.ID),
				zap.String("serviceAccount", serviceAccountName))

			sa, err := m.client.CoreV1().ServiceAccounts(dep.Namespace).Get(ctx, serviceAccountName, metav1.GetOptions{})
			if err != nil {
				if !errors.IsNotFound(err) {
					m.logger.Error("Failed to check service account status",
						zap.String("deploymentID", dep.ID),
						zap.String("serviceAccount", serviceAccountName),
						zap.Error(err))
					return fmt.Errorf("failed to check service account %s: %w", serviceAccountName, err)
				}
				m.logger.Info("Service account not found, skipping deletion",
					zap.String("deploymentID", dep.ID),
					zap.String("serviceAccount", serviceAccountName))
				continue
			}

			m.logger.Info("Deleting service account",
				zap.String("deploymentID", dep.ID),
				zap.String("serviceAccount", serviceAccountName),
				zap.String("resourceVersion", sa.ResourceVersion))

			err = m.client.CoreV1().ServiceAccounts(dep.Namespace).Delete(ctx, serviceAccountName, metav1.DeleteOptions{})
			if err != nil && !errors.IsNotFound(err) {
				m.logger.Error("Failed to delete service account",
					zap.String("deploymentID", dep.ID),
					zap.String("serviceAccount", serviceAccountName),
					zap.Error(err))
				return fmt.Errorf("failed to delete service account %s: %w", serviceAccountName, err)
			}
			m.logger.Info("Service account deletion completed",
				zap.String("deploymentID", dep.ID),
				zap.String("serviceAccount", serviceAccountName))
		}
	}

	// Delete namespace
	m.logger.Info("Checking namespace status before deletion",
		zap.String("deploymentID", dep.ID),
		zap.String("namespace", dep.Namespace))

	ns, err := m.client.CoreV1().Namespaces().Get(ctx, dep.Namespace, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			m.logger.Error("Failed to check namespace status",
				zap.String("deploymentID", dep.ID),
				zap.String("namespace", dep.Namespace),
				zap.Error(err))
			return fmt.Errorf("failed to check namespace: %w", err)
		}
		m.logger.Info("Namespace not found, skipping deletion",
			zap.String("deploymentID", dep.ID),
			zap.String("namespace", dep.Namespace))
		return nil
	}

	m.logger.Info("Deleting namespace",
		zap.String("deploymentID", dep.ID),
		zap.String("namespace", dep.Namespace),
		zap.String("phase", string(ns.Status.Phase)),
		zap.String("resourceVersion", ns.ResourceVersion))

	propagationPolicy := metav1.DeletePropagationForeground
	err = m.client.CoreV1().Namespaces().Delete(ctx, dep.Namespace, metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	if err != nil && !errors.IsNotFound(err) {
		m.logger.Error("Failed to delete namespace",
			zap.String("deploymentID", dep.ID),
			zap.String("namespace", dep.Namespace),
			zap.Error(err))
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	m.logger.Info("Kubernetes resources deletion completed",
		zap.String("deploymentID", dep.ID),
		zap.String("namespace", dep.Namespace),
		zap.Time("endTime", time.Now()))

	return nil
}

// deleteKubernetesResources deletes Kubernetes resources for a deployment
func (m *DeploymentManager) deleteKubernetesResources(ctx context.Context, dep *clusterTypes.ManagedDeployment) error {
	m.logger.Info("Starting deletion of Kubernetes resources",
		zap.String("deploymentID", dep.ID),
		zap.String("namespace", dep.Namespace),
		zap.Time("startTime", time.Now()))

	// Delete deployments
	for groupIdx, group := range dep.Manifest.Groups {
		for serviceIdx := range group.Services {
			serviceName := fmt.Sprintf("%s-group-%d-service-%d", dep.Name, groupIdx, serviceIdx)
			m.logger.Info("Checking deployment status before deletion",
				zap.String("deploymentID", dep.ID),
				zap.String("service", serviceName))

			// Check if deployment exists before trying to delete
			deployment, err := m.client.AppsV1().Deployments(dep.Namespace).Get(ctx, serviceName, metav1.GetOptions{})
			if err != nil {
				if !errors.IsNotFound(err) {
					m.logger.Error("Failed to check deployment status",
						zap.String("deploymentID", dep.ID),
						zap.String("service", serviceName),
						zap.Error(err))
					return fmt.Errorf("failed to check deployment %s: %w", serviceName, err)
				}
				m.logger.Info("Deployment not found, skipping deletion",
					zap.String("deploymentID", dep.ID),
					zap.String("service", serviceName))
				continue
			}

			m.logger.Info("Deleting deployment",
				zap.String("deploymentID", dep.ID),
				zap.String("service", serviceName),
				zap.Int32("replicas", *deployment.Spec.Replicas),
				zap.String("resourceVersion", deployment.ResourceVersion))

			propagationPolicy := metav1.DeletePropagationForeground
			err = m.client.AppsV1().Deployments(dep.Namespace).Delete(ctx, serviceName, metav1.DeleteOptions{
				PropagationPolicy: &propagationPolicy,
			})
			if err != nil && !errors.IsNotFound(err) {
				m.logger.Error("Failed to delete deployment",
					zap.String("deploymentID", dep.ID),
					zap.String("service", serviceName),
					zap.Error(err))
				return fmt.Errorf("failed to delete deployment for service %s: %w", serviceName, err)
			}
			m.logger.Info("Deployment deletion initiated",
				zap.String("deploymentID", dep.ID),
				zap.String("service", serviceName))
		}
	}

	// Delete services
	for groupIdx, group := range dep.Manifest.Groups {
		for serviceIdx := range group.Services {
			serviceName := fmt.Sprintf("%s-group-%d-service-%d", dep.Name, groupIdx, serviceIdx)
			m.logger.Info("Checking service status before deletion",
				zap.String("deploymentID", dep.ID),
				zap.String("service", serviceName))

			// Check if service exists before trying to delete
			svc, err := m.client.CoreV1().Services(dep.Namespace).Get(ctx, serviceName, metav1.GetOptions{})
			if err != nil {
				if !errors.IsNotFound(err) {
					m.logger.Error("Failed to check service status",
						zap.String("deploymentID", dep.ID),
						zap.String("service", serviceName),
						zap.Error(err))
					return fmt.Errorf("failed to check service %s: %w", serviceName, err)
				}
				m.logger.Info("Service not found, skipping deletion",
					zap.String("deploymentID", dep.ID),
					zap.String("service", serviceName))
				continue
			}

			m.logger.Info("Deleting service",
				zap.String("deploymentID", dep.ID),
				zap.String("service", serviceName),
				zap.String("type", string(svc.Spec.Type)),
				zap.String("clusterIP", svc.Spec.ClusterIP))

			err = m.client.CoreV1().Services(dep.Namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
			if err != nil && !errors.IsNotFound(err) {
				m.logger.Error("Failed to delete service",
					zap.String("deploymentID", dep.ID),
					zap.String("service", serviceName),
					zap.Error(err))
				return fmt.Errorf("failed to delete service for service %s: %w", serviceName, err)
			}
			m.logger.Info("Service deletion completed",
				zap.String("deploymentID", dep.ID),
				zap.String("service", serviceName))

			// Delete service account
			serviceAccountName := fmt.Sprintf("%s-sa", serviceName)
			m.logger.Info("Checking service account status before deletion",
				zap.String("deploymentID", dep.ID),
				zap.String("serviceAccount", serviceAccountName))

			sa, err := m.client.CoreV1().ServiceAccounts(dep.Namespace).Get(ctx, serviceAccountName, metav1.GetOptions{})
			if err != nil {
				if !errors.IsNotFound(err) {
					m.logger.Error("Failed to check service account status",
						zap.String("deploymentID", dep.ID),
						zap.String("serviceAccount", serviceAccountName),
						zap.Error(err))
					return fmt.Errorf("failed to check service account %s: %w", serviceAccountName, err)
				}
				m.logger.Info("Service account not found, skipping deletion",
					zap.String("deploymentID", dep.ID),
					zap.String("serviceAccount", serviceAccountName))
				continue
			}

			m.logger.Info("Deleting service account",
				zap.String("deploymentID", dep.ID),
				zap.String("serviceAccount", serviceAccountName),
				zap.String("resourceVersion", sa.ResourceVersion))

			err = m.client.CoreV1().ServiceAccounts(dep.Namespace).Delete(ctx, serviceAccountName, metav1.DeleteOptions{})
			if err != nil && !errors.IsNotFound(err) {
				m.logger.Error("Failed to delete service account",
					zap.String("deploymentID", dep.ID),
					zap.String("serviceAccount", serviceAccountName),
					zap.Error(err))
				return fmt.Errorf("failed to delete service account %s: %w", serviceAccountName, err)
			}
			m.logger.Info("Service account deletion completed",
				zap.String("deploymentID", dep.ID),
				zap.String("serviceAccount", serviceAccountName))
		}
	}

	// Delete namespace
	m.logger.Info("Checking namespace status before deletion",
		zap.String("deploymentID", dep.ID),
		zap.String("namespace", dep.Namespace))

	ns, err := m.client.CoreV1().Namespaces().Get(ctx, dep.Namespace, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			m.logger.Error("Failed to check namespace status",
				zap.String("deploymentID", dep.ID),
				zap.String("namespace", dep.Namespace),
				zap.Error(err))
			return fmt.Errorf("failed to check namespace: %w", err)
		}
		m.logger.Info("Namespace not found, skipping deletion",
			zap.String("deploymentID", dep.ID),
			zap.String("namespace", dep.Namespace))
		return nil
	}

	m.logger.Info("Deleting namespace",
		zap.String("deploymentID", dep.ID),
		zap.String("namespace", dep.Namespace),
		zap.String("phase", string(ns.Status.Phase)),
		zap.String("resourceVersion", ns.ResourceVersion))

	propagationPolicy := metav1.DeletePropagationForeground
	err = m.client.CoreV1().Namespaces().Delete(ctx, dep.Namespace, metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	if err != nil && !errors.IsNotFound(err) {
		m.logger.Error("Failed to delete namespace",
			zap.String("deploymentID", dep.ID),
			zap.String("namespace", dep.Namespace),
			zap.Error(err))
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	m.logger.Info("Kubernetes resources deletion completed",
		zap.String("deploymentID", dep.ID),
		zap.String("namespace", dep.Namespace),
		zap.Time("endTime", time.Now()))

	return nil
}

// createHealthChecks creates health check probes for a service
func (m *DeploymentManager) createHealthChecks(ctx context.Context, namespace, serviceName string, health *manifest.HealthParams) error {
	if health == nil {
		return nil
	}

	maxRetries := 3
	var lastErr error

	for retry := 0; retry < maxRetries; retry++ {
		// Get the deployment
		deployment, err := m.client.AppsV1().Deployments(namespace).Get(ctx, serviceName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get deployment: %w", err)
		}

		// Update container with health checks
		if len(deployment.Spec.Template.Spec.Containers) > 0 {
			container := &deployment.Spec.Template.Spec.Containers[0]

			// Add readiness probe
			if health.Readiness != nil {
				container.ReadinessProbe = &corev1.Probe{
					InitialDelaySeconds: health.Readiness.InitialDelaySeconds,
					PeriodSeconds:       health.Readiness.PeriodSeconds,
					TimeoutSeconds:      health.Readiness.TimeoutSeconds,
					SuccessThreshold:    health.Readiness.SuccessThreshold,
					FailureThreshold:    health.Readiness.FailureThreshold,
				}

				if health.Readiness.HTTP != nil {
					container.ReadinessProbe.HTTPGet = &corev1.HTTPGetAction{
						Path: health.Readiness.HTTP.Path,
						Port: intstr.FromInt(int(health.Readiness.HTTP.Port)),
					}
				}
			}

			// Add liveness probe
			if health.Liveness != nil {
				container.LivenessProbe = &corev1.Probe{
					InitialDelaySeconds: health.Liveness.InitialDelaySeconds,
					PeriodSeconds:       health.Liveness.PeriodSeconds,
					TimeoutSeconds:      health.Liveness.TimeoutSeconds,
					SuccessThreshold:    health.Liveness.SuccessThreshold,
					FailureThreshold:    health.Liveness.FailureThreshold,
				}

				if health.Liveness.HTTP != nil {
					container.LivenessProbe.HTTPGet = &corev1.HTTPGetAction{
						Path: health.Liveness.HTTP.Path,
						Port: intstr.FromInt(int(health.Liveness.HTTP.Port)),
					}
				}
			}

			// Update the deployment with retry
			_, err = m.client.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
			if err == nil {
				return nil
			}

			lastErr = err
			m.logger.Warn("Failed to update deployment with health checks, retrying...",
				zap.String("service", serviceName),
				zap.Int("retry", retry+1),
				zap.Error(err))

			// Add a small delay before retrying
			time.Sleep(time.Duration(retry+1) * time.Second)
		}
	}

	return fmt.Errorf("failed to update deployment with health checks after %d retries: %w", maxRetries, lastErr)
}

// applyPVC applies a persistent volume claim to the cluster
func (m *DeploymentManager) applyPVC(ctx context.Context, b *builder.PVC) (*corev1.PersistentVolumeClaim, *corev1.PersistentVolumeClaim, *corev1.PersistentVolumeClaim, error) {
	oobj, err := m.client.CoreV1().PersistentVolumeClaims(b.NS()).Get(ctx, b.Name(), metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, nil, nil, err
	}

	var nobj *corev1.PersistentVolumeClaim
	var uobj *corev1.PersistentVolumeClaim

	switch {
	case err == nil:
		uobj, err = b.Update(oobj)
		if err == nil && (!reflect.DeepEqual(&uobj.Spec, &oobj.Spec) ||
			!reflect.DeepEqual(uobj.Labels, oobj.Labels)) {
			uobj, err = m.client.CoreV1().PersistentVolumeClaims(b.NS()).Update(ctx, uobj, metav1.UpdateOptions{})
		}
	case errors.IsNotFound(err):
		nobj, err = b.Create()
		if err == nil {
			nobj, err = m.client.CoreV1().PersistentVolumeClaims(b.NS()).Create(ctx, nobj, metav1.CreateOptions{})
		}
	}

	return nobj, uobj, oobj, err
}

// recover attempts to rollback changes in case of error
func (po *previousObj) recover(ctx context.Context, client kubernetes.Interface, crdClient *crd.Client) error {
	// Delete newly created resources
	for _, obj := range po.nDeployments {
		if err := client.AppsV1().Deployments(obj.Namespace).Delete(ctx, obj.Name, metav1.DeleteOptions{}); err != nil {
			return fmt.Errorf("failed to delete deployment: %v", err)
		}
	}
	for _, obj := range po.nStatefulSets {
		if err := client.AppsV1().StatefulSets(obj.Namespace).Delete(ctx, obj.Name, metav1.DeleteOptions{}); err != nil {
			return fmt.Errorf("failed to delete statefulset: %v", err)
		}
	}
	for _, obj := range po.nLocalServices {
		if err := client.CoreV1().Services(obj.Namespace).Delete(ctx, obj.Name, metav1.DeleteOptions{}); err != nil {
			return fmt.Errorf("failed to delete local service: %v", err)
		}
	}
	for _, obj := range po.nGlobalServices {
		if err := client.CoreV1().Services(obj.Namespace).Delete(ctx, obj.Name, metav1.DeleteOptions{}); err != nil {
			return fmt.Errorf("failed to delete global service: %v", err)
		}
	}
	for _, obj := range po.nServiceCreds {
		if err := client.CoreV1().Secrets(obj.Namespace).Delete(ctx, obj.Name, metav1.DeleteOptions{}); err != nil {
			return fmt.Errorf("failed to delete service credentials: %v", err)
		}
	}
	for _, obj := range po.nNetPolicies {
		if err := client.NetworkingV1().NetworkPolicies(obj.Namespace).Delete(ctx, obj.Name, metav1.DeleteOptions{}); err != nil {
			return fmt.Errorf("failed to delete network policy: %v", err)
		}
	}

	// Restore updated resources
	for _, obj := range po.uDeployments {
		if _, err := client.AppsV1().Deployments(obj.Namespace).Update(ctx, &obj, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("failed to restore deployment: %v", err)
		}
	}
	for _, obj := range po.uStatefulSets {
		if _, err := client.AppsV1().StatefulSets(obj.Namespace).Update(ctx, &obj, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("failed to restore statefulset: %v", err)
		}
	}
	for _, obj := range po.uLocalServices {
		if _, err := client.CoreV1().Services(obj.Namespace).Update(ctx, &obj, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("failed to restore local service: %v", err)
		}
	}
	for _, obj := range po.uGlobalServices {
		if _, err := client.CoreV1().Services(obj.Namespace).Update(ctx, &obj, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("failed to restore global service: %v", err)
		}
	}
	for _, obj := range po.uServiceCreds {
		if _, err := client.CoreV1().Secrets(obj.Namespace).Update(ctx, &obj, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("failed to restore service credentials: %v", err)
		}
	}

	// Restore manifest
	if po.omani != nil {
		if _, err := crdClient.Update(ctx, po.omani); err != nil {
			return fmt.Errorf("failed to restore manifest: %v", err)
		}
	}

	return nil
}

// applyServiceCredentials applies service credentials for a deployment
func (m *DeploymentManager) applyServiceCredentials(ctx context.Context, deployment *clusterTypes.ManagedDeployment) error {
	serviceNames := make([]string, 0, len(deployment.Manifest.Groups[0].Services))
	for i := range deployment.Manifest.Groups[0].Services {
		serviceName := fmt.Sprintf("service-%d", i)
		serviceNames = append(serviceNames, serviceName)
	}
	for groupIdx := range deployment.Manifest.Groups {
		for serviceIdx := range deployment.Manifest.Groups[groupIdx].Services {
			settings := builder.Settings{
				Client: m.client,
				Logger: m.logger,
			}
			builder, err := builder.NewServiceCredentialsBuilder(m.logger, settings, deployment, groupIdx, serviceIdx)
			if err != nil {
				return fmt.Errorf("failed to create service credentials builder: %w", err)
			}

			secret, err := builder.Create()
			if err != nil {
				return fmt.Errorf("failed to create service credentials: %w", err)
			}

			existing, err := m.client.CoreV1().Secrets(builder.NS()).Get(ctx, secret.Name, metav1.GetOptions{})
			if err != nil {
				if !errors.IsNotFound(err) {
					return fmt.Errorf("failed to get existing service credentials: %w", err)
				}

				// Create new secret
				_, err = m.client.CoreV1().Secrets(builder.NS()).Create(ctx, secret, metav1.CreateOptions{})
				if err != nil {
					return fmt.Errorf("failed to create service credentials: %w", err)
				}
				continue
			}

			// Update existing secret
			updated, err := builder.Update(existing)
			if err != nil {
				return fmt.Errorf("failed to update service credentials: %w", err)
			}

			_, err = m.client.CoreV1().Secrets(builder.NS()).Update(ctx, updated, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("failed to update service credentials: %w", err)
			}
		}
	}

	return nil
}
