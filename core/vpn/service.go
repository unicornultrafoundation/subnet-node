package vpn

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
)

const VPNProtocol = "/vpn/1.0.0"

var log = logrus.WithField("service", "vpn")

type Service struct {
	mu             sync.Mutex
	cfg            *config.C
	enable         bool
	IsProvider     bool
	accountService *account.AccountService
	PeerHost       p2phost.Host
	DHT            *ddht.DHT
	virtualIP      string
	subnet         int
	routes         []string
	mtu            int
	unallowedPorts map[string]bool
	streamCache    map[string]network.Stream
	peerIDCache    *cache.Cache

	// IP-specific mutexes for synchronous packet sending per destination IP
	ipMutexes sync.Map
	// Cache to track activity for IP cleanup
	ipActivityCache *cache.Cache
	// Mutex to protect access to ipActivityCache
	ipActivityMu sync.Mutex
	// Map of packet queues for each destination IP:Port
	packetQueues sync.Map
	// Mutex to protect access to packetQueues
	packetQueuesMu sync.Mutex

	stopChan chan struct{} // Channel to stop background tasks
}

// Initializes the Service
func New(cfg *config.C, peerHost p2phost.Host, dht *ddht.DHT, accountService *account.AccountService) *Service {
	unallowedPortList := cfg.GetStringSlice("vpn.unallowed_ports", []string{})
	unallowedPorts := make(map[string]bool, len(unallowedPortList))
	for _, port := range unallowedPortList {
		unallowedPorts[port] = true
	}

	return &Service{
		cfg:             cfg,
		PeerHost:        peerHost,
		accountService:  accountService,
		enable:          cfg.GetBool("vpn.enable", false),
		mtu:             cfg.GetInt("vpn.mtu", 1400),
		virtualIP:       cfg.GetString("vpn.virtual_ip", ""),
		subnet:          cfg.GetInt("vpn.subnet", 8),
		routes:          cfg.GetStringSlice("vpn.routes", []string{"10.0.0.0/8"}),
		unallowedPorts:  unallowedPorts,
		IsProvider:      cfg.GetBool("provider.enable", false),
		DHT:             dht,
		streamCache:     make(map[string]network.Stream),
		peerIDCache:     cache.New(1*time.Minute, 30*time.Second),
		ipActivityCache: cache.New(10*time.Minute, 1*time.Minute),
		stopChan:        make(chan struct{}),
	}
}

func (s *Service) Start(ctx context.Context) error {
	if !s.enable {
		return nil
	}

	if s.virtualIP == "" {
		log.Error("virtual IP is not set")
		return nil
	}

	if len(s.routes) == 0 {
		log.Error("routes are not set")
		return nil
	}

	// Wait until there are some peers connected
	WaitUntilPeerConnected(ctx, s.PeerHost)
	time.Sleep(1 * time.Second)

	// Start a goroutine to periodically clean up unused IP mutexes
	go s.cleanupIPMutexes(ctx)

	go func() {
		err := s.start(ctx)
		if err != nil {
			log.Errorf("Something went wrong when running VPN: %v", err)
		}
	}()

	return nil
}

func (s *Service) start(ctx context.Context) error {
	// Wait to connect to peer to sync DHT
	err := s.SyncPeerIDToDHT(ctx)
	if err != nil {
		return fmt.Errorf("failed to sync peer id to DHT: %v", err)
	}

	iface, err := s.SetupTUN()
	if err != nil {
		return fmt.Errorf("failed to setup TUN interface: %v", err)
	}
	defer iface.Close()

	s.HandleP2PTraffic(iface)

	// Listen packets from TUN interface
	return s.listenFromTUN(ctx, iface)
}

// cleanupIPMutexes periodically checks for and removes unused IP:Port mutexes
func (s *Service) cleanupIPMutexes(_ context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Use a separate cache to track last activity time for each IP:Port
	s.ipActivityMu.Lock()
	if s.ipActivityCache == nil {
		s.ipActivityCache = cache.New(10*time.Minute, 1*time.Minute)
	}
	s.ipActivityMu.Unlock()
	activityCache := s.ipActivityCache

	// Update activity cache when packets are sent
	s.ipMutexes.Range(func(key, value interface{}) bool {
		syncKey := key.(string)
		activityCache.Set(syncKey, time.Now(), cache.DefaultExpiration)
		return true
	})

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			// Clean up IP:Port mutexes that have expired from the activity cache
			s.ipMutexes.Range(func(key, value interface{}) bool {
				syncKey := key.(string)
				_, found := activityCache.Get(syncKey)
				if !found {
					// IP:Port has expired from activity cache, remove the mutex
					s.ipMutexes.Delete(syncKey)
					log.Debugf("Cleaned up mutex for unused key: %s", syncKey)
				}
				return true
			})
		}
	}
}

func (s *Service) Stop() error {
	if !s.enable || s.virtualIP == "" || len(s.routes) == 0 {
		return nil
	}

	if len(s.streamCache) > 0 {
		// Close all cached streams
		var wg sync.WaitGroup
		for _, stream := range s.streamCache {
			wg.Add(1)
			go func(str io.Closer) {
				defer wg.Done()
				str.Close()
			}(stream)
		}
		wg.Wait()
	}

	// Clear IP:Port mutexes, packet queues, and activity cache
	s.ipMutexes = sync.Map{}
	s.ipActivityMu.Lock()
	s.ipActivityCache = nil
	s.ipActivityMu.Unlock()

	// Close and clear all packet queues
	s.packetQueuesMu.Lock()
	s.packetQueues.Range(func(key, value interface{}) bool {
		queue := value.(chan *QueuedPacket)
		close(queue)
		s.packetQueues.Delete(key)
		return true
	})
	s.packetQueuesMu.Unlock()

	// Close stopChan to stop all background tasks
	close(s.stopChan)

	log.Infoln("VPN service stopped successfully!")
	return nil
}
