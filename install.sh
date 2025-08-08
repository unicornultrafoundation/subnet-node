#!/bin/bash

# Easy Subnet Node Installer
# Simple installation script for non-technical users

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  Subnet Node Easy Installer${NC}"
    echo -e "${BLUE}================================${NC}"
}

print_step() {
    echo -e "${BLUE}[STEP $1]${NC} $2"
}

# Function to show welcome message
show_welcome() {
    print_header
    echo
    echo "Welcome to Subnet Node Easy Installer!"
    echo
    echo "This script will help you install Subnet Node on your K3s cluster."
    echo "You will need to provide some information during the installation."
    echo
    print_warning "Before starting, make sure you have:"
    echo "  ✓ K3s installed and running"
    echo "  ✓ kubectl installed"
    echo "  ✓ Your private key ready"
    echo "  ✓ Your provider ID and machine ID"
    echo
    echo "Press Enter to continue or Ctrl+C to cancel..."
    read -r
}

# Function to check prerequisites
check_prerequisites() {
    print_step "1" "Checking system requirements..."
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed. Please install kubectl first."
        echo "You can install it by running:"
        echo "  curl -LO \"https://dl.k8s.io/release/\$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl\""
        echo "  sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl"
        exit 1
    fi
    print_status "✓ kubectl found"
    
    # Check K3s connection
    if ! kubectl cluster-info &> /dev/null; then
        print_error "Cannot connect to K3s cluster. Please make sure K3s is running."
        echo "You can start K3s by running:"
        echo "  sudo systemctl start k3s"
        exit 1
    fi
    print_status "✓ Connected to K3s cluster"
    
    echo
}

# Function to get private key
get_private_key() {
    print_step "2" "Setting up your private key..."
    echo
    echo "You have 3 options to provide your private key:"
    echo
    echo "1. Create a key file (Recommended)"
    echo "2. Enter key manually (Most secure)"
    echo "3. Use existing key file"
    echo
    echo -n "Choose option (1-3): "
    read -r choice
    
    case $choice in
        1)
            echo
            echo "Creating a new key file..."
            echo -n "Enter your private key: "
            read -s private_key
            echo
            
            # Create .subnet directory if it doesn't exist
            mkdir -p ~/.subnet
            
            # Save to file
            echo "$private_key" > ~/.subnet/private.key
            chmod 600 ~/.subnet/private.key
            
            PRIVATE_KEY_FILE="$HOME/.subnet/private.key"
            print_status "✓ Private key saved to $PRIVATE_KEY_FILE"
            ;;
        2)
            echo
            echo -n "Enter your private key (input will be hidden): "
            read -s PRIVATE_KEY
            echo
            print_status "✓ Private key entered"
            ;;
        3)
            echo
            echo -n "Enter the path to your existing key file: "
            read -r PRIVATE_KEY_FILE
            if [[ ! -f "$PRIVATE_KEY_FILE" ]]; then
                print_error "File not found: $PRIVATE_KEY_FILE"
                exit 1
            fi
            print_status "✓ Using existing key file: $PRIVATE_KEY_FILE"
            ;;
        *)
            print_error "Invalid choice. Please run the script again."
            exit 1
            ;;
    esac
    echo
}

# Function to get keystore password
get_keystore_password() {
    print_step "2.5" "Setting up keystore password..."
    echo
    echo "The private key will be imported into a keystore file."
    echo "You can set a password to protect the keystore (recommended) or leave it empty."
    echo
    echo -n "Enter keystore password (or press Enter for no password): "
    read -s KEYSTORE_PASSWORD
    echo
    
    if [[ -z "$KEYSTORE_PASSWORD" ]]; then
        print_warning "No password set for keystore (less secure)"
    else
        print_status "✓ Keystore password set"
    fi
    echo
}

# Function to get provider and machine IDs
get_ids() {
    print_step "3" "Setting up your node information..."
    echo
    
    echo -n "Enter your Provider ID: "
    read -r PROVIDER_ID
    
    echo -n "Enter your Machine ID: "
    read -r MACHINE_ID
    
    if [[ -z "$PROVIDER_ID" || -z "$MACHINE_ID" ]]; then
        print_error "Provider ID and Machine ID cannot be empty."
        exit 1
    fi
    
    print_status "✓ Provider ID: $PROVIDER_ID"
    print_status "✓ Machine ID: $MACHINE_ID"
    echo
}

