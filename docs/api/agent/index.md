---
layout: default
title: Agent
nav_order: 1
parent: API Reference
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

## Interactive API Documentation

The complete API specification is available through Swagger UI:

```
http://localhost:8080/swagger/index.html
```

### OpenAPI Specification

- **JSON Format**: `http://localhost:8080/swagger/doc.json`
- **YAML Format**: `http://localhost:8080/swagger/doc.yaml`

The Swagger UI provides:
- Interactive API testing
- Complete request/response schemas
- Authentication configuration
- Real-time API exploration

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

## License

This API is licensed under the Business Source License (BSL 1.1).

**License URL**: [https://mariadb.com/bsl11/](https://mariadb.com/bsl11/)