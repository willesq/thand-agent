---
layout: default
title: Roles
parent: Configuration
nav_order: 9
description: Detailed documentation for Thand Agent roles
has_children: true
---

# Roles

Roles are the core authorization mechanism in Thand Agent that define what permissions users can request and under what conditions. They act as templates that specify the scope of access, workflows for approval, and inheritance relationships that enable flexible permission management.

## Role Concepts

### What is a Role?

A Thand role is a configuration that defines:
- **Permissions**: What actions can be performed (allow/deny rules)
- **Resources**: Which resources can be accessed (with allow/deny rules)
- **Inheritance**: Which other roles this role builds upon
- **Providers**: Which provider instances can be used with this role
- **Scopes**: Who can request this role (users/groups)
- **Workflows**: How access requests are processed and approved

Roles are **templates** that users can request to be granted temporarily. When a user requests a role, the agent creates a temporary authorization based on the role's definition.

### Role vs Provider Roles

It's important to distinguish between:
- **Thand Roles**: Defined in your agent configuration (documented here)
- **Provider Roles**: Native roles in external systems (AWS IAM roles, Azure roles, etc.)

Thand roles can **inherit** from other roles and provider roles to leverage existing cloud IAM configurations.

## Role Structure

{: .note}
When mixing multiple providers into a single role you can use the providers name as a prefix to avoid ambiguity. For example, to inherit from an AWS role in the `aws-prod` provider instance, use `aws-prod:arn:aws:iam::aws:policy/ReadOnlyAccess`.

### Basic Configuration

```yaml
version: "1.0"
roles:
  role-name:
    name: Human Readable Name
    description: Description of what this role provides
    enabled: true
    
    # Core role definition
    permissions: # What actions are allowed/denied
    resources:   # What resources can be accessed
    inherits:    # What other roles to inherit from
    providers:   # Which providers can be used
    
    # Access control
    scopes:      # Who can request this role
    
    # Process control  
    workflows:   # How requests are processed
    authenticators: # Which auth providers are valid
```

### Complete Role Example

```yaml
version: "1.0"
roles:
  aws-developer:
    name: AWS Developer Access
    description: Developer access to AWS resources with approval workflow
    enabled: true
    
    # Inheritance - build upon existing roles
    inherits:
      - aws-basic-user                    # Local role
      - aws-dev:arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess  # AWS managed policy
    
    # Explicit permissions
    permissions:
      allow:
        - ec2:*Instance*
        - s3:GetObject
        - s3:PutObject
      deny:
        - ec2:TerminateInstances
    
    # Resource restrictions
    resources:
      allow:
        - "arn:aws:ec2:us-east-1:123456789012:instance/*"
        - "arn:aws:s3:::dev-bucket/*"
      deny:
        - "arn:aws:s3:::prod-bucket/*"
    
    # Provider restrictions
    providers:
      - aws-dev
      - aws-staging
    
    # Who can request this role
    scopes:
      users:
        - developer@company.com
      groups:
        - oidc:developers
        - ad:dev-team
    
    # Approval process
    workflows:
      - manager-approval
      - security-review
    
    # Valid authentication methods
    authenticators:
      - google-oauth
      - saml-sso
```

## Role Inheritance

Role inheritance is a powerful feature that allows roles to build upon each other, promoting reusability and consistent security patterns.

### How Inheritance Works

When a role inherits from other roles:

1. **Permission Merging**: All `allow` and `deny` permissions from inherited roles are combined
2. **Resource Merging**: All resource `allow` and `deny` rules are combined  
3. **Additive Model**: Child roles can add more permissions but cannot remove inherited ones
4. **Override Protection**: Inherited permissions cannot be overridden, only supplemented

### Inheritance Types

#### 1. Local Role Inheritance

Inherit from other Thand roles:

```yaml
roles:
  basic-user:
    name: Basic User
    permissions:
      allow:
        - ec2:DescribeInstances
        - s3:ListBuckets
  
  power-user:
    name: Power User
    inherits:
      - basic-user  # Inherits basic-user permissions
    permissions:
      allow:
        - ec2:StartInstances  # Additional permissions
        - ec2:StopInstances
```

