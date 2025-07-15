package manifest

import (
	dtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/deployment/v1"

	maniv2beta1 "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/manifest/v1"
)

type submitRequest struct {
	Deployment dtypes.DeploymentID  `json:"deployment"`
	Manifest   maniv2beta1.Manifest `json:"manifest"`
}
