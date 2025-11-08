# Cloudflare Provider - Access Control Models# Cloudflare Provider - Access Control Models# Cloudflare Provider - Access Control Models# Cloudflare Provider - Access Control Models Comparison



> **Note**: The Cloudflare provider only supports role-based access control. Unlike some other providers, Cloudflare does not support granular permission-level assignments. All access is managed through Cloudflare's predefined roles.



This document demonstrates the two different access control models available in the Cloudflare provider, both using Cloudflare's predefined roles.> **Note**: The Cloudflare provider only supports role-based access control. Unlike some other providers, Cloudflare does not support granular permission-level assignments. All access is managed through Cloudflare's predefined roles.



## Model 1: Account-Wide Roles (Traditional RBAC)



### When to UseThis document demonstrates the two different access control models available in the Cloudflare provider, both using Cloudflare's predefined roles.> **Note**: The Cloudflare provider only supports role-based access control. Unlike some other providers, Cloudflare does not support granular permission-level assignments. All access is managed through Cloudflare's predefined roles.This document demonstrates the two different access control models available in the Cloudflare provider.

- Need broad, account-wide permissions

- Using standard Cloudflare role definitions

- Simple administrative access patterns

- No need for resource-specific restrictions## Model 1: Account-Wide Roles (Traditional RBAC)



### Example Configuration



```yaml### When to UseThis document demonstrates the two different access control models available in the Cloudflare provider, both using Cloudflare's predefined roles.## Model 1: Account-Wide Roles (Traditional RBAC)

version: "1.0"

roles:- Need broad, account-wide permissions

  # Account-wide administrator - full access to everything

  cloudflare-admin:- Using standard Cloudflare role definitions

    name: Cloudflare Administrator

    description: Full administrative access to entire account- Simple administrative access patterns

    providers:

      - cloudflare-prod- No need for resource-specific restrictions## Model 1: Account-Wide Roles (Traditional RBAC)### When to Use

    inherits:

      - Administrator  # Cloudflare's built-in Administrator role

    resources:

      allow:### Example Configuration- Need broad, account-wide permissions

        - account:*  # Account-wide access (required)

    enabled: true



  # Account-wide read-only - can view but not edit```yaml### When to Use- Using standard Cloudflare role definitions

  cloudflare-readonly:

    name: Cloudflare Read-Onlyversion: "1.0"

    description: Read-only access across entire account

    providers:roles:- Need broad, account-wide permissions- Simple administrative access patterns

      - cloudflare-prod

    inherits:  # Account-wide administrator - full access to everything

      - Administrator Read Only  # Cloudflare's read-only admin role

    resources:  cloudflare-admin:- Using standard Cloudflare role definitions- No need for resource-specific restrictions

      allow:

        - account:*  # Account-wide access (required)    name: Cloudflare Administrator

    enabled: true

```    description: Full administrative access to entire account- Simple administrative access patterns



### How It Works    providers:

1. Provider looks up the predefined Cloudflare role by name

2. Creates resource group for account-wide access (`account:*`)      - cloudflare-prod- No need for resource-specific restrictions### Example Configuration

3. Creates account member with policies combining the role and resource

4. User gets all permissions associated with that role across the entire account    inherits:



### Available Predefined Roles      - Administrator  # Cloudflare's built-in Administrator role

Use `agent providers roles list --provider cloudflare-prod` to see all 94 available roles, including:

    resources:

- Administrator

- Administrator Read Only      allow:### Example Configuration```yaml

- Super Administrator - All Privileges

- Analytics        - account:*  # Account-wide access (required)

- Billing

- Cache Purge    enabled: trueversion: "1.0"

- DNS

- Firewall

- Load Balancer

- SSL/TLS, Caching, Performance, Page Rules, and Customization  # Account-wide read-only - can view but not edit```yamlroles:

- Workers Platform Admin

- Workers Platform (Read-only)  cloudflare-readonly:

- Workers Editor

- Cloudflare Access    name: Cloudflare Read-Onlyversion: "1.0"  # Account-wide administrator - full access to everything

- Cloudflare Zero Trust

- Cloudflare Zero Trust Read Only    description: Read-only access across entire account

- And 70+ more specialized roles

    providers:roles:  cloudflare-admin:

---

      - cloudflare-prod

## Model 2: Resource-Scoped Roles (Granular RBAC)

    inherits:  # Account-wide administrator - full access to everything    name: Cloudflare Administrator

### When to Use

- Need fine-grained access control      - Administrator Read Only  # Cloudflare's read-only admin role

- Want to limit access to specific zones/resources

- Implementing least-privilege security    resources:  cloudflare-admin:    description: Full administrative access to entire account

- Managing multi-tenant environments

- Different teams manage different zones      allow:



### How Resource-Scoped Roles Work        - account:*  # Account-wide access (required)    name: Cloudflare Administrator    providers:



When you specify specific `resources` along with `inherits`:    enabled: true

1. The inherited Cloudflare roles define what permissions the user has

2. The resources field limits WHERE those permissions can be applied```    description: Full administrative access to entire account      - cloudflare-prod

3. User gets the role's permissions, but only for the specified zones/resources



**Important:** The `resources.allow` field is always required. There is no default.

### How It Works    providers:    # NOTE: No resources or permissions specified

### Example Configuration

1. Provider looks up the predefined Cloudflare role by name

```yaml

version: "1.0"2. Creates resource group for account-wide access (`account:*`)      - cloudflare-prod    # This uses the predefined "Administrator" role in Cloudflare

roles:

  # DNS role scoped to specific zones3. Creates account member with policies combining the role and resource

  cloudflare-dns-specific:

    name: DNS Editor - Specific Zones4. User gets all permissions associated with that role across the entire account    inherits:    enabled: true

    description: DNS management for specific production zones

    providers:

      - cloudflare-prod

    inherits:### Available Predefined Roles      - Administrator  # Cloudflare's built-in Administrator role

      - DNS        # Cloudflare DNS role

      - Analytics  # Cloudflare Analytics role- Administrator

    resources:

      allow:- Administrator Read Only    # NOTE: No resources specified = account-wide role assignment  # Account-wide read-only - can view but not edit

        - zone:example.com

        - zone:api.example.com- Super Administrator - All Privileges

    enabled: true

- Analytics    enabled: true  cloudflare-readonly:

  # DNS role for all zones

  cloudflare-dns-all:- Billing

    name: DNS Manager - All Zones

    description: DNS and cache management across all zones in the account- Cache Purge    name: Cloudflare Read-Only

    providers:

      - cloudflare-prod- DNS

    inherits:

      - DNS          # Cloudflare DNS role- Firewall  # Account-wide read-only - can view but not edit    description: Read-only access across entire account

      - Cache Purge  # Cloudflare Cache Purge role

    resources:- Load Balancer

      allow:

        - zone:*  # Wildcard - all zones- SSL/TLS, Caching, Performance, Page Rules, and Customization  cloudflare-readonly:    providers:

    enabled: true

