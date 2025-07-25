package k8s

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

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

	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/runner"
	sdlutil "github.com/unicornultrafoundation/subnet-node/pkg/k8s/sdl/util"

	ctypes "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1"
	cinventory "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/clients/inventory"
	cip "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/clients/ip"
	cfromctx "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/fromctx"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/event"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/operator/waiter"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/tools/fromctx"
)

var (
	// errReservationNotFound is the new error with message "not found"
	errReservationNotFound      = errors.New("reservation not found")
	errInventoryNotAvailableYet = errors.New("inventory status not available yet")
	errInventoryReservation     = errors.New("inventory error")
	errNoLeasedIPsAvailable     = fmt.Errorf("%w: no leased IPs available", errInventoryReservation)
	errInsufficientIPs          = fmt.Errorf("%w: insufficient number of IPs", errInventoryReservation)
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
	waiter                 waiter.OperatorWaiter
	availableExternalPorts uint

	clients struct {
		ip        cip.Client
		inventory cinventory.Client
	}
}

func newInventoryService(
	ctx context.Context,
	config Config,
	log *logrus.Logger,
	sub pubsub.Subscriber,
	client Client,
	waiter waiter.OperatorWaiter,
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
		log:                    log.WithField("cmp", "inventory-service").Logger,
		lc:                     lifecycle.New(),
		availableExternalPorts: config.InventoryExternalPortQuantity,
		waiter:                 waiter,
	}

	is.clients.inventory = cfromctx.ClientInventoryFromContext(ctx)
	is.clients.ip = cfromctx.ClientIPFromContext(ctx)

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
			is.log.Debug("reservation count", "cnt", cnt)
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
			is.log.Debug("reservation count", "cnt", cnt)
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
	ipAddrUsage  cip.AddressUsage
	reservations []*reservation
}

func countPendingIPs(state *inventoryServiceState) uint {
	pending := uint(0)
	for _, entry := range state.reservations {
		if !entry.ipsConfirmed {
			pending += entry.endpointQuantity
		}
	}

	return pending
}

