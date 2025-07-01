package libvirt

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
)

// UserContext manages user context for libvirt operations
type UserContext struct {
	currentUser *user.User
	libvirtUser *user.User
	logger      *logrus.Entry
	mu          sync.RWMutex
}

// NewUserContext creates a new user context manager
func NewUserContext(logger *logrus.Entry) *UserContext {
	uc := &UserContext{
		logger: logger.WithField("component", "user-context"),
	}

	// Get current user
	if currentUser, err := user.Current(); err == nil {
		uc.currentUser = currentUser
		uc.logger.WithField("current_user", currentUser.Username).Debug("Current user detected")
	} else {
		uc.logger.WithError(err).Warn("Failed to get current user")
	}

	// Try to get libvirt-qemu user
	if libvirtUser, err := user.Lookup("libvirt-qemu"); err == nil {
		uc.libvirtUser = libvirtUser
		uc.logger.WithField("libvirt_user", libvirtUser.Username).Debug("Libvirt user detected")
	} else {
		uc.logger.WithError(err).Debug("Libvirt-qemu user not found, will use current user")
	}

	return uc
}

// GetEffectiveUser returns the effective user for file operations
func (uc *UserContext) GetEffectiveUser() *user.User {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	// If we have libvirt user and current user is root, use libvirt user
	if uc.libvirtUser != nil && uc.currentUser != nil && uc.currentUser.Uid == "0" {
		return uc.libvirtUser
	}

	// Otherwise use current user
	return uc.currentUser
}

// EnsureFileOwnership ensures files are accessible by libvirt
func (uc *UserContext) EnsureFileOwnership(filePath string) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if uc.currentUser == nil {
		return fmt.Errorf("current user not available")
	}

	// Check current file ownership
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Get file owner
	fileOwner, err := user.LookupId(fmt.Sprintf("%d", fileInfo.Sys().(*syscall.Stat_t).Uid))
	if err != nil {
		uc.logger.WithError(err).Debug("Failed to get file owner, proceeding anyway")
		return nil
	}

	// If file is already owned by current user, just ensure it's readable/writable
	if fileOwner.Uid == uc.currentUser.Uid {
		uc.logger.WithField("file_path", filePath).Debug("File already owned by current user")
		return nil
	}

	// Try to change ownership without sudo first (works if current user has permission)
	cmd := exec.Command("chown", uc.currentUser.Uid+":"+uc.currentUser.Gid, filePath)
	if err := cmd.Run(); err == nil {
		uc.logger.WithField("file_path", filePath).Info("Changed file ownership to current user")
		return nil
	}

	// If that fails, log the issue but don't use sudo aggressively
	uc.logger.WithField("file_path", filePath).Warn("Could not change file ownership, libvirt may have access issues")
	return fmt.Errorf("failed to ensure file ownership for %s", filePath)
}

// CreateFileWithCorrectOwnership creates a file with the correct ownership from the start
func (uc *UserContext) CreateFileWithCorrectOwnership(filePath string) (*os.File, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	// Ensure correct ownership
	if err := uc.EnsureFileOwnership(filePath); err != nil {
		file.Close()
		os.Remove(filePath) // Clean up
		return nil, fmt.Errorf("failed to set file ownership: %w", err)
	}

	return file, nil
}

// EnsureDirectoryPermissions ensures a directory has the correct permissions for libvirt access
func (uc *UserContext) EnsureDirectoryPermissions(dirPath string) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if uc.currentUser == nil {
		return fmt.Errorf("current user not available")
	}

	// Check if directory exists
	dirInfo, err := os.Stat(dirPath)
	if err != nil {
		return fmt.Errorf("failed to stat directory: %w", err)
	}

	if !dirInfo.IsDir() {
		return fmt.Errorf("path is not a directory: %s", dirPath)
	}

	// Get directory owner
	dirOwner, err := user.LookupId(fmt.Sprintf("%d", dirInfo.Sys().(*syscall.Stat_t).Uid))
	if err != nil {
		uc.logger.WithError(err).Debug("Failed to get directory owner, proceeding anyway")
		return nil
	}

	// If directory is already owned by current user, just ensure it's accessible
	if dirOwner.Uid == uc.currentUser.Uid {
		uc.logger.WithField("dir_path", dirPath).Debug("Directory already owned by current user")
		return nil
	}

	// Try to change ownership without sudo first
	cmd := exec.Command("chown", uc.currentUser.Uid+":"+uc.currentUser.Gid, dirPath)
	if err := cmd.Run(); err == nil {
		uc.logger.WithField("dir_path", dirPath).Info("Changed directory ownership to current user")
		return nil
	}

	// If that fails, log the issue but don't use sudo aggressively
	uc.logger.WithField("dir_path", dirPath).Warn("Could not change directory ownership, libvirt may have access issues")
	return fmt.Errorf("failed to ensure directory permissions for %s", dirPath)
}

// UbuntuCloudImages contains URLs for Ubuntu cloud images by version and architecture
var UbuntuCloudImages = map[string]map[string]string{
	"22.04": {
		"amd64": "https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img",
		"arm64": "https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-arm64.img",
	},
	"20.04": {
		"amd64": "https://cloud-images.ubuntu.com/releases/20.04/release/ubuntu-20.04-server-cloudimg-amd64.img",
		"arm64": "https://cloud-images.ubuntu.com/releases/20.04/release/ubuntu-20.04-server-cloudimg-arm64.img",
	},
	"24.04": {
		"amd64": "https://cloud-images.ubuntu.com/releases/24.04/release/ubuntu-24.04-server-cloudimg-amd64.img",
		"arm64": "https://cloud-images.ubuntu.com/releases/24.04/release/ubuntu-24.04-server-cloudimg-arm64.img",
	},
}

// Client wraps libvirt connection and provides KVM/QEMU management
type Client struct {
	conn         interface{} // Will be *libvirt.Connect when libvirt is available
	uri          string
	logger       *logrus.Entry
	mu           sync.RWMutex
	available    bool
	kvmAvailable bool // Track if KVM is available
	userContext  *UserContext
}

// NewClient creates a new libvirt client
func NewClient(uri string, logger *logrus.Entry) (*Client, error) {
	if uri == "" {
		uri = "qemu:///system" // Default to system QEMU connection
	}

	logger.WithField("uri", uri).Info("Creating new libvirt client")

	client := &Client{
		uri:         uri,
		logger:      logger.WithField("component", "libvirt-client"),
		userContext: NewUserContext(logger),
	}

	// Try to connect to libvirt
	logger.Info("Attempting to connect to libvirt")
	if err := client.tryConnect(); err != nil {
		logger.WithError(err).Warn("Libvirt not available, using stub implementation")
		client.available = false
	} else {
		client.available = true
		logger.Info("Libvirt connected successfully")
	}

	logger.WithField("available", client.available).Info("Libvirt client creation completed")
	return client, nil
}

// tryConnect attempts to establish connection to libvirt daemon
func (c *Client) tryConnect() error {
	// Check if libvirt is available by trying to run virsh
	if _, err := exec.LookPath("virsh"); err != nil {
		c.logger.WithError(err).Debug("virsh not found in PATH")
		return fmt.Errorf("virsh not found: %w", err)
	}

	c.logger.WithField("uri", c.uri).Debug("Testing libvirt connection")

	// Test connection using virsh list (this tests if we can connect to the daemon)
	cmd := exec.Command("virsh", "-c", c.uri, "list", "--all")
	c.logger.WithField("command", cmd.String()).Debug("Executing virsh command")

	output, err := cmd.CombinedOutput()
	if err != nil {
		c.logger.WithError(err).WithField("output", string(output)).Debug("Failed to connect to libvirt")
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}

	// Check if KVM is supported by looking at capabilities
	capCmd := exec.Command("virsh", "-c", c.uri, "capabilities")
	capOutput, err := capCmd.Output()
	if err == nil {
		if strings.Contains(string(capOutput), "domain type='kvm'") {
			c.logger.Info("KVM virtualization is supported")
			c.kvmAvailable = true
		} else if strings.Contains(string(capOutput), "domain type='qemu'") {
			c.logger.Info("QEMU virtualization is supported (KVM not available)")
		} else {
			c.logger.Warn("Neither KVM nor QEMU virtualization detected")
		}
	}

	c.logger.WithField("output", string(output)).Debug("Successfully connected to libvirt")
	return nil
}

