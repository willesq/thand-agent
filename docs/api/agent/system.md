---
layout: default
title: System Endpoints
parent: Agent
grand_parent: API Reference
nav_order: 1
---

# System Endpoints

System health, readiness, and metrics endpoints.

## Health Check

Check service health and dependencies.

**GET** `/health`

### Response

```json
{
  "status": "healthy",
  "path": "/api/v1",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0 (git: abc123)",
  "services": {
    "temporal": "healthy",
    "llm": "healthy",
    "encryption": "healthy",
    "vault": "healthy",
    "scheduler": "healthy",
    "storage": "healthy"
  }
}
```

## Ready Check

Check if service is ready to handle requests.

**GET** `/ready`

### Response

```json
{
  "status": "ready",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0 (git: abc123)"
}
```

## Metrics

Get service metrics and statistics.

**GET** `/metrics`

### Response

```json
{
  "uptime": "2h30m15s",
  "total_requests": 1567,
  "roles_count": 25,
  "workflows_count": 12,
  "providers_count": 8,
  "elevate_requests": 234
}
```