---
layout: default
title: Configuration
nav_order: 4  
description: "Thand Agent configuration reference and examples"
has_children: true
---

# Configuration
{: .no_toc }

Complete configuration reference for Thand Agent.
{: .fs-6 .fw-300 }

## Overview

Thand Agent uses YAML configuration files to define behavior, providers, roles, and workflows. Configuration can be provided via:

- Configuration files (`~/.thand/config.yaml`)
- Environment variables 
- Command line flags

---

## Configuration Hierarchy

Configuration is loaded in this order (later sources override earlier ones):

1. Default values
2. Configuration file (`~/.thand/config.yaml`)
3. Environment variables (prefixed with `THAND_`)
4. Command line flags

---

## Basic Configuration Structure

```yaml
# Server connection
server:
  url: "https://thand.example.com"
  timeout: 30s
  
# Agent settings  
agent:
  listen_port: 8080
  session_timeout: "1h"
  
# Logging configuration
logging:
  level: "info"      # debug, info, warn, error
  format: "json"     # json, text
  file: "/var/log/thand/agent.log"
  
# Provider configurations
providers:
  aws:
    region: "us-east-1"
  gcp:
    project_id: "my-project"
    
# Security settings
security:
  tls:
    cert_file: "/etc/thand/tls.crt"
    key_file: "/etc/thand/tls.key"
```

---

## Environment Variables

All configuration options can be set via environment variables using the `THAND_` prefix:

```bash
export THAND_SERVER_URL="https://thand.example.com"
export THAND_AGENT_LISTEN_PORT="8080"
export THAND_LOGGING_LEVEL="debug"
```

Nested configuration uses underscores:

```bash
export THAND_PROVIDERS_AWS_REGION="us-west-2"
export THAND_SECURITY_TLS_CERT_FILE="/etc/ssl/thand.crt"
```

---

## Sections

- **[Server Configuration](server)** - Server connection settings
- **[Providers](providers)** - Cloud provider configurations  
- **[Roles](roles)** - Role definitions and mappings
- **[Workflows](workflows)** - Custom approval workflows
- **[Security](security)** - TLS and authentication settings