- Workers Admin

  # Firewall role for all zones

  cloudflare-firewall-all:- Workers Platform Admin    name: Cloudflare Read-Only      - cloudflare-prod

    name: Firewall Administrator

    description: Manage firewall rules and WAF across all zones- Workers Platform (Read-only)

    providers:

      - cloudflare-prod- Cloudflare Access    description: Read-only access across entire account    # Uses predefined "Administrator Read Only" role

    inherits:

      - Firewall   # Cloudflare Firewall role- Cloudflare Zero Trust

      - Analytics  # Cloudflare Analytics role

    resources:- Cloudflare Zero Trust Read Only    providers:    enabled: true

      allow:

        - zone:*  # All zones for security management- And 40+ more specialized roles

    enabled: true

      - cloudflare-prod```

  # Workers role (account-level resource)

  cloudflare-workers:Use `agent providers roles list --provider cloudflare-prod` to see all available roles.

    name: Workers Developer

    description: Deploy and manage Cloudflare Workers    inherits:

    providers:

      - cloudflare-prod---

    inherits:

      - Workers Platform Admin  # Cloudflare Workers Platform Admin role      - Administrator Read Only  # Cloudflare's read-only admin role### How It Works

    resources:

      allow:## Model 2: Resource-Scoped Roles (Granular RBAC)

        - account:*  # Account-level Workers access

    enabled: true    enabled: true1. Provider looks up the predefined Cloudflare role by name



  # Zero Trust role (account-level)### When to Use

  cloudflare-zerotrust:

    name: Zero Trust Administrator- Need fine-grained access control```2. Creates account member with that role assigned

    description: Manage Cloudflare Zero Trust and Access policies

    providers:- Want to limit access to specific zones/resources

      - cloudflare-prod

    inherits:- Implementing least-privilege security3. User gets all permissions associated with that role

      - Cloudflare Zero Trust  # Cloudflare Zero Trust role

      - Cloudflare Access      # Cloudflare Access role- Managing multi-tenant environments

    resources:

      allow:- Different teams manage different zones### How It Works4. Access applies to entire account (all zones, all resources)

        - account:*  # Account-level Zero Trust access

    enabled: true



  # SSL/TLS for specific zones### How Resource-Scoped Roles Work1. Provider looks up the predefined Cloudflare role by name

  cloudflare-ssl:

    name: SSL/TLS Manager

    description: Manage SSL certificates and TLS settings for specific zones

    providers:When you specify specific `resources` along with `inherits`:2. Creates account member with that role assigned### Available Predefined Roles

      - cloudflare-prod

    inherits:1. The inherited Cloudflare roles define what permissions the user has

      - SSL/TLS, Caching, Performance, Page Rules, and Customization

    resources:2. The resources field limits WHERE those permissions can be applied3. User gets all permissions associated with that role- Administrator

      allow:

        - zone:example.com3. User gets the role's permissions, but only for the specified zones/resources

        - zone:api.example.com

        - zone:secure.example.com4. Access applies to entire account (all zones, all resources)- Administrator Read Only

    enabled: true

```**Important:** The `resources.allow` field is always required. There is no default.



---- Super Administrator - All Privileges



## Side-by-Side Comparison### Example Configuration



| Feature | Account-Wide Roles | Resource-Scoped Roles |### Available Predefined Roles- Analytics

|---------|-------------------|----------------------|

| **Configuration** | `inherits` + `resources: [account:*]` | `inherits` + `resources: [zone:...]` |```yaml

| **Permissions** | All permissions for the role | All permissions for the role |

| **Scope** | Entire account | Specific zones/resources |version: "1.0"- Administrator- Billing

| **Flexibility** | Low (broad access) | High (limited to resources) |

| **Security** | Less restrictive | More restrictive (least privilege) |roles:

| **Use Case** | Admins, full access | Teams with specific zone responsibilities |

| **Setup Complexity** | Simple | Moderate |  # DNS role scoped to specific zones- Administrator Read Only- Cache Purge



---  cloudflare-dns-specific:



## Real-World Examples    name: DNS Editor - Specific Zones- Super Administrator - All Privileges- DNS



### Scenario 1: Small Team - Use Account-Wide Roles    description: DNS management for specific production zones



**Situation**: Small startup with 5 engineers, everyone needs access to everything    providers:- Analytics- Firewall



```yaml      - cloudflare-prod

roles:

  team-admin:    inherits:- Billing- Load Balancer

    name: Team Administrator

    description: Full access for all team members      - DNS        # Cloudflare DNS role

    providers:

      - cloudflare-prod      - Analytics  # Cloudflare Analytics role- Cache Purge- SSL and Certificates

    inherits:

      - Administrator  # Cloudflare Administrator role    resources:

    resources:

      allow:      allow:- DNS- Workers Admin

        - account:*  # Account-wide access

    scopes:        - zone:example.com

      groups:

        - oidc:engineers        - zone:api.example.com- Firewall- Access

    enabled: true

```    enabled: true



**Why**: Simple, fast, everyone trusted with full access- Load Balancer- Zero Trust



---  # DNS role for all zones



### Scenario 2: Large Organization - Use Resource-Scoped Roles  cloudflare-dns-all:- SSL/TLS, Caching, Performance, Page Rules, and Customization



**Situation**: Large company with multiple teams managing different zones    name: DNS Manager - All Zones



