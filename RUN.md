# Development Environment Setup

This guide covers how to run CLIProxyAPI for quick feature testing using Docker.

## Prerequisites

- Docker and Docker Compose installed
- Configuration file at `config.yaml` (copy from `.local-config/config.yaml` if needed)
- Auth files in `.local-config/` directory

## Quick Start

```bash
# Move your config to root (if not already there)
cp .local-config/config.yaml config.yaml

# Start development environment
make dev

# View logs
make dev-logs
```

## Available Commands

```bash
# Start development containers
make dev

# Stop development containers
make dev-stop

# View live logs
make dev-logs

# Rebuild after code changes
make dev-rebuild

# Build locally (outside Docker)
make build

# Run tests
make test
make test-verbose

# Clean build artifacts
make clean
```

## Manual Docker Commands

If you prefer not to use the Makefile:

```bash
# Start dev environment
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose -f docker-compose.dev.yml logs -f

# Rebuild after changes
docker-compose -f docker-compose.dev.yml down
docker-compose -f docker-compose.dev.yml up -d --build

# Stop
docker-compose -f docker-compose.dev.yml down
```

## Development Workflow

1. **Make code changes** in your local project
2. **Rebuild**: `make dev-rebuild` (this rebuilds the container with your changes)
3. **Test**: Access API at `http://localhost:8317`
4. **View logs**: `make dev-logs` to see what's happening

## Development Configuration

The `docker-compose.dev.yml` setup:

- **Source code**: Mounted at `/app` inside container
- **Config file**: Uses `./config.yaml` from project root
- **Auth files**: Mounts `.local-config/` to `/root/.cli-proxy-api`
- **Logs**: Mounts `./logs/` to `/CLIProxyAPI/logs`
- **Environment**: Sets `DEPLOY=local`
- **Ports exposed**: 8317, 8085, 1455, 54545, 51121, 11451

## Configuration Files

- `docker-compose.dev.yml` - Development Docker Compose configuration
- `Makefile` - Development command shortcuts
- `config.yaml` - Main configuration (copy from `.local-config/config.yaml`)

## Testing Your Features

After making code changes:

```bash
# Rebuild and restart
make dev-rebuild

# Check logs to verify startup
make dev-logs

# Test your API endpoints
curl http://localhost:8317/v1/models
```

## Troubleshooting

**Container not starting?**
```bash
# Check logs
make dev-logs

# Rebuild from scratch
make dev-stop
docker rmi cli-proxy-api-dev
make dev
```

**Config changes not applied?**
- Ensure `config.yaml` exists in project root
- Restart containers: `make dev-rebuild`

**Need to run tests locally?**
```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Run a specific test
go test -run TestSpecificName ./test/
```