func (is *inventoryService) handleRequest(req inventoryRequest, state *inventoryServiceState) {
	// convert the resources to the committed amount
	resourcesToCommit := is.resourcesToCommit(req.resources)
	// create new registration if capacity available
	reservation := newReservation(req.order, resourcesToCommit)

	{
		jReservation, _ := json.Marshal(req.resources.GetResourceUnits())
		is.log.Debug(fmt.Sprintf("reservation requested. order=%s, resources=%s", req.order, jReservation))
	}

	if reservation.endpointQuantity != 0 {
		if is.clients.ip == nil {
			req.ch <- inventoryResponse{err: errNoLeasedIPsAvailable}
			return
		}
		// numIPUnused := state.ipAddrUsage.Available - state.ipAddrUsage.InUse
		// pending := countPendingIPs(state)
		// if reservation.endpointQuantity > (numIPUnused - pending) {
		// 	is.log.Info("insufficient number of IP addresses available", "order", req.order)
		// 	req.ch <- inventoryResponse{err: fmt.Errorf("%w: unable to reserve %d", errInsufficientIPs, reservation.endpointQuantity)}
		// 	return
		// }

		// is.log.Info("reservation used leased IPs", "used", reservation.endpointQuantity, "available", state.ipAddrUsage.Available, "in-use", state.ipAddrUsage.InUse, "pending", pending)
	} else {
		reservation.ipsConfirmed = true // No IPs, just mark it as confirmed implicitly
	}

	err := state.inventory.Adjust(reservation)
	if err != nil {
		is.log.Info("insufficient capacity for reservation", "order", req.order)
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

	rctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state := &inventoryServiceState{
		inventory:    nil,
		reservations: reservationsArg,
	}
	is.log.Info("starting with existing reservations", "qty", len(state.reservations))

	// wait on the operators to be ready
	err := is.waiter.WaitForAll(ctx)
	if err != nil {
		is.lc.ShutdownInitiated(err)
		return
	}

	var runch <-chan runner.Result
	var currinv ctypes.Inventory

	invupch := make(chan ctypes.Inventory, 1)

	invch := is.clients.inventory.ResultChan()
	var reservech <-chan inventoryRequest

	resumeProcessingReservations := func() {
		reservech = is.reservech
	}

	t := timer.NewStoppedTimer()

	updateIPs := func() {
		if is.clients.ip != nil {
			reservech = nil
			if runch == nil {
				t.Stop()
				runch = is.runCheck(rctx, state)
			}
		} else if reservech == nil && state.inventory != nil {
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
			is.log.Debug("received shutdown request", "err", err)
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

						is.log.Debug("reservation status update",
							"order", res.OrderID(),
							"resource-group", res.Resources().GetName(),
							"allocated", res.allocated)

						if currinv != nil {
							select {
							case invupch <- currinv:
							default:
							}
						}
					}

					break
				}

				updateIPs()
			}
		case <-t.C:
			updateIPs()
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
			is.log.Debug("unreserving capacity", "order", req.order)
			// remove reservation

			is.log.Info("attempting to removing reservation", "order", req.order)

			for idx, res := range state.reservations {
				if !res.OrderID().Equals(req.order) {
					continue
				}

				is.log.Info("removing reservation", "order", res.OrderID())

				state.reservations = append(state.reservations[:idx], state.reservations[idx+1:]...)
				// reclaim availableExternalPorts if unreserving allocated resources
				if res.allocated {
					is.availableExternalPorts += reservationCountEndpoints(res)
				}

				req.ch <- inventoryResponse{value: res}
				is.log.Info("unreserve capacity complete", "order", req.order)
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

			if err == nil {
				// inventoryRequestsCounter.WithLabelValues("status", "success").Inc()
			} else {
				// inventoryRequestsCounter.WithLabelValues("status", "error").Inc()
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

			updateIPs()

			// readjust inventory accordingly with pending leases
			for _, r := range state.reservations {
				if !r.allocated {
					if err := state.inventory.Adjust(r); err != nil {
						is.log.Error("adjust inventory for pending reservation", "error", err.Error())
					}
				}
			}

			trySignal()
		case run := <-runch:
			runch = nil
			t.Reset(5 * time.Second)
			if err := run.Error(); err == nil {
				res := run.Value().(runCheckResult)
				state.ipAddrUsage = res.ipResult

				// Process confirmed IP addresses usage
				for _, confirmedOrderID := range res.confirmedResult {
					for i, entry := range state.reservations {
						if entry.order.Equals(confirmedOrderID) {
							state.reservations[i].ipsConfirmed = true
							is.log.Info("confirmed IP allocation", "orderID", confirmedOrderID)
							break
						}
					}
				}
			} else {
				is.log.Error("checking IP addresses", "err", err)
			}

			resumeProcessingReservations()

			trySignal()
		case <-signalch:
			inv, err := is.getStatusV1(state)
			if err != nil {
				continue
			}

			bus.Pub(inv, []string{ptypes.PubSubTopicInventoryStatus}, tpubsub.WithRetain())
		}
	}
	is.log.Debug("shutting down")
	if runch != nil {
		<-runch
	}

	if is.clients.ip != nil {
		is.clients.ip.Stop()
	}

	is.log.Debug("shutdown complete")
}

type confirmationItem struct {
	orderID          mtypes.OrderID
	expectedQuantity uint
}

type runCheckResult struct {
	ipResult        cip.AddressUsage
	confirmedResult []mtypes.OrderID
}

func (is *inventoryService) runCheck(ctx context.Context, state *inventoryServiceState) <-chan runner.Result {
	// Look for unconfirmed IPs, these are IPs that have a deployment created
	// event and are marked allocated. But until the IP address operator has reported
	// that it has actually created the associated resources, we need to consider the total number of end
	// points as pending

	confirm := make([]confirmationItem, 0, len(state.reservations))

	for _, entry := range state.reservations {
		// Skip anything already confirmed or not allocated
		if entry.ipsConfirmed || !entry.allocated {
			continue
		}

		confirm = append(confirm, confirmationItem{
			orderID:          entry.OrderID(),
			expectedQuantity: entry.endpointQuantity,
		})
	}

	return runner.Do(func() runner.Result {
		retval := runCheckResult{}
		var err error

		retval.ipResult, err = is.clients.ip.GetIPAddressUsage(ctx)
		if err != nil {
			return runner.NewResult(nil, err)
		}

		for _, confirmItem := range confirm {
			status, err := is.clients.ip.GetIPAddressStatus(ctx, confirmItem.orderID)
			if err != nil {
				// This error is not really fatal, so don't bail on this entirely. The other results
				// retrieved in this code are still valid
				is.log.Error("failed checking IP address usage", "orderID", confirmItem.orderID, "error", err)
				break
			}

			numConfirmed := uint(len(status))
			if numConfirmed == confirmItem.expectedQuantity {
				retval.confirmedResult = append(retval.confirmedResult, confirmItem.orderID)
			}
		}

		return runner.NewResult(retval, nil)
	})
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

	for _, nd := range state.inventory.Metrics().Nodes {
		status.Available.Nodes = append(status.Available.Nodes, nd)
	}

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
