package docker

import (
	"context"
	"io"
	"math/big"
	"time"

	dtypes "github.com/docker/docker/api/types"
	ctypes "github.com/docker/docker/api/types/container"
	itypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
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
	return r.client.ContainerList(ctx, options)
}

func (r *RealDockerClient) ImagePull(ctx context.Context, ref string, options itypes.PullOptions) (io.ReadCloser, error) {
	return r.client.ImagePull(ctx, ref, options)
}

func (r *RealDockerClient) ContainerStats(ctx context.Context, containerID string, stream bool) (dtypes.ContainerStats, error) {
	return r.client.ContainerStats(ctx, containerID, stream)
}

func (r *RealDockerClient) Close() error {
	return r.client.Close()
}

// MockDockerClient is a mock implementation of DockerClient for testing.
type MockDockerClient struct {
	InspectResponse dtypes.ContainerJSON
	InspectError    error
	StopError       error
	RemoveError     error
	ListResponse    []dtypes.Container
	ListError       error
	PullResponse    io.ReadCloser
	PullError       error
	StatsResponse   dtypes.ContainerStats
	StatsError      error
	CloseError      error
}

// ContainerInspect returns a mock container inspection result.
func (m *MockDockerClient) ContainerInspect(ctx context.Context, appId *big.Int) (dtypes.ContainerJSON, error) {
	return m.InspectResponse, m.InspectError
}

// ContainerStop simulates stopping a container.
func (m *MockDockerClient) ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error {
	return m.StopError
}

// ContainerRemove simulates removing a container.
func (m *MockDockerClient) ContainerRemove(ctx context.Context, containerID string, options ctypes.RemoveOptions) error {
	return m.RemoveError
}

// ContainerList returns a mock list of containers.
func (m *MockDockerClient) ContainerList(ctx context.Context, options ctypes.ListOptions) ([]dtypes.Container, error) {
	return m.ListResponse, m.ListError
}

// ImagePull simulates pulling an image.
func (m *MockDockerClient) ImagePull(ctx context.Context, ref string, options itypes.PullOptions) (io.ReadCloser, error) {
	return m.PullResponse, m.PullError
}

// ContainerStats returns mock container stats.
func (m *MockDockerClient) ContainerStats(ctx context.Context, containerID string, stream bool) (dtypes.ContainerStats, error) {
	return m.StatsResponse, m.StatsError
}

// Close simulates closing the client.
func (m *MockDockerClient) Close() error {
	return m.CloseError
}
