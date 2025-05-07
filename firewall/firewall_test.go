package firewall

import (
	"bytes"
	"errors"
	"math"
	"net/netip"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/test"
)

func TestNewFirewall(t *testing.T) {
	l := test.NewLogger()
	fw := NewFirewall(l, time.Second, time.Minute, time.Hour, []netip.Prefix{}, []netip.Prefix{})
	conntrack := fw.Conntrack
	assert.NotNil(t, conntrack)
	assert.NotNil(t, conntrack.Conns)
	assert.NotNil(t, conntrack.TimerWheel)
	assert.NotNil(t, fw.InRules)
	assert.NotNil(t, fw.OutRules)
	assert.Equal(t, time.Second, fw.TCPTimeout)
	assert.Equal(t, time.Minute, fw.UDPTimeout)
	assert.Equal(t, time.Hour, fw.DefaultTimeout)

	assert.Equal(t, time.Hour, conntrack.TimerWheel.wheelDuration)
	assert.Equal(t, time.Hour, conntrack.TimerWheel.wheelDuration)
	assert.Equal(t, 3602, conntrack.TimerWheel.wheelLen)

	fw = NewFirewall(l, time.Second, time.Hour, time.Minute, []netip.Prefix{}, []netip.Prefix{})
	assert.Equal(t, time.Hour, conntrack.TimerWheel.wheelDuration)
	assert.Equal(t, 3602, conntrack.TimerWheel.wheelLen)

	fw = NewFirewall(l, time.Hour, time.Second, time.Minute, []netip.Prefix{}, []netip.Prefix{})
	assert.Equal(t, time.Hour, conntrack.TimerWheel.wheelDuration)
	assert.Equal(t, 3602, conntrack.TimerWheel.wheelLen)

	fw = NewFirewall(l, time.Hour, time.Minute, time.Second, []netip.Prefix{}, []netip.Prefix{})
	assert.Equal(t, time.Hour, conntrack.TimerWheel.wheelDuration)
	assert.Equal(t, 3602, conntrack.TimerWheel.wheelLen)

	fw = NewFirewall(l, time.Minute, time.Hour, time.Second, []netip.Prefix{}, []netip.Prefix{})
	assert.Equal(t, time.Hour, conntrack.TimerWheel.wheelDuration)
	assert.Equal(t, 3602, conntrack.TimerWheel.wheelLen)

	fw = NewFirewall(l, time.Minute, time.Second, time.Hour, []netip.Prefix{}, []netip.Prefix{})
	assert.Equal(t, time.Hour, conntrack.TimerWheel.wheelDuration)
	assert.Equal(t, 3602, conntrack.TimerWheel.wheelLen)
}