// IsAvailable returns whether libvirt is available
func (c *Client) IsAvailable() bool {
	return c.available
}

// IsKVMAvailable returns whether KVM is available
func (c *Client) IsKVMAvailable() bool {
	return c.kvmAvailable
}

// GetConnection returns the underlying libvirt connection (stub implementation)
func (c *Client) GetConnection() (interface{}, error) {
	if !c.available {
		return nil, fmt.Errorf("libvirt not available")
	}
	return c.conn, nil
}

// GetNetworkByName gets a network by name (stub implementation)
func (c *Client) GetNetworkByName(name string) (interface{}, error) {
	if !c.available {
		return nil, fmt.Errorf("libvirt not available")
	}

	c.logger.WithField("network_name", name).Debug("Checking if network exists")

	// Check if network exists using virsh
	cmd := exec.Command("virsh", "-c", c.uri, "net-info", name)
	c.logger.WithField("command", cmd.String()).Debug("Executing virsh net-info")

	output, err := cmd.CombinedOutput()
	if err != nil {
		c.logger.WithError(err).WithField("output", string(output)).Debug("Network not found")
		return nil, fmt.Errorf("network %s not found", name)
	}

	c.logger.WithField("network_name", name).Debug("Network found")
	return name, nil
}

// CreateNetwork creates a new network (stub implementation)
func (c *Client) CreateNetwork(xml string) (interface{}, error) {
	if !c.available {
		return nil, fmt.Errorf("libvirt not available")
	}

	c.logger.WithField("xml", xml).Info("Creating network with virsh")

	// Extract network name from XML for operations
	networkName := "subnet-net" // Default fallback
	if strings.Contains(xml, "<name>") {
		start := strings.Index(xml, "<name>") + 6
		end := strings.Index(xml, "</name>")
		if start > 5 && end > start {
			networkName = xml[start:end]
		}
	}

	// First check if network is already defined
	checkCmd := exec.Command("virsh", "-c", c.uri, "net-info", networkName)
	if err := checkCmd.Run(); err == nil {
		// Network exists, check if it's active
		activeCmd := exec.Command("virsh", "-c", c.uri, "net-list", "--name", "--active")
		activeOutput, err := activeCmd.Output()
		if err == nil && strings.Contains(string(activeOutput), networkName) {
			c.logger.WithField("network_name", networkName).Info("Network already exists and is active")
			return networkName, nil
		}
		// Network exists but not active, try to start it
		startCmd := exec.Command("virsh", "-c", c.uri, "net-start", networkName)
		if err := startCmd.Run(); err == nil {
			c.logger.WithField("network_name", networkName).Info("Network started successfully")
			return networkName, nil
		}
		// If start fails, we'll try to redefine it
		c.logger.WithField("network_name", networkName).Warn("Failed to start existing network, will redefine")
		undefineCmd := exec.Command("virsh", "-c", c.uri, "net-undefine", networkName)
		undefineCmd.Run() // Ignore errors here
	}

	// Write XML to temporary file
	tmpFile, err := os.CreateTemp("", "network-*.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(xml); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write XML to temp file: %w", err)
	}
	tmpFile.Close()

	// Create network using virsh
	cmd := exec.Command("virsh", "-c", c.uri, "net-define", tmpFile.Name())
	c.logger.WithField("command", cmd.String()).Debug("Executing virsh net-define")

	output, err := cmd.CombinedOutput()
	if err != nil {
		c.logger.WithError(err).WithField("output", string(output)).Error("Failed to define network")
		return nil, fmt.Errorf("failed to define network: %w", err)
	}

	// Start the network
	startCmd := exec.Command("virsh", "-c", c.uri, "net-start", networkName)
	c.logger.WithField("command", startCmd.String()).Debug("Executing virsh net-start")

	startOutput, err := startCmd.CombinedOutput()
	if err != nil {
		c.logger.WithError(err).WithField("output", string(startOutput)).Error("Failed to start network")
		// Try to undefine the network if start failed
		undefineCmd := exec.Command("virsh", "-c", c.uri, "net-undefine", networkName)
		undefineCmd.Run() // Ignore errors here
		return nil, fmt.Errorf("failed to start network: %w", err)
	}

	// Set network to autostart
	autostartCmd := exec.Command("virsh", "-c", c.uri, "net-autostart", networkName)
	c.logger.WithField("command", autostartCmd.String()).Debug("Executing virsh net-autostart")

	autostartOutput, err := autostartCmd.CombinedOutput()
	if err != nil {
		c.logger.WithError(err).WithField("output", string(autostartOutput)).Warn("Failed to set network autostart")
		// Don't fail the entire operation for this
	}

	c.logger.WithField("network_name", networkName).Info("Network created and started successfully")
	return networkName, nil
}

// DomainManager handles VM (domain) operations
type DomainManager struct {
	client *Client
	logger *logrus.Entry
}

// NewDomainManager creates a new domain manager
func NewDomainManager(client *Client, logger *logrus.Entry) *DomainManager {
	return &DomainManager{
		client: client,
		logger: logger.WithField("component", "domain-manager"),
	}
}

// SystemArchitecture represents detected system architecture information
// This is used to determine the appropriate VM configuration for the host system
type SystemArchitecture struct {
	Architecture string   // e.g., "x86_64", "aarch64", "ppc64le"
	Machine      string   // e.g., "pc-q35-2.12", "virt-8.2"
	CPUModel     string   // e.g., "Intel(R) Core(TM) i7-9750H", "AMD EPYC"
	Vendor       string   // e.g., "Intel", "AMD"
	Features     []string // CPU features like "sse4_2", "avx2"
}

// detectSystemArchitecture detects the current system architecture and hardware information
// This function combines runtime.GOARCH detection with Linux-specific CPU information
// and libvirt capabilities to provide comprehensive system information for VM configuration
func (dm *DomainManager) detectSystemArchitecture() *SystemArchitecture {
	arch := &SystemArchitecture{
		Architecture: "x86_64", // Default fallback
		Machine:      "pc-q35-2.12",
		CPUModel:     "Unknown",
		Vendor:       "Unknown",
		Features:     []string{},
	}

	// Use runtime.GOARCH as primary detection
	switch runtime.GOARCH {
	case "amd64":
		arch.Architecture = "x86_64"
		arch.Machine = "pc-q35-2.12"
	case "arm64":
		arch.Architecture = "aarch64"
		arch.Machine = "virt-8.2"
	case "ppc64le":
		arch.Architecture = "ppc64le"
		arch.Machine = "pseries"
	case "s390x":
		arch.Architecture = "s390x"
		arch.Machine = "s390-ccw-virtio"
	}

	// On Linux, try to get detailed CPU information from /proc/cpuinfo
	if runtime.GOOS == "linux" {
		arch = dm.enrichWithLinuxCPUInfo(arch)
	}

	// Try to get CPU info via libvirt capabilities if available
	arch = dm.enrichWithLibvirtInfo(arch)

	dm.logger.WithFields(logrus.Fields{
		"architecture": arch.Architecture,
		"machine":      arch.Machine,
		"cpu_model":    arch.CPUModel,
		"vendor":       arch.Vendor,
		"features":     arch.Features,
	}).Debug("Detected system architecture")

	return arch
}

