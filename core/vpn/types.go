package vpn

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// PacketInfo contains extracted source and destination IPs and ports
type PacketInfo struct {
	SrcIP    net.IP
	DstIP    net.IP
	SrcPort  *int // pointer to allow nil
	DstPort  *int
	Protocol uint8
}

// QueuedPacket represents a packet in the sending queue
type QueuedPacket struct {
	Ctx    context.Context
	DestIP string
	Data   []byte
	DoneCh chan error // Channel to signal when packet processing is complete
}

// PacketWorker handles packets for a specific destination IP:Port
type PacketWorker struct {
	// The unique key for this worker (IP:Port or just IP)
	SyncKey string
	// The destination IP
	DestIP string
	// The peer ID associated with this destination
	PeerID peer.ID
	// The stream to the peer
	Stream VPNStream
	// Channel for receiving packets
	PacketChan chan *QueuedPacket
	// Last activity time
	LastActivity time.Time
	// Mutex for protecting access to the worker
	Mu sync.Mutex
	// Context for the worker
	Ctx context.Context
	// Cancel function for the worker context
	Cancel context.CancelFunc
	// Whether the worker is running
	Running bool
	// PacketCount tracks the number of packets processed by this worker
	PacketCount int64
	// ErrorCount tracks the number of errors encountered by this worker
	ErrorCount int64
}