func TestFirewall_AddRule(t *testing.T) {
	l := test.NewLogger()
	ob := &bytes.Buffer{}
	l.SetOutput(ob)

	fw := NewFirewall(l, time.Second, time.Minute, time.Hour, []netip.Prefix{}, []netip.Prefix{})
	assert.NotNil(t, fw.InRules)
	assert.NotNil(t, fw.OutRules)

	ti, err := netip.ParsePrefix("1.2.3.4/32")
	require.NoError(t, err)

	require.NoError(t, fw.AddRule(true, ProtoTCP, 1, 1, []string{}, "", netip.Prefix{}, netip.Prefix{}))
	// An empty rule is any
	assert.True(t, fw.InRules.TCP[1].Any.Any.Any)
	assert.Empty(t, fw.InRules.TCP[1].Any.Groups)
	assert.Empty(t, fw.InRules.TCP[1].Any.Hosts)

	fw = NewFirewall(l, time.Second, time.Minute, time.Hour, []netip.Prefix{}, []netip.Prefix{})
	require.NoError(t, fw.AddRule(true, ProtoUDP, 1, 1, []string{"g1"}, "", netip.Prefix{}, netip.Prefix{}))
	assert.Nil(t, fw.InRules.UDP[1].Any.Any)
	assert.Contains(t, fw.InRules.UDP[1].Any.Groups[0].Groups, "g1")
	assert.Empty(t, fw.InRules.UDP[1].Any.Hosts)

	fw = NewFirewall(l, time.Second, time.Minute, time.Hour, []netip.Prefix{}, []netip.Prefix{})
	require.NoError(t, fw.AddRule(true, ProtoICMP, 1, 1, []string{}, "h1", netip.Prefix{}, netip.Prefix{}))
	assert.Nil(t, fw.InRules.ICMP[1].Any.Any)
	assert.Empty(t, fw.InRules.ICMP[1].Any.Groups)
	assert.Contains(t, fw.InRules.ICMP[1].Any.Hosts, "h1")

	fw = NewFirewall(l, time.Second, time.Minute, time.Hour, []netip.Prefix{}, []netip.Prefix{})
	require.NoError(t, fw.AddRule(false, ProtoAny, 1, 1, []string{}, "", ti, netip.Prefix{}))
	assert.Nil(t, fw.OutRules.AnyProto[1].Any.Any)
	_, ok := fw.OutRules.AnyProto[1].Any.CIDR.Get(ti)
	assert.True(t, ok)

	fw = NewFirewall(l, time.Second, time.Minute, time.Hour, []netip.Prefix{}, []netip.Prefix{})
	require.NoError(t, fw.AddRule(false, ProtoAny, 1, 1, []string{}, "", netip.Prefix{}, ti))
	assert.NotNil(t, fw.OutRules.AnyProto[1].Any.Any)
	_, ok = fw.OutRules.AnyProto[1].Any.Any.LocalCIDR.Get(ti)
	assert.True(t, ok)

	fw = NewFirewall(l, time.Second, time.Minute, time.Hour, []netip.Prefix{}, []netip.Prefix{})
	require.NoError(t, fw.AddRule(true, ProtoUDP, 1, 1, []string{"g1"}, "", netip.Prefix{}, netip.Prefix{}))

	fw = NewFirewall(l, time.Second, time.Minute, time.Hour, []netip.Prefix{}, []netip.Prefix{})
	require.NoError(t, fw.AddRule(true, ProtoUDP, 1, 1, []string{"g1"}, "", netip.Prefix{}, netip.Prefix{}))

	fw = NewFirewall(l, time.Second, time.Minute, time.Hour, []netip.Prefix{}, []netip.Prefix{})
	require.NoError(t, fw.AddRule(false, ProtoAny, 0, 0, []string{}, "any", netip.Prefix{}, netip.Prefix{}))
	assert.True(t, fw.OutRules.AnyProto[0].Any.Any.Any)

	fw = NewFirewall(l, time.Second, time.Minute, time.Hour, []netip.Prefix{}, []netip.Prefix{})
	anyIp, err := netip.ParsePrefix("0.0.0.0/0")
	require.NoError(t, err)

	require.NoError(t, fw.AddRule(false, ProtoAny, 0, 0, []string{}, "", anyIp, netip.Prefix{}))
	assert.True(t, fw.OutRules.AnyProto[0].Any.Any.Any)

	// Test error conditions
	fw = NewFirewall(l, time.Second, time.Minute, time.Hour, []netip.Prefix{}, []netip.Prefix{})
	require.Error(t, fw.AddRule(true, math.MaxUint8, 0, 0, []string{}, "", netip.Prefix{}, netip.Prefix{}))
	require.Error(t, fw.AddRule(true, ProtoAny, 10, 0, []string{}, "", netip.Prefix{}, netip.Prefix{}))
}