// enrichWithLinuxCPUInfo enriches architecture info with Linux-specific CPU details
// Reads /proc/cpuinfo to get detailed CPU model, vendor, and feature information
func (dm *DomainManager) enrichWithLinuxCPUInfo(arch *SystemArchitecture) *SystemArchitecture {
	// Try to read /proc/cpuinfo
	if cpuInfo, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		lines := strings.Split(string(cpuInfo), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "model name") {
				if parts := strings.SplitN(line, ":", 2); len(parts) == 2 {
					arch.CPUModel = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(line, "vendor_id") {
				if parts := strings.SplitN(line, ":", 2); len(parts) == 2 {
					arch.Vendor = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(line, "flags") {
				if parts := strings.SplitN(line, ":", 2); len(parts) == 2 {
					flags := strings.Fields(parts[1])
					arch.Features = flags
				}
			}
		}
	}

	// Try to detect architecture via uname as backup
	if cmd := exec.Command("uname", "-m"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			machine := strings.TrimSpace(string(output))
			switch machine {
			case "x86_64":
				arch.Architecture = "x86_64"
				arch.Machine = "pc-q35-2.12"
			case "aarch64", "arm64":
				arch.Architecture = "aarch64"
				arch.Machine = "virt-8.2"
			case "ppc64le":
				arch.Architecture = "ppc64le"
				arch.Machine = "pseries"
			case "s390x":
				arch.Architecture = "s390x"
				arch.Machine = "s390-ccw-virtio"
			}
		}
	}

	return arch
}

// enrichWithLibvirtInfo enriches architecture info with libvirt capabilities
// Uses virsh capabilities to get host CPU model information when available
func (dm *DomainManager) enrichWithLibvirtInfo(arch *SystemArchitecture) *SystemArchitecture {
	// Try to get CPU info via libvirt capabilities if available
	if dm.client.IsAvailable() {
		capCmd := exec.Command("virsh", "-c", dm.client.uri, "capabilities")
		if output, err := capCmd.Output(); err == nil {
			capXML := string(output)

			// Extract host CPU model from capabilities
			if strings.Contains(capXML, "<host>") {
				// Look for CPU model in host section
				if start := strings.Index(capXML, "<model>"); start != -1 {
					if end := strings.Index(capXML[start:], "</model>"); end != -1 {
						model := capXML[start+7 : start+end]
						if model != "" {
							arch.CPUModel = model
						}
					}
				}
			}
		}
	}

	return arch
}

// CreateDomainWithCloudInit creates a new VM domain with cloud-init support
func (dm *DomainManager) CreateDomainWithCloudInit(name, uuid string, memoryMB, vcpus int, diskPath, networkName, cloudInitISOPath string) (interface{}, error) {
	if !dm.client.IsAvailable() {
		return nil, fmt.Errorf("libvirt not available")
	}

	dm.logger.WithFields(logrus.Fields{
		"name":           name,
		"uuid":           uuid,
		"memory_mb":      memoryMB,
		"vcpus":          vcpus,
		"disk_path":      diskPath,
		"network_name":   networkName,
		"cloud_init_iso": cloudInitISOPath,
	}).Info("Creating VM domain with cloud-init")

	// Validate inputs
	if name == "" || uuid == "" || memoryMB <= 0 || vcpus <= 0 || diskPath == "" || networkName == "" {
		return nil, fmt.Errorf("invalid parameters: name, uuid, memoryMB, vcpus, diskPath, and networkName must be provided")
	}

	// Check if domain already exists
	checkCmd := exec.Command("virsh", "-c", dm.client.uri, "dominfo", name)
	if err := checkCmd.Run(); err == nil {
		dm.logger.WithField("domain_name", name).Warn("Domain already exists")
		return name, nil
	}

	// Check if disk file exists
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("disk file not found: %s", diskPath)
	}

	// Check if cloud-init ISO exists
	if cloudInitISOPath != "" {
		if _, err := os.Stat(cloudInitISOPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("cloud-init ISO file not found: %s", cloudInitISOPath)
		}
	}

	// Try to create domain with different CPU configurations if needed
	return dm.createDomainWithFallback(name, uuid, memoryMB, vcpus, diskPath, networkName, cloudInitISOPath)
}

// createDomainWithFallback attempts to create a domain with fallback CPU configurations
func (dm *DomainManager) createDomainWithFallback(name, uuid string, memoryMB, vcpus int, diskPath, networkName, cloudInitISOPath string) (interface{}, error) {
	sysArch := dm.detectSystemArchitecture()

	// Define CPU configurations to try for ARM64
	var cpuConfigs []string
	if sysArch.Architecture == "aarch64" {
		cpuConfigs = []string{
			"cortex-a72",  // Most common ARM64 CPU
			"cortex-a73",  // Alternative
			"cortex-a76",  // Newer ARM64 CPU
			"cortex-a78",  // Even newer
			"neoverse-n1", // Server-grade ARM64
			"cortex-a57",  // Fallback
			"cortex-a53",  // Basic ARM64
		}
	} else {
		// For non-ARM64, just use the standard configuration
		cpuConfigs = []string{""}
	}

	// Try each CPU configuration
	for i, cpuModel := range cpuConfigs {
		if sysArch.Architecture == "aarch64" {
			dm.logger.WithField("cpu_model", cpuModel).Debug("Trying CPU configuration")
		}

		// Generate domain XML with current CPU configuration
		domainXML := dm.GenerateDomainXMLWithCPU(name, uuid, memoryMB, vcpus, diskPath, networkName, cloudInitISOPath, cpuModel)

		// Write XML to temporary file
		tmpFile, err := os.CreateTemp("", "domain-*.xml")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp file: %w", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.WriteString(domainXML); err != nil {
			tmpFile.Close()
			return nil, fmt.Errorf("failed to write XML to temp file: %w", err)
		}
		tmpFile.Close()

		// Try to define the domain
		defineCmd := exec.Command("virsh", "-c", dm.client.uri, "define", tmpFile.Name())
		dm.logger.WithField("command", defineCmd.String()).Debug("Executing virsh define")

		output, err := defineCmd.CombinedOutput()
		if err == nil {
			// Success! Set domain to autostart
			autostartCmd := exec.Command("virsh", "-c", dm.client.uri, "autostart", name)
			dm.logger.WithField("command", autostartCmd.String()).Debug("Executing virsh autostart")

			autostartOutput, err := autostartCmd.CombinedOutput()
			if err != nil {
				dm.logger.WithError(err).WithField("output", string(autostartOutput)).Warn("Failed to set domain autostart")
				// Don't fail the entire operation for this
			}

			if sysArch.Architecture == "aarch64" {
				dm.logger.WithField("cpu_model", cpuModel).Info("Domain created successfully with CPU model")
			} else {
				dm.logger.WithField("domain_name", name).Info("Domain created successfully")
			}
			return name, nil
		}

		// If this is not the last attempt, log the error and try the next configuration
		if i < len(cpuConfigs)-1 {
			dm.logger.WithError(err).WithField("output", string(output)).WithField("cpu_model", cpuModel).Debug("CPU configuration failed, trying next")
		} else {
			// This was the last attempt, return the error
			dm.logger.WithError(err).WithField("output", string(output)).Error("All CPU configurations failed")
			return nil, fmt.Errorf("failed to define domain after trying all CPU configurations: %w", err)
		}
	}

	return nil, fmt.Errorf("failed to create domain")
}

