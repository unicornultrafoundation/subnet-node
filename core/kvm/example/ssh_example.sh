#!/bin/bash

# KVM Service SSH Example Script
# This script demonstrates how to connect to VMs created by the KVM service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SSH_KEY_PATH="/var/lib/libvirt/ssh/subnet-key"
DEFAULT_USERNAME="ubuntu"
DEFAULT_PORT=22

echo -e "${BLUE}=== KVM Service SSH Example ===${NC}"

# Check if SSH key exists
if [ ! -f "$SSH_KEY_PATH" ]; then
    echo -e "${RED}SSH key not found at: $SSH_KEY_PATH${NC}"
    echo -e "${YELLOW}The KVM service should generate this automatically when creating VMs.${NC}"
    echo -e "${YELLOW}Make sure you have created at least one VM using the KVM service.${NC}"
    exit 1
fi

# Set proper permissions for SSH key
echo -e "${BLUE}Setting SSH key permissions...${NC}"
chmod 600 "$SSH_KEY_PATH"
chmod 644 "$SSH_KEY_PATH.pub"

echo -e "${GREEN}SSH key permissions set correctly${NC}"

# Function to test SSH connectivity
test_ssh_connectivity() {
    local ip_address=$1
    local timeout=3
    
    echo -e "${BLUE}Testing SSH connectivity to $ip_address...${NC}"
    
    if timeout $timeout bash -c "</dev/tcp/$ip_address/$DEFAULT_PORT" 2>/dev/null; then
        echo -e "${GREEN}✓ SSH is accessible on $ip_address:$DEFAULT_PORT${NC}"
        return 0
    else
        echo -e "${RED}✗ SSH is not accessible on $ip_address:$DEFAULT_PORT${NC}"
        return 1
    fi
}

# Function to connect to VM
connect_to_vm() {
    local ip_address=$1
    local vm_name=${2:-"VM"}
    
    echo -e "${BLUE}Connecting to $vm_name at $ip_address...${NC}"
    echo -e "${YELLOW}SSH Command: ssh -i $SSH_KEY_PATH $DEFAULT_USERNAME@$ip_address${NC}"
    echo -e "${YELLOW}Default credentials:${NC}"
    echo -e "${YELLOW}  Username: $DEFAULT_USERNAME${NC}"
    echo -e "${YELLOW}  Password: password${NC}"
    echo -e "${YELLOW}  SSH Key: Automatically configured${NC}"
    echo ""
    
    # Test connectivity first
    if test_ssh_connectivity "$ip_address"; then
        echo -e "${GREEN}Attempting SSH connection...${NC}"
        echo -e "${YELLOW}Press Ctrl+C to exit SSH session${NC}"
        echo ""
        
        # Connect via SSH
        ssh -i "$SSH_KEY_PATH" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "$DEFAULT_USERNAME@$ip_address"
    else
        echo -e "${RED}Cannot connect to VM. Please check:${NC}"
        echo -e "${YELLOW}1. VM is running and has an IP address${NC}"
        echo -e "${YELLOW}2. Network connectivity to the VM${NC}"
        echo -e "${YELLOW}3. SSH service is running on the VM${NC}"
        echo -e "${YELLOW}4. Firewall allows SSH connections${NC}"
    fi
}

# Function to list available VMs (if virsh is available)
list_vms() {
    if command -v virsh >/dev/null 2>&1; then
        echo -e "${BLUE}Available VMs (from virsh):${NC}"
        virsh list --all | grep -E "(running|idle)" | while read -r line; do
            if [[ $line =~ ([0-9]+)\ +([a-zA-Z0-9_-]+)\ +([a-zA-Z]+) ]]; then
                vm_id="${BASH_REMATCH[1]}"
                vm_name="${BASH_REMATCH[2]}"
                vm_state="${BASH_REMATCH[3]}"
                echo -e "${GREEN}  ID: $vm_id, Name: $vm_name, State: $vm_state${NC}"
            fi
        done
    else
        echo -e "${YELLOW}virsh not available, cannot list VMs${NC}"
    fi
}

# Function to get VM IP address (if virsh is available)
get_vm_ip() {
    local vm_name=$1
    
    if command -v virsh >/dev/null 2>&1; then
        echo -e "${BLUE}Getting IP address for VM: $vm_name${NC}"
        
        # Try to get IP from domain network info
        local ip_output=$(virsh domifaddr "$vm_name" 2>/dev/null | grep "ipv4" | awk '{print $4}' | cut -d'/' -f1)
        
        if [ -n "$ip_output" ]; then
            echo -e "${GREEN}Found IP: $ip_output${NC}"
            echo "$ip_output"
        else
            echo -e "${YELLOW}Could not get IP address for VM: $vm_name${NC}"
            echo ""
        fi
    else
        echo -e "${YELLOW}virsh not available, cannot get VM IP${NC}"
        echo ""
    fi
}

# Main script logic
echo -e "${BLUE}SSH Key Path: $SSH_KEY_PATH${NC}"
echo -e "${BLUE}Default Username: $DEFAULT_USERNAME${NC}"
echo -e "${BLUE}Default Port: $DEFAULT_PORT${NC}"
echo ""

# List available VMs
list_vms
echo ""

# Check if VM name is provided as argument
if [ $# -eq 0 ]; then
    echo -e "${YELLOW}Usage: $0 <vm_name>${NC}"
    echo -e "${YELLOW}Example: $0 my-vm${NC}"
    echo ""
    echo -e "${YELLOW}Or provide IP address directly:${NC}"
    echo -e "${YELLOW}Usage: $0 --ip <ip_address>${NC}"
    echo -e "${YELLOW}Example: $0 --ip 192.168.123.45${NC}"
    echo ""
    
    # Interactive mode
    read -p "Enter VM name (or press Enter to skip): " vm_name
    if [ -n "$vm_name" ]; then
        ip_address=$(get_vm_ip "$vm_name")
        if [ -n "$ip_address" ]; then
            connect_to_vm "$ip_address" "$vm_name"
        else
            echo -e "${RED}Could not get IP address for VM: $vm_name${NC}"
            echo -e "${YELLOW}Please make sure the VM is running and has an IP address.${NC}"
        fi
    else
        echo -e "${YELLOW}No VM name provided. Exiting.${NC}"
    fi
else
    if [ "$1" = "--ip" ] && [ -n "$2" ]; then
        # Direct IP connection
        connect_to_vm "$2" "VM at $2"
    else
        # VM name provided
        vm_name="$1"
        ip_address=$(get_vm_ip "$vm_name")
        if [ -n "$ip_address" ]; then
            connect_to_vm "$ip_address" "$vm_name"
        else
            echo -e "${RED}Could not get IP address for VM: $vm_name${NC}"
            echo -e "${YELLOW}Please make sure the VM is running and has an IP address.${NC}"
        fi
    fi
fi

echo ""
echo -e "${BLUE}=== SSH Example Completed ===${NC}" 