func TestFirewall_Drop(t *testing.T) {
	l := test.NewLogger()
	ob := &bytes.Buffer{}
	l.SetOutput(ob)

	p := Packet{
		LocalAddr:  netip.MustParseAddr("1.2.3.4"),
		RemoteAddr: netip.MustParseAddr("1.2.3.4"),
		LocalPort:  10,
		RemotePort: 90,
		Protocol:   ProtoUDP,
		Fragment:   false,
	}

	h := HostInfo{
		vpnAddrs: []netip.Addr{netip.MustParseAddr("1.2.3.4")},
	}
	h.buildNetworks([]netip.Prefix{netip.MustParsePrefix("1.2.3.4/24")}, []netip.Prefix{})

	fw := NewFirewall(l, time.Second, time.Minute, time.Hour, []netip.Prefix{netip.MustParsePrefix("1.2.3.4/24")}, []netip.Prefix{})
	require.NoError(t, fw.AddRule(true, ProtoAny, 0, 0, []string{"any"}, "", netip.Prefix{}, netip.Prefix{}))

	// Drop outbound
	assert.Equal(t, ErrNoMatchingRule, fw.Drop(p, false, &h, nil))
	// Allow inbound
	resetConntrack(fw)
	require.NoError(t, fw.Drop(p, true, &h, nil))
	// Allow outbound because conntrack
	require.NoError(t, fw.Drop(p, false, &h, nil))

	// test remote mismatch
	oldRemote := p.RemoteAddr
	p.RemoteAddr = netip.MustParseAddr("1.2.3.10")
	assert.Equal(t, fw.Drop(p, false, &h, nil), ErrInvalidRemoteIP)
	p.RemoteAddr = oldRemote

	// ensure signer doesn't get in the way of group checks
	fw = NewFirewall(l, time.Second, time.Minute, time.Hour, []netip.Prefix{netip.MustParsePrefix("1.2.3.4/24")}, []netip.Prefix{})
	require.NoError(t, fw.AddRule(true, ProtoAny, 0, 0, []string{"nope"}, "", netip.Prefix{}, netip.Prefix{}))
	require.NoError(t, fw.AddRule(true, ProtoAny, 0, 0, []string{"default-group"}, "", netip.Prefix{}, netip.Prefix{}))
	assert.Equal(t, fw.Drop(p, true, &h, nil), nil)
}

func BenchmarkFirewallTable_match(b *testing.B) {
	f := &Firewall{}
	ft := FirewallTable{
		TCP: firewallPort{},
	}

	pfix := netip.MustParsePrefix("172.1.1.1/32")
	_ = ft.TCP.addRule(f, 10, 10, []string{"good-group"}, "good-host", pfix, netip.Prefix{})
	_ = ft.TCP.addRule(f, 100, 100, []string{"good-group"}, "good-host", netip.Prefix{}, pfix)

	b.Run("fail on proto", func(b *testing.B) {
		// This benchmark is showing us the cost of failing to match the protocol
		for n := 0; n < b.N; n++ {
			assert.False(b, ft.match(Packet{Protocol: ProtoUDP}, true))
		}
	})

	b.Run("pass proto, fail on port", func(b *testing.B) {
		// This benchmark is showing us the cost of matching a specific protocol but failing to match the port
		for n := 0; n < b.N; n++ {
			assert.False(b, ft.match(Packet{Protocol: ProtoTCP, LocalPort: 1}, true))
		}
	})

	b.Run("pass proto, port, fail on local CIDR", func(b *testing.B) {
		ip := netip.MustParsePrefix("9.254.254.254/32")
		for n := 0; n < b.N; n++ {
			assert.False(b, ft.match(Packet{Protocol: ProtoTCP, LocalPort: 100, LocalAddr: ip.Addr()}, true))
		}
	})
}

func TestFirewall_Drop2(t *testing.T) {
	l := test.NewLogger()
	ob := &bytes.Buffer{}
	l.SetOutput(ob)

	p := Packet{
		LocalAddr:  netip.MustParseAddr("1.2.3.4"),
		RemoteAddr: netip.MustParseAddr("1.2.3.4"),
		LocalPort:  10,
		RemotePort: 90,
		Protocol:   ProtoUDP,
		Fragment:   false,
	}

	network := netip.MustParsePrefix("1.2.3.4/24")

	h1 := HostInfo{
		vpnAddrs: []netip.Addr{network.Addr()},
	}
	h1.buildNetworks([]netip.Prefix{network}, []netip.Prefix{})

	fw := NewFirewall(l, time.Second, time.Minute, time.Hour, []netip.Prefix{network}, []netip.Prefix{})
	// inbound should fail because the network is not in the networks list
	require.ErrorIs(t, fw.Drop(p, true, &h1, nil), ErrNoMatchingRule)
}