```yaml    description: DNS and cache management across all zones in the account- Workers Admin---

roles:

  # Marketing team - only their zones    providers:

  marketing-zones:

    name: Marketing DNS Manager      - cloudflare-prod- Workers Platform Admin

    description: DNS and Analytics for marketing domains

    providers:    inherits:

      - cloudflare-prod

    inherits:      - DNS          # Cloudflare DNS role- Workers Platform (Read-only)## Model 2: Resource-Scoped Policies (Granular RBAC)

      - DNS        # Cloudflare DNS role

      - Analytics  # Cloudflare Analytics role      - Cache Purge  # Cloudflare Cache Purge role

    resources:

      allow:    resources:- Cloudflare Access

        - zone:marketing.example.com

        - zone:blog.example.com      allow:

    scopes:

      groups:        - zone:*  # Wildcard - all zones- Cloudflare Zero Trust### When to Use

        - oidc:marketing-team

    enabled: true    enabled: true



  # API team - their zones only- Cloudflare Zero Trust Read Only- Need fine-grained access control

  api-zones:

    name: API Zone Manager  # Firewall role for all zones

    description: DNS, Firewall, and Load Balancing for API zones

    providers:  cloudflare-firewall-all:- And 40+ more specialized roles- Want to limit access to specific zones/resources

      - cloudflare-prod

    inherits:    name: Firewall Administrator

      - DNS           # Cloudflare DNS role

      - Firewall      # Cloudflare Firewall role    description: Manage firewall rules and WAF across all zones- Implementing least-privilege security

      - Load Balancer # Cloudflare Load Balancer role

      - Analytics     # Cloudflare Analytics role    providers:

    resources:

      allow:      - cloudflare-prodUse `agent providers roles list --provider cloudflare-prod` to see all available roles.- Managing multi-tenant environments

        - zone:api.example.com

        - zone:api-staging.example.com    inherits:

    scopes:

      groups:      - Firewall   # Cloudflare Firewall role- Different teams manage different zones

        - oidc:api-team

    enabled: true      - Analytics  # Cloudflare Analytics role



  # Security team - all zones, security features only    resources:---

  security-all-zones:

    name: Security Team Access      allow:

    description: Firewall and monitoring across all zones

    providers:        - zone:*  # All zones for security management### How Inherits and Permissions Work Together

      - cloudflare-prod

    inherits:    enabled: true

      - Firewall              # Cloudflare Firewall role

      - Cloudflare Zero Trust # Cloudflare Zero Trust role## Model 2: Resource-Scoped Roles (Granular RBAC)

      - Analytics             # Cloudflare Analytics role

      - Audit Logs Viewer     # Cloudflare Audit Logs Viewer role  # Workers role (account-level resource)

    resources:

      allow:  cloudflare-workers:**Inherits Only (Simple Role Assignment)**

        - zone:*      # All zones for security monitoring

        - account:*   # Account-level for Zero Trust    name: Workers Developer

    scopes:

      groups:    description: Deploy and manage Cloudflare Workers### When to Use- When only `inherits` is specified (no permissions), the role IDs are assigned directly

        - oidc:security-team

    enabled: true    providers:

```

      - cloudflare-prod- Need fine-grained access control- User gets all permissions from the inherited Cloudflare roles

**Why**: Isolation between teams, least privilege, audit-friendly

    inherits:

---

      - Workers Platform Admin  # Cloudflare Workers Platform Admin role- Want to limit access to specific zones/resources- Simple and efficient for standard role assignments

### Scenario 3: Hybrid Approach

    resources:

**Situation**: Most users need specific access, but some need full admin

      allow:- Implementing least-privilege security

```yaml

roles:        - account:*  # Account-level Workers access

  # Account-wide for admins

  platform-admin:    enabled: true- Managing multi-tenant environments**Inherits + Permissions (Granular Control)**

    name: Platform Administrator

    description: Full account access for platform team

    providers:

      - cloudflare-prod  # Zero Trust role (account-level)- Different teams manage different zones- When `permissions` are specified along with `inherits`:

    inherits:

      - Administrator  # Cloudflare Administrator role  cloudflare-zerotrust:

    resources:

      allow:    name: Zero Trust Administrator  1. All inherited roles are examined to build a map of available permissions

        - account:*  # Account-wide access

    scopes:    description: Manage Cloudflare Zero Trust and Access policies

      groups:

        - oidc:platform-admins    providers:### How Resource-Scoped Roles Work  2. Only permissions that exist in the inherited roles can be used in allow/deny

    enabled: true

      - cloudflare-prod

  # Resource-scoped for everyone else

  developer-zones:    inherits:  3. `permissions.allow`: Explicitly grants specific permissions from the inherited roles

    name: Developer Zone Access

    description: DNS and cache for dev zones      - Cloudflare Zero Trust  # Cloudflare Zero Trust role

    providers:

      - cloudflare-prod      - Cloudflare Access      # Cloudflare Access roleWhen you specify `resources` along with `inherits`:  4. `permissions.deny`: Explicitly denies specific permissions from the inherited roles

    inherits:

      - DNS          # Cloudflare DNS role    resources:

      - Cache Purge  # Cloudflare Cache Purge role

      - Analytics    # Cloudflare Analytics role      allow:1. The inherited Cloudflare roles define what permissions the user has- This provides fine-grained control while ensuring permissions are valid for the inherited roles

    resources:

      allow:        - account:*  # Account-level Zero Trust access

        - zone:dev.example.com

        - zone:staging.example.com    enabled: true2. The resources field limits WHERE those permissions can be applied

    scopes:

      groups:

        - oidc:developers

    enabled: true  # SSL/TLS for specific zones3. User gets the role's permissions, but only for the specified zones/resources### Example Configuration

```

  cloudflare-ssl:

**Why**: Best of both worlds - simple for admins, restricted for others

    name: SSL/TLS Manager

---

    description: Manage SSL certificates and TLS settings for specific zones

## Emergency Access Roles

    providers:### Example Configuration```yaml

For on-call engineers and incident response:

      - cloudflare-prod

```yaml

roles:    inherits:version: "1.0"

  cloudflare-oncall:

    name: On-Call Engineer Access      - SSL/TLS, Caching, Performance, Page Rules, and Customization

    description: Emergency access for on-call engineers across all zones

    providers:    resources:```yamlroles:

      - cloudflare-prod

    workflows:      allow:

      - oncall_verification  # Auto-approve if on-call schedule

    inherits:        - zone:example.comversion: "1.0"  # Inherits only - assigns roles directly (no permissions specified)

      - DNS                       # Cloudflare DNS role

      - Cache Purge               # Cloudflare Cache Purge role        - zone:api.example.com

      - Analytics                 # Cloudflare Analytics role

      - Administrator Read Only   # Cloudflare read-only for monitoring        - zone:secure.example.comroles:  cloudflare-dns-simple:

    resources:

      allow:    enabled: true

        - zone:*  # All zones for incident management

    scopes:```  # DNS role scoped to specific zones    name: DNS Editor - Simple Assignment

      groups:

        - oidc:oncall

        - oidc:sre

    enabled: true---  cloudflare-dns-specific:    description: DNS management using direct role assignment

```



---

## Side-by-Side Comparison    name: DNS Editor - Specific Zones    providers:

## Customer/Tenant-Specific Roles



For multi-tenant or managed service scenarios:

| Feature | Account-Wide Roles | Resource-Scoped Roles |    description: DNS management for specific production zones      - cloudflare-prod

```yaml

roles:|---------|-------------------|----------------------|

  cloudflare-customer-a:

    name: DNS Editor - Customer A| **Configuration** | `inherits` + `resources: [account:*]` | `inherits` + `resources: [zone:...]` |    providers:    inherits:

    description: DNS management for Customer A's zones

    providers:| **Permissions** | All permissions for the role | All permissions for the role |

      - cloudflare-prod

    workflows:| **Scope** | Entire account | Specific zones/resources |      - cloudflare-prod      - DNS        # Assign DNS role directly

      - customer_approval

    inherits:| **Flexibility** | Low (broad access) | High (limited to resources) |

      - DNS        # Cloudflare DNS role

      - Analytics  # Cloudflare Analytics role| **Security** | Less restrictive | More restrictive (least privilege) |    inherits:      - Analytics  # Assign Analytics role directly

    resources:

      allow:| **Use Case** | Admins, full access | Teams with specific zone responsibilities |

        - zone:customer-a.com

        - zone:www.customer-a.com| **Setup Complexity** | Simple | Moderate |      - DNS        # Cloudflare DNS role    resources:

        - zone:api.customer-a.com

    scopes:

      users:

        - customer-a-admin@example.com---      - Analytics  # Cloudflare Analytics role      allow:

        - customer-a-lead@example.com

    enabled: true

```