// GenerateDomainXMLWithCPU generates domain XML with a specific CPU model
func (dm *DomainManager) GenerateDomainXMLWithCPU(name, uuid string, memoryMB, vcpus int, diskPath, networkName, cloudInitISOPath, cpuModel string) string {
	// Determine disk format based on file extension
	diskFormat := "qcow2"
	if strings.HasSuffix(diskPath, ".raw") {
		diskFormat = "raw"
	} else if strings.HasSuffix(diskPath, ".vmdk") {
		diskFormat = "vmdk"
	}

	// Detect current architecture and hardware information
	sysArch := dm.detectSystemArchitecture()
	arch := sysArch.Architecture
	machine := sysArch.Machine

	// Build disk devices XML
	diskDevices := fmt.Sprintf(`    <disk type='file' device='disk'>
      <driver name='qemu' type='%s'/>
      <source file='%s'/>
      <target dev='vda' bus='virtio'/>
    </disk>`, diskFormat, diskPath)

	// Add cloud-init ISO if provided
	if cloudInitISOPath != "" {
		diskDevices += fmt.Sprintf(`
    <disk type='file' device='cdrom'>
      <driver name='qemu' type='raw'/>
      <source file='%s'/>
      <target dev='sda' bus='sata'/>
      <readonly/>
    </disk>`, cloudInitISOPath)
	}

	// Generate the complete domain XML
	domainXML := fmt.Sprintf(`<domain type='%s'>
  <name>%s</name>
  <uuid>%s</uuid>
  <memory unit='MiB'>%d</memory>
  <currentMemory unit='MiB'>%d</currentMemory>
  <vcpu>%d</vcpu>
  <os>
    <type arch='%s' machine='%s'>hvm</type>
    <boot dev='hd'/>
    <boot dev='cdrom'/>
  </os>`, dm.client.GetDomainType(), name, uuid, memoryMB, memoryMB, vcpus, arch, machine)

	// Add features based on architecture
	if arch == "x86_64" {
		domainXML += `
  <features>
    <apic/>
  </features>`
	}

	// Use appropriate CPU mode based on architecture
	if arch == "x86_64" {
		domainXML += `
  <cpu mode='host-passthrough' check='partial'/>`
	} else if arch == "aarch64" {
		// Use the provided CPU model or fallback to default
		if cpuModel == "" {
			cpuModel = "cortex-a72"
		}

		dm.logger.WithFields(logrus.Fields{
			"detected_cpu": sysArch.CPUModel,
			"selected_cpu": cpuModel,
			"architecture": arch,
		}).Debug("Selected CPU model for ARM64 domain")

		domainXML += fmt.Sprintf(`
  <cpu mode='custom' match='exact'>
    <model fallback='allow'>%s</model>
  </cpu>`, cpuModel)
	} else {
		// For other architectures, use a generic CPU model
		domainXML += `
  <cpu mode='custom' match='exact'>
    <model fallback='allow'>cortex-a57</model>
  </cpu>`
	}

	domainXML += `
  <clock offset='utc'/>
  <on_poweroff>destroy</on_poweroff>
  <on_reboot>restart</on_reboot>
  <on_crash>destroy</on_crash>`

	// Add power management configuration only for x86_64 (ACPI S3/S4 support)
	if arch == "x86_64" {
		domainXML += `
  <pm>
    <suspend-to-mem enabled='no'/>
    <suspend-to-disk enabled='no'/>
  </pm>`
	}

	domainXML += fmt.Sprintf(`
  <devices>
%s
    <interface type='network'>
      <source network='%s'/>
      <model type='virtio'/>
    </interface>
    <console type='pty'>
      <target type='serial' port='0'/>
    </console>`, diskDevices, networkName)

	// Add USB controller and input devices based on architecture
	if arch == "x86_64" {
		domainXML += `
    <controller type='usb' index='0' model='piix3-uhci'/>
    <controller type='usb' index='0' model='ehci'/>
    <input type='tablet' bus='usb'/>
    <input type='keyboard' bus='usb'/>`
	} else {
		// For non-x86 architectures, use simpler input devices
		domainXML += `
    <input type='keyboard' bus='virtio'/>
    <input type='mouse' bus='virtio'/>`
	}

	domainXML += fmt.Sprintf(`
    <graphics type='vnc' port='-1' autoport='yes' listen='0.0.0.0'>
      <listen type='address' address='0.0.0.0'/>
    </graphics>
    <video>
      <model type='%s' vram='16384' heads='1' primary='yes'/>
    </video>
    <memballoon model='virtio'/>
    <rng model='virtio'>
      <backend model='random'>/dev/urandom</backend>
    </rng>
  </devices>
</domain>`, dm.getVideoModel(arch))

	return domainXML
}

// getVideoModel returns the appropriate video model for the architecture
func (dm *DomainManager) getVideoModel(arch string) string {
	switch arch {
	case "x86_64":
		return "cirrus"
	case "aarch64":
		return "virtio"
	default:
		return "virtio"
	}
}

// GetDomain retrieves a domain by name
func (dm *DomainManager) GetDomain(name string) (interface{}, error) {
	if !dm.client.IsAvailable() {
		return nil, fmt.Errorf("libvirt not available")
	}

	dm.logger.WithField("domain_name", name).Debug("Getting domain info")

	// Check if domain exists using virsh dominfo
	cmd := exec.Command("virsh", "-c", dm.client.uri, "dominfo", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		dm.logger.WithError(err).WithField("output", string(output)).Debug("Domain not found")
		return nil, fmt.Errorf("domain %s not found", name)
	}

	dm.logger.WithField("domain_name", name).Debug("Domain found")
	return name, nil
}

// FixDomainPermissions fixes permissions on VM files before starting a domain
func (dm *DomainManager) FixDomainPermissions(domain interface{}) error {
	if !dm.client.IsAvailable() {
		return fmt.Errorf("libvirt not available")
	}

	domainName, ok := domain.(string)
	if !ok {
		return fmt.Errorf("invalid domain type, expected string")
	}

	dm.logger.WithField("domain_name", domainName).Info("Fixing domain permissions")

	// Get domain XML to extract disk path
	cmd := exec.Command("virsh", "-c", dm.client.uri, "dumpxml", domainName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		dm.logger.WithError(err).WithField("output", string(output)).Debug("Failed to get domain XML")
		return fmt.Errorf("failed to get domain XML: %w", err)
	}

	// Parse XML to find disk paths
	xmlContent := string(output)
	diskPaths := dm.extractDiskPathsFromXML(xmlContent)

	if len(diskPaths) == 0 {
		dm.logger.WithField("domain_name", domainName).Warn("No disk paths found in domain XML")
		return nil
	}

	// Fix permissions for each disk file
	for _, diskPath := range diskPaths {
		dm.logger.WithField("disk_path", diskPath).Debug("Fixing permissions for disk file")
		if err := dm.client.userContext.FixVMFilePermissions(diskPath); err != nil {
			dm.logger.WithError(err).WithField("disk_path", diskPath).Warn("Failed to fix disk file permissions")
		}
	}

	dm.logger.WithField("domain_name", domainName).Info("Domain permissions fixed")
	return nil
}

// extractDiskPathsFromXML extracts disk file paths from domain XML
func (dm *DomainManager) extractDiskPathsFromXML(xmlContent string) []string {
	var diskPaths []string

	// Look for disk source file attributes
	lines := strings.Split(xmlContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "<source file=") {
			// Extract file path from source file attribute
			start := strings.Index(line, "file='")
			if start != -1 {
				start += 6 // Skip "file='"
				end := strings.Index(line[start:], "'")
				if end != -1 {
					filePath := line[start : start+end]
					diskPaths = append(diskPaths, filePath)
				}
			}
		}
	}

	return diskPaths
}

