package k8scluster

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

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

// setupClusterResources sets up the cluster resources according to configuration
func (s *Service) setupClusterResources(ctx context.Context) error {
	s.logger.Info("Setting up cluster resources",
		zap.Int64("cpu", s.config.ClusterResources.CPU),
		zap.String("memory", s.config.ClusterResources.Memory),
		zap.String("storage", s.config.ClusterResources.Storage),
		zap.Int64("gpu", s.config.ClusterResources.GPU))

	// Get nodes
	nodes, err := s.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	if len(nodes.Items) == 0 {
		return fmt.Errorf("no nodes found in cluster")
	}

	// Check and setup storage
	if s.config.ClusterResources.Storage != "" {
		if err := s.setupStorage(ctx); err != nil {
			return fmt.Errorf("failed to setup storage: %w", err)
		}
	}

	// Check and setup GPU if needed
	if s.config.ClusterResources.GPU > 0 {
		if err := s.setupGPU(ctx); err != nil {
			return fmt.Errorf("failed to setup GPU: %w", err)
		}
	}

	// Verify CPU and memory resources
	for _, node := range nodes.Items {
		// Check CPU
		if cpu, ok := node.Status.Capacity.Cpu().AsInt64(); ok {
			if cpu < s.config.ClusterResources.CPU {
				s.logger.Warn("Node has insufficient CPU",
					zap.String("node", node.Name),
					zap.Int64("available", cpu),
					zap.Int64("required", s.config.ClusterResources.CPU))
			}
		}

		// Check Memory
		if mem, ok := node.Status.Capacity.Memory().AsInt64(); ok {
			requiredMem, err := resource.ParseQuantity(s.config.ClusterResources.Memory)
			if err != nil {
				return fmt.Errorf("invalid memory configuration: %w", err)
			}
			if mem < requiredMem.Value() {
				s.logger.Warn("Node has insufficient memory",
					zap.String("node", node.Name),
					zap.Int64("available", mem),
					zap.Int64("required", requiredMem.Value()))
			}
		}
	}

	return nil
}

// setupStorage sets up the storage resources
func (s *Service) setupStorage(ctx context.Context) error {
	// Skip storage setup if storage requirement is 0
	if s.config.ClusterResources.Storage == "0Gi" || s.config.ClusterResources.Storage == "0" {
		s.logger.Info("Skipping storage setup as storage requirement is 0")
		return nil
	}

	// Get nodes for node affinity
	nodes, err := s.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}
	if len(nodes.Items) == 0 {
		return fmt.Errorf("no nodes found in cluster")
	}

	// Create storage class if it doesn't exist
	waitForFirstConsumer := storagev1.VolumeBindingWaitForFirstConsumer
	retain := corev1.PersistentVolumeReclaimRetain
	storageClass := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "local-path",
		},
		Provisioner:       "rancher.io/local-path",
		VolumeBindingMode: &waitForFirstConsumer,
		ReclaimPolicy:     &retain,
	}

	_, err = s.client.StorageV1().StorageClasses().Create(ctx, storageClass, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create storage class: %w", err)
	}

	// Create persistent volume
	storageQuantity, err := resource.ParseQuantity(s.config.ClusterResources.Storage)
	if err != nil {
		return fmt.Errorf("invalid storage configuration: %w", err)
	}

	filesystem := corev1.PersistentVolumeFilesystem
	hostPathType := corev1.HostPathDirectoryOrCreate
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "local-pv",
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: storageQuantity,
			},
			VolumeMode:                    &filesystem,
			AccessModes:                   []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
			StorageClassName:              "local-path",
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/mnt/data",
					Type: &hostPathType,
				},
			},
			NodeAffinity: &corev1.VolumeNodeAffinity{
				Required: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "kubernetes.io/hostname",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{nodes.Items[0].Name},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err = s.client.CoreV1().PersistentVolumes().Create(ctx, pv, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create persistent volume: %w", err)
	}

	// Create storage directory
	cmd := exec.Command("sudo", "mkdir", "-p", "/mnt/data")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	cmd = exec.Command("sudo", "chmod", "777", "/mnt/data")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set storage directory permissions: %w", err)
	}

	return nil
}

// setupGPU sets up GPU resources
func (s *Service) setupGPU(ctx context.Context) error {
	// Check if NVIDIA device plugin is installed
	_, err := s.client.AppsV1().DaemonSets("kube-system").Get(ctx, "nvidia-device-plugin-daemonset", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			s.logger.Warn("NVIDIA device plugin not found. GPU support may not be available")
		} else {
			return fmt.Errorf("failed to check NVIDIA device plugin: %w", err)
		}
	}

	// Check if AMD device plugin is installed
	_, err = s.client.AppsV1().DaemonSets("kube-system").Get(ctx, "amd-device-plugin-daemonset", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			s.logger.Warn("AMD device plugin not found. GPU support may not be available")
		} else {
			return fmt.Errorf("failed to check AMD device plugin: %w", err)
		}
	}

	return nil
}
