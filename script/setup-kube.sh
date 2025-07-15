#!/usr/bin/env bash

set -e

rootdir="$(dirname "$0")/.."

CRD_FILE=$rootdir/pkg/k8s/apis/crd.yaml
STORAGE_CLASS_FILE=$rootdir/pkg/k8s/apis/storageclass.yaml

install_crd() {
    set -x
    kubectl apply -f "$CRD_FILE"
    kubectl apply -f "$STORAGE_CLASS_FILE"
}

install_crd