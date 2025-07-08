package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// PathDetector provides utilities to detect VirtualBox-related paths
type PathDetector struct{}

// NewPathDetector creates a new path detector
func NewPathDetector() *PathDetector {
	return &PathDetector{}
}

// DetectVBoxManagePath detects the path to VBoxManage executable
func (pd *PathDetector) DetectVBoxManagePath() string {
	// Common paths for VBoxManage on different platforms
	commonPaths := pd.getCommonVBoxManagePaths()

	// First check if VBoxManage is in PATH
	if path, err := exec.LookPath("VBoxManage"); err == nil {
		return path
	}

	// Check common installation paths
	for _, path := range commonPaths {
		if pd.isExecutable(path) {
			return path
		}
	}

	// Fallback to just "VBoxManage" and let the system handle it
	return "VBoxManage"
}

// DetectDefaultVMPath detects the default VM storage path
func (pd *PathDetector) DetectDefaultVMPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// VirtualBox default VM directory
	vmPath := filepath.Join(homeDir, "VirtualBox VMs")

	// Check if the directory exists or can be created
	if pd.isDirWritable(vmPath) {
		return vmPath
	}

	// Fallback to a subdirectory in home
	fallbackPath := filepath.Join(homeDir, ".subnet-node", "vms")
	if pd.isDirWritable(fallbackPath) {
		return fallbackPath
	}

	return ""
}

// DetectTerraformPath detects the path to terraform executable
func (pd *PathDetector) DetectTerraformPath() string {
	// Check if terraform is in PATH
	if path, err := exec.LookPath("terraform"); err == nil {
		return path
	}

	// Common installation paths for terraform
	commonPaths := pd.getCommonTerraformPaths()
	for _, path := range commonPaths {
		if pd.isExecutable(path) {
			return path
		}
	}

	// Fallback to just "terraform" and let the system handle it
	return "terraform"
}

// DetectTerraformWorkDir detects the default terraform working directory
func (pd *PathDetector) DetectTerraformWorkDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	workDir := filepath.Join(homeDir, ".subnet-node", "terraform")
	if pd.isDirWritable(workDir) {
		return workDir
	}

	return ""
}

// DetectBaseImagePath detects the default base image path
func (pd *PathDetector) DetectBaseImagePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	imagePath := filepath.Join(homeDir, ".subnet-node", "images")
	if pd.isDirWritable(imagePath) {
		return imagePath
	}

	return ""
}

// DetectSnapshotPath detects the default snapshot path
func (pd *PathDetector) DetectSnapshotPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	snapshotPath := filepath.Join(homeDir, ".subnet-node", "snapshots")
	if pd.isDirWritable(snapshotPath) {
		return snapshotPath
	}

	return ""
}

// getCommonVBoxManagePaths returns common VBoxManage installation paths
func (pd *PathDetector) getCommonVBoxManagePaths() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/Applications/VirtualBox.app/Contents/MacOS/VBoxManage",
			"/usr/local/bin/VBoxManage",
			"/opt/homebrew/bin/VBoxManage",
		}
	case "linux":
		return []string{
			"/usr/bin/VBoxManage",
			"/usr/local/bin/VBoxManage",
			"/opt/VirtualBox/VBoxManage",
		}
	case "windows":
		return []string{
			"C:\\Program Files\\Oracle\\VirtualBox\\VBoxManage.exe",
			"C:\\Program Files (x86)\\Oracle\\VirtualBox\\VBoxManage.exe",
		}
	default:
		return []string{}
	}
}

// getCommonTerraformPaths returns common terraform installation paths
func (pd *PathDetector) getCommonTerraformPaths() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/usr/local/bin/terraform",
			"/opt/homebrew/bin/terraform",
		}
	case "linux":
		return []string{
			"/usr/local/bin/terraform",
			"/usr/bin/terraform",
		}
	case "windows":
		return []string{
			"C:\\terraform\\terraform.exe",
		}
	default:
		return []string{}
	}
}

// isExecutable checks if a file exists and is executable
func (pd *PathDetector) isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return false
	}

	// On Unix-like systems, check if it's executable
	if runtime.GOOS != "windows" {
		return info.Mode()&0111 != 0
	}

	// On Windows, just check if it exists and has .exe extension
	return strings.HasSuffix(strings.ToLower(path), ".exe")
}

// isDirWritable checks if a directory exists and is writable, or can be created
func (pd *PathDetector) isDirWritable(path string) bool {
	// Check if directory exists
	if info, err := os.Stat(path); err == nil {
		return info.IsDir() && pd.isWritable(path)
	}

	// Try to create the directory
	if err := os.MkdirAll(path, 0755); err == nil {
		return true
	}

	return false
}

// isWritable checks if a path is writable
func (pd *PathDetector) isWritable(path string) bool {
	// Try to create a temporary file to test write permissions
	testFile := filepath.Join(path, ".write_test")
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	defer func() {
		file.Close()
		os.Remove(testFile)
	}()

	return true
}
