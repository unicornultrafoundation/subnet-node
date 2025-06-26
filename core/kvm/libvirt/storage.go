//go:build libvirt
// +build libvirt

package libvirt

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"libvirt.org/go/libvirt"
)

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

// StoragePoolXML represents libvirt storage pool XML
type StoragePoolXML struct {
	XMLName xml.Name         `xml:"pool"`
	Type    string           `xml:"type,attr"`
	Name    string           `xml:"name"`
	Target  StorageTargetXML `xml:"target"`
}

type StorageTargetXML struct {
	Path string `xml:"path"`
}

// StorageVolXML represents libvirt storage volume XML
type StorageVolXML struct {
	XMLName      xml.Name                `xml:"volume"`
	Name         string                  `xml:"name"`
	Capacity     StorageCapacityXML      `xml:"capacity"`
	Target       StorageVolTargetXML     `xml:"target"`
	BackingStore *StorageBackingStoreXML `xml:"backingStore,omitempty"`
}

type StorageCapacityXML struct {
	Unit  string `xml:"unit,attr"`
	Value string `xml:",chardata"`
}

type StorageVolTargetXML struct {
	Path   string           `xml:"path"`
	Format StorageFormatXML `xml:"format"`
}

type StorageFormatXML struct {
	Type string `xml:"type,attr"`
}

type StorageBackingStoreXML struct {
	Path   string           `xml:"path"`
	Format StorageFormatXML `xml:"format"`
}

// CreateStoragePool creates a new storage pool
func (sm *StorageManager) CreateStoragePool(name, path string) (*libvirt.StoragePool, error) {
	// Ensure directory exists
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	poolXML := StoragePoolXML{
		Type: "dir",
		Name: name,
		Target: StorageTargetXML{
			Path: path,
		},
	}

	xmlData, err := xml.MarshalIndent(poolXML, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal storage pool XML: %w", err)
	}

	xmlString := `<?xml version="1.0" encoding="UTF-8"?>` + "\n" + string(xmlData)

	pool, err := sm.client.CreateStoragePool(xmlString)
	if err != nil {
		return nil, err
	}

	sm.logger.WithFields(logrus.Fields{
		"pool": name,
		"path": path,
	}).Info("Storage pool created")

	return pool, nil
}

// GetOrCreateStoragePool gets existing pool or creates new one
func (sm *StorageManager) GetOrCreateStoragePool(name, path string) (*libvirt.StoragePool, error) {
	// Try to get existing pool
	pool, err := sm.client.GetStoragePoolByName(name)
	if err == nil {
		// Pool exists, make sure it's active
		active, err := pool.IsActive()
		if err != nil {
			return nil, fmt.Errorf("failed to check pool status: %w", err)
		}
		if !active {
			if err := pool.Create(0); err != nil {
				return nil, fmt.Errorf("failed to activate pool: %w", err)
			}
		}
		return pool, nil
	}

	// Pool doesn't exist, create it
	return sm.CreateStoragePool(name, path)
}

// CreateDiskImage creates a new disk image
func (sm *StorageManager) CreateDiskImage(poolName, volumeName string, sizeGB int, backingFile string) (string, error) {
	pool, err := sm.client.GetStoragePoolByName(poolName)
	if err != nil {
		return "", fmt.Errorf("failed to get storage pool: %w", err)
	}

	// Get pool path
	poolXMLStr, err := pool.GetXMLDesc(0)
	if err != nil {
		return "", fmt.Errorf("failed to get pool XML: %w", err)
	}

	var poolXML StoragePoolXML
	if err := xml.Unmarshal([]byte(poolXMLStr), &poolXML); err != nil {
		return "", fmt.Errorf("failed to parse pool XML: %w", err)
	}

	diskPath := filepath.Join(poolXML.Target.Path, volumeName+".qcow2")

	// Create volume XML
	volXML := StorageVolXML{
		Name: volumeName,
		Capacity: StorageCapacityXML{
			Unit:  "GiB",
			Value: fmt.Sprintf("%d", sizeGB),
		},
		Target: StorageVolTargetXML{
			Path:   diskPath,
			Format: StorageFormatXML{Type: "qcow2"},
		},
	}

	// Add backing store if specified
	if backingFile != "" {
		volXML.BackingStore = &StorageBackingStoreXML{
			Path:   backingFile,
			Format: StorageFormatXML{Type: "qcow2"},
		}
	}

	xmlData, err := xml.MarshalIndent(volXML, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal volume XML: %w", err)
	}

	xmlString := `<?xml version="1.0" encoding="UTF-8"?>` + "\n" + string(xmlData)

	// Create volume
	_, err = pool.StorageVolCreateXML(xmlString, 0)
	if err != nil {
		return "", fmt.Errorf("failed to create storage volume: %w", err)
	}

	sm.logger.WithFields(logrus.Fields{
		"volume": volumeName,
		"path":   diskPath,
		"size":   fmt.Sprintf("%dGB", sizeGB),
	}).Info("Disk image created")

	return diskPath, nil
}

