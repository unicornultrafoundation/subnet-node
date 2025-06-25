package deployment

import (
	"context"
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/config"
)

// Mock implementations
type MockTenantManager struct {
	mock.Mock
}

func (m *MockTenantManager) CreateTenant(ctx context.Context, tenant *Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *MockTenantManager) GetTenant(ctx context.Context, tenantID string) (*Tenant, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(*Tenant), args.Error(1)
}

func (m *MockTenantManager) ListTenants(ctx context.Context) ([]*Tenant, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*Tenant), args.Error(1)
}

func (m *MockTenantManager) UpdateTenant(ctx context.Context, tenant *Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *MockTenantManager) DeleteTenant(ctx context.Context, tenantID string) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

func (m *MockTenantManager) GetTenantResourceUsage(ctx context.Context, tenantID string) (*TenantResourceUsage, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(*TenantResourceUsage), args.Error(1)
}

type MockPortManager struct {
	mock.Mock
}

func (m *MockPortManager) AllocatePort(ctx context.Context, tenantID string, preferredPort int) (int, error) {
	args := m.Called(ctx, tenantID, preferredPort)
	return args.Int(0), args.Error(1)
}

func (m *MockPortManager) ReleasePort(ctx context.Context, port int) error {
	args := m.Called(ctx, port)
	return args.Error(0)
}

func (m *MockPortManager) GetPortPool(ctx context.Context) *PortPool {
	args := m.Called(ctx)
	return args.Get(0).(*PortPool)
}

func (m *MockPortManager) ReservePortRange(ctx context.Context, tenantID string, startPort, endPort int) error {
	args := m.Called(ctx, tenantID, startPort, endPort)
	return args.Error(0)
}

func (m *MockPortManager) ReleasePortRange(ctx context.Context, startPort, endPort int) error {
	args := m.Called(ctx, startPort, endPort)
	return args.Error(0)
}

func (m *MockPortManager) GetTenantPorts(ctx context.Context, tenantID string) ([]int, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]int), args.Error(1)
}

type MockManifestManager struct {
	mock.Mock
}

func (m *MockManifestManager) ValidateManifest(ctx context.Context, manifest *ComposeManifest) error {
	args := m.Called(ctx, manifest)
	return args.Error(0)
}

func (m *MockManifestManager) ProcessManifest(ctx context.Context, manifest *ComposeManifest, tenantID string) (*ComposeManifest, error) {
	args := m.Called(ctx, manifest, tenantID)
	return args.Get(0).(*ComposeManifest), args.Error(1)
}

func (m *MockManifestManager) GenerateManifest(ctx context.Context, template string, params map[string]interface{}) (*ComposeManifest, error) {
	args := m.Called(ctx, template, params)
	return args.Get(0).(*ComposeManifest), args.Error(1)
}

func (m *MockManifestManager) MergeManifests(ctx context.Context, manifests ...*ComposeManifest) (*ComposeManifest, error) {
	args := m.Called(ctx, manifests)
	return args.Get(0).(*ComposeManifest), args.Error(1)
}

func (m *MockManifestManager) SerializeManifest(ctx context.Context, manifest *ComposeManifest) ([]byte, error) {
	args := m.Called(ctx, manifest)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockManifestManager) DeserializeManifest(ctx context.Context, data []byte) (*ComposeManifest, error) {
	args := m.Called(ctx, data)
	return args.Get(0).(*ComposeManifest), args.Error(1)
}

type MockResourceManager struct {
	mock.Mock
}

func (m *MockResourceManager) CheckResourceAvailability(ctx context.Context, tenantID string, manifest *ComposeManifest) error {
	args := m.Called(ctx, tenantID, manifest)
	return args.Error(0)
}

func (m *MockResourceManager) AllocateResources(ctx context.Context, tenantID string, manifest *ComposeManifest) error {
	args := m.Called(ctx, tenantID, manifest)
	return args.Error(0)
}

func (m *MockResourceManager) ReleaseResources(ctx context.Context, tenantID string, manifest *ComposeManifest) error {
	args := m.Called(ctx, tenantID, manifest)
	return args.Error(0)
}

func (m *MockResourceManager) GetSystemResources(ctx context.Context) (*TenantResourceUsage, error) {
	args := m.Called(ctx)
	return args.Get(0).(*TenantResourceUsage), args.Error(1)
}

