---
layout: default
title: Configuration
parent: API Reference
nav_order: 11
---

# Configuration Endpoints

Agent registration and configuration management.

## Pre-flight Check

Validate configuration before registration.

**POST** `/preflight`

Currently a stub endpoint for future pre-flight validation.

## Register Agent

Register an agent with the server (server mode only).

**POST** `/register`

### Request Body

```json
{
  "environment": {
    "name": "production",
    "description": "Production environment configuration"
  }
}
```

### Response

```json
{
  "success": true,
  "services": {
    "temporal": {
      "enabled": true,
      "namespace": "thand",
      "task_queue": "thand-task-queue"
    },
    "llm": {
      "enabled": true,
      "provider": "openai"
    }
  }
}
```

## Post-flight Check

Validate configuration after registration.

**POST** `/postflight`

Currently a stub endpoint for future post-flight validation.

## Sync Configuration

Get current configuration state.

**GET** `/sync`

### Response

```json
{
  "version": "1.0.0 (git: abc123)",
  "timestamp": "2024-01-15T10:30:00Z"
}
```