// CreateDiskImageFromTemplate creates a disk image from a template
func (sm *StorageManager) CreateDiskImageFromTemplate(poolName, volumeName, templatePath string, sizeGB int) (string, error) {
	if templatePath == "" {
		return sm.CreateDiskImage(poolName, volumeName, sizeGB, "")
	}

	pool, err := sm.client.GetStoragePoolByName(poolName)
	if err != nil {
		return "", fmt.Errorf("failed to get storage pool: %w", err)
	}

	// Get pool path
	poolXMLStr, err := pool.GetXMLDesc(0)
	if err != nil {
		return "", fmt.Errorf("failed to get pool XML: %w", err)
	}

	var poolXML StoragePoolXML
	if err := xml.Unmarshal([]byte(poolXMLStr), &poolXML); err != nil {
		return "", fmt.Errorf("failed to parse pool XML: %w", err)
	}

	diskPath := filepath.Join(poolXML.Target.Path, volumeName+".qcow2")

	// Use qemu-img to create image from template
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", "-F", "qcow2", "-b", templatePath, diskPath, fmt.Sprintf("%dG", sizeGB))
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create disk from template: %w", err)
	}

	sm.logger.WithFields(logrus.Fields{
		"volume":   volumeName,
		"path":     diskPath,
		"template": templatePath,
		"size":     fmt.Sprintf("%dGB", sizeGB),
	}).Info("Disk image created from template")

	return diskPath, nil
}

// DeleteDiskImage removes a disk image
func (sm *StorageManager) DeleteDiskImage(poolName, volumeName string) error {
	pool, err := sm.client.GetStoragePoolByName(poolName)
	if err != nil {
		return fmt.Errorf("failed to get storage pool: %w", err)
	}

	vol, err := pool.LookupStorageVolByName(volumeName)
	if err != nil {
		return fmt.Errorf("failed to lookup volume: %w", err)
	}

	if err := vol.Delete(0); err != nil {
		return fmt.Errorf("failed to delete volume: %w", err)
	}

	sm.logger.WithField("volume", volumeName).Info("Disk image deleted")
	return nil
}

// GetDiskImagePath returns the path of a disk image
func (sm *StorageManager) GetDiskImagePath(poolName, volumeName string) (string, error) {
	pool, err := sm.client.GetStoragePoolByName(poolName)
	if err != nil {
		return "", fmt.Errorf("failed to get storage pool: %w", err)
	}

	vol, err := pool.LookupStorageVolByName(volumeName)
	if err != nil {
		return "", fmt.Errorf("failed to lookup volume: %w", err)
	}

	path, err := vol.GetPath()
	if err != nil {
		return "", fmt.Errorf("failed to get volume path: %w", err)
	}

	return path, nil
}

// ListVolumes returns all volumes in a storage pool
func (sm *StorageManager) ListVolumes(poolName string) ([]string, error) {
	pool, err := sm.client.GetStoragePoolByName(poolName)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage pool: %w", err)
	}

	volumeNames, err := pool.ListStorageVolumes()
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}

	return volumeNames, nil
}

