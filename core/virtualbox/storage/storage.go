package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

var storageLog = logrus.WithField("package", "storage")

// StorageManager handles image downloads and storage operations
type StorageManager struct {
	imageGenerator *UbuntuImageGenerator
	client         *http.Client
	vmDir          string
	imagesDir      string // Base directory for storing images
}

// NewStorageManager creates a new storage manager instance
func NewStorageManager(vmDir string) *StorageManager {
	imagesDir := filepath.Join(vmDir, "Images")
	return &StorageManager{
		imageGenerator: NewUbuntuImageGenerator(),
		client:         &http.Client{}, // No timeout - let the queue system handle timeouts
		imagesDir:      imagesDir,
		vmDir:          vmDir,
	}
}

// GetOrDownloadUbuntuImage downloads an Ubuntu cloud image if it doesn't exist, or returns existing one
func (sm *StorageManager) GetOrDownloadUbuntuImage(ctx context.Context, version, arch string) (string, error) {
	// Create the organized directory structure
	imagePath := sm.getImagePath(version, arch)

	// Check if image already exists
	if err := sm.ValidateImageExists(imagePath); err == nil {
		storageLog.Infof("Ubuntu image already exists: %s", imagePath)
		return imagePath, nil
	}

	// Image doesn't exist, download it
	storageLog.Infof("Ubuntu image not found, downloading: version=%s, arch=%s", version, arch)
	return sm.DownloadUbuntuImage(ctx, version, arch)
}

// DownloadUbuntuImage downloads an Ubuntu cloud image to the organized directory structure
func (sm *StorageManager) DownloadUbuntuImage(ctx context.Context, version, arch string) (string, error) {
	storageLog.Infof("Starting Ubuntu image download for version %s, arch %s", version, arch)

	// Create the organized directory structure
	imageDir := sm.getImageDir(version)
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create image directory: %w", err)
	}

	// Generate image URL
	req := ImageRequest{
		OS:      "ubuntu",
		Version: version,
	}

	imageURL := sm.imageGenerator.GenerateImageURLWithFallback(req, arch)
	if imageURL == "" {
		return "", fmt.Errorf("failed to generate image URL for Ubuntu %s %s", version, arch)
	}

	storageLog.Infof("Generated image URL: %s", imageURL)

	// Determine output filename in organized structure
	outputFilename := sm.getImagePath(version, arch)

	// Download the image
	if err := sm.downloadFile(ctx, imageURL, outputFilename); err != nil {
		return "", fmt.Errorf("failed to download Ubuntu image: %w", err)
	}

	storageLog.Infof("Successfully downloaded Ubuntu image to: %s", outputFilename)
	return outputFilename, nil
}

// ConvertImageToVDI converts a downloaded .img file to .vdi format for VirtualBox
func (sm *StorageManager) ConvertImageToVDI(ctx context.Context, imagePath, vmName, vmFolder string) (string, error) {
	storageLog.Infof("Converting image to VDI: %s -> %s", imagePath, vmName)

	// Validate source image exists
	if err := sm.ValidateImageExists(imagePath); err != nil {
		return "", fmt.Errorf("source image validation failed: %w", err)
	}

	// Create VM folder if it doesn't exist
	if err := os.MkdirAll(vmFolder, 0755); err != nil {
		return "", fmt.Errorf("failed to create VM folder: %w", err)
	}

	// Create VDI filename in the VM folder
	vdiPath := filepath.Join(vmFolder, vmName+".vdi")

	// Check if VDI already exists
	if _, err := os.Stat(vdiPath); err == nil {
		storageLog.Infof("VDI file already exists: %s", vdiPath)
		return vdiPath, nil
	}

	// Execute qemu-img convert command
	cmd := exec.CommandContext(ctx, "qemu-img", "convert", "-f", "qcow2", "-O", "vdi", imagePath, vdiPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	storageLog.Infof("Executing: %s", cmd.String())
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to convert image to VDI: %w", err)
	}

	storageLog.Infof("Successfully converted image to VDI: %s", vdiPath)
	return vdiPath, nil
}

