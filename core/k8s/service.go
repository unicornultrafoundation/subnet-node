package k8s

import (
	"context"
	"io"

	"github.com/boz/go-lifecycle"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	tpubsub "github.com/troian/pubsub"
	apclient "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/provider/client"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/pkg/errors"

	dtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/deployment/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"
	provider "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/provider/v1"

	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/apitypes"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/pubsub"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/sdl"

	"github.com/unicornultrafoundation/subnet-node/core/k8s/manifest"
	ctypes "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1"
	crd "github.com/unicornultrafoundation/subnet-node/pkg/k8s/apis/subnet.node/v1"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/event"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/session"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/tools/fromctx"
	ptypes "github.com/unicornultrafoundation/subnet-node/pkg/k8s/types"
	maniv1 "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/manifest/v1"
)

// ErrNotRunning is the error when service is not running
var (
	ErrNotRunning      = errors.New("not running")
	ErrInvalidResource = errors.New("invalid resource")
	errNoManifestGroup = errors.New("no manifest group could be found")
)

type service struct {
	session session.Session
	client  Client
	bus     pubsub.Bus
	sub     pubsub.Subscriber

	inventory *inventoryService
	// hostnames *hostnameService

	checkDeploymentExistsRequestCh chan checkDeploymentExistsRequest
	statusch                       chan chan<- *apclient.ClusterStatus
	statusV1ch                     chan chan<- uint32
	managers                       map[mtypes.LeaseIDKey]*deploymentManager

	managerch chan *deploymentManager

	log *logrus.Logger
	lc  lifecycle.Lifecycle

	config Config

	manifestService *manifest.Service
}

type checkDeploymentExistsRequest struct {
	owner common.Address
	dseq  uint64
	gseq  uint32

	responseCh chan<- mtypes.LeaseID
}

// Cluster is the interface that wraps Reserve and Unreserve methods
//
//go:generate mockery --name Cluster
type Cluster interface {
	Reserve(mtypes.OrderID, dtypes.ResourceGroup) (ctypes.Reservation, error)
	Unreserve(mtypes.OrderID) error
}

// StatusClient is the interface which includes status of service
type StatusClient interface {
	Status(context.Context) (*apclient.ClusterStatus, error)
	StatusV1(context.Context) (*provider.ClusterStatus, error)
	FindActiveLease(ctx context.Context, owner common.Address, dseq uint64, gseq uint32) (bool, mtypes.LeaseID, crd.ManifestGroup, error)
}

// Service manage compute cluster for the provider.  Will eventually integrate with kubernetes, etc...
//
//go:generate mockery --name Service
type Service interface {
	StatusClient
	Cluster
	Close() error
	Ready() <-chan struct{}
	Done() <-chan struct{}

	// RequestDeployment requests a deployment to be created
	RequestDeployment(ctx context.Context, deploymentID dtypes.DeploymentID, sdlManifest sdl.SDL) error

	// GetAllLeaseStatus returns the status of all leases
	GetAllLeaseStatus(ctx context.Context) ([]apitypes.DeploymentStatus, error)

	// DeleteDeployment deletes a deployment
	DeleteDeployment(lid mtypes.LeaseID) error

	// GetLeaseStatus returns the status of a lease
	GetLeaseStatus(ctx context.Context, leaseID mtypes.LeaseID) (apclient.LeaseStatus, error)

	// Exec executes a command in a lease
	Exec(ctx context.Context,
		lID mtypes.LeaseID,
		service string,
		podIndex uint,
		cmd []string,
		stdin io.Reader,
		stdout io.Writer,
		stderr io.Writer,
		tty bool,
		tsq remotecommand.TerminalSizeQueue) (ctypes.ExecResult, error)

	// ServiceStatus returns the status of a service
	ServiceStatus(ctx context.Context, leaseID mtypes.LeaseID, service string) (*apclient.ServiceStatus, error)

	// LeaseLogs returns the logs of a lease
	LeaseLogs(context.Context, mtypes.LeaseID, string, bool, *int64) ([]*ctypes.ServiceLog, error)

	// GetManifestGroup returns the manifest group of a lease
	GetManifestGroup(ctx context.Context, leaseID mtypes.LeaseID) (bool, crd.ManifestGroup, error)
}

