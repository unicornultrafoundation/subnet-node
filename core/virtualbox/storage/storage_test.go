package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStorageManager_GetOrDownloadUbuntuImage(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage manager
	sm := NewStorageManager(tempDir)

	// Test downloading Ubuntu 22.04 image
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	imagePath, err := sm.GetOrDownloadUbuntuImage(ctx, "22.04", "amd64")
	if err != nil {
		t.Skipf("Download test skipped (likely network issue): %v", err)
	}

	// Validate the downloaded image
	if err := sm.validateImageExists(imagePath); err != nil {
		t.Errorf("Image validation failed: %v", err)
	}

	// Check if file exists and has reasonable size
	fileInfo, err := os.Stat(imagePath)
	if err != nil {
		t.Errorf("Failed to get file info: %v", err)
	}

	if fileInfo.Size() < 100*1024*1024 { // Less than 100MB
		t.Errorf("Downloaded image seems too small: %d bytes", fileInfo.Size())
	}

	// Test that the image is in the correct organized directory structure
	expectedPath := filepath.Join(tempDir, "Images", "Ubuntu", "22.04", "ubuntu-22.04-server-cloudimg-amd64.img")
	if imagePath != expectedPath {
		t.Errorf("Image path mismatch. Expected: %s, Got: %s", expectedPath, imagePath)
	}

	t.Logf("Successfully downloaded Ubuntu image: %s (%d bytes)", imagePath, fileInfo.Size())

	// Test that calling GetOrDownloadUbuntuImage again returns the existing image
	imagePath2, err := sm.GetOrDownloadUbuntuImage(ctx, "22.04", "amd64")
	if err != nil {
		t.Errorf("Failed to get existing image: %v", err)
	}

	if imagePath != imagePath2 {
		t.Errorf("Expected same image path for existing image. Expected: %s, Got: %s", imagePath, imagePath2)
	}
}

func TestStorageManager_GetAvailableUbuntuVersions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm := NewStorageManager(tempDir)
	versions := sm.GetAvailableUbuntuVersions()

	if len(versions) == 0 {
		t.Error("No Ubuntu versions returned")
	}

	// Check if common versions are included
	expectedVersions := []string{"22.04", "20.04", "18.04"}
	for _, expected := range expectedVersions {
		found := false
		for _, version := range versions {
			if version == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected version %s not found in available versions", expected)
		}
	}

	t.Logf("Available Ubuntu versions: %v", versions)
}

func TestStorageManager_ValidateImageExists(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm := NewStorageManager(tempDir)

	// Test with non-existent file
	err = sm.validateImageExists("/path/to/nonexistent/file.img")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test with existing file (create a dummy file)
	tempFile, err := os.CreateTemp("", "test-image-*.img")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write some data to make it a reasonable size
	dummyData := make([]byte, 200*1024*1024) // 200MB
	_, err = tempFile.Write(dummyData)
	if err != nil {
		t.Fatalf("Failed to write dummy data: %v", err)
	}
	tempFile.Close()

	// Test validation
	if err := sm.validateImageExists(tempFile.Name()); err != nil {
		t.Errorf("Image validation failed for valid file: %v", err)
	}
}

func TestStorageManager_GetImagePath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm := NewStorageManager(tempDir)

	// Test image path generation
	expectedPath := filepath.Join(tempDir, "Images", "Ubuntu", "22.04", "ubuntu-22.04-server-cloudimg-amd64.img")
	actualPath := sm.getImagePath("22.04", "amd64")

	if expectedPath != actualPath {
		t.Errorf("Image path mismatch. Expected: %s, Got: %s", expectedPath, actualPath)
	}

	t.Logf("Image path generated correctly: %s", actualPath)
}

func TestStorageManager_ListDownloadedImages(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm := NewStorageManager(tempDir)

	// Create some dummy image files in the organized structure
	imageDir := filepath.Join(tempDir, "Images", "Ubuntu", "22.04")
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		t.Fatalf("Failed to create image directory: %v", err)
	}

	// Create dummy image files
	dummyImages := []string{
		"ubuntu-22.04-server-cloudimg-amd64.img",
		"ubuntu-22.04-server-cloudimg-arm64.img",
	}

	for _, imageName := range dummyImages {
		imagePath := filepath.Join(imageDir, imageName)
		file, err := os.Create(imagePath)
		if err != nil {
			t.Fatalf("Failed to create dummy image %s: %v", imageName, err)
		}
		file.Close()
	}

	// Test listing downloaded images
	images, err := sm.ListDownloadedImages()
	if err != nil {
		t.Errorf("Failed to list downloaded images: %v", err)
	}

	if len(images) == 0 {
		t.Error("No images found in listing")
	}

	if images["22.04"] == nil {
		t.Error("Expected images for version 22.04")
	}

	if len(images["22.04"]) != 2 {
		t.Errorf("Expected 2 images for version 22.04, got %d", len(images["22.04"]))
	}

	t.Logf("Listed downloaded images: %v", images)
}

