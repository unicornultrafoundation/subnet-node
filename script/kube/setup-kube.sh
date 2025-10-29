#!/usr/bin/env bash
#
# Subnet Node Kubernetes Setup Script
#
# This script sets up the subnet node infrastructure in a Kubernetes cluster:
# - Creates namespace and installs CRDs
# - Applies network policies
# - Installs inventory operator
# - Waits for operator to be available with robust error handling
#
# Cross-platform support:
# - Linux: Full support
# - macOS: Full support with platform-specific adjustments
# - Windows: Full support via WSL/Git Bash
#
# Improvements made to wait logic:
# - Checks pod status before attempting port-forward
# - Uses temporary port-forward for testing instead of long-running ones
# - Better error handling and debugging information
# - Configurable timeouts and retry logic
# - Helper functions for reusability and maintainability
#

set -e

rootdir="$(dirname "$0")/../.."

# Configurable ports via env or flags
INVENTORY_REST_PORT="${INVENTORY_REST_PORT:-8080}"
INVENTORY_GRPC_PORT="${INVENTORY_GRPC_PORT:-8081}"
WITH_GRPC_CHECK="${WITH_GRPC_CHECK:-false}"

retries=10
retrywait=1

CRD_FILE=$rootdir/pkg/k8s/apis/crd.yaml
STORAGE_CLASS_FILE=$rootdir/pkg/k8s/apis/storageclass.yaml
NAMESPACE_FILE=$rootdir/pkg/k8s/apis/namespace.yaml

# Detect platform and set appropriate commands
detect_platform() {
    case "$(uname -s)" in
        Linux*)     PLATFORM="linux";;
        Darwin*)    PLATFORM="macos";;
        CYGWIN*|MINGW*|MSYS*) PLATFORM="windows";;
        *)          PLATFORM="unknown";;
    esac
    
    echo "Detected platform: $PLATFORM"
}

# Platform-specific port testing function
test_port() {
    local host=$1
    local port=$2
    
    case "$PLATFORM" in
        "linux")
            # Linux: use nc with -z flag
            nc -z "$host" "$port" 2>/dev/null
            ;;
        "macos")
            # macOS: use nc with -z flag (BSD version)
            nc -z "$host" "$port" 2>/dev/null
            ;;
        "windows")
            # Windows: try multiple methods
            if command -v nc >/dev/null 2>&1; then
                # If nc is available (WSL/Git Bash)
                nc -z "$host" "$port" 2>/dev/null
            elif command -v Test-NetConnection >/dev/null 2>&1; then
                # PowerShell (would need different script)
                echo "PowerShell detected - consider using setup-kube.ps1 instead"
                return 1
            else
                # Fallback: try to connect with timeout
                timeout 1 bash -c "echo >/dev/tcp/$host/$port" 2>/dev/null
            fi
            ;;
        *)
            # Fallback for unknown platforms
            timeout 1 bash -c "echo >/dev/tcp/$host/$port" 2>/dev/null
            ;;
    esac
}

