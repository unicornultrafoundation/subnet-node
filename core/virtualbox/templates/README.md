# Cloud-Init Templates

This directory contains cloud-init templates for VirtualBox VM creation and cloning.

## Template Files

### Regular VM Templates

- `cloud-init-user-data.tmpl` - User data template for regular VM creation
- `cloud-init-meta-data.tmpl` - Meta data template for regular VM creation

### Clone VM Templates

- `cloud-init-user-data-clone-vm.tmpl` - User data template for clone VM creation
- `cloud-init-meta-data-clone-vm.tmpl` - Meta data template for clone VM creation

## Template Data Structure

All templates use the `CloudInitData` structure:

```go
type CloudInitData struct {
    InstanceID string
    Hostname   string
    Username   string
    Password   string
}
```

## Usage

### Regular VM Creation

Regular VMs use the standard templates and are created with full OS installation:

```go
// Generate cloud-init files for regular VM
_, _, cloudInitDir, err := vboxExec.GenerateCloudInitFiles(vmName, hostname, username, password)
```

### Clone VM Creation

Clone VMs use the clone-specific templates and are created by cloning from a template VM:

```go
// Generate cloud-init files for clone VM
_, _, cloudInitDir, err := vboxExec.GenerateCloneVMCloudInitFiles(vmName, hostname, username, password)
```

## Clone VM Workflow

1. **Template Creation**: Create a template VM named `template_sample` with ubuntu/ubuntu credentials
2. **Clone Creation**: Use `CreateAndStartVM` to clone from template with custom credentials
3. **Cloud-Init**: Clone VMs use the clone-specific templates for faster setup

### Example

```go
// Create template VM (one-time setup)
templateReq := vbtypes.VMCreateRequest{
    Name:       "template_sample",
    CPUCores:   2,
    MemoryMB:   2048,
    DiskSizeGB: 20,
    OSType:     "Ubuntu_ARM64",
    Username:   "ubuntu",
    Password:   "ubuntu",
}
templateVM, err := vboxService.CreateVM(ctx, templateReq)

// Create clone VM with custom credentials
cloneReq := vbtypes.VMCreateRequest{
    Name:       "my-cloned-vm",
    CPUCores:   2,
    MemoryMB:   4096,
    DiskSizeGB: 30,
    OSType:     "Ubuntu_ARM64",
    Username:   "admin",
    Password:   "mypassword123",
}
clonedVM, err := vboxService.CreateAndStartVM(ctx, cloneReq)
```

## Template Differences

### Regular VM Templates

- Full cloud-init configuration
- Complete user setup
- Network configuration
- Package installation

### Clone VM Templates

- Simplified user setup (since OS is already installed)
- Focus on user credentials and basic configuration
- Faster initialization
- Optimized for cloned environments

## Validation

The template manager validates all required templates on service startup:

```go
func (tm *TemplateManager) ValidateTemplates() error {
    requiredTemplates := []string{
        "cloud-init-user-data.tmpl",
        "cloud-init-meta-data.tmpl",
        "cloud-init-user-data-clone-vm.tmpl",
        "cloud-init-meta-data-clone-vm.tmpl",
    }
    // ... validation logic
}
```

## Testing

Use the test script to verify template functionality:

```bash
cd core/virtualbox/script
go run vm_template.go
```

This script demonstrates:

1. Creating a template VM
2. Cloning a new VM from the template
3. Using the new clone VM cloud-init templates
