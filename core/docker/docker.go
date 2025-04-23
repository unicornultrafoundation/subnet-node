package docker

import (
	"context"
	"io"
	"strings"

	dtypes "github.com/docker/docker/api/types"
	ctypes "github.com/docker/docker/api/types/container"
	itypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	vtypes "github.com/docker/docker/api/types/volume"
	dockerCli "github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type DockerClient interface {
	ContainerInspect(ctx context.Context, containerID string) (dtypes.ContainerJSON, error)
	ContainerCreate(ctx context.Context, config *ctypes.Config, hostConfig *ctypes.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (ctypes.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options ctypes.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options ctypes.StopOptions) error
	ContainerRemove(ctx context.Context, containerID string, options ctypes.RemoveOptions) error
	ContainerList(ctx context.Context, options ctypes.ListOptions) ([]dtypes.Container, error)
	ImagePull(ctx context.Context, ref string, options itypes.PullOptions) (io.ReadCloser, error)
	ContainerStats(ctx context.Context, containerID string, stream bool) (dtypes.ContainerStats, error)
	VolumeCreate(ctx context.Context, options vtypes.CreateOptions) (vtypes.Volume, error)
	VolumeRemove(ctx context.Context, volumeID string, force bool) error
	DiskUsage(ctx context.Context, options dtypes.DiskUsageOptions) (dtypes.DiskUsage, error)
	Close() error
}

type RealDockerClient struct {
	client *dockerCli.Client
}

func (r *RealDockerClient) ContainerInspect(ctx context.Context, containerID string) (dtypes.ContainerJSON, error) {
	return r.client.ContainerInspect(ctx, containerID)
}

func (r *RealDockerClient) ContainerCreate(ctx context.Context, config *ctypes.Config, hostConfig *ctypes.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (ctypes.CreateResponse, error) {
	return r.client.ContainerCreate(ctx, config, hostConfig, networkingConfig, platform, containerName)
}

func (r *RealDockerClient) ContainerStart(ctx context.Context, containerID string, options ctypes.StartOptions) error {
	return r.client.ContainerStart(ctx, containerID, options)
}

func (r *RealDockerClient) ContainerStop(ctx context.Context, containerID string, options ctypes.StopOptions) error {
	return r.client.ContainerStop(ctx, containerID, options)
}

func (r *RealDockerClient) ContainerRemove(ctx context.Context, containerID string, options ctypes.RemoveOptions) error {
	return r.client.ContainerRemove(ctx, containerID, options)
}

func (r *RealDockerClient) ContainerList(ctx context.Context, options ctypes.ListOptions) ([]dtypes.Container, error) {
	containers, err := r.client.ContainerList(ctx, options)
	if err != nil {
		return nil, err
	}

	var filteredContainers []dtypes.Container
	for _, container := range containers {
		for _, name := range container.Names {
			if strings.HasPrefix(name, "/subnet-") {
				filteredContainers = append(filteredContainers, container)
				break
			}
		}
	}

	return filteredContainers, nil
}

func (r *RealDockerClient) ImagePull(ctx context.Context, ref string, options itypes.PullOptions) (io.ReadCloser, error) {
	return r.client.ImagePull(ctx, ref, options)
}

func (r *RealDockerClient) ContainerStats(ctx context.Context, containerID string, stream bool) (dtypes.ContainerStats, error) {
	return r.client.ContainerStats(ctx, containerID, stream)
}

func (r *RealDockerClient) VolumeCreate(ctx context.Context, options vtypes.CreateOptions) (vtypes.Volume, error) {
	return r.client.VolumeCreate(ctx, options)
}

func (r *RealDockerClient) VolumeRemove(ctx context.Context, volumeID string, force bool) error {
	return r.client.VolumeRemove(ctx, volumeID, force)
}

func (r *RealDockerClient) DiskUsage(ctx context.Context, options dtypes.DiskUsageOptions) (dtypes.DiskUsage, error) {
	return r.client.DiskUsage(ctx, options)
}

func (r *RealDockerClient) Close() error {
	return r.client.Close()
}

// MockDockerClient is a mock implementation of DockerClient for testing.
type MockDockerClient struct {
	InspectResponse   dtypes.ContainerJSON
	InspectError      error
	CreateResponse    ctypes.CreateResponse
	CreateError       error
	StartError        error
	StopError         error
	RemoveError       error
	ListResponse      []dtypes.Container
	ListError         error
	PullResponse      io.ReadCloser
	PullError         error
	StatsResponse     dtypes.ContainerStats
	VolumeResponse    vtypes.Volume
	VolumeError       error
	StatsError        error
	CloseError        error
	DiskUsageResponse dtypes.DiskUsage
	DiskUsageError    error
}

// ContainerInspect mocks container inspection.
func (m *MockDockerClient) ContainerInspect(ctx context.Context, containerID string) (dtypes.ContainerJSON, error) {
	return m.InspectResponse, m.InspectError
}

// ContainerCreate mocks container creation.
func (m *MockDockerClient) ContainerCreate(ctx context.Context, config *ctypes.Config, hostConfig *ctypes.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (ctypes.CreateResponse, error) {
	return m.CreateResponse, m.CreateError
}

// ContainerStart mocks starting a container.
func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, options ctypes.StartOptions) error {
	return m.StartError
}

// ContainerStop mocks stopping a container.
func (m *MockDockerClient) ContainerStop(ctx context.Context, containerID string, options ctypes.StopOptions) error {
	return m.StopError
}

// ContainerRemove mocks removing a container.
func (m *MockDockerClient) ContainerRemove(ctx context.Context, containerID string, options ctypes.RemoveOptions) error {
	return m.RemoveError
}

// ContainerList mocks listing containers.
func (m *MockDockerClient) ContainerList(ctx context.Context, options ctypes.ListOptions) ([]dtypes.Container, error) {
	return m.ListResponse, m.ListError
}

// ImagePull mocks pulling an image.
func (m *MockDockerClient) ImagePull(ctx context.Context, ref string, options itypes.PullOptions) (io.ReadCloser, error) {
	return m.PullResponse, m.PullError
}

// ContainerStats mocks getting container stats.
func (m *MockDockerClient) ContainerStats(ctx context.Context, containerID string, stream bool) (dtypes.ContainerStats, error) {
	return m.StatsResponse, m.StatsError
}

func (m *MockDockerClient) VolumeCreate(ctx context.Context, options vtypes.CreateOptions) (vtypes.Volume, error) {
	return m.VolumeResponse, m.VolumeError
}

func (m *MockDockerClient) VolumeRemove(ctx context.Context, volumeID string, force bool) error {
	return m.VolumeError
}

func (m *MockDockerClient) DiskUsage(ctx context.Context, options dtypes.DiskUsageOptions) (dtypes.DiskUsage, error) {
	return m.DiskUsageResponse, m.DiskUsageError
}

// Close mocks closing the Docker client.
func (m *MockDockerClient) Close() error {
	return m.CloseError
}
