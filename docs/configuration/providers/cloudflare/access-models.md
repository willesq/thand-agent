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

### How Inherits and Permissions Work Together

**Inherits Only (Simple Role Assignment)**
- When only `inherits` is specified (no permissions), the role IDs are assigned directly
- User gets all permissions from the inherited Cloudflare roles
- Simple and efficient for standard role assignments

**Inherits + Permissions (Granular Control)**
- When `permissions` are specified along with `inherits`:
  1. All inherited roles are examined to build a map of available permissions
  2. Only permissions that exist in the inherited roles can be used in allow/deny
  3. `permissions.allow`: Explicitly grants specific permissions from the inherited roles
  4. `permissions.deny`: Explicitly denies specific permissions from the inherited roles
- This provides fine-grained control while ensuring permissions are valid for the inherited roles

### Example Configuration

```yaml
version: "1.0"
roles:
  # Inherits only - assigns roles directly (no permissions specified)
  cloudflare-dns-simple:
    name: DNS Editor - Simple Assignment
    description: DNS management using direct role assignment
    providers:
      - cloudflare-prod
    inherits:
      - DNS        # Assign DNS role directly
      - Analytics  # Assign Analytics role directly
    resources:
      allow:
        - zone:example.com
        - zone:api.example.com
    enabled: true

  # Inherits + Permissions - granular control over inherited permissions
  cloudflare-dns-granular:
    name: DNS Editor - Granular Control
    description: Limited DNS permissions from DNS role
    providers:
      - cloudflare-prod
    inherits:
      - DNS        # Provides available DNS permissions
      - Analytics  # Provides available Analytics permissions
    permissions:
      allow:
        - dns_records:read   # Only allow reading DNS records
        - dns_records:edit   # And editing DNS records
        - analytics:read     # And reading analytics
      # Other DNS role permissions (like zone settings) are not granted
    resources:
      allow:
        - zone:example.com
    enabled: true

  # All zones with both inherited roles and additional permissions
  cloudflare-firewall-all:
    name: Firewall Manager - All Zones
    description: Firewall management with analytics across all zones
    providers:
      - cloudflare-prod
    inherits:
      - Firewall  # Provides all Firewall-related permissions
    permissions:
      allow:
        - firewall_services:read
        - firewall_services:edit
        - waf:read
        - waf:edit
        - analytics:read  # From Firewall role (if available)
        - logs:read       # From Firewall role (if available)
    resources:
      allow:
        - zone:*  # Wildcard for all zones
    enabled: true

  # Account-level access with permissions
  cloudflare-workers:
    name: Workers Developer
    description: Workers management only (no KV or durable objects)
    providers:
      - cloudflare-prod
    inherits:
      - Workers Platform Admin  # Provides all Workers-related permissions
    permissions:
      allow:
        - workers_scripts:read
        - workers_scripts:edit
      # Deny KV and Durable Objects even though they're in Workers Platform Admin
      deny:
        - workers_kv_storage:read
        - workers_kv_storage:edit
        - workers_durable_objects:read
        - workers_durable_objects:edit
    resources:
      allow:
        - account:*  # Account-level (Workers are account resources)
    enabled: true

  # Using deny permissions to restrict inherited role permissions
  cloudflare-restricted-admin:
    name: Restricted Administrator
    description: Admin access without billing or organization changes
    providers:
      - cloudflare-prod
    inherits:
      - Administrator  # Provides all admin permissions
    permissions:
      # Only specify what to deny - everything else from Administrator is allowed
      deny:
        - billing:read
        - billing:edit
        - organization:read
        - organization:edit
    resources:
      allow:
        - zone:*
    enabled: true
```

### How It Works

**Inherits Only Mode:**
1. Extracts role IDs from inherited Cloudflare roles
2. Assigns those role IDs directly to the account member
3. User gets all permissions associated with those roles
4. Access scoped to specified resources

**Inherits + Permissions Mode:**
1. **Build Permission Map**: Loops through all inherited roles and collects all available permissions
2. **Validate Permissions**: Each permission in allow/deny lists is checked against available permissions
3. **Create Permission Groups**: Valid permissions are converted to Cloudflare permission groups
4. **Build Resource Groups**: Creates resource groups for each specified resource (zone or account)
5. **Generate Policies**: Combines permission groups with resource groups to create policies
   - Allow permissions → policies with `access: "allow"`
   - Deny permissions → policies with `access: "deny"`
6. **Apply**: Creates account member with the generated policies

---