func (m *MockResourceManager) GetAvailableResources(ctx context.Context) (*TenantResourceUsage, error) {
	args := m.Called(ctx)
	return args.Get(0).(*TenantResourceUsage), args.Error(1)
}

func (m *MockResourceManager) GetTenantResourceUsage(ctx context.Context, tenantID string) (*TenantResourceUsage, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(*TenantResourceUsage), args.Error(1)
}

func (m *MockResourceManager) UpdateResourceUsage(ctx context.Context, tenantID string, usage *TenantResourceUsage) error {
	args := m.Called(ctx, tenantID, usage)
	return args.Error(0)
}

type MockEventManager struct {
	mock.Mock
}

func (m *MockEventManager) EmitEvent(ctx context.Context, event *DeploymentEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventManager) GetEvents(ctx context.Context, deploymentID string, limit int) ([]*DeploymentEvent, error) {
	args := m.Called(ctx, deploymentID, limit)
	return args.Get(0).([]*DeploymentEvent), args.Error(1)
}

func (m *MockEventManager) SubscribeToEvents(ctx context.Context, deploymentID string) (<-chan *DeploymentEvent, error) {
	args := m.Called(ctx, deploymentID)
	return args.Get(0).(<-chan *DeploymentEvent), args.Error(1)
}

func (m *MockEventManager) UnsubscribeFromEvents(ctx context.Context, deploymentID string) error {
	args := m.Called(ctx, deploymentID)
	return args.Error(0)
}

type MockComposeExecutor struct {
	mock.Mock
}

func (m *MockComposeExecutor) Up(ctx context.Context, deploymentID string, manifest *ComposeManifest, workDir string) error {
	args := m.Called(ctx, deploymentID, manifest, workDir)
	return args.Error(0)
}

func (m *MockComposeExecutor) Down(ctx context.Context, deploymentID string, workDir string) error {
	args := m.Called(ctx, deploymentID, workDir)
	return args.Error(0)
}

func (m *MockComposeExecutor) Restart(ctx context.Context, deploymentID string, workDir string) error {
	args := m.Called(ctx, deploymentID, workDir)
	return args.Error(0)
}

func (m *MockComposeExecutor) Status(ctx context.Context, deploymentID string, workDir string) (map[string]string, error) {
	args := m.Called(ctx, deploymentID, workDir)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockComposeExecutor) Logs(ctx context.Context, deploymentID string, serviceName string, workDir string, tail int) (io.ReadCloser, error) {
	args := m.Called(ctx, deploymentID, serviceName, workDir, tail)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockComposeExecutor) Scale(ctx context.Context, deploymentID string, serviceName string, replicas int, workDir string) error {
	args := m.Called(ctx, deploymentID, serviceName, replicas, workDir)
	return args.Error(0)
}

func (m *MockComposeExecutor) Pull(ctx context.Context, deploymentID string, manifest *ComposeManifest, workDir string) error {
	args := m.Called(ctx, deploymentID, manifest, workDir)
	return args.Error(0)
}

type MockStorageManager struct {
	mock.Mock
}

func (m *MockStorageManager) SaveDeployment(ctx context.Context, deployment *Deployment) error {
	args := m.Called(ctx, deployment)
	return args.Error(0)
}

func (m *MockStorageManager) LoadDeployment(ctx context.Context, deploymentID string) (*Deployment, error) {
	args := m.Called(ctx, deploymentID)
	return args.Get(0).(*Deployment), args.Error(1)
}

func (m *MockStorageManager) DeleteDeployment(ctx context.Context, deploymentID string) error {
	args := m.Called(ctx, deploymentID)
	return args.Error(0)
}

func (m *MockStorageManager) ListDeployments(ctx context.Context) ([]*Deployment, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*Deployment), args.Error(1)
}

func (m *MockStorageManager) SaveTenant(ctx context.Context, tenant *Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *MockStorageManager) LoadTenant(ctx context.Context, tenantID string) (*Tenant, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(*Tenant), args.Error(1)
}

func (m *MockStorageManager) DeleteTenant(ctx context.Context, tenantID string) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

func (m *MockStorageManager) ListTenants(ctx context.Context) ([]*Tenant, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*Tenant), args.Error(1)
}

