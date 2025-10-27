---
layout: default
title: CLI Reference
parent: Configuration
nav_order: 2
description: "Complete reference for all Thand Agent CLI commands and options"
---

# CLI Reference
{: .no_toc }

Complete reference for all Thand Agent command-line interface options and subcommands.
{: .fs-6 .fw-300 }

## Table of Contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Overview

The Thand Agent CLI provides multiple commands for different use cases:
- **Client mode**: Request just-in-time access to resources
- **Server mode**: Run the agent as a service for handling requests
- **Service management**: Install and manage the agent as a system service

---

## Global Flags

These flags are available for all commands:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--config` | - | string | Config file (default is `$HOME/.config/thand/config.yaml`) |
| `--verbose` | `-v` | boolean | Enable verbose output for debugging |
| `--login-server` | - | string | Override the default login server URL |
| `--api-key` | - | string | Provide API key for login server authentication |
| `--help` | `-h` | boolean | Show help for any command |

### Examples

```bash
# Use custom config file
agent --config /path/to/config.yaml roles

# Override login server
agent --login-server https://auth.example.com login

# Enable verbose logging
agent --verbose server

# Use API key authentication
agent --api-key "your-api-key" roles
```

---

## Main Command

### `agent`

The main command runs an interactive request wizard when called without subcommands.

```bash
agent [reason for access]
```

**Examples:**
```bash
# Interactive wizard
agent

# Direct request with reason
agent "Need access to production database for debugging"
```

**Behavior:**
- If no login server is configured, prompts for setup
- If configured, launches interactive access request wizard
- Collects provider, role, duration, and reason for access
- Submits elevation request automatically

---

## Authentication Commands

### `login`

Authenticate with the login server and establish a session.

```bash
agent login
```

**What it does:**
- Opens browser to login server authentication page
- Establishes local callback server to receive auth tokens
- Stores session for future CLI operations
- Validates successful authentication

**Examples:**
```bash
# Login to configured server
agent login

# Login with custom server
agent --login-server https://auth.example.com login
```

### `sessions`

Interactive session management interface.

```bash
agent sessions
```

**Features:**
- List all active authentication sessions
- Create new provider-specific sessions
- Remove expired or unwanted sessions
- Refresh existing sessions
- Interactive menu-driven interface

---

## Access Request Commands

### `request`

Make AI-powered access requests using natural language.

```bash
agent request [reason]
```

**Examples:**
```bash
agent request "Need to debug production issue in AWS"
agent request "Quarterly analysis requires Snowflake access"
agent request "Emergency database maintenance required"
```

**How it works:**
- Sends natural language reason to login server LLM
- AI determines appropriate role and resources
- Automatically submits elevation request
- Returns request status and next steps

### `request access`

Make structured access requests with specific parameters.

```bash
agent request access --resource <resource> --role <role> --duration <duration> --reason <reason>
```

**Required Flags:**

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--resource` | `-r` | Resource/provider to access | `aws-prod`, `snowflake-dev` |
| `--role` | `-o` | Role to assume | `admin`, `analyst`, `readonly` |
| `--duration` | `-d` | Access duration | `1h`, `4h`, `8h` |
| `--reason` | `-e` | Justification for access | `Emergency maintenance` |

**Examples:**
```bash
# Request AWS admin access
agent request access \
  --resource aws-prod \
  --role admin \
  --duration 2h \
  --reason "Emergency security patch deployment"

# Request read-only Snowflake access
agent request access \
  -r snowflake-prod \
  -o analyst \
  -d 4h \
  -e "Monthly report generation"
```

---

## Information Commands

### `roles`

List available roles and their descriptions.

```bash
agent roles [flags]
```

**Flags:**

| Flag | Type | Description |
|------|------|-------------|
| `--provider` | string | Filter roles by provider |

**Examples:**
```bash
# List all roles
agent roles

# List AWS-specific roles
agent roles --provider aws

# List roles for multiple providers
agent roles --provider gcp
```

