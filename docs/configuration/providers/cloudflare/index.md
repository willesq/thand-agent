---
layout: default
title: Cloudflare Provider
description: Cloudflare provider with account-wide roles and resource-scoped policies
parent: Providers
grand_parent: Configuration
---

# Cloudflare Provider

The Cloudflare provider enables integration with Cloudflare accounts, providing both traditional account-wide role-based access control (RBAC) and granular resource-scoped policy-based access control.

## Capabilities

- **Role-Based Access Control (RBAC)**: Supports Cloudflare account-wide roles
- **Policy-Based Access Control**: Granular resource-scoped permissions for zones and accounts
- **Account Member Management**: Invite, assign roles/policies, and remove account members
- **Permission Discovery**: Access to 40+ Cloudflare permissions across all services
- **Identity Management**: List and manage Cloudflare account members
- **Full-text Search**: Search for permissions and roles

## Prerequisites

### Cloudflare Account Setup

1. **Cloudflare Account**: Active Cloudflare account with appropriate permissions
2. **Account ID**: Your Cloudflare account ID (found in dashboard URL or account settings)
3. **API Token/Key**: Authentication credentials with required permissions

### Required Cloudflare Permissions

The API token or API key must have the following permissions:

- **Account Settings**: Edit access (includes member management)

To create an API token with these permissions:

