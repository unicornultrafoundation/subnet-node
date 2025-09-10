#!/bin/bash

# Subnet Node K3s Uninstaller
# This script removes Subnet Node from K3s cluster
#
# Options:
#   --namespace NS        Kubernetes namespace (default: subnet)
#   --force              Force deletion without confirmation
#   --keep-data          Keep persistent data (don't delete PVC)
#   --help               Show this help message

set -e

# Default values
NAMESPACE="subnet"
FORCE=false
KEEP_DATA=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  Subnet Node K3s Uninstaller${NC}"
    echo -e "${BLUE}================================${NC}"
}

# Function to show help
show_help() {
    cat << EOF
Subnet Node K3s Uninstaller

Usage: $0 [OPTIONS]

Options:
  --namespace NS        Kubernetes namespace (default: subnet)
  --force              Force deletion without confirmation
  --keep-data          Keep persistent data (don't delete PVC)
  --help               Show this help message

Examples:
  $0 --namespace subnet
  $0 --namespace my-subnet --force
  $0 --keep-data

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        --force)
            FORCE=true
            shift
            ;;
        --keep-data)
            KEEP_DATA=true
            shift
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Function to check if kubectl is installed
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    print_status "kubectl found: $(kubectl version --client --short)"
}

# Function to check if namespace exists
check_namespace() {
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        print_error "Namespace '$NAMESPACE' does not exist."
        exit 1
    fi
    print_status "Namespace '$NAMESPACE' found"
}

# Function to show what will be deleted
show_deletion_plan() {
    print_status "The following resources will be deleted from namespace '$NAMESPACE':"
    echo
    
    # Check for deployments
    if kubectl get deployments -n "$NAMESPACE" --no-headers 2>/dev/null | grep -q .; then
        print_status "Deployments:"
        kubectl get deployments -n "$NAMESPACE" --no-headers | awk '{print "  - " $1}'
    fi
    
    # Check for services
    if kubectl get services -n "$NAMESPACE" --no-headers 2>/dev/null | grep -q .; then
        print_status "Services:"
        kubectl get services -n "$NAMESPACE" --no-headers | awk '{print "  - " $1}'
    fi
    
    # Check for ingress
    if kubectl get ingress -n "$NAMESPACE" --no-headers 2>/dev/null | grep -q .; then
        print_status "Ingress:"
        kubectl get ingress -n "$NAMESPACE" --no-headers | awk '{print "  - " $1}'
    fi
    
    # Check for configmaps
    if kubectl get configmaps -n "$NAMESPACE" --no-headers 2>/dev/null | grep -q .; then
        print_status "ConfigMaps:"
        kubectl get configmaps -n "$NAMESPACE" --no-headers | awk '{print "  - " $1}'
    fi
    
    # Check for secrets
    if kubectl get secrets -n "$NAMESPACE" --no-headers 2>/dev/null | grep -q .; then
        print_status "Secrets:"
        kubectl get secrets -n "$NAMESPACE" --no-headers | awk '{print "  - " $1}'
    fi
    
    # Check for PVCs
    if kubectl get pvc -n "$NAMESPACE" --no-headers 2>/dev/null | grep -q .; then
        if [[ "$KEEP_DATA" == true ]]; then
            print_warning "PVCs (will be kept due to --keep-data option):"
        else
            print_status "PVCs:"
        fi
        kubectl get pvc -n "$NAMESPACE" --no-headers | awk '{print "  - " $1}'
    fi
    
    echo
}

# Function to confirm deletion
confirm_deletion() {
    if [[ "$FORCE" == true ]]; then
        print_warning "Force mode enabled, skipping confirmation"
        return 0
    fi
    
    echo -n "Are you sure you want to proceed with the deletion? (y/N): "
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        return 0
    else
        print_status "Deletion cancelled"
        exit 0
    fi
}

# Function to delete resources
delete_resources() {
    print_status "Deleting Subnet Node resources..."
    
    # Delete deployment
    if kubectl get deployment subnet-node -n "$NAMESPACE" &> /dev/null; then
        print_status "Deleting deployment: subnet-node"
        kubectl delete deployment subnet-node -n "$NAMESPACE"
    fi
    
    # Delete service
    if kubectl get service subnet-node-service -n "$NAMESPACE" &> /dev/null; then
        print_status "Deleting service: subnet-node-service"
        kubectl delete service subnet-node-service -n "$NAMESPACE"
    fi
    
    # Delete ingress
    if kubectl get ingress subnet-node-ingress -n "$NAMESPACE" &> /dev/null; then
        print_status "Deleting ingress: subnet-node-ingress"
        kubectl delete ingress subnet-node-ingress -n "$NAMESPACE"
    fi
    
    # Delete configmap
    if kubectl get configmap subnet-config -n "$NAMESPACE" &> /dev/null; then
        print_status "Deleting configmap: subnet-config"
        kubectl delete configmap subnet-config -n "$NAMESPACE"
    fi
    
    # Delete keystore configmap
    if kubectl get configmap subnet-keystore -n "$NAMESPACE" &> /dev/null; then
        print_status "Deleting configmap: subnet-keystore"
        kubectl delete configmap subnet-keystore -n "$NAMESPACE"
    fi
    
    # Delete secret
    if kubectl get secret subnet-keystore-password -n "$NAMESPACE" &> /dev/null; then
        print_status "Deleting secret: subnet-keystore-password"
        kubectl delete secret subnet-keystore-password -n "$NAMESPACE"
    fi
    
    # Delete PVC if not keeping data
    if [[ "$KEEP_DATA" == false ]]; then
        if kubectl get pvc subnet-data -n "$NAMESPACE" &> /dev/null; then
            print_status "Deleting PVC: subnet-data"
            kubectl delete pvc subnet-data -n "$NAMESPACE"
        fi
    else
        print_warning "Keeping PVC: subnet-data (use --keep-data=false to delete)"
    fi
}

# Function to delete namespace
delete_namespace() {
    # Check if namespace is empty
    if kubectl get all -n "$NAMESPACE" --no-headers 2>/dev/null | grep -q .; then
        print_warning "Namespace '$NAMESPACE' still contains resources, not deleting namespace"
        return 0
    fi
    
    print_status "Deleting namespace: $NAMESPACE"
    kubectl delete namespace "$NAMESPACE"
    print_status "Namespace '$NAMESPACE' deleted"
}

# Function to show final status
show_final_status() {
    print_status "Uninstallation completed!"
    echo
    
    if [[ "$KEEP_DATA" == true ]]; then
        print_warning "Note: Persistent data was kept. To delete it later, run:"
        echo "kubectl delete pvc subnet-data -n $NAMESPACE"
        echo
    fi
    
    print_status "To verify cleanup, run:"
    echo "kubectl get all -n $NAMESPACE"
    echo "kubectl get pvc -n $NAMESPACE"
}

# Main uninstallation process
main() {
    print_header
    print_status "Starting Subnet Node uninstallation from K3s"
    print_status "Namespace: $NAMESPACE"
    print_status "Force mode: $FORCE"
    print_status "Keep data: $KEEP_DATA"
    echo

    # Check prerequisites
    check_kubectl
    check_namespace
    echo

    # Show deletion plan
    show_deletion_plan
    
    # Confirm deletion
    confirm_deletion
    echo

    # Delete resources
    delete_resources
    echo

    # Try to delete namespace
    delete_namespace
    echo

    # Show final status
    show_final_status
}

# Run main function
main "$@" 