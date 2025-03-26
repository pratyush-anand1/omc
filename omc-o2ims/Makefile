# Project variables
PROJECT_NAME := omc-o2ims
VERSION ?= $(shell git describe --tags --always --dirty)
REGISTRY ?= omc-o2ims-registry
IMAGE_NAME := $(REGISTRY)/$(PROJECT_NAME)

# Go related variables
GO := go
GOPATH := $(shell go env GOPATH)
GOBIN := $(GOPATH)/bin
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)


CONTROLLER_GEN := $(GOBIN)/controller-gen
CLIENT_GEN := $(GOBIN)/client-gen
# Build variables
BUILD_DIR := build
BINARY_NAME := $(PROJECT_NAME)
MAIN_PACKAGE := ./cmd/server

# Kubernetes related variables
KUBE_NAMESPACE ?= default
HELM_NAMESPACE ?= default
HELM_RELEASE_NAME ?= omc-o2ims

# Test variables
COVERAGE_DIR := coverage
COVERAGE_PROFILE := $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML := $(COVERAGE_DIR)/coverage.html

# Tool installation
.PHONY: tools
tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
	go install k8s.io/code-generator/cmd/client-gen@latest

# Build targets
.PHONY: build
build: clean generate fmt vet
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build \
		-ldflags "-X main.version=$(VERSION)" \
		-o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	cp -r config $(BUILD_DIR)
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(COVERAGE_DIR)

# Development targets
.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: vet
vet:
	$(GO) vet ./...

LINT_FLAGS := --exclude-files=".*_test.go" --exclude-files=".*_mock.go"
.PHONY: lint
lint:
	golangci-lint run $(LINT_FLAGS)
	go fmt  -n ./...
	go vet   ./... 

.PHONY: generate
generate:
	echo aa
	#$(CONTROLLER_GEN) object paths="./pkg/apis/..."
	# $(CLIENT_GEN) \
	# 	--input-base "" \
	# 	--input "pkg/apis/o2ims/v1alpha1" \
	# 	--output-dir "pkg/client" \
	# 	--output-pkg "client" \
	# 	--clientset-name "versioned"

# Test targets
.PHONY: test
test: test-unit test-integration test-e2e

.PHONY: test-unit
test-unit:
	mkdir -p $(COVERAGE_DIR)
	$(GO) test -race -coverprofile=$(COVERAGE_PROFILE) ./internal/... 
	$(GO) tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)

.PHONY: test-integration
test-integration:
	$(GO) test -v ./test/integration/...

.PHONY: test-e2e
test-e2e:
	$(GO) test -v ./test/e2e/...

# Docker targets
.PHONY: docker-build
docker-build:
	docker volume create package-cache
	DOCKER_BUILDKIT=1 docker build --build-arg BUILDKIT_INLINE_CACHE=1 -t $(IMAGE_NAME):$(VERSION) .
	docker tag $(IMAGE_NAME):$(VERSION) $(IMAGE_NAME):latest


.PHONY: docker-run
docker-run:
	docker rm -f $(PROJECT_NAME) || true
	docker run --name $(PROJECT_NAME) -p 9090:9090  --network=host  -it $(IMAGE_NAME):$(VERSION)

CRD_NAME ?= "provisioningrequests.o2ims.provisioning.oran.org"
.PHONY: docker-stop
docker-stop:
	docker stop  $(PROJECT_NAME) || true
	kubectl get $(CRD_NAME) -A -o name | \
  xargs -I {} kubectl patch {} -p '{"metadata":{"finalizers":[]}}' --type=merge
	kubectl delete crd $(CRD_NAME)

.PHONY: docker-push
docker-push:
	docker push $(IMAGE_NAME):$(VERSION)
	docker push $(IMAGE_NAME):latest

# Kubernetes deployment targets
.PHONY: deploy-k8s
deploy-k8s:
	kubectl apply -f deployments/kubernetes/configmap.yaml -n $(KUBE_NAMESPACE)
	kubectl apply -f deployments/kubernetes/deployment.yaml -n $(KUBE_NAMESPACE)
	kubectl apply -f deployments/kubernetes/service.yaml -n $(KUBE_NAMESPACE)


EXCLUDE_DIRS="(^|/)(build|coverage|test|testdata)(/|$$)"
EXCLUDE_FILES=*.test.go|*_mock.go|*.generated.go|*test.go  
.PHONY: gocloc
gocloc:
	@echo "Running goloc..."
	@gocloc .   --not-match-d=$(EXCLUDE_DIRS)


.PHONY: deploy-helm
deploy-helm:
	helm upgrade --install $(HELM_RELEASE_NAME) ./deployments/helm \
		--namespace $(HELM_NAMESPACE) \
		--create-namespace \
		--set image.tag=$(VERSION)

.PHONY: undeploy-k8s
undeploy-k8s:
	kubectl delete -f deployments/kubernetes/ -n $(KUBE_NAMESPACE)

.PHONY: undeploy-helm
undeploy-helm:
	helm uninstall $(HELM_RELEASE_NAME) -n $(HELM_NAMESPACE)

# Development workflow targets
.PHONY: dev
dev: generate fmt vet lint test build

.PHONY: ci
ci: tools generate fmt vet lint test docker-build

# Run locally
.PHONY: run
run:
	$(GO) run $(MAIN_PACKAGE)


.PHONY: vuln-check
vuln-check:
	@echo "Running vulnerability test on latest image $(IMAGE_NAME):latest"
	trivy image --exit-code 0 --no-progress --ignore-unfixed $(IMAGE_NAME):latest
	@echo "govulncheck  $(GOPATH)/pkg/mod/ "
	govulncheck  ./...

.PHONY: check-updates
check-updates:
	@go list -m -u -f '{{if .Update}}{{.Path}}|{{.Version}}|{{.Update.Version}}{{end}}' all | awk -F'|' '\
	{ \
		split($$2, curr, "."); \
		split($$3, new, "."); \
		if (curr[1] != new[1]) { \
			printf "\033[33m%s | \033[31m%s | \033[32m%s\033[0m\n", $$1, $$2, $$3; \
		} else { \
			printf "%s | \033[31m%s | \033[32m%s\033[0m\n", $$1, $$2, $$3; \
		} \
	}'
.PHONY: update-deps
update-deps:
	@go get -u ./...          # Update all dependencies
	@go mod tidy              # Remove unused dependencies

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  clean          - Clean build artifacts"
	@echo "  dev            - Run full development workflow"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-push    - Push Docker image"
	@echo "  deploy-k8s     - Deploy to Kubernetes using raw manifests"
	@echo "  deploy-helm    - Deploy to Kubernetes using Helm"
	@echo "  fmt            - Run go fmt"
	@echo "  generate       - Generate code"
	@echo "  help           - Show this help"
	@echo "  lint           - Run linter"
	@echo "  run            - Run locally"
	@echo "  test           - Run all tests"
	@echo "  test-unit      - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-e2e       - Run end-to-end tests"
	@echo "  tools          - Install required tools"
	@echo "  vuln-check 	- Run vulnerability test on latest image"

# Default target
.DEFAULT_GOAL := help
