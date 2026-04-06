SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

COMPOSE_FILE := deploy/local/docker-compose.yml
export COMPOSE_FILE
GOCACHE ?= $(CURDIR)/.cache/go-build
export GOCACHE

ifneq (,$(wildcard .env))
include .env
export
endif

SERVICE ?=
GO_PACKAGES := ./...
GO_FILES := $(shell find . -type f -name '*.go' -not -path './.git/*')

.PHONY: help up down logs ps build test lint fmt check smoke seed kafka-topics kafka-test check-structure tree openapi-summary compose-up compose-down topics-create kafka-topology terraform-fmt-check go-build go-test helm-lint helm-template helm-template-dev helm-template-prod microk8s-deploy microk8s-smoke microk8s-status

help: ## Show available targets
	@printf '\nLocal developer workflow:\n\n'
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z0-9_.-]+:.*## / {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@printf '\nExamples:\n'
	@printf '  make up\n'
	@printf '  make smoke\n'
	@printf '  make logs SERVICE=query-api\n\n'

up: ## Start the local stack, wait for services and ensure Kafka topics exist
	@./scripts/up-local.sh

down: ## Stop the local stack and remove orphan containers
	@docker compose -f $(COMPOSE_FILE) down --remove-orphans

logs: ## Tail local stack logs. Optionally set SERVICE=query-api
	@docker compose -f $(COMPOSE_FILE) logs --tail=200 -f $(SERVICE)

ps: ## Show the current local stack status
	@docker compose -f $(COMPOSE_FILE) ps

build: ## Build all Go packages
	@go build $(GO_PACKAGES)

test: ## Run the Go unit test suite
	@go test $(GO_PACKAGES)

go-build: ## Backward-compatible alias for make build
	@$(MAKE) build

go-test: ## Backward-compatible alias for make test
	@$(MAKE) test

lint: ## Run formatting and vet checks for Go sources
	@unformatted="$$(gofmt -l $(GO_FILES))"; \
	if [[ -n "$$unformatted" ]]; then \
		echo 'Go files need formatting:'; \
		echo "$$unformatted"; \
		exit 1; \
	fi
	@go vet $(GO_PACKAGES)

fmt: ## Format Go sources with gofmt
	@gofmt -w $(GO_FILES)

check: ## Run lint, unit tests and structure validation
	@$(MAKE) lint
	@$(MAKE) test
	@$(MAKE) check-structure

smoke: ## Run the local end-to-end smoke test against the running stack
	@./scripts/smoke-test.sh

seed: ## Seed the local stack with realistic demo business events
	@./scripts/seed-local.sh

kafka-topics: ## Create required Kafka topics and print their metadata
	@./scripts/create-topics.sh

kafka-test: ## Verify Kafka topics, offsets and consumer groups in the local broker
	@./scripts/kafka-test.sh

check-structure: ## Validate that the repository scaffold is complete
	@./scripts/check-structure.sh

tree: ## Print the repository tree up to depth 3
	@./scripts/render-tree.sh

openapi-summary: ## Print available OpenAPI contracts
	@printf 'OpenAPI contracts:\n'
	@for file in api/openapi/*.yaml; do \
		echo "--- $$file"; \
		grep -E '^(openapi:|  title:|  version:)' "$$file"; \
		echo; \
	done

compose-up: ## Backward-compatible alias for make up
	@$(MAKE) up

compose-down: ## Backward-compatible alias for make down
	@$(MAKE) down

topics-create: ## Backward-compatible alias for make kafka-topics
	@$(MAKE) kafka-topics

kafka-topology: ## Print the Kafka topology catalog from repository docs
	@echo 'Topics:'
	@sed -n '1,240p' platform/kafka/topic-catalog.yaml
	@echo
	@echo 'Consumer groups:'
	@sed -n '1,240p' platform/kafka/consumer-groups.yaml

terraform-fmt-check: ## Run terraform fmt in check mode when terraform is installed
	@terraform fmt -check -recursive terraform

helm-lint: ## Run helm lint for the application chart when helm is installed
	@command -v helm >/dev/null 2>&1 || { echo 'helm is not installed'; exit 1; }
	@helm lint deploy/helm/platform

helm-template: ## Render the application chart with base values
	@command -v helm >/dev/null 2>&1 || { echo 'helm is not installed'; exit 1; }
	@helm template sap-integration-platform deploy/helm/platform -f deploy/helm/platform/values.yaml

helm-template-dev: ## Render the application chart with the dev overlay
	@command -v helm >/dev/null 2>&1 || { echo 'helm is not installed'; exit 1; }
	@helm template sap-integration-platform deploy/helm/platform -f deploy/helm/platform/values.yaml -f deploy/helm/platform/values-dev.yaml

helm-template-prod: ## Render the application chart with the prod overlay
	@command -v helm >/dev/null 2>&1 || { echo 'helm is not installed'; exit 1; }
	@helm template sap-integration-platform deploy/helm/platform -f deploy/helm/platform/values.yaml -f deploy/helm/platform/values-prod.yaml

microk8s-deploy: ## Build local images, import them into MicroK8s and deploy the platform
	@./scripts/deploy-microk8s.sh

microk8s-smoke: ## Run an end-to-end smoke test against the MicroK8s deployment
	@./scripts/smoke-test-microk8s.sh

microk8s-status: ## Print MicroK8s cluster and workload status
	@./scripts/lib/microk8s.sh >/dev/null 2>&1 || true
	@sg microk8s -c 'microk8s status'
	@sg microk8s -c 'microk8s kubectl get pods -A'