// GetOrCreateVDI gets an existing VDI or creates one from the image
func (sm *StorageManager) GetOrCreateVDI(ctx context.Context, version, arch, vmName string) (string, error) {
	vmFolder := filepath.Join(sm.vmDir, vmName)
	// First, ensure we have the image
	imagePath, err := sm.GetOrDownloadUbuntuImage(ctx, version, arch)
	if err != nil {
		return "", fmt.Errorf("failed to get/download image: %w", err)
	}

	// Create VDI filename in the VM folder
	vdiPath := filepath.Join(vmFolder, vmName+".vdi")

	// Check if VDI already exists
	if _, err := os.Stat(vdiPath); err == nil {
		storageLog.Infof("VDI file already exists: %s", vdiPath)
		return vdiPath, nil
	}

	// Convert image to VDI
	return sm.ConvertImageToVDI(ctx, imagePath, vmName, vmFolder)
}

// downloadFile downloads a file from URL to local path with progress tracking
func (sm *StorageManager) downloadFile(ctx context.Context, url, outputPath string) error {
	storageLog.Infof("Downloading file from %s to %s", url, outputPath)

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add User-Agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; SubnetNode/1.0)")

	// Execute request
	resp, err := sm.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, resp.Status)
	}

	// Get content length for progress tracking
	contentLength := resp.ContentLength
	storageLog.Infof("File size: %d bytes (%.2f MB)", contentLength, float64(contentLength)/(1024*1024))

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Download with progress tracking
	downloadedBytes := int64(0)
	lastProgressTime := time.Now()

	buffer := make([]byte, 32*1024) // 32KB buffer
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Read chunk
			n, err := resp.Body.Read(buffer)
			if n > 0 {
				// Write chunk to file
				if _, writeErr := outputFile.Write(buffer[:n]); writeErr != nil {
					return fmt.Errorf("failed to write to file: %w", writeErr)
				}
				downloadedBytes += int64(n)

				// Log progress every 10 seconds or every 10MB
				if time.Since(lastProgressTime) > 10*time.Second || downloadedBytes%(10*1024*1024) < int64(n) {
					if contentLength > 0 {
						progress := float64(downloadedBytes) / float64(contentLength) * 100
						storageLog.Infof("Download progress: %.1f%% (%d/%d bytes)",
							progress, downloadedBytes, contentLength)
					} else {
						storageLog.Infof("Downloaded: %d bytes", downloadedBytes)
					}
					lastProgressTime = time.Now()
				}
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("failed to read response body: %w", err)
			}
		}
	}
}

// GetAvailableUbuntuVersions returns a list of available Ubuntu versions
func (sm *StorageManager) GetAvailableUbuntuVersions() []string {
	versions := make([]string, 0, len(sm.imageGenerator.versionToCodename))
	for version := range sm.imageGenerator.versionToCodename {
		versions = append(versions, version)
	}
	return versions
}

// ValidateImageExists checks if an image file exists and is valid
func (sm *StorageManager) ValidateImageExists(imagePath string) error {
	// Check if file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return fmt.Errorf("image file does not exist: %s", imagePath)
	}

	// Check if file is readable
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Check if file has reasonable size (at least 100MB)
	if fileInfo.Size() < 100*1024*1024 {
		return fmt.Errorf("image file seems too small: %d bytes", fileInfo.Size())
	}

	storageLog.Infof("Image validation passed: %s (%d bytes)", imagePath, fileInfo.Size())
	return nil
}

// getImageDir returns the directory path for a specific Ubuntu version
func (sm *StorageManager) getImageDir(version string) string {
	return filepath.Join(sm.imagesDir, "Ubuntu", version)
}

// getImagePath returns the full path for an Ubuntu image file
func (sm *StorageManager) getImagePath(version, arch string) string {
	imageDir := sm.getImageDir(version)
	filename := fmt.Sprintf("ubuntu-%s-server-cloudimg-%s.img", version, arch)
	return filepath.Join(imageDir, filename)
}

