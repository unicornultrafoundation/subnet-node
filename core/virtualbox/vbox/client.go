package vbox

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/config"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

// Client represents a VirtualBox client
type Client struct {
	config *config.ServiceConfig
	logger *logrus.Logger
}

// NewClient creates a new VirtualBox client
func NewClient(cfg *config.ServiceConfig, logger *logrus.Logger) *Client {
	return &Client{
		config: cfg,
		logger: logger,
	}
}

// CreateVM creates a VM using VBoxManage
func (c *Client) CreateVM(ctx context.Context, vmID string, vmConfig *types.VMConfig) (*types.VMInfo, error) {
	vmName := fmt.Sprintf("subnet-vm-%s", vmID)

	// Create VM
	if err := c.createVM(ctx, vmName, vmConfig); err != nil {
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}

	// Configure VM settings
	if err := c.configureVM(ctx, vmName, vmConfig); err != nil {
		// Clean up on failure
		c.DeleteVM(ctx, vmName)
		return nil, fmt.Errorf("failed to configure VM: %w", err)
	}

	// Get VM info
	vmInfo, err := c.GetVMInfo(ctx, vmName)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM info: %w", err)
	}

	c.logger.WithField("vmID", vmID).WithField("vmName", vmName).Info("Successfully created VM")
	return vmInfo, nil
}

// createVM creates the basic VM
func (c *Client) createVM(ctx context.Context, vmName string, vmConfig *types.VMConfig) error {
	args := []string{
		"createvm",
		"--name", vmName,
		"--ostype", vmConfig.OSType,
		"--register",
	}

	cmd := exec.CommandContext(ctx, c.config.VBoxManagePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create VM: %s, %w", string(output), err)
	}

	return nil
}

// configureVM configures VM settings
func (c *Client) configureVM(ctx context.Context, vmName string, vmConfig *types.VMConfig) error {
	// Set memory
	if err := c.setMemory(ctx, vmName, vmConfig.MemoryMB); err != nil {
		return err
	}

	// Set CPU count
	if err := c.setCPUs(ctx, vmName, vmConfig.CPUs); err != nil {
		return err
	}

	// Create and attach storage
	if err := c.configureStorage(ctx, vmName, vmConfig); err != nil {
		return err
	}

	// Configure network
	if err := c.configureNetwork(ctx, vmName, vmConfig); err != nil {
		return err
	}

	// Configure advanced settings
	if err := c.configureAdvancedSettings(ctx, vmName, vmConfig); err != nil {
		return err
	}

	return nil
}

// setMemory sets the VM memory
func (c *Client) setMemory(ctx context.Context, vmName string, memoryMB int) error {
	args := []string{
		"modifyvm",
		vmName,
		"--memory", strconv.Itoa(memoryMB),
	}

	cmd := exec.CommandContext(ctx, c.config.VBoxManagePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set memory: %s, %w", string(output), err)
	}

	return nil
}

// setCPUs sets the VM CPU count
func (c *Client) setCPUs(ctx context.Context, vmName string, cpus int) error {
	args := []string{
		"modifyvm",
		vmName,
		"--cpus", strconv.Itoa(cpus),
	}

	cmd := exec.CommandContext(ctx, c.config.VBoxManagePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set CPUs: %s, %w", string(output), err)
	}

	return nil
}

// configureStorage configures VM storage
func (c *Client) configureStorage(ctx context.Context, vmName string, vmConfig *types.VMConfig) error {
	// Create storage controller
	controllerName := "SATA Controller"
	args := []string{
		"storagectl",
		vmName,
		"--name", controllerName,
		"--add", "sata",
		"--controller", "IntelAhci",
	}

	cmd := exec.CommandContext(ctx, c.config.VBoxManagePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create storage controller: %s, %w", string(output), err)
	}

	// Create virtual disk
	diskPath := fmt.Sprintf("%s/%s.vdi", c.config.DefaultVMPath, vmName)
	args = []string{
		"createhd",
		"--filename", diskPath,
		"--size", strconv.Itoa(vmConfig.DiskSizeGB * 1024), // Convert GB to MB
	}

	cmd = exec.CommandContext(ctx, c.config.VBoxManagePath, args...)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create virtual disk: %s, %w", string(output), err)
	}

	// Attach disk to controller
	args = []string{
		"storageattach",
		vmName,
		"--storagectl", controllerName,
		"--port", "0",
		"--device", "0",
		"--type", "hdd",
		"--medium", diskPath,
	}

	cmd = exec.CommandContext(ctx, c.config.VBoxManagePath, args...)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to attach disk: %s, %w", string(output), err)
	}

	// Attach base image if specified
	if vmConfig.BaseImage != "" {
		args = []string{
			"storageattach",
			vmName,
			"--storagectl", controllerName,
			"--port", "1",
			"--device", "0",
			"--type", "hdd",
			"--medium", vmConfig.BaseImage,
		}

		cmd = exec.CommandContext(ctx, c.config.VBoxManagePath, args...)
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to attach base image: %s, %w", string(output), err)
		}
	}

	return nil
}