**Output Format:**
```
Available roles:

NAME                 PROVIDERS       DESCRIPTION
----                 ---------       -----------
aws-admin           aws             Full administrative access to AWS
aws-readonly        aws             Read-only access to AWS resources
snowflake-analyst   snowflake       Data analysis access to Snowflake
gcp-developer       gcp             Development access to GCP

Total: 4 roles
```

### `config`

Display current agent configuration.

```bash
agent config
```

**Shows:**
- Server host and port settings
- Login server endpoint
- Current logging level
- Other key configuration values

### `version`

Display version information and check for updates.

```bash
agent version
```

**Output includes:**
- Current version number
- Git commit hash (if available)
- Update availability status
- Instructions for updating

---

## Server Commands

### `server`

Run the agent server in the foreground.

```bash
agent server
```

**What it does:**
- Starts HTTP server on configured host:port
- Loads roles, workflows, and providers
- Handles authentication callbacks
- Provides API endpoints for elevation requests
- Runs until interrupted (Ctrl+C)

**Output includes:**
- Environment information
- Server startup status
- Request handling logs

---

## Service Management Commands

The service commands manage the Thand Agent as a system service.

### `service install`

Install the agent as a system service.

```bash
agent service install
```

**Requirements:**
- Administrative/root privileges on most systems
- Service will start automatically on boot

**Platform-specific instructions:**
```bash
# Linux/macOS
sudo agent service install

# Windows (run as Administrator)
agent service install
```

### `service start`

Start the agent system service.

```bash
agent service start
```

### `service stop`

Stop the agent system service.

```bash
agent service stop
```

### `service status`

Check the agent service status.

```bash
agent service status
```

**Output:**
- 🟢 Running: Service is active
- Stopped: Service is not running
- 🟡 Unknown: Service state unclear

### `service remove`

Uninstall the agent system service.

```bash
agent service remove
```

**What it does:**
- Stops the service if running
- Removes service from system startup
- Cleans up service files

---

## Update Commands

### `update`

Update the agent to the latest version.

```bash
agent update [flags]
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Update without confirmation prompt |
| `--check` | `-c` | Check for updates without installing |

**Examples:**
```bash
# Interactive update
agent update

# Force update without prompts
agent update --force

# Check for updates only
agent update --check
```

**Update Process:**
1. Checks GitHub for latest release
2. Shows release notes and version info
3. Prompts for confirmation (unless `--force`)
4. Downloads and installs new version
5. Requires agent restart to use new version

---

## Environment Variables

Configuration options can be set via environment variables with the `THAND_` prefix. However, CLI-specific flags like `--verbose`, `--config`, `--login-server`, and `--api-key` are **only available as command-line flags** and do not have corresponding environment variables.

The following environment variables are available for configuration settings:

```bash
# Environment settings
export THAND_ENVIRONMENT_PLATFORM="aws"
export THAND_ENVIRONMENT_CONFIG_API_KEY="your-api-key"
export THAND_ENVIRONMENT_CONFIG_TIMEOUT="10s"

# Cloud provider settings (AWS)
export THAND_ENVIRONMENT_CONFIG_PROFILE="default"
export THAND_ENVIRONMENT_CONFIG_REGION="us-west-2"
export THAND_ENVIRONMENT_CONFIG_ACCESS_KEY_ID="your-access-key"
export THAND_ENVIRONMENT_CONFIG_SECRET_ACCESS_KEY="your-secret-key"
export THAND_ENVIRONMENT_CONFIG_KMS_ARN="arn:aws:kms:..."
export THAND_ENVIRONMENT_CONFIG_IMSD_DISABLE="true"

# Cloud provider settings (GCP)
export THAND_ENVIRONMENT_CONFIG_PROJECT_ID="my-project"
export THAND_ENVIRONMENT_CONFIG_LOCATION="us-central1"
export THAND_ENVIRONMENT_CONFIG_KEY_RING="my-key-ring"
export THAND_ENVIRONMENT_CONFIG_KEY_NAME="my-key"

