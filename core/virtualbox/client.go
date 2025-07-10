package virtualbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("service", "virtualbox-client")

// VBoxClientImpl implements the VBoxClient interface using VBoxManage commands
type VBoxClientImpl struct {
	homeDir string
}

// NewVBoxClient creates a new VirtualBox client using VBoxManage commands
func NewVBoxClient() (*VBoxClientImpl, error) {
	// Get home directory for VM storage
	homeDir, err := getHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	return &VBoxClientImpl{
		homeDir: homeDir,
	}, nil
}

// getHomeDir returns the home directory for storing VMs
func getHomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return homeDir, nil
}

// CreateVM creates a new VM with the specified name and OS type
func (c *VBoxClientImpl) CreateVM(name, ostype string) error {
	// Create machine using VBoxManage
	baseFolder := filepath.Join(c.homeDir, "VirtualBox VMs")
	cmd := exec.Command("VBoxManage", "createvm", "--name", name, "--register", "--basefolder", baseFolder)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("VBoxManage createvm failed: %w, output: %s", err, string(output))
	}
	return nil
}

// ModifyVM modifies VM settings
func (c *VBoxClientImpl) ModifyVM(name string, params map[string]string) error {
	// Apply parameters using VBoxManage
	for key, value := range params {
		switch key {
		case "memory":
			cmd := exec.Command("VBoxManage", "modifyvm", name, "--memory", value)
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to set memory: %w, output: %s", err, string(output))
			}
		case "cpus":
			cmd := exec.Command("VBoxManage", "modifyvm", name, "--cpus", value)
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to set CPUs: %w, output: %s", err, string(output))
			}
		case "vram":
			cmd := exec.Command("VBoxManage", "modifyvm", name, "--vram", value)
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to set VRAM: %w, output: %s", err, string(output))
			}
		}
	}

	return nil
}

// DeleteVM deletes a VM
func (c *VBoxClientImpl) DeleteVM(name string) error {
	cmd := exec.Command("VBoxManage", "unregistervm", name, "--delete")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("VBoxManage unregistervm failed: %w, output: %s", err, string(output))
	}
	return nil
}

// ListVMs lists all registered VMs
func (c *VBoxClientImpl) ListVMs() ([]string, error) {
	cmd := exec.Command("VBoxManage", "list", "vms")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("VBoxManage list vms failed: %w", err)
	}

	var vmNames []string
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse "VMName" {uuid} format
		if strings.Contains(line, `"`) {
			start := strings.Index(line, `"`)
			end := strings.LastIndex(line, `"`)
			if start != -1 && end != -1 && start != end {
				vmName := line[start+1 : end]
				vmNames = append(vmNames, vmName)
			}
		}
	}

	return vmNames, nil
}

// GetVMInfo gets detailed information about a VM
func (c *VBoxClientImpl) GetVMInfo(name string) (map[string]string, error) {
	cmd := exec.Command("VBoxManage", "showvminfo", name, "--machinereadable")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("VBoxManage showvminfo failed: %w", err)
	}

	info := make(map[string]string)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse key=value format
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.Trim(parts[0], `"`)
				value := strings.Trim(parts[1], `"`)
				info[key] = value
			}
		}
	}

	return info, nil
}

// CreateHD creates a new hard disk
func (c *VBoxClientImpl) CreateHD(filename string, size int) error {
	cmd := exec.Command("VBoxManage", "createhd", "--filename", filename, "--size", strconv.Itoa(size*1024))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("VBoxManage createhd failed: %w, output: %s", err, string(output))
	}
	return nil
}

// StorageController adds a storage controller to a VM
func (c *VBoxClientImpl) StorageController(name, controller string) error {
	cmd := exec.Command("VBoxManage", "storagectl", name, "--name", controller, "--add", "sata", "--controller", "IntelAhci")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("VBoxManage storagectl failed: %w, output: %s", err, string(output))
	}
	return nil
}

