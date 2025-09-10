#!/bin/bash

# Subnet Node K3s Installation Script
# This script installs Subnet Node on K3s with all necessary resources
#
# Options:
#   --private-key KEY     Private key for the node (required)
#   --provider-id ID      Provider ID (required)
#   --machine-id ID       Machine ID (required)
#   --namespace NS        Kubernetes namespace (default: subnet)
#   --image IMAGE         Docker image (default: subnet:latest)
#   --replicas N          Number of replicas (default: 1)
#   --storage-size SIZE   Storage size (default: 10Gi)
#   --api-port PORT       API port (default: 8080)
#   --swarm-port PORT     Swarm port (default: 4001)
#   --config-path PATH    Config file path (default: /data/config.yaml)
#   --data-dir DIR        Data directory (default: /data)
#   --help                Show this help message

set -e

# Default values
NAMESPACE="subnet"
IMAGE="subnet:latest"
REPLICAS=1
STORAGE_SIZE="10Gi"
API_PORT=8080
SWARM_PORT=4001
CONFIG_PATH="/data/config.yaml"
DATA_DIR="/data"
PRIVATE_KEY=""
PROVIDER_ID=""
MACHINE_ID=""

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
    echo -e "${BLUE}  Subnet Node K3s Installer${NC}"
    echo -e "${BLUE}================================${NC}"
}

# Function to show help
show_help() {
    cat << EOF
Subnet Node K3s Installation Script

Usage: $0 [OPTIONS]

Required Options:
  --private-key KEY     Private key for the node
  --provider-id ID      Provider ID
  --machine-id ID       Machine ID

Optional Options:
  --namespace NS        Kubernetes namespace (default: subnet)
  --image IMAGE         Docker image (default: subnet:latest)
  --replicas N          Number of replicas (default: 1)
  --storage-size SIZE   Storage size (default: 10Gi)
  --api-port PORT       API port (default: 8080)
  --swarm-port PORT     Swarm port (default: 4001)
  --config-path PATH    Config file path (default: /data/config.yaml)
  --data-dir DIR        Data directory (default: /data)
  --help                Show this help message

Examples:
  $0 --private-key "your-private-key" --provider-id "provider123" --machine-id "machine456"
  $0 --private-key "key" --provider-id "p123" --machine-id "m456" --namespace "my-subnet" --replicas 3

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --private-key)
            PRIVATE_KEY="$2"
            shift 2
            ;;
        --provider-id)
            PROVIDER_ID="$2"
            shift 2
            ;;
        --machine-id)
            MACHINE_ID="$2"
            shift 2
            ;;
        --namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        --image)
            IMAGE="$2"
            shift 2
            ;;
        --replicas)
            REPLICAS="$2"
            shift 2
            ;;
        --storage-size)
            STORAGE_SIZE="$2"
            shift 2
            ;;
        --api-port)
            API_PORT="$2"
            shift 2
            ;;
        --swarm-port)
            SWARM_PORT="$2"
            shift 2
            ;;
        --config-path)
            CONFIG_PATH="$2"
            shift 2
            ;;
        --data-dir)
            DATA_DIR="$2"
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

# Validate required parameters
if [[ -z "$PRIVATE_KEY" ]]; then
    print_error "Private key is required. Use --private-key option."
    exit 1
fi

if [[ -z "$PROVIDER_ID" ]]; then
    print_error "Provider ID is required. Use --provider-id option."
    exit 1
fi

if [[ -z "$MACHINE_ID" ]]; then
    print_error "Machine ID is required. Use --machine-id option."
    exit 1
fi

# Function to check if kubectl is installed
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    print_status "kubectl found: $(kubectl version --client --short)"
}

# Function to check if k3s is running
check_k3s() {
    if ! kubectl cluster-info &> /dev/null; then
        print_error "Cannot connect to Kubernetes cluster. Please ensure k3s is running."
        exit 1
    fi
    print_status "Connected to Kubernetes cluster: $(kubectl cluster-info | head -n 1)"
}

# Function to create namespace
create_namespace() {
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        print_status "Creating namespace: $NAMESPACE"
        kubectl create namespace "$NAMESPACE"
    else
        print_status "Namespace $NAMESPACE already exists"
    fi
}

# Function to create config map
create_config_map() {
    print_status "Creating ConfigMap for subnet configuration"
    
    cat << EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: subnet-config
  namespace: $NAMESPACE
data:
  config.yaml: |
    blockchain:
      rpc_url: "https://rpc.u2u.network"
      chain_id: 1000

    contracts:
      provider: "$PROVIDER_ID"
      bid_market: "0xabcdef123456789abcdef123456789abcdef1234"

    bidengine:
      min_bid_percent: 70
      max_bid_percent: 95
      price_factor: 1
      min_profit: 1000000000
      max_profit: 1000000000000000000
      max_concurrent_bids: 10
      bid_timeout: "30s"
      order_sync_interval: "30s"
      bid_check_interval: "60s"
      
      min_requirements:
        cpu_cores: 1
        memory_mb: 1024
        disk_gb: 10
        gpu_cores: 0
        upload_speed: 0
        download_speed: 0

      resource_cost:
        cpu_core: 1000000000000000
        memory_gb: 2000000000000000
        disk_gb: 100000000000000
        gpu_core: 5000000000000000
        upload_speed: 50000000000000
        download_speed: 50000000000000

      machine_type_multipliers:
        shared: 10
        vm: 15
        bare_metal: 25

      strategy:
        min_profit_margin: 0.05
        max_profit_margin: 0.20
        competitive_factor: 0.10
        market_adjustment: 0.05
        
        resource_weights:
          cpu: 0.25
          gpu: 0.35
          memory: 0.20
          disk: 0.15
          network: 0.05

    account:
      keystore_path: "/data/keystore"
      keystore_password: ""

    logging:
      level: "info"
      format: "text"
      output: "stdout"
      file_path: "/data/logs/subnet-node.log"
EOF
}

