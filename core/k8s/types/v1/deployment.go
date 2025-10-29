package v1

import (
	maniv1 "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/manifest/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"
)

// IDeployment interface defined with LeaseID and ManifestGroup methods
//
//go:generate mockery --name IDeployment --output ./mocks
type IDeployment interface {
	LeaseID() mtypes.LeaseID
	ManifestGroup() *maniv1.Group
	ClusterParams() interface{}
	ResourceVersion() string
}

type Deployment struct {
	Lid         mtypes.LeaseID
	MGroup      *maniv1.Group
	CParams     interface{}
	ResourceVer string
}

var _ IDeployment = (*Deployment)(nil)

func (d *Deployment) LeaseID() mtypes.LeaseID {
	return d.Lid
}

func (d *Deployment) ManifestGroup() *maniv1.Group {
	return d.MGroup
}

func (d *Deployment) ClusterParams() interface{} {
	return d.CParams
}

func (d *Deployment) ResourceVersion() string {
	return d.ResourceVer
}
