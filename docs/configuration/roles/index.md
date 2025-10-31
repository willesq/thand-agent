---
layout: default
title: Roles
parent: Configuration
nav_order: 9
description: Comprehensive documentation for Thand Agent roles with intelligent inheritance and permission merging
has_children: true
---

# Roles

Roles are the core authorization mechanism in Thand Agent that define what permissions users can request and under what conditions. They act as templates that specify the scope of access, workflows for approval, and inheritance relationships that enable flexible permission management.

## Quick Start

A basic role definition:

```yaml
version: "1.0"
roles:
  aws-developer:
    name: AWS Developer Access
    description: Developer access to AWS resources
    enabled: true
    
    permissions:
      allow:
        - ec2:DescribeInstances
        - s3:GetObject
        - s3:ListBuckets
    
    scopes:
      groups:
        - developers
```

## Core Concepts

### What is a Role?

A Thand role is a configuration template that defines:
- **Permissions**: What actions can be performed (allow/deny rules)
- **Resources**: Which resources can be accessed (with allow/deny rules)  
- **Inheritance**: Which other roles this role builds upon
- **Providers**: Which provider instances can be used with this role
- **Scopes**: Who can request this role (users/groups)
- **Workflows**: How access requests are processed and approved

### Role vs Provider Roles

It's important to distinguish between:
- **Thand Roles**: Defined in your agent configuration (documented here)
- **Provider Roles**: Native roles in external systems (AWS IAM roles, Azure roles, etc.)

Thand roles can **inherit** from other roles and provider roles to leverage existing cloud IAM configurations.

### Intelligent Permission Merging

Thand Agent features intelligent permission merging that:
- **Consolidates condensed actions**: `k8s:pods:get,list` + `k8s:pods:create,update` = `k8s:pods:create,get,list,update`
- **Resolves Allow/Deny conflicts**: Deny permissions remove conflicting actions from Allow permissions
- **Handles complex inheritance**: Multi-level role inheritance with proper conflict resolution
- **Supports provider-specific naming**: AWS ARNs, GCP service accounts, Azure resource IDs with complex naming patterns

---

## Table of Contents