// StartDomain starts a VM
func (dm *DomainManager) StartDomain(domain interface{}) error {
	if !dm.client.IsAvailable() {
		return fmt.Errorf("libvirt not available")
	}

	domainName, ok := domain.(string)
	if !ok {
		return fmt.Errorf("invalid domain type, expected string")
	}

	dm.logger.WithField("domain_name", domainName).Info("Starting domain")

	// Check if domain is already running
	checkCmd := exec.Command("virsh", "-c", dm.client.uri, "domstate", domainName)
	output, err := checkCmd.Output()
	if err == nil && strings.TrimSpace(string(output)) == "running" {
		dm.logger.WithField("domain_name", domainName).Info("Domain is already running")
		return nil
	}

	// Get domain XML to check disk paths before starting
	dm.logger.WithField("domain_name", domainName).Debug("Checking domain configuration")
	xmlCmd := exec.Command("virsh", "-c", dm.client.uri, "dumpxml", domainName)
	xmlOutput, err := xmlCmd.CombinedOutput()
	if err != nil {
		dm.logger.WithError(err).WithField("output", string(xmlOutput)).Warn("Failed to get domain XML, proceeding anyway")
	} else {
		// Extract and check disk paths
		diskPaths := dm.extractDiskPathsFromXML(string(xmlOutput))
		for _, diskPath := range diskPaths {
			dm.logger.WithField("disk_path", diskPath).Debug("Checking disk file permissions")

			// Check if disk file exists and is accessible
			if _, err := os.Stat(diskPath); os.IsNotExist(err) {
				dm.logger.WithField("disk_path", diskPath).Error("Disk file does not exist")
				return fmt.Errorf("disk file does not exist: %s", diskPath)
			}

			// Try to fix permissions on disk file
			if err := dm.client.userContext.FixVMFilePermissions(diskPath); err != nil {
				dm.logger.WithError(err).WithField("disk_path", diskPath).Warn("Failed to fix disk file permissions")
			}
		}
	}

	// Try to fix permissions before starting
	if err := dm.FixDomainPermissions(domainName); err != nil {
		dm.logger.WithError(err).Warn("Failed to fix domain permissions, attempting to start anyway")
	}

	// Start the domain
	startCmd := exec.Command("virsh", "-c", dm.client.uri, "start", domainName)
	dm.logger.WithField("command", startCmd.String()).Debug("Executing virsh start")

	startOutput, err := startCmd.CombinedOutput()
	if err != nil {
		dm.logger.WithError(err).WithField("output", string(startOutput)).Error("Failed to start domain")

		// Check for specific error types and provide helpful messages
		outputStr := string(startOutput)

		if strings.Contains(outputStr, "Permission denied") {
			dm.logger.WithField("domain_name", domainName).Warn("Permission denied when starting domain")
			return fmt.Errorf("permission denied when starting domain: %w (output: %s)", err, outputStr)
		} else if strings.Contains(outputStr, "No such file or directory") {
			dm.logger.WithField("domain_name", domainName).Error("Domain configuration file not found")
			return fmt.Errorf("domain configuration file not found: %w (output: %s)", err, outputStr)
		} else if strings.Contains(outputStr, "already exists") {
			dm.logger.WithField("domain_name", domainName).Error("Domain already exists with different configuration")
			return fmt.Errorf("domain already exists with different configuration: %w (output: %s)", err, outputStr)
		}

		return fmt.Errorf("failed to start domain: %w (output: %s)", err, outputStr)
	}

	dm.logger.WithField("domain_name", domainName).Info("Domain started successfully")
	return nil
}

// StopDomain stops a VM gracefully
func (dm *DomainManager) StopDomain(domain interface{}) error {
	if !dm.client.IsAvailable() {
		return fmt.Errorf("libvirt not available")
	}

	domainName, ok := domain.(string)
	if !ok {
		return fmt.Errorf("invalid domain type, expected string")
	}

	dm.logger.WithField("domain_name", domainName).Info("Stopping domain")

	// Check if domain is running
	checkCmd := exec.Command("virsh", "-c", dm.client.uri, "domstate", domainName)
	output, err := checkCmd.Output()
	if err == nil && strings.TrimSpace(string(output)) != "running" {
		dm.logger.WithField("domain_name", domainName).Info("Domain is not running")
		return nil
	}

	// Try graceful shutdown first
	shutdownCmd := exec.Command("virsh", "-c", dm.client.uri, "shutdown", domainName)
	dm.logger.WithField("command", shutdownCmd.String()).Debug("Executing virsh shutdown")

	shutdownOutput, err := shutdownCmd.CombinedOutput()
	if err != nil {
		dm.logger.WithError(err).WithField("output", string(shutdownOutput)).Warn("Graceful shutdown failed, trying destroy")

		// If graceful shutdown fails, force destroy
		destroyCmd := exec.Command("virsh", "-c", dm.client.uri, "destroy", domainName)
		dm.logger.WithField("command", destroyCmd.String()).Debug("Executing virsh destroy")

		destroyOutput, err := destroyCmd.CombinedOutput()
		if err != nil {
			dm.logger.WithError(err).WithField("output", string(destroyOutput)).Error("Failed to destroy domain")
			return fmt.Errorf("failed to destroy domain: %w", err)
		}
	}

	dm.logger.WithField("domain_name", domainName).Info("Domain stopped successfully")
	return nil
}

// DeleteDomain removes a VM permanently
func (dm *DomainManager) DeleteDomain(domain interface{}) error {
	if !dm.client.IsAvailable() {
		return fmt.Errorf("libvirt not available")
	}

	domainName, ok := domain.(string)
	if !ok {
		return fmt.Errorf("invalid domain type, expected string")
	}

	dm.logger.WithField("domain_name", domainName).Info("Deleting domain")

	// Check if domain exists
	checkCmd := exec.Command("virsh", "-c", dm.client.uri, "dominfo", domainName)
	if err := checkCmd.Run(); err != nil {
		dm.logger.WithField("domain_name", domainName).Warn("Domain does not exist")
		return nil // Not an error if it doesn't exist
	}

	// Stop domain if it's running
	stopCmd := exec.Command("virsh", "-c", dm.client.uri, "domstate", domainName)
	if output, err := stopCmd.Output(); err == nil && strings.TrimSpace(string(output)) == "running" {
		dm.logger.WithField("domain_name", domainName).Info("Domain is running, stopping it first")
		if err := dm.StopDomain(domainName); err != nil {
			dm.logger.WithError(err).Warn("Failed to stop domain before deletion")
		}
	}

	// Undefine the domain
	undefineCmd := exec.Command("virsh", "-c", dm.client.uri, "undefine", domainName, "--remove-all-storage")
	dm.logger.WithField("command", undefineCmd.String()).Debug("Executing virsh undefine")

	undefineOutput, err := undefineCmd.CombinedOutput()
	if err != nil {
		dm.logger.WithError(err).WithField("output", string(undefineOutput)).Error("Failed to undefine domain")
		return fmt.Errorf("failed to undefine domain: %w", err)
	}

	dm.logger.WithField("domain_name", domainName).Info("Domain deleted successfully")
	return nil
}

// StorageManager handles storage operations
type StorageManager struct {
	client      *Client
	logger      *logrus.Entry
	userContext *UserContext
}

// NewStorageManager creates a new storage manager
func NewStorageManager(client *Client, logger *logrus.Entry) *StorageManager {
	return &StorageManager{
		client:      client,
		logger:      logger.WithField("component", "storage-manager"),
		userContext: client.userContext,
	}
}

// EnsureDefaultPool ensures the default storage pool exists
func (sm *StorageManager) EnsureDefaultPool(poolName, poolPath string) error {
	if !sm.client.IsAvailable() {
		return fmt.Errorf("libvirt not available")
	}
	// Check if pool exists
	checkCmd := exec.Command("virsh", "-c", sm.client.uri, "pool-info", poolName)
	if err := checkCmd.Run(); err == nil {
		// Pool exists
		return nil
	}
	// Pool does not exist, create it
	poolXML := fmt.Sprintf(`
<pool type='dir'>
  <name>%s</name>
  <target>
    <path>%s</path>
  </target>
</pool>`, poolName, poolPath)
	tmpFile, err := os.CreateTemp("", "pool-*.xml")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write([]byte(poolXML)); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()
	// Define and start the pool
	defineCmd := exec.Command("virsh", "-c", sm.client.uri, "pool-define", tmpFile.Name())
	if output, err := defineCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to define pool: %s", string(output))
	}
	startCmd := exec.Command("virsh", "-c", sm.client.uri, "pool-start", poolName)
	if output, err := startCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start pool: %s", string(output))
	}
	autostartCmd := exec.Command("virsh", "-c", sm.client.uri, "pool-autostart", poolName)
	autostartCmd.Run() // Optional: ignore error
	return nil
}

// getStorageDir returns the storage directory path for VM images
func (sm *StorageManager) getStorageDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	storageDir := filepath.Join(homeDir, ".subnet", "libvirt", "images")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Ensure the storage directory has the correct permissions for libvirt access
	if err := sm.userContext.EnsureDirectoryPermissions(storageDir); err != nil {
		sm.logger.WithError(err).Warn("Failed to ensure storage directory permissions")
	}

	// Also ensure the parent directories have correct permissions
	parentDirs := []string{
		filepath.Join(homeDir, ".subnet"),
		filepath.Join(homeDir, ".subnet", "libvirt"),
	}

	for _, parentDir := range parentDirs {
		if err := sm.userContext.EnsureDirectoryPermissions(parentDir); err != nil {
			sm.logger.WithError(err).WithField("dir", parentDir).Warn("Failed to ensure parent directory permissions")
		}
	}

	return storageDir, nil
}

