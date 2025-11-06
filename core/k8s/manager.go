package k8s

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/boz/go-lifecycle"
	"github.com/sirupsen/logrus"

	"github.com/unicornultrafoundation/subnet-node/core/k8s/util"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/pubsub"
	mani "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/manifest/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"

	ctypes "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1"
	etypes "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/expiry"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/event"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/manifest"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/session"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/tools/fromctx"
)

// ScalingState represents the current scaling state of a deployment
type ScalingState int

const (
	// ScalingStateActive indicates the deployment is running normally
	ScalingStateActive ScalingState = iota
	// ScalingStateScalingDown indicates the deployment is being scaled down
	ScalingStateScalingDown
	// ScalingStateScaledDown indicates the deployment has been scaled to zero
	ScalingStateScaledDown
	// ScalingStateScalingUp indicates the deployment is being scaled back up
	ScalingStateScalingUp
)

func (s ScalingState) String() string {
	switch s {
	case ScalingStateActive:
		return "active"
	case ScalingStateScalingDown:
		return "scaling-down"
	case ScalingStateScaledDown:
		return "scaled-down"
	case ScalingStateScalingUp:
		return "scaling-up"
	default:
		return "unknown"
	}
}

var (
	ErrLeaseInactive = errors.New("inactive Lease")
)

const uncleanShutdownGracePeriod = 30 * time.Second

type deploymentState string

const (
	dsDeployActive     deploymentState = "deploy-active"
	dsDeployPending    deploymentState = "deploy-pending"
	dsDeployComplete   deploymentState = "deploy-complete"
	dsTeardownActive   deploymentState = "teardown-active"
	dsTeardownPending  deploymentState = "teardown-pending"
	dsTeardownComplete deploymentState = "teardown-complete"
)

type deploymentManager struct {
	bus                 pubsub.Bus
	client              Client
	session             session.Session
	state               deploymentState
	deployment          ctypes.IDeployment
	monitor             *deploymentMonitor
	wg                  sync.WaitGroup
	updatech            chan ctypes.IDeployment
	teardownch          chan struct{}
	currentHostnames    map[string]struct{}
	log                 *logrus.Logger
	lc                  lifecycle.Lifecycle
	hostnameService     ctypes.HostnameServiceClient
	config              Config
	isNewLease          bool
	serviceShuttingDown <-chan struct{}
	messages            []string
	expiryService       *expiryService

	// Scaling state management
	scalingState     ScalingState
	originalReplicas map[string]int32
	scalingMutex     sync.RWMutex
}

func newDeploymentManager(s *service, deployment ctypes.IDeployment, isNewLease bool, expiryService *expiryService) *deploymentManager {
	lid := deployment.LeaseID()
	mgroup := deployment.ManifestGroup()
	logger := s.log.WithField("module", "deployment-manager").WithField("lease", lid).WithField("manifest-group", mgroup.GetName()).Logger

	dm := &deploymentManager{
		bus:                 s.bus,
		client:              s.client,
		session:             s.session,
		state:               dsDeployActive,
		deployment:          deployment,
		wg:                  sync.WaitGroup{},
		updatech:            make(chan ctypes.IDeployment),
		teardownch:          make(chan struct{}),
		log:                 logger,
		lc:                  lifecycle.New(),
		hostnameService:     s.HostnameService(),
		config:              s.config,
		serviceShuttingDown: s.lc.ShuttingDown(),
		isNewLease:          isNewLease,
		currentHostnames:    make(map[string]struct{}),
		expiryService:       expiryService,
		scalingState:        ScalingStateActive,
		originalReplicas:    make(map[string]int32),
		scalingMutex:        sync.RWMutex{},
	}

	go dm.lc.WatchChannel(s.lc.ShuttingDown())
	go dm.run(context.Background())

	go func() {
		<-dm.lc.Done()
		dm.log.Debug("sending manager into channel")
		s.managerch <- dm
	}()

	err := s.bus.Publish(event.LeaseAddFundsMonitor{LeaseID: lid, IsNewLease: isNewLease})
	if err != nil {
		s.log.WithError(err).WithField("lease", lid).Error("Unable to publish LeaseAddFundsMonitor event")
	}

	return dm
}

