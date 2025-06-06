package deployment

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/manifest"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
)

func TestDeploymentStore(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create temporary directory for store
	storeDir, err := os.MkdirTemp("", "deployment-store-test")
	require.NoError(t, err)
	defer os.RemoveAll(storeDir)

	// Create deployment store
	store, err := NewDeploymentStore(&DeploymentStoreConfig{
		StoreDir:   storeDir,
		MaxRetries: 3,
	})
	require.NoError(t, err)
	require.NotNil(t, store)

	// Create test deployment
	dep := &types.ManagedDeployment{
		ID:           "test-deployment",
		LeaseID:      "test-lease",
		Requester:    common.HexToAddress("0x123"),
		Status:       types.DeploymentStatusPending,
		HealthStatus: types.HealthStatusUnknown,
		Version:      1,
		Resources: manifest.ResourceRequirements{
			CPU: manifest.ResourceLimit{
				Request: "1",
			},
			Memory: manifest.ResourceLimit{
				Request: "1Gi",
			},
			Storage: []manifest.StorageVolume{
				{
					Name: "data",
					Size: "10Gi",
				},
			},
		},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		LastHealth: time.Now(),
	}

	t.Run("Store and Retrieve", func(t *testing.T) {
		select {
		case <-ctx.Done():
			t.Fatal("test timed out")
		default:
			// Test storing deployment
			err = store.StoreDeployment(dep)
			require.NoError(t, err)

			// Test getting deployment
			retrieved, err := store.GetDeployment(dep.ID)
			require.NoError(t, err)
			assert.Equal(t, dep.ID, retrieved.ID)
			assert.Equal(t, dep.LeaseID, retrieved.LeaseID)
			assert.Equal(t, dep.Requester, retrieved.Requester)
			assert.Equal(t, dep.Status, retrieved.Status)
			assert.Equal(t, dep.HealthStatus, retrieved.HealthStatus)
			assert.Equal(t, dep.Version, retrieved.Version)
			assert.Equal(t, dep.Resources.CPU.Request, retrieved.Resources.CPU.Request)
			assert.Equal(t, dep.Resources.Memory.Request, retrieved.Resources.Memory.Request)
			assert.Equal(t, len(dep.Resources.Storage), len(retrieved.Resources.Storage))
			if len(dep.Resources.Storage) > 0 {
				assert.Equal(t, dep.Resources.Storage[0].Name, retrieved.Resources.Storage[0].Name)
				assert.Equal(t, dep.Resources.Storage[0].Size, retrieved.Resources.Storage[0].Size)
			}
		}
	})

	t.Run("List Deployments", func(t *testing.T) {
		// Test listing deployments
		deployments, err := store.ListDeployments()
		require.NoError(t, err)
		assert.Len(t, deployments, 1)
		assert.Equal(t, dep.ID, deployments[0].ID)
	})

	t.Run("Update Deployment", func(t *testing.T) {
		// Test updating deployment
		dep.Status = types.DeploymentStatusRunning
		dep.UpdatedAt = time.Now()
		err = store.UpdateDeployment(dep)
		require.NoError(t, err)

		// Verify update
		retrieved, err := store.GetDeployment(dep.ID)
		require.NoError(t, err)
		assert.Equal(t, types.DeploymentStatusRunning, retrieved.Status)
	})

	t.Run("Get Deployments By Status", func(t *testing.T) {
		// Test getting deployments by status
		running, err := store.GetDeploymentsByStatus(types.DeploymentStatusRunning)
		require.NoError(t, err)
		assert.Len(t, running, 1)
		assert.Equal(t, dep.ID, running[0].ID)
	})

	t.Run("Health Status Management", func(t *testing.T) {
		// Test health status management
		health, err := store.GetDeploymentHealth(dep.ID)
		require.NoError(t, err)
		assert.Equal(t, types.HealthStatusUnknown, health)

		err = store.UpdateDeploymentHealth(dep.ID, types.HealthStatusHealthy)
		require.NoError(t, err)

		health, err = store.GetDeploymentHealth(dep.ID)
		require.NoError(t, err)
		assert.Equal(t, types.HealthStatusHealthy, health)
	})

	t.Run("Error Management", func(t *testing.T) {
		// Test error management
		errMsg, err := store.GetDeploymentError(dep.ID)
		require.NoError(t, err)
		assert.Empty(t, errMsg)

		testErr := assert.AnError
		err = store.UpdateDeploymentError(dep.ID, testErr)
		require.NoError(t, err)

		errMsg, err = store.GetDeploymentError(dep.ID)
		require.NoError(t, err)
		assert.Equal(t, testErr.Error(), errMsg)
	})

	t.Run("Delete Deployment", func(t *testing.T) {
		// Test deleting deployment
		err = store.DeleteDeployment(dep.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = store.GetDeployment(dep.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read deployment file")
	})
}

func TestDeploymentStoreConcurrent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create temporary directory for store
	storeDir, err := os.MkdirTemp("", "deployment-store-concurrent-test")
	require.NoError(t, err)
	defer os.RemoveAll(storeDir)

	// Create deployment store
	store, err := NewDeploymentStore(&DeploymentStoreConfig{
		StoreDir:   storeDir,
		MaxRetries: 3,
	})
	require.NoError(t, err)
	require.NotNil(t, store)

	// Reduce number of concurrent operations
	numDeployments := 5
	deployments := make([]*types.ManagedDeployment, numDeployments)
	for i := 0; i < numDeployments; i++ {
		deployments[i] = &types.ManagedDeployment{
			ID:           fmt.Sprintf("test-deployment-%d", i),
			LeaseID:      fmt.Sprintf("test-lease-%d", i),
			Requester:    common.HexToAddress("0x123"),
			Status:       types.DeploymentStatusPending,
			HealthStatus: types.HealthStatusUnknown,
			Version:      1,
			Resources: manifest.ResourceRequirements{
				CPU: manifest.ResourceLimit{
					Request: "1",
				},
				Memory: manifest.ResourceLimit{
					Request: "1Gi",
				},
				Storage: []manifest.StorageVolume{
					{
						Name: "data",
						Size: "10Gi",
					},
				},
			},
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			LastHealth: time.Now(),
		}
	}

	// Test concurrent store operations with context
	var wg sync.WaitGroup
	errCh := make(chan error, numDeployments)

	for i := 0; i < numDeployments; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				errCh <- fmt.Errorf("test timed out")
				return
			default:
				// Store deployment
				if err := store.StoreDeployment(deployments[i]); err != nil {
					errCh <- err
					return
				}

				// Get deployment
				if _, err := store.GetDeployment(deployments[i].ID); err != nil {
					errCh <- err
					return
				}

				// Update deployment
				deployments[i].Status = types.DeploymentStatusRunning
				if err := store.UpdateDeployment(deployments[i]); err != nil {
					errCh <- err
					return
				}

				// Delete deployment
				if err := store.DeleteDeployment(deployments[i].ID); err != nil {
					errCh <- err
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		t.Fatal("test timed out")
	case <-done:
		// Check for errors
		close(errCh)
		for err := range errCh {
			if err != nil {
				t.Error(err)
			}
		}
	}

	// Verify all deployments are deleted
	list, err := store.ListDeployments()
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestDeploymentStoreInvalid(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create temporary directory for store
	storeDir, err := os.MkdirTemp("", "deployment-store-invalid-test")
	require.NoError(t, err)
	defer os.RemoveAll(storeDir)

	// Create deployment store
	store, err := NewDeploymentStore(&DeploymentStoreConfig{
		StoreDir:   storeDir,
		MaxRetries: 3,
	})
	require.NoError(t, err)
	require.NotNil(t, store)

	select {
	case <-ctx.Done():
		t.Fatal("test timed out")
	default:
		// Test invalid deployment ID
		_, err = store.GetDeployment("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read deployment file")

		// Test invalid deployment data
		filePath := filepath.Join(storeDir, "invalid.json")
		err = os.WriteFile(filePath, []byte("invalid json"), 0644)
		require.NoError(t, err)

		_, err = store.GetDeployment("invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal deployment")
	}
}
