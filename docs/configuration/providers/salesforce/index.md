---
layout: default
title: Salesforce Provider
description: Salesforce CRM provider for user and permission management
parent: Providers
grand_parent: Configuration
---

# Salesforce Provider

The Salesforce provider enables integration with Salesforce CRM, providing role-based access control and user management capabilities through the Salesforce API.

## Capabilities

- **Authentication**: Username/password-based authentication with Salesforce
- **Role-Based Access Control (RBAC)**: Salesforce profile and permission set management
- **User Management**: Access to Salesforce users and organizational data
- **Permission Integration**: Support for Salesforce profiles, permission sets, and roles

## Prerequisites

### Salesforce Org Setup

1. **Salesforce Org**: Active Salesforce organization (Production, Sandbox, or Developer)
2. **API Access**: API access enabled in your Salesforce org
3. **User Credentials**: Salesforce user with API access and appropriate permissions
4. **Security Token**: User security token for API authentication

### Required Salesforce Permissions

The Salesforce user should have the following permissions:
- **API Enabled**: User must have API access
- **View All Users**: To read user information
- **Modify All Data**: For full RBAC capabilities (or specific object permissions)
- **Manage Profiles and Permission Sets**: To manage user permissions

## Authentication Methods

### 1. Username/Password with Security Token

Uses Salesforce username, password, and security token:

```yaml
providers:
  salesforce:
    name: Salesforce
    provider: salesforce
    config:
      domain: login.salesforce.com
      username: YOUR_SALESFORCE_USERNAME
      password: YOUR_SALESFORCE_PASSWORD
      security_token: YOUR_SECURITY_TOKEN
```

### 2. Sandbox Environment

For Salesforce sandbox organizations:

```yaml
providers:
  salesforce-sandbox:
    name: Salesforce Sandbox
    provider: salesforce
    config:
      domain: test.salesforce.com
      username: YOUR_SANDBOX_USERNAME
      password: YOUR_SANDBOX_PASSWORD
      security_token: YOUR_SANDBOX_SECURITY_TOKEN
```

### 3. Custom Domain

For organizations with custom domains:

```yaml
providers:
  salesforce-custom:
    name: Salesforce Custom
    provider: salesforce
    config:
      domain: YOUR_CUSTOM_DOMAIN.my.salesforce.com
      username: YOUR_USERNAME
      password: YOUR_PASSWORD
      security_token: YOUR_SECURITY_TOKEN
```

## Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `domain` | string | Yes | - | Salesforce login domain |
| `username` | string | Yes | - | Salesforce username |
| `password` | string | Yes | - | Salesforce password |
| `security_token` | string | Yes | - | Salesforce security token |
| `version` | string | No | `latest` | Salesforce API version |

## Getting Credentials

### Security Token Setup

1. **Login to Salesforce**: Go to your Salesforce org

2. **Reset Security Token**:
   - Click your profile → Settings
   - In the left sidebar, click "Reset My Security Token"
   - Click "Reset Security Token"
   - Check your email for the new security token

3. **API Access**: Ensure your user profile has "API Enabled" permission

### Connected App Setup (Advanced)

For enhanced security, you can create a Connected App:

1. **Create Connected App**: Setup → App Manager → New Connected App

2. **Configure OAuth Settings**:
   - Enable OAuth Settings
   - Add required OAuth scopes
   - Set callback URL if needed

3. **Get Consumer Key/Secret**: Use for OAuth-based authentication

### Trusted IP Ranges

To avoid security token requirements:

1. **Setup Trusted IP Ranges**: Setup → Network Access → Trusted IP Ranges
2. **Add Agent IP**: Add your agent's IP address range
3. **Remove Security Token**: Security token not required from trusted IPs

## Example Configurations

### Production Salesforce