## Real-World Examples    resources:        - zone:example.com

---



## Development/Testing Roles

### Scenario 1: Small Team - Use Account-Wide Roles      allow:        - zone:api.example.com

For development and staging environments:



```yaml

roles:**Situation**: Small startup with 5 engineers, everyone needs access to everything        - zone:example.com    enabled: true

  cloudflare-dev-fullaccess:

    name: Development Full Access

    description: Combined role access for full development environment access

    providers:```yaml        - zone:api.example.com

      - cloudflare-dev

      - cloudflare-stagingroles:

    workflows:

      - self_service  # Instant access in dev  team-admin:    enabled: true  # Inherits + Permissions - granular control over inherited permissions

    inherits:

      - DNS                                                           # Cloudflare DNS role    name: Team Administrator

      - Firewall                                                      # Cloudflare Firewall role

      - Workers Platform Admin                                        # Cloudflare Workers Platform Admin role    description: Full access for all team members  cloudflare-dns-granular:

      - SSL/TLS, Caching, Performance, Page Rules, and Customization # Cloudflare SSL/TLS role

      - Cache Purge                                                   # Cloudflare Cache Purge role    providers:

      - Analytics                                                     # Cloudflare Analytics role

    resources:      - cloudflare-prod  # DNS role for all zones    name: DNS Editor - Granular Control

      allow:

        - zone:*  # All zones in dev account    inherits:

    scopes:

      groups:      - Administrator  # Cloudflare Administrator role  cloudflare-dns-all:    description: Limited DNS permissions from DNS role

        - oidc:developers

        - oidc:qa-team    resources:

    enabled: true

```      allow:    name: DNS Manager - All Zones    providers:



---        - account:*  # Account-wide access



## Migration Path    scopes:    description: DNS and cache management across all zones in the account      - cloudflare-prod



### From Account-Wide to Resource-Scoped      groups:



If you start with account-wide roles and want to lock down access:        - oidc:engineers    providers:    inherits:



**Step 1**: Current state (account-wide)    enabled: true

```yaml

cloudflare-access:```      - cloudflare-prod      - DNS        # Provides available DNS permissions

  name: Cloudflare Access

  providers:

    - cloudflare-prod

  inherits:**Why**: Simple, fast, everyone trusted with full access    inherits:      - Analytics  # Provides available Analytics permissions

    - Administrator  # Cloudflare Administrator role

  resources:

    allow:

      - account:*  # Account-wide---      - DNS          # Cloudflare DNS role    permissions:

  enabled: true

```



**Step 2**: Identify actual usage### Scenario 2: Large Organization - Use Resource-Scoped Roles      - Cache Purge  # Cloudflare Cache Purge role      allow:

- What zones do users actually access?

- What roles do they actually need?



**Step 3**: Create resource-scoped roles**Situation**: Large company with multiple teams managing different zones    resources:        - dns_records:read   # Only allow reading DNS records

```yaml

cloudflare-access-locked:

  name: Cloudflare Limited Access

  providers:```yaml      allow:        - dns_records:edit   # And editing DNS records

    - cloudflare-prod

  inherits:roles:

    - DNS        # Cloudflare DNS role only

    - Analytics  # Cloudflare Analytics role  # Marketing team - only their zones        - zone:*  # Wildcard - all zones        - analytics:read     # And reading analytics

  resources:

    allow:  marketing-zones:

      - zone:app.example.com  # Only zones they use

  enabled: true    name: Marketing DNS Manager    enabled: true      # Other DNS role permissions (like zone settings) are not granted

```

    description: DNS and Analytics for marketing domains

**Step 4**: Gradual rollout

- Test with a small group first    providers:    resources:

- Monitor for access issues

- Adjust roles as needed      - cloudflare-prod

- Roll out to everyone

    inherits:  # Firewall role for all zones      allow:

---

      - DNS        # Cloudflare DNS role

## Best Practices

      - Analytics  # Cloudflare Analytics role  cloudflare-firewall-all:        - zone:example.com

1. **Start Restrictive**: Use resource-scoped roles by default, only grant account-wide when needed

2. **Zone Ownership**: Map resources to team ownership    resources:

3. **Role Minimization**: Only grant required roles

4. **Regular Audits**: Review and revoke unused access      allow:    name: Firewall Administrator    enabled: true

5. **Document Decisions**: Comment why specific roles were granted

6. **Use Workflows**: Require approvals for sensitive roles        - zone:marketing.example.com

7. **Time-Limited Access**: Use workflows with expiration for temporary needs

        - zone:blog.example.com    description: Manage firewall rules and WAF across all zones

---

    scopes:

## Quick Reference

      groups:    providers:  # All zones with both inherited roles and additional permissions

### Use Account-Wide When:

- ✅ Full admin needed        - oidc:marketing-team

- ✅ Small, trusted team

- ✅ Simple access patterns    enabled: true      - cloudflare-prod  cloudflare-firewall-all:

- ✅ Development/staging environments



### Use Resource-Scoped When:

- ✅ Multiple teams/zones  # API team - their zones only    inherits:    name: Firewall Manager - All Zones

- ✅ Need least privilege

- ✅ Production environments  api-zones:

- ✅ Compliance requirements

- ✅ Multi-tenant scenarios    name: API Zone Manager      - Firewall   # Cloudflare Firewall role    description: Firewall management with analytics across all zones

- ✅ Per-application access

    description: DNS, Firewall, and Load Balancing for API zones

---

    providers:      - Analytics  # Cloudflare Analytics role    providers:

## Common Patterns

      - cloudflare-prod

### Pattern 1: Read-Only Monitoring

```yaml    inherits:    resources:      - cloudflare-prod

cloudflare-monitoring:

  name: Monitoring Access      - DNS           # Cloudflare DNS role

  inherits:

    - Administrator Read Only  # Cloudflare read-only admin      - Firewall      # Cloudflare Firewall role      allow:    inherits:

    - Analytics                # Cloudflare Analytics

  resources:      - Load Balancer # Cloudflare Load Balancer role

    allow:

      - zone:*      - Analytics     # Cloudflare Analytics role        - zone:*  # All zones for security management      - Firewall  # Provides all Firewall-related permissions

```

    resources:

### Pattern 2: DNS Only

```yaml      allow:    enabled: true    permissions:

