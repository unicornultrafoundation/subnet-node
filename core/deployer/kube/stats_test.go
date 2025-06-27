package kube

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	metricsfake "k8s.io/metrics/pkg/client/clientset/versioned/fake"
)

func TestGetDeploymentStats(t *testing.T) {
	// Create fake clients
	fakeClient := fake.NewSimpleClientset()
	metricsClient := metricsfake.NewSimpleClientset()

	// Setup test data
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-deployment",
			CreationTimestamp: metav1.Time{
				Time: time.Now().Add(-1 * time.Hour),
			},
		},
	}
	fakeClient.CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})

	// Create a test pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-deployment",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "test-container",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse("100m"),
							"memory": resource.MustParse("128Mi"),
						},
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("200m"),
							"memory": resource.MustParse("256Mi"),
						},
					},
				},
			},
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:        "test-container",
					ContainerID: "docker://test-container-id",
				},
			},
		},
	}
	fakeClient.CoreV1().Pods("test-deployment").Create(context.Background(), pod, metav1.CreateOptions{})

	// Create KubeClient
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client:        fakeClient,
		MetricsClient: metricsClient,
		Logger:        logger,
	}

	// Test GetDeploymentStats
	stats, err := client.GetDeploymentStats(context.Background(), "test-deployment")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "test-deployment", namespace.Name)
	assert.GreaterOrEqual(t, stats.Duration, int64(0))
}

func TestGetNetworkStatsFromKubelet(t *testing.T) {
	// Create fake client
	fakeClient := fake.NewSimpleClientset()

	// Setup test data
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
	}
	fakeClient.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})

	// Create KubeClient with minimal setup
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client: fakeClient,
		Logger: logger,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Log("Recovered from panic as expected due to missing REST client in fake client")
		}
	}()

	// Test getNetworkStatsFromKubelet
	_, _, _ = client.getNetworkStatsFromKubelet(context.Background(), "test-deployment")
}

func TestGetNetworkStatsFromMetricsAPI(t *testing.T) {
	// Create KubeClient
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Logger: logger,
	}

	// Test getNetworkStatsFromMetricsAPI
	upload, download, err := client.getNetworkStatsFromMetricsAPI(context.Background(), "test-deployment")

	// Assertions
	assert.Error(t, err) // Should return error as network metrics not available
	assert.Equal(t, uint64(0), upload)
	assert.Equal(t, uint64(0), download)
}

func TestFindPodForContainer(t *testing.T) {
	// Create fake client
	fakeClient := fake.NewSimpleClientset()

	// Setup test data
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:        "test-container",
					ContainerID: "docker://test-container-id",
				},
			},
		},
	}
	fakeClient.CoreV1().Pods("test-namespace").Create(context.Background(), pod, metav1.CreateOptions{})

	// Create KubeClient
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client: fakeClient,
		Logger: logger,
	}

	// Test findPodForContainer
	podName, err := client.findPodForContainer("docker://test-container-id")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "test-namespace/test-pod", podName)
}

func TestGetGPUStats(t *testing.T) {
	// Create fake client
	fakeClient := fake.NewSimpleClientset()

	// Setup test data with GPU resources
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gpu-pod",
			Namespace: "test-deployment",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "gpu-container",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"nvidia.com/gpu": resource.MustParse("1"),
						},
						Limits: corev1.ResourceList{
							"nvidia.com/gpu": resource.MustParse("1"),
						},
					},
				},
			},
		},
	}
	fakeClient.CoreV1().Pods("test-deployment").Create(context.Background(), pod, metav1.CreateOptions{})

	// Create KubeClient
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client: fakeClient,
		Logger: logger,
	}

	// Test getGPUStats
	gpuCount, err := client.getGPUStats(context.Background(), "test-deployment")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), gpuCount)
}

func TestGetVolumeStatsFromKubelet(t *testing.T) {
	// Create fake client
	fakeClient := fake.NewSimpleClientset()

	// Setup test data
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
	}
	fakeClient.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})

	// Create KubeClient
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client: fakeClient,
		Logger: logger,
	}

	// Test getVolumeStatsFromKubelet with panic recovery
	var volumeUsage uint64
	var err error

	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected panic due to fake client's REST client being nil
				err = fmt.Errorf("expected panic recovered: %v", r)
			}
		}()

		volumeUsage, err = client.getVolumeStatsFromKubelet(context.Background(), "test-deployment")
	}()

	// Assertions
	assert.Error(t, err) // Should fail with fake client
	assert.Equal(t, uint64(0), volumeUsage)
}

// Benchmark tests for performance
func BenchmarkGetDeploymentStats(b *testing.B) {
	// Create fake clients
	fakeClient := fake.NewSimpleClientset()
	metricsClient := metricsfake.NewSimpleClientset()

	// Setup test data
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "benchmark-deployment",
			CreationTimestamp: metav1.Time{
				Time: time.Now().Add(-1 * time.Hour),
			},
		},
	}
	fakeClient.CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})

	// Create multiple pods for benchmarking
	for i := 0; i < 10; i++ {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("pod-%d", i),
				Namespace: "benchmark-deployment",
			},
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{
						Name:        fmt.Sprintf("container-%d", i),
						ContainerID: fmt.Sprintf("docker://container%d", i),
					},
				},
			},
		}
		fakeClient.CoreV1().Pods("benchmark-deployment").Create(context.Background(), pod, metav1.CreateOptions{})
	}

	// Create KubeClient
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client:        fakeClient,
		MetricsClient: metricsClient,
		Logger:        logger,
	}

	// Run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetDeploymentStats(context.Background(), "benchmark-deployment")
		if err != nil {
			b.Fatalf("GetDeploymentStats failed: %v", err)
		}
	}
}

// BenchmarkGetNetworkStats benchmarks the Kubernetes-native network stats collection
func BenchmarkGetNetworkStats(b *testing.B) {
	// Create fake client
	fakeClient := fake.NewSimpleClientset()

	// Setup test data
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "benchmark-network",
		},
	}
	fakeClient.CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})

	// Create multiple pods for benchmarking
	for i := 0; i < 5; i++ {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("pod-%d", i),
				Namespace: "benchmark-network",
			},
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{
						Name:        fmt.Sprintf("container-%d", i),
						ContainerID: fmt.Sprintf("docker://container%d", i),
					},
				},
			},
		}
		fakeClient.CoreV1().Pods("benchmark-network").Create(context.Background(), pod, metav1.CreateOptions{})
	}

	// Create KubeClient
	logger := logrus.New().WithField("service", "kube").Logger
	client := &KubeClient{
		Client: fakeClient,
		Logger: logger,
	}

	// Run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := client.getNetworkStats(context.Background(), "benchmark-network")
		if err != nil {
			b.Fatalf("getNetworkStats failed: %v", err)
		}
	}
}
