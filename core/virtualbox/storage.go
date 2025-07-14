package virtualbox

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

var storageLog = logrus.WithField("service", "virtualbox-storage")

// StorageManagerImpl implements the StorageManager interface
type StorageManagerImpl struct {
	isoDir string
	config *ServiceConfig
}

// NewStorageManager creates a new storage manager
func NewStorageManager() (*StorageManagerImpl, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	isoDir := filepath.Join(homeDir, "VirtualBox VMs", "ISOs")
	if err := os.MkdirAll(isoDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create ISO directory: %w", err)
	}

	return &StorageManagerImpl{
		isoDir: isoDir,
	}, nil
}

// NewStorageManagerWithConfig creates a new storage manager with configuration
func NewStorageManagerWithConfig(config *ServiceConfig) (*StorageManagerImpl, error) {
	storageMgr, err := NewStorageManager()
	if err != nil {
		return nil, err
	}
	storageMgr.config = config
	return storageMgr, nil
}

// DownloadFile downloads a file from URL to the specified destination
func (s *StorageManagerImpl) DownloadFile(ctx context.Context, url, destPath string) error {
	storageLog.Infof("Downloading %s to %s", url, destPath)

	// Create the destination directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create the destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Minute, // Long timeout for large ISO files
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add user agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; VirtualBox-ISO-Downloader/1.0)")

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Copy the response body to the destination file
	_, err = io.Copy(destFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	storageLog.Infof("Successfully downloaded %s", url)
	return nil
}

// GetFileInfo gets information about a file
func (s *StorageManagerImpl) GetFileInfo(filePath string) (*vbtypes.ISOInfo, error) {
	if !s.FileExists(filePath) {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}

	// Get file size
	size, err := s.GetFileSize(filePath)
	if err != nil {
		return nil, err
	}

	// Calculate checksum
	checksum, err := s.CalculateChecksum(filePath)
	if err != nil {
		return nil, err
	}

	// Get file modification time
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return &vbtypes.ISOInfo{
		Path:         filePath,
		Size:         size,
		Checksum:     checksum,
		DownloadedAt: fileInfo.ModTime(),
	}, nil
}

// ListFiles lists all files in a directory
func (s *StorageManagerImpl) ListFiles(dirPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			// Check if it's an ISO file
			if strings.HasSuffix(strings.ToLower(path), ".iso") {
				files = append(files, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}

// DeleteFile deletes a file
func (s *StorageManagerImpl) DeleteFile(filePath string) error {
	if !s.FileExists(filePath) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	storageLog.Infof("Deleted file: %s", filePath)
	return nil
}

// FileExists checks if a file exists
func (s *StorageManagerImpl) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// GetFileSize gets the size of a file in bytes
func (s *StorageManagerImpl) GetFileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}

	return fileInfo.Size(), nil
}

// CalculateChecksum calculates SHA256 checksum of a file
func (s *StorageManagerImpl) CalculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// GetISODir returns the ISO directory path
func (s *StorageManagerImpl) GetISODir() string {
	return s.isoDir
}

// GetISOFileName extracts the filename from a URL
func (s *StorageManagerImpl) GetISOFileName(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "unknown.iso"
}

// GetISOPath returns the full path for an ISO file
func (s *StorageManagerImpl) GetISOPath(isoURL string) string {
	fileName := s.GetISOFileName(isoURL)
	return filepath.Join(s.isoDir, fileName)
}

// ISO OS Type Management Methods

// DetermineOSTypeAndISO handles the complete process of determining OS type and ISO URL
func (s *StorageManagerImpl) DetermineOSTypeAndISO(ctx context.Context, req vbtypes.VMCreateRequest) (string, string, error) {
	storageLog.Infof("Starting OS type and ISO determination process...")

	// Step 1: Determine OS type
	osType, err := s.determineOSType(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to determine OS type: %w", err)
	}
	storageLog.Infof("Determined OS type: %s", osType)

	// Step 2: Determine ISO URL
	isoURL, err := s.determineISOURL(req, osType)
	if err != nil {
		return "", "", fmt.Errorf("failed to determine ISO URL: %w", err)
	}
	storageLog.Infof("Determined ISO URL: %s", isoURL)

	// Step 3: Validate the combination
	if err := s.validateOSTypeAndISOCombination(osType, isoURL); err != nil {
		return "", "", fmt.Errorf("invalid OS type and ISO combination: %w", err)
	}

	storageLog.Infof("OS type and ISO determination completed successfully")
	return osType, isoURL, nil
}