cloudflare-dns:

  name: DNS Management        - zone:api.example.com

  inherits:

    - DNS        # Cloudflare DNS role        - zone:api-staging.example.com      allow:

    - Analytics  # Cloudflare Analytics role

  resources:    scopes:

    allow:

      - zone:example.com      groups:  # Workers role (account-level resource)        - firewall_services:read

```

        - oidc:api-team

### Pattern 3: Security Operations

```yaml    enabled: true  cloudflare-workers:        - firewall_services:edit

cloudflare-security:

  name: Security Operations

  inherits:

    - Firewall                  # Cloudflare Firewall role  # Security team - all zones, security features only    name: Workers Developer        - waf:read

    - Cloudflare Zero Trust     # Cloudflare Zero Trust role

    - Audit Logs Viewer         # Cloudflare Audit Logs Viewer  security-all-zones:

  resources:

    allow:    name: Security Team Access    description: Deploy and manage Cloudflare Workers        - waf:edit

      - zone:*

      - account:*    description: Firewall and monitoring across all zones

```

    providers:    providers:        - analytics:read  # From Firewall role (if available)

### Pattern 4: Application Development

```yaml      - cloudflare-prod

cloudflare-app-dev:

  name: Application Development    inherits:      - cloudflare-prod        - logs:read       # From Firewall role (if available)

  inherits:

    - Workers Platform Admin  # Cloudflare Workers Platform Admin      - Firewall              # Cloudflare Firewall role

    - DNS                     # Cloudflare DNS role

    - Cache Purge             # Cloudflare Cache Purge role      - Cloudflare Zero Trust # Cloudflare Zero Trust role    inherits:    resources:

  resources:

    allow:      - Analytics             # Cloudflare Analytics role

      - account:*  # For Workers

      - zone:app.example.com      - Audit Logs Viewer     # Cloudflare Audit Logs Viewer role      - Workers Platform Admin  # Cloudflare Workers Platform Admin role      allow:

```

    resources:

      allow:    resources:        - zone:*  # Wildcard for all zones

        - zone:*      # All zones for security monitoring

        - account:*   # Account-level for Zero Trust      allow:    enabled: true

    scopes:

      groups:        - account:*  # Account-level Workers access

        - oidc:security-team

    enabled: true    enabled: true  # Account-level access with permissions

```

  cloudflare-workers:

**Why**: Isolation between teams, least privilege, audit-friendly

  # Zero Trust role (account-level)    name: Workers Developer

---

  cloudflare-zerotrust:    description: Workers management only (no KV or durable objects)

### Scenario 3: Hybrid Approach

    name: Zero Trust Administrator    providers:

**Situation**: Most users need specific access, but some need full admin

    description: Manage Cloudflare Zero Trust and Access policies      - cloudflare-prod

```yaml

roles:    providers:    inherits:

  # Account-wide for admins

  platform-admin:      - cloudflare-prod      - Workers Platform Admin  # Provides all Workers-related permissions

    name: Platform Administrator

    description: Full account access for platform team    inherits:    permissions:

    providers:

      - cloudflare-prod      - Cloudflare Zero Trust  # Cloudflare Zero Trust role      allow:

    inherits:

      - Administrator  # Cloudflare Administrator role      - Cloudflare Access      # Cloudflare Access role        - workers_scripts:read

    resources:

      allow:    resources:        - workers_scripts:edit

        - account:*  # Account-wide access

    scopes:      allow:      # Deny KV and Durable Objects even though they're in Workers Platform Admin

      groups:

        - oidc:platform-admins        - account:*  # Account-level Zero Trust access      deny:

    enabled: true

    enabled: true        - workers_kv_storage:read

  # Resource-scoped for everyone else

  developer-zones:        - workers_kv_storage:edit

    name: Developer Zone Access

    description: DNS and cache for dev zones  # SSL/TLS for specific zones        - workers_durable_objects:read

    providers:

      - cloudflare-prod  cloudflare-ssl:        - workers_durable_objects:edit

    inherits:

      - DNS          # Cloudflare DNS role    name: SSL/TLS Manager    resources:

      - Cache Purge  # Cloudflare Cache Purge role

      - Analytics    # Cloudflare Analytics role    description: Manage SSL certificates and TLS settings for specific zones      allow:

    resources:

      allow:    providers:        - account:*  # Account-level (Workers are account resources)

        - zone:dev.example.com

        - zone:staging.example.com      - cloudflare-prod    enabled: true

    scopes:

      groups:    inherits:

        - oidc:developers

    enabled: true      - SSL/TLS, Caching, Performance, Page Rules, and Customization  # Using deny permissions to restrict inherited role permissions

```

    resources:  cloudflare-restricted-admin:

**Why**: Best of both worlds - simple for admins, restricted for others

      allow:    name: Restricted Administrator

---

        - zone:example.com    description: Admin access without billing or organization changes

## Emergency Access Roles

        - zone:api.example.com    providers:

For on-call engineers and incident response:

        - zone:secure.example.com      - cloudflare-prod

```yaml

roles:    enabled: true    inherits:

  cloudflare-oncall:

    name: On-Call Engineer Access```      - Administrator  # Provides all admin permissions

    description: Emergency access for on-call engineers across all zones

    providers:    permissions:

      - cloudflare-prod

    workflows:---      # Only specify what to deny - everything else from Administrator is allowed

      - oncall_verification  # Auto-approve if on-call schedule

    inherits:      deny:

      - DNS                       # Cloudflare DNS role

      - Cache Purge               # Cloudflare Cache Purge role## Side-by-Side Comparison        - billing:read

      - Analytics                 # Cloudflare Analytics role

      - Administrator Read Only   # Cloudflare read-only for monitoring        - billing:edit

    resources:

      allow:| Feature | Account-Wide Roles | Resource-Scoped Roles |        - organization:read

        - zone:*  # All zones for incident management

    scopes:|---------|-------------------|----------------------|        - organization:edit

      groups:

        - oidc:oncall| **Configuration** | Role name only (inherits) | Inherits + Resources |    resources:

        - oidc:sre

    enabled: true| **Permissions** | All permissions for the role | All permissions for the role |      allow:

```

| **Scope** | Entire account | Specific zones/resources |        - zone:*

---

| **Flexibility** | Low (broad access) | High (limited to resources) |    enabled: true

## Customer/Tenant-Specific Roles

| **Security** | Less restrictive | More restrictive (least privilege) |```

For multi-tenant or managed service scenarios:

| **Use Case** | Admins, full access | Teams with specific zone responsibilities |