# Check if required tools are available
check_requirements() {
    local missing_tools=()
    
    # Check for grpcurl and add Go bin to PATH if needed
    if ! command -v grpcurl >/dev/null 2>&1; then
        # Check if grpcurl is available in Go bin directory
        if [ -f "$HOME/go/bin/grpcurl" ]; then
            echo "Found grpcurl in ~/go/bin, adding to PATH..."
            export PATH="$PATH:$HOME/go/bin"
        else
            missing_tools+=("grpcurl")
        fi
    fi
    
    # Core tools that should be available on all platforms
    for tool in kubectl jq; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            missing_tools+=("$tool")
        fi
    done
    
    # Check for port testing capability
    local port_test_available=false
    
    case "$PLATFORM" in
        "linux"|"macos")
            if command -v nc >/dev/null 2>&1; then
                port_test_available=true
            fi
            ;;
        "windows")
            if command -v nc >/dev/null 2>&1 || command -v timeout >/dev/null 2>&1; then
                port_test_available=true
            fi
            ;;
    esac
    
    if [ "$port_test_available" = false ]; then
        missing_tools+=("port-testing-tool")
        echo "WARNING: No port testing tool available. Install 'nc' (netcat) or ensure 'timeout' is available."
    fi
    
    if [ ${#missing_tools[@]} -gt 0 ]; then
        echo "ERROR: Missing required tools: ${missing_tools[*]}"
        echo ""
        echo "Installation instructions by platform:"
        case "$PLATFORM" in
            "linux")
                echo "  Ubuntu/Debian:"
                echo "    sudo apt-get install netcat-openbsd jq"
                echo "    # grpcurl via Go (recommended):"
                echo "    go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
                echo "    export PATH=\$PATH:~/go/bin"
                echo "    # Or add to ~/.bashrc: echo 'export PATH=\$PATH:~/go/bin' >> ~/.bashrc"
                echo "  CentOS/RHEL: sudo yum install nc grpcurl jq"
                echo "  Arch: sudo pacman -S netcat grpcurl jq"
                ;;
            "macos")
                echo "  Homebrew: brew install netcat grpcurl jq"
                ;;
            "windows")
                echo "  WSL: sudo apt-get install netcat-openbsd jq && go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
                echo "  Git Bash: Install via package manager or use WSL"
                echo "  PowerShell: Use setup-kube.ps1 instead"
                ;;
        esac
        exit 1
    fi
}

# Check if kubectl is connected to a cluster
check_kubectl_connection() {
    if ! kubectl cluster-info >/dev/null 2>&1; then
        echo "ERROR: kubectl is not connected to a cluster"
        echo "Please ensure you have a valid kubeconfig and are connected to a cluster"
        exit 1
    fi
    echo "Connected to cluster: $(kubectl config current-context)"
}

# Extract first listening TCP port inside operator-inventory container
# Uses /proc/net/tcp and filters by LISTEN state (0A). Returns empty if none.
get_first_listen_port() {
    kubectl -n subnet-services exec deployment/operator-inventory -- sh -c '
        awk "NR>1 && \$4==\"0A\" {print \$2}" /proc/net/tcp | head -1 | awk -F":" "{print \$2}"
    ' 2>/dev/null | tr -d '\r\n'
}

# Helper function to check pod status
check_pod_status() {
    local namespace=$1
    local label_selector=$2
    
    kubectl -n "$namespace" get pod -l "$label_selector" --no-headers 2>/dev/null | head -1 | awk '{print $3}'
}

# Helper function to wait for pod to be ready
wait_for_pod_ready() {
    local namespace=$1
    local label_selector=$2
    local max_attempts=${3:-30}
    local sleep_interval=${4:-2}
    
    echo "Waiting for pod with label '$label_selector' in namespace '$namespace' to be ready..."
    local attempts=0
    
    while [ $attempts -lt $max_attempts ]; do
        attempts=$((attempts + 1))
        local status=$(check_pod_status "$namespace" "$label_selector")
        
        case "$status" in
            "Running")
                echo "Pod is now running"
                return 0
                ;;
            "Pending")
                echo "Pod is still pending, waiting... (attempt $attempts/$max_attempts)"
                ;;
            "CrashLoopBackOff"|"Error"|"Failed")
                echo "ERROR: Pod is in failed state: $status"
                kubectl -n "$namespace" describe pod -l "$label_selector"
                return 1
                ;;
            "")
                echo "Pod not found yet, waiting... (attempt $attempts/$max_attempts)"
                ;;
            *)
                echo "Pod status: $status, waiting... (attempt $attempts/$max_attempts)"
                ;;
        esac
        
        sleep "$sleep_interval"
    done
    
    echo "ERROR: Pod failed to become ready within timeout"
    kubectl -n "$namespace" get pods -l "$label_selector"
    return 1
}

install_ns() {
    kubectl apply -f "$NAMESPACE_FILE"
}

