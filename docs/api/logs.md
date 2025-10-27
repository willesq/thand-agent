---
layout: default
title: Audit and Logging
parent: API Reference
nav_order: 10
---

# Audit and Logging

Access system logs and audit events.

## Get Logs

Get system logs and audit events.

**GET** `/logs`

### Response

```json
{
  "logs": [
    {
      "timestamp": "2024-01-15T10:30:00Z",
      "level": "info",
      "message": "User alice@example.com requested elevation to admin role",
      "provider": "aws",
      "user": "alice@example.com",
      "workflow_id": "wf_abc123"
    }
  ]
}
```

### Notes

- Requires authentication in server mode
- Returns up to 500 recent log entries
- Includes system events, user actions, and audit trail
- Available in both JSON and HTML formats