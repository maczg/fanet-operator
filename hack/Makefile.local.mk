##@ Kind Cluster

KIND_CLUSTER_NAME ?= demo-fanet-cluster
KIND_VERSION ?= v0.30.0
KIND_NODE_IMAGE_VERSION ?= v1.34.0
KIND_IN_PATH := $(shell command -v kind 2>/dev/null)
KIND := $(if $(KIND_IN_PATH),$(KIND_IN_PATH),$(shell go env GOPATH)/bin/kind)

.PHONY: kind-ensure
kind-ensure: ## Ensure kind is installed locally if not found in PATH
	@if [ -z "$(KIND)" ] || ! [ -x "$(KIND)" ]; then \
		echo "INFO: kind not found in PATH or locally. Installing..."; \
		go install sigs.k8s.io/kind@$(KIND_VERSION); \
	else \
		echo "INFO: kind is already installed at $(KIND)"; \
	fi

.PHONY: kind-config
kind-config: ## Ensure kind-config.yaml is present
	@if ! [ -f "./kind/config.yaml" ]; then \
		echo "ERROR: kind config not found in ./kind dir, please create a valid kind config"; \
		exit 1; \
	fi

.PHONY: kind-cluster
kind-cluster: kind-ensure kind-config ## Create kind cluster if not already present
	@if ! $(KIND) get clusters 2>/dev/null | grep -q "^$(KIND_CLUSTER_NAME)$$"; then \
		echo "INFO: Creating kind cluster '$(KIND_CLUSTER_NAME)'..."; \
		$(KIND) create cluster --name $(KIND_CLUSTER_NAME) --image kindest/node:$(KIND_NODE_IMAGE_VERSION) --config kind/config.yaml; \
	else \
		echo "INFO: Kind cluster '$(KIND_CLUSTER_NAME)' already exists."; \
	fi

.PHONY: kind-delete-cluster
kind-delete-cluster: ## Delete kind cluster if present
	@if command -v $(KIND) >/dev/null 2>&1 && $(KIND) get clusters 2>/dev/null | grep -q "^$(KIND_CLUSTER_NAME)$$"; then \
		echo "INFO: Deleting kind cluster '$(KIND_CLUSTER_NAME)'..."; \
		$(KIND) delete cluster --name $(KIND_CLUSTER_NAME); \
	else \
		echo "INFO: Kind cluster '$(KIND_CLUSTER_NAME)' not found."; \
	fi

.PHONY: kind-deploy
kind-deploy: kind-cluster docker-build kind-load deploy ## Build, load and deploy the operator into the kind cluster
	@echo "INFO: Operator deployed successfully to kind cluster."

.PHONY: kind-load
kind-load: ##loads the operator image into the kind cluster.
	@echo "INFO: Loading operator image '$(IMG)' into kind cluster '$(KIND_CLUSTER_NAME)'..."
	@$(KIND) load docker-image $(IMG) --name $(KIND_CLUSTER_NAME)
