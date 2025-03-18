package apps

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/libp2p/go-libp2p/core/network"
)

// Reverse request to specific port inside an app in Docker container
func (s *Service) onReverseRequestReceive(stream network.Stream) {
	defer stream.Close()

	// Read HTTP request from P2P stream
	req, err := http.ReadRequest(bufio.NewReader(stream))
	if err != nil {
		writeErrorResponse(stream, http.StatusBadRequest, "Failed to read request: "+err.Error())
		return
	}

	// Extract target AppId and AppPort from headers
	appIdStr := req.Header.Get("X-App-Id")
	appPort := req.Header.Get("X-App-Port")

	if appIdStr == "" || appPort == "" {
		writeErrorResponse(stream, http.StatusBadRequest, "Missing AppId or AppPort in request headers")
		return
	}

	// Parse AppId
	appId := new(big.Int)
	appId, ok := appId.SetString(appIdStr, 16)
	if !ok {
		writeErrorResponse(stream, http.StatusBadRequest, "Failed to parse appId")
		return
	}

	// Check if container is running
	container, err := s.ContainerInspect(context.Background(), appId)
	if err != nil {
		writeErrorResponse(stream, http.StatusBadRequest, "Failed to read container:"+err.Error())
		return
	}

	containerIP := container.NetworkSettings.IPAddress

	targetURL := "http://" + containerIP + ":" + appPort
	log.Debugln("Forwarding request to", targetURL)

	// Reverse proxy to the selected container
	target, _ := url.Parse(targetURL)
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Capture the response
	rec := &responseRecorder{
		header:     make(http.Header),
		body:       new(bytes.Buffer),
		statusCode: http.StatusOK,
	}

	// Serve request using proxy
	proxy.ServeHTTP(rec, req)

	// Write response back to stream
	writeResponse(stream, rec)
}

// Utility function to write an HTTP response over P2P
func writeResponse(stream network.Stream, rec *responseRecorder) {
	writer := bufio.NewWriter(stream)

	// Write status line
	fmt.Fprintf(writer, "HTTP/1.1 %d %s\r\n", rec.statusCode, http.StatusText(rec.statusCode))

	// Write headers
	for k, v := range rec.header {
		for _, val := range v {
			fmt.Fprintf(writer, "%s: %s\r\n", k, val)
		}
	}
	fmt.Fprint(writer, "\r\n") // End of headers

	// Write body
	writer.Write(rec.body.Bytes())
	writer.Flush()
}

// Utility function to write an error response
func writeErrorResponse(stream network.Stream, statusCode int, message string) {
	log.Debugln(message)
	writer := bufio.NewWriter(stream)
	fmt.Fprintf(writer, "HTTP/1.1 %d %s\r\n", statusCode, http.StatusText(statusCode))
	fmt.Fprintf(writer, "Content-Length: %d\r\n", len(message))
	fmt.Fprint(writer, "\r\n") // End of headers
	fmt.Fprint(writer, message)
	writer.Flush()
}

// Custom ResponseRecorder to capture proxy response
type responseRecorder struct {
	header     http.Header
	body       *bytes.Buffer
	statusCode int
}

func (r *responseRecorder) Header() http.Header {
	return r.header
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}
