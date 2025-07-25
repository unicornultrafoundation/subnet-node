package apitypes

import (
	apclient "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/provider/client"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"
)

type ManifestGroup struct {
	Group    string
	Services *apclient.ServiceStatus
}

type DeploymentStatus struct {
	LeaseID        mtypes.LeaseID
	ManifestGroups []ManifestGroup
}
