---
layout: default
title: AWS Provider
description: Amazon Web Services provider with IAM and SSO integration
parent: Providers
grand_parent: Configuration
---

# AWS Provider

The AWS provider enables integration with Amazon Web Services, providing role-based access control (RBAC) capabilities through AWS IAM and AWS SSO (Identity Center).

## Capabilities

- **Role-Based Access Control (RBAC)**: Supports AWS IAM roles and AWS SSO permission sets
- **Permission Management**: Access to AWS IAM permissions and policies
- **Identity Integration**: Support for AWS SSO Identity Center users and groups
- **Multi-Account Support**: Can be configured for different AWS accounts and regions

## Prerequisites

### AWS Account Setup

1. **AWS Account**: Active AWS account with appropriate permissions
2. **IAM Permissions**: The agent needs permissions to:
   - List and describe IAM roles and policies
   - Access STS for account information
   - Read AWS SSO configurations (if using SSO)

### Required AWS Permissions

The following AWS permissions are required for the agent to function properly:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iam:ListRoles",
        "iam:ListPolicies",
        "iam:GetRole",
        "iam:GetPolicy",
        "iam:ListAttachedRolePolicies",
        "sts:GetCallerIdentity",
        "sso:ListPermissionSets",
        "sso:DescribePermissionSet",
        "identitystore:ListUsers",
        "identitystore:ListGroups"
      ],
      "Resource": "*"
    }
  ]
}
```

## Authentication Methods

The AWS provider supports multiple authentication methods:

### 1. AWS Profile (Recommended)

Uses AWS shared credentials profile:

```yaml
providers:
  aws-prod:
    name: AWS Production
    provider: aws
    config:
      region: us-east-1
      profile: my-aws-profile
```

### 2. Static Credentials

Uses explicit access key and secret:

```yaml
providers:
  aws-prod:
    name: AWS Production
    provider: aws
    config:
      region: us-east-1
      access_key_id: YOUR_AWS_ACCESS_KEY_ID
      secret_access_key: YOUR_AWS_SECRET_ACCESS_KEY
```

### 3. IAM Role (Default)

When no credentials are provided, uses the default AWS credential chain (environment variables, EC2 instance profile, etc.):

```yaml
providers:
  aws-prod:
    name: AWS Production
    provider: aws
    config:
      region: us-east-1
```

## Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `region` | string | No | `us-east-1` | AWS region to use |
| `profile` | string | No | - | AWS shared credentials profile name |
| `access_key_id` | string | No | - | AWS access key ID (requires secret_access_key) |
| `secret_access_key` | string | No | - | AWS secret access key (requires access_key_id) |
| `account_id` | string | No | - | AWS account ID (auto-detected if not provided) |

## Getting Credentials

### AWS CLI Setup

1. **Install AWS CLI**: Follow the [AWS CLI installation guide](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)

2. **Configure Profile**:
   ```bash
   aws configure --profile my-aws-profile
   ```

3. **Set Profile in Configuration**:
   ```yaml
   config:
     profile: my-aws-profile
   ```

### IAM User Setup

1. **Create IAM User**: In AWS Console → IAM → Users → Create User
2. **Attach Policy**: Attach the required permissions policy (see above)
3. **Generate Access Keys**: Create access key and secret access key
4. **Configure in Agent**: Use the access key and secret in your configuration

### IAM Role Setup (EC2/ECS/Lambda)

1. **Create IAM Role**: In AWS Console → IAM → Roles → Create Role
2. **Attach Policy**: Attach the required permissions policy
3. **Associate Role**: Attach to EC2 instance, ECS task, or Lambda function
4. **No Configuration Needed**: Agent will automatically use the role

## Example Configurations

### Production Environment with Profile

```yaml
version: "1.0"
providers:
  aws-prod:
    name: AWS Production
    description: Production AWS environment
    provider: aws
    enabled: true
    config:
      region: us-east-1
      profile: prod-aws-profile
```

### Development Environment with Static Credentials

```yaml
version: "1.0"
providers:
  aws-dev:
    name: AWS Development
    description: Development AWS environment
    provider: aws
    enabled: true
    config:
      region: us-west-2
      access_key_id: YOUR_AWS_ACCESS_KEY_ID
      secret_access_key: YOUR_AWS_SECRET_ACCESS_KEY
```

### Multi-Account Setup

```yaml
version: "1.0"
providers:
  aws-prod:
    name: AWS Production
    description: Production account
    provider: aws
    enabled: true
    config:
      region: us-east-1
      profile: prod-profile
      account_id: "YOUR_PROD_ACCOUNT_ID"
  
  aws-staging:
    name: AWS Staging
    description: Staging account
    provider: aws
    enabled: true
    config:
      region: us-east-1
      profile: staging-profile
      account_id: "YOUR_STAGING_ACCOUNT_ID"
```

## Features

### IAM Role Discovery

The AWS provider automatically discovers and indexes IAM roles in your account, making them available for role elevation requests.

### AWS SSO Integration

When AWS SSO is configured, the provider can integrate with Identity Center to provide:
- User and group discovery
- Permission set management
- Federated identity support

### Permission Indexing

The provider includes a comprehensive database of AWS IAM permissions, enabling:
- Permission search and discovery
- Policy analysis and recommendations
- Role permission mapping

## Troubleshooting

### Common Issues

1. **Authentication Failures**
   - Verify AWS credentials are correctly configured
   - Check IAM permissions for the agent
   - Ensure account ID is valid (12 digits)

2. **Region Issues**
   - Verify the specified region exists and is accessible
   - Check if IAM roles exist in the specified region

3. **SSO Integration**
   - Ensure AWS SSO is enabled in the account
   - Verify SSO permissions for the agent role

### Debugging

Enable debug logging to troubleshoot AWS provider issues:

```yaml
logging:
  level: debug
```

Look for AWS-specific log entries to identify authentication and permission issues.
