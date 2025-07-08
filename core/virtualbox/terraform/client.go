package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/config"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

// Client represents a Terraform client for VirtualBox VM provisioning
type Client struct {
	config *config.ServiceConfig
	logger *logrus.Logger
}

// NewClient creates a new Terraform client
func NewClient(cfg *config.ServiceConfig, logger *logrus.Logger) *Client {
	return &Client{
		config: cfg,
		logger: logger,
	}
}

// CreateVM creates a VM using Terraform
func (c *Client) CreateVM(ctx context.Context, vmID string, tfConfig *types.TerraformConfig) (*types.TerraformOutput, error) {
	// Create work directory for this VM
	workDir := filepath.Join(c.config.GetTerraformWorkDir(), vmID)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	// Generate Terraform configuration files
	if err := c.generateTerraformFiles(workDir, tfConfig); err != nil {
		return nil, fmt.Errorf("failed to generate Terraform files: %w", err)
	}

	// Initialize Terraform
	if err := c.terraformInit(ctx, workDir); err != nil {
		return nil, fmt.Errorf("failed to initialize Terraform: %w", err)
	}

	// Apply Terraform configuration
	if err := c.terraformApply(ctx, workDir, tfConfig); err != nil {
		return nil, fmt.Errorf("failed to apply Terraform configuration: %w", err)
	}

	// Get Terraform output
	output, err := c.terraformOutput(ctx, workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get Terraform output: %w", err)
	}

	c.logger.WithField("vmID", vmID).Info("Successfully created VM with Terraform")
	return output, nil
}

// DestroyVM destroys a VM using Terraform
func (c *Client) DestroyVM(ctx context.Context, vmID string) error {
	workDir := filepath.Join(c.config.GetTerraformWorkDir(), vmID)

	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		c.logger.WithField("vmID", vmID).Warn("Terraform work directory does not exist")
		return nil
	}

	// Destroy Terraform resources
	if err := c.terraformDestroy(ctx, workDir); err != nil {
		return fmt.Errorf("failed to destroy Terraform resources: %w", err)
	}

	// Clean up work directory
	if err := os.RemoveAll(workDir); err != nil {
		c.logger.WithError(err).WithField("vmID", vmID).Warn("Failed to remove Terraform work directory")
	}

	c.logger.WithField("vmID", vmID).Info("Successfully destroyed VM with Terraform")
	return nil
}

// generateTerraformFiles generates the necessary Terraform configuration files
func (c *Client) generateTerraformFiles(workDir string, tfConfig *types.TerraformConfig) error {
	// Generate main.tf
	mainTf := c.generateMainTf(tfConfig)
	mainTfPath := filepath.Join(workDir, "main.tf")
	if err := os.WriteFile(mainTfPath, []byte(mainTf), 0644); err != nil {
		return fmt.Errorf("failed to write main.tf: %w", err)
	}

	// Generate variables.tf
	variablesTf := c.generateVariablesTf()
	variablesTfPath := filepath.Join(workDir, "variables.tf")
	if err := os.WriteFile(variablesTfPath, []byte(variablesTf), 0644); err != nil {
		return fmt.Errorf("failed to write variables.tf: %w", err)
	}

	// Generate outputs.tf
	outputsTf := c.generateOutputsTf()
	outputsTfPath := filepath.Join(workDir, "outputs.tf")
	if err := os.WriteFile(outputsTfPath, []byte(outputsTf), 0644); err != nil {
		return fmt.Errorf("failed to write outputs.tf: %w", err)
	}

	// Generate terraform.tfvars
	tfvars := c.generateTfvars(tfConfig)
	tfvarsPath := filepath.Join(workDir, "terraform.tfvars")
	if err := os.WriteFile(tfvarsPath, []byte(tfvars), 0644); err != nil {
		return fmt.Errorf("failed to write terraform.tfvars: %w", err)
	}

	return nil
}

// generateMainTf generates the main Terraform configuration
func (c *Client) generateMainTf(tfConfig *types.TerraformConfig) string {
	networkConfig := ""
	if tfConfig.NetworkType == "bridged" {
		networkConfig = fmt.Sprintf(`
  network_interface {
    type = "bridged"
    bridge_adapter = "%s"
  }`, tfConfig.BridgeName)
	} else {
		networkConfig = `
  network_interface {
    type = "nat"
  }`
	}

	return fmt.Sprintf(`terraform {
  required_providers {
    virtualbox = {
      source = "terra-farm/virtualbox"
      version = "~> 0.2.2-alpha.1"
    }
  }
}

resource "virtualbox_vm" "vm" {
  name   = var.vm_name
  image  = var.base_image_path
  cpus   = var.cpus
  memory = var.memory_mb

  network_adapter {
    type           = "%s"
    host_interface = "%s"
  }

  disk {
    file = "${var.vm_name}.vdi"
    size = var.disk_size_gb * 1024
  }

  %s
}`, tfConfig.NetworkType, tfConfig.BridgeName, networkConfig)
}

// generateVariablesTf generates the variables configuration
func (c *Client) generateVariablesTf() string {
	return `variable "vm_name" {
  description = "Name of the VM"
  type        = string
}

variable "os_type" {
  description = "Operating system type"
  type        = string
  default     = "Linux_64"
}

variable "memory_mb" {
  description = "Memory in MB"
  type        = number
  default     = 2048
}

variable "cpus" {
  description = "Number of CPUs"
  type        = number
  default     = 2
}

variable "disk_size_gb" {
  description = "Disk size in GB"
  type        = number
  default     = 20
}

variable "network_type" {
  description = "Network type"
  type        = string
  default     = "bridged"
}

variable "bridge_name" {
  description = "Bridge adapter name"
  type        = string
  default     = ""
}

variable "base_image_path" {
  description = "Path to base image"
  type        = string
}`
}

