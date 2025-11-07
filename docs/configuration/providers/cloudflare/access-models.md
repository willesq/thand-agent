# Cloudflare Provider - Access Control Models Comparison

This document demonstrates the two different access control models available in the Cloudflare provider.

## Model 1: Account-Wide Roles (Traditional RBAC)

### When to Use
- Need broad, account-wide permissions
- Using standard Cloudflare role definitions
- Simple administrative access patterns
- No need for resource-specific restrictions

### Example Configuration

```yaml
version: "1.0"
roles:
  # Account-wide administrator - full access to everything
  cloudflare-admin:
    name: Cloudflare Administrator
    description: Full administrative access to entire account
    providers:
      - cloudflare-prod
    # NOTE: No resources or permissions specified
    # This uses the predefined "Administrator" role in Cloudflare
    enabled: true

  # Account-wide read-only - can view but not edit
  cloudflare-readonly:
    name: Cloudflare Read-Only
    description: Read-only access across entire account
    providers:
      - cloudflare-prod
    # Uses predefined "Administrator Read Only" role
    enabled: true
```

### How It Works
1. Provider looks up the predefined Cloudflare role by name
2. Creates account member with that role assigned
3. User gets all permissions associated with that role
4. Access applies to entire account (all zones, all resources)

### Available Predefined Roles
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

---

## Model 2: Resource-Scoped Policies (Granular RBAC)

### When to Use
- Need fine-grained access control
- Want to limit access to specific zones/resources
- Implementing least-privilege security
- Managing multi-tenant environments
- Different teams manage different zones

### Example Configuration

```yaml
version: "1.0"
roles:
  # Specific zone access with inherited permissions
  cloudflare-dns-specific:
    name: DNS Editor - Specific Zones
    description: DNS management for production zones only
    providers:
      - cloudflare-prod
    inherits:
      - DNS        # Inherit all DNS permissions
      - Analytics  # Inherit all Analytics permissions
    resources:
      allow:
        - zone:example.com
        - zone:api.example.com
    enabled: true

  # All zones with specific permissions
  cloudflare-firewall-all:
    name: Firewall Manager - All Zones
    description: Firewall management across all zones
    providers:
      - cloudflare-prod
    inherits:
      - Firewall  # Inherit all Firewall permissions
    permissions:
      allow:
        - analytics:read  # Add analytics read access
        - logs:read       # Add log read access
    resources:
      allow:
        - zone:*  # Wildcard for all zones
    enabled: true

  # Account-level access with permissions
  cloudflare-workers:
    name: Workers Developer
    description: Workers and KV storage management
    providers:
      - cloudflare-prod
    inherits:
      - Workers Platform Admin  # Inherit Workers permissions
    resources:
      allow:
        - account:*  # Account-level (Workers are account resources)
    enabled: true

  # Using deny permissions to restrict
  cloudflare-restricted-admin:
    name: Restricted Administrator
    description: Admin access without billing
    providers:
      - cloudflare-prod
    inherits:
      - Administrator  # Inherit all admin permissions
    permissions:
      deny:
        - billing:read  # Explicitly deny billing read
        - billing:edit  # Explicitly deny billing edit
    resources:
      allow:
        - zone:*
    enabled: true
```

### How It Works
1. **Inherits Processing**: Extracts permission groups from inherited Cloudflare roles (e.g., DNS, Firewall, Workers Platform Admin)
2. **Permission Addition**: Maps additional permission names (analytics, logs, etc.) to Cloudflare permission IDs
3. Creates Resource Groups for each specified resource
4. Builds Cloudflare Policies combining Permission Groups + Resource Groups
5. For inherited + allow permissions: creates policies with `access: "allow"`
6. For deny permissions: creates separate policies with `access: "deny"`
7. Creates account member with policies (not traditional roles)

---

## Side-by-Side Comparison

| Feature | Account-Wide Roles | Resource-Scoped Policies |
|---------|-------------------|-------------------------|
| **Permissions** | All permissions for the role | Only specified permissions |
| **Scope** | Entire account | Specific zones/resources |
| **Configuration** | Role name only | Permissions + Resources |
| **Flexibility** | Low (predefined roles) | High (custom combinations) |
| **Security** | Less restrictive | More restrictive (least privilege) |
| **Use Case** | Admins, full access | Teams, specific zones |
| **Setup Complexity** | Simple | Moderate |

---

## Real-World Examples

### Scenario 1: Small Team - Use Account-Wide Roles

**Situation**: Small startup with 5 engineers, everyone needs access to everything