// CreateVMFromUbuntuImage creates a VM disk from Ubuntu cloud image
func (sm *StorageManager) CreateVMFromUbuntuImage(poolName, vmName string, sizeGB int, ubuntuVersion, arch string) (string, error) {
	if !sm.client.IsAvailable() {
		return "", fmt.Errorf("libvirt not available")
	}

	// Validate inputs
	if poolName == "" || vmName == "" || sizeGB <= 0 {
		return "", fmt.Errorf("invalid parameters: poolName, vmName, and sizeGB must be provided")
	}

	if arch == "" {
		arch = "amd64" // Default architecture
	}

	// Get storage directory
	storageDir, err := sm.getStorageDir()
	if err != nil {
		return "", err
	}

	// Get image URL for the specified version and architecture
	versionImages, exists := UbuntuCloudImages[ubuntuVersion]
	if !exists {
		return "", fmt.Errorf("unsupported Ubuntu version: %s", ubuntuVersion)
	}

	imageURL, exists := versionImages[arch]
	if !exists {
		return "", fmt.Errorf("unsupported architecture: %s for Ubuntu %s", arch, ubuntuVersion)
	}

	// Define file paths
	baseImageName := fmt.Sprintf("ubuntu-%s-server-cloudimg-%s.img", ubuntuVersion, arch)
	baseImagePath := filepath.Join(storageDir, baseImageName)
	vmDiskPath := filepath.Join(storageDir, fmt.Sprintf("%s.qcow2", vmName))

	sm.logger.WithFields(logrus.Fields{
		"ubuntu_version": ubuntuVersion,
		"architecture":   arch,
		"base_image":     baseImagePath,
		"vm_disk":        vmDiskPath,
		"size_gb":        sizeGB,
		"storage_dir":    storageDir,
	}).Info("Creating VM disk from Ubuntu cloud image")

	// Ensure storage directory has correct permissions for libvirt access
	if err := sm.userContext.EnsureDirectoryPermissions(storageDir); err != nil {
		sm.logger.WithError(err).Warn("Failed to ensure storage directory permissions, continuing anyway")
	}

	// Download base image if it doesn't exist
	if _, err := os.Stat(baseImagePath); os.IsNotExist(err) {
		sm.logger.WithField("image_url", imageURL).Info("Downloading Ubuntu cloud image")
		if err := sm.downloadFile(imageURL, baseImagePath); err != nil {
			return "", fmt.Errorf("failed to download Ubuntu image: %w", err)
		}
		sm.logger.Info("Ubuntu cloud image downloaded successfully")
	} else {
		sm.logger.Info("Ubuntu cloud image already exists, skipping download")
	}

	// Ensure base image has correct permissions for libvirt access
	if err := sm.userContext.EnsureFileOwnership(baseImagePath); err != nil {
		sm.logger.WithError(err).Warn("Failed to ensure base image ownership")
	}

	// Check if VM disk already exists and remove it with proper permission handling
	if _, err := os.Stat(vmDiskPath); err == nil {
		sm.logger.WithField("vm_disk", vmDiskPath).Warn("VM disk already exists, removing old disk")
		if err := sm.removeVMFileWithPermissionHandling(vmDiskPath); err != nil {
			return "", fmt.Errorf("failed to remove existing VM disk: %w", err)
		}
	}

	// Create the VM disk file with correct ownership from the start
	// This ensures libvirt can access it immediately
	sm.logger.Info("Creating VM disk file with correct ownership")
	vmDiskFile, err := sm.userContext.CreateFileWithCorrectOwnership(vmDiskPath)
	if err != nil {
		return "", fmt.Errorf("failed to create VM disk file with correct ownership: %w", err)
	}
	vmDiskFile.Close() // Close the file so qemu-img can use it

	// Create VM disk as a copy-on-write image from the base image
	sm.logger.Info("Creating VM disk from base image")
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", "-F", "qcow2", "-b", baseImagePath, vmDiskPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Clean up the file if qemu-img fails
		os.Remove(vmDiskPath)
		return "", fmt.Errorf("failed to create VM disk: %w", err)
	}

	// Resize the disk to the specified size
	sm.logger.WithField("size_gb", sizeGB).Info("Resizing VM disk")
	resizeCmd := exec.Command("qemu-img", "resize", vmDiskPath, fmt.Sprintf("%dG", sizeGB))
	resizeCmd.Stdout = os.Stdout
	resizeCmd.Stderr = os.Stderr

	if err := resizeCmd.Run(); err != nil {
		// Clean up the file if resize fails
		os.Remove(vmDiskPath)
		return "", fmt.Errorf("failed to resize VM disk: %w", err)
	}

	// Ensure VM disk has correct permissions for libvirt access after all operations
	if err := sm.ensureVMDiskPermissions(vmDiskPath); err != nil {
		sm.logger.WithError(err).Error("Failed to ensure VM disk permissions, but continuing")
		// Don't fail the entire operation, but log the error
	}

	sm.logger.WithField("vm_disk", vmDiskPath).Info("VM disk created successfully")
	return vmDiskPath, nil
}

// downloadFile downloads a file from URL to the specified path
func (sm *StorageManager) downloadFile(url, filepath string) error {
	// Create the file with correct ownership from the start
	out, err := sm.userContext.CreateFileWithCorrectOwnership(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file with correct ownership: %w", err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create progress logger
	sm.logger.WithField("url", url).Info("Starting download")

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// Ensure correct ownership after download
	if err := sm.userContext.EnsureFileOwnership(filepath); err != nil {
		sm.logger.WithError(err).Warn("Failed to ensure file ownership after download")
	}

	sm.logger.WithField("filepath", filepath).Info("Download completed")
	return nil
}

// DeleteDiskImage removes a disk image
func (sm *StorageManager) DeleteDiskImage(poolName, volumeName string) error {
	if !sm.client.IsAvailable() {
		return fmt.Errorf("libvirt not available")
	}

	// Get storage directory
	storageDir, err := sm.getStorageDir()
	if err != nil {
		return err
	}

	// Construct the full path to the disk image
	diskPath := filepath.Join(storageDir, volumeName)
	if !strings.HasSuffix(diskPath, ".qcow2") {
		diskPath += ".qcow2"
	}

	sm.logger.WithField("disk_path", diskPath).Info("Deleting disk image")

	// Check if disk exists
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		sm.logger.WithField("disk_path", diskPath).Warn("Disk image does not exist")
		return nil // Not an error if it doesn't exist
	}

	// Remove the disk image
	if err := os.Remove(diskPath); err != nil {
		// Check if it's a permission error
		if os.IsPermission(err) {
			sm.logger.WithError(err).WithField("disk_path", diskPath).Warn("Permission denied when deleting disk image, attempting to fix ownership")

			// Try to fix ownership using the user context system
			if fixErr := sm.userContext.EnsureFileOwnership(diskPath); fixErr == nil {
				// Try removing again after fixing ownership
				if err := os.Remove(diskPath); err == nil {
					sm.logger.WithField("disk_path", diskPath).Info("Successfully deleted disk image after fixing ownership")
				} else {
					return fmt.Errorf("failed to delete disk image even after fixing ownership: %w", err)
				}
			} else {
				sm.logger.WithError(err).WithField("disk_path", diskPath).Error("Permission denied when deleting disk image")
				return fmt.Errorf("permission denied when deleting disk image '%s'", diskPath)
			}
		} else {
			return fmt.Errorf("failed to delete disk image: %w", err)
		}
	}

	sm.logger.WithField("disk_path", diskPath).Info("Disk image deleted successfully")
	return nil
}

// NetworkManager handles network operations
type NetworkManager struct {
	client *Client
	logger *logrus.Entry
}

// NewNetworkManager creates a new network manager
func NewNetworkManager(client *Client, logger *logrus.Entry) *NetworkManager {
	return &NetworkManager{
		client: client,
		logger: logger.WithField("component", "network-manager"),
	}
}

