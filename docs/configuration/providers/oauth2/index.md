---
title: OAuth2 Provider
description: Generic OAuth2 provider for authentication with various services
parent: Providers
grand_parent: Configuration
---

# OAuth2 Provider

The OAuth2 provider enables integration with any OAuth2-compliant service, providing generic authentication capabilities through the OAuth2 authorization framework.

## Capabilities

- **Authentication**: OAuth2 authorization code flow authentication
- **Generic Integration**: Works with any OAuth2-compliant service
- **Token Management**: Access token and refresh token handling
- **Customizable Endpoints**: Configurable authorization and token endpoints

## Prerequisites

### OAuth2 Service Setup

1. **OAuth2 Service**: Access to an OAuth2-compliant service
2. **Application Registration**: Registered application with OAuth2 provider
3. **Client Credentials**: Client ID and client secret from the OAuth2 provider
4. **Redirect URI**: Configured redirect URI in the OAuth2 provider

### Required OAuth2 Configuration

- **Authorization Endpoint**: OAuth2 authorization URL
- **Token Endpoint**: OAuth2 token exchange URL
- **Client ID**: OAuth2 application client identifier
- **Client Secret**: OAuth2 application client secret

## Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `authority` | string | Yes | - | OAuth2 authority/base URL |
| `client.id` | string | Yes | - | OAuth2 client ID |
| `client.secret` | string | Yes | - | OAuth2 client secret |
| `endpoints.auth` | string | No | `/auth` | Authorization endpoint path |
| `endpoints.token` | string | No | `/token` | Token endpoint path |
| `grant` | string | No | `authorization_code` | OAuth2 grant type |
| `scopes` | array | No | `[]` | Requested OAuth2 scopes |

## Example Configurations

### Generic OAuth2 Service

```yaml
version: "1.0"
providers:
  oauth2-service:
    name: OAuth2 Service
    description: Generic OAuth2 authentication
    provider: oauth2
    enabled: true
    config:
      authority: https://oauth.example.com
      client:
        id: YOUR_CLIENT_ID
        secret: YOUR_CLIENT_SECRET
      endpoints:
        auth: /oauth/authorize
        token: /oauth/token
      scopes:
        - openid
        - profile
        - email
```

### Google OAuth2 (Alternative to oauth2.google)

```yaml
version: "1.0"
providers:
  google-oauth2:
    name: Google OAuth2
    description: Google OAuth2 authentication
    provider: oauth2
    enabled: true
    config:
      authority: https://accounts.google.com/o/oauth2
      client:
        id: YOUR_GOOGLE_CLIENT_ID
        secret: YOUR_GOOGLE_CLIENT_SECRET
      endpoints:
        auth: /auth
        token: /token
      scopes:
        - openid
        - profile
        - email
```

For detailed OAuth2 setup instructions, refer to your specific OAuth2 provider's documentation.