package virtualbox

import (
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/gorilla/websocket"
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

// / Implementation for SSHConnection
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

	// config, exists := sc.SSHServer.GetVMConfig(sc.VMID)
	// if !exists {
	// 	return fmt.Errorf("VM config not found for ID: %s", sc.VMID)
	// }

	// hostport := fmt.Sprintf("%s:%s", config.Host, config.Port)
	// fmt.Println("hostport", hostport)

	hostport := "127.0.0.1:58431"

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

	if err := session.RequestPty("xterm", 80, 40, ssh.TerminalModes{}); err != nil {
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
