package firewall

import "net/netip"

type addRuleCall struct {
	incoming  bool
	proto     uint8
	startPort int32
	endPort   int32
	groups    []string
	ip        netip.Prefix
	localIp   netip.Prefix
}

type MockFirewall struct {
	lastCall       addRuleCall
	nextCallReturn error
}

func (mf *MockFirewall) AddRule(incoming bool, proto uint8, startPort int32, endPort int32, groups []string, ip, localIp netip.Prefix) error {
	mf.lastCall = addRuleCall{
		incoming:  incoming,
		proto:     proto,
		startPort: startPort,
		endPort:   endPort,
		groups:    groups,
		ip:        ip,
		localIp:   localIp,
	}

	err := mf.nextCallReturn
	mf.nextCallReturn = nil
	return err
}

func (mf *MockFirewall) AddNetwork(network netip.Prefix) error {
	return nil
}

func (mf *MockFirewall) Drop(fp Packet, incoming bool, localCache ConntrackCache) error {
	return nil
}
