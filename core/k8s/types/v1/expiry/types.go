package expiry

import mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"

type DeploymentExpiryInfo struct {
	DeploymentID string
	Status       DeploymentExpiryStatus

	// in seconds, if the status is active, the time left is the time left until the deployment expires
	// if the status is expired, the time left is the time left until the deployment is deleted
	// if the status is deleted, the time left is 0
	// if the status is unknown, the time left is 0
	TimeLeft int64
}

type DeploymentExpiryStatus string

const (
	DeploymentExpiryStatusExpired DeploymentExpiryStatus = "expired"
	DeploymentExpiryStatusActive  DeploymentExpiryStatus = "active"
	DeploymentExpiryStatusDeleted DeploymentExpiryStatus = "deleted"
	DeploymentExpiryStatusUnknown DeploymentExpiryStatus = "unknown"
)

type DeploymentExpiry struct {
	LeaseID mtypes.LeaseID
	Status  DeploymentExpiryStatus
}
