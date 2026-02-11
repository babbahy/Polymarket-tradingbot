#!/bin/bash

# Development helper script for Polymarket Arbitrage Bot

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Polymarket Arbitrage Bot - Development Helper${NC}"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    echo "Please install Go 1.22 or later from https://go.dev/dl/"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}')
echo -e "${GREEN}✓${NC} Go version: $GO_VERSION"

# Function to run with error handling
run_cmd() {
    echo -e "${YELLOW}Running:${NC} $1"
    if eval "$1"; then
        echo -e "${GREEN}✓${NC} Success"
    else
        echo -e "${RED}✗${NC} Failed"
        exit 1
    fi
    echo ""
}

# Parse command line arguments
case "${1:-help}" in
    setup)
        echo "Setting up development environment..."
        run_cmd "go mod download"
        run_cmd "go mod tidy"
        echo -e "${GREEN}Setup complete!${NC}"
        ;;
    
    build)
        echo "Building binaries..."
        run_cmd "go build -v -o bin/polymarket-arbitrage ./cmd/trader"
        run_cmd "go build -v -o bin/backtest ./cmd/backtest"
        echo -e "${GREEN}Build complete! Binaries in ./bin/${NC}"
        ;;
    
    test)
        echo "Running tests..."
        run_cmd "go test -v -race ./..."
        ;;
    
    lint)
        echo "Running linter..."
        if ! command -v golangci-lint &> /dev/null; then
            echo -e "${YELLOW}Warning: golangci-lint not installed${NC}"
            echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
            exit 1
        fi
        run_cmd "golangci-lint run --timeout=5m ./..."
        ;;
    
    run-paper)
        echo "Starting in paper trading mode..."
        if [ ! -f "bin/polymarket-arbitrage" ]; then
            echo -e "${YELLOW}Binary not found, building first...${NC}"
            ./scripts/dev.sh build
        fi
        ./bin/polymarket-arbitrage --mode=paper --config=configs/config.yaml
        ;;
    
    run-dry)
        echo "Starting in dry-run mode..."
        if [ ! -f "bin/polymarket-arbitrage" ]; then
            echo -e "${YELLOW}Binary not found, building first...${NC}"
            ./scripts/dev.sh build
        fi
        ./bin/polymarket-arbitrage --mode=dry-run --config=configs/config.yaml
        ;;
    
    docker-build)
        echo "Building Docker image..."
        run_cmd "docker build -t polymarket-arbitrage:latest ."
        ;;
    
    clean)
        echo "Cleaning build artifacts..."
        rm -rf bin/ coverage.out coverage.html
        echo -e "${GREEN}Clean complete!${NC}"
        ;;
    
    help|*)
        echo "Usage: ./scripts/dev.sh [command]"
        echo ""
        echo "Commands:"
        echo "  setup         - Download dependencies and setup environment"
        echo "  build         - Build all binaries"
        echo "  test          - Run tests"
        echo "  lint          - Run linter"
        echo "  run-paper     - Run in paper trading mode"
        echo "  run-dry       - Run in dry-run mode"
        echo "  docker-build  - Build Docker image"
        echo "  clean         - Clean build artifacts"
        echo "  help          - Show this help"
        ;;
esac
