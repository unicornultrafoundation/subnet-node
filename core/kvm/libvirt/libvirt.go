package libvirt

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// Client wraps libvirt connection and provides KVM/QEMU management
type Client struct {
	conn      interface{} // Will be *libvirt.Connect when libvirt is available
	uri       string
	logger    *logrus.Entry
	mu        sync.RWMutex
	available bool
}

// NewClient creates a new libvirt client
func NewClient(uri string, logger *logrus.Entry) (*Client, error) {
	if uri == "" {
		uri = "qemu:///system" // Default to system QEMU connection
	}

	logger.WithField("uri", uri).Info("Creating new libvirt client")

	client := &Client{
		uri:    uri,
		logger: logger.WithField("component", "libvirt-client"),
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

	c.logger.WithField("output", string(output)).Debug("Successfully connected to libvirt")
	return nil
}

// IsAvailable returns whether libvirt is available
func (c *Client) IsAvailable() bool {
	return c.available
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

	// Generate domain XML
	domainXML := dm.generateDomainXML(name, uuid, memoryMB, vcpus, diskPath, networkName, cloudInitISOPath)

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

	// Define domain using virsh
	defineCmd := exec.Command("virsh", "-c", dm.client.uri, "define", tmpFile.Name())
	dm.logger.WithField("command", defineCmd.String()).Debug("Executing virsh define")

	output, err := defineCmd.CombinedOutput()
	if err != nil {
		dm.logger.WithError(err).WithField("output", string(output)).Error("Failed to define domain")
		return nil, fmt.Errorf("failed to define domain: %w", err)
	}

	// Set domain to autostart
	autostartCmd := exec.Command("virsh", "-c", dm.client.uri, "autostart", name)
	dm.logger.WithField("command", autostartCmd.String()).Debug("Executing virsh autostart")

	autostartOutput, err := autostartCmd.CombinedOutput()
	if err != nil {
		dm.logger.WithError(err).WithField("output", string(autostartOutput)).Warn("Failed to set domain autostart")
		// Don't fail the entire operation for this
	}

	dm.logger.WithField("domain_name", name).Info("Domain created successfully")
	return name, nil
}

// generateDomainXML generates the XML definition for a VM domain
func (dm *DomainManager) generateDomainXML(name, uuid string, memoryMB, vcpus int, diskPath, networkName, cloudInitISOPath string) string {
	// Determine disk format based on file extension
	diskFormat := "qcow2"
	if strings.HasSuffix(diskPath, ".raw") {
		diskFormat = "raw"
	} else if strings.HasSuffix(diskPath, ".vmdk") {
		diskFormat = "vmdk"
	}

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
	domainXML := fmt.Sprintf(`<domain type='kvm'>
  <name>%s</name>
  <uuid>%s</uuid>
  <memory unit='MiB'>%d</memory>
  <currentMemory unit='MiB'>%d</currentMemory>
  <vcpu>%d</vcpu>
  <os>
    <type arch='x86_64' machine='pc-q35-2.12'>hvm</type>
    <boot dev='hd'/>
    <boot dev='cdrom'/>
  </os>
  <features>
    <acpi/>
    <apic/>
    <vmx state='on'/>
  </features>
  <cpu mode='host-model' check='partial'/>
  <clock offset='utc'/>
  <on_poweroff>destroy</on_poweroff>
  <on_reboot>restart</on_reboot>
  <on_crash>destroy</on_crash>
  <pm>
    <suspend-to-mem enabled='no'/>
    <suspend-to-disk enabled='no'/>
  </pm>
  <devices>
%s
    <interface type='network'>
      <source network='%s'/>
      <model type='virtio'/>
    </interface>
    <console type='pty'>
      <target type='serial' port='0'/>
    </console>
    <input type='tablet' bus='usb'/>
    <input type='keyboard' bus='usb'/>
    <graphics type='vnc' port='-1' autoport='yes' listen='0.0.0.0'>
      <listen type='address' address='0.0.0.0'/>
    </graphics>
    <video>
      <model type='cirrus' vram='16384' heads='1' primary='yes'/>
    </video>
    <memballoon model='virtio'/>
    <rng model='virtio'>
      <backend model='random'>/dev/urandom</backend>
    </rng>
  </devices>
</domain>`, name, uuid, memoryMB, memoryMB, vcpus, diskDevices, networkName)

	return domainXML
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

	// Start the domain
	startCmd := exec.Command("virsh", "-c", dm.client.uri, "start", domainName)
	dm.logger.WithField("command", startCmd.String()).Debug("Executing virsh start")

	startOutput, err := startCmd.CombinedOutput()
	if err != nil {
		dm.logger.WithError(err).WithField("output", string(startOutput)).Error("Failed to start domain")
		return fmt.Errorf("failed to start domain: %w", err)
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
	client *Client
	logger *logrus.Entry
}

// NewStorageManager creates a new storage manager
func NewStorageManager(client *Client, logger *logrus.Entry) *StorageManager {
	return &StorageManager{
		client: client,
		logger: logger.WithField("component", "storage-manager"),
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
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create storage directory: %w", err)
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

	// Define Ubuntu cloud image URLs and checksums
	ubuntuImages := map[string]map[string]string{
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

	// Get image URL for the specified version and architecture
	versionImages, exists := ubuntuImages[ubuntuVersion]
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

	// Check if VM disk already exists
	if _, err := os.Stat(vmDiskPath); err == nil {
		sm.logger.WithField("vm_disk", vmDiskPath).Warn("VM disk already exists, removing old disk")
		if err := os.Remove(vmDiskPath); err != nil {
			return "", fmt.Errorf("failed to remove existing VM disk: %w", err)
		}
	}

	// Create VM disk as a copy-on-write image from the base image
	sm.logger.Info("Creating VM disk from base image")
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", "-F", "qcow2", "-b", baseImagePath, vmDiskPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create VM disk: %w", err)
	}

	// Resize the disk to the specified size
	sm.logger.WithField("size_gb", sizeGB).Info("Resizing VM disk")
	resizeCmd := exec.Command("qemu-img", "resize", vmDiskPath, fmt.Sprintf("%dG", sizeGB))
	resizeCmd.Stdout = os.Stdout
	resizeCmd.Stderr = os.Stderr

	if err := resizeCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to resize VM disk: %w", err)
	}

	// Set proper permissions (readable by libvirt-qemu user)
	if err := os.Chmod(vmDiskPath, 0644); err != nil {
		sm.logger.WithError(err).Warn("Failed to set disk permissions")
	}

	sm.logger.WithField("vm_disk", vmDiskPath).Info("VM disk created successfully")
	return vmDiskPath, nil
}

// downloadFile downloads a file from URL to the specified path
func (sm *StorageManager) downloadFile(url, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
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
		return fmt.Errorf("failed to delete disk image: %w", err)
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
	keyDir := keyPath[:strings.LastIndex(keyPath, "/")]
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

	// Read and return public key
	publicKeyPath := keyPath + ".pub"
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

	// Try genisoimage first, then mkisofs
	var cmd *exec.Cmd
	if _, err := exec.LookPath("genisoimage"); err == nil {
		cmd = exec.Command("genisoimage", "-output", isoPath, "-volid", "cidata", "-joliet", "-rock", userDataPath, metaDataPath)
	} else if _, err := exec.LookPath("mkisofs"); err == nil {
		cmd = exec.Command("mkisofs", "-output", isoPath, "-volid", "cidata", "-joliet", "-rock", userDataPath, metaDataPath)
	} else {
		return "", fmt.Errorf("neither genisoimage nor mkisofs found, cannot create ISO")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create ISO: %s, %w", string(output), err)
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
