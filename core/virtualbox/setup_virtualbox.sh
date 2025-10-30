#!/bin/bash

# VirtualBox Setup Script
# This script detects the operating system and installs the required packages
# for the VirtualBox service as specified in the README

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to detect OS
detect_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if command -v apt-get &> /dev/null; then
            echo "ubuntu"
        elif command -v yum &> /dev/null; then
            echo "rhel"
        elif command -v pacman &> /dev/null; then
            echo "arch"
        else
            echo "linux"
        fi
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        echo "windows"
    else
        echo "unknown"
    fi
}

# Function to check if command exists
command_exists() {
    command -v "$1" &> /dev/null
}

# Function to check if package is installed (macOS)
check_brew_package() {
    brew list "$1" &> /dev/null
}

# Function to check if package is installed (Ubuntu/Debian)
check_apt_package() {
    dpkg -l | grep -q "^ii.*$1"
}

# Function to install packages on macOS
install_macos_packages() {
    print_status "Installing packages for macOS..."
    
    # Check if Homebrew is installed
    if ! command_exists brew; then
        print_error "Homebrew is not installed. Please install Homebrew first:"
        echo "  /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
        exit 1
    fi
    
    # Required packages for macOS
    packages=("virtualbox" "qemu" "cdrtools")
    
    for package in "${packages[@]}"; do
        if check_brew_package "$package"; then
            print_success "$package is already installed"
        else
            print_status "Installing $package..."
            if brew install "$package"; then
                print_success "$package installed successfully"
            else
                print_error "Failed to install $package"
                exit 1
            fi
        fi
    done
}

# Function to install packages on Ubuntu/Debian
install_ubuntu_packages() {
    print_status "Installing packages for Ubuntu/Debian..."
    
    # Update package list
    print_status "Updating package list..."
    sudo apt-get update
    
    # Required packages for Ubuntu/Debian
    packages=("virtualbox-7.1" "qemu-utils" "genisoimage")
    
    for package in "${packages[@]}"; do
        if check_apt_package "$package"; then
            print_success "$package is already installed"
        else
            print_status "Installing $package..."
            if sudo apt-get install -y "$package"; then
                print_success "$package installed successfully"
            else
                print_error "Failed to install $package"
                exit 1
            fi
        fi
    done
}

# Function to install packages on RHEL/CentOS/Fedora
install_rhel_packages() {
    print_status "Installing packages for RHEL/CentOS/Fedora..."
    
    # Check if dnf or yum is available
    if command_exists dnf; then
        PKG_MANAGER="dnf"
    elif command_exists yum; then
        PKG_MANAGER="yum"
    else
        print_error "Neither dnf nor yum package manager found"
        exit 1
    fi
    
    # Update package list
    print_status "Updating package list..."
    sudo $PKG_MANAGER update -y
    
    # Required packages for RHEL/CentOS/Fedora
    packages=("VirtualBox" "qemu-img" "genisoimage")
    
    for package in "${packages[@]}"; do
        if rpm -q "$package" &> /dev/null; then
            print_success "$package is already installed"
        else
            print_status "Installing $package..."
            if sudo $PKG_MANAGER install -y "$package"; then
                print_success "$package installed successfully"
            else
                print_error "Failed to install $package"
                exit 1
            fi
        fi
    done
}

# Function to install packages on Arch Linux
install_arch_packages() {
    print_status "Installing packages for Arch Linux..."
    
    # Update package list
    print_status "Updating package list..."
    sudo pacman -Sy
    
    # Required packages for Arch Linux
    packages=("virtualbox" "qemu" "cdrtools")
    
    for package in "${packages[@]}"; do
        if pacman -Q "$package" &> /dev/null; then
            print_success "$package is already installed"
        else
            print_status "Installing $package..."
            if sudo pacman -S --noconfirm "$package"; then
                print_success "$package installed successfully"
            else
                print_error "Failed to install $package"
                exit 1
            fi
        fi
    done
}

