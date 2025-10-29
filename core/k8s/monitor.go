package k8s

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/boz/go-lifecycle"

	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/pubsub"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/runner"

	ctypes "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1"
	etypes "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/expiry"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/event"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/session"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/tools/fromctx"
)

type deploymentMonitor struct {
	bus     pubsub.Bus
	session session.Session
	client  Client

	deployment ctypes.IDeployment

	attempts uint
	log      *logrus.Logger
	lc       lifecycle.Lifecycle

	expiryService *expiryService

	config Config

	// Reference to the deployment manager for accessing scaling state
	manager *deploymentManager
}

func newDeploymentMonitor(dm *deploymentManager) *deploymentMonitor {
	m := &deploymentMonitor{
		bus:           dm.bus,
		session:       dm.session,
		client:        dm.client,
		deployment:    dm.deployment,
		log:           dm.log.WithField("module", "deployment-monitor").Logger,
		lc:            lifecycle.New(),
		config:        dm.config,
		expiryService: dm.expiryService,
		manager:       dm,
	}

	go m.lc.WatchChannel(dm.lc.ShuttingDown())
	go m.run()

	return m
}

func (m *deploymentMonitor) shutdown() {
	m.lc.ShutdownAsync(nil)
}

func (m *deploymentMonitor) done() <-chan struct{} {
	return m.lc.Done()
}

func (m *deploymentMonitor) run() {
	defer m.lc.ShutdownCompleted()
	ctx, cancel := context.WithCancel(context.Background())

	var (
		runch       <-chan runner.Result
		runExpirych <-chan runner.Result
		closech     <-chan runner.Result
	)

	tickch := m.scheduleRetry()
	prevStatus := event.ClusterDeploymentUnknown

	expiryTickch := m.scheduleExpiryCheck()

loop:
	for {
		select {
		case err := <-m.lc.ShutdownRequest():
			m.lc.ShutdownInitiated(err)
			break loop

		case <-tickch:
			tickch = nil
			runch = m.runCheck(ctx)

		case <-expiryTickch:
			expiryTickch = nil
			runExpirych = m.runExpiryCheck(ctx)

		case result := <-runExpirych:
			runExpirych = nil
			if err := result.Error(); err != nil {
				m.log.WithError(err).Error("Expiry check")
			}
			currExpiryStatus := result.Value().(etypes.DeploymentExpiryInfo)

			m.publishExpiryStatus(currExpiryStatus.Status)

			expiryTickch = m.scheduleExpiryCheck()

		case result := <-runch:
			runch = nil

			if err := result.Error(); err != nil {
				m.log.WithError(err).Error("Monitor check")
			}

			var currStatus event.ClusterDeploymentStatus

			healthy := result.Value().(bool)

			if healthy {
				currStatus = event.ClusterDeploymentDeployed

				m.attempts = 0
				tickch = m.scheduleHealthcheck()

			} else {
				currStatus = event.ClusterDeploymentPending

				m.log.WithField("ok", false).WithField("attempt", m.attempts).Info("Check result")
			}

			if currStatus != prevStatus {
				m.publishStatus(currStatus)
				prevStatus = currStatus
			}

			if !healthy {
				if m.attempts <= m.config.MonitorMaxRetries {
					// unhealthy.  retry
					tickch = m.scheduleRetry()
					break
				}

				m.log.Error("Deployment failed. Closing lease.")
				closech = m.runCloseLease(ctx)
			}
		case <-closech:
			closech = nil
		}
	}
	cancel()

	if runch != nil {
		<-runch
	}

	if closech != nil {
		<-closech
	}
}

func (m *deploymentMonitor) runCheck(ctx context.Context) <-chan runner.Result {
	m.attempts++
	return runner.Do(func() runner.Result {
		return runner.NewResult(m.doCheck(ctx))
	})
}

func (m *deploymentMonitor) doCheck(ctx context.Context) (bool, error) {
	ctx = fromctx.ApplyToContext(ctx, m.config.ClusterSettings)

	status, err := m.client.LeaseStatus(ctx, m.deployment.LeaseID())

	if err != nil {
		m.log.WithError(err).Error("Lease status")
		return false, err
	}

	// Check if deployment is scaled down - if so, consider it healthy
	if m.manager != nil && m.manager.isScaledDown() {
		m.log.Debug("Deployment is scaled down, considering it healthy")
		return true, nil
	}

	badsvc := 0

	for _, spec := range m.deployment.ManifestGroup().Services {
		service, foundService := status[spec.Name]
		if foundService {
			if uint32(service.Available) < spec.Count { // nolint: gosec
				badsvc++
				m.log.WithField("service", spec.Name).WithField("available", service.Available).WithField("target", spec.Count).Debug("Service available replicas below target")
			}
		}

		if !foundService {
			badsvc++
			m.log.WithField("service", spec.Name).Debug("Service status not found")
		}
	}

	return badsvc == 0, nil
}

func (m *deploymentMonitor) runCloseLease(ctx context.Context) <-chan runner.Result {
	return runner.Do(func() runner.Result {
		// Still keep the lease open for debugging purposes
		res := true
		return runner.NewResult(res, nil)
	})
}

func (m *deploymentMonitor) runExpiryCheck(ctx context.Context) <-chan runner.Result {
	return runner.Do(func() runner.Result {
		return runner.NewResult(m.doExpiryCheck(ctx))
	})
}

func (m *deploymentMonitor) doExpiryCheck(ctx context.Context) (etypes.DeploymentExpiryInfo, error) {
	deploymentID := fmt.Sprintf("%d", m.deployment.LeaseID().DSeq)
	return m.expiryService.CheckDeploymentExpiry(ctx, deploymentID)
}

func (m *deploymentMonitor) publishExpiryStatus(status etypes.DeploymentExpiryStatus) {
	if err := m.bus.Publish(etypes.DeploymentExpiry{
		LeaseID: m.deployment.LeaseID(),
		Status:  status,
	}); err != nil {
		m.log.WithError(err).WithField("status", status).Error("Publish expiry status event failed")
	}
}

func (m *deploymentMonitor) publishStatus(status event.ClusterDeploymentStatus) {
	if err := m.bus.Publish(event.ClusterDeployment{
		LeaseID: m.deployment.LeaseID(),
		Group:   m.deployment.ManifestGroup(),
		Status:  status,
	}); err != nil {
		m.log.WithError(err).WithField("status", status).Error("Publishing manifest group deployed event")
	}
}

func (m *deploymentMonitor) scheduleRetry() <-chan time.Time {
	return m.schedule(m.config.MonitorRetryPeriod, m.config.MonitorRetryPeriodJitter)
}

func (m *deploymentMonitor) scheduleHealthcheck() <-chan time.Time {
	return m.schedule(m.config.MonitorHealthcheckPeriod, m.config.MonitorHealthcheckPeriodJitter)
}

func (m *deploymentMonitor) scheduleExpiryCheck() <-chan time.Time {
	return m.schedule(m.config.MonitorExpiryCheckPeriod, m.config.MonitorExpiryCheckPeriodJitter)
}

func (m *deploymentMonitor) schedule(minTime, jitter time.Duration) <-chan time.Time {
	period := minTime + time.Duration(rand.Int63n(int64(jitter))) // nolint: gosec
	return time.After(period)
}
