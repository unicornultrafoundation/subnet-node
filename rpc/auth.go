package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/crypto/authchain"
)

var authLog = logrus.WithField("module", "rpc-auth")

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled        bool
	AllowedOwners  []common.Address
	RequireAuthFor []string // List of RPC methods that require authentication
	PublicMethods  []string // List of methods that don't require auth
}

// DefaultAuthConfig returns the default authentication configuration
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		Enabled: false,
		// Default public methods that don't require authentication
		PublicMethods: []string{
			"rpc_modules",
			"web3_clientVersion",
			"net_version",
			"eth_chainId",
		},
	}
}

// LoadAuthConfigFromEnv loads auth configuration from environment variables
func LoadAuthConfigFromEnv() *AuthConfig {
	config := DefaultAuthConfig()

	// Check if authentication is enabled
	if enabled := os.Getenv("RPC_AUTH_ENABLED"); enabled == "true" || enabled == "1" {
		config.Enabled = true
	}

	// Load allowed owner addresses
	if ownersStr := os.Getenv("RPC_ALLOWED_OWNERS"); ownersStr != "" {
		addresses := strings.Split(ownersStr, ",")
		for _, addr := range addresses {
			addr = strings.TrimSpace(addr)
			if addr != "" {
				config.AllowedOwners = append(config.AllowedOwners, common.HexToAddress(addr))
			}
		}
	}

	// Load methods that require authentication
	if methodsStr := os.Getenv("RPC_REQUIRE_AUTH_FOR"); methodsStr != "" {
		methods := strings.Split(methodsStr, ",")
		for _, method := range methods {
			method = strings.TrimSpace(method)
			if method != "" {
				config.RequireAuthFor = append(config.RequireAuthFor, method)
			}
		}
	}

	// Load public methods (override defaults if provided)
	if publicStr := os.Getenv("RPC_PUBLIC_METHODS"); publicStr != "" {
		config.PublicMethods = []string{}
		methods := strings.Split(publicStr, ",")
		for _, method := range methods {
			method = strings.TrimSpace(method)
			if method != "" {
				config.PublicMethods = append(config.PublicMethods, method)
			}
		}
	}

	return config
}

// AuthMiddleware provides authentication middleware for RPC server
type AuthMiddleware struct {
	config         *AuthConfig
	authChainCache map[string]*authchain.AuthChainInfo // Cache for validated auth chains
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(config *AuthConfig) *AuthMiddleware {
	if config == nil {
		config = DefaultAuthConfig()
	}

	return &AuthMiddleware{
		config:         config,
		authChainCache: make(map[string]*authchain.AuthChainInfo),
	}
}

// Wrap wraps the RPC server with authentication middleware
func (a *AuthMiddleware) Wrap(server *Server) http.Handler {
	if !a.config.Enabled {
		// If authentication is disabled, return the server as-is
		return server
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if authentication is required for this request
		if !a.requiresAuth(r) {
			server.ServeHTTP(w, r)
			return
		}

		// Authenticate the request
		if err := a.authenticate(r); err != nil {
			a.sendUnauthorized(w, err.Error())
			return
		}

		// Authentication passed, continue
		server.ServeHTTP(w, r)
	})
}

// requiresAuth checks if the request requires authentication
func (a *AuthMiddleware) requiresAuth(r *http.Request) bool {
	// Allow OPTIONS requests
	if r.Method == http.MethodOptions {
		return false
	}

	// Allow GET requests for health checks
	if r.Method == http.MethodGet && r.ContentLength == 0 {
		return false
	}

	// Try to parse the RPC method from the request body
	// Note: This reads the body, but we'll need to restore it
	method := a.extractRPCMethod(r)
	if method == "" {
		// If we can't determine the method, require auth to be safe
		return true
	}

	// Check if method is in public methods list
	for _, publicMethod := range a.config.PublicMethods {
		if method == publicMethod || strings.HasPrefix(method, publicMethod) {
			return false
		}
	}

	// Check if method is specifically marked as requiring auth
	if len(a.config.RequireAuthFor) > 0 {
		for _, authMethod := range a.config.RequireAuthFor {
			if method == authMethod || strings.HasPrefix(method, authMethod) {
				return true
			}
		}
		// If RequireAuthFor is specified and method is not in it, don't require auth
		return false
	}

	// By default, require authentication
	return true
}

