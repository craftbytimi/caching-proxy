.PHONY: help build test run clean redis redis-stop install lint fmt

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@./scripts/build.sh

test: ## Run all tests
	@./scripts/test.sh

test-unit: ## Run unit tests only
	go test ./internal/... -v -race

test-integration: ## Run integration tests only
	go test ./test/integration/... -v -race

run: ## Run the proxy locally
	go run ./cmd/proxy

redis: ## Start Redis using Docker
	@./scripts/run-redis.sh

redis-stop: ## Stop Redis container
	docker stop caching-proxy-redis || true
	docker rm caching-proxy-redis || true

install: ## Install dependencies
	go mod download
	go mod tidy

lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	go fmt ./...
	gofmt -s -w .

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf tmp/
	rm -f coverage.out coverage.html

dev: ## Run with live reload using Air
	air

.DEFAULT_GOAL := help
