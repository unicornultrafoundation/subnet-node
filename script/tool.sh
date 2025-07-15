#!/usr/bin/env bash

ROOT_DIR=$(git rev-parse --show-toplevel)

function run_k8s_gen() {
	rm -rf "${ROOT_DIR}/pkg/client/*"

	source "$AP_DEVCACHE_BIN/kube_codegen.sh"

	kube::codegen::gen_helpers \
		--boilerplate "${ROOT_DIR}/pkg/apis/boilerplate.go.txt" \
		"${ROOT_DIR}/pkg/apis"

	kube::codegen::gen_client \
		--output-pkg github.com/unicornultrafoundation/subnet-node/pkg/client \
		--output-dir "${ROOT_DIR}/pkg/client" \
		--boilerplate "${ROOT_DIR}/pkg/apis/boilerplate.go.txt" \
		--with-watch \
		--with-applyconfig \
		"${ROOT_DIR}/pkg/apis"
}


run_k8s_gen