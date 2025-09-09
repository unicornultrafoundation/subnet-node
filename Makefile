###############################################################################
###                                Protobuf                                 ###
###############################################################################

check-proto-deps:
ifeq (,$(shell which protoc-gen-gogofaster))
	@go install github.com/gogo/protobuf/protoc-gen-gogofaster@latest
endif
.PHONY: check-proto-deps

check-proto-format-deps:
ifeq (,$(shell which clang-format))
	$(error "clang-format is required for Protobuf formatting. See instructions for your platform on how to install it.")
endif
.PHONY: check-proto-format-deps

proto-gen: check-proto-deps
	@echo "Generating Protobuf files"
	@go run github.com/bufbuild/buf/cmd/buf generate --exclude-path proto/subnet/k8s
.PHONY: proto-gen

proto-gen-k8s: check-proto-deps
	@echo "Generating Protobuf files (k8s only)"
	cd proto/subnet/k8s && buf generate --exclude-path k8s.io
	@echo "Cleaning up unused imports in generated Go files"
	@goimports -w proto/subnet/k8s/
.PHONY: proto-gen-k8s

mock-gen:
	@go run github.com/vektra/mockery/v2@latest \
		--dir=core/k8s/types/v1/clients/ip \
		--output=core/k8s/types/v1/clients/ip/mocks \
		--name=Client
.PHONY: mock-gen

# These targets are provided for convenience and are intended for local
# execution only.
proto-lint: check-proto-deps
	@echo "Linting Protobuf files"
	@go run github.com/bufbuild/buf/cmd/buf lint
.PHONY: proto-lint

proto-format: check-proto-format-deps
	@echo "Formatting Protobuf files"
	@find . -name '*.proto' -path "./proto/*" -exec clang-format -i {} \;
.PHONY: proto-format

proto-check-breaking: check-proto-deps
	@echo "Checking for breaking changes in Protobuf files against local branch"
	@echo "Note: This is only useful if your changes have not yet been committed."
	@echo "      Otherwise read up on buf's \"breaking\" command usage:"
	@echo "      https://docs.buf.build/breaking/usage"
	@go run github.com/bufbuild/buf/cmd/buf breaking --against ".git"
.PHONY: proto-check-breaking

proto-check-breaking-ci:
	@go run github.com/bufbuild/buf/cmd/buf breaking --against $(HTTPS_GIT)#branch=v0.34.x
.PHONY: proto-check-breaking-ci

kustomize-deploy-subnet-operator-inventory:
	@echo "Deploying subnet-operator-inventory"
	@kubectl kustomize pkg/k8s/kustomize/subnet-operator-inventory | kubectl apply -f-
.PHONY: kustomize-deploy-subnet-operator-inventory

###############################################################################
###                         k8s-services & Setup                            ###
###############################################################################

# Run kube setup (namespace, CRDs, operator, policies)
kube-setup:
	@echo "Running kube setup"
	@script/kube/setup-kube.sh
.PHONY: kube-setup

# Build local k8s-services image using Dockerfile at script/k8s-services/
k8s-services-build:
	@echo "Building local k8s-services image"
	@script/k8s-services/build-k8s-services.sh
.PHONY: k8s-services-build

# Build image and then setup kube 
# (only for local development - need to update the image from u2udepin/k8s-services:latest to k8s-services:latest)
k8s-services-build-and-kube-setup: k8s-services-build
	@echo "Building k8s-services image..."
	@$(MAKE) k8s-services-build
	@echo "Ensuring local registry is running..."
	@./script/manage-local-registry.sh start
	@echo "Pushing k8s-services to local registry..."
	@./script/manage-local-registry.sh push k8s-services:latest
	@echo "Running kube setup..."
	@$(MAKE) kube-setup
.PHONY: k8s-services-build-and-kube-setup

###############################################################################
###                                 Rook                                    ###
###############################################################################

# Deploy Rook/Ceph; override profile via ROOK_PROFILE=dev (default: prod)
# NOTE: This is now experimental, use at your own risk
ROOK_PROFILE ?= prod

rook-setup:
	@echo "Deploying Rook/Ceph with profile=$(ROOK_PROFILE)"
	@script/rook/rook.sh --profile $(ROOK_PROFILE) deploy
.PHONY: rook-setup