// NewService returns new Service instance
func NewService(
	ctx context.Context,
	session session.Session,
	bus pubsub.Bus,
	client Client,
	cfg Config,
) (Service, error) {
	log := session.Log().WithField("module", "provider-cluster").WithField("cmp", "service").Logger

	lc := lifecycle.New()

	sub, err := bus.Subscribe()
	if err != nil {
		return nil, err
	}

	deployments, err := findDeployments(ctx, log, client)
	if err != nil {
		sub.Close()
		return nil, err
	}

	inventory, err := newInventoryService(ctx, cfg, log, sub, client, deployments)
	if err != nil {
		sub.Close()
		return nil, err
	}

	manifestService := manifest.NewService(bus, log, session.Provider().Address())

	s := &service{
		session: session,
		client:  client,
		// hostnames:                      hostnames,
		bus:                            bus,
		sub:                            sub,
		inventory:                      inventory,
		statusch:                       make(chan chan<- *apclient.ClusterStatus),
		statusV1ch:                     make(chan chan<- uint32),
		managers:                       make(map[mtypes.LeaseIDKey]*deploymentManager),
		managerch:                      make(chan *deploymentManager),
		checkDeploymentExistsRequestCh: make(chan checkDeploymentExistsRequest),
		log:                            log.WithField("service", "k8s").Logger,
		lc:                             lc,
		config:                         cfg,
		manifestService:                manifestService,
	}

	go s.lc.WatchContext(ctx)
	go s.run(ctx, deployments)

	return s, nil
}

func (s *service) FindActiveLease(ctx context.Context, owner common.Address, dseq uint64, gseq uint32) (bool, mtypes.LeaseID, crd.ManifestGroup, error) {
	response := make(chan mtypes.LeaseID, 1)
	req := checkDeploymentExistsRequest{
		responseCh: response,
		dseq:       dseq,
		gseq:       gseq,
		owner:      owner,
	}
	select {
	case s.checkDeploymentExistsRequestCh <- req:
	case <-ctx.Done():
		return false, mtypes.LeaseID{}, crd.ManifestGroup{}, ctx.Err()
	}

	var leaseID mtypes.LeaseID
	var ok bool
	select {
	case leaseID, ok = <-response:
		if !ok {
			return false, mtypes.LeaseID{}, crd.ManifestGroup{}, nil
		}

	case <-ctx.Done():
		return false, mtypes.LeaseID{}, crd.ManifestGroup{}, ctx.Err()
	}

	found, mgroup, err := s.client.GetManifestGroup(ctx, leaseID)
	if err != nil {
		return false, mtypes.LeaseID{}, crd.ManifestGroup{}, err
	}

	if !found {
		return false, mtypes.LeaseID{}, crd.ManifestGroup{}, errNoManifestGroup
	}

	return true, leaseID, mgroup, nil
}

func (s *service) Close() error {
	s.lc.Shutdown(nil)
	return s.lc.Error()
}

func (s *service) Done() <-chan struct{} {
	return s.lc.Done()
}

func (s *service) Ready() <-chan struct{} {
	return s.inventory.ready()
}

func (s *service) Reserve(order mtypes.OrderID, resources dtypes.ResourceGroup) (ctypes.Reservation, error) {
	return s.inventory.reserve(order, resources)
}

func (s *service) Unreserve(order mtypes.OrderID) error {
	return s.inventory.unreserve(order)
}

func (s *service) Status(ctx context.Context) (*apclient.ClusterStatus, error) {
	istatus, err := s.inventory.status(ctx)
	if err != nil {
		return nil, err
	}

	ch := make(chan *apclient.ClusterStatus, 1)

	select {
	case <-s.lc.Done():
		return nil, ErrNotRunning
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.statusch <- ch:
	}

	select {
	case <-s.lc.Done():
		return nil, ErrNotRunning
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-ch:
		result.Inventory = istatus
		return result, nil
	}
}

