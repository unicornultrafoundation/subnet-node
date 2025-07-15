package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemplateManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "template-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test template files
	userDataTemplate := `#cloud-config
hostname: {{.Hostname}}
username: {{.Username}}
password: {{.Password}}
`
	metaDataTemplate := `instance-id: {{.InstanceID}}
local-hostname: {{.Hostname}}
`

	// Write test templates
	userDataPath := filepath.Join(tempDir, "cloud-init-user-data.tmpl")
	if err := os.WriteFile(userDataPath, []byte(userDataTemplate), 0644); err != nil {
		t.Fatalf("Failed to write user-data template: %v", err)
	}

	metaDataPath := filepath.Join(tempDir, "cloud-init-meta-data.tmpl")
	if err := os.WriteFile(metaDataPath, []byte(metaDataTemplate), 0644); err != nil {
		t.Fatalf("Failed to write meta-data template: %v", err)
	}

	// Create template manager
	tm := NewTemplateManager(tempDir)

	// Test data
	data := CloudInitData{
		InstanceID: "test-vm-123",
		Hostname:   "test-vm",
		Username:   "ubuntu",
		Password:   "test-password",
	}

	// Test user-data generation
	userData, err := tm.GenerateUserData(data)
	if err != nil {
		t.Fatalf("Failed to generate user-data: %v", err)
	}

	// Verify user-data contains expected values
	if !strings.Contains(userData, "hostname: test-vm") {
		t.Error("User-data does not contain expected hostname")
	}
	if !strings.Contains(userData, "username: ubuntu") {
		t.Error("User-data does not contain expected username")
	}
	if !strings.Contains(userData, "password: test-password") {
		t.Error("User-data does not contain expected password")
	}

	// Test meta-data generation
	metaData, err := tm.GenerateMetaData(data)
	if err != nil {
		t.Fatalf("Failed to generate meta-data: %v", err)
	}

	// Verify meta-data contains expected values
	if !strings.Contains(metaData, "instance-id: test-vm-123") {
		t.Error("Meta-data does not contain expected instance-id")
	}
	if !strings.Contains(metaData, "local-hostname: test-vm") {
		t.Error("Meta-data does not contain expected hostname")
	}

	// Test template validation
	if err := tm.ValidateTemplates(); err != nil {
		t.Errorf("Template validation failed: %v", err)
	}

	t.Logf("User-data generated: %s", userData)
	t.Logf("Meta-data generated: %s", metaData)
}

func TestTemplateValidation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "template-validation-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tm := NewTemplateManager(tempDir)

	// Test validation with missing templates (should fail)
	if err := tm.ValidateTemplates(); err == nil {
		t.Error("Expected validation to fail with missing templates")
	}

	// Create only user-data template
	userDataPath := filepath.Join(tempDir, "cloud-init-user-data.tmpl")
	if err := os.WriteFile(userDataPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write user-data template: %v", err)
	}

	// Test validation with partial templates (should fail)
	if err := tm.ValidateTemplates(); err == nil {
		t.Error("Expected validation to fail with missing meta-data template")
	}

	// Create meta-data template
	metaDataPath := filepath.Join(tempDir, "cloud-init-meta-data.tmpl")
	if err := os.WriteFile(metaDataPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write meta-data template: %v", err)
	}

	// Test validation with all templates (should pass)
	if err := tm.ValidateTemplates(); err != nil {
		t.Errorf("Template validation should pass with all templates: %v", err)
	}
}
