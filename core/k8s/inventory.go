package k8s

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/boz/go-lifecycle"
	"github.com/desertbit/timer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/pubsub"

	tpubsub "github.com/troian/pubsub"
	ptypes "github.com/unicornultrafoundation/subnet-node/pkg/k8s/types"
	atypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/base/v1"
	dtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/deployment/v1"
	inventoryV1 "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/inventory/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"
	provider "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/provider/v1"

	sdlutil "github.com/unicornultrafoundation/subnet-node/pkg/k8s/sdl/util"

	ctypes "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1"
	cinventory "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/clients/inventory"
	cfromctx "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/fromctx"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/event"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/tools/fromctx"
)

var (
	// errReservationNotFound is the new error with message "not found"
	errReservationNotFound      = errors.New("reservation not found")
	errInventoryNotAvailableYet = errors.New("inventory status not available yet")
)

type invSnapshotResp struct {
	res *provider.Inventory
	err error
}

type inventoryRequest struct {
	order     mtypes.OrderID
	resources dtypes.ResourceGroup
	ch        chan<- inventoryResponse
}

type inventoryResponse struct {
	value ctypes.Reservation
	err   error
}

type inventoryService struct {
	config                 Config
	client                 Client
	sub                    pubsub.Subscriber
	statusch               chan chan<- inventoryV1.InventoryMetrics
	statusV1ch             chan chan<- invSnapshotResp
	lookupch               chan inventoryRequest
	reservech              chan inventoryRequest
	unreservech            chan inventoryRequest
	reservationCount       int64
	readych                chan struct{}
	log                    *logrus.Logger
	lc                     lifecycle.Lifecycle
	availableExternalPorts uint

	clients struct {
		inventory cinventory.Client
	}
}

func newInventoryService(
	ctx context.Context,
	config Config,
	log *logrus.Logger,
	sub pubsub.Subscriber,
	client Client,
	deployments []ctypes.IDeployment,
) (*inventoryService, error) {
	sub, err := sub.Clone()
	if err != nil {
		return nil, err
	}

	is := &inventoryService{
		config:                 config,
		client:                 client,
		sub:                    sub,
		statusch:               make(chan chan<- inventoryV1.InventoryMetrics),
		statusV1ch:             make(chan chan<- invSnapshotResp),
		lookupch:               make(chan inventoryRequest),
		reservech:              make(chan inventoryRequest),
		unreservech:            make(chan inventoryRequest),
		readych:                make(chan struct{}),
		log:                    log.WithField("module", "inventory").Logger,
		lc:                     lifecycle.New(),
		availableExternalPorts: config.InventoryExternalPortQuantity,
	}

	is.clients.inventory = cfromctx.ClientInventoryFromContext(ctx)

	reservations := make([]*reservation, 0, len(deployments))
	for _, d := range deployments {
		res := newReservation(d.LeaseID().OrderID(), d.ManifestGroup())
		res.SetClusterParams(d.ClusterParams())

		reservations = append(reservations, res)
	}

	go is.lc.WatchChannel(ctx.Done())
	go is.run(ctx, reservations)

	return is, nil
}

func (is *inventoryService) done() <-chan struct{} {
	return is.lc.Done()
}

func (is *inventoryService) ready() <-chan struct{} {
	return is.readych
}

func (is *inventoryService) lookup(order mtypes.OrderID, resources dtypes.ResourceGroup) (ctypes.Reservation, error) {
	ch := make(chan inventoryResponse, 1)
	req := inventoryRequest{
		order:     order,
		resources: resources,
		ch:        ch,
	}

	select {
	case is.lookupch <- req:
		response := <-ch
		return response.value, response.err
	case <-is.lc.ShuttingDown():
		return nil, ErrNotRunning
	}
}

