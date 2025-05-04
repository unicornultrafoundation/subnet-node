package overlay

import (
	"io"
	"net/netip"

	"github.com/unicornultrafoundation/subnet-node/core/vpn/routing"
)

type Device interface {
	io.ReadWriteCloser
	Activate() error
	Networks() []netip.Prefix
	Name() string
	RoutesFor(netip.Addr) routing.Gateways
	NewMultiQueueReader() (io.ReadWriteCloser, error)
}
