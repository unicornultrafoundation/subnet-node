#!/usr/bin/env bash

set -e

rootdir="$(dirname "$0")/.."

retries=10
retrywait=1

CRD_FILE=$rootdir/pkg/k8s/apis/crd.yaml
STORAGE_CLASS_FILE=$rootdir/pkg/k8s/apis/storageclass.yaml
NAMESPACE_FILE=$rootdir/pkg/k8s/apis/namespace.yaml

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
    kubectl patch node "${CONTROL_PLANE_NODE}" -p '{"metadata":{"labels":{"subnet.node/storageclasses":"default.v1","ingress-ready":"true"}}}'
}

install_network_policies() {
    set -x
    kubectl kustomize "$rootdir/pkg/k8s/kustomize/subnet-services/" | kubectl apply -f-
}

wait_inventory_available() {
    set -x

    local pid

    kubectl -n subnet-services port-forward --address 0.0.0.0 service/operator-inventory 8455:grpc &
    pid=$!

    # shellcheck disable=SC2064
    trap "kill -SIGINT ${pid}" EXIT

    timeout 10 bash -c -- 'while ! nc -vz localhost 8455 > /dev/null 2>&1 ; do sleep 0.1; done'

    local r=0

    while ! grpcurl -plaintext localhost:8455 subnet.inventory.v1.ClusterRPC.QueryCluster | jq '(.nodes | length > 0) and (.storage | length > 0)' --exit-status > /dev/null 2>&1; do
        r=$((r+1))
        if [ ${r} -eq "${retries}" ]; then
            exit 0
        fi

        # shellcheck disable=SC2086
        sleep $retrywait
    done
}

wait() {
    case "${1}" in
    inventory-available)
        shift
        wait_inventory_available
        ;;
    *)
        echo "invalid wait command"
        exit 1
        ;;
    esac
}

install_ns
install_crd
install_network_policies

case "${1}" in
"wait")
    shift
    wait "$@"
    ;;
esac