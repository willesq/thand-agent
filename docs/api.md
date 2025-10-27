---
layout: default
title: API Reference
nav_order: 5
description: "Thand Agent REST API documentation and examples"
---

# API Reference
{: .no_toc }

Complete REST API documentation for Thand Agent.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Base URL

All API endpoints are relative to the agent's base URL:

```
http://localhost:8080/api/v1
```

## Authentication

API requests require authentication via:

- **Bearer Token**: Include in `Authorization: Bearer <token>` header
- **Session Cookie**: Automatically set after web login

---

## Endpoints

### Health Check

Check agent health and status.

**GET** `/health`

#### Response

```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "2h30m15s",
  "server_connected": true
}
```

### Authentication

#### Login

Initiate authentication flow.

**POST** `/auth/login`

#### Request Body

```json
{
  "redirect_uri": "http://localhost:8080/callback"
}
```

#### Response

```json
{
  "auth_url": "https://server.com/oauth/authorize?...",
  "device_code": "abc123",
  "expires_in": 600
}
```

#### Logout

**POST** `/auth/logout`

### Access Requests

#### Create Access Request

Request access to a resource.

**POST** `/requests`

#### Request Body

```json
{
  "provider": "aws",
  "resource": {
    "account": "123456789012",
    "role": "ReadOnlyAccess"
  },
  "duration": "1h",
  "reason": "Investigating production issue #123"
}
```

#### Response

```json
{
  "id": "req_abc123",
  "status": "pending",
  "provider": "aws",
  "resource": {
    "account": "123456789012", 
    "role": "ReadOnlyAccess"
  },
  "duration": "1h",
  "reason": "Investigating production issue #123",
  "created_at": "2024-01-15T10:30:00Z",
  "expires_at": "2024-01-15T11:30:00Z"
}
```

#### List Requests

Get all access requests for the authenticated user.

**GET** `/requests`

#### Query Parameters

- `status` - Filter by status (`pending`, `approved`, `denied`, `expired`)
- `provider` - Filter by provider (`aws`, `gcp`, `azure`)
- `limit` - Maximum number of results (default: 50)
- `offset` - Pagination offset

#### Response

```json
{
  "requests": [
    {
      "id": "req_abc123",
      "status": "approved",
      "provider": "aws",
      "resource": {
        "account": "123456789012",
        "role": "ReadOnlyAccess"
      },
      "duration": "1h",
      "reason": "Investigating production issue #123",
      "created_at": "2024-01-15T10:30:00Z",
      "approved_at": "2024-01-15T10:32:00Z",
      "expires_at": "2024-01-15T11:30:00Z"
    }
  ],
  "total": 1,
  "has_more": false
}
```

#### Get Request Details

**GET** `/requests/{id}`

#### Response

```json
{
  "id": "req_abc123", 
  "status": "approved",
  "provider": "aws",
  "resource": {
    "account": "123456789012",
    "role": "ReadOnlyAccess"
  },
  "duration": "1h",
  "reason": "Investigating production issue #123",
  "created_at": "2024-01-15T10:30:00Z",
  "approved_at": "2024-01-15T10:32:00Z", 
  "expires_at": "2024-01-15T11:30:00Z",
  "credentials": {
    "aws_access_key_id": "AKIAIOSFODNN7EXAMPLE",
    "aws_secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
    "aws_session_token": "AQoDYXdzEPT//////////wEXAMPLEtc764bNrC9SAPBSM22wDOk4x4HIZ8j4FZTwdQWLWsKWHGBuFqwAeMicRXmxfpSPfIeoIYRqTflfKD8YUuwthAx7mSEI/qkPpKPi/kMcGdQrmGdeehM4IC1NtBmUpp2wUE8phUZampKsburEDy0KPkyQDYwT7WZ0wq5VSXDvp75YU9HFvlRd8Tx6q6fE8YQcHNVXAkiY9q6d+xo0rKwT38xVqr7ZD0u0iPPkUL64lIZbqBAz+scqKmlzm8FDrypNC9Yjc8fPOLn9FX9KSYvKTr4rvx3iSIlTJabIQwj2ICCR/oLxBA=="
  },
  "audit_trail": [
    {
      "action": "request_created",
      "timestamp": "2024-01-15T10:30:00Z",
      "user": "alice@example.com"
    },
    {
      "action": "request_approved", 
      "timestamp": "2024-01-15T10:32:00Z",
      "approver": "bob@example.com"
    }
  ]
}
```