// determineOSType determines the appropriate OS type based on request and system architecture
func (s *StorageManagerImpl) determineOSType(req vbtypes.VMCreateRequest) (string, error) {
	// If OS type is explicitly provided in the request, use it
	if req.OSType != "" {
		storageLog.Infof("Using provided OS type: %s", req.OSType)
		return req.OSType, nil
	}

	// Use default from configuration if available
	if s.config != nil && s.config.DefaultOSType != "" {
		storageLog.Infof("Using default OS type from config: %s", s.config.DefaultOSType)
		return s.config.DefaultOSType, nil
	}

	// Fallback to architecture-based detection
	arch := runtime.GOARCH
	osType := s.getOSTypeForArchitecture(arch)
	storageLog.Infof("Using architecture-based OS type: %s (for %s)", osType, arch)
	return osType, nil
}

// determineISOURL determines the appropriate ISO URL based on request and OS type
func (s *StorageManagerImpl) determineISOURL(req vbtypes.VMCreateRequest, osType string) (string, error) {
	// If ISO URL is explicitly provided in the request, use it
	if req.ISOURL != "" {
		storageLog.Infof("Using provided ISO URL: %s", req.ISOURL)
		return req.ISOURL, nil
	}

	// Determine ISO URL based on OS type
	isoURL := s.getISOURLForOSType(osType)
	if isoURL == "" {
		return "", fmt.Errorf("no ISO URL available for OS type: %s", osType)
	}

	storageLog.Infof("Using ISO URL for OS type %s: %s", osType, isoURL)
	return isoURL, nil
}

// validateOSTypeAndISOCombination validates that the OS type and ISO URL are compatible
func (s *StorageManagerImpl) validateOSTypeAndISOCombination(osType, isoURL string) error {
	// Basic validation - check if both are provided
	if osType == "" {
		return fmt.Errorf("OS type cannot be empty")
	}
	if isoURL == "" {
		return fmt.Errorf("ISO URL cannot be empty")
	}

	// Validate OS type format
	if !s.isValidOSType(osType) {
		return fmt.Errorf("invalid OS type format: %s", osType)
	}

	// Validate ISO URL format
	if !s.isValidISOURL(isoURL) {
		return fmt.Errorf("invalid ISO URL format: %s", isoURL)
	}

	// Check architecture compatibility
	if err := s.validateArchitectureCompatibility(osType); err != nil {
		return fmt.Errorf("architecture compatibility check failed: %w", err)
	}

	return nil
}

// getOSTypeForArchitecture returns the appropriate OS type for the given architecture
func (s *StorageManagerImpl) getOSTypeForArchitecture(arch string) string {
	switch arch {
	case "arm64", "aarch64":
		return "Ubuntu_ARM64"
	case "amd64", "x86_64":
		return "Ubuntu_64"
	case "arm":
		return "Ubuntu"
	case "386", "i386":
		return "Ubuntu"
	default:
		// Default to Ubuntu 64-bit for unknown architectures
		return "Ubuntu_64"
	}
}

// isValidOSType checks if the OS type has a valid format
func (s *StorageManagerImpl) isValidOSType(osType string) bool {
	// Basic validation - check if it contains valid characters and format
	if osType == "" {
		return false
	}

	// Check for common OS type patterns
	validPatterns := []string{
		"Ubuntu", "Ubuntu_64", "Ubuntu_ARM64",
		"Debian", "Debian_64", "Debian_ARM64",
		"Fedora", "Fedora_64", "Fedora_ARM64",
		"CentOS", "CentOS_64", "CentOS_ARM64",
		"RedHat", "RedHat_64", "RedHat_ARM64",
		"Oracle", "Oracle_64", "Oracle_ARM64",
		"Linux", "Linux_64", "Linux_ARM64",
		"Windows", "Windows_64", "Windows_ARM64",
		"FreeBSD", "FreeBSD_64", "FreeBSD_ARM64",
		"NetBSD", "NetBSD_64", "NetBSD_ARM64",
		"BSD", "BSD_64", "BSD_ARM64",
		"Other", "Other_64", "Other_ARM64",
	}

	for _, pattern := range validPatterns {
		if osType == pattern {
			return true
		}
	}

	return false
}