#### 2. Provider Role Inheritance

Inherit from cloud provider managed roles:

```yaml
roles:
  aws-admin:
    name: AWS Administrator
    inherits:
      - arn:aws:iam::aws:policy/AdministratorAccess
    # No additional permissions needed - inherits full admin access
  
  gcp-viewer:
    name: GCP Viewer
    inherits:
      - roles/viewer  # GCP predefined role
    permissions:
      allow:
        - compute.instances.start  # Additional specific permission
```

#### 3. Mixed Inheritance

Combine local and provider roles:

```yaml
roles:
  hybrid-admin:
    name: Hybrid Cloud Admin
    inherits:
      - local-base-role
      - arn:aws:iam::aws:policy/ReadOnlyAccess
      - roles/compute.viewer
    permissions:
      allow:
        - custom:specific-action
```

### Inheritance Resolution

The inheritance system resolves permissions in this order:

1. **Direct Permissions**: Permissions defined directly in the role
2. **Inherited Permissions**: Permissions from all inherited roles (merged)
3. **Validation**: Ensures no conflicts and all inherited roles exist

Example of inheritance resolution:

```yaml
# Base role
base-role:
  permissions:
    allow: [ec2:Describe*]
    deny: [ec2:Terminate*]

# Child role  
child-role:
  inherits: [base-role]
  permissions:
    allow: [s3:GetObject]
    deny: [s3:DeleteObject]

# Resolved permissions for child-role:
# allow: [ec2:Describe*, s3:GetObject]
# deny: [ec2:Terminate*, s3:DeleteObject]
```

## Role Configuration Fields

### Core Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Human-readable role name |
| `description` | string | Yes | Description of role purpose |
| `enabled` | boolean | No | Whether role is active (default: true) |

### Permission Fields

| Field | Type | Description |
|-------|------|-------------|
| `permissions.allow` | array | List of allowed permissions/actions |
| `permissions.deny` | array | List of explicitly denied permissions |
| `resources.allow` | array | List of allowed resources (ARNs, URLs, etc.) |
| `resources.deny` | array | List of denied resources |
| `inherits` | array | List of roles to inherit from |

### Access Control Fields

| Field | Type | Description |
|-------|------|-------------|
| `providers` | array | List of provider instances this role can use |
| `scopes.users` | array | Specific users who can request this role |
| `scopes.groups` | array | Groups whose members can request this role |
| `authenticators` | array | Valid authentication providers for this role |

### Process Control Fields

| Field | Type | Description |
|-------|------|-------------|
| `workflows` | array | Approval workflows to execute for this role |

## Permission Patterns

### Cloud Provider Patterns

#### AWS Permissions
```yaml
permissions:
  allow:
    - ec2:*                    # All EC2 actions
    - s3:GetObject             # Specific S3 action
    - iam:PassRole             # IAM role assumption
  deny:
    - ec2:TerminateInstances   # Explicit denial
```

#### Azure Permissions
```yaml
permissions:
  allow:
    - Microsoft.Compute/virtualMachines/*
    - Microsoft.Storage/storageAccounts/read
    - Microsoft.Authorization/roleAssignments/read
  deny:
    - Microsoft.Compute/virtualMachines/delete
```

#### GCP Permissions
```yaml
permissions:
  allow:
    - compute.instances.*
    - storage.buckets.list
    - iam.serviceAccounts.get
  deny:
    - compute.instances.delete
```

### Resource Patterns

#### AWS Resources
```yaml
resources:
  allow:
    - "arn:aws:ec2:*:*:instance/*"           # All EC2 instances
    - "arn:aws:s3:::my-bucket/*"             # Specific S3 bucket
    - "arn:aws:iam::123456789012:role/*"     # IAM roles in account
  deny:
    - "arn:aws:s3:::sensitive-bucket/*"      # Sensitive data
```

#### Azure Resources
```yaml
resources:
  allow:
    - "/subscriptions/*/resourceGroups/dev-*" # Dev resource groups
    - "/subscriptions/12345/resourceGroups/prod/providers/Microsoft.Compute/virtualMachines/*"
  deny:
    - "/subscriptions/*/resourceGroups/prod-*" # Production resources
```