```yaml
version: "1.0"
providers:
  salesforce-prod:
    name: Salesforce Production
    description: Production Salesforce org
    provider: salesforce
    enabled: true
    config:
      domain: login.salesforce.com
      username: YOUR_PROD_USERNAME
      password: YOUR_PROD_PASSWORD
      security_token: YOUR_PROD_SECURITY_TOKEN
```

### Sandbox Environment

```yaml
version: "1.0"
providers:
  salesforce-sandbox:
    name: Salesforce Sandbox
    description: Sandbox Salesforce org
    provider: salesforce
    enabled: true
    config:
      domain: test.salesforce.com
      username: YOUR_SANDBOX_USERNAME
      password: YOUR_SANDBOX_PASSWORD
      security_token: YOUR_SANDBOX_SECURITY_TOKEN
```

### Custom Domain

```yaml
version: "1.0"
providers:
  salesforce-custom:
    name: Salesforce Custom Domain
    description: Custom domain Salesforce org
    provider: salesforce
    enabled: true
    config:
      domain: mycompany.my.salesforce.com
      username: YOUR_USERNAME
      password: YOUR_PASSWORD
      security_token: YOUR_SECURITY_TOKEN
```

### Multi-Environment Setup

```yaml
version: "1.0"
providers:
  salesforce-prod:
    name: Salesforce Production
    description: Production environment
    provider: salesforce
    enabled: true
    config:
      domain: login.salesforce.com
      username: YOUR_PROD_USERNAME
      password: YOUR_PROD_PASSWORD
      security_token: YOUR_PROD_TOKEN
  
  salesforce-staging:
    name: Salesforce Staging
    description: Staging environment
    provider: salesforce
    enabled: true
    config:
      domain: test.salesforce.com
      username: YOUR_STAGING_USERNAME
      password: YOUR_STAGING_PASSWORD
      security_token: YOUR_STAGING_TOKEN
```

## Features

### Profile and Permission Set Management

The Salesforce provider can manage:
- User profiles and their permissions
- Permission sets and assignments
- Role hierarchies
- Organization-wide defaults

### User and Role Discovery

Access to Salesforce organizational data:
- User accounts and their roles
- Profile assignments
- Permission set assignments
- Role hierarchy structure

### API Integration

Full integration with Salesforce APIs:
- REST API for modern integrations
- SOAP API for comprehensive access
- Metadata API for configuration management

## Troubleshooting

### Common Issues

1. **Authentication Failures**
   - Verify username and password are correct
   - Check if security token is current and correct
   - Ensure API access is enabled for the user

2. **Security Token Issues**
   - Reset security token if login fails
   - Check if IP address is in trusted ranges
   - Verify email delivery of security token

3. **Permission Issues**
   - Verify user has "API Enabled" permission
   - Check profile permissions for required objects
   - Ensure user has appropriate role in role hierarchy

4. **Domain Issues**
   - Verify correct login domain (login.salesforce.com vs test.salesforce.com)
   - Check custom domain configuration
   - Ensure domain is accessible from agent location

### Debugging

Enable debug logging to troubleshoot Salesforce provider issues:

```yaml
logging:
  level: debug
```

Look for Salesforce API-specific log entries to identify authentication and permission issues.

### API Limits

Salesforce has API call limits:
- **Developer Edition**: 5,000 calls per day
- **Professional/Enterprise**: 1,000 calls per user per day
- **Unlimited**: 5,000 calls per user per day

### Environment Variables

The Salesforce provider supports these environment variables:
- `SALESFORCE_USERNAME`: Salesforce username
- `SALESFORCE_PASSWORD`: Salesforce password
- `SALESFORCE_SECURITY_TOKEN`: Security token
- `SALESFORCE_DOMAIN`: Login domain

## Security Considerations

1. **Credential Security**: Store passwords and security tokens securely
2. **Least Privilege**: Use user accounts with minimal required permissions
3. **IP Restrictions**: Configure trusted IP ranges when possible
4. **Token Rotation**: Regularly reset security tokens
5. **Audit Logging**: Monitor API usage and access patterns
