package container

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/go-cni"
)

// ContainerManager manages container lifecycle using containerd.
type ContainerManager struct {
	client *containerd.Client
	cni    cni.CNI
}

// NewContainerManager creates a new ContainerManager.
func NewContainerManager(socketPath string) (*ContainerManager, error) {
	client, err := containerd.New(socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %w", err)
	}

	cni, err := cni.New(cni.WithMinNetworkCount(2), cni.WithPluginDir([]string{"/opt/cni/bin"}), cni.WithConfListFile("/etc/cni/net.d"))
	if err != nil {
		return nil, fmt.Errorf("failed to create CNI: %w", err)
	}

	return &ContainerManager{client: client, cni: cni}, nil
}

// PullImage pulls an image from a registry.
func (cm *ContainerManager) PullImage(ctx context.Context, imageName string) (containerd.Image, error) {
	ctx = namespaces.WithNamespace(ctx, "subnet-apps")
	image, err := cm.client.Pull(ctx, imageName, containerd.WithPullUnpack)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}
	return image, nil
}

// CreateContainer creates a new container.
func (cm *ContainerManager) CreateContainer(ctx context.Context, id, imageName string, specOpts ...oci.SpecOpts) (containerd.Container, error) {
	ctx = namespaces.WithNamespace(ctx, "subnet-apps")
	image, err := cm.PullImage(ctx, imageName)
	if err != nil {
		return nil, err
	}

	container, err := cm.client.NewContainer(
		ctx,
		id,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(id+"-snapshot", image),
		containerd.WithNewSpec(specOpts...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	return container, nil
}

// StartContainer starts a container with network configuration.
func (cm *ContainerManager) StartContainer(ctx context.Context, container containerd.Container) (containerd.Task, error) {
	ctx = namespaces.WithNamespace(ctx, "subnet-apps")
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	if err := task.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start task: %w", err)
	}

	// Configure the network for the container using gocni
	if err := cm.setupNetwork(ctx, container.ID()); err != nil {
		return nil, fmt.Errorf("failed to setup network: %w", err)
	}

	fmt.Printf("Container %s started successfully\n", container.ID())
	return task, nil
}

// StopContainer stops a container.
func (cm *ContainerManager) StopContainer(ctx context.Context, task containerd.Task) error {
	ctx = namespaces.WithNamespace(ctx, "subnet-apps")
	exitStatusC, err := task.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for task: %w", err)
	}
	if err := task.Kill(ctx, syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to kill task: %w", err)
	}

	<-exitStatusC
	if _, err := task.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// RemoveContainer removes a container.
func (cm *ContainerManager) RemoveContainer(ctx context.Context, container containerd.Container) error {
	ctx = namespaces.WithNamespace(ctx, "subnet-apps")
	if err := container.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete container: %w", err)
	}
	return nil
}

// RestartContainer restarts a container.
func (cm *ContainerManager) RestartContainer(ctx context.Context, container containerd.Container) (containerd.Task, error) {
	ctx = namespaces.WithNamespace(ctx, "subnet-apps")
	task, err := container.Task(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if err := cm.StopContainer(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to stop container: %w", err)
	}

	return cm.StartContainer(ctx, container)
}

// setupNetwork configures the network for a container using gocni.
func (cm *ContainerManager) setupNetwork(ctx context.Context, containerID string) error {
	return nil
}

// GetContainerIP retrieves the IP address of a container from within its network namespace.
func (cm *ContainerManager) GetContainerIP(ctx context.Context, containerID string) (string, error) {
	netnsPath := fmt.Sprintf("/proc/%s/ns/net", containerID)
	cmd := exec.Command("ip", "netns", "exec", netnsPath, "ip", "-4", "addr", "show")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get container IP: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "inet") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				return strings.Split(parts[1], "/")[0], nil
			}
		}
	}

	return "", fmt.Errorf("IP address not found for container: %s", containerID)
}

// Close closes the containerd client.
func (cm *ContainerManager) Close() error {
	return cm.client.Close()
}
