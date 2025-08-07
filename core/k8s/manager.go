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

	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/pubsub"
	mani "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/manifest/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"

	ctypes "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/event"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/session"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/tools/fromctx"
)

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
	config              Config
	isNewLease          bool
	serviceShuttingDown <-chan struct{}
	messages            []string
}

func newDeploymentManager(s *service, deployment ctypes.IDeployment, isNewLease bool) *deploymentManager {
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
		config:              s.config,
		serviceShuttingDown: s.lc.ShuttingDown(),
		isNewLease:          isNewLease,
		currentHostnames:    make(map[string]struct{}),
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
		s.log.Error("unable to publish LeaseAddFundsMonitor event", "error", err, "lease", lid)
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

	dm.log.Info("Shutdown complete")
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
		err := dm.doDeploy(ctx)
		if err != nil {
			chErr <- err
			return
		}

		groupCopy := *dm.deployment.ManifestGroup()
		ev := event.ClusterDeployment{
			LeaseID: dm.deployment.LeaseID(),
			Group:   &groupCopy,
			Status:  event.ClusterDeploymentUpdated,
		}
		err = dm.bus.Publish(ev)
		if err != nil {
			dm.log.Error("failed publishing event", "err", err)
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

func (dm *deploymentManager) doDeploy(ctx context.Context) error {
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
		cancel()
	}()

	if err = dm.checkLeaseActive(ctx); err != nil {
		return err
	}

	// Don't use a context tied to the lifecycle, as we don't want to cancel Kubernetes operations
	deployCtx := fromctx.ApplyToContext(context.Background(), dm.config.ClusterSettings)

	err = dm.client.Deploy(deployCtx, dm.deployment)
	if err != nil {
		dm.log.WithError(err).Error("deploying workload")
		return err
	}

	return nil
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
	// TODO - remove this once we have a way to check if the lease is active
	lease := mtypes.QueryLeaseResponse{
		Lease: mtypes.Lease{
			State: mtypes.LeaseActive,
		},
	}

	// TODO: Check if the lease is active on-chain

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
