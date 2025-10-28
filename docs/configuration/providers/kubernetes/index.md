---
layout: default
title: Kubernetes Provider
description: Kubernetes provider for cluster authentication and RBAC
parent: Providers
grand_parent: Configuration
---

# Kubernetes Provider

The Kubernetes provider enables integration with Kubernetes clusters for authentication and role-based access control.

## Capabilities

- **Authentication**: Kubernetes cluster authentication
- **RBAC Integration**: Kubernetes Role-Based Access Control
- **Cluster Management**: Access to multiple Kubernetes clusters
- **Service Account Support**: Integration with Kubernetes service accounts

## Configuration Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `kubeconfig` | string | No | Path to kubeconfig file |
| `context` | string | No | Kubernetes context to use |
| `namespace` | string | No | Default namespace |
| `cluster_url` | string | No | Kubernetes API server URL |
| `token` | string | No | Service account token |

## Example Configuration

```yaml
version: "1.0"
providers:
  kubernetes:
    name: Kubernetes
    description: Kubernetes cluster access
    provider: kubernetes
    enabled: true
    config:
      kubeconfig: /path/to/kubeconfig
      context: production-cluster
      namespace: default
```

For detailed setup instructions, refer to the [Kubernetes documentation](https://kubernetes.io/docs/).