# k8s-services Local Build & Registry Scripts

This package helps build the `k8s-services` image locally and run it via a local Docker registry when a remote registry image is unavailable.

## Prerequisites

- Kubernetes cluster (e.g., minikube, kind, or production cluster)
- `kubectl` configured and working
- Docker installed and running
- Go 1.24+ (for building the application)

## Files

- `.dockerignore.k8s-services`: Docker ignore rules for the build context (reference file)
- `Dockerfile.k8s-services`: Multi-stage Dockerfile to build the `k8s-services` binary and image
- `build-k8s-services.sh`: Helper to build the Docker image from the repository root context

## Quick Start

```bash
# 1) Start local registry (from repo root)
./manage-local-registry.sh start
./manage-local-registry.sh status

# 2) Build the image locally (uses repo-root build context)
script/k8s-services/build-k8s-services.sh

# 3) Push to local registry
./manage-local-registry.sh push k8s-services latest

# 4) Apply network policies and operator manifests
kubectl apply -f pkg/k8s/kustomize/subnet-services/network-policies.yaml
kubectl apply -k pkg/k8s/kustomize/subnet-operator-inventory/

# 5) Verify
kubectl get all -n subnet-services
kubectl logs -n subnet-services deployment/operator-inventory --tail=20
```

## Notes

- Dockerfile path: `script/k8s-services/Dockerfile.k8s-services`
- Build context: repository root
- Ensure manifests reference `k8s-services:latest` (or your tag) and imagePullPolicy allows using locally built images.

## Troubleshooting

- Image Pull Errors:
  - Ensure local registry is running: `./manage-local-registry.sh status`
  - Check image exists: `./manage-local-registry.sh list`
- Network Policy Issues:
  - Verify policies: `kubectl get networkpolicies -n subnet-services`