// configureNetwork configures VM networking
func (c *Client) configureNetwork(ctx context.Context, vmName string, vmConfig *types.VMConfig) error {
	networkType := vmConfig.NetworkType
	if networkType == "" {
		networkType = c.config.DefaultNetworkType
	}

	args := []string{
		"modifyvm",
		vmName,
		"--nic1", networkType,
	}

	if networkType == "bridged" && vmConfig.BridgeName != "" {
		args = append(args, "--bridgeadapter1", vmConfig.BridgeName)
	}

	if vmConfig.MACAddress != "" {
		args = append(args, "--macaddress1", vmConfig.MACAddress)
	}

	cmd := exec.CommandContext(ctx, c.config.VBoxManagePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to configure network: %s, %w", string(output), err)
	}

	return nil
}

// configureAdvancedSettings configures advanced VM settings
func (c *Client) configureAdvancedSettings(ctx context.Context, vmName string, vmConfig *types.VMConfig) error {
	args := []string{"modifyvm", vmName}

	// Audio settings
	if vmConfig.EnableAudio {
		args = append(args, "--audio", "pulse")
	} else {
		args = append(args, "--audio", "none")
	}

	// USB settings
	if vmConfig.EnableUSB {
		args = append(args, "--usb", "on")
	}

	// VRDE settings
	if vmConfig.EnableVRDE {
		args = append(args, "--vrde", "on")
		if vmConfig.VRDEPort > 0 {
			args = append(args, "--vrdeport", strconv.Itoa(vmConfig.VRDEPort))
		}
	}

	// PAE settings
	if vmConfig.EnablePAE {
		args = append(args, "--pae", "on")
	}

	// Nested paging settings
	if vmConfig.EnableNestedPaging {
		args = append(args, "--nested-hw-virt", "on")
	}

	// Custom settings
	for key, value := range vmConfig.CustomSettings {
		args = append(args, fmt.Sprintf("--%s", key), value)
	}

	cmd := exec.CommandContext(ctx, c.config.VBoxManagePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to configure advanced settings: %s, %w", string(output), err)
	}

	return nil
}

// StartVM starts a VM
func (c *Client) StartVM(ctx context.Context, vmName string) error {
	args := []string{"startvm", vmName}
	if c.config.VBoxHeadless {
		args = append(args, "--type", "headless")
	}

	cmd := exec.CommandContext(ctx, c.config.VBoxManagePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start VM: %s, %w", string(output), err)
	}

	c.logger.WithField("vmName", vmName).Info("Started VM")
	return nil
}

// StopVM stops a VM
func (c *Client) StopVM(ctx context.Context, vmName string) error {
	args := []string{"controlvm", vmName, "poweroff"}

	cmd := exec.CommandContext(ctx, c.config.VBoxManagePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop VM: %s, %w", string(output), err)
	}

	c.logger.WithField("vmName", vmName).Info("Stopped VM")
	return nil
}

// ShutdownVM gracefully shuts down a VM
func (c *Client) ShutdownVM(ctx context.Context, vmName string) error {
	args := []string{"controlvm", vmName, "acpipowerbutton"}

	cmd := exec.CommandContext(ctx, c.config.VBoxManagePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to shutdown VM: %s, %w", string(output), err)
	}

	c.logger.WithField("vmName", vmName).Info("Shutdown VM")
	return nil
}