func (s *service) StatusV1(ctx context.Context) (*provider.ClusterStatus, error) {
	istatus, err := s.inventory.statusV1(ctx)
	if err != nil {
		return nil, err
	}

	ch := make(chan uint32, 1)

	select {
	case <-s.lc.Done():
		return nil, ErrNotRunning
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.statusV1ch <- ch:
	}

	select {
	case <-s.lc.Done():
		return nil, ErrNotRunning
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-ch:
		res := &provider.ClusterStatus{
			Leases: provider.Leases{
				Active: result,
			},
			Inventory: *istatus,
		}

		return res, nil
	}
}

func (s *service) Exec(ctx context.Context,
	lID mtypes.LeaseID,
	service string,
	podIndex uint,
	cmd []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	tty bool,
	tsq remotecommand.TerminalSizeQueue) (ctypes.ExecResult, error) {

	return s.client.Exec(ctx, lID, service, podIndex, cmd, stdin, stdout, stderr, tty, tsq)
}

func (s *service) LeaseLogs(ctx context.Context, leaseID mtypes.LeaseID, service string, follow bool, tailLines *int64) ([]*ctypes.ServiceLog, error) {
	return s.client.LeaseLogs(ctx, leaseID, service, follow, tailLines)
}

func (s *service) ServiceStatus(ctx context.Context, leaseID mtypes.LeaseID, service string) (*apclient.ServiceStatus, error) {
	return s.client.ServiceStatus(ctx, leaseID, service)
}

func (s *service) GetManifestGroup(ctx context.Context, leaseID mtypes.LeaseID) (bool, crd.ManifestGroup, error) {
	return s.client.GetManifestGroup(ctx, leaseID)
}

