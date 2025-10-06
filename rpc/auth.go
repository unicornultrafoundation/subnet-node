package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/crypto/authchain"
)

var authLog = logrus.WithField("module", "rpc-auth")

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled       bool
	AllowedOwners []common.Address
}

// AuthMiddleware provides authentication middleware for RPC server
type AuthMiddleware struct {
	config         AuthConfig
	authChainCache map[string]*authchain.AuthChainInfo // Cache for validated auth chains
	mu             sync.RWMutex                         // Protects config during reload
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(config AuthConfig) *AuthMiddleware {
	return &AuthMiddleware{
		config:         config,
		authChainCache: make(map[string]*authchain.AuthChainInfo),
	}
}

// UpdateConfig updates the auth configuration (for reload)
func (a *AuthMiddleware) UpdateConfig(config AuthConfig) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.config = config
}

// Wrap wraps the RPC server with authentication middleware
func (a *AuthMiddleware) Wrap(server *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read config with lock
		a.mu.RLock()
		enabled := a.config.Enabled
		a.mu.RUnlock()

		if !enabled {
			// If authentication is disabled, pass through
			server.ServeHTTP(w, r)
			return
		}

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
	// Allow OPTIONS requests (CORS preflight)
	if r.Method == http.MethodOptions {
		return false
	}

	// Allow GET requests (health checks, status endpoints)
	if r.Method == http.MethodGet {
		return false
	}

	// All POST requests require authentication
	if r.Method == http.MethodPost {
		return true
	}

	// Default: require authentication for safety
	return true
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
	a.mu.RLock()
	allowedOwners := a.config.AllowedOwners
	a.mu.RUnlock()

	if len(allowedOwners) == 0 {
		// If no specific owners are configured, allow all authenticated users
		return true
	}

	for _, allowed := range allowedOwners {
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
