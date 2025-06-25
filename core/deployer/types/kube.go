package types

import (
	"bufio"
	"io"
)

// ServiceLog stores name, stream and scanner
type ServiceLog struct {
	Name    string
	Stream  io.ReadCloser
	Scanner *bufio.Scanner
}

// ServiceLogMessage represents a log message from a service,
// including the service name and the log message content.
type ServiceLogMessage struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

type ExecResult interface {
	ExitCode() int
}
