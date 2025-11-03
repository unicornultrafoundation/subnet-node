package ssh_connection

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketHandler handles WebSocket connections for SSH access
type WebSocketHandler struct {
	sshServer     *SSHServer
	tokens        map[string]*SSHToken
	tokenMu       sync.RWMutex
	cleanupTicker *time.Ticker
	stopChan      chan struct{}
}

// NewWebSocketHandler creates a new WebSocket handler for SSH connections
func NewWebSocketHandler(sshServer *SSHServer) *WebSocketHandler {
	return &WebSocketHandler{
		sshServer:     sshServer,
		tokens:        make(map[string]*SSHToken),
		cleanupTicker: time.NewTicker(5 * time.Minute), // Clean up expired tokens every 5 minutes
		stopChan:      make(chan struct{}),
	}
}

// StartCleanup starts the background token cleanup routine
func (h *WebSocketHandler) StartCleanup(stopChan chan struct{}) {
	go func() {
		for {
			select {
			case <-h.cleanupTicker.C:
				h.CleanupExpiredTokens()
			case <-stopChan:
				return
			}
		}
	}()
}

// StopCleanup stops the token cleanup routine
func (h *WebSocketHandler) StopCleanup() {
	if h.cleanupTicker != nil {
		h.cleanupTicker.Stop()
	}
}

// GenerateSSHToken generates a one-time access token for SSH connections
func (h *WebSocketHandler) GenerateSSHToken(vmID string, username string, password string) (*SSHTokenResponse, error) {
	// Validate VM exists in SSHServer configs
	_, exists := h.sshServer.GetVMConfig(vmID)
	if !exists {
		return nil, fmt.Errorf("VM not found in SSH configuration")
	}

	// Generate a secure random token
	token := generateSecureToken()

	// Set expiration time (15 minutes from now)
	expiresAt := time.Now().Add(15 * time.Minute)

	// Store the token
	h.tokenMu.Lock()
	h.tokens[token] = &SSHToken{
		Token:     token,
		VMID:      vmID,
		Username:  username,
		Password:  password,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		Used:      false,
	}
	h.tokenMu.Unlock()

	return &SSHTokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// ValidateAndConsumeSSHToken validates a token and returns the credentials if valid
func (h *WebSocketHandler) ValidateAndConsumeSSHToken(token string) (*SSHToken, error) {
	h.tokenMu.Lock()
	defer h.tokenMu.Unlock()

	accessToken, exists := h.tokens[token]
	if !exists {
		return nil, fmt.Errorf("invalid token")
	}

	if accessToken.Used {
		return nil, fmt.Errorf("token already used")
	}

	if time.Now().After(accessToken.ExpiresAt) {
		// Remove expired token
		delete(h.tokens, token)
		return nil, fmt.Errorf("token expired")
	}

	// Mark token as used and remove it
	accessToken.Used = true
	delete(h.tokens, token)

	return accessToken, nil
}

// CleanupExpiredTokens removes expired tokens from memory
func (h *WebSocketHandler) CleanupExpiredTokens() {
	h.tokenMu.Lock()
	defer h.tokenMu.Unlock()

	now := time.Now()
	for token, accessToken := range h.tokens {
		if now.After(accessToken.ExpiresAt) {
			delete(h.tokens, token)
		}
	}
}

// HandleWebSocket handles the WebSocket connection for SSH access
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request, vmID string, vmValidator func(string) error) error {
	// Validate VM exists and is running
	if err := vmValidator(vmID); err != nil {
		return fmt.Errorf("VM validation failed: %w", err)
	}

	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		return fmt.Errorf("SSH access token is required as query parameter")
	}

	// Validate and consume the token
	vmPayload, err := h.ValidateAndConsumeSSHToken(token)
	if err != nil {
		return fmt.Errorf("invalid or expired token: %w", err)
	}

	// Verify the token is for the correct VM
	if vmPayload.VMID != vmID {
		return fmt.Errorf("token is not valid for this VM")
	}

	username := vmPayload.Username
	password := vmPayload.Password

	// Upgrade to WebSocket connection
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return fmt.Errorf("WebSocket upgrade failed: %w", err)
	}
	defer conn.Close()

	// Send initial connection message
	err = conn.WriteJSON(map[string]string{"message": "WebSocket connected successfully"})
	if err != nil {
		return fmt.Errorf("failed to send initial message: %w", err)
	}

	// Create SSH connection
	sshConn := NewSSHConnection(vmID, conn, h.sshServer)

	// Verify VM exists in SSHServer configs
	_, exists := h.sshServer.GetVMConfig(vmID)
	if !exists {
		return fmt.Errorf("VM SSH configuration not found")
	}

	// Connect to SSH
	if err := sshConn.Connect(username, password); err != nil {
		// Send error message to WebSocket
		errorMsg := map[string]string{"error": fmt.Sprintf("SSH connection failed: %v", err)}
		if writeErr := conn.WriteJSON(errorMsg); writeErr != nil {
		}
		return fmt.Errorf("SSH connection failed: %w", err)
	}

	// Start SSH connection handling
	sshConn.Start()

	// Wait for connection to close (handled by SSH connection)
	// The connection will remain active until WebSocket closes or SSH connection is terminated
	return nil
}