#### Revoke Request

Revoke active access.

**DELETE** `/requests/{id}`

#### Response

```json
{
  "id": "req_abc123",
  "status": "revoked",
  "revoked_at": "2024-01-15T10:45:00Z"
}
```

### Sessions

#### Get Current Session

**GET** `/sessions/current`

#### Response

```json
{
  "user": {
    "id": "user_123",
    "email": "alice@example.com",
    "name": "Alice Smith"
  },
  "permissions": [
    "aws:request",
    "gcp:request"
  ],
  "active_requests": [
    "req_abc123"
  ]
}
```

### Audit Logs

#### Get Audit Events

**GET** `/audit`

#### Query Parameters

- `user` - Filter by user email
- `action` - Filter by action type
- `start_time` - Start time (ISO 8601)
- `end_time` - End time (ISO 8601)
- `limit` - Maximum results (default: 100)

#### Response

```json
{
  "events": [
    {
      "id": "evt_123",
      "timestamp": "2024-01-15T10:30:00Z",
      "user": "alice@example.com",
      "action": "access_granted",
      "resource": {
        "provider": "aws",
        "account": "123456789012",
        "role": "ReadOnlyAccess"
      },
      "metadata": {
        "request_id": "req_abc123",
        "ip_address": "192.168.1.100",
        "user_agent": "thand-agent/1.0.0"
      }
    }
  ],
  "total": 1
}
```

---

## Error Responses

All error responses follow this format:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "The request is missing required parameters",
    "details": {
      "missing_fields": ["provider", "resource"]
    }
  }
}
```

### Common Error Codes

- `invalid_request` - Request validation failed
- `unauthorized` - Authentication required
- `forbidden` - Insufficient permissions  
- `not_found` - Resource not found
- `conflict` - Resource already exists
- `rate_limited` - Too many requests
- `internal_error` - Server error

---

## SDK Examples

### Go

```go
package main

import (
    "context"
    "github.com/thand-io/agent/sdk/go"
)

func main() {
    client := thand.NewClient("http://localhost:8080")
    
    // Authenticate
    err := client.Login(context.Background())
    if err != nil {
        panic(err)
    }
    
    // Request access
    req := &thand.AccessRequest{
        Provider: "aws",
        Resource: map[string]string{
            "account": "123456789012",
            "role": "ReadOnlyAccess",
        },
        Duration: "1h",
        Reason: "Debug production issue",
    }
    
    resp, err := client.RequestAccess(context.Background(), req)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Request ID: %s\n", resp.ID)
}
```

### Python

```python
import thand

client = thand.Client("http://localhost:8080")

# Authenticate  
client.login()

# Request access
response = client.request_access({
    "provider": "aws",
    "resource": {
        "account": "123456789012", 
        "role": "ReadOnlyAccess"
    },
    "duration": "1h",
    "reason": "Debug production issue"
})

print(f"Request ID: {response['id']}")
```

### cURL

```bash
# Authenticate and get token
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"redirect_uri": "http://localhost:8080/callback"}' | \
  jq -r '.access_token')

# Request access
curl -X POST http://localhost:8080/api/v1/requests \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "aws",
    "resource": {
      "account": "123456789012",
      "role": "ReadOnlyAccess"  
    },
    "duration": "1h",
    "reason": "Debug production issue"
  }'
```