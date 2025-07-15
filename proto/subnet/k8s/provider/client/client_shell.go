package client

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"k8s.io/client-go/tools/remotecommand"

	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"
)

var (
	errLeaseShell              = errors.New("lease shell failed")
	ErrLeaseShellProviderError = fmt.Errorf("%w: the provider encountered an unknown error", errLeaseShell)
)

func (c *client) LeaseShell(
	ctx context.Context,
	lID mtypes.LeaseID,
	service string,
	podIndex uint,
	cmd []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	tty bool,
	terminalResize <-chan remotecommand.TerminalSize,
) error {
	return nil
}

func processRemoteError(input io.Reader) error {
	dec := json.NewDecoder(input)
	var v LeaseShellResponse
	err := dec.Decode(&v)
	if err != nil {
		return fmt.Errorf("%w: failed parsing response data from provider", err)
	}

	if 0 != len(v.Message) {
		return fmt.Errorf("%w: %s", errLeaseShell, v.Message)
	}

	if 0 != v.ExitCode {
		return fmt.Errorf("%w: remote process exited with code %d", errLeaseShell, v.ExitCode)
	}

	return nil
}

func handleStdin(ctx context.Context, input io.Reader, output io.Writer, saveError func(string, error)) {
	data := make([]byte, 4096)

	for {
		n, err := input.Read(data)
		if err != nil {
			saveError("reading from stdin", err)
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		_, err = output.Write(data[0:n])
		if err != nil {
			saveError("writing stdin data to remote", err)
			return
		}
	}
}

func handleTerminalResize(ctx context.Context, wg *sync.WaitGroup, input <-chan remotecommand.TerminalSize, output io.Writer, saveError func(string, error)) {
	defer wg.Done()

	buf := &bytes.Buffer{}
	for {
		var size remotecommand.TerminalSize
		var ok bool
		select {
		case <-ctx.Done():
			return
		case size, ok = <-input:
			if !ok { // Channel has closed
				return
			}

		}

		// Clear the buffer, then pack in both values
		buf.Reset()
		err := binary.Write(buf, binary.BigEndian, size.Width)
		if err != nil {
			saveError("encoding terminal size width", err)
			return
		}
		err = binary.Write(buf, binary.BigEndian, size.Height)
		if err != nil {
			saveError("encoding terminal size height", err)
			return
		}

		_, err = output.Write((buf).Bytes())
		if err != nil {
			saveError("sending terminal size to remote", err)
			return
		}
	}
}
