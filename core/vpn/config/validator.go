package config

import (
	"errors"
	"net"
	"strconv"
)

// Error codes
var (
	ErrVirtualIPNotSet  = errors.New("virtual IP is not set")
	ErrRoutesNotSet     = errors.New("routes are not set")
	ErrInvalidVirtualIP = errors.New("virtual IP is invalid")
	ErrInvalidRoutes    = errors.New("routes are invalid")
	ErrInvalidMTU       = errors.New("invalid MTU (must be between 576 and 9000)")
)

// Validate checks if the configuration is valid
func (c *VPNConfig) Validate() error {
	if !c.Enable {
		return nil
	}

	if c.VirtualIP == "" {
		return ErrVirtualIPNotSet
	} else if _, _, err := net.ParseCIDR(c.VirtualIP + "/" + strconv.Itoa(c.Subnet)); err != nil {
		return ErrInvalidVirtualIP
	}

	if len(c.Routes) == 0 {
		return ErrRoutesNotSet
	} else {
		for _, route := range c.Routes {
			if _, _, err := net.ParseCIDR(route); err != nil {
				return ErrInvalidRoutes
			}
		}
	}

	if c.MTU < 576 || c.MTU > 9000 {
		return ErrInvalidMTU
	}

	return nil
}
