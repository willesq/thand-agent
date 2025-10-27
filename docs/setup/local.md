---
layout: default
title: Local Development
parent: Setup Guides
nav_order: 4
description: "Set up Thand Agent for local development and testing"
---

# Local Development Setup
{: .no_toc }

Set up Thand Agent for local development and testing.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Prerequisites

- Go 1.21 or later
- Git
- Docker (optional, for containerized setup)
- Make (optional, for build automation)

---

## Clone and Build

### Clone Repository

```bash
git clone https://github.com/thand-io/agent.git
cd agent
```

### Build from Source

```bash
# Build the agent
go build -o bin/agent .

# Or use make (if available)
make build
```

### Run Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Local Configuration

### Basic Config

Create `~/.thand/config.yaml`:

```yaml
server:
  url: "http://localhost:8081"  # Local server
  
agent:
  listen_port: 8080
  session_timeout: "24h"       # Longer for development
  
logging:
  level: "debug"               # Verbose logging
  format: "text"               # Human-readable
  
development:
  enabled: true
  mock_providers: true         # Use mock providers
  skip_tls_verify: true        # Allow self-signed certs
```

### Environment Variables

```bash
# Set development environment
export THAND_ENV="development"
export THAND_LOG_LEVEL="debug"
export THAND_DEVELOPMENT_ENABLED="true"
```

---

## Running Locally

### Start the Server

```bash
# Terminal 1: Start server
./bin/agent server start --port 8081 --dev

# Or with live reload (if you have air installed)
air -c .air.toml server
```

### Start the Agent

```bash
# Terminal 2: Start agent
./bin/agent agent start --port 8080 --dev

# Or with configuration file
./bin/agent agent start --config ~/.thand/config.yaml
```

---

## Docker Development

### Using Docker Compose

Create `docker-compose.dev.yml`:

```yaml
version: '3.8'

services:
  thand-server:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8081:8080"
    environment:
      - THAND_ENV=development
      - THAND_LOG_LEVEL=debug
    command: ["server", "start", "--dev"]
    volumes:
      - ./config:/app/config
      
  thand-agent:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080" 
    environment:
      - THAND_SERVER_URL=http://thand-server:8080
      - THAND_ENV=development
    command: ["agent", "start", "--dev"]
    depends_on:
      - thand-server
    volumes:
      - ./config:/app/config
```

Run with:

```bash
docker-compose -f docker-compose.dev.yml up --build
```

---

## Testing Access Requests

### Mock Providers

In development mode, you can test with mock providers:

```bash
# Request mock AWS access
./bin/agent request aws \
  --account mock-account \
  --role ReadOnlyAccess \
  --duration 1h \
  --dev

# Request mock GCP access  
./bin/agent request gcp \
  --project mock-project \
  --role viewer \
  --duration 30m \
  --dev
```

### Local Sudo Testing

Test local privilege escalation:

```bash
# Request sudo access (will prompt for approval)
./bin/agent request sudo \
  --duration 15m \
  --reason "Testing local elevation"
```

---

## Development Workflows

### Code Generation

```bash
# Generate mocks
go generate ./...

# Generate protobuf (if using gRPC)
protoc --go_out=. --go-grpc_out=. api/v1/*.proto
```

### Database Migrations

```bash
# Create migration
./bin/agent migrate create add_user_table

# Run migrations
./bin/agent migrate up

# Rollback migration
./bin/agent migrate down
```

### Hot Reloading

Install Air for automatic rebuilds:

```bash
go install github.com/cosmtrek/air@latest
```

Create `.air.toml`:

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = ["server", "start", "--dev"]
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false
```

Run with hot reload:

```bash
air
```

---

## Debugging

### VS Code Launch Configuration

Create `.vscode/launch.json`:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Server",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}",
            "args": ["server", "start", "--dev", "--port", "8081"],
            "env": {
                "THAND_ENV": "development",
                "THAND_LOG_LEVEL": "debug"
            }
        },
        {
            "name": "Launch Agent",
            "type": "go", 
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}",
            "args": ["agent", "start", "--dev", "--port", "8080"],
            "env": {
                "THAND_SERVER_URL": "http://localhost:8081",
                "THAND_ENV": "development"
            }
        }
    ]
}
```

### Debugging with Delve

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Start server with debugger
dlv debug . -- server start --dev

# In another terminal, start agent with debugger
dlv debug . -- agent start --dev
```

---

## Contributing

### Pre-commit Hooks

Set up pre-commit hooks:

```bash
# Install pre-commit
pip install pre-commit

# Install hooks
pre-commit install

# Run manually
pre-commit run --all-files
```

### Code Quality

```bash
# Run linter
golangci-lint run

# Format code
gofmt -w .
goimports -w .

# Check for security issues
gosec ./...
```

---

## Next Steps

- **[GCP Setup](gcp)** - Deploy to Google Cloud
- **[Configuration](../configuration/)** - Advanced configuration
- **[API Reference](../api)** - REST API documentation