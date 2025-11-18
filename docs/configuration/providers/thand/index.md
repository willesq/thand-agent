---
layout: default
title: Thand
description: Thand provider for federated OIDC authentication
parent: Providers
grand_parent: Configuration
---

# Thand Provider

The Thand provider enables authentication through Thand's federated OIDC service, providing seamless user authentication and session management.

## Capabilities

- **Authentication**: Federated OIDC authentication flow
- **Session Management**: Automatic session creation and validation
- **User Information**: Retrieves user profile including email, name, and groups
- **Token Validation**: Bearer token authentication and validation

## Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `endpoint` | string | No | `https://auth.thand.io` | Thand authentication endpoint URL |

## Example Configurations

### Production

```yaml
version: "1.0"
providers:
  thand:
    name: Thand Production
    description: Thand federated authentication
    provider: thand
    enabled: true
    config:
      endpoint: "https://auth.thand.io"
```

### Local Development

```yaml
version: "1.0"
providers:
  thand:
    name: Thand Local
    description: Thand local development
    provider: thand
    enabled: true
    config:
      endpoint: "http://localhost:3000"
```

## How It Works

The Thand provider implements a federated OIDC authentication flow:

1. **Authorization**: Users are redirected to the Thand authentication endpoint
2. **Authentication**: User authenticates with Thand service
3. **Code Exchange**: Authorization code is exchanged for user information
4. **Session Creation**: A session is created with 1-hour expiry containing:
   - User ID (sub)
   - Email address
   - Username
   - Full name
   - Groups/roles (if available)
5. **Validation**: Sessions are validated using the bearer token against the userinfo endpoint

## User Information

The provider retrieves the following user information from the Thand service:

| Field | Description |
|-------|-------------|
| `sub` | Unique user identifier |
| `email` | User's email address |
| `email_verified` | Email verification status |
| `name` | User's full name |
| `preferred_username` | User's preferred username |
| `groups` | User's groups/roles |

## Session Expiry

Sessions created by the Thand provider have a default expiry of **1 hour**. After expiration, users will need to re-authenticate.

For detailed information about Thand's authentication service, refer to the Thand documentation.
