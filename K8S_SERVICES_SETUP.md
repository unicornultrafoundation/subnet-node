# K8S Services Setup Guide

This document provides a complete step-by-step guide to set up the k8s services infrastructure from scratch.

## Prerequisites

- Kubernetes cluster (e.g., minikube, kind, or production cluster)
- `kubectl` configured and working
- Docker installed and running
- Go 1.24+ (for building the application)
- `kustomize` (optional, for applying manifests)

## Overview

The k8s services consist of:
1. **Inventory Operator**: Main operator that manages hardware discovery
2. **Hardware Discovery Pods**: Dynamic pods that discover CPU, GPU, and memory information
3. **Local Docker Registry**: For storing and serving container images
4. **Network Policies**: For controlling pod-to-pod communication

## Step-by-Step Setup

### 1. Create Namespace

```bash
kubectl create namespace subnet-services
```

### 2. Set Up Local Docker Registry

The local registry is used to store the `k8s-services` Docker image.

```bash
# Start the local registry
./manage-local-registry.sh start

# Verify it's running
./manage-local-registry.sh status
```

**Registry Management Commands:**
- `./manage-local-registry.sh start` - Start the registry
- `./manage-local-registry.sh stop` - Stop the registry
- `./manage-local-registry.sh status` - Check registry status
- `./manage-local-registry.sh push <image> <tag>` - Push image to registry
- `./manage-local-registry.sh pull <image> <tag>` - Pull image from registry
- `./manage-local-registry.sh list` - List repositories

### 3. Build the k8s-services Docker Image

```bash
# Build the Docker image
./build-k8s-services.sh

# Push to local registry
./manage-local-registry.sh push k8s-services latest
```

**Build Script Details:**
- Uses multi-stage Docker build
- Stage 1: Builds Go binary using `golang:1.24`
- Stage 2: Creates minimal runtime container using `ubuntu:22.04`
- Binary is built from `./pkg/k8s/cmd/k8s-services`
- Image is tagged as `k8s-services:latest`

### 4. Apply Network Policies

Create the base network policies for the `subnet-services` namespace:

```bash
kubectl apply -f pkg/k8s/kustomize/subnet-services/network-policies.yaml
```

**Network Policies Applied:**
- `subnet-services-default-deny-ingress`: Denies all ingress traffic by default
- `subnet-services-allow-subnet-services`: Allows traffic between pods in the same namespace
- `subnet-services-allow-ingress-nginx`: Allows traffic from ingress-nginx

### 5. Deploy the Inventory Operator

Apply all the Kubernetes manifests for the inventory operator:

```bash
# Apply using kustomize (recommended)
kubectl apply -k pkg/k8s/kustomize/subnet-operator-inventory/

# Or apply individual files
kubectl apply -f pkg/k8s/kustomize/subnet-operator-inventory/service-accounts.yaml
kubectl apply -f pkg/k8s/kustomize/subnet-operator-inventory/cluster-roles.yaml
kubectl apply -f pkg/k8s/kustomize/subnet-operator-inventory/role-bindings.yaml
kubectl apply -f pkg/k8s/kustomize/subnet-operator-inventory/service.yaml
kubectl apply -f pkg/k8s/kustomize/subnet-operator-inventory/deployment.yaml
```

**Resources Created:**
- **Service Accounts:**
  - `operator-inventory`: For the main inventory operator
  - `operator-inventory-hardware-discovery`: For hardware discovery pods

- **Cluster Roles:**
  - `subnet-operator-inventory`: Full permissions for inventory management
  - `subnet-operator-inventory-hardware-discovery`: Limited permissions for hardware discovery

- **Cluster Role Bindings:**
  - `operator-inventory`: Binds operator service account to cluster role
  - `operator-inventory-hardware-discovery`: Binds hardware discovery service account to cluster role

- **Service:**
  - `operator-inventory`: ClusterIP service exposing ports 8080 (REST) and 8081 (gRPC)

- **Deployment:**
  - `operator-inventory`: Runs the main inventory operator

### 6. Create Additional Network Policy for Proxy Access

Create a network policy to allow the Kubernetes proxy to access hardware discovery pods:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-proxy-access
  namespace: subnet-services
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/instance: inventory-hardware-discovery
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: kube-system
    ports:
    - protocol: TCP
      port: 8081
```

Apply this policy:
```bash
kubectl apply -f allow-proxy-access.yaml
```

### 7. Verify Deployment

Check that all components are running correctly:

```bash
# Check namespace
kubectl get namespace subnet-services

# Check all resources in the namespace
kubectl get all -n subnet-services

# Check service accounts
kubectl get serviceaccount -n subnet-services

# Check network policies
kubectl get networkpolicies -n subnet-services

# Check cluster roles and bindings
kubectl get clusterrole,clusterrolebinding | grep inventory

