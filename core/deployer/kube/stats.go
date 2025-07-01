package kube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// GetDeploymentStats retrieves real-time statistics for a deployment by deploymentID.
func (s *KubeClient) GetDeploymentStats(ctx context.Context, deploymentID string) (*types.DeploymentStats, error) {
	// Check if the namespace (deployment) exists and get it once
	ns, err := s.Client.CoreV1().Namespaces().Get(ctx, deploymentID, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("deployment (namespace) %q not found", deploymentID)
		}
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}

	// Get pod metrics from metrics API
	podMetricsList, err := s.MetricsClient.MetricsV1beta1().PodMetricses(deploymentID).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %w", err)
	}

	var usedCPU, usedMemory, usedEphemeralStorage uint64
	for _, podMetrics := range podMetricsList.Items {
		for _, container := range podMetrics.Containers {
			cpu := container.Usage.Cpu().MilliValue()    // millicores
			mem := container.Usage.Memory().Value()      // bytes
			storage := container.Usage.Storage().Value() // bytes (ephemeral)
			usedCPU += uint64(cpu)
			usedMemory += uint64(mem)
			usedEphemeralStorage += uint64(storage)
		}
	}

	// Get actual persistent storage usage from kubelet volume stats
	usedPersistentStorage, err := s.getVolumeStatsFromKubelet(ctx, deploymentID)
	if err != nil {
		s.Logger.WithError(err).Warn("Failed to get volume stats from kubelet, falling back to PVC resource requests")
		// Fallback to PVC resource requests (old method)
		pvcs, err := s.Client.CoreV1().PersistentVolumeClaims(deploymentID).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, pvc := range pvcs.Items {
				if req, ok := pvc.Spec.Resources.Requests["storage"]; ok {
					usedPersistentStorage += uint64(req.Value()) // bytes
				}
			}
		}
	}

	// Get GPU usage
	usedGPU, err := s.getGPUStats(ctx, deploymentID)
	if err != nil {
		s.Logger.WithError(err).Warn("Failed to get GPU stats, using zero values")
		usedGPU = 0
	}

	// Get network usage from optimized method with shorter timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second) // Reduced from 15s to 5s
	defer cancel()

	usedUploadBytes, usedDownloadBytes, err := s.getNetworkStats(ctx, deploymentID)
	if err != nil {
		s.Logger.WithError(err).Warn("Failed to get network stats, using zero values")
		usedUploadBytes = 0
		usedDownloadBytes = 0
	}

	// Calculate duration from namespace creation time
	duration := int64(time.Since(ns.CreationTimestamp.Time).Seconds())

	return &types.DeploymentStats{
		UsedCpu:           usedCPU,
		UsedMemory:        usedMemory,
		UsedStorage:       usedEphemeralStorage + usedPersistentStorage,
		UsedUploadBytes:   usedUploadBytes,
		UsedDownloadBytes: usedDownloadBytes,
		UsedGpu:           usedGPU,
		Duration:          duration,
	}, nil
}

// getNetworkStats retrieves network statistics for a deployment using Kubernetes APIs
func (s *KubeClient) getNetworkStats(ctx context.Context, deploymentID string) (upload, download uint64, err error) {
	// Get all pods in the deployment
	pods, err := s.Client.CoreV1().Pods(deploymentID).List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return 0, 0, nil
	}

	// Try to get network stats from kubelet stats summary API
	upload, download, err = s.getNetworkStatsFromKubelet(ctx, deploymentID)
	if err == nil && (upload > 0 || download > 0) {
		s.Logger.WithField("deployment", deploymentID).Debugf("Got network stats from kubelet: Upload=%d, Download=%d", upload, download)
		return upload, download, nil
	}

	s.Logger.WithError(err).WithField("deployment", deploymentID).Debug("Failed to get network stats from kubelet, trying container runtime")

	// Fallback: try to get stats from container runtime (Docker/containerd)
	upload, download, err = s.getNetworkStatsFromContainerRuntime(ctx, deploymentID)
	if err == nil && (upload > 0 || download > 0) {
		s.Logger.WithField("deployment", deploymentID).Debugf("Got network stats from container runtime: Upload=%d, Download=%d", upload, download)
		return upload, download, nil
	}

	s.Logger.WithError(err).WithField("deployment", deploymentID).Debug("Failed to get network stats from container runtime, trying metrics API")

	// Fallback: try to get stats from container metrics API
	upload, download, err = s.getNetworkStatsFromMetricsAPI(ctx, deploymentID)
	if err == nil && (upload > 0 || download > 0) {
		s.Logger.WithField("deployment", deploymentID).Debugf("Got network stats from metrics API: Upload=%d, Download=%d", upload, download)
		return upload, download, nil
	}

	s.Logger.WithError(err).WithField("deployment", deploymentID).Warn("Failed to get network stats from all Kubernetes sources")
	return 0, 0, nil
}

