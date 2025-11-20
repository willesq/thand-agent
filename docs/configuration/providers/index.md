---
layout: default
title: Providers
parent: Configuration
nav_order: 8
description: Detailed documentation for Thand Agent providers
has_children: true
---

# Providers

Providers are the core integration components of the Thand Agent that enable connectivity with various external services, identity providers, and cloud platforms. Each provider implements specific capabilities to support authentication, authorization, role-based access control, and notifications.

## Provider Capabilities

The Thand Agent supports four main provider capabilities:

### 1. Authentication (Authorizor)
**Capability**: `authorizor`

Enables user authentication and session management. Providers with this capability can:
- Authenticate users against external identity systems
- Create and manage user sessions
- Handle OAuth2, SAML, and other authentication flows
- Validate and renew authentication tokens

### 2. Role-Based Access Control (RBAC)
**Capability**: `rbac`

Provides role and permission management capabilities. Providers with this capability can:
- Grant temporary access to roles and permissions
- Manage user entitlements and access levels
- Integrate with cloud IAM systems (AWS, Azure, GCP)
- Handle role elevation and access requests

### 3. Notifications
**Capability**: `notifier`

Enables sending notifications and messages. Providers with this capability can:
- Send notifications via email, Slack, or other channels
- Deliver access request notifications
- Provide audit and alert messaging
- Support custom notification templates

### 4. Identity Management
**Capability**: `identities`

Provides user and group discovery capabilities. Providers with this capability can:
- Discover users and groups from directory services
- Sync identity information from external systems
- Provide user profile and membership data
- Support identity federation scenarios

## Available Providers

### Cloud & Infrastructure Providers

| Provider | Capabilities | Description |
|----------|-------------|-------------|
| [AWS](aws/) | RBAC | Amazon Web Services IAM and SSO integration |
| [Azure](azure/) | RBAC | Microsoft Azure RBAC and subscription management |
| [GCP](gcp/) | RBAC | Google Cloud Platform IAM and resource management |
| [Cloudflare](cloudflare/) | RBAC, Identities | Cloudflare account management with role and policy-based access control |
| [Kubernetes](kubernetes/) | Authorizor, RBAC | Kubernetes cluster authentication and RBAC |

### Development & DevOps

| Provider | Capabilities | Description |
|----------|-------------|-------------|
| [GitHub](github/) | Authorizor, RBAC | GitHub repository and organization management |
| [Terraform](terraform/) | Authorizor, RBAC | Terraform Cloud/Enterprise workspace management |

### Enterprise Authentication

| Provider | Capabilities | Description |
|----------|-------------|-------------|
| [SAML](saml/) | Authorizor | SAML 2.0 SSO integration for enterprise identity providers |
| [OAuth2](oauth2/) | Authorizor | Generic OAuth2 authentication for any compliant service |
| [Google OAuth2](oauth2.google/) | Authorizor | Google account authentication with OAuth2 |
| [Thand](thand/) | Authorizor | Thand federated OIDC authentication service |
| [Okta](okta/) | RBAC, Identities | Okta identity management and administrator role control |

### Business Applications

| Provider | Capabilities | Description |
|----------|-------------|-------------|
| [Salesforce](salesforce/) | RBAC, Authorizor | Salesforce CRM user and permission management |
| [Google Workspace](gsuite/) | Authorizor, Identities | Google Workspace user and group management |

### Infrastructure & Communication

| Provider | Capabilities | Description |
|----------|-------------|-------------|
| [Slack](slack/) | Notifier | Slack team communication and notifications |
| [Email](email/) | Notifier | SMTP email notifications and communication |

## Provider Configuration

All providers follow a common configuration structure:

```yaml
version: "1.0"
providers:
  provider-name:
    name: Display Name
    description: Provider description
    provider: provider-type
    enabled: true
    config:
      # Provider-specific configuration options
```

### Common Configuration Options

- **name**: Human-readable display name for the provider
- **description**: Brief description of the provider's purpose
- **provider**: The provider type (e.g., `aws`, `azure`, `github`)
- **enabled**: Whether the provider is active
- **config**: Provider-specific configuration parameters

## Getting Started

1. **Choose Your Providers**: Select providers based on your infrastructure and requirements
2. **Review Prerequisites**: Check each provider's documentation for setup requirements
3. **Obtain Credentials**: Follow the credential setup guides for each provider
4. **Configure**: Add provider configurations to your agent configuration file
5. **Test**: Verify provider connectivity and functionality

## Security Best Practices

- **Use Placeholders**: Never commit actual credentials to version control
- **Least Privilege**: Grant minimal required permissions to provider credentials
- **Rotate Credentials**: Regularly rotate API keys, tokens, and certificates
- **Secure Storage**: Use secure credential storage solutions (vaults, secrets managers)
- **Monitor Access**: Enable audit logging and monitor provider access patterns

## Troubleshooting

For provider-specific troubleshooting:
1. Check the individual provider documentation
2. Enable debug logging in your agent configuration
3. Verify credentials and permissions
4. Check network connectivity and firewall settings
5. Review provider service status and API limits