// EnsureDefaultPool ensures the default storage pool exists
func (sm *StorageManager) EnsureDefaultPool(poolName, poolPath string) error {
	_, err := sm.GetOrCreateStoragePool(poolName, poolPath)
	if err != nil {
		return fmt.Errorf("failed to ensure default storage pool: %w", err)
	}

	sm.logger.WithFields(logrus.Fields{
		"pool": poolName,
		"path": poolPath,
	}).Info("Default storage pool ensured")

	return nil
}

// GetPoolInfo returns information about a storage pool
func (sm *StorageManager) GetPoolInfo(poolName string) (*libvirt.StoragePoolInfo, error) {
	pool, err := sm.client.GetStoragePoolByName(poolName)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage pool: %w", err)
	}

	info, err := pool.GetInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get pool info: %w", err)
	}

	return info, nil
}

// DownloadUbuntuCloudImage downloads Ubuntu cloud image if not already present
func (sm *StorageManager) DownloadUbuntuCloudImage(version, arch string, poolPath string) (string, error) {
	// Default to Ubuntu 22.04 LTS if version not specified
	if version == "" {
		version = "jammy"
	}
	if arch == "" {
		arch = "amd64"
	}

	// Construct image filename
	imageName := fmt.Sprintf("ubuntu-%s.qcow2", version)
	imagePath := filepath.Join(poolPath, imageName)

	// Check if image already exists
	if _, err := os.Stat(imagePath); err == nil {
		sm.logger.WithField("image_path", imagePath).Info("Ubuntu cloud image already exists")
		return imagePath, nil
	}

	// Construct download URL
	url := fmt.Sprintf("https://cloud-images.ubuntu.com/%s/current/%s-server-cloudimg-%s.img", version, version, arch)

	sm.logger.WithFields(logrus.Fields{
		"url":        url,
		"image_path": imagePath,
	}).Info("Downloading Ubuntu cloud image")

	// Download the image
	if err := sm.downloadFile(url, imagePath); err != nil {
		return "", fmt.Errorf("failed to download Ubuntu cloud image: %w", err)
	}

	sm.logger.WithField("image_path", imagePath).Info("Ubuntu cloud image downloaded successfully")
	return imagePath, nil
}

// downloadFile downloads a file from URL to local path
func (sm *StorageManager) downloadFile(url, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// CreateVMFromUbuntuImage creates a VM disk from Ubuntu cloud image
func (sm *StorageManager) CreateVMFromUbuntuImage(poolName, vmName string, sizeGB int, ubuntuVersion, arch string) (string, error) {
	// Get pool path
	pool, err := sm.client.GetStoragePoolByName(poolName)
	if err != nil {
		return "", fmt.Errorf("failed to get storage pool: %w", err)
	}

	poolXMLStr, err := pool.GetXMLDesc(0)
	if err != nil {
		return "", fmt.Errorf("failed to get pool XML: %w", err)
	}

	var poolXML StoragePoolXML
	if err := xml.Unmarshal([]byte(poolXMLStr), &poolXML); err != nil {
		return "", fmt.Errorf("failed to parse pool XML: %w", err)
	}

	// Download Ubuntu cloud image if not present
	baseImagePath, err := sm.DownloadUbuntuCloudImage(ubuntuVersion, arch, poolXML.Target.Path)
	if err != nil {
		return "", fmt.Errorf("failed to get Ubuntu cloud image: %w", err)
	}

	// Create VM disk using qemu-img
	vmDiskPath := filepath.Join(poolXML.Target.Path, vmName+".qcow2")

	// Use qemu-img to create a new disk based on the Ubuntu image
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", "-b", baseImagePath, vmDiskPath, fmt.Sprintf("%dG", sizeGB))
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("qemu-img failed: %s, %w", string(output), err)
	}

	sm.logger.WithFields(logrus.Fields{
		"vm_name":    vmName,
		"disk_path":  vmDiskPath,
		"size":       fmt.Sprintf("%dGB", sizeGB),
		"base_image": baseImagePath,
	}).Info("VM disk created from Ubuntu cloud image")

	return vmDiskPath, nil
}