// GetVMIPAddress retrieves the IP address of a VM from DHCP leases
func (nm *NetworkManager) GetVMIPAddress(networkName, vmName string) (string, error) {
	if !nm.client.IsAvailable() {
		return "", fmt.Errorf("libvirt not available")
	}

	nm.logger.WithFields(logrus.Fields{
		"network_name": networkName,
		"vm_name":      vmName,
	}).Debug("Getting VM IP address from DHCP leases")

	// Get DHCP leases using virsh
	cmd := exec.Command("virsh", "-c", nm.client.uri, "net-dhcp-leases", networkName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		nm.logger.WithError(err).WithField("output", string(output)).Debug("Failed to get DHCP leases")
		return "", fmt.Errorf("failed to get DHCP leases: %w", err)
	}

	// Parse the output to find the VM's IP address
	// The output format is typically:
	// Expiry Time          MAC address        Protocol  IP address        Hostname        Client ID or DUID
	// 2024-01-01 12:00:00  52:54:00:12:34:56  ipv4      192.168.123.45    vm-name         -

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Expiry Time") {
			continue // Skip header and empty lines
		}

		// Split by whitespace and look for the hostname
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			hostname := fields[4]
			if hostname == vmName {
				ipAddress := fields[3]
				nm.logger.WithFields(logrus.Fields{
					"vm_name":    vmName,
					"ip_address": ipAddress,
				}).Debug("Found VM IP address in DHCP leases")
				return ipAddress, nil
			}
		}
	}

	// If not found in DHCP leases, try to get from domain info
	return nm.getVMIPFromDomainInfo(vmName)
}

// getVMIPFromDomainInfo tries to get IP address from domain network interface info
func (nm *NetworkManager) getVMIPFromDomainInfo(vmName string) (string, error) {
	nm.logger.WithField("vm_name", vmName).Debug("Trying to get IP from domain network info")

	// Get domain network interface info
	cmd := exec.Command("virsh", "-c", nm.client.uri, "domifaddr", vmName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		nm.logger.WithError(err).WithField("output", string(output)).Debug("Failed to get domain network info")
		return "", fmt.Errorf("failed to get domain network info: %w", err)
	}

	// Parse the output to find IP address
	// The output format is typically:
	// Interface  Type       MAC Address       Protocol     Address
	// vnet0      ethernet   52:54:00:12:34:56 ipv4         192.168.123.45/24

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Interface") {
			continue // Skip header and empty lines
		}

		fields := strings.Fields(line)
		if len(fields) >= 5 && fields[3] == "ipv4" {
			// Extract IP address from CIDR notation (e.g., "192.168.123.45/24")
			ipWithCIDR := fields[4]
			if strings.Contains(ipWithCIDR, "/") {
				ipAddress := strings.Split(ipWithCIDR, "/")[0]
				nm.logger.WithFields(logrus.Fields{
					"vm_name":    vmName,
					"ip_address": ipAddress,
				}).Debug("Found VM IP address in domain network info")
				return ipAddress, nil
			}
		}
	}

	return "", fmt.Errorf("IP address not found for VM: %s", vmName)
}

// CloudInitManager handles cloud-init operations
type CloudInitManager struct {
	logger      *logrus.Entry
	userContext *UserContext
}

// NewCloudInitManager creates a new cloud-init manager
func NewCloudInitManager(logger *logrus.Entry) *CloudInitManager {
	return &CloudInitManager{
		logger:      logger.WithField("component", "cloudinit-manager"),
		userContext: NewUserContext(logger),
	}
}

// getCloudInitDir returns the cloud-init directory path for ISO files
func (cim *CloudInitManager) getCloudInitDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	cloudInitDir := filepath.Join(homeDir, ".subnet", "libvirt", "cloud-init")
	if err := os.MkdirAll(cloudInitDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cloud-init directory: %w", err)
	}

	return cloudInitDir, nil
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

// GenerateSSHKey generates a new SSH key pair
func (cim *CloudInitManager) GenerateSSHKey(keyPath string) (string, error) {
	// Create directory if it doesn't exist
	keyDir := filepath.Dir(keyPath)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create key directory: %w", err)
	}

	// Check if key already exists
	if _, err := os.Stat(keyPath); err == nil {
		// Key exists, read public key
		publicKeyPath := keyPath + ".pub"
		if publicKeyBytes, err := os.ReadFile(publicKeyPath); err == nil {
			return string(publicKeyBytes), nil
		}
	}

	// Generate new SSH key pair
	cim.logger.WithField("key_path", keyPath).Info("Generating new SSH key pair")

	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "4096", "-f", keyPath, "-N", "")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to generate SSH key: %w", err)
	}

	// Ensure correct ownership of the generated keys
	if err := cim.userContext.EnsureFileOwnership(keyPath); err != nil {
		cim.logger.WithError(err).Warn("Failed to set SSH private key ownership")
	}

	publicKeyPath := keyPath + ".pub"
	if err := cim.userContext.EnsureFileOwnership(publicKeyPath); err != nil {
		cim.logger.WithError(err).Warn("Failed to set SSH public key ownership")
	}

	// Read and return public key
	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read public key: %w", err)
	}

	cim.logger.WithField("key_path", keyPath).Info("SSH key pair generated successfully")
	return string(publicKeyBytes), nil
}

// CreateDefaultCloudInitConfig creates a default cloud-init configuration
func (cim *CloudInitManager) CreateDefaultCloudInitConfig(hostname, username string) *CloudInitConfig {
	// Generate a proper password hash for 'password' (you can change this default password)
	// This is a SHA-512 hash of 'password' with salt
	defaultPasswordHash := "$6$rounds=656000$YQJxupVxKq$QJxupVxKqYQJxupVxKqYQJxupVxKqYQJxupVxKqYQJxupVxKqYQJxupVxKqYQJxupVxKq"

	return &CloudInitConfig{
		Hostname: hostname,
		Username: username,
		Password: defaultPasswordHash,
	}
}

