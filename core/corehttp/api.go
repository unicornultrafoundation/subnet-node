package corehttp

import (
	"net"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core"
	"github.com/unicornultrafoundation/subnet-node/internal/api"
	"github.com/unicornultrafoundation/subnet-node/rpc"
)

// APIPath is the path at which the API is mounted.
const APIPath = "/"

// parseAuthConfig parses auth configuration from config.C
func parseAuthConfig(cfg *config.C) rpc.AuthConfig {
	authConfig := rpc.AuthConfig{
		Enabled:       cfg.GetBool("api.auth.enabled", false),
		AllowedOwners: []common.Address{},
	}

	// Parse allowed owners
	ownersSlice := cfg.GetStringSlice("api.auth.allowed_owners", []string{})
	for _, ownerStr := range ownersSlice {
		if ownerStr != "" {
			authConfig.AllowedOwners = append(authConfig.AllowedOwners, common.HexToAddress(ownerStr))
		}
	}

	return authConfig
}

func APIOption() ServeOption {
	return func(n *core.SubnetNode, _ net.Listener, smux *http.ServeMux) (*http.ServeMux, error) {
		cfg := n.Repo.Config()

		server := rpc.NewServer()
		server.RegisterName("version", api.NewVersionAPI())         // Register the VersionAPI
		server.RegisterName("keystore", api.NewKeystoreAPI(n.Repo)) // Register the KeystoreAPI

		// Parse auth config from config.C with reload support
		authConfig := parseAuthConfig(cfg)
		authMiddleware := rpc.NewAuthMiddleware(authConfig)
		
		// Register reload callback to update auth config when config changes
		cfg.RegisterReloadCallback(func(reloadedCfg *config.C) {
			newAuthConfig := parseAuthConfig(reloadedCfg)
			authMiddleware.UpdateConfig(newAuthConfig)
		})

		// Wrap with auth middleware first, then CORS
		handler := WithCORSHeaders(cfg, authMiddleware.Wrap(server))
		smux.Handle(APIPath, handler)
		return smux, nil
	}
}