func TestFirewall_Drop3(t *testing.T) {
	l := test.NewLogger()
	ob := &bytes.Buffer{}
	l.SetOutput(ob)

	p := Packet{
		LocalAddr:  netip.MustParseAddr("1.2.3.4"),
		RemoteAddr: netip.MustParseAddr("1.2.3.4"),
		LocalPort:  1,
		RemotePort: 1,
		Protocol:   ProtoUDP,
		Fragment:   false,
	}

	network := netip.MustParsePrefix("1.2.3.4/24")

	h1 := HostInfo{
		vpnAddrs: []netip.Addr{network.Addr()},
	}
	h1.buildNetworks([]netip.Prefix{network}, []netip.Prefix{})

	h2 := HostInfo{
		vpnAddrs: []netip.Addr{network.Addr()},
	}
	h2.buildNetworks([]netip.Prefix{network}, []netip.Prefix{})

	h3 := HostInfo{
		vpnAddrs: []netip.Addr{network.Addr()},
	}
	h3.buildNetworks([]netip.Prefix{network}, []netip.Prefix{})

	fw := NewFirewall(l, time.Second, time.Minute, time.Hour, []netip.Prefix{network}, []netip.Prefix{})
	require.NoError(t, fw.AddRule(true, ProtoAny, 1, 1, []string{}, "host1", netip.Prefix{}, netip.Prefix{}))
	require.NoError(t, fw.AddRule(true, ProtoAny, 1, 1, []string{}, "", netip.Prefix{}, netip.Prefix{}))

	require.NoError(t, fw.Drop(p, true, &h1, nil))
	resetConntrack(fw)
	require.NoError(t, fw.Drop(p, true, &h2, nil))
	resetConntrack(fw)
	assert.Equal(t, fw.Drop(p, true, &h3, nil), nil)

	// Test a remote address match
	fw = NewFirewall(l, time.Second, time.Minute, time.Hour, []netip.Prefix{network}, []netip.Prefix{})
	require.NoError(t, fw.AddRule(true, ProtoAny, 1, 1, []string{}, "", netip.MustParsePrefix("1.2.3.4/24"), netip.Prefix{}))
	require.NoError(t, fw.Drop(p, true, &h1, nil))
}

func TestFirewall_DropConntrackReload(t *testing.T) {
	l := test.NewLogger()
	ob := &bytes.Buffer{}
	l.SetOutput(ob)

	p := Packet{
		LocalAddr:  netip.MustParseAddr("1.2.3.4"),
		RemoteAddr: netip.MustParseAddr("1.2.3.4"),
		LocalPort:  10,
		RemotePort: 90,
		Protocol:   ProtoUDP,
		Fragment:   false,
	}
	network := netip.MustParsePrefix("1.2.3.4/24")

	h := HostInfo{

		vpnAddrs: []netip.Addr{network.Addr()},
	}
	h.buildNetworks([]netip.Prefix{network}, []netip.Prefix{})

	fw := NewFirewall(l, time.Second, time.Minute, time.Hour, []netip.Prefix{network}, []netip.Prefix{})
	require.NoError(t, fw.AddRule(true, ProtoAny, 0, 0, []string{"any"}, "", netip.Prefix{}, netip.Prefix{}))

	// Drop outbound
	assert.Equal(t, fw.Drop(p, false, &h, nil), ErrNoMatchingRule)
	// Allow inbound
	resetConntrack(fw)
	require.NoError(t, fw.Drop(p, true, &h, nil))
	// Allow outbound because conntrack
	require.NoError(t, fw.Drop(p, false, &h, nil))

	oldFw := fw
	fw = NewFirewall(l, time.Second, time.Minute, time.Hour, []netip.Prefix{network}, []netip.Prefix{})
	require.NoError(t, fw.AddRule(true, ProtoAny, 10, 10, []string{"any"}, "", netip.Prefix{}, netip.Prefix{}))
	fw.Conntrack = oldFw.Conntrack
	fw.rulesVersion = oldFw.rulesVersion + 1

	// Allow outbound because conntrack and new rules allow port 10
	require.NoError(t, fw.Drop(p, false, &h, nil))

	oldFw = fw
	fw = NewFirewall(l, time.Second, time.Minute, time.Hour, []netip.Prefix{network}, []netip.Prefix{})
	require.NoError(t, fw.AddRule(true, ProtoAny, 11, 11, []string{"any"}, "", netip.Prefix{}, netip.Prefix{}))
	fw.Conntrack = oldFw.Conntrack
	fw.rulesVersion = oldFw.rulesVersion + 1

	// Drop outbound because conntrack doesn't match new ruleset
	assert.Equal(t, fw.Drop(p, false, &h, nil), ErrNoMatchingRule)
}

