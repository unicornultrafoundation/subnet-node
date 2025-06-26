package deployments

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	bidenginetypes "github.com/unicornultrafoundation/subnet-node/bidengine/types"
	"github.com/unicornultrafoundation/subnet-node/config"
)

// Server represents the deployment API server
type Server struct {
	cfg        *config.C
	service    ServiceManager
	bidMarket  bidenginetypes.BidMarketContract
	logger     *logrus.Logger
	router     *mux.Router
	httpServer *http.Server
	api        *API
}

// NewServer creates a new deployment API server
func NewServer(cfg *config.C, service ServiceManager, bidMarket bidenginetypes.BidMarketContract, logger *logrus.Logger) *Server {
	router := mux.NewRouter()

	// Get provider configuration from config
	configProviderID := big.NewInt(int64(cfg.GetInt("deployment.provider_id", 0)))

	api := NewAPI(service, bidMarket, logger, configProviderID)

	server := &Server{
		cfg:       cfg,
		service:   service,
		bidMarket: bidMarket,
		logger:    logger,
		router:    router,
		api:       api,
	}

	// Set up middleware
	server.setupMiddleware()

	// Register routes
	api.RegisterRoutes(router)

	// Set up HTTP server
	server.httpServer = &http.Server{
		Addr:         server.getServerAddr(),
		Handler:      router,
		ReadTimeout:  server.getReadTimeout(),
		WriteTimeout: server.getWriteTimeout(),
		IdleTimeout:  server.getIdleTimeout(),
	}

	return server
}

// setupMiddleware sets up the server middleware
func (s *Server) setupMiddleware() {
	// CORS middleware
	if s.isCORSEnabled() {
		s.router.Use(s.corsMiddleware)
	}

	// Rate limiting middleware
	if s.isRateLimitEnabled() {
		s.router.Use(s.rateLimitMiddleware)
	}

	// Logging middleware
	s.router.Use(s.loggingMiddleware)

	// Recovery middleware
	s.router.Use(s.recoveryMiddleware)

	// Health check endpoint
	s.router.HandleFunc("/health", s.healthCheck).Methods("GET")
	s.router.HandleFunc("/ready", s.readyCheck).Methods("GET")
}

// corsMiddleware handles CORS requests
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		allowedOrigins := s.getCORSAllowedOrigins()
		if len(allowedOrigins) > 0 {
			origin := r.Header.Get("Origin")
			if origin != "" {
				// Check if origin is allowed
				for _, allowedOrigin := range allowedOrigins {
					if allowedOrigin == "*" || allowedOrigin == origin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// rateLimitMiddleware implements basic rate limiting
func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	// Simple in-memory rate limiter (in production, use Redis or similar)
	clients := make(map[string]*rateLimiter)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr

		limiter, exists := clients[clientIP]
		if !exists {
			limiter = newRateLimiter(s.getRateLimitRequestsPer())
			clients[clientIP] = limiter
		}

		if !limiter.Allow() {
			s.logger.WithField("client_ip", clientIP).Warn("Rate limit exceeded")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		s.logger.WithFields(logrus.Fields{
			"method":     r.Method,
			"path":       r.URL.Path,
			"status":     wrapped.statusCode,
			"duration":   duration,
			"user_agent": r.UserAgent(),
			"remote_ip":  r.RemoteAddr,
		}).Info("HTTP request")
	})
}

// recoveryMiddleware recovers from panics
func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.logger.WithField("error", err).Error("Panic recovered")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// healthCheck handles health check requests
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
}

// readyCheck handles readiness check requests
func (s *Server) readyCheck(w http.ResponseWriter, r *http.Request) {
	// Check if the service is ready
	// You can add more readiness checks here
	// For example, check if the deployment service is running
	// check if database connections are healthy, etc.

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
}

// Start starts the API server
func (s *Server) Start(ctx context.Context) error {
	s.logger.WithFields(logrus.Fields{
		"host": s.getServerHost(),
		"port": s.getServerPort(),
		"tls":  s.isTLSEnabled(),
	}).Info("Starting deployment API server")

	// Start the server in a goroutine
	go func() {
		var err error
		if s.isTLSEnabled() {
			err = s.httpServer.ListenAndServeTLS(s.getTLSCertFile(), s.getTLSKeyFile())
		} else {
			err = s.httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Error("Server error")
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	return s.Stop(ctx)
}

// Stop stops the API server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping deployment API server")

	// Create a context with timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.WithError(err).Error("Error during server shutdown")
		return err
	}

	s.logger.Info("Deployment API server stopped")
	return nil
}

// GetRouter returns the router for testing or additional route registration
func (s *Server) GetRouter() *mux.Router {
	return s.router
}

// GetConfig returns the server configuration
func (s *Server) GetConfig() *config.C {
	return s.cfg
}

// Configuration helper methods

func (s *Server) getServerAddr() string {
	host := s.getServerHost()
	port := s.getServerPort()
	return fmt.Sprintf("%s:%d", host, port)
}

func (s *Server) getServerHost() string {
	return s.cfg.GetString("deployment.api.server.host", "0.0.0.0")
}

func (s *Server) getServerPort() int {
	return s.cfg.GetInt("deployment.api.server.port", 8080)
}

func (s *Server) getReadTimeout() time.Duration {
	return s.cfg.GetDuration("deployment.api.server.read_timeout", 30*time.Second)
}

func (s *Server) getWriteTimeout() time.Duration {
	return s.cfg.GetDuration("deployment.api.server.write_timeout", 30*time.Second)
}

func (s *Server) getIdleTimeout() time.Duration {
	return s.cfg.GetDuration("deployment.api.server.idle_timeout", 120*time.Second)
}

func (s *Server) isTLSEnabled() bool {
	return s.cfg.GetBool("deployment.api.server.tls.enabled", false)
}

func (s *Server) getTLSCertFile() string {
	return s.cfg.GetString("deployment.api.server.tls.cert_file", "")
}

func (s *Server) getTLSKeyFile() string {
	return s.cfg.GetString("deployment.api.server.tls.key_file", "")
}

func (s *Server) isCORSEnabled() bool {
	return s.cfg.GetBool("deployment.api.server.cors.enabled", true)
}

func (s *Server) getCORSAllowedOrigins() []string {
	return s.cfg.GetStringSlice("deployment.api.server.cors.allowed_origins", []string{"*"})
}

func (s *Server) isRateLimitEnabled() bool {
	return s.cfg.GetBool("deployment.api.server.rate_limit.enabled", true)
}

func (s *Server) getRateLimitRequestsPer() int {
	return s.cfg.GetInt("deployment.api.server.rate_limit.requests_per_minute", 100)
}

// Helper types and functions

// rateLimiter implements a simple rate limiter
type rateLimiter struct {
	requests chan time.Time
	limit    int
}

func newRateLimiter(limit int) *rateLimiter {
	rl := &rateLimiter{
		requests: make(chan time.Time, limit),
		limit:    limit,
	}

	// Fill the channel with old timestamps
	for i := 0; i < limit; i++ {
		rl.requests <- time.Now().Add(-time.Minute)
	}

	return rl
}

func (rl *rateLimiter) Allow() bool {
	now := time.Now()

	select {
	case oldest := <-rl.requests:
		// Check if the oldest request is within the last minute
		if now.Sub(oldest) < time.Minute {
			// Put the oldest back and deny
			rl.requests <- oldest
			return false
		}
		// Put current time and allow
		rl.requests <- now
		return true
	default:
		// Channel is full, deny
		return false
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}
