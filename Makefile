.PHONY: help build test lint clean run-paper run-dry install-tools fmt vet integration-test docker-build

BINARY_NAME=polymarket-arbitrage
BACKTEST_BINARY=backtest
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-X main.Version=$(shell git describe --tags --always --dirty) -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install-tools: ## Install development tools
	@echo "Installing development tools..."
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install gotest.tools/gotestsum@latest

build: ## Build the trader binary
	@echo "Building trader..."
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/trader
	@echo "Building backtest tool..."
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/$(BACKTEST_BINARY) ./cmd/backtest

test: ## Run unit tests
	@echo "Running unit tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests and show coverage
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

integration-test: ## Run integration tests
	@echo "Running integration tests..."
	$(GO) test -v -tags=integration ./test/integration/...

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run --timeout=5m ./...

fmt: ## Format code
	@echo "Formatting code..."
	$(GO) fmt ./...
	gofmt -s -w .

vet: ## Run go vet
	@echo "Running go vet..."
	$(GO) vet ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

run-paper: build ## Run in paper trading mode
	@echo "Starting paper trading mode..."
	./bin/$(BINARY_NAME) --mode=paper --config=configs/config.yaml

run-dry: build ## Run in dry-run mode (no order placement)
	@echo "Starting dry-run mode..."
	./bin/$(BINARY_NAME) --mode=dry-run --config=configs/config.yaml

docker-build: ## Build Docker image
	docker build -t $(BINARY_NAME):latest .

docker-run-paper: docker-build ## Run paper trading in Docker
	docker run --rm -v $(PWD)/configs:/app/configs $(BINARY_NAME):latest --mode=paper

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

verify: fmt vet lint test ## Run all verification steps

ci: verify integration-test ## Run CI pipeline

.DEFAULT_GOAL := help
