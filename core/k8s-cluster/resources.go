package k8scluster

import (
	"context"
	"fmt"

	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/manifest"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// calculateResourceRequirements calculates the total resource requirements for a deployment
func (s *Service) calculateResourceRequirements(sdl *manifest.SDL) (manifest.ResourceRequirements, error) {
	var totalResources manifest.ResourceRequirements

	// Initialize resource maps
	totalResources.CPU = manifest.ResourceLimit{}
	totalResources.Memory = manifest.ResourceLimit{}
	totalResources.Storage = make([]manifest.StorageVolume, 0)

	// Calculate resources for each deployment group
	for _, deployment := range sdl.Deployment {
		computeProfile, ok := sdl.Profiles.Compute[deployment.Profile]
		if !ok {
			return manifest.ResourceRequirements{}, fmt.Errorf("compute profile %s not found", deployment.Profile)
		}

		// Multiply resources by deployment count
		count := int64(deployment.Count)

		// Add CPU resources
		if computeProfile.Resources.CPU.Request != "" {
			cpu, err := resource.ParseQuantity(computeProfile.Resources.CPU.Request)
			if err != nil {
				return manifest.ResourceRequirements{}, fmt.Errorf("invalid CPU request: %w", err)
			}
			cpuValue := cpu.Value() * count
			totalResources.CPU.Request = fmt.Sprintf("%dm", cpuValue)
		}
		if computeProfile.Resources.CPU.Limit != "" {
			cpu, err := resource.ParseQuantity(computeProfile.Resources.CPU.Limit)
			if err != nil {
				return manifest.ResourceRequirements{}, fmt.Errorf("invalid CPU limit: %w", err)
			}
			cpuValue := cpu.Value() * count
			totalResources.CPU.Limit = fmt.Sprintf("%dm", cpuValue)
		}

		// Add Memory resources
		if computeProfile.Resources.Memory.Request != "" {
			mem, err := resource.ParseQuantity(computeProfile.Resources.Memory.Request)
			if err != nil {
				return manifest.ResourceRequirements{}, fmt.Errorf("invalid memory request: %w", err)
			}
			memValue := mem.Value() * count
			totalResources.Memory.Request = fmt.Sprintf("%dMi", memValue)
		}
		if computeProfile.Resources.Memory.Limit != "" {
			mem, err := resource.ParseQuantity(computeProfile.Resources.Memory.Limit)
			if err != nil {
				return manifest.ResourceRequirements{}, fmt.Errorf("invalid memory limit: %w", err)
			}
			memValue := mem.Value() * count
			totalResources.Memory.Limit = fmt.Sprintf("%dMi", memValue)
		}

		// Add Storage resources
		for _, storage := range computeProfile.Resources.Storage {
			storageCopy := storage
			if storage.Size != "" {
				size, err := resource.ParseQuantity(storage.Size)
				if err != nil {
					return manifest.ResourceRequirements{}, fmt.Errorf("invalid storage size: %w", err)
				}
				sizeValue := size.Value() * count
				storageCopy.Size = fmt.Sprintf("%dGi", sizeValue)
			}
			totalResources.Storage = append(totalResources.Storage, storageCopy)
		}

		// Add GPU resources if present
		if computeProfile.Resources.GPU != nil {
			if totalResources.GPU == nil {
				totalResources.GPU = &manifest.GPUResource{
					Units:      computeProfile.Resources.GPU.Units * int32(count),
					Attributes: computeProfile.Resources.GPU.Attributes,
				}
			} else {
				totalResources.GPU.Units += computeProfile.Resources.GPU.Units * int32(count)
			}
		}
	}

	return totalResources, nil
}

// cleanupResources cleans up resources for a deployment
func (s *Service) cleanupResources(ctx context.Context, dep *types.ManagedDeployment) error {
	// Get Kubernetes client
	clientset, err := kubernetes.NewForConfig(s.config.KubeConfig)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Get namespace
	namespace := dep.Namespace

	// Delete all config maps
	configMaps, err := clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list config maps: %w", err)
	}
	for _, cm := range configMaps.Items {
		if cm.Name == "manifest-default" {
			// Skip the manifest-default ConfigMap
			continue
		}
		if err := clientset.CoreV1().ConfigMaps(namespace).Delete(ctx, cm.Name, metav1.DeleteOptions{}); err != nil {
			return fmt.Errorf("failed to delete config map %s: %w", cm.Name, err)
		}
	}

	// Delete all secrets
	secrets, err := clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}
	for _, secret := range secrets.Items {
		if err := clientset.CoreV1().Secrets(namespace).Delete(ctx, secret.Name, metav1.DeleteOptions{}); err != nil {
			return fmt.Errorf("failed to delete secret %s: %w", secret.Name, err)
		}
	}

	// Delete all persistent volume claims
	pvcs, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list persistent volume claims: %w", err)
	}
	for _, pvc := range pvcs.Items {
		if err := clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, pvc.Name, metav1.DeleteOptions{}); err != nil {
			return fmt.Errorf("failed to delete persistent volume claim %s: %w", pvc.Name, err)
		}
	}

	return nil
}
