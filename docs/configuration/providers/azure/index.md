---
layout: default
title: Azure
description: Microsoft Azure provider with RBAC and subscription management
parent: Providers
grand_parent: Configuration
---

# Azure Provider

The Azure provider enables integration with Microsoft Azure, providing role-based access control (RBAC) capabilities through Azure Active Directory and Azure Resource Manager.

## Capabilities

- **Role-Based Access Control (RBAC)**: Supports Azure RBAC roles and assignments
- **Subscription Management**: Access to Azure subscriptions and resource groups
- **Permission Management**: Access to Azure resource provider operations and permissions
- **Identity Integration**: Support for Azure AD authentication and service principals

## Prerequisites

### Azure Account Setup

1. **Azure Subscription**: Active Azure subscription with appropriate permissions
2. **Azure AD Permissions**: The agent needs permissions to:
   - Read role assignments and definitions
   - Access subscription and resource group information
   - List Azure resource provider operations

### Required Azure Permissions

The following Azure permissions are required for the agent to function properly:

```json
{
  "permissions": [
    {
      "actions": [
        "Microsoft.Authorization/roleAssignments/read",
        "Microsoft.Authorization/roleDefinitions/read",
        "Microsoft.Resources/subscriptions/read",
        "Microsoft.Resources/subscriptions/resourceGroups/read"
      ],
      "notActions": [],
      "dataActions": [],
      "notDataActions": []
    }
  ]
}
```

## Authentication Methods

The Azure provider supports multiple authentication methods:

### 1. Service Principal (Recommended)

Uses Azure service principal with client credentials:

```yaml
providers:
  azure-prod:
    name: Azure Production
    provider: azure
    config:
      subscription_id: YOUR_SUBSCRIPTION_ID
      tenant_id: YOUR_TENANT_ID
      client_id: YOUR_CLIENT_ID
      client_secret: YOUR_CLIENT_SECRET
```

### 2. Managed Identity (Default)

When no credentials are provided, uses the default Azure credential chain (managed identity, environment variables, etc.):

```yaml
providers:
  azure-prod:
    name: Azure Production
    provider: azure
    config:
      subscription_id: YOUR_SUBSCRIPTION_ID
```

### 3. Resource Group Scoped

Optionally scope to a specific resource group:

```yaml
providers:
  azure-prod:
    name: Azure Production
    provider: azure
    config:
      subscription_id: YOUR_SUBSCRIPTION_ID
      resource_group: YOUR_RESOURCE_GROUP_NAME
      tenant_id: YOUR_TENANT_ID
      client_id: YOUR_CLIENT_ID
      client_secret: YOUR_CLIENT_SECRET
```

## Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `subscription_id` | string | Yes | - | Azure subscription ID |
| `tenant_id` | string | No | - | Azure AD tenant ID (required for service principal) |
| `client_id` | string | No | - | Service principal client ID |
| `client_secret` | string | No | - | Service principal client secret |
| `resource_group` | string | No | - | Resource group name (optional scoping) |

## Getting Credentials

### Service Principal Setup

1. **Create Service Principal**: In Azure Portal → Azure Active Directory → App Registrations → New Registration

2. **Get Application Details**:
   - **Application (client) ID**: Copy this value for `client_id`
   - **Directory (tenant) ID**: Copy this value for `tenant_id`

3. **Create Client Secret**: In the app registration → Certificates & secrets → New client secret
   - Copy the secret value for `client_secret`

4. **Assign Permissions**: In Azure Portal → Subscriptions → Your Subscription → Access Control (IAM)
   - Add role assignment with "Reader" role minimum
   - Assign to your service principal

### Azure CLI Setup

1. **Install Azure CLI**: Follow the [Azure CLI installation guide](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli)

2. **Login**:
   ```bash
   az login
   ```

3. **Get Subscription ID**:
   ```bash
   az account show --query id --output tsv
   ```

### Managed Identity Setup (Azure Resources)

1. **Enable Managed Identity**: On your Azure resource (VM, App Service, etc.)
2. **Assign Permissions**: Grant the managed identity appropriate RBAC permissions
3. **No Configuration Needed**: Agent will automatically use the managed identity

## Example Configurations

### Production Environment with Service Principal

```yaml
version: "1.0"
providers:
  azure-prod:
    name: Azure Production
    description: Production Azure environment
    provider: azure
    enabled: true
    config:
      subscription_id: YOUR_SUBSCRIPTION_ID
      tenant_id: YOUR_TENANT_ID
      client_id: YOUR_CLIENT_ID
      client_secret: YOUR_CLIENT_SECRET
```

### Development Environment with Managed Identity

```yaml
version: "1.0"
providers:
  azure-dev:
    name: Azure Development
    description: Development Azure environment
    provider: azure
    enabled: true
    config:
      subscription_id: YOUR_SUBSCRIPTION_ID
```

### Resource Group Scoped Configuration

```yaml
version: "1.0"
providers:
  azure-rg:
    name: Azure Resource Group
    description: Scoped to specific resource group
    provider: azure
    enabled: true
    config:
      subscription_id: YOUR_SUBSCRIPTION_ID
      resource_group: YOUR_RESOURCE_GROUP_NAME
      tenant_id: YOUR_TENANT_ID
      client_id: YOUR_CLIENT_ID
      client_secret: YOUR_CLIENT_SECRET
```

### Multi-Subscription Setup

```yaml
version: "1.0"
providers:
  azure-prod:
    name: Azure Production
    description: Production subscription
    provider: azure
    enabled: true
    config:
      subscription_id: YOUR_PROD_SUBSCRIPTION_ID
      tenant_id: YOUR_TENANT_ID
      client_id: YOUR_CLIENT_ID
      client_secret: YOUR_CLIENT_SECRET
  
  azure-staging:
    name: Azure Staging
    description: Staging subscription
    provider: azure
    enabled: true
    config:
      subscription_id: YOUR_STAGING_SUBSCRIPTION_ID
      tenant_id: YOUR_TENANT_ID
      client_id: YOUR_CLIENT_ID
      client_secret: YOUR_CLIENT_SECRET
```

## Features

### Azure RBAC Integration

The Azure provider automatically discovers and indexes Azure built-in and custom roles, making them available for role elevation requests.

### Resource Provider Operations

Access to comprehensive Azure resource provider operations and permissions for fine-grained access control.

### Subscription and Resource Group Management

Support for managing access across multiple Azure subscriptions and resource groups.

## Troubleshooting

### Common Issues

1. **Authentication Failures**
   - Verify service principal credentials are correct
   - Check Azure AD permissions for the service principal
   - Ensure subscription ID is valid and accessible

2. **Permission Issues**
   - Verify the service principal has Reader role on the subscription
   - Check if resource group exists and is accessible
   - Ensure tenant ID matches the subscription's tenant

3. **Managed Identity Issues**
   - Verify managed identity is enabled on the Azure resource
   - Check if managed identity has appropriate RBAC permissions
   - Ensure the Azure resource can access Azure metadata service

### Debugging

Enable debug logging to troubleshoot Azure provider issues:

```yaml
logging:
  level: debug
```

Look for Azure-specific log entries to identify authentication and permission issues.

### Environment Variables

The Azure provider also supports standard Azure environment variables:

- `AZURE_CLIENT_ID`: Service principal client ID
- `AZURE_CLIENT_SECRET`: Service principal client secret
- `AZURE_TENANT_ID`: Azure AD tenant ID
- `AZURE_SUBSCRIPTION_ID`: Azure subscription ID
