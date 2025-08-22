#!/bin/bash

# Build script for k8s-services Docker image (local registry friendly)

set -euo pipefail

IMAGE_NAME=${IMAGE_NAME:-k8s-services}
TAG=${TAG:-latest}

# Determine repo root (two dirs up from this script)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

DOCKERFILE_PATH="${SCRIPT_DIR}/Dockerfile.k8s-services"

echo "Building ${IMAGE_NAME}:${TAG} using Dockerfile: ${DOCKERFILE_PATH}"

docker build \
  -f "${DOCKERFILE_PATH}" \
  -t "${IMAGE_NAME}:${TAG}" \
  "${REPO_ROOT}"

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