func (is *inventoryService) reserve(order mtypes.OrderID, resources dtypes.ResourceGroup) (ctypes.Reservation, error) {
	for idx, res := range resources.GetResourceUnits() {
		if res.CPU == nil {
			return nil, fmt.Errorf("%w: CPU resource at idx %d is nil", ErrInvalidResource, idx)
		}
		if res.GPU == nil {
			return nil, fmt.Errorf("%w: GPU resource at idx %d is nil", ErrInvalidResource, idx)
		}
		if res.Memory == nil {
			return nil, fmt.Errorf("%w: Memory resource at idx %d is nil", ErrInvalidResource, idx)
		}
	}

	ch := make(chan inventoryResponse, 1)
	req := inventoryRequest{
		order:     order,
		resources: resources,
		ch:        ch,
	}

	select {
	case is.reservech <- req:
		response := <-ch
		if response.err == nil {
			cnt := atomic.AddInt64(&is.reservationCount, 1)
			is.log.WithField("cnt", cnt).Debug("Reservation count")
		}
		return response.value, response.err
	case <-is.lc.ShuttingDown():
		return nil, ErrNotRunning
	}
}

func (is *inventoryService) unreserve(order mtypes.OrderID) error { // nolint: golint,unparam
	ch := make(chan inventoryResponse, 1)
	req := inventoryRequest{
		order: order,
		ch:    ch,
	}

	select {
	case is.unreservech <- req:
		response := <-ch
		if response.err == nil {
			cnt := atomic.AddInt64(&is.reservationCount, -1)
			is.log.WithField("cnt", cnt).Debug("Reservation count")
		}
		return response.err
	case <-is.lc.ShuttingDown():
		return ErrNotRunning
	}
}

func (is *inventoryService) status(ctx context.Context) (inventoryV1.InventoryMetrics, error) {
	ch := make(chan inventoryV1.InventoryMetrics, 1)

	select {
	case <-is.lc.Done():
		return inventoryV1.InventoryMetrics{}, ErrNotRunning
	case <-ctx.Done():
		return inventoryV1.InventoryMetrics{}, ctx.Err()
	case is.statusch <- ch:
	}

	select {
	case <-is.lc.Done():
		return inventoryV1.InventoryMetrics{}, ErrNotRunning
	case <-ctx.Done():
		return inventoryV1.InventoryMetrics{}, ctx.Err()
	case result := <-ch:
		return result, nil
	}
}

func (is *inventoryService) statusV1(ctx context.Context) (*provider.Inventory, error) {
	ch := make(chan invSnapshotResp, 1)

	select {
	case <-is.lc.Done():
		return nil, ErrNotRunning
	case <-ctx.Done():
		return nil, ctx.Err()
	case is.statusV1ch <- ch:
	}

	select {
	case <-is.lc.Done():
		return nil, ErrNotRunning
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-ch:
		return result.res, result.err
	}
}

func (is *inventoryService) resourcesToCommit(rgroup dtypes.ResourceGroup) dtypes.ResourceGroup {
	replacedResources := make(dtypes.ResourceUnits, 0)

	for _, resource := range rgroup.GetResourceUnits() {
		runits := atypes.Resources{
			ID: resource.ID,
			CPU: &atypes.CPU{
				Units:      sdlutil.ComputeCommittedResources(is.config.CPUCommitLevel, resource.Resources.GetCPU().GetUnits()),
				Attributes: resource.Resources.GetCPU().GetAttributes(),
			},
			GPU: &atypes.GPU{
				Units:      sdlutil.ComputeCommittedResources(is.config.GPUCommitLevel, resource.Resources.GetGPU().GetUnits()),
				Attributes: resource.Resources.GetGPU().GetAttributes(),
			},
			Memory: &atypes.Memory{
				Quantity:   sdlutil.ComputeCommittedResources(is.config.MemoryCommitLevel, resource.Resources.GetMemory().GetQuantity()),
				Attributes: resource.Resources.GetMemory().GetAttributes(),
			},
			Endpoints: resource.Resources.GetEndpoints(),
		}

		storage := make(atypes.Volumes, 0, len(resource.Resources.GetStorage()))

		for _, volume := range resource.Resources.GetStorage() {
			storage = append(storage, atypes.Storage{
				Name:       volume.Name,
				Quantity:   sdlutil.ComputeCommittedResources(is.config.StorageCommitLevel, volume.GetQuantity()),
				Attributes: volume.GetAttributes(),
			})
		}

		runits.Storage = storage

		v := dtypes.ResourceUnit{
			Resources: runits,
			Count:     resource.Count,
		}

		replacedResources = append(replacedResources, v)
	}

	result := dtypes.GroupSpec{
		Name:         rgroup.GetName(),
		Requirements: atypes.PlacementRequirements{},
		Resources:    replacedResources,
	}

	return result
}