# Cloud provider settings (Azure)
export THAND_ENVIRONMENT_CONFIG_VAULT_URL="https://vault.vault.azure.net/"

# Vault settings
export THAND_ENVIRONMENT_CONFIG_SECRET_PATH="secret/path"
export THAND_ENVIRONMENT_CONFIG_MOUNT_PATH="secret"

# Logging
export THAND_LOGGING_LEVEL="debug"
export THAND_LOGGING_FORMAT="json"
export THAND_LOGGING_OUTPUT="stdout"

# Services
export THAND_SERVICES_LLM_PROVIDER="openai"
export THAND_SERVICES_LLM_API_KEY="sk-..."
export THAND_SERVICES_LLM_BASE_URL="https://api.openai.com/v1"
export THAND_SERVICES_LLM_MODEL="gpt-4"

# Temporal
export THAND_SERVICES_TEMPORAL_HOST="temporal.example.com"
export THAND_SERVICES_TEMPORAL_PORT="7233"
export THAND_SERVICES_TEMPORAL_NAMESPACE="production"
export THAND_SERVICES_TEMPORAL_MTLS_PEM="-----BEGIN CERTIFICATE-----..."
export THAND_SERVICES_TEMPORAL_API_KEY="your-temporal-api-key"

# External sources
export THAND_ROLES_VAULT="secret/roles"
export THAND_WORKFLOWS_VAULT="secret/workflows"
export THAND_PROVIDERS_VAULT="secret/providers"
```

**Note**: CLI flags like `--verbose`, `--config`, `--login-server`, and `--api-key` must be specified on the command line and cannot be set via environment variables.

---

## Configuration Integration

The CLI integrates with the configuration system:

### Precedence Order
1. Command-line flags (highest priority)
2. Environment variables
3. Configuration file
4. Default values (lowest priority)

### Example Workflow
```bash
# 1. Configure login server
cat > ~/.config/thand/config.yaml << EOF
login:
  endpoint: "https://auth.company.com"
  api_key: "your-api-key"
EOF

# 2. Login and authenticate
agent login

# 3. View available access options
agent roles

# 4. Request access
agent request "Need production access for incident response"
```

---

## Error Handling

### Common Error Messages

**"No login server configured"**
- Solution: Configure login server in config file or use `--login-server` flag

**"Authentication required but login was declined"**
- Solution: Run `agent login` to authenticate

**"Role not found"**
- Solution: Check available roles with `agent roles`

**"Failed to install service"**
- Solution: Run with elevated privileges (`sudo` or "Run as Administrator")

### Debug Mode

Enable verbose output for troubleshooting:

```bash
agent --verbose [command]
```

This provides detailed logging for:
- Configuration loading
- API requests/responses
- Authentication flows
- Error diagnostics

---

## Interactive Features

### Request Wizard

The main `agent` command provides an interactive wizard:

1. **Provider Selection**: Choose from configured providers
2. **Role Selection**: Pick appropriate role for selected provider
3. **Duration**: Select access duration (1h, 2h, 4h, 8h, custom)
4. **Reason**: Enter justification for access
5. **Summary**: Review and confirm request

### Session Manager

The `agent sessions` command provides interactive session management:

- Navigate with arrow keys
- Select actions from menu
- View detailed session information
- Manage multiple provider sessions

---

## Integration Examples

### CI/CD Pipeline
```bash
# Automated access request in pipeline
agent --api-key "$THAND_API_KEY" request access \
  --resource aws-prod \
  --role deployer \
  --duration 1h \
  --reason "Automated deployment pipeline"
```

### Emergency Access
```bash
# Quick emergency access
agent request "Production outage - need immediate admin access"
```

### Scheduled Maintenance
```bash
# Planned maintenance window
agent request access \
  --resource all-systems \
  --role maintenance \
  --duration 4h \
  --reason "Scheduled maintenance window MW-2024-10-27"
```