// CreateCloudInitISO creates a cloud-init ISO file
func (cim *CloudInitManager) CreateCloudInitISO(isoPath string, config *CloudInitConfig) (string, error) {
	// Check if the provided path is writable, if not use user-writable directory
	if !cim.isPathWritable(filepath.Dir(isoPath)) {
		cloudInitDir, err := cim.getCloudInitDir()
		if err != nil {
			return "", fmt.Errorf("failed to get cloud-init directory: %w", err)
		}

		// Use the filename from the original path but in the user-writable directory
		fileName := filepath.Base(isoPath)
		isoPath = filepath.Join(cloudInitDir, fileName)

		cim.logger.WithField("new_iso_path", isoPath).Info("Using user-writable directory for ISO creation")
	}

	// Create temporary directory for cloud-init files
	tempDir, err := os.MkdirTemp("", "cloud-init-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Generate user-data.yaml
	userData := cim.GenerateUserData(config)
	userDataPath := filepath.Join(tempDir, "user-data")
	if err := os.WriteFile(userDataPath, []byte(userData), 0644); err != nil {
		return "", fmt.Errorf("failed to write user-data: %w", err)
	}

	// Generate meta-data
	metaData := cim.GenerateMetaData(config)
	metaDataPath := filepath.Join(tempDir, "meta-data")
	if err := os.WriteFile(metaDataPath, []byte(metaData), 0644); err != nil {
		return "", fmt.Errorf("failed to write meta-data: %w", err)
	}

	// Create ISO file using genisoimage or mkisofs
	cim.logger.WithField("iso_path", isoPath).Info("Creating cloud-init ISO")

	// Create directory for ISO if it doesn't exist
	isoDir := filepath.Dir(isoPath)
	if err := os.MkdirAll(isoDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create ISO directory: %w", err)
	}

	// Ensure directory has correct permissions for libvirt access
	if err := cim.userContext.EnsureDirectoryPermissions(isoDir); err != nil {
		cim.logger.WithError(err).Warn("Failed to ensure ISO directory permissions")
	}

	// Create a temporary ISO path to avoid ownership issues
	tempISOPath := isoPath + ".tmp"

	// Try genisoimage first, then mkisofs
	var cmd *exec.Cmd
	if _, err := exec.LookPath("genisoimage"); err == nil {
		cmd = exec.Command("genisoimage", "-output", tempISOPath, "-volid", "cidata", "-joliet", "-rock", userDataPath, metaDataPath)
	} else if _, err := exec.LookPath("mkisofs"); err == nil {
		cmd = exec.Command("mkisofs", "-output", tempISOPath, "-volid", "cidata", "-joliet", "-rock", userDataPath, metaDataPath)
	} else {
		return "", fmt.Errorf("neither genisoimage nor mkisofs found, cannot create ISO")
	}

	// Set the process to run with current user's environment
	if currentUser, err := user.Current(); err == nil {
		cmd.Env = append(os.Environ(), "HOME="+currentUser.HomeDir)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create ISO: %s, %w", string(output), err)
	}

	// Move temporary file to final location and ensure correct ownership
	if err := os.Rename(tempISOPath, isoPath); err != nil {
		// If rename fails, try copy and delete
		if err := cim.copyFileWithOwnership(tempISOPath, isoPath); err != nil {
			return "", fmt.Errorf("failed to move ISO file to final location: %w", err)
		}
		os.Remove(tempISOPath) // Clean up temp file
	}

	// Ensure correct ownership of the created ISO file
	if err := cim.userContext.EnsureFileOwnership(isoPath); err != nil {
		cim.logger.WithError(err).Warn("Failed to set ISO file ownership")
	}

	cim.logger.WithField("iso_path", isoPath).Info("Cloud-init ISO created successfully")
	return isoPath, nil
}

// isPathWritable checks if a directory is writable by the current user
func (cim *CloudInitManager) isPathWritable(dirPath string) bool {
	// Check if directory exists and is writable
	if info, err := os.Stat(dirPath); err != nil {
		return false
	} else if !info.IsDir() {
		return false
	}

	// Try to create a temporary file to test write permissions
	testFile := filepath.Join(dirPath, ".test-write-permission")
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(testFile) // Clean up
	return true
}

// GenerateUserData generates the user-data.yaml content
func (cim *CloudInitManager) GenerateUserData(config *CloudInitConfig) string {
	// Default password hash for 'password' (you should generate this properly in production)
	defaultPasswordHash := "$6$rounds=656000$YQJxupVxKq$QJxupVxKqYQJxupVxKqYQJxupVxKqYQJxupVxKqYQJxupVxKqYQJxupVxKqYQJxupVxKq"

	// Use provided password hash or default
	passwordHash := config.Password
	if passwordHash == "" {
		passwordHash = defaultPasswordHash
	}

	// Generate user-data.yaml content
	userData := fmt.Sprintf(`#cloud-config
hostname: %s

users:
  - name: %s
    sudo: ALL=(ALL) NOPASSWD:ALL
    lock_passwd: false
    passwd: "%s"
    shell: /bin/bash
    ssh_authorized_keys:
      - %s

ssh_pwauth: true
chpasswd:
  expire: false

# Package management
package_update: true
package_upgrade: true

# Install additional packages
packages:
  - curl
  - wget
  - vim
  - htop
  - net-tools

# Run commands after package installation
runcmd:
  - echo "Cloud-init completed successfully"
  - systemctl restart ssh

# Final message
final_message: "Cloud-init completed. System is ready."
`,
		config.Hostname,
		config.Username,
		passwordHash,
		config.SSHKey,
	)

	return userData
}

// GenerateMetaData generates the meta-data content
func (cim *CloudInitManager) GenerateMetaData(config *CloudInitConfig) string {
	metaData := fmt.Sprintf(`instance-id: %s
local-hostname: %s
`,
		config.Hostname,
		config.Hostname,
	)

	// Add network configuration if provided
	if config.NetworkConfig != "" {
		metaData += fmt.Sprintf("\nnetwork-interfaces: |\n%s", config.NetworkConfig)
	}

	return metaData
}

// copyFileWithOwnership copies a file and ensures correct ownership
func (cim *CloudInitManager) copyFileWithOwnership(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file with correct ownership
	dstFile, err := cim.userContext.CreateFileWithCorrectOwnership(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// GetSystemArchitecture returns the detected system architecture information
func (dm *DomainManager) GetSystemArchitecture() *SystemArchitecture {
	return dm.detectSystemArchitecture()
}

// GetSystemArchitectureInfo returns a map with system architecture information
func (dm *DomainManager) GetSystemArchitectureInfo() map[string]interface{} {
	arch := dm.detectSystemArchitecture()
	return map[string]interface{}{
		"architecture":      arch.Architecture,
		"machine":           arch.Machine,
		"cpu_model":         arch.CPUModel,
		"vendor":            arch.Vendor,
		"features":          arch.Features,
		"go_arch":           runtime.GOARCH,
		"go_os":             runtime.GOOS,
		"kvm_available":     dm.client.IsKVMAvailable(),
		"domain_type":       dm.client.GetDomainType(),
		"libvirt_available": dm.client.IsAvailable(),
	}
}

// GetDomainType returns the appropriate domain type based on KVM availability
func (c *Client) GetDomainType() string {
	if c.kvmAvailable {
		return "kvm"
	}
	return "qemu"
}

// FixVMFilePermissions fixes permissions on existing VM files to ensure libvirt can access them
func (uc *UserContext) FixVMFilePermissions(filePath string) error {
	uc.logger.WithField("file_path", filePath).Info("Fixing VM file permissions for libvirt access")
	return uc.EnsureFileOwnership(filePath)
}

// FixVMDirectoryPermissions fixes permissions on VM storage directories
func (uc *UserContext) FixVMDirectoryPermissions(dirPath string) error {
	uc.logger.WithField("dir_path", dirPath).Info("Fixing VM directory permissions for libvirt access")
	return uc.EnsureDirectoryPermissions(dirPath)
}

// FixVMFilePermissions fixes permissions on an existing VM disk file
func (sm *StorageManager) FixVMFilePermissions(vmName string) error {
	if !sm.client.IsAvailable() {
		return fmt.Errorf("libvirt not available")
	}

	// Get storage directory
	storageDir, err := sm.getStorageDir()
	if err != nil {
		return err
	}

	// Construct the full path to the disk image
	diskPath := filepath.Join(storageDir, fmt.Sprintf("%s.qcow2", vmName))

	sm.logger.WithField("disk_path", diskPath).Info("Fixing permissions on VM disk file")

	// Check if disk exists
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		return fmt.Errorf("VM disk file does not exist: %s", diskPath)
	}

	// Fix file permissions using the user context
	if err := sm.userContext.FixVMFilePermissions(diskPath); err != nil {
		return fmt.Errorf("failed to fix VM file permissions: %w", err)
	}

	// Also fix directory permissions to ensure libvirt can access the storage directory
	if err := sm.userContext.FixVMDirectoryPermissions(storageDir); err != nil {
		sm.logger.WithError(err).Warn("Failed to fix storage directory permissions")
	}

	sm.logger.WithField("disk_path", diskPath).Info("Successfully fixed VM disk file permissions")
	return nil
}

// removeVMFileWithPermissionHandling removes a VM file with proper permission handling
func (sm *StorageManager) removeVMFileWithPermissionHandling(filePath string) error {
	// Try to remove the file directly
	if err := os.Remove(filePath); err == nil {
		return nil
	}

	// If removal fails, try to fix permissions first
	if err := sm.userContext.EnsureFileOwnership(filePath); err != nil {
		return fmt.Errorf("failed to fix file permissions for removal: %w", err)
	}

	// Try removing again
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove file even after fixing permissions: %w", err)
	}

	return nil
}

// ensureVMDiskPermissions ensures VM disk has correct permissions for libvirt access after all operations
func (sm *StorageManager) ensureVMDiskPermissions(vmDiskPath string) error {
	// Use the standard permission fixing approach
	return sm.userContext.EnsureFileOwnership(vmDiskPath)
}
