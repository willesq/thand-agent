---
layout: default
title: Cloudflare Quick Start
description: Get started with Cloudflare provider in 5 minutes
parent: Cloudflare Provider
grand_parent: Providers
---

# Cloudflare Provider - Quick Start

Get up and running with the Cloudflare provider in 5 minutes.

## Prerequisites

- Cloudflare account with admin access
- Cloudflare account ID
- Access to create API tokens

## Step 1: Get Your Account ID

1. Log in to [Cloudflare Dashboard](https://dash.cloudflare.com/)
2. Select your account
3. Copy the Account ID from the URL or Account Home page
   - URL format: `https://dash.cloudflare.com/{ACCOUNT_ID}`
   - Or find it in **Manage Account → Overview**

## Step 2: Create an API Token

1. Go to [API Tokens](https://dash.cloudflare.com/profile/api-tokens)
2. Click **Create Token**
3. Click **Get started** next to **Custom token**
4. Configure the token:
   - **Token name**: `Thand Agent - Production`
   - **Permissions**:
     - Account → **Account Settings** → **Edit**
   - **Account Resources**: Select your account
   - **TTL**: Choose expiration (recommended: 1 year)
   
   > **Note**: Account member management requires the "Account Settings" edit permission. Member management is included within account settings permissions in Cloudflare's API.
5. Click **Continue to summary**
6. Click **Create Token**
7. **Copy the token** (you won't be able to see it again!)

## Step 3: Configure the Provider

Create or edit `config/providers/cloudflare.yaml`:

```yaml
version: "1.0"
providers:
  cloudflare-prod:
    name: Cloudflare Production
    description: Production Cloudflare Account
    provider: cloudflare
    enabled: true
    config:
      account_id: "your-account-id"
      api_token: "your-api-token"
```

**Security Best Practice**: Use environment variables instead:

```yaml
version: "1.0"
providers:
  cloudflare-prod:
    name: Cloudflare Production
    description: Production Cloudflare Account
    provider: cloudflare
    enabled: true
    config:
      account_id: "your-account-id"
      api_token: "your-api-token"
```

Then set environment variables:
```bash
export CLOUDFLARE_ACCOUNT_ID="your-account-id"
export CLOUDFLARE_API_TOKEN="your-api-token"
```

## Step 4: Create a Role Definition

Create `config/roles/cloudflare-dns.yaml`:

```yaml
version: "1.0"
roles:
  cloudflare-dns-editor:
    name: DNS Editor
    description: DNS management for specific zones
    authenticators:
      - google_oauth2
    workflows:
      - slack_approval
    inherits:
      - DNS        # Inherit all DNS role permissions
      - Analytics  # Inherit all Analytics role permissions
    resources:
      allow:
        - zone:example.com  # Replace with your domain
    providers:
      - cloudflare-prod
    scopes:
      groups:
        - oidc:engineering
    enabled: true
```

## Step 5: Test the Configuration

### List Available Permissions

```bash
agent providers permissions list --provider cloudflare-prod
```

Expected output:
```
DNS Read
DNS Edit
Firewall Services Read
Firewall Services Edit
...
```

### List Available Roles

```bash
agent providers roles list --provider cloudflare-prod
```

### List Account Members

```bash
agent providers identities list --provider cloudflare-prod
```

## Step 6: Authorize a User

### Using Account-Wide Role

```bash
agent providers authorize \
  --provider cloudflare-prod \
  --user engineer@example.com \
  --role "Administrator Read Only"
```

### Using Resource-Scoped Policy

```bash
agent providers authorize \
  --provider cloudflare-prod \
  --user engineer@example.com \
  --role cloudflare-dns-editor
```

## Common Use Cases

### Use Case 1: Read-Only Access

For engineers who need visibility but no edit permissions:

```yaml
roles:
  cloudflare-viewer:
    name: Cloudflare Viewer
    description: Read-only access to all zones
    workflows:
      - self_service  # Instant access
    providers:
      - cloudflare-prod
    scopes:
      groups:
        - oidc:all-engineers
    enabled: true
```

### Use Case 2: Emergency DNS Access

For on-call engineers during incidents:

```yaml
roles:
  cloudflare-oncall-dns:
    name: On-Call DNS Access
    description: Emergency DNS access for on-call engineers
    workflows:
      - oncall_auto_approve
    inherits:
      - DNS          # All DNS permissions
      - Cache Purge  # Cache purge permissions
      - Analytics    # Analytics access
    resources:
      allow:
        - zone:*  # All zones
    providers:
      - cloudflare-prod
    scopes:
      groups:
        - oidc:oncall-engineers
    enabled: true
```

### Use Case 3: Zone-Specific Access

For developers managing specific applications:

```yaml
roles:
  cloudflare-app-team:
    name: App Team - Zone Access
    description: DNS and firewall for app-specific zones
    workflows:
      - manager_approval
    inherits:
      - DNS       # DNS permissions
      - Firewall  # Firewall permissions
    permissions:
      allow:
        - analytics  # Add analytics access
    resources:
      allow:
        - zone:app.example.com
        - zone:api.app.example.com
    providers:
      - cloudflare-prod
    scopes:
      groups:
        - oidc:app-team
    enabled: true
```

## Troubleshooting

### "Failed to verify credentials"

**Cause**: Invalid API token or account ID

**Solution**:
```bash
# Test your credentials manually
curl -X GET "https://api.cloudflare.com/client/v4/accounts/YOUR_ACCOUNT_ID" \
  -H "Authorization: Bearer YOUR_API_TOKEN"
```

### "No matching permission groups found"

**Cause**: Permission names don't match Cloudflare's names

**Solution**: List available permissions first:
```bash
agent providers permissions list --provider cloudflare-prod
```

### "Failed to get zone ID"

**Cause**: Zone name doesn't exist or isn't accessible

**Solution**:
```bash
# List your zones
curl -X GET "https://api.cloudflare.com/client/v4/zones" \
  -H "Authorization: Bearer YOUR_API_TOKEN"
```

## Next Steps

1. [Full Cloudflare Provider Documentation](index.md)
2. [Configure Approval Workflows]({{ site.baseurl }}{% link configuration/workflows/index.md %})
3. [Set Up Slack Notifications]({{ site.baseurl }}{% link configuration/providers/slack/index.md %})
4. [Learn About Role Inheritance]({{ site.baseurl }}{% link configuration/roles/index.md %})

## Need Help?

- Check the [full documentation](index.md)
- Review [example configurations](https://github.com/thand-io/agent/tree/main/examples/roles/cloudflare.example.yml)
- See [Cloudflare API docs](https://developers.cloudflare.com/api/)