// isValidISOURL checks if the ISO URL has a valid format
func (s *StorageManagerImpl) isValidISOURL(isoURL string) bool {
	if isoURL == "" {
		return false
	}

	// Check if it's a valid HTTP/HTTPS URL
	if !strings.HasPrefix(isoURL, "http://") && !strings.HasPrefix(isoURL, "https://") {
		return false
	}

	// Check if it ends with .iso
	if !strings.HasSuffix(isoURL, ".iso") {
		return false
	}

	return true
}

// validateArchitectureCompatibility checks if the OS type is compatible with the current architecture
func (s *StorageManagerImpl) validateArchitectureCompatibility(osType string) error {
	currentArch := runtime.GOARCH

	// Define architecture compatibility rules
	compatibilityMap := map[string][]string{
		"arm64": {"Ubuntu_ARM64", "Debian_ARM64", "Fedora_ARM64", "CentOS_ARM64", "RedHat_ARM64", "Oracle_ARM64", "Linux_ARM64", "FreeBSD_ARM64", "NetBSD_ARM64", "BSD_ARM64", "Other_ARM64"},
		"amd64": {"Ubuntu_64", "Debian_64", "Fedora_64", "CentOS_64", "RedHat_64", "Oracle_64", "Linux_64", "Windows_64", "FreeBSD_64", "NetBSD_64", "BSD_64", "Other_64"},
		"arm":   {"Ubuntu", "Debian", "Fedora", "CentOS", "RedHat", "Oracle", "Linux", "FreeBSD", "NetBSD", "BSD", "Other"},
		"386":   {"Ubuntu", "Debian", "Fedora", "CentOS", "RedHat", "Oracle", "Linux", "Windows", "FreeBSD", "NetBSD", "BSD", "Other"},
	}

	// Get compatible OS types for current architecture
	compatibleTypes, exists := compatibilityMap[currentArch]
	if !exists {
		// If architecture not found, assume all types are compatible
		storageLog.Warnf("Unknown architecture %s, assuming all OS types are compatible", currentArch)
		return nil
	}

	// Check if the OS type is compatible
	for _, compatibleType := range compatibleTypes {
		if osType == compatibleType {
			return nil
		}
	}

	return fmt.Errorf("OS type %s is not compatible with architecture %s", osType, currentArch)
}