type inventoryServiceState struct {
	inventory    ctypes.Inventory
	reservations []*reservation
}

func (is *inventoryService) handleRequest(req inventoryRequest, state *inventoryServiceState) {
	// convert the resources to the committed amount
	resourcesToCommit := is.resourcesToCommit(req.resources)
	// create new registration if capacity available
	reservation := newReservation(req.order, resourcesToCommit)

	{
		jReservation, _ := json.Marshal(req.resources.GetResourceUnits())
		is.log.WithField("order", req.order).WithField("resources", jReservation).Debug("Reservation requested")
	}

	err := state.inventory.Adjust(reservation)
	if err != nil {
		is.log.WithField("order", req.order).Info("Insufficient capacity for reservation")
		req.ch <- inventoryResponse{err: err}
		return
	}

	// Add the reservation to the list
	state.reservations = append(state.reservations, reservation)
	req.ch <- inventoryResponse{value: reservation}

}

func (is *inventoryService) run(ctx context.Context, reservationsArg []*reservation) {
	defer is.lc.ShutdownCompleted()
	defer is.sub.Close()

	state := &inventoryServiceState{
		inventory:    nil,
		reservations: reservationsArg,
	}
	is.log.WithField("qty", len(state.reservations)).Info("Starting with existing reservations")

	var currinv ctypes.Inventory

	invupch := make(chan ctypes.Inventory, 1)

	invch := is.clients.inventory.ResultChan()
	var reservech <-chan inventoryRequest

	t := timer.NewStoppedTimer()

	resumeRevesech := func() {
		if reservech == nil && state.inventory != nil {
			reservech = is.reservech
		}
	}

	bus := fromctx.MustPubSubFromCtx(ctx)

	signalch := make(chan struct{}, 1)
	trySignal := func() {
		select {
		case signalch <- struct{}{}:
		case <-is.lc.ShutdownRequest():
		default:
		}
	}
loop:
	for {
		select {
		case err := <-is.lc.ShutdownRequest():
			is.log.WithError(err).Debug("received shutdown request")
			is.lc.ShutdownInitiated(err)
			break loop
		case ev := <-is.sub.Events():
			switch ev := ev.(type) { // nolint: gocritic
			case event.ClusterDeployment:
				// mark reservation allocated if deployment successful
				for _, res := range state.reservations {
					if !res.OrderID().Equals(ev.LeaseID.OrderID()) {
						continue
					}
					if res.Resources().GetName() != ev.Group.Name {
						continue
					}

					allocatedPrev := res.allocated
					res.allocated = ev.Status == event.ClusterDeploymentDeployed

					if res.allocated != allocatedPrev {
						externalPortCount := reservationCountEndpoints(res)
						if ev.Status == event.ClusterDeploymentDeployed {
							is.availableExternalPorts -= externalPortCount
						} else {
							is.availableExternalPorts += externalPortCount
						}

						is.log.WithField("order", res.OrderID()).WithField("resource-group", res.Resources().GetName()).WithField("allocated", res.allocated).Debug("Reservation status update")

						if currinv != nil {
							select {
							case invupch <- currinv:
							default:
							}
						}
					}

					break
				}

				resumeRevesech()
			}
		case <-t.C:
			resumeRevesech()
		case req := <-reservech:
			is.handleRequest(req, state)
		case req := <-is.lookupch:
			// lookup registration
			for _, res := range state.reservations {
				if !res.OrderID().Equals(req.order) {
					continue
				}
				if res.Resources().GetName() != req.resources.GetName() {
					continue
				}
				req.ch <- inventoryResponse{value: res}
				continue loop
			}

			req.ch <- inventoryResponse{err: errReservationNotFound}
		case req := <-is.unreservech:
			is.log.WithField("order", req.order).Info("Attempting to remove reservation")

			for idx, res := range state.reservations {
				if !res.OrderID().Equals(req.order) {
					continue
				}

				state.reservations = append(state.reservations[:idx], state.reservations[idx+1:]...)
				// reclaim availableExternalPorts if unreserving allocated resources
				if res.allocated {
					is.availableExternalPorts += reservationCountEndpoints(res)
				}

				req.ch <- inventoryResponse{value: res}
				is.log.WithField("order", req.order).Info("Unreserve capacity complete")
				continue loop
			}

			req.ch <- inventoryResponse{err: errReservationNotFound}
		case responseCh := <-is.statusch:
			select {
			case responseCh <- is.getStatus(state):
			default:
			}
			// inventoryRequestsCounter.WithLabelValues("status", "success").Inc()
		case responseCh := <-is.statusV1ch:
			resp, err := is.getStatusV1(state)
			select {
			case responseCh <- invSnapshotResp{
				res: resp,
				err: err,
			}:
			default:
			}
		case inv := <-invch:
			if inv == nil {
				continue
			}

			select {
			case <-invupch:
			default:
			}

			invupch <- inv
		case inv := <-invupch:
			currinv = inv.Dup()
			state.inventory = inv

			resumeRevesech()

			// readjust inventory accordingly with pending leases
			for _, r := range state.reservations {
				if !r.allocated {
					if err := state.inventory.Adjust(r); err != nil {
						is.log.WithError(err).Error("Adjust inventory for pending reservation")
					}
				}
			}

			trySignal()
		case <-signalch:
			inv, err := is.getStatusV1(state)
			if err != nil {
				continue
			}

			bus.Pub(inv, []string{ptypes.PubSubTopicInventoryStatus}, tpubsub.WithRetain())
		}
	}
	is.log.Debug("Shutdown complete")
}

