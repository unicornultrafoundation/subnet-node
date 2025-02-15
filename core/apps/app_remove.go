package apps

import (
	"context"
	"fmt"
	"math/big"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/errdefs"
)

// Stops and removes the container for the specified app.
func (s *Service) RemoveApp(ctx context.Context, appId *big.Int) (*App, error) {
	// Set the namespace for the container
	ctx = namespaces.WithNamespace(ctx, NAMESPACE)

	// Retrieve app details from the Ethereum contract
	app, err := s.GetApp(ctx, appId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch app details: %w", err)
	}

	// Load the container for the app
	container, err := s.containerdClient.LoadContainer(ctx, app.ContainerId())
	if err != nil {
		return nil, fmt.Errorf("failed to load container: %w", err)
	}

	// Stop the container task
	task, err := container.Task(ctx, nil)
	taskNotFound := false
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return nil, fmt.Errorf("failed to get task: %w", err)
		}
		taskNotFound = true
	}

	if !taskNotFound {
		_, err = task.Delete(ctx, containerd.WithProcessKill)
		if err != nil {
			return nil, fmt.Errorf("failed to delete task: %w", err)
		}
	}

	// Remove the container and its snapshot
	err = container.Delete(ctx, containerd.WithSnapshotCleanup)
	if err != nil {
		return nil, fmt.Errorf("failed to delete container: %w", err)
	}

	log.Printf("App %s removed successfully: Container ID: %s", app.Symbol, app.ContainerId())
	app.Status = NotFound
	return app, nil
}