// Tests
func TestNewService(t *testing.T) {
	logger := logrus.New()
	cfg := config.NewC(logger)

	service := NewService(cfg, logger)

	assert.NotNil(t, service)
	assert.Equal(t, cfg, service.cfg)
	assert.Equal(t, logger, service.logger)
	assert.NotNil(t, service.stopChan)
}

func TestCreateTenant(t *testing.T) {
	logger := logrus.New()
	cfg := config.NewC(logger)
	service := NewService(cfg, logger)

	// Create mocks
	mockTenantManager := &MockTenantManager{}
	mockStorageManager := &MockStorageManager{}
	mockResourceManager := &MockResourceManager{}

	// Set up expectations
	tenant := &Tenant{
		ID:          "test-tenant",
		Name:        "Test Tenant",
		Description: "Test tenant for testing",
	}

	mockTenantManager.On("CreateTenant", mock.Anything, tenant).Return(nil)
	mockStorageManager.On("SaveTenant", mock.Anything, tenant).Return(nil)
	mockResourceManager.On("UpdateResourceUsage", mock.Anything, tenant.ID, mock.Anything).Return(nil)

	// Set managers
	service.SetManagers(
		mockTenantManager,
		nil, // deploymentManager
		nil, // manifestManager
		nil, // portManager
		nil, // networkManager
		mockResourceManager,
		nil, // eventManager
		nil, // composeExecutor
		mockStorageManager,
	)

	// Test
	err := service.CreateTenant(context.Background(), tenant)

	assert.NoError(t, err)
	mockTenantManager.AssertExpectations(t)
	mockStorageManager.AssertExpectations(t)
	mockResourceManager.AssertExpectations(t)
}

func TestCreateDeployment(t *testing.T) {
	logger := logrus.New()
	cfg := config.NewC(logger)
	service := NewService(cfg, logger)

	// Create mocks
	mockManifestManager := &MockManifestManager{}
	mockResourceManager := &MockResourceManager{}
	mockPortManager := &MockPortManager{}
	mockStorageManager := &MockStorageManager{}
	mockEventManager := &MockEventManager{}

	// Set up test data
	manifest := &ComposeManifest{
		Version: "3.8",
		Services: map[string]ServiceConfig{
			"web": {
				Image: "nginx:alpine",
				Ports: []PortMapping{
					{ContainerPort: 80, Protocol: "tcp"},
				},
			},
		},
	}

	deployment := &Deployment{
		TenantID:    "test-tenant",
		Name:        "test-deployment",
		Description: "Test deployment",
		Manifest:    *manifest,
	}

	// Set up expectations
	mockManifestManager.On("ValidateManifest", mock.Anything, manifest).Return(nil)
	mockManifestManager.On("ProcessManifest", mock.Anything, manifest, deployment.TenantID).Return(manifest, nil)
	mockResourceManager.On("CheckResourceAvailability", mock.Anything, deployment.TenantID, manifest).Return(nil)
	mockPortManager.On("AllocatePort", mock.Anything, deployment.TenantID, 0).Return(8080, nil)
	mockStorageManager.On("SaveDeployment", mock.Anything, deployment).Return(nil)
	mockEventManager.On("EmitEvent", mock.Anything, mock.Anything).Return(nil)

	// Set managers
	service.SetManagers(
		nil, // tenantManager
		nil, // deploymentManager
		mockManifestManager,
		mockPortManager,
		nil, // networkManager
		mockResourceManager,
		mockEventManager,
		nil, // composeExecutor
		mockStorageManager,
	)

	// Test
	err := service.CreateDeployment(context.Background(), deployment)

	assert.NoError(t, err)
	assert.NotEmpty(t, deployment.ID)
	assert.Equal(t, string(DeploymentStatusPending), deployment.Status)

	mockManifestManager.AssertExpectations(t)
	mockResourceManager.AssertExpectations(t)
	mockPortManager.AssertExpectations(t)
	mockStorageManager.AssertExpectations(t)
	mockEventManager.AssertExpectations(t)
}