func BenchmarkLookup(b *testing.B) {
	ml := func(m map[string]struct{}, a [][]string) {
		for n := 0; n < b.N; n++ {
			for _, sg := range a {
				found := false

				for _, g := range sg {
					if _, ok := m[g]; !ok {
						found = false
						break
					}

					found = true
				}

				if found {
					return
				}
			}
		}
	}

	b.Run("array to map best", func(b *testing.B) {
		m := map[string]struct{}{
			"1ne": {},
			"2wo": {},
			"3hr": {},
			"4ou": {},
			"5iv": {},
			"6ix": {},
		}

		a := [][]string{
			{"1ne", "2wo", "3hr", "4ou", "5iv", "6ix"},
			{"one", "2wo", "3hr", "4ou", "5iv", "6ix"},
			{"one", "two", "3hr", "4ou", "5iv", "6ix"},
			{"one", "two", "thr", "4ou", "5iv", "6ix"},
			{"one", "two", "thr", "fou", "5iv", "6ix"},
			{"one", "two", "thr", "fou", "fiv", "6ix"},
			{"one", "two", "thr", "fou", "fiv", "six"},
		}

		for n := 0; n < b.N; n++ {
			ml(m, a)
		}
	})

	b.Run("array to map worst", func(b *testing.B) {
		m := map[string]struct{}{
			"one": {},
			"two": {},
			"thr": {},
			"fou": {},
			"fiv": {},
			"six": {},
		}

		a := [][]string{
			{"1ne", "2wo", "3hr", "4ou", "5iv", "6ix"},
			{"one", "2wo", "3hr", "4ou", "5iv", "6ix"},
			{"one", "two", "3hr", "4ou", "5iv", "6ix"},
			{"one", "two", "thr", "4ou", "5iv", "6ix"},
			{"one", "two", "thr", "fou", "5iv", "6ix"},
			{"one", "two", "thr", "fou", "fiv", "6ix"},
			{"one", "two", "thr", "fou", "fiv", "six"},
		}

		for n := 0; n < b.N; n++ {
			ml(m, a)
		}
	})
}

func Test_parsePort(t *testing.T) {
	_, _, err := parsePort("")
	require.EqualError(t, err, "was not a number; ``")

	_, _, err = parsePort("  ")
	require.EqualError(t, err, "was not a number; `  `")

	_, _, err = parsePort("-")
	require.EqualError(t, err, "appears to be a range but could not be parsed; `-`")

	_, _, err = parsePort(" - ")
	require.EqualError(t, err, "appears to be a range but could not be parsed; ` - `")

	_, _, err = parsePort("a-b")
	require.EqualError(t, err, "beginning range was not a number; `a`")

	_, _, err = parsePort("1-b")
	require.EqualError(t, err, "ending range was not a number; `b`")

	s, e, err := parsePort(" 1 - 2    ")
	assert.Equal(t, int32(1), s)
	assert.Equal(t, int32(2), e)
	require.NoError(t, err)

	s, e, err = parsePort("0-1")
	assert.Equal(t, int32(0), s)
	assert.Equal(t, int32(0), e)
	require.NoError(t, err)

	s, e, err = parsePort("9919")
	assert.Equal(t, int32(9919), s)
	assert.Equal(t, int32(9919), e)
	require.NoError(t, err)

	s, e, err = parsePort("any")
	assert.Equal(t, int32(0), s)
	assert.Equal(t, int32(0), e)
	require.NoError(t, err)
}