// getISOURLForOSType returns the ISO URL for a given OS type
func (s *StorageManagerImpl) getISOURLForOSType(osType string) string {
	switch osType {
	case "Ubuntu_64":
		// Use a working Ubuntu 24.04 LTS server ISO URL
		return "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-amd64.iso"
	case "Ubuntu_ARM64":
		// Use Ubuntu 24.04 LTS server ARM64 ISO
		return "https://cdimage.ubuntu.com/releases/24.04/release/ubuntu-24.04.2-live-server-arm64.iso"
	case "Ubuntu":
		// Use appropriate Ubuntu ISO based on architecture
		arch := runtime.GOARCH
		switch arch {
		case "arm64", "aarch64":
			return "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-arm64.iso"
		case "amd64", "x86_64":
			return "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-amd64.iso"
		case "arm":
			return "https://cdimage.ubuntu.com/releases/24.04/release/ubuntu-24.04.2-live-server-arm64.iso"
		case "386", "i386":
			return "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-i386.iso"
		default:
			return "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-amd64.iso"
		}
	case "Debian_64":
		return "https://cdimage.debian.org/debian-cd/current/amd64/iso-cd/debian-12.5.0-amd64-netinst.iso"
	case "Debian_ARM64":
		return "https://cdimage.debian.org/debian-cd/current/arm64/iso-cd/debian-12.5.0-arm64-netinst.iso"
	case "Debian":
		// Use appropriate Debian ISO based on architecture
		arch := runtime.GOARCH
		switch arch {
		case "arm64", "aarch64":
			return "https://cdimage.debian.org/debian-cd/current/arm64/iso-cd/debian-12.5.0-arm64-netinst.iso"
		case "amd64", "x86_64":
			return "https://cdimage.debian.org/debian-cd/current/amd64/iso-cd/debian-12.5.0-amd64-netinst.iso"
		case "arm":
			return "https://cdimage.debian.org/debian-cd/current/armhf/iso-cd/debian-12.5.0-armhf-netinst.iso"
		case "386", "i386":
			return "https://cdimage.debian.org/debian-cd/current/i386/iso-cd/debian-12.5.0-i386-netinst.iso"
		default:
			return "https://cdimage.debian.org/debian-cd/current/amd64/iso-cd/debian-12.5.0-amd64-netinst.iso"
		}
	case "Fedora_64":
		return "https://download.fedoraproject.org/pub/fedora/linux/releases/40/Server/x86_64/iso/Fedora-Server-dvd-x86_64-40-1.14.iso"
	case "Fedora_ARM64":
		return "https://download.fedoraproject.org/pub/fedora/linux/releases/40/Server/aarch64/iso/Fedora-Server-dvd-aarch64-40-1.14.iso"
	case "Fedora":
		// Use appropriate Fedora ISO based on architecture
		arch := runtime.GOARCH
		switch arch {
		case "arm64", "aarch64":
			return "https://download.fedoraproject.org/pub/fedora/linux/releases/40/Server/aarch64/iso/Fedora-Server-dvd-aarch64-40-1.14.iso"
		case "amd64", "x86_64":
			return "https://download.fedoraproject.org/pub/fedora/linux/releases/40/Server/x86_64/iso/Fedora-Server-dvd-x86_64-40-1.14.iso"
		default:
			return "https://download.fedoraproject.org/pub/fedora/linux/releases/40/Server/x86_64/iso/Fedora-Server-dvd-x86_64-40-1.14.iso"
		}
	case "CentOS_64":
		return "https://mirror.stream.centos.org/9-stream/BaseOS/x86_64/iso/CentOS-Stream-9-latest-x86_64-boot.iso"
	case "CentOS_ARM64":
		return "https://mirror.stream.centos.org/9-stream/BaseOS/aarch64/iso/CentOS-Stream-9-latest-aarch64-boot.iso"
	case "CentOS":
		// Use appropriate CentOS ISO based on architecture
		arch := runtime.GOARCH
		switch arch {
		case "arm64", "aarch64":
			return "https://mirror.stream.centos.org/9-stream/BaseOS/aarch64/iso/CentOS-Stream-9-latest-aarch64-boot.iso"
		case "amd64", "x86_64":
			return "https://mirror.stream.centos.org/9-stream/BaseOS/x86_64/iso/CentOS-Stream-9-latest-x86_64-boot.iso"
		default:
			return "https://mirror.stream.centos.org/9-stream/BaseOS/x86_64/iso/CentOS-Stream-9-latest-x86_64-boot.iso"
		}
	case "RedHat_64":
		return "https://access.redhat.com/downloads/content/rhel/9.0/x86_64/iso/rhel-9.0-x86_64-boot.iso"
	case "RedHat_ARM64":
		return "https://access.redhat.com/downloads/content/rhel/9.0/aarch64/iso/rhel-9.0-aarch64-boot.iso"
	case "Oracle_64":
		return "https://yum.oracle.com/ISOS/OracleLinux/OL9/u0/x86_64/OracleLinux-R9-U0-x86_64-boot.iso"
	case "Oracle_ARM64":
		return "https://yum.oracle.com/ISOS/OracleLinux/OL9/u0/aarch64/OracleLinux-R9-U0-aarch64-boot.iso"
	case "Linux_64":
		return "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-amd64.iso"
	case "Linux_ARM64":
		return "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-arm64.iso"
	case "Linux":
		// Use appropriate Linux ISO based on architecture
		arch := runtime.GOARCH
		switch arch {
		case "arm64", "aarch64":
			return "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-arm64.iso"
		case "amd64", "x86_64":
			return "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-amd64.iso"
		default:
			return "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-amd64.iso"
		}
	case "Windows_64":
		return "https://software-download.microsoft.com/download/pr/888969d5-f34g-4e03-ac9d-1f9786c69161/22000.318.211104-1236.co_release_svc_refresh_CLIENTENTERPRISEEVAL_OEMRET_x64FRE_en-us.iso"
	case "Windows_ARM64":
		return "https://software-download.microsoft.com/download/pr/888969d5-f34g-4e03-ac9d-1f9786c69161/22000.318.211104-1236.co_release_svc_refresh_CLIENTENTERPRISEEVAL_OEMRET_ARM64FRE_en-us.iso"
	case "FreeBSD_64":
		return "https://download.freebsd.org/ftp/releases/amd64/amd64/13.2-RELEASE/FreeBSD-13.2-RELEASE-amd64-bootonly.iso"
	case "FreeBSD_ARM64":
		return "https://download.freebsd.org/ftp/releases/arm64/aarch64/13.2-RELEASE/FreeBSD-13.2-RELEASE-arm64-aarch64-bootonly.iso"
	case "FreeBSD":
		// Use appropriate FreeBSD ISO based on architecture
		arch := runtime.GOARCH
		switch arch {
		case "arm64", "aarch64":
			return "https://download.freebsd.org/ftp/releases/arm64/aarch64/13.2-RELEASE/FreeBSD-13.2-RELEASE-arm64-aarch64-bootonly.iso"
		case "amd64", "x86_64":
			return "https://download.freebsd.org/ftp/releases/amd64/amd64/13.2-RELEASE/FreeBSD-13.2-RELEASE-amd64-bootonly.iso"
		default:
			return "https://download.freebsd.org/ftp/releases/amd64/amd64/13.2-RELEASE/FreeBSD-13.2-RELEASE-amd64-bootonly.iso"
		}
	case "NetBSD_64":
		return "https://cdn.netbsd.org/pub/NetBSD/NetBSD-9.3/amd64/installation/cdrom/NetBSD-9.3-amd64.iso"
	case "NetBSD_ARM64":
		return "https://cdn.netbsd.org/pub/NetBSD/NetBSD-9.3/evbarm-aarch64/installation/cdrom/NetBSD-9.3-evbarm-aarch64.iso"
	case "BSD_64":
		return "https://download.freebsd.org/ftp/releases/amd64/amd64/13.2-RELEASE/FreeBSD-13.2-RELEASE-amd64-bootonly.iso"
	case "BSD_ARM64":
		return "https://download.freebsd.org/ftp/releases/arm64/aarch64/13.2-RELEASE/FreeBSD-13.2-RELEASE-arm64-aarch64-bootonly.iso"
	case "Other_64":
		return "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-amd64.iso"
	case "Other_ARM64":
		return "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-arm64.iso"
	case "Other":
		// Use appropriate generic ISO based on architecture
		arch := runtime.GOARCH
		switch arch {
		case "arm64", "aarch64":
			return "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-arm64.iso"
		case "amd64", "x86_64":
			return "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-amd64.iso"
		default:
			return "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-amd64.iso"
		}
	default:
		// Default to Ubuntu 64-bit
		return "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-amd64.iso"
	}
}