func TestStorageManager_GetImageInfo(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm := NewStorageManager(tempDir)

	// Create a dummy image file
	imageDir := filepath.Join(tempDir, "Images", "Ubuntu", "22.04")
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		t.Fatalf("Failed to create image directory: %v", err)
	}

	imagePath := filepath.Join(imageDir, "ubuntu-22.04-server-cloudimg-amd64.img")
	file, err := os.Create(imagePath)
	if err != nil {
		t.Fatalf("Failed to create dummy image: %v", err)
	}
	file.Close()

	// Test getting image info
	imageInfo, err := sm.GetImageInfo("22.04", "amd64")
	if err != nil {
		t.Errorf("Failed to get image info: %v", err)
	}

	if imageInfo.Path != imagePath {
		t.Errorf("Image path mismatch. Expected: %s, Got: %s", imagePath, imageInfo.Path)
	}

	if imageInfo.Version != "22.04" {
		t.Errorf("Version mismatch. Expected: 22.04, Got: %s", imageInfo.Version)
	}

	if imageInfo.Architecture != "amd64" {
		t.Errorf("Architecture mismatch. Expected: amd64, Got: %s", imageInfo.Architecture)
	}

	t.Logf("Image info retrieved successfully: %+v", imageInfo)
}

func TestStorageManager_GenerateCloudInitISO(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cloud-init-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm := NewStorageManager(tempDir)

	// Test cloud-init ISO generation
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	vmName := "test-vm"
	username := "ubuntu"
	password := "password123"

	isoPath, err := sm.GenerateCloudInitISO(ctx, vmName, username, password)
	if err != nil {
		// Check if it's because the tools are not available
		if strings.Contains(err.Error(), "neither isogenimage nor mkisofs found") {
			t.Skipf("Skipping test - ISO creation tools not available: %v", err)
		}
		t.Errorf("Failed to generate cloud-init ISO: %v", err)
		return
	}

	// Check if ISO file exists
	if _, err := os.Stat(isoPath); os.IsNotExist(err) {
		t.Errorf("Generated ISO file does not exist: %s", isoPath)
	}

	// Check if ISO file has reasonable size (at least 1KB)
	fileInfo, err := os.Stat(isoPath)
	if err != nil {
		t.Errorf("Failed to get ISO file info: %v", err)
	} else if fileInfo.Size() < 1024 {
		t.Errorf("Generated ISO file seems too small: %d bytes", fileInfo.Size())
	}

	t.Logf("Successfully generated cloud-init ISO: %s (%d bytes)", isoPath, fileInfo.Size())
}

// Example usage of the StorageManager
func ExampleStorageManager_GetOrDownloadUbuntuImage() {
	// Create a storage manager
	vmDir := "/path/to/virtualbox/directory"
	sm := NewStorageManager(vmDir)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Get or download Ubuntu 22.04 image
	imagePath, err := sm.GetOrDownloadUbuntuImage(ctx, "22.04", "amd64")
	if err != nil {
		// Handle error
		return
	}

	// Validate the downloaded image
	if err := sm.validateImageExists(imagePath); err != nil {
		// Handle validation error
		return
	}

	// Use the image path for VM creation
	_ = imagePath
}

// Example usage of VDI conversion
func ExampleStorageManager_GetOrCreateVDI() {
	// Create a storage manager
	vmDir := "/path/to/virtualbox/directory"
	sm := NewStorageManager(vmDir)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	// Get or create VDI for a specific VM
	vmName := "my-ubuntu-vm"
	vdiPath, err := sm.GetOrCreateVDI(ctx, "22.04", "amd64", vmName)
	if err != nil {
		// Handle error
		return
	}

	// Use the VDI path for VirtualBox VM creation
	_ = vdiPath
}

// Example usage of cloud-init ISO generation
func ExampleStorageManager_GenerateCloudInitISO() {
	// Create a storage manager
	vmDir := "/path/to/virtualbox/directory"
	sm := NewStorageManager(vmDir)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Generate cloud-init ISO for a VM
	vmName := "my-ubuntu-vm"
	username := "ubuntu"
	password := "password123"

	isoPath, err := sm.GenerateCloudInitISO(ctx, vmName, username, password)
	if err != nil {
		// Handle error
		return
	}

	// Use the ISO path for VirtualBox VM creation
	_ = isoPath
}
