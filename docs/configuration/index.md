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
```

---

## Sections

- **[Configuration Reference](file)** - Complete reference for all configuration options
- **[Environment](environment)** - Environment-specific configuration
- **[CLI](cli)** - Command-line interface configuration
- **[Providers](providers)** - Provider configurations  
- **[Roles](roles)** - Role definitions and mappings
- **[Workflows](workflows)** - Custom approval workflows
