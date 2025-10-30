package ipmanager

import (
	"context"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/client/dynamic"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/client/static"
	"go.uber.org/fx"
)

var log = logrus.WithField("service", "vpn-ip-manager")

type IPManager interface {
	WatchIP() (IPManagerObserver, error)
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type IPManagerImpl struct {
	ipFromConfig  string
	peerID        string
	cfgIpCh       chan string
	ipCh          chan string
	dynamicClient dynamic.DynamicIPClient
	staticClient  static.StaticIPClient

	state        IPManagerState
	stateCh      chan IPManagerStateType
	observerList []IPManagerObserver
	mu           sync.RWMutex
	stopped      bool

	// Manager context
	ctx      context.Context
	cancel   context.CancelFunc
	stopChan chan struct{}

	// dispatcher lifecycle
	dispatcherOnce sync.Once
}

func NewIPManager(lc fx.Lifecycle, cfg *config.C, host host.Host, dynamicClient dynamic.DynamicIPClient, staticClient static.StaticIPClient) IPManager {
	ipCh := make(chan string)
	cfgIpCh := make(chan string)

	// Register callback to update the IP from the config
	cfg.RegisterReloadCallback(func(c *config.C) {
		if c.HasChanged("vpn.virtual_ip") {
			newIp := c.GetString("vpn.virtual_ip", "")
			log.WithField("ip", newIp).Info("VPN virtual IP changed, updating IP manager")
			cfgIpCh <- newIp
		}
	})

	i := &IPManagerImpl{
		ipCh:          ipCh,
		cfgIpCh:       cfgIpCh,
		ipFromConfig:  cfg.GetString("vpn.virtual_ip", ""),
		peerID:        host.ID().String(),
		dynamicClient: dynamicClient,
		staticClient:  staticClient,
		stopChan:      make(chan struct{}),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return i.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return i.Stop(ctx)
		},
	})

	return i
}

func (i *IPManagerImpl) Start(ctx context.Context) error {
	go func() {
		i.ctx, i.cancel = context.WithCancel(context.Background())
		// Initialize the IP from the config
		if err := i.start(i.ctx, i.ipFromConfig); err != nil {
			log.WithField("ip", i.ipFromConfig).Errorf("failed to start IP manager: %v", err)
			return
		}

		// Start the IP manager loop
		for {
			select {
			case ip := <-i.cfgIpCh:
				// If the IP is the same as the one from the config, skip
				if ip == i.ipFromConfig {
					continue
				}

				// Start the new IP
				log.Infof("IP changed to %s", ip)
				if err := i.start(i.ctx, ip); err != nil {
					log.WithField("ip", ip).Errorf("failed to start IP manager: %v", err)
					return
				}

				// Update the IP from the config
				i.ipFromConfig = ip

			case <-i.ctx.Done():
				log.Info("IP manager context done")
				return
			case <-i.stopChan:
				log.Info("IP manager stop channel closed")
				return
			}
		}
	}()

	return nil
}

func (i *IPManagerImpl) Stop(ctx context.Context) error {
	if i.isStopped() {
		log.Warn("IP manager is already stopped")
		return nil
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	// Cancel the manager context
	if i.cancel != nil {
		i.cancel()
	}
	i.ctx = nil
	i.cancel = nil
	close(i.stopChan)
	i.stopped = true

	// Close all observers synchronously to prevent goroutine leaks
	for _, observer := range i.observerList {
		observer.Close()
	}
	i.observerList = nil

	return nil
}

func (i *IPManagerImpl) start(ctx context.Context, ip string) error {
	if i.isStopped() {
		log.Warn("IP manager is stopped")
		return fmt.Errorf("IP manager is stopped")
	}

	// Stop the previous state lifecycle
	if i.stateCh != nil {
		close(i.stateCh)
		i.stateCh = nil
		i.setState(nil)
	}

	i.stateCh = make(chan IPManagerStateType)
	// ensure dispatcher is running
	i.startDispatcher()
	// Start the new state lifecycle
	go func() {
		// Initialize the states
		staticIPManager := NewStaticIPManager(i.staticClient, ip, i.peerID, i.stateCh, i.ipCh)
		dynamicIPManager := NewDynamicIPManager(i.dynamicClient, i.stateCh, i.ipCh)

		// Cleanup the states on exit
		defer func() {
			err := i.cleanupState(ctx)
			if err != nil {
				log.WithField("ip", ip).WithError(err).Errorf("failed to cleanup previous state")
			}

			staticIPManager = nil
			dynamicIPManager = nil
		}()

		// Start the new state lifecycle
		for {
			select {
			case state := <-i.stateCh:
				// Cleanup the previous state
				if err := i.cleanupState(ctx); err != nil {
					log.WithField("ip", ip).WithError(err).Errorf("failed to cleanup previous state")
					return
				}

				if i.isStopped() {
					return
				}

				// Set the new state
				switch state {
				case IPManagerStateTypeStatic:
					i.setState(staticIPManager)
				case IPManagerStateTypeDynamic:
					i.setState(dynamicIPManager)
				default:
					log.WithField("state", state).Errorf("invalid state")
					return
				}

				// Start the new state
				if err := i.getState().Start(ctx); err != nil {
					log.WithField("ip", ip).WithField("state", i.getState().GetType()).Errorf("failed to start state")
					return
				}
			case <-ctx.Done():
				return
			case <-i.stopChan:
				return
			}
		}
	}()

	// Send the initial state
	i.stateCh <- IPManagerStateTypeStatic

	return nil
}

func (i *IPManagerImpl) cleanupState(ctx context.Context) error {
	// Stop the current state
	if i.getState() != nil {
		if err := i.getState().Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop state: %w", err)
		}

		i.ipCh <- ""
	}

	return nil
}

func (i *IPManagerImpl) getState() IPManagerState {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.state
}

func (i *IPManagerImpl) setState(state IPManagerState) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.state = state
}

func (i *IPManagerImpl) isStopped() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.stopped
}

// startDispatcher launches a single goroutine that fans out IP updates to observers
func (i *IPManagerImpl) startDispatcher() {
	i.dispatcherOnce.Do(func() {
		go func() {
			for {
				select {
				case ip := <-i.ipCh:
					// broadcast to all observers non-blockingly
					i.mu.RLock()
					observers := append([]IPManagerObserver(nil), i.observerList...)
					i.mu.RUnlock()
					for _, obs := range observers {
						obs.SendIP(ip)
					}
				case <-i.stopChan:
					return
				case <-i.ctx.Done():
					return
				}
			}
		}()
	})
}

// removeObserver removes an observer by id; safe to call concurrently
func (i *IPManagerImpl) removeObserver(id string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	// filter slice without changing order
	dst := i.observerList[:0]
	for _, o := range i.observerList {
		// type assert to access id
		if ob, ok := o.(*observer); ok {
			if ob.id == id {
				continue
			}
		}
		dst = append(dst, o)
	}
	// allow GC of removed observers
	i.observerList = append([]IPManagerObserver(nil), dst...)
}
