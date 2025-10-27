---
layout: default
title: Getting Started
nav_order: 2
description: "Get started with Thand Agent - installation, configuration, and first steps"
---

# Getting Started
{: .no_toc }

Get up and running with Thand Agent quickly and easily.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Installation

### Prerequisites

Before installing Thand Agent, ensure you have:

- Go 1.21 or later (if building from source)
- Access to your target infrastructure (GCP, AWS, etc.)
- Network connectivity to the Thand server

### Install via Binary Release

Download the latest binary for your platform from the [GitHub Releases](https://github.com/thand-io/agent/releases) page:

```bash
# Linux/macOS
curl -L -o thand-agent https://github.com/thand-io/agent/releases/latest/download/agent-$(uname -s)-$(uname -m)
chmod +x thand-agent
sudo mv thand-agent /usr/local/bin/
```

```bash
# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/thand-io/agent/releases/latest/download/agent-windows-amd64.exe" -OutFile "thand-agent.exe"
```

### Build from Source

Clone the repository and build the agent:

```bash
git clone https://github.com/thand-io/agent.git
cd agent
go build -o bin/agent .
```

The binary will be available at `bin/agent`.

---

## Configuration

### Basic Configuration

Create a configuration file at `~/.thand/config.yaml`:

```yaml
server:
  url: "https://your-thand-server.com"
  
agent:
  listen_port: 8080
  session_timeout: "1h"
  
logging:
  level: "info"
  format: "json"
```

### Environment Variables

You can also configure the agent using environment variables:

```bash
export THAND_SERVER_URL="https://your-thand-server.com"
export THAND_AGENT_PORT="8080"
export THAND_LOG_LEVEL="info"
```

---

## First Steps

### 1. Start the Agent

```bash
thand-agent start
```

The agent will start and listen on the configured port (default: 8080).

### 2. Authenticate

Initialize your authentication with the Thand server:

```bash
thand-agent auth login
```

This will open your browser for authentication or provide a device code flow.

### 3. Request Access

Request access to a resource:

```bash
# Request AWS access for 1 hour
thand-agent request aws --role ReadOnlyAccess --duration 1h

# Request GCP access
thand-agent request gcp --project my-project --role viewer
```

### 4. Use Your Access

Once approved, the agent will provide temporary credentials:

```bash
# AWS credentials are automatically configured
aws s3 ls

# GCP credentials are set via environment
gcloud projects list
```

---

## Common Use Cases

### Local Sudo Access

Request temporary sudo access on your local machine:

```bash
thand-agent request sudo --duration 30m --reason "System maintenance"
```

### Cloud Infrastructure Access

Request cloud access for specific tasks:

```bash
# AWS - Deploy to production
thand-agent request aws \
  --account 123456789012 \
  --role DeploymentRole \
  --duration 2h \
  --reason "Production deployment v1.2.3"

# GCP - Debug application
thand-agent request gcp \
  --project production-app \
  --role roles/logging.viewer \
  --duration 1h \
  --reason "Investigating error logs"
```

### Application Access

Request access to SaaS applications:

```bash
thand-agent request app \
  --app salesforce \
  --role admin \
  --duration 30m \
  --reason "User account recovery"
```

---

## Troubleshooting

### Agent Won't Start

1. Check if the port is already in use:
   ```bash
   lsof -i :8080
   ```

2. Verify configuration file syntax:
   ```bash
   thand-agent config validate
   ```

3. Check logs for detailed error messages:
   ```bash
   thand-agent logs
   ```

### Authentication Issues

1. Verify server URL is correct and accessible:
   ```bash
   curl -I https://your-thand-server.com/health
   ```

2. Clear cached authentication and re-login:
   ```bash
   thand-agent auth logout
   thand-agent auth login
   ```

### Access Requests Denied

1. Check your user permissions with your administrator
2. Verify the requested role exists and you're eligible
3. Review audit logs for denial reasons:
   ```bash
   thand-agent audit list --user $(whoami)
   ```

---

## Next Steps

- **[Setup Guides](../setup/)** - Configure Thand for specific platforms
- **[Configuration](../configuration/)** - Detailed configuration options
- **[API Reference](../api/)** - REST API documentation
- **[Workflows](../workflows/)** - Custom workflow configuration