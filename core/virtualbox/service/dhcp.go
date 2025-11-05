package vbox_service

import (
	"bufio"
	"net"
	"strings"
)

type DHCP struct {
	NetworkName string
	IPv4        net.IPNet
	LowerIP     net.IP
	UpperIP     net.IP
	Enabled     bool
}

func (s *VBoxService) addDHCP(kind, name string, d DHCP) error {
	args := []string{"dhcpserver", "add",
		kind, name,
		"--ip", d.IPv4.IP.String(),
		"--netmask", net.IP(d.IPv4.Mask).String(),
		"--lowerip", d.LowerIP.String(),
		"--upperip", d.UpperIP.String(),
	}
	if d.Enabled {
		args = append(args, "--enable")
	} else {
		args = append(args, "--disable")
	}
	_, _, err := s.vBoxCmd.Run(args...)
	return err
}

// AddInternalDHCP adds a DHCP server to an internal network.
func (s *VBoxService) AddInternalDHCP(netname string, d DHCP) error {
	return s.addDHCP("--netname", netname, d)
}

// AddHostonlyDHCP adds a DHCP server to a host-only network.
func (s *VBoxService) AddHostonlyDHCP(ifname string, d DHCP) error {
	return s.addDHCP("--ifname", ifname, d)
}

// DHCPs gets all DHCP server settings in a map keyed by DHCP.NetworkName.
func (s *VBoxService) DHCPs() (map[string]*DHCP, error) {
	out, _, err := s.vBoxCmd.Run("list", "dhcpservers")
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(strings.NewReader(out))
	m := map[string]*DHCP{}
	dhcp := &DHCP{}
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			m[dhcp.NetworkName] = dhcp
			dhcp = &DHCP{}
			continue
		}
		res := reColonLine.FindStringSubmatch(line)
		if res == nil {
			continue
		}
		switch key, val := res[1], res[2]; key {
		case "NetworkName":
			dhcp.NetworkName = val
		case "IP":
			dhcp.IPv4.IP = net.ParseIP(val)
		case "upperIPAddress":
			dhcp.UpperIP = net.ParseIP(val)
		case "lowerIPAddress":
			dhcp.LowerIP = net.ParseIP(val)
		case "NetworkMask":
			dhcp.IPv4.Mask = parseIPv4Mask(val)
		case "Enabled":
			dhcp.Enabled = (val == "Yes")
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return m, nil
}

// ParseIPv4Mask parses IPv4 netmask written in IP form (e.g. 255.255.255.0).
// This function should really belong to the net package.
func parseIPv4Mask(s string) net.IPMask {
	mask := net.ParseIP(s)
	if mask == nil {
		return nil
	}
	return net.IPv4Mask(mask[12], mask[13], mask[14], mask[15])
}
