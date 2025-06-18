package kube

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *KubeClient) CleanupResources(ctx context.Context, deploymentID string) error {
	c.Logger.Info("cleaning up resources for deployment", zap.String("deploymentID", deploymentID))

	// List and delete all resources with the deploymentID label
	// We'll use the deploymentID as the namespace name as well
	namespace := deploymentID

	// Delete Deployments
	deployments, err := c.Client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		c.Logger.Error("failed to list deployments for cleanup", zap.Error(err))
	} else {
		for _, deployment := range deployments.Items {
			c.Logger.Info("deleting deployment", zap.String("name", deployment.Name))
			err := c.Client.AppsV1().Deployments(namespace).Delete(ctx, deployment.Name, metav1.DeleteOptions{})
			if err != nil {
				c.Logger.Error("failed to delete deployment",
					zap.String("name", deployment.Name),
					zap.Error(err))
			}
		}
	}

	// Delete Services
	services, err := c.Client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		c.Logger.Error("failed to list services for cleanup", zap.Error(err))
	} else {
		for _, service := range services.Items {
			c.Logger.Info("deleting service", zap.String("name", service.Name))
			err := c.Client.CoreV1().Services(namespace).Delete(ctx, service.Name, metav1.DeleteOptions{})
			if err != nil {
				c.Logger.Error("failed to delete service",
					zap.String("name", service.Name),
					zap.Error(err))
			}
		}
	}

	// Delete Ingresses
	ingresses, err := c.Client.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		c.Logger.Error("failed to list ingresses for cleanup", zap.Error(err))
	} else {
		for _, ingress := range ingresses.Items {
			c.Logger.Info("deleting ingress", zap.String("name", ingress.Name))
			err := c.Client.NetworkingV1().Ingresses(namespace).Delete(ctx, ingress.Name, metav1.DeleteOptions{})
			if err != nil {
				c.Logger.Error("failed to delete ingress",
					zap.String("name", ingress.Name),
					zap.Error(err))
			}
		}
	}

	// Delete Secrets (registry auth secrets)
	secrets, err := c.Client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		c.Logger.Error("failed to list secrets for cleanup", zap.Error(err))
	} else {
		for _, secret := range secrets.Items {
			c.Logger.Info("deleting secret", zap.String("name", secret.Name))
			err := c.Client.CoreV1().Secrets(namespace).Delete(ctx, secret.Name, metav1.DeleteOptions{})
			if err != nil {
				c.Logger.Error("failed to delete secret",
					zap.String("name", secret.Name),
					zap.Error(err))
			}
		}
	}

	// Also delete registry auth secrets by name pattern
	allSecrets, err := c.Client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		c.Logger.Error("failed to list all secrets for cleanup", zap.Error(err))
	} else {
		for _, secret := range allSecrets.Items {
			if strings.HasPrefix(secret.Name, "registry-auth-") {
				c.Logger.Info("deleting registry secret", zap.String("name", secret.Name))
				err := c.Client.CoreV1().Secrets(namespace).Delete(ctx, secret.Name, metav1.DeleteOptions{})
				if err != nil {
					c.Logger.Error("failed to delete registry secret",
						zap.String("name", secret.Name),
						zap.Error(err))
				}
			}
		}
	}

	// Delete PersistentVolumeClaims
	pvcs, err := c.Client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		c.Logger.Error("failed to list PVCs for cleanup", zap.Error(err))
	} else {
		for _, pvc := range pvcs.Items {
			c.Logger.Info("deleting PVC", zap.String("name", pvc.Name))
			err := c.Client.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, pvc.Name, metav1.DeleteOptions{})
			if err != nil {
				c.Logger.Error("failed to delete PVC",
					zap.String("name", pvc.Name),
					zap.Error(err))
			}
		}
	}

	// Also delete PVCs by name pattern (storage volumes)
	allPVCs, err := c.Client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		c.Logger.Error("failed to list all PVCs for cleanup", zap.Error(err))
	} else {
		for _, pvc := range allPVCs.Items {
			if strings.HasPrefix(pvc.Name, "storage-") {
				c.Logger.Info("deleting storage PVC", zap.String("name", pvc.Name))
				err := c.Client.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, pvc.Name, metav1.DeleteOptions{})
				if err != nil {
					c.Logger.Error("failed to delete storage PVC",
						zap.String("name", pvc.Name),
						zap.Error(err))
				}
			}
		}
	}

	// Delete Pods (in case there are orphaned pods)
	pods, err := c.Client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		c.Logger.Error("failed to list pods for cleanup", zap.Error(err))
	} else {
		for _, pod := range pods.Items {
			c.Logger.Info("deleting pod", zap.String("name", pod.Name))
			err := c.Client.CoreV1().Pods(namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
			if err != nil {
				c.Logger.Error("failed to delete pod",
					zap.String("name", pod.Name),
					zap.Error(err))
			}
		}
	}

	// Finally, delete the namespace itself
	c.Logger.Info("deleting namespace", zap.String("namespace", namespace))
	err = c.Client.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			c.Logger.Info("namespace already deleted", zap.String("namespace", namespace))
		} else {
			c.Logger.Error("failed to delete namespace",
				zap.String("namespace", namespace),
				zap.Error(err))
			return fmt.Errorf("failed to delete namespace %s: %w", namespace, err)
		}
	}

	c.Logger.Info("cleanup completed for deployment", zap.String("deploymentID", deploymentID))
	return nil
}
