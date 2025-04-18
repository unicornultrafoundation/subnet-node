package vpn

import (
	"io"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
)

// VPNStream is an interface for the network.Stream to make it easier to mock
type VPNStream interface {
	io.ReadWriteCloser
	Reset() error
	SetDeadline(time.Time) error
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

// Ensure network.Stream implements VPNStream
var _ VPNStream = (network.Stream)(nil)