# Function to verify installations
verify_installations() {
    print_status "Verifying installations..."
    
    # Check VirtualBox
    if command_exists VBoxManage; then
        print_success "VirtualBox (VBoxManage) is available"
        VBoxManage --version
    else
        print_error "VirtualBox (VBoxManage) is not available"
        return 1
    fi
    
    # Check QEMU tools
    if command_exists qemu-img; then
        print_success "QEMU tools (qemu-img) are available"
        qemu-img --version
    else
        print_error "QEMU tools (qemu-img) are not available"
        return 1
    fi
    
    # Check ISO generation tools
    if command_exists genisoimage || command_exists mkisofs; then
        print_success "ISO generation tools are available"
        if command_exists genisoimage; then
            genisoimage --version
        else
            mkisofs --version
        fi
    else
        print_error "ISO generation tools are not available"
        return 1
    fi
    
    return 0
}

# Function to show architecture information
show_architecture_info() {
    print_status "System Architecture Information:"
    echo "  OS Type: $(uname -s)"
    echo "  Architecture: $(uname -m)"
    echo "  Kernel: $(uname -r)"
    
    # Check for Apple Silicon
    if [[ "$OSTYPE" == "darwin"* ]] && [[ "$(uname -m)" == "arm64" ]]; then
        print_warning "Detected Apple Silicon (ARM64). VirtualBox may have limited support."
        print_warning "Consider using UTM or VMware Fusion for better ARM64 virtualization."
    fi
}

# Function to show post-installation notes
show_post_install_notes() {
    print_status "Post-installation Notes:"
    echo ""
    echo "1. VirtualBox Extension Pack:"
    echo "   - Download from: https://www.virtualbox.org/wiki/Downloads"
    echo "   - Install for additional features (USB 3.0, RDP, etc.)"
    echo ""
    echo "2. VirtualBox User Groups:"
    echo "   - Add your user to the vboxusers group (Linux):"
    echo "     sudo usermod -a -G vboxusers \$USER"
    echo "   - Log out and log back in for changes to take effect"
    echo ""
    echo "3. VirtualBox Directory:"
    echo "   - Default VM directory: ~/VirtualBox VMs/"
    echo "   - The service will create Images/ subdirectory for cloud images"
    echo ""
    echo "4. Testing the installation:"
    echo "   - Run: VBoxManage --version"
    echo "   - Run: qemu-img --version"
    echo "   - Run: genisoimage --version (or mkisofs --version)"
    echo ""
}

# Main execution
main() {
    echo "=========================================="
    echo "VirtualBox Setup Script"
    echo "=========================================="
    echo ""
    
    # Detect OS
    OS=$(detect_os)
    print_status "Detected OS: $OS"
    
    # Show architecture info
    show_architecture_info
    echo ""
    
    # Install packages based on OS
    case $OS in
        "macos")
            install_macos_packages
            ;;
        "ubuntu")
            install_ubuntu_packages
            ;;
        "rhel")
            install_rhel_packages
            ;;
        "arch")
            install_arch_packages
            ;;
        "windows")
            print_error "Windows is not supported by this script."
            print_error "Please install VirtualBox manually from: https://www.virtualbox.org/wiki/Downloads"
            exit 1
            ;;
        *)
            print_error "Unsupported operating system: $OS"
            print_error "Please install the following packages manually:"
            echo "  - VirtualBox"
            echo "  - QEMU tools (qemu-img)"
            echo "  - ISO generation tools (genisoimage or cdrtools)"
            exit 1
            ;;
    esac
    
    echo ""
    
    # Verify installations
    if verify_installations; then
        print_success "All required packages are installed and working!"
    else
        print_error "Some packages failed verification. Please check the installation."
        exit 1
    fi
    
    echo ""
    
    # Show post-installation notes
    show_post_install_notes
    
    print_success "VirtualBox setup completed successfully!"
    print_status "You can now use the VirtualBox service with your subnet-node application."
}

# Run main function
main "$@"
