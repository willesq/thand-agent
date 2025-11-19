---
layout: default
title: GitHub
description: GitHub provider for repository management and authentication
parent: Providers
grand_parent: Configuration
---

# GitHub Provider

The GitHub provider enables integration with GitHub, providing authentication and role-based access control capabilities for GitHub repositories and organizations.

## Capabilities

- **Authentication**: OAuth-based user authentication with GitHub
- **Role-Based Access Control (RBAC)**: Repository and organization-level access management
- **Repository Management**: Access to GitHub repositories and organizational data
- **User and Organization Integration**: Support for GitHub users and organization membership

## Prerequisites

### GitHub Account Setup

1. **GitHub Account**: Active GitHub account (personal or organization)
2. **Repository/Organization Access**: Appropriate permissions to repositories or organizations you want to manage
3. **API Access**: For token-based authentication, personal access token or GitHub App credentials

### Required GitHub Permissions

The following GitHub scopes/permissions are recommended:

For **Personal Access Token**:
- `repo`: Full control of private repositories
- `user`: Read user profile data
- `read:org`: Read organization membership
- `admin:org`: Full control of organizations (if managing org access)

For **GitHub App**:
- Repository permissions: Read/Write as needed
- Organization permissions: Read as needed
- User permissions: Read user profile

## Authentication Methods

The GitHub provider supports multiple authentication methods:

### 1. Personal Access Token (Recommended)

Uses a GitHub personal access token for API access:

```yaml
providers:
  github:
    name: GitHub
    provider: github
    config:
      token: YOUR_GITHUB_TOKEN
      endpoint: https://api.github.com
```

### 2. OAuth Application

Uses GitHub OAuth application for user authentication:

```yaml
providers:
  github:
    name: GitHub
    provider: github
    config:
      client_id: YOUR_OAUTH_APP_CLIENT_ID
      client_secret: YOUR_OAUTH_APP_CLIENT_SECRET
      endpoint: https://api.github.com
```

### 3. GitHub Enterprise Server

For GitHub Enterprise Server installations:

```yaml
providers:
  github-enterprise:
    name: GitHub Enterprise
    provider: github
    config:
      token: YOUR_GITHUB_TOKEN
      endpoint: https://your-github-enterprise.com/api/v3
```

## Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `token` | string | Yes* | - | GitHub personal access token |
| `client_id` | string | Yes* | - | OAuth application client ID |
| `client_secret` | string | Yes* | - | OAuth application client secret |
| `endpoint` | string | No | `https://api.github.com` | GitHub API endpoint URL |
| `organization` | string | No | - | GitHub organization name (optional) |
| `scopes` | array | No | `["repo", "user", "read:org"]` | OAuth scopes for authentication |

*Either `token` or both `client_id` and `client_secret` are required.

## Getting Credentials

### Personal Access Token Setup

1. **Create Token**: Go to GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic) → Generate new token

2. **Select Scopes**: Choose appropriate scopes based on your needs:
   - `repo`: For repository access
   - `user`: For user profile access
   - `read:org`: For organization membership
   - `admin:org`: For organization management

