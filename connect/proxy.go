package connect

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

// forwardTraffic listens on localPort and forwards traffic to target node
func forwardTraffic(h host.Host, peerID peer.ID, localPort, appName, appPort string) {
	listener, err := net.Listen("tcp", "127.0.0.1:"+localPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", localPort, err)
	}
	defer listener.Close()
	log.Printf("Forwarding local port %s to peer %s (App: %s, Port: %s)\n", localPort, peerID, appName, appPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(h, peerID, conn, appName, appPort)
	}
}

// handleConnection forwards a single TCP connection via P2P
func handleConnection(h host.Host, peerID peer.ID, conn net.Conn, appId, appPort string) {
	defer conn.Close()

	stream, err := h.NewStream(context.Background(), peerID, atypes.ProtocolProxyReverse)
	if err != nil {
		log.Println("Failed to create stream:", err)
		return
	}
	defer stream.Close()

	// Read the incoming request from conn
	reader := bufio.NewReader(conn)
	req, err := http.ReadRequest(reader)
	if err != nil {
		log.Println("❌ Failed to read request:", err)
		return
	}

	// Modify request headers to include X-App-Id and X-App-Port
	req.Header.Set("X-App-Id", appId)
	req.Header.Set("X-App-Port", appPort)

	// Write the modified request to the P2P stream
	err = req.Write(stream)
	if err != nil {
		log.Println("❌ Failed to send request over P2P stream:", err)
		return
	}

	// Forward the TCP traffic (for body and response)

	// Forward the TCP traffic
	go io.Copy(stream, conn)
	io.Copy(conn, stream)
}
