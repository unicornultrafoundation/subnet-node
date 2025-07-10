package virtualbox

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

var storageLog = logrus.WithField("service", "virtualbox-storage")

// StorageManagerImpl implements the StorageManager interface
type StorageManagerImpl struct {
	isoDir string
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
