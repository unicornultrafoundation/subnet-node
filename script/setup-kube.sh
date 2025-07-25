#!/usr/bin/env bash

set -e

rootdir="$(dirname "$0")/.."

CRD_FILE=$rootdir/pkg/k8s/apis/crd.yaml
STORAGE_CLASS_FILE=$rootdir/pkg/k8s/apis/storageclass.yaml
NAMESPACE_FILE=$rootdir/pkg/k8s/apis/namespace.yaml

# Ingress Nginx
INGRESS_CONFIG_PATH=$rootdir/pkg/k8s/apis/ingress-nginx.yaml
KUBE_ROLLOUT_TIMEOUT=180
METALLB_CONFIG_PATH=$rootdir/pkg/k8s/apis/metallb.yaml
METALLB_IP_CONFIG_PATH=$rootdir/pkg/k8s/apis/kube-config-metal-lb-ip.yaml
METALLB_SERVICE_PATH=$rootdir/pkg/k8s/apis/metallb-service.yaml

install_ns() {
    set -x
    kubectl apply -f "$NAMESPACE_FILE"
}

install_crd() {
    set -x
    kubectl apply -f "$CRD_FILE"
    kubectl apply -f "$STORAGE_CLASS_FILE"

    # TODO: Config exactly the control plane node name
    CONTROL_PLANE_NODE=$(kubectl get nodes -l node-role.kubernetes.io/control-plane=true -o jsonpath='{.items[0].metadata.name}')
    kubectl patch node "${CONTROL_PLANE_NODE}" -p '{"metadata":{"labels":{"subnet.node/storageclasses":"v1.default","ingress-ready":"true"}}}'
}

setup_ingress() {
    set -x
    kubectl apply -f "$INGRESS_CONFIG_PATH"
    kubectl rollout status deployment -n ingress-nginx ingress-nginx-controller --timeout="$KUBE_ROLLOUT_TIMEOUT"s
    kubectl apply -f "$METALLB_CONFIG_PATH"
	kubectl apply -f "$METALLB_IP_CONFIG_PATH"
	kubectl apply -f "$METALLB_SERVICE_PATH"
}

install_ns
install_crd
setup_ingress