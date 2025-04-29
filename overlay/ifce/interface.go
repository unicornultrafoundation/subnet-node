package ifce

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"runtime"
	"sync"
	"sync/atomic"

	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/firewall"
	"github.com/unicornultrafoundation/subnet-node/overlay"
)

const mtu = 9001

type InterfaceConfig struct {
	Inside   overlay.Device
	Routines int
	P2phost  p2phost.Host
	Logger   *logrus.Logger
}

type Interface struct {
	inside     overlay.Device
	readers    []io.ReadWriteCloser
	routines   int
	closed     atomic.Bool
	l          *logrus.Logger
	p2phost    p2phost.Host
	streams    map[string]network.Stream
	streamLock sync.Mutex
}

// NewInterface creates a new VPN interface with the given configuration.
func NewInterface(ctx context.Context, c *InterfaceConfig) *Interface {
	ifce := &Interface{
		inside:     c.Inside,
		readers:    make([]io.ReadWriteCloser, c.Routines),
		routines:   c.Routines,
		p2phost:    c.P2phost,
		l:          c.Logger,
		streams:    make(map[string]network.Stream),
		streamLock: sync.Mutex{},
	}
	return ifce
}

func (f *Interface) activate() {

	var err error
	// Prepare n tun queues
	var reader io.ReadWriteCloser = f.inside
	for i := 0; i < f.routines; i++ {
		if i > 0 {
			reader, err = f.inside.NewMultiQueueReader()
			if err != nil {
				f.l.Fatal(err)
			}
		}
		f.readers[i] = reader
	}

	if err := f.inside.Activate(); err != nil {
		f.inside.Close()
		f.l.Fatal(err)
	}

	f.p2phost.SetStreamHandler("/vpn/1.0.0", f.listenOut)
}

func (f *Interface) run() {
	// Launch n queues to read packets from tun dev
	for i := 0; i < f.routines; i++ {
		go f.listenIn(f.readers[i], i)
	}
}

func (f *Interface) listenOut(s network.Stream) {
	defer s.Close()

	peer := s.Conn().RemotePeer().String()
	f.l.WithField("peer", peer).Debug("New VPN stream established")

	// Pre-allocate all buffers to reduce GC pressure
	combinedBuf := make([]byte, mtu+4) // 4 bytes for length + max packet size

	// Process the packet with a dedicated fwPacket instance per packet
	// to avoid shared state issues that can cause duplicate packet processing
	fwPacket := &firewall.Packet{}

	for {
		// Read the packet length and data in potentially multiple operations
		// but minimize allocations
		_, err := io.ReadAtLeast(s, combinedBuf[:4], 4)
		if err != nil {
			if err != io.EOF {
				f.l.WithError(err).WithField("peer", peer).Error("Error reading packet length")
			} else {
				f.l.WithField("peer", peer).Debug("Stream closed by remote peer")
			}
			return
		}

		packetLength := int(binary.BigEndian.Uint32(combinedBuf[:4]))

		// Validate packet length
		if packetLength > mtu || packetLength <= 0 {
			f.l.WithFields(logrus.Fields{
				"length": packetLength,
				"peer":   peer,
			}).Error("Invalid packet length")
			return
		}

		// Read the actual packet data
		_, err = io.ReadFull(s, combinedBuf[4:4+packetLength])
		if err != nil {
			if err != io.EOF {
				f.l.WithError(err).WithField("peer", peer).Error("Error reading packet data")
			} else {
				f.l.WithField("peer", peer).Debug("Unexpected EOF while reading packet")
			}
			return
		}

		// Process the packet - make a copy if needed for async processing
		// to avoid race conditions with the buffer
		packetCopy := make([]byte, packetLength)
		copy(packetCopy, combinedBuf[4:4+packetLength])

		f.readOutsidePackets(packetCopy, fwPacket, peer)
	}
}

func (f *Interface) listenIn(reader io.ReadWriteCloser, i int) {
	runtime.LockOSThread()
	packet := make([]byte, mtu)
	fwPacket := &firewall.Packet{}

	// Read packets from the tun device
	for {

		n, err := reader.Read(packet)
		if err != nil {
			if errors.Is(err, os.ErrClosed) && f.closed.Load() {
				return
			}

			f.l.WithError(err).Error("Error while reading outbound packet")
			// This only seems to happen when something fatal happens to the fd, so exit.
			os.Exit(2)
		}

		f.consumeInsidePacket(packet[:n], fwPacket, i)
	}
}

func (f *Interface) Close() error {
	f.closed.Store(true)
	// Release the tun device
	return f.inside.Close()
}