// getNetworkStatsFromKubelet gets network stats from kubelet stats summary API
func (s *KubeClient) getNetworkStatsFromKubelet(ctx context.Context, deploymentID string) (upload, download uint64, err error) {
	// Get node name (assuming single node cluster for now)
	nodes, err := s.Client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to list nodes: %w", err)
	}
	if len(nodes.Items) == 0 {
		return 0, 0, fmt.Errorf("no nodes found")
	}
	nodeName := nodes.Items[0].Name

	// Check if REST client is available
	if s.Client.CoreV1().RESTClient() == nil {
		return 0, 0, fmt.Errorf("REST client not available")
	}

	// Get kubelet stats summary
	statsURL := fmt.Sprintf("/api/v1/nodes/%s/proxy/stats/summary", nodeName)
	req := s.Client.CoreV1().RESTClient().Get().RequestURI(statsURL)
	raw, err := req.DoRaw(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get kubelet stats: %w", err)
	}

	// Parse the JSON response to find pods in our namespace
	var statsSummary struct {
		Pods []struct {
			PodRef struct {
				Namespace string `json:"namespace"`
			} `json:"podRef"`
			Network struct {
				RxBytes uint64 `json:"rxBytes"`
				TxBytes uint64 `json:"txBytes"`
			} `json:"network"`
		} `json:"pods"`
	}

	if err := json.Unmarshal(raw, &statsSummary); err != nil {
		return 0, 0, fmt.Errorf("failed to parse kubelet stats: %w", err)
	}

	var totalUpload, totalDownload uint64
	for _, pod := range statsSummary.Pods {
		if pod.PodRef.Namespace == deploymentID {
			totalUpload += pod.Network.TxBytes
			totalDownload += pod.Network.RxBytes
		}
	}

	return totalUpload, totalDownload, nil
}

// getNetworkStatsFromMetricsAPI gets network stats from Kubernetes metrics API
func (s *KubeClient) getNetworkStatsFromMetricsAPI(ctx context.Context, deploymentID string) (upload, download uint64, err error) {
	// Note: The standard metrics API doesn't include network stats
	// This is a placeholder for when network metrics become available
	// For now, return zero values
	return 0, 0, fmt.Errorf("network metrics not available in metrics API")
}

// getNetworkStatsFromContainerRuntime gets network stats from container runtime (Docker/containerd)
func (s *KubeClient) getNetworkStatsFromContainerRuntime(ctx context.Context, deploymentID string) (upload, download uint64, err error) {
	// Get all pods in the deployment
	pods, err := s.Client.CoreV1().Pods(deploymentID).List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to list pods: %w", err)
	}

	var totalUpload, totalDownload uint64

	for _, pod := range pods.Items {
		for _, container := range pod.Status.ContainerStatuses {
			if container.ContainerID == "" {
				continue
			}

			// Extract container ID
			containerID := container.ContainerID
			if strings.Contains(container.ContainerID, "://") {
				parts := strings.Split(container.ContainerID, "://")
				if len(parts) == 2 {
					containerID = parts[1]
				}
			}

			// Try to get stats from Docker
			if strings.Contains(container.ContainerID, "docker://") {
				containerUpload, containerDownload, err := s.getDockerContainerNetworkStats(containerID)
				if err == nil {
					totalUpload += containerUpload
					totalDownload += containerDownload
					continue
				}
			}

			// Try to get stats from containerd
			if strings.Contains(container.ContainerID, "containerd://") {
				containerUpload, containerDownload, err := s.getContainerdContainerNetworkStats(containerID)
				if err == nil {
					totalUpload += containerUpload
					totalDownload += containerDownload
					continue
				}
			}
		}
	}

	return totalUpload, totalDownload, nil
}