// DeleteVM deletes a VM
func (c *Client) DeleteVM(ctx context.Context, vmName string) error {
	// First, stop the VM if it's running
	state, err := c.GetVMState(ctx, vmName)
	if err == nil && state == types.VMStateRunning {
		if err := c.StopVM(ctx, vmName); err != nil {
			c.logger.WithError(err).WithField("vmName", vmName).Warn("Failed to stop VM before deletion")
		}
	}

	// Delete the VM
	args := []string{"unregistervm", vmName, "--delete"}

	cmd := exec.CommandContext(ctx, c.config.VBoxManagePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete VM: %s, %w", string(output), err)
	}

	c.logger.WithField("vmName", vmName).Info("Deleted VM")
	return nil
}

// GetVMState gets the current state of a VM
func (c *Client) GetVMState(ctx context.Context, vmName string) (types.VMState, error) {
	args := []string{"showvminfo", vmName, "--machinereadable"}

	cmd := exec.CommandContext(ctx, c.config.VBoxManagePath, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get VM state: %w", err)
	}

	// Parse the output to find VMState
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "VMState=") {
			state := strings.Trim(strings.TrimPrefix(line, "VMState="), "\"")
			return types.VMState(state), nil
		}
	}

	return "", fmt.Errorf("VM state not found in output")
}

// GetVMInfo gets detailed information about a VM
func (c *Client) GetVMInfo(ctx context.Context, vmName string) (*types.VMInfo, error) {
	args := []string{"showvminfo", vmName, "--machinereadable"}

	cmd := exec.CommandContext(ctx, c.config.VBoxManagePath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get VM info: %w", err)
	}

	vmInfo := &types.VMInfo{
		Name: vmName,
	}

	// Parse the output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := strings.Trim(parts[1], "\"")

		switch key {
		case "UUID":
			vmInfo.UUID = value
		case "VMState":
			vmInfo.State = types.VMState(value)
		case "OSType":
			vmInfo.OSType = value
		case "Memory":
			if memory, err := strconv.Atoi(value); err == nil {
				vmInfo.MemoryMB = memory
			}
		case "CPUs":
			if cpus, err := strconv.Atoi(value); err == nil {
				vmInfo.CPUs = cpus
			}
		}
	}

	return vmInfo, nil
}

// ListVMs lists all VMs
func (c *Client) ListVMs(ctx context.Context) ([]string, error) {
	args := []string{"list", "vms"}

	cmd := exec.CommandContext(ctx, c.config.VBoxManagePath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs: %w", err)
	}

	var vmNames []string
	lines := strings.Split(string(output), "\n")
	vmRegex := regexp.MustCompile(`"([^"]+)"`)

	for _, line := range lines {
		if line = strings.TrimSpace(line); line == "" {
			continue
		}

		matches := vmRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			vmNames = append(vmNames, matches[1])
		}
	}

	return vmNames, nil
}

// ValidateVirtualBox validates that VirtualBox is available and working
func (c *Client) ValidateVirtualBox(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, c.config.VBoxManagePath, "version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("VirtualBox is not available: %w", err)
	}

	c.logger.WithField("version", strings.TrimSpace(string(output))).Info("VirtualBox is available")
	return nil
}

// GetVMStats gets resource usage statistics for a VM
func (c *Client) GetVMStats(ctx context.Context, vmName string) (*types.VMStatus, error) {
	args := []string{"metrics", "collect", vmName}

	cmd := exec.CommandContext(ctx, c.config.VBoxManagePath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get VM stats: %w", err)
	}

	status := &types.VMStatus{}

	// Parse metrics output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "CPU/Usage/Total":
			if usage, err := strconv.ParseFloat(value, 64); err == nil {
				status.CPUUsage = usage
			}
		case "RAM/Usage/Total":
			if usage, err := strconv.ParseFloat(value, 64); err == nil {
				status.MemoryUsage = usage
			}
		case "VBox/VM/Uptime":
			if uptime, err := strconv.ParseInt(value, 10, 64); err == nil {
				status.Uptime = time.Duration(uptime) * time.Second
			}
		}
	}

	// Get current state
	state, err := c.GetVMState(ctx, vmName)
	if err == nil {
		status.State = state
	}

	return status, nil
}
