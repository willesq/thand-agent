---
title: Google Workspace (GSuite) Provider
description: Google Workspace provider for user and group management
parent: Providers
grand_parent: Configuration
---

# Google Workspace (GSuite) Provider

The Google Workspace provider enables integration with Google Workspace (formerly GSuite) for user management and authentication.

## Capabilities

- **User Management**: Google Workspace user and group administration
- **Authentication**: Google OAuth2 authentication
- **Directory Integration**: Access to Google Workspace directory
- **Domain Management**: Multi-domain Google Workspace support

## Configuration Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `domain` | string | Yes | Google Workspace domain |
| `service_account_key` | string | Yes | Service account key JSON |
| `admin_email` | string | Yes | Admin email for delegation |
| `client_id` | string | No | OAuth2 client ID |
| `client_secret` | string | No | OAuth2 client secret |

## Example Configuration

```yaml
version: "1.0"
providers:
  gsuite:
    name: Google Workspace
    description: Google Workspace integration
    provider: gsuite
    enabled: true
    config:
      domain: your-company.com
      admin_email: admin@your-company.com
      service_account_key: |
        {
          "type": "service_account",
          "project_id": "YOUR_PROJECT_ID",
          "private_key_id": "YOUR_PRIVATE_KEY_ID",
          "private_key": "-----BEGIN PRIVATE KEY-----\nYOUR_PRIVATE_KEY\n-----END PRIVATE KEY-----\n",
          "client_email": "agent@YOUR_PROJECT_ID.iam.gserviceaccount.com",
          "client_id": "YOUR_CLIENT_ID",
          "auth_uri": "https://accounts.google.com/o/oauth2/auth",
          "token_uri": "https://oauth2.googleapis.com/token"
        }
```

For detailed setup instructions, refer to the [Google Workspace Admin SDK documentation](https://developers.google.com/admin-sdk/).