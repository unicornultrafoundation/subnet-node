#!/bin/bash

# Build script for k8s-services Docker image

set -e

IMAGE_NAME="k8s-services"
TAG="latest"

echo "Building k8s-services Docker image..."

# Build the Docker image
docker build -f Dockerfile.k8s-services -t ${IMAGE_NAME}:${TAG} .

echo "Build completed successfully!"
echo ""
echo "To run the container:"
echo "  docker run ${IMAGE_NAME}:${TAG}"
echo ""
echo "To run with specific commands (like the debug configuration):"
echo "  docker run ${IMAGE_NAME}:${TAG} tools psutil serve"
echo ""
echo "To run interactively:"
echo "  docker run -it ${IMAGE_NAME}:${TAG} /bin/bash" 