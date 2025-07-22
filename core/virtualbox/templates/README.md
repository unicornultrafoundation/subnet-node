# Cloud-Init Templates for VirtualBox VMs

This directory contains **cloud-init templates** used for initializing VirtualBox virtual machines (VMs) in two main scenarios:

- **Template VM Creation** (full OS installation)
- **Clone VM Creation** (cloning from a pre-installed template VM)

These templates are used to generate the `user-data` and `meta-data` files required by cloud-init, which are then packaged into an ISO file and attached to the VM for automated provisioning.

---

## Template Types & Files

### 1. Template VM (Full OS Install)

- **User Data Template:** `cloud-init-user-data-template-vm.tmpl`
- **Meta Data Template:** `cloud-init-meta-data-template-vm.tmpl`

**Purpose:**

- Used when creating a new VM from scratch (with full OS installation).
- Provides full cloud-init configuration, including user setup, network, and package installation.

**Example structure:**

```yaml
#cloud-config
runcmd:
  - [eval, 'echo $(cat /proc/cmdline) "autoinstall" > /root/cmdline']
  ...
autoinstall:
  identity:
    hostname: {{.Hostname}}
    username: {{.Username}}
    password: {{.Password}}
  ...
user-data:
  users:
    - name: {{.Username}}
      sudo: ALL=(ALL) NOPASSWD:ALL
      shell: /bin/bash
```

**Meta Data:**

```yaml
instance-id: { { .InstanceID } }
local-hostname: { { .Hostname } }
```

### 2. Clone VM (From Template VM)

- **User Data Template:** `cloud-init-user-data-clone-vm.tmpl`
- **Meta Data Template:** `cloud-init-meta-data-clone-vm.tmpl`

**Purpose:**

- Used when creating a VM by cloning from a pre-installed template VM.
- Provides minimal configuration, focusing on user credentials and basic setup for faster initialization.

**Example structure:**

```yaml
#cloud-config
users:
  - name: { { .Username } }
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: sudo
    shell: /bin/bash
    lock_passwd: false
    passwd: { { .Password } }
ssh_pwauth: true
```

**Meta Data:**

```yaml
instance-id: { { .InstanceID } }
local-hostname: { { .Hostname } }
```

---

## Template Data Structure

All templates use the following Go struct for data injection:

```go
type CloudInitData struct {
    InstanceID string
    Hostname   string
    Username   string
    Password   string
}
```

---

## How to Generate and Build the Cloud-Init ISO

The process for generating the cloud-init ISO (to be attached to a VM) is as follows:

1. **Generate cloud-init files from templates:**

   - For a template VM:
     ```go
     _, _, cloudInitDir, err := vboxExec.GenerateCloudInitFiles(vmName, hostname, username, password)
     ```
   - For a clone VM:
     ```go
     _, _, cloudInitDir, err := vboxExec.GenerateCloneVMCloudInitFiles(vmName, hostname, username, password)
     ```
   - This will create `user-data` and `meta-data` files in a directory like `<vmDir>/<vmName>/cloud-init/`.

2. **Build the ISO file:**

   - The ISO is created from the generated `user-data` and `meta-data` files using the following logic (see `GenerateCloudInitISO` in `vbox_manage.go`):
     ```go
     isoPath, err := vboxExec.GenerateCloudInitISO(cloudInitDir)
     ```
   - This runs (pseudocode):
     ```sh
     mkisofs -o cloud-init.iso -V cidata -r -J user-data meta-data
     # or, as fallback:
     genisoimage -output cloud-init.iso -volid cidata -joliet -rock user-data meta-data
     ```
   - The resulting `cloud-init.iso` is placed in the same directory.

3. **Attach the ISO to the VM:**
   - The ISO is attached as a DVD drive to the VM for use by cloud-init at boot.

---

## Template Validation

The system validates the presence of all required templates on startup:

- `cloud-init-user-data-template-vm.tmpl`
- `cloud-init-meta-data-template-vm.tmpl`
- `cloud-init-user-data-clone-vm.tmpl`
- `cloud-init-meta-data-clone-vm.tmpl`

---

## Example Workflow

### Creating a Template VM (one-time setup)

```go
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
```

### Creating a Clone VM (from template)

```go
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

---

## Testing

To test template functionality:

```sh
cd core/virtualbox/script
go run vm_template.go
```

This script demonstrates:

- Creating a template VM
- Cloning a new VM from the template
- Using the correct cloud-init templates for each case
