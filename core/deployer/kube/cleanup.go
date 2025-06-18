package kube

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *KubeClient) CleanupResources(ctx context.Context, deploymentID string) error {
	c.Logger.WithField("deploymentID", deploymentID).Info("cleaning up resources for deployment")

	// List and delete all resources with the deploymentID label
	// We'll use the deploymentID as the namespace name as well
	namespace := deploymentID

	// Delete Deployments
	deployments, err := c.Client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		c.Logger.Error("failed to list deployments for cleanup", err)
	} else {
		for _, deployment := range deployments.Items {
			c.Logger.WithField("name", deployment.Name).Info("deleting deployment")
			err := c.Client.AppsV1().Deployments(namespace).Delete(ctx, deployment.Name, metav1.DeleteOptions{})
			if err != nil {
				c.Logger.Error("failed to delete deployment",
					"name", deployment.Name,
					"error", err)
			}
		}
	}

	// Delete Services
	services, err := c.Client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		c.Logger.Error("failed to list services for cleanup", err)
	} else {
		for _, service := range services.Items {
			c.Logger.WithField("name", service.Name).Info("deleting service")
			err := c.Client.CoreV1().Services(namespace).Delete(ctx, service.Name, metav1.DeleteOptions{})
			if err != nil {
				c.Logger.Error("failed to delete service",
					"name", service.Name,
					"error", err)
			}
		}
	}

	// Delete Ingresses
	ingresses, err := c.Client.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		c.Logger.Error("failed to list ingresses for cleanup", err)
	} else {
		for _, ingress := range ingresses.Items {
			c.Logger.WithField("name", ingress.Name).Info("deleting ingress")
			err := c.Client.NetworkingV1().Ingresses(namespace).Delete(ctx, ingress.Name, metav1.DeleteOptions{})
			if err != nil {
				c.Logger.WithField("name", ingress.Name).Error("failed to delete ingress", err)
			}
		}
	}

	// Delete Secrets (registry auth secrets)
	secrets, err := c.Client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		c.Logger.Error("failed to list secrets for cleanup", err)
	} else {
		for _, secret := range secrets.Items {
			c.Logger.WithField("name", secret.Name).Info("deleting secret")
			err := c.Client.CoreV1().Secrets(namespace).Delete(ctx, secret.Name, metav1.DeleteOptions{})
			if err != nil {
				c.Logger.WithField("name", secret.Name).Error("failed to delete secret", err)
			}
		}
	}

	// Also delete registry auth secrets by name pattern
	allSecrets, err := c.Client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		c.Logger.Error("failed to list all secrets for cleanup", err)
	} else {
		for _, secret := range allSecrets.Items {
			if strings.HasPrefix(secret.Name, "registry-auth-") {
				c.Logger.WithField("name", secret.Name).Info("deleting registry secret")
				err := c.Client.CoreV1().Secrets(namespace).Delete(ctx, secret.Name, metav1.DeleteOptions{})
				if err != nil {
					c.Logger.WithField("name", secret.Name).Error("failed to delete registry secret", err)
				}
			}
		}
	}

	// Delete PersistentVolumeClaims
	pvcs, err := c.Client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		c.Logger.Error("failed to list PVCs for cleanup", err)
	} else {
		for _, pvc := range pvcs.Items {
			c.Logger.WithField("name", pvc.Name).Info("deleting PVC")
			err := c.Client.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, pvc.Name, metav1.DeleteOptions{})
			if err != nil {
				c.Logger.WithField("name", pvc.Name).Error("failed to delete PVC", err)
			}
		}
	}

	// Also delete PVCs by name pattern (storage volumes)
	allPVCs, err := c.Client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		c.Logger.Error("failed to list all PVCs for cleanup", err)
	} else {
		for _, pvc := range allPVCs.Items {
			if strings.HasPrefix(pvc.Name, "storage-") {
				c.Logger.WithField("name", pvc.Name).Info("deleting storage PVC")
				err := c.Client.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, pvc.Name, metav1.DeleteOptions{})
				if err != nil {
					c.Logger.WithField("name", pvc.Name).Error("failed to delete storage PVC", err)
				}
			}
		}
	}

	// Delete Pods (in case there are orphaned pods)
	pods, err := c.Client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("deploymentID=%s", deploymentID),
	})
	if err != nil {
		c.Logger.Error("failed to list pods for cleanup", err)
	} else {
		for _, pod := range pods.Items {
			c.Logger.WithField("name", pod.Name).Info("deleting pod")
			err := c.Client.CoreV1().Pods(namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
			if err != nil {
				c.Logger.WithField("name", pod.Name).Error("failed to delete pod", err)
			}
		}
	}

	// Finally, delete the namespace itself
	c.Logger.WithField("namespace", namespace).Info("deleting namespace")
	err = c.Client.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			c.Logger.WithField("namespace", namespace).Info("namespace already deleted")
		} else {
			c.Logger.WithField("namespace", namespace).Error("failed to delete namespace", err)
			return fmt.Errorf("failed to delete namespace %s: %w", namespace, err)
		}
	}

	c.Logger.WithField("deploymentID", deploymentID).Info("cleanup completed for deployment")
	return nil
}
