package templates

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/sirupsen/logrus"
)

var templateLog = logrus.WithField("component", "cloudinit-templates")

// CloudInitData represents the data to be used in cloud-init templates
type CloudInitData struct {
	InstanceID string
	Hostname   string
	Username   string
	Password   string
}

// TemplateManager handles cloud-init template processing
type TemplateManager struct {
	templateDir string
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(templateDir string) *TemplateManager {
	return &TemplateManager{
		templateDir: templateDir,
	}
}

// GenerateUserData generates user-data content from template
func (tm *TemplateManager) GenerateUserData(data CloudInitData) (string, error) {
	templateLog.Infof("Generating user-data for VM: %s", data.Hostname)

	templatePath := filepath.Join(tm.templateDir, "cloud-init-user-data-template-vm.tmpl")
	content, err := tm.processTemplate(templatePath, data)
	if err != nil {
		return "", fmt.Errorf("failed to process user-data template: %w", err)
	}

	templateLog.Infof("Successfully generated user-data for VM: %s", data.Hostname)
	return content, nil
}

// GenerateMetaData generates meta-data content from template
func (tm *TemplateManager) GenerateMetaData(data CloudInitData) (string, error) {
	templateLog.Infof("Generating meta-data for VM: %s", data.Hostname)

	templatePath := filepath.Join(tm.templateDir, "cloud-init-meta-data-template-vm.tmpl")
	content, err := tm.processTemplate(templatePath, data)
	if err != nil {
		return "", fmt.Errorf("failed to process meta-data template: %w", err)
	}

	templateLog.Infof("Successfully generated meta-data for VM: %s", data.Hostname)
	return content, nil
}

// GenerateCloneVMUserData generates user-data content from clone VM template
func (tm *TemplateManager) GenerateCloneVMUserData(data CloudInitData) (string, error) {
	templateLog.Infof("Generating clone VM user-data for VM: %s", data.Hostname)

	templatePath := filepath.Join(tm.templateDir, "cloud-init-user-data-clone-vm.tmpl")
	content, err := tm.processTemplate(templatePath, data)
	if err != nil {
		return "", fmt.Errorf("failed to process clone VM user-data template: %w", err)
	}

	templateLog.Infof("Successfully generated clone VM user-data for VM: %s", data.Hostname)
	return content, nil
}

// GenerateCloneVMMetaData generates meta-data content from clone VM template
func (tm *TemplateManager) GenerateCloneVMMetaData(data CloudInitData) (string, error) {
	templateLog.Infof("Generating clone VM meta-data for VM: %s", data.Hostname)

	templatePath := filepath.Join(tm.templateDir, "cloud-init-meta-data-clone-vm.tmpl")
	content, err := tm.processTemplate(templatePath, data)
	if err != nil {
		return "", fmt.Errorf("failed to process clone VM meta-data template: %w", err)
	}

	templateLog.Infof("Successfully generated clone VM meta-data for VM: %s", data.Hostname)
	return content, nil
}

// processTemplate processes a template file with the given data
func (tm *TemplateManager) processTemplate(templatePath string, data CloudInitData) (string, error) {
	// Read template file
	templateContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	// Parse template
	tmpl, err := template.New(filepath.Base(templatePath)).Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templatePath, err)
	}

	return buf.String(), nil
}

// ValidateTemplates checks if all required template files exist
func (tm *TemplateManager) ValidateTemplates() error {
	requiredTemplates := []string{
		"cloud-init-user-data-template-vm.tmpl",
		"cloud-init-meta-data-template-vm.tmpl",
		"cloud-init-user-data-clone-vm.tmpl",
		"cloud-init-meta-data-clone-vm.tmpl",
	}

	for _, templateName := range requiredTemplates {
		templatePath := filepath.Join(tm.templateDir, templateName)
		if _, err := ioutil.ReadFile(templatePath); err != nil {
			return fmt.Errorf("required template file not found: %s", templatePath)
		}
	}

	templateLog.Info("All required template files validated successfully")
	return nil
}