// getDockerContainerNetworkStats gets network stats from Docker container
func (s *KubeClient) getDockerContainerNetworkStats(containerID string) (upload, download uint64, err error) {
	// Try to get stats using docker CLI
	cmd := exec.Command("docker", "stats", containerID, "--no-stream", "--format", "{{.NetIO}}")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get docker stats: %w", err)
	}

	// Parse the output format: "1.23MB / 456KB" or "0B / 0B"
	netIO := strings.TrimSpace(string(output))
	if netIO == "" {
		return 0, 0, fmt.Errorf("empty docker stats output")
	}

	// If Docker stats show 0B / 0B, try to get stats from /proc/net/dev inside the container
	if netIO == "0B / 0B" {
		return s.getContainerNetworkStatsFromProcNetDev(containerID)
	}

	// Parse the format "upload / download"
	parts := strings.Split(netIO, " / ")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid docker stats format: %s", netIO)
	}

	uploadStr := strings.TrimSpace(parts[0])
	downloadStr := strings.TrimSpace(parts[1])

	// Convert to bytes
	uploadBytes, err := s.parseDockerBytes(uploadStr)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse upload bytes: %w", err)
	}

	downloadBytes, err := s.parseDockerBytes(downloadStr)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse download bytes: %w", err)
	}

	return uploadBytes, downloadBytes, nil
}

// getContainerNetworkStatsFromProcNetDev gets network stats from /proc/net/dev inside the container
func (s *KubeClient) getContainerNetworkStatsFromProcNetDev(containerID string) (upload, download uint64, err error) {
	// Execute cat /proc/net/dev inside the container
	cmd := exec.Command("docker", "exec", containerID, "cat", "/proc/net/dev")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read /proc/net/dev from container: %w", err)
	}

	// Parse the /proc/net/dev output
	lines := strings.Split(string(output), "\n")
	var totalUpload, totalDownload uint64

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Inter-|") || strings.HasPrefix(line, "face |") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		// Skip loopback interface
		if strings.HasPrefix(fields[0], "lo:") {
			continue
		}

		// Parse receive and transmit bytes
		// Format: interface | bytes packets errs drop fifo frame compressed multicast | bytes packets errs drop fifo colls carrier compressed
		receiveBytes, _ := strconv.ParseUint(fields[1], 10, 64)
		transmitBytes, _ := strconv.ParseUint(fields[9], 10, 64)

		totalDownload += receiveBytes
		totalUpload += transmitBytes
	}

	return totalUpload, totalDownload, nil
}

// getContainerdContainerNetworkStats gets network stats from containerd container
func (s *KubeClient) getContainerdContainerNetworkStats(containerID string) (upload, download uint64, err error) {
	// Try to get stats using ctr CLI (containerd)
	cmd := exec.Command("ctr", "containers", "info", containerID)
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get containerd container info: %w", err)
	}

	// Parse containerd output to extract network stats
	// This is a simplified approach - in production you'd need more sophisticated parsing
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "rx_bytes") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if bytes, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
					download = bytes
				}
			}
		} else if strings.Contains(line, "tx_bytes") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if bytes, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
					upload = bytes
				}
			}
		}
	}

	return upload, download, nil
}

// parseDockerBytes converts Docker's byte format to uint64
func (s *KubeClient) parseDockerBytes(bytesStr string) (uint64, error) {
	bytesStr = strings.TrimSpace(bytesStr)

	// Handle "0B" case
	if bytesStr == "0B" {
		return 0, nil
	}

	// Parse format like "1.23MB", "456KB", etc.
	var multiplier uint64 = 1
	var numericPart string

	if strings.HasSuffix(bytesStr, "B") {
		multiplier = 1
		numericPart = strings.TrimSuffix(bytesStr, "B")
	} else if strings.HasSuffix(bytesStr, "KB") {
		multiplier = 1024
		numericPart = strings.TrimSuffix(bytesStr, "KB")
	} else if strings.HasSuffix(bytesStr, "MB") {
		multiplier = 1024 * 1024
		numericPart = strings.TrimSuffix(bytesStr, "MB")
	} else if strings.HasSuffix(bytesStr, "GB") {
		multiplier = 1024 * 1024 * 1024
		numericPart = strings.TrimSuffix(bytesStr, "GB")
	} else {
		// Try to parse as plain number
		numericPart = bytesStr
	}

	// Parse the numeric part
	value, err := strconv.ParseFloat(numericPart, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse numeric value: %s", numericPart)
	}

	return uint64(value * float64(multiplier)), nil
}