# Function to get installation options
get_options() {
    print_step "4" "Setting up installation options..."
    echo
    
    echo "Default settings:"
    echo "  - Namespace: subnet"
    echo "  - Storage: 10GB"
    echo "  - Replicas: 1"
    echo
    
    echo -n "Do you want to use default settings? (y/N): "
    read -r use_default
    
    if [[ "$use_default" =~ ^[Yy]$ ]]; then
        NAMESPACE="subnet"
        STORAGE_SIZE="10Gi"
        REPLICAS=1
        print_status "✓ Using default settings"
    else
        echo
        echo -n "Enter namespace name (default: subnet): "
        read -r namespace_input
        NAMESPACE="${namespace_input:-subnet}"
        
        echo -n "Enter storage size (default: 10Gi): "
        read -r storage_input
        STORAGE_SIZE="${storage_input:-10Gi}"
        
        echo -n "Enter number of replicas (default: 1): "
        read -r replicas_input
        REPLICAS="${replicas_input:-1}"
        
        print_status "✓ Custom settings applied"
    fi
    echo
}

# Function to confirm installation
confirm_installation() {
    print_step "5" "Confirming installation..."
    echo
    echo "Installation Summary:"
    echo "  Provider ID: $PROVIDER_ID"
    echo "  Machine ID: $MACHINE_ID"
    echo "  Namespace: $NAMESPACE"
    echo "  Storage: $STORAGE_SIZE"
    echo "  Replicas: $REPLICAS"
    if [[ -n "$PRIVATE_KEY_FILE" ]]; then
        echo "  Private Key: From file ($PRIVATE_KEY_FILE)"
    else
        echo "  Private Key: Entered manually"
    fi
    if [[ -n "$KEYSTORE_PASSWORD" ]]; then
        echo "  Keystore Password: Set"
    else
        echo "  Keystore Password: None (less secure)"
    fi
    echo
    
    echo -n "Do you want to proceed with the installation? (y/N): "
    read -r confirm
    
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        print_status "Installation cancelled."
        exit 0
    fi
    echo
}

# Function to run installation
run_installation() {
    print_step "6" "Installing Subnet Node..."
    echo
    
    # Get the directory where this script is located
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    SETUP_SCRIPT="$SCRIPT_DIR/setup/setup-subnet-k3s.sh"
    
    if [[ ! -f "$SETUP_SCRIPT" ]]; then
        print_error "Installation script not found. Please make sure you're running this from the correct directory."
        exit 1
    fi
    
    # Make script executable
    chmod +x "$SETUP_SCRIPT"
    
    # Prepare private key
    if [[ -n "$PRIVATE_KEY_FILE" ]]; then
        PRIVATE_KEY=$(cat "$PRIVATE_KEY_FILE")
    fi
    
    # Run the installation
    print_status "Starting installation... This may take a few minutes."
    echo
    
    "$SETUP_SCRIPT" \
        --private-key "$PRIVATE_KEY" \
        --provider-id "$PROVIDER_ID" \
        --machine-id "$MACHINE_ID" \
        --namespace "$NAMESPACE" \
        --replicas "$REPLICAS" \
        --storage-size "$STORAGE_SIZE" \
        --keystore-password "$KEYSTORE_PASSWORD"
}

# Function to show completion
show_completion() {
    print_step "7" "Installation completed!"
    echo
    print_status "Your Subnet Node has been successfully installed!"
    echo
    echo "Next steps:"
    echo "  1. Check if the node is running:"
    echo "     kubectl get pods -n $NAMESPACE"
    echo
    echo "  2. View the logs:"
    echo "     kubectl logs -f deployment/subnet-node -n $NAMESPACE"
    echo
    echo "  3. Access the API:"
    echo "     kubectl port-forward service/subnet-node-service 8080:8080 -n $NAMESPACE"
    echo
    echo "  4. To uninstall later:"
    echo "     ./setup/uninstall-subnet-k3s.sh --namespace $NAMESPACE"
    echo
    print_status "Thank you for using Subnet Node Easy Installer!"
}

# Function to cleanup
cleanup() {
    # Clear private key from memory
    PRIVATE_KEY=""
}

# Main function
main() {
    # Trap to cleanup on exit
    trap cleanup EXIT
    
    show_welcome
    check_prerequisites
    get_private_key
    get_keystore_password
    get_ids
    get_options
    confirm_installation
    run_installation
    show_completion
}

# Run main function
main "$@"
