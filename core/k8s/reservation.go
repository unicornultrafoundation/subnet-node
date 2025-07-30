package k8s

import (
	atypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/base/v1"
	dtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/deployment/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"

	ctypes "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1"
	"github.com/unicornultrafoundation/subnet-node/core/k8s/util"
)

func newReservation(order mtypes.OrderID, resources dtypes.ResourceGroup) *reservation {
	return &reservation{
		order:            order,
		resources:        resources,
		endpointQuantity: util.GetEndpointQuantityOfResourceGroup(resources, atypes.Endpoint_LEASED_IP)}
}

type reservation struct {
	order             mtypes.OrderID
	resources         dtypes.ResourceGroup
	adjustedResources dtypes.ResourceUnits
	clusterParams     interface{}
	endpointQuantity  uint
	allocated         bool
}

var _ ctypes.Reservation = (*reservation)(nil)

func (r *reservation) OrderID() mtypes.OrderID {
	return r.order
}

func (r *reservation) Resources() dtypes.ResourceGroup {
	return r.resources
}

func (r *reservation) SetAllocatedResources(val dtypes.ResourceUnits) {
	r.adjustedResources = val
}

func (r *reservation) GetAllocatedResources() dtypes.ResourceUnits {
	return r.adjustedResources
}

func (r *reservation) SetClusterParams(val interface{}) {
	r.clusterParams = val
}

func (r *reservation) ClusterParams() interface{} {
	return r.clusterParams
}

func (r *reservation) Allocated() bool {
	return r.allocated
}