// findPodForContainer finds the pod that contains a specific container
func (s *KubeClient) findPodForContainer(containerID string) (string, error) {
	// Extract short container ID
	shortID := containerID
	if strings.Contains(containerID, "://") {
		parts := strings.Split(containerID, "://")
		if len(parts) == 2 {
			shortID = parts[1]
		}
	}

	// List all pods across all namespaces to find the one with this container
	// This is expensive, so we should cache this information
	pods, err := s.Client.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list pods: %w", err)
	}

	for _, pod := range pods.Items {
		for _, container := range pod.Status.ContainerStatuses {
			if strings.Contains(container.ContainerID, shortID) {
				return fmt.Sprintf("%s/%s", pod.Namespace, pod.Name), nil
			}
		}
	}

	return "", fmt.Errorf("pod not found for container %s", containerID)
}

// getVolumeStatsFromKubelet gets actual volume usage from kubelet stats summary
func (s *KubeClient) getVolumeStatsFromKubelet(ctx context.Context, deploymentID string) (uint64, error) {
	// Get node name (assuming single node cluster for now)
	nodes, err := s.Client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to list nodes: %w", err)
	}
	if len(nodes.Items) == 0 {
		return 0, fmt.Errorf("no nodes found")
	}
	nodeName := nodes.Items[0].Name

	// Check if REST client is available
	if s.Client.CoreV1().RESTClient() == nil {
		return 0, fmt.Errorf("REST client not available")
	}

	// Get kubelet stats summary
	statsURL := fmt.Sprintf("/api/v1/nodes/%s/proxy/stats/summary", nodeName)
	req := s.Client.CoreV1().RESTClient().Get().RequestURI(statsURL)
	raw, err := req.DoRaw(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get kubelet stats: %w", err)
	}

	// Parse the JSON response to find pods in our namespace
	var statsSummary struct {
		Pods []struct {
			PodRef struct {
				Namespace string `json:"namespace"`
			} `json:"podRef"`
			Volume []struct {
				Name          string `json:"name"`
				UsedBytes     uint64 `json:"usedBytes"`
				CapacityBytes uint64 `json:"capacityBytes"`
			} `json:"volume"`
		} `json:"pods"`
	}

	if err := json.Unmarshal(raw, &statsSummary); err != nil {
		return 0, fmt.Errorf("failed to parse kubelet stats: %w", err)
	}

	var totalVolumeUsage uint64
	var hasVolumeStats bool

	for _, pod := range statsSummary.Pods {
		if pod.PodRef.Namespace == deploymentID {
			for _, volume := range pod.Volume {
				// Only count persistent volumes (not ephemeral volumes like kube-api-access)
				if !strings.Contains(volume.Name, "kube-api-access") {
					totalVolumeUsage += volume.UsedBytes
					hasVolumeStats = true
				}
			}
		}
	}

	// If kubelet stats don't provide accurate volume usage (e.g., for emptyDir volumes),
	// fall back to getting actual storage usage from containers
	if !hasVolumeStats || totalVolumeUsage == 0 {
		s.Logger.WithField("deployment", deploymentID).Debug("Kubelet volume stats not available, getting storage from containers")
		return s.getContainerStorageUsage(ctx, deploymentID)
	}

	return totalVolumeUsage, nil
}

// getContainerStorageUsage gets actual storage usage from containers using du command
func (s *KubeClient) getContainerStorageUsage(ctx context.Context, deploymentID string) (uint64, error) {
	// Get all pods in the deployment
	pods, err := s.Client.CoreV1().Pods(deploymentID).List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to list pods: %w", err)
	}

	var totalStorageUsage uint64

	for _, pod := range pods.Items {
		for _, container := range pod.Status.ContainerStatuses {
			if container.ContainerID == "" {
				continue
			}

			// Extract container ID
			containerID := container.ContainerID
			if strings.Contains(container.ContainerID, "://") {
				parts := strings.Split(container.ContainerID, "://")
				if len(parts) == 2 {
					containerID = parts[1]
				}
			}

			// Get storage usage from container using du command
			containerUsage, err := s.getDockerContainerStorageUsage(containerID)
			if err == nil {
				totalStorageUsage += containerUsage
			} else {
				s.Logger.WithError(err).WithField("container", containerID).Debug("Failed to get container storage usage")
			}
		}
	}

	return totalStorageUsage, nil
}

// getDockerContainerStorageUsage gets storage usage from Docker container using du command
func (s *KubeClient) getDockerContainerStorageUsage(containerID string) (uint64, error) {
	// Get storage usage for the /data directory (common mount point for storage)
	cmd := exec.Command("docker", "exec", containerID, "du", "-sb", "/data")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get container storage usage: %w", err)
	}

	// Parse the output (format: "size_in_bytes\t/data")
	outputStr := strings.TrimSpace(string(output))
	parts := strings.Fields(outputStr)
	if len(parts) < 1 {
		return 0, fmt.Errorf("invalid du output format: %s", outputStr)
	}

	// Parse the size in bytes
	sizeBytes, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse storage size: %w", err)
	}

	return sizeBytes, nil
}

