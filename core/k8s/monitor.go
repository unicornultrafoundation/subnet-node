package k8s

import (
	"context"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/boz/go-lifecycle"

	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/pubsub"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/runner"

	ctypes "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1"
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

	config Config
}

func newDeploymentMonitor(dm *deploymentManager) *deploymentMonitor {
	m := &deploymentMonitor{
		bus:        dm.bus,
		session:    dm.session,
		client:     dm.client,
		deployment: dm.deployment,
		log:        dm.log.WithField("module", "deployment-monitor").Logger,
		lc:         lifecycle.New(),
		config:     dm.config,
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
		runch   <-chan runner.Result
		closech <-chan runner.Result
	)

	tickch := m.scheduleRetry()

	prevStatus := event.ClusterDeploymentUnknown

loop:
	for {
		select {
		case err := <-m.lc.ShutdownRequest():
			m.lc.ShutdownInitiated(err)
			break loop

		case <-tickch:
			tickch = nil
			runch = m.runCheck(ctx)

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

			// TODO: Add case check lease on-chain status
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
		// TODO: retry, timeout
		// msg := &mtypes.MsgCloseBid{
		// 	BidID: m.deployment.LeaseID().BidID(),
		// }
		// res, err := m.session.Client().Tx().Broadcast(ctx, []sdk.Msg{msg}, aclient.WithResultCodeAsError())
		// if err != nil {
		// 	m.log.Error("closing deployment", "err", err)
		// } else {
		// 	m.log.Info("bidding on lease closed")
		// }

		// Still keep the lease open for debugging purposes
		res := true
		return runner.NewResult(res, nil)
	})
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

func (m *deploymentMonitor) schedule(minTime, jitter time.Duration) <-chan time.Time {
	period := minTime + time.Duration(rand.Int63n(int64(jitter))) // nolint: gosec
	return time.After(period)
}
