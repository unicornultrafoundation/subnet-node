package cmd_exec

import (
	"fmt"
	"testing"
)

func TestVBoxManageVersion(t *testing.T) {
	cmd := GetVBoxCmd()
	if cmd == nil {
		t.Errorf("GetVBoxCmd() returned nil")
	}

	stdout, stderr, err := cmd.Run("--version")
	if err != nil {
		t.Errorf("Failed to run VBoxManage --version: %v", err)
	}

	fmt.Printf("VBoxManage version: %s", stdout)
	if stderr != "" {
		fmt.Printf("Stderr: %s", stderr)
	}
}