// getGPUStats retrieves GPU statistics for a deployment
func (s *KubeClient) getGPUStats(ctx context.Context, deploymentID string) (uint64, error) {
	// Get all pods in the deployment
	pods, err := s.Client.CoreV1().Pods(deploymentID).List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to list pods: %w", err)
	}

	var totalGPUCount uint64
	var totalGPUUsage uint64

	for _, pod := range pods.Items {
		// Check if pod has GPU resources
		for _, container := range pod.Spec.Containers {
			if s.hasGPUResources(container.Resources) {
				// Count GPU resources
				gpuCount := s.getGPUResourceCount(container.Resources)
				totalGPUCount += gpuCount
			}
		}

		// Try to get actual GPU usage from within the pod
		if pod.Status.Phase == corev1.PodRunning {
			podGPUUsage, err := s.getPodGPUUsage(ctx, pod)
			if err == nil {
				totalGPUUsage += podGPUUsage
			} else {
				s.Logger.WithError(err).WithField("pod", pod.Name).Debug("Failed to get GPU usage from pod")
			}
		}
	}

	// If we got actual usage from pods, return that
	if totalGPUUsage > 0 {
		return totalGPUUsage, nil
	}

	// Otherwise return the GPU count as fallback
	return totalGPUCount, nil
}

// getGPUResourceCount gets the total count of GPU resources for a container
func (s *KubeClient) getGPUResourceCount(resources corev1.ResourceRequirements) uint64 {
	gpuResourceTypes := []string{
		"nvidia.com/gpu", // NVIDIA GPUs
		"amd.com/gpu",    // AMD GPUs
		"intel.com/gpu",  // Intel GPUs
		"apple.com/gpu",  // Apple Silicon GPUs
		"gpu",            // Generic GPU resource
	}

	var totalCount uint64

	// Check requests first
	for _, gpuType := range gpuResourceTypes {
		if quantity, ok := resources.Requests[corev1.ResourceName(gpuType)]; ok {
			totalCount += uint64(quantity.Value())
		}
	}

	// If no requests, check limits
	if totalCount == 0 {
		for _, gpuType := range gpuResourceTypes {
			if quantity, ok := resources.Limits[corev1.ResourceName(gpuType)]; ok {
				totalCount += uint64(quantity.Value())
			}
		}
	}

	return totalCount
}

// getPodGPUUsage gets GPU usage from within a specific pod
func (s *KubeClient) getPodGPUUsage(ctx context.Context, pod corev1.Pod) (uint64, error) {
	// Check if pod has GPU resources
	hasGPU := false
	for _, container := range pod.Spec.Containers {
		if s.hasGPUResources(container.Resources) {
			hasGPU = true
			break
		}
	}

	if !hasGPU {
		return 0, nil
	}

	// Try to get GPU usage from the first container with GPU
	for _, container := range pod.Spec.Containers {
		if s.hasGPUResources(container.Resources) {
			return s.getContainerGPUUsage(ctx, pod, container.Name)
		}
	}

	return 0, fmt.Errorf("no GPU containers found in pod")
}

// hasGPUResources checks if a container has any GPU resources
func (s *KubeClient) hasGPUResources(resources corev1.ResourceRequirements) bool {
	gpuResourceTypes := []string{
		"nvidia.com/gpu", // NVIDIA GPUs
		"amd.com/gpu",    // AMD GPUs
		"intel.com/gpu",  // Intel GPUs
		"apple.com/gpu",  // Apple Silicon GPUs
		"gpu",            // Generic GPU resource
	}

	// Check requests
	for _, gpuType := range gpuResourceTypes {
		if _, ok := resources.Requests[corev1.ResourceName(gpuType)]; ok {
			return true
		}
	}

	// Check limits
	for _, gpuType := range gpuResourceTypes {
		if _, ok := resources.Limits[corev1.ResourceName(gpuType)]; ok {
			return true
		}
	}

	return false
}

