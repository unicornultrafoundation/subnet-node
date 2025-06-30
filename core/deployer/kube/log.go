package kube

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewServiceLog creates and returns a service log with provided details
func NewServiceLog(name string, stream io.ReadCloser) *types.ServiceLog {
	return &types.ServiceLog{
		Name:    name,
		Stream:  stream,
		Scanner: bufio.NewScanner(stream),
	}
}

// GetDeploymentLog retrieves logs from all pods in a deployment namespace
func (c *KubeClient) GetDeploymentLogs(ctx context.Context, deploymentID string, opts *corev1.PodLogOptions) ([]*types.ServiceLog, error) {
	if opts == nil {
		opts = &corev1.PodLogOptions{
			Follow:     true,
			Timestamps: false,
		}
	}

	// List all pods in the deployment namespace
	pods, err := c.Client.CoreV1().Pods(deploymentID).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace %s: %w", deploymentID, err)
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods found in deployment namespace %s", deploymentID)
	}

	var logStreams []*types.ServiceLog

	// Get logs from each pod
	for _, pod := range pods.Items {
		c.Logger.WithField("pod", pod.Name).WithField("phase", pod.Status.Phase).Debug("Processing pod for logs")

		stream, err := c.Client.CoreV1().Pods(deploymentID).GetLogs(pod.Name, opts).Stream(ctx)

		if err != nil {
			c.Logger.WithField("pod", pod.Name).WithError(err).Error("Failed to get log stream")
			continue
		}

		logStreams = append(logStreams, NewServiceLog(pod.Name, stream))
	}

	if len(logStreams) == 0 {
		return nil, fmt.Errorf("no log streams available for deployment %s", deploymentID)
	}

	return logStreams, nil
}
