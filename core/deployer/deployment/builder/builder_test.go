package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestBuildNamespace(t *testing.T) {
	// Create test settings
	settings := Settings{
		Client: fake.NewSimpleClientset(),
		Logger: zap.NewNop(),
	}

	// Create test deployment
	deployment := &types.ManagedDeployment{
		ID:        "test-deployment",
		Namespace: "test-namespace",
		Name:      "test-name",
	}

	// Build namespace
	nsBuilder := BuildNamespace(settings, deployment, settings.Client)
	require.NotNil(t, nsBuilder)

	// Create namespace
	ns, err := nsBuilder.Create()
	require.NoError(t, err)
	require.NotNil(t, ns)

	// Verify namespace properties
	assert.Equal(t, deployment.Namespace, ns.Name)
	assert.Equal(t, map[string]string{
		"app": deployment.Name,
	}, ns.Labels)
}

func TestBuildNetPol(t *testing.T) {
	// Create test settings
	settings := Settings{
		Client: fake.NewSimpleClientset(),
		Logger: zap.NewNop(),
	}

	// Create test deployment
	deployment := &types.ManagedDeployment{
		ID:        "test-deployment",
		Namespace: "test-namespace",
		Name:      "test-name",
	}

	// Build network policy
	netPolBuilder := BuildNetPol(settings, deployment)
	require.NotNil(t, netPolBuilder)

	// Create network policy
	netPol, err := netPolBuilder.Create()
	require.NoError(t, err)
	require.NotNil(t, netPol)

	// Verify network policy properties
	assert.Equal(t, deployment.Name, netPol.Name)
	assert.Equal(t, deployment.Namespace, netPol.Namespace)
	assert.Equal(t, map[string]string{
		"app": deployment.Name,
	}, netPol.Labels)
}

func TestBuildService(t *testing.T) {
	settings := Settings{
		DeploymentServiceType: corev1.ServiceTypeClusterIP,
	}
	deployment := &types.ManagedDeployment{
		Name:      "test-deployment",
		Namespace: "test-namespace",
	}

	svcBuilder := BuildService(settings, deployment, 0, 0, false)
	if svcBuilder == nil {
		t.Fatal("service builder is nil")
	}

	if svcBuilder.Name() != "test-deployment-service-0" {
		t.Errorf("expected service name to be test-deployment-service-0, got %s", svcBuilder.Name())
	}

	if svcBuilder.NS() != "test-namespace" {
		t.Errorf("expected namespace to be test-namespace, got %s", svcBuilder.NS())
	}
}

func TestBuildStatefulSet(t *testing.T) {
	// Create test settings
	settings := Settings{
		Client: fake.NewSimpleClientset(),
		Logger: zap.NewNop(),
	}

	// Create test deployment
	deployment := &types.ManagedDeployment{
		ID:        "test-deployment",
		Namespace: "test-namespace",
		Name:      "test-name",
	}

	// Build statefulset
	stsBuilder := BuildStatefulSet(settings, deployment)
	require.NotNil(t, stsBuilder)

	// Create statefulset
	sts, err := stsBuilder.Create()
	require.NoError(t, err)
	require.NotNil(t, sts)

	// Verify statefulset properties
	assert.Equal(t, deployment.Name, sts.Name)
	assert.Equal(t, deployment.Namespace, sts.Namespace)
	assert.Equal(t, map[string]string{
		"app": deployment.Name,
	}, sts.Labels)
}

func TestBuildDeployment(t *testing.T) {
	// Create test settings
	settings := Settings{
		Client: fake.NewSimpleClientset(),
		Logger: zap.NewNop(),
	}

	// Create test deployment
	deployment := &types.ManagedDeployment{
		ID:        "test-deployment",
		Namespace: "test-namespace",
		Name:      "test-name",
	}

	// Build deployment
	depBuilder := BuildDeployment(settings, deployment)
	require.NotNil(t, depBuilder)

	// Create deployment
	dep, err := depBuilder.Create()
	require.NoError(t, err)
	require.NotNil(t, dep)

	// Verify deployment properties
	assert.Equal(t, deployment.Name, dep.Name)
	assert.Equal(t, deployment.Namespace, dep.Namespace)
	assert.Equal(t, map[string]string{
		"app": deployment.Name,
	}, dep.Labels)
}

func TestBuildServiceCredentials(t *testing.T) {
	// Create test settings
	settings := Settings{
		Client: fake.NewSimpleClientset(),
		Logger: zap.NewNop(),
	}

	// Create test deployment
	deployment := &types.ManagedDeployment{
		ID:        "test-deployment",
		Namespace: "test-namespace",
		Name:      "test-name",
	}

	// Build service credentials
	credsBuilder := BuildServiceCredentials(settings, deployment)
	require.NotNil(t, credsBuilder)

	// Create service credentials
	creds, err := credsBuilder.Create()
	require.NoError(t, err)
	require.NotNil(t, creds)

	// Verify service credentials properties
	assert.Equal(t, deployment.Name+"-sa", creds.Name)
	assert.Equal(t, deployment.Namespace, creds.Namespace)
	assert.Equal(t, map[string]string{
		"app": deployment.Name,
	}, creds.Labels)
}