// getContainerGPUUsage gets GPU usage from within a specific container
func (s *KubeClient) getContainerGPUUsage(ctx context.Context, pod corev1.Pod, containerName string) (uint64, error) {
	// Try NVIDIA GPU first
	gpuUsage, err := s.getNvidiaGPUUsage(ctx, pod, containerName)
	if err == nil && gpuUsage > 0 {
		return gpuUsage, nil
	}

	// Try Apple Silicon GPU
	gpuUsage, err = s.getAppleSiliconGPUUsage(ctx, pod, containerName)
	if err == nil && gpuUsage > 0 {
		return gpuUsage, nil
	}

	// Fallback: check if GPU is available but usage can't be measured
	if s.isGPUAvailable(ctx, pod, containerName) {
		return 1, nil // Return 1 to indicate GPU is available
	}

	return 0, fmt.Errorf("no GPU usage data available")
}

// getNvidiaGPUUsage gets NVIDIA GPU usage from within a container
func (s *KubeClient) getNvidiaGPUUsage(ctx context.Context, pod corev1.Pod, containerName string) (uint64, error) {
	// Execute nvidia-smi inside the container
	cmd := []string{"nvidia-smi", "--query-gpu=utilization.gpu", "--format=csv,noheader,nounits"}

	output, err := s.execInContainer(ctx, pod.Namespace, pod.Name, containerName, cmd)
	if err != nil {
		return 0, fmt.Errorf("failed to execute nvidia-smi in container: %w", err)
	}

	var totalUsage uint64
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		usage, err := strconv.ParseUint(line, 10, 64)
		if err != nil {
			continue
		}

		totalUsage += usage
	}

	return totalUsage, nil
}

// getAppleSiliconGPUUsage gets Apple Silicon GPU usage from within a container
func (s *KubeClient) getAppleSiliconGPUUsage(ctx context.Context, pod corev1.Pod, containerName string) (uint64, error) {
	// Check if we're on Apple Silicon
	cmd := []string{"uname", "-m"}
	output, err := s.execInContainer(ctx, pod.Namespace, pod.Name, containerName, cmd)
	if err != nil {
		return 0, fmt.Errorf("failed to check architecture: %w", err)
	}

	arch := strings.TrimSpace(string(output))
	if arch != "arm64" && !strings.Contains(arch, "aarch64") {
		return 0, fmt.Errorf("not Apple Silicon architecture")
	}

	// Try to get GPU info from system_profiler (if available in container)
	cmd = []string{"system_profiler", "SPDisplaysDataType"}
	output, err = s.execInContainer(ctx, pod.Namespace, pod.Name, containerName, cmd)
	if err != nil {
		// Fallback: check for GPU-related environment variables
		return s.checkGPUEnvironment(ctx, pod, containerName)
	}

	// Parse system_profiler output
	outputStr := string(output)
	if strings.Contains(outputStr, "Metal") || strings.Contains(outputStr, "Apple") {
		return 1, nil // Return 1 to indicate GPU is available
	}

	return 0, fmt.Errorf("no Apple Silicon GPU detected")
}

// isGPUAvailable checks if GPU is available in the container
func (s *KubeClient) isGPUAvailable(ctx context.Context, pod corev1.Pod, containerName string) bool {
	// Check for GPU-related environment variables
	usage, err := s.checkGPUEnvironment(ctx, pod, containerName)
	return err == nil && usage > 0
}

// checkGPUEnvironment checks for GPU-related environment variables in the container
func (s *KubeClient) checkGPUEnvironment(ctx context.Context, pod corev1.Pod, containerName string) (uint64, error) {
	// Execute env command to check for GPU-related environment variables
	cmd := []string{"env"}
	output, err := s.execInContainer(ctx, pod.Namespace, pod.Name, containerName, cmd)
	if err != nil {
		return 0, fmt.Errorf("failed to check environment: %w", err)
	}

	outputStr := string(output)
	gpuEnvVars := []string{
		"NVIDIA_VISIBLE_DEVICES",
		"CUDA_VISIBLE_DEVICES",
		"GPU_DEVICE_ORDINAL",
		"METAL_DEVICE",
	}

	for _, envVar := range gpuEnvVars {
		if strings.Contains(outputStr, envVar+"=") {
			return 1, nil // Return 1 to indicate GPU environment is available
		}
	}

	return 0, fmt.Errorf("no GPU environment detected")
}

// execInContainer executes a command inside a container
func (s *KubeClient) execInContainer(ctx context.Context, namespace, podName, containerName string, cmd []string) ([]byte, error) {
	// Create exec request
	req := s.Client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   cmd,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	// Execute the command
	exec, err := remotecommand.NewSPDYExecutor(s.Config, "POST", req.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}
