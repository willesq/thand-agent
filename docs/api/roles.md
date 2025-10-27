---
layout: default
title: Role Management
parent: API Reference
nav_order: 6
---

# Role Management

Manage access roles and their configurations.

## List Roles

Get all available roles.

**GET** `/roles`

### Query Parameters

- `provider` - Filter by provider (comma-separated)

### Response

```json
{
  "version": "1.0",
  "roles": {
    "admin": {
      "role": {
        "name": "admin",
        "description": "Administrative access to all systems",
        "providers": ["aws", "gcp"],
        "workflows": ["default", "emergency"],
        "enabled": true,
        "permissions": {
          "users": ["alice@example.com", "bob@example.com"],
          "groups": ["admins"]
        }
      }
    },
    "developer": {
      "role": {
        "name": "developer",
        "description": "Development environment access",
        "providers": ["aws"],
        "workflows": ["development"],
        "enabled": true,
        "permissions": {
          "groups": ["developers"]
        }
      }
    }
  }
}
```

## Get Role Details

**GET** `/role/{role}`

### Response

```json
{
  "name": "admin",
  "description": "Administrative access to all systems",
  "providers": ["aws", "gcp"],
  "workflows": ["default", "emergency"],
  "enabled": true,
  "permissions": {
    "users": ["alice@example.com"],
    "groups": ["admins"]
  }
}
```