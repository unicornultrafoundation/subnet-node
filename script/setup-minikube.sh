#!/bin/bash

# Minikube Setup Script
# This script installs and configures Minikube for local Kubernetes development

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
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if [[ -f /etc/os-release ]]; then
            . /etc/os-release
            OS=$NAME
            VER=$VERSION_ID
        else
            OS=$(uname -s)
            VER=$(uname -r)
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macOS"
        VER=$(sw_vers -productVersion)
    else
        print_error "Unsupported operating system: $OSTYPE"
        exit 1
    fi
    print_status "Detected OS: $OS $VER"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check system requirements
check_requirements() {
    print_status "Checking system requirements..."
    
    # Check CPU cores
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        CPU_CORES=$(nproc)
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        CPU_CORES=$(sysctl -n hw.ncpu)
    fi
    
    if [[ $CPU_CORES -lt 2 ]]; then
        print_warning "Minikube requires at least 2 CPU cores. Found: $CPU_CORES"
    else
        print_success "CPU cores: $CPU_CORES"
    fi
    
    # Check memory
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        MEM_GB=$(free -g | awk '/^Mem:/{print $2}')
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        MEM_GB=$(sysctl -n hw.memsize | awk '{print int($1/1024/1024/1024)}')
    fi
    
    if [[ $MEM_GB -lt 2 ]]; then
        print_warning "Minikube requires at least 2GB RAM. Found: ${MEM_GB}GB"
    else
        print_success "Memory: ${MEM_GB}GB"
    fi
    
    # Check disk space
    DISK_GB=$(df -BG . | awk 'NR==2 {print $4}' | sed 's/G//')
    if [[ $DISK_GB -lt 20 ]]; then
        print_warning "Minikube requires at least 20GB free disk space. Found: ${DISK_GB}GB"
    else
        print_success "Free disk space: ${DISK_GB}GB"
    fi
}

# Function to install kubectl
install_kubectl() {
    if command_exists kubectl; then
        print_status "kubectl is already installed"
        kubectl version --client --short
        return 0
    fi
    
    print_status "Installing kubectl..."
    
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Download kubectl
        curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
        chmod +x kubectl
        sudo mv kubectl /usr/local/bin/
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        if command_exists brew; then
            brew install kubectl
        else
            # Download kubectl for macOS
            curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/amd64/kubectl"
            chmod +x kubectl
            sudo mv kubectl /usr/local/bin/
        fi
    fi
    
    print_success "kubectl installed successfully"
    kubectl version --client --short
}

# Function to install Docker
install_docker() {
    if command_exists docker; then
        print_status "Docker is already installed"
        docker --version
        return 0
    fi
    
    print_status "Installing Docker..."
    
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if [[ "$OS" == *"Ubuntu"* ]] || [[ "$OS" == *"Debian"* ]]; then
            # Install Docker on Ubuntu/Debian
            sudo apt-get update
            sudo apt-get install -y apt-transport-https ca-certificates curl gnupg lsb-release
            curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
            echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
            sudo apt-get update
            sudo apt-get install -y docker-ce docker-ce-cli containerd.io
            sudo usermod -aG docker $USER
            print_warning "Please log out and back in for Docker group changes to take effect"
        elif [[ "$OS" == *"CentOS"* ]] || [[ "$OS" == *"Red Hat"* ]]; then
            # Install Docker on CentOS/RHEL
            sudo yum install -y yum-utils
            sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
            sudo yum install -y docker-ce docker-ce-cli containerd.io
            sudo systemctl start docker
            sudo systemctl enable docker
            sudo usermod -aG docker $USER
            print_warning "Please log out and back in for Docker group changes to take effect"
        else
            print_error "Unsupported Linux distribution for Docker installation"
            exit 1
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        if command_exists brew; then
            brew install --cask docker
        else
            print_error "Please install Docker Desktop for macOS manually"
            exit 1
        fi
    fi
    
    print_success "Docker installed successfully"
}

# Function to install Minikube
install_minikube() {
    if command_exists minikube; then
        print_status "Minikube is already installed"
        minikube version
        return 0
    fi
    
    print_status "Installing Minikube..."
    
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Download and install Minikube for Linux
        curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
        sudo install minikube-linux-amd64 /usr/local/bin/minikube
        rm minikube-linux-amd64
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        if command_exists brew; then
            brew install minikube
        else
            # Download and install Minikube for macOS
            curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-darwin-amd64
            sudo install minikube-darwin-amd64 /usr/local/bin/minikube
            rm minikube-darwin-amd64
        fi
    fi
    
    print_success "Minikube installed successfully"
    minikube version
}

# Function to start Minikube
start_minikube() {
    print_status "Starting Minikube cluster..."
    
    # Check if minikube is already running
    if minikube status | grep -q "Running"; then
        print_status "Minikube is already running"
        return 0
    fi
    
    # Start minikube with recommended settings
    minikube start \
        --driver=docker \
        --cpus=2 \
        --memory=4096 \
        --disk-size=20g \
        --addons=ingress \
        --addons=metrics-server \
        --addons=dashboard
    
    print_success "Minikube cluster started successfully"
}

