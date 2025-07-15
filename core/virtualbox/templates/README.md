# Cloud-Init Templates

This directory contains template files for generating cloud-init configuration for VirtualBox VMs.

## Template Files

### `cloud-init-user-data.tmpl`

The main cloud-init configuration file that defines:

- Ubuntu autoinstaller configuration
- User account setup
- SSH configuration
- Package installation
- System configuration

### `cloud-init-meta-data.tmpl`

Metadata file containing:

- Instance ID
- Hostname

## Template Variables

The templates use Go template syntax with the following variables:

- `{{.InstanceID}}` - Unique identifier for the VM
- `{{.Hostname}}` - Hostname for the VM
- `{{.Username}}` - Username for the VM account
- `{{.Password}}` - Password for the VM account

## Customization

You can customize the templates by editing the .tmpl files:

### Adding Packages

To install additional packages, add them to the `packages` section in `cloud-init-user-data.tmpl`:

```yaml
packages:
  - openssh-server
  - curl
  - wget
  - htop # Add your packages here
  - vim
  - git
```

### Adding Commands

To run additional commands during setup, add them to the `runcmd` section:

```yaml
runcmd:
  - [eval, 'echo $(cat /proc/cmdline) "autoinstall" > /root/cmdline']
  - [eval, "mount -n --bind -o ro /root/cmdline /proc/cmdline"]
  - [eval, "snap restart subiquity.subiquity-server"]
  - [eval, "snap restart subiquity.subiquity-service"]
  - echo "Custom command executed" # Add your commands here
  - apt-get update
```

### Modifying User Configuration

To change user settings, modify the `identity` and `user-data.users` sections:

```yaml
identity:
  hostname: { { .Hostname } }
  username: { { .Username } }
  password: { { .Password } }

user-data:
  users:
    - name: { { .Username } }
      sudo: ALL=(ALL) NOPASSWD:ALL
      shell: /bin/bash
      # Add additional user configuration here
```

## Template Processing

The templates are processed by the `TemplateManager` in `template.go`, which:

1. Reads the template files
2. Substitutes variables with VM-specific values
3. Generates the final cloud-init files

## File Structure

```
templates/
├── README.md                    # This file
├── template.go                  # Template processing logic
├── template_test.go             # Unit tests
├── cloud-init-user-data.tmpl    # User data template
└── cloud-init-meta-data.tmpl    # Meta data template
```

## Usage

Templates are automatically used when creating VMs. Each VM gets its own cloud-init configuration based on these templates with VM-specific values substituted.

## Validation

The system validates that all required template files exist during service initialization. If any template files are missing, the service will fail to start with a clear error message.