1. [Role Structure](#role-structure)
2. [Permissions](#permissions)
3. [Resources](#resources)
4. [Inheritance](#inheritance)
5. [Scopes & Access Control](#scopes--access-control)
6. [Provider Integration](#provider-integration)
7. [Workflow Integration](#workflow-integration)
8. [Configuration Management](#configuration-management)
9. [Best Practices](#best-practices)
10. [Troubleshooting](#troubleshooting)

---

## Role Structure

### Basic Configuration

```yaml
version: "1.0"
roles:
  role-name:
    name: Human Readable Name
    description: Description of what this role provides
    enabled: true                    # Optional, defaults to true
    
    # Core role definition
    permissions:     # What actions are allowed/denied
      allow: []
      deny: []
    resources:       # What resources can be accessed
      allow: []
      deny: []
    inherits: []     # What other roles to inherit from
    providers: []    # Which providers can be used
    
    # Access control
    scopes:          # Who can request this role
      users: []
      groups: []
    
    # Process control  
    workflows: []         # How requests are processed
    authenticators: []    # Which auth providers are valid
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
    
    # Explicit permissions with intelligent merging
    permissions:
      allow:
        - ec2:DescribeInstances,StartInstances,StopInstances  # Condensed actions
        - s3:GetObject,PutObject          # Multiple S3 actions
        - logs:DescribeLogGroups,DescribeLogStreams
      deny:
        - ec2:TerminateInstances          # Explicit denial
    
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
        - developer@example.com
      groups:
        - developers
        - engineering
    
    # Approval process
    workflows:
      - manager-approval
      - security-review
    
    # Valid authentication methods
    authenticators:
      - google-oauth
      - saml-sso
```

### Configuration Fields Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Human-readable role name |
| `description` | string | Yes | Description of role purpose |
| `enabled` | boolean | No | Whether role is active (default: true) |
| `permissions` | object | No | Allow/deny permission rules |
| `resources` | object | No | Allow/deny resource rules |
| `inherits` | array | No | List of roles to inherit from |
| `providers` | array | No | List of provider instances this role can use |
| `scopes` | object | No | User/group access restrictions |
| `workflows` | array | No | Approval workflows to execute |
| `authenticators` | array | No | Valid authentication providers |

---

## Permissions

Permissions define **what actions** can be performed when a role is activated. Thand Agent supports intelligent permission merging that handles condensed actions, inheritance conflicts, and provider-specific permission formats.

### Basic Permission Structure

```yaml
permissions:
  allow:    # List of allowed actions
    - action1
    - action2
  deny:     # List of explicitly denied actions  
    - action3
    - action4
```

### Condensed Actions

Thand Agent intelligently handles condensed actions where multiple related actions are specified in a single permission string:

```yaml
permissions:
  allow:
    # Condensed format - multiple actions in one string
    - "k8s:pods:get,list,watch,create,update,delete"
    - "s3:GetObject,PutObject,ListBucket"
    - "ec2:DescribeInstances,StartInstances,StopInstances"
    
    # Individual format - also supported
    - "logs:DescribeLogGroups"
    - "logs:DescribeLogStreams"
```

#### Intelligent Merging

When roles are inherited or merged, condensed actions are intelligently combined:

```yaml
# Base role
base-role:
  permissions:
    allow:
      - "k8s:pods:get,list,watch"
      - "s3:GetObject,ListBucket"

# Child role
child-role:
  inherits: [base-role]
  permissions:
    allow:
      - "k8s:pods:create,update,delete"  # Will merge with base
      - "s3:PutObject,DeleteObject"      # Will merge with base

# Resulting merged permissions:
# - "k8s:pods:create,delete,get,list,update,watch"  (merged and sorted)
# - "s3:DeleteObject,GetObject,ListBucket,PutObject" (merged and sorted)
```

### Cloud Provider Permission Patterns

#### AWS Permissions
```yaml
permissions:
  allow:
    - "ec2:*"                          # All EC2 actions
    - "s3:GetObject,PutObject"         # Specific S3 actions
    - "iam:PassRole"                   # IAM role assumption
    - "logs:DescribeLogGroups,DescribeLogStreams,CreateLogStream"
  deny:
    - "ec2:TerminateInstances"         # Explicit denial
    - "s3:DeleteBucket"                # Protect against deletion
```

#### Azure Permissions
```yaml
permissions:
  allow:
    - "Microsoft.Compute/virtualMachines/read,start,restart"
    - "Microsoft.Storage/storageAccounts/read"
    - "Microsoft.Authorization/roleAssignments/read"
  deny:
    - "Microsoft.Compute/virtualMachines/delete"
    - "Microsoft.Storage/storageAccounts/delete"
```

#### GCP Permissions
```yaml
permissions:
  allow:
    - "compute.instances.get,list,start,stop"
    - "storage.buckets.list,get"
    - "iam.serviceAccounts.get,list"
  deny:
    - "compute.instances.delete"
    - "storage.buckets.delete"
```

#### Kubernetes Permissions
```yaml
permissions:
  allow:
    - "k8s:pods:get,list,watch,create,update,patch"
    - "k8s:services:get,list,create,update,delete"
    - "k8s:configmaps:get,list,create,update,delete"
    - "k8s:secrets:get,list"  # Read-only for secrets
  deny:
    - "k8s:secrets:create,update,delete"  # No secret modifications
    - "k8s:pods:delete"                   # Cannot delete pods
```

### Allow/Deny Conflict Resolution

When the same action appears in both `allow` and `deny` lists, the system resolves conflicts intelligently:

#### Single Role Conflicts
```yaml
# Within a single role, deny takes precedence
role:
  permissions:
    allow:
      - "k8s:pods:get,list,create,update,delete"
    deny:
      - "k8s:pods:delete"  # Removes 'delete' from the allow list

# Resolves to:
# allow: ["k8s:pods:create,get,list,update"]
# deny: []  (deny removed since conflict was resolved)
```

#### Inheritance Conflicts
```yaml
# In inheritance, parent permissions take precedence
parent-role:
  permissions:
    allow: ["ec2:StartInstances"]
    deny: ["ec2:TerminateInstances"]

child-role:
  inherits: [parent-role]
  permissions:
    allow: ["ec2:TerminateInstances"]  # Conflicts with parent deny
    deny: ["ec2:StartInstances"]      # Conflicts with parent allow

# Resolves to (parent wins):
# allow: ["ec2:StartInstances"]      (parent allow preserved)
# deny: ["ec2:TerminateInstances"]   (parent deny preserved)
```

### Wildcard Permissions

Support for wildcard patterns varies by provider:

```yaml
permissions:
  allow:
    # AWS wildcards
    - "ec2:*"                    # All EC2 actions
    - "s3:*Object*"              # All object-related S3 actions
    
    # Kubernetes wildcards  
    - "k8s:*:*"                  # All Kubernetes actions
    - "k8s:pods:*"               # All pod actions
    
    # Azure wildcards
    - "Microsoft.Compute/*"      # All compute actions
```

---

## Resources

Resources define **what resources** the permissions can be applied to. They provide fine-grained control over which specific cloud resources, files, databases, or other assets can be accessed.

### Basic Resource Structure

```yaml
resources:
  allow:    # List of allowed resources
    - resource1
    - resource2
  deny:     # List of explicitly denied resources
    - resource3
    - resource4
```

### Cloud Provider Resource Patterns

#### AWS Resources (ARNs)
```yaml
resources:
  allow:
    # EC2 resources
    - "arn:aws:ec2:*:*:instance/*"           # All EC2 instances
    - "arn:aws:ec2:us-east-1:123456789012:instance/*"  # Specific region/account
    
    # S3 resources
    - "arn:aws:s3:::dev-bucket/*"            # Specific bucket contents
    - "arn:aws:s3:::app-*/*"                 # Pattern-based bucket access
    
    # IAM resources
    - "arn:aws:iam::123456789012:role/app-*" # Application roles only
    
    # RDS resources
    - "arn:aws:rds:*:*:db:dev-*"            # Development databases
  deny:
    - "arn:aws:s3:::prod-bucket/*"          # Sensitive production data
    - "arn:aws:iam::*:role/admin-*"         # Administrative roles
```

#### Azure Resources
```yaml
resources:
  allow:
    # Virtual machines
    - "/subscriptions/*/resourceGroups/dev-*/providers/Microsoft.Compute/virtualMachines/*"
    - "/subscriptions/12345/resourceGroups/app-*/providers/Microsoft.Compute/*"
    
    # Storage accounts
    - "/subscriptions/*/resourceGroups/*/providers/Microsoft.Storage/storageAccounts/dev*"
    
    # Resource groups
    - "/subscriptions/*/resourceGroups/development-*"
    - "/subscriptions/*/resourceGroups/staging-*"
  deny:
    - "/subscriptions/*/resourceGroups/production-*"  # No production access
    - "/subscriptions/*/resourceGroups/*/providers/Microsoft.Storage/storageAccounts/prod*"
```

#### GCP Resources
```yaml
resources:
  allow:
    # Compute instances
    - "projects/dev-project/zones/*/instances/*"
    - "projects/*/zones/us-central1-*/instances/app-*"
    
    # Storage buckets
    - "projects/*/global/buckets/dev-*"
    - "projects/my-project/global/buckets/staging-*"
    
    # Networks
    - "projects/*/global/networks/default"
    - "projects/*/regions/*/subnetworks/dev-*"
  deny:
    - "projects/prod-project/*"               # No production project access
    - "projects/*/global/buckets/sensitive-*" # Sensitive buckets
```

#### Kubernetes Resources
```yaml
resources:
  allow:
    # Namespace-scoped resources
    - "namespace:development"
    - "namespace:staging"
    - "namespace:feature-*"
    
    # Specific resource types
    - "namespace:dev/pods/*"
    - "namespace:dev/services/*"
    - "namespace:*/configmaps/app-*"
  deny:
    - "namespace:production"           # No production namespace
    - "namespace:*/secrets/*"          # No secret access
    - "namespace:kube-system"          # No system namespace
```

### Resource Inheritance and Merging

Resources from inherited roles are merged using the same intelligent system as permissions:

```yaml
# Base role
base-role:
  resources:
    allow:
      - "arn:aws:s3:::app-bucket/*"
      - "arn:aws:ec2:*:*:instance/i-dev-*"
    deny:
      - "arn:aws:s3:::app-bucket/sensitive/*"

# Child role
child-role:
  inherits: [base-role]
  resources:
    allow:
      - "arn:aws:s3:::logs-bucket/*"     # Additional resource
      - "arn:aws:ec2:*:*:instance/i-staging-*"
    deny:
      - "arn:aws:s3:::logs-bucket/audit/*"  # Additional restriction

# Merged result:
# allow: 
#   - "arn:aws:s3:::app-bucket/*"
#   - "arn:aws:s3:::logs-bucket/*"  
#   - "arn:aws:ec2:*:*:instance/i-dev-*"
#   - "arn:aws:ec2:*:*:instance/i-staging-*"
# deny:
#   - "arn:aws:s3:::app-bucket/sensitive/*"
#   - "arn:aws:s3:::logs-bucket/audit/*"
```

### Resource Pattern Matching

#### Wildcards
```yaml
resources:
  allow:
    - "arn:aws:s3:::*-dev/*"           # Any bucket ending with '-dev'
    - "arn:aws:ec2:*:*:instance/*"     # All instances in any region
    - "projects/*/zones/us-*/*"        # US zones only
```

#### Path-based Patterns
```yaml
resources:
  allow:
    # Hierarchical access
    - "projects/my-project/zones/us-central1-a/*"
    - "/subscriptions/12345/resourceGroups/dev-*/providers/*"
    
    # File-system style paths
    - "/app/data/dev/*"
    - "/shared/logs/application-*"
```

---

## Inheritance

Role inheritance is a powerful feature that allows roles to build upon each other, promoting reusability and consistent security patterns. Thand Agent features intelligent inheritance that properly handles complex permission merging, provider-specific role names, and conflict resolution.

### How Inheritance Works

When a role inherits from other roles:

1. **Permission Expansion**: Condensed actions are expanded to individual permissions
2. **Intelligent Merging**: All `allow` and `deny` permissions from inherited roles are combined at the action level
3. **Resource Merging**: All resource `allow` and `deny` rules are combined
4. **Conflict Resolution**: Allow/Deny conflicts are resolved with parent permissions taking precedence
5. **Action Condensing**: Final permissions are condensed back for clean output
6. **Scope Validation**: Inherited roles must be applicable to the requesting identity

### Inheritance Types

#### 1. Local Role Inheritance

Inherit from other Thand roles:

```yaml
roles:
  base-user:
    name: Base User
    permissions:
      allow:
        - "ec2:DescribeInstances,DescribeImages"
        - "s3:ListBuckets,GetBucketLocation"
  
  power-user:
    name: Power User
    inherits:
      - base-user  # Inherits base-user permissions
    permissions:
      allow:
        - "ec2:StartInstances,StopInstances,RebootInstances"  # Additional permissions
        - "s3:GetObject,PutObject"

# Resulting power-user permissions (intelligently merged):
# allow:
#   - "ec2:DescribeImages,DescribeInstances,RebootInstances,StartInstances,StopInstances"
#   - "s3:GetBucketLocation,GetObject,ListBuckets,PutObject"
```

#### 2. Provider Role Inheritance

Inherit from cloud provider managed roles using provider-specific syntax:

```yaml
roles:
  aws-admin:
    name: AWS Administrator
    inherits:
      # Direct AWS managed policy
      - "arn:aws:iam::aws:policy/AdministratorAccess"
      
      # Provider-scoped inheritance
      - "aws-prod:arn:aws:iam::aws:policy/ReadOnlyAccess"
    
  gcp-viewer:
    name: GCP Viewer
    inherits:
      # GCP predefined role
      - "roles/viewer"
      
      # Provider-scoped GCP role
      - "gcp-prod:roles/compute.viewer"
    permissions:
      allow:
        - "compute.instances.start,stop"  # Additional specific permissions

  azure-contributor:
    name: Azure Contributor
    inherits:
      # Azure built-in role
      - "Contributor"
      
      # Provider-scoped Azure role
      - "azure-prod:/subscriptions/12345/providers/Microsoft.Authorization/roleDefinitions/b24988ac-6180-42a0-ab88-20f7382dd24c"
```

#### 3. Complex Provider-Specific Inheritance

Handle complex role names with multiple colons (AWS ARNs, service accounts):

```yaml
roles:
  kubernetes-admin:
    name: Kubernetes Administrator
    inherits:
      # AWS ARN with multiple colons - uses first colon as delimiter
      - "aws-prod:arn:aws:iam::123456789012:role/KubernetesAdmin"
      
      # GCP service account with @ symbol
      - "gcp-prod:k8s-admin@my-project.iam.gserviceaccount.com"
      
      # Azure resource ID with multiple path segments
      - "azure-prod:/subscriptions/12345/resourceGroups/k8s/providers/Microsoft.ManagedIdentity/userAssignedIdentities/k8s-admin"

  multi-cloud-viewer:
    name: Multi-Cloud Viewer
    inherits:
      - local-base-viewer           # Local role
      - "aws-prod:arn:aws:iam::aws:policy/ReadOnlyAccess"
      - "gcp-prod:roles/viewer"
      - "azure-prod:Reader"
    permissions:
      allow:
        - "custom:audit,monitor"     # Additional custom permissions
```

#### 4. Mixed Inheritance with Intelligent Merging

Combine local and provider roles with complex permission merging:

```yaml
roles:
  base-k8s:
    name: Base Kubernetes
    permissions:
      allow:
        - "k8s:pods:get,list,watch"
        - "k8s:services:get,list"

  k8s-developer:
    name: Kubernetes Developer
    inherits: [base-k8s]
    permissions:
      allow:
        - "k8s:pods:create,update,patch"      # Merges with inherited get,list,watch
        - "k8s:services:create,update,delete" # Merges with inherited get,list
        - "k8s:configmaps:get,list,create,update,delete"
      deny:
        - "k8s:pods:delete"                   # Prevents pod deletion

  k8s-admin:
    name: Kubernetes Administrator  
    inherits: [k8s-developer]
    permissions:
      allow:
        - "k8s:pods:delete"                   # Overrides parent deny
        - "k8s:secrets:get,list,create"
        - "k8s:*:*"                          # Admin access to all
      deny:
        - "k8s:secrets:delete"                # Even admins can't delete secrets

# Final k8s-admin permissions after intelligent merging:
# allow:
#   - "k8s:*:*"  (covers everything including specific permissions)
# deny:  
#   - "k8s:secrets:delete"  (explicit restriction even for admin)
```

### Inheritance Resolution Process

The inheritance system resolves permissions in this order:

1. **Parse Inheritance**: Extract provider prefixes and role names
2. **Scope Validation**: Ensure each inherited role is applicable to the requesting identity
3. **Recursive Resolution**: Resolve inheritance chains (A inherits B inherits C)
4. **Permission Expansion**: Expand all condensed actions to individual permissions
5. **Intelligent Merging**: Combine permissions from all inheritance levels
6. **Conflict Resolution**: Apply Allow/Deny conflict resolution rules
7. **Action Condensing**: Condense related actions back for clean output
8. **Final Cleanup**: Remove redundant or conflicting permissions

### Provider-Specific Inheritance Syntax

When inheriting from provider roles, use the provider name as a prefix:

```yaml
# Format: provider-name:role-identifier
inherits:
  - "aws-prod:arn:aws:iam::123456789012:role/MyRole"      # AWS role
  - "gcp-prod:roles/storage.admin"                         # GCP role  
  - "azure-prod:Storage Blob Data Contributor"             # Azure role
  - "k8s-prod:cluster-admin"                              # Kubernetes role
```

**Parser Behavior:**
- Uses the **first colon** as the delimiter between provider and role
- Everything before first colon = provider name
- Everything after first colon = role identifier
- Handles complex identifiers like AWS ARNs with multiple colons correctly

### Inheritance Validation

The system validates inheritance chains:

#### Cyclic Inheritance Detection
```yaml
# This will be detected and rejected
role-a:
  inherits: [role-b]
role-b:
  inherits: [role-c]  
role-c:
  inherits: [role-a]  # Cycle detected!
```

#### Missing Role Detection
```yaml
# This will fail if 'nonexistent-role' doesn't exist
my-role:
  inherits: [nonexistent-role]  # Error: role not found
```

#### Scope Compatibility
```yaml
# Inherited role must be applicable to the requesting user
admin-role:
  scopes:
    groups: [admins]
    
developer-role:
  inherits: [admin-role]  # Will fail if user is not in 'admins' group
  scopes:
    groups: [developers]
```

### Inheritance Best Practices

#### 1. Build Role Hierarchies
```yaml
# Base roles with minimal permissions
readonly-base:
  permissions:
    allow: ["*:Describe*", "*:List*", "*:Get*"]

# Specialized roles building on base
ec2-readonly:
  inherits: [readonly-base]
  resources:
    allow: ["arn:aws:ec2:*:*:*"]

# Team-specific roles
dev-team-ec2:
  inherits: [ec2-readonly]
  scopes:
    groups: [developers]
```

#### 2. Use Provider Managed Roles
```yaml
# Leverage existing cloud roles
aws-power-user:
  inherits:
    - "aws-prod:arn:aws:iam::aws:policy/PowerUserAccess"
  # Add company-specific restrictions
  resources:
    deny:
      - "arn:aws:s3:::sensitive-*"
```

#### 3. Layer Security Controls
```yaml
restrictive-admin:
  name: Restrictive Admin
  inherits:
    - "aws-prod:arn:aws:iam::aws:policy/AdministratorAccess"
  # Add explicit denials even for admins
  permissions:
    deny:
      - "iam:DeleteUser"
      - "iam:DeleteRole"
      - "s3:DeleteBucket"
  resources:
    deny:
      - "arn:aws:s3:::critical-*"
```

---

## Scopes & Access Control

Scopes control **who** can request a role. This enables role-based access control at the user/group level, ensuring that only authorized identities can request specific roles.

### Scope Structure

```yaml
scopes:
  users:    # Specific user identities
    - user1@example.com
    - user2@example.com
  groups:   # Group memberships
    - group1
    - group2
```

### User Scopes

Grant access to specific users using various identity formats:

```yaml
scopes:
  users:
    - alice@example.com           # Email address
    - bob.smith@company.com       # Full name email
    - service-account@project.iam.gserviceaccount.com  # Service account
    - "123456789"                 # User ID
    - "alice.smith"               # Username
```

### Group Scopes

Grant access to groups (depends on identity provider):

```yaml
scopes:
  groups:
    - developers                  # Simple group name
    - engineering                 # Department
    - on-call                     # Role-based group
    - team-alpha                  # Team designation
    - contractors                 # Employment type
```

### Identity Provider Integration

Different identity providers may have different group formats:

```yaml
scopes:
  groups:
    # Active Directory groups
    - "DOMAIN\\Domain Users"
    - "CORP\\Engineering"
    
    # OIDC/OAuth groups  
    - "developers"
    - "admin-users"
    
    # SAML groups
    - "cn=developers,ou=groups,dc=company,dc=com"
    
    # GitHub teams
    - "my-org/developers"
    - "my-org/admin-team"
```

### Mixed Scopes

Combine users and groups for flexible access control:

```yaml
scopes:
  users:
    - emergency-admin@example.com     # Emergency access user
    - service-bot@example.com         # Automated service
  groups:
    - on-call                         # On-call team members
    - security-team                   # Security personnel
    - senior-engineers                # Senior staff
```

### Public Roles

Omit `scopes` to allow any authenticated user to request the role:

```yaml
roles:
  basic-viewer:
    name: Basic Viewer Access
    description: Read-only access available to all authenticated users
    # No 'scopes' field - available to all users
    permissions:
      allow:
        - "*:Describe*"
        - "*:List*"
        - "*:Get*"
```

### Scope Inheritance

When roles inherit from other roles, scope checking is applied to each role in the inheritance chain:

```yaml
roles:
  admin-base:
    name: Admin Base
    scopes:
      groups: [admins]
    permissions:
      allow: ["*:*"]
  
  senior-admin:
    name: Senior Admin
    inherits: [admin-base]          # User must be in 'admins' group
    scopes:
      groups: [senior-staff]        # AND in 'senior-staff' group
    permissions:
      allow: ["sensitive:*"]

# For senior-admin role to work, user must be in BOTH groups:
# - 'admins' (required by admin-base)
# - 'senior-staff' (required by senior-admin)
```

### Scope Validation Examples

#### Successful Access
```yaml
# Role definition
developer-role:
  scopes:
    groups: [developers, engineering]

# User identity
user:
  email: alice@example.com
  groups: [developers, qa-team]

# Result: ✅ Access granted (user in 'developers' group)
```

#### Failed Access
```yaml
# Role definition  
admin-role:
  scopes:
    users: [admin@example.com]
    groups: [administrators]

# User identity
user:
  email: alice@example.com
  groups: [developers]

# Result: ❌ Access denied (user not in allowed users or groups)
```

---

## Provider Integration

Roles specify which provider instances can be used for role elevation. This enables multi-cloud and multi-environment access control.

### Single Provider

Restrict a role to a specific provider instance:

```yaml
roles:
  aws-dev-access:
    name: AWS Development Access
    providers:
      - aws-dev  # Only the aws-dev provider instance
    permissions:
      allow:
        - "ec2:*"
        - "s3:*"
```

### Multi-Provider

Allow a role to work across multiple provider instances:

```yaml
roles:
  multi-cloud-viewer:
    name: Multi-Cloud Viewer Access
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

### Environment-Specific Providers

Organize providers by environment:

```yaml
roles:
  development-admin:
    name: Development Administrator
    providers:
      - aws-dev
      - azure-dev
      - gcp-dev
      - k8s-dev
    permissions:
      allow: ["*:*"]
  
  production-readonly:
    name: Production Read-Only
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

### Provider Inheritance Compatibility

When inheriting from provider roles, ensure the provider supports the inherited role:

```yaml
roles:
  aws-ec2-admin:
    name: EC2 Administrator
    providers:
      - aws-prod
    inherits:
      # This AWS managed policy must exist in the aws-prod provider
      - "aws-prod:arn:aws:iam::aws:policy/AmazonEC2FullAccess"
    
  gcp-compute-admin:
    name: GCP Compute Administrator  
    providers:
      - gcp-prod
    inherits:
      # This GCP role must be available in the gcp-prod provider
      - "gcp-prod:roles/compute.admin"
```

### Provider Validation

The system validates provider compatibility:

```yaml
# This will fail if aws-staging doesn't have the specified role
problematic-role:
  providers:
    - aws-staging
  inherits:
    - "aws-prod:arn:aws:iam::123456789012:role/CustomRole"  # Different provider!
```

**Correct approach:**
```yaml
correct-role:
  providers:
    - aws-staging
  inherits:
    - "aws-staging:arn:aws:iam::123456789012:role/CustomRole"  # Same provider
```

---

## Workflow Integration

Roles integrate with [workflows](../workflows/) to define approval processes, time limits, and other governance controls.

### Basic Workflow Assignment

```yaml
roles:
  sensitive-admin:
    name: Sensitive Admin Access
    workflows:
      - manager-approval     # Requires manager approval
      - security-review      # Additional security review
    permissions:
      allow: ["*:*"]
```

### Multiple Workflows

Workflows are executed in sequence:

```yaml
roles:
  production-access:
    name: Production Access
    workflows:
      - identity-verification    # Step 1: Verify identity
      - manager-approval         # Step 2: Manager approval  
      - security-approval        # Step 3: Security team approval
      - time-limit               # Step 4: Apply time limits
    permissions:
      allow: ["*:*"]
```

### Conditional Workflows

Workflows can implement conditional logic:

```yaml
roles:
  escalated-access:
    name: Escalated Access
    workflows:
      - risk-assessment          # Determines approval path based on risk
      # Workflow logic can route to different approval chains
    permissions:
      allow: ["*:*"]
```

### Workflow Context

Workflows receive context about the role request:

- **Role name**: Which role is being requested
- **User identity**: Who is requesting access
- **Duration**: How long access is requested for
- **Justification**: User-provided reason for access
- **Resources**: Specific resources if applicable
- **Provider**: Which provider instance will be used

### Integration Examples

#### Emergency Access
```yaml
roles:
  break-glass:
    name: Emergency Break Glass Access
    workflows:
      - emergency-notification   # Immediately notify security team
      - post-incident-review     # Schedule follow-up review
    scopes:
      groups: [on-call, security-team]
    permissions:
      allow: ["*:*"]
```

#### Development Access
```yaml
roles:
  dev-access:
    name: Development Access
    workflows:
      - self-approval           # Automatic approval for dev
      - usage-tracking          # Track usage patterns
    scopes:
      groups: [developers]
    permissions:
      allow: ["*:*"]
    resources:
      allow: ["*dev*", "*staging*"]
```

#### Audit Access
```yaml
roles:
  audit-access:
    name: Audit Access
    workflows:
      - compliance-approval     # Compliance team approval
      - audit-logging           # Enhanced audit logging
      - time-restriction        # Strict time limits
    scopes:
      groups: [auditors, compliance]
    permissions:
      allow: ["*:List*", "*:Describe*", "*:Get*"]
```

---

## Configuration Management

### File Structure Options

Roles can be organized in multiple ways to suit different organizational needs:

#### Single File Approach
```yaml
# roles.yaml
version: "1.0"
roles:
  aws-developer:
    name: AWS Developer
    permissions: { ... }
  gcp-admin:
    name: GCP Administrator  
    permissions: { ... }
  azure-viewer:
    name: Azure Viewer
    permissions: { ... }
```

#### Multiple Files by Provider
```
config/roles/
├── aws.yaml          # AWS-specific roles
├── azure.yaml        # Azure-specific roles
├── gcp.yaml          # GCP-specific roles
├── kubernetes.yaml   # Kubernetes-specific roles
└── common.yaml       # Cross-provider roles
```

**aws.yaml:**
```yaml
version: "1.0"
roles:
  aws-ec2-admin:
    name: AWS EC2 Administrator
    providers: [aws-prod, aws-dev]
    permissions:
      allow: ["ec2:*"]
  
  aws-s3-readonly:
    name: AWS S3 Read-Only
    providers: [aws-prod]
    permissions:
      allow: ["s3:Get*", "s3:List*"]
```

#### Multiple Files by Team/Function
```
config/roles/
├── developers.yaml    # Developer roles
├── admins.yaml       # Administrative roles
├── security.yaml     # Security team roles
├── readonly.yaml     # Read-only access roles
└── emergency.yaml    # Break-glass access roles
```

**developers.yaml:**
```yaml
version: "1.0"
roles:
  frontend-developer:
    name: Frontend Developer
    scopes:
      groups: [frontend-team]
    permissions:
      allow: ["s3:GetObject", "cloudfront:*"]
  
  backend-developer:
    name: Backend Developer
    scopes:
      groups: [backend-team]
    permissions:
      allow: ["ec2:*", "rds:Describe*"]
```

### Loading Configuration

Configure role loading in the main agent configuration:

#### Directory-Based Loading
```yaml
# Load all YAML files from directory
roles:
  path: "./config/roles"
  # Recursively loads all *.yaml and *.yml files
```

#### URL-Based Loading
```yaml
# Load from remote URL
roles:
  url:
    uri: "https://config.company.com/roles.yaml"
    headers:
      Authorization: "Bearer ${VAULT_TOKEN}"
    refresh_interval: "5m"      # Refresh every 5 minutes
```

#### Vault Integration
```yaml
# Load from HashiCorp Vault
roles:
  vault:
    path: "secret/agent/roles"
    key: "roles"              # Key within the secret
    refresh_interval: "10m"    # Refresh interval
```

#### Inline Definitions
```yaml
# Define roles directly in main config
roles:
  admin:
    name: Administrator
    permissions:
      allow: ["*:*"]
  readonly:
    name: Read-Only User
    permissions:
      allow: ["*:Describe*", "*:List*", "*:Get*"]
```

#### Combined Loading
```yaml
# Load from multiple sources
roles:
  sources:
    - path: "./config/roles/local"
    - url:
        uri: "https://config.company.com/shared-roles.yaml"
        headers:
          Authorization: "Bearer ${CONFIG_TOKEN}"
    - vault:
        path: "secret/team/roles"
        key: "definitions"
```

### Configuration Validation

The agent validates role configurations on startup:

#### Syntax Validation
- YAML syntax correctness
- Required field presence
- Data type validation
- Reference integrity

#### Semantic Validation
- Inheritance cycle detection
- Provider compatibility
- Permission format validation
- Resource pattern validation

#### Runtime Validation
- Provider role existence
- User/group scope resolution
- Workflow availability
- Authentication provider integration

### Hot Reloading

For certain loading methods, roles can be updated without restarting:

```yaml
roles:
  path: "./config/roles"
  auto_reload: true           # Enable hot reloading
  reload_interval: "30s"      # Check for changes every 30 seconds
```

**Supported for hot reloading:**
- File-based loading (`path`)
- URL-based loading (`url`)
- Vault-based loading (`vault`)

**Not supported for hot reloading:**
- Inline definitions
- Combined loading with inline components

---

## Best Practices

### 1. Role Design Principles

#### Principle of Least Privilege
```yaml
# ✅ Good - specific permissions
ec2-restart-role:
  name: EC2 Instance Restart
  permissions:
    allow:
      - "ec2:DescribeInstances,StartInstances,StopInstances,RebootInstances"
      - "ec2:DescribeInstanceStatus"
  resources:
    allow:
      - "arn:aws:ec2:*:*:instance/i-app-*"  # Only app instances

# ❌ Avoid - overly broad permissions  
ec2-admin-role:
  name: EC2 Admin
  permissions:
    allow: ["ec2:*"]  # Too broad
```

#### Time-Bounded Access
```yaml
# Configure time limits in workflows, not roles
time-limited-admin:
  name: Time-Limited Admin
  workflows:
    - time-limited-approval  # Implements max 2-hour access
  permissions:
    allow: ["*:*"]
```

#### Clear Naming and Documentation
```yaml
# ✅ Good - descriptive names and documentation
aws-rds-backup-operator:
  name: AWS RDS Backup Operator
  description: |
    Allows operators to manage RDS backups including:
    - Creating manual snapshots
    - Restoring from snapshots  
    - Managing automated backup settings
    - Read access to backup status and logs
    
    Does NOT allow:
    - Deleting production databases
    - Modifying database configurations
    - Creating new database instances
  
# ❌ Avoid - unclear names
role1:
  name: Some Database Access
  description: Database stuff
```

### 2. Inheritance Patterns

#### Build Logical Role Hierarchies
```yaml
# Base roles with fundamental permissions
cloud-readonly-base:
  name: Cloud Read-Only Base
  permissions:
    allow: ["*:Describe*", "*:List*", "*:Get*"]

# Service-specific roles
aws-readonly:
  name: AWS Read-Only
  inherits: [cloud-readonly-base]
  providers: [aws-prod, aws-dev]
  
ec2-readonly:
  name: EC2 Read-Only
  inherits: [aws-readonly]
  permissions:
    allow: ["ec2:*"]
  resources:
    allow: ["arn:aws:ec2:*:*:*"]

# Team-specific roles
dev-team-ec2:
  name: Development Team EC2 Access
  inherits: [ec2-readonly]
  scopes:
    groups: [developers]
  permissions:
    allow: ["ec2:StartInstances", "ec2:StopInstances"]
  resources:
    allow: ["arn:aws:ec2:*:*:instance/i-dev-*"]
```

#### Leverage Provider Managed Roles
```yaml
# ✅ Good - use existing cloud roles as foundation
aws-power-user:
  name: AWS Power User
  inherits:
    - "aws-prod:arn:aws:iam::aws:policy/PowerUserAccess"
  # Add company-specific restrictions
  permissions:
    deny: ["iam:*User*", "iam:*Role*"]  # No user/role management
  resources:
    deny: ["arn:aws:s3:::sensitive-*"]   # No sensitive buckets
```

### 3. Security Patterns

#### Defense in Depth
```yaml
production-admin:
  name: Production Administrator
  description: High-privilege production access with multiple security layers
  
  # Multiple approval layers
  workflows:
    - identity-verification
    - manager-approval
    - security-approval
    - time-restriction
  
  # Strict scope limitation
  scopes:
    users: [emergency-admin@example.com]
    groups: [senior-sre, security-team]
  
  # Explicit resource restrictions even for admin
  resources:
    allow: ["arn:aws:*:us-east-1:123456789012:*"]  # Single region only
    deny: 
      - "arn:aws:s3:::audit-*"                      # No audit data
      - "arn:aws:kms:*:*:key/*"                     # No key access
  
  permissions:
    allow: ["*:*"]
    deny:
      - "iam:DeleteUser"                            # No user deletion
      - "iam:DeleteRole"                            # No role deletion
      - "s3:DeleteBucket"                           # No bucket deletion
```

#### Explicit Denials for High-Risk Actions
```yaml
developer-access:
  name: Developer Access
  permissions:
    allow:
      - "ec2:*"
      - "s3:*"
      - "rds:*"
    deny:
      # Explicit denials for dangerous actions
      - "ec2:TerminateInstances"
      - "s3:DeleteBucket"
      - "rds:DeleteDBInstance"
      - "iam:*"                                     # No IAM access at all
```

### 4. Operational Patterns

#### Environment Separation
```yaml
# ✅ Good - clear environment separation
development-admin:
  name: Development Administrator
  providers: [aws-dev, azure-dev, gcp-dev]
  workflows: [self-approval]                        # Minimal approval for dev
  
staging-admin:
  name: Staging Administrator  
  providers: [aws-staging, azure-staging, gcp-staging]
  workflows: [lead-approval]                        # Team lead approval
  
production-readonly:
  name: Production Read-Only
  providers: [aws-prod, azure-prod, gcp-prod]
  workflows: [manager-approval, audit-logging]      # Strict controls for prod
  permissions:
    allow: ["*:Describe*", "*:List*", "*:Get*"]
```

#### Emergency Access Patterns
```yaml
break-glass-access:
  name: Emergency Break Glass Access
  description: |
    EMERGENCY USE ONLY
    This role provides unrestricted access for critical incidents.
    All usage is heavily audited and requires post-incident review.
  
  workflows:
    - emergency-notification     # Immediate alerts
    - break-glass-logging        # Enhanced audit logging
    - post-incident-review       # Mandatory follow-up
  
  scopes:
    groups: [on-call, incident-commanders]
  
  permissions:
    allow: ["*:*"]
    
  # Even emergency access has some limits
  resources:
    deny: 
      - "arn:aws:s3:::customer-data-*"             # Customer data protection
      - "arn:aws:kms:*:*:key/*"                     # Encryption key protection
```

#### Service Account Patterns
```yaml
ci-cd-deployment:
  name: CI/CD Deployment Access
  description: Automated deployment service access
  
  scopes:
    users:
      - ci-service@example.com
      - deployment-bot@example.com
  
  workflows:
    - automated-approval        # No human approval needed
    - deployment-logging        # Track all deployments
  
  permissions:
    allow:
      - "ec2:*Instance*"
      - "s3:GetObject,PutObject"
      - "ecs:*Service*"
      - "lambda:UpdateFunctionCode"
    deny:
      - "*:Delete*"             # No deletion permissions for automation
      - "*:Create*User*"        # No user creation
```

### 5. Maintenance Patterns

#### Regular Permission Audits
```yaml
# Use descriptive comments for audit trails
quarterly-access-review:
  name: Quarterly Access Review
  description: |
    Last reviewed: 2025-01-15
    Next review: 2025-04-15
    Approved by: Security Committee
    
    This role provides quarterly access review capabilities
    for compliance auditing purposes.
```

#### Version Control Integration
```yaml
# Include metadata for tracking
developer-role:
  name: Developer Access
  description: |
    Version: 2.1.0
    Last modified: 2025-01-15
    Modified by: alice@example.com
    Change reason: Added S3 read access for new logging requirements
    
    Change log:
    - 2.1.0: Added S3 read permissions
    - 2.0.0: Migrated to intelligent permission merging
    - 1.0.0: Initial role definition
```

---

## Troubleshooting

### Common Issues and Solutions

#### 1. Role Inheritance Errors

**Error:** `role admin inherits from non-existent role user`

**Cause:** Referenced role doesn't exist or isn't loaded yet

**Solution:**
```yaml
# ✅ Ensure base roles are defined before child roles
base-user:
  name: Base User
  permissions:
    allow: ["*:Describe*", "*:List*"]

admin-user:
  name: Administrator
  inherits: [base-user]  # Now this will work
  permissions:
    allow: ["*:*"]
```

#### 2. Provider Role Not Found

**Error:** `role inherits from arn:aws:iam::aws:policy/NonexistentPolicy`

**Cause:** Provider role ARN is incorrect or doesn't exist in target account

**Solutions:**
```bash
# Verify AWS managed policies
aws iam list-policies --scope AWS --query 'Policies[?PolicyName==`PowerUserAccess`]'

# Verify custom policies  
aws iam get-policy --policy-arn arn:aws:iam::123456789012:policy/CustomPolicy

# Check GCP roles
gcloud iam roles list --filter="name:roles/compute.viewer"

# Check Azure roles
az role definition list --name "Virtual Machine Contributor"
```

#### 3. Permission Validation Errors

**Error:** `permission ec2:InvalidAction not found in provider`

**Cause:** Permission name is incorrect or not supported by provider

**Solutions:**
```yaml
# ✅ Use correct AWS permission names
permissions:
  allow:
    - "ec2:DescribeInstances"     # Correct
    # - "ec2:ListInstances"       # Incorrect - no such permission

# ✅ Check provider documentation for correct names
# AWS: https://docs.aws.amazon.com/service-authorization/
# Azure: https://docs.microsoft.com/en-us/azure/role-based-access-control/
# GCP: https://cloud.google.com/iam/docs/understanding-roles
```

#### 4. Scope Resolution Issues

**Error:** `user alice@example.com cannot access role admin`

**Cause:** User not included in role scopes

**Solutions:**
```yaml
# ✅ Check role scopes include the user
admin-role:
  scopes:
    users: [alice@example.com]  # Direct user access
    groups: [administrators]    # Or group membership

# ✅ Verify user's group memberships
# Check identity provider for user's group assignments
```

#### 5. Provider-Specific Inheritance Issues

**Error:** `provider aws-prod does not support role arn:aws:iam::456:role/Role`

**Cause:** Cross-account role inheritance without proper trust relationship

**Solutions:**
```yaml
# ✅ Ensure role is in the correct account
aws-role:
  providers: [aws-prod]
  inherits:
    # Use role from same account as provider
    - "aws-prod:arn:aws:iam::123456789012:role/MyRole"  # Correct account

# ✅ Set up cross-account trust if needed
# In the target role's trust policy:
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::123456789012:root"  # Trust the source account
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

#### 6. Condensed Action Parsing Issues

**Error:** `invalid condensed action format: k8s:pods:get,list,`

**Cause:** Trailing comma or empty action in condensed format

**Solutions:**
```yaml
# ❌ Incorrect - trailing comma
permissions:
  allow:
    - "k8s:pods:get,list,"     # Trailing comma

# ✅ Correct format
permissions:
  allow:
    - "k8s:pods:get,list"      # No trailing comma
    - "k8s:services:create,delete,get,list,update"  # Properly formatted
```

### Debugging Tools and Techniques

#### Enable Debug Logging
```yaml
# In main agent configuration
logging:
  level: debug
  components:
    - roles
    - inheritance
    - permissions
```

#### Use CLI Tools for Testing
```bash
# Test role resolution (hypothetical CLI commands)
agent roles list                                    # List all available roles
agent roles describe aws-developer                  # Show role details
agent roles test alice@example.com aws-developer    # Test user access
agent roles inheritance aws-developer               # Show inheritance chain
```

#### Validate Configuration
```bash
# Validate role configuration files
agent config validate --roles-only
agent config validate --file ./config/roles/aws.yaml
```

#### Test Inheritance Resolution
```yaml
# Add temporary debug role to test inheritance
debug-inheritance:
  name: Debug Inheritance Test
  inherits: [problematic-role]
  permissions:
    allow: ["debug:test"]
  # This will show inheritance resolution issues
```

### Getting Help

#### Enable Verbose Logging
```yaml
logging:
  level: trace
  format: json
  outputs:
    - type: file
      path: /var/log/agent/roles.log
    - type: console
```

#### Check System Health
```bash
# Check provider connectivity
agent providers status

# Check identity provider integration  
agent auth status

# Check workflow system
agent workflows status
```

#### Contact Information
- **Documentation:** [https://docs.thand.io](https://docs.thand.io)
- **Community:** [GitHub Discussions](https://github.com/thand-io/agent/discussions)
- **Support:** [support@thand.io](mailto:support@thand.io)
- **Security Issues:** [security@thand.io](mailto:security@thand.io)

---

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

---

{: .note}
**Provider Prefix Syntax:** When mixing multiple providers into a single role, you can use the provider name as a prefix to avoid ambiguity. For example, to inherit from an AWS role in the `aws-prod` provider instance, use `aws-prod:arn:aws:iam::aws:policy/ReadOnlyAccess`. The system uses the first colon as the delimiter between provider name and role identifier.

