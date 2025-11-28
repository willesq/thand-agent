---
layout: default
title: Cloudflare
description: Cloudflare provider with role-based access control
parent: Providers
grand_parent: Configuration
---

# Cloudflare Provider

> **Note**: The Cloudflare provider only supports role-based access control. Unlike some other providers, Cloudflare does not support granular permission-level assignments. All access is managed through Cloudflare's predefined roles, which can be assigned either account-wide or scoped to specific resources (zones).

The Cloudflare provider enables integration with Cloudflare accounts, providing role-based access control (RBAC) that can be applied either account-wide or scoped to specific resources.

## Capabilities

- **Role-Based Access Control (RBAC)**: Supports Cloudflare's predefined account roles
- **Resource-Scoped Roles**: Assign roles to specific zones or account-level resources
- **Account Member Management**: Invite, assign roles, and remove account members
- **Role Discovery**: Access to 60+ predefined Cloudflare roles
- **Identity Management**: List and manage Cloudflare account members
- **Full-text Search**: Search for roles

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

The Cloudflare provider supports two distinct access control models using Cloudflare's predefined roles:

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
    inherits:
      - Administrator  # Assigns Cloudflare's built-in Administrator role
    resources:
      allow:
        - account:*  # Account-wide access (required)
    enabled: true
```

**Note:** The `resources.allow` field is required. Use `account:*` or `*` for account-wide access.

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

### 2. Resource-Scoped Roles (Granular RBAC)

Assigns Cloudflare roles scoped to specific resources (zones or account-level resources). This provides fine-grained access control following the principle of least privilege.

**When to use:**
- Limiting access to specific zones/domains
- Implementing least-privilege security model
- Managing multi-tenant or multi-zone environments
- Different teams manage different zones

**Example configuration:**

```yaml
roles:
  cloudflare-dns-editor:
    name: DNS Editor for Production Zones
    description: DNS and analytics access for specific production zones
    providers:
      - cloudflare-prod
    inherits:
      - DNS              # Cloudflare DNS role
      - Analytics        # Cloudflare Analytics role
    resources:
      allow:
        - zone:example.com      # Specific zone
        - zone:api.example.com  # Another specific zone
    enabled: true
```

## Resource Specification Format

**Important:** The `resources.allow` field is always required for Cloudflare roles.

When using resource-scoped roles, you can specify resources in the following formats:

| Format | Description | Example |
|--------|-------------|---------|
| `account:*` | Entire account (all resources) | `account:*` |
| `*` | Entire account (same as account:*) | `*` |
| `zone:domain.com` | Specific zone by domain name | `zone:example.com` |
| `zone:*` | All zones in the account | `zone:*` |
| Custom key | Cloudflare resource key | `com.cloudflare.api.account.zone.abc123` |

## Role Assignment

### Using the `inherits` Field

The `inherits` field specifies which Cloudflare predefined roles to assign. You can inherit multiple roles:

```yaml
roles:
  cloudflare-multi-role:
    name: DNS and Firewall Manager
    description: Manage DNS and firewall for specific zones
    providers:
      - cloudflare-prod
    inherits:
      - DNS              # Cloudflare DNS role
      - Firewall         # Cloudflare Firewall role
      - Cache Purge      # Cloudflare Cache Purge role
    resources:
      allow:
        - zone:example.com
    enabled: true
```

Use `agent providers roles list --provider cloudflare-prod` to see all available role names.

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
    inherits:
      - Administrator Read Only  # Cloudflare's read-only admin role
    resources:
      allow:
        - account:*  # Account-wide access
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
      - DNS          # Cloudflare DNS role
      - Analytics    # Cloudflare Analytics role
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
      - Workers Platform Admin  # Cloudflare Workers Platform Admin role
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

### Example 5: Security Team Access

```yaml
roles:
  cloudflare-security:
    name: Security Team Access
    description: Firewall and security monitoring across all zones
    authenticators:
      - google_oauth2
    workflows:
      - security_lead_approval
    inherits:
      - Firewall                  # Cloudflare Firewall role
      - Cloudflare Zero Trust     # Cloudflare Zero Trust role
      - Analytics                 # Cloudflare Analytics role
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
thand providers roles list --provider cloudflare-prod
```

### Authorize a User (Account-Wide Role)

```bash
thand providers authorize \
  --provider cloudflare-prod \
  --user user@example.com \
  --role cloudflare-readonly