func TestNewFirewallFromConfig(t *testing.T) {
	l := test.NewLogger()

	conf := config.NewC(l)
	conf.Settings["firewall"] = map[string]any{"outbound": "asdf"}
	_, err := NewFirewallFromConfig(l, conf, []netip.Prefix{}, []netip.Prefix{})
	require.EqualError(t, err, "firewall.outbound failed to parse, should be an array of rules")

	// Test both port and code
	conf = config.NewC(l)
	conf.Settings["firewall"] = map[string]any{"outbound": []any{map[string]any{"port": "1", "code": "2"}}}
	_, err = NewFirewallFromConfig(l, conf, []netip.Prefix{}, []netip.Prefix{})
	require.EqualError(t, err, "firewall.outbound rule #0; only one of port or code should be provided")

	// Test missing host, group, cidr, ca_name and ca_sha
	conf = config.NewC(l)
	conf.Settings["firewall"] = map[string]any{"outbound": []any{map[string]any{}}}
	_, err = NewFirewallFromConfig(l, conf, []netip.Prefix{}, []netip.Prefix{})
	require.EqualError(t, err, "firewall.outbound rule #0; at least one of host, group, cidr, local_cidr, ca_name, or ca_sha must be provided")

	// Test code/port error
	conf = config.NewC(l)
	conf.Settings["firewall"] = map[string]any{"outbound": []any{map[string]any{"code": "a", "host": "testh"}}}
	_, err = NewFirewallFromConfig(l, conf, []netip.Prefix{}, []netip.Prefix{})
	require.EqualError(t, err, "firewall.outbound rule #0; code was not a number; `a`")

	conf.Settings["firewall"] = map[string]any{"outbound": []any{map[string]any{"port": "a", "host": "testh"}}}
	_, err = NewFirewallFromConfig(l, conf, []netip.Prefix{}, []netip.Prefix{})
	require.EqualError(t, err, "firewall.outbound rule #0; port was not a number; `a`")

	// Test proto error
	conf = config.NewC(l)
	conf.Settings["firewall"] = map[string]any{"outbound": []any{map[string]any{"code": "1", "host": "testh"}}}
	_, err = NewFirewallFromConfig(l, conf, []netip.Prefix{}, []netip.Prefix{})
	require.EqualError(t, err, "firewall.outbound rule #0; proto was not understood; ``")

	// Test cidr parse error
	conf = config.NewC(l)
	conf.Settings["firewall"] = map[string]any{"outbound": []any{map[string]any{"code": "1", "cidr": "testh", "proto": "any"}}}
	_, err = NewFirewallFromConfig(l, conf, []netip.Prefix{}, []netip.Prefix{})
	require.EqualError(t, err, "firewall.outbound rule #0; cidr did not parse; netip.ParsePrefix(\"testh\"): no '/'")

	// Test local_cidr parse error
	conf = config.NewC(l)
	conf.Settings["firewall"] = map[string]any{"outbound": []any{map[string]any{"code": "1", "local_cidr": "testh", "proto": "any"}}}
	_, err = NewFirewallFromConfig(l, conf, []netip.Prefix{}, []netip.Prefix{})
	require.EqualError(t, err, "firewall.outbound rule #0; local_cidr did not parse; netip.ParsePrefix(\"testh\"): no '/'")

	// Test both group and groups
	conf = config.NewC(l)
	conf.Settings["firewall"] = map[string]any{"inbound": []any{map[string]any{"port": "1", "proto": "any", "group": "a", "groups": []string{"b", "c"}}}}
	_, err = NewFirewallFromConfig(l, conf, []netip.Prefix{}, []netip.Prefix{})
	require.EqualError(t, err, "firewall.inbound rule #0; only one of group or groups should be defined, both provided")
}

