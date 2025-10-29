#!/bin/bash

# Subnet Node K3s Complete Setup Script
# This script will:
# 1. Build the Docker image (if needed)
# 2. Import private key into subnet account using import-hex
# 3. Create keystore and store in ConfigMap/Secret
# 4. Install Subnet Node on K3s
# 5. Configure all necessary resources
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
#   --keystore-password PASSWORD Password for keystore (default: empty)
#   --skip-build          Skip Docker image build
#   --build-only          Only build Docker image, don't install
#   --install-only        Only install, don't build Docker image
#   --help                Show this help message

set -e

# Default values
NAMESPACE="subnet"
IMAGE="subnet:latest"
REPLICAS=1
STORAGE_SIZE="10Gi"
API_PORT=8080
SWARM_PORT=4001
KEYSTORE_PASSWORD=""
PRIVATE_KEY=""
PROVIDER_ID=""
MACHINE_ID=""
SKIP_BUILD=false
BUILD_ONLY=false
INSTALL_ONLY=false

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
    echo -e "${BLUE}  Subnet Node K3s Setup${NC}"
    echo -e "${BLUE}================================${NC}"
}

# Function to show help
show_help() {
    cat << EOF
Subnet Node K3s Complete Setup Script

This script will build the Docker image and install Subnet Node on K3s.

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
  --keystore-password PASSWORD Password for keystore (default: empty)
  --skip-build          Skip Docker image build
  --build-only          Only build Docker image, don't install
  --install-only        Only install, don't build Docker image
  --help                Show this help message

Examples:
  $0 --private-key "your-private-key" --provider-id "provider123" --machine-id "machine456"
  $0 --private-key "key" --provider-id "p123" --machine-id "m456" --keystore-password "mypassword"
  $0 --build-only --image "my-registry.com/subnet:v1.0.0"

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
        --keystore-password)
            KEYSTORE_PASSWORD="$2"
            shift 2
            ;;
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        --build-only)
            BUILD_ONLY=true
            shift
            ;;
        --install-only)
            INSTALL_ONLY=true
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

# Validate required parameters for installation
validate_install_params() {
    if [[ -z "$PRIVATE_KEY" ]]; then
        print_error "Private key is required for installation. Use --private-key option."
        exit 1
    fi

    if [[ -z "$PROVIDER_ID" ]]; then
        print_error "Provider ID is required for installation. Use --provider-id option."
        exit 1
    fi

    if [[ -z "$MACHINE_ID" ]]; then
        print_error "Machine ID is required for installation. Use --machine-id option."
        exit 1
    fi
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if Docker is installed (for building)
    if [[ "$SKIP_BUILD" == false && "$BUILD_ONLY" == true ]]; then
        if ! command -v docker &> /dev/null; then
            print_error "Docker is not installed. Please install Docker first."
            exit 1
        fi
        print_status "Docker found: $(docker --version)"
    fi
    
    # Check if kubectl is installed (for installation)
    if [[ "$BUILD_ONLY" == false ]]; then
        if ! command -v kubectl &> /dev/null; then
            print_error "kubectl is not installed. Please install kubectl first."
            exit 1
        fi
        print_status "kubectl found: $(kubectl version --client --short)"
        
        # Check if k3s is running
        if ! kubectl cluster-info &> /dev/null; then
            print_error "Cannot connect to Kubernetes cluster. Please ensure k3s is running."
            exit 1
        fi
        print_status "Connected to Kubernetes cluster: $(kubectl cluster-info | head -n 1)"
    fi
}

# Function to build Docker image
build_image() {
    if [[ "$SKIP_BUILD" == true ]]; then
        print_warning "Skipping Docker image build (--skip-build flag)"
        return 0
    fi
    
    print_status "Building Docker image: $IMAGE"
    
    # Check if image already exists
    if docker images "$IMAGE" --format "table {{.Repository}}:{{.Tag}}" | grep -q "$IMAGE"; then
        print_warning "Image $IMAGE already exists. Use --skip-build to skip building."
        echo -n "Do you want to rebuild the image? (y/N): "
        read -r response
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            print_status "Using existing image: $IMAGE"
            return 0
        fi
    fi
    
    # Build the image
    docker build \
        --platform "linux/amd64" \
        --tag "$IMAGE" \
        --file Dockerfile \
        .
    
    print_status "Docker image built successfully!"
}

