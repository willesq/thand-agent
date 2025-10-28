---
title: Terraform Provider
description: Terraform Cloud/Enterprise provider for workspace and organization management
parent: Providers
grand_parent: Configuration
---

# Terraform Provider

The Terraform provider enables integration with Terraform Cloud and Terraform Enterprise, providing role-based access control and workspace management capabilities.

## Capabilities

- **Authentication**: API token-based authentication with Terraform Cloud/Enterprise
- **Role-Based Access Control (RBAC)**: Organization and workspace-level access management
- **Workspace Management**: Access to Terraform workspaces and runs
- **Organization Integration**: Support for Terraform Cloud organizations and teams

## Prerequisites

### Terraform Cloud/Enterprise Setup

1. **Terraform Cloud Account**: Active Terraform Cloud account or Terraform Enterprise installation
2. **Organization Access**: Access to a Terraform Cloud organization
3. **API Token**: User or team API token with appropriate permissions

### Required Terraform Permissions

The API token should have the following permissions:
- **Organization access**: Read organization details
- **Workspace access**: Read/write workspace configurations
- **Team management**: Manage team memberships (if applicable)
- **Run access**: View and manage Terraform runs

## Authentication Methods

### 1. User API Token (Recommended)

Uses a Terraform Cloud user API token:

```yaml
providers:
  terraform:
    name: Terraform Cloud
    provider: terraform
    config:
      token: YOUR_TERRAFORM_CLOUD_TOKEN
```

### 2. Team API Token

Uses a Terraform Cloud team API token:

```yaml
providers:
  terraform:
    name: Terraform Cloud
    provider: terraform
    config:
      token: YOUR_TEAM_API_TOKEN
      organization: YOUR_ORG_NAME
```

### 3. Terraform Enterprise

For Terraform Enterprise installations:

```yaml
providers:
  terraform-enterprise:
    name: Terraform Enterprise
    provider: terraform
    config:
      token: YOUR_TFE_TOKEN
      hostname: YOUR_TFE_HOSTNAME
```

## Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `token` | string | Yes | - | Terraform Cloud/Enterprise API token |
| `organization` | string | No | - | Terraform organization name |
| `hostname` | string | No | `app.terraform.io` | Terraform hostname (for Enterprise) |

## Getting Credentials

### User API Token Setup

1. **Login to Terraform Cloud**: Go to [Terraform Cloud](https://app.terraform.io/)

2. **Create API Token**: 
   - Go to User Settings → Tokens
   - Click "Create an API token"
   - Enter a description and click "Create"

3. **Copy Token**: Copy the generated token immediately

4. **Configure Agent**: Use the token in your configuration

### Team API Token Setup

1. **Access Organization**: Go to your organization in Terraform Cloud

2. **Create Team Token**:
   - Go to Settings → Teams
   - Select a team or create a new one
   - Go to Team Settings → Team API Token
   - Click "Generate a token"

3. **Copy Token**: Copy the generated token

4. **Configure Agent**: Use the token and organization name in your configuration

### Terraform Enterprise Token

1. **Access TFE Instance**: Go to your Terraform Enterprise instance

2. **Create User Token**: Follow the same process as Terraform Cloud

3. **Configure Hostname**: Include your TFE hostname in the configuration

## Example Configurations

### Terraform Cloud with User Token

```yaml
version: "1.0"
providers:
  terraform-cloud:
    name: Terraform Cloud
    description: Terraform Cloud integration
    provider: terraform
    enabled: true
    config:
      token: YOUR_TERRAFORM_CLOUD_USER_TOKEN
```

### Terraform Cloud with Organization

```yaml
version: "1.0"
providers:
  terraform-org:
    name: Terraform Organization
    description: Terraform Cloud organization access
    provider: terraform
    enabled: true
    config:
      token: YOUR_TEAM_API_TOKEN
      organization: YOUR_ORGANIZATION_NAME
```

### Terraform Enterprise

```yaml
version: "1.0"
providers:
  terraform-enterprise:
    name: Terraform Enterprise
    description: Terraform Enterprise integration
    provider: terraform
    enabled: true
    config:
      token: YOUR_TFE_TOKEN
      hostname: terraform.your-company.com
```

### Multi-Organization Setup

```yaml
version: "1.0"
providers:
  terraform-prod:
    name: Terraform Production
    description: Production organization
    provider: terraform
    enabled: true
    config:
      token: YOUR_PROD_TEAM_TOKEN
      organization: YOUR_PROD_ORG
  
  terraform-dev:
    name: Terraform Development
    description: Development organization
    provider: terraform
    enabled: true
    config:
      token: YOUR_DEV_TEAM_TOKEN
      organization: YOUR_DEV_ORG
```

## Features

### Workspace Access Management

The Terraform provider can manage access to:
- Terraform workspaces
- Workspace variables and configurations
- Terraform runs and state files

### Organization Integration

When configured with an organization:
- Team and user discovery
- Organization-wide permissions
- Cross-workspace access management

### Permission Levels

The provider supports Terraform's built-in permission levels:
- **Admin**: Full administrative access
- **Read**: Read-only access to workspaces and runs
- **Write**: Read/write access to workspaces and runs

## Troubleshooting

### Common Issues

1. **Authentication Failures**
   - Verify API token is valid and not expired
   - Check if token has appropriate permissions
   - Ensure organization name is correct (if specified)

2. **Permission Issues**
   - Verify token has access to the specified organization
   - Check if user/team has workspace permissions
   - Ensure API access is enabled for the organization

3. **Terraform Enterprise Connection**
   - Verify hostname is correct and accessible
   - Check if TLS/SSL certificates are valid
   - Ensure firewall/network access is configured

### Debugging

Enable debug logging to troubleshoot Terraform provider issues:

```yaml
logging:
  level: debug
```

Look for Terraform-specific log entries to identify authentication and API issues.

### API Rate Limits

Terraform Cloud has API rate limits:
- **Free tier**: 30 requests per minute
- **Paid tiers**: Higher limits based on plan

### Environment Variables

The Terraform provider supports these environment variables:
- `TFE_TOKEN`: Terraform API token
- `TFE_HOSTNAME`: Terraform Enterprise hostname

## Security Considerations

1. **Token Security**: Store API tokens securely and rotate them regularly
2. **Least Privilege**: Use team tokens with minimal required permissions
3. **Organization Policies**: Respect organization security policies
4. **Audit Logging**: Monitor workspace and run access patterns