#### GCP Resources
```yaml
resources:
  allow:
    - "projects/my-dev-project/zones/*/instances/*"
    - "projects/*/global/networks/default"
  deny:
    - "projects/prod-project/*"
```

## Scope Management

Scopes control **who** can request a role. This enables role-based access control at the user/group level.

### User Scopes

Grant access to specific users:

```yaml
scopes:
  users:
    - alice@company.com
    - bob@company.com
    - service-account@company.com
```

### Group Scopes

Grant access to groups (depends on identity provider):

```yaml
scopes:
  groups:
    - oidc:developers      # OIDC group
    - ad:engineering       # Active Directory group  
    - saml:admins         # SAML group
    - github:my-org/team  # GitHub team
```

### Mixed Scopes

Combine users and groups:

```yaml
scopes:
  users:
    - emergency@company.com
  groups:
    - oidc:on-call
    - ad:security-team
```

### Public Roles

Omit `scopes` to allow any authenticated user to request the role:

```yaml
roles:
  basic-viewer:
    name: Basic Viewer
    # No 'scopes' field - available to all users
    permissions:
      allow:
        - ec2:DescribeInstances
```

## Workflow Integration

Roles integrate with [workflows](../workflows/) to define approval processes.

### Workflow Assignment

```yaml
roles:
  sensitive-admin:
    name: Sensitive Admin Access
    workflows:
      - manager-approval     # Requires manager approval
      - security-review      # Additional security review
    permissions:
      allow:
        - "*:*"
```

### Workflow Conditions

Workflows can implement conditional logic based on:
- Requested role
- User identity
- Time of request
- Duration requested
- Resource scope

## Provider Integration

Roles specify which provider instances can be used for role elevation.

### Single Provider

```yaml
roles:
  aws-dev-access:
    name: AWS Development Access
    providers:
      - aws-dev  # Only the aws-dev provider instance
    permissions:
      allow:
        - ec2:*
```

### Multi-Provider

```yaml
roles:
  multi-cloud-viewer:
    name: Multi-Cloud Viewer
    providers:
      - aws-prod
      - azure-prod  
      - gcp-prod
    permissions:
      allow:
        - "*:Describe*"
        - "*:List*"
        - "*:Get*"
```

### Provider Inheritance

When inheriting from provider roles, ensure the provider supports the inherited role:

```yaml
roles:
  aws-ec2-admin:
    name: EC2 Administrator
    providers:
      - aws-prod
    inherits:
      - arn:aws:iam::aws:policy/AmazonEC2FullAccess  # Must exist in aws-prod
```

## Configuration Management

### File Structure

Roles can be organized in multiple ways:

#### Single File
```yaml
# roles.yaml
version: "1.0"
roles:
  role1: { ... }
  role2: { ... }
```

#### Multiple Files by Provider
```
config/roles/
├── aws.yaml
├── azure.yaml
├── gcp.yaml
└── common.yaml
```

#### Multiple Files by Team
```
config/roles/
├── developers.yaml
├── admins.yaml
├── security.yaml
└── readonly.yaml
```

### Loading Configuration

Configure role loading in the main config:

```yaml
# Load from directory
roles:
  path: "./config/roles"

# Load from URL
roles:
  url:
    uri: "https://config.company.com/roles.yaml"
    headers:
      Authorization: "Bearer TOKEN"

# Load from Vault
roles:
  vault: "secret/agent/roles"

# Inline definitions
roles:
  admin:
    name: Admin
    permissions:
      allow: ["*:*"]
```

## Best Practices

### 1. Role Design Principles

**Principle of Least Privilege**
```yaml
# Good - specific permissions
permissions:
  allow:
    - ec2:StartInstances
    - ec2:StopInstances
    - ec2:DescribeInstances

# Avoid - overly broad permissions  
permissions:
  allow:
    - ec2:*
```

