package apitypes

import (
	apclient "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/provider/client"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"
)

type ManifestGroup struct {
	Group    string                  `json:"group"`
	Services *apclient.ServiceStatus `json:"services"`
}

type DeploymentStatus struct {
	LeaseID        mtypes.LeaseID  `json:"lease_id"`
	ManifestGroups []ManifestGroup `json:"manifest_groups"`
}