```yaml

roles:| **Setup Complexity** | Simple | Moderate |### How It Works

  cloudflare-customer-a:

    name: DNS Editor - Customer A

    description: DNS management for Customer A's zones

    providers:---**Inherits Only Mode:**

      - cloudflare-prod

    workflows:1. Extracts role IDs from inherited Cloudflare roles

      - customer_approval

    inherits:## Real-World Examples2. Assigns those role IDs directly to the account member

      - DNS        # Cloudflare DNS role

      - Analytics  # Cloudflare Analytics role3. User gets all permissions associated with those roles

    resources:

      allow:### Scenario 1: Small Team - Use Account-Wide Roles4. Access scoped to specified resources

        - zone:customer-a.com

        - zone:www.customer-a.com

        - zone:api.customer-a.com

    scopes:**Situation**: Small startup with 5 engineers, everyone needs access to everything**Inherits + Permissions Mode:**

      users:

        - customer-a-admin@example.com1. **Build Permission Map**: Loops through all inherited roles and collects all available permissions

        - customer-a-lead@example.com

    enabled: true```yaml2. **Validate Permissions**: Each permission in allow/deny lists is checked against available permissions

```

roles:3. **Create Permission Groups**: Valid permissions are converted to Cloudflare permission groups

---

  team-admin:4. **Build Resource Groups**: Creates resource groups for each specified resource (zone or account)

## Development/Testing Roles

    name: Team Administrator5. **Generate Policies**: Combines permission groups with resource groups to create policies

For development and staging environments:

    description: Full access for all team members   - Allow permissions → policies with `access: "allow"`

```yaml

roles:    providers:   - Deny permissions → policies with `access: "deny"`

  cloudflare-dev-fullaccess:

    name: Development Full Access      - cloudflare-prod6. **Apply**: Creates account member with the generated policies

    description: Combined role access for full development environment access

    providers:    inherits:

      - cloudflare-dev

      - cloudflare-staging      - Administrator  # Cloudflare Administrator role---

    workflows:

      - self_service  # Instant access in dev    scopes:

    inherits:

      - DNS                                                           # Cloudflare DNS role      groups:## Side-by-Side Comparison

      - Firewall                                                      # Cloudflare Firewall role

      - Workers Platform Admin                                        # Cloudflare Workers Platform Admin role        - oidc:engineers

      - SSL/TLS, Caching, Performance, Page Rules, and Customization # Cloudflare SSL/TLS role

      - Cache Purge                                                   # Cloudflare Cache Purge role    enabled: true| Feature | Account-Wide Roles | Inherits Only | Inherits + Permissions |

      - Analytics                                                     # Cloudflare Analytics role

    resources:```|---------|-------------------|---------------|----------------------|

      allow:

        - zone:*  # All zones in dev account| **Configuration** | Role name only | Inherits + Resources | Inherits + Permissions + Resources |

    scopes:

      groups:**Why**: Simple, fast, everyone trusted with full access| **Permissions** | All permissions for the role | All permissions from inherited roles | Only specified permissions (validated against inherited roles) |

        - oidc:developers

        - oidc:qa-team| **Scope** | Entire account | Specific zones/resources | Specific zones/resources |

    enabled: true

```---| **Flexibility** | Low (predefined roles) | Medium (role combinations) | High (granular control) |



---| **Security** | Less restrictive | Moderately restrictive | Most restrictive (least privilege) |



## Migration Path### Scenario 2: Large Organization - Use Resource-Scoped Roles| **Validation** | N/A | N/A | Validates permissions exist in inherited roles |



### From Account-Wide to Resource-Scoped| **Use Case** | Admins, full access | Teams with standard role needs | Fine-grained, compliance scenarios |



If you start with account-wide roles and want to lock down access:**Situation**: Large company with multiple teams managing different zones| **Setup Complexity** | Simple | Moderate | More complex |



**Step 1**: Current state (account-wide)

```yaml

cloudflare-access:```yaml---

  name: Cloudflare Access

  providers:roles:

    - cloudflare-prod

  inherits:  # Marketing team - only their zones## Understanding Permission Validation

    - Administrator  # Cloudflare Administrator role

  resources:  marketing-zones:

    allow:

      - account:*  # Account-wide    name: Marketing DNS ManagerWhen using `inherits` + `permissions` together, the system validates that requested permissions actually exist in the inherited roles:

  enabled: true

```    description: DNS and Analytics for marketing domains



**Step 2**: Identify actual usage    providers:### Example: Valid Permissions

- What zones do users actually access?

- What roles do they actually need?      - cloudflare-prod



**Step 3**: Create resource-scoped roles    inherits:```yaml

```yaml

cloudflare-access-locked:      - DNS        # Cloudflare DNS rolecloudflare-limited-dns:

  name: Cloudflare Limited Access

  providers:      - Analytics  # Cloudflare Analytics role  name: Limited DNS Editor

    - cloudflare-prod

  inherits:    resources:  providers:

    - DNS        # Cloudflare DNS role only

    - Analytics  # Cloudflare Analytics role      allow:    - cloudflare-prod

  resources:

    allow:        - zone:marketing.example.com  inherits:

      - zone:app.example.com  # Only zones they use

  enabled: true        - zone:blog.example.com    - DNS  # Provides permissions like: dns_records:read, dns_records:edit, zone:read, etc.

```

    scopes:  permissions:

**Step 4**: Gradual rollout

- Test with a small group first      groups:    allow:

- Monitor for access issues

- Adjust roles as needed        - oidc:marketing-team      - dns_records:read   # ✅ Valid - exists in DNS role

- Roll out to everyone

    enabled: true      - dns_records:edit   # ✅ Valid - exists in DNS role

---

      - analytics:read     # ⚠️  Warning logged - NOT in DNS role, skipped

## Best Practices

  # API team - their zones only  resources:

1. **Start Restrictive**: Use resource-scoped roles by default, only grant account-wide when needed

2. **Zone Ownership**: Map resources to team ownership  api-zones:    allow:

3. **Role Minimization**: Only grant required roles

4. **Regular Audits**: Review and revoke unused access    name: API Zone Manager      - zone:example.com

5. **Document Decisions**: Comment why specific roles were granted

6. **Use Workflows**: Require approvals for sensitive roles    description: DNS, Firewall, and Load Balancing for API zones```

7. **Time-Limited Access**: Use workflows with expiration for temporary needs

    providers:

---

      - cloudflare-prod**Result**: User gets only `dns_records:read` and `dns_records:edit`. The `analytics:read` permission is logged as a warning and skipped because it's not available in the DNS role.

## Quick Reference

    inherits:

### Use Account-Wide When:

- ✅ Full admin needed      - DNS           # Cloudflare DNS role### Example: Using Multiple Inherits for Permission Pool

- ✅ Small, trusted team

- ✅ Simple access patterns      - Firewall      # Cloudflare Firewall role

- ✅ Development/staging environments

      - Load Balancer # Cloudflare Load Balancer role```yaml

### Use Resource-Scoped When:

- ✅ Multiple teams/zones      - Analytics     # Cloudflare Analytics rolecloudflare-multi-role:

- ✅ Need least privilege

- ✅ Production environments    resources:  name: Multi-Role Access

- ✅ Compliance requirements

- ✅ Multi-tenant scenarios      allow:  providers:

- ✅ Per-application access

        - zone:api.example.com    - cloudflare-prod

