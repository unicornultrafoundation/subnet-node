#!/bin/bash
set -e

# --- Default values ---
NAMESPACE="subnet-provider"
PROVIDER_NAME="subnet-provider"
IMAGE_NAME="subnet-node"
IMAGE_TAG="latest"
REPLICAS=1
CPU_REQUEST="100m"
CPU_LIMIT="500m"
MEMORY_REQUEST="128Mi"
MEMORY_LIMIT="512Mi"
PORT=8080
HEALTH_CHECK_PORT=8081
OWNER_ADDR=""
ACCOUNT_SECRET=""
DRY_RUN=false
DELETE_MODE=false

# --- Colors ---
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info()    { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $1"; }
error()   { echo -e "${RED}[ERROR]${NC} $1"; }

usage() {
  echo "Usage: $0 [options]"
  echo "  --owner-addr <addr>       Owner Ethereum address (required)"
  echo "  --account-secret <file>   Path to keystore secret JSON (required)"
}

# --- Parse only required args ---
while [[ $# -gt 0 ]]; do
  case $1 in
    --owner-addr) OWNER_ADDR="$2"; shift 2;;
    --account-secret) ACCOUNT_SECRET="$2"; shift 2;;
    --dry-run) DRY_RUN=true; shift;;
    --delete) DELETE_MODE=true; shift;;
    --help) usage; exit 0;;
    *) error "Unknown option: $1"; usage; exit 1;;
  esac
done

# --- Prerequisites ---
if ! command -v kubectl >/dev/null 2>&1; then error "kubectl not found"; exit 1; fi
if ! kubectl cluster-info >/dev/null 2>&1; then error "Cannot connect to cluster"; exit 1; fi
if [[ -z "$OWNER_ADDR" ]]; then error "--owner-addr is required"; exit 1; fi
if [[ -z "$ACCOUNT_SECRET" ]]; then error "--account-secret is required"; exit 1; fi
if [[ ! -f "$ACCOUNT_SECRET" ]]; then error "account-secret file not found: $ACCOUNT_SECRET"; exit 1; fi

# --- YAML Generation ---
generate_yaml() {
cat <<EOF
---
apiVersion: v1
kind: Secret
metadata:
  name: ${PROVIDER_NAME}-account-secret
  namespace: $NAMESPACE
type: Opaque
data:
  keystore.json: $(base64 -w 0 < "$ACCOUNT_SECRET")
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ${PROVIDER_NAME}-sa
  namespace: $NAMESPACE
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: ${PROVIDER_NAME}-clusterrole
rules:
  - apiGroups: ["*"]
    resources: ["pods", "pods/log", "services", "endpoints", "configmaps", "secrets", "deployments", "nodes", "namespaces"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ${PROVIDER_NAME}-clusterrolebinding
subjects:
  - kind: ServiceAccount
    name: ${PROVIDER_NAME}-sa
    namespace: $NAMESPACE
roleRef:
  kind: ClusterRole
  name: ${PROVIDER_NAME}-clusterrole
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: v1
kind: Service
metadata:
  name: $PROVIDER_NAME
  namespace: $NAMESPACE
  labels:
    app: $PROVIDER_NAME
spec:
  selector:
    app: $PROVIDER_NAME
  ports:
    - name: http
      port: $PORT
      targetPort: $PORT
      protocol: TCP
    - name: health
      port: $HEALTH_CHECK_PORT
      targetPort: $HEALTH_CHECK_PORT
      protocol: TCP
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $PROVIDER_NAME
  namespace: $NAMESPACE
  labels:
    app: $PROVIDER_NAME
spec:
  replicas: $REPLICAS
  selector:
    matchLabels:
      app: $PROVIDER_NAME
  template:
    metadata:
      labels:
        app: $PROVIDER_NAME
    spec:
      serviceAccountName: ${PROVIDER_NAME}-sa
      containers:
      - name: $PROVIDER_NAME
        image: $IMAGE_NAME:$IMAGE_TAG
        ports:
        - containerPort: $PORT
          name: http
        - containerPort: $HEALTH_CHECK_PORT
          name: health
        env:
        - name: PROVIDER_NAME
          value: "$PROVIDER_NAME"
        - name: PROVIDER_PORT
          value: "$PORT"
        - name: HEALTH_PORT
          value: "$HEALTH_CHECK_PORT"
        - name: OWNER_ADDR
          value: "$OWNER_ADDR"
        volumeMounts:
        - name: account-secret-volume
          mountPath: /keystore
          readOnly: true
        resources:
          requests:
            cpu: $CPU_REQUEST
            memory: $MEMORY_REQUEST
          limits:
            cpu: $CPU_LIMIT
            memory: $MEMORY_LIMIT
        livenessProbe:
          httpGet:
            path: /health
            port: health
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: health
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
      volumes:
      - name: account-secret-volume
        secret:
          secretName: ${PROVIDER_NAME}-account-secret
      restartPolicy: Always
EOF
EOF
}

# --- Delete resources ---
delete_resources() {
  info "Deleting resources in namespace: $NAMESPACE"
  kubectl delete deployment,service,secret "$PROVIDER_NAME" -n "$NAMESPACE" --ignore-not-found
  success "Deleted provider resources."
}

# --- Main logic ---
main() {
  info "Provider Setup for Kubernetes"
  if $DELETE_MODE; then
    delete_resources
    exit 0
  fi
  if $DRY_RUN; then
    generate_yaml
    exit 0
  fi
  generate_yaml | kubectl apply -f -
  success "Provider deployed!"
  info "Check status: kubectl get pods -n $NAMESPACE -l app=$PROVIDER_NAME"
  info "View logs: kubectl logs -f deployment/$PROVIDER_NAME -n $NAMESPACE"
  info "Port-forward: kubectl port-forward service/$PROVIDER_NAME $PORT:$PORT -n $NAMESPACE"
}

main "$@"