# Check pod logs
kubectl logs -n subnet-services deployment/operator-inventory --tail=20
```

### 8. Test the Setup

Test that the inventory operator is working correctly:

```bash
# Test CPU discovery
kubectl get --raw "/api/v1/namespaces/subnet-services/pods/operator-inventory-hardware-discovery-colima-provider:8081/proxy/cpu"

# Test GPU discovery
kubectl get --raw "/api/v1/namespaces/subnet-services/pods/operator-inventory-hardware-discovery-colima-provider:8081/proxy/gpu"

# Test Memory discovery
kubectl get --raw "/api/v1/namespaces/subnet-services/pods/operator-inventory-hardware-discovery-colima-provider:8081/proxy/memory"
```

## Updating the Services

When you make changes to the code, follow this workflow:

### 1. Build and Push Updated Image

```bash
# Build the updated Docker image
./build-k8s-services.sh

# Push to local registry
./manage-local-registry.sh push k8s-services latest
```

### 2. Restart the Deployment

```bash
# Restart the inventory operator deployment
kubectl rollout restart deployment operator-inventory -n subnet-services

# Monitor the rollout
kubectl rollout status deployment operator-inventory -n subnet-services

# Verify the new pod is running
kubectl get pods -n subnet-services | grep operator-inventory

# Check logs
kubectl logs -n subnet-services deployment/operator-inventory --tail=20
```

## Troubleshooting

### Common Issues

1. **Pod in CrashLoopBackOff:**
   ```bash
   kubectl describe pod <pod-name> -n subnet-services
   kubectl logs <pod-name> -n subnet-services
   ```

2. **Image Pull Errors:**
   - Ensure local registry is running: `./manage-local-registry.sh status`
   - Check image exists: `./manage-local-registry.sh list`

3. **Network Policy Issues:**
   - Verify network policies are applied: `kubectl get networkpolicies -n subnet-services`
   - Check if proxy access policy is needed for hardware discovery

4. **Permission Issues:**
   - Verify service accounts and cluster roles: `kubectl get serviceaccount,clusterrole,clusterrolebinding -A | grep inventory`

### Debugging Commands

```bash
# Get detailed pod information
kubectl describe pod <pod-name> -n subnet-services

# Check events in namespace
kubectl get events -n subnet-services --sort-by='.lastTimestamp'

# Port forward to debug directly
kubectl port-forward -n subnet-services <pod-name> 8082:8081

# Test connectivity
curl localhost:8082/cpu
```

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                       │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────────────────────┐ │
│  │ Local Registry  │    │      subnet-services            │ │
│  │ localhost:5000  │    │                                 │ │
│  └─────────────────┘    │  ┌─────────────────────────────┐ │ │
│                         │  │   operator-inventory        │ │ │
│                         │  │   (Deployment)              │ │ │
│                         │  │   - REST API (8080)         │ │ │
│                         │  │   - gRPC API (8081)         │ │ │
│                         │  └─────────────────────────────┘ │ │
│                         │                                 │ │
│                         │  ┌─────────────────────────────┐ │ │
│                         │  │ hardware-discovery-*        │ │ │
│                         │  │ (Dynamic Pods)              │ │ │
│                         │  │ - CPU Discovery             │ │ │
│                         │  │ - GPU Discovery             │ │ │
│                         │  │ - Memory Discovery          │ │ │
│                         │  └─────────────────────────────┘ │ │
│                         └─────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## File Structure

```
pkg/k8s/
├── cmd/k8s-services/           # Main application entry point
├── operator/inventory/         # Inventory operator logic
│   ├── cmd.go                  # Command definitions
│   ├── node-discovery.go       # Hardware discovery logic
│   └── gpus.go                 # GPU registry (converted from JSON)
├── kustomize/                  # Kubernetes manifests
│   ├── subnet-services/        # Base namespace resources
│   └── subnet-operator-inventory/  # Operator resources
└── tools/                      # Utility tools

Scripts:
├── build-k8s-services.sh       # Docker build script
├── manage-local-registry.sh    # Registry management
└── Dockerfile.k8s-services     # Multi-stage Docker build
```

## Security Considerations

1. **Network Policies**: Restrict pod-to-pod communication
2. **RBAC**: Use least-privilege service accounts and roles
3. **Image Security**: Use local registry for controlled image distribution
4. **Pod Security**: Run containers with appropriate security contexts

## Monitoring and Logging

The inventory operator provides:
- REST API on port 8080 for metrics and health checks
- gRPC API on port 8081 for internal communication
- Structured logging with logrus
- Kubernetes events for deployment tracking

## Next Steps

After setup, you can:
1. Integrate with external monitoring systems
2. Add custom hardware discovery plugins
3. Implement additional inventory types
4. Scale the operator for multi-node clusters 