```

### Authorize a User (Resource-Scoped Role)

```bash
# First, ensure the role is defined in your roles configuration
# Then authorize the user
thand providers authorize \
  --provider cloudflare-prod \
  --user user@example.com \
  --role cloudflare-dns-prod
```

### Revoke User Access

```bash
thand providers revoke \
  --provider cloudflare-prod \
  --user user@example.com \
  --role cloudflare-dns-prod
```

**Note:** Revoking removes the member entirely from the account, removing all their access.

### List Account Members

```bash
# List all members of the Cloudflare account
thand providers identities list --provider cloudflare-prod

# Search for a specific member
thand providers identities list --provider cloudflare-prod --filter "user@example.com"
```

## Implementation Details

### How Role-Based Access Works

When you define a role with `inherits` and `resources`:

1. **Role Lookup**: Fetches the role IDs from Cloudflare's predefined roles (e.g., "DNS", "Firewall")
2. **Resource Processing**: Creates Resource Groups for each specified resource
   - `account:*` or `*`: Full account access
   - `zone:*`: All zones
   - `zone:example.com`: Specific zone
3. **Policy Construction**: Combines role permissions with Resource Groups into Cloudflare Policies
4. **Member Creation**: Invites the user as an account member with the assigned policies

**Important:** The `resources.allow` field is always required. There is no default - you must explicitly specify which resources the role applies to.

### Role Specification

- Use the `inherits` field to specify which Cloudflare roles to assign
- Role names must match Cloudflare's predefined role names exactly
- Multiple roles can be inherited for combined permissions
- Use `thand providers roles list` to see all available role names

### Caching and Performance

- **Identity Caching**: Account members are cached to reduce API calls
- **Role Indexing**: Roles are indexed in the background using Bleve for fast searching
- **API Rate Limiting**: The provider respects Cloudflare API rate limits

### Security Considerations

1. **API Token Security**: 
   - Store API tokens securely using environment variables or secret management
   - Never commit tokens to version control
   - Rotate tokens regularly

2. **Principle of Least Privilege**:
   - Use resource-scoped roles instead of account-wide roles when possible
   - Grant only necessary roles
   - Limit resource access to specific zones when appropriate

3. **Audit Logging**:
   - Monitor member changes through Cloudflare's audit logs
   - Track authorization and revocation events in the agent logs

4. **Token Rotation**:
   - Regularly rotate API tokens
   - Use token expiration when creating tokens
   - Monitor token usage in Cloudflare dashboard

## Limitations

- **Role-Based Only**: Cloudflare only supports role-based access control; granular permission-level assignments are not available
- **Revocation**: Currently removes the member entirely from the account (doesn't support partial role removal)
- **Predefined Roles**: Only Cloudflare's predefined roles can be assigned; custom role creation is not supported
- **Zone Lookups**: Wildcard zone access (`zone:*`) may be slow for accounts with many zones

## Troubleshooting

### Authentication Errors

**Problem**: `failed to verify credentials` or `unauthorized`

**Solutions**:
- Verify your API token or API key is correct
- Check that the token has required permissions (Account Settings)
- Ensure the account ID is correct
- For API key authentication, verify the email is correct

### Member Not Found

**Problem**: `user X not found in account members`

**Solutions**:
- User may need to accept the invitation first
- Check if the user email is correct
- Verify the user has a Cloudflare account

### Role Not Found

**Problem**: `role not found` or `no matching role`

**Solutions**:
- Check role names match Cloudflare's role names exactly
- Use `thand providers roles list` to see available roles
- Verify your API token has permission to list roles

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