func (dm *deploymentManager) update(deployment ctypes.IDeployment) error {
	select {
	case dm.updatech <- deployment:
		return nil
	case <-dm.lc.ShuttingDown():
		return ErrNotRunning
	}
}

func (dm *deploymentManager) teardown() error {
	select {
	case dm.teardownch <- struct{}{}:
		return nil
	case <-dm.lc.ShuttingDown():
		return ErrNotRunning
	}
}

func (dm *deploymentManager) handleUpdate(ctx context.Context) <-chan error {
	switch dm.state {
	case dsDeployActive:
		dm.state = dsDeployPending
	case dsDeployComplete:
		// start update
		return dm.startDeploy(ctx)
	case dsDeployPending, dsTeardownActive, dsTeardownPending, dsTeardownComplete:
		// do nothing
	}

	return nil
}

func (dm *deploymentManager) run(ctx context.Context) {
	defer dm.lc.ShutdownCompleted()
	var shutdownErr error

	runch := dm.startDeploy(ctx)

	defer func() {
		err := dm.hostnameService.ReleaseHostnames(dm.deployment.LeaseID())
		if err != nil {
			dm.log.WithError(err).Error("failed releasing hostnames")
		}
		dm.log.Debug("hostnames released")
	}()

	var teardownErr error

loop:
	for {
		select {
		case shutdownErr = <-dm.lc.ShutdownRequest():
			dm.log.WithError(shutdownErr).Debug("received shutdown request")
			break loop
		case deployment := <-dm.updatech:
			dm.deployment = deployment
			newch := dm.handleUpdate(ctx)
			if newch != nil {
				runch = newch
			}

		case result := <-runch:
			runch = nil
			if result != nil {
				dm.log.WithField("state", dm.state).WithError(result).Error("Execution error")
			}
			switch dm.state {
			case dsDeployActive:
				// regardless of error from deploy, mark as complete
				// save the last error if any for user to retrieve status of the deployment
				dm.log.Debug("deploy complete")
				dm.state = dsDeployComplete
				dm.startMonitor()

				if result != nil {
					dm.messages = []string{result.Error()}
				} else {
					dm.messages = nil
				}
			case dsDeployPending:
				if result != nil {
					break loop
				}
				// start update
				runch = dm.startDeploy(ctx)
			case dsDeployComplete:
				panic(fmt.Sprintf("INVALID STATE: runch read on %v", dm.state))
			case dsTeardownActive:
				teardownErr = result
				dm.state = dsTeardownComplete
				dm.log.Debug("teardown complete")
				break loop
			case dsTeardownPending:
				// start teardown
				runch = dm.startTeardown()
			case dsTeardownComplete:
				panic(fmt.Sprintf("INVALID STATE: runch read on %v", dm.state))
			}

		case <-dm.teardownch:
			dm.log.Debug("Teardown request")
			dm.stopMonitor()
			switch dm.state {
			case dsDeployActive:
				dm.state = dsTeardownPending
			case dsDeployPending:
				dm.state = dsTeardownPending
			case dsDeployComplete:
				// start teardown
				runch = dm.startTeardown()
			case dsTeardownActive, dsTeardownPending, dsTeardownComplete:
			}
		}
	}

	dm.log.Debug("shutting down")
	dm.lc.ShutdownInitiated(shutdownErr)
	if runch != nil {
		<-runch
		dm.log.Debug("read from runch during shutdown")
	}

	dm.log.Debug("waiting on dm.wg")
	dm.wg.Wait()

	if dm.isNewLease && (dm.state < dsDeployComplete) {
		dm.log.Info("shutting down unclean, running teardown now")
		ctx, cancel := context.WithTimeout(context.Background(), uncleanShutdownGracePeriod)
		defer cancel()
		teardownErr = dm.doTeardown(ctx)
	}

	if teardownErr != nil {
		dm.log.WithError(teardownErr).Error("lease teardown failed")
	}

	dm.log.WithField("lease", dm.deployment.LeaseID()).Info("Shutdown complete")
}