3. **Copy Token**: Copy the generated token immediately (it won't be shown again)

4. **Configure Agent**: Use the token in your configuration

### OAuth Application Setup

1. **Create OAuth App**: Go to GitHub → Settings → Developer settings → OAuth Apps → New OAuth App

2. **Configure Application**:
   - **Application name**: Your application name
   - **Homepage URL**: Your application homepage
   - **Authorization callback URL**: Your agent's callback URL (e.g., `https://your-agent.com/auth/github/callback`)

3. **Get Credentials**:
   - **Client ID**: Copy from the OAuth app settings
   - **Client Secret**: Generate and copy the client secret

4. **Configure Agent**: Use the client ID and secret in your configuration

### GitHub App Setup (Advanced)

1. **Create GitHub App**: Go to GitHub → Settings → Developer settings → GitHub Apps → New GitHub App

2. **Configure Permissions**: Set repository, organization, and user permissions as needed

3. **Generate Private Key**: Download the private key file

4. **Configure Agent**: Use the app ID, installation ID, and private key

## Example Configurations

### Personal Access Token Configuration

```yaml
version: "1.0"
providers:
  github:
    name: GitHub
    description: GitHub integration with personal access token
    provider: github
    enabled: true
    config:
      token: YOUR_GITHUB_PERSONAL_ACCESS_TOKEN
      endpoint: https://api.github.com
      organization: YOUR_ORGANIZATION_NAME
```

### OAuth Application Configuration

```yaml
version: "1.0"
providers:
  github-oauth:
    name: GitHub OAuth
    description: GitHub integration with OAuth
    provider: github
    enabled: true
    config:
      client_id: YOUR_OAUTH_CLIENT_ID
      client_secret: YOUR_OAUTH_CLIENT_SECRET
      endpoint: https://api.github.com
      scopes:
        - repo
        - user
        - read:org
```

### GitHub Enterprise Server Configuration

```yaml
version: "1.0"
providers:
  github-enterprise:
    name: GitHub Enterprise
    description: GitHub Enterprise Server integration
    provider: github
    enabled: true
    config:
      token: YOUR_ENTERPRISE_TOKEN
      endpoint: https://github.your-company.com/api/v3
      organization: YOUR_ENTERPRISE_ORG
```

### Multi-Organization Setup

```yaml
version: "1.0"
providers:
  github-org1:
    name: GitHub Org 1
    description: First organization
    provider: github
    enabled: true
    config:
      token: YOUR_ORG1_TOKEN
      organization: YOUR_FIRST_ORG
  
  github-org2:
    name: GitHub Org 2
    description: Second organization
    provider: github
    enabled: true
    config:
      token: YOUR_ORG2_TOKEN
      organization: YOUR_SECOND_ORG
```

## Features

### Repository Access Management

The GitHub provider can manage access to:
- Public and private repositories
- Organization repositories
- Repository teams and collaborators

### Organization Integration

When configured with an organization:
- User and team discovery
- Organization membership management
- Repository access within the organization

### Authentication Flow

For OAuth-based authentication:
1. User is redirected to GitHub for authorization
2. GitHub redirects back with authorization code
3. Agent exchanges code for access token
4. Access token is used for API requests

## Troubleshooting

### Common Issues

1. **Authentication Failures**
   - Verify personal access token is valid and has correct scopes
   - Check if token has expired
   - Ensure endpoint URL is correct for GitHub Enterprise

2. **Permission Issues**
   - Verify token has required scopes for intended operations
   - Check if organization has restrictions on third-party applications
   - Ensure user has appropriate permissions in the organization

3. **Rate Limiting**
   - GitHub has API rate limits (5000 requests/hour for authenticated requests)
   - Use appropriate caching and request patterns
   - Consider using GitHub App authentication for higher limits

### Debugging

Enable debug logging to troubleshoot GitHub provider issues:

```yaml
logging:
  level: debug
```

Look for GitHub-specific log entries to identify authentication and API issues.

### API Rate Limits

GitHub enforces rate limits:
- **Personal Access Token**: 5,000 requests per hour
- **OAuth Application**: 5,000 requests per hour per user
- **GitHub App**: 15,000 requests per hour per installation

### Environment Variables

The GitHub provider supports these environment variables:
- `GITHUB_TOKEN`: Personal access token
- `GITHUB_CLIENT_ID`: OAuth client ID
- `GITHUB_CLIENT_SECRET`: OAuth client secret

## Security Considerations

1. **Token Security**: Store personal access tokens securely and rotate them regularly
2. **Scope Limitation**: Use minimal required scopes for tokens
3. **Organization Policies**: Respect organization security policies and restrictions
4. **Audit Logging**: Monitor and log access to repositories and organization resources
