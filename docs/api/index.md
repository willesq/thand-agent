---
layout: default
title: API Reference
nav_order: 10
description: "Thand Agent REST API documentation"
has_children: true
---

# API Reference
{: .no_toc }

Complete REST API documentation for Thand Agent.
{: .fs-6 .fw-300 }

## Base URL

All API endpoints are relative to the agent's base URL:

```
http://localhost:8080/api/v1
```

## Authentication

The API supports multiple authentication methods depending on the mode:

- **Session Cookies**: Automatically managed for browser-based authentication
- **OAuth2 Flows**: Provider-specific authentication via redirect flows
- **Local Sessions**: Encrypted session tokens for cross-service communication

## Content Types

The API supports multiple content types:

- **JSON API**: `application/json` (default for API responses)
- **HTML Pages**: `text/html` (for web interface)
- **Form Data**: `application/x-www-form-urlencoded` (for form submissions)
- **Multipart Forms**: `multipart/form-data` (for file uploads)

## Mode-Specific Behavior

The API behaves differently based on the service mode:

### Server Mode
- Full API available
- User authentication required
- Multi-user session management
- Workflow execution and management
- Identity provider integration

### Agent Mode  
- Limited API endpoints
- Local session management
- Proxy requests to login server
- Session synchronization

### Client Mode
- Minimal API surface
- Session creation and management
- Authentication delegation

## Error Responses

All error responses follow this format:

```json
{
  "code": 400,
  "title": "Invalid Request",
  "message": "The request is missing required parameters"
}
```

### Common HTTP Status Codes

- `200` - Success
- `201` - Created
- `302` - Redirect (for OAuth flows)
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `500` - Internal Server Error