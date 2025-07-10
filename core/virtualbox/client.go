package virtualbox

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("service", "virtualbox-client")

// VBoxClientImpl implements the VBoxClient interface
type VBoxClientImpl struct {
	vboxManagePath string
	homeDir        string
}

// NewVBoxClient creates a new VirtualBox client
func NewVBoxClient() (*VBoxClientImpl, error) {
	// Find VBoxManage executable
	vboxPath, err := findVBoxManage()
	if err != nil {
		return nil, fmt.Errorf("failed to find VBoxManage: %w", err)
	}

	// Get home directory for VM storage
	homeDir, err := getHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	return &VBoxClientImpl{
		vboxManagePath: vboxPath,
		homeDir:        homeDir,
	}, nil
}

// findVBoxManage locates the VBoxManage executable
func findVBoxManage() (string, error) {
	// Common paths for VBoxManage
	paths := []string{
		"VBoxManage",
		"/usr/bin/VBoxManage",
		"/usr/local/bin/VBoxManage",
		"/opt/homebrew/bin/VBoxManage", // macOS ARM
		"C:\\Program Files\\Oracle\\VirtualBox\\VBoxManage.exe",
	}

	for _, path := range paths {
		if _, err := exec.LookPath(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("VBoxManage not found in PATH or common locations")
}

// getHomeDir returns the home directory for storing VMs
func getHomeDir() (string, error) {
	homeDir, err := exec.Command("echo", "$HOME").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(homeDir)), nil
}

// ExecuteCommand executes a VBoxManage command
func (c *VBoxClientImpl) ExecuteCommand(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.vboxManagePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("VBoxManage command failed: %w, output: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

// CreateVM creates a new VM with the specified name and OS type
func (c *VBoxClientImpl) CreateVM(name, ostype string) error {
	args := []string{"createvm", "--name", name, "--ostype", ostype, "--register"}
	_, err := c.ExecuteCommand(args...)
	return err
}

// ModifyVM modifies VM settings
func (c *VBoxClientImpl) ModifyVM(name string, params map[string]string) error {
	args := []string{"modifyvm", name}
	for key, value := range params {
		args = append(args, fmt.Sprintf("--%s", key), value)
	}
	_, err := c.ExecuteCommand(args...)
	return err
}

// DeleteVM deletes a VM
func (c *VBoxClientImpl) DeleteVM(name string) error {
	args := []string{"unregistervm", name, "--delete"}
	_, err := c.ExecuteCommand(args...)
	return err
}

// ListVMs lists all registered VMs
func (c *VBoxClientImpl) ListVMs() ([]string, error) {
	output, err := c.ExecuteCommand("list", "vms")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(output, "\n")
	var vms []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Parse "VMName" {uuid} format
		if idx := strings.Index(line, " "); idx > 0 {
			vmName := strings.Trim(line[:idx], "\"")
			vms = append(vms, vmName)
		}
	}

	return vms, nil
}

// GetVMInfo gets detailed information about a VM
func (c *VBoxClientImpl) GetVMInfo(name string) (map[string]string, error) {
	output, err := c.ExecuteCommand("showvminfo", name)
	if err != nil {
		return nil, err
	}

	info := make(map[string]string)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if idx := strings.Index(line, ":"); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			info[key] = value
		}
	}

	return info, nil
}

// CreateHD creates a new hard disk
func (c *VBoxClientImpl) CreateHD(filename string, size int) error {
	args := []string{"createhd", "--filename", filename, "--size", strconv.Itoa(size), "--format", "VDI"}
	_, err := c.ExecuteCommand(args...)
	return err
}

// StorageController adds a storage controller to a VM
func (c *VBoxClientImpl) StorageController(name, controller string) error {
	args := []string{"storagectl", name, "--name", controller, "--add", "virtio-scsi", "--bootable", "on"}
	_, err := c.ExecuteCommand(args...)
	return err
}

// StorageAttach attaches storage to a VM
func (c *VBoxClientImpl) StorageAttach(name, controller string, port, device int, mediumType, medium string) error {
	args := []string{
		"storageattach", name,
		"--storagectl", controller,
		"--port", strconv.Itoa(port),
		"--device", strconv.Itoa(device),
		"--type", mediumType,
		"--medium", medium,
	}
	_, err := c.ExecuteCommand(args...)
	return err
}

// StartVM starts a VM
func (c *VBoxClientImpl) StartVM(name string, headless bool) error {
	args := []string{"startvm", name}
	if headless {
		args = append(args, "--type", "headless")
	}
	_, err := c.ExecuteCommand(args...)
	return err
}

// StopVM stops a VM
func (c *VBoxClientImpl) StopVM(name string) error {
	args := []string{"controlvm", name, "poweroff"}
	_, err := c.ExecuteCommand(args...)
	return err
}

// PauseVM pauses a VM
func (c *VBoxClientImpl) PauseVM(name string) error {
	args := []string{"controlvm", name, "pause"}
	_, err := c.ExecuteCommand(args...)
	return err
}

// ResumeVM resumes a VM
func (c *VBoxClientImpl) ResumeVM(name string) error {
	args := []string{"controlvm", name, "resume"}
	_, err := c.ExecuteCommand(args...)
	return err
}

// ResetVM resets a VM
func (c *VBoxClientImpl) ResetVM(name string) error {
	args := []string{"controlvm", name, "reset"}
	_, err := c.ExecuteCommand(args...)
	return err
}

// GetVMStatus gets the current status of a VM
func (c *VBoxClientImpl) GetVMStatus(name string) (string, error) {
	info, err := c.GetVMInfo(name)
	if err != nil {
		return "", err
	}

	if state, ok := info["VMState"]; ok {
		return strings.ToLower(state), nil
	}

	return "unknown", nil
}

// GetVBoxVersion gets the VirtualBox version
func (c *VBoxClientImpl) GetVBoxVersion() (string, error) {
	output, err := c.ExecuteCommand("--version")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// GetHostInfo gets information about the host system
func (c *VBoxClientImpl) GetHostInfo() (map[string]string, error) {
	info := make(map[string]string)

	// OS and architecture
	info["os"] = runtime.GOOS
	info["arch"] = runtime.GOARCH

	// CPU info
	if cpuInfo, err := c.ExecuteCommand("list", "hostcpuids"); err == nil {
		info["cpu_info"] = cpuInfo
	}

	// Memory info
	if memInfo, err := c.ExecuteCommand("list", "hostinfo"); err == nil {
		info["memory_info"] = memInfo
	}

	return info, nil
}