func (dm *deploymentManager) startMonitor() {
	dm.wg.Add(1)
	dm.monitor = newDeploymentMonitor(dm)
	go func(m *deploymentMonitor) {
		defer dm.wg.Done()
		<-m.done()
	}(dm.monitor)
}

func (dm *deploymentManager) stopMonitor() {
	if dm.monitor != nil {
		dm.monitor.shutdown()
	}
}

func (dm *deploymentManager) startDeploy(ctx context.Context) <-chan error {
	dm.stopMonitor()
	dm.state = dsDeployActive

	chErr := make(chan error, 1)

	go func() {
		hostnames, endpoints, err := dm.doDeploy(ctx)
		if err != nil {
			chErr <- err
			return
		}

		if len(hostnames) != 0 {
			// Some hostnames have been withheld
			dm.log.WithField("cnt", len(hostnames)).WithField("lease", dm.deployment.LeaseID()).Warn("hostnames withheld from deployment")
		}

		if len(endpoints) != 0 {
			// Some endpoints have been withheld
			dm.log.WithField("cnt", len(endpoints)).WithField("lease", dm.deployment.LeaseID()).Warn("endpoints withheld from deployment")
		}

		groupCopy := *dm.deployment.ManifestGroup()
		ev := event.ClusterDeployment{
			LeaseID: dm.deployment.LeaseID(),
			Group:   &groupCopy,
			Status:  event.ClusterDeploymentUpdated,
		}
		err = dm.bus.Publish(ev)
		if err != nil {
			dm.log.WithError(err).Error("Failed publishing event")
		}

		close(chErr)
	}()

	return chErr
}

func (dm *deploymentManager) startTeardown() <-chan error {
	dm.stopMonitor()
	dm.state = dsTeardownActive
	return dm.do(func() error {
		// Don't use a context tied to the lifecycle, as we don't want to cancel Kubernetes operations
		return dm.doTeardown(context.Background())
	})
}

type serviceExposeWithServiceName struct {
	expose mani.ServiceExpose
	name   string
}

func (dm *deploymentManager) doDeploy(ctx context.Context) ([]string, []string, error) {
	cleanupHelper := newDeployCleanupHelper(dm.deployment.LeaseID(), dm.client, dm.log)
	var err error
	ctx, cancel := context.WithCancel(context.Background())

	// Weird hack to tie this context to the lifecycle of the parent service, so this doesn't
	// block forever or anything like that
	go func() {
		select {
		case <-dm.serviceShuttingDown:
			cancel()
		case <-ctx.Done():
		}
	}()

	defer func() {
		// TODO - run on an isolated context
		cleanupHelper.purgeAll(ctx)
		cancel()
	}()

	if err = dm.checkLeaseActive(ctx); err != nil {
		return nil, nil, err
	}

	// Either reserve the hostnames, or confirm that they already are held
	allHostnames := manifest.AllHostnamesOfManifestGroup(*dm.deployment.ManifestGroup())
	withheldHostnames, err := dm.hostnameService.ReserveHostnames(ctx, allHostnames, dm.deployment.LeaseID())

	if err != nil {
		dm.log.Error("deploy hostname reservation error", "state", dm.state, "err", err)
		return nil, nil, err
	}

	dm.log.Info("hostnames withheld", "cnt", len(withheldHostnames))

	hostnamesInThisRequest := make(map[string]struct{})
	for _, hostname := range allHostnames {
		hostnamesInThisRequest[hostname] = struct{}{}
	}

	// Figure out what hostnames were removed from the manifest if any
	for hostnameInUse := range dm.currentHostnames {
		_, stillInUse := hostnamesInThisRequest[hostnameInUse]
		if !stillInUse {
			cleanupHelper.addHostname(hostnameInUse)
		}
	}

	// Don't use a context tied to the lifecycle, as we don't want to cancel Kubernetes operations
	deployCtx := fromctx.ApplyToContext(context.Background(), dm.config.ClusterSettings)

	err = dm.client.Deploy(deployCtx, dm.deployment)
	if err != nil {
		dm.log.WithError(err).Error("deploying workload")
		return nil, nil, err
	}

	// Figure out what hostnames to declare
	blockedHostnames := make(map[string]struct{})
	for _, hostname := range withheldHostnames {
		blockedHostnames[hostname] = struct{}{}
	}
	hosts := make(map[string]mani.ServiceExpose)
	leasedIPs := make([]serviceExposeWithServiceName, 0)
	hostToServiceName := make(map[string]string)

	ipsInThisRequest := make(map[string]serviceExposeWithServiceName)
	// clear this out so it gets repopulated
	dm.currentHostnames = make(map[string]struct{})
	// Iterate over each entry, extracting the ingress services & leased IPs
	for _, service := range dm.deployment.ManifestGroup().Services {
		for _, expose := range service.Expose {
			if expose.IsIngress() {
				if dm.config.DeploymentIngressStaticHosts {
					uid := manifest.IngressHost(dm.deployment.LeaseID(), service.Name)
					host := fmt.Sprintf("%s.%s", uid, dm.config.DeploymentIngressDomain)
					hosts[host] = expose
					hostToServiceName[host] = service.Name
				}

				for _, host := range expose.Hosts {
					_, blocked := blockedHostnames[host]
					if !blocked {
						dm.currentHostnames[host] = struct{}{}
						hosts[host] = expose
						hostToServiceName[host] = service.Name
					}
				}
			}

			if expose.Global && len(expose.IP) != 0 {
				v := serviceExposeWithServiceName{expose: expose, name: service.Name}
				leasedIPs = append(leasedIPs, v)
				sharingKey := util.MakeIPSharingKey(dm.deployment.LeaseID(), expose.IP)
				ipsInThisRequest[sharingKey] = v
			}
		}
	}

	for host, serviceExpose := range hosts {
		externalPort := uint32(serviceExpose.GetExternalPort()) // nolint: gosec
		err = dm.client.DeclareHostname(ctx, dm.deployment.LeaseID(), host, hostToServiceName[host], externalPort)
		if err != nil {
			// TODO - counter
			return withheldHostnames, nil, err
		}
	}

	return withheldHostnames, nil, nil
}

