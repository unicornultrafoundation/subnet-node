//go:build libvirt
// +build libvirt

package libvirt

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
)

// CloudInitManager handles cloud-init operations
type CloudInitManager struct {
	logger *logrus.Entry
}

// NewCloudInitManager creates a new cloud-init manager
func NewCloudInitManager(logger *logrus.Entry) *CloudInitManager {
	return &CloudInitManager{
		logger: logger.WithField("component", "cloudinit-manager"),
	}
}

// CloudInitConfig represents cloud-init configuration
type CloudInitConfig struct {
	Hostname      string
	Username      string
	Password      string
	SSHKey        string
	NetworkConfig string
	UserData      string
	MetaData      string
}

// CreateCloudInitISO creates a cloud-init ISO file
func (cim *CloudInitManager) CreateCloudInitISO(isoPath string, config *CloudInitConfig) error {
	// Create temporary directory for cloud-init files
	tempDir, err := os.MkdirTemp("", "cloud-init-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create user-data file
	userDataPath := filepath.Join(tempDir, "user-data")
	if err := cim.createUserDataFile(userDataPath, config); err != nil {
		return fmt.Errorf("failed to create user-data: %w", err)
	}

	// Create meta-data file
	metaDataPath := filepath.Join(tempDir, "meta-data")
	if err := cim.createMetaDataFile(metaDataPath, config); err != nil {
		return fmt.Errorf("failed to create meta-data: %w", err)
	}

	// Create network-config file if provided
	if config.NetworkConfig != "" {
		networkConfigPath := filepath.Join(tempDir, "network-config")
		if err := os.WriteFile(networkConfigPath, []byte(config.NetworkConfig), 0644); err != nil {
			return fmt.Errorf("failed to create network-config: %w", err)
		}
	}

	// Create ISO using genisoimage or mkisofs
	if err := cim.createISO(tempDir, isoPath); err != nil {
		return fmt.Errorf("failed to create ISO: %w", err)
	}

	cim.logger.WithField("iso_path", isoPath).Info("Cloud-init ISO created successfully")
	return nil
}

// createUserDataFile creates the user-data file for cloud-init
func (cim *CloudInitManager) createUserDataFile(path string, config *CloudInitConfig) error {
	userDataTemplate := `#cloud-config
hostname: {{.Hostname}}
manage_etc_hosts: true

users:
  - name: {{.Username}}
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
    {{- if .Password}}
    lock_passwd: false
    passwd: {{.Password}}
    {{- else}}
    lock_passwd: true
    {{- end}}
    {{- if .SSHKey}}
    ssh_authorized_keys:
      - {{.SSHKey}}
    {{- end}}

{{- if .UserData}}
{{.UserData}}
{{- end}}

package_update: true
package_upgrade: true

packages:
  - qemu-guest-agent
  - cloud-init

runcmd:
  - systemctl enable qemu-guest-agent
  - systemctl start qemu-guest-agent
  - cloud-init clean
  - cloud-init init
`

	tmpl, err := template.New("user-data").Parse(userDataTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse user-data template: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create user-data file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, config); err != nil {
		return fmt.Errorf("failed to execute user-data template: %w", err)
	}

	return nil
}

// createMetaDataFile creates the meta-data file for cloud-init
func (cim *CloudInitManager) createMetaDataFile(path string, config *CloudInitConfig) error {
	metaDataTemplate := `instance-id: {{.Hostname}}
local-hostname: {{.Hostname}}
{{- if .MetaData}}
{{.MetaData}}
{{- end}}
`

	tmpl, err := template.New("meta-data").Parse(metaDataTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse meta-data template: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create meta-data file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, config); err != nil {
		return fmt.Errorf("failed to execute meta-data template: %w", err)
	}

	return nil
}

// createISO creates an ISO file from the cloud-init files
func (cim *CloudInitManager) createISO(sourceDir, isoPath string) error {
	// Try genisoimage first
	if _, err := exec.LookPath("genisoimage"); err == nil {
		cmd := exec.Command("genisoimage", "-output", isoPath, "-volid", "cidata", "-joliet", "-rock", sourceDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("genisoimage failed: %s, %w", string(output), err)
		}
		return nil
	}

	// Try mkisofs as fallback
	if _, err := exec.LookPath("mkisofs"); err == nil {
		cmd := exec.Command("mkisofs", "-output", isoPath, "-volid", "cidata", "-joliet", "-rock", sourceDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("mkisofs failed: %s, %w", string(output), err)
		}
		return nil
	}

	return fmt.Errorf("neither genisoimage nor mkisofs found")
}

// GenerateSSHKey generates a new SSH key pair
func (cim *CloudInitManager) GenerateSSHKey(keyPath string) (string, error) {
	// Check if key already exists
	if _, err := os.Stat(keyPath); err == nil {
		// Key exists, read public key
		publicKeyPath := keyPath + ".pub"
		if _, err := os.Stat(publicKeyPath); err == nil {
			publicKey, err := os.ReadFile(publicKeyPath)
			if err != nil {
				return "", fmt.Errorf("failed to read existing public key: %w", err)
			}
			return strings.TrimSpace(string(publicKey)), nil
		}
	}

	// Generate new key pair
	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "4096", "-f", keyPath, "-N", "")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to generate SSH key: %s, %w", string(output), err)
	}

	// Read public key
	publicKeyPath := keyPath + ".pub"
	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read generated public key: %w", err)
	}

	cim.logger.WithField("key_path", keyPath).Info("SSH key pair generated")
	return strings.TrimSpace(string(publicKey)), nil
}

// CreateDefaultCloudInitConfig creates a default cloud-init configuration
func (cim *CloudInitManager) CreateDefaultCloudInitConfig(hostname, username string) *CloudInitConfig {
	return &CloudInitConfig{
		Hostname: hostname,
		Username: username,
		Password: "$6$rounds=656000$salt$hashedpassword", // Default hashed password
		UserData: `# Additional user data can be added here
final_message: "Cloud-init completed successfully"
`,
		MetaData: `# Additional metadata can be added here
`,
	}
}
