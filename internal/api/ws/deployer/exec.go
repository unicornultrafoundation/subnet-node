package deployer

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/internal/api/ws"
	"k8s.io/client-go/tools/remotecommand"
)

func (api *DeployerWSAPI) GetExecHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := chi.URLParam(r, "orderID")
		logger := api.logger.WithField("orderID", orderID)

		// get podName, serviceName, cmd from query params
		vars := r.URL.Query()
		var cmd []string

		for i := 0; true; i++ {
			v := vars.Get(fmt.Sprintf("cmd%d", i))
			if len(v) == 0 {
				break
			}
			cmd = append(cmd, v)
		}

		tty := vars.Get("tty")
		if len(tty) == 0 {
			logger.Error("missing parameter tty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		isTty := tty == "1"

		serviceName := vars.Get("service")
		if len(serviceName) == 0 {
			logger.Error("missing parameter service")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		stdin := vars.Get("stdin")
		if len(stdin) == 0 {
			logger.Error("missing parameter stdin")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		connectStdin := stdin == "1"

		podName := vars.Get("pod")
		if len(podName) == 0 {
			logger.Error("missing parameter pod")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		api.HandleExec(w, r, orderID, podName, serviceName, cmd, isTty, connectStdin)
	}
}

func (api *DeployerWSAPI) HandleExec(w http.ResponseWriter, r *http.Request, orderID string, podName string, serviceName string, cmd []string, isTty bool, connectStdin bool) {
	logger := api.logger.WithField("orderID", orderID)
	logger.Debug("WebSocket exec connection requested")

	conn, err := ws.SetupWebSocket(w, r, logger)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var stdinPipeOut *io.PipeWriter
	var stdinPipeIn *io.PipeReader
	wg := &sync.WaitGroup{}

	var tsq remotecommand.TerminalSizeQueue
	var terminalSizeUpdate chan remotecommand.TerminalSize
	if isTty {
		terminalSizeUpdate = make(chan remotecommand.TerminalSize, 1)
		tsq = channelToTerminalSizeQueue(terminalSizeUpdate)
	}

	if connectStdin {
		stdinPipeIn, stdinPipeOut = io.Pipe()

		wg.Add(1)
		go deploymentShellWebsocketHandler(logger, wg, conn, stdinPipeOut, terminalSizeUpdate)
	}

	responseData := deploymentShellResponse{}
	l := &sync.Mutex{}

	resultWriter := ws.NewWsWriterWrapper(conn, DeploymentShellCodeResult, l)

	encodeData := true

	status, err := api.deployerService.GetServiceStatus(ctx, orderID, serviceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if status.ReadyReplicas == 0 {
		err = errors.New("no active replicase for service")
		responseData.Message = err.Error()
	}

	if err == nil {
		stdout := ws.NewWsWriterWrapper(conn, DeploymentShellCodeStdout, l)
		stderr := ws.NewWsWriterWrapper(conn, DeploymentShellCodeStderr, l)

		subctx, subcancel := context.WithCancel(r.Context())
		wg.Add(1)
		go deploymentShellPingHandler(subctx, wg, conn)

		var stdinForExec io.Reader
		if connectStdin {
			stdinForExec = stdinPipeIn
		}
		result, err := api.deployerService.Exec(subctx, orderID, podName, serviceName, cmd, stdinForExec, stdout, stderr, isTty, tsq)
		subcancel()

		if result != nil {
			responseData.ExitCode = result.ExitCode()

			logger.Info("deployment shell completed", "exitcode", result.ExitCode())
		} else {
			responseData.Message = err.Error()
			resultWriter = ws.NewWsWriterWrapper(conn, DeploymentShellCodeFailure, l)
			// Don't return errors like this to the client, they could contain information
			// that should not be let out
			encodeData = false

			logger.Error("deployment exec failed", "err", err)
		}
	}

	if encodeData {
		encoder := json.NewEncoder(resultWriter)
		err = encoder.Encode(responseData)
	} else {
		// Just send an empty message so the remote knows things are over
		_, err = resultWriter.Write([]byte{})
	}

	_ = conn.Close()

	if err != nil {
		logger.Error("failed writing response to client after exec", "err", err)
	}

	wg.Wait()

	if stdinPipeOut != nil {
		_ = stdinPipeOut.Close()
	}
	if stdinPipeIn != nil {
		_ = stdinPipeIn.Close()
	}

	if terminalSizeUpdate != nil {
		close(terminalSizeUpdate)
	}
}

func deploymentShellWebsocketHandler(log *logrus.Entry, wg *sync.WaitGroup, shellWs *websocket.Conn, stdinPipeOut io.Writer, terminalSizeUpdate chan<- remotecommand.TerminalSize) {
	defer wg.Done()
	for {
		shellWs.SetPongHandler(func(string) error {
			return shellWs.SetReadDeadline(time.Now().Add(ws.PingWait))
		})

		msgType, data, err := shellWs.ReadMessage()
		if err != nil {
			return
		}

		// Just ignore anything not a binary message or that is empty
		if msgType != websocket.BinaryMessage || len(data) == 0 {
			continue
		}

		msgID := data[0]
		msg := data[1:]
		switch msgID {
		case DeploymentShellCodeStdin:
			_, err := stdinPipeOut.Write(msg)
			if err != nil {
				return
			}
		case DeploymentShellCodeTerminalResize:
			var size remotecommand.TerminalSize
			r := bytes.NewReader(msg)
			// Unpack data, its just binary encoded data in big endian
			err = binary.Read(r, binary.BigEndian, &size.Width)
			if err != nil {
				return
			}
			err = binary.Read(r, binary.BigEndian, &size.Height)
			if err != nil {
				return
			}

			log.Debug("terminal resize received", "width", size.Width, "height", size.Height)
			if terminalSizeUpdate != nil {
				terminalSizeUpdate <- size
			}
		default:
			log.Error("unknown message ID on websocket", "code", msgID)
			return
		}

	}
}

func deploymentShellPingHandler(ctx context.Context, wg *sync.WaitGroup, conn *websocket.Conn) {
	defer wg.Done()
	pingTicker := time.NewTicker(ws.PingPeriod)
	defer pingTicker.Stop()

	for {
		select {
		case <-pingTicker.C:
			if err := ws.SendPing(conn); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

type channelToTerminalSizeQueue <-chan remotecommand.TerminalSize

type deploymentShellResponse struct {
	ExitCode int    `json:"exit_code"`
	Message  string `json:"message,omitempty"`
}

func (sq channelToTerminalSizeQueue) Next() *remotecommand.TerminalSize {
	v, ok := <-sq
	if !ok {
		return nil
	}

	return &v // Interface is dumb and use a pointer
}