## Side-by-Side Comparison

| Feature | Account-Wide Roles | Inherits Only | Inherits + Permissions |
|---------|-------------------|---------------|----------------------|
| **Configuration** | Role name only | Inherits + Resources | Inherits + Permissions + Resources |
| **Permissions** | All permissions for the role | All permissions from inherited roles | Only specified permissions (validated against inherited roles) |
| **Scope** | Entire account | Specific zones/resources | Specific zones/resources |
| **Flexibility** | Low (predefined roles) | Medium (role combinations) | High (granular control) |
| **Security** | Less restrictive | Moderately restrictive | Most restrictive (least privilege) |
| **Validation** | N/A | N/A | Validates permissions exist in inherited roles |
| **Use Case** | Admins, full access | Teams with standard role needs | Fine-grained, compliance scenarios |
| **Setup Complexity** | Simple | Moderate | More complex |

---

## Understanding Permission Validation

When using `inherits` + `permissions` together, the system validates that requested permissions actually exist in the inherited roles:

### Example: Valid Permissions

```yaml
cloudflare-limited-dns:
  name: Limited DNS Editor
  providers:
    - cloudflare-prod
  inherits:
    - DNS  # Provides permissions like: dns_records:read, dns_records:edit, zone:read, etc.
  permissions:
    allow:
      - dns_records:read   # ✅ Valid - exists in DNS role
      - dns_records:edit   # ✅ Valid - exists in DNS role
      - analytics:read     # ⚠️  Warning logged - NOT in DNS role, skipped
  resources:
    allow:
      - zone:example.com
```

**Result**: User gets only `dns_records:read` and `dns_records:edit`. The `analytics:read` permission is logged as a warning and skipped because it's not available in the DNS role.

### Example: Using Multiple Inherits for Permission Pool

```yaml
cloudflare-multi-role:
  name: Multi-Role Access
  providers:
    - cloudflare-prod
  inherits:
    - DNS        # Provides: dns_records:*, zone:read, etc.
    - Firewall   # Provides: firewall_services:*, waf:*, etc.
    - Analytics  # Provides: analytics:*, logs:read, etc.
  permissions:
    allow:
      - dns_records:edit      # ✅ From DNS role
      - firewall_services:edit # ✅ From Firewall role
      - analytics:read        # ✅ From Analytics role
  resources:
    allow:
      - zone:example.com
```

**Result**: All three permissions are valid because they exist in their respective inherited roles. The system builds a permission pool from all inherited roles.

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
  # Marketing team - only their zones, simple role assignment
  marketing-zones:
    name: Marketing DNS Manager
    description: DNS and Analytics for marketing domains
    providers:
      - cloudflare-prod
    inherits:
      - DNS        # Provides all DNS permissions
      - Analytics  # Provides all Analytics permissions
    resources:
      allow:
        - zone:marketing.example.com
        - zone:blog.example.com
    scopes:
      groups:
        - oidc:marketing-team
    enabled: true

  # API team - granular permission control
  api-zones:
    name: API Zone Manager
    description: Specific DNS and firewall permissions for API zones
    providers:
      - cloudflare-prod
    inherits:
      - DNS           # Provides DNS permissions to choose from
      - Firewall      # Provides Firewall permissions to choose from
      - Load Balancer # Provides Load Balancer permissions
    permissions:
      allow:
        - dns_records:read
        - dns_records:edit
        - firewall_services:read
        - firewall_services:edit
        - load_balancers:read
        - load_balancers:edit
        - analytics:read
    resources:
      allow:
        - zone:api.example.com
        - zone:api-staging.example.com
    scopes:
      groups:
        - oidc:api-team
    enabled: true

  # Security team - all zones, security features only with deny rules
  security-all-zones:
    name: Security Team Access
    description: Firewall and monitoring across all zones (no configuration changes)
    providers:
      - cloudflare-prod
    inherits:
      - Firewall              # Provides Firewall permissions
      - Cloudflare Zero Trust # Provides Zero Trust permissions
      - Analytics             # Provides Analytics permissions
    permissions:
      allow:
        - firewall_services:read
        - waf:read
        - analytics:read
        - logs:read
      deny:
        - firewall_services:edit  # Can view but not modify
        - waf:edit                # Can view but not modify
    resources:
      allow:
        - zone:*  # All zones for security monitoring
    scopes:
      groups:
        - oidc:security-team
    enabled: true
```

**Why**: Isolation between teams, granular least privilege, audit-friendly

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