func (is *inventoryService) getStatus(state *inventoryServiceState) inventoryV1.InventoryMetrics {
	status := inventoryV1.InventoryMetrics{}

	if state.inventory == nil {
		status.Error = errInventoryNotAvailableYet
		return status
	}

	for _, reservation := range state.reservations {
		total := inventoryV1.MetricTotal{
			Storage: make(map[string]int64),
		}

		for _, resources := range reservation.Resources().GetResourceUnits() {
			total.AddResources(resources)
		}

		if reservation.allocated {
			status.Active = append(status.Active, total)
		} else {
			status.Pending = append(status.Pending, total)
		}
	}

	status.Available.Nodes = append(status.Available.Nodes, state.inventory.Metrics().Nodes...)

	for class, size := range state.inventory.Metrics().TotalAvailable.Storage {
		status.Available.Storage = append(status.Available.Storage, inventoryV1.StorageStatus{Class: class, Size: size})
	}

	return status
}

func (is *inventoryService) getStatusV1(state *inventoryServiceState) (*provider.Inventory, error) {
	if state.inventory == nil {
		return nil, errInventoryNotAvailableYet
	}

	status := &provider.Inventory{
		Cluster: state.inventory.Snapshot(),
		Reservations: provider.Reservations{
			Pending: provider.ReservationsMetric{
				Count:     0,
				Resources: provider.NewResourcesMetric(),
			},
			Active: provider.ReservationsMetric{
				Count:     0,
				Resources: provider.NewResourcesMetric(),
			},
		},
	}

	for _, reservation := range state.reservations {
		runits := reservation.Resources().GetResourceUnits()
		if reservation.allocated {
			status.Reservations.Active.Resources.AddResourceUnits(runits)
			status.Reservations.Active.Count++
		} else {
			status.Reservations.Pending.Resources.AddResourceUnits(runits)
			status.Reservations.Pending.Count++
		}
	}

	return status, nil
}

func reservationCountEndpoints(reservation *reservation) uint {
	var externalPortCount uint

	resources := reservation.Resources().GetResourceUnits()
	// Count the number of endpoints per resource. The number of instances does not affect
	// the number of ports
	for _, resource := range resources {
		for _, endpoint := range resource.Resources.Endpoints {
			if endpoint.Kind == atypes.Endpoint_RANDOM_PORT {
				externalPortCount++
			}
		}
	}

	return externalPortCount
}
