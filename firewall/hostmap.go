package firewall

import (
	"net/netip"

	"github.com/gaissmai/bart"
	"github.com/sirupsen/logrus"
)

type HostInfo struct {
	// vpnAddrs is a list of vpn addresses assigned to this host that are within our own vpn networks
	// The host may have other vpn addresses that are outside our
	// vpn networks but were removed because they are not usable
	vpnAddrs []netip.Addr

	// networks are both all vpn and unsafe networks assigned to this host
	networks *bart.Lite
}

func (i *HostInfo) buildNetworks(networks, unsafeNetworks []netip.Prefix) {
	if len(networks) == 1 && len(unsafeNetworks) == 0 {
		// Simple case, no CIDRTree needed
		return
	}

	i.networks = new(bart.Lite)
	for _, network := range networks {
		i.networks.Insert(network)
	}

	for _, network := range unsafeNetworks {
		i.networks.Insert(network)
	}
}

func (i *HostInfo) logger(l *logrus.Logger) *logrus.Entry {
	if i == nil {
		return logrus.NewEntry(l)
	}

	li := l.WithField("vpnAddrs", i.vpnAddrs)

	return li
}
