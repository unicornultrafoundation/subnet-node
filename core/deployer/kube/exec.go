package kube

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	executil "k8s.io/client-go/util/exec"
)

type execResult struct {
	exitCode int
}

func (er execResult) ExitCode() int {
	return er.exitCode
}

func (c *KubeClient) Exec(ctx context.Context, deploymentID string, podName string, serviceName string, cmd []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, tty bool, tsq remotecommand.TerminalSizeQueue) (types.ExecResult, error) {
	if c.Config == nil {
		return nil, fmt.Errorf("KubeClient.Config is nil; exec requires rest.Config")
	}

	// Find the pod to get the container name if needed
	pod, err := c.Client.CoreV1().Pods(deploymentID).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	containerName := serviceName
	if containerName == "" && len(pod.Spec.Containers) > 0 {
		containerName = pod.Spec.Containers[0].Name
	}

	req := c.Client.CoreV1().RESTClient().
		Post().
		Namespace(deploymentID).
		Resource("pods").
		Name(podName).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   cmd,
			Stdin:     stdin != nil,
			Stdout:    stdout != nil,
			Stderr:    stderr != nil,
			TTY:       tty,
		}, scheme.ParameterCodec)

	// Make the request with SPDY
	exec, err := remotecommand.NewSPDYExecutor(c.Config, "POST", req.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to create SPDY executor: %w", err)
	}

	// Run, passing in the streams and everything else. This runs until the remote end closes
	// or the streams close
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            stdout,
		Stderr:            stderr,
		Tty:               tty,
		TerminalSizeQueue: tsq,
	})
	if err == nil {
		// No error means the process returned a 0 exit code
		return execResult{exitCode: 0}, nil
	}

	// Check to see if the process ran & returned an exit code
	// If this is true, don't return an error. Something ran in the
	// container which is what this code was trying to do
	terr := executil.CodeExitError{}
	if errors.As(err, &terr) {
		return execResult{exitCode: terr.Code}, nil
	}

	// Some errors are untyped, use string matching to give better answers
	if strings.Contains(err.Error(), "error executing command in container") {
		if strings.Contains(err.Error(), "no such file or directory") || strings.Contains(err.Error(), "executable file not found in $PATH") {
			return nil, fmt.Errorf("remote command execute error: command could not be executed because it does not exist")
		}
		// Don't send the full text of unknown errors back to the user
		// Log the error here so this can be tracked down somehow in the provider logs at least
		c.Logger.Error("command execution failed", "err", err)
		return nil, fmt.Errorf("remote command execute error: command execution failed")
	}

	return nil, err
}
