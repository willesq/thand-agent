---
layout: default
title: Google OAuth2
description: Google OAuth2 provider for Google account authentication
parent: Providers
grand_parent: Configuration
---

# Google OAuth2 Provider

The Google OAuth2 provider enables authentication with Google accounts using OAuth2.

## Capabilities

- **Authentication**: Google account OAuth2 authentication
- **Google Services**: Access to Google user profile and services
- **OpenID Connect**: Support for OpenID Connect with Google
- **Scope Management**: Configurable OAuth2 scopes

## Configuration Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `client_id` | string | Yes | Google OAuth2 client ID |
| `client_secret` | string | Yes | Google OAuth2 client secret |
| `scopes` | array | No | OAuth2 scopes to request |
| `hosted_domain` | string | No | Restrict to specific Google Workspace domain |

## Example Configuration

```yaml
version: "1.0"
providers:
  google-oauth:
    name: Google OAuth2
    description: Google account authentication
    provider: oauth2.google
    enabled: true
    config:
      client_id: YOUR_GOOGLE_CLIENT_ID.apps.googleusercontent.com
      client_secret: YOUR_GOOGLE_CLIENT_SECRET
      scopes:
        - openid
        - profile
        - email
      hosted_domain: your-company.com
```

### Getting Google OAuth2 Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create or select a project
3. Enable the Google+ API
4. Go to Credentials → Create Credentials → OAuth 2.0 Client IDs
5. Configure the OAuth consent screen
6. Create OAuth2 credentials for a web application
7. Add authorized redirect URIs

For detailed setup instructions, refer to the [Google OAuth2 documentation](https://developers.google.com/identity/protocols/oauth2/).