# Cloudflare Provider

The Cloudflare provider implements role-based access control (RBAC) for Cloudflare accounts, enabling automated management of account members with both account-wide roles and granular resource-scoped policies.

## Features

- **Role-Based Access Control**: Manage Cloudflare account roles and permissions
- **Policy-Based Access Control**: Granular resource-scoped permissions (zones, accounts, specific resources)
- **Identity Management**: List and manage account members
- **Account Member Management**: Invite, assign roles/policies, and remove account members
- **Permission and Role Discovery**: Search and filter Cloudflare permissions and roles
- **Full-text Search**: Bleve-powered search for permissions and roles

## Access Control Models

The provider supports two types of access control:

### 1. Account-Wide Roles (Traditional)

Assigns predefined roles that grant permissions across the entire account:

```yaml
roles:
  cloudflare-readonly:
    name: Cloudflare Read-Only
    description: Read-only access to Cloudflare account
    providers:
      - cloudflare-prod
    # No resources specified = account-wide role
```

### 2. Resource-Scoped Policies (Granular)

Assigns specific permissions to specific resources (zones, accounts) using Cloudflare's Policy system:

```yaml
roles:
  cloudflare-dns-admin:
    description: DNS management for specific zones
    providers:
      - cloudflare-prod
    inherits:
      - DNS
      - Analytics
    resources:
      allow:
        - zone:example.com      # Specific zone
        - zone:another.com      # Another specific zone
```

```yaml
roles:
  cloudflare-all-zones:
    description: Access to all zones
    providers:
      - cloudflare-prod
    inherits:
      - DNS
      - Firewall
    resources:
      allow:
        - zone:*                # All zones
```

```yaml
roles:
  cloudflare-full-account:
    description: Full account access with specific permissions
    providers:
      - cloudflare-prod
    inherits:
      - DNS
      - Workers
      - Analytics
    resources:
      allow:
        - account:*             # Entire account
        # or just:
        - "*"                   # Same as account:*
```

## Configuration

The Cloudflare provider can be configured using either API tokens (recommended) or API keys with email.

### Using API Token (Recommended)

```yaml
providers:
  - name: cloudflare-prod
    provider: cloudflare
    description: Production Cloudflare Account
    config:
      account_id: "your-account-id"
      api_token: "your-api-token"
```

### Using API Key with Email (Legacy)

```yaml
providers:
  - name: cloudflare-prod
    provider: cloudflare
    description: Production Cloudflare Account
    config:
      account_id: "your-account-id"
      api_key: "your-api-key"
      email: "your-email@example.com"
```

## Configuration Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `account_id` | Yes | Your Cloudflare account ID |
| `api_token` | Yes* | Cloudflare API token (recommended) |
| `api_key` | Yes* | Cloudflare API key (legacy, requires email) |
| `email` | Yes* | Email associated with API key (legacy) |

*Either `api_token` OR both `api_key` and `email` must be provided.

## Supported Permissions

The provider includes a comprehensive set of Cloudflare permissions, including:

- **Account Management**: Read/edit account settings
- **Access Control**: Manage Cloudflare Access applications and policies
- **Analytics**: Access analytics data
- **Billing**: Manage billing and subscriptions
- **DNS**: Manage DNS records
- **Firewall**: Configure WAF rules and firewall services
- **Load Balancing**: Manage load balancers
- **Workers**: Manage Workers scripts, KV storage, and R2
- **Zero Trust**: Configure Zero Trust settings
- **And many more...**

## Supported Roles

The provider supports standard Cloudflare account roles:

- Administrator
- Administrator Read Only
- Super Administrator - All Privileges
- Analytics
- Billing
- Cache Purge
- DNS
- Firewall
- Load Balancer
- SSL and Certificates
- Workers Admin
- Access
- Zero Trust
- Zero Trust PII
- Zero Trust Read Only
- Zero Trust Reporting
- Magic Transit
- Stream
- Images
- API Gateway
- R2 Storage
- Workers KV Storage

## API Endpoints Used

