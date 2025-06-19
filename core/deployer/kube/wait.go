package kube

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// WaitForDeployment waits for all services in a deployment to be ready using the Watch API
func (c *KubeClient) WaitForDeployment(ctx context.Context, deploymentID string, timeout time.Duration) error {
	c.Logger.WithField("deploymentID", deploymentID).Debug("Waiting for deployment to be ready")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	watcher, err := c.Client.AppsV1().Deployments(deploymentID).Watch(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}
	defer watcher.Stop()

	ready := make(map[string]bool)
	var total int

	// Get the initial set of deployments and their state
	deployments, err := c.Client.AppsV1().Deployments(deploymentID).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}
	total = len(deployments.Items)
	for _, d := range deployments.Items {
		ready[d.Name] = isDeploymentReady(&d)
	}
	if total == 0 {
		return fmt.Errorf("no deployments found for deploymentID=%s", deploymentID)
	}
	// Check if already ready
	allReady := true
	for _, r := range ready {
		if !r {
			allReady = false
			break
		}
	}
	if allReady {
		c.Logger.WithField("deploymentID", deploymentID).Debug("All deployments are ready")
		return nil
	}

	// Now watch for changes
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("deployment %s failed to become ready within %v: %w", deploymentID, timeout, ctx.Err())
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return fmt.Errorf("watch channel closed unexpectedly")
			}
			dep, ok := event.Object.(*appsv1.Deployment)
			if !ok {
				continue
			}
			c.Logger.WithField("deploymentID", deploymentID).
				WithField("eventType", event.Type).
				WithField("deployment", dep.Name).
				WithField("readyReplicas", dep.Status.ReadyReplicas).
				WithField("desiredReplicas", dep.Status.Replicas).
				Debug("received deployment event")
			ready[dep.Name] = isDeploymentReady(dep)
			allReady := true
			for _, r := range ready {
				if !r {
					allReady = false
					break
				}
			}
			if allReady && len(ready) == total {
				c.Logger.WithField("deploymentID", deploymentID).Debug("All deployments are ready")
				return nil
			}
		}
	}
}

// isDeploymentReady checks if a deployment is ready
func isDeploymentReady(deployment *appsv1.Deployment) bool {
	return deployment.Status.ReadyReplicas == deployment.Status.Replicas &&
		deployment.Status.Replicas > 0 &&
		deployment.Status.UpdatedReplicas == deployment.Status.Replicas &&
		deployment.Status.AvailableReplicas == deployment.Status.Replicas
}

// WaitForNamespaceTermination waits for a namespace to be fully deleted
func (c *KubeClient) WaitForNamespaceTermination(ctx context.Context, namespace string, timeout time.Duration) error {
	c.Logger.WithField("namespace", namespace).Debug("Waiting for namespace termination")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Check if namespace is already gone
	ns, err := c.Client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			c.Logger.WithField("namespace", namespace).Debug("namespace already deleted")
			return nil
		}
		return fmt.Errorf("failed to get namespace: %w", err)
	}
	c.Logger.WithField("namespace", namespace).WithField("phase", ns.Status.Phase).Debug("namespace found, waiting for deletion")

	watcher, err := c.Client.CoreV1().Namespaces().Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", namespace),
	})
	if err != nil {
		return fmt.Errorf("failed to start namespace watcher: %w", err)
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("namespace %s was not terminated within %v: %w", namespace, timeout, ctx.Err())
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return fmt.Errorf("namespace watch channel closed unexpectedly")
			}
			c.Logger.WithField("namespace", namespace).
				WithField("eventType", event.Type).
				Debug("received namespace event (termination wait)")
			if event.Type == watch.Deleted {
				c.Logger.WithField("namespace", namespace).Debug("namespace deleted")
				return nil
			}
		}
	}
}
