//go:build e2e_testing
// +build e2e_testing

package subnet

// This file contains functions used to export information to the e2e testing framework

import (
	"net/netip"
)

func (i *HostInfo) GetVpnIp() netip.Addr {
	return i.vpnIp
}

func (i *HostInfo) GetLocalIndex() uint32 {
	return i.localIndexId
}

func (i *HostInfo) GetRemoteIndex() uint32 {
	return i.remoteIndexId
}

func (i *HostInfo) GetRelayState() *RelayState {
	return &i.relayState
}
