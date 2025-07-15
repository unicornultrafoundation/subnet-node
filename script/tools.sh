#!/usr/bin/env bash

ROOT_DIR=$(pwd)
AP_DEVCACHE_BIN="${ROOT_DIR}/vendor/k8s.io/code-generator"

function run_k8s_gen() {
	rm -rf "${ROOT_DIR}/pkg/k8s/client/*"

	# Temporarily disable vendor mode for code generation
	export GOFLAGS="-mod=mod"
	
	source "$AP_DEVCACHE_BIN/kube_codegen.sh"

	kube::codegen::gen_helpers \
		--boilerplate "${ROOT_DIR}/pkg/k8s/apis/boilerplate.go.txt" \
		"${ROOT_DIR}/pkg/k8s/apis"

	kube::codegen::gen_client \
		--output-pkg github.com/unicornultrafoundation/subnet-node/pkg/k8s/client \
		--output-dir "${ROOT_DIR}/pkg/k8s/client" \
		--boilerplate "${ROOT_DIR}/pkg/k8s/apis/boilerplate.go.txt" \
		--with-watch \
		--with-applyconfig \
		"${ROOT_DIR}/pkg/k8s/apis"
	
	# Restore vendor mode
	unset GOFLAGS
}

case "$1" in
	k8s-gen)
		run_k8s_gen
		;;
esac