**Time-Bounded Access**
```yaml
# Roles should be used with time limits
# Configure in workflows, not roles themselves
workflows:
  - time-limited-approval  # Implements 2-hour max
```

**Clear Naming**
```yaml
# Good - descriptive names
aws-ec2-restart-access:
  name: AWS EC2 Instance Restart Access

# Avoid - unclear names
role1:
  name: Some Access
```

### 2. Inheritance Patterns

**Build Role Hierarchies**
```yaml
# Base roles
readonly-base:
  permissions:
    allow: ["*:Describe*", "*:List*", "*:Get*"]

# Specialized roles
ec2-readonly:
  inherits: [readonly-base]
  permissions:
    allow: ["ec2:*"]  # More specific EC2 permissions

# Team roles  
dev-team-ec2:
  inherits: [ec2-readonly]
  scopes:
    groups: [oidc:developers]
```

**Use Provider Managed Roles**
```yaml
# Leverage existing cloud roles
aws-power-user:
  inherits:
    - arn:aws:iam::aws:policy/PowerUserAccess
  # Add company-specific restrictions
  resources:
    deny:
      - "arn:aws:s3:::sensitive-*"
```

### 3. Security Patterns

**Defense in Depth**
```yaml
sensitive-admin:
  name: Sensitive Admin
  # Multiple approval layers
  workflows:
    - manager-approval
    - security-approval
  # Specific time windows
  # Limited scope
  scopes:
    users: [google-prod:emergency@company.com]
  # Explicit resource limits
  resources:
    allow: ["arn:aws:*:us-east-1:123456789012:*"]
```

**Explicit Denials**
```yaml
prod-access:
  permissions:
    allow: ["*:*"]
    deny:
      - "*:Delete*"
      - "*:Terminate*"
      - "iam:CreateUser"
```

### 4. Operational Patterns

**Environment Separation**
```yaml
# Use different providers for different environments
dev-admin:
  providers: [aws-dev, azure-dev]
  
prod-readonly:
  providers: [aws-prod, azure-prod]
  workflows: [security-approval]
```

**Emergency Access**
```yaml
break-glass:
  name: Emergency Break Glass Access
  workflows: [emergency-approval]
  scopes:
    groups: [emergency-responders]
  permissions:
    allow: ["*:*"]
  # Should be heavily audited
```

## Troubleshooting

### Common Issues

#### 1. Role Inheritance Errors
```
Error: role admin inherits from non-existent role user
```

**Solution**: Ensure all inherited roles exist:
```yaml
# Define base role first
user:
  name: Basic User
  permissions: { ... }

# Then inherit from it
admin:
  inherits: [user]
  permissions: { ... }
```

#### 2. Provider Role Not Found
```
Error: role inherits from arn:aws:iam::aws:policy/NonexistentPolicy
```

**Solution**: Verify provider role ARNs are correct and exist in the target account.

#### 3. Permission Validation Errors
```
Error: permission ec2:InvalidAction not found in provider
```

**Solution**: Check permission names against provider documentation.

#### 4. Scope Resolution Issues
```
Error: user alice@company.com cannot access role admin
```

**Solution**: Check role scopes:
```yaml
admin:
  scopes:
    users: [alice@company.com]  # Must include user
```

### Debugging Role Issues

Enable debug logging:
```yaml
logging:
  level: debug
```

Use the CLI to test roles:
```bash
# List available roles
agent roles
```

## Examples

For practical examples and templates of role configurations, see the [Role Examples](examples/) page which includes:

- **Basic Development Role** - Simple developer access patterns
- **Inherited Admin Role** - Multi-cloud administrative access using inheritance
- **Emergency Access Role** - Break-glass access for incidents
- **Read-Only Auditor Role** - Compliance and auditing access
- **Database Administrator Role** - Specialized database permissions
- **DevOps Engineer Role** - Infrastructure and deployment management
- **Security Analyst Role** - Security monitoring and investigation
- **Temporary Contractor Role** - Time-limited external access
- **Multi-Environment Role** - Different access across environments
- **Application-Specific Role** - Fine-grained application permissions

Each example includes complete YAML configurations with explanations of the patterns used.
