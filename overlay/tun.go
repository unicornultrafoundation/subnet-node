package overlay

import (
	"net/netip"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-masternode/config"
	"github.com/unicornultrafoundation/subnet-masternode/util"
)

const DefaultMTU = 1300

// TODO: We may be able to remove routines
type DeviceFactory func(c *config.C, l *logrus.Logger, tunCidr netip.Prefix, routines int) (Device, error)

func NewDeviceFromConfig(c *config.C, l *logrus.Logger, tunCidr netip.Prefix, routines int) (Device, error) {
	switch {
	case c.GetBool("tun.disabled", false):
		tun := newDisabledTun(tunCidr, c.GetInt("tun.tx_queue", 500), c.GetBool("stats.message_metrics", false), l)
		return tun, nil

	default:
		return newTun(c, l, tunCidr, routines > 1)
	}
}

func NewFdDeviceFromConfig(fd *int) DeviceFactory {
	return func(c *config.C, l *logrus.Logger, tunCidr netip.Prefix, routines int) (Device, error) {
		return newTunFromFd(c, l, *fd, tunCidr)
	}
}

func getAllRoutesFromConfig(c *config.C, cidr netip.Prefix, initial bool) (bool, []Route, error) {
	if !initial && !c.HasChanged("tun.routes") && !c.HasChanged("tun.unsafe_routes") {
		return false, nil, nil
	}

	routes, err := parseRoutes(c, cidr)
	if err != nil {
		return true, nil, util.NewContextualError("Could not parse tun.routes", nil, err)
	}

	unsafeRoutes, err := parseUnsafeRoutes(c, cidr)
	if err != nil {
		return true, nil, util.NewContextualError("Could not parse tun.unsafe_routes", nil, err)
	}

	routes = append(routes, unsafeRoutes...)
	return true, routes, nil
}

// findRemovedRoutes will return all routes that are not present in the newRoutes list and would affect the system route table.
// Via is not used to evaluate since it does not affect the system route table.
func findRemovedRoutes(newRoutes, oldRoutes []Route) []Route {
	var removed []Route
	has := func(entry Route) bool {
		for _, check := range newRoutes {
			if check.Equal(entry) {
				return true
			}
		}
		return false
	}

	for _, oldEntry := range oldRoutes {
		if !has(oldEntry) {
			removed = append(removed, oldEntry)
		}
	}

	return removed
}
