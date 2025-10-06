package rpc

import (
	"net/http"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestAuthMiddleware_requiresAuth(t *testing.T) {
	middleware := NewAuthMiddleware(AuthConfig{Enabled: true})

	tests := []struct {
		name         string
		httpMethod   string
		expectedAuth bool
	}{
		{
			name:         "POST request - requires auth",
			httpMethod:   http.MethodPost,
			expectedAuth: true,
		},
		{
			name:         "GET request - no auth",
			httpMethod:   http.MethodGet,
			expectedAuth: false,
		},
		{
			name:         "OPTIONS request - no auth",
			httpMethod:   http.MethodOptions,
			expectedAuth: false,
		},
		{
			name:         "PUT request - requires auth",
			httpMethod:   http.MethodPut,
			expectedAuth: true,
		},
		{
			name:         "DELETE request - requires auth",
			httpMethod:   http.MethodDelete,
			expectedAuth: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.httpMethod, "/", nil)
			result := middleware.requiresAuth(req)
			if result != tt.expectedAuth {
				t.Errorf("requiresAuth() = %v, want %v for HTTP method %s", result, tt.expectedAuth, tt.httpMethod)
			}
		})
	}
}

func TestAuthMiddleware_isOwnerAllowed(t *testing.T) {
	addr1 := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	addr2 := common.HexToAddress("0x8ba1f109551bD432803012645Ac136ddd64DBA72")
	addr3 := common.HexToAddress("0x0000000000000000000000000000000000000001")

	tests := []struct {
		name     string
		config   AuthConfig
		address  common.Address
		expected bool
	}{
		{
			name: "No whitelist - allow all",
			config: AuthConfig{
				Enabled:       true,
				AllowedOwners: []common.Address{},
			},
			address:  addr1,
			expected: true,
		},
		{
			name: "Address in whitelist",
			config: AuthConfig{
				Enabled:       true,
				AllowedOwners: []common.Address{addr1, addr2},
			},
			address:  addr1,
			expected: true,
		},
		{
			name: "Address not in whitelist",
			config: AuthConfig{
				Enabled:       true,
				AllowedOwners: []common.Address{addr1, addr2},
			},
			address:  addr3,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewAuthMiddleware(tt.config)
			result := middleware.isOwnerAllowed(tt.address)
			if result != tt.expected {
				t.Errorf("isOwnerAllowed() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAuthMiddleware_UpdateConfig(t *testing.T) {
	addr1 := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	addr2 := common.HexToAddress("0x8ba1f109551bD432803012645Ac136ddd64DBA72")

	initialConfig := AuthConfig{
		Enabled:       false,
		AllowedOwners: []common.Address{addr1},
	}

	middleware := NewAuthMiddleware(initialConfig)

	// Verify initial config
	if middleware.config.Enabled {
		t.Error("Expected initial Enabled to be false")
	}
	if len(middleware.config.AllowedOwners) != 1 {
		t.Errorf("Expected 1 initial owner, got %d", len(middleware.config.AllowedOwners))
	}

	// Update config
	newConfig := AuthConfig{
		Enabled:       true,
		AllowedOwners: []common.Address{addr1, addr2},
	}
	middleware.UpdateConfig(newConfig)

	// Verify updated config
	middleware.mu.RLock()
	if !middleware.config.Enabled {
		t.Error("Expected Enabled to be true after update")
	}
	if len(middleware.config.AllowedOwners) != 2 {
		t.Errorf("Expected 2 owners after update, got %d", len(middleware.config.AllowedOwners))
	}
	middleware.mu.RUnlock()
}

func TestAuthMiddleware_Wrap(t *testing.T) {
	server := NewServer()

	t.Run("Auth disabled - no middleware", func(t *testing.T) {
		config := AuthConfig{Enabled: false}
		middleware := NewAuthMiddleware(config)
		handler := middleware.Wrap(server)

		if handler == nil {
			t.Error("Expected non-nil handler")
		}
	})

	t.Run("Auth enabled - middleware active", func(t *testing.T) {
		config := AuthConfig{Enabled: true}
		middleware := NewAuthMiddleware(config)
		handler := middleware.Wrap(server)

		if handler == nil {
			t.Error("Expected non-nil handler")
		}
	})
}