func (s *service) run(ctx context.Context, deployments []ctypes.IDeployment) {
	defer s.lc.ShutdownCompleted()
	defer s.sub.Close()

	bus := fromctx.MustPubSubFromCtx(ctx)

	inventorych := bus.Sub(ptypes.PubSubTopicInventoryStatus)

	for _, deployment := range deployments {
		s.managers[mtypes.LeaseIDToKey(deployment.LeaseID())] = newDeploymentManager(s, deployment, false)
	}

	signalch := make(chan struct{}, 1)

	trySignal := func() {
		select {
		case signalch <- struct{}{}:
		case <-ctx.Done():
		default:
		}
	}

	trySignal()

loop:
	for {
		select {
		case err := <-s.lc.ShutdownRequest():
			s.log.Debug("received shutdown request", "err", err)
			s.lc.ShutdownInitiated(err)
			break loop
		case ev := <-s.sub.Events():
			switch ev := ev.(type) {
			case event.ManifestReceived:
				s.log.WithField("lease", ev.LeaseID).Info("Manifest received")

				mgroup := ev.ManifestGroup()
				if mgroup == nil {
					s.log.WithField("lease", ev.LeaseID).WithField("group-name", ev.Group.GroupSpec.Name).Error("Indeterminate manifest group")
					break
				}

				getDeployment := func(leaseID mtypes.LeaseID, mgroup *maniv1.Group) (*ctypes.Deployment, error) {
					reservation, err := s.inventory.lookup(leaseID.OrderID(), mgroup)
					if err != nil {
						return nil, err
					}

					deployment := &ctypes.Deployment{
						Lid:     leaseID,
						MGroup:  mgroup,
						CParams: reservation.ClusterParams(),
					}

					return deployment, nil
				}

				key := mtypes.LeaseIDToKey(ev.LeaseID)
				if manager := s.managers[key]; manager != nil {
					// If the lease is already managed, update the deployment
					deployment, err := getDeployment(ev.LeaseID, mgroup)
					if err != nil {
						s.log.WithField("lease", ev.LeaseID).WithField("group-name", mgroup.Name).WithField("err", err).Error("Error getting deployment")
						break
					}

					if err := manager.update(deployment); err != nil {
						s.log.WithField("lease", ev.LeaseID).WithField("group-name", mgroup.Name).WithField("err", err).Error("Error updating deployment")
					}
					break
				}

				// If the lease is new, reserve the inventory
				_, err := s.Reserve(ev.LeaseID.OrderID(), mgroup)
				if err != nil {
					s.log.WithField("lease", ev.LeaseID).WithField("group-name", mgroup.Name).WithField("err", err).Error("Error reserving inventory")
					break
				}

				// Create a new deployment manager
				deployment, err := getDeployment(ev.LeaseID, mgroup)
				if err != nil {
					s.log.WithField("lease", ev.LeaseID).WithField("group-name", mgroup.Name).WithField("err", err).Error("Error getting deployment")
					break
				}
				s.managers[key] = newDeploymentManager(s, deployment, true)

				trySignal()
			case mtypes.EventLeaseClosed:
				_ = s.bus.Publish(event.LeaseRemoveFundsMonitor{LeaseID: ev.ID})
				s.teardownLease(ev.ID)
			}
		case ch := <-s.statusch:
			ch <- &apclient.ClusterStatus{
				Leases: uint32(len(s.managers)), // nolint: gosec
			}
		case ch := <-s.statusV1ch:
			ch <- uint32(len(s.managers)) // nolint: gosec
		case <-signalch:
			istatus, _ := s.inventory.statusV1(ctx)

			if istatus == nil {
				continue
			}

			msg := provider.ClusterStatus{
				Leases:    provider.Leases{Active: uint32(len(s.managers))}, // nolint: gosec
				Inventory: *istatus,
			}
			bus.Pub(msg, []string{ptypes.PubSubTopicClusterStatus}, tpubsub.WithRetain())
		case update := <-inventorych:
			inv, valid := update.(*provider.Inventory)
			if !valid {
				continue
			}

			msg := provider.ClusterStatus{
				Leases:    provider.Leases{Active: uint32(len(s.managers))}, // nolint: gosec
				Inventory: *inv,
			}
			bus.Pub(msg, []string{ptypes.PubSubTopicClusterStatus}, tpubsub.WithRetain())
		case dm := <-s.managerch:
			s.log.Info("manager done", "lease", dm.deployment.LeaseID())

			// unreserve resources
			if err := s.inventory.unreserve(dm.deployment.LeaseID().OrderID()); err != nil {
				s.log.WithField("lease", dm.deployment.LeaseID()).WithField("err", err).Error("Error unreserving inventory")
			}

			delete(s.managers, mtypes.LeaseIDToKey(dm.deployment.LeaseID()))
			trySignal()
		case req := <-s.checkDeploymentExistsRequestCh:
			s.doCheckDeploymentExists(req)
		}
	}

	s.log.Debug("draining deployment managers...", "qty", len(s.managers))
	for _, manager := range s.managers {
		if manager != nil {
			manager := <-s.managerch
			s.log.Debug("manager done", "lease", manager.deployment.LeaseID())
		}
	}

	<-s.inventory.done()
	s.session.Log().Info("shutdown complete")
}

func (s *service) doCheckDeploymentExists(req checkDeploymentExistsRequest) {
	for leaseID := range s.managers {
		// Check for a match
		if leaseID.GSeq == req.gseq && leaseID.DSeq == req.dseq && leaseID.Owner == req.owner.String() {
			req.responseCh <- leaseID.ToLeaseID()
			return
		}
	}

	close(req.responseCh)
}

func (s *service) teardownLease(lid mtypes.LeaseID) {
	if manager := s.managers[mtypes.LeaseIDToKey(lid)]; manager != nil {
		if err := manager.teardown(); err != nil {
			s.log.Error("tearing down lease deployment", "err", err, "lease", lid)
		}
		return
	}

	// unreserve resources if no manager present yet.
	if lid.Provider == s.session.Provider().Owner {
		s.log.Info("unreserving unmanaged order", "lease", lid)
		err := s.inventory.unreserve(lid.OrderID())
		if err != nil && !errors.Is(errReservationNotFound, err) {
			s.log.Error("unreserve failed", "lease", lid, "err", err)
		}
	}
}

func findDeployments(
	ctx context.Context,
	log *logrus.Logger,
	client Client,
) ([]ctypes.IDeployment, error) {
	deployments, err := client.Deployments(ctx)
	if err != nil {
		log.Error("fetching deployments", "err", err)
		return nil, err
	}

	return deployments, nil
}