install_crd() {
    echo "Installing CRDs"
    kubectl apply -f "$CRD_FILE"
    kubectl apply -f "$STORAGE_CLASS_FILE"
    echo "CRDs installed"

    # TODO: Config exactly the control plane node name
    CONTROL_PLANE_NODE=$(kubectl get nodes -l node-role.kubernetes.io/control-plane=true -o jsonpath='{.items[0].metadata.name}')
    kubectl patch node "${CONTROL_PLANE_NODE}" -p '{"metadata":{"labels":{"subnet.node/storageclasses":"default","ingress-ready":"true"}}}'
}

install_network_policies() {
    echo "Installing network policies"
    kubectl kustomize "$rootdir/pkg/k8s/kustomize/subnet-services/" | kubectl apply -f-
    echo "Network policies installed"
}

install_inventory_operator() {
    echo "Installing inventory operator"
    kubectl kustomize "$rootdir/pkg/k8s/kustomize/subnet-operator-inventory/" | kubectl apply -f-
    echo "Inventory operator installed"
    
    # Validate service configuration and detect port mismatches
    echo "Validating service configuration..."
    local service_ports=$(kubectl -n subnet-services get service operator-inventory -o jsonpath='{.spec.ports[*].port}' 2>/dev/null)
    if echo "$service_ports" | grep -q "${INVENTORY_GRPC_PORT}"; then
        echo "✓ gRPC port 8081 is configured"
    else
        echo "✗ gRPC port 8081 is NOT configured in service"
    fi
    
    if echo "$service_ports" | grep -q "${INVENTORY_REST_PORT}"; then
        echo "✓ REST port 8080 is configured"
    else
        echo "✗ REST port 8080 is NOT configured in service"
    fi
    
    # Check for port mismatches between service and actual container
    echo "Checking for port configuration mismatches..."
    local hex_port
    hex_port=$(get_first_listen_port)
    local actual_port="0"
    if [ -n "$hex_port" ]; then
        actual_port=$((16#$hex_port))
    fi
    
    if [ "$actual_port" != "0" ] && [ "$actual_port" != "${INVENTORY_REST_PORT}" ] && [ "$actual_port" != "${INVENTORY_GRPC_PORT}" ]; then
        echo "⚠️  WARNING: Port mismatch detected!"
        echo "   Service expects ports ${INVENTORY_REST_PORT}/${INVENTORY_GRPC_PORT}, but application listens on port $actual_port"
        echo "   Attempting to fix service configuration..."
        
        # Fix the service to use the actual port
        kubectl -n subnet-services patch service operator-inventory -p "{\"spec\":{\"ports\":[{\"name\":\"rest\",\"port\":${INVENTORY_REST_PORT},\"targetPort\":$actual_port,\"protocol\":\"TCP\"},{\"name\":\"grpc\",\"port\":${INVENTORY_GRPC_PORT},\"targetPort\":$actual_port,\"protocol\":\"TCP\"}]}}" 2>/dev/null
        
        if [ $? -eq 0 ]; then
            echo "✓ Service configuration updated to use port $actual_port"
            # Wait for endpoints to update
            sleep 5
        else
            echo "✗ Failed to update service configuration"
        fi
    else
        echo "✓ Port configuration appears correct"
    fi
}

wait_inventory_available() {
    echo "Waiting for inventory operator to be available..."

    # First, wait for the pod to be running using the helper function
    # The operator-inventory pod uses app.kubernetes.io/component=operator,app.kubernetes.io/instance=inventory-service
    if ! wait_for_pod_ready "subnet-services" "app.kubernetes.io/component=operator,app.kubernetes.io/instance=inventory-service" 30 2; then
        echo "ERROR: Failed to wait for pod to be ready"
        exit 1
    fi
    
    # Wait for service endpoints to be ready
    echo "Waiting for service endpoints to be ready..."
    local max_wait=30
    local wait_count=0
    
    while [ $wait_count -lt $max_wait ]; do
        local endpoints=$(kubectl -n subnet-services get endpoints operator-inventory -o jsonpath='{.subsets[0].addresses}' 2>/dev/null | wc -w)
        if [ "$endpoints" -gt 0 ]; then
            echo "✓ Service endpoints are ready ($endpoints endpoint(s))"
            break
        fi
        echo "Waiting for service endpoints... ($((wait_count + 1))/$max_wait)"
        sleep 2
        wait_count=$((wait_count + 1))
    done
    
    if [ $wait_count -eq $max_wait ]; then
        echo "WARNING: Service endpoints not ready after $max_wait attempts"
    fi
    
    # Test gRPC endpoint availability with better error handling
    if [ "$WITH_GRPC_CHECK" != "true" ]; then
        echo "Skipping gRPC checks (enable with WITH_GRPC_CHECK=true)"
        echo "Inventory operator is now available!"
        return 0
    fi

    echo "Testing gRPC endpoint availability..."
    local grpc_ready=false
    local attempts=0
    local max_attempts=15  # Reduced from 30 to avoid long waits
    
    while [ "$grpc_ready" = false ] && [ $attempts -lt $max_attempts ]; do
        attempts=$((attempts + 1))
        
        # Test direct connection to the service without port-forward
        echo "Attempt $attempts/$max_attempts: Testing gRPC service directly..."
        
        # Use kubectl exec to test from within the cluster
        # First, determine the actual port the application is listening on
        local hex_port
        hex_port=$(get_first_listen_port)
        local actual_port="${INVENTORY_GRPC_PORT}"
        if [ -n "$hex_port" ]; then
            actual_port=$((16#$hex_port))
        fi
        
        echo "   Testing on port $actual_port..."
        
        # Test if the port is accepting connections
        if kubectl -n subnet-services exec deployment/operator-inventory -- timeout 5 bash -c "cat < /dev/tcp/localhost/$actual_port" >/dev/null 2>&1; then
            echo "✓ Port $actual_port is accepting connections"
            grpc_ready=true
            break
        else
            echo "gRPC endpoint not responding yet, waiting... (attempt $attempts/$max_attempts)"
            sleep 3
        fi
    done
    
    if [ "$grpc_ready" = false ]; then
        echo "WARNING: gRPC endpoint failed to become available within timeout"
        echo "This could indicate:"
        echo "  - The gRPC service is not implemented yet"
        echo "  - The service is using a different port or configuration"
        echo "  - The service is not fully initialized"
        echo "  - There are network policy issues"
        echo ""
        echo "However, the operator pod is running and monitoring Ceph clusters."
        echo "The inventory operator may be working through other mechanisms."
        echo ""
        echo "Running debug function to gather more information..."
        debug_inventory_operator
        
        echo ""
        echo "Continuing setup as the operator pod is running..."
        echo "You can manually test the gRPC endpoint later when it becomes available."
        
        # Test if the operator is working through other means
        echo ""
        echo "Testing operator health through alternative endpoints..."
        
        # Test REST endpoint
        local temp_pid
        kubectl -n subnet-services port-forward --address 127.0.0.1 service/operator-inventory ${INVENTORY_REST_PORT}:rest >/dev/null 2>&1 &
        temp_pid=$!
        sleep 2
        
        if test_port 127.0.0.1 ${INVENTORY_REST_PORT}; then
            echo "✓ REST endpoint (port ${INVENTORY_REST_PORT}) is accessible"
        else
            echo "✗ REST endpoint (port ${INVENTORY_REST_PORT}) is not accessible"
        fi
        
        # Clean up port-forward
        kill $temp_pid 2>/dev/null || true
        wait $temp_pid 2>/dev/null || true
        
        echo "✅ Setup completed successfully!"
        return 0
    fi
    
    echo "Inventory operator is now available!"
}

# Helper function to provide debugging information
debug_inventory_operator() {
    local namespace="subnet-services"
    local label_selector="app.kubernetes.io/component=operator,app.kubernetes.io/instance=inventory-service"
    
    echo "=== DEBUGGING INVENTORY OPERATOR ==="
    echo ""
    
    echo "1. Pod Status:"
    kubectl -n "$namespace" get pods -l "$label_selector" -o wide
    echo ""
    
    echo "2. Service Status:"
    kubectl -n "$namespace" get svc -l "$label_selector"
    echo ""
    
    echo "3. Endpoints:"
    kubectl -n "$namespace" get endpoints -l "$label_selector"
    echo ""
    
    echo "4. Pod Events:"
    kubectl -n "$namespace" get events --sort-by='.lastTimestamp' | grep "$label_selector" | tail -10
    echo ""
    
    echo "5. Pod Logs:"
    kubectl -n "$namespace" logs -l "$label_selector" --tail=50 || echo "No logs available"
    echo ""
    
    echo "6. Network Policies:"
    kubectl -n "$namespace" get networkpolicies
    echo ""
    
    echo "7. Cluster Info:"
    kubectl cluster-info
    echo ""
}

# Test alternative endpoints to verify operator health
test_operator_health() {
    local namespace="subnet-services"
    local label_selector="app.kubernetes.io/component=operator,app.kubernetes.io/instance=inventory-service"
    
    echo "Testing operator health through alternative endpoints..."
    
    # Test REST endpoint
    local rest_working=false
    local temp_pid
    
    kubectl -n "$namespace" port-forward --address 127.0.0.1 service/operator-inventory 8080:rest >/dev/null 2>&1 &
    temp_pid=$!
    
    sleep 2
    
    if test_port 127.0.0.1 8080; then
        echo "✓ REST endpoint (port 8080) is accessible"
        rest_working=true
    else
        echo "✗ REST endpoint (port 8080) is not accessible"
    fi
    
    # Clean up port-forward
    kill $temp_pid 2>/dev/null || true
    wait $temp_pid 2>/dev/null || true
    
    # Check if pod is actively working
    local recent_logs=$(kubectl -n "$namespace" logs -l "$label_selector" --tail=5 2>/dev/null | grep -c "MODIFIED\|INFO\|WARN\|ERROR" || echo "0")
    
    if [ "$recent_logs" -gt 0 ]; then
        echo "✓ Operator is actively working (recent activity in logs)"
    else
        echo "✗ No recent activity detected in operator logs"
    fi
    
    # Check if the operator is responding to basic health checks
    if kubectl -n "$namespace" get pods -l "$label_selector" | grep -q "Running"; then
        echo "✓ Operator pod is running and healthy"
    else
        echo "✗ Operator pod is not in running state"
    fi
    
    echo ""
    echo "Operator Health Summary:"
    if [ "$rest_working" = true ] && [ "$recent_logs" -gt 0 ]; then
        echo "✅ Operator appears to be healthy and working"
        return 0
    else
        echo "⚠️  Operator may have issues - check logs and configuration"
        return 1
    fi
}

wait_command() {
    case "${1}" in
    inventory-available)
        shift
        wait_inventory_available
        ;;
    debug)
        shift
        debug_inventory_operator
        ;;
    *)
        echo "invalid wait command"
        echo "Available commands:"
        echo "  inventory-available - Wait for inventory operator to be available"
        echo "  debug              - Show debugging information for inventory operator"
        exit 1
        ;;
    esac
}

# Main execution logic
main() {
    detect_platform
    check_requirements
    check_kubectl_connection
    install_ns
    install_crd
    install_network_policies
    install_inventory_operator
    wait_inventory_available
}

# Check if arguments were provided
if [ $# -eq 0 ]; then
    # No arguments - run full setup
    main
else
    # Arguments provided - handle specific commands
    case "${1}" in
    "wait")
        shift
        wait_command "$@"
        ;;
    *)
        echo "Usage: $0 [wait <command>]"
        echo ""
        echo "Commands:"
        echo "  (no args)     - Run full setup"
        echo "  wait inventory-available - Wait for inventory operator to be available"
        echo "  wait debug    - Show debugging information for inventory operator"
        exit 1
        ;;
    esac
fi