# Function to configure kubectl
configure_kubectl() {
    print_status "Configuring kubectl..."
    
    # Set kubectl context to minikube
    kubectl config use-context minikube
    
    # Verify connection
    if kubectl cluster-info; then
        print_success "kubectl configured successfully"
    else
        print_error "Failed to configure kubectl"
        exit 1
    fi
}

# Function to install additional tools
install_additional_tools() {
    print_status "Installing additional Kubernetes tools..."
    
    # Install helm if not present
    if ! command_exists helm; then
        print_status "Installing Helm..."
        if [[ "$OSTYPE" == "linux-gnu"* ]]; then
            curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
        elif [[ "$OSTYPE" == "darwin"* ]]; then
            if command_exists brew; then
                brew install helm
            else
                curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
            fi
        fi
        print_success "Helm installed successfully"
    else
        print_status "Helm is already installed"
    fi
    
    # Install kubectx and kubens if not present
    if ! command_exists kubectx; then
        print_status "Installing kubectx and kubens..."
        if [[ "$OSTYPE" == "linux-gnu"* ]]; then
            sudo git clone https://github.com/ahmetb/kubectx /opt/kubectx
            sudo ln -s /opt/kubectx/kubectx /usr/local/bin/kubectx
            sudo ln -s /opt/kubectx/kubens /usr/local/bin/kubens
        elif [[ "$OSTYPE" == "darwin"* ]]; then
            if command_exists brew; then
                brew install kubectx
            else
                sudo git clone https://github.com/ahmetb/kubectx /opt/kubectx
                sudo ln -s /opt/kubectx/kubectx /usr/local/bin/kubectx
                sudo ln -s /opt/kubectx/kubens /usr/local/bin/kubens
            fi
        fi
        print_success "kubectx and kubens installed successfully"
    else
        print_status "kubectx and kubens are already installed"
    fi
}

# Function to verify installation
verify_installation() {
    print_status "Verifying installation..."
    
    echo
    print_status "=== Installation Verification ==="
    
    # Check kubectl
    if command_exists kubectl; then
        print_success "✓ kubectl is installed"
        kubectl version --client --short
    else
        print_error "✗ kubectl is not installed"
    fi
    
    # Check minikube
    if command_exists minikube; then
        print_success "✓ minikube is installed"
        minikube version
    else
        print_error "✗ minikube is not installed"
    fi
    
    # Check docker
    if command_exists docker; then
        print_success "✓ docker is installed"
        docker --version
    else
        print_error "✗ docker is not installed"
    fi
    
    # Check cluster status
    if minikube status | grep -q "Running"; then
        print_success "✓ Minikube cluster is running"
        kubectl get nodes
    else
        print_warning "⚠ Minikube cluster is not running"
    fi
    
    # Check helm
    if command_exists helm; then
        print_success "✓ helm is installed"
        helm version --short
    else
        print_warning "⚠ helm is not installed"
    fi
    
    echo
}

# Function to show useful commands
show_useful_commands() {
    print_status "=== Useful Commands ==="
    echo
    echo "Minikube commands:"
    echo "  minikube start          - Start the cluster"
    echo "  minikube stop           - Stop the cluster"
    echo "  minikube delete         - Delete the cluster"
    echo "  minikube status         - Show cluster status"
    echo "  minikube dashboard      - Open Kubernetes dashboard"
    echo "  minikube tunnel         - Enable LoadBalancer access"
    echo
    echo "Kubernetes commands:"
    echo "  kubectl get nodes       - List cluster nodes"
    echo "  kubectl get pods        - List pods in default namespace"
    echo "  kubectl get pods -A     - List pods in all namespaces"
    echo "  kubectl get services    - List services"
    echo "  kubectl describe pod    - Describe a pod"
    echo "  kubectl logs <pod>      - Show pod logs"
    echo
    echo "Helm commands:"
    echo "  helm repo add <name> <url>  - Add a Helm repository"
    echo "  helm install <name> <chart> - Install a Helm chart"
    echo "  helm list               - List installed releases"
    echo "  helm uninstall <name>   - Uninstall a release"
    echo
}


# --- Subcommand logic ---
case "$1" in
  install)
    echo "=========================================="
    echo "    Minikube Install"
    echo "=========================================="
    detect_os
    check_requirements
    install_kubectl
    install_docker
    install_minikube
    install_additional_tools
    print_success "Minikube and dependencies installed!"
    ;;
  up)
    echo "=========================================="
    echo "    Minikube Up"
    echo "=========================================="
    detect_os
    start_minikube
    configure_kubectl
    verify_installation
    show_useful_commands
    ;;
  clean)
    echo "=========================================="
    echo "    Minikube Clean"
    echo "=========================================="
    if command_exists minikube; then
      minikube delete
      print_success "Minikube cluster deleted."
    else
      print_warning "Minikube is not installed. Nothing to clean."
    fi
    ;;
  ""|help|-h|--help)
    echo "Usage: $0 <install|up|clean>"
    echo "  install   Install minikube, kubectl, docker, helm, etc."
    echo "  up        Start minikube cluster and verify setup"
    echo "  clean     Delete minikube cluster"
    ;;
  *)
    print_error "Unknown command: $1"
    echo "Usage: $0 <install|up|clean>"
    exit 1
    ;;
esac
