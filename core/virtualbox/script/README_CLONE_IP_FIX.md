# Clone VM IP Address Fix

## Problem

When cloning VMs using VirtualBox's `clonevm` command, the cloned VM inherits the exact same network configuration as the original template VM, including the MAC address. This causes IP address conflicts because multiple VMs end up with the same MAC address, leading to network issues.

## Solution

The fix implements the following changes:

### 1. Unique MAC Address Generation

- **File**: `core/virtualbox/vbox_manage.go`
- **Function**: `generateUniqueMACAddress()`
- **Purpose**: Generates a unique MAC address in the VirtualBox range `08:00:27:xx:xx:xx`

### 2. MAC Address Update After Cloning

- **File**: `core/virtualbox/vbox_manage.go`
- **Function**: `CloneTemplateVM()` (enhanced)
- **Purpose**: After cloning, updates the network adapter with a new unique MAC address

### 3. Enhanced Cloud-Init Templates

- **Files**:
  - `core/virtualbox/templates/cloud-init-meta-data-clone-vm.tmpl`
  - `core/virtualbox/templates/cloud-init-user-data-clone-vm.tmpl`
- **Purpose**: Include proper network configuration to ensure each VM gets a unique IP address

### 4. IP Address Retrieval

- **File**: `core/virtualbox/vbox_manage.go`
- **Functions**:
  - `GetVMIPAddress()` - Gets current IP address
  - `WaitForVMIPAddress()` - Waits for VM to get IP address
- **Purpose**: Retrieve and verify IP addresses for cloned VMs

### 5. Service Integration

- **File**: `core/virtualbox/service.go`
- **Function**: `CreateAndStartVM()` (enhanced)
- **Purpose**: Waits for and stores IP addresses after VM creation

## How It Works

1. **Clone Creation**: When `CreateAndStartVM` is called, it clones the template VM
2. **MAC Address Assignment**: A unique MAC address is generated and assigned to the cloned VM
3. **Cloud-Init Setup**: New cloud-init files are generated with proper network configuration
4. **VM Start**: The VM is started with the new network configuration
5. **IP Address Wait**: The service waits for the VM to get an IP address
6. **Verification**: The IP address is retrieved and stored in the VM metadata

## Testing the Fix

### Prerequisites

1. VirtualBox must be installed
2. A template VM named `template_sample` must exist
3. The template VM should be properly configured with Ubuntu

### Test Steps

1. **Create Template VM** (if not exists):

   ```bash
   cd core/virtualbox/script
   go run vm_template.go
   ```

2. **Test Multiple Clones**:

   ```bash
   # Use the API to create multiple cloned VMs
   # Each should get a unique IP address
   ```

3. **Verify IP Uniqueness**:
   - Check that each cloned VM has a different IP address
   - Verify no IP conflicts in the network

### Expected Results

- Each cloned VM should have a unique MAC address
- Each cloned VM should get a unique IP address from DHCP
- No network conflicts between cloned VMs
- All VMs should be able to communicate on the network

## Code Changes Summary

### New Functions Added

1. `generateUniqueMACAddress()` - Generates unique MAC addresses
2. `UpdateNetworkAdapterMAC()` - Updates VM MAC address
3. `GetVMIPAddress()` - Retrieves VM IP address
4. `WaitForVMIPAddress()` - Waits for IP address assignment

### Enhanced Functions

1. `CloneTemplateVM()` - Now assigns unique MAC addresses
2. `CreateAndStartVM()` - Now waits for and stores IP addresses
3. `GetVM()` - Now refreshes IP addresses for running VMs

### Template Updates

1. **meta-data template**: Added network configuration
2. **user-data template**: Added network setup and restart commands

## Benefits

- **No IP Conflicts**: Each cloned VM gets a unique IP address
- **Network Stability**: Eliminates network issues caused by duplicate MAC addresses
- **Better Monitoring**: IP addresses are tracked and stored
- **Improved Reliability**: VMs can coexist on the same network without conflicts

## Troubleshooting

### Common Issues

1. **VM doesn't get IP address**:

   - Check if template VM has proper network configuration
   - Verify cloud-init templates are valid
   - Check VirtualBox network adapter settings

2. **IP address conflicts still occur**:

   - Ensure the fix is properly applied
   - Check that MAC addresses are being updated
   - Verify DHCP server configuration

3. **VM fails to start**:
   - Check VirtualBox installation
   - Verify template VM exists and is properly configured
   - Check system resources availability

### Debugging

- Check VirtualBox logs for network-related errors
- Verify MAC address assignment using `VBoxManage showvminfo`
- Monitor DHCP server logs for IP assignment issues
