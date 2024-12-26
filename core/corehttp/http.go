package corehttp

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ...existing code...

// CORSOptions holds the configuration for CORS.
type CORSOptions struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// Authorization holds the configuration for authorization.
type Authorization struct {
	AuthSecret     string
	AllowedMethods []string
}

// WithCORS adds CORS headers to the response.
func WithCORS(corsOptions CORSOptions, next http.Handler) http.Handler {
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
func WithAuthorization(authConfig map[string]Authorization, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		authParts := strings.SplitN(authHeader, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := authParts[1]
		for _, auth := range authConfig {
			if auth.AuthSecret == token {
				var body struct {
					Method string `json:"method"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					http.Error(w, "Bad request", http.StatusBadRequest)
					return
				}

				if contains(auth.AllowedMethods, "*") || contains(auth.AllowedMethods, body.Method) {
					next.ServeHTTP(w, r)
					return
				}
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
		}

		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

// contains checks if a slice contains a specific string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