---

        - zone:api-staging.example.com  inherits:

## Common Patterns

    scopes:    - DNS        # Provides: dns_records:*, zone:read, etc.

### Pattern 1: Read-Only Monitoring

```yaml      groups:    - Firewall   # Provides: firewall_services:*, waf:*, etc.

cloudflare-monitoring:

  name: Monitoring Access        - oidc:api-team    - Analytics  # Provides: analytics:*, logs:read, etc.

  inherits:

    - Administrator Read Only  # Cloudflare read-only admin    enabled: true  permissions:

    - Analytics                # Cloudflare Analytics

  resources:    allow:

    allow:

      - zone:*  # Security team - all zones, security features only      - dns_records:edit      # ✅ From DNS role

```

  security-all-zones:      - firewall_services:edit # ✅ From Firewall role

### Pattern 2: DNS Only

```yaml    name: Security Team Access      - analytics:read        # ✅ From Analytics role

cloudflare-dns:

  name: DNS Management    description: Firewall and monitoring across all zones  resources:

  inherits:

    - DNS        # Cloudflare DNS role    providers:    allow:

    - Analytics  # Cloudflare Analytics role

  resources:      - cloudflare-prod      - zone:example.com

    allow:

      - zone:example.com    inherits:```

```

      - Firewall              # Cloudflare Firewall role

### Pattern 3: Security Operations

```yaml      - Cloudflare Zero Trust # Cloudflare Zero Trust role**Result**: All three permissions are valid because they exist in their respective inherited roles. The system builds a permission pool from all inherited roles.

cloudflare-security:

  name: Security Operations      - Analytics             # Cloudflare Analytics role

  inherits:

    - Firewall                  # Cloudflare Firewall role      - Audit Logs Viewer     # Cloudflare Audit Logs Viewer role---

    - Cloudflare Zero Trust     # Cloudflare Zero Trust role

    - Audit Logs Viewer         # Cloudflare Audit Logs Viewer    resources:

  resources:

    allow:      allow:## Real-World Examples

      - zone:*

      - account:*        - zone:*      # All zones for security monitoring

```

        - account:*   # Account-level for Zero Trust### Scenario 1: Small Team - Use Account-Wide Roles

### Pattern 4: Application Development

```yaml    scopes:

cloudflare-app-dev:

  name: Application Development      groups:**Situation**: Small startup with 5 engineers, everyone needs access to everything

  inherits:

    - Workers Platform Admin  # Cloudflare Workers Platform Admin        - oidc:security-team

    - DNS                     # Cloudflare DNS role

    - Cache Purge             # Cloudflare Cache Purge role    enabled: true```yaml

  resources:

    allow:```roles:

      - account:*  # For Workers

      - zone:app.example.com  team-admin:

```

**Why**: Isolation between teams, least privilege, audit-friendly    name: Team Administrator

    description: Full access for all team members

---    providers:

      - cloudflare-prod

### Scenario 3: Hybrid Approach    scopes:

      groups:

**Situation**: Most users need specific access, but some need full admin        - oidc:engineers

    enabled: true

```yaml```

roles:

  # Account-wide for admins**Why**: Simple, fast, everyone trusted with full access

  platform-admin:

    name: Platform Administrator---

    description: Full account access for platform team

    providers:### Scenario 2: Large Organization - Use Resource-Scoped Policies

      - cloudflare-prod

    inherits:**Situation**: Large company with multiple teams managing different zones

      - Administrator  # Cloudflare Administrator role

    scopes:```yaml

      groups:roles:

        - oidc:platform-admins  # Marketing team - only their zones, simple role assignment

    enabled: true  marketing-zones:

    name: Marketing DNS Manager

  # Resource-scoped for everyone else    description: DNS and Analytics for marketing domains

  developer-zones:    providers:

    name: Developer Zone Access      - cloudflare-prod

    description: DNS and cache for dev zones    inherits:

    providers:      - DNS        # Provides all DNS permissions

      - cloudflare-prod      - Analytics  # Provides all Analytics permissions

    inherits:    resources:

      - DNS          # Cloudflare DNS role      allow:

      - Cache Purge  # Cloudflare Cache Purge role        - zone:marketing.example.com

      - Analytics    # Cloudflare Analytics role        - zone:blog.example.com

    resources:    scopes:

      allow:      groups:

        - zone:dev.example.com        - oidc:marketing-team

        - zone:staging.example.com    enabled: true

    scopes:

      groups:  # API team - granular permission control

        - oidc:developers  api-zones:

    enabled: true    name: API Zone Manager

```    description: Specific DNS and firewall permissions for API zones

    providers:

**Why**: Best of both worlds - simple for admins, restricted for others      - cloudflare-prod

    inherits:

---      - DNS           # Provides DNS permissions to choose from

      - Firewall      # Provides Firewall permissions to choose from

## Emergency Access Roles      - Load Balancer # Provides Load Balancer permissions

    permissions:

For on-call engineers and incident response:      allow:

        - dns_records:read

```yaml        - dns_records:edit

roles:        - firewall_services:read

  cloudflare-oncall:        - firewall_services:edit

    name: On-Call Engineer Access        - load_balancers:read

    description: Emergency access for on-call engineers across all zones        - load_balancers:edit

    providers:        - analytics:read

      - cloudflare-prod    resources:

    workflows:      allow:

      - oncall_verification  # Auto-approve if on-call schedule        - zone:api.example.com

    inherits:        - zone:api-staging.example.com

      - DNS                       # Cloudflare DNS role    scopes:

      - Cache Purge               # Cloudflare Cache Purge role      groups:

      - Analytics                 # Cloudflare Analytics role        - oidc:api-team

      - Administrator Read Only   # Cloudflare read-only for monitoring    enabled: true

    resources:

      allow:  # Security team - all zones, security features only with deny rules

        - zone:*  # All zones for incident management  security-all-zones:

    scopes:    name: Security Team Access

      groups:    description: Firewall and monitoring across all zones (no configuration changes)

        - oidc:oncall    providers:

        - oidc:sre      - cloudflare-prod

    enabled: true    inherits:

```      - Firewall              # Provides Firewall permissions

      - Cloudflare Zero Trust # Provides Zero Trust permissions

---      - Analytics             # Provides Analytics permissions

    permissions:

## Customer/Tenant-Specific Roles      allow:

        - firewall_services:read

For multi-tenant or managed service scenarios:        - waf:read

        - analytics:read

