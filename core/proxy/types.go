package proxy

import (
	"fmt"
	"strconv"

	"github.com/libp2p/go-libp2p/core/peer"
	p2peer "github.com/libp2p/go-libp2p/core/peer"
)

// PortMapping represents a parsed port mapping
type PortMapping struct {
	HostIP    string
	LocalPort int
	AppPort   int
	Protocol  string
}

type ProxyApp struct {
	ID          string   `yaml:"id"`
	Ports       []string `yaml:"ports"`
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

func ParseProxyConfig(data map[interface{}]interface{}) (ProxyConfig, error) {
	proxyConfig := ProxyConfig{}

	// Parse "enable" field
	if enable, ok := data["enable"].(bool); ok {
		proxyConfig.Enable = enable
	}

	// Parse "peers" field
	if peers, ok := data["peers"].([]interface{}); ok {
		for _, p := range peers {
			peerMap, ok := p.(map[interface{}]interface{})
			if !ok {
				continue
			}

			peer := ProxyPeer{}
			if id, ok := peerMap["id"].(string); ok {
				peer.ID = id
				peerId, err := p2peer.Decode(id)
				if err != nil {
					return proxyConfig, fmt.Errorf("failed to decode peerID %s: %v", id, err)
				}
				peer.ParsedId = peerId
			}

			// Parse "apps" field
			if apps, ok := peerMap["apps"].([]interface{}); ok {
				for _, a := range apps {
					appMap, ok := a.(map[interface{}]interface{})
					if !ok {
						continue
					}

					app := ProxyApp{}
					if id, ok := appMap["id"].(string); ok {
						app.ID = id
					} else if id, ok := appMap["id"].(int64); ok {
						app.ID = strconv.FormatInt(id, 10) // Convert int to string
					} else {
						return proxyConfig, fmt.Errorf("failed to parse appId for the peerId %s", peer.ID)
					}

					// Parse "ports" field
					if ports, ok := appMap["ports"].([]interface{}); ok {
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
						app.ParsedPorts = parsedPorts
					}

					peer.Apps = append(peer.Apps, app)
				}
			}

			proxyConfig.Peers = append(proxyConfig.Peers, peer)
		}
	}

	return proxyConfig, nil
}
