package v1

import (
	dtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/deployment/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"
)

//go:generate mockery --name ReservationGroup --output ./mocks
type ReservationGroup interface {
	Resources() dtypes.ResourceGroup
	SetAllocatedResources(dtypes.ResourceUnits)
	GetAllocatedResources() dtypes.ResourceUnits
	SetClusterParams(interface{})
	ClusterParams() interface{}
}

// Reservation interface implements orders and resources
//
//go:generate mockery --name Reservation --output ./mocks
type Reservation interface {
	OrderID() mtypes.OrderID
	Allocated() bool
	ReservationGroup
}