func (dm *deploymentManager) getCleanupRetryOpts(ctx context.Context) []retry.Option {
	retryFn := func(err error) bool {
		isCanceled := errors.Is(err, context.Canceled)
		isDeadlineExceeded := errors.Is(err, context.DeadlineExceeded)
		return !isCanceled && !isDeadlineExceeded
	}
	return []retry.Option{
		retry.Attempts(50),
		retry.Delay(100 * time.Millisecond),
		retry.MaxDelay(3000 * time.Millisecond),
		retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(true),
		retry.RetryIf(retryFn),
		retry.Context(ctx),
	}
}

func (dm *deploymentManager) doTeardown(ctx context.Context) error {
	const teardownActivityCount = 1
	teardownResults := make(chan error, teardownActivityCount)

	go func() {
		result := retry.Do(func() error {
			err := dm.client.TeardownLease(ctx, dm.deployment.LeaseID())
			if err != nil {
				dm.log.WithError(err).Error("lease teardown failed")
			}
			return err
		}, dm.getCleanupRetryOpts(ctx)...)

		teardownResults <- result
	}()

	go func() {
		result := retry.Do(func() error {
			err := dm.client.PurgeDeclaredHostnames(ctx, dm.deployment.LeaseID())
			if err != nil {
				dm.log.Error("purge declared hostname failure", "err", err)
			}
			return err
		}, dm.getCleanupRetryOpts(ctx)...)
		// TODO - counter

		if result == nil {
			dm.log.Debug("purged hostnames")
		}
		teardownResults <- result
	}()

	var firstError error
	for i := 0; i != teardownActivityCount; i++ {
		select {
		case err := <-teardownResults:
			if err != nil && firstError == nil {
				firstError = err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return firstError
}

func (dm *deploymentManager) checkLeaseActive(ctx context.Context) error {
	lease := mtypes.QueryLeaseResponse{
		Lease: mtypes.Lease{
			State: mtypes.LeaseActive,
		},
	}

	// Check if the lease is active on-chain
	deploymentID := fmt.Sprintf("%d", dm.deployment.LeaseID().DSeq)
	status, err := dm.expiryService.CheckDeploymentExpiry(ctx, deploymentID)
	if err != nil {
		dm.log.WithError(err).Error("error checking deployment expiry")
		return err
	}
	if status.Status != etypes.DeploymentExpiryStatusActive {
		lease.Lease.State = mtypes.LeaseClosed
	}

	if status.Status == etypes.DeploymentExpiryStatusDeleted {
		// If the lease is expired and deleted, we should close the lease
		dm.log.WithField("lease", dm.deployment.LeaseID()).Info("Deployment deleted. Closing lease...")
		if err := dm.bus.Publish(mtypes.EventLeaseClosed{
			ID: dm.deployment.LeaseID(),
		}); err != nil {
			dm.log.WithError(err).Error("Send lease closed request failed")
		}
	}

	if lease.GetLease().State != mtypes.LeaseActive {
		dm.log.Error("lease not active, not deploying")
		return fmt.Errorf("%w: %s", ErrLeaseInactive, dm.deployment.LeaseID())
	}

	return nil
}

func (dm *deploymentManager) do(fn func() error) <-chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- fn()
	}()
	return ch
}

func TieContextToChannel(parentCtx context.Context, donech <-chan struct{}) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parentCtx)

	go func() {
		select {
		case <-donech:
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx, cancel
}

// setScalingState transitions the deployment to a new scaling state
func (dm *deploymentManager) setScalingState(newState ScalingState) error {
	dm.scalingMutex.Lock()
	defer dm.scalingMutex.Unlock()

	oldState := dm.scalingState

	// Validate state transition
	if !dm.isValidStateTransition(oldState, newState) {
		return fmt.Errorf("invalid state transition from %s to %s", oldState, newState)
	}

	// Store original replica counts when starting to scale down
	if newState == ScalingStateScalingDown {
		dm.originalReplicas = make(map[string]int32)
		for _, service := range dm.deployment.ManifestGroup().Services {
			dm.originalReplicas[service.Name] = int32(service.Count)
		}
		dm.log.WithField("originalReplicas", dm.originalReplicas).Debug("Stored original replica counts")
	}

	dm.scalingState = newState
	dm.log.WithField("oldState", oldState).WithField("newState", newState).Debug("Updated scaling state")
	return nil
}

// isValidStateTransition validates if a state transition is allowed
func (dm *deploymentManager) isValidStateTransition(from, to ScalingState) bool {
	switch from {
	case ScalingStateActive:
		return to == ScalingStateScalingDown
	case ScalingStateScalingDown:
		return to == ScalingStateScaledDown || to == ScalingStateActive // Allow cancellation
	case ScalingStateScaledDown:
		return to == ScalingStateScalingUp
	case ScalingStateScalingUp:
		return to == ScalingStateActive || to == ScalingStateScaledDown // Allow cancellation
	default:
		return false
	}
}

// getScalingState returns the current scaling state
func (dm *deploymentManager) getScalingState() ScalingState {
	dm.scalingMutex.RLock()
	defer dm.scalingMutex.RUnlock()
	return dm.scalingState
}

// getOriginalReplicas returns the original replica counts for each service
func (dm *deploymentManager) getOriginalReplicas() map[string]int32 {
	dm.scalingMutex.RLock()
	defer dm.scalingMutex.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]int32)
	for service, replicas := range dm.originalReplicas {
		result[service] = replicas
	}
	return result
}

// isScaledDown returns whether the deployment is currently scaled down
func (dm *deploymentManager) isScaledDown() bool {
	dm.scalingMutex.RLock()
	defer dm.scalingMutex.RUnlock()
	return dm.scalingState == ScalingStateScaledDown
}

// canScaleDown returns whether the deployment can be scaled down
func (dm *deploymentManager) canScaleDown() bool {
	dm.scalingMutex.RLock()
	defer dm.scalingMutex.RUnlock()
	return dm.scalingState == ScalingStateActive || dm.scalingState == ScalingStateScalingUp
}

// canScaleUp returns whether the deployment can be scaled up
func (dm *deploymentManager) canScaleUp() bool {
	dm.scalingMutex.RLock()
	defer dm.scalingMutex.RUnlock()
	return dm.scalingState == ScalingStateScaledDown || dm.scalingState == ScalingStateScalingDown
}