```yaml        - logs:read

roles:      deny:

  cloudflare-customer-a:        - firewall_services:edit  # Can view but not modify

    name: DNS Editor - Customer A        - waf:edit                # Can view but not modify

    description: DNS management for Customer A's zones    resources:

    providers:      allow:

      - cloudflare-prod        - zone:*  # All zones for security monitoring

    workflows:    scopes:

      - customer_approval      groups:

    inherits:        - oidc:security-team

      - DNS        # Cloudflare DNS role    enabled: true

      - Analytics  # Cloudflare Analytics role```

    resources:

      allow:**Why**: Isolation between teams, granular least privilege, audit-friendly

        - zone:customer-a.com

        - zone:www.customer-a.com---

        - zone:api.customer-a.com

    scopes:### Scenario 3: Hybrid Approach

      users:

        - customer-a-admin@example.com**Situation**: Most users need specific access, but some need full admin

        - customer-a-lead@example.com

    enabled: true```yaml

```roles:

  # Account-wide for admins

---  platform-admin:

    name: Platform Administrator

## Development/Testing Roles    description: Full account access for platform team

    providers:

For development and staging environments:      - cloudflare-prod

    scopes:

```yaml      groups:

roles:        - oidc:platform-admins

  cloudflare-dev-fullaccess:    enabled: true

    name: Development Full Access

    description: Combined role access for full development environment access  # Resource-scoped for everyone else

    providers:  developer-zones:

      - cloudflare-dev    name: Developer Zone Access

      - cloudflare-staging    description: DNS and cache for dev zones

    workflows:    providers:

      - self_service  # Instant access in dev      - cloudflare-prod

    inherits:    inherits:

      - DNS                                                           # Cloudflare DNS role      - DNS          # DNS permissions

      - Firewall                                                      # Cloudflare Firewall role      - Cache Purge  # Cache purge permissions

      - Workers Platform Admin                                        # Cloudflare Workers Platform Admin role      - Analytics    # Analytics access

      - SSL/TLS, Caching, Performance, Page Rules, and Customization # Cloudflare SSL/TLS role    resources:

      - Cache Purge                                                   # Cloudflare Cache Purge role      allow:

      - Analytics                                                     # Cloudflare Analytics role        - zone:dev.example.com

    resources:        - zone:staging.example.com

      allow:    scopes:

        - zone:*  # All zones in dev account      groups:

    scopes:        - oidc:developers

      groups:    enabled: true

        - oidc:developers```

        - oidc:qa-team

    enabled: true**Why**: Best of both worlds - simple for admins, restricted for others

```

---

---

## Migration Path

## Migration Path

### From Account-Wide to Resource-Scoped

### From Account-Wide to Resource-Scoped

If you start with account-wide roles and want to lock down access:

If you start with account-wide roles and want to lock down access:

**Step 1**: Current state (account-wide)

**Step 1**: Current state (account-wide)```yaml

```yamlcloudflare-access:

cloudflare-access:  name: Cloudflare Access

  name: Cloudflare Access  providers:

  providers:    - cloudflare-prod

    - cloudflare-prod  enabled: true

  inherits:```

    - Administrator  # Cloudflare Administrator role

  enabled: true**Step 2**: Identify actual usage

```- What zones do users actually access?

- What permissions do they actually need?

**Step 2**: Identify actual usage

- What zones do users actually access?**Step 3**: Create resource-scoped roles

- What roles do they actually need?```yaml

cloudflare-access-locked:

**Step 3**: Create resource-scoped roles  name: Cloudflare Limited Access

```yaml  providers:

cloudflare-access-locked:    - cloudflare-prod

  name: Cloudflare Limited Access  inherits:

  providers:    - DNS        # Only DNS permissions

    - cloudflare-prod    - Analytics  # And analytics

  inherits:  resources:

    - DNS        # Cloudflare DNS role only    allow:

    - Analytics  # Cloudflare Analytics role      - zone:app.example.com  # Only zones they use

  resources:  enabled: true

    allow:```

      - zone:app.example.com  # Only zones they use

  enabled: true**Step 4**: Gradual rollout

```- Test with a small group first

- Monitor for access issues

**Step 4**: Gradual rollout- Adjust permissions as needed

- Test with a small group first- Roll out to everyone

- Monitor for access issues

- Adjust roles as needed---

- Roll out to everyone

## Best Practices

---

1. **Start Restrictive**: Use resource-scoped policies by default, only grant account-wide when needed

## Best Practices2. **Zone Ownership**: Map resources to team ownership

3. **Permission Minimization**: Only grant required permissions

1. **Start Restrictive**: Use resource-scoped roles by default, only grant account-wide when needed4. **Regular Audits**: Review and revoke unused access

2. **Zone Ownership**: Map resources to team ownership5. **Document Decisions**: Comment why specific permissions were granted

3. **Role Minimization**: Only grant required roles6. **Use Workflows**: Require approvals for sensitive permissions

4. **Regular Audits**: Review and revoke unused access7. **Time-Limited Access**: Use workflows with expiration for temporary needs

5. **Document Decisions**: Comment why specific roles were granted

6. **Use Workflows**: Require approvals for sensitive roles---

7. **Time-Limited Access**: Use workflows with expiration for temporary needs

## Quick Reference

---

### Use Account-Wide When:

## Quick Reference- ✅ Full admin needed

- ✅ Small, trusted team

### Use Account-Wide When:- ✅ Simple access patterns

- ✅ Full admin needed- ✅ Development/staging environments

- ✅ Small, trusted team

- ✅ Simple access patterns### Use Resource-Scoped When:

- ✅ Development/staging environments- ✅ Multiple teams/zones

- ✅ Need least privilege

### Use Resource-Scoped When:- ✅ Production environments

- ✅ Multiple teams/zones- ✅ Compliance requirements

- ✅ Need least privilege- ✅ Multi-tenant scenarios

- ✅ Production environments- ✅ Per-application access

- ✅ Compliance requirements
- ✅ Multi-tenant scenarios
- ✅ Per-application access

---

## Common Patterns

### Pattern 1: Read-Only Monitoring
```yaml
cloudflare-monitoring:
  name: Monitoring Access
  inherits:
    - Administrator Read Only  # Cloudflare read-only admin
    - Analytics                # Cloudflare Analytics
  resources:
    allow:
      - zone:*
```

### Pattern 2: DNS Only
```yaml
cloudflare-dns:
  name: DNS Management
  inherits:
    - DNS        # Cloudflare DNS role
    - Analytics  # Cloudflare Analytics role
  resources:
    allow:
      - zone:example.com
```

### Pattern 3: Security Operations
```yaml
cloudflare-security:
  name: Security Operations
  inherits:
    - Firewall                  # Cloudflare Firewall role
    - Cloudflare Zero Trust     # Cloudflare Zero Trust role
    - Audit Logs Viewer         # Cloudflare Audit Logs Viewer
  resources:
    allow:
      - zone:*
      - account:*
```

### Pattern 4: Application Development
```yaml
cloudflare-app-dev:
  name: Application Development
  inherits:
    - Workers Platform Admin  # Cloudflare Workers Platform Admin
    - DNS                     # Cloudflare DNS role
    - Cache Purge             # Cloudflare Cache Purge role
  resources:
    allow:
      - account:*  # For Workers
      - zone:app.example.com
```
