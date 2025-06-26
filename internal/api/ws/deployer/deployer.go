package deployer

import (
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/deployer"
)

// Common WebSocket message codes for deployment operations
const (
	DeploymentShellCodeStdout         = 100
	DeploymentShellCodeStderr         = 101
	DeploymentShellCodeResult         = 102
	DeploymentShellCodeFailure        = 103
	DeploymentShellCodeStdin          = 104
	DeploymentShellCodeTerminalResize = 105
)

// DeployerWSAPI is the API for the deployer web socket
type DeployerWSAPI struct {
	deployerService *deployer.Service
	logger          *logrus.Logger
}

func NewDeployerWSAPI(deployerService *deployer.Service) *DeployerWSAPI {
	return &DeployerWSAPI{
		deployerService: deployerService,
		logger:          logrus.New(),
	}
}