// GetSupportedOSTypes returns a list of supported OS types for the current architecture
func (s *StorageManagerImpl) GetSupportedOSTypes() []string {
	currentArch := runtime.GOARCH

	compatibilityMap := map[string][]string{
		"arm64": {"Ubuntu_ARM64", "Debian_ARM64", "Fedora_ARM64", "CentOS_ARM64", "RedHat_ARM64", "Oracle_ARM64", "Linux_ARM64", "FreeBSD_ARM64", "NetBSD_ARM64", "BSD_ARM64", "Other_ARM64"},
		"amd64": {"Ubuntu_64", "Debian_64", "Fedora_64", "CentOS_64", "RedHat_64", "Oracle_64", "Linux_64", "Windows_64", "FreeBSD_64", "NetBSD_64", "BSD_64", "Other_64"},
		"arm":   {"Ubuntu", "Debian", "Fedora", "CentOS", "RedHat", "Oracle", "Linux", "FreeBSD", "NetBSD", "BSD", "Other"},
		"386":   {"Ubuntu", "Debian", "Fedora", "CentOS", "RedHat", "Oracle", "Linux", "Windows", "FreeBSD", "NetBSD", "BSD", "Other"},
	}

	if compatibleTypes, exists := compatibilityMap[currentArch]; exists {
		return compatibleTypes
	}

	// Return all types if architecture not found
	return []string{
		"Ubuntu", "Ubuntu_64", "Ubuntu_ARM64",
		"Debian", "Debian_64", "Debian_ARM64",
		"Fedora", "Fedora_64", "Fedora_ARM64",
		"CentOS", "CentOS_64", "CentOS_ARM64",
		"RedHat", "RedHat_64", "RedHat_ARM64",
		"Oracle", "Oracle_64", "Oracle_ARM64",
		"Linux", "Linux_64", "Linux_ARM64",
		"Windows", "Windows_64", "Windows_ARM64",
		"FreeBSD", "FreeBSD_64", "FreeBSD_ARM64",
		"NetBSD", "NetBSD_64", "NetBSD_ARM64",
		"BSD", "BSD_64", "BSD_ARM64",
		"Other", "Other_64", "Other_ARM64",
	}
}
