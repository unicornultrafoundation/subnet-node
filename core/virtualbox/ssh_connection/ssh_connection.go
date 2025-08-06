package ssh_connection

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// VMPort represents a virtual machine port configuration
type VMPort struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

// SSHServer manages VM SSH configurations
type SSHServer struct {
	vmConfigs map[string]*VMPort
	mu        sync.RWMutex
}

// SSHConnection represents an active SSH connection
type SSHConnection struct {
	VMID      string
	SSHConn   *ssh.Client
	Session   *ssh.Session
	WSConn    *websocket.Conn
	SSHServer *SSHServer
	SSHIn     io.WriteCloser
	SSHOut    io.Reader
	IsActive  bool
	mu        sync.Mutex
}

// SSHToken represents an SSH access token
type SSHToken struct {
	Token     string    `json:"token"`
	VMID      string    `json:"vm_id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
}

// SSHTokenResponse represents the response for token generation
type SSHTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// WebSocketHandler handles WebSocket connections for SSH access
type WebSocketHandler struct {
	sshServer     *SSHServer
	tokens        map[string]*SSHToken
	tokenMu       sync.RWMutex
	logger        *logrus.Entry
	cleanupTicker *time.Ticker
	stopChan      chan struct{}
}

// NewSSHServer creates a new SSHServer instance
func NewSSHServer() *SSHServer {
	return &SSHServer{
		vmConfigs: make(map[string]*VMPort),
	}
}

// AddVMConfig adds a VM configuration to the SSH server
func (s *SSHServer) AddVMConfig(vmID string, host string, port string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vmConfigs[vmID] = &VMPort{
		Host: host,
		Port: port,
	}
}

// GetVMConfig retrieves a VM configuration by ID
func (s *SSHServer) GetVMConfig(vmID string) (*VMPort, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	config, exists := s.vmConfigs[vmID]
	return config, exists
}

// RemoveVMConfig removes a VM configuration by ID
func (s *SSHServer) RemoveVMConfig(vmID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.vmConfigs, vmID)
}

// GetAllVMConfigs returns all VM configurations
func (s *SSHServer) GetAllVMConfigs() map[string]*VMPort {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a copy to avoid race conditions
	configs := make(map[string]*VMPort)
	for k, v := range s.vmConfigs {
		configs[k] = &VMPort{
			Host: v.Host,
			Port: v.Port,
		}
	}
	return configs
}

// NewWebSocketHandler creates a new WebSocket handler for SSH connections
func NewWebSocketHandler(sshServer *SSHServer, logger *logrus.Entry) *WebSocketHandler {
	return &WebSocketHandler{
		sshServer:     sshServer,
		tokens:        make(map[string]*SSHToken),
		logger:        logger,
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

	h.logger.WithField("vmID", vmID).WithField("token", token[:8]+"...").Info("Generated SSH access token")

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

	h.logger.WithField("vmID", accessToken.VMID).WithField("token", token[:8]+"...").Info("SSH access token consumed")

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
			h.logger.WithField("token", token[:8]+"...").Debug("Cleaned up expired SSH access token")
		}
	}
}

// HandleWebSocket handles the WebSocket connection for SSH access
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request, vmID string, vmValidator func(string) error) error {
	// Validate VM exists and is running
	if err := vmValidator(vmID); err != nil {
		return fmt.Errorf("VM validation failed: %w", err)
	}

	logger := h.logger.WithField("vmID", vmID)

	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		logger.Error("Missing SSH access token")
		return fmt.Errorf("SSH access token is required as query parameter")
	}

	// Validate and consume the token
	vmPayload, err := h.ValidateAndConsumeSSHToken(token)
	if err != nil {
		logger.WithError(err).Error("Invalid or expired SSH access token")
		return fmt.Errorf("invalid or expired token: %w", err)
	}

	// Verify the token is for the correct VM
	if vmPayload.VMID != vmID {
		logger.Error("Token is for a different VM")
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
		logger.WithError(err).Error("Failed to upgrade WebSocket connection")
		return fmt.Errorf("WebSocket upgrade failed: %w", err)
	}
	defer conn.Close()

	// Send initial connection message
	err = conn.WriteJSON(map[string]string{"message": "WebSocket connected successfully"})
	if err != nil {
		logger.WithError(err).Error("Failed to send initial message")
		return fmt.Errorf("failed to send initial message: %w", err)
	}

	// Create SSH connection
	sshConn := NewSSHConnection(vmID, conn, h.sshServer)

	// Verify VM exists in SSHServer configs
	_, exists := h.sshServer.GetVMConfig(vmID)
	if !exists {
		logger.Error("VM not found in SSHServer configs")
		return fmt.Errorf("VM SSH configuration not found")
	}

	// Connect to SSH
	if err := sshConn.Connect(username, password); err != nil {
		logger.WithError(err).Error("Failed to establish SSH connection")
		// Send error message to WebSocket
		errorMsg := map[string]string{"error": fmt.Sprintf("SSH connection failed: %v", err)}
		if writeErr := conn.WriteJSON(errorMsg); writeErr != nil {
			logger.WithError(writeErr).Error("Failed to send error response")
		}
		return fmt.Errorf("SSH connection failed: %w", err)
	}

	logger.Info("SSH connection established successfully")

	// Start SSH connection handling
	sshConn.Start()

	// Wait for connection to close (handled by SSH connection)
	// The connection will remain active until WebSocket closes or SSH connection is terminated
	return nil
}

// NewSSHConnection creates a new SSH connection
func NewSSHConnection(vmID string, wsConn *websocket.Conn, sshServer *SSHServer) *SSHConnection {
	return &SSHConnection{
		VMID:      vmID,
		WSConn:    wsConn,
		IsActive:  false,
		SSHServer: sshServer,
	}
}

// Connect establishes the SSH connection
func (sc *SSHConnection) Connect(user, pass string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	config, exists := sc.SSHServer.GetVMConfig(sc.VMID)
	if !exists {
		return fmt.Errorf("VM config not found for ID: %s", sc.VMID)
	}

	hostport := fmt.Sprintf("%s:%s", config.Host, config.Port)

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshConn, err := ssh.Dial("tcp", hostport, sshConfig)
	if err != nil {
		return fmt.Errorf("SSH dial error: %v", err)
	}

	session, err := sshConn.NewSession()
	if err != nil {
		sshConn.Close()
		return fmt.Errorf("SSH session error: %v", err)
	}

	sshOut, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		sshConn.Close()
		return fmt.Errorf("STDOUT pipe error: %v", err)
	}

	sshIn, err := session.StdinPipe()
	if err != nil {
		session.Close()
		sshConn.Close()
		return fmt.Errorf("STDIN pipe error: %v", err)
	}

	if err := session.RequestPty("xterm", 80, 40, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.ECHOCTL:       0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		session.Close()
		sshConn.Close()
		return fmt.Errorf("Request PTY error: %v", err)
	}

	if err := session.Shell(); err != nil {
		session.Close()
		sshConn.Close()
		return fmt.Errorf("Start shell error: %v", err)
	}

	sc.SSHConn = sshConn
	sc.Session = session
	sc.SSHIn = sshIn
	sc.SSHOut = sshOut
	sc.IsActive = true

	log.Printf("SSH connection established for VM %s to %s", sc.VMID, hostport)
	return nil
}

// Start starts the SSH connection handling
func (sc *SSHConnection) Start() {
	// Handle SSH output
	go sc.handleSSHOutput()

	// Handle WebSocket input
	go sc.handleWebSocketInput()
}

// handleSSHOutput handles data from SSH to WebSocket
func (sc *SSHConnection) handleSSHOutput() {
	defer sc.Close()

	buf := make([]byte, 1024)
	for {
		n, err := sc.SSHOut.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Read from SSH stdout error for VM %s: %v", sc.VMID, err)
			}
			return
		}
		if n > 0 {
			err = sc.WSConn.WriteMessage(websocket.BinaryMessage, buf[:n])
			if err != nil {
				log.Printf("Write to WebSocket error for VM %s: %v", sc.VMID, err)
				return
			}
		}
	}
}

// handleWebSocketInput handles data from WebSocket to SSH
func (sc *SSHConnection) handleWebSocketInput() {
	defer sc.Close()

	for {
		messageType, p, err := sc.WSConn.ReadMessage()
		if err != nil {
			if err != io.EOF {
				log.Printf("Read from WebSocket error for VM %s: %v", sc.VMID, err)
			}
			return
		}

		sc.mu.Lock()
		if !sc.IsActive {
			sc.mu.Unlock()
			return
		}
		sc.mu.Unlock()

		if messageType == websocket.BinaryMessage || messageType == websocket.TextMessage {
			_, err = sc.SSHIn.Write(p)
			if err != nil {
				log.Printf("Write to SSH stdin error for VM %s: %v", sc.VMID, err)
				return
			}
		}
	}
}

// Close closes the SSH connection
func (sc *SSHConnection) Close() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.IsActive {
		sc.IsActive = false
		if sc.Session != nil {
			sc.Session.Close()
		}
		if sc.SSHConn != nil {
			sc.SSHConn.Close()
		}
		log.Printf("SSH connection closed for VM %s", sc.VMID)
	}
}

// generateSecureToken generates a secure random token for SSH access
func generateSecureToken() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const tokenLength = 32

	b := make([]byte, tokenLength)
	for i := range b {
		// Generate a random index into the charset
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return ""
		}
		b[i] = charset[idx.Int64()]
	}
	return string(b)
}