// ListDownloadedImages returns a list of all downloaded images
func (sm *StorageManager) ListDownloadedImages() (map[string][]string, error) {
	result := make(map[string][]string)

	ubuntuDir := filepath.Join(sm.imagesDir, "Ubuntu")
	if _, err := os.Stat(ubuntuDir); os.IsNotExist(err) {
		return result, nil // No images downloaded yet
	}

	// Walk through the Ubuntu directory
	err := filepath.Walk(ubuntuDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".img" {
			// Extract version from path
			relPath, err := filepath.Rel(ubuntuDir, path)
			if err != nil {
				return err
			}

			parts := filepath.SplitList(relPath)
			if len(parts) >= 2 {
				version := parts[0]
				filename := filepath.Base(path)
				result[version] = append(result[version], filename)
			}
		}
		return nil
	})

	return result, err
}

func (sm *StorageManager) GenerateCloudInitISO(ctx context.Context, vmName, username, password string) (string, error) {
	storageLog.Infof("Generating cloud-init ISO for VM: %s", vmName)

	// Create VM folder if it doesn't exist
	vmFolder := filepath.Join(sm.vmDir, vmName)
	if err := os.MkdirAll(vmFolder, 0755); err != nil {
		return "", fmt.Errorf("failed to create VM folder: %w", err)
	}

	// Create temporary directory for cloud-init files
	tempDir, err := os.MkdirTemp("", "cloud-init-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	// Clean up temp directory and all sensitive files immediately after use
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			storageLog.Warnf("Failed to clean up temporary cloud-init files: %v", err)
		}
	}()

	// Generate user-data file
	userDataPath := filepath.Join(tempDir, "user-data")
	userData := fmt.Sprintf(`#cloud-config

users:
  - name: %s
    plain_text_passwd: %s
    lock_passwd: false
    groups: sudo
    shell: /bin/bash

chpasswd:
  list: |
    %s:%s
  expire: false
`, username, password, username, password)

	if err := os.WriteFile(userDataPath, []byte(userData), 0644); err != nil {
		return "", fmt.Errorf("failed to write user-data file: %w", err)
	}

	// Generate meta-data file
	metaDataPath := filepath.Join(tempDir, "meta-data")
	metaData := fmt.Sprintf(`instance-id: %s
local-hostname: %s
`, vmName, vmName)

	if err := os.WriteFile(metaDataPath, []byte(metaData), 0644); err != nil {
		return "", fmt.Errorf("failed to write meta-data file: %w", err)
	}

	// Generate ISO file
	isoPath := filepath.Join(vmFolder, "cloud-init.iso")

	// Try isogenimage first, then fallback to mkisofs
	var cmd *exec.Cmd
	if _, err := exec.LookPath("isogenimage"); err == nil {
		// Use isogenimage
		cmd = exec.CommandContext(ctx, "isogenimage", "-o", isoPath, "-V", "cidata", "-r", tempDir)
		storageLog.Infof("Using isogenimage to create cloud-init ISO")
	} else if _, err := exec.LookPath("mkisofs"); err == nil {
		// Use mkisofs
		cmd = exec.CommandContext(ctx, "mkisofs", "-o", isoPath, "-V", "cidata", "-r", "-J", tempDir)
		storageLog.Infof("Using mkisofs to create cloud-init ISO")
	} else {
		return "", fmt.Errorf("neither isogenimage nor mkisofs found in PATH")
	}

	// Execute the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create cloud-init ISO: %w, output: %s", err, string(output))
	}

	// Immediately clean up sensitive files after ISO creation
	if err := os.RemoveAll(tempDir); err != nil {
		storageLog.Warnf("Failed to clean up temporary cloud-init files: %v", err)
	}

	storageLog.Infof("Successfully generated cloud-init ISO: %s", isoPath)
	return isoPath, nil
}

// GetImageInfo returns information about a specific image
func (sm *StorageManager) GetImageInfo(version, arch string) (*ImageInfo, error) {
	imagePath := sm.getImagePath(version, arch)

	fileInfo, err := os.Stat(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get image info: %w", err)
	}

	return &ImageInfo{
		Path:         imagePath,
		Size:         fileInfo.Size(),
		ModifiedTime: fileInfo.ModTime(),
		Version:      version,
		Architecture: arch,
	}, nil
}

// ImageInfo contains information about a downloaded image
type ImageInfo struct {
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	ModifiedTime time.Time `json:"modified_time"`
	Version      string    `json:"version"`
	Architecture string    `json:"architecture"`
}
