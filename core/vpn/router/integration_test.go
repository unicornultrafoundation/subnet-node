package router

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Tests
func TestIntegration_IsEnabled(t *testing.T) {
	// Create config
	cfg := NewMockConfig(map[string]interface{}{
		"vpn.router.enable": true,
	})

	// Create integration
	integration := NewIntegration(cfg)

	// Test IsEnabled
	assert.True(t, integration.IsEnabled())
}

func TestConfigService_GetStreamRouterConfig(t *testing.T) {
	// Create config
	cfg := NewMockConfig(map[string]interface{}{
		"vpn.router.min_streams_per_peer": 2,
		"vpn.router.max_streams_per_peer": 20,
		"vpn.router.connection_ttl":       "15s",
		"vpn.router.cleanup_interval":     "5s",
	})

	// Create config service
	configService := NewConfigService(cfg)

	// Test GetStreamRouterConfig
	routerConfig := configService.GetStreamRouterConfig()
	assert.NotNil(t, routerConfig)
	assert.Equal(t, 2, routerConfig.MinStreamsPerPeer)
	assert.Equal(t, 20, routerConfig.MaxStreamsPerPeer)
	assert.Equal(t, 15*time.Second, routerConfig.ConnectionTTL)
	assert.Equal(t, 5*time.Second, routerConfig.CleanupInterval)
}
