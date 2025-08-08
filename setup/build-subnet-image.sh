#!/bin/bash

# Subnet Node Docker Image Builder
# This script builds Docker image for Subnet Node
#
# Options:
#   --tag TAG           Docker image tag (default: subnet:latest)
#   --platform PLATFORM Target platform (default: linux/amd64)
#   --push              Push image to registry after building
#   --registry REGISTRY Registry URL (default: none)
#   --help              Show this help message

set -e

# Default values
TAG="subnet:latest"
PLATFORM="linux/amd64"
PUSH=false
REGISTRY=""

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
    echo -e "${BLUE}  Subnet Node Image Builder${NC}"
    echo -e "${BLUE}================================${NC}"
}

# Function to show help
show_help() {
    cat << EOF
Subnet Node Docker Image Builder

Usage: $0 [OPTIONS]

Options:
  --tag TAG           Docker image tag (default: subnet:latest)
  --platform PLATFORM Target platform (default: linux/amd64)
  --push              Push image to registry after building
  --registry REGISTRY Registry URL (default: none)
  --help              Show this help message

Examples:
  $0 --tag subnet:v1.0.0
  $0 --tag my-registry.com/subnet:latest --push --registry my-registry.com
  $0 --platform linux/arm64 --tag subnet:arm64

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --tag)
            TAG="$2"
            shift 2
            ;;
        --platform)
            PLATFORM="$2"
            shift 2
            ;;
        --push)
            PUSH=true
            shift
            ;;
        --registry)
            REGISTRY="$2"
            shift 2
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

# Function to check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    print_status "Docker found: $(docker --version)"
}

# Function to check if Docker daemon is running
check_docker_daemon() {
    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running. Please start Docker first."
        exit 1
    fi
    print_status "Docker daemon is running"
}

# Function to build the image
build_image() {
    print_status "Building Subnet Node Docker image..."
    print_status "Tag: $TAG"
    print_status "Platform: $PLATFORM"
    
    # Build the image
    docker build \
        --platform "$PLATFORM" \
        --tag "$TAG" \
        --file Dockerfile \
        .
    
    print_status "Image built successfully!"
}

# Function to push the image
push_image() {
    if [[ "$PUSH" == true ]]; then
        if [[ -z "$REGISTRY" ]]; then
            print_error "Registry URL is required when using --push option"
            exit 1
        fi
        
        print_status "Pushing image to registry: $REGISTRY"
        docker push "$TAG"
        print_status "Image pushed successfully!"
    fi
}

# Function to show image info
show_image_info() {
    print_status "Image information:"
    docker images "$TAG"
    echo
    print_status "To run the image locally:"
    echo "docker run -d --name subnet-node -p 8080:8080 -p 4001:4001 $TAG"
}

# Main build process
main() {
    print_header
    print_status "Starting Subnet Node Docker image build"
    print_status "Tag: $TAG"
    print_status "Platform: $PLATFORM"
    print_status "Push: $PUSH"
    if [[ -n "$REGISTRY" ]]; then
        print_status "Registry: $REGISTRY"
    fi
    echo

    # Check prerequisites
    check_docker
    check_docker_daemon
    echo

    # Build image
    build_image
    echo

    # Push image if requested
    if [[ "$PUSH" == true ]]; then
        push_image
        echo
    fi

    # Show image info
    show_image_info
    echo
    print_status "Docker image build completed successfully!"
}

# Run main function
main "$@" 