```yaml
roles:
  team-admin:
    name: Team Administrator
    description: Full access for all team members
    providers:
      - cloudflare-prod
    scopes:
      groups:
        - oidc:engineers
    enabled: true
```

**Why**: Simple, fast, everyone trusted with full access

---

### Scenario 2: Large Organization - Use Resource-Scoped Policies

**Situation**: Large company with multiple teams managing different zones

```yaml
roles:
  # Marketing team - only their zones
  marketing-zones:
    name: Marketing DNS Manager
    description: DNS for marketing domains
    providers:
      - cloudflare-prod
    inherits:
      - DNS        # Inherit DNS permissions
      - Analytics  # Inherit Analytics permissions
    resources:
      allow:
        - zone:marketing.example.com
        - zone:blog.example.com
    scopes:
      groups:
        - oidc:marketing-team
    enabled: true

  # API team - only their zones
  api-zones:
    name: API Zone Manager
    description: Full management of API zones
    providers:
      - cloudflare-prod
    inherits:
      - DNS           # DNS permissions
      - Firewall      # Firewall permissions
      - Load Balancer # Load balancer permissions
    permissions:
      allow:
        - analytics:read  # Add analytics read access
    resources:
      allow:
        - zone:api.example.com
        - zone:api-staging.example.com
    scopes:
      groups:
        - oidc:api-team
    enabled: true

  # Security team - all zones, security features only
  security-all-zones:
    name: Security Team Access
    description: Firewall and WAF across all zones
    providers:
      - cloudflare-prod
    inherits:
      - Firewall              # Firewall permissions
      - Cloudflare Zero Trust # Zero Trust permissions
    permissions:
      allow:
        - analytics:read  # Add analytics read access
        - logs:read       # Add log read access
    resources:
      allow:
        - zone:*  # All zones for security monitoring
    scopes:
      groups:
        - oidc:security-team
    enabled: true
```

**Why**: Isolation between teams, least privilege, audit-friendly

---

### Scenario 3: Hybrid Approach

**Situation**: Most users need specific access, but some need full admin

```yaml
roles:
  # Account-wide for admins
  platform-admin:
    name: Platform Administrator
    description: Full account access for platform team
    providers:
      - cloudflare-prod
    scopes:
      groups:
        - oidc:platform-admins
    enabled: true

  # Resource-scoped for everyone else
  developer-zones:
    name: Developer Zone Access
    description: DNS and cache for dev zones
    providers:
      - cloudflare-prod
    inherits:
      - DNS          # DNS permissions
      - Cache Purge  # Cache purge permissions
      - Analytics    # Analytics access
    resources:
      allow:
        - zone:dev.example.com
        - zone:staging.example.com
    scopes:
      groups:
        - oidc:developers
    enabled: true
```

**Why**: Best of both worlds - simple for admins, restricted for others

---

## Migration Path

### From Account-Wide to Resource-Scoped

If you start with account-wide roles and want to lock down access:

**Step 1**: Current state (account-wide)
```yaml
cloudflare-access:
  name: Cloudflare Access
  providers:
    - cloudflare-prod
  enabled: true
```

**Step 2**: Identify actual usage
- What zones do users actually access?
- What permissions do they actually need?

**Step 3**: Create resource-scoped roles
```yaml
cloudflare-access-locked:
  name: Cloudflare Limited Access
  providers:
    - cloudflare-prod
  inherits:
    - DNS        # Only DNS permissions
    - Analytics  # And analytics
  resources:
    allow:
      - zone:app.example.com  # Only zones they use
  enabled: true
```

**Step 4**: Gradual rollout
- Test with a small group first
- Monitor for access issues
- Adjust permissions as needed
- Roll out to everyone

---

## Best Practices

1. **Start Restrictive**: Use resource-scoped policies by default, only grant account-wide when needed
2. **Zone Ownership**: Map resources to team ownership
3. **Permission Minimization**: Only grant required permissions
4. **Regular Audits**: Review and revoke unused access
5. **Document Decisions**: Comment why specific permissions were granted
6. **Use Workflows**: Require approvals for sensitive permissions
7. **Time-Limited Access**: Use workflows with expiration for temporary needs

---

## Quick Reference

### Use Account-Wide When:
- ✅ Full admin needed
- ✅ Small, trusted team
- ✅ Simple access patterns
- ✅ Development/staging environments

### Use Resource-Scoped When:
- ✅ Multiple teams/zones
- ✅ Need least privilege
- ✅ Production environments
- ✅ Compliance requirements
- ✅ Multi-tenant scenarios
- ✅ Per-application access