# Function to create secret for private key
create_secret() {
    print_status "Creating Secret for private key"
    
    cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: subnet-private-key
  namespace: $NAMESPACE
type: Opaque
data:
  private-key: $(echo -n "$PRIVATE_KEY" | base64)
EOF
}

# Function to create persistent volume claim
create_pvc() {
    print_status "Creating PersistentVolumeClaim"
    
    cat << EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: subnet-data
  namespace: $NAMESPACE
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: $STORAGE_SIZE
EOF
}

# Function to create deployment
create_deployment() {
    print_status "Creating Subnet Node Deployment"
    
    cat << EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: subnet-node
  namespace: $NAMESPACE
  labels:
    app: subnet-node
    provider-id: $PROVIDER_ID
    machine-id: $MACHINE_ID
spec:
  replicas: $REPLICAS
  selector:
    matchLabels:
      app: subnet-node
  template:
    metadata:
      labels:
        app: subnet-node
        provider-id: $PROVIDER_ID
        machine-id: $MACHINE_ID
    spec:
      containers:
      - name: subnet-node
        image: $IMAGE
        command: ["./subnet"]
        args:
        - "--datadir"
        - "$DATA_DIR"
        - "--init"
        ports:
        - containerPort: $API_PORT
          name: api
          protocol: TCP
        - containerPort: $SWARM_PORT
          name: swarm
          protocol: TCP
        env:
        - name: SUBNET_PROVIDER_ID
          value: "$PROVIDER_ID"
        - name: SUBNET_MACHINE_ID
          value: "$MACHINE_ID"
        - name: SUBNET_PRIVATE_KEY
          valueFrom:
            secretKeyRef:
              name: subnet-private-key
              key: private-key
        volumeMounts:
        - name: subnet-data
          mountPath: $DATA_DIR
        - name: subnet-config
          mountPath: /etc/subnet
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: $API_PORT
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: $API_PORT
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: subnet-data
        persistentVolumeClaim:
          claimName: subnet-data
      - name: subnet-config
        configMap:
          name: subnet-config
EOF
}

# Function to create service
create_service() {
    print_status "Creating Service"
    
    cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: subnet-node-service
  namespace: $NAMESPACE
  labels:
    app: subnet-node
spec:
  type: NodePort
  ports:
  - port: $API_PORT
    targetPort: $API_PORT
    protocol: TCP
    name: api
  - port: $SWARM_PORT
    targetPort: $SWARM_PORT
    protocol: TCP
    name: swarm
  selector:
    app: subnet-node
EOF
}

# Function to create ingress (optional)
create_ingress() {
    print_status "Creating Ingress (if ingress controller is available)"
    
    cat << EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: subnet-node-ingress
  namespace: $NAMESPACE
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: subnet-node.local
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: subnet-node-service
            port:
              number: $API_PORT
EOF
}

# Function to wait for deployment
wait_for_deployment() {
    print_status "Waiting for deployment to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/subnet-node -n "$NAMESPACE"
    print_status "Deployment is ready!"
}

# Function to show status
show_status() {
    print_status "Subnet Node deployment status:"
    echo
    kubectl get pods -n "$NAMESPACE" -l app=subnet-node
    echo
    kubectl get services -n "$NAMESPACE"
    echo
    print_status "To view logs: kubectl logs -f deployment/subnet-node -n $NAMESPACE"
    print_status "To access API: kubectl port-forward service/subnet-node-service $API_PORT:$API_PORT -n $NAMESPACE"
}

# Main installation process
main() {
    print_header
    print_status "Starting Subnet Node installation on K3s"
    print_status "Provider ID: $PROVIDER_ID"
    print_status "Machine ID: $MACHINE_ID"
    print_status "Namespace: $NAMESPACE"
    print_status "Image: $IMAGE"
    print_status "Replicas: $REPLICAS"
    echo

    # Check prerequisites
    check_kubectl
    check_k3s
    echo

    # Create resources
    create_namespace
    create_config_map
    create_secret
    create_pvc
    create_deployment
    create_service
    
    # Try to create ingress (may fail if no ingress controller)
    if create_ingress 2>/dev/null; then
        print_status "Ingress created successfully"
    else
        print_warning "Ingress creation failed (no ingress controller available)"
    fi
    
    echo
    wait_for_deployment
    echo
    show_status
    echo
    print_status "Subnet Node installation completed successfully!"
}

# Run main function
main "$@" 