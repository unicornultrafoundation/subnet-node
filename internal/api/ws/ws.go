package ws

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	// Time allowed writing the file to the client.
	PingWait = 15 * time.Second

	// Time allowed reading the next pong message from the client.
	PongWait = 15 * time.Second

	// Send pings to a client with this period. Must be less than pongWait.
	PingPeriod = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WsConn = websocket.Conn

// Add a helper to setup the websocket connection
func SetupWebSocket(w http.ResponseWriter, r *http.Request, logger *logrus.Entry) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.WithError(err).Error("Failed to upgrade WebSocket connection")
		// Return a proper error response instead of letting the handler continue
		http.Error(w, fmt.Sprintf("WebSocket upgrade failed: %v", err), http.StatusBadRequest)
		return nil, err
	}

	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(PingWait))
	})

	logger.Debug("WebSocket connection established")

	// Set up connection with proper read deadline
	conn.SetReadLimit(512)                                 // Limit message size
	conn.SetReadDeadline(time.Now().Add(30 * time.Second)) // Initial read deadline

	return conn, nil
}

// Add a helper to setup the websocket connection
func SetupWebSocketForVirtualBox(w http.ResponseWriter, r *http.Request, logger *logrus.Entry) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.WithError(err).Error("Failed to upgrade WebSocket connection")
		// Return a proper error response instead of letting the handler continue
		http.Error(w, fmt.Sprintf("WebSocket upgrade failed: %v", err), http.StatusBadRequest)
		return nil, err
	}

	logger.Debug("WebSocket connection established for VirtualBox")

	return conn, nil
}

// SetupWebSocketForMetrics sets up a WebSocket connection specifically for metrics streaming
// with read deadlines for faster disconnection detection
func SetupWebSocketForMetrics(w http.ResponseWriter, r *http.Request, logger *logrus.Entry) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.WithError(err).Error("Failed to upgrade WebSocket connection")
		// Return a proper error response instead of letting the handler continue
		http.Error(w, fmt.Sprintf("WebSocket upgrade failed: %v", err), http.StatusBadRequest)
		return nil, err
	}

	// Set up connection with proper read deadline and handlers for faster disconnection detection
	conn.SetReadLimit(512)                                 // Limit message size
	conn.SetReadDeadline(time.Now().Add(30 * time.Second)) // Initial read deadline

	// Set up pong handler to reset read deadline
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	})

	logger.Debug("WebSocket connection established for VirtualBox metrics")

	return conn, nil
}

// SetupWebSocketWithPingHandler sets up a WebSocket connection with ping handling
func SetupWebSocketWithPingHandler(w http.ResponseWriter, r *http.Request, logger *logrus.Entry) (*websocket.Conn, error) {
	conn, err := SetupWebSocket(w, r, logger)
	if err != nil {
		return nil, err
	}

	// Set up ping handler
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(PingWait))
	})

	return conn, nil
}

// SendErrorResponse sends an error message as JSON to the WebSocket connection
func SendErrorResponse(conn *websocket.Conn, errorMsg string) error {
	errorResponse := map[string]string{"error": errorMsg}
	return conn.WriteJSON(errorResponse)
}

// SendPing sends a ping message to the WebSocket connection
func SendPing(conn *websocket.Conn) error {
	const pingWriteWaitTime = 5 * time.Second
	return conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(pingWriteWaitTime))
}

// WrappedConnection interface exposes the single method that the wrapper uses
type WrappedConnection interface {
	WriteMessage(int, []byte) error
}

// WsWriterWrapper provides a thread-safe way to write binary messages with a type prefix
type WsWriterWrapper struct {
	connection WrappedConnection
	id         byte
	buf        bytes.Buffer
	l          sync.Locker
}

// NewWsWriterWrapper creates a new WebSocket writer wrapper
func NewWsWriterWrapper(conn WrappedConnection, id byte, l sync.Locker) *WsWriterWrapper {
	return &WsWriterWrapper{
		connection: conn,
		l:          l,
		id:         id,
	}
}

// Write implements io.Writer interface
func (wsw *WsWriterWrapper) Write(data []byte) (int, error) {
	myBuf := &wsw.buf
	myBuf.Reset()
	_ = myBuf.WriteByte(wsw.id)
	_, _ = myBuf.Write(data)

	wsw.l.Lock()
	defer wsw.l.Unlock()
	err := wsw.connection.WriteMessage(websocket.BinaryMessage, myBuf.Bytes())

	return len(data), err
}
