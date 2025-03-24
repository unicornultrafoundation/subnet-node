package proxy_test

import (
	"reflect"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/proxy"
)

func TestParseProxyConfig(t *testing.T) {
	// Mock YAML data as map[interface{}]interface{}
	peerId := "12D3KooWNNtWuCNRJQxyxJwqMqgSqcKcwXg2N2U5aghbRfp3JS2x"
	yamlData := map[interface{}]interface{}{
		"enable": true,
		"peers": []interface{}{
			map[interface{}]interface{}{
				"id": peerId,
				"apps": []interface{}{
					map[interface{}]interface{}{
						"id":    "2",
						"ports": []interface{}{"9933:8080", "9934:8080", "9935:8080"},
					},
				},
			},
		},
	}

	parsedPeerId, _ := peer.Decode(peerId)

	// Expected result
	expectedConfig := proxy.ProxyConfig{
		Enable: true,
		Peers: []proxy.ProxyPeer{
			{
				ID:       peerId,
				ParsedId: parsedPeerId,
				Apps: []proxy.ProxyApp{
					{
						ID:    "2",
						Ports: []string{"9933:8080", "9934:8080", "9935:8080"},
						ParsedPorts: []proxy.PortMapping{{
							HostIP: "", LocalPort: 9933, AppPort: 8080, Protocol: "tcp",
						}, {
							HostIP: "", LocalPort: 9934, AppPort: 8080, Protocol: "tcp",
						}, {
							HostIP: "", LocalPort: 9935, AppPort: 8080, Protocol: "tcp",
						}},
					},
				},
			},
		},
	}

	// Call the function
	parsedConfig, err := proxy.ParseProxyConfig(yamlData)

	if err != nil {
		t.Errorf("Parsed config failed: %v", err)
	}

	// Compare results
	if !reflect.DeepEqual(parsedConfig, expectedConfig) {
		t.Errorf("Parsed config does not match expected config.\nExpected: %+v\nGot: %+v", expectedConfig, parsedConfig)
	}
}
