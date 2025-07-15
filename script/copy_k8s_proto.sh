#!/bin/bash
# Copies k8s.io/apimachinery proto files from vendor to proto/k8s.io
# Usage: ./copy_k8s_proto.sh

set -euo pipefail

SRC_DIR="$(dirname "$0")/../vendor/k8s.io"
DST_DIR="$(dirname "$0")/../proto/subnet/k8s/k8s.io"

# Only include k8s.io/apimachinery
INCLUDE="k8s.io/apimachinery"

mkdir -p "$DST_DIR"

# Download rsync if not installed
if ! command -v rsync &> /dev/null; then
    echo "rsync could not be found, installing..."
    sudo apt-get update
    sudo apt-get install -y rsync
fi

rsync -av --include="$INCLUDE/" --include="$INCLUDE/**" --include="*/" --include="*.proto" --exclude="*" "$SRC_DIR/" "$DST_DIR/"

echo "Copied k8s.io/apimachinery proto files to proto/k8s.io/" 