The provider interacts with the following Cloudflare API endpoints:

- `GET /accounts` - Verify account access
- `GET /accounts/{account_id}/members` - List account members
- `POST /accounts/{account_id}/members` - Invite account members
- `DELETE /accounts/{account_id}/members/{member_id}` - Remove account members

## Permissions Required

To use this provider, the API token or API key must have the following permissions:

- **Account**: Read and Edit access
- **Account Memberships**: Read and Edit access

## Example Usage

### List Available Roles

```bash
agent providers roles list --provider cloudflare-prod
```

### Authorize a User with Account-Wide Role

```bash
agent providers authorize \
  --provider cloudflare-prod \
  --user user@example.com \
  --role "Administrator Read Only"
```

### Authorize a User with Resource-Scoped Policy

Create a role definition with resources:

```yaml
# roles/cloudflare-dns-editor.yaml
version: "1.0"
roles:
  cloudflare-dns-editor:
    description: DNS editor for specific zones
    providers:
    - cloudflare-prod
    inherits:
    - DNS
    - Analytics
    resources:
    allow:
        - zone:example.com
```

Then authorize:

```bash
agent providers authorize \
  --provider cloudflare-prod \
  --user user@example.com \
  --role cloudflare-dns-editor
```

### Revoke a User's Access

```bash
agent providers revoke \
  --provider cloudflare-prod \
  --user user@example.com \
  --role "Administrator Read Only"
```

**Note**: Revoking removes the member entirely from the account, removing all their access (both roles and policies).

### List Account Members

```bash
agent providers identities list --provider cloudflare-prod
```

## Implementation Notes

- The provider uses the official `cloudflare-go` SDK
- Supports both **account-wide roles** (traditional) and **resource-scoped policies** (granular)
- **Policy-based RBAC**: When a role specifies `resources.allow`, the provider creates Cloudflare Policies that scope permissions to specific resources
- **Role-based RBAC**: When no resources are specified, the provider assigns traditional account-wide roles
- Resource formats supported:
  - `zone:example.com` - Specific zone by name
  - `zone:*` - All zones in the account
  - `account:*` or `*` - Entire account
  - Custom resource keys (e.g., `com.cloudflare.api.account.zone.abc123`)
- The provider includes full-text search capabilities using Bleve for permissions and roles
- Identity caching is implemented to reduce API calls
- Policy creation dynamically maps permissions to Cloudflare Permission Groups via API

## Permissions and Resources

### Permission Mapping

When using policy-based access, the provider:
1. Fetches available Cloudflare Permission Groups via API
2. Maps your role's permissions to matching Permission Groups
3. Creates policies combining those Permission Groups with your Resource Groups

### Resource Specification

Resources can be specified in role definitions:

- **Wildcard zone access**: `zone:*` - Grants access to all zones
- **Specific zone**: `zone:example.com` - Grants access to a single zone
- **Account access**: `account:*` or `*` - Grants account-level access
- **Multiple resources**: List multiple resources for fine-grained control

## Error Handling

The provider includes comprehensive error handling for:

- Invalid or expired credentials
- Missing required permissions
- Account access errors
- Member management failures

## Security Considerations

1. **API Token Storage**: Store API tokens securely, never commit them to version control
2. **Principle of Least Privilege**: Use API tokens with only the required permissions
3. **Token Rotation**: Regularly rotate API tokens
4. **Audit Logging**: Monitor account member changes through Cloudflare's audit logs

## Limitations

- Custom Cloudflare roles are not currently supported (uses predefined role IDs)
- Revocation removes the member entirely (doesn't support partial policy removal)
- Zone lookups for `zone:*` may be slow for accounts with many zones
- Permission Group matching is name/key-based and may need refinement

## Future Enhancements

- Partial access revocation (update member instead of delete)
- Support for custom Cloudflare roles
- Enhanced permission group matching with fuzzy search
- Caching of Permission Groups to reduce API calls
- Support for policy updates (not just creation)
- Temporary access grants with expiration
- Bulk member operations