1. Go to [Cloudflare Dashboard](https://dash.cloudflare.com/) → Profile → API Tokens
2. Click "Create Token"
3. Use "Custom Token" template
4. Add permissions:
   - Account → **Account Settings** → **Edit**
5. Select your account under "Account Resources"
6. Create token and save it securely

> **Note**: Account member management (inviting, modifying, and removing members) is included in the "Account Settings" permission scope in Cloudflare's API.

## Authentication Methods

The Cloudflare provider supports two authentication methods:

### 1. API Token (Recommended)

Uses a Cloudflare API token for authentication:

```yaml
providers:
  cloudflare-prod:
    name: Cloudflare Production
    description: Production Cloudflare Account
    provider: cloudflare
    enabled: true
    config:
      account_id: "account-id-here"
      api_token: "your-cloudflare-api-token"
```

**Advantages of API Tokens:**
- Scoped permissions (principle of least privilege)
- Can be rotated without affecting other services
- No email association required
- More secure than global API keys

### 2. API Key with Email (Legacy)

Uses a global API key with email address:

```yaml
providers:
  cloudflare-prod:
    name: Cloudflare Production
    description: Production Cloudflare Account
    provider: cloudflare
    enabled: true
    config:
      account_id: "account-id-here"
      api_key: "your-global-api-key"
      email: "your-email@example.com"
```

**Note:** This method uses your Global API Key, which has full account access. API tokens are recommended for better security.

## Configuration Parameters

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `account_id` | Yes | string | Your Cloudflare account ID |
| `api_token` | Yes* | string | Cloudflare API token (recommended) |
| `api_key` | Yes* | string | Cloudflare global API key (legacy) |
| `email` | With api_key | string | Email associated with API key (required when using api_key) |

*Either `api_token` OR both `api_key` and `email` must be provided.

## Access Control Models

The Cloudflare provider supports two distinct access control models that can be used based on your security requirements:

### 1. Account-Wide Roles (Traditional RBAC)

Assigns predefined Cloudflare roles that grant permissions across the entire account. This is useful for broad administrative access.

**When to use:**
- Granting full account access
- Using standard Cloudflare roles
- Simple access patterns without resource restrictions

**Example configuration:**

```yaml
roles:
  cloudflare-admin:
    name: Cloudflare Administrator
    description: Full administrative access to Cloudflare account
    providers:
      - cloudflare-prod
    # No resources specified = account-wide role assignment
    enabled: true
```

**Available Account-Wide Roles:**
- Administrator
- Administrator Read Only
- Super Administrator - All Privileges
- Minimal Account Access
- Analytics
- API Gateway
- API Gateway Read
- Application Security Reports Read
- Audit Logs Viewer
- Billing
- Bot Management (Account-Wide)
- Brand Protection
- Cache Purge
- Cloudchamber Admin
- Cloudchamber Admin Read Only
- Cloudflare Access
- Cloudflare CASB
- Cloudflare CASB Read
- Cloudflare DEX
- Cloudflare Gateway
- Cloudflare Images
- Cloudflare R2 Admin
- Cloudflare R2 Read
- Cloudflare Stream
- Cloudflare Zero Trust
- Cloudflare Zero Trust PII
- Cloudflare Zero Trust Read Only
- Cloudflare Zero Trust Reporting
- Cloudflare Zero Trust Secure DNS Locations Write
- Cloudforce One Admin
- Cloudforce One Read
- Connectivity Directory Admin
- Connectivity Directory Bind
- Connectivity Directory Read
- DNS
- Email Configuration Admin
- Email Integration Admin
- Email Security Analyst
- Email Security Policy Admin
- Email Security Readonly
- Email Security Reporting
- Firewall
- HTTP Applications
- HTTP Applications Read
- Hyperdrive Admin
- Hyperdrive Readonly
- Load Balancer
- Load Balancing Account Read
- Log Share
- Log Share Reader
- Magic Network Monitoring
- Magic Network Monitoring Admin
- Magic Network Monitoring Read-Only
- Network Services Read (Magic)
- Network Services Write (Magic)
- Page Shield
- Page Shield Read
- Realtime
- Realtime Admin
- SSL/TLS, Caching, Performance, Page Rules, and Customization
- Secrets Store Admin
- Secrets Store Deployer
- Secrets Store Reporter
- Trust and Safety
- Turnstile
- Turnstile Read
- Vectorize Admin
- Vectorize Readonly
- Waiting Room Admin
- Waiting Room Read
- Workers Editor
- Workers Platform (Read-only)
- Workers Platform Admin
- Zaraz Admin
- Zaraz Edit
- Zaraz Readonly
- Zone Versioning (Account-Wide)
- Zone Versioning Read (Account-Wide)

### 2. Resource-Scoped Policies (Granular RBAC)

Creates Cloudflare Policies that combine specific permissions with specific resources (zones, accounts). This provides fine-grained access control following the principle of least privilege.

**When to use:**
- Limiting access to specific zones/domains
- Granting only necessary permissions
- Implementing least-privilege security model
- Managing multi-tenant or multi-zone environments

**Example configuration using inherits:**

```yaml
roles:
  cloudflare-dns-editor:
    name: DNS Editor for Production Zones
    description: DNS and analytics access for specific production zones
    providers:
      - cloudflare-prod
    inherits:
      - DNS              # Inherit all DNS role permissions
      - Analytics        # Inherit all Analytics role permissions
    resources:
      allow:
        - zone:example.com      # Specific zone
        - zone:api.example.com  # Another specific zone
    enabled: true
```

**Example configuration using explicit permissions:**

```yaml
roles:
  cloudflare-dns-editor:
    name: DNS Editor for Production Zones
    description: DNS and analytics access for specific production zones
    providers:
      - cloudflare-prod
    permissions:
      allow:
        - dns          # DNS read/write
        - dns_records  # DNS records management
        - analytics    # View analytics
    resources:
      allow:
        - zone:example.com      # Specific zone
        - zone:api.example.com  # Another specific zone
    enabled: true
```

**Combining inherits with permission overrides:**

You can inherit a Cloudflare role's permissions and then add or deny specific permissions:

```yaml
roles:
  cloudflare-restricted-admin:
    name: Restricted Administrator
    description: Admin access but without billing permissions
    providers:
      - cloudflare-prod
    inherits:
      - Administrator  # Inherit all Administrator permissions
    permissions:
      deny:
        - billing       # Explicitly deny billing access
        - organization  # Explicitly deny organization settings
    resources:
      allow:
        - zone:*
    enabled: true
```

## Resource Specification Format

When using resource-scoped policies, you can specify resources in the following formats:

| Format | Description | Example |
|--------|-------------|---------|
| `zone:domain.com` | Specific zone by domain name | `zone:example.com` |
| `zone:*` | All zones in the account | `zone:*` |
| `account:*` | Entire account (all resources) | `account:*` |
| `*` | Entire account (same as account:*) | `*` |
| Custom key | Cloudflare resource key | `com.cloudflare.api.account.zone.abc123` |

## Supported Permissions

Cloudflare provides two ways to specify permissions for resource-scoped policies:

### 1. Using Inherits (Recommended for Standard Roles)

Inherit permissions from predefined Cloudflare roles. This is the easiest way to grant a standard set of permissions:

```yaml
inherits:
  - DNS              # All DNS-related permissions
  - Analytics        # All analytics permissions
  - Firewall         # All firewall permissions
```

### 2. Using Explicit Permissions (Granular Control)

Specify individual permission keys for fine-grained control. Available permission keys:

- `analytics` - View analytics and reports
- `billing` - Manage billing and subscriptions
- `cache_purge` - Purge cache
- `dns` - Manage DNS settings
- `dns_records` - Manage DNS records
- `lb` - Manage load balancers
- `logs` - Access and manage logs
- `organization` - Manage organization settings
- `ssl` - Manage SSL/TLS certificates
- `waf` - Manage Web Application Firewall
- `zone_settings` - Manage zone settings
- `zones` - Manage zones

### 3. Combining Both Approaches

You can inherit a role's permissions and then override specific permissions:

```yaml
inherits:
  - Administrator    # Inherit all admin permissions
permissions:
  deny:
    - billing       # But deny billing access
    - organization  # And deny org settings
```

Use `agent providers permissions list --provider cloudflare-prod` to see all available permissions with their exact key names.

## Example Role Configurations

### Example 1: Read-Only Account Access

```yaml
roles:
  cloudflare-readonly:
    name: Cloudflare Read-Only
    description: Read-only access to entire Cloudflare account
    authenticators:
      - google_oauth2
    workflows:
      - slack_approval
    providers:
      - cloudflare-prod
    scopes:
      groups:
        - oidc:engineering
    enabled: true
```

### Example 2: DNS Management for Specific Zones

```yaml
roles:
  cloudflare-dns-prod:
    name: DNS Manager - Production Zones
    description: DNS and analytics access for production domains only
    authenticators:
      - google_oauth2
    workflows:
      - manager_approval
    inherits:
      - DNS          # Inherit all DNS permissions
      - Analytics    # Inherit all Analytics permissions
    resources:
      allow:
        - zone:example.com
        - zone:www.example.com
        - zone:api.example.com
    providers:
      - cloudflare-prod
    scopes:
      groups:
        - oidc:devops
        - oidc:sre
    enabled: true
```

### Example 3: Wildcard Zone Access

```yaml
roles:
  cloudflare-all-zones:
    name: All Zones Manager
    description: DNS and firewall management across all zones
    authenticators:
      - google_oauth2
    workflows:
      - senior_engineer_approval
    inherits:
      - DNS
      - Firewall
      - Cache Purge
    resources:
      allow:
        - zone:*  # All zones in the account
    providers:
      - cloudflare-prod
    scopes:
      groups:
        - oidc:senior-engineers
    enabled: true
```

### Example 4: Workers Development Access

```yaml
roles:
  cloudflare-workers-dev:
    name: Workers Developer
    description: Deploy and manage Workers scripts
    authenticators:
      - github_oauth
    workflows:
      - self_service  # Instant access for developers
    inherits:
      - Workers Platform Admin  # All Workers permissions
    resources:
      allow:
        - account:*  # Account-level Workers access
    providers:
      - cloudflare-prod
    scopes:
      groups:
        - oidc:developers
    enabled: true
```

### Example 5: Security Team Access with Permission Override

```yaml
roles:
  cloudflare-security:
    name: Security Team Access
    description: Firewall, WAF, and security settings management
    authenticators:
      - google_oauth2
    workflows:
      - security_lead_approval
    inherits:
      - Firewall                      # All firewall permissions
      - Cloudflare Zero Trust         # Zero Trust permissions
    permissions:
      allow:
        - logs                         # Add log access
        - analytics                    # Add analytics
      deny:
        - billing                      # Explicitly deny billing
    resources:
      allow:
        - zone:*  # All zones for security monitoring
    providers:
      - cloudflare-prod
    scopes:
      groups:
        - oidc:security
      users:
        - security-lead@example.com
    enabled: true
```

## CLI Usage Examples

### List Available Roles

```bash
# List all predefined Cloudflare roles
agent providers roles list --provider cloudflare-prod
```

### List Available Permissions

```bash
# List all Cloudflare permissions
agent providers permissions list --provider cloudflare-prod

# Search for specific permissions
agent providers permissions list --provider cloudflare-prod --filter "DNS"
```

### Authorize a User (Account-Wide Role)

```bash
agent providers authorize \
  --provider cloudflare-prod \
  --user user@example.com \
  --role cloudflare-readonly
```

### Authorize a User (Resource-Scoped Policy)

```bash
# First, ensure the role is defined in your roles configuration
# Then authorize the user
agent providers authorize \
  --provider cloudflare-prod \
  --user user@example.com \
  --role cloudflare-dns-prod
```

### Revoke User Access

```bash
agent providers revoke \
  --provider cloudflare-prod \
  --user user@example.com \
  --role cloudflare-dns-prod
```

**Note:** Revoking removes the member entirely from the account, removing all their access (both roles and policies).

### List Account Members

```bash
# List all members of the Cloudflare account
agent providers identities list --provider cloudflare-prod

# Search for a specific member
agent providers identities list --provider cloudflare-prod --filter "user@example.com"
```

## Implementation Details

### How Policy-Based Access Works

When you define a role with `inherits`, `permissions`, and `resources`:

1. **Role Inheritance Processing**: If `inherits` is specified, fetches permission groups from the inherited Cloudflare roles (e.g., "DNS", "Firewall")
2. **Permission Groups Creation**: 
   - For `inherits`: Extracts all permission groups from the specified Cloudflare roles
   - For `permissions.allow`: Creates additional permission group objects for each specified permission key
   - For `permissions.deny`: Creates permission group objects to explicitly deny
3. **Resource Group Creation**: Creates Resource Groups for each specified resource:
   - Zones are looked up by domain name
   - Wildcard zones are expanded to all zones in the account
   - Account resources are created with account-level scope
4. **Policy Construction**: 
   - Combines allowed Permission Groups (from both inherits and permissions.allow) with Resource Groups into Cloudflare Policies with `access: "allow"`
   - If deny permissions are specified, creates separate policies with `access: "deny"`
5. **Member Creation**: Invites the user as an account member with the generated policies

### Permission Allow/Deny Behavior

- **Inherits**: Brings in all permissions from one or more Cloudflare roles
- **Allow permissions**: Adds additional individual permissions beyond what's inherited
- **Deny permissions**: Explicitly denies specific permissions (overrides inherited permissions)
- Both allow and deny policies are applied to the same resource groups specified in `resources.allow`

### Recommended Approaches

1. **Use `inherits` for standard access patterns**: When you need a well-defined set of permissions that match a Cloudflare role (e.g., DNS, Firewall, Workers Platform Admin)
2. **Use `permissions.allow` for custom combinations**: When you need a specific mix of permissions that doesn't match any single role
3. **Use both together for fine-tuning**: Inherit a role's permissions, then use `permissions.allow` to add extras or `permissions.deny` to remove specific permissions
4. **Use `inherits` with optional role IDs**: The `inherits` field can also accept Cloudflare role IDs for backward compatibility

### Caching and Performance

- **Identity Caching**: Account members are cached to reduce API calls
- **Background Indexing**: Permissions and roles are indexed in the background using Bleve for fast searching
- **API Rate Limiting**: The provider respects Cloudflare API rate limits

### Security Considerations

1. **API Token Security**: 
   - Store API tokens securely using environment variables or secret management
   - Never commit tokens to version control
   - Rotate tokens regularly

2. **Principle of Least Privilege**:
   - Use resource-scoped policies instead of account-wide roles when possible
   - Grant only necessary permissions
   - Limit resource access to specific zones when appropriate

3. **Audit Logging**:
   - Monitor member changes through Cloudflare's audit logs
   - Track authorization and revocation events in the agent logs

4. **Token Rotation**:
   - Regularly rotate API tokens
   - Use token expiration when creating tokens
   - Monitor token usage in Cloudflare dashboard

## Limitations

- **Revocation**: Currently removes the member entirely from the account (doesn't support partial policy removal)
- **Custom Roles**: Predefined Cloudflare role IDs are hardcoded (may need updates if Cloudflare changes them)
- **Zone Lookups**: Wildcard zone access (`zone:*`) may be slow for accounts with many zones
- **Permission Matching**: Permission group matching is name/key-based and may need refinement

## Troubleshooting

### Authentication Errors

**Problem**: `failed to verify credentials` or `unauthorized`

**Solutions**:
- Verify your API token or API key is correct
- Check that the token has required permissions (Account Settings, Account Memberships)
- Ensure the account ID is correct
- For API key authentication, verify the email is correct

### Member Not Found

**Problem**: `user X not found in account members`

**Solutions**:
- User may need to accept the invitation first
- Check if the user email is correct
- Verify the user has a Cloudflare account

### Permission Mapping Errors

**Problem**: `no matching permission groups found for permissions`

**Solutions**:
- Check permission names match Cloudflare's permission names
- Use `agent providers permissions list` to see available permissions
- Verify your API token has permission to list Permission Groups

### Zone Not Found

**Problem**: `failed to get zone ID for X`

**Solutions**:
- Verify the zone domain name is correct
- Ensure the zone exists in your Cloudflare account
- Check that your API token has access to the zone

## API Endpoints Used

The provider interacts with the following Cloudflare API endpoints:

- `GET /accounts/{account_id}` - Verify account access
- `GET /accounts/{account_id}/members` - List account members
- `POST /accounts/{account_id}/members` - Invite account members (with roles or policies)
- `DELETE /accounts/{account_id}/members/{member_id}` - Remove account members
- `GET /accounts/{account_id}/roles` - List account roles
- `GET /accounts/{account_id}/access/groups` - List permission groups
- `GET /zones` - List zones (for wildcard resource expansion)

## Further Resources

- [Cloudflare API Documentation](https://developers.cloudflare.com/api/)
- [Cloudflare Account Roles](https://developers.cloudflare.com/fundamentals/account-and-billing/account-setup/manage-account-members/)
- [Cloudflare API Tokens](https://developers.cloudflare.com/fundamentals/api/get-started/create-token/)
- [Cloudflare Access Policies](https://developers.cloudflare.com/cloudflare-one/policies/)

## Next Steps

1. [Configure your Cloudflare provider]({{ site.baseurl }}{% link configuration/providers/cloudflare/index.md %})
2. [Create role definitions]({{ site.baseurl }}{% link configuration/roles/index.md %})
3. [Set up workflows for approval]({{ site.baseurl }}{% link configuration/workflows/index.md %})
4. [Test with CLI commands](#cli-usage-examples)
