# Makefile for CLIProxyAPI development tasks

# Variables
BINARY_NAME = cli-proxy-api
DOCKER_COMPOSE_DEV = docker-compose -f docker-compose.dev.yml

# Phony targets
.PHONY: help dev dev-logs dev-stop dev-rebuild build test test-verbose clean

# Default target
help:
	@echo "Available targets:"
	@echo "  make dev          - Start development environment using docker-compose.dev.yml"
	@echo "  make dev-logs     - Show logs from dev containers"
	@echo "  make dev-stop     - Stop dev containers"
	@echo "  make dev-rebuild  - Rebuild and restart dev containers"
	@echo "  make build        - Build the server binary locally"
	@echo "  make test         - Run all tests"
	@echo "  make test-verbose - Run tests with verbose output"
	@echo "  make clean        - Clean up build artifacts and remove built binary"

# Development targets
dev:
	$(DOCKER_COMPOSE_DEV) up -d

dev-logs:
	$(DOCKER_COMPOSE_DEV) logs -f

dev-stop:
	$(DOCKER_COMPOSE_DEV) down

dev-rebuild:
	$(DOCKER_COMPOSE_DEV) down
	$(DOCKER_COMPOSE_DEV) up -d --build

# Build targets
build:
	go build -o $(BINARY_NAME) ./cmd/server/

# Test targets
test:
	go test ./...

test-verbose:
	go test -v ./...

# Cleanup targets
clean:
	rm -f $(BINARY_NAME)