func TestAddFirewallRulesFromConfig(t *testing.T) {
	l := test.NewLogger()
	// Test adding tcp rule
	conf := config.NewC(l)
	mf := &mockFirewall{}
	conf.Settings["firewall"] = map[string]any{"outbound": []any{map[string]any{"port": "1", "proto": "tcp", "host": "a"}}}
	require.NoError(t, AddFirewallRulesFromConfig(l, false, conf, mf))
	assert.Equal(t, addRuleCall{incoming: false, proto: ProtoTCP, startPort: 1, endPort: 1, groups: nil, host: "a", ip: netip.Prefix{}, localIp: netip.Prefix{}}, mf.lastCall)

	// Test adding udp rule
	conf = config.NewC(l)
	mf = &mockFirewall{}
	conf.Settings["firewall"] = map[string]any{"outbound": []any{map[string]any{"port": "1", "proto": "udp", "host": "a"}}}
	require.NoError(t, AddFirewallRulesFromConfig(l, false, conf, mf))
	assert.Equal(t, addRuleCall{incoming: false, proto: ProtoUDP, startPort: 1, endPort: 1, groups: nil, host: "a", ip: netip.Prefix{}, localIp: netip.Prefix{}}, mf.lastCall)

	// Test adding icmp rule
	conf = config.NewC(l)
	mf = &mockFirewall{}
	conf.Settings["firewall"] = map[string]any{"outbound": []any{map[string]any{"port": "1", "proto": "icmp", "host": "a"}}}
	require.NoError(t, AddFirewallRulesFromConfig(l, false, conf, mf))
	assert.Equal(t, addRuleCall{incoming: false, proto: ProtoICMP, startPort: 1, endPort: 1, groups: nil, host: "a", ip: netip.Prefix{}, localIp: netip.Prefix{}}, mf.lastCall)

	// Test adding any rule
	conf = config.NewC(l)
	mf = &mockFirewall{}
	conf.Settings["firewall"] = map[string]any{"inbound": []any{map[string]any{"port": "1", "proto": "any", "host": "a"}}}
	require.NoError(t, AddFirewallRulesFromConfig(l, true, conf, mf))
	assert.Equal(t, addRuleCall{incoming: true, proto: ProtoAny, startPort: 1, endPort: 1, groups: nil, host: "a", ip: netip.Prefix{}, localIp: netip.Prefix{}}, mf.lastCall)

	// Test adding rule with cidr
	cidr := netip.MustParsePrefix("10.0.0.0/8")
	conf = config.NewC(l)
	mf = &mockFirewall{}
	conf.Settings["firewall"] = map[string]any{"inbound": []any{map[string]any{"port": "1", "proto": "any", "cidr": cidr.String()}}}
	require.NoError(t, AddFirewallRulesFromConfig(l, true, conf, mf))
	assert.Equal(t, addRuleCall{incoming: true, proto: ProtoAny, startPort: 1, endPort: 1, groups: nil, ip: cidr, localIp: netip.Prefix{}}, mf.lastCall)

	// Test adding rule with local_cidr
	conf = config.NewC(l)
	mf = &mockFirewall{}
	conf.Settings["firewall"] = map[string]any{"inbound": []any{map[string]any{"port": "1", "proto": "any", "local_cidr": cidr.String()}}}
	require.NoError(t, AddFirewallRulesFromConfig(l, true, conf, mf))
	assert.Equal(t, addRuleCall{incoming: true, proto: ProtoAny, startPort: 1, endPort: 1, groups: nil, ip: netip.Prefix{}, localIp: cidr}, mf.lastCall)

	// Test single group
	conf = config.NewC(l)
	mf = &mockFirewall{}
	conf.Settings["firewall"] = map[string]any{"inbound": []any{map[string]any{"port": "1", "proto": "any", "group": "a"}}}
	require.NoError(t, AddFirewallRulesFromConfig(l, true, conf, mf))
	assert.Equal(t, addRuleCall{incoming: true, proto: ProtoAny, startPort: 1, endPort: 1, groups: []string{"a"}, ip: netip.Prefix{}, localIp: netip.Prefix{}}, mf.lastCall)

	// Test single groups
	conf = config.NewC(l)
	mf = &mockFirewall{}
	conf.Settings["firewall"] = map[string]any{"inbound": []any{map[string]any{"port": "1", "proto": "any", "groups": "a"}}}
	require.NoError(t, AddFirewallRulesFromConfig(l, true, conf, mf))
	assert.Equal(t, addRuleCall{incoming: true, proto: ProtoAny, startPort: 1, endPort: 1, groups: []string{"a"}, ip: netip.Prefix{}, localIp: netip.Prefix{}}, mf.lastCall)

	// Test multiple AND groups
	conf = config.NewC(l)
	mf = &mockFirewall{}
	conf.Settings["firewall"] = map[string]any{"inbound": []any{map[string]any{"port": "1", "proto": "any", "groups": []string{"a", "b"}}}}
	require.NoError(t, AddFirewallRulesFromConfig(l, true, conf, mf))
	assert.Equal(t, addRuleCall{incoming: true, proto: ProtoAny, startPort: 1, endPort: 1, groups: []string{"a", "b"}, ip: netip.Prefix{}, localIp: netip.Prefix{}}, mf.lastCall)

	// Test Add error
	conf = config.NewC(l)
	mf = &mockFirewall{}
	mf.nextCallReturn = errors.New("test error")
	conf.Settings["firewall"] = map[string]any{"inbound": []any{map[string]any{"port": "1", "proto": "any", "host": "a"}}}
	require.EqualError(t, AddFirewallRulesFromConfig(l, true, conf, mf), "firewall.inbound rule #0; `test error`")
}