func TestStartDeployment(t *testing.T) {
	logger := logrus.New()
	cfg := config.NewC(logger)
	service := NewService(cfg, logger)

	// Create mocks
	mockResourceManager := &MockResourceManager{}
	mockComposeExecutor := &MockComposeExecutor{}
	mockStorageManager := &MockStorageManager{}
	mockEventManager := &MockEventManager{}

	// Set up test data
	deployment := &Deployment{
		ID:       "test-deployment",
		TenantID: "test-tenant",
		Name:     "test-deployment",
		Status:   string(DeploymentStatusPending),
		Manifest: ComposeManifest{
			Version: "3.8",
			Services: map[string]ServiceConfig{
				"web": {
					Image: "nginx:alpine",
				},
			},
		},
	}

	// Add deployment to active deployments
	service.mu.Lock()
	service.activeDeployments[deployment.ID] = deployment
	service.mu.Unlock()

	// Set up expectations
	mockResourceManager.On("AllocateResources", mock.Anything, deployment.TenantID, &deployment.Manifest).Return(nil)
	mockComposeExecutor.On("Up", mock.Anything, deployment.ID, &deployment.Manifest, mock.Anything).Return(nil)
	mockStorageManager.On("SaveDeployment", mock.Anything, deployment).Return(nil)
	mockEventManager.On("EmitEvent", mock.Anything, mock.Anything).Return(nil)

	// Set managers
	service.SetManagers(
		nil, // tenantManager
		nil, // deploymentManager
		nil, // manifestManager
		nil, // portManager
		nil, // networkManager
		mockResourceManager,
		mockEventManager,
		mockComposeExecutor,
		mockStorageManager,
	)

	// Test
	err := service.StartDeployment(context.Background(), deployment.ID)

	assert.NoError(t, err)
	assert.Equal(t, string(DeploymentStatusRunning), deployment.Status)

	mockResourceManager.AssertExpectations(t)
	mockComposeExecutor.AssertExpectations(t)
	mockStorageManager.AssertExpectations(t)
	mockEventManager.AssertExpectations(t)
}

func TestGetDeployment(t *testing.T) {
	logger := logrus.New()
	cfg := config.NewC(logger)
	service := NewService(cfg, logger)

	// Create mocks
	mockStorageManager := &MockStorageManager{}

	// Set up test data
	deployment := &Deployment{
		ID:       "test-deployment",
		TenantID: "test-tenant",
		Name:     "test-deployment",
	}

	// Test case 1: Deployment in active deployments
	service.mu.Lock()
	service.activeDeployments[deployment.ID] = deployment
	service.mu.Unlock()

	result, err := service.GetDeployment(context.Background(), deployment.ID)

	assert.NoError(t, err)
	assert.Equal(t, deployment, result)

	// Test case 2: Deployment not in active deployments, load from storage
	service.mu.Lock()
	delete(service.activeDeployments, deployment.ID)
	service.mu.Unlock()

	mockStorageManager.On("LoadDeployment", mock.Anything, deployment.ID).Return(deployment, nil)

	service.SetManagers(
		nil, // tenantManager
		nil, // deploymentManager
		nil, // manifestManager
		nil, // portManager
		nil, // networkManager
		nil, // resourceManager
		nil, // eventManager
		nil, // composeExecutor
		mockStorageManager,
	)

	result, err = service.GetDeployment(context.Background(), deployment.ID)

	assert.NoError(t, err)
	assert.Equal(t, deployment, result)

	mockStorageManager.AssertExpectations(t)
}

func TestGetTenantResourceUsage(t *testing.T) {
	logger := logrus.New()
	cfg := config.NewC(logger)
	service := NewService(cfg, logger)

	// Create mocks
	mockResourceManager := &MockResourceManager{}

	// Set up test data
	usage := &TenantResourceUsage{
		TenantID:       "test-tenant",
		CPUUsage:       2.5,
		MemoryUsage:    1024 * 1024 * 1024,      // 1GB
		DiskUsage:      10 * 1024 * 1024 * 1024, // 10GB
		NetworkUsage:   100 * 1024 * 1024,       // 100MB
		PortCount:      5,
		ContainerCount: 3,
	}

	// Set up expectations
	mockResourceManager.On("GetTenantResourceUsage", mock.Anything, "test-tenant").Return(usage, nil)

	// Set managers
	service.SetManagers(
		nil, // tenantManager
		nil, // deploymentManager
		nil, // manifestManager
		nil, // portManager
		nil, // networkManager
		mockResourceManager,
		nil, // eventManager
		nil, // composeExecutor
		nil, // storageManager
	)

	// Test
	result, err := service.GetTenantResourceUsage(context.Background(), "test-tenant")

	assert.NoError(t, err)
	assert.Equal(t, usage, result)

	mockResourceManager.AssertExpectations(t)
}
