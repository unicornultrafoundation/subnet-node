#!/bin/bash

# Local Docker Registry Management Script
# Usage: ./manage-local-registry.sh {start|stop|status|push|pull|list}

REGISTRY_NAME="local-registry"
REGISTRY_PORT="5000"
REGISTRY_URL="localhost:${REGISTRY_PORT}"

case "$1" in
    "start")
        echo "Checking if local Docker registry is already running..."
        if docker ps | grep -q ${REGISTRY_NAME}; then
            echo "✅ Registry is already running at http://${REGISTRY_URL}"
        else
            echo "Starting local Docker registry..."
            docker run -d -p ${REGISTRY_PORT}:5000 --name ${REGISTRY_NAME} registry:2
            echo "Registry started at http://${REGISTRY_URL}"
        fi
        ;;
    "stop")
        echo "Stopping local Docker registry..."
        docker stop ${REGISTRY_NAME}
        docker rm ${REGISTRY_NAME}
        echo "Registry stopped and removed"
        ;;
    "status")
        echo "Checking registry status..."
        if docker ps | grep -q ${REGISTRY_NAME}; then
            echo "✅ Registry is running at http://${REGISTRY_URL}"
            echo "Repositories:"
            curl -s http://${REGISTRY_URL}/v2/_catalog | jq -r '.repositories[]' 2>/dev/null || curl -s http://${REGISTRY_URL}/v2/_catalog
        else
            echo "❌ Registry is not running"
        fi
        ;;
    "push")
        if [ -z "$2" ]; then
            echo "Usage: $0 push <image-name> [tag]"
            echo "Example: $0 push k8s-services latest"
            exit 1
        fi
        IMAGE_NAME=$2
        TAG=${3:-latest}
        echo "Pushing ${IMAGE_NAME}:${TAG} to local registry..."
        docker tag ${IMAGE_NAME}:${TAG} ${REGISTRY_URL}/${IMAGE_NAME}:${TAG}
        docker push ${REGISTRY_URL}/${IMAGE_NAME}:${TAG}
        echo "✅ Image pushed successfully"
        ;;
    "pull")
        if [ -z "$2" ]; then
            echo "Usage: $0 pull <image-name> [tag]"
            echo "Example: $0 pull k8s-services latest"
            exit 1
        fi
        IMAGE_NAME=$2
        TAG=${3:-latest}
        echo "Pulling ${IMAGE_NAME}:${TAG} from local registry..."
        docker pull ${REGISTRY_URL}/${IMAGE_NAME}:${TAG}
        echo "✅ Image pulled successfully"
        ;;
    "list")
        echo "Repositories in local registry:"
        curl -s http://${REGISTRY_URL}/v2/_catalog | jq -r '.repositories[]' 2>/dev/null || curl -s http://${REGISTRY_URL}/v2/_catalog
        ;;
    *)
        echo "Local Docker Registry Management Script"
        echo ""
        echo "Usage: $0 {start|stop|status|push|pull|list}"
        echo ""
        echo "Commands:"
        echo "  start   - Start the local registry"
        echo "  stop    - Stop and remove the local registry"
        echo "  status  - Check registry status"
        echo "  push    - Push an image to the registry"
        echo "  pull    - Pull an image from the registry"
        echo "  list    - List repositories in the registry"
        echo ""
        echo "Examples:"
        echo "  $0 start"
        echo "  $0 push k8s-services latest"
        echo "  $0 pull k8s-services latest"
        echo "  $0 status"
        ;;
esac 