func TestFirewall_convertRule(t *testing.T) {
	l := test.NewLogger()
	ob := &bytes.Buffer{}
	l.SetOutput(ob)

	// Ensure group array of 1 is converted and a warning is printed
	c := map[string]any{
		"group": []any{"group1"},
	}

	r, err := convertRule(l, c, "test", 1)
	assert.Contains(t, ob.String(), "test rule #1; group was an array with a single value, converting to simple value")
	require.NoError(t, err)
	assert.Equal(t, "group1", r.Group)

	// Ensure group array of > 1 is errord
	ob.Reset()
	c = map[string]any{
		"group": []any{"group1", "group2"},
	}

	r, err = convertRule(l, c, "test", 1)
	assert.Empty(t, ob.String())
	require.Error(t, err, "group should contain a single value, an array with more than one entry was provided")

	// Make sure a well formed group is alright
	ob.Reset()
	c = map[string]any{
		"group": "group1",
	}

	r, err = convertRule(l, c, "test", 1)
	require.NoError(t, err)
	assert.Equal(t, "group1", r.Group)
}

type addRuleCall struct {
	incoming  bool
	proto     uint8
	startPort int32
	endPort   int32
	groups    []string
	host      string
	ip        netip.Prefix
	localIp   netip.Prefix
}

type mockFirewall struct {
	lastCall       addRuleCall
	nextCallReturn error
}

func (mf *mockFirewall) AddRule(incoming bool, proto uint8, startPort int32, endPort int32, groups []string, host string, ip netip.Prefix, localIp netip.Prefix) error {
	mf.lastCall = addRuleCall{
		incoming:  incoming,
		proto:     proto,
		startPort: startPort,
		endPort:   endPort,
		groups:    groups,
		host:      host,
		ip:        ip,
		localIp:   localIp,
	}

	err := mf.nextCallReturn
	mf.nextCallReturn = nil
	return err
}

func resetConntrack(fw *Firewall) {
	fw.Conntrack.Lock()
	fw.Conntrack.Conns = map[Packet]*conn{}
	fw.Conntrack.Unlock()
}