// extractRPCMethod extracts the RPC method from the request
func (a *AuthMiddleware) extractRPCMethod(r *http.Request) string {
	// We can check the X-RPC-Method header if set
	if method := r.Header.Get("X-RPC-Method"); method != "" {
		return method
	}

	// For batch requests or if we can't determine, return empty
	// The actual validation will happen in requiresAuth
	return ""
}

// authenticate validates the auth chain in the request
func (a *AuthMiddleware) authenticate(r *http.Request) error {
	// Extract auth chain from header
	authChainHeader := r.Header.Get("X-Auth-Chain")
	if authChainHeader == "" {
		return fmt.Errorf("X-Auth-Chain header required")
	}

	// Check cache first
	if cachedInfo, exists := a.authChainCache[authChainHeader]; exists {
		if cachedInfo.IsValid && !cachedInfo.IsExpired {
			// Verify owner is in allowed list
			if !a.isOwnerAllowed(cachedInfo.OwnerAddress) {
				return fmt.Errorf("owner address not authorized: %s", cachedInfo.OwnerAddress.Hex())
			}
			return nil
		}
		// Remove expired from cache
		delete(a.authChainCache, authChainHeader)
	}

	// Deserialize auth chain
	authChain, err := authchain.DeserializeAuthChain(authChainHeader)
	if err != nil {
		authLog.WithError(err).Warn("Failed to deserialize auth chain")
		return fmt.Errorf("invalid auth chain format: %w", err)
	}

	// Validate auth chain format
	if err := authchain.ValidateAuthChainFormat(authChain); err != nil {
		authLog.WithError(err).Warn("Invalid auth chain format")
		return fmt.Errorf("invalid auth chain format: %w", err)
	}

	// Get auth chain info
	info := authchain.GetAuthChainInfo(authChain)
	if !info.IsValid {
		authLog.WithField("error", info.ValidationError).Warn("Auth chain validation failed")
		return fmt.Errorf("auth chain validation failed: %s", info.ValidationError)
	}

	if info.IsExpired {
		authLog.WithField("owner", info.OwnerAddress.Hex()).Warn("Auth chain expired")
		return fmt.Errorf("auth chain expired")
	}

	// Check if owner is in allowed list
	if !a.isOwnerAllowed(info.OwnerAddress) {
		authLog.WithField("owner", info.OwnerAddress.Hex()).Warn("Owner not authorized")
		return fmt.Errorf("owner address not authorized: %s", info.OwnerAddress.Hex())
	}

	// Cache the validated auth chain
	a.authChainCache[authChainHeader] = info

	authLog.WithFields(logrus.Fields{
		"owner":      info.OwnerAddress.Hex(),
		"ephemeral":  info.EphemeralAddress.Hex(),
		"expires_in": info.TimeUntilExpiry.String(),
	}).Debug("Authentication successful")

	return nil
}

// isOwnerAllowed checks if the owner address is in the allowed list
func (a *AuthMiddleware) isOwnerAllowed(ownerAddr common.Address) bool {
	if len(a.config.AllowedOwners) == 0 {
		// If no specific owners are configured, allow all authenticated users
		return true
	}

	for _, allowed := range a.config.AllowedOwners {
		if ownerAddr == allowed {
			return true
		}
	}

	return false
}

// sendUnauthorized sends an unauthorized error response
func (a *AuthMiddleware) sendUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"error": map[string]interface{}{
			"code":    -32001,
			"message": fmt.Sprintf("Unauthorized: %s", message),
		},
		"id": nil,
	}

	json.NewEncoder(w).Encode(response)
}

// GetOwnerFromRequest extracts the owner address from a request
func (a *AuthMiddleware) GetOwnerFromRequest(r *http.Request) (common.Address, error) {
	authChainHeader := r.Header.Get("X-Auth-Chain")
	if authChainHeader == "" {
		return common.Address{}, fmt.Errorf("no auth chain in request")
	}

	// Check cache first
	if cachedInfo, exists := a.authChainCache[authChainHeader]; exists {
		if cachedInfo.IsValid && !cachedInfo.IsExpired {
			return cachedInfo.OwnerAddress, nil
		}
	}

	// Deserialize and validate
	authChain, err := authchain.DeserializeAuthChain(authChainHeader)
	if err != nil {
		return common.Address{}, err
	}

	return authchain.GetOwnerAddress(authChain)
}

// ClearCache clears the auth chain cache
func (a *AuthMiddleware) ClearCache() {
	a.authChainCache = make(map[string]*authchain.AuthChainInfo)
}

// GetCacheSize returns the current cache size
func (a *AuthMiddleware) GetCacheSize() int {
	return len(a.authChainCache)
}
