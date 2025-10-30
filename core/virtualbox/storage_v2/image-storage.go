package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/cmd_exec"
)

var imageStorageLog = logrus.WithField("package", "imageStorage")

type ImageStorage struct {
	vmDir      string
	imageDir   string
	httpClient *http.Client
	QemuCmd    cmd_exec.Command
	IsogenCmd  cmd_exec.Command
}

func NewImageStorage(vmDir, imageDir string) *ImageStorage {
	return &ImageStorage{
		vmDir:      vmDir,
		imageDir:   imageDir,
		httpClient: &http.Client{},
		QemuCmd:    cmd_exec.GetQemuCmd(),
		IsogenCmd:  cmd_exec.GetIsogenCmd(),
	}
}

func (is *ImageStorage) GetOrCreateVDI(ctx context.Context, version, arch, vmName string) (string, error) {

	vmFolder := filepath.Join(is.vmDir, vmName)
	// First, ensure we have the image
	imagePath, err := is.GetOrDownloadUbuntuImage(ctx, version, arch)

	if err != nil {
		return "", fmt.Errorf("failed to get/download image: %w", err)
	}

	// Create VDI filename in the VM folder
	vdiPath := filepath.Join(vmFolder, vmName+".vdi")

	// Check if VDI already exists
	if _, err := os.Stat(vdiPath); err == nil {
		imageStorageLog.Infof("VDI file already exists: %s", vdiPath)
		return vdiPath, nil
	}

	// Convert image to VDI
	return is.ConvertImageToVDI(ctx, imagePath, vdiPath)
}

func (is *ImageStorage) GetOrDownloadUbuntuImage(ctx context.Context, version, arch string) (string, error) {
	// {imageDir}/Ubuntu/{version}/{version}-server-cloudimg-{arch}.img
	ubuntuVersionPath := filepath.Join(is.imageDir, "Ubuntu", version)
	imagePath := filepath.Join(ubuntuVersionPath, fmt.Sprintf("ubuntu-%s-server-cloudimg-%s.img", version, arch))

	// Check if image already exists
	if err := is.validateImageExists(imagePath); err == nil {
		return imagePath, nil
	}

	// Image doesn't exist, download it
	if err := os.MkdirAll(ubuntuVersionPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create image directory: %w", err)
	}

	imageURL := is.GenerateImageURLWithFallback(ImageRequest{
		OS:      "ubuntu",
		Version: version,
	})
	if imageURL == "" {
		return "", fmt.Errorf("failed to generate image URL for Ubuntu %s %s", version, arch)
	}

	// Download the image
	if err := is.downloadFile(ctx, imageURL, imagePath); err != nil {
		return "", fmt.Errorf("failed to download Ubuntu image: %w", err)

	}
	return imagePath, nil
}

// ConvertImageToVDI converts a downloaded .img file to .vdi format for VirtualBox
func (is *ImageStorage) ConvertImageToVDI(ctx context.Context, imagePath, vdiPath string) (string, error) {
	// Validate source image exists
	if err := is.validateImageExists(imagePath); err != nil {
		return "", fmt.Errorf("source image validation failed: %w", err)
	}

	fmt.Println("imagePath", imagePath)
	fmt.Println("vdiPath", vdiPath)

	// Execute qemu-img convert command
	args := []string{"convert", "-f", "qcow2", "-O", "vdi", imagePath, vdiPath}
	_, _, err := is.QemuCmd.Run(args...)
	if err != nil {
		return "", fmt.Errorf("failed to convert image to VDI: %w", err)
	}

	imageStorageLog.Infof("Successfully converted image to VDI: %s", vdiPath)
	return vdiPath, nil
}

// ValidateImageExists checks if an image file exists and is valid
func (is *ImageStorage) validateImageExists(imagePath string) error {
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

	return nil
}

// downloadFile downloads a file from URL to local path with progress tracking
func (is *ImageStorage) downloadFile(ctx context.Context, url, outputPath string) error {
	imageStorageLog.Infof("Downloading file from %s to %s", url, outputPath)

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add User-Agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; SubnetNode/1.0)")

	// Execute request
	resp, err := is.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, resp.Status)
	}

	// Get content length for progress tracking
	contentLength := resp.ContentLength
	imageStorageLog.Infof("File size: %d bytes (%.2f MB)", contentLength, float64(contentLength)/(1024*1024))

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

				// Log progress every 10 seconds
				if time.Since(lastProgressTime) > 10*time.Second {
					if contentLength > 0 {
						progress := float64(downloadedBytes) / float64(contentLength) * 100
						imageStorageLog.Infof("Download progress: %.1f%% (%d/%d bytes)",
							progress, downloadedBytes, contentLength)
					} else {
						imageStorageLog.Infof("Downloaded: %d bytes", downloadedBytes)
					}
					lastProgressTime = time.Now()
				}
				// print when file is downloaded
				if downloadedBytes == contentLength {
					imageStorageLog.Infof("File downloaded: %d bytes", downloadedBytes)
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

// Cloudinit functions
func (is *ImageStorage) GenerateCloudInitISO(ctx context.Context, vmName, username, password string) (string, error) {
	vmFolder := filepath.Join(is.vmDir, vmName)

	// Create temporary directory for cloud-init files
	tempDir, err := os.MkdirTemp("", "cloud-init-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	// Clean up temp directory and all sensitive files immediately after use
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			imageStorageLog.Warnf("Failed to clean up temporary cloud-init files: %v", err)
		}
	}()

	// Generate user-data file
	userDataPath := filepath.Join(tempDir, "user-data")
	userData := fmt.Sprintf(userDataTemplate, username, password, username, password)
	if err := os.WriteFile(userDataPath, []byte(userData), 0644); err != nil {
		return "", fmt.Errorf("failed to write user-data file: %w", err)
	}

	// Generate meta-data file
	metaDataPath := filepath.Join(tempDir, "meta-data")
	metaData := fmt.Sprintf(metaDataTemplate, vmName, vmName)
	if err := os.WriteFile(metaDataPath, []byte(metaData), 0644); err != nil {
		return "", fmt.Errorf("failed to write meta-data file: %w", err)
	}

	// Generate ISO file
	isoPath := filepath.Join(vmFolder, "cloud-init.iso")

	// Execute isogenimage command
	args := []string{"-o", isoPath, "-V", "cidata", "-r", tempDir}
	stdout, _, err := is.IsogenCmd.Run(args...)
	if err != nil {
		return "", fmt.Errorf("failed to generate cloud-init ISO: %w", err)
	}
	if stdout != "" {
		imageStorageLog.Infof("Isogenimage output: %s", stdout)
	}

	// Immediately clean up sensitive files (including user data) after ISO creation

	if err := os.RemoveAll(tempDir); err != nil {
		imageStorageLog.Warnf("Failed to clean up temporary cloud-init files: %v", err)
	}

	return isoPath, nil
}
