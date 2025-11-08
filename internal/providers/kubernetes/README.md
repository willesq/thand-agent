# Kubernetes Provider

This provider implements RBAC (Role-Based Access Control) for Kubernetes clusters, supporting both namespace-scoped and cluster-wide permissions.

## Configuration

```yaml
providers:
  - name: kubernetes-prod
    description: Production Kubernetes cluster
    provider: kubernetes
    config:
      # No configuration needed - uses kubeconfig or in-cluster config
    enabled: true
```

## Authentication

The provider automatically detects the environment and uses:

1. **In-cluster configuration** when running inside a Kubernetes cluster
2. **Kubeconfig file** when running outside (uses `KUBECONFIG` env var or `~/.kube/config`)

## Service Account Setup

When deploying thand-agent inside Kubernetes, create a service account with proper RBAC permissions:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: thand-agent
  namespace: thand-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: thand-agent
rules:
- apiGroups: ["rbac.authorization.k8s.io"]
  resources: ["roles", "rolebindings", "clusterroles", "clusterrolebindings"]
  verbs: ["get", "list", "create", "update", "delete"]
- apiGroups: [""]
  resources: ["namespaces"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: thand-agent
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: thand-agent
subjects:
- kind: ServiceAccount
  name: thand-agent
  namespace: thand-system
```

## Permission Discovery

The provider **dynamically discovers** available permissions from the Kubernetes API server using the discovery API. This means:

- ✅ Always up-to-date with your cluster's available resources
- ✅ Includes custom resources and operators
- ✅ Automatically detects API versions and groups
- ✅ Fallback to static list if discovery fails

## Permission Format

Kubernetes permissions follow the format: `k8s:[apiGroup/]resource:verb`

Examples:
- `k8s:pods:get` - Read access to pods
- `k8s:apps/deployments:create` - Create deployments
- `k8s:secrets:list` - List secrets
- `k8s:rbac.authorization.k8s.io/roles:update` - Update roles
- `k8s:networking.k8s.io/ingresses:delete` - Delete ingresses

## Role Scoping

### Namespace-Scoped Roles
To create namespace-scoped roles, include a namespace resource in the role definition:

```yaml
version: "1.0"
roles:
  dev-pod-reader:
    name: Dev Pod Reader
    description: Read pods in development namespace
    resources:
      allow:
        - "namespace:development"
    permissions:
      allow:
        - "k8s:pods:get,list,watch"
        - "k8s:services:get,list"
```

### Cluster-Wide Roles
For cluster-wide access, omit namespace resources:

```yaml
version: "1.0"
roles:
  cluster-viewer:
    description: View resources across all namespaces
    permissions:
      allow:
        - "k8s:pods:get,list,watch"
        - "k8s:namespaces:get,list"
```

## Built-in Roles

The provider includes several built-in roles based on Kubernetes defaults:

- `view` - Read-only access to most objects in a namespace
- `edit` - Read/write access to most objects in a namespace  
- `admin` - Full admin access within a namespace
- `cluster-admin` - Full admin access across the entire cluster
- `pod-reader` - Read-only access to pods
- `deployment-manager` - Manage deployments and related resources
- `secret-manager` - Manage secrets and config maps
- `network-admin` - Manage networking resources
- `storage-admin` - Manage storage resources

## Security Features

### Permission Validation
The provider validates permissions to prevent privilege escalation:

- Blocks dangerous wildcard permissions like `k8s:*:*` for cluster roles
- Validates cluster-level resource access
- Restricts access to RBAC resources

### User Identification
Users are identified by:
1. Email (preferred for OIDC integration)
2. Username (fallback)

### Resource Naming
Kubernetes resource names are automatically sanitized:
- Email domains converted: `user@example.com` → `user-at-example-com`
- Special characters replaced with hyphens
- Names converted to lowercase

## Example Usage

1. **Grant namespace access:**
```bash
thand elevate --provider kubernetes-prod --role dev-pod-reader --user john@company.com
```

2. **Grant cluster access:**
```bash
thand elevate --provider kubernetes-prod --role cluster-viewer --user admin@company.com
```

3. **Revoke access:**
```bash
thand revoke --provider kubernetes-prod --role dev-pod-reader --user john@company.com
```

## Integration with OIDC

For production use, integrate with OIDC providers:

1. Configure Kubernetes API server with OIDC
2. Users authenticate via OIDC provider
3. Thand manages RBAC bindings using user emails from OIDC tokens

## Troubleshooting

### Permission Denied
- Ensure thand-agent has proper ClusterRole permissions
- Check if the target namespace exists
- Verify user email/username format

### Role Not Found
- Check if the role is defined in your configuration
- Verify role name matches exactly (case-sensitive)

### In-Cluster vs External
- In-cluster: Uses service account token automatically
- External: Requires valid kubeconfig file with cluster access