// generateOutputsTf generates the outputs configuration
func (c *Client) generateOutputsTf() string {
	return `output "vm_uuid" {
  description = "VM UUID"
  value       = virtualbox_vm.vm.id
}

output "vm_name" {
  description = "VM name"
  value       = virtualbox_vm.vm.name
}

output "state" {
  description = "VM state"
  value       = virtualbox_vm.vm.state
}

output "ip_address" {
  description = "VM IP address"
  value       = virtualbox_vm.vm.network_adapter[0].ipv4_address
}

output "ssh_port" {
  description = "SSH port"
  value       = 22
}

output "vrde_port" {
  description = "VRDE port"
  value       = 3389
}

output "disk_path" {
  description = "Disk file path"
  value       = virtualbox_vm.vm.disk[0].file
}

output "network_config" {
  description = "Network configuration"
  value       = jsonencode(virtualbox_vm.vm.network_adapter)
}`
}

// generateTfvars generates the terraform.tfvars file
func (c *Client) generateTfvars(tfConfig *types.TerraformConfig) string {
	return fmt.Sprintf(`vm_name = "%s"
os_type = "%s"
memory_mb = %d
cpus = %d
disk_size_gb = %d
network_type = "%s"
bridge_name = "%s"
base_image_path = "%s"`,
		tfConfig.VMName,
		tfConfig.OSType,
		tfConfig.MemoryMB,
		tfConfig.CPUs,
		tfConfig.DiskSizeGB,
		tfConfig.NetworkType,
		tfConfig.BridgeName,
		tfConfig.BaseImagePath)
}

// terraformInit initializes Terraform
func (c *Client) terraformInit(ctx context.Context, workDir string) error {
	cmd := exec.CommandContext(ctx, c.config.GetTerraformPath(), "init")
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	c.logger.WithField("workDir", workDir).Debug("Initializing Terraform")
	return cmd.Run()
}

// terraformApply applies Terraform configuration
func (c *Client) terraformApply(ctx context.Context, workDir string, tfConfig *types.TerraformConfig) error {
	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, c.config.TerraformTimeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, c.config.GetTerraformPath(), "apply", "-auto-approve")
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment variables for custom variables
	if len(tfConfig.CustomVars) > 0 {
		for key, value := range tfConfig.CustomVars {
			cmd.Env = append(cmd.Env, fmt.Sprintf("TF_VAR_%s=%s", key, value))
		}
	}

	c.logger.WithField("workDir", workDir).Debug("Applying Terraform configuration")
	return cmd.Run()
}

// terraformDestroy destroys Terraform resources
func (c *Client) terraformDestroy(ctx context.Context, workDir string) error {
	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, c.config.TerraformTimeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, c.config.GetTerraformPath(), "destroy", "-auto-approve")
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	c.logger.WithField("workDir", workDir).Debug("Destroying Terraform resources")
	return cmd.Run()
}

// terraformOutput gets Terraform output
func (c *Client) terraformOutput(ctx context.Context, workDir string) (*types.TerraformOutput, error) {
	cmd := exec.CommandContext(ctx, c.config.GetTerraformPath(), "output", "-json")
	cmd.Dir = workDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get Terraform output: %w", err)
	}

	// Parse JSON output
	var tfOutput map[string]interface{}
	if err := json.Unmarshal(output, &tfOutput); err != nil {
		return nil, fmt.Errorf("failed to parse Terraform output: %w", err)
	}

	// Extract values
	result := &types.TerraformOutput{}

	if vmUUID, ok := tfOutput["vm_uuid"].(map[string]interface{}); ok {
		if value, ok := vmUUID["value"].(string); ok {
			result.VMUUID = value
		}
	}

	if vmName, ok := tfOutput["vm_name"].(map[string]interface{}); ok {
		if value, ok := vmName["value"].(string); ok {
			result.VMName = value
		}
	}

	if state, ok := tfOutput["state"].(map[string]interface{}); ok {
		if value, ok := state["value"].(string); ok {
			result.State = value
		}
	}

	if ipAddress, ok := tfOutput["ip_address"].(map[string]interface{}); ok {
		if value, ok := ipAddress["value"].(string); ok {
			result.IPAddress = value
		}
	}

	if sshPort, ok := tfOutput["ssh_port"].(map[string]interface{}); ok {
		if value, ok := sshPort["value"].(float64); ok {
			result.SSHPort = int(value)
		}
	}

	if vrdePort, ok := tfOutput["vrde_port"].(map[string]interface{}); ok {
		if value, ok := vrdePort["value"].(float64); ok {
			result.VRDEPort = int(value)
		}
	}

	if diskPath, ok := tfOutput["disk_path"].(map[string]interface{}); ok {
		if value, ok := diskPath["value"].(string); ok {
			result.DiskPath = value
		}
	}

	if networkConfig, ok := tfOutput["network_config"].(map[string]interface{}); ok {
		if value, ok := networkConfig["value"].(string); ok {
			result.NetworkConfig = value
		}
	}

	return result, nil
}

// ValidateTerraform validates that Terraform is available and working
func (c *Client) ValidateTerraform(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, c.config.GetTerraformPath(), "version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Terraform is not available: %w", err)
	}

	c.logger.WithField("version", strings.TrimSpace(string(output))).Info("Terraform is available")
	return nil
}