// StorageAttach attaches storage to a VM
func (c *VBoxClientImpl) StorageAttach(name, controller string, port, device int, mediumType, medium string) error {
	cmd := exec.Command("VBoxManage", "storageattach", name,
		"--storagectl", controller,
		"--port", strconv.Itoa(port),
		"--device", strconv.Itoa(device),
		"--type", mediumType,
		"--medium", medium)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("VBoxManage storageattach failed: %w, output: %s", err, string(output))
	}
	return nil
}

// StartVM starts a VM
func (c *VBoxClientImpl) StartVM(name string, headless bool) error {
	var cmd *exec.Cmd
	if headless {
		cmd = exec.Command("VBoxManage", "startvm", name, "--type", "headless")
	} else {
		cmd = exec.Command("VBoxManage", "startvm", name)
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("VBoxManage startvm failed: %w, output: %s", err, string(output))
	}
	return nil
}

// StopVM stops a VM
func (c *VBoxClientImpl) StopVM(name string) error {
	cmd := exec.Command("VBoxManage", "controlvm", name, "poweroff")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("VBoxManage controlvm poweroff failed: %w, output: %s", err, string(output))
	}
	return nil
}

// PauseVM pauses a VM
func (c *VBoxClientImpl) PauseVM(name string) error {
	cmd := exec.Command("VBoxManage", "controlvm", name, "pause")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("VBoxManage controlvm pause failed: %w, output: %s", err, string(output))
	}
	return nil
}

// ResumeVM resumes a VM
func (c *VBoxClientImpl) ResumeVM(name string) error {
	cmd := exec.Command("VBoxManage", "controlvm", name, "resume")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("VBoxManage controlvm resume failed: %w, output: %s", err, string(output))
	}
	return nil
}

// ResetVM resets a VM
func (c *VBoxClientImpl) ResetVM(name string) error {
	cmd := exec.Command("VBoxManage", "controlvm", name, "reset")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("VBoxManage controlvm reset failed: %w, output: %s", err, string(output))
	}
	return nil
}

// GetVMStatus gets the current status of a VM
func (c *VBoxClientImpl) GetVMStatus(name string) (string, error) {
	cmd := exec.Command("VBoxManage", "showvminfo", name, "--machinereadable")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("VBoxManage showvminfo failed: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, `VMState="`) {
			start := strings.Index(line, `"`) + 1
			end := strings.LastIndex(line, `"`)
			if start > 0 && end > start {
				return line[start:end], nil
			}
		}
	}

	return "unknown", nil
}

// GetVBoxVersion gets the VirtualBox version
func (c *VBoxClientImpl) GetVBoxVersion() (string, error) {
	cmd := exec.Command("VBoxManage", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("VBoxManage --version failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetHostInfo gets host system information
func (c *VBoxClientImpl) GetHostInfo() (map[string]string, error) {
	info := make(map[string]string)
	info["OS"] = runtime.GOOS
	info["Arch"] = runtime.GOARCH
	info["CPUs"] = strconv.Itoa(runtime.NumCPU())

	// Get VirtualBox version
	if vboxVersion, err := c.GetVBoxVersion(); err == nil {
		info["VBoxVersion"] = vboxVersion
	}

	return info, nil
}

// ListOSTypes lists available OS types
func (c *VBoxClientImpl) ListOSTypes() ([]string, error) {
	cmd := exec.Command("VBoxManage", "list", "ostypes")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("VBoxManage list ostypes failed: %w", err)
	}

	var osTypes []string
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "ID:") {
			osTypes = append(osTypes, line)
		}
	}

	return osTypes, nil
}

// ExecuteCommand executes a VBoxManage command
func (c *VBoxClientImpl) ExecuteCommand(args ...string) (string, error) {
	cmd := exec.Command("VBoxManage", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("VBoxManage command failed: %w, output: %s", err, string(output))
	}
	return string(output), nil
}
