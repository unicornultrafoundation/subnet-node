package proxy

import (
	"fmt"
	"strconv"

	"github.com/libp2p/go-libp2p/core/peer"
)

// PortMapping represents a parsed port mapping
type PortMapping struct {
	HostIP    string
	LocalPort int
	AppPort   int
	Protocol  string
	AllowIPs  []string // List of allowed source IPs (CIDR or single IP)
}

type ProxyApp struct {
	ID          string   `yaml:"id"`
	Ports       []string `yaml:"ports"`
	AllowIPs    []string `yaml:"allow_ips"` // List of allowed source IPs for the app
	ParsedPorts []PortMapping
}

type ProxyPeer struct {
	ID       string `yaml:"id"`
	ParsedId peer.ID
	Apps     []ProxyApp `yaml:"apps"`
}

type ProxyConfig struct {
	Enable bool        `yaml:"enable"`
	Peers  []ProxyPeer `yaml:"peers"`
}

func ParseProxyConfig(data map[string]any) (ProxyConfig, error) {
	proxyConfig := ProxyConfig{}

	// Parse "enable" field
	if enable, ok := data["enable"].(bool); ok {
		proxyConfig.Enable = enable
	}

	// Parse "peers" field
	if peers, ok := data["peers"].([]any); ok {
		for _, p := range peers {
			peerMap, ok := p.(map[string]any)
			if !ok {
				continue
			}

			proxyPeer := ProxyPeer{}
			if id, ok := peerMap["id"].(string); ok {
				proxyPeer.ID = id
				parsedId, err := peer.Decode(id)
				if err != nil {
					return proxyConfig, fmt.Errorf("failed to decode peerID %s: %v", id, err)
				}
				proxyPeer.ParsedId = parsedId
			}

			// Parse "apps" field
			if apps, ok := peerMap["apps"].([]any); ok {
				for _, a := range apps {
					appMap, ok := a.(map[string]any)
					if !ok {
						continue
					}

					app := ProxyApp{}
					if id, ok := appMap["id"].(string); ok {
						app.ID = id
					} else if id, ok := appMap["id"].(int64); ok {
						app.ID = strconv.FormatInt(id, 10) // Convert int to string
					} else {
						return proxyConfig, fmt.Errorf("failed to parse appId for the peerId %s", proxyPeer.ID)
					}

					// Parse "allow_ips" field
					if allowIPs, ok := appMap["allow_ips"].([]any); ok {
						for _, ip := range allowIPs {
							if ipStr, ok := ip.(string); ok {
								app.AllowIPs = append(app.AllowIPs, ipStr)
							}
						}
					}

					// Parse "ports" field
					if ports, ok := appMap["ports"].([]any); ok {
						for _, port := range ports {
							if portStr, ok := port.(string); ok {
								app.Ports = append(app.Ports, portStr)
							} else if portInt, ok := port.(int); ok {
								app.Ports = append(app.Ports, strconv.Itoa(portInt))
							}
						}

						if len(app.Ports) == 0 {
							return proxyConfig, fmt.Errorf("this appId %s needs to config port mappings", app.ID)
						}

						parsedPorts, err := ParsePortMappings(app.Ports)
						if err != nil {
							return proxyConfig, fmt.Errorf("failed to parse port mapping for this appId %s: %v", app.ID, err)
						}
						// Apply app-level AllowIPs to each port mapping if no port-specific AllowIPs are defined
						for i := range parsedPorts {
							parsedPorts[i].AllowIPs = app.AllowIPs
						}
						app.ParsedPorts = parsedPorts
					}

					proxyPeer.Apps = append(proxyPeer.Apps, app)
				}
			}

			proxyConfig.Peers = append(proxyConfig.Peers, proxyPeer)
		}
	}

	return proxyConfig, nil
}
