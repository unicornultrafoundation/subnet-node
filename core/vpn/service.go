package vpn

import (
	"context"
	"fmt"
	"sync"
	"time"

	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	p2phost "github.com/libp2p/go-libp2p/core/host"
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
	config         *VPNConfig
	IsProvider     bool
	accountService *account.AccountService
	PeerHost       p2phost.Host
	DHT            *ddht.DHT
	peerIDCache    *cache.Cache

	// Packet dispatcher for managing workers and routing packets
	dispatcher *PacketDispatcher
	// Buffer pool for packet processing
	bufferPool sync.Pool
	// Metrics for monitoring
	metrics struct {
		packetsReceived int64
		packetsSent     int64
		packetsDropped  int64
		bytesReceived   int64
		bytesSent       int64
		streamErrors    int64
		mutex           sync.Mutex
	}

	stopChan chan struct{} // Channel to stop background tasks
}

// Initializes the Service
func New(cfg *config.C, peerHost p2phost.Host, dht *ddht.DHT, accountService *account.AccountService) *Service {
	// Create the VPN configuration
	vpnConfig := NewVPNConfig(cfg)

	// Initialize the buffer pool with MTU-sized buffers
	service := &Service{
		cfg:            cfg,
		config:         vpnConfig,
		PeerHost:       peerHost,
		accountService: accountService,
		IsProvider:     cfg.GetBool("provider.enable", false),
		DHT:            dht,
		peerIDCache:    cache.New(1*time.Minute, 30*time.Second),
		bufferPool: sync.Pool{
			New: func() interface{} {
				buffer := make([]byte, vpnConfig.MTU)
				return &buffer
			},
		},
		stopChan: make(chan struct{}),
	}

	// Create the packet dispatcher
	service.dispatcher = NewPacketDispatcher(service, 0) // The second parameter is ignored now

	return service
}

func (s *Service) Start(ctx context.Context) error {
	if !s.config.Enable {
		return nil
	}

	// Validate configuration
	if err := s.config.Validate(); err != nil {
		log.Errorf("Invalid VPN configuration: %v", err)
		return err
	}

	// Wait until there are some peers connected
	WaitUntilPeerConnected(ctx, s.PeerHost)
	time.Sleep(1 * time.Second)

	// Start the packet dispatcher
	s.dispatcher.Start()

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

func (s *Service) Stop() error {
	if !s.config.Enable || s.config.VirtualIP == "" || len(s.config.Routes) == 0 {
		return nil
	}

	// Stop the packet dispatcher
	s.dispatcher.Stop()

	// Close stopChan to stop all background tasks
	close(s.stopChan)

	log.Infoln("VPN service stopped successfully!")
	return nil
}

// GetMetrics returns the current VPN metrics
func (s *Service) GetMetrics() map[string]int64 {
	s.metrics.mutex.Lock()
	defer s.metrics.mutex.Unlock()

	// Count active workers
	activeWorkers := 0
	s.dispatcher.workers.Range(func(_, _ interface{}) bool {
		activeWorkers++
		return true
	})

	return map[string]int64{
		"packets_received": s.metrics.packetsReceived,
		"packets_sent":     s.metrics.packetsSent,
		"packets_dropped":  s.metrics.packetsDropped,
		"bytes_received":   s.metrics.bytesReceived,
		"bytes_sent":       s.metrics.bytesSent,
		"active_workers":   int64(activeWorkers),
		"stream_errors":    s.metrics.streamErrors,
	}
}

// IncrementPacketsSent updates the packets sent metric
func (s *Service) IncrementPacketsSent(bytes int) {
	s.metrics.mutex.Lock()
	defer s.metrics.mutex.Unlock()
	s.metrics.packetsSent++
	s.metrics.bytesSent += int64(bytes)
}

// IncrementPacketsDropped updates the packets dropped metric
func (s *Service) IncrementPacketsDropped() {
	s.metrics.mutex.Lock()
	defer s.metrics.mutex.Unlock()
	s.metrics.packetsDropped++
}

// IncrementStreamErrors updates the stream errors metric
func (s *Service) IncrementStreamErrors() {
	s.metrics.mutex.Lock()
	defer s.metrics.mutex.Unlock()
	s.metrics.streamErrors++
}

// GetWorkerBufferSize returns the worker buffer size
func (s *Service) GetWorkerBufferSize() int {
	return s.config.WorkerBufferSize
}

// GetWorkerIdleTimeout returns the worker idle timeout in seconds
func (s *Service) GetWorkerIdleTimeout() int {
	return s.config.WorkerIdleTimeout
}
