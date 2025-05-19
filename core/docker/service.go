package docker

import (
	dockerCli "github.com/docker/docker/client"
	"github.com/unicornultrafoundation/subnet-node/config"
)

// var log = logrus.WithField("service", "docker")

// DockerService is a service to handle Docker interaction
type Service struct {
	client DockerClient
}

// NewDockerService initializes a new DockerService
func NewDockerService(cfg *config.C) (*Service, error) {
	client, err := dockerCli.NewClientWithOpts(dockerCli.FromEnv, dockerCli.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	s := Service{
		client: &RealDockerClient{client: client},
	}
	return &s, nil
}

// GetClient retrieves the ethclient instance
func (s *Service) GetClient() *DockerClient {
	return &s.client
}
