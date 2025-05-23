package corehttp

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/unicornultrafoundation/subnet-node/config"
)

// ...existing code...

// CORSOptions holds the configuration for CORS.
type CORSOptions struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// convertToStringSlice converts a slice of interface{} to a slice of string.
func convertToStringSlice(slice []any) []string {
	strSlice := make([]string, len(slice))
	for i, v := range slice {
		strSlice[i] = v.(string)
	}
	return strSlice
}

// Authorization holds the configuration for authorization.
type Authorization struct {
	AuthSecret     string
	AllowedMethods []string
}

// parseCORSFromConfig parses CORS options from the provided config.
func parseCORSFromConfig(cfg *config.C) CORSOptions {
	return CORSOptions{
		AllowedOrigins: cfg.GetStringSlice("api.cors.allowed_origins", []string{"*"}),
		AllowedMethods: cfg.GetStringSlice("api.cors.allowed_methods", []string{"*"}),
		AllowedHeaders: cfg.GetStringSlice("api.cors.allowed_headers", []string{"*"}),
	}
}

// WithCORS adds CORS headers to the response.
func WithCORSHeaders(cfg *config.C, next http.Handler) http.Handler {
	corsOptions := parseCORSFromConfig(cfg)
	cfg.RegisterReloadCallback(func(cfg *config.C) {
		corsOptions = parseCORSFromConfig(cfg)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			for _, allowedOrigin := range corsOptions.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
					break
				}
			}
		}

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(corsOptions.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(corsOptions.AllowedHeaders, ", "))
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

// WithAuthorization adds authorization based on the provided config.
func WithAuth(cfg *config.C, next http.Handler) http.Handler {
	authConfig := parseAuthorizationsFromConfig(cfg)
	cfg.RegisterReloadCallback(func(cfg *config.C) {
		authConfig = parseAuthorizationsFromConfig(cfg)
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(authConfig) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		authParts := strings.SplitN(authHeader, " ", 2)

		if len(authParts) != 2 || (authParts[0] != "Bearer" && authParts[0] != "Basic") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := authParts[1]
		_, exists := authConfig[token]
		if !exists {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func parseAuthorizationsFromConfig(cfg *config.C) map[string]Authorization {
	// authorizations is a map where we can just check for the header value to match.
	authorizations := map[string]Authorization{}
	authScopes := cfg.GetMap("api.authorizations", map[string]any{})

	for user, authScope := range authScopes {
		if scopeMap, ok := authScope.(map[string]any); ok {
			expectedHeader := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, scopeMap["auth_secret"])))
			// Encode the auth secret to base64
			authorizations[expectedHeader] = Authorization{
				AuthSecret:     scopeMap["auth_secret"].(string),
				AllowedMethods: convertToStringSlice(scopeMap["allowed_methods"].([]any)),
			}
		}
	}

	return authorizations
}