# Function to create keystore from private key
create_keystore() {
    if [[ "$BUILD_ONLY" == true ]]; then
        print_warning "Skipping keystore creation (--build-only flag)"
        return 0
    fi
    
    print_status "Creating keystore from private key..."
    
    # Create temporary directory for keystore
    TEMP_DIR=$(mktemp -d)
    KEYSTORE_DIR="$TEMP_DIR/keystore"
    mkdir -p "$KEYSTORE_DIR"
    
    # Create a temporary subnet binary or use existing one
    SUBNET_BIN=""
    if [[ -f "./build/subnet" ]]; then
        SUBNET_BIN="./build/subnet"
    elif [[ -f "./subnet" ]]; then
        SUBNET_BIN="./subnet"
    else
        print_error "Subnet binary not found. Please build the project first."
        exit 1
    fi
    
    # Make sure subnet binary is executable
    chmod +x "$SUBNET_BIN"
    
    # Initialize subnet with temporary datadir
    print_status "Initializing subnet node..."
    "$SUBNET_BIN" --datadir "$TEMP_DIR" --init
    
    # Import private key using import-hex command
    print_status "Importing private key into keystore..."
    
    # Create a temporary file with the password
    PASSWORD_FILE="$TEMP_DIR/password.txt"
    echo "$KEYSTORE_PASSWORD" > "$PASSWORD_FILE"
    
    # Import the private key
    if [[ -n "$KEYSTORE_PASSWORD" ]]; then
        echo "$KEYSTORE_PASSWORD" | "$SUBNET_BIN" --datadir "$TEMP_DIR" account import-hex "$PRIVATE_KEY"
    else
        # If no password provided, use empty password
        echo "" | "$SUBNET_BIN" --datadir "$TEMP_DIR" account import-hex "$PRIVATE_KEY"
    fi
    
    # Check if keystore was created
    if [[ ! -d "$KEYSTORE_DIR" ]] || [[ -z "$(ls -A "$KEYSTORE_DIR" 2>/dev/null)" ]]; then
        print_error "Failed to create keystore. Please check your private key format."
        rm -rf "$TEMP_DIR"
        exit 1
    fi
    
    # Get the keystore file
    KEYSTORE_FILE=$(ls "$KEYSTORE_DIR"/*.json | head -n 1)
    if [[ ! -f "$KEYSTORE_FILE" ]]; then
        print_error "Keystore file not found after import."
        rm -rf "$TEMP_DIR"
        exit 1
    fi
    
    print_status "Keystore created successfully: $(basename "$KEYSTORE_FILE")"
    
    # Store keystore file path for later use
    KEYSTORE_CONTENT=$(cat "$KEYSTORE_FILE")
    
    # Clean up temporary directory
    rm -rf "$TEMP_DIR"
}

# Function to install on K3s
install_on_k3s() {
    if [[ "$BUILD_ONLY" == true ]]; then
        print_warning "Skipping installation (--build-only flag)"
        return 0
    fi
    
    print_status "Installing Subnet Node on K3s..."
    
    # Create namespace
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        print_status "Creating namespace: $NAMESPACE"
        kubectl create namespace "$NAMESPACE"
    else
        print_status "Namespace $NAMESPACE already exists"
    fi
    
    # Create ConfigMap for subnet configuration
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

    # Create ConfigMap for keystore
    if [[ -n "$KEYSTORE_CONTENT" ]]; then
        print_status "Creating ConfigMap for keystore"
        cat << EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: subnet-keystore
  namespace: $NAMESPACE
data:
  keystore.json: |
$(echo "$KEYSTORE_CONTENT" | sed 's/^/    /')
EOF
    fi

    # Create Secret for keystore password
    print_status "Creating Secret for keystore password"
    cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: subnet-keystore-password
  namespace: $NAMESPACE
type: Opaque
data:
  password: $(echo -n "$KEYSTORE_PASSWORD" | base64)
EOF

    # Create PVC
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

    # Create Deployment
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
        - "/data"
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
        - name: SUBNET_KEYSTORE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: subnet-keystore-password
              key: password
        volumeMounts:
        - name: subnet-data
          mountPath: /data
        - name: subnet-config
          mountPath: /etc/subnet
        - name: subnet-keystore
          mountPath: /data/keystore
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
      - name: subnet-keystore
        configMap:
          name: subnet-keystore
EOF

    # Create Service
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

    # Wait for deployment
    print_status "Waiting for deployment to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/subnet-node -n "$NAMESPACE"
    print_status "Deployment is ready!"
}

# Function to show final status
show_final_status() {
    if [[ "$BUILD_ONLY" == false ]]; then
        print_status "Subnet Node deployment status:"
        echo
        kubectl get pods -n "$NAMESPACE" -l app=subnet-node
        echo
        kubectl get services -n "$NAMESPACE"
        echo
        print_status "Useful commands:"
        echo "  View logs: kubectl logs -f deployment/subnet-node -n $NAMESPACE"
        echo "  Access API: kubectl port-forward service/subnet-node-service $API_PORT:$API_PORT -n $NAMESPACE"
        echo "  Check status: kubectl get all -n $NAMESPACE"
        echo "  Uninstall: ./uninstall-subnet-k3s.sh --namespace $NAMESPACE"
    fi
}

# Main setup process
main() {
    print_header
    
    if [[ "$BUILD_ONLY" == true ]]; then
        print_status "Build-only mode: Building Docker image only"
        print_status "Image: $IMAGE"
    elif [[ "$INSTALL_ONLY" == true ]]; then
        print_status "Install-only mode: Installing on K3s only"
        validate_install_params
        print_status "Provider ID: $PROVIDER_ID"
        print_status "Machine ID: $MACHINE_ID"
        print_status "Namespace: $NAMESPACE"
        print_status "Image: $IMAGE"
    else
        print_status "Full setup mode: Building image and installing on K3s"
        validate_install_params
        print_status "Provider ID: $PROVIDER_ID"
        print_status "Machine ID: $MACHINE_ID"
        print_status "Namespace: $NAMESPACE"
        print_status "Image: $IMAGE"
        print_status "Replicas: $REPLICAS"
    fi
    echo

    # Check prerequisites
    check_prerequisites
    echo

    # Build image
    build_image
    echo

    # Create keystore
    create_keystore
    echo

    # Install on K3s
    install_on_k3s
    echo

    # Show final status
    show_final_status
    echo
    
    if [[ "$BUILD_ONLY" == true ]]; then
        print_status "Docker image build completed successfully!"
    elif [[ "$INSTALL_ONLY" == true ]]; then
        print_status "Subnet Node installation completed successfully!"
    else
        print_status "Subnet Node setup completed successfully!"
    fi
}

# Run main function
main "$@" 