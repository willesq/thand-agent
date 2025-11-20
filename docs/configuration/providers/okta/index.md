---
layout: default
title: Okta
description: Okta provider with identity and administrator role management
parent: Providers
grand_parent: Configuration
---

# Okta Provider

The Okta provider enables integration with Okta for identity and access management. It supports RBAC (Role-Based Access Control) and identity management capabilities through Okta's administrator roles and user management.

## Capabilities

- **Role-Based Access Control (RBAC)**: Supports Okta's predefined administrator roles
- **Identity Management**: Synchronizes users and groups from Okta
- **Permission Management**: Fine-grained permission controls for Okta resources
- **Resource Management**: Tracks Okta applications and resources
- **Search & Discovery**: Fast search across users, groups, and roles

## Prerequisites

### Okta Organization Setup

1. **Okta Account**: Active Okta organization (e.g., `https://your-domain.okta.com`)
2. **Administrator Access**: Admin privileges to create API tokens
3. **API Token**: API token with appropriate permissions

### Required Permissions

The API token must have sufficient permissions to:
- Read users and groups
- Read applications
- Read administrator roles
- Manage role assignments (if you want to grant roles through the agent)

**Recommended**: Use a token from an account with **Read-Only Administrator** or **Super Administrator** privileges for full functionality.

## Authentication Method

The Okta provider uses API token authentication.

### Generating an API Token

To create an API token in Okta:

1. **Sign in** to your Okta organization as an administrator
2. **Navigate** to **Security** > **API** in the Admin Console
3. **Click** on the **Tokens** tab
4. **Click** **Create Token**
5. **Enter** a name for your token (e.g., "Thand Agent Integration")
6. **Click** **Create Token**
7. **⚠️ Important**: Copy the token value immediately - you won't be able to see it again
8. **Store** the token securely (e.g., in a password manager or secrets management system)

## Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `endpoint` | string | Yes | - | Your Okta organization URL (e.g., `https://your-domain.okta.com`) |
| `token` | string | Yes | - | The API token generated from your Okta organization |

## Example Configurations

### Production Environment

```yaml
version: "1.0"
providers:
  okta-prod:
    name: Okta Production
    description: Production Okta environment
    provider: okta
    enabled: true
    config:
      endpoint: https://your-domain.okta.com
      token: <your-api-token-here>
```

### Development Environment

```yaml
version: "1.0"
providers:
  okta-dev:
    name: Okta Development
    description: Development Okta environment
    provider: okta
    enabled: true
    config:
      endpoint: https://your-domain-dev.okta.com
      token: 00XyZaBcDeFgHiJkLmNoPqRsTuVw9876543210xy
```

### Multi-Environment Setup

```yaml
version: "1.0"
providers:
  okta-prod:
    name: Okta Production
    description: Production Okta environment
    provider: okta
    enabled: true
    config:
      endpoint: https://your-domain.okta.com
      token: 00aBcDeFgHiJkLmNoPqRsTuVwXyZ1234567890ab
  
  okta-dev:
    name: Okta Development
    description: Development Okta environment
    provider: okta
    enabled: true
    config:
      endpoint: https://your-domain-dev.okta.com
      token: 00XyZaBcDeFgHiJkLmNoPqRsTuVw9876543210xy
```

## Available Administrator Roles

The Okta provider supports the following built-in Okta administrator roles:

| Role ID | Role Name | Description |
|---------|-----------|-------------|
| `SUPER_ADMIN` | Super Administrator | Full administrative access to the Okta organization. Can perform all administrative tasks including managing other administrators. |
| `ORG_ADMIN` | Organization Administrator | Full administrative access except for managing super administrators. Can manage users, groups, apps, and most org settings. |
| `APP_ADMIN` | Application Administrator | Can create and manage applications and their assignments. Cannot manage users or groups unless they are assigned to apps. |
| `USER_ADMIN` | User Administrator | Can create and manage users and groups. Cannot manage applications or advanced settings. |
| `GROUP_ADMIN` | Group Administrator | Can create, manage, and delete groups. Can manage group membership. |
| `GROUP_MEMBERSHIP_ADMIN` | Group Membership Administrator | Can manage group membership but cannot create or delete groups. |
| `HELP_DESK_ADMIN` | Help Desk Administrator | Can reset passwords and MFA factors for users. Limited administrative capabilities for support purposes. |
| `READ_ONLY_ADMIN` | Read-Only Administrator | Can view all aspects of the Okta organization but cannot make changes. |
| `MOBILE_ADMIN` | Mobile Administrator | Can manage mobile device management settings and policies. |
| `API_ACCESS_MANAGEMENT_ADMIN` | API Access Management Administrator | Can manage authorization servers, scopes, and claims for API access management. |
| `REPORT_ADMIN` | Report Administrator | Can create and view reports about the Okta organization. |

## Role Configuration

Configure access to Okta administrator roles in your `config/roles/okta.yaml` file.

### Configuration Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Human-readable name for the role |
| `description` | Yes | Description of what the role provides access to |
| `workflows` | No | List of approval workflows required for this role |
| `providers` | Yes | List of Okta provider instances this role applies to |
| `enabled` | Yes | Whether this role is active |
| `inherits` | No | List of roles or groups to inherit permissions from |
| `permissions` | No | Fine-grained permission controls |
| `resources` | No | Resource-level access controls |
| `groups` | No | Group-based access controls |

### Permission Controls

Common Okta permissions include:

- `okta.users.manage` - Manage users
- `okta.users.read` - Read user information
- `okta.users.lifecycle.manage` - Manage user lifecycle (activate, deactivate, etc.)
- `okta.users.credentials.manage` - Manage user credentials
- `okta.users.credentials.resetPassword` - Reset user passwords
- `okta.users.credentials.resetFactors` - Reset MFA factors
- `okta.groups.manage` - Manage groups
- `okta.groups.read` - Read group information
- `okta.groups.members.manage` - Manage group membership
- `okta.apps.manage` - Manage applications
- `okta.apps.read` - Read application information
- `okta.policies.manage` - Manage policies
- `okta.authzServers.manage` - Manage authorization servers
- `okta.identityProviders.manage` - Manage identity providers

### Resource Patterns

Resource controls support the following patterns:

- `okta:*` - All Okta resources
- `okta:user:*` - All users
- `okta:user:john.doe@company.com` - Specific user
- `okta:group:*` - All groups
- `okta:group:Engineers` - Specific group
- `okta:app:*` - All applications
- `okta:authorizationServer:*` - All authorization servers
- `okta:role:ROLE_ID` - Specific admin role

### Example Role Configurations

#### Super Administrator

```yaml
okta_super_admin:
  name: Okta Super Administrator
  description: Full administrative access to the Okta organization
  workflows: 
    - email_approval
  permissions:
    allow:
      - okta.users.manage
      - okta.groups.manage
      - okta.apps.manage
      - okta.policies.manage
  resources:
    allow:
      - "okta:*"
  providers:
    - okta-prod
  enabled: true
```

#### User Administrator

```yaml
okta_user_admin:
  name: Okta User Administrator
  description: Can create and manage users and groups
  workflows: 
    - slack_approval
  permissions:
    allow:
      - okta.users.manage
      - okta.users.lifecycle.manage
      - okta.users.credentials.manage
      - okta.groups.manage
      - okta.policies.read
      - okta.apps.read
  resources:
    allow:
      - "okta:user:*"
      - "okta:group:*"
  providers:
    - okta-prod
    - okta-dev
  enabled: true
```

#### Help Desk Support

```yaml
okta_help_desk:
  name: Okta Help Desk Support
  description: Can reset passwords and MFA factors
  workflows: 
    - auto_approve
  permissions:
    allow:
      - okta.users.read
      - okta.users.credentials.resetPassword
      - okta.users.credentials.resetFactors
      - okta.users.lifecycle.unlock
      - okta.groups.read
  resources:
    allow:
      - "okta:user:*"
  providers:
    - okta-prod
  enabled: true
```

#### Group-Based Access

```yaml
okta_engineering_admin:
  name: Okta Engineering Group Admin
  description: Can manage the Engineering group
  workflows: 
    - slack_approval
  groups:
    allow:
      - Engineers
  permissions:
    allow:
      - okta.groups.members.manage
      - okta.users.read
  resources:
    allow:
      - "okta:group:Engineering"
  providers:
    - okta-prod
  enabled: true
```

## Features

### Administrator Role Discovery

The Okta provider automatically discovers and indexes Okta's predefined administrator roles, making them available for role elevation requests.

### Identity Synchronization

The provider synchronizes users and groups from your Okta organization, enabling:
- User discovery and search
- Group membership tracking
- Identity-based access controls

### Application Management

Tracks Okta applications and makes them available as resources for fine-grained access control.

### Permission Indexing

Includes comprehensive Okta permission mappings for:
- Permission search and discovery
- Role permission analysis
- Access recommendations

## Security Best Practices

1. **API Token Security**
   - Store API tokens in environment variables or a secrets manager
   - Never commit tokens to version control
   - Rotate tokens regularly
   - Use read-only tokens when write access is not needed

2. **Principle of Least Privilege**
   - Grant only the minimum required permissions for each role
   - Use resource restrictions to limit scope
   - Implement approval workflows for sensitive roles

3. **Monitoring**
   - Review access logs regularly
   - Monitor for unusual role assignment patterns
   - Set up alerts for high-privilege role grants

4. **Multi-Environment Setup**
   - Use separate Okta organizations for dev/staging/prod
   - Use different API tokens for each environment
   - Test role configurations in non-production environments first

## Troubleshooting

### Common Issues

**Issue**: `endpoint is required for Okta provider`
- **Solution**: Ensure the `endpoint` field is set in your provider configuration

**Issue**: `token is required for Okta provider`
- **Solution**: Ensure the `token` field is set in your provider configuration

**Issue**: `failed to create Okta client`
- **Solution**: Verify your `endpoint` is correct and includes the full URL (e.g., `https://your-domain.okta.com`)

**Issue**: API token authentication failures
- **Solution**: Verify your API token is valid and hasn't been revoked. Generate a new token if needed.

**Issue**: Role not found errors
- **Solution**: Ensure you're using the correct role ID from the Available Administrator Roles table above

### Debugging

Enable debug logging to troubleshoot Okta provider issues:

```yaml
logging:
  level: debug
```

Look for Okta-specific log entries to identify authentication and permission issues.

## References

- [Okta Administrator Roles Documentation](https://help.okta.com/en-us/content/topics/security/administrators-admin-comparison.htm)
- [Okta API Authentication Guide](https://developer.okta.com/docs/guides/create-an-api-token/)
- [Okta SDK for Go](https://github.com/okta/okta-sdk-golang)
